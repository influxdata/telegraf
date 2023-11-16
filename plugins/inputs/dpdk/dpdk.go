//go:generate ../../../tools/readme_config_includer/generator
//go:build linux

package dpdk

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/internal/choice"
	"github.com/influxdata/telegraf/internal/globpath"
	"github.com/influxdata/telegraf/plugins/inputs"
	jsonparser "github.com/influxdata/telegraf/plugins/parsers/json"
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

	dpdkMetadataFieldPidName     = "dpdk_pid"
	dpdkMetadataFieldVersionName = "dpdk_version"

	dpdkPluginOptionInMemory = "dpdk_in_memory"

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
		return fmt.Errorf("socket_access_timeout should be positive number or equal to 0 (to disable timeouts)")
	}

	if len(dpdk.AdditionalCommands) == 0 && len(dpdk.DeviceTypes) == 0 {
		return fmt.Errorf("plugin was configured with nothing to read")
	}

	dpdk.ethdevExcludedCommandsFilter, err = filter.Compile(dpdk.EthdevConfig.EthdevExcludeCommands)
	if err != nil {
		return fmt.Errorf("error occurred during filter preparation for ethdev excluded commands - %w", err)
	}

	if err = choice.Check(dpdk.UnreachableSocketBehavior, []string{unreachableSocketBehaviorError, unreachableSocketBehaviorIgnore}); err != nil {
		return fmt.Errorf("unreachable_socket_behavior: %w", err)
	}

	glob, err := prepareGlob(dpdk.SocketPath)
	if err != nil {
		return err
	}
	dpdk.socketGlobPath = glob

	if err = dpdk.maintainConnections(); err != nil {
		if dpdk.UnreachableSocketBehavior == unreachableSocketBehaviorError {
			return err
		}
		dpdk.Log.Warnf("Unreachable socket issue occurred: %v", err)
	}

	return nil
}

// Start is empty to implement ServiceInput interface
func (dpdk *dpdk) Start(telegraf.Accumulator) error {
	return nil
}

// Gather function gathers all unique commands and processes each command sequentially
// Parallel processing could be achieved by running several instances of this plugin with different settings
func (dpdk *dpdk) Gather(acc telegraf.Accumulator) error {
	if err := dpdk.maintainConnections(); err != nil {
		if dpdk.UnreachableSocketBehavior == unreachableSocketBehaviorError {
			return err
		}
		dpdk.Log.Warnf("Unreachable socket issue occurred: %v", err)
		return nil
	}

	for _, dpdkConn := range dpdk.connectors {
		commands := dpdk.gatherCommands(dpdkConn, acc)
		for _, command := range commands {
			dpdk.processCommand(dpdkConn, acc, command)
		}
	}
	return nil
}

func (dpdk *dpdk) Stop() {
	var err error
	for _, connector := range dpdk.connectors {
		if err = connector.tryClose(); err != nil {
			dpdk.Log.Warnf("Couldn't close connection for: %s. Err: %v", connector.pathToSocket, err)
		}
	}
	dpdk.connectors = nil
}

// Setup default values for dpdk
func (dpdk *dpdk) setupDefaultValues() {
	if dpdk.SocketPath == "" {
		dpdk.SocketPath = defaultPathToSocket
		dpdk.Log.Debugf("Using default %q value for socket_path", defaultPathToSocket)
	}

	if dpdk.DeviceTypes == nil {
		dpdk.DeviceTypes = []string{"ethdev"}
		dpdk.Log.Debugf("Using default %q value for device_types", dpdk.DeviceTypes)
	}

	if dpdk.MetadataFields == nil {
		dpdk.MetadataFields = []string{dpdkMetadataFieldPidName}
		dpdk.Log.Debugf("Using default %q value for metadata_fields", dpdk.MetadataFields)
	}

	if dpdk.PluginOptions == nil {
		dpdk.PluginOptions = []string{dpdkPluginOptionInMemory}
		dpdk.Log.Debugf("Using default %q value for plugin_options", dpdk.PluginOptions)
	}

	if len(dpdk.UnreachableSocketBehavior) == 0 {
		dpdk.UnreachableSocketBehavior = unreachableSocketBehaviorError
		dpdk.Log.Debugf("Using default %q value for unreachable_socket_behavior", unreachableSocketBehaviorError)
	}

	dpdk.rawdevCommands = []string{"/rawdev/xstats"}
	dpdk.ethdevCommands = []string{"/ethdev/stats", "/ethdev/xstats", "/ethdev/info", ethdevLinkStatusCommand}
}

// Checks that user-supplied commands are unique and match DPDK commands format
func (dpdk *dpdk) validateAdditionalCommands() error {
	dpdk.AdditionalCommands = uniqueValues(dpdk.AdditionalCommands)

	for _, fullCommandWithParams := range dpdk.AdditionalCommands {
		if len(fullCommandWithParams) == 0 {
			return fmt.Errorf("got empty command")
		}

		if fullCommandWithParams[0] != '/' {
			return fmt.Errorf("%q command should start with '/'", fullCommandWithParams)
		}

		if commandWithoutParams := stripParams(fullCommandWithParams); len(commandWithoutParams) >= maxCommandLength {
			return fmt.Errorf("%q command is too long. It shall be less than %v characters", commandWithoutParams, maxCommandLength)
		}

		if len(fullCommandWithParams) >= maxCommandLengthWithParams {
			return fmt.Errorf("command with parameters %q shall be less than %v characters", fullCommandWithParams, maxCommandLengthWithParams)
		}
	}

	return nil
}

