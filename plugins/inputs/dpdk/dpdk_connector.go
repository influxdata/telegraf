//go:build linux
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
func (conn *dpdkConnector) connect() (*initMessage, error) {
	connection, err := net.Dial("unixpacket", conn.pathToSocket)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the socket - %v", err)
	}

	conn.connection = connection
	result, err := conn.readMaxOutputLen()
	if err != nil {
		if closeErr := conn.tryClose(); closeErr != nil {
			return nil, fmt.Errorf("%v and failed to close connection - %v", err, closeErr)
		}
		return nil, err
	}

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

	err = conn.setTimeout()
	if err != nil {
		return nil, fmt.Errorf("failed to set timeout for %v command - %v", fullCommand, err)
	}

	_, err = connection.Write([]byte(fullCommand))
	if err != nil {
		if closeErr := conn.tryClose(); closeErr != nil {
			return nil, fmt.Errorf("failed to send '%v' command - %v and failed to close connection - %v",
				fullCommand, err, closeErr)
		}
		return nil, fmt.Errorf("failed to send '%v' command - %v", fullCommand, err)
	}

	buf := make([]byte, conn.maxOutputLen)
	messageLength, err := connection.Read(buf)
	if err != nil {
		if closeErr := conn.tryClose(); closeErr != nil {
			return nil, fmt.Errorf("failed read response of '%v' command - %v and failed to close connection - %v",
				fullCommand, err, closeErr)
		}
		return nil, fmt.Errorf("failed to read response of '%v' command - %v", fullCommand, err)
	}

	if messageLength == 0 {
		return nil, fmt.Errorf("got empty response during execution of '%v' command", fullCommand)
	}
	return buf[:messageLength], nil
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
		return fmt.Errorf("connection had not been established before")
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
func (conn *dpdkConnector) readMaxOutputLen() (*initMessage, error) {
	buf := make([]byte, maxInitMessageLength)
	err := conn.setTimeout()
	if err != nil {
		return nil, fmt.Errorf("failed to set timeout - %v", err)
	}

	messageLength, err := conn.connection.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to read InitMessage - %v", err)
	}

	var initMessage initMessage
	err = json.Unmarshal(buf[:messageLength], &initMessage)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response - %v", err)
	}

	if initMessage.MaxOutputLen == 0 {
		return nil, fmt.Errorf("failed to read maxOutputLen information")
	}

	if !conn.messageShowed {
		conn.maxOutputLen = initMessage.MaxOutputLen
		conn.messageShowed = true
		return &initMessage, nil
	}

	return nil, nil
}
