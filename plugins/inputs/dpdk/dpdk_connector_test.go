//go:build linux

package dpdk

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/plugins/inputs/dpdk/mocks"
	"github.com/influxdata/telegraf/testutil"
)

func Test_readMaxOutputLen(t *testing.T) {
	t.Run("should return error if timeout occurred", func(t *testing.T) {
		conn := &mocks.Conn{}
		conn.On("Read", mock.Anything).Return(0, errors.New("timeout"))
		conn.On("SetDeadline", mock.Anything).Return(nil)
		connector := dpdkConnector{connection: conn}

		initMessage, err := connector.readInitMessage()

		require.Error(t, err)
		require.Contains(t, err.Error(), "timeout")
		require.Empty(t, initMessage)
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

		initMsg, err := connector.readInitMessage()

		require.NoError(t, err)
		require.Equal(t, maxOutputLen, initMsg.MaxOutputLen)
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

		_, err := connector.readInitMessage()

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

		_, err = connector.readInitMessage()

		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to read maxOutputLen information")
	})
}

func Test_connect(t *testing.T) {
	t.Run("should pass if PathToSocket points to socket", func(t *testing.T) {
		pathToSocket, socket := createSocketForTest(t, "")
		dpdk := dpdk{
			SocketPath: pathToSocket,
			connectors: []*dpdkConnector{newDpdkConnector(pathToSocket, 0)},
		}
		go simulateSocketResponse(socket, t)

		_, err := dpdk.connectors[0].connect()

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

		for _, connector := range dpdk.connectors {
			buf, err := connector.getCommandResponse(command)

			require.NoError(t, err)
			require.Equal(t, len(response), len(buf))
			require.Equal(t, response, string(buf))
		}
	})

	t.Run("should return error if failed to get connection handler", func(t *testing.T) {
		_, dpdk, _ := prepareEnvironment()
		dpdk.connectors[0].connection = nil

		buf, err := dpdk.connectors[0].getCommandResponse(command)

		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to get connection to execute \"/\" command")
		require.Empty(t, buf)
	})

	t.Run("should return error if failed to set timeout duration", func(t *testing.T) {
		mockConn, dpdk, _ := prepareEnvironment()
		defer mockConn.AssertExpectations(t)
		mockConn.On("SetDeadline", mock.Anything).Return(errors.New("deadline error"))

		buf, err := dpdk.connectors[0].getCommandResponse(command)

		require.Error(t, err)
		require.Contains(t, err.Error(), "deadline error")
		require.Empty(t, buf)
	})

	t.Run("should return error if timeout occurred during Write operation", func(t *testing.T) {
		mockConn, dpdk, _ := prepareEnvironment()
		defer mockConn.AssertExpectations(t)
		mockConn.On("Write", mock.Anything).Return(0, errors.New("write timeout"))
		mockConn.On("SetDeadline", mock.Anything).Return(nil)
		mockConn.On("Close").Return(nil)

		buf, err := dpdk.connectors[0].getCommandResponse(command)

		require.Error(t, err)
		require.Contains(t, err.Error(), "write timeout")
		require.Empty(t, buf)
	})

	t.Run("should return error if timeout occurred during Read operation", func(t *testing.T) {
		mockConn, dpdk, _ := prepareEnvironment()
		defer mockConn.AssertExpectations(t)
		simulateResponse(mockConn, "", errors.New("read timeout"))

		buf, err := dpdk.connectors[0].getCommandResponse(command)

		require.Error(t, err)
		require.Contains(t, err.Error(), "read timeout")
		require.Empty(t, buf)
	})

	t.Run("should return error if got empty response", func(t *testing.T) {
		mockConn, dpdk, _ := prepareEnvironment()
		defer mockConn.AssertExpectations(t)
		simulateResponse(mockConn, "", nil)

		buf, err := dpdk.connectors[0].getCommandResponse(command)

		require.Error(t, err)
		require.Empty(t, buf)
		require.Contains(t, err.Error(), "got empty response during execution of")
	})
}

func Test_processCommand(t *testing.T) {
	t.Run("should pass if received valid response", func(t *testing.T) {
		mockConn, dpdk, mockAcc := prepareEnvironment()
		defer mockConn.AssertExpectations(t)
		response := `{"/": ["/", "/eal/app_params", "/eal/params", "/ethdev/link_status, /ethdev/info"]}`
		simulateResponse(mockConn, response, nil)

		for _, dpdkConn := range dpdk.connectors {
			dpdkConn.processCommand(mockAcc, testutil.Logger{}, "/", nil)
		}

		require.Empty(t, mockAcc.Errors)
	})

	t.Run("if received a non-JSON object then should return error", func(t *testing.T) {
		mockConn, dpdk, mockAcc := prepareEnvironment()
		defer mockConn.AssertExpectations(t)
		response := `notAJson`
		simulateResponse(mockConn, response, nil)

		for _, dpdkConn := range dpdk.connectors {
			dpdkConn.processCommand(mockAcc, testutil.Logger{}, "/", nil)
		}

		require.Len(t, mockAcc.Errors, 1)
		require.Contains(t, mockAcc.Errors[0].Error(), "invalid character")
	})

	t.Run("if failed to get command response then accumulator should contain error", func(t *testing.T) {
		mockConn, dpdk, mockAcc := prepareEnvironment()
		defer mockConn.AssertExpectations(t)
		mockConn.On("Write", mock.Anything).Return(0, errors.New("deadline exceeded"))
		mockConn.On("SetDeadline", mock.Anything).Return(nil)
		mockConn.On("Close").Return(nil)
		for _, dpdkConn := range dpdk.connectors {
			dpdkConn.processCommand(mockAcc, testutil.Logger{}, "/", nil)
		}

		require.Len(t, mockAcc.Errors, 1)
		require.Contains(t, mockAcc.Errors[0].Error(), "deadline exceeded")
	})

	t.Run("if response contains nil or empty value then error shouldn't be returned in accumulator", func(t *testing.T) {
		mockConn, dpdk, mockAcc := prepareEnvironment()
		defer mockConn.AssertExpectations(t)
		response := `{"/test": null}`
		simulateResponse(mockConn, response, nil)
		for _, dpdkConn := range dpdk.connectors {
			dpdkConn.processCommand(mockAcc, testutil.Logger{}, "/test,param", nil)
		}

		require.Empty(t, mockAcc.Errors)
	})
}
