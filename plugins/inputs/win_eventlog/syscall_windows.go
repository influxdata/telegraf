//go:build windows

package win_eventlog

import "syscall"

// event log error codes.
// https://msdn.microsoft.com/en-us/library/windows/desktop/ms681382(v=vs.85).aspx
const (
	errInsufficientBuffer syscall.Errno = 122
	errNoMoreItems        syscall.Errno = 259
	errInvalidOperation   syscall.Errno = 4317
)

// evtSubscribeFlag defines the possible values that specify when to start subscribing to events.
type evtSubscribeFlag uint32

// EVT_SUBSCRIBE_FLAGS enumeration
// https://msdn.microsoft.com/en-us/library/windows/desktop/aa385588(v=vs.85).aspx
const (
	evtSubscribeToFutureEvents      evtSubscribeFlag = 1
	evtSubscribeStartAtOldestRecord evtSubscribeFlag = 2
	evtSubscribeStartAfterBookmark  evtSubscribeFlag = 3
)

// evtRenderFlag uint32
type evtRenderFlag uint32

// EVT_RENDER_FLAGS enumeration
// https://msdn.microsoft.com/en-us/library/windows/desktop/aa385563(v=vs.85).aspx
const (
	// Render the event as an XML string. For details on the contents of the XML string, see the event schema.
	evtRenderEventXML evtRenderFlag = 1
	// Render bookmark
	evtRenderBookmark evtRenderFlag = 2
)
