package dcos

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Dcos struct {
	ClusterURL        string   `toml:"cluster_url"`
	AuthToken         string   `toml:"auth_token"`
	FileSystemMounts  []string `toml:"file_system_mounts"`
	NetworkInterfaces []string `toml:"network_interfaces"`
}

var sampleConfig = `
	# Base URL of DC/OS cluster, e.g. http://dcos.example.com
	cluster_url=
	# Authentication token, obtained by running: dcos config show core.dcos_acs_token
	auth_token=
	# DC/OS agent node file system mount for which related metrics should be gathered
	file_system_mounts = []
	# DC/OS agent node network interface names for which related metrics should be gathered
	network_interfaces = []
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
	Value     float64           `json:"value"`
	Unit      string            `json:"unit"`
	TimeStamp string            `json:"timestamp"`
	Tags      map[string]string `json:"tags"`
}

type metric struct {
	Datapoints []datapoint       `json:"datapoints"`
	Dimensions map[string]string `json:"dimensions"`
}

const (
	Node      = "node"
	Container = "container"
)

const MesosMasterStateSummaryUrl = "/mesos/master/state-summary"
const AgentContainersUrl = "/system/v1/agent/%s/metrics/v0/containers"
const AgentContainerMetricsUrl = "/system/v1/agent/%s/metrics/v0/containers/%s"
const AgentNodeMetricsUrl = "/system/v1/agent/%s/metrics/v0/node"

//validateConfiguration tests whether important configuration params are not empty
func (m *Dcos) validateConfiguration() error {
	errorStrings := []string{}

	if len(m.ClusterURL) == 0 {
		errorStrings = append(errorStrings, "Invalid configuration, cluster_url is empty")
	}
	if len(m.AuthToken) == 0 {
		errorStrings = append(errorStrings, "Invalid configuration, auth_token is empty")
	}

	if len(errorStrings) > 0 {
		return errors.New(strings.Join(errorStrings, "\n"))
	} else {
		return nil
	}

}

func (m *Dcos) Gather(acc telegraf.Accumulator) error {
	err := m.validateConfiguration()
	if err != nil {
		return err
	}
	//get agents
	url := m.ClusterURL + MesosMasterStateSummaryUrl
	var state mesosMasterStateSummary
	err = m.handleJsonRequest(url, &state)
	if err != nil {
		return err
	}

	if len(state.Slaves) == 0 {
		return fmt.Errorf("DC/OS cluster %s does not have any running slave", m.ClusterURL)
	}

	var wg sync.WaitGroup

	for _, a := range state.Slaves {
		wg.Add(1)
		go func(agent slave) {
			acc.AddError(m.gatherAgentMetrics(agent, acc))
			wg.Done()
		}(a)
		wg.Add(1)
		go func(agent slave) {
			acc.AddError(m.gatherAgentContainers(agent, acc))
			wg.Done()
		}(a)
	}

	wg.Wait()
	return nil
}

//gatherAgentMetrics collects metric for agent
func (m *Dcos) gatherAgentMetrics(agent slave, acc telegraf.Accumulator) error {
	url := m.ClusterURL + fmt.Sprintf(AgentNodeMetricsUrl, agent.Id)

	var nodeMetric metric
	err := m.handleJsonRequest(url, &nodeMetric)

	if err != nil {
		return err
	}

	m.processMetric(&nodeMetric, acc, Node)
	return nil
}

//gatherAgentContainers manages gathering metrics for containers on agent.
func (m *Dcos) gatherAgentContainers(agent slave, acc telegraf.Accumulator) error {
	url := m.ClusterURL + fmt.Sprintf(AgentContainersUrl, agent.Id)
	var containerIds []string
	err := m.handleJsonRequest(url, &containerIds)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup

	for _, c := range containerIds {
		wg.Add(1)
		go func(cid string) {
			acc.AddError(m.gatherContainerMetrics(agent.Id, cid, acc))
			wg.Done()
		}(c)
	}
	wg.Wait()

	return nil
}

//gatherContainerMetrics collects metric for given container
func (m *Dcos) gatherContainerMetrics(agentId string, containerId string, acc telegraf.Accumulator) error {
	url := m.ClusterURL + fmt.Sprintf(AgentContainerMetricsUrl, agentId, containerId)
	var metric metric
	err := m.handleJsonRequest(url, &metric)
	if err != nil {
		return err
	}
	m.processMetric(&metric, acc, Container)
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
		switch resp.StatusCode {
		case 401:
			return fmt.Errorf("Authentication error. Verify the auth_token is correct")
		default:
			return fmt.Errorf("HTTP request to %s has failed. HTTP status code: %d", url, resp.StatusCode)
		}

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

//processMetric validated metric and fills accumulator with metric data
func (m *Dcos) processMetric(metric *metric, acc telegraf.Accumulator, metricType string) {
	tags := make(map[string]string)
	fields := make(map[string]interface{})

	for k, v := range metric.Dimensions {
		tags[k] = v
	}
	tags["metric_scope"] = metricType
	tags["cluster_url"] = m.ClusterURL

	//fmt.Println("Datapoints")

	for _, d := range metric.Datapoints {
		if m.preProcessDataPoint(&d, metricType) {
			continue
		}
		fields[d.Name] = d.Value
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
	//fmt.Println("Fields")
	//fmt.Printf("  %v\n",fields)
	//fmt.Println("Tags")
	//fmt.Printf("  %v\n", tags)
	acc.AddFields("dcos", fields, tags)
}

//preProcessDataPoint checks and filters metric and modifies name in case of group metrics for file system and network. Return true if datapoint  should be added to measurement
func (m *Dcos) preProcessDataPoint(datapoint *datapoint, metricType string) bool {
	if metricType == Node {
		for k, v := range datapoint.Tags {
			switch k {
			case "interface":
				if isItemFiltered(m.NetworkInterfaces, v) {
					return true
				}
				datapoint.Name = fmt.Sprintf("%s.%s", datapoint.Name, nameSanitizer.Replace(v))
				delete(datapoint.Tags, k)
			case "path":
				if isItemFiltered(m.FileSystemMounts, v) {
					return true
				}
				datapoint.Name = datapoint.Name + nameSanitizer.Replace(v)
				delete(datapoint.Tags, k)
			}
		}
	}
	return false
}

//isItemFiltered tests whether item is part of non empty array, if not, returns true
func isItemFiltered(array []string, item string) bool {
	if len(array) > 0 {
		for _, i := range array {
			if i == item {
				return false
			}
		}
		return true
	}
	return false
}

func init() {
	inputs.Add("dcos", func() telegraf.Input { return &Dcos{} })
}
