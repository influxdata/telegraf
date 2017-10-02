package particle

import (
	"github.com/influxdata/telegraf/testutil"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func postWebhooks(rb *ParticleWebhook, eventBody string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest("POST", "/", strings.NewReader(eventBody))
	log.Printf("eventBody: %s\n", eventBody)
	w := httptest.NewRecorder()
	w.Code = 500

	rb.eventHandler(w, req)

	return w
}

func TestNewItem(t *testing.T) {
	var acc testutil.Accumulator
	rb := &ParticleWebhook{Path: "/particle", acc: &acc}
	resp := postWebhooks(rb, NewItemJSON())
	log.Printf("Respnse: %s\n", resp.Body)
	if resp.Code != http.StatusOK {
		t.Errorf("POST new_item returned HTTP status code %v.\nExpected %v", resp.Code, http.StatusOK)
	}

	fields := map[string]interface{}{
		"temp_c":    26.680000,
		"temp_f":    80.024001,
		"humidity":  44.937500,
		"pressure":  998.998901,
		"altitude":  119.331436,
		"broadband": 1266,
		"infrared":  528,
		"lux":       0,
	}

	tags := map[string]string{
		"id":       "230035001147343438323536",
		"location": "TravelingWilbury",
	}

	acc.AssertContainsTaggedFields(t, "particle_webhooks", fields, tags)
}
func TestUnknowItem(t *testing.T) {
	rb := &ParticleWebhook{Path: "/particle"}
	resp := postWebhooks(rb, UnknowJSON())
	log.Printf("Response: %s\n", resp.Body)
	if resp.Code != http.StatusOK {
		t.Errorf("POST unknown returned HTTP status code %v.\nExpected %v", resp.Code, http.StatusOK)
	}
}
