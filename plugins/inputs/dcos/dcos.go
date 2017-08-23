package dcos

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Dcos struct {
	ClusterURL        string            `toml:"cluster_url"`
	AuthToken         string            `toml:"auth_token"`
	Agents            []string          `toml:"agents"`
	FileSystemMounts  []string          `toml:"file_system_mounts"`
	NetworkInterfaces []string          `toml:"network_interfaces"`
	ClientTimeout     internal.Duration `toml:"client_timeout"`
}

var sampleConfig = `
  # Base URL of DC/OS cluster, e.g. http://dcos.example.com
  cluster_url=""
  # Authentication token, obtained by running: dcos config show core.dcos_acs_token
  auth_token=""
  # List of  DC/OS agent hostnames from which the metrics should be gathered. Leave empty for all.
  agents = []
  # DC/OS agent node file system mount for which related metrics should be gathered. Leave empty for all.
  file_system_mounts = []
  # DC/OS agent node network interface names for which related metrics should be gathered. Leave empty for all.
  network_interfaces = []
  # HTTP Response timeout, value must be more than a second
  #client_timeout = 4s
`

func (m *Dcos) Description() string {
	return "Input plugin for gathering DC/OS agent metrics"
}

func (m *Dcos) SampleConfig() string {
	return sampleConfig
}

var snakeCaser = strings.NewReplacer(".", "_")

type mesosMasterStateSummary struct {
	Hostname string  `json:"hostname"`
	Slaves   []slave `json:"slaves"`
}

type slave struct {
	Id       string `json:"id"`
	Hostname string `json:"hostname"`
}

type datapoint struct {
	Name      string                 `json:"name"`
	Value     interface{}            `json:"value"`
	Unit      string                 `json:"unit"`
	TimeStamp string                 `json:"timestamp"`
	Tags      map[string]interface{} `json:"tags"`
}

type metric struct {
	Datapoints []datapoint            `json:"datapoints"`
	Dimensions map[string]interface{} `json:"dimensions"`
}

const (
	Node      = "node"
	Container = "container"
	App       = "app"
)

const MesosMasterStateSummaryUrl = "/mesos/master/state-summary"
const AgentContainersUrl = "/system/v1/agent/%s/metrics/v0/containers"
const AgentContainerMetricsUrl = "/system/v1/agent/%s/metrics/v0/containers/%s"
const AgentContainerAppMetricsUrl = "/system/v1/agent/%s/metrics/v0/containers/%s/app"
const AgentNodeMetricsUrl = "/system/v1/agent/%s/metrics/v0/node"

var tr *http.Transport
var client *http.Client

func getClient(timeout time.Duration) *http.Client {

	if client == nil {
		tr = &http.Transport{
			ResponseHeaderTimeout: timeout - time.Second,
		}
		client = &http.Client{
			Transport: tr,
			Timeout:   timeout,
		}
	}
	return client
}

