package application_insights

import "github.com/Microsoft/ApplicationInsights-Go/appinsights"

type Transmitter struct {
	client appinsights.TelemetryClient
}

func NewTransmitter(ikey string) *Transmitter {
	return &Transmitter{client: appinsights.NewTelemetryClient(ikey)}
}

func (t *Transmitter) Track(telemetry appinsights.Telemetry) {
	t.client.Track(telemetry)
}

func (t *Transmitter) Close() <-chan struct{} {
	return t.client.Channel().Close(0)
}
