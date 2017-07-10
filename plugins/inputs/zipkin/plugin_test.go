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

var tests []func(*testing.T)

func testBasicSpans(t *testing.T) {
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

	postTestData(t, "testdata/threespans.dat")
	acc.Wait(6)

	if len(acc.Errors) != 0 {
		for _, e := range acc.Errors {
			fmt.Println(e)
		}
		t.Fatal("Errors were added during request")
	}

	// Actual testing:

	// The tags we will be querying by:

	// Test for the first span in the trace:
	tags := map[string]string{
		"id":               "8090652509916334619",
		"parent_id":        "22964302721410078",
		"trace_id":         "2505404965370368069",
		"name":             "Child",
		"service_name":     "trivial",
		"annotation_value": "trivial",
		"endpoint_host":    "2130706433:0",
		"key":              "lc",
		"type":             "STRING",
	}

	// assertContainsTaggedDuration asserts that the specified field which corresponds to `tags` has
	// the specified value. In this case, we are testing that the measurement zipkin with tags `tags` has a
	// field called `duration` with the value 53106 microseconds

	assertContainsTaggedDuration(t, &acc, "zipkin", "duration", time.Duration(53106)*time.Microsecond, tags)

	// Test for the second span in the trace:

	//Update tags in order to perform our next test

	tags = map[string]string{
		"id":               "103618986556047333",
		"parent_id":        "22964302721410078",
		"trace_id":         "2505404965370368069",
		"name":             "Child",
		"service_name":     "trivial",
		"annotation_value": "trivial",
		"endpoint_host":    "2130706433:0",
		"key":              "lc",
		"type":             "STRING",
	}

	//Similar test as above, but with different tags
	assertContainsTaggedDuration(t, &acc, "zipkin", "duration", time.Duration(50410)*time.Microsecond, tags)

	//test for the third span in the trace (with three annotations)

	tags = map[string]string{
		"id":               "22964302721410078",
		"parent_id":        "22964302721410078",
		"trace_id":         "2505404965370368069",
		"name":             "Parent",
		"service_name":     "trivial",
		"annotation_value": "Starting child #0",
		"endpoint_host":    "2130706433:0",
	}

	// test for existence of annotation specific fields
	assertContainsTaggedInt64(t, &acc, "zipkin", "annotation_timestamp", 1498688360851325, tags)

	// test for existence of annotation specific fields
	tags["annotation_value"] = "Starting child #1"
	assertContainsTaggedInt64(t, &acc, "zipkin", "annotation_timestamp", 1498688360904545, tags)

	//test for existence of annotation specific fields
	tags["annotation_value"] = "A Log"
	assertContainsTaggedInt64(t, &acc, "zipkin", "annotation_timestamp", 1498688360954992, tags)

	tags = map[string]string{
		"id":               "22964302721410078",
		"parent_id":        "22964302721410078",
		"trace_id":         "2505404965370368069",
		"name":             "Parent",
		"service_name":     "trivial",
		"annotation_value": "trivial",
		"endpoint_host":    "2130706433:0",
		"key":              "lc",
		"type":             "STRING",
	}
	// test for existence of span time stamp in third span, using binary annotation specific values.
	assertContainsTaggedDuration(t, &acc, "zipkin", "duration", time.Duration(103680)*time.Microsecond, tags)
	log.Println("end")
	log.Println("TIMESTAMP: ", acc.Metrics[5].Time)
	assertTimeIs(t, &acc, "zipkin", time.Unix(1498688360, 851318*int64(time.Microsecond)), tags)

}

func TestZipkinPlugin(t *testing.T) {
	tests = append(tests, testBasicSpans)

	// iterate through all tests to run each test in sequence
	for _, test := range tests {
		t.Run("Trivial Test", test)
	}
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

			t.Errorf("Expected value %d\n got value %d\n", expectedValue, actualValue)

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
	log.Println("going through tagged ")
	var actualValue interface{}
	log.Println(acc.Metrics)
	for _, pt := range acc.Metrics {
		log.Println("looping, point is : ", pt)
		log.Println("point tags are : ", pt.Tags)
		if pt.Measurement == measurement && reflect.DeepEqual(pt.Tags, tags) {
			log.Println("found measurement")
			for fieldname, value := range pt.Fields {
				fmt.Println("looping through fields")
				if fieldname == field {
					fmt.Println("found field: ", field)
					actualValue = value
					fmt.Println("Value: ", value)
					if value == expectedValue {
						return
					}
					t.Errorf("Expected value %d\n got value %d\n", expectedValue, value)
				}
			}
		}
	}
	msg := fmt.Sprintf(
		"Could not find measurement \"%s\" with requested tags within %s, Actual: %d",
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
		if pt.Measurement == measurement && reflect.DeepEqual(pt.Tags, tags) {
			log.Println("found measurement")
			for fieldname, value := range pt.Fields {
				fmt.Println("looping through fields")
				if fieldname == field {
					fmt.Println("found field: ", field)
					actualValue = value
					fmt.Println("Value: ", value)
					if value == expectedValue {
						return
					}
					t.Errorf("Expected value %v\n got value %v\n", expectedValue, value)
				}
			}
		}
	}
	msg := fmt.Sprintf(
		"Could not find measurement \"%s\" with requested tags within %s, Actual: %d",
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
	client := &http.Client{}
	_, err = client.Do(req)
	if err != nil {
		t.Fatal("http request failed")
	}
}
