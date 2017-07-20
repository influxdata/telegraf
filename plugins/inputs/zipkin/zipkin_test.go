package zipkin

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/influxdata/telegraf/testutil"
)

func (u UnitTest) Run(t *testing.T, acc *testutil.Accumulator) {
	postTestData(t, u.datafile)
	if u.waitPoints == 0 {
		acc.Wait(len(u.expected))
	} else {
		acc.Wait(u.waitPoints)
	}

	for _, data := range u.expected {
		for key, value := range data.expectedValues {
			switch value.(type) {
			case int64:
				//assertContainsTaggedInt64(t, acc, u.measurement, key, value.(int64), data.expectedTags)
				break
			case time.Duration:
				assertContainsTaggedDuration(t, acc, u.measurement, key, value.(time.Duration), data.expectedTags)
				break
			case time.Time:
				if key == "time" {
					assertTimeIs(t, acc, u.measurement, value.(time.Time), data.expectedTags)
				} else {
					assertContainsTaggedTime(t, acc, u.measurement, key, value.(time.Time), data.expectedTags)
				}
				break
			default:
				t.Fatalf("Invalid type for field %v\n", reflect.TypeOf(value))
				break
			}
		}
	}
}

func TestZipkin(t *testing.T) {
	var acc testutil.Accumulator
	z := &Zipkin{
		Path: "/api/v1/test",
		Port: 9411,
	}
	err := z.Start(&acc)
	if err != nil {
		t.Fatal("Failed to start zipkin server")
	}
	defer z.Stop()

	for _, test := range tests {
		test.Run(t, &acc)
	}

	//t.Fatal("ERROR!")
}

func assertTimeIs(t *testing.T, acc *testutil.Accumulator,
	measurement string,
	expectedValue time.Time,
	tags map[string]string) {
	var actualValue time.Time
	for _, pt := range acc.Metrics {
		if pt.Measurement == measurement && reflect.DeepEqual(pt.Tags, tags) {
			actualValue = pt.Time
			if actualValue == expectedValue {
				return
			}

			t.Errorf("Expected value %v\n got value %v\n", expectedValue, actualValue)

		}
	}

	msg := fmt.Sprintf(
		"Could not find measurement \"%s\" with requested tags and time: %v, Actual: %v",
		measurement, expectedValue, actualValue)
	t.Fatal(msg)
}

func assertContainsTaggedDuration(
	t *testing.T,
	acc *testutil.Accumulator,
	measurement string,
	field string,
	expectedValue time.Duration,
	tags map[string]string,
) {
	var actualValue interface{}
	for _, pt := range acc.Metrics {
		if pt.Measurement == measurement && reflect.DeepEqual(pt.Tags, tags) {
			for fieldname, value := range pt.Fields {
				if fieldname == field {
					actualValue = value
					if value == expectedValue {
						return
					}
					t.Errorf("Expected value %d\n got value %d\n", expectedValue, value)
				}
			}
		}
	}
	msg := fmt.Sprintf(
		"assertContainsTaggedDuration: Could not find measurement \"%s\" with requested tags within %s, Actual: %d",
		measurement, field, actualValue)
	t.Fatal(msg)
}

func assertContainsTaggedInt64(
	t *testing.T,
	acc *testutil.Accumulator,
	measurement string,
	field string,
	expectedValue int64,
	tags map[string]string,
) {
	log.Println("going through tagged ")
	var actualValue interface{}
	log.Println(acc.Metrics)
	for _, pt := range acc.Metrics {
		log.Println("looping, point is : ", pt)
		log.Println("point tags are : ", pt.Tags)
		log.Println("point fields are:", pt.Fields)
		log.Println("tags are: ", tags)
		if pt.Measurement == measurement && reflect.DeepEqual(pt.Tags, tags) {
			log.Println("found measurement")
			for fieldname, value := range pt.Fields {
				fmt.Println("looping through fields, fieldname is: ", fieldname)
				fmt.Println("user input field is: ", field)
				if fieldname == field {
					fmt.Println("found field: ", field)
					actualValue = value
					fmt.Println("Value: ", value)
					if value == expectedValue {
						return
					}
					t.Errorf("Expected value %v\n got value %v\n", expectedValue, value)
				} else {
					t.Errorf("Fieldname != field %s", fieldname)
				}
			}
		} else if !reflect.DeepEqual(pt.Tags, tags) {
			log.Printf("%s\n%s", pt.Tags, tags)
		}
	}
	msg := fmt.Sprintf(
		"assertContainsTaggedInt64: Could not find measurement \"%s\" with requested tags within %s, Actual: %d ,Expected: %d",
		measurement, field, actualValue, expectedValue)
	t.Fatal(msg)
}

func assertContainsTaggedTime(
	t *testing.T,
	acc *testutil.Accumulator,
	measurement string,
	field string,
	expectedValue time.Time,
	tags map[string]string,
) {
	var actualValue interface{}
	for _, pt := range acc.Metrics {
		if pt.Measurement == measurement && reflect.DeepEqual(pt.Tags, tags) {
			for fieldname, value := range pt.Fields {
				if fieldname == field {
					actualValue = value
					if value == expectedValue {
						return
					}
					t.Errorf("Expected value %v\n got value %v\n", expectedValue, value)
				}
			}
		}
	}
	msg := fmt.Sprintf(
		"assertContainsTaggedTime: Could not find measurement \"%s\" with requested tags within %s, Actual: %d",
		measurement, field, actualValue)
	t.Fatal(msg)
}

func postTestData(t *testing.T, datafile string) {
	dat, err := ioutil.ReadFile(datafile)
	if err != nil {
		t.Fatal("Could not read from data file")
	}

	req, err := http.NewRequest("POST", "http://localhost:9411/api/v1/test", bytes.NewReader(dat))

	if err != nil {
		t.Fatal("bad http request")
	}

	req.Header.Set("Content-Type", "application/x-thrift")
	client := &http.Client{}
	_, err = client.Do(req)
	if err != nil {
		t.Fatal("http request failed")
	}
}

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
					"time":     time.Unix(1498688360, 851318*int64(time.Microsecond)).UTC(),
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
