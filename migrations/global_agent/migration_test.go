package global_agent_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/config"
	_ "github.com/influxdata/telegraf/migrations/global_agent" // register migration
)

func TestNoMigration(t *testing.T) {
	fn := filepath.Join("testcases", "default.conf")

	// Read the input data
	input, remote, err := config.LoadConfigFile(fn)
	require.NoError(t, err)
	require.False(t, remote)
	require.NotEmpty(t, input)

	// Expect the output to be equal to the input
	expectedBuffer, err := os.ReadFile(fn)
	require.NoError(t, err)
	expected := config.NewConfig()
	require.NoError(t, expected.LoadConfigData(expectedBuffer, config.EmptySourcePath))
	require.NotNil(t, expected.Agent)

	// Migrate
	output, n, err := config.ApplyMigrations(input)
	require.NoError(t, err)
	require.NotEmpty(t, output)
	require.Zero(t, n)

	actual := config.NewConfig()
	require.NoError(t, actual.LoadConfigData(output, config.EmptySourcePath))
	require.NotNil(t, actual.Agent)

	// Test the output
	require.EqualValues(t, expected.Agent, actual.Agent, string(output))
	require.Equal(t, string(expectedBuffer), string(output))
}

func TestLogTargetEventlogCollision(t *testing.T) {
	fn := filepath.Join("testcases", "logtarget_eventlog_collision.conf")

	// Read the input data
	input, remote, err := config.LoadConfigFile(fn)
	require.NoError(t, err)
	require.False(t, remote)
	require.NotEmpty(t, input)

	// Migrate
	_, n, err := config.ApplyMigrations(input)
	require.ErrorContains(t, err, "contradicting setting for 'logtarget' and 'logformat'")
	require.Zero(t, n)
}

func TestCases(t *testing.T) {
	// Get all directories in testdata
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
			require.NotNil(t, expected.Agent)

			// Read the input data
			input, remote, err := config.LoadConfigFile(inputFile)
			require.NoError(t, err)
			require.False(t, remote)
			require.NotEmpty(t, input)

			// Migrate
			output, n, err := config.ApplyMigrations(input)
			require.NoError(t, err)
			require.NotEmpty(t, output)
			require.Positive(t, n, "expected migration application but none applied")
			actual := config.NewConfig()
			require.NoError(t, actual.LoadConfigData(output, config.EmptySourcePath))
			require.NotNil(t, actual.Agent)

			// Test the output
			require.EqualValues(t, expected.Agent, actual.Agent, string(output))

			expectedBuffer, err := os.ReadFile(expectedFile)
			require.NoError(t, err)
			require.Equal(t, string(expectedBuffer), string(output))
		})
	}
}
