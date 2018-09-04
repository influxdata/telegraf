package appinsights

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/Microsoft/ApplicationInsights-Go/appinsights/contracts"
)

const float_precision = 1e-4

func checkDataContract(t *testing.T, property string, actual, expected interface{}) {
	if x, ok := actual.(float64); ok {
		if y, ok := expected.(float64); ok {
			if math.Abs(x-y) > float_precision {
				t.Errorf("Float property %s mismatched; got %f, want %f.\n", property, actual, expected)
			}

			return
		}
	}

	if actual != expected {
		t.Errorf("Property %s mismatched; got %v, want %v.\n", property, actual, expected)
	}
}

func checkNotNullOrEmpty(t *testing.T, property string, actual interface{}) {
	if actual == nil {
		t.Errorf("Property %s was expected not to be null.\n", property)
	} else if str, ok := actual.(string); ok && str == "" {
		t.Errorf("Property %s was expected not to be an empty string.\n", property)
	}
}

func TestTraceTelemetry(t *testing.T) {
	mockClock()
	defer resetClock()

	telem := NewTraceTelemetry("~my message~", Error)
	telem.Properties["prop1"] = "value1"
	telem.Properties["prop2"] = "value2"
	d := telem.TelemetryData().(*contracts.MessageData)

	checkDataContract(t, "Message", d.Message, "~my message~")
	checkDataContract(t, "SeverityLevel", d.SeverityLevel, Error)
	checkDataContract(t, "Properties[prop1]", d.Properties["prop1"], "value1")
	checkDataContract(t, "Properties[prop2]", d.Properties["prop2"], "value2")
	checkDataContract(t, "Timestamp", telem.Time(), currentClock.Now())
	checkNotNullOrEmpty(t, "ContextTags", telem.ContextTags())

	telem2 := &TraceTelemetry{
		Message:       "~my-2nd-message~",
		SeverityLevel: Critical,
	}
	d2 := telem2.TelemetryData().(*contracts.MessageData)

	checkDataContract(t, "Message", d2.Message, "~my-2nd-message~")
	checkDataContract(t, "SeverityLevel", d2.SeverityLevel, Critical)

	var telemInterface Telemetry
	if telemInterface = telem; telemInterface.GetMeasurements() != nil {
		t.Errorf("Trace.(Telemetry).GetMeasurements should return nil")
	}
}

func TestEventTelemetry(t *testing.T) {
	mockClock()
	defer resetClock()

	telem := NewEventTelemetry("~my event~")
	telem.Properties["prop1"] = "value1"
	telem.Properties["prop2"] = "value2"
	telem.Measurements["measure1"] = 1234.0
	telem.Measurements["measure2"] = 5678.0
	d := telem.TelemetryData().(*contracts.EventData)

	checkDataContract(t, "Name", d.Name, "~my event~")
	checkDataContract(t, "Properties[prop1]", d.Properties["prop1"], "value1")
	checkDataContract(t, "Properties[prop2]", d.Properties["prop2"], "value2")
	checkDataContract(t, "Measurements[measure1]", d.Measurements["measure1"], 1234.0)
	checkDataContract(t, "Measurements[measure2]", d.Measurements["measure2"], 5678.0)
	checkDataContract(t, "Timestamp", telem.Time(), currentClock.Now())
	checkNotNullOrEmpty(t, "ContextTags", telem.ContextTags())

	telem2 := &EventTelemetry{
		Name: "~my-event~",
	}
	d2 := telem2.TelemetryData().(*contracts.EventData)

	checkDataContract(t, "Name", d2.Name, "~my-event~")
}

