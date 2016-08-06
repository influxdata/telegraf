package particle

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/influxdata/telegraf/testutil"
)

func ParticleWebhookRequest(event string, urlEncodedString string, t *testing.T) {
	var acc testutil.Accumulator
	pwh := &ParticleWebhook{Path: "/particle", acc: &acc}
	req, _ := http.NewRequest("POST", "/particle", bytes.NewBufferString(urlEncodedString))
	w := httptest.NewRecorder()
	pwh.eventHandler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("POST "+event+" returned HTTP status code %v.\nExpected %v", w.Code, http.StatusOK)
	}
}

func TestParticleEvent(t *testing.T) {
	ParticleWebhookRequest("particle_event", NewEventURLEncoded(), t)
}
