//go:generate ../../../tools/readme_config_includer/generator
package airthings

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
	"golang.org/x/oauth2/clientcredentials"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"time"
)

// DO NOT REMOVE THE NEXT TWO LINES! This is required to embed the sampleConfig data.
//
//go:embed sample.conf
var sampleConfig string

type DeviceList struct {
	Devices []struct {
		ID         string        `json:"id"`
		DeviceType string        `json:"deviceType"`
		Sensors    []interface{} `json:"sensors"`
		Segment    struct {
			ID      string `json:"id"`
			Name    string `json:"name"`
			Started string `json:"started"`
			Active  bool   `json:"active"`
		} `json:"segment"`
		Location struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"location"`
	} `json:"devices"`
}

const (
	ID         = "id"
	DeviceType = "deviceType"
	Location   = "location"
	Segment    = "segment"
	Sensors    = "sensors"

	TagName           = "name"
	TagID             = ID
	TagDeviceType     = DeviceType
	TagSegmentID      = Segment + ".id"
	TagSegmentName    = Segment + ".name"
	TagSegmentActive  = Segment + ".active"
	TagSegmentStarted = Segment + ".started"

	timeParseFormat = "2006-01-02T15:04:05"
	timeZonedFormat = time.RFC3339
)

type Airthings struct {
	Log telegraf.Logger `toml:"-"`

	URL          string          `toml:"url"`
	ShowInactive bool            `toml:"showInactive"`
	ClientID     string          `toml:"client_id"`
	ClientSecret string          `toml:"client_secret"`
	TokenURL     string          `toml:"token_url"`
	Scopes       []string        `toml:"scopes"`
	Timeout      config.Duration `toml:"timeout"`
	TimeZone     string          `toml:"timeZone"`
	tls.ClientConfig
	cfg        *clientcredentials.Config
	httpClient *http.Client
	timer      time.Time
	location   *time.Location
}

func (m *Airthings) SampleConfig() string {
	return sampleConfig
}

func (m *Airthings) Description() string {
	return "Read metrics from the devices connected to the users Airthing account"
}

func (m *Airthings) Init() error {
	m.location, _ = time.LoadLocation("Local")
	if len(m.TimeZone) > 1 {
		location, err := time.LoadLocation(m.TimeZone)
		if err != nil {
			return err
		}
		m.location = location
	}
	m.timer = time.Now().In(m.location)
	m.Log.Infof("Init with location: %v", m.location)

	if m.cfg == nil {
		m.cfg = &clientcredentials.Config{
			ClientID:     m.ClientID,
			ClientSecret: m.ClientSecret,
			TokenURL:     m.TokenURL,
			Scopes:       m.Scopes,
		}
	}

	m.httpClient = m.cfg.Client(context.Background())

	return nil
}

func (m *Airthings) Gather(acc telegraf.Accumulator) error {
	m.Log.Infof("Gather duration since last run %s", time.Since(m.timer))
	m.timer = time.Now().In(m.location)
	deviceList, err := m.deviceList()
	if err != nil {
		return err
	}
	for _, device := range deviceList.Devices {
		var segStartedTime string
		zonedTime, err := enforceTimeZone(device.Segment.Started, m.location)
		if err != nil {
			m.Log.Errorf("time stamp: '%s' not parsable with format '%s' error: %v",
				device.Segment.Started, timeParseFormat, err)
			segStartedTime = device.Segment.Started
		} else {
			segStartedTime = zonedTime.In(m.location).Format(timeZonedFormat)
		}

		var airTags = map[string]string{
			TagName:           "airthings",
			TagID:             device.ID,
			TagDeviceType:     device.DeviceType,
			TagSegmentID:      device.Segment.ID,
			TagSegmentName:    device.Segment.Name,
			TagSegmentActive:  strconv.FormatBool(device.Segment.Active),
			TagSegmentStarted: segStartedTime,
		}

		var ts time.Time
		air, ts, err := m.deviceSamples(device.ID)
		if err != nil {
			return err
		}

		details, err := m.deviceDetails(device.ID)
		if err != nil {
			return err
		}
		for k, v := range *details {
			switch k {
			case ID:
			case DeviceType:
			case Location:
			case Segment:
			case Sensors:
			default:
				air[k] = v
			}
		}

		m.Log.Debugf("Add tags and fields %v <-> %v", airTags, air)
		acc.AddFields("airthings", air, airTags, ts)
	}
	return nil
}

func (m *Airthings) deviceSamples(deviceID string) (map[string]interface{}, time.Time, error) {
	var ts = time.Now().In(m.location)
	resp, err := m.doHTTPRequest(http.MethodGet, m.URL, "/devices/", deviceID, "/latest-samples")
	if err != nil {
		return nil, ts, err
	}
	var objmap map[string]json.RawMessage
	err = json.Unmarshal(resp, &objmap)
	if err != nil {
		return nil, ts, err
	}
	if dataVal, ok := objmap["data"]; ok {
		var data map[string]interface{}
		err = json.Unmarshal(dataVal, &data)
		if err != nil {
			return nil, ts, err
		}
		var air = make(map[string]interface{})
		for k, v := range data {
			switch k {
			case "time":
				// Get the time of the sample
				ts = time.Unix(int64(v.(float64)), 0).In(m.location)
			default:
				air[k] = v
			}
		}
		return air, ts, nil
	}
	return nil, ts, fmt.Errorf("no key 'data' in json data from sensor %s", deviceID)
}

func (m *Airthings) deviceDetails(deviceID string) (*map[string]interface{}, error) {
	var objmap map[string]interface{}
	resp, err := m.doHTTPRequest(http.MethodGet, m.URL, "/devices", deviceID)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(resp, &objmap)
	if err != nil {
		return nil, err
	}
	return &objmap, nil
}

func (m *Airthings) deviceList() (*DeviceList, error) {
	u, err := url.Parse(m.URL)
	if err != nil {
		m.Log.Errorf("error parsing url %v, %v", m.URL, err)
		return nil, err
	}

	values := u.Query()
	values.Add("showInactive", strconv.FormatBool(m.ShowInactive))
	u.RawQuery = values.Encode()

	resp, err := m.doHTTPRequest(http.MethodGet, u.String(), "/devices")
	if err != nil {
		return nil, err
	}
	var dl DeviceList
	if err := json.Unmarshal(resp, &dl); err != nil {
		return nil, err
	}
	m.Log.Debugf("device list %v", dl)
	return &dl, nil
}

func (m *Airthings) doHTTPRequest(httpMethod string, baseURL string, pathComponents ...string) ([]byte, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		m.Log.Errorf("error parsing url %v, %v", m.URL, err)
		return nil, err
	}
	for _, pc := range pathComponents {
		u.Path = path.Join(u.Path, pc)
	}
	r, err := http.NewRequest(httpMethod, u.String(), nil)
	if err != nil {
		m.Log.Errorf("error creating request: %v, %v", m.URL, err)
		return nil, err
	}
	m.Log.Debugf("%s request to %s", r.Proto, u)
	r.Header.Add("Accept", "application/json")
	resp, err := m.httpClient.Do(r)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Errorf("error closing reader (%v)", err)
		}
	}(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received HTTP status code %d from %q; expected 200",
			resp.StatusCode, m.URL)
	}
	buf := &bytes.Buffer{}
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func enforceTimeZone(inTime string, location *time.Location) (*time.Time, error) {
	t0, err := time.Parse(timeParseFormat, inTime)
	if err != nil {
		return nil, err
	}
	t0 = time.Date(t0.Year(), t0.Month(), t0.Day(), t0.Hour(), t0.Minute(), t0.Second(), t0.Nanosecond(), location)
	return &t0, nil
}

func init() {
	inputs.Add("airthings", func() telegraf.Input { return &Airthings{} })
}
