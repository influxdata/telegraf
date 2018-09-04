package appinsights

import "testing"

func TestTelemetryConfiguration(t *testing.T) {
	testKey := "test"
	defaultEndpoint := "https://dc.services.visualstudio.com/v2/track"

	config := NewTelemetryConfiguration(testKey)

	if config.InstrumentationKey != testKey {
		t.Errorf("InstrumentationKey is %s, want %s", config.InstrumentationKey, testKey)
	}

	if config.EndpointUrl != defaultEndpoint {
		t.Errorf("EndpointUrl is %s, want %s", config.EndpointUrl, defaultEndpoint)
	}
}
