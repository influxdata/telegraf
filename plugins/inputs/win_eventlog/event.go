// +build windows

package win_eventlog

// Event is the event entry representation
type Event struct {
	Provider      Provider    `xml:"System>Provider"`
	EventID       int         `xml:"System>EventID"`
	Version       int         `xml:"System>Version"`
	Level         int         `xml:"System>Level"`
	Task          int         `xml:"System>Task"`
	Opcode        int         `xml:"System>Opcode"`
	Keywords      string      `xml:"System>Keywords"`
	TimeCreated   TimeCreated `xml:"System>TimeCreated"`
	EventRecordID int         `xml:"System>EventRecordID"`
	Correlation   Correlation `xml:"System>Correlation"`
	Execution     Execution   `xml:"System>Execution"`
	Channel       string      `xml:"System>Channel"`
	Computer      string      `xml:"System>Computer"`
	Security      Security    `xml:"System>Security"`
	Data          []EventData `xml:"EventData>Data"`
}

// Provider is the Event provider information
type Provider struct {
	Name string `xml:"Name,attr"`
}

// Correlation is used for the event grouping
type Correlation struct {
	ActivityID string `xml:"ActivityID,attr"`
}

// EventData is a field with optional name
type EventData struct {
	Name  string `xml:"Name,attr"`
	Value string `xml:",chardata"`
}

// Execution Info for Event
type Execution struct {
	ProcessID int32 `xml:"ProcessID,attr"`
	ThreadID  int32 `xml:"ThreadID,attr"`
}

// Security Data for Event
type Security struct {
	UserID string `xml:"UserID,attr"`
}

// TimeCreated field for Event
type TimeCreated struct {
	SystemTime string `xml:"SystemTime,attr"`
}
