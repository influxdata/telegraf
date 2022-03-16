package logzio

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	testToken = "123456789"
	testURL   = "https://logzio.com"
)

func TestConnetWithoutToken(t *testing.T) {
	l := &Logzio{
		URL: testURL,
		Log: testutil.Logger{},
	}
	err := l.Connect()
	require.Error(t, err)
}

func TestParseMetric(t *testing.T) {
	l := &Logzio{}
	for _, tm := range testutil.MockMetrics() {
		lm := l.parseMetric(tm)
		require.Equal(t, tm.Fields(), lm.Metric[tm.Name()])
		require.Equal(t, logzioType, lm.Type)
		require.Equal(t, tm.Tags(), lm.Dimensions)
		require.Equal(t, tm.Time(), lm.Time)
	}
}

func TestBadStatusCode(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	l := &Logzio{
		Token: testToken,
		URL:   ts.URL,
		Log:   testutil.Logger{},
	}

	err := l.Connect()
	require.NoError(t, err)

	err = l.Write(testutil.MockMetrics())
	require.Error(t, err)
}

func TestWrite(t *testing.T) {
	tm := testutil.TestMetric(float64(3.14), "test1")
	var body bytes.Buffer
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gz, err := gzip.NewReader(r.Body)
		require.NoError(t, err)

		_, err = io.Copy(&body, gz)
		require.NoError(t, err)

		var lm Metric
		err = json.Unmarshal(body.Bytes(), &lm)
		require.NoError(t, err)

		require.Equal(t, tm.Fields(), lm.Metric[tm.Name()])
		require.Equal(t, logzioType, lm.Type)
		require.Equal(t, tm.Tags(), lm.Dimensions)
		require.Equal(t, tm.Time(), lm.Time)

		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	l := &Logzio{
		Token: testToken,
		URL:   ts.URL,
		Log:   testutil.Logger{},
	}

	err := l.Connect()
	require.NoError(t, err)

	err = l.Write([]telegraf.Metric{tm})
	require.NoError(t, err)
}
