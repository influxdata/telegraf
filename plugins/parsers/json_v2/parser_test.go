package json_v2_test

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/inputs/file"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
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
			expectedOutputs, err := readMetricFile(fmt.Sprintf("testdata/%s/expected.out", f.Name()))
			require.NoError(t, err)
			testutil.RequireMetricsEqual(t, expectedOutputs, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
		})
	}
}

func readMetricFile(path string) ([]telegraf.Metric, error) {
	var metrics []telegraf.Metric
	expectedFile, err := os.Open(path)
	if err != nil {
		return metrics, err
	}
	defer expectedFile.Close()

	parser := influx.NewParser(influx.NewMetricHandler())
	scanner := bufio.NewScanner(expectedFile)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			m, err := parser.ParseLine(line)
			if err != nil {
				return nil, fmt.Errorf("unable to parse metric in %q failed: %v", line, err)
			}
			metrics = append(metrics, m)
		}
	}
	err = expectedFile.Close()
	if err != nil {
		return metrics, err
	}

	return metrics, nil
}
