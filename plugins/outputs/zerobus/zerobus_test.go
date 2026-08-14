package zerobus

import (
	"encoding/json"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

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
	require.Equal(t, "timestamp", plugin.TimestampColumn)
	require.Empty(t, plugin.MeasurementColumn)
	require.Empty(t, plugin.Application)
	require.Equal(t, config.Duration(defaultConnectTimeout), plugin.Timeout)
	require.NotEmpty(t, plugin.SampleConfig())
}

func TestInitRequiredOptions(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Zerobus)
		option string
	}{
		{
			name:   "endpoint",
			mutate: func(z *Zerobus) { z.Endpoint = "" },
			option: "endpoint",
		},
		{
			name:   "workspace",
			mutate: func(z *Zerobus) { z.Workspace = "" },
			option: "workspace",
		},
		{
			name:   "table",
			mutate: func(z *Zerobus) { z.Table = "" },
			option: "table",
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

func TestInitColumnOptions(t *testing.T) {
	t.Run("defaults the timeout", func(t *testing.T) {
		plugin := validPlugin()
		plugin.Timeout = 0
		require.NoError(t, plugin.Init())
		require.Equal(t, config.Duration(defaultConnectTimeout), plugin.Timeout)
	})

	t.Run("rejects a negative timeout", func(t *testing.T) {
		plugin := validPlugin()
		plugin.Timeout = -1
		require.ErrorContains(t, plugin.Init(), "timeout")
	})

	t.Run("allows an omitted timestamp column", func(t *testing.T) {
		plugin := validPlugin()
		plugin.TimestampColumn = ""
		require.NoError(t, plugin.Init())
	})

	t.Run("rejects colliding columns", func(t *testing.T) {
		plugin := validPlugin()
		plugin.MeasurementColumn = plugin.TimestampColumn
		require.ErrorContains(t, plugin.Init(), "must be different")
	})
}

func TestWriteWithoutConnection(t *testing.T) {
	plugin := validPlugin()
	require.NoError(t, plugin.Init())

	require.NoError(t, plugin.Write(nil))
	require.ErrorIs(t, plugin.Write([]telegraf.Metric{testutil.TestMetric(1)}), internal.ErrNotConnected)
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

	record, err := metricToTableSchemaJSON(input, "event_time", "measurement", nil)
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

	record, err = metricToTableSchemaJSON(input, "", "", nil)
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
			name: "unsupported field type",
			metric: metricWithFields{
				Metric: testutil.TestMetric(1),
				fields: []*telegraf.Field{{Key: "value", Value: []int{1}}},
			},
			match: "unsupported type []int",
		},
		{
			name: "nil field",
			metric: metricWithFields{
				Metric: testutil.TestMetric(1),
				fields: []*telegraf.Field{nil},
			},
			match: "nil field",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := metricToTableSchemaJSON(tt.metric, "timestamp", "", nil)
			require.ErrorContains(t, err, tt.match)
		})
	}
}

func TestSerializeMetricsRejectsInvalidMetric(t *testing.T) {
	plugin := validPlugin()
	require.NoError(t, plugin.Init())

	records, err := plugin.serializeMetrics([]telegraf.Metric{
		testutil.TestMetric(1),
		metricWithFields{
			Metric: testutil.TestMetric(2),
			fields: []*telegraf.Field{{Key: "unsupported", Value: complex(1, 2)}},
		},
		testutil.TestMetric(3),
	})
	require.Len(t, records, 2)

	var writeErr *internal.PartialWriteError
	require.ErrorAs(t, err, &writeErr)
	require.Equal(t, []int{0, 2}, writeErr.MetricsAccept)
	require.Equal(t, []int{1}, writeErr.MetricsReject)
	require.Len(t, writeErr.MetricsRejectErrors, 1)
	require.ErrorContains(t, writeErr.MetricsRejectErrors[0], "unsupported type complex128")
}

func TestSerializeMetricsRejectsOversizedMetric(t *testing.T) {
	plugin := validPlugin()
	require.NoError(t, plugin.Init())

	oversized := metric.New(
		"cpu",
		nil,
		map[string]interface{}{"value": strings.Repeat("x", maxRecordBytes)},
		time.Now(),
	)
	records, err := plugin.serializeMetrics([]telegraf.Metric{oversized})
	require.Empty(t, records)

	var writeErr *internal.PartialWriteError
	require.ErrorAs(t, err, &writeErr)
	require.Equal(t, []int{0}, writeErr.MetricsReject)
	require.ErrorContains(t, err, "exceeding the request limit")
}

func TestChunkRecords(t *testing.T) {
	records := [][]byte{[]byte("first"), []byte("second"), []byte("third")}

	t.Run("keeps a fitting batch together", func(t *testing.T) {
		chunks, err := chunkRecords(records, maxBatchRecords, maxRecordBytes)
		require.NoError(t, err)
		require.Len(t, chunks, 1)
		require.Equal(t, records, chunks[0])
	})

	t.Run("splits by record count", func(t *testing.T) {
		chunks, err := chunkRecords(records, 2, maxRecordBytes)
		require.NoError(t, err)
		require.Len(t, chunks, 2)
		require.Equal(t, records[:2], chunks[0])
		require.Equal(t, records[2:], chunks[1])
	})

	t.Run("splits by payload size", func(t *testing.T) {
		chunks, err := chunkRecords(records, maxBatchRecords, recordSize(records[0])+recordSize(records[1]))
		require.NoError(t, err)
		require.Len(t, chunks, 2)
		require.Equal(t, records[:2], chunks[0])
		require.Equal(t, records[2:], chunks[1])
	})

	t.Run("handles an empty batch", func(t *testing.T) {
		chunks, err := chunkRecords(nil, maxBatchRecords, maxRecordBytes)
		require.NoError(t, err)
		require.Empty(t, chunks)
	})
}

func validPlugin() *Zerobus {
	return &Zerobus{
		Endpoint:        "https://workspace.zerobus.example.com",
		Workspace:       "https://workspace.example.com",
		Table:           "catalog.schema.metrics",
		ClientID:        "client",
		ClientSecret:    config.NewSecret([]byte("secret")),
		Application:     "telegraf",
		TimestampColumn: "timestamp",
		Timeout:         config.Duration(defaultConnectTimeout),
		Log:             testutil.Logger{},
	}
}

type metricWithFields struct {
	telegraf.Metric
	fields []*telegraf.Field
}

func (m metricWithFields) FieldList() []*telegraf.Field {
	return m.fields
}
