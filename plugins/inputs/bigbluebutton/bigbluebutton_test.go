package bigbluebutton

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

var emptyState = false

func getXMLResponse(requestURI string) ([]byte, int) {
	apiName := strings.Split(strings.TrimPrefix(requestURI, "/bigbluebutton/api/"), "?")[0]
	xmlFile := fmt.Sprintf("./testdata/%s.xml", apiName)

	if emptyState {
		xmlFile = fmt.Sprintf("%s.empty_state", xmlFile)
	}

	code := 200
	_, err := os.Stat(xmlFile)
	if err != nil {
		return nil, 404
	}

	b, _ := ioutil.ReadFile(xmlFile)
	return b, code
}

// return mocked HTTP server
func getHTTPServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, code := getXMLResponse(r.RequestURI)
		w.WriteHeader(code)
		if code == 200 {
			w.Header().Set("Content-Type", "text/xml")
			w.Write(body)
		} else {
			w.Write([]byte(""))
		}
	}))
}

func getPlugin(url string) BigBlueButton {
	return BigBlueButton{
		URL:       url,
		SecretKey: "OxShRR1sT8FrJZq",
	}
}

func TestBigBlueButton(t *testing.T) {
	s := getHTTPServer()
	defer s.Close()

	plugin := getPlugin(s.URL)
	plugin.Init()
	acc := &testutil.Accumulator{}
	plugin.Gather(acc)

	require.Empty(t, acc.Errors)

	meetingsRecord := map[string]uint64{
		"participant_count":       15,
		"listener_count":          12,
		"voice_participant_count": 4,
		"video_count":             1,
		"active_recording":        0,
	}

	recordingsRecord := map[string]uint64{
		"recordings_count":           2,
		"published_recordings_count": 1,
	}

	tags := make(map[string]string)

	expected := []telegraf.Metric{
		testutil.MustMetric("bigbluebutton_meetings", tags, toStringMapInterface(meetingsRecord), time.Unix(0, 0)),
		testutil.MustMetric("bigbluebutton_recordings", tags, toStringMapInterface(recordingsRecord), time.Unix(0, 0)),
	}

	acc.Wait(len(expected))

	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
}

func TestBigBlueButtonEmptyState(t *testing.T) {
	emptyState = true
	s := getHTTPServer()
	defer s.Close()

	plugin := getPlugin(s.URL)
	plugin.Init()
	acc := &testutil.Accumulator{}
	plugin.Gather(acc)

	require.Empty(t, acc.Errors)

	meetingsRecord := map[string]uint64{
		"participant_count":       0,
		"listener_count":          0,
		"voice_participant_count": 0,
		"video_count":             0,
		"active_recording":        0,
	}

	recordingsRecord := map[string]uint64{
		"recordings_count":           0,
		"published_recordings_count": 0,
	}

	tags := make(map[string]string)

	expected := []telegraf.Metric{
		testutil.MustMetric("bigbluebutton_meetings", tags, toStringMapInterface(meetingsRecord), time.Unix(0, 0)),
		testutil.MustMetric("bigbluebutton_recordings", tags, toStringMapInterface(recordingsRecord), time.Unix(0, 0)),
	}

	acc.Wait(len(expected))

	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
}
