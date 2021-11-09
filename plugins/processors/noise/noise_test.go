package noise

import (
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/rand"
	"gonum.org/v1/gonum/stat/distuv"
)

// Helper function which returns a map for metrics for easy lookup
func createTestMetrics() map[string]telegraf.Metric {
	m := make(map[string]telegraf.Metric)
	m["cpu"] = testutil.MustMetric(
		"cpu",
		map[string]string{},
		map[string]interface{}{
			"usage_guest":  2.5,
			"usage_system": 1.5,
			"usage_nice":   0.5,
			"usage_irq":    1.0,
		},
		time.Unix(0, 0),
	)
	m["disk"] = testutil.MustMetric(
		"disk",
		map[string]string{},
		map[string]interface{}{
			"free":        250,
			"inodes_free": 1500.0,
			"inodes_used": 1337,
		},
		time.Unix(0, 0),
	)
	return m
}

// Verifies that field values are modified by the Laplace noise
func TestAddNoiseToMetric(t *testing.T) {
	generators := []string{"laplace", "gaussian", "uniform"}
	for _, generator := range generators {
		p := Noise{
			NoiseType: generator,
			Scale:     1.0,
			Mu:        0.0,
			Sigma:     0.2,
			Min:       -1,
			Max:       1,
			Log:       testutil.Logger{},
		}
		_ = p.Init()
		metrics := testutil.MockMetrics()
		after := p.Apply(metrics[0].Copy())[0]
		require.NotEqual(t, metrics[0], after)
	}
}

// Verifies that no int64 or uint64 overflow occurs while adding noise
func TestAddNoiseOverflowCheck(t *testing.T) {
	p := Noise{
		NoiseType: "laplace",
		Scale:     1.0,
		Log:       testutil.Logger{},
	}
	_ = p.Init()
	p.Generator = &distuv.Laplace{Mu: p.Mu, Scale: p.Scale, Src: rand.NewSource(4)}
	require.Equal(t, math.MaxInt64, p.addNoise(math.MaxInt64))

	// rand should be -1.3528540286500519, a underflow happens for uint64
	require.Equal(t, uint64(0), p.addNoise(uint64(math.MaxUint64)))

	// rand should be 0.813394603644003, a positive overflow happens
	p.Generator = &distuv.Laplace{Mu: p.Mu, Scale: p.Scale, Src: rand.NewSource(2)}
	require.Equal(t, math.MaxInt64, p.addNoise(math.MaxInt64))
	require.Equal(t, uint64(math.MaxUint64), p.addNoise(uint64(math.MaxUint64)))
}

// Verifies that even addNoise() modifies 0 values as well
func TestAddNoiseWithZeroValue(t *testing.T) {
	p := Noise{
		NoiseType: "laplace",
		Scale:     1.0,
		Log:       testutil.Logger{},
	}
	_ = p.Init()
	noise := p.addNoise(0)
	require.NotEqual(t, noise, 0)
}

// Verifies that any invalid generator setting (not "laplace", "gaussian" or
// "uniform") raises an error
func TestInvalidDistributionFunction(t *testing.T) {
	p := Noise{
		NoiseType: "invalid",
		Log:       testutil.Logger{},
	}
	require.Error(t, p.Init())
}

// Tests that fields in the IgnoreField set are not affected by the Laplace
// noise
func TestLaplaceNoiseWithIgnoreField(t *testing.T) {
	p := Noise{
		Scale:         1.0,
		IncludeFields: []string{},
		ExcludeFields: []string{"usage_guest", "usage_system"},
		Log:           testutil.Logger{},
	}

	// call Init as we want to initialize the excludeLists
	_ = p.Init()
	metric := createTestMetrics()
	after := p.Apply(metric["cpu"].Copy())[0]

	// check that some values in the struct have been changed
	require.NotEqual(t, metric, after)

	// check that ignore values were not changed
	for _, ignore := range p.ExcludeFields {
		have, _ := metric["cpu"].GetField(ignore)
		should, _ := after.GetField(ignore)
		require.Equal(t, have, should)
	}
}

func TestAddNoiseToValue(t *testing.T) {
	p := Noise{
		Scale: 2.0,
		Log:   testutil.Logger{},
	}
	_ = p.Init()
	haveValues := []interface{}{
		int64(-51232),
		uint64(45123),
		uint64(100),
		uint64(2),
		uint64(999862),
		float64(1.337),
	}

	haveValuesInvalid := []interface{}{
		string("helloworld"),
	}

	for _, value := range haveValues {
		after := p.addNoise(value)
		// check value is not the same
		require.NotEqual(t, value, after)

		// check type is still the same
		require.Equal(t, reflect.TypeOf(value), reflect.TypeOf(after))
	}

	// check that nothing happens to non numerical types:
	for _, value := range haveValuesInvalid {
		after := p.addNoise(value)
		require.Equal(t, value, after)
		require.Equal(t, reflect.TypeOf(value), reflect.TypeOf(after))
	}
}
