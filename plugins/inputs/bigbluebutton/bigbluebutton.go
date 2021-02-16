package bigbluebutton

import (
	"crypto/sha1"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/common/proxy"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type BigBlueButton struct {
	URL              string `toml:"url"`
	PathPrefix       string `toml:"path_prefix"`
	SecretKey        string `toml:"secret_key"`
	getMeetingsURL   string
	getRecordingsURL string

	tls.ClientConfig
	proxy.HTTPProxy
	client *http.Client
}

var defaultPathPrefix = "/bigbluebutton"

var sampleConfig = `
	## Required BigBlueButton server url
	url = "http://localhost:8090"

	## BigBlueButton path prefix. Default is "/bigbluebutton"
	# path_prefix = "/bigbluebutton"

	## Required BigBlueButton secret key
	# secret_key =

	## Optional HTTP Proxy support
	# http_proxy_url = ""

	## Optional TLS Config
	# tls_ca = "/etc/telegraf/ca.pem"
	# tls_cert = "/etc/telegraf/cert.pem"
	# tls_key = "/etc/telegraf/key.pem"
	## Use TLS but skip chain & host verification
	# insecure_skip_verify = false
`

func (b *BigBlueButton) Init() error {
	if b.SecretKey == "" {
		return fmt.Errorf("BigBlueButton secret key is required")
	}

	if b.PathPrefix == "" {
		b.PathPrefix = defaultPathPrefix
	}

	b.getMeetingsURL = b.getURL("getMeetings")
	b.getRecordingsURL = b.getURL("getRecordings")

	tlsCfg, err := b.ClientConfig.TLSConfig()
	if err != nil {
		return err
	}

	proxy, err := b.HTTPProxy.Proxy()
	if err != nil {
		return err
	}

	transport := &http.Transport{
		TLSClientConfig: tlsCfg,
		Proxy:           proxy,
	}

	b.client = &http.Client{
		Transport: transport,
	}

	return nil
}

func (b *BigBlueButton) SampleConfig() string {
	return sampleConfig
}

func (b *BigBlueButton) Description() string {
	return "Gather BigBlueButton web conferencing server metrics"
}

func (b *BigBlueButton) Gather(acc telegraf.Accumulator) error {
	if err := b.gatherMeetings(acc); err != nil {
		return err
	}

	return b.gatherRecordings(acc)
}

// BigBlueButton uses an authentication based on a SHA1 checksum processed from api call name and server secret key
func (b *BigBlueButton) checksum(apiCallName string) []byte {
	hash := sha1.New()
	hash.Write([]byte(fmt.Sprintf("%s%s", apiCallName, b.SecretKey)))
	return hash.Sum(nil)
}

func (b *BigBlueButton) getURL(apiCallName string) string {
	endpoint := fmt.Sprintf("%s/api/%s", b.PathPrefix, apiCallName)
	return fmt.Sprintf("%s%s?checksum=%x", b.URL, endpoint, b.checksum(apiCallName))
}

// Call BBB server api
func (b *BigBlueButton) api(url string) ([]byte, error) {
	resp, err := b.client.Get(url)

	if err != nil || resp.StatusCode != 200 {
		return nil, fmt.Errorf("error getting bbb metrics: %s status %d", err, resp.StatusCode)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	return body, nil
}

func (b *BigBlueButton) gatherMeetings(acc telegraf.Accumulator) error {
	body, err := b.api(b.getMeetingsURL)
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
	body, err := b.api(b.getRecordingsURL)
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
