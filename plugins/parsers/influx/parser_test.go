package influx

import (
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	"github.com/influxdata/telegraf"

	"github.com/stretchr/testify/assert"
)

var (
	ms         []telegraf.Metric
	writer     = ioutil.Discard
	metrics500 []byte
	exptime    = time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC).UnixNano()
)

const (
	validInflux          = "cpu_load_short,cpu=cpu0 value=10 1257894000000000000\n"
	negativeFloat        = "cpu_load_short,cpu=cpu0 value=-13.4 1257894000000000000\n"
	validInfluxNewline   = "\ncpu_load_short,cpu=cpu0 value=10 1257894000000000000\n"
	validInfluxNoNewline = "cpu_load_short,cpu=cpu0 value=10 1257894000000000000"
	invalidInflux        = "I don't think this is line protocol\n"
	invalidInflux2       = "{\"a\": 5, \"b\": {\"c\": 6}}\n"
	invalidInflux3       = `name text="unescaped "quote" ",value=1 1498077493081000000`
	invalidInflux4       = `name text="unbalanced "quote" 1498077493081000000`
)

const influxMulti = `
cpu,host=foo,datacenter=us-east usage_idle=99,usage_busy=1
cpu,host=foo,datacenter=us-east usage_idle=99,usage_busy=1
cpu,host=foo,datacenter=us-east usage_idle=99,usage_busy=1
cpu,host=foo,datacenter=us-east usage_idle=99,usage_busy=1
cpu,host=foo,datacenter=us-east usage_idle=99,usage_busy=1
cpu,host=foo,datacenter=us-east usage_idle=99,usage_busy=1
cpu,host=foo,datacenter=us-east usage_idle=99,usage_busy=1
`

const influxMultiSomeInvalid = `
cpu,host=foo,datacenter=us-east usage_idle=99,usage_busy=1
cpu,host=foo,datacenter=us-east usage_idle=99,usage_busy=1
cpu,host=foo,datacenter=us-east usage_idle=99,usage_busy=1
cpu,cpu=cpu3, host=foo,datacenter=us-east usage_idle=99,usage_busy=1
cpu,cpu=cpu4 , usage_idle=99,usage_busy=1
cpu,host=foo,datacenter=us-east usage_idle=99,usage_busy=1
`

func TestParseValidInflux(t *testing.T) {
	parser := InfluxParser{}

	metrics, err := parser.Parse([]byte(validInflux))
	assert.NoError(t, err)
	assert.Len(t, metrics, 1)
	assert.Equal(t, "cpu_load_short", metrics[0].Name())
	assert.Equal(t, map[string]interface{}{
		"value": float64(10),
	}, metrics[0].Fields())
	assert.Equal(t, map[string]string{
		"cpu": "cpu0",
	}, metrics[0].Tags())
	assert.Equal(t, exptime, metrics[0].Time().UnixNano())

	metrics, err = parser.Parse([]byte(validInfluxNewline))
	assert.NoError(t, err)
	assert.Len(t, metrics, 1)
	assert.Equal(t, "cpu_load_short", metrics[0].Name())
	assert.Equal(t, map[string]interface{}{
		"value": float64(10),
	}, metrics[0].Fields())
	assert.Equal(t, map[string]string{
		"cpu": "cpu0",
	}, metrics[0].Tags())
	assert.Equal(t, exptime, metrics[0].Time().UnixNano())

	metrics, err = parser.Parse([]byte(validInfluxNoNewline))
	assert.NoError(t, err)
	assert.Len(t, metrics, 1)
	assert.Equal(t, "cpu_load_short", metrics[0].Name())
	assert.Equal(t, map[string]interface{}{
		"value": float64(10),
	}, metrics[0].Fields())
	assert.Equal(t, map[string]string{
		"cpu": "cpu0",
	}, metrics[0].Tags())
	assert.Equal(t, exptime, metrics[0].Time().UnixNano())

	metrics, err = parser.Parse([]byte(negativeFloat))
	assert.NoError(t, err)
	assert.Len(t, metrics, 1)
	assert.Equal(t, "cpu_load_short", metrics[0].Name())
	assert.Equal(t, map[string]interface{}{
		"value": float64(-13.4),
	}, metrics[0].Fields())
	assert.Equal(t, map[string]string{
		"cpu": "cpu0",
	}, metrics[0].Tags())
	assert.Equal(t, exptime, metrics[0].Time().UnixNano())
}

func TestParseLineValidInflux(t *testing.T) {
	parser := InfluxParser{}

	metric, err := parser.ParseLine(validInflux)
	assert.NoError(t, err)
	assert.Equal(t, "cpu_load_short", metric.Name())
	assert.Equal(t, map[string]interface{}{
		"value": float64(10),
	}, metric.Fields())
	assert.Equal(t, map[string]string{
		"cpu": "cpu0",
	}, metric.Tags())
	assert.Equal(t, exptime, metric.Time().UnixNano())

	metric, err = parser.ParseLine(validInfluxNewline)
	assert.NoError(t, err)
	assert.Equal(t, "cpu_load_short", metric.Name())
	assert.Equal(t, map[string]interface{}{
		"value": float64(10),
	}, metric.Fields())
	assert.Equal(t, map[string]string{
		"cpu": "cpu0",
	}, metric.Tags())
	assert.Equal(t, exptime, metric.Time().UnixNano())
}

