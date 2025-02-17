// Go API over pdh syscalls
//go:build windows

package win_perf_counters

import (
	"errors"
	"syscall"
	"time"
	"unsafe"
)

// Initial buffer size for return buffers
const initialBufferSize = uint32(1024) // 1kB

var errBufferLimitReached = errors.New("buffer limit reached")

// counterValue is abstraction for pdhFmtCountervalueItemDouble
type counterValue struct {
	instanceName string
	value        interface{}
}

// performanceQuery provides wrappers around Windows performance counters API for easy usage in GO
//
//nolint:interfacebloat // conditionally allow to contain more methods
type performanceQuery interface {
	open() error
	close() error
	addCounterToQuery(counterPath string) (pdhCounterHandle, error)
	addEnglishCounterToQuery(counterPath string) (pdhCounterHandle, error)
	getCounterPath(counterHandle pdhCounterHandle) (string, error)
	expandWildCardPath(counterPath string) ([]string, error)
	getFormattedCounterValueDouble(hCounter pdhCounterHandle) (float64, error)
	getRawCounterValue(hCounter pdhCounterHandle) (int64, error)
	getFormattedCounterArrayDouble(hCounter pdhCounterHandle) ([]counterValue, error)
	getRawCounterArray(hCounter pdhCounterHandle) ([]counterValue, error)
	collectData() error
	collectDataWithTime() (time.Time, error)
	isVistaOrNewer() bool
}

type performanceQueryCreator interface {
	newPerformanceQuery(string, uint32) performanceQuery
}

// pdhError represents error returned from Performance Counters API
type pdhError struct {
	errorCode uint32
	errorText string
}

func (m *pdhError) Error() string {
	return m.errorText
}

func newPdhError(code uint32) error {
	return &pdhError{
		errorCode: code,
		errorText: pdhFormatError(code),
	}
}

// performanceQueryImpl is implementation of performanceQuery interface, which calls phd.dll functions
type performanceQueryImpl struct {
	maxBufferSize uint32
	query         pdhQueryHandle
}

type performanceQueryCreatorImpl struct{}

func (performanceQueryCreatorImpl) newPerformanceQuery(_ string, maxBufferSize uint32) performanceQuery {
	return &performanceQueryImpl{maxBufferSize: maxBufferSize}
}

// open creates a new counterPath that is used to manage the collection of performance data.
// It returns counterPath handle used for subsequent calls for adding counters and querying data
func (m *performanceQueryImpl) open() error {
	if m.query != 0 {
		err := m.close()
		if err != nil {
			return err
		}
	}
	var handle pdhQueryHandle

	if ret := pdhOpenQuery(0, 0, &handle); ret != errorSuccess {
		return newPdhError(ret)
	}
	m.query = handle
	return nil
}

// close closes the counterPath, releases associated counter handles and frees resources
func (m *performanceQueryImpl) close() error {
	if m.query == 0 {
		return errors.New("uninitialized query")
	}

	if ret := pdhCloseQuery(m.query); ret != errorSuccess {
		return newPdhError(ret)
	}
	m.query = 0
	return nil
}

func (m *performanceQueryImpl) addCounterToQuery(counterPath string) (pdhCounterHandle, error) {
	var counterHandle pdhCounterHandle
	if m.query == 0 {
		return 0, errors.New("uninitialized query")
	}

	if ret := pdhAddCounter(m.query, counterPath, 0, &counterHandle); ret != errorSuccess {
		return 0, newPdhError(ret)
	}
	return counterHandle, nil
}

func (m *performanceQueryImpl) addEnglishCounterToQuery(counterPath string) (pdhCounterHandle, error) {
	var counterHandle pdhCounterHandle
	if m.query == 0 {
		return 0, errors.New("uninitialized query")
	}
	if ret := pdhAddEnglishCounter(m.query, counterPath, 0, &counterHandle); ret != errorSuccess {
		return 0, newPdhError(ret)
	}
	return counterHandle, nil
}

// getCounterPath returns counter information for given handle
func (m *performanceQueryImpl) getCounterPath(counterHandle pdhCounterHandle) (string, error) {
	for buflen := initialBufferSize; buflen <= m.maxBufferSize; buflen *= 2 {
		buf := make([]byte, buflen)

		// Get the info with the current buffer size
		size := buflen
		ret := pdhGetCounterInfo(counterHandle, 0, &size, &buf[0])
		if ret == errorSuccess {
			ci := (*pdhCounterInfo)(unsafe.Pointer(&buf[0])) //nolint:gosec // G103: Valid use of unsafe call to create PDH_COUNTER_INFO
			return utf16PtrToString(ci.SzFullPath), nil
		}

		// Use the size as a hint if it exceeds the current buffer size
		if size > buflen {
			buflen = size
		}

		// We got a non-recoverable error so exit here
		if ret != pdhMoreData {
			return "", newPdhError(ret)
		}
	}

	return "", errBufferLimitReached
}

// expandWildCardPath examines local computer and returns those counter paths that match the given counter path which contains wildcard characters.
func (m *performanceQueryImpl) expandWildCardPath(counterPath string) ([]string, error) {
	for buflen := initialBufferSize; buflen <= m.maxBufferSize; buflen *= 2 {
		buf := make([]uint16, buflen)

		// Get the info with the current buffer size
		size := buflen
		ret := pdhExpandWildCardPath(counterPath, &buf[0], &size)
		if ret == errorSuccess {
			return utf16ToStringArray(buf), nil
		}

		// Use the size as a hint if it exceeds the current buffer size
		if size > buflen {
			buflen = size
		}

		// We got a non-recoverable error so exit here
		if ret != pdhMoreData {
			return nil, newPdhError(ret)
		}
	}

	return nil, errBufferLimitReached
}

