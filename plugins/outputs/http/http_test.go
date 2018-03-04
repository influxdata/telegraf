package http

import (
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/serializers/graphite"
	"github.com/influxdata/telegraf/plugins/serializers/json"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

var (
	cpuTags = map[string]string{
		"host":       "localhost",
		"cpu":        "cpu0",
		"datacenter": "us-west-2",
	}

	cpuField = map[string]interface{}{
		"usage_idle": float64(91.5),
	}

	memTags = map[string]string{
		"host":       "localhost",
		"cpu":        "mem",
		"datacenter": "us-west-2",
	}

	memField = map[string]interface{}{
		"used": float64(91.5),
	}

	count int
)

type TestOkHandler struct {
	T        *testing.T
	Expected []string
}

// The handler gets a new variable each time it receives a request, so it fetches an expected string based on global variable.
func (h TestOkHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	actual, _ := ioutil.ReadAll(r.Body)

	assert.Equal(h.T, h.Expected[count], string(actual), fmt.Sprintf("%d Expected fail!", count))

	count++

	fmt.Fprint(w, "ok")
}

type TestNotFoundHandler struct {
}

func (h TestNotFoundHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	http.NotFound(w, r)
}

func TestWriteAllInputMetric(t *testing.T) {
	now := time.Now()

	server := httptest.NewServer(&TestOkHandler{
		T: t,
		Expected: []string{
			fmt.Sprintf("telegraf.cpu0.us-west-2.localhost.cpu.usage_idle 91.5 %d\ntelegraf.mem.us-west-2.localhost.mem.used 91.5 %d\n", now.Unix(), now.Unix()),
		},
	})
	defer server.Close()
	defer resetCount()

	m1, _ := metric.New("cpu", cpuTags, cpuField, now)
	m2, _ := metric.New("mem", memTags, memField, now)
	metrics := []telegraf.Metric{m1, m2}

	http := &Http{
		URL:                server.URL,
		HttpHeaders:        []string{"Content-Type:plain/text"},
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

	server := httptest.NewServer(&TestNotFoundHandler{})
	defer server.Close()

	m, _ := metric.New("cpu", cpuTags, cpuField, now)
	metrics := []telegraf.Metric{m}

	http := &Http{
		URL:                server.URL,
		HttpHeaders:        []string{"Content-Type:application/json"},
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

	server := httptest.NewServer(&TestNotFoundHandler{})
	defer server.Close()

	m, _ := metric.New("cpu", cpuTags, cpuField, now)
	metrics := []telegraf.Metric{m}

	http := &Http{
		URL:                server.URL,
		HttpHeaders:        []string{"Content-Type:application/json"},
	}

	http.SetSerializer(&graphite.GraphiteSerializer{
		Prefix:   "telegraf",
		Template: "tags.measurement.field",
	})

	http.Connect()
	err := http.Write(metrics)

	assert.Error(t, err)
}

func TestHttpWriteWithIncorrectServerPort(t *testing.T) {
	now := time.Now()

	m, _ := metric.New("cpu", cpuTags, cpuField, now)
	metrics := []telegraf.Metric{m}

	http := &Http{
		URL:                "http://127.0.0.1:56879/incorrect/url",
		HttpHeaders:        []string{"Content-Type:application/json"},
	}

	http.SetSerializer(&graphite.GraphiteSerializer{
		Prefix:   "telegraf",
		Template: "tags.measurement.field",
	})

	http.Connect()
	err := http.Write(metrics)

	assert.Error(t, err)
}

func TestMakeReqBody_IfSerializerIsJsonSerializerAndJsonContentType(t *testing.T) {
	// given
	now := time.Now()
	m, _ := metric.New("cpu", cpuTags, cpuField, now)

	var reqBodyBuf []byte

	jsonSerializer := json.JsonSerializer{}
	serializedMetric, _ := jsonSerializer.Serialize(m)
	reqBodyBuf = append(reqBodyBuf, serializedMetric...)

	// when
	reqBody, err := makeReqBody("application/json", reqBodyBuf, 1)

	// then
	assert.Equal(t, fmt.Sprintf("[{\"fields\":{\"usage_idle\":91.5},\"name\":\"cpu\",\"tags\":{\"cpu\":\"cpu0\",\"datacenter\":\"us-west-2\",\"host\":\"localhost\"},\"timestamp\":%d}\n]", now.Unix()), string(reqBody))
	assert.NoError(t, err)
}

func TestMakeReqBody_IfContentTypeIsInvalid(t *testing.T) {
	_, err := makeReqBody("plain/tex", nil, 1)

	// then
	assert.Equal(t, "E! HTTP plain/tex content-type is not supported!", err.Error())
}

func TestMakeReqBody_IfSerializerIsGraphiteSerializer(t *testing.T) {
	// given
	now := time.Now()
	m, _ := metric.New("cpu", cpuTags, cpuField, now)

	var reqBodyBuf []byte

	graphiteSerializer := graphite.GraphiteSerializer{}
	serializedMetric1, _ := graphiteSerializer.Serialize(m)
	serializedMetric2, _ := graphiteSerializer.Serialize(m)
	reqBodyBuf = append(reqBodyBuf, serializedMetric1...)
	reqBodyBuf = append(reqBodyBuf, serializedMetric2...)

	// when
	reqBody, err := makeReqBody("plain/text", reqBodyBuf, 1)

	// then
	assert.Equal(t, fmt.Sprintf("localhost.cpu0.us-west-2.cpu.usage_idle 91.5 %d\nlocalhost.cpu0.us-west-2.cpu.usage_idle 91.5 %d\n", now.Unix(), now.Unix()), string(reqBody))
	assert.NoError(t, err)
}

func TestImplementedInterfaceFunction(t *testing.T) {
	http := &Http{
		URL:         "http://127.0.0.1:56879/incorrect/url",
		HttpHeaders: []string{"Content-Type:application/json"},
	}

	assert.NotNil(t, http.SampleConfig())
	assert.NotNil(t, http.Description())
}

func resetCount() {
	count = 0
}
