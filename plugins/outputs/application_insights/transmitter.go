package application_insights

import (
	"github.com/Microsoft/ApplicationInsights-Go/appinsights"
)

type Transmitter struct {
	client appinsights.TelemetryClient
}

func NewTransmitter(ikey string, EndpointUrl string) *Transmitter {

	if len(EndpointUrl) == 0 {
		return &Transmitter{client: appinsights.NewTelemetryClient(ikey)}
	}

	telemetryConfig := appinsights.NewTelemetryConfiguration(ikey)
	telemetryConfig.EndpointUrl = EndpointUrl
	return &Transmitter{client: appinsights.NewTelemetryClientFromConfig(telemetryConfig)}
}

func (t *Transmitter) Track(telemetry appinsights.Telemetry) {
	t.client.Track(telemetry)
}

func (t *Transmitter) Close() <-chan struct{} {
	return t.client.Channel().Close(0)
}
