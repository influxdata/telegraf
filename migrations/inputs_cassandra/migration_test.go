package inputs_cassandra_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/influxdata/telegraf/config"
	_ "github.com/influxdata/telegraf/migrations/inputs_cassandra" // register migration
	"github.com/stretchr/testify/require"
)

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
			expected, remote, err := config.LoadConfigFile(expectedFile)
			require.NoError(t, err)
			require.False(t, remote)
			require.NotEmpty(t, expected)

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

			// Test the output
			require.Equal(t, string(expected), string(output))
		})
	}
}
