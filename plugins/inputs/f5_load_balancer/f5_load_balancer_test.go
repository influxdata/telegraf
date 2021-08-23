package f5_load_balancer

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
	plugin := &F5LoadBalancer{
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
		plugin   *F5LoadBalancer
		expected string
	}{
		{
			name:     "all empty",
			plugin:   &F5LoadBalancer{},
			expected: "Username cannot be empty",
		},
		{
			name:     "no username",
			plugin:   &F5LoadBalancer{Password: "testpass", URL: "http://example.com"},
			expected: "Username cannot be empty",
		},
		{
			name:     "no password",
			plugin:   &F5LoadBalancer{Username: "testuser", URL: "http://example.com"},
			expected: "Password cannot be empty",
		},
		{
			name:     "no url",
			plugin:   &F5LoadBalancer{Username: "testuser", Password: "testpass"},
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
				if r.URL.Path == "/mgmt/shared/authn/login" {
					w.WriteHeader(http.StatusOK)
					_, err := fmt.Fprintln(w, authenticateResponse)
					require.NoError(t, err)
				} else if r.URL.Path == "/mgmt/tm/ltm/pool" {
					w.WriteHeader(http.StatusOK)
					_, err := fmt.Fprintln(w, sampleGetPoolsUrlResponse)
					require.NoError(t, err)
				} else if r.URL.Path == "/mgmt/tm/ltm/pool/~Common~POOL_TEST_1/stats" {
					w.WriteHeader(http.StatusOK)
					_, err := fmt.Fprintln(w, samplePoolResponseOne)
					require.NoError(t, err)
				}
			},
		),
	)
	defer ts.Close()

	tests := []struct {
		name     string
		plugin   *F5LoadBalancer
		expected []telegraf.Metric
	}{
		{
			name: "gather pool only",
			plugin: &F5LoadBalancer{
				Username:   "testuser",
				Password:   "testpass",
				URL:        ts.URL,
				Collectors: []string{"pool"},
			},
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"f5_load_balancer",
					map[string]string{
						"name": "POOL_TEST_1",
					},
					map[string]interface{}{
						"pool_active_member_count":            6,
						"pool_available":                      1,
						"pool_current_sessions":               27,
						"pool_serverside_bits_in":             4335162092552,
						"pool_serverside_bits_out":            7086935980136,
						"pool_serverside_current_connections": 1541,
						"pool_serverside_packets_in":          1097041172,
						"pool_serverside_packets_out":         1177604238,
						"pool_serverside_total_connections":   42132223,
						"pool_total_requests":                 450843983,
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
		plugin   *F5LoadBalancer
		expected string
	}{
		{
			name: "authentication failed",
			plugin: &F5LoadBalancer{
				Username: "usertest",
				Password: "userpass",
				URL:      ts.URL,
			},
			expected: "No Authentication Token. Exiting...",
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

func TestGetTagsFailed(t *testing.T) {
	ts := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/mgmt/shared/authn/login" {
					w.WriteHeader(http.StatusOK)
					_, err := fmt.Fprintln(w, authenticateResponse)
					require.NoError(t, err)
				} else if r.URL.Path == "/mgmt/tm/ltm/pool" {
					w.WriteHeader(http.StatusOK)
					_, err := fmt.Fprintln(w, sampleGetPoolsUrlResponse)
					require.NoError(t, err)
				} else if r.URL.Path == "/mgmt/tm/ltm/pool/~Common~POOL_TEST_1/stats" {
					w.WriteHeader(http.StatusOK)
					_, err := fmt.Fprintln(w, samplePoolResponseTwo)
					require.NoError(t, err)
				}
			},
		),
	)
	defer ts.Close()
	tests := []struct {
		name        string
		plugin      *F5LoadBalancer
		expected    []telegraf.Metric
		expectedErr string
	}{
		{
			name: "get tags failed",
			plugin: &F5LoadBalancer{
				Username:   "usertest",
				Password:   "userpass",
				URL:        ts.URL,
				Collectors: []string{"pool"},
			},
			expected:    []telegraf.Metric{},
			expectedErr: "Bad or malformed response",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var acc testutil.Accumulator

			tt.plugin.Log = testutil.Logger{}
			require.NoError(t, tt.plugin.Init())

			err := tt.plugin.Gather(&acc)
			require.NoError(t, err)
			require.Len(t, acc.Errors, 1)
			require.EqualError(t, acc.Errors[0], tt.expectedErr)
			testutil.RequireMetricsEqual(t, tt.expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
		})
	}
}

var authenticateResponse = `
{
	"loginProviderName": "tmos",
	"token": {
		"token": "FUL2V33SRR2JBF4NKKADC8RHGX",
		"name": "FUL2V33SRR2JBF4NKKADC8RHGX"
	}
}
`

