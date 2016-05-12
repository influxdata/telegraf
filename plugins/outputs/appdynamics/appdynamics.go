package influxdb

import (
	"encoding/json"
	"fmt"
	"github.com/influxdata/telegraf"
	"io/ioutil"
	"log"
	"net/http"
	"github.com/influxdata/telegraf/plugins/outputs"
)

var sampleConfig = `
  ## controller information tor connect and retrieve tier-id value
  controllerTierURL = "https://foo.saas.appdynamics.com/controller/rest/applications/bar/tiers/baz?output=JSON"
  controllerUserName = "apiuser@account.com"
  controllerPassword = "apipassword"
  ## Machine agent custom metrics listener url format string
  ## |Component:%d| gets transformed into |Component:id| during initialization - where 'id' is a tier-id for
  ## this controller application/tier combination
  agentURL = "http://localhost:8293/machineagent/metrics?name=Server|Component:%d|Custom+Metrics|"
`

type Appdynamics struct {
	// Controller values for retrieving tier-id from the controller
	ControllerTierURL     string
	ControllerUserName    string
	ControllerPassword    string

	// Machine agent URL format string
	AgentURL string

	// Tier id value retrieved from the controller for this application/tier
	tierId int64
}

// Close - There is nothing to close here, but need to comply with output interface
func (a *Appdynamics) Close() error{
	return nil
}

// Connect - initialize appdynamics plugin by retrieving tier-id value from the appdynamics controller
// for this application/tier combination and updating (reformatting) agent url string with tier-id value
func (a *Appdynamics) Connect() (err error) {
	a.tierId, err = a.getTierId()
	if err != nil {
		return err
	}
	fmt.Printf("Agent Tier ID: %d\n", a.tierId)
	a.AgentURL = fmt.Sprintf(a.AgentURL, a.tierId)
	fmt.Printf("Agent URL: %s\n", a.AgentURL)
	return err
}

// Description - describing what this is
func (a *Appdynamics) Description() string {
	return "Configuration for Appdynamics controller/listener to send metrics to"
}

// getTierId - retrieve tier id value for this application/tier combination from the appdynamics controller
func (a *Appdynamics) getTierId() (int64, error) {
	client := &http.Client{}

	/* Auth */
	req, err := http.NewRequest("GET", a.ControllerTierURL, nil)
	req.SetBasicAuth(a.ControllerUserName, a.ControllerPassword)

	res, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return 0, err
	}

	var tiers []struct {
		Id int64 `json:"id"`
	}

	err = json.Unmarshal(body, &tiers)
	if err != nil {
		return 0, err
	}
	if len(tiers) != 1 {
		fmt.Println("Invalid reply: ", tiers)
	}

	return tiers[0].Id, nil
}

// Write - post telegraf metrics to appdynamics machine agent listener using
// http.Get per https://docs.appdynamics.com/display/PRO40/Standalone+Machine+Agent+HTTP+Listener
func (a *Appdynamics) Write(metrics []telegraf.Metric) error {
	for _, metric := range metrics {
		if metric.Fields()["value"] == nil {
			log.Println("WARNING: missing value:", metric)
		} else {
			var appdType string
			switch metric.Tags()["metric_type"] {
			case "gauge":
				appdType = "average"
			default:
				appdType = "sum"
			}
			url := a.AgentURL + metric.Name() + fmt.Sprintf("&value=%v&type=%s", metric.Fields()["value"], appdType)
			fmt.Printf("Calling %s ...\n", url)
			_, err := http.Get(url)
			if err != nil {
				log.Println("ERROR: " + err.Error())
			}
		}
	}
	return nil
}

func (a *Appdynamics) SampleConfig() string {
	return sampleConfig
}

func init() {
	outputs.Add("appdynamics", func() telegraf.Output {
		return &Appdynamics{}
	})
}
