package inputs_tail_test

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/config"
	_ "github.com/influxdata/telegraf/migrations/inputs_tail" // register migration
	"github.com/influxdata/telegraf/plugins/inputs/tail"      // register plugin
	_ "github.com/influxdata/telegraf/plugins/parsers/influx" // register parser
)

func TestNoMigration(t *testing.T) {
	plugin := &tail.Tail{}
	defaultCfg := plugin.SampleConfig()

	// Migrate and check that nothing changed
	output, n, err := config.ApplyMigrations([]byte(defaultCfg))
	require.NoError(t, err)
	require.NotEmpty(t, output)
	require.Zero(t, n)
	require.Equal(t, defaultCfg, string(output))
}

func TestConflictWarning(t *testing.T) {
	tests := []struct {
		name string
		conf string
		warn bool
	}{
		{
			name: "conflicting values warn",
			conf: `
[[inputs.tail]]
  files = ["/var/mymetrics.out"]
  from_beginning = true
  initial_read_offset = "end"
  data_format = "influx"
`,
			warn: true,
		},
		{
			name: "agreeing values do not warn",
			conf: `
[[inputs.tail]]
  files = ["/var/mymetrics.out"]
  from_beginning = true
  initial_read_offset = "beginning"
  data_format = "influx"
`,
			warn: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			log.SetOutput(&buf)
			t.Cleanup(func() { log.SetOutput(os.Stderr) })

			output, n, err := config.ApplyMigrations([]byte(tt.conf))
			require.NoError(t, err)
			require.NotEmpty(t, output)
			require.Equal(t, uint64(1), n)

			if tt.warn {
				require.Contains(t, buf.String(), `conflicts with 'initial_read_offset = "end"'`)
			} else {
				require.NotContains(t, buf.String(), "conflicts with")
			}
		})
	}
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
