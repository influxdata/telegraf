//go:build linux && amd64

package intel_baseband

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"strconv"
	"strings"
)

type fileType int

const (
	log fileType = iota
	socket
)

func validatePath(pathToRead string, ft fileType) (string, error) {
	if pathToRead == "" {
		return "", errors.New("required path not specified")
	}
	cleanPath := path.Clean(pathToRead)
	if (ft == log && path.Ext(cleanPath) != logFileExtension) || (ft == socket && path.Ext(cleanPath) != socketExtension) {
		return "", fmt.Errorf("wrong file extension: %q", cleanPath)
	}
	if !path.IsAbs(cleanPath) {
		return "", fmt.Errorf("path is not absolute %q", cleanPath)
	}
	return cleanPath, nil
}

func checkFile(pathToFile string, fileType fileType) error {
	pathInfo, err := os.Lstat(pathToFile)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("provided path does not exist: %q", pathToFile)
		}
		if errors.Is(err, fs.ErrPermission) {
			return fmt.Errorf("user doesn't have enough privileges to file %q", pathToFile)
		}

		return fmt.Errorf("couldn't get system information of file %q: %w", pathToFile, err)
	}

	mode := pathInfo.Mode()
	switch fileType {
	case socket:
		if mode&os.ModeSocket != os.ModeSocket {
			return fmt.Errorf("provided path does not point to a socket file: %q", pathToFile)
		}
	case log:
		if !(mode.IsRegular()) {
			return fmt.Errorf("provided path does not point to a log file: %q", pathToFile)
		}
	}
	return nil
}

// Replace metric name to snake case
// Example: Code Blocks -> code_blocks
func metricNameToTagName(metricName string) string {
	cleanedStr := strings.Replace(strings.Replace(strings.Replace(metricName, "(", "", -1), ")", "", -1), " ", "_", -1)
	return strings.ToLower(cleanedStr)
}

func logMetricDataToValue(data string) (int, error) {
	value, err := strconv.Atoi(data)
	if err != nil {
		return 0, err
	}

	if value < 0 {
		return 0, fmt.Errorf("metric can't be negative")
	}

	return value, nil
}
