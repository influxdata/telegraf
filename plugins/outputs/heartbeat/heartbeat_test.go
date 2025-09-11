package heartbeat

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/testutil"
)

func TestInitSuccess(t *testing.T) {
	u := config.NewSecret([]byte("http://localhost/heartbeat"))
	t.Cleanup(u.Destroy)

	plugin := &Heartbeat{
		URL:        u,
		InstanceID: "telegraf",
		Interval:   config.Duration(10 * time.Second),
	}

	require.NoError(t, plugin.Init())
}

func TestInitFail(t *testing.T) {
	u := config.NewSecret([]byte("http://localhost/heartbeat"))
	t.Cleanup(u.Destroy)

	tests := []struct {
		name     string
		plugin   *Heartbeat
		expected string
	}{
		{
			name:     "missing URL",
			plugin:   &Heartbeat{},
			expected: "url required",
		},
		{
			name: "missing instance ID",
			plugin: &Heartbeat{
				URL: u,
			},
			expected: "instance ID required",
		},
		{
			name: "invalid interval",
			plugin: &Heartbeat{
				URL:        u,
				InstanceID: "telegraf",
			},
			expected: "invalid interval",
		},
		{
			name: "invalid include",
			plugin: &Heartbeat{
				URL:        u,
				InstanceID: "telegraf",
				Interval:   config.Duration(10 * time.Second),
				Include:    []string{"foo"},
			},
			expected: "invalid 'include' setting",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.ErrorContains(t, tt.plugin.Init(), tt.expected)
		})
	}
}

func TestIncludes(t *testing.T) {
	u := config.NewSecret([]byte("http://localhost/heartbeat"))
	t.Cleanup(u.Destroy)

	plugin := &Heartbeat{
		URL:        u,
		InstanceID: "telegraf",
		Interval:   config.Duration(10 * time.Second),
		Include:    []string{"configs", "hostname" /* , logs */, "metrics", "status"},
		Log:        &testutil.Logger{},
	}

	require.NoError(t, plugin.Init())
}

func TestIncludedExtraData(t *testing.T) {
	// Get the hostname for test-data construction
	hostname, err := os.Hostname()
	require.NoError(t, err)

	// Add a dummy http server for configs
	cfgServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer cfgServer.Close()

	// Add some dummy configuration files
	absdir, err := filepath.Abs("testdata")
	require.NoError(t, err)
	cfgs := []string{
		"testdata/telegraf.conf",
		absdir + "/telegraf.d/inputs.conf",
		"http://user:password@" + cfgServer.Listener.Addr().String(),
		absdir + "/telegraf.d/outputs.conf",
		"http://" + cfgServer.Listener.Addr().String() + "/myconfigs",
		"testdata/non_existing.conf",
	}
	cfg := config.NewConfig()
	for _, c := range cfgs {
		//nolint:errcheck // Ignore error on purpose as some endpoints won't be loadable
		cfg.LoadConfig(c)
	}
	cfgServer.Close()

	// Expected configs
	cfgsExpected := strings.Join([]string{
		`"testdata/telegraf.conf"`,
		`"` + absdir + `/telegraf.d/inputs.conf"`,
		`"http://user:xxxxx@` + cfgServer.Listener.Addr().String() + `"`,
		`"` + absdir + `/telegraf.d/outputs.conf"`,
		`"http://` + cfgServer.Listener.Addr().String() + `/myconfigs"`,
	}, ",")

	// Prepare a string replacer to replace dynamic content such as the hostname
	// in the expected strings
	replacer := strings.NewReplacer(
		"$HOSTNAME", hostname,
		"$VERSION", internal.FormatFullVersion(),
		"$CONFIGS", cfgsExpected,
	)

	tests := []struct {
		name     string
		includes []string
		expected string
	}{
		{
			name: "minimal",
			expected: `{
			  "id": "telegraf",
			  "version": "$VERSION"
			}`,
		},
		{
			name:     "hostname",
			includes: []string{"hostname"},
			expected: `{
			  "id": "telegraf",
			  "version": "$VERSION",
			  "hostname": "$HOSTNAME"
			}`,
		},
		{
			name:     "metrics",
			includes: []string{"metrics"},
			expected: `{
			  "id": "telegraf",
			  "version": "$VERSION",
			  "metrics": 5
			}`,
		},
		{
			name:     "status",
			includes: []string{"status"},
			expected: `{
			  "id": "telegraf",
			  "version": "$VERSION",
			  "status": "OK"
			}`,
		},
		{
			name:     "configurations",
			includes: []string{"configs"},
			expected: `{
			  "id": "telegraf",
			  "version": "$VERSION",
			  "configurations": [$CONFIGS]
			}`,
		},
		{
			name:     "all",
			includes: []string{"configs", "hostname" /* , logs */, "metrics", "status"},
			expected: `{
			  "id": "telegraf",
			  "version": "$VERSION",
			  "hostname": "$HOSTNAME",
			  "metrics": 5,
			  "status": "OK",
			  "configurations": [$CONFIGS]
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Inject dynamic content into the expectation
			expected := replacer.Replace(tt.expected)

			// Create a test server to validate the data sent
			var actual string
			var done atomic.Bool
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Fail()
					w.WriteHeader(http.StatusMethodNotAllowed)
				}

				// Decode the body
				body, err := io.ReadAll(r.Body)
				if err != nil {
					t.Fail()
					w.WriteHeader(http.StatusInternalServerError)
				}
				actual = string(body)
				done.Store(true)
				w.WriteHeader(http.StatusOK)
			}))
			defer ts.Close()

			// Initialize the plugin
			u := config.NewSecret([]byte("http://" + ts.Listener.Addr().String()))
			defer u.Destroy()

			plugin := &Heartbeat{
				URL:        u,
				InstanceID: "telegraf",
				Interval:   config.Duration(time.Second),
				Include:    tt.includes,
				Log:        &testutil.Logger{},
			}
			require.NoError(t, plugin.Init())

			// Write 5 metrics to be able to test the metric count. This has to
			// be done before "connecting" to avoid race conditions between
			// the Write function and the heartbeat ticker.
			require.NoError(t, plugin.Write([]telegraf.Metric{
				testutil.TestMetric(0),
				testutil.TestMetric(1),
				testutil.TestMetric(2),
				testutil.TestMetric(3),
				testutil.TestMetric(4),
			}))

			require.NoError(t, plugin.Connect())
			defer plugin.Close()

			// Wait for the data to arrive at the test-server and check the
			// payload we got.
			require.Eventually(t, func() bool {
				return done.Load()
			}, 3*time.Second, 100*time.Millisecond)
			require.JSONEq(t, expected, actual)
		})
	}
}
