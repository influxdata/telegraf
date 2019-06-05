package hobolink

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// dataPath is the default enpoint for metrics collection
const dataPath = "/data/json"

// customTime is our mapping for the timestamp returned by HOBOlink
const customTime = "2006-01-02 15:04:05"

// customTZ is the default timezone for processing the data points
var customTZ = "UTC"

// Description returns the plugin description.
func (h *HOBOlink) Description() string {
	return "Read stats from the HOBOlink API for a given user account"
}

// SampleConfig returns sample configuration for this plugin.
func (h *HOBOlink) SampleConfig() string {
	return sampleConfig
}

// NewHOBOlink returns a new instance of the HOBOlink collector
func NewHOBOlink() *HOBOlink {
	return &HOBOlink{
		HttpTimeout: internal.Duration{Duration: time.Second * 5},
		Server:      fmt.Sprintf("https://webservice.hobolink.com/restv2%s", dataPath),
		Window:      1,
		Timezone:    "UTC",
	}
}

func init() {
	inputs.Add("hobolink", func() telegraf.Input {
		return NewHOBOlink()
	})
}

// Gather reads the stats from the HOBOlink API and writes it to the Accumulator.
func (h *HOBOlink) Gather(acc telegraf.Accumulator) error {
	if h.client == nil {
		client, err := h.createHTTPClient()
		if err != nil {
			return err
		}

		h.client = client
		customTZ = h.Timezone
	}

	var wg sync.WaitGroup
	wg.Add(len(h.SerialNumbers))

	for _, serv := range h.SerialNumbers {
		go func(s string, acc telegraf.Accumulator) {
			defer wg.Done()

			if err := h.gatherStats(s, acc); err != nil {
				return
			}
		}(serv, acc)
	}

	wg.Wait()
	return nil
}

// parseJSON handles the request to the API and processing the return values
func (h *HOBOlink) parseJSON() (*Observations, error) {
	loc, err := time.LoadLocation(h.Timezone)
	if err != nil {
		return nil, err
	}

	// prep the JSON request
	t := time.Now()
	ts, err := time.ParseInLocation(customTime, t.Format("2006-01-02 15:04:05"), loc)
	if err != nil {
		return nil, err
	}

	payload, err := json.Marshal(APIRequest{
		Authentication: Authentication{
			User:     h.User,
			Password: h.Password,
			Token:    h.Token,
		},
		Query: Query{
			EndDateTime:   ts.Format("2006-01-02 15:04:05"),
			StartDateTime: ts.Add(-(time.Duration(h.Window)) * time.Hour).Format("2006-01-02 15:04:05"),
			Loggers:       h.SerialNumbers,
		},
	})
	if err != nil {
		return nil, err
	}

	// enable the request
	req, err := http.NewRequest("POST", h.Server, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}

	// set the additional headers
	req.Header.Set("content-type", "application/json")
	req.Header.Set("accept", "application/json")

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var v APIError
		if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
			return nil, fmt.Errorf("hobolink: API status %d, expected %d", resp.StatusCode, http.StatusOK)
		}

		return nil, fmt.Errorf("hobolink: API status %d, expected %d; msg: %s", resp.StatusCode, http.StatusOK, v.Message)
	}

	// parse the response body to our payload
	var v Observations
	if err = json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return nil, err
	}

	return &v, nil
}

// gatherStats does the actual collection of the stats from the endpoint
func (h *HOBOlink) gatherStats(serial string, acc telegraf.Accumulator) error {
	v, err := h.parseJSON()
	if err != nil {
		return err
	}

	for _, obs := range v.ObservationList {
		tags := map[string]string{
			"serial":        obs.LoggerSerialNumber,
			"sensor_serial": obs.SensorSerialNumber,
			"channel":       fmt.Sprintf("%d", obs.ChannelNumber),
			"data_type":     obs.DataType,
			"si_unit":       obs.SIUnit,
			"us_unit":       obs.USUnit,
			"scaled_unit":   obs.ScaluedUnit,
		}

		fields := map[string]interface{}{
			"si_value":     obs.SIValue,
			"us_value":     obs.USValue,
			"scaled_value": obs.ScaledValue,
		}

		acc.AddFields("hobolink", fields, tags, obs.Timestamp)
	}

	return nil
}

