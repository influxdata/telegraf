//go:build linux
// +build linux

package dpdk

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/plugins/inputs/dpdk/mocks"
	"github.com/influxdata/telegraf/testutil"
)

func Test_Init(t *testing.T) {
	t.Run("when SocketPath field isn't set then it should be set to default value", func(t *testing.T) {
		_, dpdk, _ := prepareEnvironment()
		dpdk.SocketPath = ""
		require.Equal(t, "", dpdk.SocketPath)

		_ = dpdk.Init()

		require.Equal(t, defaultPathToSocket, dpdk.SocketPath)
	})

	t.Run("when commands are in invalid format (doesn't start with '/') then error should be returned", func(t *testing.T) {
		pathToSocket, socket := createSocketForTest(t)
		defer socket.Close()
		dpdk := dpdk{
			SocketPath:         pathToSocket,
			AdditionalCommands: []string{"invalid"},
		}

		err := dpdk.Init()

		require.Error(t, err)
		require.Contains(t, err.Error(), "command should start with '/'")
	})

	t.Run("when all values are valid, then no error should be returned", func(t *testing.T) {
		pathToSocket, socket := createSocketForTest(t)
		defer socket.Close()
		dpdk := dpdk{
			SocketPath:  pathToSocket,
			DeviceTypes: []string{"ethdev"},
			Log:         testutil.Logger{},
		}
		go simulateSocketResponse(socket, t)

		err := dpdk.Init()

		require.NoError(t, err)
	})

	t.Run("when device_types and additional_commands are empty, then error should be returned", func(t *testing.T) {
		pathToSocket, socket := createSocketForTest(t)
		defer socket.Close()
		dpdk := dpdk{
			SocketPath:         pathToSocket,
			DeviceTypes:        []string{},
			AdditionalCommands: []string{},
			Log:                testutil.Logger{},
		}

		err := dpdk.Init()

		require.Error(t, err)
		require.Contains(t, err.Error(), "plugin was configured with nothing to read")
	})
}

func Test_validateCommands(t *testing.T) {
	t.Run("when validating commands in correct format then no error should be returned", func(t *testing.T) {
		dpdk := dpdk{
			AdditionalCommands: []string{"/test", "/help"},
		}

		err := dpdk.validateCommands()

		require.NoError(t, err)
	})

	t.Run("when validating command that doesn't begin with slash then error should be returned", func(t *testing.T) {
		dpdk := dpdk{
			AdditionalCommands: []string{
				"/test", "commandWithoutSlash",
			},
		}

		err := dpdk.validateCommands()

		require.Error(t, err)
		require.Contains(t, err.Error(), "command should start with '/'")
	})

	t.Run("when validating long command (without parameters) then error should be returned", func(t *testing.T) {
		dpdk := dpdk{
			AdditionalCommands: []string{
				"/test", "/" + strings.Repeat("a", maxCommandLength),
			},
		}

		err := dpdk.validateCommands()

		require.Error(t, err)
		require.Contains(t, err.Error(), "command is too long")
	})

	t.Run("when validating long command (with params) then error should be returned", func(t *testing.T) {
		dpdk := dpdk{
			AdditionalCommands: []string{
				"/test", "/," + strings.Repeat("a", maxCommandLengthWithParams),
			},
		}

		err := dpdk.validateCommands()

		require.Error(t, err)
		require.Contains(t, err.Error(), "shall be less than 1024 characters")
	})

	t.Run("when validating empty command then error should be returned", func(t *testing.T) {
		dpdk := dpdk{
			AdditionalCommands: []string{
				"/test", "",
			},
		}

		err := dpdk.validateCommands()

		require.Error(t, err)
		require.Contains(t, err.Error(), "got empty command")
	})

	t.Run("when validating commands with duplicates then duplicates should be removed and no error should be returned", func(t *testing.T) {
		dpdk := dpdk{
			AdditionalCommands: []string{
				"/test", "/test",
			},
		}
		require.Equal(t, 2, len(dpdk.AdditionalCommands))

		err := dpdk.validateCommands()

		require.Equal(t, 1, len(dpdk.AdditionalCommands))
		require.NoError(t, err)
	})
}

func prepareEnvironment() (*mocks.Conn, dpdk, *testutil.Accumulator) {
	mockConnection := &mocks.Conn{}
	dpdk := dpdk{
		connector: &dpdkConnector{
			connection:    mockConnection,
			maxOutputLen:  1024,
			accessTimeout: 2 * time.Second,
		},
		Log: testutil.Logger{},
	}
	mockAcc := &testutil.Accumulator{}
	return mockConnection, dpdk, mockAcc
}

