package rollbar_webhooks

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/influxdata/telegraf/testutil"
)

func postWebhooks(rb *RollbarWebhooks, eventBody string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest("POST", "/", strings.NewReader(eventBody))
	w := httptest.NewRecorder()

	rb.eventHandler(w, req)

	return w
}

func TestNewItem(t *testing.T) {
	rb := NewRollbarWebhooks()
	resp := postWebhooks(rb, NewItemJSON())
	if resp.Code != http.StatusOK {
		t.Errorf("POST new_item returned HTTP status code %v.\nExpected %v", resp.Code, http.StatusOK)
	}
}

func TestGather(t *testing.T) {
	var acc testutil.Accumulator
	rb := NewRollbarWebhooks()

	postWebhooks(rb, NewItemJSON())
	rb.Gather(&acc)

	fields := map[string]interface{}{
		"id": 272716944,
	}

	tags := map[string]string{
		"event":       "new_item",
		"environment": "production",
		"project_id":  "90",
		"language":    "python",
	}

	acc.AssertContainsTaggedFields(t, "rollbar_webhooks", fields, tags)
}
