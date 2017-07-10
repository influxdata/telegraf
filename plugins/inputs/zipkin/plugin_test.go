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

type UnitTest struct {
	datfile        string
	tags           []map[string]string
	expectedFields []map[string]interface{}
}

func TestZipkinPlugin(t *testing.T) {
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

	for _, m := range acc.Metrics {
		log.Println(m)
	}

	if len(acc.Errors) != 0 {
		for _, e := range acc.Errors {
			fmt.Println(e)
		}
		t.Fatal("Errors were added during request")
	}

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

	//map[endpoint_host:2130706433:0 name:Child service_name:trivial trace_id:2505404965370368069 annotation_value:trivial key:lc type:STRING id:8090652509916334619 parent_id:22964302721410078]

	//map[trace_id:2505404965370368069 key:lc type:STRING id:103618986556047333 name:Child service_name:trivial annotation_value:trivial endpoint_host:2130706433:0 parent_id:22964302721410078]

	assertContainsTaggedDuration(t, &acc, "zipkin", "duration", time.Duration(53106)*time.Microsecond, tags)

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
	assertContainsTaggedDuration(t, &acc, "zipkin", "duration", time.Duration(50410)*time.Microsecond, tags)

	tags = map[string]string{
		"id":               "22964302721410078",
		"parent_id":        "22964302721410078",
		"trace_id":         "2505404965370368069",
		"name":             "Parent",
		"service_name":     "trivial",
		"annotation_value": "Starting child #0",
		"endpoint_host":    "2130706433:0",
	}
	assertContainsTaggedInt64(t, &acc, "zipkin", "annotation_timestamp", 1498688360851325, tags)
	tags["annotation_value"] = "Starting child #1"
	assertContainsTaggedInt64(t, &acc, "zipkin", "annotation_timestamp", 1498688360904545, tags)
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

	assertContainsTaggedDuration(t, &acc, "zipkin", "duration", time.Duration(103680)*time.Microsecond, tags)
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
	dat, err := ioutil.ReadFile("testdata/threespans.dat")
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
