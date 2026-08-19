package inputs_zookeeper_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/config"
	_ "github.com/influxdata/telegraf/migrations/inputs_zookeeper" // register migration
	"github.com/influxdata/telegraf/plugins/inputs/zookeeper"      // register plugin
)

func TestNoMigration(t *testing.T) {
	plugin := &zookeeper.Zookeeper{}
	defaultCfg := plugin.SampleConfig()

	// Migrate and check that nothing changed
	output, n, err := config.ApplyMigrations([]byte(defaultCfg))
	require.NoError(t, err)
	require.NotEmpty(t, output)
	require.Zero(t, n)
	require.Equal(t, defaultCfg, string(output))
}

func TestEnableTLSConflict(t *testing.T) {
	cfg := []byte(`
[[inputs.zookeeper]]
  servers = ["localhost:2181"]
  enable_tls = true
  tls_enable = false
	`)
	// Migrate and check that the contradicting settings are caught
	output, n, err := config.ApplyMigrations(cfg)
	require.ErrorContains(t, err, "contradicting setting for 'tls_enable' and 'enable_tls'")
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
