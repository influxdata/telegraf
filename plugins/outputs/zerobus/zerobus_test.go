package zerobus

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"slices"
	"sync"
	"testing"
	"time"

	sdkzerobus "github.com/databricks/zerobus-sdk/purego/zerobus"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/testutil"
)

func TestDefaults(t *testing.T) {
	creator, found := outputs.Outputs["zerobus"]
	require.True(t, found)

	plugin, ok := creator().(*Zerobus)
	require.True(t, ok)
	require.Empty(t, plugin.ApplicationName)
	require.Equal(t, schemaModeStatic, plugin.SchemaMode)
	require.Equal(t, "timestamp", plugin.TimestampColumn)
	require.Equal(t, config.Duration(defaultConnectTimeout), plugin.ConnectTimeout)
	require.Equal(t, 1, plugin.ConcurrentStreams)
	require.NotEmpty(t, plugin.SampleConfig())
}

func TestInitSetsProtocolLimits(t *testing.T) {
	plugin := validPlugin()
	plugin.batchRecordLimit = 0
	plugin.payloadByteLimit = 0
	plugin.bufferedByteLimit = 0
	require.NoError(t, plugin.Init())
	require.Equal(t, defaultMaxBatchRecords, plugin.batchRecordLimit)
	require.Equal(t, defaultMaxPayloadBytes, plugin.payloadByteLimit)
	require.EqualValues(t, defaultMaxBufferedPayloadBytes, plugin.bufferedByteLimit)
}

func TestInitRequiredOptions(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Zerobus)
		option string
	}{
		{
			name:   "server endpoint",
			mutate: func(z *Zerobus) { z.ServerEndpoint = "" },
			option: "server_endpoint",
		},
		{
			name:   "workspace URL",
			mutate: func(z *Zerobus) { z.WorkspaceURL = "" },
			option: "workspace_url",
		},
		{
			name:   "table name",
			mutate: func(z *Zerobus) { z.TableName = "" },
			option: "table_name",
		},
		{
			name:   "client ID",
			mutate: func(z *Zerobus) { z.ClientID = "" },
			option: "client_id",
		},
		{
			name:   "client secret",
			mutate: func(z *Zerobus) { z.ClientSecret = config.Secret{} },
			option: "client_secret",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := validPlugin()
			tt.mutate(plugin)
			require.ErrorContains(t, plugin.Init(), tt.option)
		})
	}
}

func TestInitRejectsInvalidTuning(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Zerobus)
		option string
	}{
		{
			name:   "negative concurrent streams",
			mutate: func(z *Zerobus) { z.ConcurrentStreams = -1 },
			option: "concurrent_streams",
		},
		{
			name:   "too many concurrent streams",
			mutate: func(z *Zerobus) { z.ConcurrentStreams = maxConcurrentStreams + 1 },
			option: "concurrent_streams",
		},
		{
			name:   "negative recovery retries",
			mutate: func(z *Zerobus) { z.RecoveryRetries = -1 },
			option: "recovery_retries",
		},
		{
			name:   "negative ack timeout",
			mutate: func(z *Zerobus) { z.LackOfAckTimeout = -1 },
			option: "lack_of_ack_timeout",
		},
		{
			name:   "negative flush timeout",
			mutate: func(z *Zerobus) { z.FlushTimeout = -1 },
			option: "flush_timeout",
		},
		{
			name:   "negative connect timeout",
			mutate: func(z *Zerobus) { z.ConnectTimeout = -1 },
			option: "connect_timeout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := validPlugin()
			tt.mutate(plugin)
			require.ErrorContains(t, plugin.Init(), tt.option)
		})
	}
}

func TestInitSchemaMode(t *testing.T) {
	t.Run("normalizes mode", func(t *testing.T) {
		plugin := validPlugin()
		plugin.SchemaMode = " TABLE_SCHEMA "
		require.NoError(t, plugin.Init())
		require.Equal(t, schemaModeTableSchema, plugin.SchemaMode)
	})

	t.Run("rejects unknown mode", func(t *testing.T) {
		plugin := validPlugin()
		plugin.SchemaMode = "automatic"
		require.ErrorContains(t, plugin.Init(), "schema_mode")
	})

	t.Run("allows omitted timestamp column", func(t *testing.T) {
		plugin := validPlugin()
		plugin.SchemaMode = schemaModeTableSchema
		plugin.TimestampColumn = ""
		require.NoError(t, plugin.Init())
	})

	t.Run("rejects reserved column collision", func(t *testing.T) {
		plugin := validPlugin()
		plugin.SchemaMode = schemaModeTableSchema
		plugin.MeasurementColumn = plugin.TimestampColumn
		require.ErrorContains(t, plugin.Init(), "must be different")
	})
}

func TestMessageDescriptor(t *testing.T) {
	raw, err := messageDescriptor()
	require.NoError(t, err)

	var descriptor descriptorpb.DescriptorProto
	require.NoError(t, proto.Unmarshal(raw, &descriptor))
	require.Equal(t, "TelegrafMetric", descriptor.GetName())
	require.Len(t, descriptor.Field, 4)

	require.Equal(t, "measurement", descriptor.Field[0].GetName())
	require.Equal(t, int32(1), descriptor.Field[0].GetNumber())
	require.Equal(t, descriptorpb.FieldDescriptorProto_LABEL_REQUIRED, descriptor.Field[0].GetLabel())
	require.Equal(t, "timestamp_ns", descriptor.Field[1].GetName())
	require.Equal(t, int32(2), descriptor.Field[1].GetNumber())
	require.Equal(t, "tags", descriptor.Field[2].GetName())
	require.Equal(t, "fields", descriptor.Field[3].GetName())
	require.Equal(t, descriptorpb.FieldDescriptorProto_TYPE_STRING, descriptor.Field[3].GetType())
	require.Equal(t, descriptorpb.FieldDescriptorProto_LABEL_REQUIRED, descriptor.Field[3].GetLabel())

	require.Len(t, descriptor.NestedType, 1)
	require.Equal(
		t,
		"TagsEntry",
		descriptor.NestedType[0].GetName(),
		"tags must be nested so the descriptor is self-contained",
	)
}

