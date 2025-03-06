package whois

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/testutil"
)

// Make sure Whois implements telegraf.Input
var _ telegraf.Input = &Whois{}

func TestInit(t *testing.T) {
	// Setup the plugin
	plugin := &Whois{
		Domains: []string{"example.com", "google.com"},
		Server:  "whois.example.org",
		Timeout: config.Duration(5 * time.Second),
		Log:     testutil.Logger{},
	}

	// Test init
	require.NoError(t, plugin.Init())
}

func TestInitFail(t *testing.T) {
	tests := []struct {
		name     string
		domains  []string
		server   string
		timeout  config.Duration
		expected string
	}{
		{
			name:     "missing domains",
			timeout:  config.Duration(5 * time.Second),
			expected: "no domains configured",
		},
		{
			name:     "invalid timeout",
			domains:  []string{"example.com"},
			timeout:  config.Duration(0),
			expected: "timeout has to be greater than zero",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup the plugin
			plugin := &Whois{
				Domains: tt.domains,
				Server:  tt.server,
				Timeout: tt.timeout,
				Log:     testutil.Logger{},
			}
			// Test for the expected error message
			require.ErrorContains(t, plugin.Init(), tt.expected)
		})
	}
}

func TestWhoisCasesIntegration(t *testing.T) {
	// Get all directories in testcases
	folders, err := os.ReadDir("testcases")
	require.NoError(t, err, "failed to read testcases directory")

	// Prepare the influx parser for expectations
	parser := &influx.Parser{}
	require.NoError(t, parser.Init())

	for _, f := range folders {
		if !f.IsDir() {
			continue
		}

		testcasePath := filepath.Join("testcases", f.Name())
		configFilename := filepath.Join(testcasePath, "telegraf.conf")
		expectedFilename := filepath.Join(testcasePath, "expected.out")
		expectedErrorFilename := filepath.Join(testcasePath, "expected.err")

		// Compare options for metrics
		options := []cmp.Option{
			testutil.IgnoreTime(),
			testutil.SortMetrics(),
			// Ignore `expiry` due to possibility of fail on tests if CI is under high load
			testutil.IgnoreFields("expiry"),
		}

		t.Run(f.Name(), func(t *testing.T) {
			// Read input data
			inputData, inputErrors, err := readInputData(testcasePath)
			require.NoError(t, err)

			// Start a mock WHOIS server that serves test case data
			mockServerAddr, shutdown := startMockWhoisServer(inputData, inputErrors)
			defer shutdown()

			// Read expected output
			var expectedMetrics []telegraf.Metric
			if _, err := os.Stat(expectedFilename); err == nil {
				expectedMetrics, err = testutil.ParseMetricsFromFile(expectedFilename, parser)
				require.NoError(t, err)
			}

			// Read expected errors
			var expectedErrors []string
			if _, err := os.Stat(expectedErrorFilename); err == nil {
				expectedErrors, err = testutil.ParseLinesFromFile(expectedErrorFilename)
				require.NoError(t, err)
			}

			// Load Telegraf plugin config
			cfg := config.NewConfig()
			require.NoError(t, cfg.LoadConfig(configFilename))
			require.Len(t, cfg.Inputs, 1)

			// Get WHOIS plugin instance
			plugin := cfg.Inputs[0].Input.(*Whois)
			plugin.Server = mockServerAddr
			require.NoError(t, plugin.Init())

			var acc testutil.Accumulator
			require.NoError(t, plugin.Gather(&acc))

			var actualErrorMsgs []string
			for _, err := range acc.Errors {
				actualErrorMsgs = append(actualErrorMsgs, err.Error())
			}
			require.ElementsMatch(t, actualErrorMsgs, expectedErrors)

			// Compare expected metrics
			actualMetrics := acc.GetTelegrafMetrics()
			testutil.RequireMetricsEqual(t, expectedMetrics, actualMetrics, options...)
		})
	}
}

func readInputData(path string) (map[string][]byte, map[string]error, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, nil, err
	}

	data := make(map[string][]byte)
	errs := make(map[string]error)

	for _, e := range entries {
		if e.IsDir() || !strings.HasPrefix(e.Name(), "input") {
			continue
		}

		filename := filepath.Join(path, e.Name())
		ext := filepath.Ext(e.Name())
		server := strings.TrimPrefix(e.Name(), "input")
		server = strings.TrimPrefix(server, "_")
		server = strings.TrimSuffix(server, ext)

		switch ext {
		case ".txt":
			d, err := os.ReadFile(filename)
			if err != nil {
				return nil, nil, fmt.Errorf("reading %q failed: %w", filename, err)
			}
			data[server] = d
		case ".err":
			msgs, err := testutil.ParseLinesFromFile(filename)
			if err != nil {
				return nil, nil, fmt.Errorf("reading error %q failed: %w", filename, err)
			}
			if len(msgs) != 1 {
				return nil, nil, fmt.Errorf("unexpected number of errors: %d", len(msgs))
			}
			errs[server] = errors.New(msgs[0])
		default:
			return nil, nil, fmt.Errorf("unexpected input %q", filename)
		}
	}

	return data, errs, nil
}

func startMockWhoisServer(responses map[string][]byte, errResponses map[string]error) (string, func()) {
	listener, err := net.Listen("tcp", "127.0.0.1:0") // Random available port
	if err != nil {
		panic(fmt.Sprintf("failed to start mock WHOIS server: %v", err))
	}

	serverAddr := listener.Addr().String()
	shutdown := func() { _ = listener.Close() } // Cleanup after test

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return // Stop accepting new connections on shutdown
			}
			go func(c net.Conn) {
				defer c.Close()
				buffer := make([]byte, 1024)
				n, err := c.Read(buffer)
				if err != nil {
					return
				}

				domain := strings.TrimSpace(string(buffer[:n])) // Read WHOIS query
				if err, exists := errResponses[domain]; exists {
					if _, err := c.Write([]byte(err.Error() + "\n")); err != nil {
						fmt.Printf("failed to write error response: %v\n", err)
						return
					}
				}

				if response, exists := responses[domain]; exists {
					if _, err := c.Write(response); err != nil {
						fmt.Printf("failed to write WHOIS response: %v\n", err)
					}
				} else {
					if _, err := c.Write([]byte("ERROR: No data available\n")); err != nil {
						fmt.Printf("failed to write default error response: %v\n", err)
					}
				}
			}(conn)
		}
	}()
	return serverAddr, shutdown
}
