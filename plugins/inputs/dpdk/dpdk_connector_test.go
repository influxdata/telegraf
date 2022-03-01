//go:build linux
// +build linux

package dpdk

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/plugins/inputs/dpdk/mocks"
)

func Test_readMaxOutputLen(t *testing.T) {
	t.Run("should return error if timeout occurred", func(t *testing.T) {
		conn := &mocks.Conn{}
		conn.On("Read", mock.Anything).Return(0, fmt.Errorf("timeout"))
		conn.On("SetDeadline", mock.Anything).Return(nil)
		connector := dpdkConnector{connection: conn}

		_, err := connector.readMaxOutputLen()

		require.Error(t, err)
		require.Contains(t, err.Error(), "timeout")
	})

	t.Run("should pass and set maxOutputLen if provided with valid InitMessage", func(t *testing.T) {
		maxOutputLen := uint32(4567)
		initMessage := initMessage{
			Version:      "DPDK test version",
			Pid:          1234,
			MaxOutputLen: maxOutputLen,
		}
		message, err := json.Marshal(initMessage)
		require.NoError(t, err)
		conn := &mocks.Conn{}
		conn.On("Read", mock.Anything).Run(func(arg mock.Arguments) {
			elem := arg.Get(0).([]byte)
			copy(elem, message)
		}).Return(len(message), nil)
		conn.On("SetDeadline", mock.Anything).Return(nil)
		connector := dpdkConnector{connection: conn}

		_, err = connector.readMaxOutputLen()

		require.NoError(t, err)
		require.Equal(t, maxOutputLen, connector.maxOutputLen)
	})

	t.Run("should fail if received invalid json", func(t *testing.T) {
		message := `{notAJson}`
		conn := &mocks.Conn{}
		conn.On("Read", mock.Anything).Run(func(arg mock.Arguments) {
			elem := arg.Get(0).([]byte)
			copy(elem, message)
		}).Return(len(message), nil)
		conn.On("SetDeadline", mock.Anything).Return(nil)
		connector := dpdkConnector{connection: conn}

		_, err := connector.readMaxOutputLen()

		require.Error(t, err)
		require.Contains(t, err.Error(), "looking for beginning of object key string")
	})

	t.Run("should fail if received maxOutputLen equals to 0", func(t *testing.T) {
		message, err := json.Marshal(initMessage{
			Version:      "test",
			Pid:          1,
			MaxOutputLen: 0,
		})
		require.NoError(t, err)
		conn := &mocks.Conn{}
		conn.On("Read", mock.Anything).Run(func(arg mock.Arguments) {
			elem := arg.Get(0).([]byte)
			copy(elem, message)
		}).Return(len(message), nil)
		conn.On("SetDeadline", mock.Anything).Return(nil)
		connector := dpdkConnector{connection: conn}

		_, err = connector.readMaxOutputLen()

		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to read maxOutputLen information")
	})
}

func Test_connect(t *testing.T) {
	t.Run("should pass if PathToSocket points to socket", func(t *testing.T) {
		pathToSocket, socket := createSocketForTest(t)
		defer socket.Close()
		dpdk := dpdk{
			SocketPath: pathToSocket,
			connector:  newDpdkConnector(pathToSocket, 0),
		}
		go simulateSocketResponse(socket, t)

		_, err := dpdk.connector.connect()

		require.NoError(t, err)
	})
}

func Test_getCommandResponse(t *testing.T) {
	command := "/"
	response := "myResponseString"

	t.Run("should return proper buffer size and value if no error occurred", func(t *testing.T) {
		mockConn, dpdk, _ := prepareEnvironment()
		defer mockConn.AssertExpectations(t)
		simulateResponse(mockConn, response, nil)

		buf, err := dpdk.connector.getCommandResponse(command)

		require.NoError(t, err)
		require.Equal(t, len(response), len(buf))
		require.Equal(t, response, string(buf))
	})

	t.Run("should return error if failed to get connection handler", func(t *testing.T) {
		_, dpdk, _ := prepareEnvironment()
		dpdk.connector.connection = nil

		buf, err := dpdk.connector.getCommandResponse(command)

		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to get connection to execute / command")
		require.Equal(t, 0, len(buf))
	})

	t.Run("should return error if failed to set timeout duration", func(t *testing.T) {
		mockConn, dpdk, _ := prepareEnvironment()
		defer mockConn.AssertExpectations(t)
		mockConn.On("SetDeadline", mock.Anything).Return(fmt.Errorf("deadline error"))

		buf, err := dpdk.connector.getCommandResponse(command)

		require.Error(t, err)
		require.Contains(t, err.Error(), "deadline error")
		require.Equal(t, 0, len(buf))
	})

	t.Run("should return error if timeout occurred during Write operation", func(t *testing.T) {
		mockConn, dpdk, _ := prepareEnvironment()
		defer mockConn.AssertExpectations(t)
		mockConn.On("Write", mock.Anything).Return(0, fmt.Errorf("write timeout"))
		mockConn.On("SetDeadline", mock.Anything).Return(nil)
		mockConn.On("Close").Return(nil)

		buf, err := dpdk.connector.getCommandResponse(command)

		require.Error(t, err)
		require.Contains(t, err.Error(), "write timeout")
		require.Equal(t, 0, len(buf))
	})

	t.Run("should return error if timeout occurred during Read operation", func(t *testing.T) {
		mockConn, dpdk, _ := prepareEnvironment()
		defer mockConn.AssertExpectations(t)
		simulateResponse(mockConn, "", fmt.Errorf("read timeout"))

		buf, err := dpdk.connector.getCommandResponse(command)

		require.Error(t, err)
		require.Contains(t, err.Error(), "read timeout")
		require.Equal(t, 0, len(buf))
	})

	t.Run("should return error if got empty response", func(t *testing.T) {
		mockConn, dpdk, _ := prepareEnvironment()
		defer mockConn.AssertExpectations(t)
		simulateResponse(mockConn, "", nil)

		buf, err := dpdk.connector.getCommandResponse(command)

		require.Error(t, err)
		require.Equal(t, 0, len(buf))
		require.Contains(t, err.Error(), "got empty response during execution of")
	})
}
