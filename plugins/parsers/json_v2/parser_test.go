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

func TestData(t *testing.T) {
	var tests = []struct {
		name string
		test string
	}{
		{
			name: "Test using just fields and tags",
			test: "fields_and_tags",
		},
		{
			name: "Test gathering from array of nested objects",
			test: "nested_array_of_objects",
		},
		{
			name: "Test setting timestamp",
			test: "timestamp",
		},
		{
			name: "Test setting measurement name from int",
			test: "measurement_name_int",
		},
		{
			name: "Test multiple types",
			test: "types",
		},
		{
			name: "Test settings tags in nested object",
			test: "nested_tags",
		},
		{
			name: "Test settings tags in nested and non-nested objects",
			test: "nested_and_nonnested_tags",
		},
		{
			name: "Test a more complex nested tag retrieval",
			test: "nested_tags_complex",
		},
		{
			name: "Test multiple arrays in object",
			test: "multiple_arrays_in_object",
		},
		{
			name: "Test fields and tags complex",
			test: "fields_and_tags_complex",
		},
		{
			name: "Test object",
			test: "object",
		},
		{
			name: "Test multiple timestamps",
			test: "multiple_timestamps",
		},
		{
			name: "Test field with null",
			test: "null",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Process the telegraf config file for the test
			buf, err := ioutil.ReadFile(fmt.Sprintf("testdata/%s/telegraf.conf", tc.test))
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
			require.NoError(t, err)

			// Process expected metrics and compare with resulting metrics
			expectedOutputs, err := readMetricFile(fmt.Sprintf("testdata/%s/expected.out", tc.test))
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
