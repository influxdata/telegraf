// +build windows

package win_eventlog

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
	"syscall"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"golang.org/x/sys/windows"
)

var sampleConfig = `
  ## Telegraf should have Administrator permissions to subscribe for Windows Events

  ## LCID (Locale ID) for event rendering
  ## 1033 to force English language
  ## 0 to use default Windows locale
  # locale = 0

  ## Name of eventlog, used only if xpath_query is empty
  ## Example: "Application"
  eventlog_name = ""

  ## xpath_query can be in defined short form like "Event/System[EventID=999]"
  ## or you can form a xml query. Refer to the Consuming Events article:
  ## https://docs.microsoft.com/en-us/windows/win32/wes/consuming-events
  xpath_query = '''
  <QueryList>
    <Query Id="0" Path="Security">
      <Select Path="Security">*</Select>
      <Suppress Path="Security">*[System[( (EventID &gt;= 5152 and EventID &lt;= 5158) or EventID=5379 or EventID=4672)]]</Suppress>
    </Query>
    <Query Id="1" Path="Application">
      <Select Path="Application">*[System[(Level &lt; 4)]]</Select>
      <Select Path="OpenSSH/Admin">*[System[(Level &lt; 4)]]</Select>
      <Select Path="Windows PowerShell">*[System[(Level &lt; 4)]]</Select>
      <Select Path="Key Management Service">*[System[(Level &lt; 4)]]</Select>
      <Select Path="HardwareEvents">*[System[(Level &lt; 4)]]</Select>
    </Query>
    <Query Id="2" Path="Windows PowerShell">
      <Select Path="Windows PowerShell">*[System[(Level &lt; 4)]]</Select>
    </Query>
    <Query Id="3" Path="System">
      <Select Path="System">*</Select>
    </Query>
    <Query Id="4" Path="Setup">
      <Select Path="Setup">*</Select>
    </Query>
  </QueryList>
  '''
`

// WinEventLog config
type WinEventLog struct {
	Locale       uint32 `toml:"locale"`
	EventlogName string `toml:"eventlog_name"`
	Query        string `toml:"xpath_query"`
	subscription EvtHandle
	buf          []byte
	Log          telegraf.Logger
}

var bufferSize = 1 << 14

var description = "Input plugin to collect Windows Event Log messages"

// Description for win_eventlog
func (w *WinEventLog) Description() string {
	return description
}

// SampleConfig for win_eventlog
func (w *WinEventLog) SampleConfig() string {
	return sampleConfig
}

// Gather Windows Event Log entries
func (w *WinEventLog) Gather(acc telegraf.Accumulator) error {

	var err error
	if w.subscription == 0 {
		w.subscription, err = w.evtSubscribe(w.EventlogName, w.Query)
		if err != nil {
			w.Log.Error("Subscription error:", err.Error())
		}
	}
	w.Log.Debug("Subscription handle id:", w.subscription)

loop:
	for {
		events, err := w.fetchEvents(w.subscription)
		if err != nil {
			switch {
			case err == ERROR_NO_MORE_ITEMS:
				break loop
			case err != nil:
				w.Log.Error("Error getting events:", err.Error())
				return err
			}
		}

		for _, event := range events {
			tags := map[string]string{
				"source":        event.Provider.Name,
				"event_id":      strconv.Itoa(int(event.EventID)),
				"level":         strconv.Itoa(int(event.Level)),
				"keywords":      event.Keywords,
				"eventlog_name": event.Channel,
			}

			// Events, forwarded from another computer
			if event.Channel == "ForwardedEvents" {
				tags["computer"] = event.Computer
			}

			fields := map[string]interface{}{
				"version":      event.Version,
				"task":         event.Task,
				"record_id":    event.EventRecordID,
				"time_created": event.TimeCreated,
			}

			if event.Execution.ProcessID != 0 {
				fields["process_id"] = event.Execution.ProcessID
				fields["thread_id"] = event.Execution.ThreadID
				_, _, processName, err := GetFromSnapProcess(event.Execution.ProcessID)
				if err == nil {
					fields["process_name"] = processName
				}
			}

			if event.Opcode != 0 {
				fields["opcode"] = event.Opcode
			}

			if event.Correlation.ActivityID != "" {
				fields["activity_id"] = event.Correlation.ActivityID
			}

			if event.Security.UserID != "" {
				fields["user_id"] = event.Security.UserID
			}

			count := 1
			// Walk EventData values
			for _, data := range event.Data {
				var key string
				if len(data.Name) < 1 {
					// Use data_<count> name format for entries without the Name attribute
					key = fmt.Sprint("data_", count)
				} else {
					if _, exists := fields[strings.ToLower(data.Name)]; exists {
						// Add "data_" prefix for field names that can override existing values
						key = fmt.Sprint("data_", data.Name)
					} else {
						key = data.Name
					}
				}
				count++
				// Values can be an array of a zero-terminated strings
				splitZero := func(c rune) bool { return c == '\x00' }
				valueArray := strings.FieldsFunc(data.Value, splitZero)
				fields[key] = strings.Join(valueArray, ",")
			}

			// Pass collected metrics
			acc.AddFields("win_eventlog", fields, tags)
		}
	}

	return nil
}

