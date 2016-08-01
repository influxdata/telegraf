package particle

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/influxdata/telegraf/testutil"
)

func ParticleWebhookRequest(urlEncodedString string, t *testing.T) {
	var acc testutil.Accumulator
	pwh := &ParticleWebhook{Path: "/particle", acc: &acc}
	req, _ := http.NewRequest("POST", "/particle", urlEncodedString)
	w := httptest.NewRecorder()
	pwh.eventHandler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("POST "+event+" returned HTTP status code %v.\nExpected %v", w.Code, http.StatusOK)
	}
}

func TestNewEvent(t *testing.T) {
	ParticleWebhookRequest(NewEventURLEncoded())
}
