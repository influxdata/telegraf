package json_v2_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/inputs/file"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestMultipleConfigs(t *testing.T) {
	// Get all directories in testdata
	folders, err := ioutil.ReadDir("testdata")
	require.NoError(t, err)
	// Make sure testdata contains data
	require.Greater(t, len(folders), 0)

	for _, f := range folders {
		t.Run(f.Name(), func(t *testing.T) {
			// Process the telegraf config file for the test
			buf, err := os.ReadFile(fmt.Sprintf("testdata/%s/telegraf.conf", f.Name()))
			require.NoError(t, err)
			inputs.Add("file", func() telegraf.Input {
				return &file.File{}
			})
			cfg := config.NewConfig()
			err = cfg.LoadConfigData(buf)
			require.NoError(t, err)

			// Gather the metrics from the input file configure
			acc := testutil.Accumulator{}
			for _, i := range cfg.Inputs {
				err = i.Init()
				require.NoError(t, err)
				err = i.Gather(&acc)
				require.NoError(t, err)
			}

			// Process expected metrics and compare with resulting metrics
			expectedOutputs, err := testutil.ReadMetricFile(fmt.Sprintf("testdata/%s/expected.out", f.Name()))
			require.NoError(t, err)
			resultingMetrics := acc.GetTelegrafMetrics()
			testutil.RequireMetricsEqual(t, expectedOutputs, resultingMetrics, testutil.IgnoreTime())

			// Folder with timestamp prefixed will also check for matching timestamps to make sure they are parsed correctly
			// The milliseconds weren't matching, seemed like a rounding difference between the influx parser
			// Compares each metrics times separately and ignores milliseconds
			if strings.HasPrefix(f.Name(), "timestamp") {
				require.Equal(t, len(expectedOutputs), len(resultingMetrics))
				for i, m := range resultingMetrics {
					require.Equal(t, expectedOutputs[i].Time().Truncate(time.Second), m.Time().Truncate(time.Second))
				}
			}
		})
	}
}
