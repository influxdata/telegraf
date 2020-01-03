// Copyright 2019, Verizon
// Licensed under the terms of the MIT License. See LICENSE file in project root for terms.

package monit

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMonit0(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		switch r.URL.Path {
		case "/_status":
			http.ServeFile(w, r, "status_response_0.xml")
		default:
			panic("Cannot handle request")
		}
	}))

	defer ts.Close()

	r := &Monit{
		Address: ts.URL,
	}

	var acc testutil.Accumulator

	r.Init()

	err := r.Gather(&acc)
	require.NoError(t, err)

	intMetrics := []string{
		"status_code",
		"monitoring_status_code",
		"mode",
	}

	floatMetrics := []string{
		"block_percent",
		"block_usage",
		"block_total",
		"inode_percent",
		"inode_usage",
		"inode_total",
	}

	assert.True(t, acc.HasMeasurement("filesystem"))
	for _, metric := range intMetrics {
		assert.True(t, acc.HasIntField("filesystem", metric))

	}

	for _, metric := range floatMetrics {
		assert.True(t, acc.HasFloatField("filesystem", metric))
	}

}

func TestMonit1(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		switch r.URL.Path {
		case "/_status":
			http.ServeFile(w, r, "status_response_1.xml")
		default:
			panic("Cannot handle request")
		}
	}))

	defer ts.Close()

	r := &Monit{
		Address: ts.URL,
	}

	var acc testutil.Accumulator

	r.Init()

	err := r.Gather(&acc)
	require.NoError(t, err)

	intMetrics := []string{
		"status_code",
		"monitoring_status_code",
		"permissions",
	}

	assert.True(t, acc.HasMeasurement("directory"))
	for _, metric := range intMetrics {
		assert.True(t, acc.HasIntField("directory", metric))
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

	r.Init()

	err := r.Gather(&acc)
	require.NoError(t, err)

	intMetrics := []string{
		"status_code",
		"monitoring_status_code",
		"permissions",
	}

	int64Metrics := []string{
		"size",
	}

	assert.True(t, acc.HasMeasurement("file"))
	for _, metric := range intMetrics {
		assert.True(t, acc.HasIntField("file", metric))
	}

	for _, metric := range int64Metrics {
		assert.True(t, acc.HasInt64Field("file", metric))
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

	r.Init()

	err := r.Gather(&acc)
	require.NoError(t, err)

	intMetrics := []string{
		"status_code",
		"monitoring_status_code",
}

	int64Metrics := []string{
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

	assert.True(t, acc.HasMeasurement("process"))
	for _, metric := range intMetrics {
		assert.True(t, acc.HasIntField("process", metric))
	}

	for _, metric := range int64Metrics {
		assert.True(t, acc.HasInt64Field("process", metric))
	}

	for _, metric := range floatMetrics {
		assert.True(t, acc.HasFloatField("process", metric))
	}

}

func TestMonit4(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		switch r.URL.Path {
		case "/_status":
			http.ServeFile(w, r, "status_response_4.xml")
		default:
			panic("Cannot handle request")
		}
	}))

	defer ts.Close()

	r := &Monit{
		Address: ts.URL,
	}

	var acc testutil.Accumulator

	r.Init()

	err := r.Gather(&acc)
	require.NoError(t, err)

	intMetrics := []string{
		"status_code",
		"monitoring_status_code",
	}

	int64Metrics := []string{
		"port_number",
	}

	float64Metrics := []string{
		"response_time",
	}

	stringMetrics := []string{
		"hostname",
		"request",
		"protocol",
		"type",
	}

	assert.True(t, acc.HasMeasurement("remote_host"))
	for _, metric := range intMetrics {
		assert.True(t, acc.HasIntField("remote_host", metric))
	}

	for _, metric := range int64Metrics {
		assert.True(t, acc.HasInt64Field("remote_host", metric))
	}

	for _, metric := range float64Metrics {
		assert.True(t, acc.HasFloatField("remote_host", metric))
	}

	for _, metric := range stringMetrics {
		assert.True(t, acc.HasField("remote_host", metric))
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

	r.Init()

	err := r.Gather(&acc)
	require.NoError(t, err)

	intMetrics := []string{
		"status_code",
		"monitoring_status_code",
	}

	int64Metrics := []string{
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

	assert.True(t, acc.HasMeasurement("system"))
	for _, metric := range intMetrics {
		assert.True(t, acc.HasIntField("system", metric))
	}

	for _, metric := range int64Metrics {
		assert.True(t, acc.HasInt64Field("system", metric))
	}

	for _, metric := range floatMetrics {
		assert.True(t, acc.HasFloatField("system", metric))
	}

}

func TestMonit6(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		switch r.URL.Path {
		case "/_status":

			http.ServeFile(w, r, "status_response_6.xml")
		default:
			panic("Cannot handle request")
		}
	}))

	defer ts.Close()

	r := &Monit{
		Address: ts.URL,
	}

	var acc testutil.Accumulator

	r.Init()

	err := r.Gather(&acc)
	require.NoError(t, err)

	intMetrics := []string{
		"status_code",
		"monitoring_status_code",
		"permissions",
	}

	assert.True(t, acc.HasMeasurement("fifo"))
	for _, metric := range intMetrics {
		assert.True(t, acc.HasIntField("fifo", metric))
	}

}

func TestMonit7(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		switch r.URL.Path {
		case "/_status":

			http.ServeFile(w, r, "status_response_7.xml")
		default:
			panic("Cannot handle request")
		}
	}))

	defer ts.Close()

	r := &Monit{
		Address: ts.URL,
	}

	var acc testutil.Accumulator

	r.Init()

	err := r.Gather(&acc)
	require.NoError(t, err)

	intMetrics := []string{
		"status_code",
		"monitoring_status_code",
		"program_status",
	}

	int64Metrics := []string{
		"last_started_time",
	}

	stringMetrics := []string{
		"output",
	}

	assert.True(t, acc.HasMeasurement("program"))
	for _, metric := range intMetrics {
		assert.True(t, acc.HasIntField("program", metric))
	}

	for _, metric := range int64Metrics {
		assert.True(t, acc.HasInt64Field("program", metric))
	}

	for _, metric := range stringMetrics {
		assert.True(t, acc.HasField("program", metric))
	}

}

func TestMonit8(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		switch r.URL.Path {
		case "/_status":

			http.ServeFile(w, r, "status_response_8.xml")
		default:
			panic("Cannot handle request")
		}
	}))

	defer ts.Close()

	r := &Monit{
		Address: ts.URL,
	}

	var acc testutil.Accumulator

	r.Init()

	err := r.Gather(&acc)
	require.NoError(t, err)

	intMetrics := []string{
		"status_code",
		"monitoring_status_code",
		"link_state",
	}

	int64Metrics := []string{
		"link_speed",
		"download_packets_now",
		"download_packets_total",
		"download_bytes_now",
		"download_bytes_total",
		"download_errors_now",
		"download_errors_total",
		"upload_packets_now",
		"upload_packets_total",
		"upload_bytes_now",
		"upload_bytes_total",
		"upload_errors_now",
		"upload_errors_total",
	}

	stringMetrics := []string{
		"link_mode",
	}

	assert.True(t, acc.HasMeasurement("network"))
	for _, metric := range intMetrics {
		assert.True(t, acc.HasIntField("network", metric))
	}

	for _, metric := range int64Metrics {
		assert.True(t, acc.HasInt64Field("network", metric))
	}

	for _, metric := range stringMetrics {
		assert.True(t, acc.HasField("network", metric))
	}

}
