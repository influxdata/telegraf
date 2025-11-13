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
	"github.com/influxdata/telegraf/selfstat"
	"github.com/influxdata/telegraf/testutil"
)

func TestInitSuccess(t *testing.T) {
	u := config.NewSecret([]byte("http://localhost/heartbeat"))
	t.Cleanup(u.Destroy)

	plugin := &Heartbeat{
		URL:        u,
		InstanceID: "telegraf",
		Interval:   config.Duration(10 * time.Second),
		Log:        &testutil.Logger{},
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
		{
			name: "invalid log level",
			plugin: &Heartbeat{
				URL:        u,
				InstanceID: "telegraf",
				Interval:   config.Duration(10 * time.Second),
				Logs: LogsConfig{
					LogLevel: "foo",
				},
				Include: []string{"logs"},
			},
			expected: "invalid log-level",
		},
		{
			name: "invalid initial status",
			plugin: &Heartbeat{
				URL:        u,
				InstanceID: "telegraf",
				Interval:   config.Duration(10 * time.Second),
				Status: StatusConfig{
					Initial: "foo",
				},
				Include: []string{"status"},
			},
			expected: "invalid status 'initial' value",
		},
		{
			name: "invalid default status",
			plugin: &Heartbeat{
				URL:        u,
				InstanceID: "telegraf",
				Interval:   config.Duration(10 * time.Second),
				Status: StatusConfig{
					Default: "foo",
				},
				Include: []string{"status"},
			},
			expected: "invalid status 'default' value",
		},
		{
			name: "invalid status in order",
			plugin: &Heartbeat{
				URL:        u,
				InstanceID: "telegraf",
				Interval:   config.Duration(10 * time.Second),
				Status: StatusConfig{
					Order: []string{"foo"},
				},
				Include: []string{"status"},
			},
			expected: "invalid status 'order' value",
		},
		{
			name: "duplicate status in order",
			plugin: &Heartbeat{
				URL:        u,
				InstanceID: "telegraf",
				Interval:   config.Duration(10 * time.Second),
				Status: StatusConfig{
					Order: []string{"ok", "warn", "ok", "fail"},
				},
				Include: []string{"status"},
			},
			expected: "duplicate value \"ok\" in status 'order'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.plugin.Log = &testutil.Logger{}
			require.ErrorContains(t, tt.plugin.Init(), tt.expected)
		})
	}
}

func TestInitStatusWarning(t *testing.T) {
	u := config.NewSecret([]byte("http://localhost/heartbeat"))
	t.Cleanup(u.Destroy)

	log := &testutil.CaptureLogger{Name: "outputs.heartbeat"}
	plugin := &Heartbeat{
		URL:        u,
		InstanceID: "telegraf",
		Interval:   config.Duration(10 * time.Second),
		Status: StatusConfig{
			Ok:    "true",
			Warn:  "true",
			Fail:  "true",
			Order: []string{"ok", "warn"},
		},
		Include: []string{"status"},
		Log:     log,
	}
	require.NoError(t, plugin.Init())

	require.Eventually(t, func() bool {
		return len(log.Warnings()) > 0
	}, 3*time.Second, 100*time.Millisecond)

	expected := `W! [outputs.heartbeat] condition for status "fail" will be ignored as it is not in the 'order' list`
	require.Contains(t, log.Warnings(), expected)
}