func Test_processCommand(t *testing.T) {
	t.Run("should pass if received valid response", func(t *testing.T) {
		mockConn, dpdk, mockAcc := prepareEnvironment()
		defer mockConn.AssertExpectations(t)
		response := `{"/": ["/", "/eal/app_params", "/eal/params", "/ethdev/link_status"]}`
		simulateResponse(mockConn, response, nil)

		dpdk.processCommand(mockAcc, "/")

		require.Equal(t, 0, len(mockAcc.Errors))
	})

	t.Run("if received a non-JSON object then should return error", func(t *testing.T) {
		mockConn, dpdk, mockAcc := prepareEnvironment()
		defer mockConn.AssertExpectations(t)
		response := `notAJson`
		simulateResponse(mockConn, response, nil)

		dpdk.processCommand(mockAcc, "/")

		require.Equal(t, 1, len(mockAcc.Errors))
		require.Contains(t, mockAcc.Errors[0].Error(), "invalid character")
	})

	t.Run("if failed to get command response then accumulator should contain error", func(t *testing.T) {
		mockConn, dpdk, mockAcc := prepareEnvironment()
		defer mockConn.AssertExpectations(t)
		mockConn.On("Write", mock.Anything).Return(0, fmt.Errorf("deadline exceeded"))
		mockConn.On("SetDeadline", mock.Anything).Return(nil)
		mockConn.On("Close").Return(nil)

		dpdk.processCommand(mockAcc, "/")

		require.Equal(t, 1, len(mockAcc.Errors))
		require.Contains(t, mockAcc.Errors[0].Error(), "deadline exceeded")
	})

	t.Run("if response contains nil or empty value then error should be returned in accumulator", func(t *testing.T) {
		mockConn, dpdk, mockAcc := prepareEnvironment()
		defer mockConn.AssertExpectations(t)
		response := `{"/test": null}`
		simulateResponse(mockConn, response, nil)

		dpdk.processCommand(mockAcc, "/test,param")

		require.Equal(t, 1, len(mockAcc.Errors))
		require.Contains(t, mockAcc.Errors[0].Error(), "got empty json on")
	})
}

func Test_appendCommandsWithParams(t *testing.T) {
	t.Run("when got valid data, then valid commands with params should be created", func(t *testing.T) {
		mockConn, dpdk, _ := prepareEnvironment()
		defer mockConn.AssertExpectations(t)
		response := `{"/testendpoint": [1,123]}`
		simulateResponse(mockConn, response, nil)
		expectedCommands := []string{"/action1,1", "/action1,123", "/action2,1", "/action2,123"}

		result, err := dpdk.appendCommandsWithParamsFromList("/testendpoint", []string{"/action1", "/action2"})

		require.NoError(t, err)
		require.Equal(t, 4, len(result))
		require.ElementsMatch(t, result, expectedCommands)
	})
}

func Test_getCommandsAndParamsCombinations(t *testing.T) {
	t.Run("when 2 ethdev commands are enabled, then 2*numberOfIds new commands should be appended", func(t *testing.T) {
		mockConn, dpdk, mockAcc := prepareEnvironment()
		defer mockConn.AssertExpectations(t)
		response := fmt.Sprintf(`{"%s": [1, 123]}`, ethdevListCommand)
		simulateResponse(mockConn, response, nil)
		expectedCommands := []string{"/ethdev/stats,1", "/ethdev/stats,123", "/ethdev/xstats,1", "/ethdev/xstats,123"}

		dpdk.DeviceTypes = []string{"ethdev"}
		dpdk.ethdevCommands = []string{"/ethdev/stats", "/ethdev/xstats"}
		dpdk.ethdevExcludedCommandsFilter, _ = filter.Compile([]string{})
		dpdk.AdditionalCommands = []string{}
		commands := dpdk.gatherCommands(mockAcc)

		require.ElementsMatch(t, commands, expectedCommands)
		require.Equal(t, 0, len(mockAcc.Errors))
	})

	t.Run("when 1 rawdev command is enabled, then 2*numberOfIds new commands should be appended", func(t *testing.T) {
		mockConn, dpdk, mockAcc := prepareEnvironment()
		defer mockConn.AssertExpectations(t)
		response := fmt.Sprintf(`{"%s": [1, 123]}`, rawdevListCommand)
		simulateResponse(mockConn, response, nil)
		expectedCommands := []string{"/rawdev/xstats,1", "/rawdev/xstats,123"}

		dpdk.DeviceTypes = []string{"rawdev"}
		dpdk.rawdevCommands = []string{"/rawdev/xstats"}
		dpdk.AdditionalCommands = []string{}
		commands := dpdk.gatherCommands(mockAcc)

		require.ElementsMatch(t, commands, expectedCommands)
		require.Equal(t, 0, len(mockAcc.Errors))
	})

	t.Run("when 2 ethdev commands are enabled but one command is disabled, then numberOfIds new commands should be appended", func(t *testing.T) {
		mockConn, dpdk, mockAcc := prepareEnvironment()
		defer mockConn.AssertExpectations(t)
		response := fmt.Sprintf(`{"%s": [1, 123]}`, ethdevListCommand)
		simulateResponse(mockConn, response, nil)
		expectedCommands := []string{"/ethdev/stats,1", "/ethdev/stats,123"}

		dpdk.DeviceTypes = []string{"ethdev"}
		dpdk.ethdevCommands = []string{"/ethdev/stats", "/ethdev/xstats"}
		dpdk.ethdevExcludedCommandsFilter, _ = filter.Compile([]string{"/ethdev/xstats"})
		dpdk.AdditionalCommands = []string{}
		commands := dpdk.gatherCommands(mockAcc)

		require.ElementsMatch(t, commands, expectedCommands)
		require.Equal(t, 0, len(mockAcc.Errors))
	})

	t.Run("when ethdev commands are enabled but params fetching command returns error then error should be logged in accumulator", func(t *testing.T) {
		mockConn, dpdk, mockAcc := prepareEnvironment()
		defer mockConn.AssertExpectations(t)
		simulateResponse(mockConn, `{notAJson}`, fmt.Errorf("some error"))

		dpdk.DeviceTypes = []string{"ethdev"}
		dpdk.ethdevCommands = []string{"/ethdev/stats", "/ethdev/xstats"}
		dpdk.ethdevExcludedCommandsFilter, _ = filter.Compile([]string{})
		dpdk.AdditionalCommands = []string{}
		commands := dpdk.gatherCommands(mockAcc)

		require.Equal(t, 0, len(commands))
		require.Equal(t, 1, len(mockAcc.Errors))
	})
}

