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
	"github.com/influxdata/telegraf/testutil"
)

func TestInitRequiredOptions(t *testing.T) {
	tests := []struct {
		name      string
		endpoint  string
		workspace string
		table     string
		clientID  string
		secret    string
		expected  string
	}{
		{
			name:      "missing endpoint",
			workspace: "https://workspace.example.com",
			table:     "catalog.schema.metrics",
			clientID:  "client",
			secret:    "secret",
			expected:  `option "endpoint" must be set`,
		},
		{
			name:     "missing workspace",
			endpoint: "https://workspace.zerobus.example.com",
			table:    "catalog.schema.metrics",
			clientID: "client",
			secret:   "secret",
			expected: `option "workspace" must be set`,
		},
		{
			name:      "missing table",
			endpoint:  "https://workspace.zerobus.example.com",
			workspace: "https://workspace.example.com",
			clientID:  "client",
			secret:    "secret",
			expected:  `option "table" must be set`,
		},
		{
			name:      "missing client ID",
			endpoint:  "https://workspace.zerobus.example.com",
			workspace: "https://workspace.example.com",
			table:     "catalog.schema.metrics",
			secret:    "secret",
			expected:  `option "client_id" must be set`,
		},
		{
			name:      "missing client secret",
			endpoint:  "https://workspace.zerobus.example.com",
			workspace: "https://workspace.example.com",
			table:     "catalog.schema.metrics",
			clientID:  "client",
			expected:  `option "client_secret" must be set`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := &Zerobus{
				Endpoint:     tt.endpoint,
				Workspace:    tt.workspace,
				Table:        tt.table,
				ClientID:     tt.clientID,
				ClientSecret: config.NewSecret([]byte(tt.secret)),
				Log:          testutil.Logger{},
			}
			require.ErrorContains(t, plugin.Init(), tt.expected)
		})
	}
}

func TestInitTimeout(t *testing.T) {
	tests := []struct {
		name     string
		timeout  config.Duration
		expected string
	}{
		{
			name: "keeps a zero timeout",
		},
		{
			name:    "keeps the configured timeout",
			timeout: config.Duration(30 * time.Second),
		},
		{
			name:     "rejects a negative timeout",
			timeout:  config.Duration(-1),
			expected: `option "timeout" cannot be negative`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := &Zerobus{
				Endpoint:        "https://workspace.zerobus.example.com",
				Workspace:       "https://workspace.example.com",
				Table:           "catalog.schema.metrics",
				ClientID:        "client",
				ClientSecret:    config.NewSecret([]byte("secret")),
				TimestampColumn: "timestamp",
				Timeout:         tt.timeout,
				Log:             testutil.Logger{},
			}

			if tt.expected != "" {
				require.ErrorContains(t, plugin.Init(), tt.expected)
				return
			}
			require.NoError(t, plugin.Init())
			require.Equal(t, tt.timeout, plugin.Timeout)
		})
	}
}