// getFormattedCounterValueDouble computes a displayable value for the specified counter
func (*performanceQueryImpl) getFormattedCounterValueDouble(hCounter pdhCounterHandle) (float64, error) {
	var counterType uint32
	var value pdhFmtCountervalueDouble

	if ret := pdhGetFormattedCounterValueDouble(hCounter, &counterType, &value); ret != errorSuccess {
		return 0, newPdhError(ret)
	}
	if value.CStatus == pdhCstatusValidData || value.CStatus == pdhCstatusNewData {
		return value.DoubleValue, nil
	}
	return 0, newPdhError(value.CStatus)
}

func (m *performanceQueryImpl) getFormattedCounterArrayDouble(hCounter pdhCounterHandle) ([]counterValue, error) {
	for buflen := initialBufferSize; buflen <= m.maxBufferSize; buflen *= 2 {
		buf := make([]byte, buflen)

		// Get the info with the current buffer size
		var itemCount uint32
		size := buflen
		ret := pdhGetFormattedCounterArrayDouble(hCounter, &size, &itemCount, &buf[0])
		if ret == errorSuccess {
			//nolint:gosec // G103: Valid use of unsafe call to create PDH_FMT_COUNTERVALUE_ITEM_DOUBLE
			items := (*[1 << 20]pdhFmtCountervalueItemDouble)(unsafe.Pointer(&buf[0]))[:itemCount]
			values := make([]counterValue, 0, itemCount)
			for _, item := range items {
				if item.FmtValue.CStatus == pdhCstatusValidData || item.FmtValue.CStatus == pdhCstatusNewData {
					val := counterValue{utf16PtrToString(item.SzName), item.FmtValue.DoubleValue}
					values = append(values, val)
				}
			}
			return values, nil
		}

		// Use the size as a hint if it exceeds the current buffer size
		if size > buflen {
			buflen = size
		}

		// We got a non-recoverable error so exit here
		if ret != pdhMoreData {
			return nil, newPdhError(ret)
		}
	}

	return nil, errBufferLimitReached
}

func (m *performanceQueryImpl) getRawCounterArray(hCounter pdhCounterHandle) ([]counterValue, error) {
	for buflen := initialBufferSize; buflen <= m.maxBufferSize; buflen *= 2 {
		buf := make([]byte, buflen)

		// Get the info with the current buffer size
		var itemCount uint32
		size := buflen
		ret := pdhGetRawCounterArray(hCounter, &size, &itemCount, &buf[0])
		if ret == errorSuccess {
			//nolint:gosec // G103: Valid use of unsafe call to create PDH_RAW_COUNTER_ITEM
			items := (*[1 << 20]pdhRawCounterItem)(unsafe.Pointer(&buf[0]))[:itemCount]
			values := make([]counterValue, 0, itemCount)
			for _, item := range items {
				if item.RawValue.CStatus == pdhCstatusValidData || item.RawValue.CStatus == pdhCstatusNewData {
					val := counterValue{utf16PtrToString(item.SzName), item.RawValue.FirstValue}
					values = append(values, val)
				}
			}
			return values, nil
		}

		// Use the size as a hint if it exceeds the current buffer size
		if size > buflen {
			buflen = size
		}

		// We got a non-recoverable error so exit here
		if ret != pdhMoreData {
			return nil, newPdhError(ret)
		}
	}

	return nil, errBufferLimitReached
}

func (m *performanceQueryImpl) collectData() error {
	var ret uint32
	if m.query == 0 {
		return errors.New("uninitialized query")
	}

	if ret = pdhCollectQueryData(m.query); ret != errorSuccess {
		return newPdhError(ret)
	}
	return nil
}

func (m *performanceQueryImpl) collectDataWithTime() (time.Time, error) {
	if m.query == 0 {
		return time.Now(), errors.New("uninitialized query")
	}
	ret, mtime := pdhCollectQueryDataWithTime(m.query)
	if ret != errorSuccess {
		return time.Now(), newPdhError(ret)
	}
	return mtime, nil
}

func (*performanceQueryImpl) isVistaOrNewer() bool {
	return pdhAddEnglishCounterSupported()
}

func (m *performanceQueryImpl) getRawCounterValue(hCounter pdhCounterHandle) (int64, error) {
	if m.query == 0 {
		return 0, errors.New("uninitialised query")
	}

	var counterType uint32
	var value pdhRawCounter
	var ret uint32

	if ret = pdhGetRawCounterValue(hCounter, &counterType, &value); ret == errorSuccess {
		if value.CStatus == pdhCstatusValidData || value.CStatus == pdhCstatusNewData {
			return value.FirstValue, nil
		}
		return 0, newPdhError(value.CStatus)
	}
	return 0, newPdhError(ret)
}

// utf16PtrToString converts Windows API LPTSTR (pointer to string) to go string
func utf16PtrToString(s *uint16) string {
	if s == nil {
		return ""
	}
	//nolint:gosec // G103: Valid use of unsafe call to create string from Windows API LPTSTR (pointer to string)
	return syscall.UTF16ToString((*[1 << 29]uint16)(unsafe.Pointer(s))[0:])
}

// utf16ToStringArray converts list of Windows API NULL terminated strings  to go string array
func utf16ToStringArray(buf []uint16) []string {
	var strings []string
	nextLineStart := 0
	stringLine := utf16PtrToString(&buf[0])
	for stringLine != "" {
		strings = append(strings, stringLine)
		nextLineStart += len([]rune(stringLine)) + 1
		remainingBuf := buf[nextLineStart:]
		stringLine = utf16PtrToString(&remainingBuf[0])
	}
	return strings
}
