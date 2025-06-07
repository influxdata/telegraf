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
	inMin  float64
	inMax  float64
	outMin float64
	outMax float64
	fields []string
}

type scalingValuesFactor struct {
	factor float64
	offset float64
	fields []string
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
					inMin:  -1,
					inMax:  1,
					outMin: 0,
					outMax: 100,
					fields: []string{"test1", "test2"},
				},
				{
					inMin:  -5,
					inMax:  0,
					outMin: 1,
					outMax: 10,
					fields: []string{"test3", "test4"},
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
			name: "Ignored fields",
			scale: []scalingValuesMinMax{
				{
					inMin:  -1,
					inMax:  1,
					outMin: 0,
					outMax: 100,
					fields: []string{"test1", "test2"},
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
					inMin:  -1,
					inMax:  1,
					outMin: 0,
					outMax: 100,
					fields: []string{"test1", "test2"},
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
			name: "Missing field fields",
			scale: []scalingValuesMinMax{
				{
					inMin:  -1,
					inMax:  1,
					outMin: 0,
					outMax: 100,
					fields: []string{"test1", "test2"},
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
				Scalings: make([]scaling, 0, len(tt.scale)),
				Log:      testutil.Logger{},
			}
			for i := range tt.scale {
				plugin.Scalings = append(plugin.Scalings, scaling{
					InMin:  &tt.scale[i].inMin,
					InMax:  &tt.scale[i].inMax,
					OutMin: &tt.scale[i].outMin,
					OutMax: &tt.scale[i].outMax,
					Fields: tt.scale[i].fields,
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
					factor: 50.0,
					offset: 50.0,
					fields: []string{"test1", "test2"},
				},
				{
					factor: 1.6,
					offset: 9.0,
					fields: []string{"test3", "test4"},
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
			name: "Ignored fields",
			scale: []scalingValuesFactor{
				{
					factor: 50.0,
					offset: 50.0,
					fields: []string{"test1", "test2"},
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
			name: "Missing field fields",
			scale: []scalingValuesFactor{
				{
					factor: 50.0,
					offset: 50.0,
					fields: []string{"test1", "test2"},
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
					factor: 50.0,
					fields: []string{"test1"},
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
					offset: 50.0,
					fields: []string{"test1"},
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
				Scalings: make([]scaling, 0, len(tt.scale)),
				Log:      testutil.Logger{},
			}
			for i := range tt.scale {
				s := scaling{
					Fields: tt.scale[i].fields,
				}
				if tt.scale[i].factor != 0.0 {
					s.Factor = &tt.scale[i].factor
				}
				if tt.scale[i].offset != 0.0 {
					s.Offset = &tt.scale[i].offset
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
		scaling          []scaling
		fields           []string
		expectedErrorMsg string
	}{
		{
			name: "Same input range values",
			scaling: []scaling{
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
			scaling: []scaling{
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
			scaling: []scaling{
				{
					Fields: []string{"test"},
				},
			},
			fields:           []string{"test"},
			expectedErrorMsg: "no scaling defined",
		},
		{
			name: "Partial minimum and maximum",
			scaling: []scaling{
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
			scaling: []scaling{
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
		Scalings: []scaling{
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