//validateConfiguration tests whether important configuration params are not empty
func (m *Dcos) validateConfiguration() error {
	errorStrings := []string{}

	if len(m.ClusterURL) == 0 {
		errorStrings = append(errorStrings, "Invalid configuration, cluster_url is empty")
	}
	if len(m.AuthToken) == 0 {
		errorStrings = append(errorStrings, "Invalid configuration, auth_token is empty")
	}

	if m.ClientTimeout.Duration.Seconds() == 0 {
		m.ClientTimeout.Duration = time.Second * 4
	} else if m.ClientTimeout.Duration.Seconds() <= 1 {
		errorStrings = append(errorStrings, "Invalid configuration, timeout value must be greater than a second")
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
		if isItemFiltered(m.Agents, a.Hostname) {
			continue
		}
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
		wg.Add(1)
		go func(cid string) {
			acc.AddError(m.gatherContainerAppMetrics(agent.Id, cid, acc))
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

//gatherContainerAppMetrics collects metric for an app in given container, if available
func (m *Dcos) gatherContainerAppMetrics(agentId string, containerId string, acc telegraf.Accumulator) error {
	url := m.ClusterURL + fmt.Sprintf(AgentContainerAppMetricsUrl, agentId, containerId)
	var metric metric

	err := m.handleJsonRequest(url, &metric)
	if err != nil {
		return err
	}

	m.processMetric(&metric, acc, App)

	return nil
}

//processMetric validated metric and fills accumulator with metric data
func (m *Dcos) processMetric(metric *metric, acc telegraf.Accumulator, metricType string) {

	measurementData := m.prepareMetric(metric, metricType, acc)
	//store current timestamp so all measurements for current set of metrics have the same timestamp
	now := time.Now()
	for measurementSuffix, tagPoints := range measurementData {
		for _, points := range tagPoints {
			tags := make(map[string]string)
			fields := make(map[string]interface{})

			fillTags(tags, metric.Dimensions)

			tags["scope"] = metricType
			tags["cluster_url"] = m.ClusterURL

			for _, dp := range points {
				fields[dp.Name] = dp.Value
				fillTags(tags, dp.Tags)
			}
			//fmt.Println(measurementSuffix, fields, tags)
			acc.AddFields("dcos_"+measurementSuffix, fields, tags, now)
		}
	}

}

//fillTags traverses source map and fills tags with non map values
func fillTags(tags map[string]string, source map[string]interface{}) {
	for k, v := range source {
		var s string
		if v != "" { //remove tags with empty value
			switch t := v.(type) {
			case map[string]interface{}:
				fillTags(tags, t)
			default:
				s = interfaceToString(v)
				tags[k] = s
			}

		}
	}
}

//prepareMetric sorts datapoints according to prefix and optional tag
func (m *Dcos) prepareMetric(metric *metric, metricType string, acc telegraf.Accumulator) map[string]map[string][]datapoint {
	//map measurement->(grouping)tag->datapoints
	measurementData := make(map[string]map[string][]datapoint)
	for _, d := range metric.Datapoints {
		if m.preProcessDataPoint(&d, metricType) {
			continue
		}
		var measurementSuffix string
		if metricType != App {
			nameSegs := strings.SplitN(d.Name, ".", 2)
			if len(nameSegs) == 2 {
				nameSegs[1] = snakeCaser.Replace(nameSegs[1])
			} else {
				//metric name could be already divided  by '_'
				nameSegs = strings.SplitN(d.Name, "_", 2)
				if len(nameSegs) != 2 {
					acc.AddError(fmt.Errorf("Unknown metric: '%s'", d.Name))
					continue
				}
			}
			measurementSuffix = nameSegs[0]
			d.Name = nameSegs[1]
		} else {
			measurementSuffix = "app"
		}
		mainTag := getMainTag(d)
		if _, ok := measurementData[measurementSuffix]; !ok {
			measurementData[measurementSuffix] = make(map[string][]datapoint)
		}
		measurementData[measurementSuffix][mainTag] = append(measurementData[measurementSuffix][mainTag], d)
	}
	return measurementData
}

//preProcessDataPoint checks fields and tags, modifies data if needed, and filters metric Return true if a datapoint should be added to measurement
func (m *Dcos) preProcessDataPoint(datapoint *datapoint, metricType string) bool {
	for k, v := range datapoint.Tags {
		switch k {
		case "interface": //filter network interfaces
			if isItemFiltered(m.NetworkInterfaces, interfaceToString(v)) {
				return true
			}
		case "path": //path
			if isItemFiltered(m.FileSystemMounts, interfaceToString(v)) {
				return true
			}
		}
	}
	//add units to name, in case if name already doesn't have that
	if len(datapoint.Unit) > 0 && !strings.HasSuffix(datapoint.Name, datapoint.Unit) {
		datapoint.Name = strings.Join([]string{datapoint.Name, datapoint.Unit}, ".")
	}
	return false
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

	resp, err := getClient(m.ClientTimeout.Duration).Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		switch resp.StatusCode {
		case 401:
			return fmt.Errorf("Authentication error. Verify that the auth_token parameter is correct")
		default:
			return fmt.Errorf("HTTP request to %s has failed. HTTP status code: %d", url, resp.StatusCode)
		}
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if len(body) > 0 {
		err = json.Unmarshal(body, &obj)
		if err != nil {
			return fmt.Errorf("Error parsing data from %s:  %s", url, err.Error())
		}
	}
	return nil
}

//interfaceToString returns string representation of given interface
func interfaceToString(v interface{}) string {
	return fmt.Sprintf("%v", v)
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

//getMainTag returns value of a tag determining membership of datapoint to a group
func getMainTag(datapoint datapoint) string {
	var mainTag interface{}
	if v, ok := datapoint.Tags["path"]; ok {
		mainTag = v
	} else if v, ok := datapoint.Tags["interface"]; ok {
		mainTag = v
	} else if v, ok := datapoint.Tags["container_id"]; ok {
		mainTag = v
	}
	return interfaceToString(mainTag)
}

func init() {
	inputs.Add("dcos", func() telegraf.Input { return &Dcos{} })
}
