// +build windows

package win_eventlog

import (
	"encoding/xml"
	"strconv"
	"strings"
	"syscall"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"golang.org/x/sys/windows"
)

var sampleConfig = `
  ## Name of eventlog
  eventlog_name = "Application"
  xpath_query = "Event/System[EventID=999]"
`

type WinEventLog struct {
	EventlogName string `toml:"eventlog_name"`
	Query        string `toml:"xpath_query"`
	subscription EvtHandle
	buf          []byte
	Log          telegraf.Logger
}

var bufferSize = 1 << 14

var description = "Input plugin to collect Windows eventlog messages"

func (w *WinEventLog) Description() string {
	return description
}

func (w *WinEventLog) SampleConfig() string {
	return sampleConfig
}

func (w *WinEventLog) Gather(acc telegraf.Accumulator) error {

	var err error
	if w.subscription == 0 {
		w.subscription, err = w.Subscribe(w.EventlogName, w.Query)
		if err != nil {
			w.Log.Error("subscribing error:", err.Error())
		}
	}
	w.Log.Debug("subscription handle id:", w.subscription)

loop:
	for {
		events, err := w.FetchEvents(w.subscription)
		if err != nil {
			switch {
			case err == ERROR_NO_MORE_ITEMS:
				break loop
			case err != nil:
				w.Log.Error("getting events error:", err.Error())
				return err
			}
		}

		for _, event := range events {

			// Pass collected metrics
			acc.AddFields("win_eventlog",
				map[string]interface{}{
					"record_id":   event.EventRecordID,
					"event_id":    event.EventID,
					"description": strings.Join(event.Data, " "),
					"source":      event.Provider,
					"created":     event.TimeCreated,
				}, map[string]string{
					"level":         strconv.Itoa(int(event.Level)),
					"eventlog_name": w.EventlogName,
				})
		}
	}

	return nil
}

func (w *WinEventLog) Subscribe(logName, xquery string) (EvtHandle, error) {
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

func (w *WinEventLog) FetchEventHandles(subsHandle EvtHandle) ([]EvtHandle, error) {
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

func (w *WinEventLog) FetchEvents(subsHandle EvtHandle) ([]Event, error) {
	var events []Event

	eventHandles, err := w.FetchEventHandles(subsHandle)
	if err != nil {
		return nil, err
	}

	for _, eventHandle := range eventHandles {
		if eventHandle != 0 {
			eventXML, err := w.RenderEvent(eventHandle)
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

func (w *WinEventLog) RenderEvent(e EvtHandle) ([]byte, error) {
	var bufferUsed, propertyCount uint32

	err := _EvtRender(0, e, EvtRenderEventXml, uint32(len(w.buf)), &w.buf[0], &bufferUsed, &propertyCount)
	if err != nil {
		return nil, err
	}

	return DecodeUTF16(w.buf[:bufferUsed])
}

func CloseEvent(e EvtHandle) error {
	err := _EvtClose(e)
	if err != nil {
		return err
	}
	return nil
}

func init() {
	inputs.Add("win_eventlog", func() telegraf.Input {
		return &WinEventLog{buf: make([]byte, bufferSize)}
	})
}
