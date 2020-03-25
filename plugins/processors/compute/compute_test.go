package compute

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

type testcase struct {
	name     string
	compute  *Compute
	input    telegraf.Metric
	expected telegraf.Metric
}

func TestComputeCorrectness(t *testing.T) {
	tests := []testcase{
		{
			name: "constant arithmetic",
			compute: &Compute{
				Missing: "ignore",
				Fields: map[string]string{
					"x1": "3 + 1",
					"x2": "3.1415 + 2.7182",
					"x3": "23 - 42",
					"x4": "23 - 42.42",
					"x5": "1 / 3",
					"x6": "12 / 3",
					"x7": "1.0 / 3.0",
					"x8": "12.345 * 2.1",
					"x9": "12 * 3",
				},
			},
			input: testutil.MustMetric(
				"test",
				map[string]string{},
				map[string]interface{}{},
				time.Unix(0, 0),
			),
			expected: testutil.MustMetric(
				"test",
				map[string]string{},
				map[string]interface{}{
					"x1": int64(3 + 1),
					"x2": float64(3.1415 + 2.7182),
					"x3": int64(23 - 42),
					"x4": float64(23 - 42.42),
					"x5": int64(1 / 3),
					"x6": int64(12 / 3),
					"x7": float64(1.0 / 3.0),
					"x8": float64(12.345 * 2.1),
					"x9": int64(12 * 3),
				},
				time.Unix(0, 0),
			),
		},
		{
			name: "operation order",
			compute: &Compute{
				Missing: "ignore",
				Fields: map[string]string{
					"x1": "(3 + 1) / 2",
					"x2": "4.2 + 3.1 * 2.0",
					"x3": "(2.1 + 1.2) / (0.5 - 8.0)",
					"x4": "12.0/3.3 * 2.0 - 3.1415",
				},
			},
			input: testutil.MustMetric(
				"test",
				map[string]string{},
				map[string]interface{}{},
				time.Unix(0, 0),
			),
			expected: testutil.MustMetric(
				"test",
				map[string]string{},
				map[string]interface{}{
					"x1": int64((3 + 1) / 2),
					"x2": float64(4.2 + 3.1*2.0),
					"x3": float64((2.1 + 1.2) / (0.5 - 8.0)),
					"x4": float64(12.0/3.3*2.0 - 3.1415),
				},
				time.Unix(0, 0),
			),
		},
		{
			name: "field variables",
			compute: &Compute{
				Missing: "ignore",
				Fields: map[string]string{
					"x1": "a + b",
					"x2": "2 * a",
					"x3": "2.1 * a",
					"x4": "a - c",
					"x5": "b / d",
					"x6": "2 * c",
				},
			},
			input: testutil.MustMetric(
				"test",
				map[string]string{},
				map[string]interface{}{
					"a": 1,
					"b": 2,
					"c": 3.1415,
					"d": 2.7182,
				},
				time.Unix(0, 0),
			),
			expected: testutil.MustMetric(
				"test",
				map[string]string{},
				map[string]interface{}{
					"a":  1,
					"b":  2,
					"c":  3.1415,
					"d":  2.7182,
					"x1": int64(1 + 2),
					"x2": int64(2 * 1),
					"x3": float64(2.1 * 1),
					"x4": float64(1 - 3.1415),
					"x5": float64(2 / 2.7182),
					"x6": float64(2 * 3.1415),
				},
				time.Unix(0, 0),
			),
		},
		{
			name: "functions",
			compute: &Compute{
				Missing: "ignore",
				Fields: map[string]string{
					"x1": "pow(a, 2)",
					"x2": "pow(a, 2.0)",
					"x3": "pow(b, 2)",
					"x4": "pow(b, 2.0)",
					"x5": "abs(-42)",
					"x6": "abs(-42.23)",
				},
			},
			input: testutil.MustMetric(
				"test",
				map[string]string{},
				map[string]interface{}{
					"a": 2,
					"b": 3.1415,
				},
				time.Unix(0, 0),
			),
			expected: testutil.MustMetric(
				"test",
				map[string]string{},
				map[string]interface{}{
					"a":  2,
					"b":  3.1415,
					"x1": float64(4.0),
					"x2": float64(4.0),
					"x3": float64(3.1415 * 3.1415),
					"x4": float64(3.1415 * 3.1415),
					"x5": int64(42),
					"x6": float64(42.23),
				},
				time.Unix(0, 0),
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup the processor
			err := tt.compute.Init()
			require.NoError(t, err)

			// Do the processing
			actual := tt.compute.Apply(tt.input)

			// We expect only one metric
			require.Len(t, actual, 1)

			// Test with floating point precision in mind
			testutil.RequireMetricEqual(t, tt.expected, actual[0], cmpopts.EquateApprox(0.0, 1e-6))
		})
	}
}

