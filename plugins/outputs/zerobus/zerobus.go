//go:generate ../../../tools/readme_config_includer/generator
//go:generate protoc --go_out=. --go_opt=paths=source_relative metric.proto
package zerobus

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"

	sdkzerobus "github.com/databricks/zerobus-sdk/purego/zerobus"
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/outputs"
)

//go:embed sample.conf
var sampleConfig string

const (
	defaultMaxBatchRecords         = 100_000
	defaultMaxPayloadBytes         = 10*1024*1024 - 64*1024
	defaultMaxBufferedPayloadBytes = 64 * 1024 * 1024
	defaultConnectTimeout          = 30 * time.Second
	batchEnvelopeReserve           = 1024
	bufferedRequestOverhead        = 512
	bufferedRecordOverhead         = 32
	maxConcurrentStreams           = 100
	schemaModeStatic               = "static"
	schemaModeTableSchema          = "table_schema"
)

// Zerobus writes metrics to a Databricks table.
type Zerobus struct {
	ServerEndpoint  string        `toml:"server_endpoint"`
	WorkspaceURL    string        `toml:"workspace_url"`
	TableName       string        `toml:"table_name"`
	ClientID        string        `toml:"client_id"`
	ClientSecret    config.Secret `toml:"client_secret"`
	ApplicationName string        `toml:"application_name"`
	SchemaMode      string        `toml:"schema_mode"`

	TimestampColumn   string          `toml:"timestamp_column"`
	MeasurementColumn string          `toml:"measurement_column"`
	ConnectTimeout    config.Duration `toml:"connect_timeout"`

	ConcurrentStreams int             `toml:"concurrent_streams"`
	RecoveryRetries   int             `toml:"recovery_retries"`
	LackOfAckTimeout  config.Duration `toml:"lack_of_ack_timeout"`
	FlushTimeout      config.Duration `toml:"flush_timeout"`

	Log telegraf.Logger `toml:"-"`

	sdk       sdkClient
	writers   []*writer
	newSDK    sdkFactory
	original  [][]byte
	confirmed [][]byte

	batchRecordLimit  int
	payloadByteLimit  int
	bufferedByteLimit int64

	descriptorMu     sync.Mutex
	descriptor       []byte
	descriptorReused bool
}

// One stream and the write state that survives a failed flush. Writers are independent, so they can be flushed concurrently.
type writer struct {
	stream  ingestStream
	pending *pendingWrite
}

type ingestStream interface {
	IngestRecordsOffset(records [][]byte, encoded bool) (int64, error)
	Flush() error
	GetUnackedBatches() ([][][]byte, error)
	IsClosed() bool
	Close() error
}

type pendingWrite struct {
	admitted  []recordBatch
	remaining []recordBatch
	waiting   bool
}

type recordBatch struct {
	records [][]byte
	encoded bool
}

type preparedWrite struct {
	records      [][]byte
	accept       []int
	reject       []int
	rejectErrors []error
}

type sdkClient interface {
	CreateStaticSchemaStream(ctx context.Context, tableName, clientID, clientSecret string, opts ...sdkzerobus.StreamOption) (ingestStream, error)
	CreateTableSchemaStream(ctx context.Context, tableName, clientID, clientSecret string, opts ...sdkzerobus.StreamOption) (ingestStream, error)
	FetchProtoDescriptor(ctx context.Context, tableName, clientID, clientSecret string) ([]byte, error)
	Close() error
}

type sdkFactory func(serverEndpoint, workspaceURL string, opts ...sdkzerobus.Option) (sdkClient, error)

type sdkAdapter struct {
	*sdkzerobus.SDK
}

func (s *sdkAdapter) CreateStaticSchemaStream(
	ctx context.Context, tableName, clientID, clientSecret string, opts ...sdkzerobus.StreamOption,
) (ingestStream, error) {
	stream, err := s.SDK.CreateStream(ctx, tableName, clientID, clientSecret, opts...)
	if err != nil {
		return nil, err
	}
	return &staticStreamAdapter{Stream: stream}, nil
}

func (s *sdkAdapter) CreateTableSchemaStream(
	ctx context.Context, tableName, clientID, clientSecret string, opts ...sdkzerobus.StreamOption,
) (ingestStream, error) {
	stream, err := s.SDK.CreateStream(ctx, tableName, clientID, clientSecret, opts...)
	if err != nil {
		return nil, err
	}
	return &tableSchemaStreamAdapter{Stream: stream}, nil
}

