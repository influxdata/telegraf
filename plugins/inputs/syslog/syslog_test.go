package syslog

import (
	"crypto/tls"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/go-syslog/v3/nontransparent"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	framing "github.com/influxdata/telegraf/internal/syslog"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/inputs"
	influx "github.com/influxdata/telegraf/plugins/parsers/influx/influx_upstream"
	"github.com/influxdata/telegraf/testutil"
)

var pki = testutil.NewPKI("../../../testutil/pki")

func TestInitFail(t *testing.T) {
	tests := []struct {
		name     string
		address  string
		expected string
	}{
		{
			name:     "no address",
			expected: "missing protocol within address",
		},
		{
			name:     "missing protocol",
			address:  "localhost:6514",
			expected: "missing protocol within address",
		},
		{
			name:     "unknown protocol",
			address:  "unsupported://example.com:6514",
			expected: "unknown protocol",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := &Syslog{
				Address: tt.address,
			}
			var acc testutil.Accumulator
			require.ErrorContains(t, plugin.Start(&acc), tt.expected)
		})
	}
}

func TestAddressDefaultPort(t *testing.T) {
	plugin := &Syslog{
		Address: "tcp://localhost",
	}

	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	// Default port is 6514
	require.Equal(t, "localhost:6514", plugin.Address)
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
	timeout := config.Duration(defaultReadTimeout)
	plugin := &Syslog{
		Address:        "unixgram://" + sock,
		Framing:        framing.OctetCounting,
		ReadTimeout:    &timeout,
		Separator:      "_",
		SyslogStandard: "RFC5424",
		Trailer:        nontransparent.LF,
		now:            getNanoNow,
	}

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
		defaultTimeout := config.Duration(defaultReadTimeout)
		return &Syslog{
			Address:        ":6514",
			now:            getNanoNow,
			ReadTimeout:    &defaultTimeout,
			Framing:        framing.OctetCounting,
			SyslogStandard: syslogRFC5424,
			Trailer:        nontransparent.LF,
			Separator:      "_",
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
			inputFilename := filepath.Join(testcasePath, "input.txt")
			expectedFilename := filepath.Join(testcasePath, "expected.out")
			expectedErrorFilename := filepath.Join(testcasePath, "expected.err")

			// Prepare the influx parser for expectations
			parser := &influx.Parser{}
			require.NoError(t, parser.Init())

			// Read the input data
			inputData, err := os.ReadFile(inputFilename)
			require.NoError(t, err)

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
			u, err := url.Parse(plugin.Address)
			require.NoError(t, err)
			if u.Scheme == "unix" {
				// Use a random socket
				sock := testutil.TempSocket(t)
				plugin.Address = "unix://" + sock
			}

			var acc testutil.Accumulator
			require.NoError(t, plugin.Start(&acc))
			defer plugin.Stop()

			// Get the address
			var addr string
			if plugin.isStream {
				addr = plugin.tcpListener.Addr().String()
			} else {
				addr = plugin.udpListener.LocalAddr().String()
			}

			// Create a fake sender
			var client net.Conn
			if srvTLS, _ := plugin.TLSConfig(); srvTLS != nil {
				tlscfg, err := pki.TLSClientConfig().TLSConfig()
				require.NoError(t, err)
				tlscfg.ServerName = "localhost"

				client, err = tls.Dial(u.Scheme, addr, tlscfg)
				require.NoError(t, err)
			} else {
				client, err = net.Dial(u.Scheme, addr)
				require.NoError(t, err)
			}
			defer client.Close()

			// Send the data and afterwards stop client and plugin
			_, err = client.Write(inputData)
			require.NoError(t, err)
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
