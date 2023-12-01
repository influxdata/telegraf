//go:build linux

package dpdk

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	jsonparser "github.com/influxdata/telegraf/plugins/parsers/json"
)

const (
	maxInitMessageLength   = 1024 // based on https://github.com/DPDK/dpdk/blob/v22.07/lib/telemetry/telemetry.c#L352
	dpdkSocketTemplateName = "dpdk_telemetry"
)

type initMessage struct {
	Version      string `json:"version"`
	Pid          int    `json:"pid"`
	MaxOutputLen uint32 `json:"max_output_len"`
}

type dpdkConnector struct {
	pathToSocket  string
	accessTimeout time.Duration
	connection    net.Conn
	initMessage   *initMessage
}

func newDpdkConnector(pathToSocket string, accessTimeout config.Duration) *dpdkConnector {
	return &dpdkConnector{
		pathToSocket:  pathToSocket,
		accessTimeout: time.Duration(accessTimeout),
	}
}

// Connects to the socket
// Since DPDK is a local unix socket, it is instantly returns error or connection, so there's no need to set timeout for it
func (conn *dpdkConnector) connect() (*initMessage, error) {
	if err := isSocket(conn.pathToSocket); err != nil {
		return nil, err
	}

	connection, err := net.Dial("unixpacket", conn.pathToSocket)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the socket: %w", err)
	}
	conn.connection = connection

	conn.initMessage, err = conn.readInitMessage()
	if err != nil {
		if closeErr := conn.tryClose(); closeErr != nil {
			return nil, fmt.Errorf("%w and failed to close connection: %w", err, closeErr)
		}
		return nil, err
	}
	return conn.initMessage, nil
}

// Add metadata fields to data
func (conn *dpdkConnector) addMetadataFields(metadataFields []string, data map[string]interface{}) {
	if conn.initMessage == nil {
		return
	}

	for _, field := range metadataFields {
		switch field {
		case dpdkMetadataFieldPidName:
			data[dpdkMetadataFieldPidName] = conn.initMessage.Pid
		case dpdkMetadataFieldVersionName:
			data[dpdkMetadataFieldVersionName] = conn.initMessage.Version
		}
	}
}

// Fetches all identifiers of devices and then creates all possible combinations of commands for each device
func (conn *dpdkConnector) appendCommandsWithParamsFromList(listCommand string, commands []string) ([]string, error) {
	response, err := conn.getCommandResponse(listCommand)
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

// Executes command using provided connection and returns response
// If error (such as timeout) occurred, then connection is discarded and recreated
// because otherwise behavior of connection is undefined (e.g. it could return result of timed out command instead of latest)
func (conn *dpdkConnector) getCommandResponse(fullCommand string) ([]byte, error) {
	connection, err := conn.getConnection()
	if err != nil {
		return nil, fmt.Errorf("failed to get connection to execute %q command: %w", fullCommand, err)
	}

	err = conn.setTimeout()
	if err != nil {
		return nil, fmt.Errorf("failed to set timeout for %q command: %w", fullCommand, err)
	}

	_, err = connection.Write([]byte(fullCommand))
	if err != nil {
		if closeErr := conn.tryClose(); closeErr != nil {
			return nil, fmt.Errorf("failed to send %q command: %w and failed to close connection: %w", fullCommand, err, closeErr)
		}
		return nil, fmt.Errorf("failed to send %q command: %w", fullCommand, err)
	}

	buf := make([]byte, conn.initMessage.MaxOutputLen)
	messageLength, err := connection.Read(buf)
	if err != nil {
		if closeErr := conn.tryClose(); closeErr != nil {
			return nil, fmt.Errorf("failed read response of %q command: %w and failed to close connection: %w", fullCommand, err, closeErr)
		}
		return nil, fmt.Errorf("failed to read response of %q command: %w", fullCommand, err)
	}

	if messageLength == 0 {
		return nil, fmt.Errorf("got empty response during execution of %q command", fullCommand)
	}
	return buf[:messageLength], nil
}

// Executes command, parses response and creates/writes metrics from response to accumulator
func (conn *dpdkConnector) processCommand(acc telegraf.Accumulator, log telegraf.Logger, commandWithParams string, metadataFields []string) {
	buf, err := conn.getCommandResponse(commandWithParams)
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
		log.Warnf("got empty json on %q command", commandWithParams)
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
		log.Warnf("Failed to process a response of the command: %s. Error: %v. Continue to handle data", command, err)
	}

	// Add metadata fields if required
	conn.addMetadataFields(metadataFields, jf.Fields)

	// Add common fields
	acc.AddFields(pluginName, jf.Fields, map[string]string{
		"command": command,
		"params":  getParams(commandWithParams),
	})
}

func (conn *dpdkConnector) tryClose() error {
	if conn.connection == nil {
		return nil
	}

	err := conn.connection.Close()
	conn.connection = nil
	if err != nil {
		return err
	}
	return nil
}

func (conn *dpdkConnector) setTimeout() error {
	if conn.connection == nil {
		return errors.New("connection had not been established before")
	}

	if conn.accessTimeout == 0 {
		return conn.connection.SetDeadline(time.Time{})
	}
	return conn.connection.SetDeadline(time.Now().Add(conn.accessTimeout))
}

// Returns connections, if connection is not created then function tries to recreate it
func (conn *dpdkConnector) getConnection() (net.Conn, error) {
	if conn.connection == nil {
		_, err := conn.connect()
		if err != nil {
			return nil, err
		}
	}
	return conn.connection, nil
}

// Reads InitMessage for connection. Should be read for each connection, otherwise InitMessage is returned as response for first command.
func (conn *dpdkConnector) readInitMessage() (*initMessage, error) {
	buf := make([]byte, maxInitMessageLength)
	err := conn.setTimeout()
	if err != nil {
		return nil, fmt.Errorf("failed to set timeout: %w", err)
	}

	messageLength, err := conn.connection.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to read InitMessage: %w", err)
	}

	var connectionInitMessage initMessage
	err = json.Unmarshal(buf[:messageLength], &connectionInitMessage)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if connectionInitMessage.MaxOutputLen == 0 {
		return nil, errors.New("failed to read maxOutputLen information")
	}

	return &connectionInitMessage, nil
}
