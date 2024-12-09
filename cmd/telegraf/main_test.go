package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/outputs"
)

var secrets = map[string]map[string][]byte{
	"yoda": {
		"episode1": []byte("member"),
		"episode2": []byte("member"),
		"episode3": []byte("member"),
	},
	"mace_windu": {
		"episode1": []byte("member"),
		"episode2": []byte("member"),
		"episode3": []byte("member"),
	},
	"oppo_rancisis": {
		"episode1": []byte("member"),
		"episode2": []byte("member"),
	},
	"coleman_kcaj": {
		"episode3": []byte("member"),
	},
}

type MockTelegraf struct {
	GlobalFlags
	WindowFlags
}

func NewMockTelegraf() *MockTelegraf {
	return &MockTelegraf{}
}

func (m *MockTelegraf) Init(_ <-chan error, _ Filters, g GlobalFlags, w WindowFlags) {
	m.GlobalFlags = g
	m.WindowFlags = w
}

func (m *MockTelegraf) Run() error {
	return nil
}

func (m *MockTelegraf) ListSecretStores() ([]string, error) {
	ids := make([]string, 0, len(secrets))
	for k := range secrets {
		ids = append(ids, k)
	}
	return ids, nil
}

func (m *MockTelegraf) GetSecretStore(id string) (telegraf.SecretStore, error) {
	v, found := secrets[id]
	if !found {
		return nil, errors.New("unknown secret store")
	}
	s := &MockSecretStore{Secrets: v}
	return s, nil
}

type MockSecretStore struct {
	Secrets map[string][]byte
}

func (s *MockSecretStore) Init() error {
	return nil
}

func (s *MockSecretStore) SampleConfig() string {
	return "I'm just a dummy"
}

func (s *MockSecretStore) Get(key string) ([]byte, error) {
	v, found := s.Secrets[key]
	if !found {
		return nil, errors.New("not found")
	}
	return v, nil
}

func (s *MockSecretStore) Set(key, value string) error {
	if strings.HasPrefix(key, "darth") {
		return errors.New("don't join the dark side")
	}
	s.Secrets[key] = []byte(value)
	return nil
}
func (s *MockSecretStore) List() ([]string, error) {
	keys := make([]string, 0, len(s.Secrets))
	for k := range s.Secrets {
		keys = append(keys, k)
	}
	return keys, nil
}

func (s *MockSecretStore) GetResolver(key string) (telegraf.ResolveFunc, error) {
	return func() ([]byte, bool, error) {
		v, err := s.Get(key)
		return v, false, err
	}, nil
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

func (m *MockConfig) CollectDeprecationInfos(_, _, _, _ []string) map[string][]config.PluginDeprecationInfo {
	return m.ExpectedDeprecatedPlugins
}

func (m *MockConfig) PrintDeprecationList(plugins []config.PluginDeprecationInfo) {
	for _, p := range plugins {
		fmt.Fprintf(m.Buffer, "plugin name: %s\n", p.Name)
	}
}

type MockServer struct {
	Address string
}

func NewMockServer() *MockServer {
	return &MockServer{}
}

func (m *MockServer) Start(_ string) {
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
			ExpectedError: "input example not found and output example not found",
		},
		{
			PluginName: "temp",
			ExpectedOutput: `
# Read metrics about temperature
[[inputs.temp]]
  ## Desired output format (Linux only)
  ## Available values are
  ##   v1 -- use pre-v1.22.4 sensor naming, e.g. coretemp_core0_input
  ##   v2 -- use v1.22.4+ sensor naming, e.g. coretemp_core_0_input
  # metric_format = "v2"

  ## Add device tag to distinguish devices with the same name (Linux only)
  # add_device_tag = false

`,
		},
	}

	for _, test := range tests {
		buf := new(bytes.Buffer)
		args := os.Args[0:1]
		args = append(args, "--usage", test.PluginName)
		err := runApp(args, buf, NewMockServer(), NewMockConfig(buf), NewMockTelegraf())
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
	err := runApp(args, buf, NewMockServer(), NewMockConfig(buf), NewMockTelegraf())
	require.NoError(t, err)
	expectedOutput := `DEPRECATED: use telegraf plugins inputs
Available Input Plugins:
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
	err := runApp(args, buf, NewMockServer(), NewMockConfig(buf), NewMockTelegraf())
	require.NoError(t, err)
	expectedOutput := `DEPRECATED: use telegraf plugins outputs
Available Output Plugins:
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
	err := runApp(args, buf, mS, mC, NewMockTelegraf())
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
	err := runApp(args, buf, m, NewMockConfig(buf), NewMockTelegraf())
	require.NoError(t, err)
	require.Equal(t, address, m.Address)
}

