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

	var rows []map[string]json.RawMessage
	json.Unmarshal(receivedBody["rows"], &rows)

	var row Row
	json.Unmarshal(rows[0]["json"], &row)

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
			decoder.Decode(&receivedBody)

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(successfulResponse))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})

	return srv
}
