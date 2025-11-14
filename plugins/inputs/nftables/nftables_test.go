//go:build linux

package nftables

import (
	"errors"
	"flag"
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

func TestCases(t *testing.T) {
	// Get all directories in testdata
	folders, err := os.ReadDir("testcases")
	require.NoError(t, err)

	// Register the plugin
	inputs.Add("nftables", func() telegraf.Input {
		return &Nftables{}
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

			// Determine the executable name of the test-binary used to mock the
			// nft command. See TestMain
			exe, err := os.Executable()
			require.NoError(t, err)

			// Configure the plugin
			cfg := config.NewConfig()
			require.NoError(t, cfg.LoadConfig(configFilename))
			require.Len(t, cfg.Inputs, 1)

			// Mock the nft executable
			plugin := cfg.Inputs[0].Input.(*Nftables)
			plugin.Binary = exe
			require.NoError(t, plugin.Init())
			plugin.args = append([]string{"--mock", "--testcase", testcasePath}, plugin.args...)

			// Gather the metrics and compare the output
			var acc testutil.Accumulator
			require.NoError(t, plugin.Gather(&acc))
			if len(acc.Errors) > 0 {
				var actualErrorMsgs []string
				for _, err := range acc.Errors {
					actualErrorMsgs = append(actualErrorMsgs, err.Error())
				}
				require.ElementsMatch(t, actualErrorMsgs, expectedErrors)
			}
			testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), options...)
		})
	}
}

func TestMain(m *testing.M) {
	// Mimic the nft command line arguments
	var mock, jsonMode bool
	var testcase string
	flag.BoolVar(&mock, "mock", false, "run test as mock")
	flag.StringVar(&testcase, "testcase", "", "path to the test directory")
	flag.BoolVar(&jsonMode, "json", false, "output as JSON")
	flag.Parse()

	if !mock {
		// Run the normal test mode
		os.Exit(m.Run())
	}

	// Run as a mock program
	if !jsonMode {
		fmt.Fprintln(os.Stderr, "JSON mode not set")
		os.Exit(1)
	}

	args := flag.Args()
	if nargs := len(args); nargs != 3 {
		fmt.Fprintf(os.Stderr, "invalid number of arguments, expected 3 got %d\n", nargs)
		os.Exit(1)
	}
	if args[0] != "list" {
		fmt.Fprintf(os.Stderr, "expected \"list\" command got %q\n", args[0])
		os.Exit(1)
	}
	if args[1] != "table" {
		fmt.Fprintf(os.Stderr, "expected \"list\" command got %q\n", args[0])
		os.Exit(1)
	}

	filename := filepath.Join(testcase, "table_"+args[2]+".json")
	buf, err := os.ReadFile(filename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			fmt.Fprintln(os.Stderr, "Error: No such file or directory")
			fmt.Fprintln(os.Stderr, "list table", args[1])
			fmt.Fprintln(os.Stderr, "          ", strings.Repeat("^", len(args[1])))
		} else {
			fmt.Fprintf(os.Stderr, "reading file %q failed: %v", filename, err)
		}
		os.Exit(1)
	}
	// This mimics the command output, do not remove!
	fmt.Print(string(buf))
	os.Exit(0)
}
