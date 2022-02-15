package xtremio

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
)

var testdataDir = getTestdataDir()

func TestInitDefault(t *testing.T) {
	// This test should succeed with the default initialization.
	plugin := &XtremIO{
		Username: "testuser",
		Password: "testpass",
		URL:      "http://example.com",
		Log:      testutil.Logger{},
	}

	// Test the initialization succeeds
	require.NoError(t, plugin.Init())

	// Also test that default values are set correctly
	require.Equal(t, "testuser", plugin.Username)
	require.Equal(t, "testpass", plugin.Password)
	require.Equal(t, "http://example.com", plugin.URL)
}

func TestInitFail(t *testing.T) {
	tests := []struct {
		name     string
		plugin   *XtremIO
		expected string
	}{
		{
			name:     "all empty",
			plugin:   &XtremIO{},
			expected: "username cannot be empty",
		},
		{
			name:     "no username",
			plugin:   &XtremIO{Password: "testpass", URL: "http://example.com"},
			expected: "username cannot be empty",
		},
		{
			name:     "no password",
			plugin:   &XtremIO{Username: "testuser", URL: "http://example.com"},
			expected: "password cannot be empty",
		},
		{
			name:     "no url",
			plugin:   &XtremIO{Username: "testuser", Password: "testpass"},
			expected: "url cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.plugin.Log = testutil.Logger{}
			err := tt.plugin.Init()
			require.Error(t, err)
			require.EqualError(t, err, tt.expected)
		})
	}
}

func TestFixedValue(t *testing.T) {
	ts := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/api/json/v3/commands/login" {
					cookie := &http.Cookie{Name: "sessid", Value: "cookie:123456789"}
					http.SetCookie(w, cookie)
					w.WriteHeader(http.StatusOK)
					_, err := fmt.Fprintln(w, "authentication succeeded")
					require.NoError(t, err)
				} else if r.URL.Path == "/api/json/v3/types/bbus" {
					sampleGetBBUsResponse, err := ioutil.ReadFile(filepath.Join(testdataDir, "sample_get_bbu_response.json"))
					require.NoError(t, err)
					w.WriteHeader(http.StatusOK)
					_, err = fmt.Fprintln(w, string(sampleGetBBUsResponse))
					require.NoError(t, err)
				} else if r.URL.Path == "/api/json/v3/types/bbus/987654321abcdef" {
					sampleBBUResponseOne, err := ioutil.ReadFile(filepath.Join(testdataDir, "sample_bbu_response.json"))
					require.NoError(t, err)
					w.WriteHeader(http.StatusOK)
					_, err = fmt.Fprintln(w, string(sampleBBUResponseOne))
					require.NoError(t, err)
				}
			},
		),
	)
	defer ts.Close()

	tests := []struct {
		name     string
		plugin   *XtremIO
		expected []telegraf.Metric
	}{
		{
			name: "gather bbus only",
			plugin: &XtremIO{
				Username:   "testuser",
				Password:   "testpass",
				URL:        ts.URL,
				Collectors: []string{"bbus"},
			},
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"xio",
					map[string]string{
						"serial_number": "A123B45678",
						"guid":          "987654321abcdef",
						"power_feed":    "PWR-A",
						"name":          "X1-BBU",
						"model_name":    "Eaton Model Name",
					},
					map[string]interface{}{
						"bbus_power":                        244,
						"bbus_average_daily_temp":           23,
						"bbus_enabled":                      true,
						"bbus_ups_need_battery_replacement": false,
						"bbus_ups_low_battery_no_input":     false,
					},
					time.Unix(0, 0),
				),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var acc testutil.Accumulator
			tt.plugin.Log = testutil.Logger{}
			require.NoError(t, tt.plugin.Init())
			require.NoError(t, tt.plugin.Gather(&acc))
			require.Len(t, acc.Errors, 0, "found errors accumulated by acc.AddError()")
			acc.Wait(len(tt.expected))
			testutil.RequireMetricsEqual(t, tt.expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
		})
	}
}

func TestAuthenticationFailed(t *testing.T) {
	ts := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				_, err := fmt.Fprintln(w, "bad request")
				require.NoError(t, err)
			},
		),
	)
	defer ts.Close()
	tests := []struct {
		name     string
		plugin   *XtremIO
		expected string
	}{
		{
			name: "authentication failed",
			plugin: &XtremIO{
				Username: "usertest",
				Password: "userpass",
				URL:      ts.URL,
			},
			expected: "no authentication cookie set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var acc testutil.Accumulator

			tt.plugin.Log = testutil.Logger{}
			require.NoError(t, tt.plugin.Init())

			err := tt.plugin.Gather(&acc)
			require.Error(t, err)
			require.EqualError(t, err, tt.expected)
		})
	}
}

func getTestdataDir() string {
	dir, err := os.Getwd()
	if err != nil {
		// if we cannot even establish the test directory, further progress is meaningless
		panic(err)
	}

	return filepath.Join(dir, "testdata")
}
