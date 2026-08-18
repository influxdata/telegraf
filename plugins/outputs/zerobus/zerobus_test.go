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

func TestInit(t *testing.T) {
	tests := []struct {
		name              string
		timestampColumn   string
		measurementColumn string
	}{
		{
			name:              "allows an omitted timestamp column",
			measurementColumn: "measurement",
		},
		{
			name: "allows omitting both columns",
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
			require.NoError(t, plugin.Init())
		})
	}
}

func TestInitFail(t *testing.T) {
	tests := []struct {
		name              string
		endpoint          string
		workspace         string
		table             string
		clientID          string
		secret            string
		timestampColumn   string
		measurementColumn string
		timeout           config.Duration
		expected          string
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
		{
			name:              "colliding columns",
			endpoint:          "https://workspace.zerobus.example.com",
			workspace:         "https://workspace.example.com",
			table:             "catalog.schema.metrics",
			clientID:          "client",
			secret:            "secret",
			timestampColumn:   "timestamp",
			measurementColumn: "timestamp",
			expected:          `options "measurement_column" and "timestamp_column" must be different`,
		},
		{
			name:            "negative timeout",
			endpoint:        "https://workspace.zerobus.example.com",
			workspace:       "https://workspace.example.com",
			table:           "catalog.schema.metrics",
			clientID:        "client",
			secret:          "secret",
			timestampColumn: "timestamp",
			timeout:         config.Duration(-1),
			expected:        `option "timeout" cannot be negative`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := &Zerobus{
				Endpoint:          tt.endpoint,
				Workspace:         tt.workspace,
				Table:             tt.table,
				ClientID:          tt.clientID,
				ClientSecret:      config.NewSecret([]byte(tt.secret)),
				TimestampColumn:   tt.timestampColumn,
				MeasurementColumn: tt.measurementColumn,
				Timeout:           tt.timeout,
				Log:               testutil.Logger{},
			}
			require.ErrorContains(t, plugin.Init(), tt.expected)
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
		time.Unix(0, 1700000000123456000),
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

func TestSerializeMetricsRejectsMetrics(t *testing.T) {
	tests := []struct {
		name     string
		metrics  []telegraf.Metric
		accepted []int
		rejected []int
		expected string
	}{
		{
			name: "tag and field of the same name",
			metrics: []telegraf.Metric{
				testutil.TestMetric(1),
				metric.New(
					"cpu",
					map[string]string{"host": "tag"},
					map[string]interface{}{"host": "field"},
					time.Now(),
				),
				testutil.TestMetric(3),
			},
			accepted: []int{0, 2},
			rejected: []int{1},
			expected: `field "host" conflicts`,
		},
		{
			name: "metric exceeding the request limit",
			metrics: []telegraf.Metric{
				metric.New(
					"cpu",
					nil,
					map[string]interface{}{"value": strings.Repeat("x", maxRequestBytes)},
					time.Now(),
				),
			},
			rejected: []int{0},
			expected: "exceeding the request limit",
		},
	}

	plugin := &Zerobus{
		TimestampColumn: "timestamp",
		Log:             testutil.Logger{},
		columns:         map[string]bool{"timestamp": true, "tag1": true, "value": true, "host": true},
		maxRecords:      maxBatchRecords,
		maxBytes:        maxRequestBytes,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			batches, err := plugin.serializeMetrics(tt.metrics)

			var batched []int
			for _, b := range batches {
				batched = append(batched, b.indices...)
			}
			require.Equal(t, tt.accepted, batched)

			var writeErr *internal.PartialWriteError
			require.ErrorAs(t, err, &writeErr)
			require.Equal(t, tt.accepted, writeErr.MetricsAccept)
			require.Equal(t, tt.rejected, writeErr.MetricsReject)
			require.ErrorContains(t, err, tt.expected)
		})
	}
}

func TestSerializeMetricsSplitsRequests(t *testing.T) {
	plugin := &Zerobus{
		TimestampColumn: "timestamp",
		Log:             testutil.Logger{},
		columns:         map[string]bool{"timestamp": true, "tag1": true, "value": true},
		maxRecords:      maxBatchRecords,
		maxBytes:        maxRequestBytes,
	}

	// Determine the size of a request holding the first two metrics
	metrics := []telegraf.Metric{testutil.TestMetric(1), testutil.TestMetric(2), testutil.TestMetric(3)}
	batches, err := plugin.serializeMetrics(metrics)
	require.NoError(t, err)
	require.Len(t, batches, 1)
	twoRecords := recordSize(batches[0].records[0]) + recordSize(batches[0].records[1])

	tests := []struct {
		name       string
		metrics    []telegraf.Metric
		maxRecords int
		maxBytes   int
		expected   [][]int
	}{
		{
			name:       "keeps a fitting batch together",
			metrics:    metrics,
			maxRecords: maxBatchRecords,
			maxBytes:   maxRequestBytes,
			expected:   [][]int{{0, 1, 2}},
		},
		{
			name:       "splits by record count",
			metrics:    metrics,
			maxRecords: 2,
			maxBytes:   maxRequestBytes,
			expected:   [][]int{{0, 1}, {2}},
		},
		{
			name:       "splits by payload size",
			metrics:    metrics,
			maxRecords: maxBatchRecords,
			maxBytes:   twoRecords,
			expected:   [][]int{{0, 1}, {2}},
		},
		{
			name:       "handles an empty batch",
			maxRecords: maxBatchRecords,
			maxBytes:   maxRequestBytes,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin.maxRecords, plugin.maxBytes = tt.maxRecords, tt.maxBytes

			batches, err := plugin.serializeMetrics(tt.metrics)
			require.NoError(t, err)
			require.Len(t, batches, len(tt.expected))
			for i, indices := range tt.expected {
				require.Equal(t, indices, batches[i].indices)
			}
		})
	}
}
