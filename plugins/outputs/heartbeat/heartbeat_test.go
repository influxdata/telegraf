package heartbeat

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/santhosh-tekuri/jsonschema/v5"
	"github.com/stretchr/testify/require"

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
		Include:    []string{"hostname"},
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
			name:     "all",
			includes: []string{"hostname"},
			expected: `{
			  "id": "telegraf",
			  "version": "$VERSION",
			  "schema": $SCHEMA,
			  "hostname": "$HOSTNAME"
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
		Include:    make([]string, 0),
		Log:        &log,
	}
	require.NoError(t, plugin.Init())
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
		"schema": $SCHEMA
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
}
