package math

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/stretchr/testify/assert"
)

func getM1() telegraf.Metric {
	var m1, _ = metric.New("m1",
		map[string]string{"foo": "bar"},
		map[string]interface{}{
			"a": float64(-1),
			"b": float64(1),
			"c": float64(1),
		},
		time.Now(),
	)
	return m1
}

func getM2() telegraf.Metric {
	var m2, _ = metric.New("m2",
		map[string]string{"foo": "bar"},
		map[string]interface{}{
			"a": float64(-1),
			"b": float64(1),
			"c": float64(1),
			"d": string("d"),
		},
		time.Now(),
	)
	return m2
}

func BenchmarkApply(b *testing.B) {
	fields := []string{"a"}
	mathProcessor := &Math{Metric: "m1", Func: "abs", Fields: fields}

	for n := 0; n < b.N; n++ {
		mathProcessor.Apply(getM1())
		mathProcessor.Apply(getM2())
	}
}

func TestAllFieldsProcess(t *testing.T) {
	var fields []string
	mathProcessor := &Math{Metric: "m1", Func: "abs", Fields: fields}

	in := mathProcessor.Apply(getM1())

	expectedFields := map[string]interface{}{
		"a":     float64(-1),
		"b":     float64(1),
		"c":     float64(1),
		"a_abs": float64(1),
		"b_abs": float64(1),
		"c_abs": float64(1),
	}
	expectedTags := map[string]string{
		"foo": "bar",
	}

	assert.Equal(t, expectedFields, in[0].Fields())
	assert.Equal(t, expectedTags, in[0].Tags())
	assert.Equal(t, in[0].Name(), "m1")

}

func TestAllFieldsWrongTypeProcess(t *testing.T) {
	var fields []string
	mathProcessor := &Math{Metric: "m2", Func: "abs", Fields: fields}

	in := mathProcessor.Apply(getM2())

	expectedFields := map[string]interface{}{
		"a":     float64(-1),
		"b":     float64(1),
		"c":     float64(1),
		"d":     string("d"),
		"a_abs": float64(1),
		"b_abs": float64(1),
		"c_abs": float64(1),
	}
	expectedTags := map[string]string{
		"foo": "bar",
	}

	assert.Equal(t, expectedFields, in[0].Fields())
	assert.Equal(t, expectedTags, in[0].Tags())
	assert.Equal(t, in[0].Name(), "m2")

}

func TestOneFieldsProcess(t *testing.T) {
	fields := []string{"a"}
	mathProcessor := &Math{Metric: "m1", Func: "abs", Fields: fields}

	in := mathProcessor.Apply(getM1())

	expectedFields := map[string]interface{}{
		"a":     float64(-1),
		"b":     float64(1),
		"c":     float64(1),
		"a_abs": float64(1),
	}
	expectedTags := map[string]string{
		"foo": "bar",
	}

	assert.Equal(t, expectedFields, in[0].Fields())
	assert.Equal(t, expectedTags, in[0].Tags())
	assert.Equal(t, in[0].Name(), "m1")

}

func TestOneFieldsWrongTypeProcess(t *testing.T) {
	fields := []string{"d"}
	mathProcessor := &Math{Metric: "m2", Func: "abs", Fields: fields}

	in := mathProcessor.Apply(getM2())

	expectedFields := map[string]interface{}{
		"a": float64(-1),
		"b": float64(1),
		"c": float64(1),
		"d": string("d"),
	}
	expectedTags := map[string]string{
		"foo": "bar",
	}

	assert.Equal(t, expectedFields, in[0].Fields())
	assert.Equal(t, expectedTags, in[0].Tags())
	assert.Equal(t, in[0].Name(), "m2")

}
