package mandrill

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/influxdata/telegraf/testutil"
)

func postWebhooks(md *MandrillWebhook, eventBody string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest("POST", "/mandrill", strings.NewReader(eventBody))
	w := httptest.NewRecorder()
	w.Code = 500

	md.eventHandler(w, req)

	return w
}

func TestSendEvent(t *testing.T) {
	var acc testutil.Accumulator
	md := &MandrillWebhook{Path: "/mandrill", acc: &acc}
	resp := postWebhooks(md, "["+ SendEventJSON() +"]")
	if resp.Code != http.StatusOK {
		t.Errorf("POST new_item returned HTTP status code %v.\nExpected %v", resp.Code, http.StatusOK)
	}

	fields := map[string]interface{}{
		"id":"1",
	}

	tags := map[string]string{
		"event": "send",
	}

	acc.AssertContainsTaggedFields(t, "mandrill_webhooks", fields, tags)
}


func TestMultipleEvents(t *testing.T) {
	var acc testutil.Accumulator
	md := &MandrillWebhook{Path: "/mandrill", acc: &acc}
	resp := postWebhooks(md, "["+ SendEventJSON() +","+ HardBounceEventJSON() +"]")
	if resp.Code != http.StatusOK {
		t.Errorf("POST new_item returned HTTP status code %v.\nExpected %v", resp.Code, http.StatusOK)
	}

	fields := map[string]interface{}{
		"id":"1",
	}

	tags := map[string]string{
		"event": "send",
	}

	acc.AssertContainsTaggedFields(t, "mandrill_webhooks", fields, tags)

	fields = map[string]interface{}{
		"id":"1",
	}

	tags = map[string]string{
		"event": "hard_bounce",
	}
	acc.AssertContainsTaggedFields(t, "mandrill_webhooks", fields, tags)
}
