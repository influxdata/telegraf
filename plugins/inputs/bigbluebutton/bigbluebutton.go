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
}

var bbbConfig = `
	## Required BigBlueButton server url
	url = "http://localhost:8090"

	## Required BigBlueButton api endpoint
	api_endpoint = "/bigbluebutton/api/"

	## Required BigBlueButton secret key
	# secret_key =
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

	if err := b.gatherMeetings(acc); err != nil {
		return err
	}

	return b.gatherRecordings(acc)
}

func (b *BigBlueButton) gatherMeetings(acc telegraf.Accumulator) error {
	body, err := b.GetMeetings()
	if err != nil {
		return err
	}

	var response MeetingsResponse
	marshalErr := xml.Unmarshal(body, &response)
	if marshalErr != nil {
		return marshalErr
	}

	record := map[string]uint64{
		"active_recording":        0,
		"listener_count":          0,
		"participant_count":       0,
		"video_count":             0,
		"voice_participant_count": 0,
	}

	if response.MessageKey == "noMeetings" {
		acc.AddFields("bigbluebutton_meetings", toStringMapInterface(record), make(map[string]string))
		return nil
	}

	for i := 0; i < len(response.Meetings.Values); i++ {
		meeting := response.Meetings.Values[i]
		record["participant_count"] += meeting.ParticipantCount
		record["listener_count"] += meeting.ListenerCount
		record["voice_participant_count"] += meeting.VoiceParticipantCount
		record["video_count"] += meeting.VideoCount
		if meeting.Recording == true {
			record["active_recording"]++
		}
	}

	acc.AddFields("bigbluebutton_meetings", toStringMapInterface(record), make(map[string]string))
	return nil
}

func (b *BigBlueButton) gatherRecordings(acc telegraf.Accumulator) error {
	body, err := b.GetRecordings()
	if err != nil {
		return err
	}

	var response RecordingsResponse
	marshalErr := xml.Unmarshal(body, &response)
	if marshalErr != nil {
		return marshalErr
	}

	record := map[string]uint64{
		"recordings_count":           0,
		"published_recordings_count": 0,
	}

	if response.MessageKey == "noRecordings" {
		acc.AddFields("bigbluebutton_recordings", toStringMapInterface(record), make(map[string]string))
		return nil
	}

	for i := 0; i < len(response.Recordings.Values); i++ {
		recording := response.Recordings.Values[i]
		record["recordings_count"]++
		if recording.Published {
			record["published_recordings_count"]++
		}
	}

	acc.AddFields("bigbluebutton_recordings", toStringMapInterface(record), make(map[string]string))
	return nil
}

func toStringMapInterface(in map[string]uint64) map[string]interface{} {
	var m = map[string]interface{}{}
	for k, v := range in {
		m[k] = v
	}
	return m
}

func init() {
	inputs.Add("bigbluebutton", func() telegraf.Input {
		return &BigBlueButton{}
	})
}
