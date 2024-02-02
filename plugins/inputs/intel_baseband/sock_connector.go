//go:build linux && amd64

package intel_baseband

import (
	"fmt"
	"net"
	"time"
)

const (
	// Command code
	clearLogCmdID   = 0x4
	deviceDataCmdID = 0x9
)

type socketConnector struct {
	pathToSocket  string
	accessTimeout time.Duration
	connection    net.Conn
}

func (sc *socketConnector) dumpTelemetryToLog() error {
	// clean the log to have only the latest metrics in the file
	err := sc.sendCommandToSocket(clearLogCmdID)
	if err != nil {
		return fmt.Errorf("failed to send clear log command: %w", err)
	}

	// fill the file with the latest metrics
	err = sc.sendCommandToSocket(deviceDataCmdID)
	if err != nil {
		return fmt.Errorf("failed to send device data command: %w", err)
	}
	return nil
}

func (sc *socketConnector) sendCommandToSocket(c byte) error {
	err := sc.connect()
	if err != nil {
		return err
	}
	defer sc.close()
	err = sc.writeCommandToSocket(c)
	if err != nil {
		return err
	}
	return nil
}

func (sc *socketConnector) writeCommandToSocket(c byte) error {
	if sc.connection == nil {
		return fmt.Errorf("connection had not been established before")
	}
	var err error
	if sc.accessTimeout == 0 {
		err = sc.connection.SetWriteDeadline(time.Time{})
	} else {
		err = sc.connection.SetWriteDeadline(time.Now().Add(sc.accessTimeout))
	}

	if err != nil {
		return fmt.Errorf("failed to set timeout for request: %w", err)
	}

	_, err = sc.connection.Write([]byte{c, 0x00})
	if err != nil {
		return fmt.Errorf("failed to send request to socket: %w", err)
	}
	return nil
}

func (sc *socketConnector) connect() error {
	connection, err := net.Dial("unix", sc.pathToSocket)
	if err != nil {
		return fmt.Errorf("failed to connect to the socket: %w", err)
	}

	sc.connection = connection
	return nil
}

func (sc *socketConnector) close() error {
	if sc.connection == nil {
		return nil
	}

	err := sc.connection.Close()
	sc.connection = nil
	if err != nil {
		return err
	}
	return nil
}

func newSocketConnector(pathToSocket string, accessTimeout time.Duration) *socketConnector {
	return &socketConnector{
		pathToSocket:  pathToSocket,
		accessTimeout: accessTimeout,
	}
}
