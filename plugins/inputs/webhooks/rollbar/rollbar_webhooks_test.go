package rollbar

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/influxdata/telegraf/testutil"
)

func postWebhooks(rb *RollbarWebhook, eventBody string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest("POST", "/", strings.NewReader(eventBody))
	w := httptest.NewRecorder()
	w.Code = 500

	rb.eventHandler(w, req)

	return w
}

func TestNewItem(t *testing.T) {
	var acc testutil.Accumulator
	rb := &RollbarWebhook{Path: "/rollbar", acc: &acc}
	resp := postWebhooks(rb, NewItemJSON())
	if resp.Code != http.StatusOK {
		t.Errorf("POST new_item returned HTTP status code %v.\nExpected %v", resp.Code, http.StatusOK)
	}

	fields := map[string]interface{}{
		"id": 272716944,
	}

	tags := map[string]string{
		"event":       "new_item",
		"environment": "production",
		"project_id":  "90",
		"language":    "python",
		"level":       "error",
	}

	acc.AssertContainsTaggedFields(t, "rollbar_webhooks", fields, tags)
}

func TestOccurrence(t *testing.T) {
	var acc testutil.Accumulator
	rb := &RollbarWebhook{Path: "/rollbar", acc: &acc}
	resp := postWebhooks(rb, OccurrenceJSON())
	if resp.Code != http.StatusOK {
		t.Errorf("POST occurrence returned HTTP status code %v.\nExpected %v", resp.Code, http.StatusOK)
	}

	fields := map[string]interface{}{
		"id": 402860571,
	}

	tags := map[string]string{
		"event":       "occurrence",
		"environment": "production",
		"project_id":  "78234",
		"language":    "php",
		"level":       "error",
	}

	acc.AssertContainsTaggedFields(t, "rollbar_webhooks", fields, tags)
}

func TestDeploy(t *testing.T) {
	var acc testutil.Accumulator
	rb := &RollbarWebhook{Path: "/rollbar", acc: &acc}
	resp := postWebhooks(rb, DeployJSON())
	if resp.Code != http.StatusOK {
		t.Errorf("POST deploy returned HTTP status code %v.\nExpected %v", resp.Code, http.StatusOK)
	}

	fields := map[string]interface{}{
		"id": 187585,
	}

	tags := map[string]string{
		"event":       "deploy",
		"environment": "production",
		"project_id":  "90",
	}

	acc.AssertContainsTaggedFields(t, "rollbar_webhooks", fields, tags)
}

func TestUnknowItem(t *testing.T) {
	rb := &RollbarWebhook{Path: "/rollbar"}
	resp := postWebhooks(rb, UnknowJSON())
	if resp.Code != http.StatusOK {
		t.Errorf("POST unknow returned HTTP status code %v.\nExpected %v", resp.Code, http.StatusOK)
	}
}
