package heartbeat

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/santhosh-tekuri/jsonschema/v5"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/logger"
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
	logtime := time.Now()
	replacer := strings.NewReplacer(
		"$HOSTNAME", hostname,
		"$VERSION", internal.FormatFullVersion(),
		"$SCHEMA", strconv.Itoa(jsonSchemaVersion),
		"$CONFIGS", cfgsExpected,
		"$LOGTIME", logtime.UTC().Format(time.RFC3339Nano),
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
			  "version": "$VERSION",
			  "schema": $SCHEMA
			}`,
		},
		{
			name:     "hostname",
			includes: []string{"hostname"},
			expected: `{
			  "id": "telegraf",
			  "version": "$VERSION",
			  "schema": $SCHEMA,
			  "hostname": "$HOSTNAME"
			}`,
		},
		{
			name:     "metrics",
			includes: []string{"metrics"},
			expected: `{
			  "id": "telegraf",
			  "version": "$VERSION",
			  "schema": $SCHEMA,
			  "metrics": 5
			}`,
		},
		{
			name:     "status",
			includes: []string{"status"},
			expected: `{
			  "id": "telegraf",
			  "version": "$VERSION",
			  "schema": $SCHEMA,
			  "status": "OK"
			}`,
		},
		{
			name:     "configurations",
			includes: []string{"configs"},
			expected: `{
			  "id": "telegraf",
			  "version": "$VERSION",
			  "schema": $SCHEMA,
			  "configurations": [$CONFIGS]
			}`,
		},
		{
			name:     "logs",
			includes: []string{"logs"},
			expected: `{
			  "id": "telegraf",
			  "version": "$VERSION",
			  "schema": $SCHEMA,
			  "logs": {
				"errors": 1,
				"warnings": 2
			  }
			}`,
		},

		{
			name:     "log-details",
			includes: []string{"log-details"},
			expected: `{
				"id": "telegraf",
				"version": "$VERSION",
				"schema": $SCHEMA,
				"logs": {
					"entries": [
						{
							"time": "$LOGTIME",
							"level": "ERROR",
							"source": "inputs.test::hbt",
							"attributes": {
								"alias": "hbt",
								"category": "inputs",
								"plugin": "test",
								"source": "heartbeat",
								"type": "testing"
							},
							"message": "An error message"
						}
					]
				}
			}`,
		},
		{
			name:     "all",
			includes: []string{"configs", "hostname", "logs", "log-details", "metrics", "status"},
			expected: `{
			  "id": "telegraf",
			  "version": "$VERSION",
			  "schema": $SCHEMA,
			  "hostname": "$HOSTNAME",
			  "metrics": 5,
			  "status": "OK",
			  "configurations": [$CONFIGS],
			  "logs": {
				"errors": 1,
				"warnings": 2,
				"entries": [
						{
							"time": "$LOGTIME",
							"level": "ERROR",
							"source": "inputs.test::hbt",
							"attributes": {
								"alias": "hbt",
								"category": "inputs",
								"plugin": "test",
								"source": "heartbeat",
								"type": "testing"
							},
							"message": "An error message"
						}
					]
			    }
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

			// Register the logging handler early to avoid race conditions
			// during testing. This is not ideal but the only way to get
			// reliable tests without a race between the Connect call,
			// registering the callback and the actual logging.
			plugin.logCallbackID = logger.AddCallback(plugin.handleLogEvent)

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

			// Log a few messages to be able to test the log-handling
			if slices.Contains(tt.includes, "logs") || slices.Contains(tt.includes, "log-details") {
				logger := logger.New("inputs", "test", "hbt")
				logger.AddAttribute("source", "heartbeat")
				logger.AddAttribute("type", "testing")
				logger.Print(telegraf.Error, logtime, "An error message")
				logger.Print(telegraf.Warn, logtime, "A first warning")
				logger.Print(telegraf.Info, logtime, "An information")
				logger.Print(telegraf.Debug, logtime, "A debug information")
				logger.Print(telegraf.Trace, logtime, "A trace information")
				logger.Print(telegraf.Warn, logtime, "A second warning")
			}

			require.NoError(t, plugin.Connect())
			defer plugin.Close()

			// Wait for the data to arrive at the test-server and check the
			// payload we got.
			require.Eventually(t, func() bool {
				return done.Load()
			}, 3*time.Second, 100*time.Millisecond)
			require.JSONEq(t, expected, actual, actual)

			// Check heartbeat message against the JSON schema
			schema, err := jsonschema.Compile(fmt.Sprintf("schema_v%d.json", jsonSchemaVersion))
			require.NoError(t, err)

			var v interface{}
			require.NoError(t, json.Unmarshal([]byte(actual), &v))
			require.NoError(t, schema.Validate(v))
		})
	}
}

