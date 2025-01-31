//go:build windows

package win_eventlog

// event is the event entry representation
// Only the most common elements are processed, human-readable data is rendered in Message
// More info on schema, if there will be need to add more:
// https://docs.microsoft.com/en-us/windows/win32/wes/eventschema-elements
type event struct {
	Source        provider       `xml:"System>Provider"`
	EventID       int            `xml:"System>EventID"`
	Version       int            `xml:"System>Version"`
	Level         int            `xml:"System>Level"`
	Task          int            `xml:"System>Task"`
	Opcode        int            `xml:"System>Opcode"`
	Keywords      string         `xml:"System>Keywords"`
	TimeCreated   timeCreated    `xml:"System>TimeCreated"`
	EventRecordID int            `xml:"System>EventRecordID"`
	Correlation   correlation    `xml:"System>Correlation"`
	Execution     execution      `xml:"System>Execution"`
	Channel       string         `xml:"System>Channel"`
	Computer      string         `xml:"System>Computer"`
	Security      security       `xml:"System>Security"`
	UserData      userData       `xml:"UserData"`
	EventData     eventData      `xml:"EventData"`
	RenderingInfo *renderingInfo `xml:"RenderingInfo"`

	Message    string
	LevelText  string
	TaskText   string
	OpcodeText string
}

// userData Application-provided XML data
type userData struct {
	InnerXML []byte `xml:",innerxml"`
}

// eventData Application-provided XML data
type eventData struct {
	InnerXML []byte `xml:",innerxml"`
}

// provider is the event provider information
type provider struct {
	Name string `xml:"Name,attr"`
}

// correlation is used for the event grouping
type correlation struct {
	ActivityID        string `xml:"ActivityID,attr"`
	RelatedActivityID string `xml:"RelatedActivityID,attr"`
}

// execution Info for event
type execution struct {
	ProcessID   uint32 `xml:"ProcessID,attr"`
	ThreadID    uint32 `xml:"ThreadID,attr"`
	ProcessName string
}

// security Data for event
type security struct {
	UserID string `xml:"UserID,attr"`
}

// timeCreated field for event
type timeCreated struct {
	SystemTime string `xml:"SystemTime,attr"`
}

// renderingInfo is provided for events forwarded by Windows event Collector
// see https://learn.microsoft.com/en-us/windows/win32/api/winevt/nf-winevt-evtformatmessage#parameters
type renderingInfo struct {
	Message  string   `xml:"Message"`
	Level    string   `xml:"Level"`
	Task     string   `xml:"Task"`
	Opcode   string   `xml:"Opcode"`
	Channel  string   `xml:"Channel"`
	Provider string   `xml:"Provider"`
	Keywords []string `xml:"Keywords>Keyword"`
}
