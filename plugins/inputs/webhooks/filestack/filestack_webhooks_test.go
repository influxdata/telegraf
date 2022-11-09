package filestack

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/testutil"
)

func postWebhooks(md *FilestackWebhook, eventBodyFile io.Reader) *httptest.ResponseRecorder {
	req, _ := http.NewRequest("POST", "/filestack", eventBodyFile)
	w := httptest.NewRecorder()

	md.eventHandler(w, req)

	return w
}

func TestDialogEvent(t *testing.T) {
	var acc testutil.Accumulator
	fs := &FilestackWebhook{Path: "/filestack", acc: &acc}
	resp := postWebhooks(fs, getFile(t, "testdata/dialog_open.json"))
	if resp.Code != http.StatusOK {
		t.Errorf("POST returned HTTP status code %v.\nExpected %v", resp.Code, http.StatusOK)
	}

	fields := map[string]interface{}{
		"id": "102",
	}

	tags := map[string]string{
		"action": "fp.dialog",
	}

	acc.AssertContainsTaggedFields(t, "filestack_webhooks", fields, tags)
}

func TestParseError(t *testing.T) {
	fs := &FilestackWebhook{Path: "/filestack"}
	resp := postWebhooks(fs, strings.NewReader(""))
	if resp.Code != http.StatusBadRequest {
		t.Errorf("POST returned HTTP status code %v.\nExpected %v", resp.Code, http.StatusBadRequest)
	}
}

func TestUploadEvent(t *testing.T) {
	var acc testutil.Accumulator
	fs := &FilestackWebhook{Path: "/filestack", acc: &acc}
	resp := postWebhooks(fs, getFile(t, "testdata/upload.json"))
	if resp.Code != http.StatusOK {
		t.Errorf("POST returned HTTP status code %v.\nExpected %v", resp.Code, http.StatusOK)
	}

	fields := map[string]interface{}{
		"id": "100946",
	}

	tags := map[string]string{
		"action": "fp.upload",
	}

	acc.AssertContainsTaggedFields(t, "filestack_webhooks", fields, tags)
}

func TestVideoConversionEvent(t *testing.T) {
	var acc testutil.Accumulator
	fs := &FilestackWebhook{Path: "/filestack", acc: &acc}
	resp := postWebhooks(fs, getFile(t, "testdata/video_conversion.json"))
	if resp.Code != http.StatusBadRequest {
		t.Errorf("POST returned HTTP status code %v.\nExpected %v", resp.Code, http.StatusBadRequest)
	}
}

func getFile(t *testing.T, filePath string) io.Reader {
	file, err := os.Open(filePath)
	require.NoErrorf(t, err, "could not read from file %s", filePath)

	return file
}
