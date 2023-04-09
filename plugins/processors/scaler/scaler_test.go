package scaler

import (
	"fmt"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func newMetric(name string, fields map[string]interface{}) telegraf.Metric {
	if fields == nil {
		fields = map[string]interface{}{}
	}
	m := metric.New(name, map[string]string{}, fields, time.Now())
	return m
}

func TestScaler(t *testing.T) {
	s := Scaler{
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
	}

	err := s.Init()
	require.NoError(t, err)

	m1 := newMetric("Name1", map[string]interface{}{"test1": int64(0), "test2": uint64(1)})
	m2 := newMetric("Name2", map[string]interface{}{"test1": float64(0.5), "test2": float32(-0.5)})
	m3 := newMetric("Name3", map[string]interface{}{"test3": int64(-3), "test4": uint64(0)})
	m4 := newMetric("Name4", map[string]interface{}{"test3": int64(-5), "test4": float32(-0.5)})

	results := s.Apply(m1, m2, m3, m4)

	val, ok := results[0].GetField("test1")
	require.True(t, ok)
	require.InEpsilon(t, float64(50), val, 1e-10)

	val, ok = results[0].GetField("test2")
	require.True(t, ok)
	require.InEpsilon(t, float64(100), val, 1e-10)

	val, ok = results[1].GetField("test1")
	require.True(t, ok)
	require.InEpsilon(t, float64(75), val, 1e-10)

	val, ok = results[1].GetField("test2")
	require.True(t, ok)
	require.InEpsilon(t, float64(25), val, 1e-10)

	val, ok = results[2].GetField("test3")
	require.True(t, ok)
	require.InEpsilon(t, float64(4.2), val, 1e-10)

	val, ok = results[2].GetField("test4")
	require.True(t, ok)
	require.InEpsilon(t, float64(9), val, 1e-10)

	val, ok = results[3].GetField("test3")
	require.True(t, ok)
	require.InEpsilon(t, float64(1), val, 1e-10)

	val, ok = results[3].GetField("test4")
	require.True(t, ok)
	require.InEpsilon(t, float64(8.2), val, 1e-10)
}

func TestOutOfInputRange(t *testing.T) {
	s := Scaler{
		Scalings: []Scaling{
			{
				InMin:  -1,
				InMax:  1,
				OutMin: 0,
				OutMax: 100,
				Fields: []string{"test1", "test2"},
			},
		},
	}

	err := s.Init()
	require.NoError(t, err)

	m1 := newMetric("Name1", map[string]interface{}{"test1": int64(-2), "test2": uint64(2)})

	results := s.Apply(m1)

	val, ok := results[0].GetField("test1")
	require.True(t, ok)
	require.InEpsilon(t, float64(-50), val, 1e-10)

	val, ok = results[0].GetField("test2")
	require.True(t, ok)
	require.InEpsilon(t, float64(150), val, 1e-10)
}

func TestNoFiltersDefined(t *testing.T) {
	s := Scaler{
		Scalings: []Scaling{
			{
				InMin:  -1,
				InMax:  1,
				OutMin: 0,
				OutMax: 100,
				Fields: []string{},
			},
		},
	}

	err := s.Init()
	require.NoError(t, err)

	m1 := newMetric("Name1", map[string]interface{}{"test1": int64(-2), "test2": uint64(2)})

	results := s.Apply(m1)

	val, ok := results[0].GetField("test1")
	require.True(t, ok)
	require.InEpsilon(t, float64(-2), val, 1e-10)

	val, ok = results[0].GetField("test2")
	require.True(t, ok)
	require.InEpsilon(t, float64(2), val, 1e-10)
}

func TestNoScalerDefined(t *testing.T) {
	s := Scaler{Log: testutil.Logger{}}

	err := s.Init()
	require.NoError(t, err)

	m1 := newMetric("Name1", map[string]interface{}{"test1": int64(-2), "test2": uint64(2)})

	results := s.Apply(m1)

	val, ok := results[0].GetField("test1")
	require.True(t, ok)
	fmt.Printf("val %v\n", val)
	require.InEpsilon(t, float64(-2), val, 1e-10)

	val, ok = results[0].GetField("test2")
	require.True(t, ok)
	fmt.Printf("val %v\n", val)
	require.InEpsilon(t, float64(2), val, 1e-10)
}
