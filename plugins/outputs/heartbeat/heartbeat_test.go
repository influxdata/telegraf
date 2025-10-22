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
	"sync"
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
		Include:    []string{"configs", "hostname", "logs", "log-details", "metrics", "status"},
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
	defer clear(config.Sources)
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

func TestSending(t *testing.T) {
	// Prepare JSON schema for validation of messages
	schema, err := jsonschema.Compile(fmt.Sprintf("schema_v%d.json", jsonSchemaVersion))
	require.NoError(t, err)

	// Prepare a string replacer to replace dynamic content such as the hostname
	// in the expected strings
	replacer := strings.NewReplacer(
		"$VERSION", internal.FormatFullVersion(),
		"$SCHEMA", strconv.Itoa(jsonSchemaVersion),
	)

	// Expect a minumal message
	expected := replacer.Replace(`{
			  "id": "telegraf",
			  "version": "$VERSION",
			  "schema": $SCHEMA
			}`)

	// Create a test server to validate the data sent
	var received atomic.Uint64
	var receivedMessage string
	var receivedMu sync.Mutex
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
		receivedMu.Lock()
		receivedMessage = string(body)
		receivedMu.Unlock()
		received.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	// Initialize the plugin
	u := config.NewSecret([]byte("http://" + ts.Listener.Addr().String()))
	defer u.Destroy()

	plugin := &Heartbeat{
		URL:        u,
		InstanceID: "telegraf",
		Interval:   config.Duration(100 * time.Millisecond),
		Log:        &testutil.Logger{},
	}
	require.NoError(t, plugin.Init())
	require.NoError(t, plugin.Connect())
	defer plugin.Close()

	// Wait for some heartbeats to arrive. Time the waiting period such that
	// there is enough headroom for receiving the requested amount of beats.
	require.Eventually(t, func() bool {
		return received.Load() > 10
	}, 2*time.Second, 100*time.Millisecond)

	receivedMu.Lock()
	actual := receivedMessage
	receivedMu.Unlock()
	require.JSONEq(t, expected, actual, actual)

	// Check heartbeat message against the JSON schema
	var v interface{}
	require.NoError(t, json.Unmarshal([]byte(actual), &v))
	require.NoError(t, schema.Validate(v))
}

