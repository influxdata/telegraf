package scale

import (
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)

type scalingValuesMinMax struct {
	InMin  float64
	InMax  float64
	OutMin float64
	OutMax float64
	Fields []string
}

type scalingValuesFactor struct {
	Factor float64
	Offset float64
	Fields []string
}

func TestMinMax(t *testing.T) {
	tests := []struct {
		name     string
		scale    []scalingValuesMinMax
		inputs   []telegraf.Metric
		expected []telegraf.Metric
	}{
		{
			name: "Field Scaling",
			scale: []scalingValuesMinMax{
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
					OutMax: 10,
					Fields: []string{"test3", "test4"},
				},
			},
			inputs: []telegraf.Metric{
				testutil.MustMetric("Name1", map[string]string{},
					map[string]interface{}{
						"test1": int64(0),
						"test2": uint64(1),
					}, time.Unix(0, 0)),
				testutil.MustMetric("Name2", map[string]string{},
					map[string]interface{}{
						"test1": "0.5",
						"test2": float32(-0.5),
					}, time.Unix(0, 0)),
				testutil.MustMetric("Name3", map[string]string{},
					map[string]interface{}{
						"test3": int64(-3),
						"test4": uint64(0),
					}, time.Unix(0, 0)),
				testutil.MustMetric("Name4", map[string]string{},
					map[string]interface{}{
						"test3": int64(-5),
						"test4": float32(-0.5),
					}, time.Unix(0, 0)),
			},
			expected: []telegraf.Metric{
				testutil.MustMetric("Name1", map[string]string{},
					map[string]interface{}{
						"test1": float64(50),
						"test2": float64(100),
					}, time.Unix(0, 0)),
				testutil.MustMetric("Name2", map[string]string{},
					map[string]interface{}{
						"test1": float64(75),
						"test2": float32(25),
					}, time.Unix(0, 0)),
				testutil.MustMetric("Name3", map[string]string{},
					map[string]interface{}{
						"test3": float64(4.6),
						"test4": float64(10),
					}, time.Unix(0, 0)),
				testutil.MustMetric("Name4", map[string]string{},
					map[string]interface{}{
						"test3": float64(1),
						"test4": float64(9.1),
					}, time.Unix(0, 0)),
			},
		},
		{
			name: "Ignored Fields",
			scale: []scalingValuesMinMax{
				{
					InMin:  -1,
					InMax:  1,
					OutMin: 0,
					OutMax: 100,
					Fields: []string{"test1", "test2"},
				},
			},
			inputs: []telegraf.Metric{
				testutil.MustMetric("Name1", map[string]string{},
					map[string]interface{}{
						"test1": int64(0),
						"test2": uint64(1),
						"test3": int64(1),
					}, time.Unix(0, 0)),
			},
			expected: []telegraf.Metric{
				testutil.MustMetric("Name1", map[string]string{},
					map[string]interface{}{
						"test1": float64(50),
						"test2": float64(100),
						"test3": int64(1),
					}, time.Unix(0, 0)),
			},
		},
		{
			name: "Out of range tests",
			scale: []scalingValuesMinMax{
				{
					InMin:  -1,
					InMax:  1,
					OutMin: 0,
					OutMax: 100,
					Fields: []string{"test1", "test2"},
				},
			},
			inputs: []telegraf.Metric{
				testutil.MustMetric("Name1", map[string]string{},
					map[string]interface{}{
						"test1": int64(-2),
						"test2": uint64(2),
					}, time.Unix(0, 0)),
			},
			expected: []telegraf.Metric{
				testutil.MustMetric("Name1", map[string]string{},
					map[string]interface{}{
						"test1": float64(-50),
						"test2": float64(150),
					}, time.Unix(0, 0)),
			},
		},
		{
			name: "Missing field Fields",
			scale: []scalingValuesMinMax{
				{
					InMin:  -1,
					InMax:  1,
					OutMin: 0,
					OutMax: 100,
					Fields: []string{"test1", "test2"},
				},
			},
			inputs: []telegraf.Metric{
				testutil.MustMetric("Name1", map[string]string{},
					map[string]interface{}{
						"test1": int64(0),
					}, time.Unix(0, 0)),
			},
			expected: []telegraf.Metric{
				testutil.MustMetric("Name1", map[string]string{},
					map[string]interface{}{
						"test1": float64(50),
					}, time.Unix(0, 0)),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := &Scale{
				Scalings: make([]Scaling, 0, len(tt.scale)),
				Log:      testutil.Logger{},
			}
			for i := range tt.scale {
				plugin.Scalings = append(plugin.Scalings, Scaling{
					InMin:  &tt.scale[i].InMin,
					InMax:  &tt.scale[i].InMax,
					OutMin: &tt.scale[i].OutMin,
					OutMax: &tt.scale[i].OutMax,
					Fields: tt.scale[i].Fields,
				})
			}
			require.NoError(t, plugin.Init())
			actual := plugin.Apply(tt.inputs...)

			testutil.RequireMetricsEqual(t, tt.expected, actual)
		})
	}
}

