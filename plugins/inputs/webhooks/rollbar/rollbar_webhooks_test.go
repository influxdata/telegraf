package rollbar

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/testutil"
)

func postWebhooks(t *testing.T, rb *Webhook, eventBody string) *httptest.ResponseRecorder {
	req, err := http.NewRequest("POST", "/", strings.NewReader(eventBody))
	require.NoError(t, err)
	w := httptest.NewRecorder()
	w.Code = 500

	rb.eventHandler(w, req)

	return w
}

func TestNewItem(t *testing.T) {
	var acc testutil.Accumulator
	rb := &Webhook{Path: "/rollbar", acc: &acc}
	resp := postWebhooks(t, rb, newItemJSON())
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
	rb := &Webhook{Path: "/rollbar", acc: &acc}
	resp := postWebhooks(t, rb, occurrenceJSON())
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
	rb := &Webhook{Path: "/rollbar", acc: &acc}
	resp := postWebhooks(t, rb, deployJSON())
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
	rb := &Webhook{Path: "/rollbar"}
	resp := postWebhooks(t, rb, unknownJSON())
	if resp.Code != http.StatusOK {
		t.Errorf("POST unknow returned HTTP status code %v.\nExpected %v", resp.Code, http.StatusOK)
	}
}
