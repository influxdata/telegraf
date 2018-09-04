// Copyright (c) 2013 Couchbase, Inc.
// Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
// except in compliance with the License. You may obtain a copy of the License at
//   http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software distributed under the
// License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
// either express or implied. See the License for the specific language governing permissions
// and limitations under the License.

// +build windows

package platform

import "syscall"

// Hide console on windows without removing it unlike -H windowsgui.
func HideConsole(hide bool) {
	var k32 = syscall.NewLazyDLL("kernel32.dll")
	var cw = k32.NewProc("GetConsoleWindow")
	var u32 = syscall.NewLazyDLL("user32.dll")
	var sw = u32.NewProc("ShowWindow")
	hwnd, _, _ := cw.Call()
	if hwnd == 0 {
		return
	}
	if hide {
		var SW_HIDE uintptr = 0
		sw.Call(hwnd, SW_HIDE)
	} else {
		var SW_RESTORE uintptr = 9
		sw.Call(hwnd, SW_RESTORE)
	}
}