// !!! DEPRECATED !!!
// TestPluginDirectoryFlag tests `--plugin-directory`
func TestPluginDirectoryFlag(t *testing.T) {
	buf := new(bytes.Buffer)
	args := os.Args[0:1]
	args = append(args, "--plugin-directory", ".")
	err := runApp(args, buf, NewMockServer(), NewMockConfig(buf), NewMockTelegraf())
	require.ErrorContains(t, err, "go plugin support is not enabled")
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
		{
			name:     "deprecated flag --sample-config",
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
		{
			name:     "test filters before config",
			commands: []string{"--input-filter", "cpu:file", "config"},
			expectedPlugins: []string{
				"[[inputs.cpu]]",
				"[[inputs.file]]",
			},
			removedPlugins: []string{
				"[[inputs.disk]]",
			},
		},
		{
			name:     "test filters before and after config",
			commands: []string{"--input-filter", "file", "config", "--input-filter", "cpu"},
			expectedPlugins: []string{
				"[[inputs.cpu]]",
				"[[inputs.file]]",
			},
			removedPlugins: []string{
				"[[inputs.disk]]",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			args := os.Args[0:1]
			args = append(args, test.commands...)
			err := runApp(args, buf, NewMockServer(), NewMockConfig(buf), NewMockTelegraf())
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
			ExpectedOutput: "Telegraf v2.0.0\n",
		},
		{
			ExpectedOutput: "Telegraf unknown\n",
		},
		{
			Version:        "v2.0.0",
			Branch:         "master",
			ExpectedOutput: "Telegraf v2.0.0 (git: master@unknown)\n",
		},
		{
			Version:        "v2.0.0",
			Branch:         "master",
			Commit:         "123",
			ExpectedOutput: "Telegraf v2.0.0 (git: master@123)\n",
		},
		{
			Version:        "v2.0.0",
			Commit:         "123",
			ExpectedOutput: "Telegraf v2.0.0 (git: unknown@123)\n",
		},
	}

	for _, test := range tests {
		buf := new(bytes.Buffer)
		args := os.Args[0:1]
		args = append(args, "version")
		internal.Version = test.Version
		internal.Branch = test.Branch
		internal.Commit = test.Commit
		err := runApp(args, buf, NewMockServer(), NewMockConfig(buf), NewMockTelegraf())
		require.NoError(t, err)
		require.Equal(t, test.ExpectedOutput, buf.String())
	}
}

// Users should use the version subcommand
func TestFlagVersion(t *testing.T) {
	tests := []struct {
		Version        string
		Branch         string
		Commit         string
		ExpectedOutput string
	}{
		{
			Version:        "v2.0.0",
			ExpectedOutput: "Telegraf v2.0.0\n",
		},
		{
			ExpectedOutput: "Telegraf unknown\n",
		},
		{
			Version:        "v2.0.0",
			Branch:         "master",
			ExpectedOutput: "Telegraf v2.0.0 (git: master@unknown)\n",
		},
		{
			Version:        "v2.0.0",
			Branch:         "master",
			Commit:         "123",
			ExpectedOutput: "Telegraf v2.0.0 (git: master@123)\n",
		},
		{
			Version:        "v2.0.0",
			Commit:         "123",
			ExpectedOutput: "Telegraf v2.0.0 (git: unknown@123)\n",
		},
	}

	for _, test := range tests {
		buf := new(bytes.Buffer)
		args := os.Args[0:1]
		args = append(args, "--version")
		internal.Version = test.Version
		internal.Branch = test.Branch
		internal.Commit = test.Commit
		err := runApp(args, buf, NewMockServer(), NewMockConfig(buf), NewMockTelegraf())
		require.NoError(t, err)
		require.Equal(t, test.ExpectedOutput, buf.String())
	}
}

func TestGlobablBoolFlags(t *testing.T) {
	commands := []string{
		"--debug",
		"--test",
		"--quiet",
		"--once",
	}

	buf := new(bytes.Buffer)
	args := os.Args[0:1]
	args = append(args, commands...)
	m := NewMockTelegraf()
	err := runApp(args, buf, NewMockServer(), NewMockConfig(buf), m)
	require.NoError(t, err)

	require.True(t, m.debug)
	require.True(t, m.test)
	require.True(t, m.once)
	require.True(t, m.quiet)
}

func TestFlagsAreSet(t *testing.T) {
	expectedInt := 1
	expectedString := "test"

	commands := []string{
		"--config", expectedString,
		"--config-directory", expectedString,
		"--debug",
		"--test",
		"--quiet",
		"--once",
		"--test-wait", strconv.Itoa(expectedInt),
		"--watch-config", expectedString,
		"--pidfile", expectedString,
	}

	buf := new(bytes.Buffer)
	args := os.Args[0:1]
	args = append(args, commands...)
	m := NewMockTelegraf()
	err := runApp(args, buf, NewMockServer(), NewMockConfig(buf), m)
	require.NoError(t, err)

	require.Equal(t, []string{expectedString}, m.config)
	require.Equal(t, []string{expectedString}, m.configDir)
	require.True(t, m.debug)
	require.True(t, m.test)
	require.True(t, m.once)
	require.True(t, m.quiet)
	require.Equal(t, expectedInt, m.testWait)
	require.Equal(t, expectedString, m.watchConfig)
	require.Equal(t, expectedString, m.pidFile)
}