func TestInitColumns(t *testing.T) {
	tests := []struct {
		name              string
		timestampColumn   string
		measurementColumn string
		expected          string
	}{
		{
			name:              "allows an omitted timestamp column",
			measurementColumn: "measurement",
		},
		{
			name: "allows omitting both columns",
		},
		{
			name:              "rejects colliding columns",
			timestampColumn:   "timestamp",
			measurementColumn: "timestamp",
			expected:          `options "measurement_column" and "timestamp_column" must be different`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := &Zerobus{
				Endpoint:          "https://workspace.zerobus.example.com",
				Workspace:         "https://workspace.example.com",
				Table:             "catalog.schema.metrics",
				ClientID:          "client",
				ClientSecret:      config.NewSecret([]byte("secret")),
				TimestampColumn:   tt.timestampColumn,
				MeasurementColumn: tt.measurementColumn,
				Timeout:           config.Duration(30 * time.Second),
				Log:               testutil.Logger{},
			}

			if tt.expected != "" {
				require.ErrorContains(t, plugin.Init(), tt.expected)
				return
			}
			require.NoError(t, plugin.Init())
		})
	}
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
	plugin := &Zerobus{TimestampColumn: "timestamp", Log: testutil.Logger{}}

	batches, err := plugin.serializeMetrics([]telegraf.Metric{
		testutil.TestMetric(1),
		metricWithFields{
			Metric: testutil.TestMetric(2),
			fields: []*telegraf.Field{{Key: "unsupported", Value: complex(1, 2)}},
		},
		testutil.TestMetric(3),
	}, maxBatchRecords, maxRequestBytes)
	require.Len(t, batches, 1)
	require.Len(t, batches[0].records, 2)
	require.Equal(t, []int{0, 2}, batches[0].indices)

	var writeErr *internal.PartialWriteError
	require.ErrorAs(t, err, &writeErr)
	require.Equal(t, []int{0, 2}, writeErr.MetricsAccept)
	require.Equal(t, []int{1}, writeErr.MetricsReject)
	require.Len(t, writeErr.MetricsRejectErrors, 1)
	require.ErrorContains(t, writeErr.MetricsRejectErrors[0], "unsupported type complex128")
}

func TestSerializeMetricsRejectsOversizedMetric(t *testing.T) {
	plugin := &Zerobus{TimestampColumn: "timestamp", Log: testutil.Logger{}}

	oversized := metric.New(
		"cpu",
		nil,
		map[string]interface{}{"value": strings.Repeat("x", maxRequestBytes)},
		time.Now(),
	)
	batches, err := plugin.serializeMetrics([]telegraf.Metric{oversized}, maxBatchRecords, maxRequestBytes)
	require.Empty(t, batches)

	var writeErr *internal.PartialWriteError
	require.ErrorAs(t, err, &writeErr)
	require.Equal(t, []int{0}, writeErr.MetricsReject)
	require.ErrorContains(t, err, "exceeding the request limit")
}

func TestSerializeMetricsSplitsRequests(t *testing.T) {
	plugin := &Zerobus{TimestampColumn: "timestamp", Log: testutil.Logger{}}

	metrics := []telegraf.Metric{testutil.TestMetric(1), testutil.TestMetric(2), testutil.TestMetric(3)}

	t.Run("keeps a fitting batch together", func(t *testing.T) {
		batches, err := plugin.serializeMetrics(metrics, maxBatchRecords, maxRequestBytes)
		require.NoError(t, err)
		require.Len(t, batches, 1)
		require.Equal(t, []int{0, 1, 2}, batches[0].indices)
	})

	t.Run("splits by record count", func(t *testing.T) {
		batches, err := plugin.serializeMetrics(metrics, 2, maxRequestBytes)
		require.NoError(t, err)
		require.Len(t, batches, 2)
		require.Equal(t, []int{0, 1}, batches[0].indices)
		require.Equal(t, []int{2}, batches[1].indices)
	})

	t.Run("splits by payload size", func(t *testing.T) {
		batches, err := plugin.serializeMetrics(metrics, maxBatchRecords, maxRequestBytes)
		require.NoError(t, err)
		require.Len(t, batches, 1)
		limit := recordSize(batches[0].records[0]) + recordSize(batches[0].records[1])

		batches, err = plugin.serializeMetrics(metrics, maxBatchRecords, limit)
		require.NoError(t, err)
		require.Len(t, batches, 2)
		require.Equal(t, []int{0, 1}, batches[0].indices)
		require.Equal(t, []int{2}, batches[1].indices)
	})

	t.Run("handles an empty batch", func(t *testing.T) {
		batches, err := plugin.serializeMetrics(nil, maxBatchRecords, maxRequestBytes)
		require.NoError(t, err)
		require.Empty(t, batches)
	})
}

type metricWithFields struct {
	telegraf.Metric
	fields []*telegraf.Field
}

func (m metricWithFields) FieldList() []*telegraf.Field {
	return m.fields
}
