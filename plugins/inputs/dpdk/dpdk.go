// +build linux

package dpdk

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/influxdata/telegraf/config"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	jsonparser "github.com/influxdata/telegraf/plugins/parsers/json"
)

const (
	dpdkDescription  = "Reads metrics from DPDK applications using v2 telemetry interface."
	dpdkSampleConfig = `
  ## Path to DPDK telemetry socket. This shall point to v2 version of DPDK telemetry interface.
  # socket_path = "/var/run/dpdk/rte/dpdk_telemetry.v2"

  ## Duration that defines how long the connected socket client will wait for a response before terminating connection.
  ## This includes both writing to and reading from socket.
  ## Setting the value to 0 disables the timeout (not recommended)
  # socket_access_timeout = "200ms"

  ## Enables telemetry data collection for selected device types.
  ## Adding "ethdev" enables collection of telemetry from DPDK NICs (stats, xstats, link_status).
  ## Adding "rawdev" enables collection of telemetry from DPDK Raw Devices (xstats).
  device_types = ["ethdev", "rawdev"]

  ## List of custom, application-specific telemetry commands to query 
  ## For e.g. L3 Forwarding with Power Management Sample Application this could be: 
  ##   additional_commands = ["/l3fwd-power/stats"]
  # additional_commands = []

  ## Allows turning off collecting data for individual "ethdev" commands.
  ## Remove "/ethdev/link_status" from list to start getting link status metrics.
  [inputs.dpdk.ethdev]
    exclude_commands = ["/ethdev/link_status"]

  ## When running multiple instances of the plugin it's recommended to add a unique tag to each instance to identify
  ## metrics exposed by an instance of DPDK application. This is useful when multiple DPDK apps run on a single host.  
  ##  [inputs.dpdk.tags]
  ##    dpdk_instance = "my-fwd-app"
`
	defaultPathToSocket        = "/var/run/dpdk/rte/dpdk_telemetry.v2"
	defaultAccessTimeout       = config.Duration(200 * time.Millisecond)
	maxCommandLength           = 56
	maxCommandLengthWithParams = 1024
	dpdkPluginName             = "dpdk"
	ethdevListCommand          = "/ethdev/list"
	rawdevListCommand          = "/rawdev/list"
)

type dpdk struct {
	SocketPath         string          `toml:"socket_path"`
	AccessTimeout      config.Duration `toml:"socket_access_timeout"`
	DeviceTypes        []string        `toml:"device_types"`
	EthdevConfig       ethdevConfig    `toml:"ethdev"`
	AdditionalCommands []string        `toml:"additional_commands"`
	Log                telegraf.Logger `toml:"-"`
	connector          *dpdkConnector  `toml:"-"`
	rawdevCommands     []string        `toml:"-"`
	ethdevCommands     []string        `toml:"-"`
}

type ethdevConfig struct {
	EthdevExcludeCommands []string `toml:"exclude_commands"`
}

func init() {
	inputs.Add(dpdkPluginName, func() telegraf.Input {
		dpdk := &dpdk{
			SocketPath:     defaultPathToSocket,
			AccessTimeout:  defaultAccessTimeout,
			rawdevCommands: []string{"/rawdev/xstats"},
			ethdevCommands: []string{"/ethdev/stats", "/ethdev/xstats", "/ethdev/link_status"},
		}
		return dpdk
	})
}

func (dpdk *dpdk) SampleConfig() string {
	return dpdkSampleConfig
}

func (dpdk *dpdk) Description() string {
	return dpdkDescription
}

// Performs validation of all parameters from configuration
func (dpdk *dpdk) Init() error {
	if dpdk.SocketPath == "" {
		dpdk.SocketPath = defaultPathToSocket
		dpdk.Log.Infof("using default '%v' path for socket_path", defaultPathToSocket)
	}
	if err := isSocket(dpdk.SocketPath); err != nil {
		return err
	}

	if err := dpdk.validateCommands(); err != nil {
		return err
	}

	if dpdk.AccessTimeout < 0 {
		return fmt.Errorf("socket_access_timeout should be positive number or equal to 0 (to disable timeouts)")
	}
	if len(dpdk.AdditionalCommands) == 0 && len(dpdk.DeviceTypes) == 0 {
		dpdk.Log.Warn("DPDK plugin is enabled and configured with nothing to read")
	}

	dpdk.connector = newDpdkConnector(dpdk.SocketPath, dpdk.AccessTimeout)
	result, err := dpdk.connector.connect()
	if result != "" {
		dpdk.Log.Info(result)
	}
	return err
}

