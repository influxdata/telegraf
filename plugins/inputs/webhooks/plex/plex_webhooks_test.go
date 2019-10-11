package plex

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/influxdata/telegraf/testutil"
)

func PlexWebhookRequest(jsonString string, t *testing.T) {
	var acc testutil.Accumulator
	p := &PlexWebhook{Path: "/plex", acc: &acc}
	values := map[string]io.Reader{
		"payload": strings.NewReader(jsonString),
	}
	writer, b, _ := BuildFormData(values)
	req, _ := http.NewRequest("POST", "/plex", &b)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	p.eventHandler(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("POST returned HTTP status code %v.\nExpected %v", w.Code, http.StatusOK)
	}
}

func TestPlexEvent(t *testing.T) {
	PlexWebhookRequest(PlexWebhookEventJSON(), t)
}

func BuildFormData(values map[string]io.Reader) (w *multipart.Writer, b bytes.Buffer, err error) {
	// Prepare a form that you will submit to that URL.
	w = multipart.NewWriter(&b)
	for key, r := range values {
		var fw io.Writer
		if x, ok := r.(io.Closer); ok {
			defer x.Close()
		}
		// Add other fields
		fw, err = w.CreateFormField(key)
		if err != nil {
			return nil, b, err
		}
		_, err = io.Copy(fw, r)
		if err != nil {
			return nil, b, err
		}

	}
	w.Close()
	return w, b, nil
}