var sampleGetPoolsUrlResponse = `
{
	"selfLink": "https://localhost/mgmt/tm/ltm/pool?ver=15.1.2.1",
	"items": [
		{
		"kind": "tm:ltm:pool:poolstate",
		"name": "POOL_TEST_1",
		"partition": "Common",
		"fullPath": "/Common/POOL_TEST_1",
		"generation": 1,
		"selfLink": "https://localhost/mgmt/tm/ltm/pool/~Common~POOL_TEST_1?ver=15.1.2.1"
		}
	]
}
`

var samplePoolResponseOne = `
{
	"kind": "tm:ltm:pool:poolstats",
	"selfLink": "https://localhost/mgmt/tm/ltm/pool/~Common~POOL_TEST_1/stats?ver=15.1.2.1",
	"entries": {
	  "https://localhost/mgmt/tm/ltm/pool/~Common~POOL_TEST_1/stats": {
		"nestedStats": {
		  "kind": "tm:ltm:pool:poolstats",
		  "selfLink": "https://localhost/mgmt/tm/ltm/pool/~Common~POOL_TEST_1/stats?ver=15.1.2.1",
		  "entries": {
			"activeMemberCnt": {
			  "value": 6
			},
			"availableMemberCnt": {
			  "value": 6
			},
			"curSessions": {
			  "value": 27
			},
			"memberCnt": {
			  "value": 6
			},
			"minActiveMembers": {
			  "value": 0
			},
			"mr.msgIn": {
			  "value": 0
			},
			"mr.msgOut": {
			  "value": 0
			},
			"mr.reqIn": {
			  "value": 0
			},
			"mr.reqOut": {
			  "value": 0
			},
			"mr.respIn": {
			  "value": 0
			},
			"mr.respOut": {
			  "value": 0
			},
			"tmName": {
			  "description": "/Common/POOL_TEST_1"
			},
			"serverside.bitsIn": {
			  "value": 4335162092552
			},
			"serverside.bitsOut": {
			  "value": 7086935980136
			},
			"serverside.curConns": {
			  "value": 1541
			},
			"serverside.maxConns": {
			  "value": 4005
			},
			"serverside.pktsIn": {
			  "value": 1097041172
			},
			"serverside.pktsOut": {
			  "value": 1177604238
			},
			"serverside.totConns": {
			  "value": 42132223
			},
			"status.availabilityState": {
			  "description": "available"
			},
			"status.enabledState": {
			  "description": "enabled"
			},
			"status.statusReason": {
			  "description": "The pool is available"
			},
			"totRequests": {
			  "value": 450843983
			}
		  }
		}
	  }
	}
  }
`

var samplePoolResponseTwo = `
{
	"kind": "tm:ltm:pool:poolstats",
	"selfLink": "https://localhost/mgmt/tm/ltm/pool/~POOL_TEST_1/stats?ver=15.1.2.1",
	"entries": {
	  "https://localhost/mgmt/tm/ltm/pool/~Common~POOL_TEST_1/stats": {
		"nestedStats": {
		  "kind": "tm:ltm:pool:poolstats",
		  "selfLink": "https://localhost/mgmt/tm/ltm/pool/~Common~POOL_TEST_1/stats?ver=15.1.2.1",
		  "entries": {
			"activeMemberCnt": {
			  "value": 6
			},
			"availableMemberCnt": {
			  "value": 6
			},
			"curSessions": {
			  "value": 27
			},
			"memberCnt": {
			  "value": 6
			},
			"minActiveMembers": {
			  "value": 0
			},
			"mr.msgIn": {
			  "value": 0
			},
			"mr.msgOut": {
			  "value": 0
			},
			"mr.reqIn": {
			  "value": 0
			},
			"mr.reqOut": {
			  "value": 0
			},
			"mr.respIn": {
			  "value": 0
			},
			"mr.respOut": {
			  "value": 0
			},
			"tmName": {
			  "description": "/Common/POOL_TEST_1"
			},
			"serverside.bitsIn": {
			  "value": 4335162092552
			},
			"serverside.bitsOut": {
			  "value": 7086935980136
			},
			"serverside.curConns": {
			  "value": 1541
			},
			"serverside.maxConns": {
			  "value": 4005
			},
			"serverside.pktsIn": {
			  "value": 1097041172
			},
			"serverside.pktsOut": {
			  "value": 1177604238
			},
			"serverside.totConns": {
			  "value": 42132223
			},
			"status.availabilityState": {
			  "description": "available"
			},
			"status.enabledState": {
			  "description": "enabled"
			},
			"status.statusReason": {
			  "description": "The pool is available"
			},
			"totRequests": {
			  "value": 450843983
			}
		  }
		}
	  }
	}
  }
`
