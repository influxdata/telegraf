package traefik

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Traefik struct {
	Server   string
	Port     int
	Instance string
}

type HealthCheck struct {
	Pid                    int       `json:"pid"`
	Uptime                 string    `json:"uptime"`
	UptimeSec              float64   `json:"uptime_sec"`
	Time                   string    `json:"time"`
	Unixtime               int       `json:"unixtime"`
	StatusCodeCount        struct{}  `json:"status_code_count"`
	TotalStatusCodeCount   HttpCodes `json:"total_status_code_count"`
	Count                  int       `json:"count"`
	TotalCount             int       `json:"total_count"`
	TotalResponseTime      string    `json:"total_response_time"`
	TotalResponseTimeSec   float64   `json:"total_response_time_sec"`
	AverageResponseTime    string    `json:"average_response_time"`
	AverageResponseTimeSec float64   `json:"average_response_time_sec"`
}

type HttpCodes map[string]int

var sampleConfig = `
	## Required Traefik server address (default: "127.0.0.1")
	# server = "127.0.0.1"
	## Required Traefik port (default "8080")
	# port = 8080
	## Required Traefik instance name (default: "default")
	# instance = "default"
	`

func (t *Traefik) Description() string {
	return "Gather health check status from services registered in Traefik"
}

func (t *Traefik) SampleConfig() string {
	return sampleConfig
}

func (t *Traefik) GatherHealthCheck(acc telegraf.Accumulator, check HealthCheck) {
	records := make(map[string]interface{})
	tags := make(map[string]string)

	for key, value := range check.TotalStatusCodeCount {
		records[key] = value
	}

	records["total_response_time_sec"] = check.TotalResponseTimeSec
	records["average_reponse_time_sec"] = check.AverageResponseTimeSec
	records["total_count"] = check.TotalCount

	tags["instance"] = t.Instance

	acc.AddFields("traefik_healthchecks", records, tags)
}

func (t *Traefik) Gather(acc telegraf.Accumulator) error {

	if t.Server == "" {
		t.Server = "127.0.0.1"
	}

	if t.Instance == "" {
		t.Instance = "default"
	}

	if t.Port == 0 {
		t.Port = 8080
	}

	client := &http.Client{}

	url := fmt.Sprintf("http://%s:%d/health", t.Server, t.Port)

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return err
	}

	resp, err := client.Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	check := HealthCheck{}

	json.NewDecoder(resp.Body).Decode(&check)

	t.GatherHealthCheck(acc, check)

	return nil
}

func init() {
	inputs.Add("traefik", func() telegraf.Input { return &Traefik{} })
}
