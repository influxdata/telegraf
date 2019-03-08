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

// +build windows

package win_perf_counters

import (
	"fmt"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
	"time"
)

// Error codes
const (
	ERROR_SUCCESS                 = 0
	ERROR_FAILURE                 = 1
	ERROR_INVALID_FUNCTION        = 1
	EPOCH_DIFFERENCE_MICROS int64 = 11644473600000000
)

type (
	HANDLE uintptr
)

// PDH error codes, which can be returned by all Pdh* functions. Taken from mingw-w64 pdhmsg.h
const (
	PDH_CSTATUS_VALID_DATA                     = 0x00000000 // The returned data is valid.
	PDH_CSTATUS_NEW_DATA                       = 0x00000001 // The return data value is valid and different from the last sample.
	PDH_CSTATUS_NO_MACHINE                     = 0x800007D0 // Unable to connect to the specified computer, or the computer is offline.
	PDH_CSTATUS_NO_INSTANCE                    = 0x800007D1
	PDH_MORE_DATA                              = 0x800007D2 // The PdhGetFormattedCounterArray* function can return this if there's 'more data to be displayed'.
	PDH_CSTATUS_ITEM_NOT_VALIDATED             = 0x800007D3
	PDH_RETRY                                  = 0x800007D4
	PDH_NO_DATA                                = 0x800007D5 // The query does not currently contain any counters (for example, limited access)
	PDH_CALC_NEGATIVE_DENOMINATOR              = 0x800007D6
	PDH_CALC_NEGATIVE_TIMEBASE                 = 0x800007D7
	PDH_CALC_NEGATIVE_VALUE                    = 0x800007D8
	PDH_DIALOG_CANCELLED                       = 0x800007D9
	PDH_END_OF_LOG_FILE                        = 0x800007DA
	PDH_ASYNC_QUERY_TIMEOUT                    = 0x800007DB
	PDH_CANNOT_SET_DEFAULT_REALTIME_DATASOURCE = 0x800007DC
	PDH_CSTATUS_NO_OBJECT                      = 0xC0000BB8
	PDH_CSTATUS_NO_COUNTER                     = 0xC0000BB9 // The specified counter could not be found.
	PDH_CSTATUS_INVALID_DATA                   = 0xC0000BBA // The counter was successfully found, but the data returned is not valid.
	PDH_MEMORY_ALLOCATION_FAILURE              = 0xC0000BBB
	PDH_INVALID_HANDLE                         = 0xC0000BBC
	PDH_INVALID_ARGUMENT                       = 0xC0000BBD // Required argument is missing or incorrect.
	PDH_FUNCTION_NOT_FOUND                     = 0xC0000BBE
	PDH_CSTATUS_NO_COUNTERNAME                 = 0xC0000BBF
	PDH_CSTATUS_BAD_COUNTERNAME                = 0xC0000BC0 // Unable to parse the counter path. Check the format and syntax of the specified path.
	PDH_INVALID_BUFFER                         = 0xC0000BC1
	PDH_INSUFFICIENT_BUFFER                    = 0xC0000BC2
	PDH_CANNOT_CONNECT_MACHINE                 = 0xC0000BC3
	PDH_INVALID_PATH                           = 0xC0000BC4
	PDH_INVALID_INSTANCE                       = 0xC0000BC5
	PDH_INVALID_DATA                           = 0xC0000BC6 // specified counter does not contain valid data or a successful status code.
	PDH_NO_DIALOG_DATA                         = 0xC0000BC7
	PDH_CANNOT_READ_NAME_STRINGS               = 0xC0000BC8
	PDH_LOG_FILE_CREATE_ERROR                  = 0xC0000BC9
	PDH_LOG_FILE_OPEN_ERROR                    = 0xC0000BCA
	PDH_LOG_TYPE_NOT_FOUND                     = 0xC0000BCB
	PDH_NO_MORE_DATA                           = 0xC0000BCC
	PDH_ENTRY_NOT_IN_LOG_FILE                  = 0xC0000BCD
	PDH_DATA_SOURCE_IS_LOG_FILE                = 0xC0000BCE
	PDH_DATA_SOURCE_IS_REAL_TIME               = 0xC0000BCF
	PDH_UNABLE_READ_LOG_HEADER                 = 0xC0000BD0
	PDH_FILE_NOT_FOUND                         = 0xC0000BD1
	PDH_FILE_ALREADY_EXISTS                    = 0xC0000BD2
	PDH_NOT_IMPLEMENTED                        = 0xC0000BD3
	PDH_STRING_NOT_FOUND                       = 0xC0000BD4
	PDH_UNABLE_MAP_NAME_FILES                  = 0x80000BD5
	PDH_UNKNOWN_LOG_FORMAT                     = 0xC0000BD6
	PDH_UNKNOWN_LOGSVC_COMMAND                 = 0xC0000BD7
	PDH_LOGSVC_QUERY_NOT_FOUND                 = 0xC0000BD8
	PDH_LOGSVC_NOT_OPENED                      = 0xC0000BD9
	PDH_WBEM_ERROR                             = 0xC0000BDA
	PDH_ACCESS_DENIED                          = 0xC0000BDB
	PDH_LOG_FILE_TOO_SMALL                     = 0xC0000BDC
	PDH_INVALID_DATASOURCE                     = 0xC0000BDD
	PDH_INVALID_SQLDB                          = 0xC0000BDE
	PDH_NO_COUNTERS                            = 0xC0000BDF
	PDH_SQL_ALLOC_FAILED                       = 0xC0000BE0
	PDH_SQL_ALLOCCON_FAILED                    = 0xC0000BE1
	PDH_SQL_EXEC_DIRECT_FAILED                 = 0xC0000BE2
	PDH_SQL_FETCH_FAILED                       = 0xC0000BE3
	PDH_SQL_ROWCOUNT_FAILED                    = 0xC0000BE4
	PDH_SQL_MORE_RESULTS_FAILED                = 0xC0000BE5
	PDH_SQL_CONNECT_FAILED                     = 0xC0000BE6
	PDH_SQL_BIND_FAILED                        = 0xC0000BE7
	PDH_CANNOT_CONNECT_WMI_SERVER              = 0xC0000BE8
	PDH_PLA_COLLECTION_ALREADY_RUNNING         = 0xC0000BE9
	PDH_PLA_ERROR_SCHEDULE_OVERLAP             = 0xC0000BEA
	PDH_PLA_COLLECTION_NOT_FOUND               = 0xC0000BEB
	PDH_PLA_ERROR_SCHEDULE_ELAPSED             = 0xC0000BEC
	PDH_PLA_ERROR_NOSTART                      = 0xC0000BED
	PDH_PLA_ERROR_ALREADY_EXISTS               = 0xC0000BEE
	PDH_PLA_ERROR_TYPE_MISMATCH                = 0xC0000BEF
	PDH_PLA_ERROR_FILEPATH                     = 0xC0000BF0
	PDH_PLA_SERVICE_ERROR                      = 0xC0000BF1
	PDH_PLA_VALIDATION_ERROR                   = 0xC0000BF2
	PDH_PLA_VALIDATION_WARNING                 = 0x80000BF3
	PDH_PLA_ERROR_NAME_TOO_LONG                = 0xC0000BF4
	PDH_INVALID_SQL_LOG_FORMAT                 = 0xC0000BF5
	PDH_COUNTER_ALREADY_IN_QUERY               = 0xC0000BF6
	PDH_BINARY_LOG_CORRUPT                     = 0xC0000BF7
	PDH_LOG_SAMPLE_TOO_SMALL                   = 0xC0000BF8
	PDH_OS_LATER_VERSION                       = 0xC0000BF9
	PDH_OS_EARLIER_VERSION                     = 0xC0000BFA
	PDH_INCORRECT_APPEND_TIME                  = 0xC0000BFB
	PDH_UNMATCHED_APPEND_COUNTER               = 0xC0000BFC
	PDH_SQL_ALTER_DETAIL_FAILED                = 0xC0000BFD
	PDH_QUERY_PERF_DATA_TIMEOUT                = 0xC0000BFE
)

