package fibaro

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

// TestUnauthorized validates that 401 (wrong credentials) is managed properly
func TestUnauthorized(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer ts.Close()

	a := Fibaro{
		URL:      ts.URL,
		Username: "user",
		Password: "pass",
		client:   &http.Client{},
	}
	require.NoError(t, a.Init())

	var acc testutil.Accumulator
	err := acc.GatherError(a.Gather)
	require.Error(t, err)
}

// TestJSONSuccess validates that module works OK with valid JSON payloads
func TestJSONSuccess(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		payload := ""
		switch r.URL.Path {
		case "/api/sections":
			content, err := os.ReadFile(path.Join("testdata", "sections.json"))
			require.NoError(t, err)
			payload = string(content)
		case "/api/rooms":
			content, err := os.ReadFile(path.Join("testdata", "rooms.json"))
			require.NoError(t, err)
			payload = string(content)
		case "/api/devices":
			content, err := os.ReadFile(path.Join("testdata", "device_hc2.json"))
			require.NoError(t, err)
			payload = string(content)
		}
		w.WriteHeader(http.StatusOK)
		_, err := fmt.Fprintln(w, payload)
		require.NoError(t, err)
	}))
	defer ts.Close()

	a := Fibaro{
		URL:      ts.URL,
		Username: "user",
		Password: "pass",
		client:   &http.Client{},
	}
	require.NoError(t, a.Init())

	var acc testutil.Accumulator
	err := acc.GatherError(a.Gather)
	require.NoError(t, err)
	require.Equal(t, uint64(5), acc.NMetrics())

	expected := []telegraf.Metric{
		testutil.MustMetric(
			"fibaro",
			map[string]string{
				"deviceId": "1",
				"section":  "Section 1",
				"room":     "Room 1",
				"name":     "Device 1",
				"type":     "com.fibaro.binarySwitch",
			},
			map[string]interface{}{
				"value": float64(0),
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"fibaro",
			map[string]string{
				"deviceId": "2",
				"section":  "Section 2",
				"room":     "Room 2",
				"name":     "Device 2",
				"type":     "com.fibaro.binarySwitch",
			},
			map[string]interface{}{
				"value": float64(1),
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"fibaro",
			map[string]string{
				"deviceId": "3",
				"section":  "Section 3",
				"room":     "Room 3",
				"name":     "Device 3",
				"type":     "com.fibaro.multilevelSwitch",
			},
			map[string]interface{}{
				"value": float64(67),
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"fibaro",
			map[string]string{
				"deviceId": "4",
				"section":  "Section 3",
				"room":     "Room 4",
				"name":     "Device 4",
				"type":     "com.fibaro.temperatureSensor",
			},
			map[string]interface{}{
				"batteryLevel": float64(100),
				"value":        float64(22.8),
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"fibaro",
			map[string]string{
				"deviceId": "5",
				"section":  "Section 3",
				"room":     "Room 4",
				"name":     "Device 5",
				"type":     "com.fibaro.FGRM222",
			},
			map[string]interface{}{
				"energy": float64(4.33),
				"power":  float64(0.7),
				"value":  float64(50),
				"value2": float64(75),
			},
			time.Unix(0, 0),
		),
	}

	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
}

func TestHC3JSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		payload := ""
		switch r.URL.Path {
		case "/api/sections":
			content, err := os.ReadFile(path.Join("testdata", "sections.json"))
			require.NoError(t, err)
			payload = string(content)
		case "/api/rooms":
			content, err := os.ReadFile(path.Join("testdata", "rooms.json"))
			require.NoError(t, err)
			payload = string(content)
		case "/api/devices":
			content, err := os.ReadFile(path.Join("testdata", "device_hc3.json"))
			require.NoError(t, err)
			payload = string(content)
		}
		w.WriteHeader(http.StatusOK)
		_, err := fmt.Fprintln(w, payload)
		require.NoError(t, err)
	}))
	defer ts.Close()

	a := Fibaro{
		URL:        ts.URL,
		Username:   "user",
		Password:   "pass",
		DeviceType: "HC3",
		client:     &http.Client{},
	}
	require.NoError(t, a.Init())

	var acc testutil.Accumulator
	err := acc.GatherError(a.Gather)
	require.NoError(t, err)
	require.Equal(t, uint64(5), acc.NMetrics())

	expected := []telegraf.Metric{
		testutil.MustMetric(
			"fibaro",
			map[string]string{
				"deviceId": "1",
				"section":  "Section 1",
				"room":     "Room 1",
				"name":     "Device 1",
				"type":     "com.fibaro.binarySwitch",
			},
			map[string]interface{}{
				"value": float64(0),
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"fibaro",
			map[string]string{
				"deviceId": "2",
				"section":  "Section 2",
				"room":     "Room 2",
				"name":     "Device 2",
				"type":     "com.fibaro.binarySwitch",
			},
			map[string]interface{}{
				"value": float64(1),
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"fibaro",
			map[string]string{
				"deviceId": "3",
				"section":  "Section 3",
				"room":     "Room 3",
				"name":     "Device 3",
				"type":     "com.fibaro.multilevelSwitch",
			},
			map[string]interface{}{
				"value": float64(67),
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"fibaro",
			map[string]string{
				"deviceId": "4",
				"section":  "Section 3",
				"room":     "Room 4",
				"name":     "Device 4",
				"type":     "com.fibaro.temperatureSensor",
			},
			map[string]interface{}{
				"batteryLevel": float64(100),
				"value":        float64(22.8),
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"fibaro",
			map[string]string{
				"deviceId": "5",
				"section":  "Section 3",
				"room":     "Room 4",
				"name":     "Device 5",
				"type":     "com.fibaro.FGRM222",
			},
			map[string]interface{}{
				"energy": float64(4.33),
				"power":  float64(0.7),
				"value":  float64(34),
			},
			time.Unix(0, 0),
		),
	}

	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
}

func TestInvalidDeviceType(t *testing.T) {
	a := Fibaro{
		DeviceType: "foobar",
	}
	require.Error(t, a.Init())
}
