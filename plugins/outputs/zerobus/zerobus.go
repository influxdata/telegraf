//go:generate ../../../tools/readme_config_includer/generator
package zerobus

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"time"

	"github.com/databricks/zerobus-sdk/purego/zerobus"
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/outputs"
)

//go:embed sample.conf
var sampleConfig string

// Zerobus writes metrics to a Databricks table.
type Zerobus struct {
	Endpoint          string          `toml:"endpoint"`
	Workspace         string          `toml:"workspace"`
	Table             string          `toml:"table"`
	ClientID          string          `toml:"client_id"`
	ClientSecret      config.Secret   `toml:"client_secret"`
	Application       string          `toml:"application"`
	TimestampColumn   string          `toml:"timestamp_column"`
	MeasurementColumn string          `toml:"measurement_column"`
	Timeout           config.Duration `toml:"timeout"`
	Log               telegraf.Logger `toml:"-"`

	sdk     *zerobus.SDK
	stream  *zerobus.Stream
	columns map[string]bool

	// Request limits the metrics are split along, lowered in tests
	maxRecords int
	maxBytes   int
}

func (*Zerobus) SampleConfig() string {
	return sampleConfig
}

func (z *Zerobus) Init() error {
	if z.Endpoint == "" {
		return errors.New(`option "endpoint" must be set`)
	}
	if z.Workspace == "" {
		return errors.New(`option "workspace" must be set`)
	}
	if z.Table == "" {
		return errors.New(`option "table" must be set`)
	}
	if z.ClientID == "" {
		return errors.New(`option "client_id" must be set`)
	}
	if z.ClientSecret.Empty() {
		return errors.New(`option "client_secret" must be set`)
	}
	if z.TimestampColumn != "" && z.MeasurementColumn != "" && z.TimestampColumn == z.MeasurementColumn {
		return errors.New(`options "measurement_column" and "timestamp_column" must be different`)
	}
	if z.Timeout < 0 {
		return errors.New(`option "timeout" cannot be negative`)
	}

	// Zerobus caps a request at 10 MiB and 100k records, so the plugin splits
	// batches along the protocol limits instead of the agent's metric_batch_size.
	// The byte budget applies to all records of a request together and reserves
	// headroom for the surrounding request fields.
	z.maxRecords = 100000
	z.maxBytes = 10*1024*1024 - 64*1024 - 1024

	return nil
}

func (z *Zerobus) Connect() error {
	applicationName := internal.ProductToken()
	if z.Application != "" {
		applicationName = z.Application
	}
	sdk, err := zerobus.New(z.Endpoint, z.Workspace, zerobus.WithApplicationName(applicationName))
	if err != nil {
		return fmt.Errorf("creating Zerobus SDK failed: %w", err)
	}
	z.sdk = sdk

	// Open the stream during startup so authentication, permission and network
	// errors surface before the first metric is written.
	if err := z.openStream(); err != nil {
		if closeErr := z.Close(); closeErr != nil {
			z.Log.Debugf("Closing after failed startup: %s", closeErr)
		}
		return &internal.StartupError{
			Err:   err,
			Retry: zerobus.Retryable(err),
		}
	}

	return nil
}

func (z *Zerobus) Write(metrics []telegraf.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	// Telegraf connects only once, so a stream lost after startup is replaced here.
	if z.stream == nil || z.stream.IsClosed() {
		if err := z.openStream(); err != nil {
			return err
		}
	}

	// Only log the metrics rejected during serialization for now. A failing write
	// below returns an error for the whole batch, so Telegraf retries them
	// together with the valid metrics until the write succeeds.
	records, err := z.serializeMetrics(metrics)
	if err != nil {
		z.Log.Errorf("Serializing metrics failed: %s", err)
	}
	if len(records) == 0 {
		return nil
	}

	// Drop the stream on failure and let Telegraf retry the whole batch on a new
	// one. Records already acknowledged by the failed attempt may therefore be
	// written twice.
	for _, batch := range z.batchRecords(records) {
		if _, err := z.stream.IngestJSONRecordsOffset(batch); err != nil {
			z.closeStream()
			return fmt.Errorf("admitting batch failed (retryable=%t): %w", zerobus.Retryable(err), err)
		}
	}
	// Report success only once Databricks acknowledged every record of the batch.
	if err := z.stream.Flush(); err != nil {
		z.closeStream()
		return fmt.Errorf("flushing batch failed (retryable=%t): %w", zerobus.Retryable(err), err)
	}

	return nil
}

func (z *Zerobus) Close() error {
	z.closeStream()
	if z.sdk == nil {
		return nil
	}
	err := z.sdk.Close()
	z.sdk = nil
	return err
}

