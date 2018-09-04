package appinsights

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"testing"
	"time"
)

const test_ikey = "01234567-0000-89ab-cdef-000000000000"

func TestJsonSerializerEvents(t *testing.T) {
	mockClock(time.Unix(1511001321, 0))
	defer resetClock()

	var buffer telemetryBufferItems

	buffer.add(
		NewTraceTelemetry("testing", Error),
		NewEventTelemetry("an-event"),
		NewMetricTelemetry("a-metric", 567),
	)

	req := NewRequestTelemetry("method", "my-url", time.Minute, "204")
	req.Name = "req-name"
	req.Id = "my-id"
	buffer.add(req)

	agg := NewAggregateMetricTelemetry("agg-metric")
	agg.AddData([]float64{1, 2, 3})
	buffer.add(agg)

	remdep := NewRemoteDependencyTelemetry("bing-remote-dep", "http", "www.bing.com", false)
	remdep.Data = "some-data"
	remdep.ResultCode = "arg"
	remdep.Duration = 4 * time.Second
	remdep.Properties["hi"] = "hello"
	buffer.add(remdep)

	avail := NewAvailabilityTelemetry("webtest", 8*time.Second, true)
	avail.RunLocation = "jupiter"
	avail.Message = "ok."
	avail.Measurements["measure"] = 88.0
	avail.Id = "avail-id"
	buffer.add(avail)

	view := NewPageViewTelemetry("name", "http://bing.com")
	view.Duration = 4 * time.Minute
	buffer.add(view)

	j, err := parsePayload(buffer.serialize())
	if err != nil {
		t.Errorf("Error parsing payload: %s", err.Error())
	}

	if len(j) != 8 {
		t.Fatal("Unexpected event count")
	}

	// Trace
	j[0].assertPath(t, "iKey", test_ikey)
	j[0].assertPath(t, "name", "Microsoft.ApplicationInsights.01234567000089abcdef000000000000.Message")
	j[0].assertPath(t, "time", "2017-11-18T10:35:21Z")
	j[0].assertPath(t, "sampleRate", 100.0)
	j[0].assertPath(t, "data.baseType", "MessageData")
	j[0].assertPath(t, "data.baseData.message", "testing")
	j[0].assertPath(t, "data.baseData.severityLevel", 3)
	j[0].assertPath(t, "data.baseData.ver", 2)

	// Event
	j[1].assertPath(t, "iKey", test_ikey)
	j[1].assertPath(t, "name", "Microsoft.ApplicationInsights.01234567000089abcdef000000000000.Event")
	j[1].assertPath(t, "time", "2017-11-18T10:35:21Z")
	j[1].assertPath(t, "sampleRate", 100.0)
	j[1].assertPath(t, "data.baseType", "EventData")
	j[1].assertPath(t, "data.baseData.name", "an-event")
	j[1].assertPath(t, "data.baseData.ver", 2)

	// Metric
	j[2].assertPath(t, "iKey", test_ikey)
	j[2].assertPath(t, "name", "Microsoft.ApplicationInsights.01234567000089abcdef000000000000.Metric")
	j[2].assertPath(t, "time", "2017-11-18T10:35:21Z")
	j[2].assertPath(t, "sampleRate", 100.0)
	j[2].assertPath(t, "data.baseType", "MetricData")
	j[2].assertPath(t, "data.baseData.metrics.<len>", 1)
	j[2].assertPath(t, "data.baseData.metrics.[0].value", 567)
	j[2].assertPath(t, "data.baseData.metrics.[0].count", 1)
	j[2].assertPath(t, "data.baseData.metrics.[0].kind", 0)
	j[2].assertPath(t, "data.baseData.metrics.[0].name", "a-metric")
	j[2].assertPath(t, "data.baseData.ver", 2)

	// Request
	j[3].assertPath(t, "iKey", test_ikey)
	j[3].assertPath(t, "name", "Microsoft.ApplicationInsights.01234567000089abcdef000000000000.Request")
	j[3].assertPath(t, "time", "2017-11-18T10:34:21Z") // Constructor subtracts duration
	j[3].assertPath(t, "sampleRate", 100.0)
	j[3].assertPath(t, "data.baseType", "RequestData")
	j[3].assertPath(t, "data.baseData.name", "req-name")
	j[3].assertPath(t, "data.baseData.duration", "0.00:01:00.0000000")
	j[3].assertPath(t, "data.baseData.responseCode", "204")
	j[3].assertPath(t, "data.baseData.success", true)
	j[3].assertPath(t, "data.baseData.id", "my-id")
	j[3].assertPath(t, "data.baseData.url", "my-url")
	j[3].assertPath(t, "data.baseData.ver", 2)

	// Aggregate metric
	j[4].assertPath(t, "iKey", test_ikey)
	j[4].assertPath(t, "name", "Microsoft.ApplicationInsights.01234567000089abcdef000000000000.Metric")
	j[4].assertPath(t, "time", "2017-11-18T10:35:21Z")
	j[4].assertPath(t, "sampleRate", 100.0)
	j[4].assertPath(t, "data.baseType", "MetricData")
	j[4].assertPath(t, "data.baseData.metrics.<len>", 1)
	j[4].assertPath(t, "data.baseData.metrics.[0].value", 6)
	j[4].assertPath(t, "data.baseData.metrics.[0].count", 3)
	j[4].assertPath(t, "data.baseData.metrics.[0].kind", 1)
	j[4].assertPath(t, "data.baseData.metrics.[0].min", 1)
	j[4].assertPath(t, "data.baseData.metrics.[0].max", 3)
	j[4].assertPath(t, "data.baseData.metrics.[0].stdDev", 0.8164)
	j[4].assertPath(t, "data.baseData.metrics.[0].name", "agg-metric")
	j[4].assertPath(t, "data.baseData.ver", 2)

	// Remote dependency
	j[5].assertPath(t, "iKey", test_ikey)
	j[5].assertPath(t, "name", "Microsoft.ApplicationInsights.01234567000089abcdef000000000000.RemoteDependency")
	j[5].assertPath(t, "time", "2017-11-18T10:35:21Z")
	j[5].assertPath(t, "sampleRate", 100.0)
	j[5].assertPath(t, "data.baseType", "RemoteDependencyData")
	j[5].assertPath(t, "data.baseData.name", "bing-remote-dep")
	j[5].assertPath(t, "data.baseData.id", "")
	j[5].assertPath(t, "data.baseData.resultCode", "arg")
	j[5].assertPath(t, "data.baseData.duration", "0.00:00:04.0000000")
	j[5].assertPath(t, "data.baseData.success", false)
	j[5].assertPath(t, "data.baseData.data", "some-data")
	j[5].assertPath(t, "data.baseData.target", "www.bing.com")
	j[5].assertPath(t, "data.baseData.type", "http")
	j[5].assertPath(t, "data.baseData.properties.hi", "hello")
	j[5].assertPath(t, "data.baseData.ver", 2)

	// Availability
	j[6].assertPath(t, "iKey", test_ikey)
	j[6].assertPath(t, "name", "Microsoft.ApplicationInsights.01234567000089abcdef000000000000.Availability")
	j[6].assertPath(t, "time", "2017-11-18T10:35:21Z")
	j[6].assertPath(t, "sampleRate", 100.0)
	j[6].assertPath(t, "data.baseType", "AvailabilityData")
	j[6].assertPath(t, "data.baseData.name", "webtest")
	j[6].assertPath(t, "data.baseData.duration", "0.00:00:08.0000000")
	j[6].assertPath(t, "data.baseData.success", true)
	j[6].assertPath(t, "data.baseData.runLocation", "jupiter")
	j[6].assertPath(t, "data.baseData.message", "ok.")
	j[6].assertPath(t, "data.baseData.id", "avail-id")
	j[6].assertPath(t, "data.baseData.ver", 2)
	j[6].assertPath(t, "data.baseData.measurements.measure", 88)

	// Page view
	j[7].assertPath(t, "iKey", test_ikey)
	j[7].assertPath(t, "name", "Microsoft.ApplicationInsights.01234567000089abcdef000000000000.PageView")
	j[7].assertPath(t, "time", "2017-11-18T10:35:21Z")
	j[7].assertPath(t, "sampleRate", 100.0)
	j[7].assertPath(t, "data.baseType", "PageViewData")
	j[7].assertPath(t, "data.baseData.name", "name")
	j[7].assertPath(t, "data.baseData.url", "http://bing.com")
	j[7].assertPath(t, "data.baseData.duration", "0.00:04:00.0000000")
	j[7].assertPath(t, "data.baseData.ver", 2)
}