func TestMetricToProtoEncodesFieldsAsVariantJSON(t *testing.T) {
	timestamp := time.Unix(1_700_000_000, 123)
	input := metric.New(
		"cpu",
		map[string]string{"host": "server-01", "region": "west"},
		map[string]interface{}{
			"z-string": "ready",
			"b-uint":   uint64(math.MaxInt64),
			"d-bool":   true,
			"a-int":    int64(-42),
			"c-float":  1.25,
		},
		timestamp,
	)

	record, err := metricToProto(input)
	require.NoError(t, err)
	require.Equal(t, "cpu", record.GetMeasurement())
	require.Equal(t, timestamp.UnixNano(), record.GetTimestampNs())
	require.Equal(t, map[string]string{"host": "server-01", "region": "west"}, record.GetTags())
	require.JSONEq(
		t,
		`{"a-int":-42,"b-uint":9223372036854775807,"c-float":1.25,"d-bool":true,`+
			`"z-string":"ready"}`,
		record.GetFields(),
	)
}

func TestMetricFieldsJSONSortsKeys(t *testing.T) {
	input := metric.New(
		"cpu",
		nil,
		map[string]interface{}{"z": int64(1), "a": int64(2), "m": int64(3)},
		time.Unix(1, 0),
	)

	fields, err := metricFieldsJSON(input)
	require.NoError(t, err)
	require.Equal(t, `{"a":2,"m":3,"z":1}`, string(fields))
}

func TestMetricToProtoAcceptsMetricWithoutFields(t *testing.T) {
	record, err := metricToProto(metric.New("cpu", nil, nil, time.Unix(1, 0)))
	require.NoError(t, err)
	require.Equal(t, "{}", record.GetFields())
}

func TestMetricToTableSchemaJSONFlattensMetric(t *testing.T) {
	input := metric.New(
		"cpu",
		map[string]string{"host": "server-01"},
		map[string]interface{}{
			"active": true,
			"count":  int64(-42),
			"ratio":  1.25,
			"status": "ready",
			"total":  uint64(math.MaxInt64),
		},
		time.Unix(1_700_000_000, 123_456_000),
	)

	record, err := metricToTableSchemaJSON(input, "event_time", "measurement")
	require.NoError(t, err)

	var values map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(record, &values))
	require.JSONEq(t, `"cpu"`, string(values["measurement"]))
	require.JSONEq(t, `"server-01"`, string(values["host"]))
	require.JSONEq(t, `1700000000123456`, string(values["event_time"]))
	require.JSONEq(t, `-42`, string(values["count"]))
	require.JSONEq(t, `9223372036854775807`, string(values["total"]))
	require.JSONEq(t, `1.25`, string(values["ratio"]))
	require.JSONEq(t, `true`, string(values["active"]))
	require.JSONEq(t, `"ready"`, string(values["status"]))

	record, err = metricToTableSchemaJSON(input, "", "")
	require.NoError(t, err)
	values = nil
	require.NoError(t, json.Unmarshal(record, &values))
	require.NotContains(t, values, "event_time")
	require.NotContains(t, values, "measurement")
}

func TestMetricToTableSchemaJSONRejectsInvalidMetric(t *testing.T) {
	tests := []struct {
		name   string
		metric telegraf.Metric
		match  string
	}{
		{
			name: "timestamp collision",
			metric: metric.New(
				"cpu",
				map[string]string{"timestamp": "tag"},
				map[string]interface{}{"value": 1.0},
				time.Now(),
			),
			match: `tag "timestamp" conflicts`,
		},
		{
			name: "tag and field collision",
			metric: metric.New(
				"cpu",
				map[string]string{"host": "tag"},
				map[string]interface{}{"host": "field"},
				time.Now(),
			),
			match: `field "host" conflicts`,
		},
		{
			name: "non-finite float",
			metric: metric.New(
				"cpu",
				nil,
				map[string]interface{}{"value": math.NaN()},
				time.Now(),
			),
			match: "non-finite float",
		},
		{
			name: "uint64 above BIGINT maximum",
			metric: metric.New(
				"cpu",
				nil,
				map[string]interface{}{"value": uint64(math.MaxInt64) + 1},
				time.Now(),
			),
			match: "exceeding Delta BIGINT maximum",
		},
		{
			name: "unsupported field",
			metric: metricWithFields{
				Metric: testutil.TestMetric(1),
				fields: []*telegraf.Field{
					{Key: "value", Value: []int{1}},
				},
			},
			match: "unsupported type []int",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := metricToTableSchemaJSON(tt.metric, "timestamp", "")
			require.ErrorContains(t, err, tt.match)
		})
	}
}

func TestFieldToVariantRejectsInvalidValues(t *testing.T) {
	tests := []struct {
		name  string
		field *telegraf.Field
		match string
	}{
		{
			name:  "unsupported type",
			field: &telegraf.Field{Key: "invalid", Value: []int{1}},
			match: "unsupported field type []int",
		},
		{
			name:  "uint64 above BIGINT maximum",
			field: &telegraf.Field{Key: "value", Value: uint64(math.MaxInt64) + 1},
			match: "exceeds Delta BIGINT maximum",
		},
		{
			name:  "non-finite float",
			field: &telegraf.Field{Key: "value", Value: math.Inf(1)},
			match: "non-finite float",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := fieldToVariant(tt.field)
			require.ErrorContains(t, err, tt.match)
		})
	}
}

func TestMetricToProtoRejectsNilField(t *testing.T) {
	input := metricWithFields{
		Metric: testutil.TestMetric(1),
		fields: []*telegraf.Field{
			nil,
			{Key: "valid", Value: int64(1)},
		},
	}
	_, err := metricToProto(input)
	require.ErrorContains(t, err, "contains a nil field")
}

func TestConnectPassesConfiguration(t *testing.T) {
	stream := &fakeStream{}
	sdk := &fakeSDK{stream: stream}
	plugin := validPlugin()
	plugin.ApplicationName = "telegraf-test"
	plugin.RecoveryRetries = 3
	plugin.LackOfAckTimeout = config.Duration(3 * time.Second)
	plugin.FlushTimeout = config.Duration(4 * time.Second)

	var gotServer, gotWorkspace string
	var sdkOptionCount int
	plugin.newSDK = func(server, workspace string, options ...sdkzerobus.Option) (sdkClient, error) {
		gotServer = server
		gotWorkspace = workspace
		sdkOptionCount = len(options)
		return sdk, nil
	}

	require.NoError(t, plugin.Init())
	require.NoError(t, plugin.Connect())
	require.Equal(t, plugin.ServerEndpoint, gotServer)
	require.Equal(t, plugin.WorkspaceURL, gotWorkspace)
	require.Equal(t, 1, sdkOptionCount)
	require.Equal(t, plugin.TableName, sdk.tableName)
	require.Equal(t, plugin.ClientID, sdk.clientID)
	require.Equal(t, "secret", sdk.clientSecret)
	require.Len(t, sdk.options, 8)
	require.Len(t, sdk.contexts, 1)
	_, hasDeadline := sdk.contexts[0].Deadline()
	require.True(t, hasDeadline)
	require.Same(t, stream, currentStream(plugin))
}

