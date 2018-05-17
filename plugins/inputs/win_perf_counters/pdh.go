// Go API over pdh syscalls
// +build windows

package win_perf_counters

import (
	"syscall"
	"unsafe"
	"errors"
)

// OpenPerformanceCountersQuery creates a new query that is used to manage the collection of performance data.
// It returns query handle used for subsequent calls for adding counters and querying data
func OpenPerformanceCountersQuery() (PDH_HQUERY, error) {
	var handle PDH_HQUERY
	ret := PdhOpenQuery(0, 0, &handle )
	if ret != ERROR_SUCCESS {
		return 0, errors.New(PdhFormatError(ret))
	}
	return handle, nil
}
// UTF16PtrToString converts Windows API LPTSTR (pointer to string) to go string
func UTF16PtrToString(s *uint16) string {
	if s == nil {
		return ""
	}
	return syscall.UTF16ToString((*[1 << 29]uint16)(unsafe.Pointer(s))[0:])
}
// UTF16ToStringArray converts list of Windows API NULL terminated strings  to go string array
func UTF16ToStringArray(buf []uint16) []string {
	var strings []string
	nextLineStart := 0
	stringLine := UTF16PtrToString(&buf[0])
	for stringLine != "" {
		strings = append(strings, stringLine)
		nextLineStart += len(stringLine) + 1
		remainingBuf := buf[nextLineStart:]
		stringLine = UTF16PtrToString(&remainingBuf[0])
	}
	return strings
}

//GetCounterInfo return counter information for given handle
func GetCounterInfo(counterHandle PDH_HCOUNTER) (*PDH_COUNTER_INFO, error) {
	var bufSize uint32
	var buff []byte
	ret := PdhGetCounterInfo(counterHandle, 0, &bufSize, nil)

	if ret == PDH_MORE_DATA {
		buff = make([]byte, bufSize)
		bufSize = uint32(len(buff))
		ret = PdhGetCounterInfo(counterHandle, 0, &bufSize, &buff[0])
		if ret == ERROR_SUCCESS {
			ci := (*PDH_COUNTER_INFO)(unsafe.Pointer(&buff[0]))
			return ci, nil
		}
	}
	return nil, errors.New(PdhFormatError(ret))
}

// ExpandWildCardPath  examines local computer and returns those counter paths that match the given counter path which contains wildcard characters.
func ExpandWildCardPath(counterPath string) ([]string, error) {
	var bufSize uint32
	var buff []uint16
	ret := PdhExpandWildCardPath(counterPath, nil, &bufSize)

	if ret == PDH_MORE_DATA {
		buff = make([]uint16, bufSize)
		bufSize = uint32(len(buff))
		ret = PdhExpandWildCardPath(counterPath, &buff[0], &bufSize)
		if ret == ERROR_SUCCESS {
			list := UTF16ToStringArray(buff)
			return list, nil
		}
	}
	return nil, errors.New(PdhFormatError(ret))
}

//GetFormattedCounterValueDouble computes a displayable value for the specified counter
func GetFormattedCounterValueDouble(hCounter PDH_HCOUNTER) (*PDH_FMT_COUNTERVALUE_DOUBLE, error) {
	var counterType uint32
	var value PDH_FMT_COUNTERVALUE_DOUBLE
	ret := PdhGetFormattedCounterValueDouble(hCounter, &counterType, &value)
	if ret == ERROR_SUCCESS {
		return &value, nil
	} else {
		return nil, errors.New(PdhFormatError(ret))
	}
}