func TestComputeMissingStrategy(t *testing.T) {
	tests := []testcase{
		{
			name: "ignore",
			compute: &Compute{
				Missing: "ignore",
				Fields: map[string]string{
					"x": "b + 1",
				},
			},
			input: testutil.MustMetric(
				"test",
				map[string]string{},
				map[string]interface{}{
					"a": 1,
				},
				time.Unix(0, 0),
			),
			expected: testutil.MustMetric(
				"test",
				map[string]string{},
				map[string]interface{}{
					"a": 1,
				},
				time.Unix(0, 0),
			),
		},
		{
			name: "constant (default)",
			compute: &Compute{
				Missing: "const",
				Fields: map[string]string{
					"x": "b + 1",
				},
			},
			input: testutil.MustMetric(
				"test",
				map[string]string{},
				map[string]interface{}{
					"a": 1,
				},
				time.Unix(0, 0),
			),
			expected: testutil.MustMetric(
				"test",
				map[string]string{},
				map[string]interface{}{
					"a": 1,
					"x": 0,
				},
				time.Unix(0, 0),
			),
		},
		{
			name: "constant (integer)",
			compute: &Compute{
				Missing:  "const",
				Constant: 42,
				Fields: map[string]string{
					"x": "b + 1",
				},
			},
			input: testutil.MustMetric(
				"test",
				map[string]string{},
				map[string]interface{}{
					"a": 1,
				},
				time.Unix(0, 0),
			),
			expected: testutil.MustMetric(
				"test",
				map[string]string{},
				map[string]interface{}{
					"a": 1,
					"x": 42,
				},
				time.Unix(0, 0),
			),
		},
		{
			name: "constant (float)",
			compute: &Compute{
				Missing:  "const",
				Constant: 42.1,
				Fields: map[string]string{
					"x": "b + 1",
				},
			},
			input: testutil.MustMetric(
				"test",
				map[string]string{},
				map[string]interface{}{
					"a": 1,
				},
				time.Unix(0, 0),
			),
			expected: testutil.MustMetric(
				"test",
				map[string]string{},
				map[string]interface{}{
					"a": 1,
					"x": 42.1,
				},
				time.Unix(0, 0),
			),
		},
		{
			name: "default (default)",
			compute: &Compute{
				Missing: "default",
				Fields: map[string]string{
					"x": "b + 1",
				},
			},
			input: testutil.MustMetric(
				"test",
				map[string]string{},
				map[string]interface{}{
					"a": 1,
				},
				time.Unix(0, 0),
			),
			expected: testutil.MustMetric(
				"test",
				map[string]string{},
				map[string]interface{}{
					"a": 1,
					"x": 1,
				},
				time.Unix(0, 0),
			),
		},
		{
			name: "default (integer)",
			compute: &Compute{
				Missing: "default",
				Default: 42,
				Fields: map[string]string{
					"x": "b + 1",
				},
			},
			input: testutil.MustMetric(
				"test",
				map[string]string{},
				map[string]interface{}{
					"a": 1,
				},
				time.Unix(0, 0),
			),
			expected: testutil.MustMetric(
				"test",
				map[string]string{},
				map[string]interface{}{
					"a": 1,
					"x": 43,
				},
				time.Unix(0, 0),
			),
		},
		{
			name: "default (float)",
			compute: &Compute{
				Missing: "default",
				Default: 42.1,
				Fields: map[string]string{
					"x": "b + 1",
				},
			},
			input: testutil.MustMetric(
				"test",
				map[string]string{},
				map[string]interface{}{
					"a": 1,
				},
				time.Unix(0, 0),
			),
			expected: testutil.MustMetric(
				"test",
				map[string]string{},
				map[string]interface{}{
					"a": 1,
					"x": 43.1,
				},
				time.Unix(0, 0),
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup the processor
			err := tt.compute.Init()
			require.NoError(t, err)

			// Do the processing
			actual := tt.compute.Apply(tt.input)

			// We expect only one metric
			require.Len(t, actual, 1)

			testutil.RequireMetricEqual(t, tt.expected, actual[0])
		})
	}
}

func TestComputeFieldCollision(t *testing.T) {
	tests := []testcase{
		{
			name: "collision",
			compute: &Compute{
				Missing: "ignore",
				Fields: map[string]string{
					"x": "a + 1",
				},
			},
			input: testutil.MustMetric(
				"test",
				map[string]string{},
				map[string]interface{}{
					"a": 1,
					"x": 25,
				},
				time.Unix(0, 0),
			),
			expected: testutil.MustMetric(
				"test",
				map[string]string{},
				map[string]interface{}{
					"a": 1,
					"x": 2,
				},
				time.Unix(0, 0),
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup the processor
			err := tt.compute.Init()
			require.NoError(t, err)

			// Do the processing
			actual := tt.compute.Apply(tt.input)

			// We expect only one metric
			require.Len(t, actual, 1)

			testutil.RequireMetricEqual(t, tt.expected, actual[0])
		})
	}
}

func TestEmptyConfigInitError(t *testing.T) {
	computer := &Compute{}
	err := computer.Init()
	require.Error(t, err)
}
