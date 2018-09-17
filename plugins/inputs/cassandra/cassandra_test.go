package cassandra

import (
	_ "fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	_ "github.com/stretchr/testify/require"
)

const validJavaMultiValueJSON = `
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

const validCassandraMultiValueJSON = `
{
	"request": {
		"mbean": "org.apache.cassandra.metrics:keyspace=test_keyspace1,name=ReadLatency,scope=test_table,type=Table",
		"type": "read"},
	"status": 200,
	"timestamp": 1458089229,
	"value": {
		"999thPercentile": 20.0,
		"99thPercentile": 10.0,
		"Count": 400,
		"DurationUnit": "microseconds",
		"Max": 30.0,
		"Mean": null,
		"MeanRate": 3.0,
		"Min": 1.0,
		"RateUnit": "events/second",
		"StdDev": null
	}
}`

const validCassandraNestedMultiValueJSON = `
{
	"request": {
		"mbean": "org.apache.cassandra.metrics:keyspace=test_keyspace1,name=ReadLatency,scope=*,type=Table",
		"type": "read"},
    "status": 200,
    "timestamp": 1458089184,
    "value": {
		"org.apache.cassandra.metrics:keyspace=test_keyspace1,name=ReadLatency,scope=test_table1,type=Table":
		{  	"999thPercentile": 1.0,
			"Count": 100,
			"DurationUnit": "microseconds",
			"OneMinuteRate": 1.0,
			"RateUnit": "events/second",
			"StdDev": null
		},
		"org.apache.cassandra.metrics:keyspace=test_keyspace2,name=ReadLatency,scope=test_table2,type=Table":
		{  	"999thPercentile": 2.0,
			"Count": 200,
			"DurationUnit": "microseconds",
			"OneMinuteRate": 2.0,
			"RateUnit": "events/second",
			"StdDev": null
		}
	}
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

const validJavaMultiTypeJSON = `
{
   "request":{
	   "mbean":"java.lang:name=ConcurrentMarkSweep,type=GarbageCollector",
	   "attribute":"CollectionCount",
	   "type":"read"
   },
   "value":1,
   "timestamp":1459316570,
   "status":200
}`

const invalidJSON = "I don't think this is JSON"

const empty = ""

var Servers = []string{"10.10.10.10:8778"}
var AuthServers = []string{"user:passwd@10.10.10.10:8778"}
var MultipleServers = []string{"10.10.10.10:8778", "10.10.10.11:8778"}
var HeapMetric = "/java.lang:type=Memory/HeapMemoryUsage"
var ReadLatencyMetric = "/org.apache.cassandra.metrics:type=Table,keyspace=test_keyspace1,scope=test_table,name=ReadLatency"
var NestedReadLatencyMetric = "/org.apache.cassandra.metrics:type=Table,keyspace=test_keyspace1,scope=*,name=ReadLatency"
var GarbageCollectorMetric1 = "/java.lang:type=GarbageCollector,name=ConcurrentMarkSweep/CollectionCount"
var GarbageCollectorMetric2 = "/java.lang:type=GarbageCollector,name=ConcurrentMarkSweep/CollectionTime"
var Context = "/jolokia/read"

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
func genJolokiaClientStub(response string, statusCode int, servers []string, metrics []string) *Cassandra {
	return &Cassandra{
		jClient: jolokiaClientStub{responseBody: response, statusCode: statusCode},
		Context: Context,
		Servers: servers,
		Metrics: metrics,
	}
}

// Test that the proper values are ignored or collected for class=Java
func TestHttpJsonJavaMultiValue(t *testing.T) {
	cassandra := genJolokiaClientStub(validJavaMultiValueJSON, 200,
		MultipleServers, []string{HeapMetric})

	var acc testutil.Accumulator
	acc.SetDebug(true)
	err := acc.GatherError(cassandra.Gather)

	assert.Nil(t, err)
	assert.Equal(t, 2, len(acc.Metrics))

	fields := map[string]interface{}{
		"HeapMemoryUsage_init":      67108864.0,
		"HeapMemoryUsage_committed": 456130560.0,
		"HeapMemoryUsage_max":       477626368.0,
		"HeapMemoryUsage_used":      203288528.0,
	}
	tags1 := map[string]string{
		"cassandra_host": "10.10.10.10",
		"mname":          "HeapMemoryUsage",
	}

	tags2 := map[string]string{
		"cassandra_host": "10.10.10.11",
		"mname":          "HeapMemoryUsage",
	}
	acc.AssertContainsTaggedFields(t, "javaMemory", fields, tags1)
	acc.AssertContainsTaggedFields(t, "javaMemory", fields, tags2)
}

func TestHttpJsonJavaMultiType(t *testing.T) {
	cassandra := genJolokiaClientStub(validJavaMultiTypeJSON, 200, AuthServers, []string{GarbageCollectorMetric1, GarbageCollectorMetric2})

	var acc testutil.Accumulator
	acc.SetDebug(true)
	err := acc.GatherError(cassandra.Gather)

	assert.Nil(t, err)
	assert.Equal(t, 2, len(acc.Metrics))

	fields := map[string]interface{}{
		"CollectionCount": 1.0,
	}

	tags := map[string]string{
		"cassandra_host": "10.10.10.10",
		"mname":          "ConcurrentMarkSweep",
	}
	acc.AssertContainsTaggedFields(t, "javaGarbageCollector", fields, tags)
}

// Test that the proper values are ignored or collected
func TestHttp404(t *testing.T) {

	jolokia := genJolokiaClientStub(invalidJSON, 404, Servers,
		[]string{HeapMetric})

	var acc testutil.Accumulator
	err := acc.GatherError(jolokia.Gather)

	assert.Error(t, err)
	assert.Equal(t, 0, len(acc.Metrics))
	assert.Contains(t, err.Error(), "has status code 404")
}

// Test that the proper values are ignored or collected for class=Cassandra
func TestHttpJsonCassandraMultiValue(t *testing.T) {
	cassandra := genJolokiaClientStub(validCassandraMultiValueJSON, 200, Servers, []string{ReadLatencyMetric})

	var acc testutil.Accumulator
	err := acc.GatherError(cassandra.Gather)

	assert.Nil(t, err)
	assert.Equal(t, 1, len(acc.Metrics))

	fields := map[string]interface{}{
		"ReadLatency_999thPercentile": 20.0,
		"ReadLatency_99thPercentile":  10.0,
		"ReadLatency_Count":           400.0,
		"ReadLatency_DurationUnit":    "microseconds",
		"ReadLatency_Max":             30.0,
		"ReadLatency_MeanRate":        3.0,
		"ReadLatency_Min":             1.0,
		"ReadLatency_RateUnit":        "events/second",
	}

	tags := map[string]string{
		"cassandra_host": "10.10.10.10",
		"mname":          "ReadLatency",
		"keyspace":       "test_keyspace1",
		"scope":          "test_table",
	}
	acc.AssertContainsTaggedFields(t, "cassandraTable", fields, tags)
}

// Test that the proper values are ignored or collected for class=Cassandra with
// nested values
func TestHttpJsonCassandraNestedMultiValue(t *testing.T) {
	cassandra := genJolokiaClientStub(validCassandraNestedMultiValueJSON, 200, Servers, []string{NestedReadLatencyMetric})

	var acc testutil.Accumulator
	acc.SetDebug(true)
	err := acc.GatherError(cassandra.Gather)

	assert.Nil(t, err)
	assert.Equal(t, 2, len(acc.Metrics))

	fields1 := map[string]interface{}{
		"ReadLatency_999thPercentile": 1.0,
		"ReadLatency_Count":           100.0,
		"ReadLatency_DurationUnit":    "microseconds",
		"ReadLatency_OneMinuteRate":   1.0,
		"ReadLatency_RateUnit":        "events/second",
	}

	fields2 := map[string]interface{}{
		"ReadLatency_999thPercentile": 2.0,
		"ReadLatency_Count":           200.0,
		"ReadLatency_DurationUnit":    "microseconds",
		"ReadLatency_OneMinuteRate":   2.0,
		"ReadLatency_RateUnit":        "events/second",
	}

	tags1 := map[string]string{
		"cassandra_host": "10.10.10.10",
		"mname":          "ReadLatency",
		"keyspace":       "test_keyspace1",
		"scope":          "test_table1",
	}

	tags2 := map[string]string{
		"cassandra_host": "10.10.10.10",
		"mname":          "ReadLatency",
		"keyspace":       "test_keyspace2",
		"scope":          "test_table2",
	}

	acc.AssertContainsTaggedFields(t, "cassandraTable", fields1, tags1)
	acc.AssertContainsTaggedFields(t, "cassandraTable", fields2, tags2)
}