func TestJsonSerializerNakedEvents(t *testing.T) {
	mockClock(time.Unix(1511001321, 0))
	defer resetClock()

	var buffer telemetryBufferItems

	buffer.add(
		&TraceTelemetry{
			Message:       "Naked telemetry",
			SeverityLevel: Warning,
		},
		&EventTelemetry{
			Name: "Naked event",
		},
		&MetricTelemetry{
			Name:  "my-metric",
			Value: 456.0,
		},
		&AggregateMetricTelemetry{
			Name:   "agg-metric",
			Value:  50,
			Min:    2,
			Max:    7,
			Count:  9,
			StdDev: 3,
		},
		&RequestTelemetry{
			Name:         "req-name",
			Url:          "req-url",
			Duration:     time.Minute,
			ResponseCode: "Response",
			Success:      true,
			Source:       "localhost",
		},
		&RemoteDependencyTelemetry{
			Name:       "dep-name",
			ResultCode: "ok.",
			Duration:   time.Hour,
			Success:    true,
			Data:       "dep-data",
			Type:       "dep-type",
			Target:     "dep-target",
		},
		&AvailabilityTelemetry{
			Name:        "avail-name",
			Duration:    3 * time.Minute,
			Success:     true,
			RunLocation: "run-loc",
			Message:     "avail-msg",
		},
		&PageViewTelemetry{
			Url:      "page-view-url",
			Duration: 4 * time.Second,
			Name:     "page-view-name",
		},
	)

	j, err := parsePayload(buffer.serialize())
	if err != nil {
		t.Errorf("Error parsing payload: %s", err.Error())
	}

	if len(j) != 8 {
		t.Fatal("Unexpected event count")
	}

	// Trace
	j[0].assertPath(t, "iKey", test_ikey)
	j[0].assertPath(t, "name", "Microsoft.ApplicationInsights.01234567000089abcdef000000000000.Message")
	j[0].assertPath(t, "time", "2017-11-18T10:35:21Z")
	j[0].assertPath(t, "sampleRate", 100)
	j[0].assertPath(t, "data.baseType", "MessageData")
	j[0].assertPath(t, "data.baseData.message", "Naked telemetry")
	j[0].assertPath(t, "data.baseData.severityLevel", 2)
	j[0].assertPath(t, "data.baseData.ver", 2)

	// Event
	j[1].assertPath(t, "iKey", test_ikey)
	j[1].assertPath(t, "name", "Microsoft.ApplicationInsights.01234567000089abcdef000000000000.Event")
	j[1].assertPath(t, "time", "2017-11-18T10:35:21Z")
	j[1].assertPath(t, "sampleRate", 100)
	j[1].assertPath(t, "data.baseType", "EventData")
	j[1].assertPath(t, "data.baseData.name", "Naked event")
	j[1].assertPath(t, "data.baseData.ver", 2)

	// Metric
	j[2].assertPath(t, "iKey", test_ikey)
	j[2].assertPath(t, "name", "Microsoft.ApplicationInsights.01234567000089abcdef000000000000.Metric")
	j[2].assertPath(t, "time", "2017-11-18T10:35:21Z")
	j[2].assertPath(t, "sampleRate", 100)
	j[2].assertPath(t, "data.baseType", "MetricData")
	j[2].assertPath(t, "data.baseData.metrics.<len>", 1)
	j[2].assertPath(t, "data.baseData.metrics.[0].value", 456)
	j[2].assertPath(t, "data.baseData.metrics.[0].count", 1)
	j[2].assertPath(t, "data.baseData.metrics.[0].kind", 0)
	j[2].assertPath(t, "data.baseData.metrics.[0].name", "my-metric")
	j[2].assertPath(t, "data.baseData.ver", 2)

	// Aggregate metric
	j[3].assertPath(t, "iKey", test_ikey)
	j[3].assertPath(t, "name", "Microsoft.ApplicationInsights.01234567000089abcdef000000000000.Metric")
	j[3].assertPath(t, "time", "2017-11-18T10:35:21Z")
	j[3].assertPath(t, "sampleRate", 100.0)
	j[3].assertPath(t, "data.baseType", "MetricData")
	j[3].assertPath(t, "data.baseData.metrics.<len>", 1)
	j[3].assertPath(t, "data.baseData.metrics.[0].value", 50)
	j[3].assertPath(t, "data.baseData.metrics.[0].count", 9)
	j[3].assertPath(t, "data.baseData.metrics.[0].kind", 1)
	j[3].assertPath(t, "data.baseData.metrics.[0].min", 2)
	j[3].assertPath(t, "data.baseData.metrics.[0].max", 7)
	j[3].assertPath(t, "data.baseData.metrics.[0].stdDev", 3)
	j[3].assertPath(t, "data.baseData.metrics.[0].name", "agg-metric")
	j[3].assertPath(t, "data.baseData.ver", 2)

	// Request
	j[4].assertPath(t, "iKey", test_ikey)
	j[4].assertPath(t, "name", "Microsoft.ApplicationInsights.01234567000089abcdef000000000000.Request")
	j[4].assertPath(t, "time", "2017-11-18T10:35:21Z") // Context takes current time since it's not supplied
	j[4].assertPath(t, "sampleRate", 100.0)
	j[4].assertPath(t, "data.baseType", "RequestData")
	j[4].assertPath(t, "data.baseData.name", "req-name")
	j[4].assertPath(t, "data.baseData.duration", "0.00:01:00.0000000")
	j[4].assertPath(t, "data.baseData.responseCode", "Response")
	j[4].assertPath(t, "data.baseData.success", true)
	j[4].assertPath(t, "data.baseData.url", "req-url")
	j[4].assertPath(t, "data.baseData.source", "localhost")
	j[4].assertPath(t, "data.baseData.ver", 2)

	if id, err := j[4].getPath("data.baseData.id"); err != nil {
		t.Errorf("Id not present")
	} else if len(id.(string)) == 0 {
		t.Errorf("Empty request id")
	}

	// Remote dependency
	j[5].assertPath(t, "iKey", test_ikey)
	j[5].assertPath(t, "name", "Microsoft.ApplicationInsights.01234567000089abcdef000000000000.RemoteDependency")
	j[5].assertPath(t, "time", "2017-11-18T10:35:21Z")
	j[5].assertPath(t, "sampleRate", 100.0)
	j[5].assertPath(t, "data.baseType", "RemoteDependencyData")
	j[5].assertPath(t, "data.baseData.name", "dep-name")
	j[5].assertPath(t, "data.baseData.id", "")
	j[5].assertPath(t, "data.baseData.resultCode", "ok.")
	j[5].assertPath(t, "data.baseData.duration", "0.01:00:00.0000000")
	j[5].assertPath(t, "data.baseData.success", true)
	j[5].assertPath(t, "data.baseData.data", "dep-data")
	j[5].assertPath(t, "data.baseData.target", "dep-target")
	j[5].assertPath(t, "data.baseData.type", "dep-type")
	j[5].assertPath(t, "data.baseData.ver", 2)

	// Availability
	j[6].assertPath(t, "iKey", test_ikey)
	j[6].assertPath(t, "name", "Microsoft.ApplicationInsights.01234567000089abcdef000000000000.Availability")
	j[6].assertPath(t, "time", "2017-11-18T10:35:21Z")
	j[6].assertPath(t, "sampleRate", 100.0)
	j[6].assertPath(t, "data.baseType", "AvailabilityData")
	j[6].assertPath(t, "data.baseData.name", "avail-name")
	j[6].assertPath(t, "data.baseData.duration", "0.00:03:00.0000000")
	j[6].assertPath(t, "data.baseData.success", true)
	j[6].assertPath(t, "data.baseData.runLocation", "run-loc")
	j[6].assertPath(t, "data.baseData.message", "avail-msg")
	j[6].assertPath(t, "data.baseData.id", "")
	j[6].assertPath(t, "data.baseData.ver", 2)

	// Page view
	j[7].assertPath(t, "iKey", test_ikey)
	j[7].assertPath(t, "name", "Microsoft.ApplicationInsights.01234567000089abcdef000000000000.PageView")
	j[7].assertPath(t, "time", "2017-11-18T10:35:21Z")
	j[7].assertPath(t, "sampleRate", 100.0)
	j[7].assertPath(t, "data.baseType", "PageViewData")
	j[7].assertPath(t, "data.baseData.name", "page-view-name")
	j[7].assertPath(t, "data.baseData.url", "page-view-url")
	j[7].assertPath(t, "data.baseData.duration", "0.00:00:04.0000000")
	j[7].assertPath(t, "data.baseData.ver", 2)
}

