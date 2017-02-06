package http

import (
	ejon "encoding/json"
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/serializers/graphite"
	"github.com/influxdata/telegraf/plugins/serializers/influx"
	"github.com/influxdata/telegraf/plugins/serializers/json"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"sync"
	"testing"
	"time"
)

var (
	tags = map[string]string{
		"host":       "localhost",
		"cpu":        "cpu0",
		"datacenter": "us-west-2",
	}

	fields = map[string]interface{}{
		"usage_idle": float64(91.5),
	}
)

func TestHttpWriteWithoutRequiredOption(t *testing.T) {
	m, _ := metric.New("cpu", tags, fields, time.Now())
	metrics := []telegraf.Metric{m}

	http := &Http{}

	http.SetSerializer(&graphite.GraphiteSerializer{
		Prefix:   "telegraf",
		Template: "tags.measurement.field",
	})

	http.Connect()

	err := http.Write(metrics)

	assert.Error(t, err)
}

func TestHttpWriteNormalCase(t *testing.T) {
	now := time.Now()
	HTTPServer(t, now, 9880)

	m, _ := metric.New("cpu", tags, fields, now)
	metrics := []telegraf.Metric{m}

	http := &Http{
		URL:            "http://127.0.0.1:9880/metric",
		HttpHeaders:    []string{"Content-Type:application/json"},
		expStatusCodes: []int{200, 204},
		BufLimit:       1,
	}

	http.SetSerializer(&graphite.GraphiteSerializer{
		Prefix:   "telegraf",
		Template: "tags.measurement.field",
	})

	http.Connect()
	err := http.Write(metrics)

	assert.NoError(t, err)
}

func TestHttpWriteWithUnexpected404StatusCode(t *testing.T) {
	now := time.Now()

	m, _ := metric.New("cpu", tags, fields, now)
	metrics := []telegraf.Metric{m}

	http := &Http{
		URL:            "http://127.0.0.1:9880/incorrect/url",
		HttpHeaders:    []string{"Content-Type:application/json"},
		expStatusCodes: []int{200},
		BufLimit:       1,
	}

	http.SetSerializer(&graphite.GraphiteSerializer{
		Prefix:   "telegraf",
		Template: "tags.measurement.field",
	})

	http.Connect()
	err := http.Write(metrics)

	assert.Error(t, err)
}

func TestHttpWriteWithExpected404StatusCode(t *testing.T) {
	now := time.Now()

	m, _ := metric.New("cpu", tags, fields, now)
	metrics := []telegraf.Metric{m}

	http := &Http{
		URL:            "http://127.0.0.1:9880/incorrect/url",
		HttpHeaders:    []string{"Content-Type:application/json"},
		expStatusCodes: []int{200, 404},
		BufLimit:       1,
	}

	http.SetSerializer(&graphite.GraphiteSerializer{
		Prefix:   "telegraf",
		Template: "tags.measurement.field",
	})

	http.Connect()
	err := http.Write(metrics)

	assert.NoError(t, err)
}

func TestHttpWriteWithIncorrectServerPort(t *testing.T) {
	now := time.Now()

	m, _ := metric.New("cpu", tags, fields, now)
	metrics := []telegraf.Metric{m}

	http := &Http{
		URL:            "http://127.0.0.1:56879/incorrect/url",
		HttpHeaders:    []string{"Content-Type:application/json"},
		expStatusCodes: []int{200},
	}

	http.SetSerializer(&graphite.GraphiteSerializer{
		Prefix:   "telegraf",
		Template: "tags.measurement.field",
	})

	http.Connect()
	err := http.Write(metrics)

	assert.Error(t, err)
}

func TestMakeReqBody(t *testing.T) {
	// given
	m, _ := metric.New("cpu", tags, fields, time.Now())

	var reqBodyBuf [][]byte

	jsonSerializer := json.JsonSerializer{}
	serializedMetric, _ := jsonSerializer.Serialize(m)
	reqBodyBuf = append(reqBodyBuf, serializedMetric)

	// when
	reqBody, err := makeReqBody(&jsonSerializer, reqBodyBuf)

	// then
	assert.NoError(t, err)
	assert.NotEmpty(t, reqBody)
}

func TestMakeReqBody2(t *testing.T) {
	// given
	m, _ := metric.New("cpu", tags, fields, time.Now())

	var reqBodyBuf [][]byte

	influxSerializer := influx.InfluxSerializer{}
	serializedMetric, _ := influxSerializer.Serialize(m)
	reqBodyBuf = append(reqBodyBuf, serializedMetric)

	// when
	reqBody, err := makeReqBody(&influxSerializer, reqBodyBuf)

	// then
	assert.NoError(t, err)
	assert.NotEmpty(t, reqBody)
}

func TestMakeJsonFormatReqBody(t *testing.T) {
	// given
	m, _ := metric.New("cpu", tags, fields, time.Now())

	var reqBodyBuf [][]byte

	jsonSerializer := json.JsonSerializer{}
	serializedMetric, _ := jsonSerializer.Serialize(m)
	reqBodyBuf = append(reqBodyBuf, serializedMetric)

	serializedMetric2, _ := jsonSerializer.Serialize(m)
	reqBodyBuf = append(reqBodyBuf, serializedMetric2)

	// when
	result, _ := makeJsonFormatReqBody(reqBodyBuf)

	// then
	var reqBodyArrayJsonO []map[string]interface{}
	ejon.Unmarshal(result, &reqBodyArrayJsonO)

	assert.Equal(t, 2, len(reqBodyArrayJsonO))
}

func TestMakeJsonFormatReqBodyWithNotJsonFormat(t *testing.T) {
	// given
	m, _ := metric.New("cpu", tags, fields, time.Now())

	var reqBodyBuf [][]byte

	influxSerializer := influx.InfluxSerializer{}
	serializedMetric, _ := influxSerializer.Serialize(m)
	reqBodyBuf = append(reqBodyBuf, serializedMetric)

	serializedMetric2, _ := influxSerializer.Serialize(m)
	reqBodyBuf = append(reqBodyBuf, serializedMetric2)

	// when
	result, err := makeJsonFormatReqBody(reqBodyBuf)

	// then
	assert.Equal(t, 0, len(result))
	assert.Error(t, err)
}

func HTTPServer(t *testing.T, now time.Time, port int) {
	http.HandleFunc("/metric", func(w http.ResponseWriter, r *http.Request) {
		body, _ := ioutil.ReadAll(r.Body)

		assert.Equal(t, fmt.Sprintf("telegraf.cpu0.us-west-2.localhost.cpu.usage_idle 91.5 %d\n", now.Unix()), string(body))

		fmt.Fprintf(w, "ok")
	})

	var wg sync.WaitGroup
	wg.Add(1)

	go http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