// Formatting options for GetFormattedCounterValue().
const (
	PDH_FMT_RAW          = 0x00000010
	PDH_FMT_ANSI         = 0x00000020
	PDH_FMT_UNICODE      = 0x00000040
	PDH_FMT_LONG         = 0x00000100 // Return data as a long int.
	PDH_FMT_DOUBLE       = 0x00000200 // Return data as a double precision floating point real.
	PDH_FMT_LARGE        = 0x00000400 // Return data as a 64 bit integer.
	PDH_FMT_NOSCALE      = 0x00001000 // can be OR-ed: Do not apply the counter's default scaling factor.
	PDH_FMT_1000         = 0x00002000 // can be OR-ed: multiply the actual value by 1,000.
	PDH_FMT_NODATA       = 0x00004000 // can be OR-ed: unknown what this is for, MSDN says nothing.
	PDH_FMT_NOCAP100     = 0x00008000 // can be OR-ed: do not cap values > 100.
	PERF_DETAIL_COSTLY   = 0x00010000
	PERF_DETAIL_STANDARD = 0x0000FFFF
)

type (
	PDH_HQUERY   HANDLE // query handle
	PDH_HCOUNTER HANDLE // counter handle
)

var (
	// Library
	libpdhDll *syscall.DLL

	// Functions
	pdh_AddCounterW               *syscall.Proc
	pdh_AddEnglishCounterW        *syscall.Proc
	pdh_CloseQuery                *syscall.Proc
	pdh_CollectQueryData          *syscall.Proc
	pdh_CollectQueryDataWithTime  *syscall.Proc
	pdh_GetFormattedCounterValue  *syscall.Proc
	pdh_GetFormattedCounterArrayW *syscall.Proc
	pdh_OpenQuery                 *syscall.Proc
	pdh_ValidatePathW             *syscall.Proc
	pdh_ExpandWildCardPathW       *syscall.Proc
	pdh_GetCounterInfoW           *syscall.Proc
)

