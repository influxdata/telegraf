package inputs_kafka_consumer_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/config"
	_ "github.com/influxdata/telegraf/migrations/inputs_kafka_consumer" // register migration
	"github.com/influxdata/telegraf/plugins/inputs/kafka_consumer"      // register plugin
	_ "github.com/influxdata/telegraf/plugins/parsers/influx"           // register parser
)

func TestNoMigration(t *testing.T) {
	plugin := &kafka_consumer.KafkaConsumer{}
	defaultCfg := plugin.SampleConfig()

	// Migrate and check that nothing changed
	output, n, err := config.ApplyMigrations([]byte(defaultCfg))
	require.NoError(t, err)
	require.NotEmpty(t, output)
	require.Zero(t, n)
	require.Equal(t, defaultCfg, string(output))
}

func TestStartupErrorBehaviorConflict(t *testing.T) {
	cfg := []byte(`
[[inputs.kafka_consumer]]
  brokers = ["localhost:9092"]
  topics = ["telegraf"]
  connection_strategy = "defer"
  startup_error_behavior = "ignore"
	`)
	// Migrate and check that the contradicting settings are caught
	output, n, err := config.ApplyMigrations(cfg)
	require.ErrorContains(t, err, "contradicting setting for 'startup_error_behavior' and 'connection_strategy'")
	require.Empty(t, output)
	require.Zero(t, n)
}

func TestInvalidConnectionStrategy(t *testing.T) {
	cfg := []byte(`
[[inputs.kafka_consumer]]
  brokers = ["localhost:9092"]
  topics = ["telegraf"]
  connection_strategy = "invalid"
	`)
	// Migrate and check that the invalid value is caught
	output, n, err := config.ApplyMigrations(cfg)
	require.ErrorContains(t, err, `invalid connection strategy "invalid"`)
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
