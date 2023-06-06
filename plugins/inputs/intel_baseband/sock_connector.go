package intel_baseband

import (
	"fmt"
	"net"
	"time"

	"github.com/influxdata/telegraf/config"
)

const (
	// Command code
	clearLogCmdID   = 0x4
	deviceDataCmdID = 0x9
)

type SocketConnector struct {
	pathToSocket          string
	accessTimeout         time.Duration
	waitForTelemetryDelay time.Duration
	connection            net.Conn
}

func (sc *SocketConnector) dumpTelemetryToLog() error {
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

	//time necessary for pf-bb-config to update the data in the file
	time.Sleep(sc.waitForTelemetryDelay)
	return nil
}

func (sc *SocketConnector) sendCommandToSocket(c byte) error {
	err := sc.connect()
	if err != nil {
		return err
	}
	defer sc.close()
	err = sc.WriteCommandToSocket(c)
	if err != nil {
		return err
	}
	return nil
}

func (sc *SocketConnector) WriteCommandToSocket(c byte) error {
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

func (sc *SocketConnector) connect() error {
	connection, err := net.Dial("unix", sc.pathToSocket)
	if err != nil {
		return fmt.Errorf("failed to connect to the socket: %w", err)
	}

	sc.connection = connection
	return nil
}

func (sc *SocketConnector) close() error {
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

func newSocketConnector(pathToSocket string, accessTimeout config.Duration, waitForTelemetryDelay config.Duration) *SocketConnector {
	return &SocketConnector{
		pathToSocket:          pathToSocket,
		accessTimeout:         time.Duration(accessTimeout),
		waitForTelemetryDelay: time.Duration(waitForTelemetryDelay),
	}
}
