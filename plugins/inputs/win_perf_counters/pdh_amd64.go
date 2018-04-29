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
