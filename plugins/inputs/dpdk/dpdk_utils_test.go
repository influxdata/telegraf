//go:build linux

package dpdk

import (
	"fmt"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/testutil"
)

func Test_isSocket(t *testing.T) {
	t.Run("when path points to non-existing file then error should be returned", func(t *testing.T) {
		err := isSocket("/tmp/file-that-doesnt-exists")

		require.Error(t, err)
		require.Contains(t, err.Error(), "provided path does not exist")
	})

	t.Run("Should pass if path points to socket", func(t *testing.T) {
		pathToSocket, socket := createSocketForTest(t, "")
		defer socket.Close()

		err := isSocket(pathToSocket)

		require.NoError(t, err)
	})

	t.Run("if path points to regular file instead of socket then error should be returned", func(t *testing.T) {
		pathToFile := "/tmp/dpdk-text-file.txt"
		_, err := os.Create(pathToFile)
		require.NoError(t, err)
		defer os.Remove(pathToFile)

		err = isSocket(pathToFile)

		require.Error(t, err)
		require.Contains(t, err.Error(), "provided path does not point to a socket file")
	})
}

func Test_stripParams(t *testing.T) {
	command := "/mycommand"
	params := "myParams"
	t.Run("when passed string without params then passed string should be returned", func(t *testing.T) {
		strippedCommand := stripParams(command)

		require.Equal(t, command, strippedCommand)
	})

	t.Run("when passed string with params then string without params should be returned", func(t *testing.T) {
		strippedCommand := stripParams(commandWithParams(command, params))

		require.Equal(t, command, strippedCommand)
	})
}

func Test_commandWithParams(t *testing.T) {
	command := "/mycommand"
	params := "myParams"
	t.Run("when passed string with params then command with comma should be returned", func(t *testing.T) {
		commandWithParams := commandWithParams(command, params)

		require.Equal(t, command+","+params, commandWithParams)
	})

	t.Run("when passed command with no params then command should be returned", func(t *testing.T) {
		commandWithParams := commandWithParams(command, "")

		require.Equal(t, command, commandWithParams)
	})
}

func Test_getParams(t *testing.T) {
	command := "/mycommand"
	params := "myParams"
	t.Run("when passed string with params then command with comma should be returned", func(t *testing.T) {
		commandParams := getParams(commandWithParams(command, params))

		require.Equal(t, params, commandParams)
	})

	t.Run("when passed command with no params then empty string (representing empty params) should be returned", func(t *testing.T) {
		commandParams := getParams(commandWithParams(command, ""))

		require.Equal(t, "", commandParams)
	})
}

func Test_jsonToArray(t *testing.T) {
	key := "/ethdev/list"
	t.Run("when got numeric array then string array should be returned", func(t *testing.T) {
		firstValue := int64(0)
		secondValue := int64(1)
		jsonString := fmt.Sprintf(`{%q: [%d, %d]}`, key, firstValue, secondValue)

		arr, err := jsonToArray([]byte(jsonString), key)

		require.NoError(t, err)
		require.Equal(t, strconv.FormatInt(firstValue, 10), arr[0])
		require.Equal(t, strconv.FormatInt(secondValue, 10), arr[1])
	})

	t.Run("if non-json string is supplied as input then error should be returned", func(t *testing.T) {
		_, err := jsonToArray([]byte("{notAJson}"), key)

		require.Error(t, err)
	})

	t.Run("when empty string is supplied as input then error should be returned", func(t *testing.T) {
		jsonString := ""

		_, err := jsonToArray([]byte(jsonString), key)

		require.Error(t, err)
		require.Contains(t, err.Error(), "got empty object instead of json")
	})

	t.Run("when valid json with json-object is supplied as input then error should be returned", func(t *testing.T) {
		jsonString := fmt.Sprintf(`{%q: {"testKey": "testValue"}}`, key)

		_, err := jsonToArray([]byte(jsonString), key)

		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to unmarshal json response")
	})
}

