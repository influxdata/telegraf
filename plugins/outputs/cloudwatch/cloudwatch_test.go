package cloudwatch

import (
	"math"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)

// Test that each tag becomes one dimension
func TestBuildDimensions(t *testing.T) {
	tests := []struct {
		name     string
		expected []types.Dimension
	}{
		{
			name: "10 max dimensions",
			expected: []types.Dimension{
				{Name: aws.String("host"), Value: aws.String("localhost")},
				{Name: aws.String("a"), Value: aws.String("1")},
				{Name: aws.String("b"), Value: aws.String("2")},
				{Name: aws.String("c"), Value: aws.String("3")},
				{Name: aws.String("d"), Value: aws.String("4")},
				{Name: aws.String("e"), Value: aws.String("5")},
				{Name: aws.String("f"), Value: aws.String("6")},
				{Name: aws.String("g"), Value: aws.String("7")},
				{Name: aws.String("h"), Value: aws.String("8")},
				{Name: aws.String("i"), Value: aws.String("9")},
			},
		},
	}

	// Define the input tags and the expected output
	input := []*telegraf.Tag{
		{Key: "a", Value: "1"},
		{Key: "b", Value: "2"},
		{Key: "c", Value: "3"},
		{Key: "d", Value: "4"},
		{Key: "e", Value: "5"},
		{Key: "f", Value: "6"},
		{Key: "g", Value: "7"},
		{Key: "h", Value: "8"},
		{Key: "i", Value: "9"},
		{Key: "j", Value: "10"},
		{Key: "k", Value: "11"},
		{Key: "host", Value: "localhost"},
		{Key: "l", Value: "12"},
		{Key: "m", Value: "13"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build the dimensions and check
			dimensions := buildDimensions(input)
			require.Len(t, dimensions, len(tt.expected))
			for i, actual := range dimensions {
				require.Equalf(t, *tt.expected[i].Name, *actual.Name, "mismatch for element %d", i)
				require.Equalf(t, *tt.expected[i].Value, *actual.Value, "mismatch for element %d", i)
			}
		})
	}
}

// Test that metrics with valid values have a MetricDatum created where as non valid do not.
// Skips "time.Time" type as something is converting the value to string.
func TestBuildMetricDatums(t *testing.T) {
	tests := []struct {
		name       string
		statistics bool
		highres    bool
		input      telegraf.Metric
		expected   int
	}{
		{
			name:     "valid int",
			input:    testutil.TestMetric(1),
			expected: 1,
		},
		{
			name:     "valid int32",
			input:    testutil.TestMetric(int32(1)),
			expected: 1,
		},
		{
			name:     "valid int64",
			input:    testutil.TestMetric(int64(1)),
			expected: 1,
		},
		{
			name:     "valid float64 zero",
			input:    testutil.TestMetric(float64(0)),
			expected: 1,
		},
		{
			name:     "valid float64 negative zero",
			input:    testutil.TestMetric(math.Copysign(0, -1)),
			expected: 1,
		},
		{
			name:     "valid float64 one",
			input:    testutil.TestMetric(float64(1)),
			expected: 1,
		},
		{
			name:     "valid float64 tiny",
			input:    testutil.TestMetric(float64(8.515920e-109)),
			expected: 1,
		},
		{
			name:     "valid float64 huge",
			input:    testutil.TestMetric(float64(1.174271e+108)),
			expected: 1,
		},
		{
			name:     "valid bool",
			input:    testutil.TestMetric(true),
			expected: 1,
		},
		{
			name:     "invalid string",
			input:    testutil.TestMetric("Foo"),
			expected: 0,
		},
		{
			name:     "invalid NaN",
			input:    testutil.TestMetric(math.NaN()),
			expected: 0,
		},
		{
			name:     "invalid too small",
			input:    testutil.TestMetric(float64(8.515919e-109)),
			expected: 0,
		},
		{
			name:     "invalid too large",
			input:    testutil.TestMetric(float64(1.174272e+108)),
			expected: 0,
		},
		{
			name:       "statistics",
			statistics: true,
			input: metric.New(
				"test1",
				map[string]string{"tag1": "value1"},
				map[string]interface{}{
					"value_max":   float64(10),
					"value_min":   float64(0),
					"value_sum":   float64(100),
					"value_count": float64(20)},
				time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
			),
			expected: 1,
		},
		{
			name:       "multiple",
			statistics: true,
			input: metric.New(
				"test1",
				map[string]string{"tag1": "value1"},
				map[string]interface{}{
					"valueA": float64(10),
					"valueB": float64(0),
					"valueC": float64(100),
					"valueD": float64(20),
				},
				time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
			),
			expected: 4,
		},
		{
			name:       "multiple statistics",
			statistics: true,
			input: metric.New(
				"test1",
				map[string]string{"tag1": "value1"},
				map[string]interface{}{
					"valueA_max":   float64(10),
					"valueA_min":   float64(0),
					"valueA_sum":   float64(100),
					"valueA_count": float64(20),
					"valueB_max":   float64(10),
					"valueB_min":   float64(0),
					"valueB_sum":   float64(100),
					"valueB_count": float64(20),
					"valueC_max":   float64(10),
					"valueC_min":   float64(0),
					"valueC_sum":   float64(100),
					"valueD":       float64(10),
					"valueE":       float64(0),
				},
				time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
			),
			expected: 7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := &CloudWatch{
				Namespace:             "foo",
				WriteStatistics:       tt.statistics,
				HighResolutionMetrics: tt.highres,
				Log:                   testutil.Logger{},
			}
			require.NoError(t, plugin.Init())
			datums := plugin.buildMetricDatum(tt.input)
			require.Len(t, datums, tt.expected)
		})
	}
}

