package inputs_smart_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/config"
	_ "github.com/influxdata/telegraf/migrations/inputs_smart" // register migration
	_ "github.com/influxdata/telegraf/plugins/inputs/smart"    // register plugin
)

func TestCases(t *testing.T) {
	testcases := []struct {
		folder     string
		shouldFail bool
	}{
		{
			folder:     "standard",
			shouldFail: false,
		},
		{
			folder:     "failed",
			shouldFail: true,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.folder, func(t *testing.T) {
			testcasePath := filepath.Join("testcases", testcase.folder)
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

			if testcase.shouldFail {
				require.Empty(t, output)
				return
			}
			require.NotEmpty(t, output)
			require.GreaterOrEqual(t, uint64(1), n)
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
			require.ElementsMatchf(t, expectedIDs, actualIDs, "generated config: %s", string(output))
		})
	}
}
