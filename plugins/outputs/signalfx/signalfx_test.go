package signalfx

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/influxdata/telegraf/testutil"

	"github.com/stretchr/testify/require"
)

func TestHTTPSignalFx(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		// https://developers.signalfx.com/docs/datapoint
		fmt.Fprintln(w, `"OK"`)
	}))
	defer ts.Close()

	i := SignalFx{
		AuthToken: "Whatever",
		Endpoint:  ts.URL,
	}

	err := i.Connect()
	require.NoError(t, err)
	err = i.Write(testutil.MockMetrics())
	require.NoError(t, err)
}