func init() {
	// Library
	libpdhDll = syscall.MustLoadDLL("pdh.dll")

	// Functions
	pdh_AddCounterW = libpdhDll.MustFindProc("PdhAddCounterW")
	pdh_AddEnglishCounterW, _ = libpdhDll.FindProc("PdhAddEnglishCounterW") // XXX: only supported on versions > Vista.
	pdh_CloseQuery = libpdhDll.MustFindProc("PdhCloseQuery")
	pdh_CollectQueryData = libpdhDll.MustFindProc("PdhCollectQueryData")
	pdh_CollectQueryDataWithTime, _ = libpdhDll.FindProc("PdhCollectQueryDataWithTime")
	pdh_GetFormattedCounterValue = libpdhDll.MustFindProc("PdhGetFormattedCounterValue")
	pdh_GetFormattedCounterArrayW = libpdhDll.MustFindProc("PdhGetFormattedCounterArrayW")
	pdh_OpenQuery = libpdhDll.MustFindProc("PdhOpenQuery")
	pdh_ValidatePathW = libpdhDll.MustFindProc("PdhValidatePathW")
	pdh_ExpandWildCardPathW = libpdhDll.MustFindProc("PdhExpandWildCardPathW")
	pdh_GetCounterInfoW = libpdhDll.MustFindProc("PdhGetCounterInfoW")
}

// PdhAddCounter adds the specified counter to the query. This is the internationalized version. Preferably, use the
// function PdhAddEnglishCounter instead. hQuery is the query handle, which has been fetched by PdhOpenQuery.
// szFullCounterPath is a full, internationalized counter path (this will differ per Windows language version).
// dwUserData is a 'user-defined value', which becomes part of the counter information. To retrieve this value
// later, call PdhGetCounterInfo() and access dwQueryUserData of the PDH_COUNTER_INFO structure.
//
// Examples of szFullCounterPath (in an English version of Windows):
//
//	\\Processor(_Total)\\% Idle Time
//	\\Processor(_Total)\\% Processor Time
//	\\LogicalDisk(C:)\% Free Space
//
// To view all (internationalized...) counters on a system, there are three non-programmatic ways: perfmon utility,
// the typeperf command, and the the registry editor. perfmon.exe is perhaps the easiest way, because it's basically a
// full implemention of the pdh.dll API, except with a GUI and all that. The registry setting also provides an
// interface to the available counters, and can be found at the following key:
//
// 	HKEY_LOCAL_MACHINE\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Perflib\CurrentLanguage
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
func PdhAddCounter(hQuery PDH_HQUERY, szFullCounterPath string, dwUserData uintptr, phCounter *PDH_HCOUNTER) uint32 {
	ptxt, _ := syscall.UTF16PtrFromString(szFullCounterPath)
	ret, _, _ := pdh_AddCounterW.Call(
		uintptr(hQuery),
		uintptr(unsafe.Pointer(ptxt)),
		dwUserData,
		uintptr(unsafe.Pointer(phCounter)))

	return uint32(ret)
}

