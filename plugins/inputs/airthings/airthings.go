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
		Id         string        `json:"id"`
		DeviceType string        `json:"deviceType"`
		Sensors    []interface{} `json:"sensors"`
		Segment    struct {
			Id      string `json:"id"`
			Name    string `json:"name"`
			Started string `json:"started"`
			Active  bool   `json:"active"`
		} `json:"segment"`
		Location struct {
			Id   string `json:"id"`
			Name string `json:"name"`
		} `json:"location"`
	} `json:"devices"`
}

const (
	Id         = "id"
	DeviceType = "deviceType"
	Location   = "location"
	Segment    = "segment"
	Sensors    = "sensors"

	TagName           = "name"
	TagId             = Id
	TagDeviceType     = DeviceType
	TagSegmentId      = Segment + ".id"
	TagSegmentName    = Segment + ".name"
	TagSegmentActive  = Segment + ".active"
	TagSegmentStarted = Segment + ".started"
)

type Airthings struct {
	Log telegraf.Logger `toml:"-"`

	URL           string          `toml:"url"`
	ShowInactive  bool            `toml:"showInactive"`
	ClientId      string          `toml:"client_id"`
	ClientSecret  string          `toml:"client_secret"`
	InsecureHttps bool            `toml:"insecureHttps"`
	TokenUrl      string          `toml:"token_url"`
	Scopes        []string        `toml:"scopes"`
	Timeout       config.Duration `toml:"timeout"`

	tls.ClientConfig
	cfg        *clientcredentials.Config
	httpClient *http.Client
	timer      time.Time
}

func (m *Airthings) SampleConfig() string {
	return sampleConfig
}

func (m *Airthings) Description() string {
	return "Read metrics from the devices connected to the users Airthing account"
}

func (m *Airthings) Init() error {

	m.Log.Info("Init")
	m.timer = time.Now()

	if m.cfg == nil {
		m.cfg = &clientcredentials.Config{
			ClientID:     m.ClientId,
			ClientSecret: m.ClientSecret,
			TokenURL:     m.TokenUrl,
			Scopes:       m.Scopes,
		}
	}

	//ctx := context.WithValue(oauth2.NoContext, oauth2.HTTPClient, myClient)
	m.httpClient = m.cfg.Client(context.Background())

	/*
		customTransport := http.DefaultTransport.(*http.Transport).Clone()
		customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		client := &http.Client{Transport: customTransport}
	*/

	return nil
}

func (m *Airthings) Gather(acc telegraf.Accumulator) error {
	m.Log.Debugf("Gather duration since last run %s", time.Since(m.timer))
	deviceList, err := m.deviceList()
	if err != nil {
		return err
	}
	for _, device := range deviceList.Devices {
		var airTags = map[string]string{
			TagName:           "airthings",
			TagId:             device.Id,
			TagDeviceType:     device.DeviceType,
			TagSegmentId:      device.Segment.Id,
			TagSegmentName:    device.Segment.Name,
			TagSegmentActive:  strconv.FormatBool(device.Segment.Active),
			TagSegmentStarted: device.Segment.Started,
		}
		var ts = time.Now()
		air, ts, err := m.deviceSamples(device.Id)
		if err != nil {
			return err
		}
		details, err := m.deviceDetails(device.Id)
		if err != nil {
			return err
		}
		for k, v := range *details {
			switch k {
			case Id:
			case DeviceType:
			case Location:
			case Segment:
			case Sensors:
			default:
				air[k] = v
			}
		}
		m.Log.Debugf("Add tags and fields %v <-> %v", airTags, air)
		acc.AddFields("airthings_connector", air, airTags, ts)
	}
	return nil
}

func (m *Airthings) deviceSamples(deviceId string) (map[string]interface{}, time.Time, error) {
	var ts = time.Now()
	resp, err := m.doHttpRequest(http.MethodGet, m.URL, "/devices/", deviceId, "/latest-samples")
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
				ts = time.Unix(int64(v.(float64)), 0)
			default:
				air[k] = v
			}
		}
		return air, ts, nil
	}
	return nil, ts, fmt.Errorf("No key 'data' in json data from sensor %s", deviceId)
}

func (m *Airthings) deviceDetails(deviceId string) (*map[string]interface{}, error) {
	var objmap map[string]interface{}
	resp, err := m.doHttpRequest(http.MethodGet, m.URL, "/devices", deviceId)
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
	u.Query().Add("showInactive", strconv.FormatBool(m.ShowInactive))
	/*
		if PathDevices == path {
			query := r.URL.Query()
			query.Add("showInactive", strconv.FormatBool(m.ShowInactive))
			r.URL.RawQuery = query.Encode()
		}
	*/
	resp, err := m.doHttpRequest(http.MethodGet, u.String(), "/devices")
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

func (m *Airthings) doHttpRequest(httpMethod string, baseUrl string, pathComponents ...string) ([]byte, error) {
	u, err := url.Parse(baseUrl)
	if err != nil {
		m.Log.Errorf("error parsing url %v, %v", m.URL, err)
		return nil, err
	}
	for _, pc := range pathComponents {
		u.Path = path.Join(u.Path, pc)
	}
	r, err := http.NewRequest(httpMethod, u.String(), nil)
	m.Log.Debugf("%s request to %s", r.Proto, u)
	r.Header.Add("Accept", "application/json")
	resp, err := m.httpClient.Do(r)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received HTTP status code %d from %q; expected 200",
			resp.StatusCode, m.URL)
	}
	buf := &bytes.Buffer{}
	buf.ReadFrom(resp.Body)
	return buf.Bytes(), nil
}

func init() {
	inputs.Add("airthings", func() telegraf.Input { return &Airthings{} })
}