// Test helpers...

func telemetryBuffer(items ...Telemetry) telemetryBufferItems {
	ctx := NewTelemetryContext(test_ikey)
	ctx.iKey = test_ikey

	var result telemetryBufferItems
	for _, item := range items {
		result = append(result, ctx.envelop(item))
	}

	return result
}

func (buffer *telemetryBufferItems) add(items ...Telemetry) {
	*buffer = append(*buffer, telemetryBuffer(items...)...)
}

type jsonMessage map[string]interface{}
type jsonPayload []jsonMessage

func parsePayload(payload []byte) (jsonPayload, error) {
	// json.Decoder can detect line endings for us but I'd like to explicitly find them.
	var result jsonPayload
	for _, item := range bytes.Split(payload, []byte("\n")) {
		if len(item) == 0 {
			continue
		}

		decoder := json.NewDecoder(bytes.NewReader(item))
		msg := make(jsonMessage)
		if err := decoder.Decode(&msg); err == nil {
			result = append(result, msg)
		} else {
			return result, err
		}
	}

	return result, nil
}

func (msg jsonMessage) assertPath(t *testing.T, path string, value interface{}) {
	const tolerance = 0.0001
	v, err := msg.getPath(path)
	if err != nil {
		t.Error(err.Error())
		return
	}

	if num, ok := value.(int); ok {
		if vnum, ok := v.(float64); ok {
			if math.Abs(float64(num)-vnum) > tolerance {
				t.Errorf("Data was unexpected at %s. Got %g want %d", path, vnum, num)
			}
		} else if vnum, ok := v.(int); ok {
			if vnum != num {
				t.Errorf("Data was unexpected at %s. Got %d want %d", path, vnum, num)
			}
		} else {
			t.Errorf("Expected value at %s to be a number, but was %T", path, v)
		}
	} else if num, ok := value.(float64); ok {
		if vnum, ok := v.(float64); ok {
			if math.Abs(num-vnum) > tolerance {
				t.Errorf("Data was unexpected at %s. Got %g want %g", path, vnum, num)
			}
		} else if vnum, ok := v.(int); ok {
			if math.Abs(num-float64(vnum)) > tolerance {
				t.Errorf("Data was unexpected at %s. Got %d want %g", path, vnum, num)
			}
		} else {
			t.Errorf("Expected value at %s to be a number, but was %T", path, v)
		}
	} else if str, ok := value.(string); ok {
		if vstr, ok := v.(string); ok {
			if str != vstr {
				t.Errorf("Data was unexpected at %s. Got '%s' want '%s'", path, vstr, str)
			}
		} else {
			t.Errorf("Expected value at %s to be a string, but was %T", path, v)
		}
	} else if bl, ok := value.(bool); ok {
		if vbool, ok := v.(bool); ok {
			if bl != vbool {
				t.Errorf("Data was unexpected at %s. Got %t want %t", path, vbool, bl)
			}
		} else {
			t.Errorf("Expected value at %s to be a bool, but was %T", path, v)
		}
	} else {
		t.Errorf("Unsupported type: %#v", value)
	}
}

