package icecast

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

var icecastStatus = `
<icestats>
	<source mount="/mount.aac">
		<fallback>/mount-fallback.aac</fallback>
		<listeners>420</listeners>
		<Connected>806794</Connected>
		<content-type>audio/aacp</content-type>
	</source>
</icestats>
`

func TestHTTPicecast(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, icecastStatus)
	}))
	defer ts.Close()

	a := Icecast{
		// Fetch it 2 times to catch possible data races.
		Urls: []string{ts.URL, ts.URL},
	}

	var acc testutil.Accumulator
	err := a.Gather(&acc)
	require.NoError(t, err)

	fields := map[string]interface{}{
		"Mount":       string("/mount.aac"),
		"Listeners":   int32(420),
		"Connected":   int32(806794),
		"ContentType": string("audio/aacp"),
	}
	acc.AssertContainsFields(t, "icecast", fields)
}
