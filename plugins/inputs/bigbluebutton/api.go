package bigbluebutton

import (
	"crypto/sha1"
	"fmt"
	"io/ioutil"
	"net/http"
)

const getMeetingCallName = "getMeetings"
const getRecordingsCallName = "getRecordings"

// BigBlueButton uses an authentication based on a SHA1 checksum processed from api call name and server secret key
func (b *BigBlueButton) checksum(apiCallName string) []byte {
	hash := sha1.New()
	hash.Write([]byte(fmt.Sprintf("%s%s", apiCallName, b.SecretKey)))
	return hash.Sum(nil)
}

// Call BBB server api
func (b *BigBlueButton) api(apiCallName string) ([]byte, error) {
	endpoint := fmt.Sprintf("%s%s", b.APIEndpoint, apiCallName)

	url := fmt.Sprintf("%s%s?checksum=%x", b.URL, endpoint, b.checksum(apiCallName))
	resp, err := http.Get(url)

	if err != nil || resp.StatusCode != 200 {
		return nil, fmt.Errorf("error getting bbb metrics: %s", err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	return body, nil
}

// GetMeetings retrieve BBB server current meetings
func (b *BigBlueButton) GetMeetings() ([]byte, error) {
	return b.api(getMeetingCallName)
}

// GetRecordings retrieve BBB server recordings
func (b *BigBlueButton) GetRecordings() ([]byte, error) {
	return b.api(getRecordingsCallName)
}