func TestFactor(t *testing.T) {
	tests := []struct {
		name     string
		scale    []scalingValuesFactor
		inputs   []telegraf.Metric
		expected []telegraf.Metric
	}{
		{
			name: "Field Scaling",
			scale: []scalingValuesFactor{
				{
					Factor: 50.0,
					Offset: 50.0,
					Fields: []string{"test1", "test2"},
				},
				{
					Factor: 1.6,
					Offset: 9.0,
					Fields: []string{"test3", "test4"},
				},
			},
			inputs: []telegraf.Metric{
				testutil.MustMetric("Name1", map[string]string{},
					map[string]interface{}{
						"test1": int64(0),
						"test2": uint64(1),
					}, time.Unix(0, 0)),
				testutil.MustMetric("Name2", map[string]string{},
					map[string]interface{}{
						"test1": "0.5",
						"test2": float32(-0.5),
					}, time.Unix(0, 0)),
				testutil.MustMetric("Name3", map[string]string{},
					map[string]interface{}{
						"test3": int64(-3),
						"test4": uint64(0),
					}, time.Unix(0, 0)),
				testutil.MustMetric("Name4", map[string]string{},
					map[string]interface{}{
						"test3": int64(-5),
						"test4": float32(-0.5),
					}, time.Unix(0, 0)),
			},
			expected: []telegraf.Metric{
				testutil.MustMetric("Name1", map[string]string{},
					map[string]interface{}{
						"test1": float64(50),
						"test2": float64(100),
					}, time.Unix(0, 0)),
				testutil.MustMetric("Name2", map[string]string{},
					map[string]interface{}{
						"test1": float64(75),
						"test2": float32(25),
					}, time.Unix(0, 0)),
				testutil.MustMetric("Name3", map[string]string{},
					map[string]interface{}{
						"test3": float64(4.2),
						"test4": float64(9),
					}, time.Unix(0, 0)),
				testutil.MustMetric("Name4", map[string]string{},
					map[string]interface{}{
						"test3": float64(1),
						"test4": float64(8.2),
					}, time.Unix(0, 0)),
			},
		},
		{
			name: "Ignored Fields",
			scale: []scalingValuesFactor{
				{
					Factor: 50.0,
					Offset: 50.0,
					Fields: []string{"test1", "test2"},
				},
			},
			inputs: []telegraf.Metric{
				testutil.MustMetric("Name1", map[string]string{},
					map[string]interface{}{
						"test1": int64(0),
						"test2": uint64(1),
						"test3": int64(1),
					}, time.Unix(0, 0)),
			},
			expected: []telegraf.Metric{
				testutil.MustMetric("Name1", map[string]string{},
					map[string]interface{}{
						"test1": float64(50),
						"test2": float64(100),
						"test3": int64(1),
					}, time.Unix(0, 0)),
			},
		},
		{
			name: "Missing field Fields",
			scale: []scalingValuesFactor{
				{
					Factor: 50.0,
					Offset: 50.0,
					Fields: []string{"test1", "test2"},
				},
			},
			inputs: []telegraf.Metric{
				testutil.MustMetric("Name1", map[string]string{},
					map[string]interface{}{
						"test1": int64(0),
					}, time.Unix(0, 0)),
			},
			expected: []telegraf.Metric{
				testutil.MustMetric("Name1", map[string]string{},
					map[string]interface{}{
						"test1": float64(50),
					}, time.Unix(0, 0)),
			},
		},
		{
			name: "No Offset",
			scale: []scalingValuesFactor{
				{
					Factor: 50.0,
					Fields: []string{"test1"},
				},
			},
			inputs: []telegraf.Metric{
				testutil.MustMetric("Name1", map[string]string{},
					map[string]interface{}{
						"test1": int64(1),
					}, time.Unix(0, 0)),
			},
			expected: []telegraf.Metric{
				testutil.MustMetric("Name1", map[string]string{},
					map[string]interface{}{
						"test1": float64(50),
					}, time.Unix(0, 0)),
			},
		},
		{
			name: "No Factor",
			scale: []scalingValuesFactor{
				{
					Offset: 50.0,
					Fields: []string{"test1"},
				},
			},
			inputs: []telegraf.Metric{
				testutil.MustMetric("Name1", map[string]string{},
					map[string]interface{}{
						"test1": int64(1),
					}, time.Unix(0, 0)),
			},
			expected: []telegraf.Metric{
				testutil.MustMetric("Name1", map[string]string{},
					map[string]interface{}{
						"test1": float64(51),
					}, time.Unix(0, 0)),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := &Scale{
				Scalings: make([]Scaling, 0, len(tt.scale)),
				Log:      testutil.Logger{},
			}
			for i := range tt.scale {
				s := Scaling{
					Fields: tt.scale[i].Fields,
				}
				if tt.scale[i].Factor != 0.0 {
					s.Factor = &tt.scale[i].Factor
				}
				if tt.scale[i].Offset != 0.0 {
					s.Offset = &tt.scale[i].Offset
				}
				plugin.Scalings = append(plugin.Scalings, s)
			}
			require.NoError(t, plugin.Init())
			actual := plugin.Apply(tt.inputs...)

			testutil.RequireMetricsEqual(t, tt.expected, actual, cmpopts.EquateApprox(0, 1e-6))
		})
	}
}