func TestBuildMetricDatumResolution(t *testing.T) {
	tests := []struct {
		name     string
		highres  bool
		expected int32
	}{
		{
			name:     "standard",
			expected: 60,
		},
		{
			name:     "high",
			highres:  true,
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup plugin
			plugin := &CloudWatch{
				Namespace:             "foo",
				HighResolutionMetrics: tt.highres,
				Log:                   testutil.Logger{},
			}
			require.NoError(t, plugin.Init())

			// Build a metric datum and check
			datum := plugin.buildMetricDatum(testutil.TestMetric(1))
			require.Len(t, datum, 1)
			require.NotNil(t, datum[0].StorageResolution)
			require.Equal(t, tt.expected, *datum[0].StorageResolution)
		})
	}
}

func TestBuildMetricDatumsSkipEmptyTags(t *testing.T) {
	// Setup plugin
	plugin := &CloudWatch{
		Namespace:       "foo",
		WriteStatistics: true,
		Log:             testutil.Logger{},
	}
	require.NoError(t, plugin.Init())

	// Build a metric datum and check
	input := metric.New(
		"cpu",
		map[string]string{
			"host": "example.org",
			"foo":  "",
		},
		map[string]interface{}{
			"value": int64(42),
		},
		time.Unix(0, 0),
	)
	datum := plugin.buildMetricDatum(input)
	require.Len(t, datum, 1)
	require.Len(t, datum[0].Dimensions, 1)
}

func TestPartitionDatums(t *testing.T) {
	const partitionSize = 2

	// Create a test-datum for keeping the tests short
	datum := types.MetricDatum{
		MetricName: aws.String("Foo"),
		Value:      aws.Float64(1),
	}

	tests := []struct {
		name     string
		input    []types.MetricDatum
		expected [][]types.MetricDatum
	}{
		{
			name:     "empty",
			input:    make([]types.MetricDatum, 0),
			expected: make([][]types.MetricDatum, 0),
		},
		{
			name: "single",
			input: []types.MetricDatum{
				datum,
			},
			expected: [][]types.MetricDatum{
				{
					datum,
				},
			},
		},
		{
			name: "two",
			input: []types.MetricDatum{
				datum,
				datum,
			},
			expected: [][]types.MetricDatum{
				{
					datum,
					datum,
				},
			},
		},
		{
			name: "three",
			input: []types.MetricDatum{
				datum,
				datum,
				datum,
			},
			expected: [][]types.MetricDatum{
				{
					datum,
					datum,
				},
				{
					datum,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, partitionDatums(tt.input, partitionSize))
		})
	}
}
