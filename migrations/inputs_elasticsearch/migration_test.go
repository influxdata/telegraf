package inputs_elasticsearch_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/config"
	_ "github.com/influxdata/telegraf/migrations/inputs_elasticsearch" // register migration
	"github.com/influxdata/telegraf/plugins/inputs/elasticsearch"
)

func TestNoMigration(t *testing.T) {
	plugin := &elasticsearch.Elasticsearch{}
	defaultCfg := []byte(plugin.SampleConfig())

	// Migrate and check that nothing changed
	output, n, err := config.ApplyMigrations(defaultCfg)
	require.NoError(t, err)
	require.NotEmpty(t, output)
	require.Zero(t, n)
	require.Equal(t, string(defaultCfg), string(output))
}

func TestTimeoutConflict(t *testing.T) {
	cfg := []byte(`
[[inputs.elasticsearch]]
  servers = ["http://localhost:9200"]
  http_timeout = "10s"
  timeout = "15s"
	`)

	// Migrate and check that it fails with conflict error
	output, n, err := config.ApplyMigrations(cfg)
	require.ErrorContains(t, err, "contradicting setting for 'http_timeout' (10s) and 'timeout' (15s)")
	require.Empty(t, output)
	require.Zero(t, n)
}

func TestCases(t *testing.T) {
	// Get all directories in testcases
	folders, err := os.ReadDir("testcases")
	require.NoError(t, err)

	for _, f := range folders {
		// Only handle folders
		if !f.IsDir() {
			continue
		}

		t.Run(f.Name(), func(t *testing.T) {
			testcasePath := filepath.Join("testcases", f.Name())
			inputFile := filepath.Join(testcasePath, "telegraf.conf")
			expectedFile := filepath.Join(testcasePath, "expected.conf")

			// Read the expected output
			expected := config.NewConfig()
			require.NoError(t, expected.LoadConfig(expectedFile))
			require.NotEmpty(t, expected.Inputs)

			// Read the input data
			input, remote, err := config.LoadConfigFile(inputFile)
			require.NoError(t, err)
			require.False(t, remote)
			require.NotEmpty(t, input)

			// Migrate
			output, n, err := config.ApplyMigrations(input)
			require.NoError(t, err)
			require.NotEmpty(t, output)
			require.GreaterOrEqual(t, n, uint64(1))

			actual := config.NewConfig()
			require.NoError(t, actual.LoadConfigData(output, config.EmptySourcePath))

			// Test the output
			require.Len(t, actual.Inputs, len(expected.Inputs))

			actualIDs := make([]string, 0, len(expected.Inputs))
			expectedIDs := make([]string, 0, len(expected.Inputs))
			for i := range actual.Inputs {
				actualIDs = append(actualIDs, actual.Inputs[i].ID())
				expectedIDs = append(expectedIDs, expected.Inputs[i].ID())
			}
			require.ElementsMatch(t, expectedIDs, actualIDs, string(output))
		})
	}
}
