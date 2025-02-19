package timestamp

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)

func TestCases(t *testing.T) {
	type testcase struct {
		name      string
		timestamp Timestamp
		input     telegraf.Metric
		expected  telegraf.Metric
	}

	testcases := []testcase{
		{
			name: "field does not exist",
			timestamp: Timestamp{
				Field:             "timestamp",
				SourceFormat:      "2006-01-02T15:04:05Z",
				DestinationFormat: "unix",
			},
			input:    metric.New("test", map[string]string{}, map[string]any{}, time.Unix(0, 0)),
			expected: metric.New("test", map[string]string{}, map[string]any{}, time.Unix(0, 0)),
		},
		{
			name: "field to unix",
			timestamp: Timestamp{
				Field:             "timestamp",
				SourceFormat:      "2006-01-02T15:04:05Z",
				DestinationFormat: "unix",
			},
			input: metric.New(
				"test",
				map[string]string{},
				map[string]any{"timestamp": "2024-03-04T10:10:32.123456789Z"},
				time.Unix(0, 0),
			),
			expected: metric.New(
				"test",
				map[string]string{},
				map[string]any{"timestamp": int64(1709547032)},
				time.Unix(0, 0),
			),
		},
		{
			name: "field to unix_ms",
			timestamp: Timestamp{
				Field:             "timestamp",
				SourceFormat:      "2006-01-02T15:04:05Z",
				DestinationFormat: "unix_ms",
			},
			input: metric.New(
				"test",
				map[string]string{},
				map[string]any{"timestamp": "2024-03-04T10:10:32.123456789Z"},
				time.Unix(0, 0),
			),
			expected: metric.New(
				"test",
				map[string]string{},
				map[string]any{"timestamp": int64(1709547032123)},
				time.Unix(0, 0),
			),
		},
		{
			name: "field to unix_us",
			timestamp: Timestamp{
				Field:             "timestamp",
				SourceFormat:      "2006-01-02T15:04:05Z",
				DestinationFormat: "unix_us",
			},
			input: metric.New(
				"test",
				map[string]string{},
				map[string]any{"timestamp": "2024-03-04T10:10:32.123456789Z"},
				time.Unix(0, 0),
			),
			expected: metric.New(
				"test",
				map[string]string{},
				map[string]any{"timestamp": int64(1709547032123456)},
				time.Unix(0, 0),
			),
		},
		{
			name: "field to unix_ns",
			timestamp: Timestamp{
				Field:             "timestamp",
				SourceFormat:      "2006-01-02T15:04:05Z",
				DestinationFormat: "unix_ns",
			},
			input: metric.New(
				"test",
				map[string]string{},
				map[string]any{"timestamp": "2024-03-04T10:10:32.123456789Z"},
				time.Unix(0, 0),
			),
			expected: metric.New(
				"test",
				map[string]string{},
				map[string]any{"timestamp": int64(1709547032123456789)},
				time.Unix(0, 0),
			),
		},
		{
			name: "field to custom format",
			timestamp: Timestamp{
				Field:             "timestamp",
				SourceFormat:      "2006-01-02T15:04:05Z",
				DestinationFormat: "2006-01-02T15:04:05",
			},
			input: metric.New(
				"test",
				map[string]string{},
				map[string]any{"timestamp": "2024-03-04T10:10:32.123456789Z"},
				time.Unix(0, 0),
			),
			expected: metric.New(
				"test",
				map[string]string{},
				map[string]any{"timestamp": "2024-03-04T10:10:32"},
				time.Unix(0, 0),
			),
		},
		{
			name: "unix_ns to unix",
			timestamp: Timestamp{
				Field:             "timestamp",
				SourceFormat:      "unix_ns",
				DestinationFormat: "unix",
			},
			input: metric.New(
				"test",
				map[string]string{},
				map[string]any{"timestamp": int64(1709547032123456789)},
				time.Unix(0, 0),
			),
			expected: metric.New(
				"test",
				map[string]string{},
				map[string]any{"timestamp": int64(1709547032)},
				time.Unix(0, 0),
			),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			processor := tc.timestamp
			require.NoError(t, processor.Init())

			output := processor.Apply(tc.input)
			require.Len(t, output, 1)
			testutil.RequireMetricsEqual(t, []telegraf.Metric{tc.expected}, output)
		})
	}
}
