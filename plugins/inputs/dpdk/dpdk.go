//go:generate ../../../tools/readme_config_includer/generator
//go:build linux

package dpdk

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/internal/choice"
	"github.com/influxdata/telegraf/internal/globpath"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

const (
	defaultPathToSocket        = "/var/run/dpdk/rte/dpdk_telemetry.v2"
	defaultAccessTimeout       = config.Duration(200 * time.Millisecond)
	maxCommandLength           = 56
	maxCommandLengthWithParams = 1024
	pluginName                 = "dpdk"
	ethdevListCommand          = "/ethdev/list"
	rawdevListCommand          = "/rawdev/list"

	dpdkMetadataFieldPidName     = "pid"
	dpdkMetadataFieldVersionName = "version"

	dpdkPluginOptionInMemory = "in_memory"

	unreachableSocketBehaviorIgnore = "ignore"
	unreachableSocketBehaviorError  = "error"
)

type dpdk struct {
	SocketPath                string          `toml:"socket_path"`
	AccessTimeout             config.Duration `toml:"socket_access_timeout"`
	DeviceTypes               []string        `toml:"device_types"`
	EthdevConfig              ethdevConfig    `toml:"ethdev"`
	AdditionalCommands        []string        `toml:"additional_commands"`
	MetadataFields            []string        `toml:"metadata_fields"`
	PluginOptions             []string        `toml:"plugin_options"`
	UnreachableSocketBehavior string          `toml:"unreachable_socket_behavior"`
	Log                       telegraf.Logger `toml:"-"`

	connectors                   []*dpdkConnector
	rawdevCommands               []string
	ethdevCommands               []string
	ethdevExcludedCommandsFilter filter.Filter
	socketGlobPath               *globpath.GlobPath
}

type ethdevConfig struct {
	EthdevExcludeCommands []string `toml:"exclude_commands"`
}

func (*dpdk) SampleConfig() string {
	return sampleConfig
}

// Init performs validation of all parameters from configuration
func (dpdk *dpdk) Init() error {
	dpdk.setupDefaultValues()

	err := dpdk.validateAdditionalCommands()
	if err != nil {
		return err
	}

	if dpdk.AccessTimeout < 0 {
		return errors.New("socket_access_timeout should be positive number or equal to 0 (to disable timeouts)")
	}

	if len(dpdk.AdditionalCommands) == 0 && len(dpdk.DeviceTypes) == 0 {
		return errors.New("plugin was configured with nothing to read")
	}

	dpdk.ethdevExcludedCommandsFilter, err = filter.Compile(dpdk.EthdevConfig.EthdevExcludeCommands)
	if err != nil {
		return fmt.Errorf("error occurred during filter preparation for ethdev excluded commands: %w", err)
	}

	if err = choice.Check(dpdk.UnreachableSocketBehavior, []string{unreachableSocketBehaviorError, unreachableSocketBehaviorIgnore}); err != nil {
		return fmt.Errorf("unreachable_socket_behavior: %w", err)
	}

	glob, err := globpath.Compile(dpdk.SocketPath + "*")
	if err != nil {
		return err
	}
	dpdk.socketGlobPath = glob
	return nil
}

// Start implements ServiceInput interface
func (dpdk *dpdk) Start(telegraf.Accumulator) error {
	return dpdk.maintainConnections()
}

func (dpdk *dpdk) Stop() {
	for _, connector := range dpdk.connectors {
		if err := connector.tryClose(); err != nil {
			dpdk.Log.Warnf("Couldn't close connection for %q: %v", connector.pathToSocket, err)
		}
	}
	dpdk.connectors = nil
}

// Gather function gathers all unique commands and processes each command sequentially
// Parallel processing could be achieved by running several instances of this plugin with different settings
func (dpdk *dpdk) Gather(acc telegraf.Accumulator) error {
	if err := dpdk.Start(acc); err != nil {
		return err
	}

	for _, dpdkConn := range dpdk.connectors {
		commands := dpdk.gatherCommands(acc, dpdkConn)
		for _, command := range commands {
			dpdkConn.processCommand(acc, dpdk.Log, command, dpdk.MetadataFields)
		}
	}
	return nil
}

// Setup default values for dpdk
func (dpdk *dpdk) setupDefaultValues() {
	if dpdk.SocketPath == "" {
		dpdk.SocketPath = defaultPathToSocket
	}

	if dpdk.DeviceTypes == nil {
		dpdk.DeviceTypes = []string{"ethdev"}
	}

	if dpdk.MetadataFields == nil {
		dpdk.MetadataFields = []string{dpdkMetadataFieldPidName, dpdkMetadataFieldVersionName}
	}

	if dpdk.PluginOptions == nil {
		dpdk.PluginOptions = []string{dpdkPluginOptionInMemory}
	}

	if len(dpdk.UnreachableSocketBehavior) == 0 {
		dpdk.UnreachableSocketBehavior = unreachableSocketBehaviorError
	}

	dpdk.rawdevCommands = []string{"/rawdev/xstats"}
	dpdk.ethdevCommands = []string{"/ethdev/stats", "/ethdev/xstats", "/ethdev/info", ethdevLinkStatusCommand}
}