// PdhAddEnglishCounterSupported returns true if PdhAddEnglishCounterW Win API function was found in pdh.dll.
// PdhAddEnglishCounterW function is not supported on pre-Windows Vista systems

func PdhAddEnglishCounterSupported() bool {
	return pdh_AddEnglishCounterW != nil
}

// PdhAddEnglishCounter adds the specified language-neutral counter to the query. See the PdhAddCounter function. This function only exists on
// Windows versions higher than Vista.
func PdhAddEnglishCounter(hQuery PDH_HQUERY, szFullCounterPath string, dwUserData uintptr, phCounter *PDH_HCOUNTER) uint32 {
	if pdh_AddEnglishCounterW == nil {
		return ERROR_INVALID_FUNCTION
	}

	ptxt, _ := syscall.UTF16PtrFromString(szFullCounterPath)
	ret, _, _ := pdh_AddEnglishCounterW.Call(
		uintptr(hQuery),
		uintptr(unsafe.Pointer(ptxt)),
		dwUserData,
		uintptr(unsafe.Pointer(phCounter)))

	return uint32(ret)
}

// PdhCloseQuery closes all counters contained in the specified query, closes all handles related to the query,
// and frees all memory associated with the query.
func PdhCloseQuery(hQuery PDH_HQUERY) uint32 {
	ret, _, _ := pdh_CloseQuery.Call(uintptr(hQuery))

	return uint32(ret)
}

// Collects the current raw data value for all counters in the specified query and updates the status
// code of each counter. With some counters, this function needs to be repeatedly called before the value
// of the counter can be extracted with PdhGetFormattedCounterValue(). For example, the following code
// requires at least two calls:
//
// 	var handle win.PDH_HQUERY
// 	var counterHandle win.PDH_HCOUNTER
// 	ret := win.PdhOpenQuery(0, 0, &handle)
//	ret = win.PdhAddEnglishCounter(handle, "\\Processor(_Total)\\% Idle Time", 0, &counterHandle)
//	var derp win.PDH_FMT_COUNTERVALUE_DOUBLE
//
//	ret = win.PdhCollectQueryData(handle)
//	fmt.Printf("Collect return code is %x\n", ret) // return code will be PDH_CSTATUS_INVALID_DATA
//	ret = win.PdhGetFormattedCounterValueDouble(counterHandle, 0, &derp)
//
//	ret = win.PdhCollectQueryData(handle)
//	fmt.Printf("Collect return code is %x\n", ret) // return code will be ERROR_SUCCESS
//	ret = win.PdhGetFormattedCounterValueDouble(counterHandle, 0, &derp)
//
// The PdhCollectQueryData will return an error in the first call because it needs two values for
// displaying the correct data for the processor idle time. The second call will have a 0 return code.
func PdhCollectQueryData(hQuery PDH_HQUERY) uint32 {
	ret, _, _ := pdh_CollectQueryData.Call(uintptr(hQuery))

	return uint32(ret)
}

// PdhCollectQueryDataWithTime queries data from perfmon, retrieving the device/windows timestamp from the node it was collected on.
// Converts the filetime structure to a GO time class and returns the native time.
//
func PdhCollectQueryDataWithTime(hQuery PDH_HQUERY) (uint32, time.Time) {
	var localFileTime FILETIME
	ret, _, _ := pdh_CollectQueryDataWithTime.Call(uintptr(hQuery), uintptr(unsafe.Pointer(&localFileTime)))

	if ret == ERROR_SUCCESS {
		var utcFileTime FILETIME
		ret, _, _ := krn_LocalFileTimeToFileTime.Call(
			uintptr(unsafe.Pointer(&localFileTime)),
			uintptr(unsafe.Pointer(&utcFileTime)))

		if ret == 0 {
			return uint32(ERROR_FAILURE), time.Now()
		}

		// First convert 100-ns intervals to microseconds, then adjust for the
		// epoch difference
		var totalMicroSeconds int64
		totalMicroSeconds = ((int64(utcFileTime.dwHighDateTime) << 32) | int64(utcFileTime.dwLowDateTime)) / 10
		totalMicroSeconds -= EPOCH_DIFFERENCE_MICROS

		retTime := time.Unix(0, totalMicroSeconds*1000)

		return uint32(ERROR_SUCCESS), retTime
	}

	return uint32(ret), time.Now()
}

