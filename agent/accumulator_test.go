package agent

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdd(t *testing.T) {
	now := time.Now()
	metrics := make(chan telegraf.Metric, 10)
	defer close(metrics)
	a := NewAccumulator(&TestMetricMaker{}, metrics)

	a.AddFields("acctest",
		map[string]interface{}{"value": float64(101)},
		map[string]string{})
	a.AddFields("acctest",
		map[string]interface{}{"value": float64(101)},
		map[string]string{"acc": "test"})
	a.AddFields("acctest",
		map[string]interface{}{"value": float64(101)},
		map[string]string{"acc": "test"}, now)

	testm := <-metrics
	actual := testm.String()
	assert.Contains(t, actual, "acctest value=101")

	testm = <-metrics
	actual = testm.String()
	assert.Contains(t, actual, "acctest,acc=test value=101")

	testm = <-metrics
	actual = testm.String()
	assert.Equal(t,
		fmt.Sprintf("acctest,acc=test value=101 %d\n", now.UnixNano()),
		actual)
}

func TestAddFields(t *testing.T) {
	now := time.Now()
	metrics := make(chan telegraf.Metric, 10)
	defer close(metrics)
	a := NewAccumulator(&TestMetricMaker{}, metrics)

	fields := map[string]interface{}{
		"usage": float64(99),
	}
	a.AddFields("acctest", fields, map[string]string{})
	a.AddGauge("acctest", fields, map[string]string{"acc": "test"})
	a.AddCounter("acctest", fields, map[string]string{"acc": "test"}, now)

	testm := <-metrics
	actual := testm.String()
	assert.Contains(t, actual, "acctest usage=99")

	testm = <-metrics
	actual = testm.String()
	assert.Contains(t, actual, "acctest,acc=test usage=99")

	testm = <-metrics
	actual = testm.String()
	assert.Equal(t,
		fmt.Sprintf("acctest,acc=test usage=99 %d\n", now.UnixNano()),
		actual)
}

func TestAccAddError(t *testing.T) {
	errBuf := bytes.NewBuffer(nil)
	log.SetOutput(errBuf)
	defer log.SetOutput(os.Stderr)

	metrics := make(chan telegraf.Metric, 10)
	defer close(metrics)
	a := NewAccumulator(&TestMetricMaker{}, metrics)

	a.AddError(fmt.Errorf("foo"))
	a.AddError(fmt.Errorf("bar"))
	a.AddError(fmt.Errorf("baz"))

	errs := bytes.Split(errBuf.Bytes(), []byte{'\n'})
	assert.EqualValues(t, int64(3), NErrors.Get())
	require.Len(t, errs, 4) // 4 because of trailing newline
	assert.Contains(t, string(errs[0]), "TestPlugin")
	assert.Contains(t, string(errs[0]), "foo")
	assert.Contains(t, string(errs[1]), "TestPlugin")
	assert.Contains(t, string(errs[1]), "bar")
	assert.Contains(t, string(errs[2]), "TestPlugin")
	assert.Contains(t, string(errs[2]), "baz")
}

func TestAddNoIntervalWithPrecision(t *testing.T) {
	now := time.Date(2006, time.February, 10, 12, 0, 0, 82912748, time.UTC)
	metrics := make(chan telegraf.Metric, 10)
	defer close(metrics)
	a := NewAccumulator(&TestMetricMaker{}, metrics)
	a.SetPrecision(0, time.Second)

	a.AddFields("acctest",
		map[string]interface{}{"value": float64(101)},
		map[string]string{})
	a.AddFields("acctest",
		map[string]interface{}{"value": float64(101)},
		map[string]string{"acc": "test"})
	a.AddFields("acctest",
		map[string]interface{}{"value": float64(101)},
		map[string]string{"acc": "test"}, now)

	testm := <-a.metrics
	actual := testm.String()
	assert.Contains(t, actual, "acctest value=101")

	testm = <-a.metrics
	actual = testm.String()
	assert.Contains(t, actual, "acctest,acc=test value=101")

	testm = <-a.metrics
	actual = testm.String()
	assert.Equal(t,
		fmt.Sprintf("acctest,acc=test value=101 %d\n", int64(1139572800000000000)),
		actual)
}

func TestAddDisablePrecision(t *testing.T) {
	now := time.Date(2006, time.February, 10, 12, 0, 0, 82912748, time.UTC)
	metrics := make(chan telegraf.Metric, 10)
	defer close(metrics)
	a := NewAccumulator(&TestMetricMaker{}, metrics)

	a.SetPrecision(time.Nanosecond, 0)
	a.AddFields("acctest",
		map[string]interface{}{"value": float64(101)},
		map[string]string{})
	a.AddFields("acctest",
		map[string]interface{}{"value": float64(101)},
		map[string]string{"acc": "test"})
	a.AddFields("acctest",
		map[string]interface{}{"value": float64(101)},
		map[string]string{"acc": "test"}, now)

	testm := <-a.metrics
	actual := testm.String()
	assert.Contains(t, actual, "acctest value=101")

	testm = <-a.metrics
	actual = testm.String()
	assert.Contains(t, actual, "acctest,acc=test value=101")

	testm = <-a.metrics
	actual = testm.String()
	assert.Equal(t,
		fmt.Sprintf("acctest,acc=test value=101 %d\n", int64(1139572800082912748)),
		actual)
}