func (s *sdkAdapter) FetchProtoDescriptor(ctx context.Context, tableName, clientID, clientSecret string) ([]byte, error) {
	return s.SDK.FetchProtoDescriptorFromUC(ctx, tableName, clientID, clientSecret)
}

type staticStreamAdapter struct {
	*sdkzerobus.Stream
}

func (s *staticStreamAdapter) IngestRecordsOffset(records [][]byte, _ bool) (int64, error) {
	return s.Stream.IngestRecordsOffset(records)
}

type tableSchemaStreamAdapter struct {
	*sdkzerobus.Stream
}

func (s *tableSchemaStreamAdapter) IngestRecordsOffset(records [][]byte, encoded bool) (int64, error) {
	if encoded {
		return s.Stream.IngestRecordsOffset(records)
	}
	return s.Stream.IngestJSONRecordsOffset(records)
}

func (*Zerobus) SampleConfig() string {
	return sampleConfig
}

// Init the output plugin.
func (z *Zerobus) Init() error {
	// Normalize the settings
	z.SchemaMode = strings.ToLower(strings.TrimSpace(z.SchemaMode))
	z.TimestampColumn = strings.TrimSpace(z.TimestampColumn)
	z.MeasurementColumn = strings.TrimSpace(z.MeasurementColumn)
	if z.SchemaMode == "" {
		z.SchemaMode = schemaModeStatic
	}
	if z.SchemaMode != schemaModeStatic && z.SchemaMode != schemaModeTableSchema {
		return fmt.Errorf(`option "schema_mode" must be %q or %q`, schemaModeStatic, schemaModeTableSchema)
	}
	if z.SchemaMode == schemaModeTableSchema && z.MeasurementColumn != "" && z.MeasurementColumn == z.TimestampColumn {
		return errors.New(`options "measurement_column" and "timestamp_column" must be different`)
	}
	// Check the required settings
	requiredStrings := []struct {
		name  string
		value string
	}{
		{"server_endpoint", z.ServerEndpoint},
		{"workspace_url", z.WorkspaceURL},
		{"table_name", z.TableName},
		{"client_id", z.ClientID},
	}
	for _, option := range requiredStrings {
		if strings.TrimSpace(option.value) == "" {
			return fmt.Errorf("option %q must be set", option.name)
		}
	}
	if z.ClientSecret.Empty() {
		return errors.New(`option "client_secret" must be set`)
	}

	// Check the streaming and recovery tuning settings
	if z.ConcurrentStreams < 0 {
		return errors.New(`option "concurrent_streams" cannot be negative`)
	}
	if z.ConcurrentStreams > maxConcurrentStreams {
		return fmt.Errorf(`option "concurrent_streams" must not exceed %d`, maxConcurrentStreams)
	}
	if z.ConcurrentStreams == 0 {
		z.ConcurrentStreams = 1
	}
	if z.RecoveryRetries < 0 {
		return errors.New(`option "recovery_retries" cannot be negative`)
	}
	if z.ConnectTimeout < 0 {
		return errors.New(`option "connect_timeout" cannot be negative`)
	}
	if z.ConnectTimeout == 0 {
		z.ConnectTimeout = config.Duration(defaultConnectTimeout)
	}
	if z.LackOfAckTimeout < 0 {
		return errors.New(`option "lack_of_ack_timeout" cannot be negative`)
	}
	if z.FlushTimeout < 0 {
		return errors.New(`option "flush_timeout" cannot be negative`)
	}

	// Zerobus caps a request at 10 MiB and the SDK bounds what it allocates and buffers per stream,
	// so these limits track the protocol rather than the agent's metric_batch_size
	z.batchRecordLimit = defaultMaxBatchRecords
	z.payloadByteLimit = defaultMaxPayloadBytes
	z.bufferedByteLimit = defaultMaxBufferedPayloadBytes

	// Setup the SDK constructor unless a test injected one
	if z.newSDK == nil {
		z.newSDK = func(serverEndpoint, workspaceURL string, opts ...sdkzerobus.Option) (sdkClient, error) {
			sdk, err := sdkzerobus.New(serverEndpoint, workspaceURL, opts...)
			if err != nil {
				return nil, err
			}
			return &sdkAdapter{SDK: sdk}, nil
		}
	}
	return nil
}