func TestSendingFail(t *testing.T) {
	// Prepare JSON schema for validation of messages
	schema, err := jsonschema.Compile(fmt.Sprintf("schema_v%d.json", jsonSchemaVersion))
	require.NoError(t, err)

	// Prepare a string replacer to replace dynamic content such as the hostname
	// in the expected strings
	logtime := time.Now()
	replacer := strings.NewReplacer(
		"$VERSION", internal.FormatFullVersion(),
		"$SCHEMA", strconv.Itoa(jsonSchemaVersion),
		"$LOGTIME", logtime.UTC().Format(time.RFC3339Nano),
	)

	// Create a test server to validate the data sent
	var received atomic.Uint64
	var receivedUpdate atomic.Uint64
	var receivedMessage string
	var receivedMu sync.Mutex
	var fail atomic.Bool
	var snapshot atomic.Bool
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fail()
			w.WriteHeader(http.StatusMethodNotAllowed)
		}

		// Count total messages received
		received.Add(1)

		// Send a fail response
		if fail.Load() {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		// Take a snapshot of the received data
		if snapshot.Swap(false) {
			// Decode the body
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fail()
				w.WriteHeader(http.StatusInternalServerError)
			}
			receivedMu.Lock()
			receivedMessage = string(body)
			receivedMu.Unlock()
			receivedUpdate.Add(1)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	// Initialize the plugin
	u := config.NewSecret([]byte("http://" + ts.Listener.Addr().String()))
	defer u.Destroy()

	var log testutil.CaptureLogger
	plugin := &Heartbeat{
		URL:        u,
		InstanceID: "telegraf",
		Interval:   config.Duration(250 * time.Millisecond),
		Include:    []string{"logs", "log-details", "metrics"},
		Log:        &log,
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
	logProducer := logger.New("inputs", "test", "hbt")
	logProducer.AddAttribute("source", "heartbeat")
	logProducer.AddAttribute("type", "testing")
	logProducer.Print(telegraf.Error, logtime, "An error message")
	logProducer.Print(telegraf.Warn, logtime, "A first warning")
	logProducer.Print(telegraf.Info, logtime, "An information")
	logProducer.Print(telegraf.Debug, logtime, "A debug information")
	logProducer.Print(telegraf.Trace, logtime, "A trace information")
	logProducer.Print(telegraf.Warn, logtime, "A second warning")

	require.NoError(t, plugin.Connect())
	defer plugin.Close()

	// Wait for the data to arrive at the test-server
	snapshot.Store(true)
	require.Eventually(t, func() bool {
		return receivedUpdate.Load() > 0
	}, time.Second, 100*time.Millisecond)

	// Test for the first successful message
	expected := replacer.Replace(`
	{
		"id": "telegraf",
		"version": "$VERSION",
		"schema": $SCHEMA,
		"metrics": 5,
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
	}`)

	receivedMu.Lock()
	actual := receivedMessage
	receivedMu.Unlock()
	require.JSONEq(t, expected, actual, actual)

	// Check heartbeat message against the JSON schema
	var v interface{}
	require.NoError(t, json.Unmarshal([]byte(actual), &v))
	require.NoError(t, schema.Validate(v))

	// Make the next message fail and wait until sending is complete
	fail.Store(true)
	require.Eventually(t, func() bool {
		for _, e := range log.Errors() {
			if strings.Contains(e, "received status 503") {
				return true
			}
		}
		return false
	}, time.Second, 100*time.Millisecond)

	// Add some more log messages
	logProducer.Print(telegraf.Error, logtime, "An error message logged during failing sends")
	received.Store(0)
	require.Eventually(t, func() bool {
		return received.Swap(0) > 0
	}, time.Second, 100*time.Millisecond)
	logProducer.Print(telegraf.Error, logtime, "Another error message logged during failing sends")
	require.Eventually(t, func() bool {
		return received.Swap(0) > 0
	}, time.Second, 100*time.Millisecond)

	// Check the update is marked as failed
	plugin.Lock()
	lastUpdate := plugin.stats.lastUpdate.Unix()
	lastUpdateFailed := plugin.stats.lastUpdateFailed
	plugin.Unlock()
	require.True(t, lastUpdateFailed, "last update not marked as failed")

	// Reset server to receive successfully and wait for the next update
	receivedUpdate.Store(0)
	snapshot.Store(true)
	fail.Store(false)
	require.Eventually(t, func() bool {
		return receivedUpdate.Load() > 0
	}, time.Second, 100*time.Millisecond)
	receivedMu.Lock()
	actual = receivedMessage
	receivedMu.Unlock()
	plugin.Close()

	// Check the message against the schema
	require.NoError(t, json.Unmarshal([]byte(actual), &v))
	require.NoError(t, schema.Validate(v))

	// Decode the received message for evaluation
	var msg message
	require.NoError(t, json.Unmarshal([]byte(actual), &msg))
	require.NotNil(t, msg.LastSuccessfulUpdate)
	require.Equal(t, *msg.LastSuccessfulUpdate, lastUpdate)
	require.NotNil(t, msg.Logs)
	require.NotZero(t, *msg.Logs.Errors)
	require.NotNil(t, msg.Logs.Entries)
	require.NotEmpty(t, *msg.Logs.Entries)

	logmsg := make([]string, 0, len(*msg.Logs.Entries))
	for _, e := range *msg.Logs.Entries {
		logmsg = append(logmsg, e.Messsage)
	}
	require.Contains(t, logmsg, "An error message logged during failing sends")
	require.Contains(t, logmsg, "Another error message logged during failing sends")
}
