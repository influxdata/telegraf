package dockerhub

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/influxdata/telegraf/testutil"
)

func DockerhubWebhookRequest(event string, jsonString string, t *testing.T) {
	var acc testutil.Accumulator
	dhwh := &DockerhubWebhook{Path: "/dockerhub", acc: &acc}
	req, _ := http.NewRequest("POST", "/dockerhub", strings.NewReader(jsonString))
	w := httptest.NewRecorder()
	dhwh.eventHandler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf(
			"POST"+event+" returned HTTP status code %v.\nExpected %v",
			w.Code,
			http.StatusOK)
	}
}

func TestNewEvent(t *testing.T) {
	DockerhubWebhookRequest("dockerhub_event", NewEventJSONEncoded(), t)
}
