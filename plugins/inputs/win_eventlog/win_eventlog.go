// +build windows

package win_eventlog

import (
	"bytes"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
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
	bookmark     EvtHandle
	buf          []byte
	out          *bytes.Buffer
	Log          telegraf.Logger
}

var description = "Input plugin to collect Windows eventlog messages"

func (w *WinEventLog) Description() string {
	return description
}

func (w *WinEventLog) SampleConfig() string {
	return sampleConfig
}

func (w *WinEventLog) Gather(acc telegraf.Accumulator) error {

	if w.subscription == 0 {
		w.subscription, err = Subscribe(w.EventlogName, w.Query)
		if err != nil {
			w.Log.Error("subscribing error:", err.Error(), w.bookmark)
		}
	}
	w.Log.Debug("subscription handle id:", w.subscription)

loop:
	for {
		events, err := FetchEvents(w.subscription)
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
					"description": strings.Join(event.Data, "\n"),
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

func init() {
	inputs.Add("win_eventlog", func() telegraf.Input {
		return &WinEventLog{
			buf: make([]byte, renderBufferSize),
			out: new(bytes.Buffer),
		}
	})
}