func TestAddNoPrecisionWithInterval(t *testing.T) {
	now := time.Date(2006, time.February, 10, 12, 0, 0, 82912748, time.UTC)
	metrics := make(chan telegraf.Metric, 10)
	defer close(metrics)
	a := NewAccumulator(&TestMetricMaker{}, metrics)

	a.SetPrecision(0, time.Second)
	a.AddFields("acctest",
		map[string]interface{}{"value": float64(101)},
		map[string]string{})
	a.AddFields("acctest",
		map[string]interface{}{"value": float64(101)},
		map[string]string{"acc": "test"})
	a.AddFields("acctest",
		map[string]interface{}{"value": float64(101)},
		map[string]string{"acc": "test"}, now)

	testm := <-a.metrics
	actual := testm.String()
	assert.Contains(t, actual, "acctest value=101")

	testm = <-a.metrics
	actual = testm.String()
	assert.Contains(t, actual, "acctest,acc=test value=101")

	testm = <-a.metrics
	actual = testm.String()
	assert.Equal(t,
		fmt.Sprintf("acctest,acc=test value=101 %d\n", int64(1139572800000000000)),
		actual)
}

func TestDifferentPrecisions(t *testing.T) {
	now := time.Date(2006, time.February, 10, 12, 0, 0, 82912748, time.UTC)
	metrics := make(chan telegraf.Metric, 10)
	defer close(metrics)
	a := NewAccumulator(&TestMetricMaker{}, metrics)

	a.SetPrecision(0, time.Second)
	a.AddFields("acctest",
		map[string]interface{}{"value": float64(101)},
		map[string]string{"acc": "test"}, now)
	testm := <-a.metrics
	actual := testm.String()
	assert.Equal(t,
		fmt.Sprintf("acctest,acc=test value=101 %d\n", int64(1139572800000000000)),
		actual)

	a.SetPrecision(0, time.Millisecond)
	a.AddFields("acctest",
		map[string]interface{}{"value": float64(101)},
		map[string]string{"acc": "test"}, now)
	testm = <-a.metrics
	actual = testm.String()
	assert.Equal(t,
		fmt.Sprintf("acctest,acc=test value=101 %d\n", int64(1139572800083000000)),
		actual)

	a.SetPrecision(0, time.Microsecond)
	a.AddFields("acctest",
		map[string]interface{}{"value": float64(101)},
		map[string]string{"acc": "test"}, now)
	testm = <-a.metrics
	actual = testm.String()
	assert.Equal(t,
		fmt.Sprintf("acctest,acc=test value=101 %d\n", int64(1139572800082913000)),
		actual)

	a.SetPrecision(0, time.Nanosecond)
	a.AddFields("acctest",
		map[string]interface{}{"value": float64(101)},
		map[string]string{"acc": "test"}, now)
	testm = <-a.metrics
	actual = testm.String()
	assert.Equal(t,
		fmt.Sprintf("acctest,acc=test value=101 %d\n", int64(1139572800082912748)),
		actual)
}

func TestAddGauge(t *testing.T) {
	now := time.Now()
	metrics := make(chan telegraf.Metric, 10)
	defer close(metrics)
	a := NewAccumulator(&TestMetricMaker{}, metrics)

	a.AddGauge("acctest",
		map[string]interface{}{"value": float64(101)},
		map[string]string{})
	a.AddGauge("acctest",
		map[string]interface{}{"value": float64(101)},
		map[string]string{"acc": "test"})
	a.AddGauge("acctest",
		map[string]interface{}{"value": float64(101)},
		map[string]string{"acc": "test"}, now)

	testm := <-metrics
	actual := testm.String()
	assert.Contains(t, actual, "acctest value=101")
	assert.Equal(t, testm.Type(), telegraf.Gauge)

	testm = <-metrics
	actual = testm.String()
	assert.Contains(t, actual, "acctest,acc=test value=101")
	assert.Equal(t, testm.Type(), telegraf.Gauge)

	testm = <-metrics
	actual = testm.String()
	assert.Equal(t,
		fmt.Sprintf("acctest,acc=test value=101 %d\n", now.UnixNano()),
		actual)
	assert.Equal(t, testm.Type(), telegraf.Gauge)
}

func TestAddCounter(t *testing.T) {
	now := time.Now()
	metrics := make(chan telegraf.Metric, 10)
	defer close(metrics)
	a := NewAccumulator(&TestMetricMaker{}, metrics)

	a.AddCounter("acctest",
		map[string]interface{}{"value": float64(101)},
		map[string]string{})
	a.AddCounter("acctest",
		map[string]interface{}{"value": float64(101)},
		map[string]string{"acc": "test"})
	a.AddCounter("acctest",
		map[string]interface{}{"value": float64(101)},
		map[string]string{"acc": "test"}, now)

	testm := <-metrics
	actual := testm.String()
	assert.Contains(t, actual, "acctest value=101")
	assert.Equal(t, testm.Type(), telegraf.Counter)

	testm = <-metrics
	actual = testm.String()
	assert.Contains(t, actual, "acctest,acc=test value=101")
	assert.Equal(t, testm.Type(), telegraf.Counter)

	testm = <-metrics
	actual = testm.String()
	assert.Equal(t,
		fmt.Sprintf("acctest,acc=test value=101 %d\n", now.UnixNano()),
		actual)
	assert.Equal(t, testm.Type(), telegraf.Counter)
}

type TestMetricMaker struct {
}

func (tm *TestMetricMaker) Name() string {
	return "TestPlugin"
}
func (tm *TestMetricMaker) MakeMetric(
	measurement string,
	fields map[string]interface{},
	tags map[string]string,
	mType telegraf.ValueType,
	t time.Time,
) telegraf.Metric {
	switch mType {
	case telegraf.Untyped:
		if m, err := metric.New(measurement, tags, fields, t); err == nil {
			return m
		}
	case telegraf.Counter:
		if m, err := metric.New(measurement, tags, fields, t, telegraf.Counter); err == nil {
			return m
		}
	case telegraf.Gauge:
		if m, err := metric.New(measurement, tags, fields, t, telegraf.Gauge); err == nil {
			return m
		}
	}
	return nil
}
