package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
)

type Icinga2 struct {
	Server   string
	Username string
	Password string
}

type Result struct {
	Results []Object
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
	# server = "localhost"
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
	icinga2 := Icinga2{Server: "https://url:5665/v1/objects/services?attrs=name&attrs=display_name&attrs=state&attrs=check_command", Username: "root", Password: "icinga"}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	req, err := http.NewRequest("GET", icinga2.Server, nil)
	req.SetBasicAuth(icinga2.Username, icinga2.Password)
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	result := Result{}
	json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		panic(err)
	}
	for _, item := range result.Results {
		res := fmt.Sprintf("services_status,name=\"%s\",display_name=\"%s\",check_command=\"%s\" state=%f", item.Attrs.Name, item.Attrs.DisplayName, item.Attrs.CheckCommand, item.Attrs.State)
		fmt.Println(res)
	}

	for _, check := range checks {
		record := make(map[string]interface{})
		tags := make(map[string]string)

		record["name"] = item.Attrs.Name
		record["status"] = item.Attrs.State

		tags["display_name"] = item.Attrs.DisplayName
		tags["check_command"] = item.Attrs.CheckCommand

		acc.AddFields("icinga2_services_status", record, tags)
	}

	return nil
}

func init() {
	inputs.Add("icinga2", func() telegraf.Input { return &Icinga2{} })
}
