package traefik

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Traefik struct {
	Address                      string
	ResponseTimeout              internal.Duration
	IncludeStatusCodeMeasurement bool
	lastRequestTiming            float64
}

type HealthCheck struct {
	UptimeSec              float64   `json:"uptime_sec"`
	Unixtime               int64     `json:"unixtime"`
	TotalStatusCodeCount   HttpCodes `json:"total_status_code_count"`
	TotalCount             int       `json:"total_count"`
	TotalResponseTimeSec   float64   `json:"total_response_time_sec"`
	AverageResponseTimeSec float64   `json:"average_response_time_sec"`
}

type HttpCodes map[string]int

var sampleConfig = `
# Required Traefik server address, host and port (default: "127.0.0.1")
address = "http://127.0.0.1:8080"

# default is false. Setting to true can increase cardinality
include_status_code_measurement = true

# Additional tags
[inputs.traefik.tags]
  instance = "prod"
`

func (t *Traefik) Description() string {
	return "Gather health check status from services registered in Traefik"
}

func (t *Traefik) SampleConfig() string {
	return sampleConfig
}

func (t *Traefik) submitStatusCodeMeasurement(acc telegraf.Accumulator, check *HealthCheck, tags map[string]string, fields map[string]interface{}) error {
	fields["total_count"] = check.TotalCount
	fields["uptime_sec"] = check.UptimeSec
	fields["unixtime"] = check.Unixtime

	for key, value := range check.TotalStatusCodeCount {
		newTags := copyTags(tags)
		newFields := copyFields(fields)
		newTags["status_code"] = key
		newFields["count"] = value
		acc.AddFields("traefik_status_codes", newFields, newTags)
	}

	return nil
}

func copyFields(m map[string]interface{}) map[string]interface{} {
	fields := make(map[string]interface{})
	for k, v := range m {
		fields[k] = v
	}
	return fields
}
func copyTags(m map[string]string) map[string]string {
	tags := make(map[string]string)
	for k, v := range m {
		tags[k] = v
	}
	return tags
}

func (t *Traefik) submitPrimaryMeasurement(acc telegraf.Accumulator, check *HealthCheck, tags map[string]string, fields map[string]interface{}) error {
	newTags := copyTags(tags)
	newFields := copyFields(fields)

	for key, value := range check.TotalStatusCodeCount {
		newFields[fmt.Sprintf("status_code_%v", key)] = value
	}

	newFields["total_response_time_sec"] = check.TotalResponseTimeSec
	newFields["average_response_time_sec"] = check.AverageResponseTimeSec
	newFields["total_count"] = check.TotalCount
	newFields["uptime_sec"] = check.UptimeSec
	newFields["unixtime"] = check.Unixtime

	acc.AddFields("traefik", newFields, newTags)
	return nil
}

func (t *Traefik) Gather(acc telegraf.Accumulator) error {
	client := &http.Client{
		Timeout: t.ResponseTimeout.Duration,
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("%v/health", t.Address), nil)
	if err != nil {
		return err
	}
	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	healthData := &HealthCheck{}
	json.NewDecoder(resp.Body).Decode(&healthData)
	t.lastRequestTiming = time.Since(start).Seconds()

	tags := map[string]string{"server": t.Address}
	fields := map[string]interface{}{"health_response_time_sec": t.lastRequestTiming}

	t.submitPrimaryMeasurement(acc, healthData, tags, fields)
	if t.IncludeStatusCodeMeasurement {
		t.submitStatusCodeMeasurement(acc, healthData, tags, fields)
	}

	return nil
}

func init() {
	inputs.Add("traefik", func() telegraf.Input {
		return &Traefik{
			Address: "127.0.0.1:8080",
		}
	})
}
