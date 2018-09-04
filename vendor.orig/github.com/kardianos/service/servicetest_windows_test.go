// Copyright 2016 Lawrence Woodman <lwoodman@vlifesystems.com>
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package service_test

import (
	"os"
	"syscall"
	"testing"
)

func interruptProcess(t *testing.T) {
	dll, err := syscall.LoadDLL("kernel32.dll")
	if err != nil {
		t.Fatalf("LoadDLL(\"kernel32.dll\") err: %s", err)
	}
	p, err := dll.FindProc("GenerateConsoleCtrlEvent")
	if err != nil {
		t.Fatalf("FindProc(\"GenerateConsoleCtrlEvent\") err: %s", err)
	}
	// Send the CTRL_BREAK_EVENT to a console process group that shares
	// the console associated with the calling process.
	// https://msdn.microsoft.com/en-us/library/windows/desktop/ms683155(v=vs.85).aspx
	pid := os.Getpid()
	r1, _, err := p.Call(syscall.CTRL_BREAK_EVENT, uintptr(pid))
	if r1 == 0 {
		t.Fatalf("Call(CTRL_BREAK_EVENT, %d) err: %s", pid, err)
	}
}
