package appinsights

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
	"testing"
	"time"
)

func BenchmarkClientBurstPerformance(b *testing.B) {
	client := NewTelemetryClient("")
	client.(*telemetryClient).channel.(*InMemoryChannel).transmitter = &nullTransmitter{}

	for i := 0; i < b.N; i++ {
		client.TrackTrace("A message", Information)
	}

	<-client.Channel().Close(time.Minute)
}

func TestClientProperties(t *testing.T) {
	client := NewTelemetryClient(test_ikey)
	defer client.Channel().Close()

	if _, ok := client.Channel().(*InMemoryChannel); !ok {
		t.Error("Client's Channel() is not InMemoryChannel")
	}

	if ikey := client.InstrumentationKey(); ikey != test_ikey {
		t.Error("Client's InstrumentationKey is not expected")
	}

	if ikey := client.Context().InstrumentationKey(); ikey != test_ikey {
		t.Error("Context's InstrumentationKey is not expected")
	}

	if client.Context() == nil {
		t.Error("Client.Context == nil")
	}

	if client.IsEnabled() == false {
		t.Error("Client.IsEnabled == false")
	}

	client.SetIsEnabled(false)
	if client.IsEnabled() == true {
		t.Error("Client.SetIsEnabled had no effect")
	}

	if client.Channel().EndpointAddress() != "https://dc.services.visualstudio.com/v2/track" {
		t.Error("Client.Channel.EndpointAddress was incorrect")
	}
}

func TestEndToEnd(t *testing.T) {
	mockClock(time.Unix(1511001321, 0))
	defer resetClock()
	xmit, server := newTestClientServer()
	defer server.Close()

	config := NewTelemetryConfiguration(test_ikey)
	config.EndpointUrl = xmit.(*httpTransmitter).endpoint
	client := NewTelemetryClientFromConfig(config)
	defer client.Channel().Close()

	// Track directly off the client
	client.TrackEvent("client-event")
	client.TrackMetric("client-metric", 44.0)
	client.TrackTrace("client-trace", Information)
	client.TrackRequest("GET", "www.testurl.org", time.Minute, "404")

	// NOTE: A lot of this is covered elsewhere, so we won't duplicate
	// *too* much.

	// Set up server response
	server.responseData = []byte(`{"itemsReceived":4, "itemsAccepted":4, "errors":[]}`)
	server.responseHeaders["Content-type"] = "application/json"

	// Wait for automatic transmit -- get the request
	slowTick(11)
	req := server.waitForRequest(t)

	// GZIP magic number
	if len(req.body) < 2 || req.body[0] != 0x1f || req.body[1] != 0x8b {
		t.Fatal("Missing gzip magic number")
	}

	// Decompress
	reader, err := gzip.NewReader(bytes.NewReader(req.body))
	if err != nil {
		t.Fatalf("Coudln't create gzip reader: %s", err.Error())
	}

	// Read payload
	body, err := ioutil.ReadAll(reader)
	reader.Close()
	if err != nil {
		t.Fatalf("Couldn't read compressed data: %s", err.Error())
	}

	// Check out payload
	j, err := parsePayload(body)
	if err != nil {
		t.Errorf("Error parsing payload: %s", err.Error())
	}

	if len(j) != 4 {
		t.Fatal("Unexpected event count")
	}

	j[0].assertPath(t, "iKey", test_ikey)
	j[0].assertPath(t, "name", "Microsoft.ApplicationInsights.01234567000089abcdef000000000000.Event")
	j[0].assertPath(t, "time", "2017-11-18T10:35:21Z")

	j[1].assertPath(t, "iKey", test_ikey)
	j[1].assertPath(t, "name", "Microsoft.ApplicationInsights.01234567000089abcdef000000000000.Metric")
	j[1].assertPath(t, "time", "2017-11-18T10:35:21Z")

	j[2].assertPath(t, "iKey", test_ikey)
	j[2].assertPath(t, "name", "Microsoft.ApplicationInsights.01234567000089abcdef000000000000.Message")
	j[2].assertPath(t, "time", "2017-11-18T10:35:21Z")

	j[3].assertPath(t, "iKey", test_ikey)
	j[3].assertPath(t, "name", "Microsoft.ApplicationInsights.01234567000089abcdef000000000000.Request")
	j[3].assertPath(t, "time", "2017-11-18T10:34:21Z")
}
