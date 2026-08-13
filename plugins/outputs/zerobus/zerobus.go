//go:generate ../../../tools/readme_config_includer/generator
package zerobus

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"strings"
	"time"

	sdkzerobus "github.com/databricks/zerobus-sdk/purego/zerobus"
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

const (
	// Zerobus caps a request at 10 MiB and 100k records, so the plugin splits
	// batches along the protocol limits instead of the agent's metric_batch_size.
	maxBatchRecords = 100_000
	maxPayloadBytes = 10*1024*1024 - 64*1024
	// Bytes left for records after reserving headroom for the surrounding
	// request fields.
	maxRecordBytes = maxPayloadBytes - 1024

	defaultConnectTimeout = 30 * time.Second
)

// Zerobus writes metrics to a Databricks table.
type Zerobus struct {
	ServerEndpoint    string          `toml:"server_endpoint"`
	WorkspaceURL      string          `toml:"workspace_url"`
	TableName         string          `toml:"table_name"`
	ClientID          string          `toml:"client_id"`
	ClientSecret      config.Secret   `toml:"client_secret"`
	ApplicationName   string          `toml:"application_name"`
	TimestampColumn   string          `toml:"timestamp_column"`
	MeasurementColumn string          `toml:"measurement_column"`
	ConnectTimeout    config.Duration `toml:"connect_timeout"`
	Log               telegraf.Logger `toml:"-"`

	sdk     *sdkzerobus.SDK
	stream  *sdkzerobus.Stream
	columns map[string]struct{}
}

func (*Zerobus) SampleConfig() string {
	return sampleConfig
}

// Init the output plugin.
func (z *Zerobus) Init() error {
	z.TimestampColumn = strings.TrimSpace(z.TimestampColumn)
	z.MeasurementColumn = strings.TrimSpace(z.MeasurementColumn)

	required := []struct {
		name  string
		value string
	}{
		{"server_endpoint", z.ServerEndpoint},
		{"workspace_url", z.WorkspaceURL},
		{"table_name", z.TableName},
		{"client_id", z.ClientID},
	}
	for _, option := range required {
		if strings.TrimSpace(option.value) == "" {
			return fmt.Errorf("option %q must be set", option.name)
		}
	}
	if z.ClientSecret.Empty() {
		return errors.New(`option "client_secret" must be set`)
	}
	if z.MeasurementColumn != "" && z.MeasurementColumn == z.TimestampColumn {
		return errors.New(`options "measurement_column" and "timestamp_column" must be different`)
	}
	if z.ConnectTimeout < 0 {
		return errors.New(`option "connect_timeout" cannot be negative`)
	}
	if z.ConnectTimeout == 0 {
		z.ConnectTimeout = config.Duration(defaultConnectTimeout)
	}

	return nil
}

// Connect to the Zerobus server.
func (z *Zerobus) Connect() error {
	applicationName := internal.ProductToken()
	if name := strings.TrimSpace(z.ApplicationName); name != "" {
		applicationName += " " + name
	}
	sdk, err := sdkzerobus.New(z.ServerEndpoint, z.WorkspaceURL, sdkzerobus.WithApplicationName(applicationName))
	if err != nil {
		return fmt.Errorf("creating Zerobus SDK failed: %w", err)
	}
	z.sdk = sdk

	// Open the stream during startup so authentication, permission and network
	// errors surface before the first metric is written.
	if err := z.openStream(); err != nil {
		startupErr := &internal.StartupError{Err: err, Retry: sdkzerobus.Retryable(err)}
		return errors.Join(startupErr, z.Close())
	}

	return nil
}

// Write the metrics to the Zerobus server.
func (z *Zerobus) Write(metrics []telegraf.Metric) error {
	if len(metrics) == 0 {
		return nil
	}
	if z.sdk == nil {
		return internal.ErrNotConnected
	}

	// Telegraf connects only once, so a stream lost after startup is replaced here.
	if z.stream == nil || z.stream.IsClosed() {
		if err := z.openStream(); err != nil {
			return err
		}
	}

	records, result := z.serializeMetrics(metrics)
	if len(records) == 0 {
		return result
	}

	chunks, err := chunkRecords(records, maxBatchRecords, maxRecordBytes)
	if err != nil {
		return err
	}
	for _, chunk := range chunks {
		if _, err := z.stream.IngestJSONRecordsOffset(chunk); err != nil {
			return z.abortWrite("admitting batch", err)
		}
	}
	// Report success only once Databricks acknowledged every record of the batch.
	if err := z.stream.Flush(); err != nil {
		return z.abortWrite("flushing batch", err)
	}

	return result
}

// Close the Zerobus connection.
func (z *Zerobus) Close() error {
	errs := []error{z.closeStream()}
	if z.sdk != nil {
		errs = append(errs, z.sdk.Close())
		z.sdk = nil
	}
	return errors.Join(errs...)
}