func (z *Zerobus) openStream() error {
	secret, err := z.ClientSecret.Get()
	if err != nil {
		return fmt.Errorf("resolving client secret failed: %w", err)
	}
	defer secret.Destroy()

	// A zero timeout means the startup requests are not bounded.
	ctx := context.Background()
	if z.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(z.Timeout))
		defer cancel()
	}

	// The descriptor is fetched per stream, so a stream opened after an ALTER
	// TABLE picks up the new columns without restarting Telegraf.
	descriptor, err := z.sdk.FetchProtoDescriptorFromUC(ctx, z.Table, z.ClientID, secret.String())
	if err != nil {
		return fmt.Errorf("fetching schema of table %q failed: %w", z.Table, err)
	}

	stream, err := z.sdk.CreateStream(ctx, z.Table, z.ClientID, secret.String(),
		zerobus.WithProto(descriptor),
		zerobus.WithWaitForReady(),
		zerobus.WithRecovery(zerobus.RecoveryDisabled),
	)
	if err != nil {
		return fmt.Errorf("creating stream for table %q failed: %w", z.Table, err)
	}
	z.stream = stream

	columns, err := columnsFromDescriptor(descriptor)
	if err != nil {
		z.closeStream()
		return fmt.Errorf("reading schema of table %q failed: %w", z.Table, err)
	}
	z.columns = columns
	z.Log.Debugf("Opened stream to table %q", z.Table)

	return nil
}

func (z *Zerobus) closeStream() {
	if z.stream == nil {
		return
	}
	if err := z.stream.Close(); err != nil {
		z.Log.Debugf("Closing the stream to table %q: %s", z.Table, err)
	}
	z.stream = nil
	z.columns = nil
}

// Serialize the metrics to the records of the destination table. Metrics that
// cannot be encoded or do not fit a single request are rejected, so the rest of
// the batch can still be written.
func (z *Zerobus) serializeMetrics(metrics []telegraf.Metric) ([][]byte, error) {
	records := make([][]byte, 0, len(metrics))
	var writeErr internal.PartialWriteError

	for i, m := range metrics {
		// Serialize the metric to a record according to the table columns
		record, err := metricToTableSchemaJSON(m, z.TimestampColumn, z.MeasurementColumn, z.columns)
		if err != nil {
			writeErr.MetricsReject = append(writeErr.MetricsReject, i)
			writeErr.MetricsRejectErrors = append(writeErr.MetricsRejectErrors, err)
			continue
		}

		// Reject records that exceed the payload size of a whole request
		if size := recordSize(record); size > z.maxBytes {
			writeErr.MetricsReject = append(writeErr.MetricsReject, i)
			err := fmt.Errorf("serialized metric requires %d bytes, exceeding the request limit of %d bytes", size, z.maxBytes)
			writeErr.MetricsRejectErrors = append(writeErr.MetricsRejectErrors, err)
			continue
		}

		records = append(records, record)
	}

	// Report the rejected metrics, the accepted ones are only known once the
	// endpoint acknowledged the records
	if len(writeErr.MetricsReject) == 0 {
		return records, nil
	}
	writeErr.Err = fmt.Errorf("rejected %d metric(s): %w", len(writeErr.MetricsReject), errors.Join(writeErr.MetricsRejectErrors...))
	return records, &writeErr
}

// Group the records into requests within the record-count and size limits. Each
// request holds at least one record as oversized ones are rejected during
// serialization.
func (z *Zerobus) batchRecords(records [][]byte) [][][]byte {
	var batches [][][]byte

	for len(records) > 0 {
		// Add records to the request until one of the limits is reached
		count, size := 1, recordSize(records[0])
		for count < len(records) && count < z.maxRecords {
			next := size + recordSize(records[count])
			if next > z.maxBytes {
				break
			}
			size = next
			count++
		}
		batches = append(batches, records[:count])
		records = records[count:]
	}

	return batches
}

// columnsFromDescriptor extracts the column names of the destination table.
// TODO: Remove this once the Zerobus SDK exposes the columns itself.
func columnsFromDescriptor(raw []byte) (map[string]bool, error) {
	var descriptor descriptorpb.DescriptorProto
	if err := proto.Unmarshal(raw, &descriptor); err != nil {
		return nil, fmt.Errorf("parsing protobuf descriptor failed: %w", err)
	}
	columns := make(map[string]bool, len(descriptor.Field))
	for _, field := range descriptor.Field {
		if name := field.GetName(); name != "" {
			columns[name] = true
		}
	}
	if len(columns) == 0 {
		return nil, errors.New("table schema descriptor has no columns")
	}
	return columns, nil
}

// Bytes a record occupies in the request, including its protobuf framing.
func recordSize(record []byte) int {
	return protowire.SizeTag(1) + protowire.SizeBytes(len(record))
}

func init() {
	outputs.Add("zerobus", func() telegraf.Output {
		return &Zerobus{
			TimestampColumn: "timestamp",
			Timeout:         config.Duration(30 * time.Second),
		}
	})
}