func TestConnectCreatesTableSchemaStream(t *testing.T) {
	stream := &fakeStream{}
	sdk := &fakeSDK{stream: stream}
	plugin := validPlugin()
	plugin.SchemaMode = schemaModeTableSchema

	var sdkOptionCount int
	plugin.newSDK = func(_ string, _ string, options ...sdkzerobus.Option) (sdkClient, error) {
		sdkOptionCount = len(options)
		return sdk, nil
	}

	require.NoError(t, plugin.Init())
	require.NoError(t, plugin.Connect())
	require.Equal(t, 1, sdkOptionCount)
	require.Zero(t, sdk.staticSchemaCalls)
	require.Equal(t, 1, sdk.tableSchemaCalls)
	require.Equal(t, 1, sdk.fetchCalls)
	require.Len(t, sdk.options, 5)
	require.Equal(t, []byte("descriptor"), plugin.descriptor)
	require.Same(t, stream, currentStream(plugin))
}

func TestConnectClosesSDKWhenStreamCreationFails(t *testing.T) {
	createErr := errors.New("create stream failed")
	sdk := &fakeSDK{createErr: createErr}
	plugin := validPlugin()
	plugin.newSDK = func(string, string, ...sdkzerobus.Option) (sdkClient, error) {
		return sdk, nil
	}

	require.NoError(t, plugin.Init())
	err := plugin.Connect()
	require.ErrorIs(t, err, createErr)
	var startupErr *internal.StartupError
	require.ErrorAs(t, err, &startupErr)
	require.Equal(t, sdkzerobus.Retryable(createErr), startupErr.Retry)
	require.Equal(t, 1, sdk.closeCalls)
}

func TestConnectReturnsSDKCreationError(t *testing.T) {
	createErr := errors.New("create SDK failed")
	plugin := validPlugin()
	plugin.newSDK = func(string, string, ...sdkzerobus.Option) (sdkClient, error) {
		return nil, createErr
	}

	require.NoError(t, plugin.Init())
	require.ErrorIs(t, plugin.Connect(), createErr)
}

func TestWriteBatchesAndFlushesOnce(t *testing.T) {
	stream := &fakeStream{}
	plugin := validPlugin()
	setStream(plugin, stream)
	metrics := []telegraf.Metric{
		metric.New(
			"cpu",
			map[string]string{"host": "a"},
			map[string]interface{}{"usage": 1.5},
			time.Unix(1, 2),
		),
		metric.New("mem", nil, map[string]interface{}{"used": uint64(7)}, time.Unix(3, 4)),
	}

	require.NoError(t, plugin.Write(metrics))
	require.Equal(t, 1, stream.ingestCalls)
	require.Equal(t, 1, stream.flushCalls)
	require.Len(t, stream.records, 2)

	var first TelegrafMetric
	require.NoError(t, proto.Unmarshal(stream.records[0], &first))
	require.Equal(t, "cpu", first.GetMeasurement())
	require.JSONEq(t, `{"usage":1.5}`, first.GetFields())
}

func TestWriteTableSchemaUsesJSON(t *testing.T) {
	stream := &fakeStream{}
	plugin := validPlugin()
	plugin.SchemaMode = schemaModeTableSchema
	plugin.MeasurementColumn = "measurement"
	setStream(plugin, stream)
	input := metric.New(
		"cpu",
		map[string]string{"host": "a"},
		map[string]interface{}{"usage": 1.5},
		time.Unix(1, 2_000),
	)

	require.NoError(t, plugin.Write([]telegraf.Metric{input}))
	require.Equal(t, 1, stream.ingestCalls)
	require.Equal(t, []bool{false}, stream.encoded)
	require.Len(t, stream.records, 1)
	require.JSONEq(
		t,
		`{"host":"a","measurement":"cpu","timestamp":1000002,"usage":1.5}`,
		string(stream.records[0]),
	)
}

func TestWriteIsDeterministic(t *testing.T) {
	stream := &fakeStream{}
	plugin := validPlugin()
	setStream(plugin, stream)
	input := metric.New(
		"cpu",
		map[string]string{"z": "last", "a": "first"},
		map[string]interface{}{"z": "last", "a": int64(1)},
		time.Unix(1, 2),
	)

	require.NoError(t, plugin.Write([]telegraf.Metric{input}))
	first := slices.Clone(stream.records[0])
	require.NoError(t, plugin.Write([]telegraf.Metric{input}))
	require.Equal(t, first, stream.records[0])
}

func TestWriteFailures(t *testing.T) {
	admissionErr := context.Canceled
	flushErr := errors.New("flush failed")

	t.Run("not connected", func(t *testing.T) {
		err := validPlugin().Write([]telegraf.Metric{testutil.TestMetric(1)})
		require.ErrorIs(t, err, internal.ErrNotConnected)
	})

	t.Run("batch is split by record count", func(t *testing.T) {
		stream := &fakeStream{}
		plugin := validPlugin()
		plugin.batchRecordLimit = 1
		setStream(plugin, stream)
		input := []telegraf.Metric{testutil.TestMetric(1), testutil.TestMetric(2)}
		require.NoError(t, plugin.Write(input))
		require.Equal(t, 2, stream.ingestCalls)
		require.Equal(t, 1, stream.flushCalls)
		require.Len(t, stream.batches, 2)
		require.Len(t, stream.batches[0], 1)
		require.Len(t, stream.batches[1], 1)
	})

	t.Run("unsupported field", func(t *testing.T) {
		stream := &fakeStream{}
		plugin := validPlugin()
		setStream(plugin, stream)
		input := metricWithFields{
			Metric: testutil.TestMetric(1),
			fields: []*telegraf.Field{
				{Key: "unsupported", Value: []int{1}},
			},
		}
		err := plugin.Write([]telegraf.Metric{input})
		require.ErrorContains(t, err, "unsupported field type []int")
		require.Zero(t, stream.ingestCalls)
	})

	t.Run("admission", func(t *testing.T) {
		stream := &fakeStream{ingestErr: admissionErr}
		plugin := validPlugin()
		setStream(plugin, stream)
		err := plugin.Write([]telegraf.Metric{testutil.TestMetric(1)})
		require.ErrorIs(t, err, admissionErr)
		require.ErrorContains(t, err, "retryable=false")
		require.Zero(t, stream.flushCalls)
	})

	t.Run("flush", func(t *testing.T) {
		stream := &fakeStream{flushErr: flushErr}
		plugin := validPlugin()
		setStream(plugin, stream)
		err := plugin.Write([]telegraf.Metric{testutil.TestMetric(1)})
		require.ErrorIs(t, err, flushErr)
		require.Equal(t, 1, stream.ingestCalls)
		require.Equal(t, 1, stream.flushCalls)
	})
}

