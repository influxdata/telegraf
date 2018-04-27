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

func TestAddFields(t *testing.T) {
	metrics := make(chan telegraf.Metric, 10)
	defer close(metrics)
	a := NewAccumulator(&TestMetricMaker{}, metrics)

	tags := map[string]string{"foo": "bar"}
	fields := map[string]interface{}{
		"usage": float64(99),
	}
	now := time.Now()
	a.AddCounter("acctest", fields, tags, now)

	testm := <-metrics

	require.Equal(t, "acctest", testm.Name())
	actual, ok := testm.GetField("usage")

	require.True(t, ok)
	require.Equal(t, float64(99), actual)

	actual, ok = testm.GetTag("foo")
	require.True(t, ok)
	require.Equal(t, "bar", actual)

	tm := testm.Time()
	// okay if monotonic clock differs
	require.True(t, now.Equal(tm))

	tp := testm.Type()
	require.Equal(t, telegraf.Counter, tp)
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

func TestSetPrecision(t *testing.T) {
	tests := []struct {
		name      string
		unset     bool
		precision time.Duration
		interval  time.Duration
		timestamp time.Time
		expected  time.Time
	}{
		{
			name:      "default precision is nanosecond",
			unset:     true,
			timestamp: time.Date(2006, time.February, 10, 12, 0, 0, 82912748, time.UTC),
			expected:  time.Date(2006, time.February, 10, 12, 0, 0, 82912748, time.UTC),
		},
		{
			name:      "second interval",
			interval:  time.Second,
			timestamp: time.Date(2006, time.February, 10, 12, 0, 0, 82912748, time.UTC),
			expected:  time.Date(2006, time.February, 10, 12, 0, 0, 0, time.UTC),
		},
		{
			name:      "microsecond interval",
			interval:  time.Microsecond,
			timestamp: time.Date(2006, time.February, 10, 12, 0, 0, 82912748, time.UTC),
			expected:  time.Date(2006, time.February, 10, 12, 0, 0, 82913000, time.UTC),
		},
		{
			name:      "2 second precision",
			precision: 2 * time.Second,
			timestamp: time.Date(2006, time.February, 10, 12, 0, 2, 4, time.UTC),
			expected:  time.Date(2006, time.February, 10, 12, 0, 2, 0, time.UTC),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := make(chan telegraf.Metric, 10)

			a := NewAccumulator(&TestMetricMaker{}, metrics)
			if !tt.unset {
				a.SetPrecision(tt.precision, tt.interval)
			}

			a.AddFields("acctest",
				map[string]interface{}{"value": float64(101)},
				map[string]string{},
				tt.timestamp,
			)

			testm := <-metrics
			require.Equal(t, tt.expected, testm.Time())

			close(metrics)
		})
	}
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