// PdhGetFormattedCounterValueDouble formats the given hCounter using a 'double'. The result is set into the specialized union struct pValue.
// This function does not directly translate to a Windows counterpart due to union specialization tricks.
func PdhGetFormattedCounterValueDouble(hCounter PDH_HCOUNTER, lpdwType *uint32, pValue *PDH_FMT_COUNTERVALUE_DOUBLE) uint32 {
	ret, _, _ := pdh_GetFormattedCounterValue.Call(
		uintptr(hCounter),
		uintptr(PDH_FMT_DOUBLE|PDH_FMT_NOCAP100),
		uintptr(unsafe.Pointer(lpdwType)),
		uintptr(unsafe.Pointer(pValue)))

	return uint32(ret)
}

// PdhGetFormattedCounterArrayDouble returns an array of formatted counter values. Use this function when you want to format the counter values of a
// counter that contains a wildcard character for the instance name. The itemBuffer must a slice of type PDH_FMT_COUNTERVALUE_ITEM_DOUBLE.
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
//		ret := win.PdhCollectQueryData(queryHandle)
//		if ret == win.ERROR_SUCCESS {
//			ret = win.PdhGetFormattedCounterArrayDouble(counterHandle, &bufSize, &bufCount, &emptyBuf[0]) // uses null ptr here according to MSDN.
//			if ret == win.PDH_MORE_DATA {
//				filledBuf := make([]win.PDH_FMT_COUNTERVALUE_ITEM_DOUBLE, bufCount*size)
//				ret = win.PdhGetFormattedCounterArrayDouble(counterHandle, &bufSize, &bufCount, &filledBuf[0])
//				for i := 0; i < int(bufCount); i++ {
//					c := filledBuf[i]
//					var s string = win.UTF16PtrToString(c.SzName)
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
func PdhGetFormattedCounterArrayDouble(hCounter PDH_HCOUNTER, lpdwBufferSize *uint32, lpdwBufferCount *uint32, itemBuffer *byte) uint32 {
	ret, _, _ := pdh_GetFormattedCounterArrayW.Call(
		uintptr(hCounter),
		uintptr(PDH_FMT_DOUBLE|PDH_FMT_NOCAP100),
		uintptr(unsafe.Pointer(lpdwBufferSize)),
		uintptr(unsafe.Pointer(lpdwBufferCount)),
		uintptr(unsafe.Pointer(itemBuffer)))

	return uint32(ret)
}

// PdhOpenQuery creates a new query that is used to manage the collection of performance data.
// szDataSource is a null terminated string that specifies the name of the log file from which to
// retrieve the performance data. If 0, performance data is collected from a real-time data source.
// dwUserData is a user-defined value to associate with this query. To retrieve the user data later,
// call PdhGetCounterInfo and access dwQueryUserData of the PDH_COUNTER_INFO structure. phQuery is
// the handle to the query, and must be used in subsequent calls. This function returns a PDH_
// constant error code, or ERROR_SUCCESS if the call succeeded.
func PdhOpenQuery(szDataSource uintptr, dwUserData uintptr, phQuery *PDH_HQUERY) uint32 {
	ret, _, _ := pdh_OpenQuery.Call(
		szDataSource,
		dwUserData,
		uintptr(unsafe.Pointer(phQuery)))

	return uint32(ret)
}

