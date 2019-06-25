package fireboard

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// Fireboard gathers statistics from the fireboard.io servers
type Fireboard struct {
	AuthToken []string
	UUID      []string

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

// Type fireboardStats represents the data that is received from Fireboard
type fireboardStats struct {
	ID          int64                    `json:id`
	Title       string                   `json:title`
	Owner       fireboardOwner           `json:owner`
	Created     string                   `json:created`
	UUID        string                   `json:uuid`
	HardwareID  string                   `json:hardware_id`
	LatestTemps []fireboardRealTimeTemps `json:latest_temps`
	LastTempLog string                   `json:last_templog`
	Model       string                   `json:model`
	ChanelCount int64                    `json:channel_count`
	DegreeType  int64                    `json:degreetype`
}

type fireboardOwner struct {
	Username    string                `json:username`
	Email       string                `json:email`
	FirstName   string                `json:first_name`
	LastName    string                `json:last_name`
	OwnerID     string                `json:id`
	UserProfile fireboardOwnerProfile `json:userprofile`
}

type fireboardOwnerProfile struct {
	Company          string `json:company`
	AlertSms         string `json:alert_sms`
	AlertEmails      string `json:alert_emails`
	NotificationTone string `json:notification_tone`
	User             string `json:user`
	Picture          string `json:picture`
	LastTemplog      string `json:last_templog`
	CommercialUser   string `json:commercial_user`
}

type fireboardRealTimeTemps struct {
	Channel    string  `json:channel`
	Created    string  `json:created`
	Temp       float64 `json:temp`
	DegreeType int64   `json:degreetype`
}

// A sample configuration to only gather stats from localhost, default port.
const sampleConfig = `
  # Specify auth token for your account
  authToken = ["ec7b2c09b5b2122151e934f6a69d6e3210be6cc6"]
`

// SampleConfig Returns a sample configuration for the plugin
func (r *Fireboard) SampleConfig() string {
	return sampleConfig
}

// Description Returns a description of the plugin
func (r *Fireboard) Description() string {
	return "Read metrics from fireboard.io servers"
}

// Gather Reads stats from all configured servers.
func (r *Fireboard) Gather(acc telegraf.Accumulator) error {
	// Default to a single server at localhost (default port) if none specified
	if len(r.AuthToken) == 0 {
		return nil
	}

	// Range over all servers, gathering stats. Returns early in case of any error.
	for _, s := range r.AuthToken {
		acc.AddError(r.gatherServer(s, acc))
	}

	return nil
}

// Gathers stats from a single server, adding them to the accumulator
func (r *Fireboard) gatherServer(s string, acc telegraf.Accumulator) error {
	// Parse the given URL to extract the server tag
	u, err := url.Parse(s)
	if err != nil {
		return fmt.Errorf("riak unable to parse given server url %s: %s", s, err)
	}

	// Perform the GET request to the riak /stats endpoint
	resp, err := r.client.Get(s + "/stats")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Successful responses will always return status code 200
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("riak responded with unexepcted status code %d", resp.StatusCode)
	}

	// Decode the response JSON into a new stats struct
	stats := &fireboardStats{}
	if err := json.NewDecoder(resp.Body).Decode(stats); err != nil {
		return fmt.Errorf("unable to decode riak response: %s", err)
	}

	// Build a map of tags
	tags := map[string]string{
		"title": stats.Title,
		"uuid":  stats.UUID,
	}

	// Build a map of field values
	fields := map[string]interface{}{
		"cpu_avg1":  stats.CpuAvg1,
		"cpu_avg15": stats.CpuAvg15,
		"cpu_avg5":  stats.CpuAvg5,
	}

	// Accumulate the tags and values
	acc.AddFields("fireboard", fields, tags)

	return nil
}

func init() {
	inputs.Add("fireboard", func() telegraf.Input {
		return NewFireboard()
	})
}