func TestWriteEmptyBatchIsNoop(t *testing.T) {
	require.NoError(t, validPlugin().Write(nil))
}

func TestWriteRetriesFlushWithoutReadmitting(t *testing.T) {
	flushErr := errors.New("flush timed out")
	stream := &fakeStream{flushErrors: []error{flushErr, nil}}
	plugin := validPlugin()
	setStream(plugin, stream)
	input := []telegraf.Metric{testutil.TestMetric(1)}

	require.ErrorIs(t, plugin.Write(input), flushErr)
	require.NotNil(t, currentPending(plugin))
	require.Equal(t, 1, stream.ingestCalls)
	require.Equal(t, 1, stream.flushCalls)

	require.NoError(t, plugin.Write(input))
	require.Nil(t, currentPending(plugin))
	require.Equal(t, 1, stream.ingestCalls, "the admitted batch must not be admitted twice")
	require.Equal(t, 2, stream.flushCalls)
}

func TestWriteAdmitsOnlyNewSuffixOnAugmentedRetry(t *testing.T) {
	flushErr := errors.New("flush timed out")
	stream := &fakeStream{flushErrors: []error{flushErr, nil, nil}}
	plugin := validPlugin()
	setStream(plugin, stream)
	original := testutil.TestMetric(1)
	added := testutil.TestMetric(2)

	require.ErrorIs(t, plugin.Write([]telegraf.Metric{original}), flushErr)
	require.NoError(t, plugin.Write([]telegraf.Metric{original, added}))
	require.Equal(t, 2, stream.ingestCalls)
	require.Len(t, stream.batches, 2)

	expectedAdded, err := serializeStaticMetric(added)
	require.NoError(t, err)
	require.Equal(t, [][]byte{expectedAdded}, stream.batches[1])
	require.Equal(t, 3, stream.flushCalls)
}

func TestWritePreservesCumulativeIdentityAcrossSuffixFailure(t *testing.T) {
	firstFlushErr := errors.New("first flush timed out")
	suffixFlushErr := errors.New("suffix flush timed out")
	stream := &fakeStream{
		flushErrors: []error{firstFlushErr, nil, suffixFlushErr, nil},
	}
	plugin := validPlugin()
	setStream(plugin, stream)
	original := testutil.TestMetric(1)
	added := testutil.TestMetric(2)
	augmented := []telegraf.Metric{original, added}

	require.ErrorIs(t, plugin.Write([]telegraf.Metric{original}), firstFlushErr)
	require.ErrorIs(t, plugin.Write(augmented), suffixFlushErr)
	require.NoError(t, plugin.Write(augmented))

	require.Equal(t, 2, stream.ingestCalls)
	require.Len(t, stream.batches, 2)
	expectedAdded, err := serializeStaticMetric(added)
	require.NoError(t, err)
	require.Equal(t, [][]byte{expectedAdded}, stream.batches[1])
	require.Equal(t, 4, stream.flushCalls)
	require.Nil(t, currentPending(plugin))
}

func TestWriteReturnsPartialErrorAfterPendingWriteSucceeds(t *testing.T) {
	flushErr := errors.New("flush timed out")
	stream := &fakeStream{flushErrors: []error{flushErr, nil, nil}}
	plugin := validPlugin()
	setStream(plugin, stream)
	original := testutil.TestMetric(1)
	unsupported := metricWithFields{
		Metric: testutil.TestMetric(2),
		fields: []*telegraf.Field{
			{Key: "unsupported", Value: []int{1}},
		},
	}
	added := testutil.TestMetric(3)

	require.ErrorIs(t, plugin.Write([]telegraf.Metric{original}), flushErr)
	err := plugin.Write([]telegraf.Metric{original, unsupported})
	var writeErr *internal.PartialWriteError
	require.ErrorAs(t, err, &writeErr)
	require.Equal(t, []int{0}, writeErr.MetricsAccept)
	require.Equal(t, []int{1}, writeErr.MetricsReject)
	require.Nil(t, plugin.confirmed)

	require.NoError(t, plugin.Write([]telegraf.Metric{added}))
	require.Equal(t, 2, stream.ingestCalls)
	expectedAdded, err := serializeStaticMetric(added)
	require.NoError(t, err)
	require.Equal(t, [][]byte{expectedAdded}, stream.batches[1])
	require.Nil(t, plugin.confirmed)
}

func TestWriteResumesAfterPartialChunkAdmission(t *testing.T) {
	admissionErr := errors.New("temporary admission failure")
	stream := &fakeStream{ingestErrors: []error{nil, admissionErr, nil}}
	plugin := validPlugin()
	plugin.batchRecordLimit = 1
	setStream(plugin, stream)
	input := []telegraf.Metric{testutil.TestMetric(1), testutil.TestMetric(2)}

	require.ErrorIs(t, plugin.Write(input), admissionErr)
	require.NotNil(t, currentPending(plugin))
	require.Equal(t, 2, stream.ingestCalls)
	require.Equal(t, 0, stream.flushCalls)

	require.NoError(t, plugin.Write(input))
	require.Nil(t, currentPending(plugin))
	require.Equal(t, 3, stream.ingestCalls)
	require.Equal(t, 2, stream.flushCalls)
}

func TestConnectOpensConfiguredStreams(t *testing.T) {
	streams := []ingestStream{&fakeStream{}, &fakeStream{}, &fakeStream{}}
	sdk := &fakeSDK{streams: slices.Clone(streams)}
	plugin := validPlugin()
	plugin.ConcurrentStreams = 3
	plugin.newSDK = func(string, string, ...sdkzerobus.Option) (sdkClient, error) {
		return sdk, nil
	}

	require.NoError(t, plugin.Init())
	require.NoError(t, plugin.Connect())
	require.Equal(t, 3, sdk.staticSchemaCalls)
	require.Len(t, plugin.writers, 3)
	for i, w := range plugin.writers {
		require.Same(t, streams[i], w.stream)
	}
}

func TestConnectFetchesTableSchemaOnceForAllStreams(t *testing.T) {
	sdk := &fakeSDK{streams: []ingestStream{&fakeStream{}, &fakeStream{}, &fakeStream{}}}
	plugin := validPlugin()
	plugin.SchemaMode = schemaModeTableSchema
	plugin.ConcurrentStreams = 3
	plugin.newSDK = func(string, string, ...sdkzerobus.Option) (sdkClient, error) {
		return sdk, nil
	}

	require.NoError(t, plugin.Init())
	require.NoError(t, plugin.Connect())
	require.Equal(t, 3, sdk.tableSchemaCalls)
	require.Equal(t, 1, sdk.fetchCalls)
}

