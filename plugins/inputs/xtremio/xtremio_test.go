package xtremio

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
)

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
			expected: "Username cannot be empty",
		},
		{
			name:     "no username",
			plugin:   &XtremIO{Password: "testpass", URL: "http://example.com"},
			expected: "Username cannot be empty",
		},
		{
			name:     "no password",
			plugin:   &XtremIO{Username: "testuser", URL: "http://example.com"},
			expected: "Password cannot be empty",
		},
		{
			name:     "no url",
			plugin:   &XtremIO{Username: "testuser", Password: "testpass"},
			expected: "URL cannot be empty",
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
					w.WriteHeader(http.StatusOK)
					_, err := fmt.Fprintln(w, sampleGetBBUsResponse)
					require.NoError(t, err)
				} else if r.URL.Path == "/api/json/v3/types/bbus/987654321abcdef" {
					w.WriteHeader(http.StatusOK)
					_, err := fmt.Fprintln(w, sampleBBUResponseOne)
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
						"bbus_power":                    244,
						"bbus_average_daily_temp":       23,
						"bbus_enabled":                  1,
						"ups_need_battery_replacement":  0,
						"bbus_ups_low_battery_no_input": 0,
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

var authenticateResponse = `
{
	"loginProviderName": "tmos",
	"token": {
		"token": "FUL2V33SRR2JBF4NKKABCDEFGH",
		"name": "FUL2V33SRR2JBF4NKKABCDEFGH"
	}
}
`

var sampleBBUResponseOne = `
{
    "content": {
        "is-low-battery-has-input": "false",
        "serial-number": "A123B45678",
        "guid": "987654321abcdef",
        "brick-name": "X1",
        "ups-battery-charge-in-percent": 100,
        "power": 244,
        "avg-daily-temp": 23,
        "fw-version": "01.02.0034",
        "sys-name": "ABCXIO001",
		"power-feed": "PWR-A",
        "ups-load-in-percent": 21,
        "name": "X1-BBU",
		"enabled-state": "enabled",
        "is-low-battery-no-input": "false",
        "ups-need-battery-replacement": "false",
        "model-name": "Eaton Model Name",
    }
}
`

var sampleGetBBUsResponse = `
{
    "bbus": [
        {
            "href": "https://127.0.0.1/api/json/v3/types/bbus/987654321abcdef", 
            "name": "X1-BBU", 
            "sys-name": "ABCXIO001"
        }
    ], 
    "links": [
        {
            "href": "https://127.0.0.1/api/json/v3/types/bbus/", 
            "rel": "self"
        }
    ]
}
`
