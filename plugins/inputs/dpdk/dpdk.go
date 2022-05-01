//go:build linux
// +build linux

package dpdk

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/internal/choice"
	"github.com/influxdata/telegraf/plugins/inputs"
	jsonparser "github.com/influxdata/telegraf/plugins/parsers/json"
)

const (
	defaultPathToSocket        = "/var/run/dpdk/rte/dpdk_telemetry.v2"
	defaultAccessTimeout       = config.Duration(200 * time.Millisecond)
	maxCommandLength           = 56
	maxCommandLengthWithParams = 1024
	pluginName                 = "dpdk"
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

	connector                    *dpdkConnector
	rawdevCommands               []string
	ethdevCommands               []string
	ethdevExcludedCommandsFilter filter.Filter
}

type ethdevConfig struct {
	EthdevExcludeCommands []string `toml:"exclude_commands"`
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

// Performs validation of all parameters from configuration
func (dpdk *dpdk) Init() error {
	if dpdk.SocketPath == "" {
		dpdk.SocketPath = defaultPathToSocket
		dpdk.Log.Debugf("using default '%v' path for socket_path", defaultPathToSocket)
	}

	if dpdk.DeviceTypes == nil {
		dpdk.DeviceTypes = []string{"ethdev"}
	}

	var err error
	if err = isSocket(dpdk.SocketPath); err != nil {
		return err
	}

	dpdk.rawdevCommands = []string{"/rawdev/xstats"}
	dpdk.ethdevCommands = []string{"/ethdev/stats", "/ethdev/xstats", "/ethdev/link_status"}

	if err = dpdk.validateCommands(); err != nil {
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
		return fmt.Errorf("error occurred during filter prepation for ethdev excluded commands - %v", err)
	}

	dpdk.connector = newDpdkConnector(dpdk.SocketPath, dpdk.AccessTimeout)
	initMessage, err := dpdk.connector.connect()
	if initMessage != nil {
		dpdk.Log.Debugf("Successfully connected to %v running as process with PID %v with len %v",
			initMessage.Version, initMessage.Pid, initMessage.MaxOutputLen)
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
func (dpdk *dpdk) Gather(acc telegraf.Accumulator) error {
	// This needs to be done during every `Gather(...)`, because DPDK can be restarted between consecutive
	// `Gather(...)` cycles which can cause that it will be exposing different set of metrics.
	commands := dpdk.gatherCommands(acc)

	for _, command := range commands {
		dpdk.processCommand(acc, command)
	}

	return nil
}

// Gathers all unique commands
func (dpdk *dpdk) gatherCommands(acc telegraf.Accumulator) []string {
	var commands []string
	if choice.Contains("ethdev", dpdk.DeviceTypes) {
		ethdevCommands := removeSubset(dpdk.ethdevCommands, dpdk.ethdevExcludedCommandsFilter)
		ethdevCommands, err := dpdk.appendCommandsWithParamsFromList(ethdevListCommand, ethdevCommands)
		if err != nil {
			acc.AddError(fmt.Errorf("error occurred during fetching of %v params - %v", ethdevListCommand, err))
		}

		commands = append(commands, ethdevCommands...)
	}

	if choice.Contains("rawdev", dpdk.DeviceTypes) {
		rawdevCommands, err := dpdk.appendCommandsWithParamsFromList(rawdevListCommand, dpdk.rawdevCommands)
		if err != nil {
			acc.AddError(fmt.Errorf("error occurred during fetching of %v params - %v", rawdevListCommand, err))
		}

		commands = append(commands, rawdevCommands...)
	}

	commands = append(commands, dpdk.AdditionalCommands...)
	return uniqueValues(commands)
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

	return result, nil
}

// Executes command, parses response and creates/writes metric from response
func (dpdk *dpdk) processCommand(acc telegraf.Accumulator, commandWithParams string) {
	buf, err := dpdk.connector.getCommandResponse(commandWithParams)
	if err != nil {
		acc.AddError(err)
		return
	}

	var parsedResponse map[string]interface{}
	err = json.Unmarshal(buf, &parsedResponse)
	if err != nil {
		acc.AddError(fmt.Errorf("failed to unmarshall json response from %v command - %v", commandWithParams, err))
		return
	}

	command := stripParams(commandWithParams)
	value := parsedResponse[command]
	if isEmpty(value) {
		acc.AddError(fmt.Errorf("got empty json on '%v' command", commandWithParams))
		return
	}

	jf := jsonparser.JSONFlattener{}
	err = jf.FullFlattenJSON("", value, true, true)
	if err != nil {
		acc.AddError(fmt.Errorf("failed to flatten response - %v", err))
		return
	}

	acc.AddFields(pluginName, jf.Fields, map[string]string{
		"command": command,
		"params":  getParams(commandWithParams),
	})
}
