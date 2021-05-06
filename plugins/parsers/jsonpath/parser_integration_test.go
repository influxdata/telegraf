package jsonpath

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
)

func TestJSONPathDockerIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode, this test requires Docker")
	}

	tests := []string{
		"types",
	}

	ctx := context.Background()

	dir, _ := os.Getwd()
	localBindDir := dir + "/testdata"

	// Build a Docker container that compiles and installs Telegraf from this repository.
	// The Dockerfile runs until its explicitly terminated.
	// The estdata dir is mounted to the container so that the resulting test data can be verified.
	req := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    dir + "/../../../",
			Dockerfile: "scripts/integration_tests.docker",
		},
		BindMounts: map[string]string{
			localBindDir: "/tmp/",
		},
	}
	telegrafDocker, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	defer func() {
		err := telegrafDocker.Terminate(ctx)
		require.NoError(t, err)
	}()

	// The integration tests require a directory within `testdata` to contain the following files:
	// 1. `telegraf.conf` file defining the telegraf configuration for the test, requires output.file plugin to be defined
	// 2. `input.json` file defining the JSON that will be parsed
	// 3. `expected.out` file defining the expected resulting metrics, the timestamp isn't necessary and will be overwritten
	for _, testName := range tests {
		_, err = telegrafDocker.Exec(ctx, []string{
			"telegraf", "--config", fmt.Sprintf("./plugins/parsers/jsonpath/testdata/%s/telegraf.conf", testName), "--once",
		})
		require.NoError(t, err)

		// Read the file in the testdata with the expected metrics
		expectedMetricsFile, err := readMetricFile(fmt.Sprintf("%s/%s/expected.out", localBindDir, testName))
		require.NoError(t, err)

		// Read the file outputed by the docker container with the expected metrics
		// All test telegraf configs need the plugin `output.file` defined to write out a file to /tmp/
		resultingMetricPath := fmt.Sprintf("%s/%s.out", localBindDir, testName)
		resultingMetricsFile, err := readMetricFile(resultingMetricPath)
		require.NoError(t, err)
		require.True(t, len(expectedMetricsFile) == len(resultingMetricsFile))
		var resultingMetrics, expectedMetrics []telegraf.Metric
		for i := range resultingMetricsFile {
			metricHandler := influx.NewMetricHandler()
			influxParser := influx.NewParser(metricHandler)
			// Extract all the resulting metrics from running the Telegraf with the test data
			resultM, err := influxParser.ParseLine(resultingMetricsFile[i])
			require.NoError(t, err)
			resultingMetrics = append(resultingMetrics, resultM)
			// Extract all the expected metrics from the testdata directory
			expectedM, err := influxParser.ParseLine(expectedMetricsFile[i])
			require.NoError(t, err)
			expectedM.SetTime(resultM.Time())
			expectedMetrics = append(expectedMetrics, expectedM)
		}
		// The resulting metrics can be in any order, loop over the slice to verify a match exists
		for _, expectedMetric := range expectedMetrics {
			var found bool
			for _, resultMetric := range resultingMetrics {
				if found = testutil.MetricEqual(expectedMetric, resultMetric); found {
					break
				}
			}
			require.True(t, found)
		}
		err = os.Remove(resultingMetricPath)
		require.NoError(t, err)
	}
}

func readMetricFile(path string) ([]string, error) {
	expectedFile, err := os.Open(path)
	if err != nil {
		return []string{}, err
	}
	defer expectedFile.Close()

	var metrics []string
	scanner := bufio.NewScanner(expectedFile)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			metrics = append(metrics, line)
		}
	}
	err = expectedFile.Close()
	if err != nil {
		return []string{}, err
	}

	return metrics, nil
}