func (w *WinEventLog) evtSubscribe(logName, xquery string) (EvtHandle, error) {
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

	subsHandle, err := _EvtSubscribe(0, uintptr(sigEvent), logNamePtr, xqueryPtr,
		0, 0, 0, EvtSubscribeToFutureEvents)
	if err != nil {
		return 0, err
	}

	return subsHandle, nil
}

func (w *WinEventLog) fetchEventHandles(subsHandle EvtHandle) ([]EvtHandle, error) {
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

func (w *WinEventLog) fetchEvents(subsHandle EvtHandle) ([]Event, error) {
	var events []Event

	eventHandles, err := w.fetchEventHandles(subsHandle)
	if err != nil {
		return nil, err
	}

	for _, eventHandle := range eventHandles {
		if eventHandle != 0 {
			eventXML, err := w.renderEvent(eventHandle)
			if err != nil {
				return nil, err
			}

			event := Event{}
			xml.Unmarshal([]byte(eventXML), &event)
			keywords, err := formatEventString(EvtFormatMessageKeyword, eventHandle,
				event.Provider.Name, w.Locale)
			if err != nil {
				// We will keep hex Keywords value just in case
				w.Log.Warnf("Error formatting keyword %s: %v", event.Keywords, err)
			} else {
				// Add comma-separated keyword strings
				event.Keywords = keywords
			}
			// w.Log.Debugf("Got event: %v", event)

			events = append(events, event)
		}
	}

	for i := 0; i < len(eventHandles); i++ {
		err := closeEvent(eventHandles[i])
		if err != nil {
			return events, err
		}
	}
	return events, nil
}

func (w *WinEventLog) renderEvent(e EvtHandle) ([]byte, error) {
	var bufferUsed, propertyCount uint32

	err := _EvtRender(0, e, EvtRenderEventXml,
		uint32(len(w.buf)), &w.buf[0], &bufferUsed, &propertyCount)
	if err != nil {
		return nil, err
	}

	return DecodeUTF16(w.buf[:bufferUsed])
}

func formatEventString(
	messageFlag EvtFormatMessageFlag,
	eventHandle EvtHandle,
	publisher string,
	lang uint32,
) (string, error) {
	ph, err := openPublisherMetadata(0, publisher, lang)
	if err != nil {
		return "", err
	}
	defer _EvtClose(ph)

	var bufferUsed uint32
	err = _EvtFormatMessage(ph, eventHandle, 0, 0, 0, messageFlag,
		0, nil, &bufferUsed)
	if err != nil && err != ERROR_INSUFFICIENT_BUFFER {
		return "", err
	}

	bufferUsed *= 2
	buffer := make([]byte, bufferUsed)
	// buffer[bufferUsed-1] = 0
	bufferUsed = 0

	err = _EvtFormatMessage(ph, eventHandle, 0, 0, 0, messageFlag,
		uint32(len(buffer)/2), &buffer[0], &bufferUsed)
	bufferUsed *= 2
	if err != nil {
		return "", err
	}

	result, err := DecodeUTF16(buffer[:bufferUsed])
	if err != nil {
		return "", err
	}

	var out string
	if messageFlag == EvtFormatMessageKeyword {
		// Keywords are returned as array of a zero-terminated strings
		splitZero := func(c rune) bool { return c == '\x00' }
		eventKeywords := strings.FieldsFunc(string(result), splitZero)
		out = strings.Join(eventKeywords, ",")
	} else {
		out = string(result)
	}
	return out, nil
}

// openPublisherMetadata opens a handle to the publisher's metadata. Close must
// be called on returned EvtHandle when finished with the handle.
func openPublisherMetadata(
	session EvtHandle,
	publisherName string,
	lang uint32,
) (EvtHandle, error) {
	p, err := syscall.UTF16PtrFromString(publisherName)
	if err != nil {
		return 0, err
	}

	h, err := _EvtOpenPublisherMetadata(session, p, nil, lang, 0)
	if err != nil {
		return 0, err
	}

	return h, nil
}

func closeEvent(e EvtHandle) error {
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
