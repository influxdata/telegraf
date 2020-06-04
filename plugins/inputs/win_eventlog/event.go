// +build windows

package win_eventlog

// Event is the event entry representation
type Event struct {
	EventRecordID int         `xml:"System>EventRecordID"`
	Provider      Provider    `xml:"System>Provider"`
	EventID       int         `xml:"System>EventID"`
	Level         int         `xml:"System>Level"`
	Data          []string    `xml:"EventData>Data"`
	TimeCreated   TimeCreated `xml:"System>TimeCreated"`
}

type Provider struct {
	Name string `xml:"Name,attr"`
}

type TimeCreated struct {
	SystemTime string `xml:"SystemTime,attr"`
}
