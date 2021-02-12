package bigbluebutton

import (
	"encoding/xml"
	"fmt"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type BigBlueButton struct {
	URL         string `toml:"url"`
	APIEndpoint string `toml:"api_endpoint"`
	SecretKey   string `toml:"secret_key"`
	ServerName  string `toml:"server_name"`
}

var bbbConfig = `
	## Required BigBlueButton server url
	url = "http://localhost:8090"

	## Required BigBlueButton api endpoint
	api_endpoint = "/bigbluebutton/api/"

	## Required BigBlueButton secret key
	# secret_key =

	## Server name. Used as "server" tag if not empty. Otherwise use url
	# server_name = 
`

func (b *BigBlueButton) SampleConfig() string {
	return bbbConfig
}

func (b *BigBlueButton) Description() string {
	return "Gather BigBlueButton web conferencing server metrics"
}

func (b *BigBlueButton) Gather(acc telegraf.Accumulator) error {
	if b.SecretKey == "" {
		return fmt.Errorf("BigBlueButton secret key is required")
	}

	if b.ServerName == "" {
		b.ServerName = b.URL
	}

	meetingsErr := b.gatherMeetings(acc)
	recordingsErr := b.gatherRecordings(acc)

	if meetingsErr != nil {
		return meetingsErr
	}

	if recordingsErr != nil {
		return recordingsErr
	}

	return nil
}

func (b *BigBlueButton) gatherMeetings(acc telegraf.Accumulator) error {
	body, err := b.GetMeetings()
	if err != nil {
		return err
	}

	var response MeetingsResponse
	xml.Unmarshal(body, &response)

	participantCount := 0
	listenerCount := 0
	voiceParticipantCount := 0
	videoCount := 0
	activeRecording := 0

	if response.MessageKey == NoMeetingsMessageKey {
		b.sendMeetingsRecord(acc, b.meetingsRecord(0, 0, 0, 0, 0))
		return nil
	}

	for i := 0; i < len(response.Meetings.Values); i++ {
		meeting := response.Meetings.Values[i]
		participantCount += meeting.ParticipantCount
		listenerCount += meeting.ListenerCount
		voiceParticipantCount += meeting.VoiceParticipantCount
		videoCount += meeting.VideoCount
		if meeting.Recording == "true" {
			activeRecording++
		}
	}

	b.sendMeetingsRecord(acc, b.meetingsRecord(participantCount, listenerCount, voiceParticipantCount, videoCount, activeRecording))
	return nil
}

func (b *BigBlueButton) meetingsRecord(participantCount int, listenerCount int, voiceParticipantCount int, videoCount int, activeRecording int) map[string]interface{} {
	record := make(map[string]interface{})
	record["participant_count"] = participantCount
	record["listener_count"] = listenerCount
	record["voice_participant_count"] = voiceParticipantCount
	record["video_count"] = videoCount
	record["active_recording"] = activeRecording
	return record
}

func (b *BigBlueButton) recordingsRecord(recordingsCount int, publishedCount int) map[string]interface{} {
	record := make(map[string]interface{})
	record["recordings_count"] = recordingsCount
	record["published_recordings_count"] = publishedCount
	return record
}

func (b *BigBlueButton) sendMeetingsRecord(acc telegraf.Accumulator, record map[string]interface{}) {
	b.sendRecord(acc, "bigbluebutton_meetings", record)
}

func (b *BigBlueButton) sendRecordingsRecord(acc telegraf.Accumulator, record map[string]interface{}) {
	b.sendRecord(acc, "bigbluebutton_recordings", record)
}

func (b *BigBlueButton) gatherRecordings(acc telegraf.Accumulator) error {
	body, err := b.GetRecordings()
	if err != nil {
		return err
	}

	var response RecordingsResponse
	xml.Unmarshal(body, &response)

	if response.MessageKey == NoRecordingMessageKey {
		b.sendRecordingsRecord(acc, b.recordingsRecord(0, 0))
		return nil
	}

	recordingsCount := 0
	publishedCount := 0

	for i := 0; i < len(response.Recordings.Values); i++ {
		recording := response.Recordings.Values[i]
		recordingsCount++
		if recording.Published {
			publishedCount++
		}
	}

	b.sendRecordingsRecord(acc, b.recordingsRecord(recordingsCount, publishedCount))

	return nil
}

func (b *BigBlueButton) tags() map[string]string {
	tags := make(map[string]string)
	tags["server_name"] = b.ServerName
	return tags
}

func (b *BigBlueButton) sendRecord(acc telegraf.Accumulator, name string, record map[string]interface{}) {
	acc.AddFields(name, record, b.tags())
}

func init() {
	inputs.Add("bigbluebutton", func() telegraf.Input {
		return &BigBlueButton{}
	})
}