// Connect to the Zerobus server.
func (z *Zerobus) Connect() error {
	secret, err := z.ClientSecret.Get()
	if err != nil {
		return fmt.Errorf("resolving client secret failed: %w", err)
	}
	defer secret.Destroy()

	// Setup the SDK client
	applicationName := internal.ProductToken()
	if name := strings.TrimSpace(z.ApplicationName); name != "" {
		applicationName += " " + name
	}
	sdk, err := z.newSDK(z.ServerEndpoint, z.WorkspaceURL, sdkzerobus.WithApplicationName(applicationName))
	if err != nil {
		return fmt.Errorf("creating Zerobus SDK failed: %w", err)
	}
	z.sdk = sdk

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(z.ConnectTimeout))
	defer cancel()
	// Open the streams; the first one fetches the table schema descriptor and the rest reuse it
	writers := make([]*writer, 0, z.ConcurrentStreams)
	for range z.ConcurrentStreams {
		w := &writer{}
		if err := z.openStream(ctx, w, secret.String()); err != nil {
			// Hand the streams opened so far to Close to clean up the partial connection.
			z.writers = writers
			closeErr := z.Close()
			startupErr := &internal.StartupError{
				Err:   fmt.Errorf("creating Zerobus stream failed: %w", err),
				Retry: sdkzerobus.Retryable(err),
			}
			return errors.Join(startupErr, closeErr)
		}
		writers = append(writers, w)
	}
	z.writers = writers
	z.Log.Debugf("Opened %d stream(s) to table %q in %s schema mode", len(writers), z.TableName, z.SchemaMode)

	return nil
}

// Write the metrics to the Zerobus server.
func (z *Zerobus) Write(metrics []telegraf.Metric) error {
	if len(metrics) == 0 {
		return nil
	}
	if !z.connected() {
		return internal.ErrNotConnected
	}

	// Resume the writers that a previous attempt left mid-write.
	if z.hasPending() {
		pendingOriginal := z.original
		if err := z.flushWriters(); err != nil {
			return err
		}
		z.confirmed = pendingOriginal
		z.original = nil
	}

	// Serialize the metrics, rejecting the ones that cannot be encoded or do not fit the budgets
	prepared := z.prepareMetrics(metrics)
	records := prepared.records
	if len(records) == 0 {
		z.confirmed = nil
		return prepared.result()
	}
	original := records

	// Telegraf retries the whole batch, so drop the leading records a previous attempt already acknowledged.
	if z.confirmed != nil {
		if recordsHavePrefix(records, z.confirmed) {
			z.Log.Debugf("Skipping %d record(s) acknowledged by the previous attempt", len(z.confirmed))
			records = records[len(z.confirmed):]
			if len(records) == 0 {
				z.confirmed = nil
				return prepared.result()
			}
		} else {
			z.confirmed = nil
		}
	}

	if err := z.assignRecords(records); err != nil {
		return err
	}
	z.original = original
	z.confirmed = nil
	if err := z.flushWriters(); err != nil {
		return err
	}
	z.original = nil
	return prepared.result()
}

// Report whether a stream is open or a pending write can still be resumed.
func (z *Zerobus) connected() bool {
	for _, w := range z.writers {
		if w.stream != nil {
			return true
		}
	}
	return z.sdk != nil && z.hasPending()
}

func (z *Zerobus) hasPending() bool {
	for _, w := range z.writers {
		if w.pending != nil {
			return true
		}
	}
	return false
}

func (z *Zerobus) assignRecords(records [][]byte) error {
	shares := partitionRecords(records, len(z.writers))
	chunked := make([][]recordBatch, len(shares))
	for i, share := range shares {
		chunks, err := z.chunkRecords(share)
		if err != nil {
			return err
		}
		chunked[i] = chunks
	}
	for i, chunks := range chunked {
		if len(chunks) > 0 {
			z.writers[i].pending = &pendingWrite{remaining: chunks}
		}
	}
	return nil
}

