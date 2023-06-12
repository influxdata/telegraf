package netflow

import (
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/testutil"
)

func TestInit(t *testing.T) {
	tests := []struct {
		name     string
		address  string
		protocol string
		errmsg   string
	}{
		{
			name:     "Netflow v5",
			address:  "udp://:2055",
			protocol: "netflow v5",
		},
		{
			name:     "Netflow v5 (uppercase)",
			address:  "udp://:2055",
			protocol: "Netflow v5",
		},
		{
			name:     "Netflow v9",
			address:  "udp://:2055",
			protocol: "netflow v9",
		},
		{
			name:     "Netflow v9 (uppercase)",
			address:  "udp://:2055",
			protocol: "Netflow v9",
		},
		{
			name:     "IPFIX",
			address:  "udp://:2055",
			protocol: "ipfix",
		},
		{
			name:     "IPFIX (uppercase)",
			address:  "udp://:2055",
			protocol: "IPFIX",
		},
		{
			name:     "invalid protocol",
			address:  "udp://:2055",
			protocol: "foo",
			errmsg:   "invalid protocol",
		},
		{
			name:     "UDP",
			address:  "udp://:2055",
			protocol: "netflow v5",
		},
		{
			name:     "UDP4",
			address:  "udp4://:2055",
			protocol: "netflow v5",
		},
		{
			name:     "UDP6",
			address:  "udp6://:2055",
			protocol: "netflow v5",
		},
		{
			name:     "empty service address",
			address:  "",
			protocol: "netflow v5",
			errmsg:   "service_address required",
		},
		{
			name:     "invalid address scheme",
			address:  "tcp://:2055",
			protocol: "netflow v5",
			errmsg:   "invalid scheme",
		},
		{
			name:     "invalid service address",
			address:  "udp://198.168.1.290:la",
			protocol: "netflow v5",
			errmsg:   "invalid service address",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := &NetFlow{
				ServiceAddress: tt.address,
				Protocol:       tt.protocol,
				Log:            testutil.Logger{},
			}
			err := plugin.Init()
			if tt.errmsg != "" {
				require.ErrorContains(t, err, tt.errmsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMissingTemplate(t *testing.T) {
	raw := "000a00bc646b84c000000000000000e7010500ac000000000001dbe100000000"
	raw += "0000038a060018bdeac0a802c8000000000001bb6810f9f90000000000000000"
	raw += "000157b8c40155f28a00005056b3e365005056b3a7f804646b8471646b84e600"
	raw += "00018843fd5cf60000018843ff232e000000000000000e00000000000007bc00"
	raw += "000005000009560000000300dc00000000000000000000000000000e3130342e"
	raw += "31362e3234392e3234390e3130342e31362e3234392e323439000000"
	msg, err := hex.DecodeString(raw)
	require.NoError(t, err)

	var acc testutil.Accumulator
	var logger testutil.CaptureLogger
	plugin := &NetFlow{
		ServiceAddress: "udp://127.0.0.1:0",
		Log:            &logger,
	}
	require.NoError(t, plugin.Init())
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	// Create a client without TLS
	addr := plugin.conn.LocalAddr()
	client, err := createClient(plugin.ServiceAddress, addr)
	require.NoError(t, err)

	// Write the message
	_, err = client.Write(msg)
	require.NoErrorf(t, err, "writing message failed: %v", err)
	require.NoError(t, client.Close())

	// We expect a warning here
	require.Eventually(t, func() bool {
		return len(logger.Warnings()) > 0
	}, 3*time.Second, 100*time.Millisecond, "did not receive expected warnings")

	var found bool
	for _, w := range logger.Warnings() {
		found = found || strings.Contains(w, "No info template 261 found for and domain id 231; skipping packet")
	}
	require.True(t, found, "warning not found")
}

func TestWrongMapping(t *testing.T) {
	var logger testutil.CaptureLogger
	plugin := &NetFlow{
		ServiceAddress: "udp://127.0.0.1:0",
		Protocol:       "ipfix",
		PENFiles:       []string{"testcases/netflow_mapping.csv"},
		Log:            &logger,
	}
	require.ErrorContains(t, plugin.Init(), "does not match pattern")
}

func TestCases(t *testing.T) {
	// Get all directories in testdata
	folders, err := os.ReadDir("testcases")
	require.NoError(t, err)

	// Register the plugin
	inputs.Add("netflow", func() telegraf.Input {
		return &NetFlow{}
	})

	// Prepare the influx parser for expectations
	parser := &influx.Parser{}
	require.NoError(t, parser.Init())

	for _, f := range folders {
		// Only handle folders
		if !f.IsDir() {
			continue
		}
		testcasePath := filepath.Join("testcases", f.Name())
		configFilename := filepath.Join(testcasePath, "telegraf.conf")
		inputFiles := filepath.Join(testcasePath, "*.bin")
		expectedFilename := filepath.Join(testcasePath, "expected.out")
		expectedErrorFilename := filepath.Join(testcasePath, "expected.err")

		// Compare options
		options := []cmp.Option{
			testutil.IgnoreTime(),
			testutil.SortMetrics(),
		}

		t.Run(f.Name(), func(t *testing.T) {
			// Read the input data
			var messages [][]byte
			matches, err := filepath.Glob(inputFiles)
			require.NoError(t, err)
			require.NotEmpty(t, matches)
			sort.Strings(matches)
			for _, fn := range matches {
				m, err := os.ReadFile(fn)
				require.NoError(t, err)
				messages = append(messages, m)
			}

			// Read the expected output if any
			var expected []telegraf.Metric
			if _, err := os.Stat(expectedFilename); err == nil {
				var err error
				expected, err = testutil.ParseMetricsFromFile(expectedFilename, parser)
				require.NoError(t, err)
			}

			// Read the expected output if any
			var expectedErrors []string
			if _, err := os.Stat(expectedErrorFilename); err == nil {
				var err error
				expectedErrors, err = testutil.ParseLinesFromFile(expectedErrorFilename)
				require.NoError(t, err)
				require.NotEmpty(t, expectedErrors)
			}

			// Configure the plugin
			cfg := config.NewConfig()
			require.NoError(t, cfg.LoadConfig(configFilename))
			require.Len(t, cfg.Inputs, 1)

			// Setup and start the plugin
			var acc testutil.Accumulator
			plugin := cfg.Inputs[0].Input.(*NetFlow)
			require.NoError(t, plugin.Init())
			require.NoError(t, plugin.Start(&acc))
			defer plugin.Stop()

			// Create a client without TLS
			addr := plugin.conn.LocalAddr()
			client, err := createClient(plugin.ServiceAddress, addr)
			require.NoError(t, err)

			// Write the given sequence
			for i, msg := range messages {
				_, err := client.Write(msg)
				require.NoErrorf(t, err, "writing message from %q failed: %v", matches[i], err)
			}
			require.NoError(t, client.Close())

			getNErrors := func() int {
				acc.Lock()
				defer acc.Unlock()
				return len(acc.Errors)
			}
			require.Eventuallyf(t, func() bool {
				return getNErrors() >= len(expectedErrors)
			}, 3*time.Second, 100*time.Millisecond, "did not receive errors (%d/%d)", getNErrors(), len(expectedErrors))

			require.Lenf(t, acc.Errors, len(expectedErrors), "got errors: %v", acc.Errors)
			sort.SliceStable(acc.Errors, func(i, j int) bool {
				return acc.Errors[i].Error() < acc.Errors[j].Error()
			})
			for i, err := range acc.Errors {
				require.ErrorContains(t, err, expectedErrors[i])
			}

			require.Eventuallyf(t, func() bool {
				acc.Lock()
				defer acc.Unlock()
				return acc.NMetrics() >= uint64(len(expected))
			}, 3*time.Second, 100*time.Millisecond, "did not receive metrics (%d/%d)", acc.NMetrics(), len(expected))

			// Check the metric nevertheless as we might get some metrics despite errors.
			actual := acc.GetTelegrafMetrics()
			testutil.RequireMetricsEqual(t, expected, actual, options...)
		})
	}
}

func createClient(endpoint string, addr net.Addr) (net.Conn, error) {
	// Determine the protocol in a crude fashion
	parts := strings.SplitN(endpoint, "://", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid endpoint %q", endpoint)
	}
	protocol := parts[0]
	return net.Dial(protocol, addr.String())
}