func TestDetailedLogging(t *testing.T) {
	// Get the hostname for test-data construction
	hostname, err := os.Hostname()
	require.NoError(t, err)

	// Prepare a string replacer to replace dynamic content such as the hostname
	// in the expected strings
	replacer := strings.NewReplacer(
		"$HOSTNAME", hostname,
		"$VERSION", internal.FormatFullVersion(),
		"$SCHEMA", strconv.Itoa(jsonSchemaVersion),
	)

	// Prepare the logging time reference
	logtime, err := time.Parse(time.RFC3339, "2025-10-22T16:02:30Z")
	require.NoError(t, err)

	tests := []struct {
		name     string
		source   string
		attrs    map[string]interface{}
		logcfg   LogsConfig
		logs     []logEvent
		expected string
	}{
		{
			name:   "default config",
			source: "inputs.test::hbt",
			logs: []logEvent{
				{
					timestamp: logtime,
					level:     telegraf.Error,
					msg:       "An error message",
				},
				{
					timestamp: logtime.Add(1 * time.Second),
					level:     telegraf.Warn,
					msg:       "A first warning",
				},
				{
					timestamp: logtime.Add(2 * time.Second),
					level:     telegraf.Info,
					msg:       "An information",
				},
				{
					timestamp: logtime.Add(3 * time.Second),
					level:     telegraf.Debug,
					msg:       "A debug information",
				},
				{
					timestamp: logtime.Add(4 * time.Second),
					level:     telegraf.Trace,
					msg:       "A trace information",
				},
				{
					timestamp: logtime.Add(5 * time.Second),
					level:     telegraf.Warn,
					msg:       "A second warning",
				},
			},
			expected: `{
				"id": "telegraf",
				"version": "$VERSION",
				"schema": $SCHEMA,
				"logs": {
					"entries": [
						{
							"time": "2025-10-22T16:02:30Z",
							"level": "ERROR",
							"source": "inputs.test::hbt",
							"attributes": {
								"alias": "hbt",
								"category": "inputs",
								"plugin": "test"
							},
							"message": "An error message"
						}
					]
				}
			}`,
		},
		{
			name:   "warn level",
			source: "inputs.test::hbt",
			logcfg: LogsConfig{
				LogLevel: "warn",
			},
			logs: []logEvent{
				{
					timestamp: logtime,
					level:     telegraf.Error,
					msg:       "An error message",
				},
				{
					timestamp: logtime.Add(1 * time.Second),
					level:     telegraf.Warn,
					msg:       "A first warning",
				},
				{
					timestamp: logtime.Add(2 * time.Second),
					level:     telegraf.Info,
					msg:       "An information",
				},
				{
					timestamp: logtime.Add(3 * time.Second),
					level:     telegraf.Debug,
					msg:       "A debug information",
				},
				{
					timestamp: logtime.Add(4 * time.Second),
					level:     telegraf.Trace,
					msg:       "A trace information",
				},
				{
					timestamp: logtime.Add(5 * time.Second),
					level:     telegraf.Warn,
					msg:       "A second warning",
				},
			},
			expected: `{
				"id": "telegraf",
				"version": "$VERSION",
				"schema": $SCHEMA,
				"logs": {
					"entries": [
						{
							"time": "2025-10-22T16:02:30Z",
							"level": "ERROR",
							"source": "inputs.test::hbt",
							"attributes": {
								"alias": "hbt",
								"category": "inputs",
								"plugin": "test"
							},
							"message": "An error message"
						},
						{
							"time": "2025-10-22T16:02:31Z",
							"level": "WARN",
							"source": "inputs.test::hbt",
							"attributes": {
								"alias": "hbt",
								"category": "inputs",
								"plugin": "test"
							},
							"message": "A first warning"
						},
						{
							"time": "2025-10-22T16:02:35Z",
							"level": "WARN",
							"source": "inputs.test::hbt",
							"attributes": {
								"alias": "hbt",
								"category": "inputs",
								"plugin": "test"
							},
							"message": "A second warning"
						}
					]
				}
			}`,
		},
		{
			name:   "info level",
			source: "inputs.test::hbt",
			logcfg: LogsConfig{
				LogLevel: "info",
			},
			logs: []logEvent{
				{
					timestamp: logtime,
					level:     telegraf.Error,
					msg:       "An error message",
				},
				{
					timestamp: logtime.Add(1 * time.Second),
					level:     telegraf.Warn,
					msg:       "A first warning",
				},
				{
					timestamp: logtime.Add(2 * time.Second),
					level:     telegraf.Info,
					msg:       "An information",
				},
				{
					timestamp: logtime.Add(3 * time.Second),
					level:     telegraf.Debug,
					msg:       "A debug information",
				},
				{
					timestamp: logtime.Add(4 * time.Second),
					level:     telegraf.Trace,
					msg:       "A trace information",
				},
				{
					timestamp: logtime.Add(5 * time.Second),
					level:     telegraf.Warn,
					msg:       "A second warning",
				},
			},
			expected: `{
				"id": "telegraf",
				"version": "$VERSION",
				"schema": $SCHEMA,
				"logs": {
					"entries": [
						{
							"time": "2025-10-22T16:02:30Z",
							"level": "ERROR",
							"source": "inputs.test::hbt",
							"attributes": {
								"alias": "hbt",
								"category": "inputs",
								"plugin": "test"
							},
							"message": "An error message"
						},
						{
							"time": "2025-10-22T16:02:31Z",
							"level": "WARN",
							"source": "inputs.test::hbt",
							"attributes": {
								"alias": "hbt",
								"category": "inputs",
								"plugin": "test"
							},
							"message": "A first warning"
						},
						{
							"time": "2025-10-22T16:02:32Z",
							"level": "INFO",
							"source": "inputs.test::hbt",
							"attributes": {
								"alias": "hbt",
								"category": "inputs",
								"plugin": "test"
							},
							"message": "An information"
						},
						{
							"time": "2025-10-22T16:02:35Z",
							"level": "WARN",
							"source": "inputs.test::hbt",
							"attributes": {
								"alias": "hbt",
								"category": "inputs",
								"plugin": "test"
							},
							"message": "A second warning"
						}
					]
				}
			}`,
		},
		{
			name:   "debug level",
			source: "inputs.test::hbt",
			logcfg: LogsConfig{
				LogLevel: "debug",
			},
			logs: []logEvent{
				{
					timestamp: logtime,
					level:     telegraf.Error,
					msg:       "An error message",
				},
				{
					timestamp: logtime.Add(1 * time.Second),
					level:     telegraf.Warn,
					msg:       "A first warning",
				},
				{
					timestamp: logtime.Add(2 * time.Second),
					level:     telegraf.Info,
					msg:       "An information",
				},
				{
					timestamp: logtime.Add(3 * time.Second),
					level:     telegraf.Debug,
					msg:       "A debug information",
				},
				{
					timestamp: logtime.Add(4 * time.Second),
					level:     telegraf.Trace,
					msg:       "A trace information",
				},
				{
					timestamp: logtime.Add(5 * time.Second),
					level:     telegraf.Warn,
					msg:       "A second warning",
				},
			},
			expected: `{
				"id": "telegraf",
				"version": "$VERSION",
				"schema": $SCHEMA,
				"logs": {
					"entries": [
						{
							"time": "2025-10-22T16:02:30Z",
							"level": "ERROR",
							"source": "inputs.test::hbt",
							"attributes": {
								"alias": "hbt",
								"category": "inputs",
								"plugin": "test"
							},
							"message": "An error message"
						},
						{
							"time": "2025-10-22T16:02:31Z",
							"level": "WARN",
							"source": "inputs.test::hbt",
							"attributes": {
								"alias": "hbt",
								"category": "inputs",
								"plugin": "test"
							},
							"message": "A first warning"
						},
						{
							"time": "2025-10-22T16:02:32Z",
							"level": "INFO",
							"source": "inputs.test::hbt",
							"attributes": {
								"alias": "hbt",
								"category": "inputs",
								"plugin": "test"
							},
							"message": "An information"
						},
						{
							"time": "2025-10-22T16:02:33Z",
							"level": "DEBUG",
							"source": "inputs.test::hbt",
							"attributes": {
								"alias": "hbt",
								"category": "inputs",
								"plugin": "test"
							},
							"message": "A debug information"
						},
						{
							"time": "2025-10-22T16:02:35Z",
							"level": "WARN",
							"source": "inputs.test::hbt",
							"attributes": {
								"alias": "hbt",
								"category": "inputs",
								"plugin": "test"
							},
							"message": "A second warning"
						}
					]
				}
			}`,
		},
		{
			name:   "trace level",
			source: "inputs.test::hbt",
			logcfg: LogsConfig{
				LogLevel: "trace",
			},
			logs: []logEvent{
				{
					timestamp: logtime,
					level:     telegraf.Error,
					msg:       "An error message",
				},
				{
					timestamp: logtime.Add(1 * time.Second),
					level:     telegraf.Warn,
					msg:       "A first warning",
				},
				{
					timestamp: logtime.Add(2 * time.Second),
					level:     telegraf.Info,
					msg:       "An information",
				},
				{
					timestamp: logtime.Add(3 * time.Second),
					level:     telegraf.Debug,
					msg:       "A debug information",
				},
				{
					timestamp: logtime.Add(4 * time.Second),
					level:     telegraf.Trace,
					msg:       "A trace information",
				},
				{
					timestamp: logtime.Add(5 * time.Second),
					level:     telegraf.Warn,
					msg:       "A second warning",
				},
			},
			expected: `{
				"id": "telegraf",
				"version": "$VERSION",
				"schema": $SCHEMA,
				"logs": {
					"entries": [
						{
							"time": "2025-10-22T16:02:30Z",
							"level": "ERROR",
							"source": "inputs.test::hbt",
							"attributes": {
								"alias": "hbt",
								"category": "inputs",
								"plugin": "test"
							},
							"message": "An error message"
						},
						{
							"time": "2025-10-22T16:02:31Z",
							"level": "WARN",
							"source": "inputs.test::hbt",
							"attributes": {
								"alias": "hbt",
								"category": "inputs",
								"plugin": "test"
							},
							"message": "A first warning"
						},
						{
							"time": "2025-10-22T16:02:32Z",
							"level": "INFO",
							"source": "inputs.test::hbt",
							"attributes": {
								"alias": "hbt",
								"category": "inputs",
								"plugin": "test"
							},
							"message": "An information"
						},
						{
							"time": "2025-10-22T16:02:33Z",
							"level": "DEBUG",
							"source": "inputs.test::hbt",
							"attributes": {
								"alias": "hbt",
								"category": "inputs",
								"plugin": "test"
							},
							"message": "A debug information"
						},
						{
							"time": "2025-10-22T16:02:34Z",
							"level": "TRACE",
							"source": "inputs.test::hbt",
							"attributes": {
								"alias": "hbt",
								"category": "inputs",
								"plugin": "test"
							},
							"message": "A trace information"
						},
						{
							"time": "2025-10-22T16:02:35Z",
							"level": "WARN",
							"source": "inputs.test::hbt",
							"attributes": {
								"alias": "hbt",
								"category": "inputs",
								"plugin": "test"
							},
							"message": "A second warning"
						}
					]
				}
			}`,
		},
		{
			name:   "trace level with limit",
			source: "inputs.test::hbt",
			logcfg: LogsConfig{
				LogLevel: "trace",
				Limit:    3,
			},
			logs: []logEvent{
				{
					timestamp: logtime,
					level:     telegraf.Error,
					msg:       "An error message",
				},
				{
					timestamp: logtime.Add(1 * time.Second),
					level:     telegraf.Warn,
					msg:       "A first warning",
				},
				{
					timestamp: logtime.Add(2 * time.Second),
					level:     telegraf.Info,
					msg:       "An information",
				},
				{
					timestamp: logtime.Add(3 * time.Second),
					level:     telegraf.Debug,
					msg:       "A debug information",
				},
				{
					timestamp: logtime.Add(4 * time.Second),
					level:     telegraf.Trace,
					msg:       "A trace information",
				},
				{
					timestamp: logtime.Add(5 * time.Second),
					level:     telegraf.Warn,
					msg:       "A second warning",
				},
			},
			expected: `{
				"id": "telegraf",
				"version": "$VERSION",
				"schema": $SCHEMA,
				"logs": {
					"entries": [
						{
							"time": "2025-10-22T16:02:30Z",
							"level": "ERROR",
							"source": "inputs.test::hbt",
							"attributes": {
								"alias": "hbt",
								"category": "inputs",
								"plugin": "test"
							},
							"message": "An error message"
						},
						{
							"time": "2025-10-22T16:02:31Z",
							"level": "WARN",
							"source": "inputs.test::hbt",
							"attributes": {
								"alias": "hbt",
								"category": "inputs",
								"plugin": "test"
							},
							"message": "A first warning"
						},
						{
							"time": "2025-10-22T16:02:35Z",
							"level": "WARN",
							"source": "inputs.test::hbt",
							"attributes": {
								"alias": "hbt",
								"category": "inputs",
								"plugin": "test"
							},
							"message": "A second warning"
						}
					]
				}
			}`,
		},
		{
			name:   "limited with most recent error messages",
			source: "inputs.test::hbt",
			logcfg: LogsConfig{
				LogLevel: "trace",
				Limit:    3,
			},
			logs: []logEvent{
				{
					timestamp: logtime,
					level:     telegraf.Error,
					msg:       "A first error message",
				},
				{
					timestamp: logtime.Add(1 * time.Second),
					level:     telegraf.Error,
					msg:       "A second error message",
				},
				{
					timestamp: logtime.Add(2 * time.Second),
					level:     telegraf.Error,
					msg:       "A third error message",
				},
				{
					timestamp: logtime.Add(3 * time.Second),
					level:     telegraf.Warn,
					msg:       "A first warning",
				},
				{
					timestamp: logtime.Add(4 * time.Second),
					level:     telegraf.Error,
					msg:       "A fourth error message",
				},
				{
					timestamp: logtime.Add(5 * time.Second),
					level:     telegraf.Warn,
					msg:       "A second warning",
				},
			},
			expected: `{
				"id": "telegraf",
				"version": "$VERSION",
				"schema": $SCHEMA,
				"logs": {
					"entries": [
						{
							"time": "2025-10-22T16:02:31Z",
							"level": "ERROR",
							"source": "inputs.test::hbt",
							"attributes": {
								"alias": "hbt",
								"category": "inputs",
								"plugin": "test"
							},
							"message": "A second error message"
						},
						{
							"time": "2025-10-22T16:02:32Z",
							"level": "ERROR",
							"source": "inputs.test::hbt",
							"attributes": {
								"alias": "hbt",
								"category": "inputs",
								"plugin": "test"
							},
							"message": "A third error message"
						},
						{
							"time": "2025-10-22T16:02:34Z",
							"level": "ERROR",
							"source": "inputs.test::hbt",
							"attributes": {
								"alias": "hbt",
								"category": "inputs",
								"plugin": "test"
							},
							"message": "A fourth error message"
						}
					]
				}
			}`,
		},
		{
			name:   "limited with most recent warning messages",
			source: "inputs.test::hbt",
			logcfg: LogsConfig{
				LogLevel: "trace",
				Limit:    5,
			},
			logs: []logEvent{
				{
					timestamp: logtime,
					level:     telegraf.Error,
					msg:       "A first error message",
				},
				{
					timestamp: logtime.Add(1 * time.Second),
					level:     telegraf.Error,
					msg:       "A second error message",
				},
				{
					timestamp: logtime.Add(2 * time.Second),
					level:     telegraf.Error,
					msg:       "A third error message",
				},
				{
					timestamp: logtime.Add(3 * time.Second),
					level:     telegraf.Warn,
					msg:       "A first warning",
				},
				{
					timestamp: logtime.Add(4 * time.Second),
					level:     telegraf.Error,
					msg:       "A fourth error message",
				},
				{
					timestamp: logtime.Add(5 * time.Second),
					level:     telegraf.Warn,
					msg:       "A second warning",
				},
			},
			expected: `{
				"id": "telegraf",
				"version": "$VERSION",
				"schema": $SCHEMA,
				"logs": {
					"entries": [
						{
							"time": "2025-10-22T16:02:30Z",
							"level": "ERROR",
							"source": "inputs.test::hbt",
							"attributes": {
								"alias": "hbt",
								"category": "inputs",
								"plugin": "test"
							},
							"message": "A first error message"
						},
						{
							"time": "2025-10-22T16:02:31Z",
							"level": "ERROR",
							"source": "inputs.test::hbt",
							"attributes": {
								"alias": "hbt",
								"category": "inputs",
								"plugin": "test"
							},
							"message": "A second error message"
						},
						{
							"time": "2025-10-22T16:02:32Z",
							"level": "ERROR",
							"source": "inputs.test::hbt",
							"attributes": {
								"alias": "hbt",
								"category": "inputs",
								"plugin": "test"
							},
							"message": "A third error message"
						},
						{
							"time": "2025-10-22T16:02:34Z",
							"level": "ERROR",
							"source": "inputs.test::hbt",
							"attributes": {
								"alias": "hbt",
								"category": "inputs",
								"plugin": "test"
							},
							"message": "A fourth error message"
						},
						{
							"time": "2025-10-22T16:02:35Z",
							"level": "WARN",
							"source": "inputs.test::hbt",
							"attributes": {
								"alias": "hbt",
								"category": "inputs",
								"plugin": "test"
							},
							"message": "A second warning"
						}
					]
				}
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
				Include:    []string{"log-details"},
				Logs:       tt.logcfg,
				Log:        &testutil.Logger{},
			}
			require.NoError(t, plugin.Init())

			// Register the logging handler early to avoid race conditions
			// during testing. This is not ideal but the only way to get
			// reliable tests without a race between the Connect call,
			// registering the callback and the actual logging.
			plugin.logCallbackID = logger.AddCallback(plugin.handleLogEvent)

			// Setup the logger
			category, remaining, found := strings.Cut(tt.source, ".")
			var name, alias string
			if found {
				name, alias, _ = strings.Cut(remaining, ":")
				alias = strings.TrimLeft(alias, ":")
			}
			logger := logger.New(category, name, alias)
			for k, v := range tt.attrs {
				logger.AddAttribute(k, v)
			}

			// Log the given messages
			for _, e := range tt.logs {
				logger.Print(e.level, e.timestamp, e.msg)
			}

			// Start processing
			require.NoError(t, plugin.Connect())
			defer plugin.Close()

			// Wait for the data to arrive at the test-server and check the
			// payload we got.
			require.Eventually(t, func() bool {
				return done.Load()
			}, 3*time.Second, 100*time.Millisecond)
			require.JSONEq(t, expected, actual, actual)
		})
	}
}