// Shares are contiguous, so every stream keeps the batch order within its own slice.
func partitionRecords(records [][]byte, count int) [][][]byte {
	if count <= 1 {
		return [][][]byte{records}
	}
	size := (len(records) + count - 1) / count
	shares := make([][][]byte, 0, count)
	for start := 0; start < len(records); start += size {
		shares = append(shares, records[start:min(start+size, len(records))])
	}
	return shares
}

// Telegraf keeps the whole batch when any writer fails, so a retry resumes the writers that did not finish and skips what the others acknowledged.
func (z *Zerobus) flushWriters() error {
	pending := make([]*writer, 0, len(z.writers))
	for _, w := range z.writers {
		if w.pending != nil {
			pending = append(pending, w)
		}
	}
	if len(pending) == 1 {
		return z.processPending(pending[0])
	}

	// Each writer reports into its own slot, keeping the joined error stable.
	var wg sync.WaitGroup
	errs := make([]error, len(pending))
	for i, w := range pending {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs[i] = z.processPending(w)
		}()
	}
	wg.Wait()
	return errors.Join(errs...)
}

// Close the Zerobus connection.
func (z *Zerobus) Close() error {
	errs := make([]error, 0, len(z.writers)+1)
	for _, w := range z.writers {
		if w.stream != nil {
			errs = append(errs, w.stream.Close())
			w.stream = nil
		}
	}
	z.writers = nil
	z.original = nil
	z.confirmed = nil
	z.descriptor = nil
	z.descriptorReused = false
	if z.sdk != nil {
		errs = append(errs, z.sdk.Close())
		z.sdk = nil
	}
	return errors.Join(errs...)
}

func (z *Zerobus) openStream(ctx context.Context, w *writer, clientSecret string) error {
	options := z.streamOptions()
	var (
		stream ingestStream
		err    error
	)
	if z.SchemaMode == schemaModeTableSchema {
		descriptor, fetchErr := z.tableDescriptor(ctx, clientSecret)
		if fetchErr != nil {
			return fetchErr
		}
		options = append([]sdkzerobus.StreamOption{sdkzerobus.WithProto(descriptor)}, options...)
		stream, err = z.sdk.CreateTableSchemaStream(ctx, z.TableName, z.ClientID, clientSecret, options...)
	} else {
		descriptor, descriptorErr := messageDescriptor()
		if descriptorErr != nil {
			return fmt.Errorf("building protobuf descriptor failed: %w", descriptorErr)
		}
		options = append([]sdkzerobus.StreamOption{sdkzerobus.WithProto(descriptor)}, options...)
		stream, err = z.sdk.CreateStaticSchemaStream(ctx, z.TableName, z.ClientID, clientSecret, options...)
	}
	if err != nil {
		return err
	}
	w.stream = stream
	return nil
}

// Fetch the descriptor only when no reusable one is cached, so concurrent openers share a single fetch.
func (z *Zerobus) tableDescriptor(ctx context.Context, clientSecret string) ([]byte, error) {
	z.descriptorMu.Lock()
	defer z.descriptorMu.Unlock()
	if z.descriptor != nil {
		return z.descriptor, nil
	}
	descriptor, err := z.sdk.FetchProtoDescriptor(ctx, z.TableName, z.ClientID, clientSecret)
	if err != nil {
		return nil, err
	}
	z.descriptor = descriptor
	z.descriptorReused = false
	return descriptor, nil
}

func (z *Zerobus) recreateStream(w *writer) error {
	unacked, err := w.stream.GetUnackedBatches()
	if err != nil {
		return z.writeError("retrieving unacknowledged batches", err)
	}
	z.Log.Warnf("Stream is closed, recreating it and replaying %d unacknowledged batch(es)", len(unacked))
	_ = w.stream.Close()
	w.stream = nil
	// Table-schema records replay as JSON, because the new stream may encode them against a newer descriptor.
	if z.SchemaMode == schemaModeTableSchema {
		z.ageDescriptor()
		if len(unacked) > len(w.pending.admitted) {
			return fmt.Errorf("rebuilding table-schema replay batches failed: SDK returned %d unacknowledged batches for %d admitted batches",
				len(unacked), len(w.pending.admitted))
		}
		if len(unacked) > 0 {
			start := len(w.pending.admitted) - len(unacked)
			replay := slices.Clone(w.pending.admitted[start:])
			w.pending.remaining = append(replay, w.pending.remaining...)
		}
	} else {
		replay := make([]recordBatch, 0, len(unacked)+len(w.pending.remaining))
		for _, batch := range unacked {
			replay = append(replay, recordBatch{records: batch, encoded: true})
		}
		w.pending.remaining = append(replay, w.pending.remaining...)
	}
	w.pending.admitted = nil
	w.pending.waiting = false

	return z.openStreamFromSecret(w)
}

