package splunk

import (
	"encoding/json"
	"github.com/influxdata/telegraf/testutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

// default config used by Tests
// AuthString == echo -n "uid:pwd" | base64
func defaultSplunk() *Splunk {
	return &Splunk{
		Prefix:          "splunk.metrics.test",
		Source:          "",
		SplunkUrl:       "http://localhost:8088/services/collector",
		AuthString:      "XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX",
		SimpleFields:    false,
		MetricSeparator: ".",
		ConvertPaths:    true,
		ConvertBool:     true,
		UseRegex:        false,
	}
}

func TestSplunk(t *testing.T) {
	s := defaultSplunk()

	// -----------------------------------------------------------------------------------------------------------------
	//  Create a Fake Server to send Splunk formatted metrics to  (reset the SplunkUrl from above)
	// -----------------------------------------------------------------------------------------------------------------
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(`{"status":"ok"}`)
	}))
	defer ts.Close()

	// Check for existence of s.AuthString to prevent race ('panic: runtime error: invalid memory address or nil pointer dereference')
	if s != nil {
		s.SplunkUrl = ts.URL

		// -----------------------------------------------------------------------------------------------------------------
		//  Call the Write method with a test metric to ensure parsing works correctly.
		// -----------------------------------------------------------------------------------------------------------------
		s.Write(testutil.MockMetrics())
	}

	return
}
