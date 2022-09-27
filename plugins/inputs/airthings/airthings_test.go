package airthings

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/testutil"
)

const (
	HTTPContentTypeKey  = "Content-Type"
	HTTPContentTypeForm = "application/x-www-form-urlencoded"
	HTTPContentTypeJSON = "application/json"

	MesBattery           = "battery"
	MesHumidity          = "humidity"
	MesMold              = "mold"
	MesRelayDeviceType   = "relayDeviceType"
	MesRssi              = "rssi"
	MesTemp              = "temp"
	MesVoc               = "voc"
	MesRadonShortTermAvg = "radonShortTermAvg"
	MesCo2               = "co2"
	MesPressure          = "pressure"
)

var (
	airthings Airthings
)

func TestMain(m *testing.M) {
	ts := setupTestServer(m)
	//cert, err := tls.X509KeyPair(testcert.LocalhostCert, testcert.LocalhostKey)
	ts.EnableHTTP2 = false
	//ts.StartTLS()
	ts.Start()
	defer ts.Close()

	airthings = Airthings{
		URL:          ts.URL,
		ShowInactive: true,
		ClientID:     "clientid",
		ClientSecret: "clientsecret",
		TokenURL:     ts.URL + "/v1/token",
		Scopes:       []string{"read:device:current_values"},
		Timeout:      config.Duration(5 * time.Second),
		Log:          testutil.Logger{},
		TimeZone:     "UTC",
	}
	err := airthings.Init()
	if err != nil {
		airthings.Log.Errorf("Test error in init(): %v", err)
		return
	}
	airthings.Log.Debugf("Server listen to %s", ts.URL)
	code := m.Run()

	os.Exit(code)
}

func setupTestServer(m *testing.M) *httptest.Server {
	rexp := regexp.MustCompile(`^/devices/([0-9]*)`)
	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		airthings.Log.Debugf("request to: %v", r.URL)
		airthings.Log.Debugf("headers to: %v", r.Header)
		var deviceID = func() string {
			devIDTmp := rexp.FindStringSubmatch(r.URL.Path)
			if devIDTmp != nil && len(devIDTmp) > 1 {
				return devIDTmp[1]
			}
			return ""
		}()

		if r.Method == http.MethodPost && r.URL.Path == "/v1/token" {
			w.Header().Set(HTTPContentTypeKey, HTTPContentTypeForm)
			fmt.Fprint(w, "access_token=acc35570d3n&scope=user&token_type=bearer")
		} else if r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/devices") {
			w.Header().Set(HTTPContentTypeKey, HTTPContentTypeJSON)
			fmt.Fprint(w, readTestData("testdata/device_list.json"))
		} else if r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/latest-samples") {
			_, serialNumber := path.Split(path.Dir(r.URL.Path))
			w.Header().Set(HTTPContentTypeKey, HTTPContentTypeJSON)
			fmt.Fprint(w, readTestData("testdata/sample_"+serialNumber+".json"))
		} else if r.Method == http.MethodGet && len(deviceID) > 0 {
			w.Header().Set(HTTPContentTypeKey, HTTPContentTypeJSON)
			fmt.Fprint(w, readTestData("testdata/device_"+deviceID+".json"))
		} else {
			fmt.Printf("request --> %v", r)
			fmt.Fprintln(w, readTestData("testdata/error.json"))
		}
	}))
	return ts
}

func readTestData(testdataFilename string) string {
	content, err := os.ReadFile(testdataFilename)
	if err != nil {
		panic(err)
	}
	return string(content)
}

// Test get mock data from device
func TestGetDeviceListAndData(t *testing.T) {
	var acc testutil.Accumulator
	err := acc.GatherError(airthings.Gather)
	require.NoError(t, err)
	assertWaveMini(t, &acc)
	assertWavePlus(t, &acc)
	assertGen2(t, &acc)
}

func assertWaveMini(t *testing.T, acc *testutil.Accumulator) {
	acc.AssertContainsTaggedFields(t, "airthings",
		map[string]interface{}{
			MesBattery:         float64(78),
			MesHumidity:        float64(24),
			MesMold:            float64(0),
			MesRelayDeviceType: "hub",
			MesRssi:            float64(-51),
			MesTemp:            float64(22.9),
			MesVoc:             float64(161),
		},
		map[string]string{
			TagName:           "airthings",
			TagID:             "9990019182",
			TagDeviceType:     "WAVE_MINI",
			TagSegmentID:      "c6ddc7f5-e052-4969-8cca-f79f6a96b4f1",
			TagSegmentName:    "VOC",
			TagSegmentActive:  "true",
			TagSegmentStarted: "2120-09-12T07:20:28Z",
		})
}

func assertWavePlus(t *testing.T, acc *testutil.Accumulator) {
	acc.AssertContainsTaggedFields(t, "airthings",
		map[string]interface{}{
			MesBattery:           float64(100),
			MesCo2:               float64(1456),
			MesHumidity:          float64(41),
			MesPressure:          float64(1000.7),
			MesRadonShortTermAvg: float64(92),
			MesRelayDeviceType:   "hub",
			MesRssi:              float64(-64),
			MesTemp:              float64(19.4),
			MesVoc:               float64(191),
		},
		map[string]string{
			TagDeviceType:     "WAVE_PLUS",
			TagID:             "9990131459",
			TagName:           "airthings",
			TagSegmentActive:  "true",
			TagSegmentID:      "2bd162ce-4470-429f-8eff-4680ed5c6197",
			TagSegmentName:    "Bedroom",
			TagSegmentStarted: "2122-10-22T20:19:18Z",
		})
}

func assertGen2(t *testing.T, acc *testutil.Accumulator) {
	acc.AssertContainsTaggedFields(t, "airthings",
		map[string]interface{}{
			MesBattery:           float64(100),
			MesHumidity:          float64(23),
			MesRadonShortTermAvg: float64(165),
			MesRelayDeviceType:   "hub",
			MesRssi:              float64(-59),
			MesTemp:              float64(23.3),
		},
		map[string]string{
			TagDeviceType:     "WAVE_GEN2",
			TagID:             "9990012993",
			TagName:           "airthings",
			TagSegmentActive:  "true",
			TagSegmentID:      "3f2f2e23-f81d-46dd-8da6-9c5ed051b6e5",
			TagSegmentName:    "Basement",
			TagSegmentStarted: "2122-11-11T17:52:43Z",
		})
}
func TestEnforceTimeZone(t *testing.T) {
	timeUTC := time.Now().UTC()
	location, err := time.LoadLocation("Europe/Stockholm")
	if err != nil {
		t.Error(err)
	}

	timeUTCStr := timeUTC.Format(time.RFC3339)
	timeUTCStr = timeUTCStr[:len(timeUTCStr)-1] // Trim away the 'Z' TimeZone
	timeZoned, err := enforceTimeZone(timeUTCStr, location)
	if err != nil {
		t.Error(err)
	}

	_, offset := timeZoned.Zone()
	t.Logf("Test: inDate: '%s' zoned date: '%s' offset: '%v'\n", timeUTCStr, timeZoned, offset)

	if 0 != timeUTC.Unix()-(timeZoned.Unix()+int64(offset)) {
		t.Logf("Fail: Date: '%s' not equal to: '%s' \n", timeUTC.In(location).Format(time.RFC3339),
			timeZoned.In(location).Format(time.RFC3339))
		t.Fail()
	}
}
