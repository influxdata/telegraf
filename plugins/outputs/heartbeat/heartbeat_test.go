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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.plugin.Log = &testutil.Logger{}
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
		Include:    []string{"configs", "hostname", "statistics"},
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
	absdir = filepath.ToSlash(absdir)
	cfgs := []string{
		"testdata/telegraf.conf",
		absdir + "/telegraf.d/inputs.conf",
		"http://user:password@" + cfgServer.Listener.Addr().String(),
		absdir + "/telegraf.d/outputs.conf",
		"http://" + cfgServer.Listener.Addr().String() + "/myconfigs",
		"testdata/non_existing.conf",
	}

	// Make sure the configuration part also works on Windows
	for i, cfg := range cfgs {
		if !strings.HasPrefix(cfg, "http://") {
			cfgs[i] = filepath.FromSlash(cfg)
		}
	}

	cfg := config.NewConfig()
	for _, c := range cfgs {
		//nolint:errcheck // Ignore error on purpose as some endpoints won't be loadable
		cfg.LoadConfig(c)
	}
	cfgServer.Close()

	// Expected configs
	cfgsExpected := []string{
		`"testdata/telegraf.conf"`,
		`"` + absdir + `/telegraf.d/inputs.conf"`,
		`"http://user:xxxxx@` + cfgServer.Listener.Addr().String() + `"`,
		`"` + absdir + `/telegraf.d/outputs.conf"`,
		`"http://` + cfgServer.Listener.Addr().String() + `/myconfigs"`,
	}

	// Make sure the configuration part also works on Windows
	for i, cfg := range cfgsExpected {
		if !strings.HasPrefix(cfg, `"http://`) {
			cfgsExpected[i] = strings.ReplaceAll(filepath.FromSlash(cfg), `\`, `\\`)
		}
	}

	// Prepare a string replacer to replace dynamic content such as the hostname
	// in the expected strings
	logtime := time.Now()
	replacer := strings.NewReplacer(
		"$HOSTNAME", hostname,
		"$VERSION", internal.FormatFullVersion(),
		"$SCHEMA", strconv.Itoa(jsonSchemaVersion),
		"$CONFIGS", strings.Join(cfgsExpected, ","),
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
			name:     "all",
			includes: []string{"configs", "hostname", "statistics"},
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
		Include:    []string{"statistics"},
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
	require.Equal(t, uint64(0), msg.Statistics.Warnings)
}