// The first recreation reuses the cached descriptor and the second discards it, in case the table schema changed.
func (z *Zerobus) ageDescriptor() {
	z.descriptorMu.Lock()
	defer z.descriptorMu.Unlock()
	switch {
	case z.descriptor == nil:
	case z.descriptorReused:
		z.Log.Debug("Discarding the cached table schema descriptor, the next stream will refetch it")
		z.descriptor = nil
		z.descriptorReused = false
	default:
		z.descriptorReused = true
	}
}

// A stream that accepted records proves the descriptor it opened with is current.
func (z *Zerobus) freshenDescriptor() {
	z.descriptorMu.Lock()
	defer z.descriptorMu.Unlock()
	z.descriptorReused = false
}

func (z *Zerobus) openStreamFromSecret(w *writer) error {
	secret, err := z.ClientSecret.Get()
	if err != nil {
		return fmt.Errorf("resolving client secret for stream recovery failed: %w", err)
	}
	defer secret.Destroy()
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(z.ConnectTimeout))
	defer cancel()
	if err := z.openStream(ctx, w, secret.String()); err != nil {
		return z.writeError("recreating stream", err)
	}
	return nil
}

func (z *Zerobus) processPending(w *writer) error {
	if w.stream == nil {
		if err := z.openStreamFromSecret(w); err != nil {
			return err
		}
	}
	if w.stream.IsClosed() {
		if err := z.recreateStream(w); err != nil {
			return err
		}
	}
	if w.pending.waiting {
		if err := w.stream.Flush(); err != nil {
			return z.writeError("flushing previously admitted batch", err)
		}
		w.pending.waiting = false
		w.pending.admitted = nil
	}

	for len(w.pending.remaining) > 0 {
		chunk := w.pending.remaining[0]
		if _, err := w.stream.IngestRecordsOffset(chunk.records, chunk.encoded); err != nil {
			return z.writeError("admitting batch", err)
		}
		w.pending.remaining = w.pending.remaining[1:]
		w.pending.admitted = append(w.pending.admitted, chunk)
		w.pending.waiting = true
	}

	if w.pending.waiting {
		if err := w.stream.Flush(); err != nil {
			return z.writeError("flushing batch", err)
		}
	}
	z.freshenDescriptor()
	w.pending = nil
	return nil
}

func (z *Zerobus) chunkRecords(records [][]byte) ([]recordBatch, error) {
	payloadBudget := z.payloadByteLimit - batchEnvelopeReserve
	chunks := make([]recordBatch, 0, (len(records)+z.batchRecordLimit-1)/z.batchRecordLimit)
	for len(records) > 0 {
		count, size := 0, 0
		for count < len(records) && count < z.batchRecordLimit {
			recordSize := protowire.SizeTag(1) + protowire.SizeBytes(len(records[count]))
			if err := z.validateRecordSize(recordSize, payloadBudget); err != nil {
				return nil, fmt.Errorf("serialized metric %d cannot be admitted: %w", count, err)
			}
			if size+recordSize > payloadBudget {
				break
			}
			if retainedPayloadSize(size+recordSize, count+1) > z.bufferedByteLimit {
				break
			}
			size += recordSize
			count++
		}
		chunks = append(chunks, recordBatch{records: records[:count]})
		records = records[count:]
	}
	return chunks, nil
}

func (z *Zerobus) prepareMetrics(metrics []telegraf.Metric) preparedWrite {
	prepared := preparedWrite{
		records: make([][]byte, 0, len(metrics)),
		accept:  make([]int, 0, len(metrics)),
	}
	payloadBudget := z.payloadByteLimit - batchEnvelopeReserve

	for i, metric := range metrics {
		record, err := z.serializeMetric(metric)
		if err == nil {
			recordSize := protowire.SizeTag(1) + protowire.SizeBytes(len(record))
			err = z.validateRecordSize(recordSize, payloadBudget)
		}
		if err != nil {
			prepared.reject = append(prepared.reject, i)
			prepared.rejectErrors = append(prepared.rejectErrors, err)
			continue
		}
		prepared.records = append(prepared.records, record)
		prepared.accept = append(prepared.accept, i)
	}
	return prepared
}