func TestParseMultipleValid(t *testing.T) {
	parser := InfluxParser{}

	metrics, err := parser.Parse([]byte(influxMulti))
	assert.NoError(t, err)
	assert.Len(t, metrics, 7)

	for _, metric := range metrics {
		assert.Equal(t, "cpu", metric.Name())
		assert.Equal(t, map[string]string{
			"datacenter": "us-east",
			"host":       "foo",
		}, metrics[0].Tags())
		assert.Equal(t, map[string]interface{}{
			"usage_idle": float64(99),
			"usage_busy": float64(1),
		}, metrics[0].Fields())
	}
}

func TestParseSomeValid(t *testing.T) {
	parser := InfluxParser{}

	metrics, err := parser.Parse([]byte(influxMultiSomeInvalid))
	assert.Error(t, err)
	assert.Len(t, metrics, 4)

	for _, metric := range metrics {
		assert.Equal(t, "cpu", metric.Name())
		assert.Equal(t, map[string]string{
			"datacenter": "us-east",
			"host":       "foo",
		}, metrics[0].Tags())
		assert.Equal(t, map[string]interface{}{
			"usage_idle": float64(99),
			"usage_busy": float64(1),
		}, metrics[0].Fields())
	}
}

// Test that default tags are applied.
func TestParseDefaultTags(t *testing.T) {
	parser := InfluxParser{
		DefaultTags: map[string]string{
			"tag": "default",
		},
	}

	metrics, err := parser.Parse([]byte(influxMultiSomeInvalid))
	assert.Error(t, err)
	assert.Len(t, metrics, 4)

	for _, metric := range metrics {
		assert.Equal(t, "cpu", metric.Name())
		assert.Equal(t, map[string]string{
			"datacenter": "us-east",
			"host":       "foo",
			"tag":        "default",
		}, metric.Tags())
		assert.Equal(t, map[string]interface{}{
			"usage_idle": float64(99),
			"usage_busy": float64(1),
		}, metric.Fields())
	}
}

// Verify that metric tags will override default tags
func TestParseDefaultTagsOverride(t *testing.T) {
	parser := InfluxParser{
		DefaultTags: map[string]string{
			"host": "default",
		},
	}

	metrics, err := parser.Parse([]byte(influxMultiSomeInvalid))
	assert.Error(t, err)
	assert.Len(t, metrics, 4)

	for _, metric := range metrics {
		assert.Equal(t, "cpu", metric.Name())
		assert.Equal(t, map[string]string{
			"datacenter": "us-east",
			"host":       "foo",
		}, metrics[0].Tags())
		assert.Equal(t, map[string]interface{}{
			"usage_idle": float64(99),
			"usage_busy": float64(1),
		}, metrics[0].Fields())
	}
}

func TestParseInvalidInflux(t *testing.T) {
	parser := InfluxParser{}

	_, err := parser.Parse([]byte(invalidInflux))
	assert.Error(t, err)
	_, err = parser.Parse([]byte(invalidInflux2))
	assert.Error(t, err)
	_, err = parser.Parse([]byte(invalidInflux3))
	assert.Error(t, err)
	fmt.Printf("%+v\n", err) // output for debug
	_, err = parser.Parse([]byte(invalidInflux4))
	assert.Error(t, err)

	_, err = parser.ParseLine(invalidInflux)
	assert.Error(t, err)
	_, err = parser.ParseLine(invalidInflux2)
	assert.Error(t, err)
	_, err = parser.ParseLine(invalidInflux3)
	assert.Error(t, err)
	_, err = parser.ParseLine(invalidInflux4)
	assert.Error(t, err)

}

func BenchmarkParse(b *testing.B) {
	var err error
	parser := InfluxParser{}
	for n := 0; n < b.N; n++ {
		// parse:
		ms, err = parser.Parse(metrics500)
		if err != nil {
			panic(err)
		}
		if len(ms) != 500 {
			panic("500 metrics not parsed!!")
		}
	}
}

func BenchmarkParseAddTagWrite(b *testing.B) {
	var err error
	parser := InfluxParser{}
	for n := 0; n < b.N; n++ {
		ms, err = parser.Parse(metrics500)
		if err != nil {
			panic(err)
		}
		if len(ms) != 500 {
			panic("500 metrics not parsed!!")
		}
		for _, tmp := range ms {
			tmp.AddTag("host", "localhost")
			writer.Write(tmp.Serialize())
		}
	}
}

func init() {
	var err error
	metrics500, err = ioutil.ReadFile("500.metrics")
	if err != nil {
		panic(err)
	}
}
