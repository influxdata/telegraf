//go:build linux && amd64

package intel_baseband

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMetricNameToTagName(t *testing.T) {
	testCases := []struct {
		metricName      string
		expectedTagName string
	}{
		{vfCodeBlocks, "code_blocks"},
		{vfDataBlock, "data_bytes"},
		{engineBlock, "per_engine"},
		{"", ""},
	}

	t.Run("check the correct transformation metric name", func(t *testing.T) {
		for _, tc := range testCases {
			tagName := metricNameToTagName(tc.metricName)
			require.Equal(t, tc.expectedTagName, tagName)
		}
	})
}

func TestValidatePath(t *testing.T) {
	t.Run("with correct file extensions checkFile shouldn't return any errors", func(t *testing.T) {
		testCases := []struct {
			path         string
			ft           fileType
			expectedPath string
		}{
			{"/tmp/socket.sock", socket, "/tmp/socket.sock"},
			{"/foo/../tmp/socket.sock", socket, "/tmp/socket.sock"},
			{"/tmp/file.log", log, "/tmp/file.log"},
			{"/foo/../tmp/file.log", log, "/tmp/file.log"},
		}

		for _, tc := range testCases {
			returnPath, err := validatePath(tc.path, tc.ft)
			require.Equal(t, tc.expectedPath, returnPath)
			require.NoError(t, err)
		}
	})
	t.Run("with empty path specified validate path should return an error", func(t *testing.T) {
		testCases := []struct {
			path                  string
			ft                    fileType
			expectedErrorContains string
		}{
			{"", socket, "required path not specified"},
			{"", log, "required path not specified"},
		}

		for _, tc := range testCases {
			returnPath, err := validatePath(tc.path, tc.ft)
			require.Equal(t, "", returnPath)
			require.ErrorContains(t, err, tc.expectedErrorContains)
		}
	})
	t.Run("with wrong extension file validatePath should return an error", func(t *testing.T) {
		testCases := []struct {
			path                  string
			ft                    fileType
			expectedErrorContains string
		}{
			{"/tmp/socket.foo", socket, "wrong file extension"},
			{"/tmp/file.foo", log, "wrong file extension"},
			{"/tmp/socket.sock", log, "wrong file extension"},
			{"/tmp/file.log", socket, "wrong file extension"},
		}

		for _, tc := range testCases {
			returnPath, err := validatePath(tc.path, tc.ft)
			require.Equal(t, "", returnPath)
			require.ErrorContains(t, err, tc.expectedErrorContains)
		}
	})
	t.Run("with not absolute path validatePath should return the error", func(t *testing.T) {
		testCases := []struct {
			path                  string
			ft                    fileType
			expectedErrorContains string
		}{
			{"foo/tmp/socket.sock", socket, "path is not absolute"},
			{"foo/tmp/file.log", log, "path is not absolute"},
		}

		for _, tc := range testCases {
			returnPath, err := validatePath(tc.path, tc.ft)
			require.Equal(t, "", returnPath)
			require.ErrorContains(t, err, tc.expectedErrorContains)
		}
	})
}

func TestCheckFile(t *testing.T) {
	t.Run("with correct file extensions checkFile shouldn't return any errors", func(t *testing.T) {
		tempSocket := newTempSocket(t)
		defer tempSocket.Close()

		testCases := []struct {
			path string
			ft   fileType
		}{
			{"testdata/logfiles/example.log", log},
			{tempSocket.pathToSocket, socket},
		}

		for _, tc := range testCases {
			err := checkFile(tc.path, tc.ft)
			require.NoError(t, err)
		}
	})
	t.Run("path does not point to the correct file type", func(t *testing.T) {
		tempSocket := newTempSocket(t)
		defer tempSocket.Close()

		testCases := []struct {
			path                  string
			ft                    fileType
			expectedErrorContains string
		}{
			{"testdata/logfiles/example.log", socket, "provided path does not point to a socket file"},
			{tempSocket.pathToSocket, log, "provided path does not point to a log file:"},
		}

		for _, tc := range testCases {
			err := checkFile(tc.path, tc.ft)
			require.ErrorContains(t, err, tc.expectedErrorContains)
		}
	})

	t.Run("with path to non existing file checkFile should return the error", func(t *testing.T) {
		testCases := []struct {
			path                  string
			ft                    fileType
			expectedErrorContains string
		}{
			{"/foo/example.log", log, "provided path does not exist"},
			{"/foo/example.sock", socket, "provided path does not exist"},
		}

		for _, tc := range testCases {
			err := checkFile(tc.path, tc.ft)
			require.ErrorContains(t, err, tc.expectedErrorContains)
		}
	})
}

func TestLogMetricDataToValue(t *testing.T) {
	testCases := []struct {
		metricData    string
		expectedValue int
		err           error
	}{
		{"010", 10, nil},
		{"00", 0, nil},
		{"5", 5, nil},
		{"-010", 0, errors.New("metric can't be negative")},
		{"", 0, fmt.Errorf("invalid syntax")},
		{"0Nax10", 0, fmt.Errorf("invalid syntax")},
	}

	t.Run("check correct returned values", func(t *testing.T) {
		for _, tc := range testCases {
			value, err := logMetricDataToValue(tc.metricData)
			if tc.err != nil {
				require.ErrorContains(t, err, tc.err.Error())
				continue
			}

			require.NoError(t, err)
			require.Equal(t, tc.expectedValue, value)
		}
	})
}