func (p *preparedWrite) result() error {
	if len(p.reject) == 0 {
		return nil
	}
	return &internal.PartialWriteError{
		Err:                 fmt.Errorf("rejected %d metric(s): %w", len(p.reject), errors.Join(p.rejectErrors...)),
		MetricsAccept:       p.accept,
		MetricsReject:       p.reject,
		MetricsRejectErrors: p.rejectErrors,
	}
}

func (z *Zerobus) validateRecordSize(recordSize, payloadBudget int) error {
	if recordSize > payloadBudget {
		return fmt.Errorf("requires %d bytes, exceeding the payload budget of %d bytes", recordSize, payloadBudget)
	}
	if retained := retainedPayloadSize(recordSize, 1); retained > z.bufferedByteLimit {
		return fmt.Errorf("requires approximately %d buffered bytes, exceeding the buffer limit of %d bytes", retained, z.bufferedByteLimit)
	}
	return nil
}

// Approximate the bytes the SDK buffers for a request, including its per-record overhead.
func retainedPayloadSize(recordBytes, recordCount int) int64 {
	return int64(recordBytes) + int64(recordCount)*bufferedRecordOverhead + bufferedRequestOverhead
}

func serializeStaticMetric(metric telegraf.Metric) ([]byte, error) {
	record, err := metricToProto(metric)
	if err != nil {
		return nil, fmt.Errorf("serializing metric failed: %w", err)
	}
	serialized, err := (proto.MarshalOptions{Deterministic: true}).Marshal(record)
	if err != nil {
		return nil, fmt.Errorf("marshaling protobuf record failed: %w", err)
	}
	return serialized, nil
}

func (z *Zerobus) serializeMetric(metric telegraf.Metric) ([]byte, error) {
	if z.SchemaMode == schemaModeTableSchema {
		return metricToTableSchemaJSON(metric, z.TimestampColumn, z.MeasurementColumn)
	}
	return serializeStaticMetric(metric)
}

func recordsHavePrefix(records, prefix [][]byte) bool {
	return len(records) >= len(prefix) && slices.EqualFunc(records[:len(prefix)], prefix, slices.Equal)
}

func (z *Zerobus) streamOptions() []sdkzerobus.StreamOption {
	// Pin the limits the plugin chunks against so a stream cannot accept less than a prepared batch
	options := []sdkzerobus.StreamOption{
		sdkzerobus.WithWaitForReady(),
		sdkzerobus.WithMaxBatchRecords(z.batchRecordLimit),
		sdkzerobus.WithMaxPayloadBytes(z.payloadByteLimit),
		sdkzerobus.WithMaxBufferedPayloadBytes(z.bufferedByteLimit),
	}
	if z.RecoveryRetries > 0 {
		options = append(options, sdkzerobus.WithRecoveryRetries(z.RecoveryRetries))
	}
	if z.LackOfAckTimeout > 0 {
		options = append(options, sdkzerobus.WithLackOfAckTimeout(time.Duration(z.LackOfAckTimeout)))
	}
	if z.FlushTimeout > 0 {
		options = append(options, sdkzerobus.WithFlushTimeout(time.Duration(z.FlushTimeout)))
	}
	return options
}

func (*Zerobus) writeError(operation string, err error) error {
	retryable := sdkzerobus.Retryable(err)
	return fmt.Errorf("%s failed (retryable=%t): %w", operation, retryable, err)
}

func messageDescriptor() ([]byte, error) {
	descriptor := protodesc.ToDescriptorProto((&TelegrafMetric{}).ProtoReflect().Descriptor())
	return proto.Marshal(descriptor)
}

func init() {
	outputs.Add("zerobus", func() telegraf.Output {
		return &Zerobus{
			SchemaMode:        schemaModeStatic,
			TimestampColumn:   "timestamp",
			ConnectTimeout:    config.Duration(defaultConnectTimeout),
			ConcurrentStreams: 1,
		}
	})
}
