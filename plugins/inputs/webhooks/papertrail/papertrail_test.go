package papertrail

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

const (
	contentType = "application/x-www-form-urlencoded"
)

func post(pt *PapertrailWebhook, contentType string, body string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest("POST", "/", strings.NewReader(body))
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()
	pt.eventHandler(w, req)
	return w
}

func TestWrongContentType(t *testing.T) {
	var acc testutil.Accumulator
	pt := &PapertrailWebhook{Path: "/papertrail", acc: &acc}
	form := url.Values{}
	form.Set("payload", sampleEventPayload)
	data := form.Encode()

	resp := post(pt, "", data)
	require.Equal(t, http.StatusUnsupportedMediaType, resp.Code)
}

func TestMissingPayload(t *testing.T) {
	var acc testutil.Accumulator
	pt := &PapertrailWebhook{Path: "/papertrail", acc: &acc}

	resp := post(pt, contentType, "")
	require.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestPayloadNotJSON(t *testing.T) {
	var acc testutil.Accumulator
	pt := &PapertrailWebhook{Path: "/papertrail", acc: &acc}

	resp := post(pt, contentType, "payload={asdf]")
	require.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestPayloadInvalidJSON(t *testing.T) {
	var acc testutil.Accumulator
	pt := &PapertrailWebhook{Path: "/papertrail", acc: &acc}

	resp := post(pt, contentType, `payload={"value": 42}`)
	require.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestEventPayload(t *testing.T) {
	var acc testutil.Accumulator
	pt := &PapertrailWebhook{Path: "/papertrail", acc: &acc}

	form := url.Values{}
	form.Set("payload", sampleEventPayload)
	resp := post(pt, contentType, form.Encode())
	require.Equal(t, http.StatusOK, resp.Code)

	fields := map[string]interface{}{
		"count": uint64(1),
	}

	tags1 := map[string]string{
		"event": "Important stuff",
		"host":  "abc",
	}
	tags2 := map[string]string{
		"event": "Important stuff",
		"host":  "def",
	}

	acc.AssertContainsTaggedFields(t, "papertrail", fields, tags1)
	acc.AssertContainsTaggedFields(t, "papertrail", fields, tags2)
}

func TestCountPayload(t *testing.T) {
	var acc testutil.Accumulator
	pt := &PapertrailWebhook{Path: "/papertrail", acc: &acc}
	form := url.Values{}
	form.Set("payload", sampleCountPayload)
	resp := post(pt, contentType, form.Encode())
	require.Equal(t, http.StatusOK, resp.Code)

	fields1 := map[string]interface{}{
		"count": uint64(5),
	}
	fields2 := map[string]interface{}{
		"count": uint64(3),
	}

	tags1 := map[string]string{
		"event": "Important stuff",
		"host":  "arthur",
	}
	tags2 := map[string]string{
		"event": "Important stuff",
		"host":  "ford",
	}

	acc.AssertContainsTaggedFields(t, "papertrail", fields1, tags1)
	acc.AssertContainsTaggedFields(t, "papertrail", fields2, tags2)
}

const sampleEventPayload = `{
  "events": [
    {
      "id": 7711561783320576,
      "received_at": "2011-05-18T20:30:02-07:00",
      "display_received_at": "May 18 20:30:02",
      "source_ip": "208.75.57.121",
      "source_name": "abc",
      "source_id": 2,
      "hostname": "abc",
      "program": "CROND",
      "severity": "Info",
      "facility": "Cron",
      "message": "message body"
    },
    {
      "id": 7711562567655424,
      "received_at": "2011-05-18T20:30:02-07:00",
      "display_received_at": "May 18 20:30:02",
      "source_ip": "208.75.57.120",
      "source_name": "server1",
      "source_id": 19,
      "hostname": "def",
      "program": "CROND",
      "severity": "Info",
      "facility": "Cron",
      "message": "A short event"
    }
  ],
  "saved_search": {
    "id": 42,
    "name": "Important stuff",
    "query": "cron OR server1",
    "html_edit_url": "https://papertrailapp.com/searches/42/edit",
    "html_search_url": "https://papertrailapp.com/searches/42"
  },
  "max_id": "7711582041804800",
  "min_id": "7711561783320576"
}`

const sampleCountPayload = `{
   "counts": [
     {
       "source_name": "arthur",
       "source_id": 4,
       "timeseries": {
         "1453248895": 5
       }
     },
     {
       "source_name": "ford",
       "source_id": 3,
       "timeseries": {
         "1453248927": 3
       }
     }
   ],
   "saved_search": {
     "id": 42,
     "name": "Important stuff",
     "query": "cron OR server1",
     "html_edit_url": "https://papertrailapp.com/searches/42/edit",
     "html_search_url": "https://papertrailapp.com/searches/42"
   },
   "max_id": "7711582041804800",
   "min_id": "7711561783320576"
}`