func (dpdk *dpdk) getDpdkInMemorySocketPaths() []string {
	filePaths := dpdk.socketGlobPath.Match()

	var results []string
	for _, filePath := range filePaths {
		fileInfo, err := os.Stat(filePath)
		if err != nil || fileInfo.IsDir() || !strings.Contains(filePath, dpdkSocketTemplateName) {
			continue
		}

		if isInMemorySocketPath(filePath, dpdk.SocketPath) {
			results = append(results, filePath)
		}
	}

	return results
}

// Checks that user-supplied commands are unique and match DPDK commands format
func (dpdk *dpdk) validateAdditionalCommands() error {
	dpdk.AdditionalCommands = uniqueValues(dpdk.AdditionalCommands)

	for _, cmd := range dpdk.AdditionalCommands {
		if len(cmd) == 0 {
			return errors.New("got empty command")
		}

		if cmd[0] != '/' {
			return fmt.Errorf("%q command should start with slash", cmd)
		}

		if commandWithoutParams := stripParams(cmd); len(commandWithoutParams) >= maxCommandLength {
			return fmt.Errorf("%q command is too long. It shall be less than %v characters", commandWithoutParams, maxCommandLength)
		}

		if len(cmd) >= maxCommandLengthWithParams {
			return fmt.Errorf("command with parameters %q shall be less than %v characters", cmd, maxCommandLengthWithParams)
		}
	}

	return nil
}

// Establishes connections do DPDK telemetry sockets
func (dpdk *dpdk) maintainConnections() error {
	candidates := []string{dpdk.SocketPath}
	if choice.Contains(dpdkPluginOptionInMemory, dpdk.PluginOptions) {
		candidates = dpdk.getDpdkInMemorySocketPaths()
	}

	// Find sockets in the connected-sockets list that are not among
	// the candidates anymore and thus need to be removed.
	for i := 0; i < len(dpdk.connectors); i++ {
		connector := dpdk.connectors[i]
		if !choice.Contains(connector.pathToSocket, candidates) {
			dpdk.Log.Debugf("Close unused connection: %s", connector.pathToSocket)
			if closeErr := connector.tryClose(); closeErr != nil {
				dpdk.Log.Warnf("Failed to close unused connection: %v", closeErr)
			}
			dpdk.connectors = append(dpdk.connectors[:i], dpdk.connectors[i+1:]...)
			i--
		}
	}

	// Find candidates that are not yet in the connected-sockets list as we
	// need to connect to those.
	for _, candidate := range candidates {
		var found bool
		for _, connector := range dpdk.connectors {
			if candidate == connector.pathToSocket {
				found = true
				break
			}
		}
		if !found {
			connector := newDpdkConnector(candidate, dpdk.AccessTimeout)
			connectionInitMessage, err := connector.connect()
			if err != nil {
				if dpdk.UnreachableSocketBehavior == unreachableSocketBehaviorError {
					return fmt.Errorf("couldn't connect to socket %s: %w", candidate, err)
				}
				dpdk.Log.Warnf("Couldn't connect to socket %s: %v", candidate, err)
				continue
			}

			dpdk.Log.Debugf("Successfully connected to the socket: %s. Version: %v running as process with PID %v with len %v",
				candidate, connectionInitMessage.Version, connectionInitMessage.Pid, connectionInitMessage.MaxOutputLen)
			dpdk.connectors = append(dpdk.connectors, connector)
		}
	}

	if len(dpdk.connectors) == 0 {
		errMsg := "no active sockets connections present"
		if dpdk.UnreachableSocketBehavior == unreachableSocketBehaviorError {
			return errors.New(errMsg)
		}
		dpdk.Log.Warnf("Unreachable socket issue occurred: %v", errMsg)
	}

	return nil
}

// Gathers all unique commands
func (dpdk *dpdk) gatherCommands(acc telegraf.Accumulator, dpdkConnector *dpdkConnector) []string {
	var commands []string
	if choice.Contains("ethdev", dpdk.DeviceTypes) {
		ethdevCommands := removeSubset(dpdk.ethdevCommands, dpdk.ethdevExcludedCommandsFilter)
		ethdevCommands, err := dpdkConnector.appendCommandsWithParamsFromList(ethdevListCommand, ethdevCommands)
		if err != nil {
			acc.AddError(fmt.Errorf("error occurred during fetching of %q params: %w", ethdevListCommand, err))
		}
		commands = append(commands, ethdevCommands...)
	}

	if choice.Contains("rawdev", dpdk.DeviceTypes) {
		rawdevCommands, err := dpdkConnector.appendCommandsWithParamsFromList(rawdevListCommand, dpdk.rawdevCommands)
		if err != nil {
			acc.AddError(fmt.Errorf("error occurred during fetching of %q params: %w", rawdevListCommand, err))
		}
		commands = append(commands, rawdevCommands...)
	}

	commands = append(commands, dpdk.AdditionalCommands...)
	return uniqueValues(commands)
}

func init() {
	inputs.Add(pluginName, func() telegraf.Input {
		dpdk := &dpdk{
			// Setting it here (rather than in `Init()`) to distinguish between "zero" value,
			// default value and don't having value in config at all.
			AccessTimeout: defaultAccessTimeout,
		}
		return dpdk
	})
}
