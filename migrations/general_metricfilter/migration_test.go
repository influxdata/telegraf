package general_metricfilter_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	_ "github.com/influxdata/telegraf/migrations/general_metricfilter" // register migration
	"github.com/influxdata/telegraf/plugins/inputs"
)

func TestNoMigration(t *testing.T) {
	cfg := []byte(`
# Dummy plugin
[[inputs.dummy]]
  ## A dummy server
  servers = ["tcp://127.0.0.1:1883"]

  ## A commented option
  # timeout = "10s"
`)

	// Migrate and check that nothing changed
	output, n, err := config.ApplyMigrations(cfg)
	require.NoError(t, err)
	require.NotEmpty(t, output)
	require.Zero(t, n)
	require.Equal(t, string(cfg), string(output))
}

func TestCases(t *testing.T) {
	// Get all directories in testdata
	folders, err := os.ReadDir("testcases")
	require.NoError(t, err)

	inputs.Add("dummy", func() telegraf.Input { return &MockupInputPlugin{} })

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
			require.NoError(t, actual.LoadConfigData(output))

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

// Implement a mock input plugin for testing
type MockupInputPlugin struct {
	Servers []string        `toml:"servers"`
	Timeout config.Duration `toml:"timeout"`
}

func (m *MockupInputPlugin) SampleConfig() string {
	return "Mockup test input plugin"
}
func (m *MockupInputPlugin) Gather(_ telegraf.Accumulator) error {
	return nil
}
