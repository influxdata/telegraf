package splunk

import (
	"encoding/json"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func fakeSplunk() *Splunk {
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
	s := fakeSplunk()

	// -----------------------------------------------------------------------------------------------------------------
	//  Create a Fake Server to send Splunk formatted metrics to
	// -----------------------------------------------------------------------------------------------------------------
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}
		log.Println(string(body))

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(`{"status":"ok"}`)
	}))
	defer ts.Close()

	// -----------------------------------------------------------------------------------------------------------------
	//  Set the SplunkUrl from above and process MockMetrics
	// -----------------------------------------------------------------------------------------------------------------
	s.SplunkUrl = ts.URL
	err := s.Connect()
	require.NoError(t, err)
	err = s.Write(testutil.MockMetrics())
	require.NoError(t, err)
}
