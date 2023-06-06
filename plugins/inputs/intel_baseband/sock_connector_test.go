package intel_baseband

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/inputs/intel_baseband/mock"
)

func TestWriteCommandToSocket(t *testing.T) {
	t.Run("correct execution of the function", func(t *testing.T) {
		conn := &mocks.Conn{}
		conn.On("Write", mock.Anything).Return(2, nil)
		conn.On("SetWriteDeadline", mock.Anything).Return(nil)
		connector := SocketConnector{connection: conn}

		err := connector.WriteCommandToSocket(0x00)
		require.NoError(t, err)

		defer conn.AssertExpectations(t)
	})

	t.Run("without setting up a connection it should return an error", func(t *testing.T) {
		connector := SocketConnector{}

		err := connector.WriteCommandToSocket(0x00)
		require.Error(t, err)
		require.ErrorContains(t, err, "connection had not been established before")
	})

	t.Run("handling timeout setting error", func(t *testing.T) {
		conn := &mocks.Conn{}
		conn.On("SetWriteDeadline", mock.Anything).Return(fmt.Errorf("deadline set error"))
		connector := SocketConnector{connection: conn}

		err := connector.WriteCommandToSocket(0x00)

		require.Error(t, err)
		require.ErrorContains(t, err, "failed to set timeout for request")
		require.ErrorContains(t, err, "deadline set error")
		defer conn.AssertExpectations(t)
	})

	t.Run("handling net.Write error", func(t *testing.T) {
		var unsupportedCommand byte = 0x99
		conn := &mocks.Conn{}
		conn.On("Write", []byte{unsupportedCommand, 0x00}).Return(0, fmt.Errorf("unsupported command"))
		conn.On("SetWriteDeadline", mock.Anything).Return(nil)
		connector := SocketConnector{connection: conn}

		err := connector.WriteCommandToSocket(unsupportedCommand)

		require.Error(t, err)
		require.ErrorContains(t, err, "failed to send request to socket")
		require.ErrorContains(t, err, "unsupported command")
		defer conn.AssertExpectations(t)
	})
}

func TestDumpTelemetryToLog(t *testing.T) {
	t.Run("with correct temporary socket should return only an error related to the inability to refresh telemetry", func(t *testing.T) {
		tempSocket := newTempSocket(t)
		defer tempSocket.Close()
		tempLogFile := newTempLogFile(t)
		defer tempLogFile.Close()
		connector := newSocketConnector(tempSocket.pathToSocket, config.Duration(5*time.Second), config.Duration(1*time.Second))

		err := connector.dumpTelemetryToLog()
		require.NoError(t, err)
	})
}