//PdhExpandWildCardPath examines the specified computer or log file and returns those counter paths that match the given counter path which contains wildcard characters.
//The general counter path format is as follows:
//
//\\computer\object(parent/instance#index)\counter
//
//The parent, instance, index, and counter components of the counter path may contain either a valid name or a wildcard character. The computer, parent, instance,
// and index components are not necessary for all counters.
//
//The following is a list of the possible formats:
//
//\\computer\object(parent/instance#index)\counter
//\\computer\object(parent/instance)\counter
//\\computer\object(instance#index)\counter
//\\computer\object(instance)\counter
//\\computer\object\counter
//\object(parent/instance#index)\counter
//\object(parent/instance)\counter
//\object(instance#index)\counter
//\object(instance)\counter
//\object\counter
//Use an asterisk (*) as the wildcard character, for example, \object(*)\counter.
//
//If a wildcard character is specified in the parent name, all instances of the specified object that match the specified instance and counter fields will be returned.
// For example, \object(*/instance)\counter.
//
//If a wildcard character is specified in the instance name, all instances of the specified object and parent object will be returned if all instance names
// corresponding to the specified index match the wildcard character. For example, \object(parent/*)\counter. If the object does not contain an instance, an error occurs.
//
//If a wildcard character is specified in the counter name, all counters of the specified object are returned.
//
//Partial counter path string matches (for example, "pro*") are supported.
func PdhExpandWildCardPath(szWildCardPath string, mszExpandedPathList *uint16, pcchPathListLength *uint32) uint32 {
	ptxt, _ := syscall.UTF16PtrFromString(szWildCardPath)
	flags := uint32(0) // expand instances and counters
	ret, _, _ := pdh_ExpandWildCardPathW.Call(
		uintptr(unsafe.Pointer(nil)), // search counters on local computer
		uintptr(unsafe.Pointer(ptxt)),
		uintptr(unsafe.Pointer(mszExpandedPathList)),
		uintptr(unsafe.Pointer(pcchPathListLength)),
		uintptr(unsafe.Pointer(&flags)))

	return uint32(ret)
}

// PdhValidatePath validates a path. Will return ERROR_SUCCESS when ok, or PDH_CSTATUS_BAD_COUNTERNAME when the path is
// erroneous.
func PdhValidatePath(path string) uint32 {
	ptxt, _ := syscall.UTF16PtrFromString(path)
	ret, _, _ := pdh_ValidatePathW.Call(uintptr(unsafe.Pointer(ptxt)))

	return uint32(ret)
}

func PdhFormatError(msgId uint32) string {
	var flags uint32 = windows.FORMAT_MESSAGE_FROM_HMODULE | windows.FORMAT_MESSAGE_ARGUMENT_ARRAY | windows.FORMAT_MESSAGE_IGNORE_INSERTS
	buf := make([]uint16, 300)
	_, err := windows.FormatMessage(flags, uintptr(libpdhDll.Handle), msgId, 0, buf, nil)
	if err == nil {
		return fmt.Sprintf("%s", UTF16PtrToString(&buf[0]))
	}
	return fmt.Sprintf("(pdhErr=%d) %s", msgId, err.Error())
}

//Retrieves information about a counter, such as data size, counter type, path, and user-supplied data values
//hCounter [in]
//Handle of the counter from which you want to retrieve information. The PdhAddCounter function returns this handle.
//
//bRetrieveExplainText [in]
//Determines whether explain text is retrieved. If you set this parameter to TRUE, the explain text for the counter is retrieved. If you set this parameter to FALSE, the field in the returned buffer is NULL.
//
//pdwBufferSize [in, out]
//Size of the lpBuffer buffer, in bytes. If zero on input, the function returns PDH_MORE_DATA and sets this parameter to the required buffer size. If the buffer is larger than the required size, the function sets this parameter to the actual size of the buffer that was used. If the specified size on input is greater than zero but less than the required size, you should not rely on the returned size to reallocate the buffer.
//
//lpBuffer [out]
//Caller-allocated buffer that receives a PDH_COUNTER_INFO structure. The structure is variable-length, because the string data is appended to the end of the fixed-format portion of the structure. This is done so that all data is returned in a single buffer allocated by the caller. Set to NULL if pdwBufferSize is zero.
func PdhGetCounterInfo(hCounter PDH_HCOUNTER, bRetrieveExplainText int, pdwBufferSize *uint32, lpBuffer *byte) uint32 {
	ret, _, _ := pdh_GetCounterInfoW.Call(
		uintptr(hCounter),
		uintptr(bRetrieveExplainText),
		uintptr(unsafe.Pointer(pdwBufferSize)),
		uintptr(unsafe.Pointer(lpBuffer)))

	return uint32(ret)
}
