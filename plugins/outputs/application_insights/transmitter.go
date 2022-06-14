package application_insights

import (
	"github.com/microsoft/ApplicationInsights-Go/appinsights"
)

type Transmitter struct {
	client appinsights.TelemetryClient
}

func NewTransmitter(ikey string, endpointURL string) *Transmitter {
	if len(endpointURL) == 0 {
		return &Transmitter{client: appinsights.NewTelemetryClient(ikey)}
	}

	telemetryConfig := appinsights.NewTelemetryConfiguration(ikey)
	telemetryConfig.EndpointUrl = endpointURL
	return &Transmitter{client: appinsights.NewTelemetryClientFromConfig(telemetryConfig)}
}

func (t *Transmitter) Track(telemetry appinsights.Telemetry) {
	t.client.Track(telemetry)
}

func (t *Transmitter) Close() <-chan struct{} {
	return t.client.Channel().Close(0)
}
