// Copyright (c) 2010 The win Authors. All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions
// are met:
// 1. Redistributions of source code must retain the above copyright
//    notice, this list of conditions and the following disclaimer.
// 2. Redistributions in binary form must reproduce the above copyright
//    notice, this list of conditions and the following disclaimer in the
//    documentation and/or other materials provided with the distribution.
// 3. The names of the authors may not be used to endorse or promote products
//    derived from this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE AUTHORS ``AS IS'' AND ANY EXPRESS OR
// IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES
// OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED.
// IN NO EVENT SHALL THE AUTHORS BE LIABLE FOR ANY DIRECT, INDIRECT,
// INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT
// NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF
// THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
//
// This is the official list of 'win' authors for copyright purposes.
//
// Alexander Neumann <an2048@googlemail.com>
// Joseph Watson <jtwatson@linux-consulting.us>
// Kevin Pors <krpors@gmail.com>

//go:build windows

package win_perf_counters

import (
	"fmt"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Error codes
const (
	errorSuccess                = 0
	errorFailure                = 1
	errorInvalidFunction        = 1
	epochDifferenceMicros int64 = 11644473600000000
)

type (
	handle uintptr
)

// PDH error codes, which can be returned by all Pdh* functions. Taken from mingw-w64 pdhmsg.h
const (
	pdhCstatusValidData                   = 0x00000000 // The returned data is valid.
	pdhCstatusNewData                     = 0x00000001 // The return data value is valid and different from the last sample.
	pdhCstatusNoMachine                   = 0x800007D0 // Unable to connect to the specified computer, or the computer is offline.
	pdhCstatusNoInstance                  = 0x800007D1
	pdhMoreData                           = 0x800007D2 // The pdhGetFormattedCounterArray* function can return this if there's 'more data to be displayed'.
	pdhCstatusItemNotValidated            = 0x800007D3
	pdhRetry                              = 0x800007D4
	pdhNoData                             = 0x800007D5 // The query does not currently contain any counters (for example, limited access)
	pdhCalcNegativeDenominator            = 0x800007D6
	pdhCalcNegativeTimebase               = 0x800007D7
	pdhCalcNegativeValue                  = 0x800007D8
	pdhDialogCancelled                    = 0x800007D9
	pdhEndOfLogFile                       = 0x800007DA
	pdhAsyncQueryTimeout                  = 0x800007DB
	pdhCannotSetDefaultRealtimeDatasource = 0x800007DC
	pdhCstatusNoObject                    = 0xC0000BB8
	pdhCstatusNoCounter                   = 0xC0000BB9 // The specified counter could not be found.
	pdhCstatusInvalidData                 = 0xC0000BBA // The counter was successfully found, but the data returned is not valid.
	pdhMemoryAllocationFailure            = 0xC0000BBB
	pdhInvalidHandle                      = 0xC0000BBC
	pdhInvalidArgument                    = 0xC0000BBD // Required argument is missing or incorrect.
	pdhFunctionNotFound                   = 0xC0000BBE
	pdhCstatusNoCountername               = 0xC0000BBF
	pdhCstatusBadCountername              = 0xC0000BC0 // Unable to parse the counter path. Check the format and syntax of the specified path.
	pdhInvalidBuffer                      = 0xC0000BC1
	pdhInsufficientBuffer                 = 0xC0000BC2
	pdhCannotConnectMachine               = 0xC0000BC3
	pdhInvalidPath                        = 0xC0000BC4
	pdhInvalidInstance                    = 0xC0000BC5
	pdhInvalidData                        = 0xC0000BC6 // specified counter does not contain valid data or a successful status code.
	pdhNoDialogData                       = 0xC0000BC7
	pdhCannotReadNameStrings              = 0xC0000BC8
	pdhLogFileCreateError                 = 0xC0000BC9
	pdhLogFileOpenError                   = 0xC0000BCA
	pdhLogTypeNotFound                    = 0xC0000BCB
	pdhNoMoreData                         = 0xC0000BCC
	pdhEntryNotInLogFile                  = 0xC0000BCD
	pdhDataSourceIsLogFile                = 0xC0000BCE
	pdhDataSourceIsRealTime               = 0xC0000BCF
	pdhUnableReadLogHeader                = 0xC0000BD0
	pdhFileNotFound                       = 0xC0000BD1
	pdhFileAlreadyExists                  = 0xC0000BD2
	pdhNotImplemented                     = 0xC0000BD3
	pdhStringNotFound                     = 0xC0000BD4
	pdhUnableMapNameFiles                 = 0x80000BD5
	pdhUnknownLogFormat                   = 0xC0000BD6
	pdhUnknownLogsvcCommand               = 0xC0000BD7
	pdhLogsvcQueryNotFound                = 0xC0000BD8
	pdhLogsvcNotOpened                    = 0xC0000BD9
	pdhWbemError                          = 0xC0000BDA
	pdhAccessDenied                       = 0xC0000BDB
	pdhLogFileTooSmall                    = 0xC0000BDC
	pdhInvalidDatasource                  = 0xC0000BDD
	pdhInvalidSqldb                       = 0xC0000BDE
	pdhNoCounters                         = 0xC0000BDF
	pdhSQLAllocFailed                     = 0xC0000BE0
	pdhSQLAllocconFailed                  = 0xC0000BE1
	pdhSQLExecDirectFailed                = 0xC0000BE2
	pdhSQLFetchFailed                     = 0xC0000BE3
	pdhSQLRowcountFailed                  = 0xC0000BE4
	pdhSQLMoreResultsFailed               = 0xC0000BE5
	pdhSQLConnectFailed                   = 0xC0000BE6
	pdhSQLBindFailed                      = 0xC0000BE7
	pdhCannotConnectWmiServer             = 0xC0000BE8
	pdhPlaCollectionAlreadyRunning        = 0xC0000BE9
	pdhPlaErrorScheduleOverlap            = 0xC0000BEA
	pdhPlaCollectionNotFound              = 0xC0000BEB
	pdhPlaErrorScheduleElapsed            = 0xC0000BEC
	pdhPlaErrorNostart                    = 0xC0000BED
	pdhPlaErrorAlreadyExists              = 0xC0000BEE
	pdhPlaErrorTypeMismatch               = 0xC0000BEF
	pdhPlaErrorFilepath                   = 0xC0000BF0
	pdhPlaServiceError                    = 0xC0000BF1
	pdhPlaValidationError                 = 0xC0000BF2
	pdhPlaValidationWarning               = 0x80000BF3
	pdhPlaErrorNameTooLong                = 0xC0000BF4
	pdhInvalidSQLLogFormat                = 0xC0000BF5
	pdhCounterAlreadyInQuery              = 0xC0000BF6
	pdhBinaryLogCorrupt                   = 0xC0000BF7
	pdhLogSampleTooSmall                  = 0xC0000BF8
	pdhOsLaterVersion                     = 0xC0000BF9
	pdhOsEarlierVersion                   = 0xC0000BFA
	pdhIncorrectAppendTime                = 0xC0000BFB
	pdhUnmatchedAppendCounter             = 0xC0000BFC
	pdhSQLAlterDetailFailed               = 0xC0000BFD
	pdhQueryPerfDataTimeout               = 0xC0000BFE
)

var pdhErrors = map[uint32]string{
	pdhCstatusValidData:                   "PDH_CSTATUS_VALID_DATA",
	pdhCstatusNewData:                     "PDH_CSTATUS_NEW_DATA",
	pdhCstatusNoMachine:                   "PDH_CSTATUS_NO_MACHINE",
	pdhCstatusNoInstance:                  "PDH_CSTATUS_NO_INSTANCE",
	pdhMoreData:                           "PDH_MORE_DATA",
	pdhCstatusItemNotValidated:            "PDH_CSTATUS_ITEM_NOT_VALIDATED",
	pdhRetry:                              "PDH_RETRY",
	pdhNoData:                             "PDH_NO_DATA",
	pdhCalcNegativeDenominator:            "PDH_CALC_NEGATIVE_DENOMINATOR",
	pdhCalcNegativeTimebase:               "PDH_CALC_NEGATIVE_TIMEBASE",
	pdhCalcNegativeValue:                  "PDH_CALC_NEGATIVE_VALUE",
	pdhDialogCancelled:                    "PDH_DIALOG_CANCELLED",
	pdhEndOfLogFile:                       "PDH_END_OF_LOG_FILE",
	pdhAsyncQueryTimeout:                  "PDH_ASYNC_QUERY_TIMEOUT",
	pdhCannotSetDefaultRealtimeDatasource: "PDH_CANNOT_SET_DEFAULT_REALTIME_DATASOURCE",
	pdhCstatusNoObject:                    "PDH_CSTATUS_NO_OBJECT",
	pdhCstatusNoCounter:                   "PDH_CSTATUS_NO_COUNTER",
	pdhCstatusInvalidData:                 "PDH_CSTATUS_INVALID_DATA",
	pdhMemoryAllocationFailure:            "PDH_MEMORY_ALLOCATION_FAILURE",
	pdhInvalidHandle:                      "PDH_INVALID_HANDLE",
	pdhInvalidArgument:                    "PDH_INVALID_ARGUMENT",
	pdhFunctionNotFound:                   "PDH_FUNCTION_NOT_FOUND",
	pdhCstatusNoCountername:               "PDH_CSTATUS_NO_COUNTERNAME",
	pdhCstatusBadCountername:              "PDH_CSTATUS_BAD_COUNTERNAME",
	pdhInvalidBuffer:                      "PDH_INVALID_BUFFER",
	pdhInsufficientBuffer:                 "PDH_INSUFFICIENT_BUFFER",
	pdhCannotConnectMachine:               "PDH_CANNOT_CONNECT_MACHINE",
	pdhInvalidPath:                        "PDH_INVALID_PATH",
	pdhInvalidInstance:                    "PDH_INVALID_INSTANCE",
	pdhInvalidData:                        "PDH_INVALID_DATA",
	pdhNoDialogData:                       "PDH_NO_DIALOG_DATA",
	pdhCannotReadNameStrings:              "PDH_CANNOT_READ_NAME_STRINGS",
	pdhLogFileCreateError:                 "PDH_LOG_FILE_CREATE_ERROR",
	pdhLogFileOpenError:                   "PDH_LOG_FILE_OPEN_ERROR",
	pdhLogTypeNotFound:                    "PDH_LOG_TYPE_NOT_FOUND",
	pdhNoMoreData:                         "PDH_NO_MORE_DATA",
	pdhEntryNotInLogFile:                  "PDH_ENTRY_NOT_IN_LOG_FILE",
	pdhDataSourceIsLogFile:                "PDH_DATA_SOURCE_IS_LOG_FILE",
	pdhDataSourceIsRealTime:               "PDH_DATA_SOURCE_IS_REAL_TIME",
	pdhUnableReadLogHeader:                "PDH_UNABLE_READ_LOG_HEADER",
	pdhFileNotFound:                       "PDH_FILE_NOT_FOUND",
	pdhFileAlreadyExists:                  "PDH_FILE_ALREADY_EXISTS",
	pdhNotImplemented:                     "PDH_NOT_IMPLEMENTED",
	pdhStringNotFound:                     "PDH_STRING_NOT_FOUND",
	pdhUnableMapNameFiles:                 "PDH_UNABLE_MAP_NAME_FILES",
	pdhUnknownLogFormat:                   "PDH_UNKNOWN_LOG_FORMAT",
	pdhUnknownLogsvcCommand:               "PDH_UNKNOWN_LOGSVC_COMMAND",
	pdhLogsvcQueryNotFound:                "PDH_LOGSVC_QUERY_NOT_FOUND",
	pdhLogsvcNotOpened:                    "PDH_LOGSVC_NOT_OPENED",
	pdhWbemError:                          "PDH_WBEM_ERROR",
	pdhAccessDenied:                       "PDH_ACCESS_DENIED",
	pdhLogFileTooSmall:                    "PDH_LOG_FILE_TOO_SMALL",
	pdhInvalidDatasource:                  "PDH_INVALID_DATASOURCE",
	pdhInvalidSqldb:                       "PDH_INVALID_SQLDB",
	pdhNoCounters:                         "PDH_NO_COUNTERS",
	pdhSQLAllocFailed:                     "PDH_SQL_ALLOC_FAILED",
	pdhSQLAllocconFailed:                  "PDH_SQL_ALLOCCON_FAILED",
	pdhSQLExecDirectFailed:                "PDH_SQL_EXEC_DIRECT_FAILED",
	pdhSQLFetchFailed:                     "PDH_SQL_FETCH_FAILED",
	pdhSQLRowcountFailed:                  "PDH_SQL_ROWCOUNT_FAILED",
	pdhSQLMoreResultsFailed:               "PDH_SQL_MORE_RESULTS_FAILED",
	pdhSQLConnectFailed:                   "PDH_SQL_CONNECT_FAILED",
	pdhSQLBindFailed:                      "PDH_SQL_BIND_FAILED",
	pdhCannotConnectWmiServer:             "PDH_CANNOT_CONNECT_WMI_SERVER",
	pdhPlaCollectionAlreadyRunning:        "PDH_PLA_COLLECTION_ALREADY_RUNNING",
	pdhPlaErrorScheduleOverlap:            "PDH_PLA_ERROR_SCHEDULE_OVERLAP",
	pdhPlaCollectionNotFound:              "PDH_PLA_COLLECTION_NOT_FOUND",
	pdhPlaErrorScheduleElapsed:            "PDH_PLA_ERROR_SCHEDULE_ELAPSED",
	pdhPlaErrorNostart:                    "PDH_PLA_ERROR_NOSTART",
	pdhPlaErrorAlreadyExists:              "PDH_PLA_ERROR_ALREADY_EXISTS",
	pdhPlaErrorTypeMismatch:               "PDH_PLA_ERROR_TYPE_MISMATCH",
	pdhPlaErrorFilepath:                   "PDH_PLA_ERROR_FILEPATH",
	pdhPlaServiceError:                    "PDH_PLA_SERVICE_ERROR",
	pdhPlaValidationError:                 "PDH_PLA_VALIDATION_ERROR",
	pdhPlaValidationWarning:               "PDH_PLA_VALIDATION_WARNING",
	pdhPlaErrorNameTooLong:                "PDH_PLA_ERROR_NAME_TOO_LONG",
	pdhInvalidSQLLogFormat:                "PDH_INVALID_SQL_LOG_FORMAT",
	pdhCounterAlreadyInQuery:              "PDH_COUNTER_ALREADY_IN_QUERY",
	pdhBinaryLogCorrupt:                   "PDH_BINARY_LOG_CORRUPT",
	pdhLogSampleTooSmall:                  "PDH_LOG_SAMPLE_TOO_SMALL",
	pdhOsLaterVersion:                     "PDH_OS_LATER_VERSION",
	pdhOsEarlierVersion:                   "PDH_OS_EARLIER_VERSION",
	pdhIncorrectAppendTime:                "PDH_INCORRECT_APPEND_TIME",
	pdhUnmatchedAppendCounter:             "PDH_UNMATCHED_APPEND_COUNTER",
	pdhSQLAlterDetailFailed:               "PDH_SQL_ALTER_DETAIL_FAILED",
	pdhQueryPerfDataTimeout:               "PDH_QUERY_PERF_DATA_TIMEOUT",
}

// Formatting options for GetFormattedCounterValue().
const (
	pdhFmtRaw          = 0x00000010
	pdhFmtAnsi         = 0x00000020
	pdhFmtUnicode      = 0x00000040
	pdhFmtLong         = 0x00000100 // Return data as a long int.
	pdhFmtDouble       = 0x00000200 // Return data as a double precision floating point real.
	pdhFmtLarge        = 0x00000400 // Return data as a 64 bit integer.
	pdhFmtNoscale      = 0x00001000 // can be OR-ed: Do not apply the counter's default scaling factor.
	pdhFmt1000         = 0x00002000 // can be OR-ed: multiply the actual value by 1,000.
	pdhFmtNodata       = 0x00004000 // can be OR-ed: unknown what this is for, MSDN says nothing.
	pdhFmtNocap100     = 0x00008000 // can be OR-ed: do not cap values > 100.
	perfDetailCostly   = 0x00010000
	perfDetailStandard = 0x0000FFFF
)

type (
	pdhQueryHandle   handle // query handle
	pdhCounterHandle handle // counter handle
)

var (
	// Library
	libPdhDll *syscall.DLL

	// Functions
	pdhAddCounterWProc               *syscall.Proc
	pdhAddEnglishCounterWProc        *syscall.Proc
	pdhCloseQueryProc                *syscall.Proc
	pdhCollectQueryDataProc          *syscall.Proc
	pdhCollectQueryDataWithTimeProc  *syscall.Proc
	pdhGetFormattedCounterValueProc  *syscall.Proc
	pdhGetFormattedCounterArrayWProc *syscall.Proc
	pdhOpenQueryProc                 *syscall.Proc
	pdhExpandWildCardPathWProc       *syscall.Proc
	pdhGetCounterInfoWProc           *syscall.Proc
	pdhGetRawCounterValueProc        *syscall.Proc
	pdhGetRawCounterArrayWProc       *syscall.Proc
)

func init() {
	// Library
	libPdhDll = syscall.MustLoadDLL("pdh.dll")

	// Functions
	pdhAddCounterWProc = libPdhDll.MustFindProc("PdhAddCounterW")
	pdhAddEnglishCounterWProc, _ = libPdhDll.FindProc("PdhAddEnglishCounterW") // XXX: only supported on versions > Vista.
	pdhCloseQueryProc = libPdhDll.MustFindProc("PdhCloseQuery")
	pdhCollectQueryDataProc = libPdhDll.MustFindProc("PdhCollectQueryData")
	pdhCollectQueryDataWithTimeProc, _ = libPdhDll.FindProc("PdhCollectQueryDataWithTime")
	pdhGetFormattedCounterValueProc = libPdhDll.MustFindProc("PdhGetFormattedCounterValue")
	pdhGetFormattedCounterArrayWProc = libPdhDll.MustFindProc("PdhGetFormattedCounterArrayW")
	pdhOpenQueryProc = libPdhDll.MustFindProc("PdhOpenQuery")
	pdhExpandWildCardPathWProc = libPdhDll.MustFindProc("PdhExpandWildCardPathW")
	pdhGetCounterInfoWProc = libPdhDll.MustFindProc("PdhGetCounterInfoW")
	pdhGetRawCounterValueProc = libPdhDll.MustFindProc("PdhGetRawCounterValue")
	pdhGetRawCounterArrayWProc = libPdhDll.MustFindProc("PdhGetRawCounterArrayW")
}

// pdhAddCounter adds the specified counter to the query. This is the internationalized version. Preferably, use the
// function pdhAddEnglishCounter instead. hQuery is the query handle, which has been fetched by pdhOpenQuery.
// szFullCounterPath is a full, internationalized counter path (this will differ per Windows language version).
// dwUserData is a 'user-defined value', which becomes part of the counter information. To retrieve this value
// later, call pdhGetCounterInfo() and access dwQueryUserData of the pdhCounterInfo structure.
//
// Examples of szFullCounterPath (in an English version of Windows):
//
//	\\Processor(_Total)\\% Idle Time
//	\\Processor(_Total)\\% Processor Time
//	\\LogicalDisk(C:)\% Free Space
//
// To view all (internationalized...) counters on a system, there are three non-programmatic ways: perfmon utility,
// the typeperf command, and the registry editor. perfmon.exe is perhaps the easiest way, because it's basically a
// full implementation of the pdh.dll API, except with a GUI and all that. The registry setting also provides an
// interface to the available counters, and can be found at the following key:
//
//	HKEY_LOCAL_MACHINE\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Perflib\CurrentLanguage
//
// This registry key contains several values as follows:
//
//	1
//	1847
//	2
//	System
//	4
//	Memory
//	6
//	% Processor Time
//	... many, many more
//
// Somehow, these numeric values can be used as szFullCounterPath too:
//
//	\2\6 will correspond to \\System\% Processor Time
//
// The typeperf command may also be pretty easy. To find all performance counters, simply execute:
//
//	typeperf -qx
func pdhAddCounter(hQuery pdhQueryHandle, szFullCounterPath string, dwUserData uintptr, phCounter *pdhCounterHandle) uint32 {
	ptxt, _ := syscall.UTF16PtrFromString(szFullCounterPath)
	ret, _, _ := pdhAddCounterWProc.Call(
		uintptr(hQuery),
		uintptr(unsafe.Pointer(ptxt)), //nolint:gosec // G103: Valid use of unsafe call to pass ptxt
		dwUserData,
		uintptr(unsafe.Pointer(phCounter))) //nolint:gosec // G103: Valid use of unsafe call to pass phCounter

	return uint32(ret)
}

// pdhAddEnglishCounterSupported returns true if PdhAddEnglishCounterW Win API function was found in pdh.dll.
// PdhAddEnglishCounterW function is not supported on pre-Windows Vista systems
func pdhAddEnglishCounterSupported() bool {
	return pdhAddEnglishCounterWProc != nil
}

// pdhAddEnglishCounter adds the specified language-neutral counter to the query. See the pdhAddCounter function. This function only exists on
// Windows versions higher than Vista.
func pdhAddEnglishCounter(hQuery pdhQueryHandle, szFullCounterPath string, dwUserData uintptr, phCounter *pdhCounterHandle) uint32 {
	if pdhAddEnglishCounterWProc == nil {
		return errorInvalidFunction
	}

	ptxt, _ := syscall.UTF16PtrFromString(szFullCounterPath)
	ret, _, _ := pdhAddEnglishCounterWProc.Call(
		uintptr(hQuery),
		uintptr(unsafe.Pointer(ptxt)), //nolint:gosec // G103: Valid use of unsafe call to pass ptxt
		dwUserData,
		uintptr(unsafe.Pointer(phCounter))) //nolint:gosec // G103: Valid use of unsafe call to pass phCounter

	return uint32(ret)
}

// pdhCloseQuery closes all counters contained in the specified query, closes all handles related to the query,
// and frees all memory associated with the query.
func pdhCloseQuery(hQuery pdhQueryHandle) uint32 {
	ret, _, _ := pdhCloseQueryProc.Call(uintptr(hQuery))

	return uint32(ret)
}

// pdhCollectQueryData collects the current raw data value for all counters in the specified query and updates the status
// code of each counter. With some counters, this function needs to be repeatedly called before the value
// of the counter can be extracted with PdhGetFormattedCounterValue(). For example, the following code
// requires at least two calls:
//
//	var handle win.PDH_HQUERY
//	var counterHandle win.PDH_HCOUNTER
//	ret := win.pdhOpenQuery(0, 0, &handle)
//	ret = win.pdhAddEnglishCounter(handle, "\\Processor(_Total)\\% Idle Time", 0, &counterHandle)
//	var derp win.PDH_FMT_COUNTERVALUE_DOUBLE
//
//	ret = win.pdhCollectQueryData(handle)
//	fmt.Printf("Collect return code is %x\n", ret) // return code will be PDH_CSTATUS_INVALID_DATA
//	ret = win.pdhGetFormattedCounterValueDouble(counterHandle, 0, &derp)
//
//	ret = win.pdhCollectQueryData(handle)
//	fmt.Printf("Collect return code is %x\n", ret) // return code will be ERROR_SUCCESS
//	ret = win.pdhGetFormattedCounterValueDouble(counterHandle, 0, &derp)
//
// The pdhCollectQueryData will return an error in the first call because it needs two values for
// displaying the correct data for the processor idle time. The second call will have a 0 return code.
func pdhCollectQueryData(hQuery pdhQueryHandle) uint32 {
	ret, _, _ := pdhCollectQueryDataProc.Call(uintptr(hQuery))

	return uint32(ret)
}

// pdhCollectQueryDataWithTime queries data from perfmon, retrieving the device/windows timestamp from the node it was collected on.
// Converts the filetime structure to a GO time class and returns the native time.
func pdhCollectQueryDataWithTime(hQuery pdhQueryHandle) (uint32, time.Time) {
	var localFileTime fileTime
	//nolint:gosec // G103: Valid use of unsafe call to pass localFileTime
	ret, _, _ := pdhCollectQueryDataWithTimeProc.Call(uintptr(hQuery), uintptr(unsafe.Pointer(&localFileTime)))

	if ret == errorSuccess {
		var utcFileTime fileTime
		ret, _, _ := kernelLocalFileTimeToFileTime.Call(
			uintptr(unsafe.Pointer(&localFileTime)), //nolint:gosec // G103: Valid use of unsafe call to pass localFileTime
			uintptr(unsafe.Pointer(&utcFileTime)))   //nolint:gosec // G103: Valid use of unsafe call to pass utcFileTime

		if ret == 0 {
			return uint32(errorFailure), time.Now()
		}

		// First convert 100-ns intervals to microseconds, then adjust for the
		// epoch difference
		var totalMicroSeconds int64
		totalMicroSeconds = ((int64(utcFileTime.dwHighDateTime) << 32) | int64(utcFileTime.dwLowDateTime)) / 10
		totalMicroSeconds -= epochDifferenceMicros

		retTime := time.Unix(0, totalMicroSeconds*1000)

		return uint32(errorSuccess), retTime
	}

	return uint32(ret), time.Now()
}

// pdhGetFormattedCounterValueDouble formats the given hCounter using a 'double'. The result is set into the specialized union struct pValue.
// This function does not directly translate to a Windows counterpart due to union specialization tricks.
func pdhGetFormattedCounterValueDouble(hCounter pdhCounterHandle, lpdwType *uint32, pValue *pdhFmtCountervalueDouble) uint32 {
	ret, _, _ := pdhGetFormattedCounterValueProc.Call(
		uintptr(hCounter),
		uintptr(pdhFmtDouble|pdhFmtNocap100),
		uintptr(unsafe.Pointer(lpdwType)), //nolint:gosec // G103: Valid use of unsafe call to pass lpdwType
		uintptr(unsafe.Pointer(pValue)))   //nolint:gosec // G103: Valid use of unsafe call to pass pValue

	return uint32(ret)
}

// pdhGetFormattedCounterArrayDouble returns an array of formatted counter values. Use this function when you want to format the counter values of a
// counter that contains a wildcard character for the instance name. The itemBuffer must a slice of type pdhFmtCountervalueItemDouble.
// An example of how this function can be used:
//
//	okPath := "\\Process(*)\\% Processor Time" // notice the wildcard * character
//
//	// omitted all necessary stuff ...
//
//	var bufSize uint32
//	var bufCount uint32
//	var size uint32 = uint32(unsafe.Sizeof(win.PDH_FMT_COUNTERVALUE_ITEM_DOUBLE{}))
//	var emptyBuf [1]win.PDH_FMT_COUNTERVALUE_ITEM_DOUBLE // need at least 1 addressable null ptr.
//
//	for {
//		// collect
//		ret := win.pdhCollectQueryData(queryHandle)
//		if ret == win.ERROR_SUCCESS {
//			ret = win.pdhGetFormattedCounterArrayDouble(counterHandle, &bufSize, &bufCount, &emptyBuf[0]) // uses null ptr here according to MSDN.
//			if ret == win.PDH_MORE_DATA {
//				filledBuf := make([]win.PDH_FMT_COUNTERVALUE_ITEM_DOUBLE, bufCount*size)
//				ret = win.pdhGetFormattedCounterArrayDouble(counterHandle, &bufSize, &bufCount, &filledBuf[0])
//				for i := 0; i < int(bufCount); i++ {
//					c := filledBuf[i]
//					var s string = win.utf16PtrToString(c.SzName)
//					fmt.Printf("Index %d -> %s, value %v\n", i, s, c.FmtValue.DoubleValue)
//				}
//
//				filledBuf = nil
//				// Need to at least set bufSize to zero, because if not, the function will not
//				// return PDH_MORE_DATA and will not set the bufSize.
//				bufCount = 0
//				bufSize = 0
//			}
//
//			time.Sleep(2000 * time.Millisecond)
//		}
//	}
func pdhGetFormattedCounterArrayDouble(hCounter pdhCounterHandle, lpdwBufferSize, lpdwBufferCount *uint32, itemBuffer *byte) uint32 {
	ret, _, _ := pdhGetFormattedCounterArrayWProc.Call(
		uintptr(hCounter),
		uintptr(pdhFmtDouble|pdhFmtNocap100),
		uintptr(unsafe.Pointer(lpdwBufferSize)),  //nolint:gosec // G103: Valid use of unsafe call to pass lpdwBufferSize
		uintptr(unsafe.Pointer(lpdwBufferCount)), //nolint:gosec // G103: Valid use of unsafe call to pass lpdwBufferCount
		uintptr(unsafe.Pointer(itemBuffer)))      //nolint:gosec // G103: Valid use of unsafe call to pass itemBuffer

	return uint32(ret)
}

// pdhOpenQuery creates a new query that is used to manage the collection of performance data.
// szDataSource is a null terminated string that specifies the name of the log file from which to
// retrieve the performance data. If 0, performance data is collected from a real-time data source.
// dwUserData is a user-defined value to associate with this query. To retrieve the user data later,
// call pdhGetCounterInfo and access dwQueryUserData of the pdhCounterInfo structure. phQuery is
// the handle to the query, and must be used in subsequent calls. This function returns a PDH_
// constant error code, or errorSuccess if the call succeeded.
func pdhOpenQuery(szDataSource, dwUserData uintptr, phQuery *pdhQueryHandle) uint32 {
	ret, _, _ := pdhOpenQueryProc.Call(
		szDataSource,
		dwUserData,
		uintptr(unsafe.Pointer(phQuery))) //nolint:gosec // G103: Valid use of unsafe call to pass phQuery

	return uint32(ret)
}

// pdhExpandWildCardPath examines the specified computer or log file and returns those counter paths that match the given counter path
// which contains wildcard characters. The general counter path format is as follows:
//
// \\computer\object(parent/instance#index)\counter
//
// The parent, instance, index, and counter components of the counter path may contain either a valid name or a wildcard character.
// The computer, parent, instance, and index components are not necessary for all counters.
//
// The following is a list of the possible formats:
//
// \\computer\object(parent/instance#index)\counter
// \\computer\object(parent/instance)\counter
// \\computer\object(instance#index)\counter
// \\computer\object(instance)\counter
// \\computer\object\counter
// \object(parent/instance#index)\counter
// \object(parent/instance)\counter
// \object(instance#index)\counter
// \object(instance)\counter
// \object\counter
// Use an asterisk (*) as the wildcard character, for example, \object(*)\counter.
//
// If a wildcard character is specified in the parent name, all instances of the specified object
// that match the specified instance and counter fields will be returned.
// For example, \object(*/instance)\counter.
//
// If a wildcard character is specified in the instance name, all instances of the specified object and parent object will be returned if all instance names
// corresponding to the specified index match the wildcard character. For example, \object(parent/*)\counter.
// If the object does not contain an instance, an error occurs.
//
// If a wildcard character is specified in the counter name, all counters of the specified object are returned.
//
// Partial counter path string matches (for example, "pro*") are supported.
func pdhExpandWildCardPath(szWildCardPath string, mszExpandedPathList *uint16, pcchPathListLength *uint32) uint32 {
	ptxt, _ := syscall.UTF16PtrFromString(szWildCardPath)
	flags := uint32(0) // expand instances and counters
	ret, _, _ := pdhExpandWildCardPathWProc.Call(
		0,                             // search counters on local computer
		uintptr(unsafe.Pointer(ptxt)), //nolint:gosec // G103: Valid use of unsafe call to pass ptxt
		uintptr(unsafe.Pointer(mszExpandedPathList)), //nolint:gosec // G103: Valid use of unsafe call to pass mszExpandedPathList
		uintptr(unsafe.Pointer(pcchPathListLength)),  //nolint:gosec // G103: Valid use of unsafe call to pass pcchPathListLength
		uintptr(unsafe.Pointer(&flags)))              //nolint:gosec // G103: Valid use of unsafe call to pass flags

	return uint32(ret)
}

func pdhFormatError(msgID uint32) string {
	var flags uint32 = windows.FORMAT_MESSAGE_FROM_HMODULE | windows.FORMAT_MESSAGE_ARGUMENT_ARRAY | windows.FORMAT_MESSAGE_IGNORE_INSERTS
	buf := make([]uint16, 300)
	_, err := windows.FormatMessage(flags, uintptr(libPdhDll.Handle), msgID, 0, buf, nil)
	if err == nil {
		return utf16PtrToString(&buf[0])
	}
	return fmt.Sprintf("(pdhErr=%d) %s", msgID, err.Error())
}

// pdhGetCounterInfo retrieves information about a counter, such as data size, counter type, path, and user-supplied data values
// hCounter [in]
// Handle of the counter from which you want to retrieve information. The pdhAddCounter function returns this handle.
//
// bRetrieveExplainText [in]
// Determines whether explain text is retrieved. If you set this parameter to TRUE, the explain text for the counter is retrieved.
// If you set this parameter to FALSE, the field in the returned buffer is NULL.
//
// pdwBufferSize [in, out]
// Size of the lpBuffer buffer, in bytes. If zero on input, the function returns PdhMoreData and sets this parameter to the required buffer size.
// If the buffer is larger than the required size, the function sets this parameter to the actual size of the buffer that was used.
// If the specified size on input is greater than zero but less than the required size, you should not rely on the returned size to reallocate the buffer.
//
// lpBuffer [out]
// Caller-allocated buffer that receives a pdhCounterInfo structure.
// The structure is variable-length, because the string data is appended to the end of the fixed-format portion of the structure.
// This is done so that all data is returned in a single buffer allocated by the caller. Set to NULL if pdwBufferSize is zero.
func pdhGetCounterInfo(hCounter pdhCounterHandle, bRetrieveExplainText int, pdwBufferSize *uint32, lpBuffer *byte) uint32 {
	ret, _, _ := pdhGetCounterInfoWProc.Call(
		uintptr(hCounter),
		uintptr(bRetrieveExplainText),
		uintptr(unsafe.Pointer(pdwBufferSize)), //nolint:gosec // G103: Valid use of unsafe call to pass pdwBufferSize
		uintptr(unsafe.Pointer(lpBuffer)))      //nolint:gosec // G103: Valid use of unsafe call to pass lpBuffer

	return uint32(ret)
}

// pdhGetRawCounterValue returns the current raw value of the counter.
// If the specified counter instance does not exist, this function will return errorSuccess
// and the CStatus member of the pdhRawCounter structure will contain PdhCstatusNoInstance.
//
// hCounter [in]
// Handle of the counter from which to retrieve the current raw value. The pdhAddCounter function returns this handle.
//
// lpdwType [out]
// Receives the counter type. For a list of counter types, see the Counter Types section of the Windows Server 2003 Deployment Kit.
// This parameter is optional.
//
// pValue [out]
// A pdhRawCounter structure that receives the counter value.
func pdhGetRawCounterValue(hCounter pdhCounterHandle, lpdwType *uint32, pValue *pdhRawCounter) uint32 {
	ret, _, _ := pdhGetRawCounterValueProc.Call(
		uintptr(hCounter),
		uintptr(unsafe.Pointer(lpdwType)), //nolint:gosec // G103: Valid use of unsafe call to pass lpdwType
		uintptr(unsafe.Pointer(pValue)))   //nolint:gosec // G103: Valid use of unsafe call to pass pValue

	return uint32(ret)
}

// pdhGetRawCounterArray returns an array of raw values from the specified counter. Use this function when you want to retrieve the raw counter values
// of a counter that contains a wildcard character for the instance name.
// hCounter
// Handle of the counter for whose current raw instance values you want to retrieve. The pdhAddCounter function returns this handle.
//
// lpdwBufferSize
// Size of the ItemBuffer buffer, in bytes. If zero on input, the function returns PdhMoreData and sets this parameter to the required buffer size.
// If the buffer is larger than the required size, the function sets this parameter to the actual size of the buffer that was used.
// If the specified size on input is greater than zero but less than the required size, you should not rely on the returned size to reallocate the buffer.
//
// lpdwItemCount
// Number of raw counter values in the ItemBuffer buffer.
//
// ItemBuffer
// Caller-allocated buffer that receives the array of pdhRawCounterItem structures; the structures contain the raw instance counter values.
// Set to NULL if lpdwBufferSize is zero.
func pdhGetRawCounterArray(hCounter pdhCounterHandle, lpdwBufferSize, lpdwBufferCount *uint32, itemBuffer *byte) uint32 {
	ret, _, _ := pdhGetRawCounterArrayWProc.Call(
		uintptr(hCounter),
		uintptr(unsafe.Pointer(lpdwBufferSize)),  //nolint:gosec // G103: Valid use of unsafe call to pass lpdwBufferSize
		uintptr(unsafe.Pointer(lpdwBufferCount)), //nolint:gosec // G103: Valid use of unsafe call to pass lpdwBufferCount
		uintptr(unsafe.Pointer(itemBuffer)))      //nolint:gosec // G103: Valid use of unsafe call to pass itemBuffer
	return uint32(ret)
}
