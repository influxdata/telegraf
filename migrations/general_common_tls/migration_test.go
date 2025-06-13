package general_common_tls_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	_ "github.com/influxdata/telegraf/migrations/general_common_tls" // register migration
	common_tls "github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

func TestNoMigration(t *testing.T) {
	tests := []struct {
		name string
		cfg  string
	}{
		{
			name: "input",
			cfg: `
# Dummy plugin
[[inputs.dummy]]
  ## A dummy server
  servers = ["tcp://127.0.0.1:1883"]

  ## A commented option
  # timeout = "10s"
`,
		},
		{
			name: "output",
			cfg: `
# Dummy plugin
[[output.dummy]]
  ## A dummy server
  servers = ["tcp://127.0.0.1:1883"]

  ## A commented option
  # timeout = "10s"
`,
		},
		{
			name: "processor",
			cfg: `
# Dummy plugin
[[processor.dummy]]
  ## A dummy server
  servers = ["tcp://127.0.0.1:1883"]

  ## A commented option
  # timeout = "10s"
`,
		},
		{
			name: "aggregator",
			cfg: `
# Dummy plugin
[[aggregator.dummy]]
  ## A dummy server
  servers = ["tcp://127.0.0.1:1883"]

  ## A commented option
  # timeout = "10s"
`,
		},
		{
			name: "no old setting",
			cfg: `
# Dummy plugin
[[input.dummy]]
  ## A dummy server
  servers = ["tcp://127.0.0.1:1883"]

  # TLS settings
  tls_ca = "/etc/ca.pem"
  tls_cert = "/etc/cert.pem"
  tls_key = "/etc/key"


  ## A commented option
  # timeout = "10s"
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := []byte(tt.cfg)

			// Migrate and check that nothing changed
			output, n, err := config.ApplyMigrations(cfg)
			require.NoError(t, err)
			require.NotEmpty(t, output)
			require.Zero(t, n)
			require.Equal(t, tt.cfg, string(output))
		})
	}
}

func TestConflict(t *testing.T) {
	tests := []struct {
		name string
		cfg  string
	}{
		{
			name: "ca",
			cfg: `
# Dummy plugin
[[inputs.dummy]]
  ## A dummy server
  servers = ["tcp://127.0.0.1:1883"]
  tls_ca = "foo.pem"
  ssl_ca = "bar.pem"

  ## A commented option
  # timeout = "10s"
`,
		},
		{
			name: "cert",
			cfg: `
# Dummy plugin
[[inputs.dummy]]
  ## A dummy server
  servers = ["tcp://127.0.0.1:1883"]
  tls_cert = "foo.pem"
  ssl_cert = "bar.pem"

  ## A commented option
  # timeout = "10s"
`,
		},
		{
			name: "key",
			cfg: `
# Dummy plugin
[[inputs.dummy]]
  ## A dummy server
  servers = ["tcp://127.0.0.1:1883"]
  tls_key = "foo.pem"
  ssl_key = "bar.pem"

  ## A commented option
  # timeout = "10s"
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := []byte(tt.cfg)

			// Migrate and check that nothing changed
			output, n, err := config.ApplyMigrations(cfg)
			require.ErrorContains(t, err, "contradicting setting")
			require.Empty(t, output)
			require.Zero(t, n)
		})
	}
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

// Implement a mock input plugin for testing
type MockupInputPlugin struct {
	Servers []string        `toml:"servers"`
	Timeout config.Duration `toml:"timeout"`
	common_tls.ClientConfig
}

func (*MockupInputPlugin) SampleConfig() string {
	return "Mockup test input plugin"
}
func (*MockupInputPlugin) Gather(telegraf.Accumulator) error {
	return nil
}
