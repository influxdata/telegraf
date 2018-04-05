package application_insights

import "github.com/Microsoft/ApplicationInsights-Go/appinsights"

type appinsightsTransmitter struct {
	client appinsights.TelemetryClient
}

func NewAppinsightsTransmitter(ikey string) *appinsightsTransmitter {
	return &appinsightsTransmitter{client: appinsights.NewTelemetryClient(ikey)}
}

func (t *appinsightsTransmitter) Track(telemetry appinsights.Telemetry) {
	t.client.Track(telemetry)
}

func (t *appinsightsTransmitter) Close() <-chan struct{} {
	return t.client.Channel().Close(0)
}
