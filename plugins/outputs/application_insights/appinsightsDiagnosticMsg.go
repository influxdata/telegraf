package application_insights

import (
	"github.com/Microsoft/ApplicationInsights-Go/appinsights"
)

type appinsightsDiagnosticsMessageSubscriber struct {
}

func (ms appinsightsDiagnosticsMessageSubscriber) Subscribe(handler appinsights.DiagnosticsMessageHandler) appinsights.DiagnosticsMessageListener {
	return appinsights.NewDiagnosticsMessageListener(handler)
}
