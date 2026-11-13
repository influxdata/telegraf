//go:build linux

package powercap

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/testutil"
)

const timeBetweenSteps = time.Second

func TestCases(t *testing.T) {
	folders, err := os.ReadDir("testcases")
	require.NoError(t, err)

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

		t.Run(f.Name(), func(t *testing.T) {
			var expected []telegraf.Metric
			if _, err := os.Stat(expectedFilename); err == nil {
				expected, err = testutil.ParseMetricsFromFile(expectedFilename, parser)
				require.NoError(t, err)
			}

			var expectedErrors []string
			if _, err := os.Stat(expectedErrorFilename); err == nil {
				expectedErrors, err = testutil.ParseLinesFromFile(expectedErrorFilename)
				require.NoError(t, err)
				require.NotEmpty(t, expectedErrors)
			}

			cfg := config.NewConfig()
			require.NoError(t, cfg.LoadConfig(configFilename))
			require.Len(t, cfg.Inputs, 1)
			plugin := cfg.Inputs[0].Input.(*Powercap)
			require.NoError(t, plugin.Init())

			var acc testutil.Accumulator
			stepTime := time.Unix(0, 0)
			plugin.timeFunc = func() time.Time { return stepTime }
			for _, step := range stepsOf(t, testcasePath) {
				t.Setenv("HOST_SYS", filepath.Join(step, "sys"))
				require.NoError(t, plugin.Gather(&acc))
				stepTime = stepTime.Add(timeBetweenSteps)
			}

			actualErrors := make([]string, 0, len(acc.Errors))
			for _, err := range acc.Errors {
				actualErrors = append(actualErrors, err.Error())
			}
			require.ElementsMatch(t, expectedErrors, actualErrors)
			testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.SortMetrics())
		})
	}
}

func stepsOf(t *testing.T, testcasePath string) []string {
	steps, err := filepath.Glob(filepath.Join(testcasePath, "step*"))
	require.NoError(t, err)
	if len(steps) == 0 {
		return []string{testcasePath}
	}
	return steps
}
