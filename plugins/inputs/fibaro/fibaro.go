package fibaro

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const defaultTimeout = 5 * time.Second

// Fibaro contains connection information
type Fibaro struct {
	URL string `toml:"url"`

	// HTTP Basic Auth Credentials
	Username string `toml:"username"`
	Password string `toml:"password"`

	Timeout config.Duration `toml:"timeout"`

	client *http.Client
}

// LinkRoomsSections links rooms to sections
type LinkRoomsSections struct {
	Name      string
	SectionID uint16
}

// Sections contains sections informations
type Sections struct {
	ID   uint16 `json:"id"`
	Name string `json:"name"`
}

// Rooms contains rooms informations
type Rooms struct {
	ID        uint16 `json:"id"`
	Name      string `json:"name"`
	SectionID uint16 `json:"sectionID"`
}

// Devices contains devices informations
type Devices struct {
	ID         uint16 `json:"id"`
	Name       string `json:"name"`
	RoomID     uint16 `json:"roomID"`
	Type       string `json:"type"`
	Enabled    bool   `json:"enabled"`
	Properties struct {
		BatteryLevel *string     `json:"batteryLevel"`
		Dead         string      `json:"dead"`
		Energy       *string     `json:"energy"`
		Power        *string     `json:"power"`
		Value        interface{} `json:"value"`
		Value2       *string     `json:"value2"`
	} `json:"properties"`
}

// getJSON connects, authenticates and reads JSON payload returned by Fibaro box
func (f *Fibaro) getJSON(path string, dataStruct interface{}) error {
	var requestURL = f.URL + path

	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return err
	}

	req.SetBasicAuth(f.Username, f.Password)
	resp, err := f.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("response from url \"%s\" has status code %d (%s), expected %d (%s)",
			requestURL,
			resp.StatusCode,
			http.StatusText(resp.StatusCode),
			http.StatusOK,
			http.StatusText(http.StatusOK))
		return err
	}

	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&dataStruct)
	if err != nil {
		return err
	}

	return nil
}

// Gather fetches all required information to output metrics
func (f *Fibaro) Gather(acc telegraf.Accumulator) error {
	if f.client == nil {
		f.client = &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
			},
			Timeout: time.Duration(f.Timeout),
		}
	}

	var tmpSections []Sections
	err := f.getJSON("/api/sections", &tmpSections)
	if err != nil {
		return err
	}
	sections := map[uint16]string{}
	for _, v := range tmpSections {
		sections[v.ID] = v.Name
	}

	var tmpRooms []Rooms
	err = f.getJSON("/api/rooms", &tmpRooms)
	if err != nil {
		return err
	}
	rooms := map[uint16]LinkRoomsSections{}
	for _, v := range tmpRooms {
		rooms[v.ID] = LinkRoomsSections{Name: v.Name, SectionID: v.SectionID}
	}

	var devices []Devices
	err = f.getJSON("/api/devices", &devices)
	if err != nil {
		return err
	}

	for _, device := range devices {
		// skip device in some cases
		if device.RoomID == 0 ||
			!device.Enabled ||
			device.Properties.Dead == "true" ||
			device.Type == "com.fibaro.zwaveDevice" {
			continue
		}

		tags := map[string]string{
			"deviceId": strconv.FormatUint(uint64(device.ID), 10),
			"section":  sections[rooms[device.RoomID].SectionID],
			"room":     rooms[device.RoomID].Name,
			"name":     device.Name,
			"type":     device.Type,
		}
		fields := make(map[string]interface{})

		if device.Properties.BatteryLevel != nil {
			if fValue, err := strconv.ParseFloat(*device.Properties.BatteryLevel, 64); err == nil {
				fields["batteryLevel"] = fValue
			}
		}

		if device.Properties.Energy != nil {
			if fValue, err := strconv.ParseFloat(*device.Properties.Energy, 64); err == nil {
				fields["energy"] = fValue
			}
		}

		if device.Properties.Power != nil {
			if fValue, err := strconv.ParseFloat(*device.Properties.Power, 64); err == nil {
				fields["power"] = fValue
			}
		}

		if device.Properties.Value != nil {
			value := device.Properties.Value
			switch value {
			case "true":
				value = "1"
			case "false":
				value = "0"
			}

			if fValue, err := strconv.ParseFloat(value.(string), 64); err == nil {
				fields["value"] = fValue
			}
		}

		if device.Properties.Value2 != nil {
			if fValue, err := strconv.ParseFloat(*device.Properties.Value2, 64); err == nil {
				fields["value2"] = fValue
			}
		}

		acc.AddFields("fibaro", fields, tags)
	}

	return nil
}

func init() {
	inputs.Add("fibaro", func() telegraf.Input {
		return &Fibaro{
			Timeout: config.Duration(defaultTimeout),
		}
	})
}
