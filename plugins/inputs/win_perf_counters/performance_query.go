// Go API over pdh syscalls
// +build windows

package win_perf_counters

import (
	"errors"
	"syscall"
	"unsafe"
)

//PerformanceQuery is abstraction for PDH_FMT_COUNTERVALUE_ITEM_DOUBLE
type CounterValue struct {
	InstanceName string
	Value        float64
}

//PerformanceQuery provides wrappers around Windows performance counters API for easy usage in GO
type PerformanceQuery interface {
	Open() error
	Close() error
	AddCounterToQuery(counterPath string) (PDH_HCOUNTER, error)
	AddEnglishCounterToQuery(counterPath string) (PDH_HCOUNTER, error)
	GetCounterPath(counterHandle PDH_HCOUNTER) (string, error)
	ExpandWildCardPath(counterPath string) ([]string, error)
	GetFormattedCounterValueDouble(hCounter PDH_HCOUNTER) (float64, error)
	GetFormattedCounterArrayDouble(hCounter PDH_HCOUNTER) ([]CounterValue, error)
	CollectData() error
	AddEnglishCounterSupported() bool
}

//PdhError represents error returned from Performance Counters API
type PdhError struct {
	ErrorCode uint32
	errorText string
}

func (m *PdhError) Error() string {
	return m.errorText
}

func NewPdhError(code uint32) error {
	return &PdhError{
		ErrorCode: code,
		errorText: PdhFormatError(code),
	}
}

//PerformanceQueryImpl is implementation of PerformanceQuery interface, which calls phd.dll functions
type PerformanceQueryImpl struct {
	query PDH_HQUERY
}

// Open creates a new counterPath that is used to manage the collection of performance data.
// It returns counterPath handle used for subsequent calls for adding counters and querying data
func (m *PerformanceQueryImpl) Open() error {
	if m.query != 0 {
		err := m.Close()
		if err != nil {
			return err
		}
	}
	var handle PDH_HQUERY
	ret := PdhOpenQuery(0, 0, &handle)
	if ret != ERROR_SUCCESS {
		return NewPdhError(ret)
	}
	m.query = handle
	return nil
}

// Close closes the counterPath, releases associated counter handles and frees resources
func (m *PerformanceQueryImpl) Close() error {
	if m.query == 0 {
		return errors.New("uninitialised query")
	}
	ret := PdhCloseQuery(m.query)
	if ret != ERROR_SUCCESS {
		return NewPdhError(ret)
	}
	m.query = 0
	return nil
}

func (m *PerformanceQueryImpl) AddCounterToQuery(counterPath string) (PDH_HCOUNTER, error) {
	var counterHandle PDH_HCOUNTER
	if m.query == 0 {
		return 0, errors.New("uninitialised query")
	}
	ret := PdhAddCounter(m.query, counterPath, 0, &counterHandle)
	if ret != ERROR_SUCCESS {
		return 0, NewPdhError(ret)
	}
	return counterHandle, nil
}

func (m *PerformanceQueryImpl) AddEnglishCounterToQuery(counterPath string) (PDH_HCOUNTER, error) {
	var counterHandle PDH_HCOUNTER
	if m.query == 0 {
		return 0, errors.New("uninitialised query")
	}
	ret := PdhAddEnglishCounter(m.query, counterPath, 0, &counterHandle)
	if ret != ERROR_SUCCESS {
		return 0, NewPdhError(ret)
	}
	return counterHandle, nil
}

//GetCounterPath return counter information for given handle
func (m *PerformanceQueryImpl) GetCounterPath(counterHandle PDH_HCOUNTER) (string, error) {
	var bufSize uint32
	var buff []byte

	ret := PdhGetCounterInfo(counterHandle, 0, &bufSize, nil)
	if ret == PDH_MORE_DATA {
		buff = make([]byte, bufSize)
		bufSize = uint32(len(buff))
		ret = PdhGetCounterInfo(counterHandle, 0, &bufSize, &buff[0])
		if ret == ERROR_SUCCESS {
			ci := (*PDH_COUNTER_INFO)(unsafe.Pointer(&buff[0]))
			return UTF16PtrToString(ci.SzFullPath), nil
		}
	}
	return "", NewPdhError(ret)
}

// ExpandWildCardPath  examines local computer and returns those counter paths that match the given counter path which contains wildcard characters.
func (m *PerformanceQueryImpl) ExpandWildCardPath(counterPath string) ([]string, error) {
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
	return nil, NewPdhError(ret)
}

//GetFormattedCounterValueDouble computes a displayable value for the specified counter
func (m *PerformanceQueryImpl) GetFormattedCounterValueDouble(hCounter PDH_HCOUNTER) (float64, error) {
	var counterType uint32
	var value PDH_FMT_COUNTERVALUE_DOUBLE
	ret := PdhGetFormattedCounterValueDouble(hCounter, &counterType, &value)
	if ret == ERROR_SUCCESS {
		if value.CStatus == PDH_CSTATUS_VALID_DATA || value.CStatus == PDH_CSTATUS_NEW_DATA {
			return value.DoubleValue, nil
		} else {
			return 0, NewPdhError(value.CStatus)
		}
	} else {
		return 0, NewPdhError(ret)
	}
}

func (m *PerformanceQueryImpl) GetFormattedCounterArrayDouble(hCounter PDH_HCOUNTER) ([]CounterValue, error) {
	var buffSize uint32
	var itemCount uint32
	ret := PdhGetFormattedCounterArrayDouble(hCounter, &buffSize, &itemCount, nil)
	if ret == PDH_MORE_DATA {
		buff := make([]byte, buffSize)
		ret = PdhGetFormattedCounterArrayDouble(hCounter, &buffSize, &itemCount, &buff[0])
		if ret == ERROR_SUCCESS {
			items := (*[1 << 20]PDH_FMT_COUNTERVALUE_ITEM_DOUBLE)(unsafe.Pointer(&buff[0]))[:itemCount]
			values := make([]CounterValue, 0, itemCount)
			for _, item := range items {
				if item.FmtValue.CStatus == PDH_CSTATUS_VALID_DATA || item.FmtValue.CStatus == PDH_CSTATUS_NEW_DATA {
					val := CounterValue{UTF16PtrToString(item.SzName), item.FmtValue.DoubleValue}
					values = append(values, val)
				}
			}
			return values, nil
		}
	}
	return nil, NewPdhError(ret)
}

func (m *PerformanceQueryImpl) CollectData() error {
	if m.query == 0 {
		return errors.New("uninitialised query")
	}
	ret := PdhCollectQueryData(m.query)
	if ret != ERROR_SUCCESS {
		return NewPdhError(ret)
	}
	return nil
}

func (m *PerformanceQueryImpl) AddEnglishCounterSupported() bool {
	return PdhAddEnglishCounterSupported()
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
		nextLineStart += len([]rune(stringLine)) + 1
		remainingBuf := buf[nextLineStart:]
		stringLine = UTF16PtrToString(&remainingBuf[0])
	}
	return strings
}
