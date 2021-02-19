// +build linux

package dpdk

import (
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/influxdata/telegraf/config"
)

const maxInitMessageLength = 1024

type initMessage struct {
	Version      string `json:"version"`
	Pid          int    `json:"pid"`
	MaxOutputLen uint32 `json:"max_output_len"`
}

type dpdkConnector struct {
	pathToSocket  string
	maxOutputLen  uint32
	messageShowed bool
	accessTimeout time.Duration
	connection    net.Conn
}

func newDpdkConnector(pathToSocket string, accessTimeout config.Duration) *dpdkConnector {
	return &dpdkConnector{
		pathToSocket:  pathToSocket,
		messageShowed: false,
		accessTimeout: time.Duration(accessTimeout),
	}
}

// Connects to the socket
// Since DPDK is a local unix socket, it is instantly returns error or connection, so there's no need to set timeout for it
func (conn *dpdkConnector) connect() (string, error) {
	connection, err := net.Dial("unixpacket", conn.pathToSocket)

	if err != nil {
		return "", fmt.Errorf("failed to connect to the socket - %v", err)
	}

	result, err := conn.readMaxOutputLen(connection)
	if err != nil {
		return result, err
	}

	conn.connection = connection
	return result, nil
}

// Executes command using provided connection and returns response
// If error (such as timeout) occurred, then connection is discarded and recreated
// because otherwise behaviour of connection is undefined (e.g. it could return result of timed out command instead of latest)
func (conn *dpdkConnector) getCommandResponse(fullCommand string) ([]byte, error) {
	connection, err := conn.getConnection()
	if err != nil {
		return nil, fmt.Errorf("failed to get connection to execute %v command - %v", fullCommand, err)
	}

	err = conn.setTimeout(connection)
	if err != nil {
		return nil, fmt.Errorf("failed to set timeout for %v command - %v", fullCommand, err)
	}

	_, err = connection.Write([]byte(fullCommand))
	if err != nil {
		return nil, conn.tryClose(fmt.Sprintf("send '%v'", fullCommand), err)
	}

	buf := make([]byte, conn.maxOutputLen)
	messageLength, err := connection.Read(buf)
	if err != nil {
		return nil, conn.tryClose(fmt.Sprintf("read response of '%v'", fullCommand), err)
	}

	if messageLength == 0 {
		return nil, fmt.Errorf("got empty response during execution of '%v' command", fullCommand)
	}
	return buf[:messageLength], nil
}

func (conn *dpdkConnector) tryClose(operation string, err error) error {
	closeErr := conn.connection.Close()
	conn.connection = nil
	newErr := fmt.Sprintf("failed to %v command - %v", operation, err)
	if closeErr != nil {
		return fmt.Errorf("%v and failed to close connection - %v", newErr, closeErr)
	}
	return fmt.Errorf(newErr)
}

func (conn *dpdkConnector) setTimeout(connection net.Conn) error {
	if conn.accessTimeout == 0 {
		return connection.SetDeadline(time.Time{})
	}
	return connection.SetDeadline(time.Now().Add(conn.accessTimeout))
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
func (conn *dpdkConnector) readMaxOutputLen(connection net.Conn) (string, error) {
	buf := make([]byte, maxInitMessageLength)
	err := conn.setTimeout(connection)
	if err != nil {
		return "", fmt.Errorf("failed to set timeout - %v", err)
	}

	messageLength, err := connection.Read(buf)
	if err != nil {
		return "", fmt.Errorf("failed to read InitMessage - %v", err)
	}

	var initMessage initMessage
	err = json.Unmarshal(buf[:messageLength], &initMessage)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal response - %v", err)
	}

	if initMessage.MaxOutputLen == 0 {
		return "", fmt.Errorf("failed to read maxOutputLen information")
	}

	if !conn.messageShowed {
		conn.maxOutputLen = initMessage.MaxOutputLen
		conn.messageShowed = true
		return fmt.Sprintf("Successfully connected to %v running as process with PID %v with len %v", initMessage.Version, initMessage.Pid, initMessage.MaxOutputLen), nil
	}

	return "", nil
}