func TestIncludes(t *testing.T) {
	u := config.NewSecret([]byte("http://localhost/heartbeat"))
	t.Cleanup(u.Destroy)

	plugin := &Heartbeat{
		URL:        u,
		InstanceID: "telegraf",
		Interval:   config.Duration(10 * time.Second),
		Include:    []string{"configs", "hostname", "logs", "statistics", "status"},
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
			name:     "statistics",
			includes: []string{"statistics"},
			expected: `{
			  "id": "telegraf",
			  "version": "$VERSION",
			  "schema": $SCHEMA,
			  "statistics": {
				"errors": 1,
				"warnings": 2,
				"metrics": 5
			  }
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
				"logs": [
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
			}`,
		},
		{
			name:     "all",
			includes: []string{"configs", "hostname", "logs", "status", "statistics"},
			expected: `{
			  "id": "telegraf",
			  "version": "$VERSION",
			  "schema": $SCHEMA,
			  "hostname": "$HOSTNAME",
			  "statistics": {
				"errors": 1,
				"warnings": 2,
				"metrics": 5
			  },
			  "status": "OK",
			  "configurations": [$CONFIGS],
			  "logs": [
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
			id, err := logger.AddCallback(plugin.handleLogEvent)
			require.NoError(t, err)
			defer logger.RemoveCallback(id)
			plugin.logCallbackID = id

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
			if slices.Contains(tt.includes, "logs") || slices.Contains(tt.includes, "statistics") {
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
				"logs": [
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
				"logs": [
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
				"logs": [
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
				"logs": [
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
				"logs": [
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
				"logs": [
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
				"logs": [
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
				"logs": [
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
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Inject dynamic content into the expectation
			expected := replacer.Replace(tt.expected)

			// Create a test server to validate the data sent
			var receivedMessage string
			var receivedMu sync.Mutex
			var snapshot atomic.Bool
			var done atomic.Bool
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Fail()
					w.WriteHeader(http.StatusMethodNotAllowed)
				}

				// Decode the body
				if snapshot.Swap(false) {
					body, err := io.ReadAll(r.Body)
					if err != nil {
						t.Fail()
						w.WriteHeader(http.StatusInternalServerError)
					}
					receivedMu.Lock()
					receivedMessage = string(body)
					receivedMu.Unlock()
					done.Store(true)
				}
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
				Include:    []string{"logs"},
				Logs:       tt.logcfg,
				Log:        &testutil.Logger{},
			}
			require.NoError(t, plugin.Init())

			// Register the logging handler early to avoid race conditions
			// during testing. This is not ideal but the only way to get
			// reliable tests without a race between the Connect call,
			// registering the callback and the actual logging.
			id, err := logger.AddCallback(plugin.handleLogEvent)
			require.NoError(t, err)
			defer logger.RemoveCallback(id)
			plugin.logCallbackID = id

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
			snapshot.Store(true)
			require.NoError(t, plugin.Connect())
			defer plugin.Close()

			// Wait for the data to arrive at the test-server and check the
			// payload we got.
			require.Eventually(t, func() bool {
				return done.Load()
			}, 3*time.Second, 100*time.Millisecond)

			receivedMu.Lock()
			actual := receivedMessage
			receivedMu.Unlock()
			require.JSONEq(t, expected, actual, actual)
		})
	}
}

func TestStatusComputation(t *testing.T) {
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

	// Compile the JSON schema for evaluation
	schema, err := jsonschema.Compile(fmt.Sprintf("schema_v%d.json", jsonSchemaVersion))
	require.NoError(t, err)

	tests := []struct {
		name     string
		stats    *statistics
		inputs   []*inputStats
		outputs  []*outputStats
		cfg      StatusConfig
		expected string
	}{
		{
			name: "default config",
			expected: `{
				"id": "telegraf",
				"version": "$VERSION",
				"schema": $SCHEMA,
				"status": "OK"
			}`,
		},
		{
			name: "initial undefined",
			cfg: StatusConfig{
				Initial: "undefined",
			},
			expected: `{
				"id": "telegraf",
				"version": "$VERSION",
				"schema": $SCHEMA,
				"status": "UNDEFINED"
			}`,
		},
		{
			name: "default undefined",
			cfg: StatusConfig{
				Default: "undefined",
			},
			expected: `{
				"id": "telegraf",
				"version": "$VERSION",
				"schema": $SCHEMA,
				"status": "UNDEFINED"
			}`,
		},
		{
			name: "default with order",
			cfg: StatusConfig{
				Default: "undefined",
				Order:   []string{"ok", "warn", "fail"},
			},
			expected: `{
				"id": "telegraf",
				"version": "$VERSION",
				"schema": $SCHEMA,
				"status": "UNDEFINED"
			}`,
		},
		{
			name: "matching ok",
			cfg: StatusConfig{
				Ok:    "true",
				Order: []string{"ok", "warn", "fail"},
			},
			expected: `{
				"id": "telegraf",
				"version": "$VERSION",
				"schema": $SCHEMA,
				"status": "OK"
			}`,
		},
		{
			name: "matching warn",
			cfg: StatusConfig{
				Warn:  "true",
				Order: []string{"ok", "warn", "fail"},
			},
			expected: `{
				"id": "telegraf",
				"version": "$VERSION",
				"schema": $SCHEMA,
				"status": "WARN"
			}`,
		},
		{
			name: "matching fail",
			cfg: StatusConfig{
				Fail:  "true",
				Order: []string{"ok", "warn", "fail"},
			},
			expected: `{
				"id": "telegraf",
				"version": "$VERSION",
				"schema": $SCHEMA,
				"status": "FAIL"
			}`,
		},
		{
			name: "none matching",
			cfg: StatusConfig{
				Ok:      "false",
				Warn:    "false",
				Fail:    "false",
				Order:   []string{"ok", "warn", "fail"},
				Default: "undefined",
			},
			expected: `{
				"id": "telegraf",
				"version": "$VERSION",
				"schema": $SCHEMA,
				"status": "UNDEFINED"
			}`,
		},
		{
			name: "matching all ascending order",
			cfg: StatusConfig{
				Ok:      "true",
				Warn:    "true",
				Fail:    "true",
				Order:   []string{"ok", "warn", "fail"},
				Default: "undefined",
			},
			expected: `{
				"id": "telegraf",
				"version": "$VERSION",
				"schema": $SCHEMA,
				"status": "OK"
			}`,
		},
		{
			name: "matching all descending order",
			cfg: StatusConfig{
				Ok:      "true",
				Warn:    "true",
				Fail:    "true",
				Order:   []string{"fail", "warn", "ok"},
				Default: "undefined",
			},
			expected: `{
				"id": "telegraf",
				"version": "$VERSION",
				"schema": $SCHEMA,
				"status": "FAIL"
			}`,
		},
		{
			name: "no logs and received enough metrics",
			cfg: StatusConfig{
				Ok:      "metrics >= 5 && log_errors == 0 && log_warnings == 0",
				Warn:    "(metrics > 0 && metrics < 5) || (log_warnings > 0 && log_errors == 0)",
				Fail:    "metrics == 0 || log_errors > 0",
				Order:   []string{"ok", "warn", "fail"},
				Default: "undefined",
			},
			stats: &statistics{
				metrics:     10,
				logErrors:   0,
				logWarnings: 0,
			},
			expected: `{
				"id": "telegraf",
				"version": "$VERSION",
				"schema": $SCHEMA,
				"status": "OK"
			}`,
		},
		{
			name: "warning logs and received enough metrics",
			cfg: StatusConfig{
				Ok:      "metrics >= 5 && log_errors == 0 && log_warnings == 0",
				Warn:    "(metrics > 0 && metrics < 5) || (log_warnings > 0 && log_errors == 0)",
				Fail:    "metrics == 0 || log_errors > 0",
				Order:   []string{"ok", "warn", "fail"},
				Default: "undefined",
			},
			stats: &statistics{
				metrics:     10,
				logErrors:   0,
				logWarnings: 4,
			},
			expected: `{
				"id": "telegraf",
				"version": "$VERSION",
				"schema": $SCHEMA,
				"status": "WARN"
			}`,
		},
		{
			name: "error logs and received enough metrics",
			cfg: StatusConfig{
				Ok:      "metrics >= 5 && log_errors == 0 && log_warnings == 0",
				Warn:    "(metrics > 0 && metrics < 5) || (log_warnings > 0 && log_errors == 0)",
				Fail:    "metrics == 0 || log_errors > 0",
				Order:   []string{"ok", "warn", "fail"},
				Default: "undefined",
			},
			stats: &statistics{
				metrics:     10,
				logErrors:   2,
				logWarnings: 4,
			},
			expected: `{
				"id": "telegraf",
				"version": "$VERSION",
				"schema": $SCHEMA,
				"status": "FAIL"
			}`,
		},
		{
			name: "not enough metrics",
			cfg: StatusConfig{
				Ok:      "metrics >= 5 && log_errors == 0 && log_warnings == 0",
				Warn:    "(metrics > 0 && metrics < 5) || (log_warnings > 0 && log_errors == 0)",
				Fail:    "metrics == 0 || log_errors > 0",
				Order:   []string{"ok", "warn", "fail"},
				Default: "undefined",
			},
			stats: &statistics{
				metrics:     2,
				logErrors:   0,
				logWarnings: 0,
			},
			expected: `{
				"id": "telegraf",
				"version": "$VERSION",
				"schema": $SCHEMA,
				"status": "WARN"
			}`,
		},
		{
			name: "no metrics",
			cfg: StatusConfig{
				Ok:      "metrics >= 5 && log_errors == 0 && log_warnings == 0",
				Warn:    "(metrics > 0 && metrics < 5) || (log_warnings > 0 && log_errors == 0)",
				Fail:    "metrics == 0 || log_errors > 0",
				Order:   []string{"ok", "warn", "fail"},
				Default: "undefined",
			},
			stats: &statistics{
				metrics:     0,
				logErrors:   0,
				logWarnings: 0,
			},
			expected: `{
				"id": "telegraf",
				"version": "$VERSION",
				"schema": $SCHEMA,
				"status": "FAIL"
			}`,
		},
		{
			name: "ambiguous conditions ascending",
			cfg: StatusConfig{
				Ok:      "metrics > 5",
				Warn:    "metrics > 2",
				Order:   []string{"ok", "warn", "fail"},
				Default: "undefined",
			},
			stats: &statistics{
				metrics:     6,
				logErrors:   0,
				logWarnings: 0,
			},
			expected: `{
				"id": "telegraf",
				"version": "$VERSION",
				"schema": $SCHEMA,
				"status": "OK"
			}`,
		},
		{
			name: "ambiguous conditions descending",
			cfg: StatusConfig{
				Ok:      "metrics > 5",
				Warn:    "metrics > 2",
				Order:   []string{"fail", "warn", "ok"},
				Default: "undefined",
			},
			stats: &statistics{
				metrics:     6,
				logErrors:   0,
				logWarnings: 0,
			},
			expected: `{
				"id": "telegraf",
				"version": "$VERSION",
				"schema": $SCHEMA,
				"status": "WARN"
			}`,
		},
		{
			name: "ambiguous conditions descending to default",
			cfg: StatusConfig{
				Ok:      "metrics > 5",
				Warn:    "metrics > 2",
				Order:   []string{"fail", "warn", "ok"},
				Default: "fail",
			},
			stats: &statistics{
				metrics:     1,
				logErrors:   0,
				logWarnings: 0,
			},
			expected: `{
				"id": "telegraf",
				"version": "$VERSION",
				"schema": $SCHEMA,
				"status": "FAIL"
			}`,
		},
		{
			name: "single plugin statistic",
			cfg: StatusConfig{
				Warn:    `outputs["file"][0].buffer_fullness > 60`,
				Fail:    `outputs["file"][0].buffer_fullness > 80`,
				Order:   []string{"fail", "warn", "ok"},
				Default: "ok",
			},
			stats: &statistics{
				metrics:     1,
				logErrors:   0,
				logWarnings: 0,
			},
			outputs: []*outputStats{
				{
					name:        "influxdb",
					alias:       "server1",
					id:          "0xabc",
					bufferSize:  1234,
					bufferLimit: 10000,
				},
				{
					name:        "influxdb",
					alias:       "server2",
					id:          "0x123",
					bufferSize:  9999,
					bufferLimit: 10000,
				},
				{
					name:        "file",
					id:          "0xdeadc0de",
					bufferSize:  0,
					bufferLimit: 5000,
				},
			},
			expected: `{
				"id": "telegraf",
				"version": "$VERSION",
				"schema": $SCHEMA,
				"status": "OK"
			}`,
		},
		{
			name: "total buffer fullness",
			cfg: StatusConfig{
				Warn:    `outputs.exists(k, outputs[k].exists(p, p.buffer_fullness > 0.6))`,
				Fail:    `outputs.exists(k, outputs[k].exists(p, p.buffer_fullness > 0.8))`,
				Order:   []string{"fail", "warn", "ok"},
				Default: "ok",
			},
			stats: &statistics{
				metrics:     1,
				logErrors:   0,
				logWarnings: 0,
			},
			outputs: []*outputStats{
				{
					name:        "influxdb",
					alias:       "server1",
					id:          "0xabc",
					bufferSize:  1234,
					bufferLimit: 10000,
				},
				{
					name:        "influxdb",
					alias:       "server2",
					id:          "0x123",
					bufferSize:  9999,
					bufferLimit: 10000,
				},
				{
					name:        "file",
					id:          "0xdeadc0de",
					bufferSize:  0,
					bufferLimit: 5000,
				},
			},
			expected: `{
				"id": "telegraf",
				"version": "$VERSION",
				"schema": $SCHEMA,
				"status": "FAIL"
			}`,
		},
		{
			name: "inputs collection errors",
			cfg: StatusConfig{
				Warn:    `inputs.exists(k, inputs[k].exists(p, p.errors > 0))`,
				Fail:    `inputs.exists(k, inputs[k].exists(p, p.errors > 5))`,
				Order:   []string{"fail", "warn", "ok"},
				Default: "ok",
			},
			stats: &statistics{
				metrics:     1,
				logErrors:   0,
				logWarnings: 0,
			},
			inputs: []*inputStats{
				{
					name: "cpu",
					id:   "0xabc",
				},
				{
					name: "mem",
					id:   "0x123",
				},
				{
					name:   "file",
					id:     "0xdeadc0de",
					errors: 3,
				},
			},
			expected: `{
				"id": "telegraf",
				"version": "$VERSION",
				"schema": $SCHEMA,
				"status": "WARN"
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Inject dynamic content into the expectation
			expected := replacer.Replace(tt.expected)

			// Create a test server to validate the data sent
			var actual string
			var actualMu sync.Mutex
			var snapshot atomic.Bool
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Fail()
					w.WriteHeader(http.StatusMethodNotAllowed)
				}

				// Decode the body
				if snapshot.Swap(false) {
					body, err := io.ReadAll(r.Body)
					if err != nil {
						t.Fail()
						w.WriteHeader(http.StatusInternalServerError)
					}
					actualMu.Lock()
					actual = string(body)
					actualMu.Unlock()
				}
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
				Include:    []string{"status"},
				Status:     tt.cfg,
				Log:        &testutil.Logger{},
			}
			require.NoError(t, plugin.Init())

			// Override the statistics for the computation
			if tt.stats != nil {
				plugin.stats.metrics = tt.stats.metrics
				plugin.stats.logErrors = tt.stats.logErrors
				plugin.stats.logWarnings = tt.stats.logWarnings
				plugin.stats.lastUpdate = tt.stats.lastUpdate
				plugin.stats.lastUpdateFailed = tt.stats.lastUpdateFailed
			}

			// Register plugin statistics of any
			for _, s := range tt.inputs {
				s.register()
			}
			defer func() {
				for _, s := range tt.inputs {
					s.unregister()
				}
			}()
			for _, s := range tt.outputs {
				s.register()
			}
			defer func() {
				for _, s := range tt.outputs {
					s.unregister()
				}
			}()

			// Start processing
			snapshot.Store(true)
			require.NoError(t, plugin.Connect())
			defer plugin.Close()

			// Wait for the data to arrive at the test-server and check the
			// payload we got.
			require.Eventually(t, func() bool {
				actualMu.Lock()
				defer actualMu.Unlock()
				return actual != ""
			}, 3*time.Second, 100*time.Millisecond)

			actualMu.Lock()
			defer actualMu.Unlock()

			// Check heartbeat message against the JSON schema
			var v interface{}
			require.NoError(t, json.Unmarshal([]byte(actual), &v))
			require.NoError(t, schema.Validate(v))
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
		Include:    []string{"logs", "statistics"},
		Log:        &log,
	}
	require.NoError(t, plugin.Init())

	// Register the logging handler early to avoid race conditions
	// during testing. This is not ideal but the only way to get
	// reliable tests without a race between the Connect call,
	// registering the callback and the actual logging.
	id, err := logger.AddCallback(plugin.handleLogEvent)
	require.NoError(t, err)
	defer logger.RemoveCallback(id)
	plugin.logCallbackID = id

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
		"statistics": {
		  	"errors": 1,
			"warnings": 2,
			"metrics": 5
		},
		"logs": [
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

	// Add some more log messages and metrics
	logProducer.Print(telegraf.Error, logtime, "An error message logged during failing sends")
	require.NoError(t, plugin.Write([]telegraf.Metric{testutil.TestMetric(5)}))
	received.Store(0)
	require.Eventually(t, func() bool {
		return received.Swap(0) > 0
	}, time.Second, 100*time.Millisecond)
	logProducer.Print(telegraf.Error, logtime, "Another error message logged during failing sends")
	require.NoError(t, plugin.Write([]telegraf.Metric{testutil.TestMetric(6)}))
	require.Eventually(t, func() bool {
		return received.Swap(0) > 0
	}, time.Second, 100*time.Millisecond)

	// Check the update is marked as failed
	plugin.stats.Lock()
	lastUpdate := plugin.stats.lastUpdate.Unix()
	lastUpdateFailed := plugin.stats.lastUpdateFailed
	plugin.stats.Unlock()
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
	require.NotNil(t, msg.Statistics)
	require.Equal(t, uint64(2), msg.Statistics.Metrics)
	require.Equal(t, uint64(5), msg.Statistics.Errors)
	require.NotNil(t, msg.Logs)
	require.NotEmpty(t, *msg.Logs)
	require.Equal(t, uint64(0), msg.Statistics.Warnings)
	logmsg := make([]string, 0, len(*msg.Logs))
	for _, e := range *msg.Logs {
		logmsg = append(logmsg, e.Message)
	}
	require.Contains(t, logmsg, "An error message logged during failing sends")
	require.Contains(t, logmsg, "Another error message logged during failing sends")
}