func TestMetricTelemetry(t *testing.T) {
	mockClock()
	defer resetClock()

	telem := NewMetricTelemetry("~my metric~", 1234.0)
	telem.Properties["prop1"] = "value!"
	d := telem.TelemetryData().(*contracts.MetricData)

	checkDataContract(t, "len(Metrics)", len(d.Metrics), 1)
	dp := d.Metrics[0]
	checkDataContract(t, "DataPoint.Name", dp.Name, "~my metric~")
	checkDataContract(t, "DataPoint.Value", dp.Value, 1234.0)
	checkDataContract(t, "DataPoint.Kind", dp.Kind, Measurement)
	checkDataContract(t, "DataPoint.Count", dp.Count, 1)
	checkDataContract(t, "Properties[prop1]", d.Properties["prop1"], "value!")
	checkDataContract(t, "Timestamp", telem.Time(), currentClock.Now())
	checkNotNullOrEmpty(t, "ContextTags", telem.ContextTags())

	telem2 := &MetricTelemetry{
		Name:  "~my metric~",
		Value: 5678.0,
	}
	d2 := telem2.TelemetryData().(*contracts.MetricData)

	checkDataContract(t, "len(Metrics)", len(d2.Metrics), 1)
	dp2 := d2.Metrics[0]
	checkDataContract(t, "DataPoint.Name", dp2.Name, "~my metric~")
	checkDataContract(t, "DataPoint.Value", dp2.Value, 5678.0)
	checkDataContract(t, "DataPoint.Kind", dp2.Kind, Measurement)
	checkDataContract(t, "DataPoint.Count", dp2.Count, 1)

	var telemInterface Telemetry
	if telemInterface = telem; telemInterface.GetMeasurements() != nil {
		t.Errorf("Metric.(Telemetry).GetMeasurements should return nil")
	}
}

type statsTest struct {
	data          []float64
	stdDev        float64
	sampledStdDev float64
	min           float64
	max           float64
}

func TestAggregateMetricTelemetry(t *testing.T) {
	statsTests := []statsTest{
		statsTest{[]float64{}, 0.0, 0.0, 0.0, 0.0},
		statsTest{[]float64{0.0}, 0.0, 0.0, 0.0, 0.0},
		statsTest{[]float64{50.0}, 0.0, 0.0, 50.0, 50.0},
		statsTest{[]float64{50.0, 50.0}, 0.0, 0.0, 50.0, 50.0},
		statsTest{[]float64{50.0, 60.0}, 5.0, 7.071, 50.0, 60.0},
		statsTest{[]float64{9.0, 10.0, 11.0, 7.0, 13.0}, 2.0, 2.236, 7.0, 13.0},
		// TODO: More tests.
	}

	for _, tst := range statsTests {
		t1 := NewAggregateMetricTelemetry("foo")
		t2 := NewAggregateMetricTelemetry("foo")
		t1.AddData(tst.data)
		t2.AddSampledData(tst.data)

		checkDataPoint(t, t1, tst, false)
		checkDataPoint(t, t2, tst, true)
	}

	// Do the same as above, but add data points one at a time.
	for _, tst := range statsTests {
		t1 := NewAggregateMetricTelemetry("foo")
		t2 := NewAggregateMetricTelemetry("foo")

		for _, x := range tst.data {
			t1.AddData([]float64{x})
			t2.AddSampledData([]float64{x})
		}

		checkDataPoint(t, t1, tst, false)
		checkDataPoint(t, t2, tst, true)
	}
}

func checkDataPoint(t *testing.T, telem *AggregateMetricTelemetry, tst statsTest, sampled bool) {
	d := telem.TelemetryData().(*contracts.MetricData)
	checkDataContract(t, "len(Metrics)", len(d.Metrics), 1)
	dp := d.Metrics[0]

	var sum float64
	for _, x := range tst.data {
		sum += x
	}

	checkDataContract(t, "DataPoint.Count", dp.Count, len(tst.data))
	checkDataContract(t, "DataPoint.Min", dp.Min, tst.min)
	checkDataContract(t, "DataPoint.Max", dp.Max, tst.max)
	checkDataContract(t, "DataPoint.Value", dp.Value, sum)

	if sampled {
		checkDataContract(t, "DataPoint.StdDev (sample)", dp.StdDev, tst.sampledStdDev)
	} else {
		checkDataContract(t, "DataPoint.StdDev (population)", dp.StdDev, tst.stdDev)
	}
}

