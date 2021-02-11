package bigquery

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"strings"
	"testing"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/option"
)

const (
	successfulResponse = "{\"kind\": \"bigquery#tableDataInsertAllResponse\"}"
)

var testingHost string
var testDuration = internal.Duration{Duration: 5 * time.Second}

func TestConnect(t *testing.T) {
	srv := localBigQueryServer(t)
	testingHost = strings.ReplaceAll(srv.URL, "http://", "")
	defer srv.Close()

	b := &BigQuery{
		Project: "test-project",
		Dataset: "test-dataset",
		Timeout: testDuration,
	}

	b.setUpTestClient()
	err := b.Connect()
	require.NoError(t, err)
}

func TestWrite(t *testing.T) {
	srv := localBigQueryServer(t)
	testingHost = strings.ReplaceAll(srv.URL, "http://", "")
	defer srv.Close()

	b := &BigQuery{
		Project: "test-project",
		Dataset: "test-dataset",
		Timeout: testDuration,
	}

	mockMetrics := testutil.MockMetrics()
	b.setUpTestClient()
	b.Connect()
	err := b.Write(mockMetrics)
	require.NoError(t, err)
}

func (b *BigQuery) setUpTestClient() error {
	noAuth := option.WithoutAuthentication()
	endpoints := option.WithEndpoint("http://" + testingHost)

	ctx := context.Background()

	c, err := bigquery.NewClient(ctx, b.Project, noAuth, endpoints)

	if err != nil {
		return err
	}

	b.client = c

	return nil
}

func localBigQueryServer(t *testing.T) *httptest.Server {
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
