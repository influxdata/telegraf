package syncthing_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs/syncthing"
	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSyncthing_Init_FailsWhenNoTokenProvided(t *testing.T) {
	err := (&syncthing.Syncthing{}).Init()
	require.Error(t, err)
	require.Contains(t, err.Error(), "token")
}

func TestSyncthing_Init_FailsWhenTokenFileIsUnreadable(t *testing.T) {
	err := (&syncthing.Syncthing{
		TokenFile: "/8y4hickjasdf",
	}).Init()
	require.Error(t, err)
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected NotExist error, but got %q", err.Error())
	}
}

func TestSyncthing_Init_PassesTokenHeader(t *testing.T) {
	const token = `laskdhlkfajhfhasdfd`
	fakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actual := r.Header.Get(syncthing.AuthHeader)
		if assert.Equal(t, actual, token) {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer fakeServer.Close()

	url := fakeServer.URL
	plugin := &syncthing.Syncthing{
		URL:     url,
		Timeout: internal.Duration{Duration: time.Second},
		Token:   token,
	}

	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	err := acc.GatherError(plugin.Gather)
	require.Error(t, err)
}

func TestSyncthing_Connections(t *testing.T) {
	testValues := &TestValues{
		FolderID:      "GXWxf-3zgnU",
		FolderLabel:   "MyFolder",
		FolderPath:    "/some/path",
		DeviceID:      "YZJBJFX-RDBL7WY-6ZGKJ2D-4MJB4E7-ZATDSYU-LDY3L63-MFLUWYE-AEMXJAC",
		InBytesTotal:  550,
		OutBytesTotal: 62763,
		TotalNeeded:   872768,
		DeviceAddress: "192.168.2.2",
		DeviceName:    "MyDevice",
		ClientVersion: "v1.60.1",
		Connected:     true,
		Crypto:        "TLS-1.3",
		Paused:        false,
		MyID:          "P56IOI7-MZJNU2Y-IQGDREY-DM2MGTI-MGL3BXN-PQ6W5BM-TBBZ4TJ-XZWICQ2",
	}

	fakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case syncthing.SystemStatusEndpoint:
			_ = systemStatusJSON.Execute(w, testValues)
		case syncthing.SysconfigEndpoint:
			_ = sysconfigJSON.Execute(w, testValues)
		case syncthing.ConnectionsEndpoint:
			_ = connectionJSON.Execute(w, testValues)
		case syncthing.NeedEndpoint:
			_ = folderNeedJSON.Execute(w, testValues)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer fakeServer.Close()

	url := fakeServer.URL + syncthing.SysconfigEndpoint
	plugin := &syncthing.Syncthing{
		URL:     url,
		Timeout: internal.Duration{Duration: time.Second},
		Token:   "token",
	}
	const metricName = "syncthing_connection"

	p, _ := parsers.NewParser(&parsers.Config{
		DataFormat: "json",
		MetricName: metricName,
	})
	plugin.SetParser(p)

	var acc testutil.Accumulator
	require.NoError(t, plugin.Init())
	require.NoError(t, acc.GatherError(plugin.Gather))

	// basic check to see if we got the right field, value and tag
	metric, ok := acc.Get(metricName)
	require.True(t, ok, "metric name was not found: %q", metricName)

	require.Equal(t, metric.Measurement, metricName)

	expectedTags := map[string]interface{}{
		"name":      testValues.DeviceName,
		"device_id": testValues.DeviceID,
	}

	for k, v := range expectedTags {
		assert.Contains(t, metric.Tags, k)
		assert.EqualValues(t, v, metric.Tags[k])
	}

	expectedFields := map[string]interface{}{
		"address":         testValues.DeviceAddress,
		"client_version":  testValues.ClientVersion,
		"connected":       testValues.Connected,
		"crypto":          testValues.Crypto,
		"in_bytes_total":  testValues.InBytesTotal,
		"out_bytes_total": testValues.OutBytesTotal,
		"paused":          testValues.Paused,
	}

	for k, v := range expectedFields {
		assert.Contains(t, metric.Fields, k)
		assert.EqualValues(t, v, metric.Fields[k])
	}
}

func TestSyncthing_Folders(t *testing.T) {
	const (
		sysconfigEndpoint   = "/rest/system/config"
		connectionsEndpoint = "/rest/system/connections"
		needEndpoint        = "/rest/db/need"
	)

	testValues := &TestValues{
		FolderID:      "GXWxf-3zgnU",
		FolderPath:    "/some/path",
		DeviceID:      "YZJBJFX-RDBL7WY-6ZGKJ2D-4MJB4E7-ZATDSYU-LDY3L63-MFLUWYE-AEMXJAC",
		InBytesTotal:  550,
		OutBytesTotal: 62763,
		TotalNeeded:   872768,
		DeviceAddress: "192.168.2.2",
		DeviceName:    "MyDevice",
		ClientVersion: "v1.60.1",
		Connected:     true,
		Crypto:        "TLS-1.3",
		Paused:        false,
		MyID:          "P56IOI7-MZJNU2Y-IQGDREY-DM2MGTI-MGL3BXN-PQ6W5BM-TBBZ4TJ-XZWICQ2",
	}

	fakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case syncthing.SystemStatusEndpoint:
			_ = systemStatusJSON.Execute(w, testValues)
		case sysconfigEndpoint:
			_ = sysconfigJSON.Execute(w, testValues)
		case connectionsEndpoint:
			_ = connectionJSON.Execute(w, testValues)
		case needEndpoint:
			_ = folderNeedJSON.Execute(w, testValues)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer fakeServer.Close()

	url := fakeServer.URL + sysconfigEndpoint
	plugin := &syncthing.Syncthing{
		URL:     url,
		Timeout: internal.Duration{Duration: time.Second},
		Token:   "token",
	}
	const metricName = "syncthing_folder"

	p, _ := parsers.NewParser(&parsers.Config{
		DataFormat: "json",
		MetricName: metricName,
	})
	plugin.SetParser(p)

	var acc testutil.Accumulator
	require.NoError(t, plugin.Init())
	require.NoError(t, acc.GatherError(plugin.Gather))

	// basic check to see if we got the right field, value and tag
	metric, ok := acc.Get(metricName)
	require.True(t, ok, "metric name was not found: %q", metricName)

	require.Equal(t, metric.Measurement, metricName)

	expectedFields := map[string]interface{}{
		"paused": testValues.Paused,
		"need":   testValues.TotalNeeded,
	}
	expectedTags := map[string]string{
		"label": testValues.FolderLabel,
		"id":    testValues.FolderID,
		"path":  testValues.FolderPath,
	}

	for k, v := range expectedTags {
		assert.Contains(t, metric.Tags, k)
		assert.EqualValues(t, v, metric.Tags[k])
	}

	for k, v := range expectedFields {
		assert.Contains(t, metric.Fields, k)
		assert.EqualValues(t, v, metric.Fields[k])
	}
}
