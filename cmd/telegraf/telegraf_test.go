package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/stretchr/testify/require"
)

type MockManager struct{}

func NewMockManager() *MockManager {
	return &MockManager{}
}

func (m *MockManager) Init(serverErr <-chan error, f Filters, g GlobalFlags, w WindowFlags) {

}

func (m *MockManager) Run() error {
	return nil
}

type MockConfig struct {
	Buffer                    io.Writer
	ExpectedDeprecatedPlugins map[string][]config.PluginDeprecationInfo
}

func NewMockConfig(buffer io.Writer) *MockConfig {
	return &MockConfig{
		Buffer: buffer,
	}
}

func (m *MockConfig) CollectDeprecationInfos(inFilter, outFilter, aggFilter, procFilter []string) map[string][]config.PluginDeprecationInfo {
	return m.ExpectedDeprecatedPlugins
}

func (m *MockConfig) PrintDeprecationList(plugins []config.PluginDeprecationInfo) {
	for _, p := range plugins {
		_, _ = m.Buffer.Write([]byte(fmt.Sprintf("plugin name: %s\n", p.Name)))
	}
}

type MockServer struct {
	Address string
}

func NewMockServer() *MockServer {
	return &MockServer{}
}

func (m *MockServer) Start(address string) {
	m.Address = "localhost:6060"
}

func (m *MockServer) ErrChan() <-chan error {
	return nil
}

func TestUsageFlag(t *testing.T) {
	tests := []struct {
		PluginName     string
		ExpectedError  string
		ExpectedOutput string
	}{
		{
			PluginName:    "example",
			ExpectedError: "E! input example not found and output example not found",
		},
		{
			PluginName: "temp",
			ExpectedOutput: `
# Read metrics about temperature
[[inputs.temp]]
  # no configuration

`,
		},
	}

	for _, test := range tests {
		buf := new(bytes.Buffer)
		args := os.Args[0:1]
		args = append(args, "--usage", test.PluginName)
		err := runApp(args, buf, NewMockServer(), NewMockConfig(buf), NewMockManager())
		if test.ExpectedError != "" {
			require.ErrorContains(t, err, test.ExpectedError)
			continue
		}
		require.NoError(t, err)
		// To run this test on windows and linux, remove windows carriage return
		o := strings.Replace(buf.String(), "\r", "", -1)
		require.Equal(t, test.ExpectedOutput, o)
	}
}

func TestInputListFlag(t *testing.T) {
	buf := new(bytes.Buffer)
	args := os.Args[0:1]
	args = append(args, "--input-list")
	temp := inputs.Inputs
	inputs.Inputs = map[string]inputs.Creator{
		"test": func() telegraf.Input { return nil },
	}
	err := runApp(args, buf, NewMockServer(), NewMockConfig(buf), NewMockManager())
	require.NoError(t, err)
	expectedOutput := `Available Input Plugins:
  test
`
	require.Equal(t, expectedOutput, buf.String())
	inputs.Inputs = temp
}

func TestOutputListFlag(t *testing.T) {
	buf := new(bytes.Buffer)
	args := os.Args[0:1]
	args = append(args, "--output-list")
	temp := outputs.Outputs
	outputs.Outputs = map[string]outputs.Creator{
		"test": func() telegraf.Output { return nil },
	}
	err := runApp(args, buf, NewMockServer(), NewMockConfig(buf), NewMockManager())
	require.NoError(t, err)
	expectedOutput := `Available Output Plugins:
  test
`
	require.Equal(t, expectedOutput, buf.String())
	outputs.Outputs = temp
}

func TestDeprecationListFlag(t *testing.T) {
	buf := new(bytes.Buffer)
	args := os.Args[0:1]
	args = append(args, "--deprecation-list")
	mS := NewMockServer()
	mC := NewMockConfig(buf)
	mC.ExpectedDeprecatedPlugins = make(map[string][]config.PluginDeprecationInfo)
	mC.ExpectedDeprecatedPlugins["inputs"] = []config.PluginDeprecationInfo{
		{
			DeprecationInfo: config.DeprecationInfo{
				Name: "test",
			},
		},
	}
	err := runApp(args, buf, mS, mC, NewMockManager())
	require.NoError(t, err)
	expectedOutput := `Deprecated Input Plugins:
plugin name: test
Deprecated Output Plugins:
Deprecated Processor Plugins:
Deprecated Aggregator Plugins:
`

	require.Equal(t, expectedOutput, buf.String())
}

func TestPprofAddressFlag(t *testing.T) {
	buf := new(bytes.Buffer)
	args := os.Args[0:1]
	address := "localhost:6060"
	args = append(args, "--pprof-addr", address)
	m := NewMockServer()
	err := runApp(args, buf, m, NewMockConfig(buf), NewMockManager())
	require.NoError(t, err)
	require.Equal(t, address, m.Address)
}

// !!! DEPRECATED !!!
// TestPluginDirectoryFlag tests `--plugin-directory`
func TestPluginDirectoryFlag(t *testing.T) {
	buf := new(bytes.Buffer)
	args := os.Args[0:1]
	args = append(args, "--plugin-directory", ".")
	err := runApp(args, buf, NewMockServer(), NewMockConfig(buf), NewMockManager())
	require.ErrorContains(t, err, "E! go plugin support is not enabled")
}

