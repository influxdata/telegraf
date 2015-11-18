package jolokia

import (
	_ "fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/influxdb/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	_ "github.com/stretchr/testify/require"
)

const validMultiValueJSON = `
{
  "request":{
    "mbean":"java.lang:type=Memory",
    "attribute":"HeapMemoryUsage",
    "type":"read"
  },
  "value":{
    "init":67108864,
    "committed":456130560,
    "max":477626368,
    "used":203288528
  },
  "timestamp":1446129191,
  "status":200
}`

const validSingleValueJSON = `
{
  "request":{
    "path":"used",
    "mbean":"java.lang:type=Memory",
    "attribute":"HeapMemoryUsage",
    "type":"read"
  },
  "value":209274376,
  "timestamp":1446129256,
  "status":200
}`

const invalidJSON = "I don't think this is JSON"

const empty = ""

var Servers = []Server{Server{Name: "as1", Host: "127.0.0.1", Port: "8080"}}
var HeapMetric = Metric{Name: "heap_memory_usage", Jmx: "/java.lang:type=Memory/HeapMemoryUsage"}
var UsedHeapMetric = Metric{Name: "heap_memory_usage", Jmx: "/java.lang:type=Memory/HeapMemoryUsage", Pass: []string{"used"}}

type jolokiaClientStub struct {
	responseBody string
	statusCode   int
}

func (c jolokiaClientStub) MakeRequest(req *http.Request) (*http.Response, error) {
	resp := http.Response{}
	resp.StatusCode = c.statusCode
	resp.Body = ioutil.NopCloser(strings.NewReader(c.responseBody))
	return &resp, nil
}

// Generates a pointer to an HttpJson object that uses a mock HTTP client.
// Parameters:
//     response  : Body of the response that the mock HTTP client should return
//     statusCode: HTTP status code the mock HTTP client should return
//
// Returns:
//     *HttpJson: Pointer to an HttpJson object that uses the generated mock HTTP client
func genJolokiaClientStub(response string, statusCode int, servers []Server, metrics []Metric) *Jolokia {
	return &Jolokia{
		jClient: jolokiaClientStub{responseBody: response, statusCode: statusCode},
		Servers: servers,
		Metrics: metrics,
	}
}

// Test that the proper values are ignored or collected
func TestHttpJsonMultiValue(t *testing.T) {

	jolokia := genJolokiaClientStub(validMultiValueJSON, 200, Servers, []Metric{HeapMetric})

	var acc testutil.Accumulator
	err := jolokia.Gather(&acc)

	assert.Nil(t, err)
	assert.Equal(t, 1, len(acc.Points))

	assert.True(t, acc.CheckFieldsValue("heap_memory_usage", map[string]interface{}{"init": 67108864.0,
		"committed": 456130560.0,
		"max":       477626368.0,
		"used":      203288528.0}))
}

// Test that the proper values are ignored or collected
func TestHttpJsonMultiValueWithPass(t *testing.T) {

	jolokia := genJolokiaClientStub(validMultiValueJSON, 200, Servers, []Metric{UsedHeapMetric})

	var acc testutil.Accumulator
	err := jolokia.Gather(&acc)

	assert.Nil(t, err)
	assert.Equal(t, 1, len(acc.Points))

	assert.True(t, acc.CheckFieldsValue("heap_memory_usage", map[string]interface{}{"used": 203288528.0}))
}

// Test that the proper values are ignored or collected
func TestHttpJsonMultiValueTags(t *testing.T) {

	jolokia := genJolokiaClientStub(validMultiValueJSON, 200, Servers, []Metric{UsedHeapMetric})

	var acc testutil.Accumulator
	err := jolokia.Gather(&acc)

	assert.Nil(t, err)
	assert.Equal(t, 1, len(acc.Points))
	assert.NoError(t, acc.ValidateTaggedFieldsValue("heap_memory_usage", map[string]interface{}{"used": 203288528.0}, map[string]string{"host": "127.0.0.1", "port": "8080", "server": "as1"}))
}

// Test that the proper values are ignored or collected
func TestHttpJsonSingleValueTags(t *testing.T) {

	jolokia := genJolokiaClientStub(validSingleValueJSON, 200, Servers, []Metric{UsedHeapMetric})

	var acc testutil.Accumulator
	err := jolokia.Gather(&acc)

	assert.Nil(t, err)
	assert.Equal(t, 1, len(acc.Points))
	assert.NoError(t, acc.ValidateTaggedFieldsValue("heap_memory_usage", map[string]interface{}{"value": 209274376.0}, map[string]string{"host": "127.0.0.1", "port": "8080", "server": "as1"}))
}

// Test that the proper values are ignored or collected
func TestHttpJsonOn404(t *testing.T) {

	jolokia := genJolokiaClientStub(validMultiValueJSON, 404, Servers, []Metric{UsedHeapMetric})

	var acc testutil.Accumulator
	err := jolokia.Gather(&acc)

	assert.Nil(t, err)
	assert.Equal(t, 0, len(acc.Points))
}