type inputStats struct {
	name  string
	id    string
	alias string

	// Model stats
	errors          int64
	metricsGathered int64
	gatherTime      int64
	gatherTimeouts  int64
	startupErrors   int64
}

func (s *inputStats) register() {
	tags := map[string]string{
		"_id":   s.id,
		"input": s.name,
	}
	if s.alias != "" {
		tags["alias"] = s.alias
	}

	// Register and set model stats
	selfstat.Register("gather", "errors", tags).Set(s.errors)
	selfstat.Register("gather", "metrics_gathered", tags).Set(s.metricsGathered)
	selfstat.RegisterTiming("gather", "gather_time_ns", tags).Set(s.gatherTime)
	selfstat.Register("gather", "gather_timeouts", tags).Set(s.gatherTimeouts)
	selfstat.Register("gather", "startup_errors", tags).Set(s.startupErrors)
}

func (s *inputStats) unregister() {
	tags := map[string]string{
		"_id":   s.id,
		"input": s.name,
	}
	if s.alias != "" {
		tags["alias"] = s.alias
	}

	// Register and set model stats
	selfstat.Unregister("gather", "errors", tags)
	selfstat.Unregister("gather", "metrics_gathered", tags)
	selfstat.Unregister("gather", "gather_time_ns", tags)
	selfstat.Unregister("gather", "gather_timeouts", tags)
	selfstat.Unregister("gather", "startup_errors", tags)
}

