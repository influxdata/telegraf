//go:build windows

// Package win_eventlog Input plugin to collect Windows Event Log messages
//
//revive:disable-next-line:var-naming
package win_eventlog

// Event is the event entry representation
// Only the most common elements are processed, human-readable data is rendered in Message
// More info on schema, if there will be need to add more:
// https://docs.microsoft.com/en-us/windows/win32/wes/eventschema-elements
type Event struct {
	Source        Provider       `xml:"System>Provider"`
	EventID       int            `xml:"System>EventID"`
	Version       int            `xml:"System>Version"`
	Level         int            `xml:"System>Level"`
	Task          int            `xml:"System>Task"`
	Opcode        int            `xml:"System>Opcode"`
	Keywords      string         `xml:"System>Keywords"`
	TimeCreated   TimeCreated    `xml:"System>TimeCreated"`
	EventRecordID int            `xml:"System>EventRecordID"`
	Correlation   Correlation    `xml:"System>Correlation"`
	Execution     Execution      `xml:"System>Execution"`
	Channel       string         `xml:"System>Channel"`
	Computer      string         `xml:"System>Computer"`
	Security      Security       `xml:"System>Security"`
	UserData      UserData       `xml:"UserData"`
	EventData     EventData      `xml:"EventData"`
	RenderingInfo *RenderingInfo `xml:"RenderingInfo"`

	Message    string
	LevelText  string
	TaskText   string
	OpcodeText string
}

// UserData Application-provided XML data
type UserData struct {
	InnerXML []byte `xml:",innerxml"`
}

// EventData Application-provided XML data
type EventData struct {
	InnerXML []byte `xml:",innerxml"`
}

// Provider is the Event provider information
type Provider struct {
	Name string `xml:"Name,attr"`
}

// Correlation is used for the event grouping
type Correlation struct {
	ActivityID        string `xml:"ActivityID,attr"`
	RelatedActivityID string `xml:"RelatedActivityID,attr"`
}

// Execution Info for Event
type Execution struct {
	ProcessID   uint32 `xml:"ProcessID,attr"`
	ThreadID    uint32 `xml:"ThreadID,attr"`
	ProcessName string
}

// Security Data for Event
type Security struct {
	UserID string `xml:"UserID,attr"`
}

// TimeCreated field for Event
type TimeCreated struct {
	SystemTime string `xml:"SystemTime,attr"`
}

// RenderingInfo is provided for events forwarded by Windows Event Collector
// see https://learn.microsoft.com/en-us/windows/win32/api/winevt/nf-winevt-evtformatmessage#parameters
type RenderingInfo struct {
	Message  string   `xml:"Message"`
	Level    string   `xml:"Level"`
	Task     string   `xml:"Task"`
	Opcode   string   `xml:"Opcode"`
	Channel  string   `xml:"Channel"`
	Provider string   `xml:"Provider"`
	Keywords []string `xml:"Keywords>Keyword"`
}
