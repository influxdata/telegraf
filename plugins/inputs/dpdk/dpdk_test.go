//go:build linux

package dpdk

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"path/filepath"
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
		dpdk := dpdk{
			Log:        testutil.Logger{},
			SocketPath: "",
		}

		require.Equal(t, "", dpdk.SocketPath)

		_ = dpdk.Init()

		require.Equal(t, defaultPathToSocket, dpdk.SocketPath)
	})

	t.Run("when Metadata Fields isn't set then it should be set to default value (dpdk_pid)", func(t *testing.T) {
		dpdk := dpdk{
			Log: testutil.Logger{},
		}
		require.Nil(t, dpdk.MetadataFields)

		_ = dpdk.Init()
		require.Equal(t, []string{dpdkMetadataFieldPidName}, dpdk.MetadataFields)
	})

	t.Run("when PluginOptions field isn't set then it should be set to default value (dpdk_in_memory)", func(t *testing.T) {
		dpdk := dpdk{
			Log: testutil.Logger{},
		}
		require.Nil(t, dpdk.PluginOptions)

		_ = dpdk.Init()
		require.Equal(t, []string{dpdkPluginOptionInMemory}, dpdk.PluginOptions)
	})

	t.Run("when commands are in invalid format (doesn't start with '/') then error should be returned", func(t *testing.T) {
		pathToSocket, socket := createSocketForTest(t, "")
		defer socket.Close()
		dpdk := dpdk{
			Log:                testutil.Logger{},
			SocketPath:         pathToSocket,
			AdditionalCommands: []string{"invalid"},
		}

		err := dpdk.Init()

		require.Error(t, err)
		require.Contains(t, err.Error(), "command should start with '/'")
	})

	t.Run("when AccessTime is < 0 then error should be returned", func(t *testing.T) {
		dpdk := dpdk{
			Log:           testutil.Logger{},
			AccessTimeout: -1,
		}
		err := dpdk.Init()

		require.Error(t, err)
		require.Contains(t, err.Error(), "socket_access_timeout should be positive number")
	})

	t.Run("when device_types and additional_commands are empty, then error should be returned", func(t *testing.T) {
		pathToSocket, socket := createSocketForTest(t, "")
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

	t.Run("when socket doesn't exist err should be returned", func(t *testing.T) {
		dpdk := dpdk{
			DeviceTypes: []string{"ethdev"},
			Log:         testutil.Logger{},
		}
		err := dpdk.Init()

		require.Error(t, err)
		require.Contains(t, err.Error(), "no active sockets connections present")
	})

	t.Run("when UnreachableSocketBehavior specified with unknown value - err should be returned", func(t *testing.T) {
		dpdk := dpdk{
			DeviceTypes:               []string{"ethdev"},
			Log:                       testutil.Logger{},
			UnreachableSocketBehavior: "whatisthat",
		}
		err := dpdk.Init()

		require.Error(t, err)
		require.Contains(t, err.Error(), "unreachable_socket_behavior")
	})

	t.Run("when socket doesn't exist, but UnreachableSocketBehavior is Ignore err shouldn't be returned", func(t *testing.T) {
		dpdk := dpdk{
			DeviceTypes:               []string{"ethdev"},
			Log:                       testutil.Logger{},
			UnreachableSocketBehavior: unreachableSocketBehaviorIgnore,
		}
		err := dpdk.Init()

		require.NoError(t, err)
	})

	t.Run("when all values are valid, then no error should be returned", func(t *testing.T) {
		pathToSocket, socket := createSocketForTest(t, "")
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
}

func TestMaintainConnections(t *testing.T) {
	t.Run("maintainConnections should return the error if socket doesn't exist", func(t *testing.T) {
		dpdk := dpdk{
			SocketPath:  "/tmp/justrandompath",
			DeviceTypes: []string{"ethdev"},
			Log:         testutil.Logger{},
		}

		require.Empty(t, dpdk.connectors)
		err := dpdk.maintainConnections()
		defer dpdk.Stop()

		require.Error(t, err)
		require.Contains(t, err.Error(), "no active sockets connections present")
	})

	t.Run("maintainConnections should return the error if socket not found with dpdkPluginOptionInMemory", func(t *testing.T) {
		dpdk := dpdk{
			SocketPath:    defaultPathToSocket,
			Log:           testutil.Logger{},
			PluginOptions: []string{dpdkPluginOptionInMemory},
		}
		var err error
		dpdk.socketGlobPath, err = prepareGlob(dpdk.SocketPath)
		require.NoError(t, err)

		require.Empty(t, dpdk.connectors)
		err = dpdk.maintainConnections()
		require.Error(t, err)
		require.Contains(t, err.Error(), "no active sockets connections present")
	})

	t.Run("maintainConnections shouldn't return error with 1 socket", func(t *testing.T) {
		pathToSocket, socket := createSocketForTest(t, "")
		defer socket.Close()
		dpdk := dpdk{
			SocketPath:  pathToSocket,
			DeviceTypes: []string{"ethdev"},
			Log:         testutil.Logger{},
		}

		go simulateSocketResponse(socket, t)

		require.Empty(t, dpdk.connectors)
		err := dpdk.maintainConnections()
		defer dpdk.Stop()

		require.NoError(t, err)
		require.Len(t, dpdk.connectors, 1)
	})

	t.Run("maintainConnections shouldn't return error with multiple sockets", func(t *testing.T) {
		numSockets := rand.Intn(5) + 1

		pathToSockets, sockets := createMultipleSocketsForTest(t, numSockets, "")
		defer func() {
			for _, socket := range sockets {
				socket.Close()
			}
		}()

		dpdk := dpdk{
			SocketPath:    pathToSockets[0],
			DeviceTypes:   []string{"ethdev"},
			Log:           testutil.Logger{},
			PluginOptions: []string{dpdkPluginOptionInMemory},
		}
		var err error
		dpdk.socketGlobPath, err = prepareGlob(dpdk.SocketPath)
		require.NoError(t, err)

		for _, socket := range sockets {
			go simulateSocketResponse(socket, t)
		}

		require.Empty(t, dpdk.connectors)
		err = dpdk.maintainConnections()
		defer dpdk.Stop()

		require.NoError(t, err)
		require.Len(t, dpdk.connectors, numSockets)
	})

	t.Run("Test maintainConnections without dpdkPluginOptionInMemory option", func(t *testing.T) {
		pathToSocket, socket := createSocketForTest(t, "")
		defer socket.Close()
		dpdk := dpdk{
			SocketPath:  pathToSocket,
			DeviceTypes: []string{"ethdev"},
			Log:         testutil.Logger{},
		}

		go simulateSocketResponse(socket, t)

		require.Empty(t, dpdk.connectors)
		err := dpdk.maintainConnections()
		require.NoError(t, err)
		require.Len(t, dpdk.connectors, 1)

		dpdk.Stop()
		require.Empty(t, dpdk.connectors)
	})

	t.Run("Test maintainConnections with dpdkPluginOptionInMemory option", func(t *testing.T) {
		pathToSocket1, socket1 := createSocketForTest(t, "")
		defer socket1.Close()
		go simulateSocketResponse(socket1, t)
		dpdk := dpdk{
			SocketPath:    pathToSocket1,
			DeviceTypes:   []string{"ethdev"},
			Log:           testutil.Logger{},
			PluginOptions: []string{dpdkPluginOptionInMemory},
		}
		var err error
		dpdk.socketGlobPath, err = prepareGlob(dpdk.SocketPath)
		require.NoError(t, err)

		require.Empty(t, dpdk.connectors)
		err = dpdk.maintainConnections()
		require.NoError(t, err)
		require.Len(t, dpdk.connectors, 1)

		// Adding 2 sockets more
		pathToSocket2, socket2 := createSocketForTest(t, filepath.Dir(pathToSocket1))
		pathToSocket3, socket3 := createSocketForTest(t, filepath.Dir(pathToSocket1))
		require.NotEqual(t, pathToSocket2, pathToSocket3)
		go simulateSocketResponse(socket2, t)
		go simulateSocketResponse(socket3, t)
		err = dpdk.maintainConnections()
		require.NoError(t, err)
		require.Len(t, dpdk.connectors, 3)

		// Close 2 new sockets
		socket2.Close()
		socket3.Close()
		err = dpdk.maintainConnections()
		require.NoError(t, err)
		require.Len(t, dpdk.connectors, 1)
		require.Equal(t, pathToSocket1, dpdk.connectors[0].pathToSocket)

		dpdk.Stop()
		require.Empty(t, dpdk.connectors)
	})
}

func TestClose(t *testing.T) {
	t.Run("Num of connections should be 0 after Stop func", func(t *testing.T) {
		pathToSocket, socket := createSocketForTest(t, "")
		defer socket.Close()
		dpdk := dpdk{
			SocketPath:  pathToSocket,
			DeviceTypes: []string{"ethdev"},
			Log:         testutil.Logger{},
		}

		go simulateSocketResponse(socket, t)

		require.Empty(t, dpdk.connectors)
		err := dpdk.maintainConnections()
		require.NoError(t, err)
		require.Len(t, dpdk.connectors, 1)

		dpdk.Stop()
		require.Empty(t, dpdk.connectors)
	})
}

func Test_validateAdditionalCommands(t *testing.T) {
	t.Run("when validating commands in correct format then no error should be returned", func(t *testing.T) {
		dpdk := dpdk{
			AdditionalCommands: []string{"/test", "/help"},
		}

		err := dpdk.validateAdditionalCommands()

		require.NoError(t, err)
	})

	t.Run("when validating command that doesn't begin with slash then error should be returned", func(t *testing.T) {
		dpdk := dpdk{
			AdditionalCommands: []string{
				"/test", "commandWithoutSlash",
			},
		}

		err := dpdk.validateAdditionalCommands()

		require.Error(t, err)
		require.Contains(t, err.Error(), "command should start with '/'")
	})

	t.Run("when validating long command (without parameters) then error should be returned", func(t *testing.T) {
		dpdk := dpdk{
			AdditionalCommands: []string{
				"/test", "/" + strings.Repeat("a", maxCommandLength),
			},
		}

		err := dpdk.validateAdditionalCommands()

		require.Error(t, err)
		require.Contains(t, err.Error(), "command is too long")
	})

	t.Run("when validating long command (with params) then error should be returned", func(t *testing.T) {
		dpdk := dpdk{
			AdditionalCommands: []string{
				"/test", "/," + strings.Repeat("a", maxCommandLengthWithParams),
			},
		}

		err := dpdk.validateAdditionalCommands()

		require.Error(t, err)
		require.Contains(t, err.Error(), "shall be less than 1024 characters")
	})

	t.Run("when validating empty command then error should be returned", func(t *testing.T) {
		dpdk := dpdk{
			AdditionalCommands: []string{
				"/test", "",
			},
		}

		err := dpdk.validateAdditionalCommands()

		require.Error(t, err)
		require.Contains(t, err.Error(), "got empty command")
	})

	t.Run("when validating commands with duplicates then duplicates should be removed and no error should be returned", func(t *testing.T) {
		dpdk := dpdk{
			AdditionalCommands: []string{
				"/test", "/test",
			},
		}
		require.Len(t, dpdk.AdditionalCommands, 2)

		err := dpdk.validateAdditionalCommands()

		require.Len(t, dpdk.AdditionalCommands, 1)
		require.NoError(t, err)
	})
}

func prepareEnvironment() (*mocks.Conn, dpdk, *testutil.Accumulator) {
	mockConnection := &mocks.Conn{}
	dpdk := dpdk{
		connectors: []*dpdkConnector{{
			connection: mockConnection,
			initMessage: &initMessage{
				Version:      "mockedDPDK",
				Pid:          1,
				MaxOutputLen: 1024,
			},
			accessTimeout: 2 * time.Second,
		}},
		Log: testutil.Logger{},
	}
	mockAcc := &testutil.Accumulator{}
	return mockConnection, dpdk, mockAcc
}

func prepareEnvironmentWithMultiSockets() ([]*mocks.Conn, dpdk, *testutil.Accumulator) {
	mockConnections := []*mocks.Conn{{}, {}}
	dpdk := dpdk{
		connectors: []*dpdkConnector{
			{
				connection: mockConnections[0],
				initMessage: &initMessage{
					Version:      "mockedDPDK",
					Pid:          1,
					MaxOutputLen: 1024,
				},
				accessTimeout: 2 * time.Second,
			},
			{
				connection: mockConnections[1],
				initMessage: &initMessage{
					Version:      "mockedDPDK",
					Pid:          2,
					MaxOutputLen: 1024,
				},
				accessTimeout: 2 * time.Second,
			},
		},
		Log: testutil.Logger{},
	}
	mockAcc := &testutil.Accumulator{}
	return mockConnections, dpdk, mockAcc
}

func prepareEnvironmentWithInitializedMessage(initMsg *initMessage) (*mocks.Conn, dpdk, *testutil.Accumulator) {
	mockConnection := &mocks.Conn{}
	dpdk := dpdk{
		connectors: []*dpdkConnector{{
			connection:    mockConnection,
			accessTimeout: 2 * time.Second,
			initMessage:   initMsg,
		}},
		Log: testutil.Logger{},
	}
	mockAcc := &testutil.Accumulator{}
	return mockConnection, dpdk, mockAcc
}

func Test_processCommand(t *testing.T) {
	t.Run("should pass if received valid response", func(t *testing.T) {
		mockConn, dpdk, mockAcc := prepareEnvironment()
		defer mockConn.AssertExpectations(t)
		response := `{"/": ["/", "/eal/app_params", "/eal/params", "/ethdev/link_status, /ethdev/info"]}`
		simulateResponse(mockConn, response, nil)

		for _, dpdkConn := range dpdk.connectors {
			dpdk.processCommand(dpdkConn, mockAcc, "/")
		}

		require.Empty(t, mockAcc.Errors)
	})

	t.Run("if received a non-JSON object then should return error", func(t *testing.T) {
		mockConn, dpdk, mockAcc := prepareEnvironment()
		defer mockConn.AssertExpectations(t)
		response := `notAJson`
		simulateResponse(mockConn, response, nil)

		for _, dpdkConn := range dpdk.connectors {
			dpdk.processCommand(dpdkConn, mockAcc, "/")
		}

		require.Len(t, mockAcc.Errors, 1)
		require.Contains(t, mockAcc.Errors[0].Error(), "invalid character")
	})

	t.Run("if failed to get command response then accumulator should contain error", func(t *testing.T) {
		mockConn, dpdk, mockAcc := prepareEnvironment()
		defer mockConn.AssertExpectations(t)
		mockConn.On("Write", mock.Anything).Return(0, fmt.Errorf("deadline exceeded"))
		mockConn.On("SetDeadline", mock.Anything).Return(nil)
		mockConn.On("Close").Return(nil)
		for _, dpdkConn := range dpdk.connectors {
			dpdk.processCommand(dpdkConn, mockAcc, "/")
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
			dpdk.processCommand(dpdkConn, mockAcc, "/test,param")
		}

		require.Empty(t, mockAcc.Errors)
	})
}

func Test_appendCommandsWithParams(t *testing.T) {
	t.Run("when got valid data, then valid commands with params should be created", func(t *testing.T) {
		mockConn, dpdk, _ := prepareEnvironment()
		defer mockConn.AssertExpectations(t)
		response := `{"/testendpoint": [1,123]}`
		simulateResponse(mockConn, response, nil)
		expectedCommands := []string{"/action1,1", "/action1,123", "/action2,1", "/action2,123"}

		for _, dpdkConn := range dpdk.connectors {
			result, err := dpdkConn.appendCommandsWithParamsFromList("/testendpoint", []string{"/action1", "/action2"})
			require.NoError(t, err)
			require.Len(t, result, 4)
			require.ElementsMatch(t, result, expectedCommands)
		}
	})
}

func Test_getCommandsAndParamsCombinations(t *testing.T) {
	t.Run("when 2 ethdev commands are enabled, then 2*numberOfIds new commands should be appended", func(t *testing.T) {
		mockConn, dpdk, mockAcc := prepareEnvironment()
		defer mockConn.AssertExpectations(t)
		response := fmt.Sprintf(`{%q: [1, 123]}`, ethdevListCommand)
		simulateResponse(mockConn, response, nil)
		expectedCommands := []string{"/ethdev/stats,1", "/ethdev/stats,123", "/ethdev/xstats,1", "/ethdev/xstats,123"}

		dpdk.DeviceTypes = []string{"ethdev"}
		dpdk.ethdevCommands = []string{"/ethdev/stats", "/ethdev/xstats"}
		dpdk.ethdevExcludedCommandsFilter, _ = filter.Compile([]string{})
		dpdk.AdditionalCommands = []string{}
		commands := dpdk.gatherCommands(dpdk.connectors[0], mockAcc)

		require.ElementsMatch(t, commands, expectedCommands)
		require.Empty(t, mockAcc.Errors)
	})

	t.Run("when 1 rawdev command is enabled, then 2*numberOfIds new commands should be appended", func(t *testing.T) {
		mockConn, dpdk, mockAcc := prepareEnvironment()
		defer mockConn.AssertExpectations(t)
		response := fmt.Sprintf(`{%q: [1, 123]}`, rawdevListCommand)
		simulateResponse(mockConn, response, nil)
		expectedCommands := []string{"/rawdev/xstats,1", "/rawdev/xstats,123"}

		dpdk.DeviceTypes = []string{"rawdev"}
		dpdk.rawdevCommands = []string{"/rawdev/xstats"}
		dpdk.AdditionalCommands = []string{}
		commands := dpdk.gatherCommands(dpdk.connectors[0], mockAcc)

		require.ElementsMatch(t, commands, expectedCommands)
		require.Empty(t, mockAcc.Errors)
	})

	t.Run("when 2 ethdev commands are enabled but one command is disabled, then numberOfIds new commands should be appended", func(t *testing.T) {
		mockConn, dpdk, mockAcc := prepareEnvironment()
		defer mockConn.AssertExpectations(t)
		response := fmt.Sprintf(`{%q: [1, 123]}`, ethdevListCommand)
		simulateResponse(mockConn, response, nil)
		expectedCommands := []string{"/ethdev/stats,1", "/ethdev/stats,123"}

		dpdk.DeviceTypes = []string{"ethdev"}
		dpdk.ethdevCommands = []string{"/ethdev/stats", "/ethdev/xstats"}
		dpdk.ethdevExcludedCommandsFilter, _ = filter.Compile([]string{"/ethdev/xstats"})
		dpdk.AdditionalCommands = []string{}
		commands := dpdk.gatherCommands(dpdk.connectors[0], mockAcc)

		require.ElementsMatch(t, commands, expectedCommands)
		require.Empty(t, mockAcc.Errors)
	})

	t.Run("when ethdev commands are enabled but params fetching command returns error then error should be logged in accumulator", func(t *testing.T) {
		mockConn, dpdk, mockAcc := prepareEnvironment()
		defer mockConn.AssertExpectations(t)
		simulateResponse(mockConn, `{notAJson}`, fmt.Errorf("some error"))

		dpdk.DeviceTypes = []string{"ethdev"}
		dpdk.ethdevCommands = []string{"/ethdev/stats", "/ethdev/xstats"}
		dpdk.ethdevExcludedCommandsFilter, _ = filter.Compile([]string{})
		dpdk.AdditionalCommands = []string{}
		commands := dpdk.gatherCommands(dpdk.connectors[0], mockAcc)

		require.Empty(t, commands)
		require.Len(t, mockAcc.Errors, 1)
	})
}

func Test_Gather(t *testing.T) {
	t.Run("Gather should return error, because socket weren't created", func(t *testing.T) {
		mockAcc := &testutil.Accumulator{}
		dpdk := dpdk{
			Log:           testutil.Logger{},
			PluginOptions: []string{},
		}

		_ = dpdk.Init()

		err := dpdk.Gather(mockAcc)
		require.Error(t, err)
		require.Contains(t, err.Error(), "couldn't connect to socket")
	})

	t.Run("Gather shouldn't return error with UnreachableSocketBehavior: Ignore option, because socket weren't created", func(t *testing.T) {
		mockAcc := &testutil.Accumulator{}
		dpdk := dpdk{
			Log:                       testutil.Logger{},
			PluginOptions:             []string{},
			UnreachableSocketBehavior: unreachableSocketBehaviorIgnore,
		}
		_ = dpdk.Init()

		err := dpdk.Gather(mockAcc)
		require.NoError(t, err)
	})

	t.Run("When parsing a plain json without nested object, then its key should be equal to \"\"", func(t *testing.T) {
		mockConn, dpdk, mockAcc := prepareEnvironment()
		defer mockConn.AssertExpectations(t)
		dpdk.AdditionalCommands = []string{"/endpoint1"}
		simulateResponse(mockConn, `{"/endpoint1":"myvalue"}`, nil)

		err := dpdk.Gather(mockAcc)

		require.NoError(t, err)
		require.Empty(t, mockAcc.Errors)

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
		require.Empty(t, mockAcc.Errors)

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

	t.Run("Test Gather with Metadata Fields dpdk_pid and dpdk_version", func(t *testing.T) {
		testInitMessage := &initMessage{
			Pid:          100,
			Version:      "DPDK 21.11.11",
			MaxOutputLen: 1024,
		}
		mockConn, dpdk, mockAcc := prepareEnvironmentWithInitializedMessage(testInitMessage)
		dpdk.MetadataFields = []string{dpdkMetadataFieldPidName, dpdkMetadataFieldVersionName}
		defer mockConn.AssertExpectations(t)
		dpdk.AdditionalCommands = []string{"/endpoint1"}
		simulateResponse(mockConn, `{"/endpoint1":"myvalue"}`, nil)

		err := dpdk.Gather(mockAcc)

		require.NoError(t, err)
		require.Empty(t, mockAcc.Errors)

		expected := []telegraf.Metric{
			testutil.MustMetric(
				"dpdk",
				map[string]string{
					"command": "/endpoint1",
					"params":  "",
				},
				map[string]interface{}{
					"":                           "myvalue",
					dpdkMetadataFieldPidName:     testInitMessage.Pid,
					dpdkMetadataFieldVersionName: testInitMessage.Version,
				},
				time.Unix(0, 0),
			),
		}

		actual := mockAcc.GetTelegrafMetrics()
		testutil.RequireMetricsEqual(t, expected, actual, testutil.IgnoreTime())
	})

	t.Run("Test Gather with Metadata Fields dpdk_pid", func(t *testing.T) {
		testInitMessage := &initMessage{
			Pid:          100,
			Version:      "DPDK 21.11.11",
			MaxOutputLen: 1024,
		}
		mockConn, dpdk, mockAcc := prepareEnvironmentWithInitializedMessage(testInitMessage)
		dpdk.MetadataFields = []string{dpdkMetadataFieldPidName}
		defer mockConn.AssertExpectations(t)
		dpdk.AdditionalCommands = []string{"/endpoint1"}
		simulateResponse(mockConn, `{"/endpoint1":"myvalue"}`, nil)

		err := dpdk.Gather(mockAcc)

		require.NoError(t, err)
		require.Empty(t, mockAcc.Errors)

		expected := []telegraf.Metric{
			testutil.MustMetric(
				"dpdk",
				map[string]string{
					"command": "/endpoint1",
					"params":  "",
				},
				map[string]interface{}{
					"":                       "myvalue",
					dpdkMetadataFieldPidName: testInitMessage.Pid,
				},
				time.Unix(0, 0),
			),
		}

		actual := mockAcc.GetTelegrafMetrics()
		testutil.RequireMetricsEqual(t, expected, actual, testutil.IgnoreTime())
	})

	t.Run("Test Gather without Metadata Fields", func(t *testing.T) {
		testInitMessage := &initMessage{
			Pid:          100,
			Version:      "DPDK 21.11.11",
			MaxOutputLen: 1024,
		}
		mockConn, dpdk, mockAcc := prepareEnvironmentWithInitializedMessage(testInitMessage)
		dpdk.MetadataFields = []string{}
		defer mockConn.AssertExpectations(t)
		dpdk.AdditionalCommands = []string{"/endpoint1"}
		simulateResponse(mockConn, `{"/endpoint1":"myvalue"}`, nil)

		err := dpdk.Gather(mockAcc)

		require.NoError(t, err)
		require.Empty(t, mockAcc.Errors)

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
}

func Test_Gather_MultiSocket(t *testing.T) {
	t.Run("Test Gather without Metadata Fields", func(t *testing.T) {
		mockConns, dpdk, mockAcc := prepareEnvironmentWithMultiSockets()
		dpdk.MetadataFields = []string{}
		defer func() {
			for _, mockConn := range mockConns {
				mockConn.AssertExpectations(t)
			}
		}()
		dpdk.AdditionalCommands = []string{"/endpoint1"}

		for _, mockConn := range mockConns {
			simulateResponse(mockConn, `{"/endpoint1":"myvalue"}`, nil)
		}

		err := dpdk.Gather(mockAcc)

		require.NoError(t, err)
		require.Empty(t, mockAcc.Errors)

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