// Checks that user-supplied commands are unique and match DPDK commands format
func (dpdk *dpdk) validateCommands() error {
	dpdk.AdditionalCommands = uniqueValues(dpdk.AdditionalCommands)

	for _, commandWithParams := range dpdk.AdditionalCommands {
		if len(commandWithParams) == 0 {
			return fmt.Errorf("got empty command")
		}

		if commandWithParams[0] != '/' {
			return fmt.Errorf("'%v' command should start with '/'", commandWithParams)
		}

		if commandWithoutParams := stripParams(commandWithParams); len(commandWithoutParams) >= maxCommandLength {
			return fmt.Errorf("'%v' command is too long. It shall be less than %v characters", commandWithoutParams, maxCommandLength)
		}

		if len(commandWithParams) >= maxCommandLengthWithParams {
			return fmt.Errorf("command with parameters '%v' shall be less than %v characters", commandWithParams, maxCommandLengthWithParams)
		}
	}

	return nil
}

// Gathers all unique commands and processes each command sequentially
// Parallel processing could be achieved by running several instances of this plugin with different settings
func (dpdk *dpdk) Gather(accumulator telegraf.Accumulator) error {
	commands := dpdk.gatherCommands(accumulator)

	for _, command := range commands {
		dpdk.processCommand(command, accumulator)
	}

	return nil
}

// Gathers all unique commands
func (dpdk *dpdk) gatherCommands(accumulator telegraf.Accumulator) []string {
	var commands []string
	if contains(dpdk.DeviceTypes, "ethdev") {
		commands = append(commands, dpdk.getCommandsAndParamsCombinations(dpdk.ethdevCommands, dpdk.EthdevConfig.EthdevExcludeCommands, ethdevListCommand, accumulator)...)
	}

	if contains(dpdk.DeviceTypes, "rawdev") {
		commands = append(commands, dpdk.getCommandsAndParamsCombinations(dpdk.rawdevCommands, []string{}, rawdevListCommand, accumulator)...)
	}

	commands = append(commands, dpdk.AdditionalCommands...)

	return uniqueValues(commands)
}

func (dpdk *dpdk) getCommandsAndParamsCombinations(appendedCommands []string, excludedCommands []string, listCommand string, accumulator telegraf.Accumulator) []string {
	validCommands := removeSubset(appendedCommands, excludedCommands)

	commands, err := dpdk.appendCommandsWithParamsFromList(listCommand, validCommands)
	if err != nil {
		accumulator.AddError(fmt.Errorf("error occurred during fetching of %v params - %v", listCommand, err))
	}
	return commands
}

// Fetches all identifiers of devices and then creates all possible combinations of commands for each device
func (dpdk *dpdk) appendCommandsWithParamsFromList(listCommand string, commands []string) ([]string, error) {
	response, err := dpdk.connector.getCommandResponse(listCommand)
	if err != nil {
		return nil, err
	}

	params, err := jsonToArray(response, listCommand)
	if err != nil {
		return nil, err
	}

	result := make([]string, 0, len(commands)*len(params))
	for _, command := range commands {
		for _, param := range params {
			result = append(result, commandWithParams(command, param))
		}
	}

	return uniqueValues(result), nil
}

// Executes command, parses response and creates/writes metric from response
func (dpdk *dpdk) processCommand(commandWithParams string, accumulator telegraf.Accumulator) {
	buf, err := dpdk.connector.getCommandResponse(commandWithParams)
	if err != nil {
		accumulator.AddError(err)
		return
	}

	var parsedResponse map[string]interface{}
	err = json.Unmarshal(buf, &parsedResponse)
	if err != nil {
		accumulator.AddError(fmt.Errorf("failed to unmarshall json response from %v command - %v", commandWithParams, err))
		return
	}

	command := stripParams(commandWithParams)
	value := parsedResponse[command]
	if isEmpty(value) {
		accumulator.AddError(fmt.Errorf("got empty json on '%v' command", commandWithParams))
		return
	}

	jf := jsonparser.JSONFlattener{}
	err = jf.FullFlattenJSON("", value, true, true)
	if err != nil {
		accumulator.AddError(fmt.Errorf("failed to flatten response - %v", err))
		return
	}

	accumulator.AddFields(dpdkPluginName, jf.Fields, map[string]string{
		"command": command,
		"params":  getParams(commandWithParams),
	})
}
