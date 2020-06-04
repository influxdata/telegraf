// +build windows

package win_eventlog

import (
	"bytes"
	"fmt"
	"unicode/utf16"
	"unicode/utf8"

	"encoding/xml"
	"syscall"

	"golang.org/x/sys/windows"
)

func DecodeUTF16(b []byte) ([]byte, error) {

	if len(b)%2 != 0 {
		return nil, fmt.Errorf("Must have even length byte slice")
	}

	u16s := make([]uint16, 1)

	ret := &bytes.Buffer{}

	b8buf := make([]byte, 4)

	lb := len(b)
	for i := 0; i < lb; i += 2 {
		u16s[0] = uint16(b[i]) + (uint16(b[i+1]) << 8)
		r := utf16.Decode(u16s)
		n := utf8.EncodeRune(b8buf, r[0])
		ret.Write(b8buf[:n])
	}

	return ret.Bytes(), nil
}

func Subscribe(logName, xquery string) (EvtHandle, error) {
	var logNamePtr, xqueryPtr *uint16

	sigEvent, err := windows.CreateEvent(nil, 0, 0, nil)
	if err != nil {
		return 0, err
	}
	defer windows.CloseHandle(sigEvent)

	logNamePtr, err = syscall.UTF16PtrFromString(logName)
	if err != nil {
		return 0, err
	}

	xqueryPtr, err = syscall.UTF16PtrFromString(xquery)
	if err != nil {
		return 0, err
	}

	subsHandle, err := _EvtSubscribe(0, uintptr(sigEvent), logNamePtr, xqueryPtr, 0, 0, 0, EvtSubscribeToFutureEvents)
	if err != nil {
		return 0, err
	}

	return subsHandle, nil
}

func FetchEventHandles(subsHandle EvtHandle) ([]EvtHandle, error) {
	var eventsNumber uint32
	var evtReturned uint32

	eventsNumber = 5

	eventHandles := make([]EvtHandle, eventsNumber)

	err := _EvtNext(subsHandle, eventsNumber, &eventHandles[0], 0, 0, &evtReturned)
	if err != nil {
		if err == ERROR_INVALID_OPERATION && evtReturned == 0 {
			return nil, ERROR_NO_MORE_ITEMS
		}
		return nil, err
	}

	return eventHandles[:evtReturned], nil
}

func FetchEvents(subsHandle EvtHandle) ([]Event, error) {
	var events []Event

	eventHandles, err := FetchEventHandles(subsHandle)
	if err != nil {
		return nil, err
	}

	for _, eventHandle := range eventHandles {
		if eventHandle != 0 {
			eventXML, err := RenderEvent(eventHandle)
			if err != nil {
				return nil, err
			}

			event := Event{}
			xml.Unmarshal(eventXML, &event)

			events = append(events, event)
		}
	}

	for i := 0; i < len(eventHandles); i++ {
		err := CloseEvent(eventHandles[i])
		if err != nil {
			return events, err
		}
	}
	return events, nil
}

func RenderEvent(e EvtHandle) ([]byte, error) {
	bufferSize := 1 << 14
	renderBuffer := make([]byte, bufferSize)
	var bufferUsed, propertyCount uint32

	err := _EvtRender(0, e, EvtRenderEventXml, uint32(len(renderBuffer)), &renderBuffer[0], &bufferUsed, &propertyCount)
	if err != nil {
		return nil, err
	}

	return DecodeUTF16(renderBuffer[:bufferUsed])
}

func QueryEventHandles(logName, xquery string) ([]EvtHandle, error) {
	// TODO
	return nil, nil
}

func QueryEvents(logName, xquery string) ([]Event, error) {
	// TODO
	return nil, nil
}

func CloseEvent(e EvtHandle) error {
	err := _EvtClose(e)
	if err != nil {
		return err
	}
	return nil
}
