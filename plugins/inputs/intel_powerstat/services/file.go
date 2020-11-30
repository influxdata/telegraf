// +build linux

package services

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf/plugins/inputs/intel_powerstat/data"
)

// FileService is responsible for handling operations on files.
type FileService interface {
	GetStringsMatchingPatternOnPath(path string) ([]string, error)
	ReadFile(path string) ([]byte, error)
	ReadFileAtOffsetToUint64(reader io.ReaderAt, offset int64) (uint64, error)
	ReadFileToFloat64(reader io.Reader) (float64, int64, error)
	GetCPUInfoStats() (map[string]*data.CPUInfo, error)
}

// FileServiceImpl is implementation of FileService.
type FileServiceImpl struct {
}

// ReadFileAtOffsetToUint64 reads 8 bytes from passed file at given offset.
func (fs *FileServiceImpl) ReadFileAtOffsetToUint64(reader io.ReaderAt, offset int64) (uint64, error) {
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

// GetStringsMatchingPatternOnPath looks for filenames and directory names on path matching given regexp.
func (fs *FileServiceImpl) GetStringsMatchingPatternOnPath(path string) ([]string, error) {
	return filepath.Glob(path)
}

// GetCPUInfoStats retrieves basic information about CPU from /proc/cpuinfo.
func (fs *FileServiceImpl) GetCPUInfoStats() (map[string]*data.CPUInfo, error) {
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

	stats := make(map[string]*data.CPUInfo)
	currentInfo := &data.CPUInfo{}

	for scanner.Scan() {
		line := scanner.Text()

		processorRes := processorRegexp.FindStringSubmatch(line)
		if len(processorRes) > 1 {
			currentInfo = &data.CPUInfo{
				CPUID: processorRes[1],
			}
		}

		vendorIDRes := vendorIDRegexp.FindStringSubmatch(line)
		if len(vendorIDRes) > 1 {
			currentInfo.VendorID = vendorIDRes[1]
		}

		physicalIDRes := physicalIDRegexp.FindStringSubmatch(line)
		if len(physicalIDRes) > 1 {
			currentInfo.PhysicalID = physicalIDRes[1]
		}

		coreIDRes := coreIDRegexp.FindStringSubmatch(line)
		if len(coreIDRes) > 1 {
			currentInfo.CoreID = coreIDRes[1]
		}

		cpuFamilyRes := cpuFamilyRegexp.FindStringSubmatch(line)
		if len(cpuFamilyRes) > 1 {
			currentInfo.CPUFamily = cpuFamilyRes[1]
		}

		modelRes := modelRegexp.FindStringSubmatch(line)
		if len(modelRes) > 1 {
			currentInfo.Model = modelRes[1]
		}

		flagsRes := flagsRegexp.FindStringSubmatch(line)
		if len(flagsRes) > 1 {
			currentInfo.Flags = flagsRes[1]

			// Flags is the last value we have to acquire, so currentInfo is added to map.
			stats[currentInfo.CPUID] = currentInfo
		}
	}

	return stats, nil
}

// ReadFile reads file on path and return string content.
func (fs *FileServiceImpl) ReadFile(path string) ([]byte, error) {
	out, err := ioutil.ReadFile(path)
	if err != nil {
		return make([]byte, 0), err
	}
	return out, nil
}

// ReadFileToFloat64 reads file on path and tries to parse content to float.
func (fs *FileServiceImpl) ReadFileToFloat64(reader io.Reader) (float64, int64, error) {
	read, err := ioutil.ReadAll(reader)
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

// NewFileService returns new FileServiceImpl struct.
func NewFileService() *FileServiceImpl {
	return &FileServiceImpl{}
}
