package whois

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/likexian/whois"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/testutil"
)

// Make sure Whois implements telegraf.Input
var _ telegraf.Input = &Whois{}

func TestWhoisConfigInitialization(t *testing.T) {
	tests := []struct {
		name               string
		domains            []string
		server             string
		IncludeNameServers bool
		timeout            config.Duration
		expectErr          bool
	}{
		{
			name:      "Valid Configuration",
			domains:   []string{"example.com", "google.com"},
			server:    "whois.example.org",
			timeout:   config.Duration(10 * time.Second),
			expectErr: false,
		},
		{
			name:      "No Domains Configured",
			domains:   nil,
			timeout:   config.Duration(5 * time.Second),
			expectErr: true,
		},
		{
			name:      "Invalid Timeout (Zero Value)",
			domains:   []string{"example.com"},
			timeout:   config.Duration(0),
			expectErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			plugin := &Whois{
				Domains: test.domains,
				Timeout: test.timeout,
				Server:  test.server,
				Log:     testutil.Logger{},
			}

			err := plugin.Init()

			if test.expectErr {
				require.Error(t, err, "Expected error but got none")
				return
			}

			require.NoError(t, err, "Unexpected error during Init()")
		})
	}
}

func TestWhoisCases(t *testing.T) {
	// Get all directories in testcases
	folders, err := os.ReadDir("testcases")
	require.NoError(t, err, "Failed to read testcases directory")

	// Register the WHOIS plugin inside the test
	inputs.Add("whois", func() telegraf.Input {
		return &Whois{}
	})

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
			testutil.IgnoreFields("expiry"),
		}

		t.Run(f.Name(), func(t *testing.T) {
			// Read input data
			inputData, inputErrors, err := readInputData(testcasePath)
			require.NoError(t, err)

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
			plugin.whoisLookup = func(_ *whois.Client, domain, _ string) (string, error) {
				return string(inputData[domain]), inputErrors[domain]
			}
			require.NoError(t, plugin.Init())

			var acc testutil.Accumulator
			require.NoError(t, plugin.Gather(&acc))
			if len(acc.Errors) > 0 {
				var actualErrorMsgs []string
				for _, err := range acc.Errors {
					actualErrorMsgs = append(actualErrorMsgs, err.Error())
				}
				require.ElementsMatch(t, actualErrorMsgs, expectedErrors)
			}

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
