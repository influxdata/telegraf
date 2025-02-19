package syslog

import (
	"crypto/tls"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/leodido/go-syslog/v4/nontransparent"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/common/socket"
	"github.com/influxdata/telegraf/plugins/inputs"
	parsers_influx_upstream "github.com/influxdata/telegraf/plugins/parsers/influx/influx_upstream"
	"github.com/influxdata/telegraf/testutil"
)

var pki = testutil.NewPKI("../../../testutil/pki")

func TestAddressMissingProtocol(t *testing.T) {
	plugin := &Syslog{
		Address: "localhost:6514",
		Log:     testutil.Logger{},
	}
	require.ErrorContains(t, plugin.Init(), "missing protocol within address")
}

func TestAddressUnknownProtocol(t *testing.T) {
	plugin := &Syslog{
		Address: "unsupported://example.com:6514",
		Log:     testutil.Logger{},
	}
	require.ErrorContains(t, plugin.Init(), "unknown protocol")
}

func TestAddressDefault(t *testing.T) {
	plugin := &Syslog{Log: testutil.Logger{}}
	require.NoError(t, plugin.Init())

	require.Equal(t, "tcp://127.0.0.1:6514", plugin.url.String())
}

func TestAddressDefaultPort(t *testing.T) {
	plugin := &Syslog{
		Address: "tcp://localhost",
		Log:     testutil.Logger{},
	}
	require.NoError(t, plugin.Init())

	// Default port is 6514
	require.Equal(t, "tcp://localhost:6514", plugin.url.String())
}