func Test_getDpdkInMemorySocketPaths(t *testing.T) {
	var err error

	t.Run("Should return nil if path doesn't exist", func(t *testing.T) {
		dpdk := dpdk{
			SocketPath: "/tmp/nothing-should-exist-here/test.socket",
			Log:        testutil.Logger{},
		}
		dpdk.socketGlobPath, err = prepareGlob(dpdk.SocketPath)
		require.NoError(t, err)

		socketsPaths := dpdk.getDpdkInMemorySocketPaths()
		require.Nil(t, socketsPaths)
	})

	t.Run("Should return nil if can't read the dir", func(t *testing.T) {
		dpdk := dpdk{
			SocketPath: "/root/no_access",
			Log:        testutil.Logger{},
		}
		dpdk.socketGlobPath, err = prepareGlob(dpdk.SocketPath)
		require.NoError(t, err)

		socketsPaths := dpdk.getDpdkInMemorySocketPaths()
		require.Nil(t, socketsPaths)
	})

	t.Run("Should return one socket from socket path", func(t *testing.T) {
		socketPath, socket := createSocketForTest(t, "")
		defer socket.Close()

		dpdk := dpdk{
			SocketPath: socketPath,
			Log:        testutil.Logger{},
		}
		dpdk.socketGlobPath, err = prepareGlob(dpdk.SocketPath)
		require.NoError(t, err)

		socketsPaths := dpdk.getDpdkInMemorySocketPaths()
		require.Len(t, socketsPaths, 1)
		require.Equal(t, socketPath, socketsPaths[0])
	})

	t.Run("Should return 2 sockets from socket path", func(t *testing.T) {
		socketPaths, sockets := createMultipleSocketsForTest(t, 2, "")
		defer func() {
			for _, socket := range sockets {
				socket.Close()
			}
		}()

		dpdk := dpdk{
			SocketPath: socketPaths[0],
			Log:        testutil.Logger{},
		}
		dpdk.socketGlobPath, err = prepareGlob(dpdk.SocketPath)
		require.NoError(t, err)

		socketsPathsFromFunc := dpdk.getDpdkInMemorySocketPaths()
		require.Len(t, socketsPathsFromFunc, 2)
		require.Equal(t, socketPaths, socketsPathsFromFunc)
	})
}

func TestGetDiffArrays(t *testing.T) {
	t.Run("Should return empty lists toAdd and toDelete", func(t *testing.T) {
		oldArray := []string{}
		newArray := []string{}

		toAdd, toDel := getDiffArrays(oldArray, newArray)
		require.Empty(t, toAdd)
		require.Empty(t, toDel)
	})

	t.Run("Should return only toDel list", func(t *testing.T) {
		oldArray := []string{"path1"}
		newArray := []string{}

		toDelExpected := []string{"path1"}
		toAdd, toDel := getDiffArrays(oldArray, newArray)

		require.Empty(t, toAdd)
		require.ElementsMatch(t, toDelExpected, toDel)
	})

	t.Run("Should return only toAdd list", func(t *testing.T) {
		oldArray := []string{}
		newArray := []string{"path1"}

		toAddExpected := []string{"path1"}
		toAdd, toDel := getDiffArrays(oldArray, newArray)

		require.ElementsMatch(t, toAddExpected, toAdd)
		require.Empty(t, toDel)
	})

	t.Run("Should return correct list toAdd and toDelete", func(t *testing.T) {
		oldArray := []string{"path1", "path2", "path3"}
		newArray := []string{"path1", "path4"}

		toAddExpected := []string{"path4"}
		toDelExpected := []string{"path2", "path3"}
		toAdd, toDel := getDiffArrays(oldArray, newArray)
		require.ElementsMatch(t, toAddExpected, toAdd)
		require.ElementsMatch(t, toDelExpected, toDel)
	})
}

func createSocketForTest(t *testing.T, dirPath string) (string, net.Listener) {
	var err error
	var pathToSocket string
	if len(dirPath) == 0 {
		dirPath, err = os.MkdirTemp("", "dpdk-test-socket")
		require.NoError(t, err)
		pathToSocket = filepath.Join(dirPath, dpdkSocketTemplateName)
	} else {
		pathToSocket = fmt.Sprintf("%s:%d", filepath.Join(dirPath, dpdkSocketTemplateName), rand.Intn(100)+1)
	}

	socket, err := net.Listen("unixpacket", pathToSocket)
	require.NoError(t, err)
	return pathToSocket, socket
}

func createMultipleSocketsForTest(t *testing.T, numSockets int, dirPath string) (socketsPaths []string, sockets []net.Listener) {
	var err error
	if len(dirPath) == 0 {
		dirPath, err = os.MkdirTemp("", "dpdk-test-socket")
	}
	require.NoError(t, err)

	for i := 0; i < numSockets; i++ {
		var pathToSocket string
		if i == 0 {
			pathToSocket = filepath.Join(dirPath, dpdkSocketTemplateName)
		} else {
			pathToSocket = filepath.Join(dirPath, fmt.Sprintf("%s:%d", dpdkSocketTemplateName, 1000+i))
		}
		socket, err := net.Listen("unixpacket", pathToSocket)
		require.NoError(t, err)
		socketsPaths = append(socketsPaths, pathToSocket)
		sockets = append(sockets, socket)
	}
	return socketsPaths, sockets
}