func TestConnectClosesOpenedStreamsWhenLaterStreamFails(t *testing.T) {
	createErr := errors.New("create stream failed")
	opened := &fakeStream{}
	sdk := &fakeSDK{
		streams:      []ingestStream{opened},
		createErrors: []error{nil, createErr},
	}
	plugin := validPlugin()
	plugin.ConcurrentStreams = 2
	plugin.newSDK = func(string, string, ...sdkzerobus.Option) (sdkClient, error) {
		return sdk, nil
	}

	require.NoError(t, plugin.Init())
	require.ErrorIs(t, plugin.Connect(), createErr)
	require.Equal(t, 1, opened.closeCalls)
	require.Equal(t, 1, sdk.closeCalls)
	require.Nil(t, plugin.writers)
}

func TestWriteSplitsRecordsAcrossStreams(t *testing.T) {
	first := &fakeStream{}
	second := &fakeStream{}
	plugin := validPlugin()
	plugin.writers = []*writer{{stream: first}, {stream: second}}
	input := []telegraf.Metric{
		testutil.TestMetric(1),
		testutil.TestMetric(2),
		testutil.TestMetric(3),
	}

	require.NoError(t, plugin.Write(input))

	// Three records over two streams leave the remainder on the first stream.
	records := make([][]byte, 0, len(input))
	for _, m := range input {
		record, err := serializeStaticMetric(m)
		require.NoError(t, err)
		records = append(records, record)
	}
	require.Equal(t, 1, first.ingestCalls)
	require.Equal(t, 1, second.ingestCalls)
	require.Equal(t, records[:2], first.batches[0])
	require.Equal(t, records[2:], second.batches[0])
	require.Equal(t, 1, first.flushCalls)
	require.Equal(t, 1, second.flushCalls)
}

func TestWriteUsesOneStreamForSingleRecord(t *testing.T) {
	first := &fakeStream{}
	second := &fakeStream{}
	plugin := validPlugin()
	plugin.writers = []*writer{{stream: first}, {stream: second}}

	require.NoError(t, plugin.Write([]telegraf.Metric{testutil.TestMetric(1)}))
	require.Equal(t, 1, first.ingestCalls)
	require.Zero(t, second.ingestCalls)
	require.Zero(t, second.flushCalls)
}

func TestWriteKeepsBatchWhenOneStreamFails(t *testing.T) {
	flushErr := errors.New("flush timed out")
	first := &fakeStream{}
	second := &fakeStream{flushErrors: []error{flushErr, nil}}
	plugin := validPlugin()
	plugin.writers = []*writer{{stream: first}, {stream: second}}
	input := []telegraf.Metric{testutil.TestMetric(1), testutil.TestMetric(2)}

	require.ErrorIs(t, plugin.Write(input), flushErr)
	require.Nil(t, plugin.writers[0].pending)
	require.NotNil(t, plugin.writers[1].pending)

	// Telegraf keeps the whole batch, so the retry resumes the stream that did
	// not finish and re-admits nothing on the stream that already succeeded.
	require.NoError(t, plugin.Write(input))
	require.Nil(t, plugin.writers[1].pending)
	require.Equal(t, 1, first.ingestCalls)
	require.Equal(t, 1, second.ingestCalls)
	require.Equal(t, 2, second.flushCalls)
}

func TestWriteRecreatesTerminalStreamsConcurrently(t *testing.T) {
	firstClosed := &fakeStream{closed: true}
	secondClosed := &fakeStream{closed: true}
	replacements := []ingestStream{&fakeStream{}, &fakeStream{}}
	sdk := &fakeSDK{streams: slices.Clone(replacements)}
	plugin := validPlugin()
	plugin.sdk = sdk
	plugin.writers = []*writer{{stream: firstClosed}, {stream: secondClosed}}
	input := []telegraf.Metric{testutil.TestMetric(1), testutil.TestMetric(2)}

	require.NoError(t, plugin.Write(input))
	require.Equal(t, 2, sdk.staticSchemaCalls)
	require.Equal(t, 1, firstClosed.closeCalls)
	require.Equal(t, 1, secondClosed.closeCalls)
	for _, w := range plugin.writers {
		require.Contains(t, replacements, w.stream)
	}
}

func TestWriteFetchesTableSchemaOnceForConcurrentRecreations(t *testing.T) {
	sdk := &fakeSDK{
		streams:    []ingestStream{&fakeStream{}, &fakeStream{}},
		fetchDelay: 50 * time.Millisecond,
	}
	plugin := validPlugin()
	plugin.SchemaMode = schemaModeTableSchema
	plugin.sdk = sdk
	plugin.writers = []*writer{
		{stream: &fakeStream{closed: true}},
		{stream: &fakeStream{closed: true}},
	}
	input := []telegraf.Metric{testutil.TestMetric(1), testutil.TestMetric(2)}

	require.NoError(t, plugin.Write(input))
	require.Equal(t, 2, sdk.tableSchemaCalls)
	require.Equal(t, 1, sdk.fetchCalls)
}

func TestCloseClosesAllStreams(t *testing.T) {
	first := &fakeStream{}
	second := &fakeStream{}
	sdk := &fakeSDK{}
	plugin := validPlugin()
	plugin.sdk = sdk
	plugin.writers = []*writer{{stream: first}, {stream: second}}

	require.NoError(t, plugin.Close())
	require.Equal(t, 1, first.closeCalls)
	require.Equal(t, 1, second.closeCalls)
	require.Equal(t, 1, sdk.closeCalls)
	require.Nil(t, plugin.writers)
}

func TestPartitionRecords(t *testing.T) {
	records := [][]byte{[]byte("a"), []byte("b"), []byte("c"), []byte("d"), []byte("e")}
	tests := []struct {
		name     string
		count    int
		expected [][][]byte
	}{
		{
			name:     "single share",
			count:    1,
			expected: [][][]byte{records},
		},
		{
			name:     "remainder on the leading shares",
			count:    2,
			expected: [][][]byte{records[:3], records[3:]},
		},
		{
			name:  "even shares",
			count: 5,
			expected: [][][]byte{
				records[:1], records[1:2], records[2:3], records[3:4], records[4:],
			},
		},
		{
			name:  "fewer records than shares",
			count: 8,
			expected: [][][]byte{
				records[:1], records[1:2], records[2:3], records[3:4], records[4:],
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, partitionRecords(records, tt.count))
		})
	}
}