func TestReadTimeoutWarning(t *testing.T) {
	logger := &testutil.CaptureLogger{}
	plugin := &Syslog{
		Address: "tcp://localhost:6514",
		Config: socket.Config{
			ReadTimeout: config.Duration(time.Second),
		},
		Log: logger,
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	plugin.Stop()

	require.Eventually(t, func() bool {
		return logger.NMessages() > 0
	}, 3*time.Second, 100*time.Millisecond)

	warnings := logger.Warnings()
	require.Contains(t, warnings, "W! [] "+readTimeoutMsg)
}

func TestUnixgram(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping test as unixgram is not supported on Windows")
	}

	// Create the socket
	sock := testutil.TempSocket(t)
	f, err := os.Create(sock)
	require.NoError(t, err)
	defer f.Close()

	// Setup plugin and start it
	plugin := &Syslog{
		Address: "unixgram://" + sock,
		Trailer: nontransparent.LF,
		Log:     testutil.Logger{},
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	// Send the message
	//nolint:lll // conditionally long lines allowed
	msg := `<29>1 2016-02-21T04:32:57+00:00 web1 someservice 2341 2 [origin][meta sequence="14125553" service="someservice"] "GET /v1/ok HTTP/1.1" 200 145 "-" "hacheck 0.9.0" 24306 127.0.0.1:40124 575`
	client, err := net.Dial("unixgram", sock)
	require.NoError(t, err)
	defer client.Close()
	_, err = client.Write([]byte(msg))
	require.NoError(t, err)

	// Do the comparison
	expected := []telegraf.Metric{
		metric.New(
			"syslog",
			map[string]string{
				"severity": "notice",
				"facility": "daemon",
				"hostname": "web1",
				"appname":  "someservice",
			},
			map[string]interface{}{
				"version":       uint16(1),
				"timestamp":     time.Unix(1456029177, 0).UnixNano(),
				"procid":        "2341",
				"msgid":         "2",
				"message":       `"GET /v1/ok HTTP/1.1" 200 145 "-" "hacheck 0.9.0" 24306 127.0.0.1:40124 575`,
				"origin":        true,
				"meta_sequence": "14125553",
				"meta_service":  "someservice",
				"severity_code": 5,
				"facility_code": 3,
			},
			time.Unix(0, 0),
		),
	}

	client.Close()

	// Check the metric nevertheless as we might get some metrics despite errors.
	require.Eventually(t, func() bool {
		return int(acc.NMetrics()) >= len(expected)
	}, 3*time.Second, 100*time.Millisecond)
	plugin.Stop()

	actual := acc.GetTelegrafMetrics()
	testutil.RequireMetricsEqual(t, expected, actual, testutil.IgnoreTime())
}

func TestCases(t *testing.T) {
	// Get all directories in testdata
	folders, err := os.ReadDir("testcases")
	require.NoError(t, err)

	// Register the plugin
	inputs.Add("syslog", func() telegraf.Input {
		return &Syslog{
			Trailer: nontransparent.LF,
			Log:     testutil.Logger{},
		}
	})

	for _, f := range folders {
		// Only handle folders
		if !f.IsDir() {
			continue
		}

		t.Run(f.Name(), func(t *testing.T) {
			testcasePath := filepath.Join("testcases", f.Name())
			configFilename := filepath.Join(testcasePath, "telegraf.conf")
			inputFilenamePattern := filepath.Join(testcasePath, "input*.txt")
			expectedFilename := filepath.Join(testcasePath, "expected.out")
			expectedErrorFilename := filepath.Join(testcasePath, "expected.err")

			// Prepare the influx parser for expectations
			parser := &parsers_influx_upstream.Parser{}
			require.NoError(t, parser.Init())

			// Read the input data
			inputFiles, err := filepath.Glob(inputFilenamePattern)
			require.NoError(t, err)
			require.NotEmpty(t, inputFiles)
			sort.Strings(inputFiles)
			messages := make([][]byte, 0, len(inputFiles))
			for _, fn := range inputFiles {
				data, err := os.ReadFile(fn)
				require.NoErrorf(t, err, "failed file: %s", fn)
				messages = append(messages, data)
			}

			// Read the expected output if any
			var expected []telegraf.Metric
			if _, err := os.Stat(expectedFilename); err == nil {
				var err error
				expected, err = testutil.ParseMetricsFromFile(expectedFilename, parser)
				require.NoError(t, err)
			}

			// Read the expected error if any
			var expectedError string
			if _, err := os.Stat(expectedErrorFilename); err == nil {
				buf, err := os.ReadFile(expectedErrorFilename)
				require.NoError(t, err)
				require.NotEmpty(t, buf)
				expectedError = string(buf)
			}

			// Configure the plugin and start it
			cfg := config.NewConfig()
			require.NoError(t, cfg.LoadConfig(configFilename))
			require.Len(t, cfg.Inputs, 1)
			plugin := cfg.Inputs[0].Input.(*Syslog)
			// Replace the TLS config with the known PKI infrastructure
			if plugin.ServerConfig.TLSCert != "" {
				plugin.ServerConfig = *pki.TLSServerConfig()
			}

			// Determine server properties. We need to parse the address before
			// calling Start() as it is modified in this function.
			if strings.HasPrefix(plugin.Address, "unix://") {
				// Use a random socket
				sock := filepath.ToSlash(testutil.TempSocket(t))
				if !strings.HasPrefix(sock, "/") {
					sock = "/" + sock
				}
				plugin.Address = "unix://" + sock
			}
			require.NoError(t, plugin.Init())

			var acc testutil.Accumulator
			require.NoError(t, plugin.Start(&acc))
			defer plugin.Stop()

			// Get the address
			addr := plugin.socket.Address().String()

			// Create a fake sender
			var client net.Conn
			srvTLS, err := plugin.TLSConfig()
			require.NoError(t, err)
			if srvTLS != nil {
				tlscfg, err := pki.TLSClientConfig().TLSConfig()
				require.NoError(t, err)
				tlscfg.ServerName = "localhost"

				client, err = tls.Dial(plugin.url.Scheme, addr, tlscfg)
				require.NoError(t, err)
			} else {
				client, err = net.Dial(plugin.url.Scheme, addr)
				require.NoError(t, err)
			}
			defer client.Close()

			// Send the data and afterwards stop client and plugin
			for i, msg := range messages {
				_, err := client.Write(msg)
				require.NoErrorf(t, err, "message %d failed with content %q", i, string(msg))
			}
			client.Close()

			// Check the metric nevertheless as we might get some metrics despite errors.
			require.Eventually(t, func() bool {
				return int(acc.NMetrics()) >= len(expected)
			}, 3*time.Second, 100*time.Millisecond)
			plugin.Stop()

			actual := acc.GetTelegrafMetrics()
			testutil.RequireMetricsEqual(t, expected, actual, testutil.IgnoreTime())

			// Check for errors
			if expectedError != "" {
				require.NotEmpty(t, acc.Errors)
				var found bool
				for _, err := range acc.Errors {
					found = found || strings.Contains(err.Error(), expectedError)
				}
				require.Truef(t, found, "expected error %q not found in errors %v", expectedError, acc.Errors)
			} else {
				require.Empty(t, acc.Errors)
			}
		})
	}
}

func TestSocketClosed(t *testing.T) {
	// Setup the plugin
	plugin := &Syslog{
		Address: "tcp://127.0.0.1:0",
		Config: socket.Config{
			ReadTimeout: config.Duration(10 * time.Millisecond),
		},
		Log: testutil.Logger{},
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	// Get the address
	addr := plugin.socket.Address().String()

	// Create a fake sender
	client, err := net.Dial("tcp", addr)
	require.NoError(t, err)
	defer client.Close()

	// Send a message to check if the socket is really active
	msg := []byte(`72 <13>1 2024-02-15T11:12:24.718151+01:00 Hugin sven - - [] Connection test`)
	_, err = client.Write(msg)
	require.NoError(t, err)

	// Stop the plugin and check if the socket is closed and unreachable
	plugin.Stop()

	require.Eventually(t, func() bool {
		_, err := client.Write(msg)
		return err != nil
	}, 3*time.Second, 100*time.Millisecond)
}

func TestIssue10121(t *testing.T) {
	// Setup the plugin
	plugin := &Syslog{
		Address: "tcp://127.0.0.1:0",
		Config: socket.Config{
			ReadTimeout: config.Duration(10 * time.Millisecond),
		},
		Log: testutil.Logger{},
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	// Get the address
	addr := plugin.socket.Address().String()

	// Create a fake sender
	client, err := net.Dial("tcp", addr)
	require.NoError(t, err)
	defer client.Close()

	// Messages should eventually timeout
	msg := []byte(`72 <13>1 2024-02-15T11:12:24.718151+01:00 Hugin sven - - [] Connection test`)
	require.Eventually(t, func() bool {
		_, err := client.Write(msg)
		return err != nil
	}, 3*time.Second, 250*time.Millisecond)
}
