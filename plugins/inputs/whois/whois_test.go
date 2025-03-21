package whois

import (
	// "errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
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

func TestCases(t *testing.T) {
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
			// Create and start a mock WHOIS server
			mockServer, err := createMockServer(testcasePath)
			require.NoError(t, err, "failed to create mock WHOIS server")

			mockServerAddr, err := mockServer.start()
			require.NoError(t, err, "failed to start mock WHOIS server")
			defer mockServer.stop() // Ensure cleanup

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

type server struct {
	responses map[string][]byte
	listener  net.Listener

	errors []error
	sync.Mutex
}

func createMockServer(path string) (*server, error) {
	// Read the input data
	matches, err := filepath.Glob(filepath.Join(path, "input_*.txt"))
	if err != nil {
		return nil, fmt.Errorf("matching input files failed: %w", err)
	}

	responses := make(map[string][]byte, len(matches))
	for _, fn := range matches {
		buf, err := os.ReadFile(fn)
		if err != nil {
			return nil, fmt.Errorf("reading %q failed: %w", fn, err)
		}
		domain := strings.TrimPrefix(filepath.Base(fn), "input_")
		domain = strings.TrimSuffix(domain, ".txt")
		responses[domain] = buf
	}
	return &server{responses: responses}, nil
}

func (s *server) start() (string, error) {
	// Create the listener
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", fmt.Errorf("starting server failed: %w", err)
	}
	s.listener = listener

	addr := listener.Addr().String()
	go func() {
		for {
			conn, err := s.listener.Accept()
			if err != nil {
				return // Stop accepting new connections on shutdown
			}

			go func(c net.Conn) {
				defer c.Close()
				// Read the requested domain
				buf := make([]byte, 1024)
				n, err := c.Read(buf)
				if err != nil {
					return
				}
				domain := strings.TrimSpace(string(buf[:n]))

				// Write the response from the input data or an error if the domain cannot be found
				response, found := s.responses[domain]
				if !found {
					response = []byte("ERROR: No data available\n")
				}

				if _, err := c.Write(response); err != nil {
					s.Lock()
					s.errors = append(s.errors, fmt.Errorf("writing response %q failed: %w", domain, err))
					s.Unlock()
				}
			}(conn)
		}
	}()
	return addr, nil
}

func (s *server) stop() {
	if s.listener != nil {
		s.listener.Close()
		s.listener = nil
	}
}
