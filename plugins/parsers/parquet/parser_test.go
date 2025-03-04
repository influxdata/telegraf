package parquet

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/testutil"
	test "github.com/influxdata/telegraf/testutil/plugin_input"
)

func TestCases(t *testing.T) {
	folders, err := os.ReadDir("testcases")
	require.NoError(t, err)
	require.NotEmpty(t, folders)

	for _, f := range folders {
		testcasePath := filepath.Join("testcases", f.Name())
		configFilename := filepath.Join(testcasePath, "telegraf.conf")
		t.Run(f.Name(), func(t *testing.T) {
			// Configure the plugin
			cfg := config.NewConfig()
			require.NoError(t, cfg.LoadConfig(configFilename))
			require.NoError(t, err)
			require.Len(t, cfg.Inputs, 1)

			// Tune the test-plugin
			plugin := cfg.Inputs[0].Input.(*test.Plugin)
			plugin.Path = testcasePath
			require.NoError(t, plugin.Init())

			// Gather the metrics and check for potential errors
			var acc testutil.Accumulator
			err := plugin.Gather(&acc)
			switch len(plugin.ExpectedErrors) {
			case 0:
				require.NoError(t, err)
			case 1:
				require.ErrorContains(t, err, plugin.ExpectedErrors[0])
			default:
				require.Contains(t, plugin.ExpectedErrors, err.Error())
			}

			// Determine checking options
			options := []cmp.Option{
				cmpopts.EquateApprox(0, 1e-6),
				testutil.SortMetrics(),
			}
			if plugin.ShouldIgnoreTimestamp {
				options = append(options, testutil.IgnoreTime())
			}

			// Process expected metrics and compare with resulting metrics
			actual := acc.GetTelegrafMetrics()
			testutil.RequireMetricsEqual(t, plugin.Expected, actual, options...)
		})
	}
}

func BenchmarkParsing(b *testing.B) {
	plugin := &Parser{}

	benchmarkData, err := os.ReadFile("testcases/benchmark/input.parquet")
	require.NoError(b, err)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		//nolint:errcheck // Benchmarking so skip the error check to avoid the unnecessary operations
		plugin.Parse(benchmarkData)
	}
}