func TestRequestTelemetry(t *testing.T) {
	mockClock()
	defer resetClock()

	telem := NewRequestTelemetry("POST", "http://testurl.org/?query=value", time.Minute, "200")
	telem.Source = "127.0.0.1"
	telem.Properties["prop1"] = "value1"
	telem.Measurements["measure1"] = 999.0
	d := telem.TelemetryData().(*contracts.RequestData)

	checkNotNullOrEmpty(t, "Id", d.Id)
	checkDataContract(t, "Name", d.Name, "POST http://testurl.org/")
	checkDataContract(t, "Url", d.Url, "http://testurl.org/?query=value")
	checkDataContract(t, "Duration", d.Duration, "0.00:01:00.0000000")
	checkDataContract(t, "Success", d.Success, true)
	checkDataContract(t, "Source", d.Source, "127.0.0.1")
	checkDataContract(t, "Properties[prop1]", d.Properties["prop1"], "value1")
	checkDataContract(t, "Measurements[measure1]", d.Measurements["measure1"], 999.0)
	checkDataContract(t, "Timestamp", telem.Time(), currentClock.Now().Add(-time.Minute))
	checkNotNullOrEmpty(t, "ContextTags", telem.ContextTags())

	startTime := currentClock.Now().Add(-time.Hour)
	endTime := startTime.Add(5 * time.Minute)
	telem.MarkTime(startTime, endTime)
	d = telem.TelemetryData().(*contracts.RequestData)
	checkDataContract(t, "Timestamp", telem.Time(), startTime)
	checkDataContract(t, "Duration", d.Duration, "0.00:05:00.0000000")
}

func TestRequestTelemetrySuccess(t *testing.T) {
	// Some of these are due to default-success
	successCodes := []string{"200", "204", "301", "302", "401", "foo", "", "55555555555555555555555555555555555555555555555555"}
	failureCodes := []string{"400", "404", "500", "430"}

	for _, code := range successCodes {
		telem := NewRequestTelemetry("GET", "https://something", time.Second, code)
		d := telem.TelemetryData().(*contracts.RequestData)
		checkDataContract(t, fmt.Sprintf("Success [%s]", code), d.Success, true)
	}

	for _, code := range failureCodes {
		telem := NewRequestTelemetry("GET", "https://something", time.Second, code)
		d := telem.TelemetryData().(*contracts.RequestData)
		checkDataContract(t, fmt.Sprintf("Success [%s]", code), d.Success, false)
	}
}

func TestRemoteDependencyTelemetry(t *testing.T) {
	mockClock()
	defer resetClock()

	telem := NewRemoteDependencyTelemetry("SQL-GET", "SQL", "myhost.name", true)
	telem.Data = "<command>"
	telem.ResultCode = "OK"
	telem.Duration = time.Minute
	telem.Properties["prop1"] = "value1"
	telem.Measurements["measure1"] = 999.0
	d := telem.TelemetryData().(*contracts.RemoteDependencyData)

	checkDataContract(t, "Id", d.Id, "") // no default
	checkDataContract(t, "Data", d.Data, "<command>")
	checkDataContract(t, "Type", d.Type, "SQL")
	checkDataContract(t, "Target", d.Target, "myhost.name")
	checkDataContract(t, "ResultCode", d.ResultCode, "OK")
	checkDataContract(t, "Name", d.Name, "SQL-GET")
	checkDataContract(t, "Duration", d.Duration, "0.00:01:00.0000000")
	checkDataContract(t, "Success", d.Success, true)
	checkDataContract(t, "Properties[prop1]", d.Properties["prop1"], "value1")
	checkDataContract(t, "Measurements[measure1]", d.Measurements["measure1"], 999.0)
	checkDataContract(t, "Timestamp", telem.Time(), currentClock.Now())
	checkNotNullOrEmpty(t, "ContextTags", telem.ContextTags())

	telem.Id = "<id>"
	telem.Success = false
	d = telem.TelemetryData().(*contracts.RemoteDependencyData)
	checkDataContract(t, "Id", d.Id, "<id>")
	checkDataContract(t, "Success", d.Success, false)

	startTime := currentClock.Now().Add(-time.Hour)
	endTime := startTime.Add(5 * time.Minute)
	telem.MarkTime(startTime, endTime)
	d = telem.TelemetryData().(*contracts.RemoteDependencyData)
	checkDataContract(t, "Timestamp", telem.Time(), startTime)
	checkDataContract(t, "Duration", d.Duration, "0.00:05:00.0000000")
}

