//go:build windows

// Package win_eventlog Input plugin to collect Windows Event Log messages
//
//revive:disable-next-line:var-naming
package win_eventlog

import "syscall"

// Event log error codes.
// https://msdn.microsoft.com/en-us/library/windows/desktop/ms681382(v=vs.85).aspx
const (
	//revive:disable:var-naming
	ERROR_INSUFFICIENT_BUFFER syscall.Errno = 122
	ERROR_NO_MORE_ITEMS       syscall.Errno = 259
	ERROR_INVALID_OPERATION   syscall.Errno = 4317
	//revive:enable:var-naming
)

// EvtSubscribeFlag defines the possible values that specify when to start subscribing to events.
type EvtSubscribeFlag uint32

// EVT_SUBSCRIBE_FLAGS enumeration
// https://msdn.microsoft.com/en-us/library/windows/desktop/aa385588(v=vs.85).aspx
const (
	EvtSubscribeToFutureEvents      EvtSubscribeFlag = 1
	EvtSubscribeStartAtOldestRecord EvtSubscribeFlag = 2
	EvtSubscribeStartAfterBookmark  EvtSubscribeFlag = 3
)

// EvtRenderFlag uint32
type EvtRenderFlag uint32

// EVT_RENDER_FLAGS enumeration
// https://msdn.microsoft.com/en-us/library/windows/desktop/aa385563(v=vs.85).aspx
const (
	// Render the event as an XML string. For details on the contents of the
	// XML string, see the Event schema.
	EvtRenderEventXml EvtRenderFlag = 1
	// Render bookmark
	EvtRenderBookmark EvtRenderFlag = 2
)
