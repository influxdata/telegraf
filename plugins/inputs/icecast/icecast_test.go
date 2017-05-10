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
<?xml version="1.0" encoding="UTF-8"?>
<icestats><source mount="/mount.aac"><fallback/><listeners>420</listeners><Connected>806794</Connected><content-type>audio/aacp</content-type></source></icestats>
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
	require.NoError(t, acc.GatherError(a.Gather))

	// Report total listeners as well
	tags := map[string]string{
		"host":  string("localhost"),
		"mount": string("/mount.aac"),
	}
	fields := map[string]interface{}{
		"listeners": int32(420),
	}
	acc.AssertContainsTaggedFields(t, "icecast", fields, tags)
}