func TestWriteRecreatesTerminalStream(t *testing.T) {
	closed := &fakeStream{closed: true}
	replacement := &fakeStream{}
	sdk := &fakeSDK{stream: replacement}
	plugin := validPlugin()
	plugin.sdk = sdk
	setStream(plugin, closed)

	require.NoError(t, plugin.Write([]telegraf.Metric{testutil.TestMetric(1)}))
	require.Equal(t, 1, closed.unackedCalls)
	require.Equal(t, 1, closed.closeCalls)
	require.Equal(t, 1, sdk.staticSchemaCalls)
	require.Equal(t, 1, replacement.ingestCalls)
	require.Equal(t, 1, replacement.flushCalls)
	require.Same(t, replacement, currentStream(plugin))
}

func TestWriteRetriesFailedStreamRecreation(t *testing.T) {
	recreateErr := errors.New("stream recreation failed")
	closed := &fakeStream{closed: true}
	replacement := &fakeStream{}
	sdk := &fakeSDK{
		streams:      []ingestStream{replacement, replacement},
		createErrors: []error{recreateErr, nil},
	}
	plugin := validPlugin()
	plugin.sdk = sdk
	setStream(plugin, closed)
	input := []telegraf.Metric{testutil.TestMetric(1)}

	require.ErrorIs(t, plugin.Write(input), recreateErr)
	require.Nil(t, currentStream(plugin))
	require.NotNil(t, currentPending(plugin))

	require.NoError(t, plugin.Write(input))
	require.Same(t, replacement, currentStream(plugin))
	require.Equal(t, 2, sdk.staticSchemaCalls)
	require.Equal(t, 1, replacement.ingestCalls)
	require.Nil(t, currentPending(plugin))
}

func TestWriteReplaysOnlyUnacknowledgedRecordsAfterTerminalFailure(t *testing.T) {
	flushErr := errors.New("stream failed")
	closed := &fakeStream{flushErrors: []error{flushErr}}
	replacement := &fakeStream{}
	sdk := &fakeSDK{stream: replacement}
	plugin := validPlugin()
	plugin.sdk = sdk
	setStream(plugin, closed)
	input := []telegraf.Metric{testutil.TestMetric(1)}

	require.ErrorIs(t, plugin.Write(input), flushErr)
	require.Len(t, closed.batches, 1)
	closed.unacked = closed.batches
	closed.closed = true

	require.NoError(t, plugin.Write(input))
	require.Equal(t, 1, closed.ingestCalls)
	require.Equal(t, 1, replacement.ingestCalls)
	require.Equal(t, closed.batches[0], replacement.batches[0])
	require.Equal(t, []bool{true}, replacement.encoded)
	require.Nil(t, currentPending(plugin))
}

func TestWriteTableSchemaReencodesRecordsAfterTerminalFailure(t *testing.T) {
	flushErr := errors.New("stream failed")
	closed := &fakeStream{flushErrors: []error{flushErr}}
	replacement := &fakeStream{}
	sdk := &fakeSDK{stream: replacement}
	plugin := validPlugin()
	plugin.SchemaMode = schemaModeTableSchema
	plugin.sdk = sdk
	setStream(plugin, closed)
	input := []telegraf.Metric{testutil.TestMetric(1)}

	require.ErrorIs(t, plugin.Write(input), flushErr)
	require.Equal(t, []bool{false}, closed.encoded)
	closed.unacked = [][][]byte{{{0x08, 0x01}}}
	closed.closed = true

	require.NoError(t, plugin.Write(input))
	require.Equal(t, 1, sdk.tableSchemaCalls)
	require.Equal(t, closed.batches[0], replacement.batches[0])
	require.Equal(t, []bool{false}, replacement.encoded)
	require.Nil(t, currentPending(plugin))
}

func TestWriteTableSchemaReusesDescriptorForReplacementStream(t *testing.T) {
	flushErr := errors.New("stream failed")
	closed := &fakeStream{flushErrors: []error{flushErr}}
	replacement := &fakeStream{}
	sdk := &fakeSDK{stream: replacement, descriptors: [][]byte{[]byte("first")}}
	plugin := validPlugin()
	plugin.SchemaMode = schemaModeTableSchema
	plugin.sdk = sdk
	setStream(plugin, closed)
	plugin.descriptor = []byte("first")
	input := []telegraf.Metric{testutil.TestMetric(1)}

	require.ErrorIs(t, plugin.Write(input), flushErr)
	closed.unacked = [][][]byte{{{0x08, 0x01}}}
	closed.closed = true

	require.NoError(t, plugin.Write(input))
	require.Equal(t, 1, sdk.tableSchemaCalls)
	require.Same(t, replacement, currentStream(plugin))
	require.Zero(t, sdk.fetchCalls)
	require.Equal(t, []byte("first"), plugin.descriptor)
	require.False(t, plugin.descriptorReused)
}

func TestWriteTableSchemaRefetchesDescriptorAfterReusedStreamFails(t *testing.T) {
	firstErr := errors.New("first stream failed")
	secondErr := errors.New("reused descriptor failed")
	closed := &fakeStream{flushErrors: []error{firstErr}}
	reused := &fakeStream{flushErrors: []error{secondErr}}
	refetched := &fakeStream{}
	sdk := &fakeSDK{
		streams:     []ingestStream{reused, refetched},
		descriptors: [][]byte{[]byte("second")},
	}
	plugin := validPlugin()
	plugin.SchemaMode = schemaModeTableSchema
	plugin.sdk = sdk
	setStream(plugin, closed)
	plugin.descriptor = []byte("first")
	input := []telegraf.Metric{testutil.TestMetric(1)}

	require.ErrorIs(t, plugin.Write(input), firstErr)
	closed.unacked = [][][]byte{{{0x08, 0x01}}}
	closed.closed = true

	require.ErrorIs(t, plugin.Write(input), secondErr)
	require.Equal(t, 1, sdk.tableSchemaCalls)
	require.Same(t, reused, currentStream(plugin))
	require.Zero(t, sdk.fetchCalls)
	require.True(t, plugin.descriptorReused)
	reused.unacked = [][][]byte{{{0x08, 0x01}}}
	reused.closed = true

	require.NoError(t, plugin.Write(input))
	require.Equal(t, 2, sdk.tableSchemaCalls)
	require.Same(t, refetched, currentStream(plugin))
	require.Equal(t, 1, sdk.fetchCalls)
	require.Equal(t, []byte("second"), plugin.descriptor)
	require.False(t, plugin.descriptorReused)
}

