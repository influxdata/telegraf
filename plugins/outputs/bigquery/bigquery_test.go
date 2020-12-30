package bigquery

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"os"
	"strings"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

const (
	successfulResponse = "{\"kind\": \"bigquery#tableDataInsertAllResponse\"}"
)

func TestMain(t *testing.M) {
	srv := localBigQueryServer(t)
	os.Setenv("BIGQUERY_TESTING_HOST", strings.ReplaceAll(srv.URL, "http://", ""))

	defer srv.Close()

	os.Exit(t.Run())
}

func TestConnect(t *testing.T) {
	b := &BigQuery{
		Project: "test-project",
		Dataset: "test-dataset",
	}

	err := b.Connect()
	require.NoError(t, err)

}

func TestWrite(t *testing.T) {
	b := &BigQuery{
		Project: "test-project",
		Dataset: "test-dataset",
	}

	mockMetrics := testutil.MockMetrics()
	b.Connect()
	err := b.Write(mockMetrics)
	require.NoError(t, err)
}

func localBigQueryServer(t *testing.M) *httptest.Server {
	srv := httptest.NewServer(http.NotFoundHandler())

	srv.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/projects/test-project/datasets/test-dataset/tables/test1/insertAll":
			requestDump, err := httputil.DumpRequest(r, true)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println(string(requestDump))
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(successfulResponse))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})

	return srv
}
