package plex

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/influxdata/telegraf/testutil"
)

func PlexWebhookRequest(jsonString string, t *testing.T) {
	var acc testutil.Accumulator
	gh := &PlexWebhook{Path: "/plex", acc: &acc}
	req, _ := http.NewRequest("POST", "/plex", strings.NewReader(jsonString))
	w := httptest.NewRecorder()
	gh.eventHandler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("POST returned HTTP status code %v.\nExpected %v", w.Code, http.StatusOK)
	}
}

func TestPlexEvent(t *testing.T) {
	PlexWebhookRequest(PlexWebhookEventJSON(), t)
}