func TestCloseClearsCachedDescriptor(t *testing.T) {
	plugin := validPlugin()
	plugin.SchemaMode = schemaModeTableSchema
	plugin.sdk = &fakeSDK{}
	setStream(plugin, &fakeStream{})
	plugin.descriptor = []byte("descriptor")
	plugin.descriptorReused = true

	require.NoError(t, plugin.Close())
	require.Nil(t, plugin.descriptor)
	require.False(t, plugin.descriptorReused)
}

func TestWriteTableSchemaReplaysOnlyCurrentSuffix(t *testing.T) {
	firstFlushErr := errors.New("first flush failed")
	suffixFlushErr := errors.New("suffix flush failed")
	closed := &fakeStream{flushErrors: []error{firstFlushErr, nil, suffixFlushErr}}
	replacement := &fakeStream{}
	sdk := &fakeSDK{stream: replacement}
	plugin := validPlugin()
	plugin.SchemaMode = schemaModeTableSchema
	plugin.sdk = sdk
	setStream(plugin, closed)
	original := testutil.TestMetric(1)
	added := testutil.TestMetric(2)

	require.ErrorIs(t, plugin.Write([]telegraf.Metric{original}), firstFlushErr)
	require.ErrorIs(t, plugin.Write([]telegraf.Metric{original, added}), suffixFlushErr)
	closed.unacked = [][][]byte{{{0x08, 0x01}}}
	closed.closed = true

	require.NoError(t, plugin.Write([]telegraf.Metric{original, added}))
	expectedAdded, err := metricToTableSchemaJSON(
		added,
		plugin.TimestampColumn,
		plugin.MeasurementColumn,
	)
	require.NoError(t, err)
	require.Equal(t, [][]byte{expectedAdded}, replacement.batches[0])
	require.Equal(t, []bool{false}, replacement.encoded)
}

func TestWriteTableSchemaReplaysOnlyUnacknowledgedChunk(t *testing.T) {
	flushErr := errors.New("flush failed")
	closed := &fakeStream{flushErrors: []error{flushErr}}
	replacement := &fakeStream{}
	sdk := &fakeSDK{stream: replacement}
	plugin := validPlugin()
	plugin.SchemaMode = schemaModeTableSchema
	plugin.batchRecordLimit = 1
	plugin.sdk = sdk
	setStream(plugin, closed)
	input := []telegraf.Metric{testutil.TestMetric(1), testutil.TestMetric(2)}

	require.ErrorIs(t, plugin.Write(input), flushErr)
	require.Len(t, closed.batches, 2)
	closed.unacked = [][][]byte{{{0x08, 0x01}}}
	closed.closed = true

	require.NoError(t, plugin.Write(input))
	require.Len(t, replacement.batches, 1)
	require.Equal(t, closed.batches[1], replacement.batches[0])
	require.Equal(t, []bool{false}, replacement.encoded)
}

func TestWriteRejectsIndividuallyOversizedMetricBeforeAdmission(t *testing.T) {
	stream := &fakeStream{}
	plugin := validPlugin()
	plugin.payloadByteLimit = batchEnvelopeReserve + 1
	setStream(plugin, stream)

	err := plugin.Write([]telegraf.Metric{testutil.TestMetric(1)})
	require.ErrorContains(t, err, "exceeding the payload budget")
	require.Zero(t, stream.ingestCalls)
}

func TestWriteRejectsInvalidMetricWithoutBlockingValidMetrics(t *testing.T) {
	stream := &fakeStream{}
	plugin := validPlugin()
	setStream(plugin, stream)
	input := []telegraf.Metric{
		testutil.TestMetric(1),
		metricWithFields{
			Metric: testutil.TestMetric(2),
			fields: []*telegraf.Field{{Key: "unsupported", Value: complex(1, 2)}},
		},
		testutil.TestMetric(3),
	}

	err := plugin.Write(input)
	var writeErr *internal.PartialWriteError
	require.ErrorAs(t, err, &writeErr)
	require.Equal(t, []int{0, 2}, writeErr.MetricsAccept)
	require.Equal(t, []int{1}, writeErr.MetricsReject)
	require.Len(t, writeErr.MetricsRejectErrors, 1)
	require.ErrorContains(t, writeErr.MetricsRejectErrors[0], "unsupported field type")
	require.Equal(t, 1, stream.ingestCalls)
	require.Len(t, stream.batches[0], 2)
	require.Equal(t, 1, stream.flushCalls)
}

func TestWriteRejectsMetricExceedingBufferedPayloadLimit(t *testing.T) {
	record, err := serializeStaticMetric(testutil.TestMetric(1))
	require.NoError(t, err)
	recordSize := protowire.SizeTag(1) + protowire.SizeBytes(len(record))

	stream := &fakeStream{}
	plugin := validPlugin()
	plugin.bufferedByteLimit = retainedPayloadSize(recordSize, 1) - 1
	setStream(plugin, stream)

	err = plugin.Write([]telegraf.Metric{testutil.TestMetric(1)})
	var writeErr *internal.PartialWriteError
	require.ErrorAs(t, err, &writeErr)
	require.Equal(t, []int{0}, writeErr.MetricsReject)
	require.ErrorContains(t, err, "exceeding the buffer limit")
	require.Zero(t, stream.ingestCalls)
}

func TestWriteSplitsBatchByPayloadSize(t *testing.T) {
	input := []telegraf.Metric{testutil.TestMetric(1), testutil.TestMetric(2)}
	first, err := serializeStaticMetric(input[0])
	require.NoError(t, err)

	stream := &fakeStream{}
	plugin := validPlugin()
	plugin.payloadByteLimit = batchEnvelopeReserve + protowire.SizeTag(1) + protowire.SizeBytes(len(first))
	setStream(plugin, stream)

	require.NoError(t, plugin.Write(input))
	require.Equal(t, 2, stream.ingestCalls)
	require.Len(t, stream.batches, 2)
	require.Len(t, stream.batches[0], 1)
	require.Len(t, stream.batches[1], 1)
	require.Equal(t, 1, stream.flushCalls)
}

func TestWriteSplitsBatchByBufferedPayloadSize(t *testing.T) {
	input := []telegraf.Metric{testutil.TestMetric(1), testutil.TestMetric(2)}
	first, err := serializeStaticMetric(input[0])
	require.NoError(t, err)
	second, err := serializeStaticMetric(input[1])
	require.NoError(t, err)
	firstSize := protowire.SizeTag(1) + protowire.SizeBytes(len(first))
	secondSize := protowire.SizeTag(1) + protowire.SizeBytes(len(second))

	stream := &fakeStream{}
	plugin := validPlugin()
	plugin.bufferedByteLimit = max(retainedPayloadSize(firstSize, 1), retainedPayloadSize(secondSize, 1))
	setStream(plugin, stream)

	require.NoError(t, plugin.Write(input))
	require.Equal(t, 2, stream.ingestCalls)
	require.Len(t, stream.batches[0], 1)
	require.Len(t, stream.batches[1], 1)
	require.Equal(t, 1, stream.flushCalls)
}