func Test_Gather(t *testing.T) {
	t.Run("When parsing a plain json without nested object, then its key should be equal to \"\"", func(t *testing.T) {
		mockConn, dpdk, mockAcc := prepareEnvironment()
		defer mockConn.AssertExpectations(t)
		dpdk.AdditionalCommands = []string{"/endpoint1"}
		simulateResponse(mockConn, `{"/endpoint1":"myvalue"}`, nil)

		err := dpdk.Gather(mockAcc)

		require.NoError(t, err)
		require.Equal(t, 0, len(mockAcc.Errors))

		expected := []telegraf.Metric{
			testutil.MustMetric(
				"dpdk",
				map[string]string{
					"command": "/endpoint1",
					"params":  "",
				},
				map[string]interface{}{
					"": "myvalue",
				},
				time.Unix(0, 0),
			),
		}

		actual := mockAcc.GetTelegrafMetrics()
		testutil.RequireMetricsEqual(t, expected, actual, testutil.IgnoreTime())
	})

	t.Run("When parsing a list of value in nested object then list should be flattened", func(t *testing.T) {
		mockConn, dpdk, mockAcc := prepareEnvironment()
		defer mockConn.AssertExpectations(t)
		dpdk.AdditionalCommands = []string{"/endpoint1"}
		simulateResponse(mockConn, `{"/endpoint1":{"myvalue":[0,1,123]}}`, nil)

		err := dpdk.Gather(mockAcc)
		require.NoError(t, err)
		require.Equal(t, 0, len(mockAcc.Errors))

		expected := []telegraf.Metric{
			testutil.MustMetric(
				"dpdk",
				map[string]string{
					"command": "/endpoint1",
					"params":  "",
				},
				map[string]interface{}{
					"myvalue_0": float64(0),
					"myvalue_1": float64(1),
					"myvalue_2": float64(123),
				},
				time.Unix(0, 0),
			),
		}

		actual := mockAcc.GetTelegrafMetrics()
		testutil.RequireMetricsEqual(t, expected, actual, testutil.IgnoreTime())
	})
}

func simulateResponse(mockConn *mocks.Conn, response string, readErr error) {
	mockConn.On("Write", mock.Anything).Return(0, nil)
	mockConn.On("Read", mock.Anything).Run(func(arg mock.Arguments) {
		elem := arg.Get(0).([]byte)
		copy(elem, response)
	}).Return(len(response), readErr)
	mockConn.On("SetDeadline", mock.Anything).Return(nil)

	if readErr != nil {
		mockConn.On("Close").Return(nil)
	}
}

func simulateSocketResponse(socket net.Listener, t *testing.T) {
	conn, err := socket.Accept()
	require.NoError(t, err)

	initMessage, err := json.Marshal(initMessage{MaxOutputLen: 1})
	require.NoError(t, err)

	_, err = conn.Write(initMessage)
	require.NoError(t, err)
}
