//go:build linux
// +build linux

package dpdk

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_isSocket(t *testing.T) {
	t.Run("when path points to non-existing file then error should be returned", func(t *testing.T) {
		err := isSocket("/tmp/file-that-doesnt-exists")

		require.Error(t, err)
		require.Contains(t, err.Error(), "provided path does not exist")
	})

	t.Run("should pass if path points to socket", func(t *testing.T) {
		pathToSocket, socket := createSocketForTest(t)
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
		jsonString := fmt.Sprintf(`{"%s": [%d, %d]}`, key, firstValue, secondValue)

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
		jsonString := fmt.Sprintf(`{"%s": {"testKey": "testValue"}}`, key)

		_, err := jsonToArray([]byte(jsonString), key)

		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to unmarshall json response")
	})
}

func createSocketForTest(t *testing.T) (string, net.Listener) {
	pathToSocket := "/tmp/dpdk-test-socket"
	socket, err := net.Listen("unixpacket", pathToSocket)
	require.NoError(t, err)
	return pathToSocket, socket
}
