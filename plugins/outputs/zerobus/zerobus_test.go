package zerobus

import (
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
	columns := map[string]bool{
		"event_time":  true,
		"measurement": true,
		"host":        true,
		"active":      true,
		"count":       true,
		"ratio":       true,
		"status":      true,
		"total":       true,
	}

	tests := []struct {
		name              string
		timestampColumn   string
		measurementColumn string
		expected          string
	}{
		{
			name:              "with timestamp and measurement column",
			timestampColumn:   "event_time",
			measurementColumn: "measurement",
			expected: `{
				"measurement": "cpu",
				"event_time": 1700000000123456,
				"host": "server-01",
				"active": true,
				"count": -42,
				"ratio": 1.25,
				"status": "ready",
				"total": 9223372036854775807
			}`,
		},
		{
			name: "without timestamp and measurement column",
			expected: `{
				"host": "server-01",
				"active": true,
				"count": -42,
				"ratio": 1.25,
				"status": "ready",
				"total": 9223372036854775807
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			record, err := metricToTableSchemaJSON(input, tt.timestampColumn, tt.measurementColumn, columns)
			require.NoError(t, err)
			require.JSONEq(t, tt.expected, string(record))
		})
	}
}

func TestMetricToTableSchemaJSONRejectsInvalidMetric(t *testing.T) {
	tests := []struct {
		name     string
		metric   telegraf.Metric
		expected string
	}{
		{
			name: "timestamp collision",
			metric: metric.New(
				"cpu",
				map[string]string{"timestamp": "tag"},
				map[string]interface{}{"value": 1.0},
				time.Now(),
			),
			expected: `tag "timestamp" conflicts`,
		},
		{
			name: "tag and field collision",
			metric: metric.New(
				"cpu",
				map[string]string{"host": "tag"},
				map[string]interface{}{"host": "field"},
				time.Now(),
			),
			expected: `field "host" conflicts`,
		},
		{
			name: "non-finite float",
			metric: metric.New(
				"cpu",
				nil,
				map[string]interface{}{"value": math.NaN()},
				time.Now(),
			),
			expected: "non-finite float",
		},
		{
			name: "uint64 above BIGINT maximum",
			metric: metric.New(
				"cpu",
				nil,
				map[string]interface{}{"value": uint64(math.MaxInt64) + 1},
				time.Now(),
			),
			expected: "exceeding Delta BIGINT maximum",
		},
	}

	columns := map[string]bool{"timestamp": true, "host": true, "value": true}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := metricToTableSchemaJSON(tt.metric, "timestamp", "", columns)
			require.ErrorContains(t, err, tt.expected)
		})
	}
}

func TestSerializeMetricsRejectsInvalidMetric(t *testing.T) {
	plugin := &Zerobus{
		TimestampColumn: "timestamp",
		Log:             testutil.Logger{},
		columns:         map[string]bool{"timestamp": true, "tag1": true, "value": true, "host": true},
	}

	batches, err := plugin.serializeMetrics([]telegraf.Metric{
		testutil.TestMetric(1),
		metric.New(
			"cpu",
			map[string]string{"host": "tag"},
			map[string]interface{}{"host": "field"},
			time.Now(),
		),
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
	require.ErrorContains(t, writeErr.MetricsRejectErrors[0], `field "host" conflicts`)
}

func TestSerializeMetricsRejectsOversizedMetric(t *testing.T) {
	plugin := &Zerobus{
		TimestampColumn: "timestamp",
		Log:             testutil.Logger{},
		columns:         map[string]bool{"timestamp": true, "value": true},
	}

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
	plugin := &Zerobus{
		TimestampColumn: "timestamp",
		Log:             testutil.Logger{},
		columns:         map[string]bool{"timestamp": true, "tag1": true, "value": true},
	}

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
