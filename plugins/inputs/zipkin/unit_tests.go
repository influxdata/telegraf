package zipkin

import "time"

// UnitTest represents contains expected test values and a data file to be
// written to the zipkin http server.
type UnitTest struct {
	expected    []TestData
	measurement string
	datafile    string
	waitPoints  int
}

// TestData contains the expected tags and values that the telegraf plugin
// should output
type TestData struct {
	expectedTags   map[string]string
	expectedValues map[string]interface{}
}

// Store all unit tests in an array to allow for iteration over all tests
var tests = []UnitTest{
	UnitTest{
		measurement: "zipkin",
		datafile:    "testdata/threespans.dat",
		expected: []TestData{
			// zipkin data points are stored in InfluxDB tagged partly //annotation specific
			//values, and partly on span specific values,
			// so there are many repeated tags. Fields have very similar tags, which is why
			// tags are relatively redundant in these tests.
			{
				expectedTags: map[string]string{
					"id":               "8090652509916334619",
					"parent_id":        "22964302721410078",
					"trace_id":         "0:2505404965370368069",
					"name":             "Child",
					"service_name":     "trivial",
					"annotation_value": "trivial",
					"endpoint_host":    "2130706433:0",
					"key":              "lc",
					"type":             "STRING",
				},
				expectedValues: map[string]interface{}{
					"duration": time.Duration(53106) * time.Microsecond,
				},
			},
			{
				expectedTags: map[string]string{
					"id":               "103618986556047333",
					"parent_id":        "22964302721410078",
					"trace_id":         "0:2505404965370368069",
					"name":             "Child",
					"service_name":     "trivial",
					"annotation_value": "trivial",
					"endpoint_host":    "2130706433:0",
					"key":              "lc",
					"type":             "STRING",
				},
				expectedValues: map[string]interface{}{
					"duration": time.Duration(50410) * time.Microsecond,
				},
			},
			{
				expectedTags: map[string]string{
					"id":               "22964302721410078",
					"parent_id":        "22964302721410078",
					"trace_id":         "0:2505404965370368069",
					"name":             "Parent",
					"service_name":     "trivial",
					"annotation_value": "Starting child #0",
					"endpoint_host":    "2130706433:0",
				},
				expectedValues: map[string]interface{}{
					"annotation_timestamp": int64(1498688360851325),
				},
			},
			{
				expectedTags: map[string]string{
					"id":               "22964302721410078",
					"parent_id":        "22964302721410078",
					"trace_id":         "0:2505404965370368069",
					"name":             "Parent",
					"service_name":     "trivial",
					"annotation_value": "Starting child #1",
					"endpoint_host":    "2130706433:0",
				},
				expectedValues: map[string]interface{}{
					"annotation_timestamp": int64(1498688360904545),
				},
			},
			{
				expectedTags: map[string]string{
					"id":               "22964302721410078",
					"parent_id":        "22964302721410078",
					"trace_id":         "0:2505404965370368069",
					"name":             "Parent",
					"service_name":     "trivial",
					"annotation_value": "A Log",
					"endpoint_host":    "2130706433:0",
				},
				expectedValues: map[string]interface{}{
					"annotation_timestamp": int64(1498688360954992),
				},
			},
			{
				expectedTags: map[string]string{
					"id":               "22964302721410078",
					"parent_id":        "22964302721410078",
					"trace_id":         "0:2505404965370368069",
					"name":             "Parent",
					"service_name":     "trivial",
					"annotation_value": "trivial",
					"endpoint_host":    "2130706433:0",
					"key":              "lc",
					"type":             "STRING",
				},
				expectedValues: map[string]interface{}{
					"duration": time.Duration(103680) * time.Microsecond,
					"time":     time.Unix(1498688360, 851318*int64(time.Microsecond)),
				},
			},
		},
	},

	// Test data from zipkin cli app:
	//https://github.com/openzipkin/zipkin-go-opentracing/tree/master/examples/cli_with_2_services
	UnitTest{
		measurement: "zipkin",
		datafile:    "testdata/cli_microservice.dat",
		expected: []TestData{
			{
				expectedTags: map[string]string{
					"id":               "3383422996321511664",
					"parent_id":        "4574092882326506380",
					"trace_id":         "0:8269862291023777619243463817635710260",
					"name":             "Concat",
					"service_name":     "cli",
					"annotation_value": "cs",
					"endpoint_host":    "0:0",
				},
				expectedValues: map[string]interface{}{
					"annotation_timestamp": int64(1499817952283903),
				},
			},
		},
	},

	// Test data from distributed trace repo sample json
	// https://github.com/mattkanwisher/distributedtrace/blob/master/testclient/sample.json
	UnitTest{
		measurement: "zipkin",
		datafile:    "testdata/distributed_trace_sample.dat",
		expected: []TestData{
			{
				expectedTags: map[string]string{
					"id":               "6802735349851856000",
					"parent_id":        "6802735349851856000",
					"trace_id":         "0:6802735349851856000",
					"name":             "main.dud",
					"service_name":     "go-zipkin-testclient",
					"annotation_value": "cs",
					"endpoint_host":    "0:9410",
				},
				expectedValues: map[string]interface{}{
					"annotation_timestamp": int64(1433330263415871),
				},
			},
			{
				expectedTags: map[string]string{
					"id":               "6802735349851856000",
					"parent_id":        "6802735349851856000",
					"trace_id":         "0:6802735349851856000",
					"name":             "main.dud",
					"service_name":     "go-zipkin-testclient",
					"annotation_value": "cr",
					"endpoint_host":    "0:9410",
				},
				expectedValues: map[string]interface{}{
					"annotation_timestamp": int64(1433330263415872),
				},
			},
		},
	},
}
