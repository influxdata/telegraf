package prometheus

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/models"
	"github.com/influxdata/telegraf/testutil"
	test "github.com/influxdata/telegraf/testutil/plugin_input"
)

func TestParseText(t *testing.T) {
	const input = "# TYPE xtars_reachability_worker_utilization_ratio gauge\n" +
		"xtars_reachability_worker_utilization_ratio 0"

	tests := []struct {
		name    string
		input   string
		header  http.Header
		wantErr bool
	}{
		{name: "missing_final_newline", input: input},
		{name: "with_final_newline", input: input + "\n"},
		{name: "text_header", input: input, header: http.Header{"Content-Type": {"text/plain; version=0.0.4"}}},
		{name: "unknown_header", input: input, header: http.Header{"Content-Type": {"application/octet-stream"}}},
		{name: "invalid_value", input: "metric invalid", wantErr: true},
		{name: "incomplete_label", input: "metric{label=\"unfinished", wantErr: true},
	}

	for _, version := range []int{0, 1, 2} {
		for _, tt := range tests {
			t.Run(fmt.Sprintf("%s_v%d", tt.name, version), func(t *testing.T) {
				parser := &Parser{MetricVersion: version, Header: tt.header, Log: testutil.Logger{}}
				require.NoError(t, parser.Init())

				metrics, err := parser.Parse([]byte(tt.input))
				if tt.wantErr {
					require.Error(t, err)
					require.Empty(t, metrics)
					return
				}
				require.NoError(t, err)
				require.Len(t, metrics, 1)
				if version == 1 {
					require.Equal(t, "xtars_reachability_worker_utilization_ratio", metrics[0].Name())
					require.Equal(t, map[string]any{"gauge": float64(0)}, metrics[0].Fields())
				} else {
					require.Equal(t, "prometheus", metrics[0].Name())
					require.Equal(t, map[string]any{"xtars_reachability_worker_utilization_ratio": float64(0)}, metrics[0].Fields())
				}
			})
		}
	}
}

func TestCases(t *testing.T) {
	// Get all directories in testcases
	folders, err := os.ReadDir("testcases")
	require.NoError(t, err)
	// Make sure testdata contains data
	require.NotEmpty(t, folders)

	for _, f := range folders {
		fname := f.Name()
		testdataPath := filepath.Join("testcases", fname)
		configFilename := filepath.Join(testdataPath, "telegraf.conf")

		// Run tests as metric version 1
		t.Run(fname+"_v1", func(t *testing.T) {
			// Load the configuration
			cfg := config.NewConfig()
			require.NoError(t, cfg.LoadConfig(configFilename))
			require.Len(t, cfg.Inputs, 1)

			// Tune plugin
			plugin := cfg.Inputs[0].Input.(*test.Plugin)
			plugin.Path = testdataPath
			plugin.UseTypeTag = "_type"
			plugin.ExpectedFilename = "expected_v1.out"

			parser := plugin.Parser.(*models.RunningParser).Parser.(*Parser)
			parser.MetricVersion = 1
			if raw, found := plugin.AdditionalParams["headers"]; found {
				headers, ok := raw.(map[string]any)
				require.Truef(t, ok, "unknown header type %T", raw)
				parser.Header = make(http.Header)
				for k, rv := range headers {
					v, ok := rv.(string)
					require.Truef(t, ok, "unknown header value type %T for %q", raw, k)
					parser.Header.Add(k, v)
				}
			}
			require.NoError(t, plugin.Init())

			// Gather data and check errors
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
				testutil.SortMetrics(),
			}
			if plugin.ShouldIgnoreTimestamp || parser.IgnoreTimestamp {
				options = append(options, testutil.IgnoreTime())
			}

			// Check the resulting metrics
			actual := acc.GetTelegrafMetrics()
			testutil.RequireMetricsEqual(t, plugin.Expected, actual, options...)

			// Special checks
			if parser.IgnoreTimestamp {
				t.Log("testing ignore-timestamp case")
				for i, m := range actual {
					expected := plugin.Expected[i]
					require.Greaterf(t, m.Time(), expected.Time(), "metric time not after prometheus value in %d", i)
				}
			}
		})

		// Run tests as metric version 2
		t.Run(fname+"_v2", func(t *testing.T) {
			// Load the configuration
			cfg := config.NewConfig()
			require.NoError(t, cfg.LoadConfig(configFilename))
			require.Len(t, cfg.Inputs, 1)

			// Tune plugin
			plugin := cfg.Inputs[0].Input.(*test.Plugin)
			plugin.Path = testdataPath
			plugin.UseTypeTag = "_type"
			plugin.ExpectedFilename = "expected_v2.out"

			parser := plugin.Parser.(*models.RunningParser).Parser.(*Parser)
			parser.MetricVersion = 2
			if raw, found := plugin.AdditionalParams["headers"]; found {
				headers, ok := raw.(map[string]any)
				require.Truef(t, ok, "unknown header type %T", raw)
				parser.Header = make(http.Header)
				for k, rv := range headers {
					v, ok := rv.(string)
					require.Truef(t, ok, "unknown header value type %T for %q", raw, k)
					parser.Header.Add(k, v)
				}
			}
			require.NoError(t, plugin.Init())

			// Gather data and check errors
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
				testutil.SortMetrics(),
			}
			if plugin.ShouldIgnoreTimestamp || parser.IgnoreTimestamp {
				options = append(options, testutil.IgnoreTime())
			}

			// Check the resulting metrics
			actual := acc.GetTelegrafMetrics()
			testutil.RequireMetricsEqual(t, plugin.Expected, actual, options...)

			// Special checks
			if parser.IgnoreTimestamp {
				t.Log("testing ignore-timestamp case")
				for i, m := range actual {
					expected := plugin.Expected[i]
					require.Greaterf(t, m.Time(), expected.Time(), "metric time not after prometheus value in %d", i)
				}
			}
		})
	}
}

func BenchmarkParsingMetricVersion1(b *testing.B) {
	plugin := &Parser{MetricVersion: 1}
	require.NoError(b, plugin.Init())

	benchmarkData, err := os.ReadFile(filepath.FromSlash("testcases/benchmark/input.txt"))
	require.NoError(b, err)
	require.NotEmpty(b, benchmarkData)

	for n := 0; n < b.N; n++ {
		//nolint:errcheck // Benchmarking so skip the error check to avoid the unnecessary operations
		plugin.Parse(benchmarkData)
	}
}

func BenchmarkParsingMetricVersion2(b *testing.B) {
	plugin := &Parser{MetricVersion: 2}
	require.NoError(b, plugin.Init())

	benchmarkData, err := os.ReadFile(filepath.FromSlash("testcases/benchmark/input.txt"))
	require.NoError(b, err)
	require.NotEmpty(b, benchmarkData)

	for n := 0; n < b.N; n++ {
		//nolint:errcheck // Benchmarking so skip the error check to avoid the unnecessary operations
		plugin.Parse(benchmarkData)
	}
}
