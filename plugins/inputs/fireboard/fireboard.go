package fireboard

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// Fireboard gathers statistics from the fireboard.io servers
type Fireboard struct {
	AuthToken string
	URL       string

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
	Temp       float64
	Channel    int64
	Degreetype int64
	Created    string
}

type fireboardStats struct {
	ID           int64
	Title        string
	Created      string
	UUID         string
	HardwareID   string `json:"hardware_id"`
	Latesttemps  []RTT  `json:"latest_temps"`
	Lasttemplog  string `json:"last_templog"`
	Model        string
	Channelcount int64 `json:"channel_count"`
	Degreetype   int64
}

// A sample configuration to only gather stats from localhost, default port.
const sampleConfig = `
  # Specify auth token for your account
  # https://docs.fireboard.io/reference/restapi.html#Authentication
  # authToken = "b4bb6e6a7b6231acb9f71b304edb2274693d8849"
  #
  # You can override the fireboard server URL if necessary
  # URL = https://fireboard.io/api/v1/devices.json
  #
`

// SampleConfig Returns a sample configuration for the plugin
func (r *Fireboard) SampleConfig() string {
	return sampleConfig
}

// Description Returns a description of the plugin
func (r *Fireboard) Description() string {
	return "Read real time temps from fireboard.io servers"
}

// Gather Reads stats from all configured servers.
func (r *Fireboard) Gather(acc telegraf.Accumulator) error {
	// Default to a single server at localhost (default port) if none specified
	if len(r.AuthToken) == 0 {
		return fmt.Errorf("You must specify an authToken")
	}
	if len(r.URL) == 0 {
		r.URL = "https://fireboard.io/api/v1/devices.json"
	}

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
		acc.AddError(r.gatherTemps(s, acc))
	}
	return nil
}

// Gathers stats from a single device, adding them to the accumulator
func (r *Fireboard) gatherTemps(s fireboardStats, acc telegraf.Accumulator) error {

	for _, t := range s.Latesttemps {
		tags := map[string]string{
			"title":   s.Title,
			"uuid":    s.UUID,
			"channel": strconv.FormatInt(t.Channel, 10),
			"scale":   strconv.FormatInt(t.Degreetype, 10),
		}
		fields := map[string]interface{}{
			"temperature": t.Temp,
		}
		acc.AddFields("fireboard", fields, tags)
	}
	return nil
}

func init() {
	inputs.Add("fireboard", func() telegraf.Input {
		return NewFireboard()
	})
}