func TestErrorCasesMinMax(t *testing.T) {
	a0, a1, a100 := float64(0.0), float64(1.0), float64(100.0)
	tests := []struct {
		name             string
		scaling          []Scaling
		fields           []string
		expectedErrorMsg string
	}{
		{
			name: "Same input range values",
			scaling: []Scaling{
				{
					InMin:  &a1,
					InMax:  &a1,
					OutMin: &a0,
					OutMax: &a100,
					Fields: []string{"test"},
				},
			},
			fields:           []string{"test"},
			expectedErrorMsg: "input minimum and maximum are equal for fields test",
		},
		{
			name: "Same input range values",
			scaling: []Scaling{
				{
					InMin:  &a0,
					InMax:  &a1,
					OutMin: &a100,
					OutMax: &a100,
					Fields: []string{"test"},
				},
			},
			fields:           []string{"test"},
			expectedErrorMsg: "output minimum and maximum are equal",
		},
		{
			name: "Nothing set",
			scaling: []Scaling{
				{
					Fields: []string{"test"},
				},
			},
			fields:           []string{"test"},
			expectedErrorMsg: "no scaling defined",
		},
		{
			name: "Partial minimum and maximum",
			scaling: []Scaling{
				{
					InMin:  &a0,
					Fields: []string{"test"},
				},
			},
			fields:           []string{"test"},
			expectedErrorMsg: "all minimum and maximum values need to be set",
		},
		{
			name: "Mixed minimum, maximum and factor",
			scaling: []Scaling{
				{
					InMin:  &a0,
					InMax:  &a1,
					OutMin: &a100,
					OutMax: &a100,
					Factor: &a1,
					Fields: []string{"test"},
				},
			},
			fields:           []string{"test"},
			expectedErrorMsg: "cannot use factor/offset and minimum/maximum at the same time",
		},
		{
			name:             "No scaling",
			expectedErrorMsg: "no valid scaling defined",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := &Scale{
				Scalings: tt.scaling,
				Log:      testutil.Logger{},
			}
			err := plugin.Init()
			require.ErrorContains(t, err, tt.expectedErrorMsg)
		})
	}
}

func TestTracking(t *testing.T) {
	inputRaw := []telegraf.Metric{
		metric.New("foo", map[string]string{}, map[string]interface{}{"value": 42}, time.Unix(0, 0)),
		metric.New("bar", map[string]string{}, map[string]interface{}{"value": 99}, time.Unix(0, 0)),
		metric.New("baz", map[string]string{}, map[string]interface{}{"value": 1}, time.Unix(0, 0)),
	}

	var mu sync.Mutex
	delivered := make([]telegraf.DeliveryInfo, 0, len(inputRaw))
	notify := func(di telegraf.DeliveryInfo) {
		mu.Lock()
		defer mu.Unlock()
		delivered = append(delivered, di)
	}

	input := make([]telegraf.Metric, 0, len(inputRaw))
	for _, m := range inputRaw {
		tm, _ := metric.WithTracking(m, notify)
		input = append(input, tm)
	}

	expected := []telegraf.Metric{
		metric.New(
			"foo",
			map[string]string{},
			map[string]interface{}{"value": float64(92)},
			time.Unix(0, 0),
		),
		metric.New(
			"bar",
			map[string]string{},
			map[string]interface{}{"value": float64(149)},
			time.Unix(0, 0),
		),
		metric.New(
			"baz",
			map[string]string{},
			map[string]interface{}{"value": float64(51)},
			time.Unix(0, 0),
		),
	}

	inMin := float64(0)
	inMax := float64(50)
	outMin := float64(50)
	outMax := float64(100)

	plugin := &Scale{
		Scalings: []Scaling{
			{
				InMin:  &inMin,
				InMax:  &inMax,
				OutMin: &outMin,
				OutMax: &outMax,
				Fields: []string{"value"},
			},
		},
	}
	require.NoError(t, plugin.Init())

	// Process expected metrics and compare with resulting metrics
	actual := plugin.Apply(input...)
	testutil.RequireMetricsEqual(t, expected, actual)

	// Simulate output acknowledging delivery
	for _, m := range actual {
		m.Accept()
	}

	// Check delivery
	require.Eventuallyf(t, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return len(input) == len(delivered)
	}, time.Second, 100*time.Millisecond, "%d delivered but %d expected", len(delivered), len(expected))
}
