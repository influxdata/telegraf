package seq

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestConnectAndWrite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ts := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusCreated)
			}))
	defer ts.Close()

	r := &Seq{
		SeqInstance: ts.URL,
	}

	err := r.Connect()
	require.NoError(t, err)

	err = r.Write(testutil.MockMetrics())
	require.NoError(t, err)
}
