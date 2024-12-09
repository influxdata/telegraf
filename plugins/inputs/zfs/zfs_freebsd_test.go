//go:build freebsd

package zfs

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/testutil"
)

// Generate testcase-data via
//
//	zpool.txt:    $ zpool list -Hp -o name,health,size,alloc,free,fragmentation,capacity,dedupratio
//	zdataset.txt: $ zfs list -Hp -o name,avail,used,usedsnap,usedds
//	sysctl.json:  $ sysctl -q kstat.zfs.misc.<kstat metrics>
func TestCases(t *testing.T) {
	// Get all testcase directories
	testpath := filepath.Join("testcases", "freebsd")
	folders, err := os.ReadDir(testpath)
	require.NoError(t, err)

	// Register the plugin
	inputs.Add("zfs", func() telegraf.Input { return &Zfs{} })

	for _, f := range folders {
		// Only handle folders
		if !f.IsDir() {
			continue
		}

		t.Run(f.Name(), func(t *testing.T) {
			testcasePath := filepath.Join(testpath, f.Name())
			configFilename := filepath.Join(testcasePath, "telegraf.conf")
			inputSysctlFilename := filepath.Join(testcasePath, "sysctl.json")
			inputZPoolFilename := filepath.Join(testcasePath, "zpool.txt")
			inputZDatasetFilename := filepath.Join(testcasePath, "zdataset.txt")
			inputUnameFilename := filepath.Join(testcasePath, "uname.txt")
			expectedFilename := filepath.Join(testcasePath, "expected.out")

			// Load the input data
			buf, err := os.ReadFile(inputSysctlFilename)
			require.NoError(t, err)
			var sysctl map[string][]string
			require.NoError(t, json.Unmarshal(buf, &sysctl))

			zpool, err := testutil.ParseLinesFromFile(inputZPoolFilename)
			require.NoError(t, err)

			zdataset, err := testutil.ParseLinesFromFile(inputZDatasetFilename)
			require.NoError(t, err)

			// Try to read release from file and default to FreeBSD 13 if
			// an error occurs.
			uname := "13.2-STABLE"
			if buf, err := os.ReadFile(inputUnameFilename); err == nil {
				uname = string(buf)
			}

			// Prepare the influx parser for expectations
			parser := &influx.Parser{}
			require.NoError(t, parser.Init())

			// Read the expected output
			expected, err := testutil.ParseMetricsFromFile(expectedFilename, parser)
			require.NoError(t, err)

			// Configure the plugin
			cfg := config.NewConfig()
			require.NoError(t, cfg.LoadConfig(configFilename))
			require.Len(t, cfg.Inputs, 1)

			// Setup the plugin
			plugin := cfg.Inputs[0].Input.(*Zfs)
			plugin.sysctl = func(metric string) ([]string, error) {
				if r, found := sysctl[metric]; found {
					return r, nil
				}
				return nil, fmt.Errorf("invalid argument")
			}
			plugin.zpool = func() ([]string, error) { return zpool, nil }
			plugin.zdataset = func(_ []string) ([]string, error) { return zdataset, nil }
			plugin.uname = func() (string, error) { return uname, nil }
			plugin.Log = testutil.Logger{}
			require.NoError(t, plugin.Init())

			// Gather and test
			var acc testutil.Accumulator
			require.NoError(t, plugin.Gather(&acc))

			actual := acc.GetTelegrafMetrics()
			testutil.RequireMetricsEqual(t, expected, actual, testutil.IgnoreTime())
		})
	}
}
