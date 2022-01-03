//go:build linux
// +build linux

package intel_powerstat

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// fileService is responsible for handling operations on files.
type fileService interface {
	getCPUInfoStats() (map[string]*cpuInfo, error)
	getStringsMatchingPatternOnPath(path string) ([]string, error)
	readFile(path string) ([]byte, error)
	readFileToFloat64(reader io.Reader) (float64, int64, error)
	readFileAtOffsetToUint64(reader io.ReaderAt, offset int64) (uint64, error)
}

type fileServiceImpl struct {
}

// getCPUInfoStats retrieves basic information about CPU from /proc/cpuinfo.
func (fs *fileServiceImpl) getCPUInfoStats() (map[string]*cpuInfo, error) {
	path := "/proc/cpuinfo"
	cpuInfoFile, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error while reading %s, err: %v", path, err)
	}
	defer cpuInfoFile.Close()

	scanner := bufio.NewScanner(cpuInfoFile)

	processorRegexp := regexp.MustCompile(`^processor\t+:\s([0-9]+)\n*$`)
	physicalIDRegexp := regexp.MustCompile(`^physical id\t+:\s([0-9]+)\n*$`)
	coreIDRegexp := regexp.MustCompile(`^core id\t+:\s([0-9]+)\n*$`)
	vendorIDRegexp := regexp.MustCompile(`^vendor_id\t+:\s([a-zA-Z]+)\n*$`)
	cpuFamilyRegexp := regexp.MustCompile(`^cpu\sfamily\t+:\s([0-9]+)\n*$`)
	modelRegexp := regexp.MustCompile(`^model\t+:\s([0-9]+)\n*$`)
	flagsRegexp := regexp.MustCompile(`^flags\t+:\s(.+)\n*$`)

	stats := make(map[string]*cpuInfo)
	currentInfo := &cpuInfo{}

	for scanner.Scan() {
		line := scanner.Text()

		processorRes := processorRegexp.FindStringSubmatch(line)
		if len(processorRes) > 1 {
			currentInfo = &cpuInfo{
				cpuID: processorRes[1],
			}
		}

		vendorIDRes := vendorIDRegexp.FindStringSubmatch(line)
		if len(vendorIDRes) > 1 {
			currentInfo.vendorID = vendorIDRes[1]
		}

		physicalIDRes := physicalIDRegexp.FindStringSubmatch(line)
		if len(physicalIDRes) > 1 {
			currentInfo.physicalID = physicalIDRes[1]
		}

		coreIDRes := coreIDRegexp.FindStringSubmatch(line)
		if len(coreIDRes) > 1 {
			currentInfo.coreID = coreIDRes[1]
		}

		cpuFamilyRes := cpuFamilyRegexp.FindStringSubmatch(line)
		if len(cpuFamilyRes) > 1 {
			currentInfo.cpuFamily = cpuFamilyRes[1]
		}

		modelRes := modelRegexp.FindStringSubmatch(line)
		if len(modelRes) > 1 {
			currentInfo.model = modelRes[1]
		}

		flagsRes := flagsRegexp.FindStringSubmatch(line)
		if len(flagsRes) > 1 {
			currentInfo.flags = flagsRes[1]

			// Flags is the last value we have to acquire, so currentInfo is added to map.
			stats[currentInfo.cpuID] = currentInfo
		}
	}

	return stats, nil
}

// getStringsMatchingPatternOnPath looks for filenames and directory names on path matching given regexp.
// It ignores file system errors such as I/O errors reading directories. The only possible returned error
// is ErrBadPattern, when pattern is malformed.
func (fs *fileServiceImpl) getStringsMatchingPatternOnPath(path string) ([]string, error) {
	return filepath.Glob(path)
}

// readFile reads file on path and return string content.
func (fs *fileServiceImpl) readFile(path string) ([]byte, error) {
	out, err := os.ReadFile(path)
	if err != nil {
		return make([]byte, 0), err
	}
	return out, nil
}

// readFileToFloat64 reads file on path and tries to parse content to float64.
func (fs *fileServiceImpl) readFileToFloat64(reader io.Reader) (float64, int64, error) {
	read, err := io.ReadAll(reader)
	if err != nil {
		return 0, 0, err
	}

	readDate := time.Now().UnixNano()

	// Remove new line character
	trimmedString := strings.TrimRight(string(read), "\n")
	// Parse result to float64
	parsedValue, err := strconv.ParseFloat(trimmedString, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("error parsing string to float for %s", trimmedString)
	}

	return parsedValue, readDate, nil
}

// readFileAtOffsetToUint64 reads 8 bytes from passed file at given offset.
func (fs *fileServiceImpl) readFileAtOffsetToUint64(reader io.ReaderAt, offset int64) (uint64, error) {
	buffer := make([]byte, 8)

	if offset == 0 {
		return 0, fmt.Errorf("file offset %d should not be 0", offset)
	}

	_, err := reader.ReadAt(buffer, offset)
	if err != nil {
		return 0, fmt.Errorf("error on reading file at offset %d, err: %v", offset, err)
	}

	return binary.LittleEndian.Uint64(buffer), nil
}

func newFileService() *fileServiceImpl {
	return &fileServiceImpl{}
}
