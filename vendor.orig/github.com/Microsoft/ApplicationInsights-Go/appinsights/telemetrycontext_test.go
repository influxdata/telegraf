package appinsights

import (
	"strings"
	"testing"
	"time"

	"github.com/Microsoft/ApplicationInsights-Go/appinsights/contracts"
)

func TestDefaultTags(t *testing.T) {
	context := NewTelemetryContext(test_ikey)
	context.Tags["test"] = "OK"
	context.Tags["no-write"] = "Fail"

	telem := NewTraceTelemetry("Hello world.", Verbose)
	telem.Tags["no-write"] = "OK"

	envelope := context.envelop(telem)

	if envelope.Tags["test"] != "OK" {
		t.Error("Default client tags did not propagate to telemetry")
	}

	if envelope.Tags["no-write"] != "OK" {
		t.Error("Default client tag overwrote telemetry item tag")
	}
}

func TestCommonProperties(t *testing.T) {
	context := NewTelemetryContext(test_ikey)
	context.CommonProperties = map[string]string{
		"test":     "OK",
		"no-write": "Fail",
	}

	telem := NewTraceTelemetry("Hello world.", Verbose)
	telem.Properties["no-write"] = "OK"

	envelope := context.envelop(telem)
	data := envelope.Data.(*contracts.Data).BaseData.(*contracts.MessageData)

	if data.Properties["test"] != "OK" {
		t.Error("Common properties did not propagate to telemetry")
	}

	if data.Properties["no-write"] != "OK" {
		t.Error("Common properties overwrote telemetry properties")
	}
}

func TestContextTags(t *testing.T) {
	// Just a quick test to make sure it works.
	tags := make(contracts.ContextTags)
	if v := tags.Session().GetId(); v != "" {
		t.Error("Failed to get empty session ID")
	}

	tags.Session().SetIsFirst("true")
	if v := tags.Session().GetIsFirst(); v != "true" {
		t.Error("Failed to get value")
	}

	if v, ok := tags["ai.session.isFirst"]; !ok || v != "true" {
		t.Error("Failed to get isFirst through raw map")
	}

	tags.Session().SetIsFirst("")
	if v, ok := tags["ai.session.isFirst"]; ok || v != "" {
		t.Error("SetIsFirst with empty string failed to remove it from the map")
	}
}

func TestSanitize(t *testing.T) {
	name := strings.Repeat("Z", 1024)
	val := strings.Repeat("Y", 10240)

	ev := NewEventTelemetry(name)
	ev.Properties[name] = val
	ev.Measurements[name] = 55.0

	ctx := NewTelemetryContext(test_ikey)
	ctx.Tags.Session().SetId(name)

	// We'll be looking for messages with these values:
	found := map[string]int{
		"EventData.Name exceeded":        0,
		"EventData.Properties has value": 0,
		"EventData.Properties has key":   0,
		"EventData.Measurements has key": 0,
		"ai.session.id exceeded":         0,
	}

	// Set up listener for the warnings.
	NewDiagnosticsMessageListener(func(msg string) error {
		for k, _ := range found {
			if strings.Contains(msg, k) {
				found[k] = found[k] + 1
				break
			}
		}

		return nil
	})

	defer resetDiagnosticsListeners()

	// This may break due to hardcoded limits... Check contracts.
	envelope := ctx.envelop(ev)

	// Make sure all the warnings were found in the output
	for k, v := range found {
		if v != 1 {
			t.Errorf("Did not find a warning containing \"%s\"", k)
		}
	}

	// Check the format of the stuff we found in the envelope
	if v, ok := envelope.Tags[contracts.SessionId]; !ok || v != name[:64] {
		t.Error("Session ID tag was not truncated")
	}

	evdata := envelope.Data.(*contracts.Data).BaseData.(*contracts.EventData)
	if evdata.Name != name[:512] {
		t.Error("Event name was not truncated")
	}

	if v, ok := evdata.Properties[name[:150]]; !ok || v != val[:8192] {
		t.Error("Event property name/value was not truncated")
	}

	if v, ok := evdata.Measurements[name[:150]]; !ok || v != 55.0 {
		t.Error("Event measurement name was not truncated")
	}
}

func TestTimestamp(t *testing.T) {
	ev := NewEventTelemetry("event")
	ev.Timestamp = time.Unix(1523667421, 500000000)

	envelope := NewTelemetryContext(test_ikey).envelop(ev)
	if envelope.Time != "2018-04-14T00:57:01.5Z" {
		t.Errorf("Unexpected timestamp: %s", envelope.Time)
	}
}
