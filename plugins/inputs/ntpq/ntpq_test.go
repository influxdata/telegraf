package ntpq

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/testutil"
)

func TestInitInvalid(t *testing.T) {
	tests := []struct {
		name     string
		plugin   *NTPQ
		expected string
	}{
		{
			name:     "invalid reach_format",
			plugin:   &NTPQ{ReachFormat: "garbage"},
			expected: `unknown 'reach_format' "garbage"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.EqualError(t, tt.plugin.Init(), tt.expected)
		})
	}
}

func TestCases(t *testing.T) {
	// Get all directories in testdata
	folders, err := os.ReadDir("testcases")
	require.NoError(t, err)

	// Register the plugin
	inputs.Add("ntpq", func() telegraf.Input {
		return &NTPQ{}
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
		expectedFilename := filepath.Join(testcasePath, "expected.out")
		expectedErrorFilename := filepath.Join(testcasePath, "expected.err")

		// Compare options
		options := []cmp.Option{
			testutil.IgnoreTime(),
			testutil.SortMetrics(),
		}

		t.Run(f.Name(), func(t *testing.T) {
			// Read the input data
			inputData, inputErrors, err := readInputData(testcasePath)
			require.NoError(t, err)

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

			// Fake the reading
			plugin := cfg.Inputs[0].Input.(*NTPQ)
			plugin.runQ = func(server string) ([]byte, error) {
				return inputData[server], inputErrors[server]
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

			// Check the metric nevertheless as we might get some metrics despite errors.
			actual := acc.GetTelegrafMetrics()
			testutil.RequireMetricsEqual(t, expected, actual, options...)
		})
	}
}

func readInputData(path string) (map[string][]byte, map[string]error, error) {
	// Get all elements in the testcase directory
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
		server = strings.TrimPrefix(server, "_") // This needs to be separate for non-server cases
		server = strings.TrimSuffix(server, ext)

		switch ext {
		case ".txt":
			// Read the input data
			d, err := os.ReadFile(filename)
			if err != nil {
				return nil, nil, fmt.Errorf("reading %q failed: %w", filename, err)
			}
			data[server] = d
		case ".err":
			// Read the input error message
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