func (z *Zerobus) openStream() error {
	secret, err := z.ClientSecret.Get()
	if err != nil {
		return fmt.Errorf("resolving client secret failed: %w", err)
	}
	defer secret.Destroy()

	timeout := time.Duration(z.ConnectTimeout)

	fetchCtx, fetchCancel := context.WithTimeout(context.Background(), timeout)
	// The descriptor is fetched per stream, so a stream opened after an ALTER
	// TABLE picks up the new columns without restarting Telegraf.
	descriptor, err := z.sdk.FetchProtoDescriptorFromUC(fetchCtx, z.TableName, z.ClientID, secret.String())
	fetchCancel()
	if err != nil {
		return fmt.Errorf("fetching schema of table %q failed: %w", z.TableName, err)
	}
	columns, err := columnsFromDescriptor(descriptor)
	if err != nil {
		return fmt.Errorf("reading schema of table %q failed: %w", z.TableName, err)
	}

	streamCtx, streamCancel := context.WithTimeout(context.Background(), timeout)
	z.stream, err = z.sdk.CreateStream(streamCtx, z.TableName, z.ClientID, secret.String(),
		sdkzerobus.WithProto(descriptor),
		sdkzerobus.WithWaitForReady(),
		sdkzerobus.WithRecovery(sdkzerobus.RecoveryDisabled),
		// Pin the stream to the limits the plugin chunks against.
		sdkzerobus.WithMaxBatchRecords(maxBatchRecords),
		sdkzerobus.WithMaxPayloadBytes(maxPayloadBytes),
	)
	streamCancel()
	if err != nil {
		return fmt.Errorf("creating stream for table %q failed: %w", z.TableName, err)
	}
	z.columns = columns
	z.Log.Debugf("Opened stream to table %q", z.TableName)

	return nil
}

func (z *Zerobus) closeStream() error {
	if z.stream == nil {
		return nil
	}
	err := z.stream.Close()
	z.stream = nil
	z.columns = nil
	return err
}

// Drop the stream and let Telegraf retry the whole batch on a new one. Records
// already acknowledged by the failed attempt may therefore be written twice.
func (z *Zerobus) abortWrite(operation string, err error) error {
	if closeErr := z.closeStream(); closeErr != nil {
		z.Log.Debugf("Closing the failed stream: %s", closeErr)
	}
	return fmt.Errorf("%s failed (retryable=%t): %w", operation, sdkzerobus.Retryable(err), err)
}

// Serialize the metrics and reject the ones that cannot be encoded or do not fit
// a single request, so the rest of the batch can still be written.
func (z *Zerobus) serializeMetrics(metrics []telegraf.Metric) ([][]byte, error) {
	records := make([][]byte, 0, len(metrics))
	var writeErr internal.PartialWriteError

	for i, m := range metrics {
		record, err := metricToTableSchemaJSON(m, z.TimestampColumn, z.MeasurementColumn, z.columns)
		if err == nil {
			if size := recordSize(record); size > maxRecordBytes {
				err = fmt.Errorf("serialized metric requires %d bytes, exceeding the request limit of %d bytes", size, maxRecordBytes)
			}
		}
		if err != nil {
			writeErr.MetricsReject = append(writeErr.MetricsReject, i)
			writeErr.MetricsRejectErrors = append(writeErr.MetricsRejectErrors, err)
			continue
		}
		records = append(records, record)
		writeErr.MetricsAccept = append(writeErr.MetricsAccept, i)
	}

	if len(writeErr.MetricsReject) == 0 {
		return records, nil
	}
	writeErr.Err = fmt.Errorf("rejected %d metric(s): %w", len(writeErr.MetricsReject), errors.Join(writeErr.MetricsRejectErrors...))
	return records, &writeErr
}

// Split the records into requests within the Zerobus record-count and size
// limits. Oversized records are rejected during serialization, so every chunk
// holds at least one record.
func chunkRecords(records [][]byte, maxRecords, maxBytes int) ([][][]byte, error) {
	chunks := make([][][]byte, 0, (len(records)+maxRecords-1)/maxRecords)
	for len(records) > 0 {
		count, size := 0, 0
		for count < len(records) && count < maxRecords {
			next := size + recordSize(records[count])
			if next > maxBytes {
				break
			}
			size = next
			count++
		}
		if count == 0 {
			return nil, fmt.Errorf("serialized record exceeds the request limit of %d bytes", maxBytes)
		}
		chunks = append(chunks, records[:count])
		records = records[count:]
	}
	return chunks, nil
}

func columnsFromDescriptor(raw []byte) (map[string]struct{}, error) {
	var descriptor descriptorpb.DescriptorProto
	if err := proto.Unmarshal(raw, &descriptor); err != nil {
		return nil, fmt.Errorf("parsing protobuf descriptor failed: %w", err)
	}
	columns := make(map[string]struct{}, len(descriptor.Field))
	for _, field := range descriptor.Field {
		if name := field.GetName(); name != "" {
			columns[name] = struct{}{}
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
			ConnectTimeout:  config.Duration(defaultConnectTimeout),
		}
	})
}