// Establishes connections do DPDK telemetry sockets
func (dpdk *dpdk) maintainConnections() error {
	socketsToConnect, socketsToDisconnect := dpdk.identifySockets()

	// Try to close connection with unused sockets
	// And delete unused elements from the dpdk.connectors list with the safe approach
	for _, socketToDisconnect := range socketsToDisconnect {
		for i := 0; i < len(dpdk.connectors); i++ {
			connector := dpdk.connectors[i]
			if socketToDisconnect == connector.pathToSocket {
				dpdk.Log.Debugf("Close unused connection: %s", socketToDisconnect)
				if closeErr := connector.tryClose(); closeErr != nil {
					dpdk.Log.Warnf("Failed to close unused connection - %v", closeErr)
				}
				dpdk.connectors = append(dpdk.connectors[:i], dpdk.connectors[i+1:]...)
				i--
				break
			}
		}
	}

	// Create connections
	for _, socketToConnect := range socketsToConnect {
		connector := newDpdkConnector(socketToConnect, dpdk.AccessTimeout)
		connectionInitMessage, err := connector.connect()
		if err != nil {
			if dpdk.UnreachableSocketBehavior == unreachableSocketBehaviorError {
				return fmt.Errorf("couldn't connect to socket %s: %w", socketToConnect, err)
			}
			dpdk.Log.Warnf("Couldn't connect to socket %s: %v", socketToConnect, err)
			continue
		}

		dpdk.Log.Debugf("Successfully connected to the socket: %s. Version: %v running as process with PID %v with len %v",
			socketToConnect, connectionInitMessage.Version, connectionInitMessage.Pid, connectionInitMessage.MaxOutputLen)
		dpdk.connectors = append(dpdk.connectors, connector)
	}

	if len(dpdk.connectors) == 0 {
		return fmt.Errorf("no active sockets connections present")
	}

	return nil
}

func (dpdk *dpdk) identifySockets() (socketsToConnect []string, socketsToDisconnect []string) {
	pathsToExistingSockets := []string{dpdk.SocketPath}
	if choice.Contains(dpdkPluginOptionInMemory, dpdk.PluginOptions) {
		pathsToExistingSockets = dpdk.getDpdkInMemorySocketPaths()
	}

	pathsToConnectedSockets := make([]string, 0, len(dpdk.connectors))
	for _, connector := range dpdk.connectors {
		pathsToConnectedSockets = append(pathsToConnectedSockets, connector.pathToSocket)
	}

	return getDiffArrays(pathsToConnectedSockets, pathsToExistingSockets)
}

// Gathers all unique commands
func (dpdk *dpdk) gatherCommands(dpdkConnector *dpdkConnector, acc telegraf.Accumulator) []string {
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

// Executes command, parses response and creates/writes metric from response
func (dpdk *dpdk) processCommand(dpdkConn *dpdkConnector, acc telegraf.Accumulator, commandWithParams string) {
	buf, err := dpdkConn.getCommandResponse(commandWithParams)
	if err != nil {
		acc.AddError(err)
		return
	}

	var parsedResponse map[string]interface{}
	err = json.Unmarshal(buf, &parsedResponse)
	if err != nil {
		acc.AddError(fmt.Errorf("failed to unmarshal json response from %q command: %w", commandWithParams, err))
		return
	}

	command := stripParams(commandWithParams)
	value := parsedResponse[command]
	if isEmpty(value) {
		dpdk.Log.Warnf("got empty json on %q command", commandWithParams)
		return
	}

	jf := jsonparser.JSONFlattener{}
	err = jf.FullFlattenJSON("", value, true, true)
	if err != nil {
		acc.AddError(fmt.Errorf("failed to flatten response: %w", err))
		return
	}

	err = processCommandResponse(command, jf.Fields)
	if err != nil {
		dpdk.Log.Warnf("Failed to process a response of the command: %s. Error: %v. Continue to handle data", command, err)
	}

	// Add metadata fields if required
	dpdk.addMetadataFields(dpdkConn, jf.Fields)

	// Add common fields
	acc.AddFields(pluginName, jf.Fields, map[string]string{
		"command": command,
		"params":  getParams(commandWithParams),
	})
}

func (dpdk *dpdk) addMetadataFields(dpdkConn *dpdkConnector, data map[string]interface{}) {
	if len(dpdk.MetadataFields) == 0 || dpdkConn.initMessage == nil {
		return
	}

	if choice.Contains(dpdkMetadataFieldPidName, dpdk.MetadataFields) {
		data[dpdkMetadataFieldPidName] = dpdkConn.initMessage.Pid
	}

	if choice.Contains(dpdkMetadataFieldVersionName, dpdk.MetadataFields) {
		data[dpdkMetadataFieldVersionName] = dpdkConn.initMessage.Version
	}
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
