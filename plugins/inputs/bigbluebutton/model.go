package bigbluebutton

import (
	"encoding/xml"
)

// MeetingsResponse is BigBlueButton XML global getMeetings api reponse type
type MeetingsResponse struct {
	XMLName    xml.Name `xml:"response"`
	ReturnCode string   `xml:"returncode"`
	MessageKey string   `xml:"messageKey"`
	Meetings   Meetings `xml:"meetings"`
}

// RecordingsResponse is BigBlueButton XML global getRecordings api response type
type RecordingsResponse struct {
	XMLName    xml.Name   `xml:"response"`
	ReturnCode string     `xml:"returncode"`
	MessageKey string     `xml:"messageKey"`
	Recordings Recordings `xml:"recordings"`
}

// Recordings is BigBlueButton XML recordings section
type Recordings struct {
	XMLName xml.Name    `xml:"recordings"`
	Values  []Recording `xml:"recording"`
}

// Recording is recording response containt information like state, record identifier, ...
type Recording struct {
	XMLName   xml.Name `xml:"recording"`
	RecordID  string   `xml:"recordID"`
	Published bool     `xml:"published"`
}

// Meetings is BigBlueButton XML meetings section
type Meetings struct {
	XMLName xml.Name  `xml:"meetings"`
	Values  []Meeting `xml:"meeting"`
}

// Meeting is a meeting response containing information like name, id, created time, created date, ...
type Meeting struct {
	XMLName               xml.Name `xml:"meeting"`
	ParticipantCount      int      `xml:"participantCount"`
	ListenerCount         int      `xml:"listenerCount"`
	VoiceParticipantCount int      `xml:"voiceParticipantCount"`
	VideoCount            int      `xml:"videoCount"`
	Recording             string   `xml:"recording"`
}
