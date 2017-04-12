package icinga2

import (
	"crypto/tls"
	"encoding/json"
	"net/http"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Icinga2 struct {
	Server   string
	Username string
	Password string
}

type Result struct {
	Results []Object `json:"results"`
}

type Object struct {
	Attrs Attribute  `json:"attrs"`
	Name  string     `json:"name"`
	Joins struct{}   `json:"joins"`
	Meta  struct{}   `json:"meta"`
	Type  ObjectType `json:"type"`
}

type Attribute struct {
	CheckCommand string  `json:"check_command"`
	DisplayName  string  `json:"display_name"`
	Name         string  `json:"name"`
	State        float32 `json:"state"`
}

const (
	SERVICE = "Service"
	HOST    = "Host"
)

type ObjectType string

var sampleConfig = `
	## Required Icinga2 server address (default: "https://localhost:5665")
	# server = "https://localhost:5665"
	## Required username used for request HTTP Basic Authentication (default: "")
	# username = ""
	## Required password used for HTTP Basic Authentication (default: "")
	# password = ""
	`

func (s *Icinga2) Description() string {
	return "Read status from Icinga2"
}

func (s *Icinga2) SampleConfig() string {
	return sampleConfig
}

func (s *Icinga2) Gather(acc telegraf.Accumulator) error {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}

	req, err := http.NewRequest("GET", s.Server+"/v1/objects/services?attrs=name&attrs=display_name&attrs=state&attrs=check_command", nil)
	req.SetBasicAuth(s.Username, s.Password)

	if err != nil {
		return err
	}

	resp, err := client.Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()
	result := Result{}
	json.NewDecoder(resp.Body).Decode(&result)

	if err != nil {
		return err
	}

	for _, check := range result.Results {
		record := make(map[string]interface{})
		tags := make(map[string]string)

		record["name"] = check.Attrs.Name
		record["status"] = check.Attrs.State

		tags["display_name"] = check.Attrs.DisplayName
		tags["check_command"] = check.Attrs.CheckCommand

		acc.AddFields("icinga2_services_status", record, tags)
	}

	return nil
}

func init() {
	inputs.Add("icinga2", func() telegraf.Input { return &Icinga2{} })
}