func TestCommandConfig(t *testing.T) {
	tests := []struct {
		name            string
		commands        []string
		expectedHeaders []string
		removedHeaders  []string
		expectedPlugins []string
		removedPlugins  []string
	}{
		// Deprecated flag replaced with command "config"
		{
			name:     "no filters",
			commands: []string{"--sample-config"},
			expectedHeaders: []string{
				outputHeader,
				inputHeader,
				aggregatorHeader,
				processorHeader,
				serviceInputHeader,
			},
		},
		{
			name:     "no filters",
			commands: []string{"config"},
			expectedHeaders: []string{
				outputHeader,
				inputHeader,
				aggregatorHeader,
				processorHeader,
				serviceInputHeader,
			},
		},
		{
			name:     "filter sections for inputs",
			commands: []string{"config", "--section-filter", "inputs"},
			expectedHeaders: []string{
				inputHeader,
			},
			removedHeaders: []string{
				outputHeader,
				aggregatorHeader,
				processorHeader,
			},
		},
		{
			name:     "filter sections for inputs,outputs",
			commands: []string{"config", "--section-filter", "inputs:outputs"},
			expectedHeaders: []string{
				inputHeader,
				outputHeader,
			},
			removedHeaders: []string{
				aggregatorHeader,
				processorHeader,
			},
		},
		{
			name:     "filter input plugins",
			commands: []string{"config", "--input-filter", "cpu:file"},
			expectedPlugins: []string{
				"[[inputs.cpu]]",
				"[[inputs.file]]",
			},
			removedPlugins: []string{
				"[[inputs.disk]]",
			},
		},
		{
			name:     "filter output plugins",
			commands: []string{"config", "--output-filter", "influxdb:http"},
			expectedPlugins: []string{
				"[[outputs.influxdb]]",
				"[[outputs.http]]",
			},
			removedPlugins: []string{
				"[[outputs.file]]",
			},
		},
		{
			name:     "filter processor plugins",
			commands: []string{"config", "--processor-filter", "date:enum"},
			expectedPlugins: []string{
				"[[processors.date]]",
				"[[processors.enum]]",
			},
			removedPlugins: []string{
				"[[processors.parser]]",
			},
		},
		{
			name:     "filter aggregator plugins",
			commands: []string{"config", "--aggregator-filter", "basicstats:starlark"},
			expectedPlugins: []string{
				"[[aggregators.basicstats]]",
				"[[aggregators.starlark]]",
			},
			removedPlugins: []string{
				"[[aggregators.minmax]]",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			args := os.Args[0:1]
			args = append(args, test.commands...)
			err := runApp(args, buf, NewMockServer(), NewMockConfig(buf), NewMockManager())
			require.NoError(t, err)
			output := buf.String()
			for _, e := range test.expectedHeaders {
				require.Contains(t, output, e, "expected header not found")
			}
			for _, r := range test.removedHeaders {
				require.NotContains(t, output, r, "removed header found")
			}
			for _, e := range test.expectedPlugins {
				require.Contains(t, output, e, "expected plugin not found")
			}
			for _, r := range test.removedPlugins {
				require.NotContains(t, output, r, "removed plugin found")
			}
		})
	}
}

func TestCommandVersion(t *testing.T) {
	tests := []struct {
		Version        string
		Branch         string
		Commit         string
		ExpectedOutput string
	}{
		{
			Version:        "v2.0.0",
			ExpectedOutput: "Telegraf v2.0.0",
		},
		{
			ExpectedOutput: "Telegraf unknown",
		},
		{
			Version:        "v2.0.0",
			Branch:         "master",
			ExpectedOutput: "Telegraf v2.0.0 (git: master@unknown)",
		},
		{
			Version:        "v2.0.0",
			Branch:         "master",
			Commit:         "123",
			ExpectedOutput: "Telegraf v2.0.0 (git: master@123)",
		},
		{
			Version:        "v2.0.0",
			Commit:         "123",
			ExpectedOutput: "Telegraf v2.0.0 (git: unknown@123)",
		},
	}

	for _, test := range tests {
		buf := new(bytes.Buffer)
		args := os.Args[0:1]
		args = append(args, "version")
		internal.Version = test.Version
		internal.Branch = test.Branch
		internal.Commit = test.Commit
		err := runApp(args, buf, NewMockServer(), NewMockConfig(buf), NewMockManager())
		require.NoError(t, err)
		require.Equal(t, test.ExpectedOutput, buf.String())
	}
}

// Deprecated in favor of command version
func TestFlagVersion(t *testing.T) {
	tests := []struct {
		Version        string
		Branch         string
		Commit         string
		ExpectedOutput string
	}{
		{
			Version:        "v2.0.0",
			ExpectedOutput: "Telegraf v2.0.0",
		},
		{
			ExpectedOutput: "Telegraf unknown",
		},
		{
			Version:        "v2.0.0",
			Branch:         "master",
			ExpectedOutput: "Telegraf v2.0.0 (git: master@unknown)",
		},
		{
			Version:        "v2.0.0",
			Branch:         "master",
			Commit:         "123",
			ExpectedOutput: "Telegraf v2.0.0 (git: master@123)",
		},
		{
			Version:        "v2.0.0",
			Commit:         "123",
			ExpectedOutput: "Telegraf v2.0.0 (git: unknown@123)",
		},
	}

	for _, test := range tests {
		buf := new(bytes.Buffer)
		args := os.Args[0:1]
		args = append(args, "--version")
		internal.Version = test.Version
		internal.Branch = test.Branch
		internal.Commit = test.Commit
		err := runApp(args, buf, NewMockServer(), NewMockConfig(buf), NewMockManager())
		require.NoError(t, err)
		require.Equal(t, test.ExpectedOutput, buf.String())
	}
}
