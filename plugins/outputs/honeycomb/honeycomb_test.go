package honeycomb

import (
	"github.com/honeycombio/libhoney-go"
	"github.com/honeycombio/libhoney-go/transmission"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"reflect"
	"testing"
	"time"
)

func MockInit() {
	libhoney.Init(libhoney.Config{
		APIKey:       "foo",
		Dataset:      "bar",
		APIHost:      "http://localhost:1234",
		Transmission: &transmission.MockSender{},
	})
}

func TestWrite(t *testing.T) {
	h := Honeycomb{}
	MockInit()

	testMetric, _ := metric.New("testName", map[string]string{"tag1": "value1", "tag2": "value2"}, map[string]interface{}{"foo": 1, "bar": 100.123}, time.Now())

	if err := h.Write([]telegraf.Metric{testMetric}); err != nil {
		t.Error(err)
	}
}

func TestBuildEventSimple(t *testing.T) {

	h := Honeycomb{}
	MockInit()

	testTime := time.Now()
	testEvent := libhoney.NewEvent()
	testEvent.Add(map[string]interface{}{
		"testName.tag1": "value1",
		"testName.tag2": "value2",
		"testName.foo":  int64(1),
		"testName.bar":  100.123,
	})
	testEvent.Timestamp = testTime
	testMetric, _ := metric.New("testName", map[string]string{"tag1": "value1", "tag2": "value2"}, map[string]interface{}{"foo": 1, "bar": 100.123}, testTime)

	retEvent, err := h.BuildEvent(testMetric)

	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(testEvent.Fields(), retEvent.Fields()) {
		t.Errorf("\nexpected\t%+v \nreceived\t%+v \n", testEvent.Fields(), retEvent.Fields())
	}
	if testEvent.Timestamp != retEvent.Timestamp {
		t.Errorf("\nexpected\t%+v \nreceived\t%+v \n", testEvent.Timestamp, retEvent.Timestamp)
	}
}

func TestBuildEventSimpleWithHost(t *testing.T) {

	h := Honeycomb{SpecialTags: []string{"host"}}
	MockInit()

	testTime := time.Now()
	testEvent := libhoney.NewEvent()
	testEvent.Add(map[string]interface{}{
		"testName.tag1": "value1",
		"testName.tag2": "value2",
		"testName.foo":  int64(1),
		"testName.bar":  100.123,
		"host":          "host-1",
	})
	testEvent.Timestamp = testTime
	testMetric, _ := metric.New("testName", map[string]string{"tag1": "value1", "tag2": "value2", "host": "host-1"}, map[string]interface{}{"foo": 1, "bar": 100.123}, testTime)

	retEvent, err := h.BuildEvent(testMetric)

	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(testEvent.Fields(), retEvent.Fields()) {
		t.Errorf("\nexpected\t%+v \nreceived\t%+v \n", testEvent.Fields(), retEvent.Fields())
	}
	if testEvent.Timestamp != retEvent.Timestamp {
		t.Errorf("\nexpected\t%+v \nreceived\t%+v \n", testEvent.Timestamp, retEvent.Timestamp)
	}
}

func TestBuildEventSimpleWithSpecialTags(t *testing.T) {

	h := Honeycomb{SpecialTags: []string{"tag1", "special2"}}
	MockInit()

	testTime := time.Now()
	testEvent := libhoney.NewEvent()
	testEvent.Add(map[string]interface{}{
		"tag1":          "value1",
		"testName.tag2": "value2",
		"testName.foo":  int64(1),
		"testName.bar":  100.123,
		"testName.host": "host-1",
		"special2":      "another-value",
	})
	testEvent.Timestamp = testTime
	testMetric, _ := metric.New("testName", map[string]string{"tag1": "value1", "tag2": "value2", "host": "host-1", "special2": "another-value"}, map[string]interface{}{"foo": 1, "bar": 100.123}, testTime)

	retEvent, err := h.BuildEvent(testMetric)

	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(testEvent.Fields(), retEvent.Fields()) {
		t.Errorf("\nexpected\t%+v \nreceived\t%+v \n", testEvent.Fields(), retEvent.Fields())
	}
	if testEvent.Timestamp != retEvent.Timestamp {
		t.Errorf("\nexpected\t%+v \nreceived\t%+v \n", testEvent.Timestamp, retEvent.Timestamp)
	}
}
