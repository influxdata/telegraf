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

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/go-syslog/v3/nontransparent"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	framing "github.com/influxdata/telegraf/internal/syslog"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/testutil"
)

const (
	address = ":6514"
)

var defaultTime = time.Unix(0, 0)
var maxP = uint8(191)
var maxV = uint16(999)
var maxTS = "2017-12-31T23:59:59.999999+00:00"
var maxH = "abcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqr" +
	"stuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabc"
var maxA = "abcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdef"
var maxPID = "abcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzabcdefghilmnopqrstuvzab"
var maxMID = "abcdefghilmnopqrstuvzabcdefghilm"
var message7681 = strings.Repeat("l", 7681)

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

func TestAddressUnixgram(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping test as unixgram is not supported on Windows")
	}

	sock := filepath.Join(t.TempDir(), "syslog.TestAddress.sock")
	plugin := &Syslog{
		Address: "unixgram://" + sock,
	}

	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	require.Equal(t, sock, plugin.Address)
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
		// Compare options
		options := []cmp.Option{
			testutil.IgnoreTime(),
			testutil.SortMetrics(),
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
			testutil.RequireMetricsEqual(t, expected, actual, options...)

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