func TestAvailabilityTelemetry(t *testing.T) {
	mockClock()
	defer resetClock()

	telem := NewAvailabilityTelemetry("Frontdoor", time.Minute, true)
	telem.RunLocation = "The moon"
	telem.Message = "OK"
	telem.Properties["prop1"] = "value1"
	telem.Measurements["measure1"] = 999.0
	d := telem.TelemetryData().(*contracts.AvailabilityData)

	checkDataContract(t, "Id", d.Id, "")
	checkDataContract(t, "Name", d.Name, "Frontdoor")
	checkDataContract(t, "Duration", d.Duration, "0.00:01:00.0000000")
	checkDataContract(t, "RunLocation", d.RunLocation, "The moon")
	checkDataContract(t, "Message", d.Message, "OK")
	checkDataContract(t, "Success", d.Success, true)
	checkDataContract(t, "Properties[prop1]", d.Properties["prop1"], "value1")
	checkDataContract(t, "Measurements[measure1]", d.Measurements["measure1"], 999.0)
	checkDataContract(t, "Timestamp", telem.Time(), currentClock.Now())
	checkNotNullOrEmpty(t, "ContextTags", telem.ContextTags())

	telem.Id = "<id>"
	telem.Success = false
	d = telem.TelemetryData().(*contracts.AvailabilityData)
	checkDataContract(t, "Id", d.Id, "<id>")
	checkDataContract(t, "Success", d.Success, false)

	startTime := currentClock.Now().Add(-time.Hour)
	endTime := startTime.Add(5 * time.Minute)
	telem.MarkTime(startTime, endTime)
	d = telem.TelemetryData().(*contracts.AvailabilityData)
	checkDataContract(t, "Timestamp", telem.Time(), startTime)
	checkDataContract(t, "Duration", d.Duration, "0.00:05:00.0000000")
}

func TestPageViewTelemetry(t *testing.T) {
	mockClock()
	defer resetClock()

	telem := NewPageViewTelemetry("Home page", "http://testuri.org/")
	telem.Duration = time.Minute
	telem.Properties["prop1"] = "value1"
	telem.Measurements["measure1"] = 999.0
	d := telem.TelemetryData().(*contracts.PageViewData)

	checkDataContract(t, "Name", d.Name, "Home page")
	checkDataContract(t, "Duration", d.Duration, "0.00:01:00.0000000")
	checkDataContract(t, "Url", d.Url, "http://testuri.org/")
	checkDataContract(t, "Properties[prop1]", d.Properties["prop1"], "value1")
	checkDataContract(t, "Measurements[measure1]", d.Measurements["measure1"], 999.0)
	checkDataContract(t, "Timestamp", telem.Time(), currentClock.Now())
	checkNotNullOrEmpty(t, "ContextTags", telem.ContextTags())

	startTime := currentClock.Now().Add(-time.Hour)
	endTime := startTime.Add(5 * time.Minute)
	telem.MarkTime(startTime, endTime)
	d = telem.TelemetryData().(*contracts.PageViewData)
	checkDataContract(t, "Timestamp", telem.Time(), startTime)
	checkDataContract(t, "Duration", d.Duration, "0.00:05:00.0000000")
}

type durationTest struct {
	duration time.Duration
	expected string
}

func TestFormatDuration(t *testing.T) {
	durationTests := []durationTest{
		durationTest{time.Hour, "0.01:00:00.0000000"},
		durationTest{time.Minute, "0.00:01:00.0000000"},
		durationTest{time.Second, "0.00:00:01.0000000"},
		durationTest{time.Millisecond, "0.00:00:00.0010000"},
		durationTest{100 * time.Nanosecond, "0.00:00:00.0000001"},
		durationTest{(31 * time.Hour) + (25 * time.Minute) + (30 * time.Second) + time.Millisecond, "1.07:25:30.0010000"},
	}

	for _, tst := range durationTests {
		actual := formatDuration(tst.duration)
		if tst.expected != actual {
			t.Errorf("Mismatch. Got %s, want %s (duration %s)\n", actual, tst.expected, tst.duration.String())
		}
	}
}
