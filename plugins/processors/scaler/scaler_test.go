package scaler

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestScaler(t *testing.T) {
	tests := []struct {
		name     string
		scaler   *Scaler
		inputs   []telegraf.Metric
		expected []telegraf.Metric
	}{
		{
			name: "Field Scaling",
			scaler: &Scaler{
				Scalings: []Scaling{
					{
						InMin:  -1,
						InMax:  1,
						OutMin: 0,
						OutMax: 100,
						Fields: []string{"test1", "test2"},
					},
					{
						InMin:  -5,
						InMax:  0,
						OutMin: 1,
						OutMax: 9,
						Fields: []string{"test3", "test4"},
					},
				},
			},
			inputs: []telegraf.Metric{
				testutil.MustMetric("Name1", map[string]string{}, map[string]interface{}{"test1": int64(0), "test2": uint64(1)}, time.Unix(0, 0)),
				testutil.MustMetric("Name2", map[string]string{}, map[string]interface{}{"test1": "0.5", "test2": float32(-0.5)}, time.Unix(0, 0)),
				testutil.MustMetric("Name3", map[string]string{}, map[string]interface{}{"test3": int64(-3), "test4": uint64(0)}, time.Unix(0, 0)),
				testutil.MustMetric("Name4", map[string]string{}, map[string]interface{}{"test3": int64(-5), "test4": float32(-0.5)}, time.Unix(0, 0)),
			},
			expected: []telegraf.Metric{
				testutil.MustMetric("Name1", map[string]string{}, map[string]interface{}{"test1": float64(50), "test2": float64(100)}, time.Unix(0, 0)),
				testutil.MustMetric("Name2", map[string]string{}, map[string]interface{}{"test1": float64(75), "test2": float32(25)}, time.Unix(0, 0)),
				testutil.MustMetric("Name3", map[string]string{}, map[string]interface{}{"test3": float64(4.2), "test4": float64(9)}, time.Unix(0, 0)),
				testutil.MustMetric("Name4", map[string]string{}, map[string]interface{}{"test3": float64(1), "test4": float64(8.2)}, time.Unix(0, 0)),
			},
		},
		{
			name: "Ignored Fileds",
			scaler: &Scaler{
				Scalings: []Scaling{
					{
						InMin:  -1,
						InMax:  1,
						OutMin: 0,
						OutMax: 100,
						Fields: []string{"test1", "test2"},
					},
				},
			},
			inputs: []telegraf.Metric{
				testutil.MustMetric("Name1", map[string]string{}, map[string]interface{}{"test1": int64(0), "test2": uint64(1), "test3": int64(1)}, time.Unix(0, 0)),
			},
			expected: []telegraf.Metric{
				testutil.MustMetric("Name1", map[string]string{}, map[string]interface{}{"test1": float64(50), "test2": float64(100), "test3": int64(1)}, time.Unix(0, 0)),
			},
		},
		{
			name: "Out of range tests",
			scaler: &Scaler{
				Scalings: []Scaling{
					{
						InMin:  -1,
						InMax:  1,
						OutMin: 0,
						OutMax: 100,
						Fields: []string{"test1", "test2"},
					},
				},
			},
			inputs: []telegraf.Metric{
				testutil.MustMetric("Name1", map[string]string{}, map[string]interface{}{"test1": int64(-2), "test2": uint64(2)}, time.Unix(0, 0)),
			},
			expected: []telegraf.Metric{
				testutil.MustMetric("Name1", map[string]string{}, map[string]interface{}{"test1": float64(-50), "test2": float64(150)}, time.Unix(0, 0)),
			},
		},
		{
			name: "Missing field Fileds",
			scaler: &Scaler{
				Scalings: []Scaling{
					{
						InMin:  -1,
						InMax:  1,
						OutMin: 0,
						OutMax: 100,
						Fields: []string{"test1", "test2"},
					},
				},
			},
			inputs: []telegraf.Metric{
				testutil.MustMetric("Name1", map[string]string{}, map[string]interface{}{"test1": int64(0)}, time.Unix(0, 0)),
			},
			expected: []telegraf.Metric{
				testutil.MustMetric("Name1", map[string]string{}, map[string]interface{}{"test1": float64(50)}, time.Unix(0, 0)),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.scaler.Log = testutil.Logger{}

			require.NoError(t, tt.scaler.Init())
			actual := tt.scaler.Apply(tt.inputs...)

			testutil.RequireMetricsEqual(t, tt.expected, actual)
		})
	}
}

func TestNoScalerDefined(t *testing.T) {
	s := Scaler{Log: testutil.Logger{}}

	require.Error(t, s.Init())
}

func TestErrorCases(t *testing.T) {
	tests := []struct {
		name     string
		scaler   *Scaler
		inputs   []telegraf.Metric
		expected []telegraf.Metric
	}{
		{
			name: "Duplicate filter fields in one scaling",
			scaler: &Scaler{
				Scalings: []Scaling{
					{
						InMin:  -1,
						InMax:  1,
						OutMin: 0,
						OutMax: 100,
						Fields: []string{"test", "test"},
					},
				},
			},
			inputs: []telegraf.Metric{
				testutil.MustMetric("Name1", map[string]string{}, map[string]interface{}{"test1": int64(0), "test2": uint64(1)}, time.Unix(0, 0)),
			},
			expected: []telegraf.Metric{
				testutil.MustMetric("Name1", map[string]string{}, map[string]interface{}{"test1": float64(50), "test2": float64(100)}, time.Unix(0, 0)),
			},
		},
		{
			name: "Duplicate filter fields in multiple scalings",
			scaler: &Scaler{
				Scalings: []Scaling{
					{
						InMin:  -1,
						InMax:  1,
						OutMin: 0,
						OutMax: 100,
						Fields: []string{"test1", "test2"},
					},
					{
						InMin:  -1,
						InMax:  1,
						OutMin: 0,
						OutMax: 100,
						Fields: []string{"test1", "test3"},
					},
				},
			},
			inputs: []telegraf.Metric{
				testutil.MustMetric("Name1", map[string]string{}, map[string]interface{}{"test1": int64(0), "test2": uint64(1), "test3": int64(1)}, time.Unix(0, 0)),
			},
			expected: []telegraf.Metric{
				testutil.MustMetric("Name1", map[string]string{}, map[string]interface{}{"test1": float64(50), "test2": float64(100), "test3": int64(1)}, time.Unix(0, 0)),
			},
		},
		{
			name:   "No scalings",
			scaler: &Scaler{Log: testutil.Logger{}},
			inputs: []telegraf.Metric{
				testutil.MustMetric("Name1", map[string]string{}, map[string]interface{}{"test1": int64(0), "test2": uint64(1), "test3": int64(1)}, time.Unix(0, 0)),
			},
			expected: []telegraf.Metric{
				testutil.MustMetric("Name1", map[string]string{}, map[string]interface{}{"test1": float64(50), "test2": float64(100), "test3": int64(1)}, time.Unix(0, 0)),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.scaler.Log = testutil.Logger{}

			require.Error(t, tt.scaler.Init())
		})
	}
}
