package bigquery

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/option"
)

const (
	successfulResponse = "{\"kind\": \"bigquery#tableDataInsertAllResponse\"}"
)

var testingHost string
var testDuration = config.Duration(5 * time.Second)
var receivedBody map[string]json.RawMessage

type Row struct {
	Tag1      string  `json:"tag1"`
	Timestamp string  `json:"timestamp"`
	Value     float64 `json:"value"`
}

func TestConnect(t *testing.T) {
	srv := localBigQueryServer(t)
	testingHost = strings.ReplaceAll(srv.URL, "http://", "")
	defer srv.Close()

	b := &BigQuery{
		Project: "test-project",
		Dataset: "test-dataset",
		Timeout: testDuration,
	}

	cerr := b.setUpTestClient()
	require.NoError(t, cerr)
	berr := b.Connect()
	require.NoError(t, berr)
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

	if err := b.setUpTestClient(); err != nil {
		require.NoError(t, err)
	}
	if err := b.Connect(); err != nil {
		require.NoError(t, err)
	}

	if err := b.Write(mockMetrics); err != nil {
		require.NoError(t, err)
	}

	var rows []map[string]json.RawMessage
	if err := json.Unmarshal(receivedBody["rows"], &rows); err != nil {
		require.NoError(t, err)
	}

	var row Row
	if err := json.Unmarshal(rows[0]["json"], &row); err != nil {
		require.NoError(t, err)
	}

	pt, _ := time.Parse(time.RFC3339, row.Timestamp)
	require.Equal(t, mockMetrics[0].Tags()["tag1"], row.Tag1)
	require.Equal(t, mockMetrics[0].Time(), pt)
	require.Equal(t, mockMetrics[0].Fields()["value"], row.Value)
}

func TestMetricToTableDefault(t *testing.T) {
	b := &BigQuery{
		Project:         "test-project",
		Dataset:         "test-dataset",
		Timeout:         testDuration,
		warnedOnHyphens: make(map[string]bool),
		ReplaceHyphenTo: "_",
		Log:             testutil.Logger{},
	}

	otn := "table-with-hyphens"
	ntn := b.metricToTable(otn)

	require.Equal(t, "table_with_hyphens", ntn)
	require.True(t, b.warnedOnHyphens[otn])
}

func TestMetricToTableCustom(t *testing.T) {
	log := testutil.Logger{}

	b := &BigQuery{
		Project:         "test-project",
		Dataset:         "test-dataset",
		Timeout:         testDuration,
		warnedOnHyphens: make(map[string]bool),
		ReplaceHyphenTo: "*",
		Log:             log,
	}

	otn := "table-with-hyphens"
	ntn := b.metricToTable(otn)

	require.Equal(t, "table*with*hyphens", ntn)
	require.True(t, b.warnedOnHyphens[otn])
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
			decoder := json.NewDecoder(r.Body)

			if err := decoder.Decode(&receivedBody); err != nil {
				require.NoError(t, err)
			}

			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte(successfulResponse)); err != nil {
				require.NoError(t, err)
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})

	return srv
}
