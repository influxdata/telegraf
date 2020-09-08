package fireboard

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// Fireboard gathers statistics from the fireboard.io servers
type Fireboard struct {
	AuthToken   string            `toml:"auth_token"`
	URL         string            `toml:"url"`
	HTTPTimeout internal.Duration `toml:"http_timeout"`

	client *http.Client
}

// NewFireboard return a new instance of Fireboard with a default http client
func NewFireboard() *Fireboard {
	tr := &http.Transport{ResponseHeaderTimeout: time.Duration(3 * time.Second)}
	client := &http.Client{
		Transport: tr,
		Timeout:   time.Duration(4 * time.Second),
	}
	return &Fireboard{client: client}
}

// RTT fireboardStats represents the data that is received from Fireboard
type RTT struct {
	Temp       float64 `json:"temp"`
	Channel    int64   `json:"channel"`
	Degreetype int     `json:"degreetype"`
	Created    string  `json:"created"`
}

type fireboardStats struct {
	Title       string `json:"title"`
	UUID        string `json:"uuid"`
	Latesttemps []RTT  `json:"latest_temps"`
}

// A sample configuration to only gather stats from localhost, default port.
const sampleConfig = `
  ## Specify auth token for your account
  auth_token = "invalidAuthToken"
  ## You can override the fireboard server URL if necessary
  # url = https://fireboard.io/api/v1/devices.json
  ## You can set a different http_timeout if you need to
  ## You should set a string using an number and time indicator
  ## for example "12s" for 12 seconds.
  # http_timeout = "4s"
`

// SampleConfig Returns a sample configuration for the plugin
func (r *Fireboard) SampleConfig() string {
	return sampleConfig
}

// Description Returns a description of the plugin
func (r *Fireboard) Description() string {
	return "Read real time temps from fireboard.io servers"
}

// Init the things
func (r *Fireboard) Init() error {

	if len(r.AuthToken) == 0 {
		return fmt.Errorf("You must specify an authToken")
	}
	if len(r.URL) == 0 {
		r.URL = "https://fireboard.io/api/v1/devices.json"
	}
	// Have a default timeout of 4s
	if r.HTTPTimeout.Duration == 0 {
		r.HTTPTimeout.Duration = time.Second * 4
	}

	r.client.Timeout = r.HTTPTimeout.Duration

	return nil
}

// Gather Reads stats from all configured servers.
func (r *Fireboard) Gather(acc telegraf.Accumulator) error {

	// Perform the GET request to the fireboard servers
	req, err := http.NewRequest("GET", r.URL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Token "+r.AuthToken)
	resp, err := r.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Successful responses will always return status code 200
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusForbidden {
			return fmt.Errorf("fireboard server responded with %d [Forbidden], verify your authToken", resp.StatusCode)
		}
		return fmt.Errorf("fireboard responded with unexepcted status code %d", resp.StatusCode)
	}
	// Decode the response JSON into a new stats struct
	var stats []fireboardStats
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		return fmt.Errorf("unable to decode fireboard response: %s", err)
	}
	// Range over all devices, gathering stats. Returns early in case of any error.
	for _, s := range stats {
		r.gatherTemps(s, acc)
	}
	return nil
}

// Return text description of degree type (scale)
func scale(n int) string {
	switch n {
	case 1:
		return "Celcius"
	case 2:
		return "Fahrenheit"
	default:
		return ""
	}
}

// Gathers stats from a single device, adding them to the accumulator
func (r *Fireboard) gatherTemps(s fireboardStats, acc telegraf.Accumulator) {
	// Construct lookup for scale values

	for _, t := range s.Latesttemps {
		tags := map[string]string{
			"title":   s.Title,
			"uuid":    s.UUID,
			"channel": strconv.FormatInt(t.Channel, 10),
			"scale":   scale(t.Degreetype),
		}
		fields := map[string]interface{}{
			"temperature": t.Temp,
		}
		acc.AddFields("fireboard", fields, tags)
	}
}

func init() {
	inputs.Add("fireboard", func() telegraf.Input {
		return NewFireboard()
	})
}
