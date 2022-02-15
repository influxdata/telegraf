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
// +build windows

package win_perf_counters

// Union specialization for double values
type PDH_FMT_COUNTERVALUE_DOUBLE struct {
	CStatus     uint32
	DoubleValue float64
}

// Union specialization for 64 bit integer values
type PDH_FMT_COUNTERVALUE_LARGE struct {
	CStatus    uint32
	LargeValue int64
}

// Union specialization for long values
type PDH_FMT_COUNTERVALUE_LONG struct {
	CStatus   uint32
	LongValue int32
	padding   [4]byte
}

type PDH_FMT_COUNTERVALUE_ITEM_DOUBLE struct {
	SzName   *uint16
	FmtValue PDH_FMT_COUNTERVALUE_DOUBLE
}

// Union specialization for 'large' values, used by PdhGetFormattedCounterArrayLarge()
type PDH_FMT_COUNTERVALUE_ITEM_LARGE struct {
	SzName   *uint16 // pointer to a string
	FmtValue PDH_FMT_COUNTERVALUE_LARGE
}

// Union specialization for long values, used by PdhGetFormattedCounterArrayLong()
type PDH_FMT_COUNTERVALUE_ITEM_LONG struct {
	SzName   *uint16 // pointer to a string
	FmtValue PDH_FMT_COUNTERVALUE_LONG
}

//PDH_COUNTER_INFO structure contains information describing the properties of a counter. This information also includes the counter path.
type PDH_COUNTER_INFO struct {
	//Size of the structure, including the appended strings, in bytes.
	DwLength uint32
	//Counter type. For a list of counter types, see the Counter Types section of the <a "href=http://go.microsoft.com/fwlink/p/?linkid=84422">Windows Server 2003 Deployment Kit</a>.
	//The counter type constants are defined in Winperf.h.
	DwType uint32
	//Counter version information. Not used.
	CVersion uint32
	//Counter status that indicates if the counter value is valid. For a list of possible values,
	//see <a href="https://msdn.microsoft.com/en-us/library/windows/desktop/aa371894(v=vs.85).aspx">Checking PDH Interface Return Values</a>.
	CStatus uint32
	//Scale factor to use when computing the displayable value of the counter. The scale factor is a power of ten.
	//The valid range of this parameter is PDH_MIN_SCALE (–7) (the returned value is the actual value times 10–⁷) to
	//PDH_MAX_SCALE (+7) (the returned value is the actual value times 10⁺⁷). A value of zero will set the scale to one, so that the actual value is returned
	LScale int32
	//Default scale factor as suggested by the counter's provider.
	LDefaultScale int32
	//The value passed in the dwUserData parameter when calling PdhAddCounter.
	DwUserData *uint32
	//The value passed in the dwUserData parameter when calling PdhOpenQuery.
	DwQueryUserData *uint32
	//Null-terminated string that specifies the full counter path. The string follows this structure in memory.
	SzFullPath *uint16 // pointer to a string
	//Null-terminated string that contains the name of the computer specified in the counter path. Is NULL, if the path does not specify a computer.
	//The string follows this structure in memory.
	SzMachineName *uint16 // pointer to a string
	//Null-terminated string that contains the name of the performance object specified in the counter path. The string follows this structure in memory.
	SzObjectName *uint16 // pointer to a string
	//Null-terminated string that contains the name of the object instance specified in the counter path. Is NULL, if the path does not specify an instance.
	//The string follows this structure in memory.
	SzInstanceName *uint16 // pointer to a string
	//Null-terminated string that contains the name of the parent instance specified in the counter path. Is NULL, if the path does not specify a parent instance.
	//The string follows this structure in memory.
	SzParentInstance *uint16 // pointer to a string
	//Instance index specified in the counter path. Is 0, if the path does not specify an instance index.
	DwInstanceIndex uint32 // pointer to a string
	//Null-terminated string that contains the counter name. The string follows this structure in memory.
	SzCounterName *uint16 // pointer to a string
	//Help text that describes the counter. Is NULL if the source is a log file.
	SzExplainText *uint16 // pointer to a string
	//Start of the string data that is appended to the structure.
	DataBuffer [1]uint32 // pointer to an extra space
}

// The PDH_RAW_COUNTER structure returns the data as it was collected from the counter provider. No translation, formatting, or other interpretation is performed on the data
type PDH_RAW_COUNTER struct {
	// Counter status that indicates if the counter value is valid. Check this member before using the data in a calculation or displaying its value. For a list of possible values,
	// see https://docs.microsoft.com/windows/desktop/PerfCtrs/checking-pdh-interface-return-values
	CStatus uint32
	// Local time for when the data was collected
	TimeStamp FILETIME
	// First raw counter value.
	FirstValue int64
	// Second raw counter value. Rate counters require two values in order to compute a displayable value.
	SecondValue int64
	// If the counter type contains the PERF_MULTI_COUNTER flag, this member contains the additional counter data used in the calculation.
	// For example, the PERF_100NSEC_MULTI_TIMER counter type contains the PERF_MULTI_COUNTER flag.
	MultiCount uint32
}

type PDH_RAW_COUNTER_ITEM struct {
	// Pointer to a null-terminated string that specifies the instance name of the counter. The string is appended to the end of this structure.
	SzName *uint16
	//A PDH_RAW_COUNTER structure that contains the raw counter value of the instance
	RawValue PDH_RAW_COUNTER
}
