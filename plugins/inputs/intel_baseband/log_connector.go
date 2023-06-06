package intel_baseband

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	infoLine     = "INFO:"
	countersLine = "counters:"

	deviceStatusStartPrefix = "Device Status::"
	deviceStatusEndPrefix   = "VFs"
)

var errFindingSubstring = errors.New("couldn't find the substring in the log file")

type LogConnector struct {
	// path to log
	path string

	// Num of VFs
	numVFs int

	// Log file data
	lines []string

	lastModTime time.Time
}

type LogMetric struct {
	operationName string
	data          []string
}

// Try to read file and fill the .lines field.
func (lc *LogConnector) readLogFile() error {
	err := lc.checkLogFreshness()
	if err != nil {
		return err
	}
	file, err := os.ReadFile(lc.path)
	if err != nil {
		lc.numVFs = -1
		return fmt.Errorf("couldn't read log file: %w", err)
	}

	// Example content of the metric file is located in testdata/example.log
	// the minimum acceptable file content consists of three lines:
	// - one line for number of VFs
	// - two lines for operation (counters name and metrics value)
	lines := strings.Split(string(file), "\n")
	if len(lines) < 3 {
		return errors.New("log file is empty or incomplete, 'wait_for_telemetry_delay' may have been set to low, try to increase it")
	}

	lc.lines = lines
	return nil
}

// function checks whether the data in the log file were updated by checking the last modification date
func (lc *LogConnector) checkLogFreshness() error {
	fileInfo, err := os.Stat(lc.path)
	if err != nil {
		return fmt.Errorf("couldn't stat log file: %w", err)
	}
	currModData := fileInfo.ModTime()
	if lc.lastModTime == currModData {
		return errors.New("failed to refresh telemetry data: 'wait_for_telemetry_delay' may have been set to low, try to increase it")
	}
	lc.lastModTime = currModData
	return nil
}

// Try to read file and return lines from it
func (lc *LogConnector) getLogLines() []string {
	return lc.lines
}

// Try to read file and return lines from it
func (lc *LogConnector) getLogLinesNum() int {
	return len(lc.lines)
}

// Return the number of VFs in the log file
func (lc *LogConnector) getNumVFs() int {
	return lc.numVFs
}

// find a line which contains Device Status. Example = Thu Apr 13 13:28:40 2023:INFO:Device Status:: 2 VFs
func (lc *LogConnector) readNumVFs() error {
	for _, line := range lc.lines {
		if !strings.Contains(line, deviceStatusStartPrefix) {
			continue
		}

		numVFs, err := lc.parseNumVFs(line)
		if err != nil {
			lc.numVFs = -1
			return err
		}
		lc.numVFs = numVFs
		return nil
	}

	return fmt.Errorf("numVFs data wasn't found in the log file")
}

// Find a line which contains a substring in the log file
func (lc *LogConnector) getSubstringLine(offsetLine int, substring string) (int, string, error) {
	if len(substring) == 0 {
		return 0, "", fmt.Errorf("substring is empty")
	}

	for i := offsetLine; i < len(lc.lines); i++ {
		if !strings.Contains(lc.lines[i], substring) {
			continue
		}

		return i, lc.lines[i], nil
	}
	return 0, "", fmt.Errorf("%q: %w", substring, errFindingSubstring)
}

func (lc *LogConnector) getMetrics(name string) (metrics []*LogMetric, err error) {
	currOffset, offset := 0, 0
	for {
		var metric *LogMetric
		currOffset, metric, err = lc.getMetric(offset, name)
		if err != nil {
			if errors.Is(err, errFindingSubstring) {
				break
			}
			return nil, err
		}
		metrics = append(metrics, metric)
		offset = currOffset
	}

	if len(metrics) == 0 {
		return nil, err
	}

	return metrics, nil
}

// Example of log file:
// Thu May 18 08:45:15 2023:INFO:5GUL counters: Code Blocks
// Thu May 18 08:45:15 2023:INFO:0 0
// Input: offsetLine, metric name (Code Blocks)
// Func will return: current offset after reading the metric (2), metric with operation name and data(5GUL, ["0", "0"]) and error
func (lc *LogConnector) getMetric(offsetLine int, name string) (int, *LogMetric, error) {
	i, line, err := lc.getSubstringLine(offsetLine, name)
	if err != nil {
		return offsetLine, nil, err
	}

	operationName := lc.parseOperationName(line)
	if len(operationName) == 0 {
		return offsetLine, nil, errors.New("valid operation name wasn't found in log")
	}

	if lc.getLogLinesNum() <= i+1 {
		return offsetLine, nil,
			fmt.Errorf("the content of the log file is incorrect, line which contains key word %q can't be the last one in log", countersLine)
	}

	// infoData eg: Thu Apr 13 13:28:40 2023:INFO:12 0
	infoData := strings.Split(lc.lines[i+1], infoLine)
	if len(infoData) != 2 {
		//info data must be in format : some data + keyword "INFO:" + metrics
		return offsetLine, nil, fmt.Errorf("the content of the log file is incorrect, couldn't find %q separator", infoLine)
	}

	dataRaw := strings.TrimSpace(infoData[1])
	if len(dataRaw) == 0 {
		return offsetLine, nil, fmt.Errorf("the content of the log file is incorrect, metric's data is incorrect")
	}

	data := strings.Split(dataRaw, " ")
	for i := range data {
		if len(data[i]) == 0 {
			return offsetLine, nil, fmt.Errorf("the content of the log file is incorrect, metric's data is empty")
		}
	}
	return i + 2, &LogMetric{operationName: operationName, data: data}, nil
}

// Example value = Thu Apr 13 13:28:40 2023:INFO:Device Status:: 2 VFs
func (lc *LogConnector) parseNumVFs(s string) (int, error) {
	i := strings.LastIndex(s, deviceStatusStartPrefix)
	if i == -1 {
		return 0, fmt.Errorf("couldn't find device status prefix in line")
	}

	j := strings.Index(s[i:], deviceStatusEndPrefix)
	if j == -1 {
		return 0, fmt.Errorf("couldn't find device end prefix in line")
	}

	startIndex := i + len(deviceStatusStartPrefix) + 1
	endIndex := i + j - 1
	if len(s) < startIndex || startIndex >= endIndex {
		return 0, fmt.Errorf("incorrect format of the line")
	}

	return strconv.Atoi(s[startIndex:endIndex])
}

// Parse Operation name
// Example = Thu Apr 13 13:28:40 2023:INFO:5GUL counters: Code Blocks
// Output: 5GUL
func (lc *LogConnector) parseOperationName(s string) string {
	i := strings.Index(s, infoLine)
	if i >= 0 {
		j := strings.Index(s[i:], countersLine)
		startIndex := i + len(infoLine)
		endIndex := i + j - 1
		if j >= 0 && startIndex < endIndex {
			return s[startIndex:endIndex]
		}
	}
	return ""
}

func newLogConnector(path string) *LogConnector {
	lastModTime := time.Time{}
	fileInfo, err := os.Stat(path)
	if err == nil {
		lastModTime = fileInfo.ModTime()
	}

	return &LogConnector{
		path:        path,
		numVFs:      -1,
		lastModTime: lastModTime,
	}
}