// createHTTPClient is a helper to generate the HTTP Client connection for our endpoint
// TODO: check if TLS support available
func (h *HOBOlink) createHTTPClient() (*http.Client, error) {
	tr := &http.Transport{
		ResponseHeaderTimeout: h.HttpTimeout.Duration,
		TLSClientConfig:       &tls.Config{},
	}

	client := &http.Client{
		Transport: tr,
		Timeout:   h.HttpTimeout.Duration,
	}

	return client, nil
}

// UnmarshalJSON to handle our custom timestamp returned from the HOBOlink API
func (o *Observation) UnmarshalJSON(data []byte) error {
	type Alias Observation
	aux := &struct {
		Timestamp string `json:"timestamp"`
		*Alias
	}{
		Alias: (*Alias)(o),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	loc, err := time.LoadLocation(customTZ)
	if err != nil {
		return err
	}

	ts, err := time.ParseInLocation(fmt.Sprintf("%sZ", customTime), aux.Timestamp, loc)
	if err != nil {
		return err
	}

	o.Timestamp = ts
	return nil
}

const sampleConfig = `
  ## NOTE: this plugin assumes a one-hour window for query for each of the
  ##       serial numbers outlined in the configuration

  ## specify the URL of the HOBOlink API (e.g. https://webservice.hobolink.com/restv2)
  server = ""

  ## specify the username of the account to query from
  user = ""

  ## specify the password of the account to query from
  password = ""

  ## specify the token of the account to query from
  token = ""

  ## specify the list of serial numbers to query within the specified account
  ## NOTE: if this list is empty, this plugin will query all devices 
  ##       as found for the given user; if this list is non-empty, _only_ the
  ##       provided list of serial numbers will be queried
  serial_numbers = [""]

  ## Define the window, in hours, of metrics to query from the API
  ## NOTE: default is 1 hour
  window = 12

  ## Timeout for HTTP requests to the HOBOlink API URL 
  # http_timeout = "5s"

  ## Timezone allows you to provide an override for timestamps that
  ## don't already include an offset
  ## e.g. 04-06-2016 12:41:45 
  ##
  ## Default: "" which renders UTC
  ## Options are as follows:
  ##   1. Local             -- interpret based on machine localtime
  ##   2. "Canada/Eastern"  -- Unix TZ values like those found in https://en.wikipedia.org/wiki/List_of_tz_database_time_zones
  ##   3. UTC               -- or blank/unspecified, will return timestamp in UTC
  # timezone = "Canada/Eastern"
`

// HOBOlink is a plugin to read stats from a list of IoT RX Sensors via the HOBOlink API
// NOTE: this agent is meant to collect from a single source user account.
type HOBOlink struct {
	User          string
	Password      string
	Server        string
	Token         string
	SerialNumbers []string
	HttpTimeout   internal.Duration
	Window        int64
	Timezone      string

	client *http.Client
}

// APIRequest is the set of values we use to format the query to the API
type APIRequest struct {
	Action         string         `json:"action"`
	Authentication Authentication `json:"authentication"`
	Query          Query          `json:"query"`
}

// Authentication handles the authentication portion of the request payload
type Authentication struct {
	User     string `json:"user"`
	Password string `json:"password"`
	Token    string `json:"token"`
}

// Query provides the query portion of the request payload
type Query struct {
	StartDateTime string   `json:"start_date_time"`
	EndDateTime   string   `json:"end_date_time"`
	Loggers       []string `json:"loggers"`
}

// Observations is the higher-level struct that encompasses all observations
type Observations struct {
	ObservationList []Observation `json:"observationList"`
	Message         string        `json:"message"`
}

// Observation is the individual observation retrieved
type Observation struct {
	LoggerSerialNumber string    `json:"logger_sn"`
	SensorSerialNumber string    `json:"sensor_sn"`
	ChannelNumber      int       `json:"channel_num"`
	Timestamp          time.Time `json:"timestamp"`
	DataType           string    `json:"data_type"`
	SIValue            float64   `json:"si_value"`
	SIUnit             string    `json:"si_unit"`
	USValue            float64   `json:"us_value"`
	USUnit             string    `json:"us_unit"`
	ScaledValue        float64   `json:"scaled_value"`
	ScaluedUnit        string    `json:"scaled_unit"`
}

// APIError is a struct to help serialize the possible error messages from the API
type APIError struct {
	Code        string `json:"code"`
	Message     string `json:"message"`
	Description string `json:"description"`
}
