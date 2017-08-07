package dcos

import (
	"encoding/json"
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type Dcos struct {
	ClusterURL string `toml: cluster_url`
	AuthToken  string `toml: auth_token`
}

var sampleConfig = `
	# Base URL of DCOS cluster, e.g. http://dcos.example.com
	cluster_url=
	# Authentication token, obtained by running: dcos config show core.dcos_acs_token
	auth_token=
`

func (m *Dcos) Description() string {
	return "Input plugin for gathering DCOS agent metrics"
}

func (m *Dcos) SampleConfig() string {
	return sampleConfig
}

var nameSanitizer = strings.NewReplacer("/", "_", `\`, "_")

var tr = &http.Transport{
	ResponseHeaderTimeout: time.Duration(3 * time.Second),
}

var client = &http.Client{
	Transport: tr,
	Timeout:   time.Duration(4 * time.Second),
}

type mesosMasterStateSummary struct {
	Hostname string  `json:"hostname"`
	Slaves   []slave `json:"slaves"`
}

type slave struct {
	Id       string `json:"id"`
	Hostname string `json:"hostname"`
}

type datapoint struct {
	Name      string            `json:"name"`
	Value     interface{}       `json:"value"`
	Unit      string            `json:"unit"`
	TimeStamp string            `json:"timestamp"`
	Tags      map[string]string `json:"tags"`
}

type metric struct {
	Datapoints []datapoint       `json:"datapoints"`
	Dimensions map[string]string `json:"dimensions"`
}

const MesosMasterStateSummaryUrl = "/mesos/master/state-summary"
const AgentContainersUrl = "/system/v1/agent/%s/metrics/v0/containers"
const AgentContainerMetricsUrl = "/system/v1/agent/%s/metrics/v0/containers/%s"
const AgentNodeMetricsUrl = "/system/v1/agent/%s/metrics/v0/node"

func (m *Dcos) Gather(acc telegraf.Accumulator) error {
	//get agents
	url := m.ClusterURL + MesosMasterStateSummaryUrl
	var state mesosMasterStateSummary
	err := m.handleJsonRequest(url, &state)
	if err != nil {
		return err
	}

	if len(state.Slaves) == 0 {
		return fmt.Errorf("DCOS cluster %s does not have any running slave", m.ClusterURL)
	}

	for _, a := range state.Slaves {
		//fmt.Println("Agent: " + a.Id)

		url = m.ClusterURL + fmt.Sprintf(AgentNodeMetricsUrl, a.Id)
		var nodeMetric metric
		err = m.handleJsonRequest(url, &nodeMetric)
		if err != nil {
			return err
		}

		processMetric(&nodeMetric, acc, "agent")

		url = m.ClusterURL + fmt.Sprintf(AgentContainersUrl, a.Id)
		var containerIds []string
		err = m.handleJsonRequest(url, &containerIds)
		if err != nil {
			return err
		}

		for _, c := range containerIds {
			//fmt.Println("Container: " + c)
			url = m.ClusterURL + fmt.Sprintf(AgentContainerMetricsUrl, a.Id, c)
			var metric metric
			err = m.handleJsonRequest(url, &metric)
			if err != nil {
				return err
			}
			processMetric(&metric, acc, "container")
		}
	}

	return nil
}

//newRequest creates http request object to given url with common headers required by DCOS
func (m *Dcos) newRequest(url string) (*http.Request, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "token="+m.AuthToken)
	req.Header.Add("Accept", "application/json")
	return req, nil
}

//handleJsonRequest takes care of sending request to given url, returning JSON data and un-marshalling JSON payload into given object
func (m *Dcos) handleJsonRequest(url string, obj interface{}) error {
	req, err := m.newRequest(url)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

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

func processMetric(metric *metric, acc telegraf.Accumulator, metricType string) {
	tags := make(map[string]string)
	fields := make(map[string]interface{})

	for k, v := range metric.Dimensions {
		tags[k] = v
	}
	tags["Type"] = metricType

	//fmt.Println("Datapoits")

	var t time.Time
	var err error
	var timeCollected = false

	for _, d := range metric.Datapoints {
		//all datapoints have the same timestamp, take the first
		if !timeCollected {
			t, err = time.Parse(time.RFC3339, d.TimeStamp)
			if err == nil {
				timeCollected = true
			}
		}
		preProcessDataPoint(&d)
		fields[d.Name] = d.Unit
		//fmt.Printf(" - [ %s: %v%s\n", d.Name, d.Value, d.Unit)
		//fmt.Printf("      %v]\n", d.Tags)
		for k, v := range d.Tags {
			tags[k] = v
		}
	}
	//fmt.Println("Dimensions")
	//fmt.Printf("  %v\n", metric.Dimensions)
	//fmt.Println("AllTags")
	//fmt.Printf("  %v\n", tags)

	if !timeCollected {
		t = time.Now()
	}
	acc.AddFields("dcos", fields, tags, t)
}

func preProcessDataPoint(datapoint *datapoint) {
	for k, v := range datapoint.Tags {
		switch k {
		case "interface":
			datapoint.Name = fmt.Sprintf("%s.%s", datapoint.Name, nameSanitizer.Replace(v))
			delete(datapoint.Tags, k)
		case "path":
			datapoint.Name = datapoint.Name + nameSanitizer.Replace(v)
			delete(datapoint.Tags, k)
		}
	}
}

func init() {
	inputs.Add("dcos", func() telegraf.Input { return &Dcos{} })
}
