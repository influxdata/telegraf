package dcos

import (
	"github.com/influxdata/telegraf"
	"net/http"
	"time"
	"io/ioutil"
	"encoding/json"
	"fmt"
)

type Dcos struct {
	ClusterURL string `toml: cluster_url`
	AuthToken	string `toml: auth_token`
}
var sampleConfig = `
	# Base URL of DCOS cluster, e.g. http://dcos.example.com
	cluster_url=
	# Authentication token, obtained by running: dcos config show core.dcos_acs_token
	auth_token=
`

func (m* Dcos) Description() string {
	return "Input plugin for gathering DCOS agent metrics"
}

func (m* Dcos) SampleConfig() string {
	return sampleConfig
}

var tr = &http.Transport{
	ResponseHeaderTimeout: time.Duration(3 * time.Second),
}

var client = &http.Client{
	Transport: tr,
	Timeout:   time.Duration(4 * time.Second),
}

type mesosMasterStateSummary struct {
	Hostname string `json:"hostname"`
	Slaves   []slave   `json:"slaves"`
}

type slave struct {
	Id       string `json:"id"`
	Hostname string `json:"hostname"`
}

type datapoint struct {
	Name string `json:"name"`
	Value interface{} `json:"value"`
	Unit string `json:"unit"`
	TimeStamp string `json:"timestamp"`
	Tags map[string]string `json:"tags"`
}

type metric struct {
	Datapoints   []datapoint `json:"datapoints"`
	Dimensions map[string]string `json:"dimensions"`
}

const MesosMasterStateSummaryUrl = "/mesos/master/state-summary"
const AgentContainersUrl = "/system/v1/agent/%s/metrics/v0/containers"
const AgentContainerMetricsUrl = "/system/v1/agent/%s/metrics/v0/containers/%s"
const AgentNodeMetricsUrl = "/system/v1/agent/%s/metrics/v0/node"

func (m* Dcos) Gather(acc telegraf.Accumulator) error {
	//get agents
	url := m.ClusterURL + MesosMasterStateSummaryUrl
	var state mesosMasterStateSummary
	err :=  m.handleJsonRequest(url, &state)
	if err != nil {
		return err
	}

	if len(state.Slaves) == 0 {
		return fmt.Errorf("DCOS cluster %s does not have any running slave", m.ClusterURL)
	}

	for _, a := range state.Slaves {
		fmt.Println("Agent: " + a.Id)

		url = m.ClusterURL + fmt.Sprintf(AgentNodeMetricsUrl, a.Id)
		var nodeMetric metric
		err = m.handleJsonRequest(url, &nodeMetric)
		if err != nil {
			return err
		}

		dumpMetric(&nodeMetric, acc)

		url = m.ClusterURL + fmt.Sprintf(AgentContainersUrl, a.Id)
		var containerIds   []string
		err = m.handleJsonRequest(url, &containerIds)
		if err != nil {
			return err
		}

		for _, c := range containerIds {
			fmt.Println("Container: " + c)
			url = m.ClusterURL + fmt.Sprintf(AgentContainerMetricsUrl, a.Id, c)
			var metric metric
			err = m.handleJsonRequest(url, &metric)
			if err != nil {
				return err
			}
			dumpMetric(&metric, acc)
		}
	}

	return nil
}

//newRequest creates http request object to given url with common headers required by DCOS
func (m* Dcos ) newRequest(url string) (*http.Request, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "token=" + m.AuthToken)
	req.Header.Add("Accept", "application/json")
	return req, nil
}

//handleJsonRequest takes care of sending request to given url, returning JSON data and unmarshalling JSON payload into given object
func (m* Dcos ) handleJsonRequest(url string, obj interface{}) ( error) {
	req, err := m.newRequest(url)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)

	defer resp.Body.Close()

	if err != nil {
		return err
	}

	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("HTTP request has failed. HTTP status code: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)


	if err != nil {
		return err
	}

	err = json.Unmarshal(body, &obj)
	if err != nil {
		return err
	}
	return nil
}

func dumpMetric(metric *metric, acc telegraf.Accumulator ) {
	fmt.Println("Datapoits")
	for _, d := range metric.Datapoints {
		fmt.Printf(" - [ %s: %v%s\n", d.Name, d.Value, d.Unit)
		fmt.Printf("      %v]\n", d.Tags)
	}
	fmt.Println("Dimensions")
	fmt.Printf("  %v\n", metric.Dimensions)
}