func (msg jsonMessage) getPath(path string) (interface{}, error) {
	parts := strings.Split(path, ".")
	var obj interface{} = msg
	for i, part := range parts {
		if strings.HasPrefix(part, "[") && strings.HasSuffix(part, "]") {
			// Array
			idxstr := part[1 : len(part)-2]
			idx, _ := strconv.Atoi(idxstr)

			if ar, ok := obj.([]interface{}); ok {
				if idx >= len(ar) {
					return nil, fmt.Errorf("Index out of bounds: %s", strings.Join(parts[0:i+1], "."))
				}

				obj = ar[idx]
			} else {
				return nil, fmt.Errorf("Path %s is not an array", strings.Join(parts[0:i], "."))
			}
		} else if part == "<len>" {
			if ar, ok := obj.([]interface{}); ok {
				return len(ar), nil
			}
		} else {
			// Map
			if dict, ok := obj.(jsonMessage); ok {
				if val, ok := dict[part]; ok {
					obj = val
				} else {
					return nil, fmt.Errorf("Key %s not found in %s", part, strings.Join(parts[0:i], "."))
				}
			} else if dict, ok := obj.(map[string]interface{}); ok {
				if val, ok := dict[part]; ok {
					obj = val
				} else {
					return nil, fmt.Errorf("Key %s not found in %s", part, strings.Join(parts[0:i], "."))
				}
			} else {
				return nil, fmt.Errorf("Path %s is not a map", strings.Join(parts[0:i], "."))
			}
		}
	}

	return obj, nil
}
