package application_insights

import (
	"github.com/microsoft/ApplicationInsights-Go/appinsights"
)

type diagnosticsMessageSubscriber struct {
}

func (ms diagnosticsMessageSubscriber) Subscribe(handler appinsights.DiagnosticsMessageHandler) appinsights.DiagnosticsMessageListener {
	return appinsights.NewDiagnosticsMessageListener(handler)
}
