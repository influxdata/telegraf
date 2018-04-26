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
	"syscall"
	"unsafe"
)

type SYSTEMTIME struct {
	wYear         uint16
	wMonth        uint16
	wDayOfWeek    uint16
	wDay          uint16
	wHour         uint16
	wMinute       uint16
	wSecond       uint16
	wMilliseconds uint16
}

type FILETIME struct {
	dwLowDateTime  uint32
	dwHighDateTime uint32
}

var (
	// Library
	libkrnDll *syscall.DLL

	// Functions
	krn_FileTimeToSystemTime    *syscall.Proc
	krn_FileTimeToLocalFileTime *syscall.Proc
	krn_LocalFileTimeToFileTime *syscall.Proc
	krn_WideCharToMultiByte     *syscall.Proc
)

func init() {
	libkrnDll = syscall.MustLoadDLL("Kernel32.dll")

	krn_FileTimeToSystemTime = libkrnDll.MustFindProc("FileTimeToSystemTime")
	krn_FileTimeToLocalFileTime = libkrnDll.MustFindProc("FileTimeToLocalFileTime")
	krn_LocalFileTimeToFileTime = libkrnDll.MustFindProc("LocalFileTimeToFileTime")
	krn_WideCharToMultiByte = libkrnDll.MustFindProc("WideCharToMultiByte")
}

// CURRENTLY UNUSED: But may be useful in the future
//
// The windows native call for converting a 16-bit wide character string (UTF-16) to a null terminated string.
//
// Note: If you call the function and not pass in an out string, the return value will be the length of the
// input string.
// Example usage:
//   cc, err := WideCharToMultiByte(65001, 0, s, -1, nil, 0)
//   if err != nil {
// 	  fmt.Println("CONVERSION ERROR: ", err)
//   }
//
//	 fmt.Println("Length bytes: ", cc)
//   n, err := WideCharToMultiByte(65001, 0, s, 1<<29, &outStr[0], 1<<29)
//   if err != nil {
// 	   fmt.Println("CONVERSION ERROR: ", err)
//   }
//   fmt.Println("Converted bytes: ", n)
//
func WideCharToMultiByte(codePage uint32, dwFlags uint32, wchar *uint16, nwchar int32, str *byte, nstr int32) (nwrite int32, err error) {
	r0, _, e1 := krn_WideCharToMultiByte.Call(
		uintptr(codePage),
		uintptr(dwFlags),
		uintptr(unsafe.Pointer(str)),
		uintptr(nstr),
		uintptr(unsafe.Pointer(wchar)),
		uintptr(nwchar),
	)

	nwrite = int32(r0)
	if nwrite == 0 {
		if e1 != nil {
			err = errnoErr(e1.(syscall.Errno))
		} else {
			err = syscall.EINVAL
		}
	}

	return nwrite, err
}
