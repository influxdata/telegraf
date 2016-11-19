package monit

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMonitAll(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		switch r.URL.Path {
		case "/_status":
			http.ServeFile(w, r, "status_response.xml")
		default:
			panic("Cannot handle request")
		}
	}))

	defer ts.Close()

	r := &Monit{
		Address: ts.URL,
	}

	var acc testutil.Accumulator

	err := r.Gather(&acc)
	require.NoError(t, err)

	intMetrics := []string{
		"status_code",
		"monitoring_status_code",
		"service_uptime",
		"mem_kb",
		"mem_kb_total",
	}

	floatMetrics := []string{
		"cpu_percent",
		"cpu_percent_total",
		"mem_percent",
		"mem_percent_total",
		"cpu_system",
		"cpu_user",
		"cpu_wait",
		"cpu_load_avg_1m",
		"cpu_load_avg_5m",
		"cpu_load_avg_15m",
		"swap_kb",
		"swap_percent",
	}

	assert.True(t, acc.HasMeasurement("monit"))
	for _, metric := range intMetrics {
		assert.True(t, acc.HasIntField("monit", metric))

	}

	for _, metric := range floatMetrics {
		assert.True(t, acc.HasFloatField("monit", metric))
	}

}

func TestMonit2(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		switch r.URL.Path {
		case "/_status":
			http.ServeFile(w, r, "status_response_2.xml")
		default:
			panic("Cannot handle request")
		}
	}))

	defer ts.Close()

	r := &Monit{
		Address: ts.URL,
	}

	var acc testutil.Accumulator

	err := r.Gather(&acc)
	require.NoError(t, err)

	intMetrics := []string{
		"status_code",
		"monitoring_status_code",
	}

	assert.True(t, acc.HasMeasurement("monit"))
	for _, metric := range intMetrics {
		assert.True(t, acc.HasIntField("monit", metric))

	}

}

func TestMonit3(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		switch r.URL.Path {
		case "/_status":
			http.ServeFile(w, r, "status_response_3.xml")
		default:
			panic("Cannot handle request")
		}
	}))

	defer ts.Close()

	r := &Monit{
		Address: ts.URL,
	}

	var acc testutil.Accumulator

	err := r.Gather(&acc)
	require.NoError(t, err)

	intMetrics := []string{
		"status_code",
		"monitoring_status_code",
		"service_uptime",
		"mem_kb",
		"mem_kb_total",
	}

	floatMetrics := []string{
		"cpu_percent",
		"cpu_percent_total",
		"mem_percent",
		"mem_percent_total",
	}

	assert.True(t, acc.HasMeasurement("monit"))
	for _, metric := range intMetrics {
		assert.True(t, acc.HasIntField("monit", metric))

	}

	for _, metric := range floatMetrics {
		assert.True(t, acc.HasFloatField("monit", metric))
	}

}

func TestMonit5(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		switch r.URL.Path {
		case "/_status":
			http.ServeFile(w, r, "status_response_5.xml")
		default:
			panic("Cannot handle request")
		}
	}))

	defer ts.Close()

	r := &Monit{
		Address: ts.URL,
	}

	var acc testutil.Accumulator

	err := r.Gather(&acc)
	require.NoError(t, err)

	intMetrics := []string{
		"status_code",
		"monitoring_status_code",
		"mem_kb",
	}

	floatMetrics := []string{
		"mem_percent",
		"cpu_system",
		"cpu_user",
		"cpu_wait",
		"cpu_load_avg_1m",
		"cpu_load_avg_5m",
		"cpu_load_avg_15m",
		"swap_kb",
		"swap_percent",
	}

	assert.True(t, acc.HasMeasurement("monit"))
	for _, metric := range intMetrics {
		assert.True(t, acc.HasIntField("monit", metric))

	}

	for _, metric := range floatMetrics {
		assert.True(t, acc.HasFloatField("monit", metric))
	}

}
