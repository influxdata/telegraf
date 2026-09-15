package fxmacrodata

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/testutil"
)

const indicatorBody = `{"data":[{"date":"2026-07-31","val":3.4,` +
	`"announcement_datetime":1786537800,"source":"BLS"}]}`

func newServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return server
}

func newPlugin(t *testing.T, baseURL string) *FXMacroData {
	t.Helper()
	plugin := &FXMacroData{
		BaseURL:    baseURL,
		Currencies: []string{"USD"},
		Indicators: []string{"inflation"},
		Log:        testutil.Logger{},
	}
	require.NoError(t, plugin.Init())
	return plugin
}

func TestGatherIndicator(t *testing.T) {
	server := newServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(indicatorBody)); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			t.Error(err)
		}
	})

	plugin := newPlugin(t, server.URL)

	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))

	acc.AssertContainsTaggedFields(t,
		"fxmacrodata_indicator",
		map[string]interface{}{"value": 3.4, "reference_date": "2026-07-31"},
		map[string]string{"currency": "USD", "indicator": "inflation", "source": "BLS"},
	)
}

func TestPointIsStampedWithThePublicationInstant(t *testing.T) {
	// A macro figure is only meaningful alongside when it became known, so the
	// point must not be stamped with collection time when the API reports the
	// publication instant.
	server := newServer(t, func(w http.ResponseWriter, _ *http.Request) {
		if _, err := w.Write([]byte(indicatorBody)); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			t.Error(err)
		}
	})

	plugin := newPlugin(t, server.URL)

	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))

	require.Len(t, acc.Metrics, 1)
	require.Equal(t, time.Unix(1786537800, 0).UTC(), acc.Metrics[0].Time)
}

func TestNullValueIsSkipped(t *testing.T) {
	// A null means the period was not reported. Recording zero would be
	// indistinguishable from a real reading of zero.
	server := newServer(t, func(w http.ResponseWriter, _ *http.Request) {
		if _, err := w.Write([]byte(`{"data":[{"date":"2026-07-31","val":null}]}`)); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			t.Error(err)
		}
	})

	plugin := newPlugin(t, server.URL)

	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))

	require.Empty(t, acc.Metrics)
}

func TestEmptySeriesProducesNoMetric(t *testing.T) {
	server := newServer(t, func(w http.ResponseWriter, _ *http.Request) {
		if _, err := w.Write([]byte(`{"data":[]}`)); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			t.Error(err)
		}
	})

	plugin := newPlugin(t, server.URL)

	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))

	require.Empty(t, acc.Metrics)
}

func TestAPIKeyTravelsAsAHeaderNotAQueryParameter(t *testing.T) {
	var gotHeader, gotQuery string
	server := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotHeader = r.Header.Get("X-API-Key")
		gotQuery = r.URL.RawQuery
		if _, err := w.Write([]byte(indicatorBody)); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			t.Error(err)
		}
	})

	plugin := newPlugin(t, server.URL)
	plugin.APIKey = config.NewSecret([]byte("test-key"))

	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))

	require.Equal(t, "test-key", gotHeader)
	require.NotContains(t, gotQuery, "test-key")
}

func TestNoAuthHeaderWithoutAKey(t *testing.T) {
	// Public USD data has to work with no credential at all.
	var hadHeader bool
	server := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, hadHeader = r.Header["X-Api-Key"]
		if _, err := w.Write([]byte(indicatorBody)); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			t.Error(err)
		}
	})

	plugin := newPlugin(t, server.URL)

	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))

	require.False(t, hadHeader)
}

func TestOneUncoveredCurrencyDoesNotLoseTheOthers(t *testing.T) {
	server := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/announcements/eur/inflation" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if _, err := w.Write([]byte(indicatorBody)); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			t.Error(err)
		}
	})

	plugin := newPlugin(t, server.URL)
	plugin.Currencies = []string{"usd", "eur"}

	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))

	require.Len(t, acc.Metrics, 1)
	require.Equal(t, "USD", acc.Metrics[0].Tags["currency"])
	require.NotEmpty(t, acc.Errors)
}

func TestGatherFX(t *testing.T) {
	server := newServer(t, func(w http.ResponseWriter, _ *http.Request) {
		if _, err := w.Write([]byte(`{"data":[{"date":"2026-09-10","val":1.1616}]}`)); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			t.Error(err)
		}
	})

	plugin := &FXMacroData{
		BaseURL: server.URL,
		FXPairs: []string{"EUR/USD"},
		Log:     testutil.Logger{},
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Gather(&acc))

	acc.AssertContainsTaggedFields(t,
		"fxmacrodata_fx",
		map[string]interface{}{"rate": 1.1616, "reference_date": "2026-09-10"},
		map[string]string{"base": "EUR", "quote": "USD"},
	)
}

func TestInitRequiresSomethingToGather(t *testing.T) {
	plugin := &FXMacroData{Log: testutil.Logger{}}

	require.ErrorContains(t, plugin.Init(), "indicators or fx_pairs")
}

func TestInitRejectsAMalformedPair(t *testing.T) {
	plugin := &FXMacroData{FXPairs: []string{"EURUSD"}, Log: testutil.Logger{}}

	require.ErrorContains(t, plugin.Init(), "BASE/QUOTE")
}

func TestInitDefaults(t *testing.T) {
	plugin := &FXMacroData{Indicators: []string{"Inflation"}, Log: testutil.Logger{}}
	require.NoError(t, plugin.Init())

	require.Equal(t, defaultBaseURL, plugin.BaseURL)
	require.Equal(t, []string{"usd"}, plugin.Currencies)
	require.Equal(t, []string{"inflation"}, plugin.Indicators)
}