type outputStats struct {
	name  string
	id    string
	alias string

	// Model stats
	errors          int64
	metricsFiltered int64
	writeTime       int64
	startupErrors   int64

	// Buffer stats
	metricsAdded    int64
	metricsWritten  int64
	metricsRejected int64
	metricsDropped  int64
	bufferSize      int64
	bufferLimit     int64
}

func (s *outputStats) register() {
	tags := map[string]string{
		"_id":    s.id,
		"output": s.name,
	}
	if s.alias != "" {
		tags["alias"] = s.alias
	}

	// Register and set model stats
	selfstat.Register("write", "errors", tags).Set(s.errors)
	selfstat.Register("write", "metrics_filtered", tags).Set(s.metricsFiltered)
	selfstat.RegisterTiming("write", "write_time_ns", tags).Set(s.writeTime)
	selfstat.Register("write", "startup_errors", tags).Set(s.startupErrors)

	// Register and set buffer stats
	selfstat.Register("write", "metrics_added", tags).Set(s.metricsAdded)
	selfstat.Register("write", "metrics_written", tags).Set(s.metricsWritten)
	selfstat.Register("write", "metrics_rejected", tags).Set(s.metricsRejected)
	selfstat.Register("write", "metrics_dropped", tags).Set(s.metricsDropped)
	selfstat.Register("write", "buffer_size", tags).Set(s.bufferSize)
	selfstat.Register("write", "buffer_limit", tags).Set(s.bufferLimit)
}

func (s *outputStats) unregister() {
	tags := map[string]string{
		"_id":    s.id,
		"output": s.name,
	}
	if s.alias != "" {
		tags["alias"] = s.alias
	}

	// Register and set model stats
	selfstat.Unregister("write", "errors", tags)
	selfstat.Unregister("write", "metrics_filtered", tags)
	selfstat.Unregister("write", "write_time_ns", tags)
	selfstat.Unregister("write", "startup_errors", tags)

	// Register and set buffer stats
	selfstat.Unregister("write", "metrics_added", tags)
	selfstat.Unregister("write", "metrics_written", tags)
	selfstat.Unregister("write", "metrics_rejected", tags)
	selfstat.Unregister("write", "metrics_dropped", tags)
	selfstat.Unregister("write", "buffer_size", tags)
	selfstat.Unregister("write", "buffer_limit", tags)
}