func TestCloseIsIdempotentAndJoinsErrors(t *testing.T) {
	streamErr := errors.New("stream close failed")
	sdkErr := errors.New("sdk close failed")
	stream := &fakeStream{closeErr: streamErr}
	sdk := &fakeSDK{closeErr: sdkErr}
	plugin := validPlugin()
	setStream(plugin, stream)
	plugin.sdk = sdk

	err := plugin.Close()
	require.ErrorIs(t, err, streamErr)
	require.ErrorIs(t, err, sdkErr)
	require.Equal(t, 1, stream.closeCalls)
	require.Equal(t, 1, sdk.closeCalls)

	require.NoError(t, plugin.Close())
	require.Equal(t, 1, stream.closeCalls)
	require.Equal(t, 1, sdk.closeCalls)
}

func validPlugin() *Zerobus {
	return &Zerobus{
		ServerEndpoint:    "https://workspace.zerobus.example.com",
		WorkspaceURL:      "https://workspace.example.com",
		TableName:         "catalog.schema.metrics",
		ClientID:          "client",
		ClientSecret:      config.NewSecret([]byte("secret")),
		ApplicationName:   "telegraf",
		SchemaMode:        schemaModeStatic,
		TimestampColumn:   "timestamp",
		ConnectTimeout:    config.Duration(defaultConnectTimeout),
		ConcurrentStreams: 1,
		Log:               testutil.Logger{},

		batchRecordLimit:  defaultMaxBatchRecords,
		payloadByteLimit:  defaultMaxPayloadBytes,
		bufferedByteLimit: defaultMaxBufferedPayloadBytes,
	}
}

func setStream(plugin *Zerobus, stream ingestStream) {
	plugin.writers = []*writer{{stream: stream}}
}

func currentStream(plugin *Zerobus) ingestStream {
	return plugin.writers[0].stream
}

func currentPending(plugin *Zerobus) *pendingWrite {
	return plugin.writers[0].pending
}

type fakeStream struct {
	records      [][]byte
	batches      [][][]byte
	unacked      [][][]byte
	ingestErr    error
	ingestErrors []error
	flushErr     error
	flushErrors  []error
	unackedErr   error
	closeErr     error
	closed       bool
	encoded      []bool
	ingestCalls  int
	flushCalls   int
	unackedCalls int
	closeCalls   int
}

func (s *fakeStream) IngestRecordsOffset(records [][]byte, encoded bool) (int64, error) {
	s.ingestCalls++
	s.records = records
	s.batches = append(s.batches, records)
	s.encoded = append(s.encoded, encoded)
	if len(s.ingestErrors) > 0 {
		err := s.ingestErrors[0]
		s.ingestErrors = s.ingestErrors[1:]
		return int64(s.ingestCalls), err
	}
	return int64(s.ingestCalls), s.ingestErr
}

func (s *fakeStream) Flush() error {
	s.flushCalls++
	if len(s.flushErrors) > 0 {
		err := s.flushErrors[0]
		s.flushErrors = s.flushErrors[1:]
		return err
	}
	return s.flushErr
}

func (s *fakeStream) GetUnackedBatches() ([][][]byte, error) {
	s.unackedCalls++
	return s.unacked, s.unackedErr
}

func (s *fakeStream) IsClosed() bool {
	return s.closed
}

func (s *fakeStream) Close() error {
	s.closeCalls++
	s.closed = true
	return s.closeErr
}

type fakeSDK struct {
	mu                sync.Mutex
	stream            ingestStream
	streams           []ingestStream
	createErr         error
	createErrors      []error
	closeErr          error
	tableName         string
	clientID          string
	clientSecret      string
	options           []sdkzerobus.StreamOption
	contexts          []context.Context
	staticSchemaCalls int
	tableSchemaCalls  int
	closeCalls        int
	descriptors       [][]byte
	descriptorErr     error
	fetchCalls        int
	fetchDelay        time.Duration
}

func (s *fakeSDK) CreateStaticSchemaStream(
	ctx context.Context,
	tableName, clientID, clientSecret string,
	options ...sdkzerobus.StreamOption,
) (ingestStream, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.staticSchemaCalls++
	s.tableName = tableName
	s.clientID = clientID
	s.clientSecret = clientSecret
	s.options = options
	s.contexts = append(s.contexts, ctx)
	err := s.createErr
	if len(s.createErrors) > 0 {
		err = s.createErrors[0]
		s.createErrors = s.createErrors[1:]
	}
	if len(s.streams) > 0 {
		stream := s.streams[0]
		s.streams = s.streams[1:]
		return stream, err
	}
	return s.stream, err
}

func (s *fakeSDK) CreateTableSchemaStream(
	ctx context.Context,
	tableName, clientID, clientSecret string,
	options ...sdkzerobus.StreamOption,
) (ingestStream, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tableSchemaCalls++
	s.tableName = tableName
	s.clientID = clientID
	s.clientSecret = clientSecret
	s.options = options
	s.contexts = append(s.contexts, ctx)
	err := s.createErr
	if len(s.createErrors) > 0 {
		err = s.createErrors[0]
		s.createErrors = s.createErrors[1:]
	}
	if len(s.streams) > 0 {
		stream := s.streams[0]
		s.streams = s.streams[1:]
		return stream, err
	}
	return s.stream, err
}

func (s *fakeSDK) FetchProtoDescriptor(
	_ context.Context,
	_, _, _ string,
) ([]byte, error) {
	s.mu.Lock()
	s.fetchCalls++
	delay, err := s.fetchDelay, s.descriptorErr
	descriptor := []byte("descriptor")
	if len(s.descriptors) > 0 {
		descriptor = s.descriptors[0]
		if len(s.descriptors) > 1 {
			s.descriptors = s.descriptors[1:]
		}
	}
	s.mu.Unlock()
	time.Sleep(delay)
	if err != nil {
		return nil, err
	}
	return descriptor, nil
}

func (s *fakeSDK) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closeCalls++
	return s.closeErr
}

type metricWithFields struct {
	telegraf.Metric
	fields []*telegraf.Field
}

func (m metricWithFields) FieldList() []*telegraf.Field {
	return m.fields
}
