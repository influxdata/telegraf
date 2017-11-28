package dcos

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Dcos struct {
	// Hostname or ip of DC/OS(mesos) master accessible from within the cluster
	MasterHostname string `toml:"master_hostname"`
	// Port of the master
	MasterPort uint16 `toml:"master_port"`
	// URL for DC/OS cluster for access from outside the cluster
	ClusterURL string `toml:"cluster_url"`
	// Authentication token for login when accessing from outside the cluster
	AuthToken string `toml:"auth_token"`
	// White list of agents to access
	Agents []string `toml:"agent_include"`
	// White list of filessytem paths on node metrics
	FileSystemMounts []string `toml:"path_include"`
	// White list of network interface node metrics
	NetworkInterfaces []string `toml:"interface_include"`
	// Timeout for a client to wait for a server response
	ResponseTimeout internal.Duration `toml:"response_timeout"`
	MetricsPort     uint16            `toml:"metrics_port"`
	// Path to CA file
	SSLCA string `toml:"ssl_ca"`
	// Path to client cert file
	SSLCert string `toml:"ssl_cert"`
	// Path to cert key file
	SSLKey string `toml:"ssl_key"`
	// Use SSL but skip chain & host verification
	InsecureSkipVerify bool `toml:"insecure_skip_verify"`
	// Flag if plugin is accessing DC/OS from inside or outside
	localAccess bool
	// HTTP client for reuse
	client *http.Client
	// Flag if plugin was already initialized
	initialized bool
	// map of HTTP headers for reuse
	headersCache map[string][]string
	// semaphore for limiting number of concurrent HTTP connections
	semaphore chan int
}

var sampleConfig = `
  # Hostname or ip address of DC/OS master for access from within DC/OS cluster
  master_hostname=""
  # Public URL of DC/OS cluster, e.g. http://dcos.example.com. Use of access from outside of the DC/OS cluster. master_hostname has higher priority, if set
  #cluster_url=""
  # Authentication token, obtained by running: dcos config show core.dcos_acs_token. Leave empty for no authentication.
  # Warning: authentication token is valid only 5 days in DC/OS 1.10.
  #auth_token=""
  # List of  DC/OS agent hostnames from which the metrics should be gathered. Leave empty for all.
  agent_include = []
  # DC/OS agent node file system mount for which related metrics should be gathered. Leave empty for all.
  path_include = []
  # DC/OS agent node network interface names for which related metrics should be gathered. Leave empty for all.
  interface_include = []
  # HTTP Response timeout, value must be more than a second
  #response_timeout = "30s"
  # Port number of Mesos component on DC/OS master for access from within DC/OS cluster
  #master_port = 5050
  # Port number of DC/OS metrics component on DC/OS agents. Must be the same on all agents
  #metrics_port = 61001
  # TLS/SSL configuration for cluster_url
  #ssl_ca = "/etc/telegraf/ca.pem"
  #ssl_cert = "/etc/telegraf/cert.cer"
  #ssl_key = "/etc/telegraf/key.key"
  #insecure_skip_verify = false
`

var tagInclude = []string{"cluster_id", "cluster_url", "path", "interface", "hostname", "mesos_id", "container_id", "executor_id", "framework_id", "scope", "package_name"}

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

type mesosAgentContainer struct {
	Id string `json:"container_id"`
}

type slave struct {
	Id        string  `json:"id"`
	Hostname  string  `json:"hostname"`
	MesosPort float64 `json:"port"`

	metricsBaseURL string
	mesosBaseURL   string
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

const (
	MesosMasterStateSummaryRemotePathTemplate  = "/mesos/master/state-summary"
	MesosSlaveContainersRemotePathTemplate     = "/agent/%s/containers"
	AgentContainersRemotePathTemplate          = "/system/v1/agent/%s/metrics/v0/containers"
	AgentContainerMetricsRemotePathTemplate    = "/system/v1/agent/%s/metrics/v0/containers/%s"
	AgentContainerAppMetricsRemotePathTemplate = "/system/v1/agent/%s/metrics/v0/containers/%s/app"
	AgentNodeMetricsRemotePathTemplate         = "/system/v1/agent/%s/metrics/v0/node"

	MesosMasterStateSummaryLocalPathTemplate  = "/state-summary"
	MesosSlaveContainersLocalPathTemplate     = "/containers"
	AgentContainersLocalPathTemplate          = "/system/v1/metrics/v0/containers"
	AgentContainerMetricsLocalPathTemplate    = "/system/v1/metrics/v0/containers/%s"
	AgentContainerAppMetricsLocalPathTemplate = "/system/v1/metrics/v0/containers/%s/app"
	AgentNodeMetricsLocalPathTemplate         = "/system/v1/metrics/v0/node"
)

const MaxIdleConnections = 50

// validateConfiguration tests whether important configuration params are not empty
func (m *Dcos) validateConfiguration() error {
	if len(m.MasterHostname) == 0 && len(m.ClusterURL) == 0 {
		return errors.New("Invalid configuration, either master_hostname or cluster_url must be set")
	}

	if 0 < m.ResponseTimeout.Duration.Seconds() && m.ResponseTimeout.Duration.Seconds() <= 1 {
		return errors.New("Invalid configuration, timeout value must be greater than a second")
	}

	return nil
}

// init validates configuration and initializes variables
func (m *Dcos) init() error {
	if err := m.validateConfiguration(); err != nil {
		return err
	}

	if len(m.MasterHostname) > 0 {
		if m.MasterPort == 0 {
			m.MasterPort = 5050
		}
		if m.MetricsPort == 0 {
			m.MetricsPort = 61001
		}
		m.localAccess = true
		murl := &url.URL{
			Scheme: "http",
			Host:   net.JoinHostPort(m.MasterHostname, strconv.Itoa(int(m.MasterPort))),
		}
		m.ClusterURL = murl.String()
	}

	if m.ResponseTimeout.Duration.Seconds() == 0 {
		m.ResponseTimeout.Duration = time.Second * 30
	}

	m.headersCache = map[string][]string{
		"Accept": {"application/json"},
	}

	if len(m.AuthToken) > 0 {
		m.headersCache["Authorization"] = []string{"token=" + m.AuthToken}
	}

	tlsCfg, err := internal.GetTLSConfig(
		m.SSLCert, m.SSLKey, m.SSLCA, m.InsecureSkipVerify)

	if err != nil {
		return err
	}

	tr := &http.Transport{
		TLSClientConfig:       tlsCfg,
		MaxIdleConns:          MaxIdleConnections,
		MaxIdleConnsPerHost:   MaxIdleConnections,
		ResponseHeaderTimeout: m.ResponseTimeout.Duration - time.Second,
	}
	m.client = &http.Client{
		Transport: tr,
		Timeout:   m.ResponseTimeout.Duration,
	}
	return nil
}

func (m *Dcos) mesosMasterStateSummaryURL() string {
	if m.localAccess {
		return m.ClusterURL + MesosMasterStateSummaryLocalPathTemplate
	} else {
		return m.ClusterURL + MesosMasterStateSummaryRemotePathTemplate
	}
}

func (m *Dcos) mesosSlaveContainersURL(agent slave) string {
	if m.localAccess {
		return agent.mesosBaseURL + MesosSlaveContainersLocalPathTemplate
	} else {
		return m.ClusterURL + fmt.Sprintf(MesosSlaveContainersRemotePathTemplate, agent.Id)
	}
}

func (m *Dcos) agentNodeMetricsURL(agent slave) string {
	if m.localAccess {
		return agent.metricsBaseURL + AgentNodeMetricsLocalPathTemplate
	} else {
		return m.ClusterURL + fmt.Sprintf(AgentNodeMetricsRemotePathTemplate, agent.Id)
	}
}

func (m *Dcos) agentContainersURL(agent slave) string {
	if m.localAccess {
		return agent.metricsBaseURL + AgentContainersLocalPathTemplate
	} else {
		return m.ClusterURL + fmt.Sprintf(AgentContainersRemotePathTemplate, agent.Id)
	}
}

func (m *Dcos) agentContainerMetricsURL(agent slave, cid string) string {
	if m.localAccess {
		return agent.metricsBaseURL + fmt.Sprintf(AgentContainerMetricsLocalPathTemplate, cid)
	} else {
		return m.ClusterURL + fmt.Sprintf(AgentContainerMetricsRemotePathTemplate, agent.Id, cid)
	}
}

func (m *Dcos) agentContainerAppMetricsURL(agent slave, cid string) string {
	if m.localAccess {
		return agent.metricsBaseURL + fmt.Sprintf(AgentContainerAppMetricsLocalPathTemplate, cid)
	} else {
		return m.ClusterURL + fmt.Sprintf(AgentContainerAppMetricsRemotePathTemplate, agent.Id, cid)
	}
}

func (m *Dcos) Gather(acc telegraf.Accumulator) error {
	if !m.initialized {
		err := m.init()
		m.initialized = true
		if err != nil {
			return err
		}
	}
	m.semaphore = make(chan int, MaxIdleConnections)
	//get agents
	var state mesosMasterStateSummary
	err := m.handleJsonRequest(m.mesosMasterStateSummaryURL(), &state)
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
		slaveURL := &url.URL{
			Scheme: "http",
			Host:   net.JoinHostPort(a.Hostname, strconv.Itoa(int(m.MetricsPort))),
		}
		a.metricsBaseURL = slaveURL.String()

		slaveURL = &url.URL{
			Scheme: "http",
			Host:   net.JoinHostPort(a.Hostname, strconv.Itoa(int(a.MesosPort))),
		}
		a.mesosBaseURL = slaveURL.String()

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

// gatherAgentMetrics collects metric for agent
func (m *Dcos) gatherAgentMetrics(agent slave, acc telegraf.Accumulator) error {
	var nodeMetric metric
	err := m.handleJsonRequest(m.agentNodeMetricsURL(agent), &nodeMetric)

	if err != nil {
		return err
	}

	m.processMetric(&nodeMetric, acc, Node)
	return nil
}

// gatherAgentContainers manages gathering metrics for containers on agent.
func (m *Dcos) gatherAgentContainers(agent slave, acc telegraf.Accumulator) error {
	var containerIds []string
	err := m.handleJsonRequest(m.agentContainersURL(agent), &containerIds)
	if err != nil {
		return err
	}

	agentContainers, err := m.getMesosAgentContainers(agent)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup

	for _, c := range containerIds {
		//metrics are returned also for removed/not running containers,
		// so filter such containers out
		if !isItemFiltered(agentContainers, c) {
			wg.Add(1)
			go func(cid string) {
				acc.AddError(m.gatherContainerMetrics(agent, cid, acc))
				wg.Done()
			}(c)
			wg.Add(1)
			go func(cid string) {
				acc.AddError(m.gatherContainerAppMetrics(agent, cid, acc))
				wg.Done()
			}(c)
		}
	}
	wg.Wait()

	return nil
}

// getMesosAgentContainers gathers ids of running mesos container
func (m *Dcos) getMesosAgentContainers(agent slave) ([]string, error) {
	var agentContainers []mesosAgentContainer
	err := m.handleJsonRequest(m.mesosSlaveContainersURL(agent), &agentContainers)
	if err != nil {
		return nil, err
	}
	agentContainerIds := make([]string, len(agentContainers))
	for _, a := range agentContainers {
		agentContainerIds = append(agentContainerIds, a.Id)
	}
	return agentContainerIds, nil
}

// gatherContainerMetrics collects metric for given container
func (m *Dcos) gatherContainerMetrics(agent slave, containerId string, acc telegraf.Accumulator) error {
	var metric metric
	err := m.handleJsonRequest(m.agentContainerMetricsURL(agent, containerId), &metric)
	if err != nil {
		return err
	}
	m.processMetric(&metric, acc, Container)
	return nil
}

// gatherContainerAppMetrics collects metric for an app in given container,
// if available
func (m *Dcos) gatherContainerAppMetrics(agent slave, containerId string, acc telegraf.Accumulator) error {
	var metric metric
	err := m.handleJsonRequest(m.agentContainerAppMetricsURL(agent, containerId), &metric)
	if err != nil {
		return err
	}

	m.processMetric(&metric, acc, App)

	return nil
}

//processMetric validated metric and fills accumulator with metric data
func (m *Dcos) processMetric(metric *metric, acc telegraf.Accumulator, metricType string) {

	measurementData := m.organizeMetric(metric, metricType, acc)
	//store current timestamp so all measurements for current set of metrics
	// have the same timestamp
	now := time.Now()
	for measurementSuffix, tagPoints := range measurementData {
		for _, points := range tagPoints {
			tags := make(map[string]string)
			fields := make(map[string]interface{})

			fillTags(tags, metric.Dimensions)

			tags["scope"] = metricType
			tags["cluster_url"] = m.ClusterURL

			for _, dp := range points {
				switch t := dp.Value.(type) {
				case bool:
					var f float64
					if t {
						f = 1
					}
					fields[dp.Name] = f
					fillTags(tags, dp.Tags)
				case float64:
					fields[dp.Name] = t
					fillTags(tags, dp.Tags)
				}
			}
			acc.AddFields("dcos_"+measurementSuffix, fields, tags, now)
		}
	}

}

// fillTags traverses source map and fills tags with non map values
func fillTags(tags map[string]string, source map[string]interface{}) {
	for k, v := range source {
		var s string
		if v != "" { //remove tags with empty value
			switch t := v.(type) {
			case map[string]interface{}:
				fillTags(tags, t)
			default:

				k = strings.TrimPrefix(strings.ToLower(k), "dcos_")
				//use hardcoded filter
				if !isItemFiltered(tagInclude, k) {
					s = tagToString(v)
					tags[k] = s
				}
			}

		}
	}
}

// organizeMetric normalizes datapoints name and sorts datapoints according to prefix and tag
func (m *Dcos) organizeMetric(metric *metric, metricType string, acc telegraf.Accumulator) map[string]map[string][]datapoint {
	//map measurement->(grouping)tag->datapoints
	measurementData := make(map[string]map[string][]datapoint)
	for _, d := range metric.Datapoints {
		if m.isDatapointFiltered(&d, metricType) {
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
					//unknown metrics
					nameSegs = []string{"general", d.Name}
				}
			}
			measurementSuffix = nameSegs[0]
			d.Name = nameSegs[1]
		} else {
			measurementSuffix = "app"
		}
		//add units to name, in case if name already doesn't have that
		if len(d.Unit) > 0 && !strings.HasSuffix(d.Name, d.Unit) {
			d.Name = strings.Join([]string{d.Name, d.Unit}, "_")
		}
		mainTag := getMainTag(d)
		if _, ok := measurementData[measurementSuffix]; !ok {
			measurementData[measurementSuffix] = make(map[string][]datapoint)
		}
		measurementData[measurementSuffix][mainTag] = append(measurementData[measurementSuffix][mainTag], d)
	}
	return measurementData
}

// isDatapointFiltered checks fields and tags, modifies data if needed,
// and filters metric Return true if a datapoint should be added to measurement
func (m *Dcos) isDatapointFiltered(datapoint *datapoint, metricType string) bool {
	for k, v := range datapoint.Tags {
		switch k {
		case "interface": //filter network interfaces
			if isItemFiltered(m.NetworkInterfaces, tagToString(v)) {
				return true
			}
		case "path": //path
			if isItemFiltered(m.FileSystemMounts, tagToString(v)) {
				return true
			}
		}
	}
	return false
}

// newRequest creates http request object to given url with common headers
// required by DCOS
func (m *Dcos) newRequest(url string) (*http.Request, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header = m.headersCache

	return req, nil
}

// handleJsonRequest takes care of sending request to given url, returning
// JSON data and un-marshalling JSON payload into given object
func (m *Dcos) handleJsonRequest(url string, obj interface{}) error {
	req, err := m.newRequest(url)
	if err != nil {
		return err
	}

	m.semaphore <- 1
	defer func() { <-m.semaphore }()

	resp, err := m.client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		switch resp.StatusCode {
		case 401:
			return fmt.Errorf("Authentication error. Verify that the auth_token parameter is correct")
		default:
			return fmt.Errorf("HTTP request to %s has failed. HTTP status code: %d", url, resp.StatusCode)
		}
	}

	if resp.StatusCode != 204 {
		jsonDecoder := json.NewDecoder(resp.Body)
		err = jsonDecoder.Decode(&obj)
		if err != nil {
			return fmt.Errorf("Error parsing data from %s:  %s", url, err.Error())
		}
	}
	return nil
}

// tagToString returns string representation of given interface
func tagToString(v interface{}) string {
	return fmt.Sprintf("%v", v)
}

// isItemFiltered tests whether item is part of non empty array, if not,
// returns true
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

// getMainTag returns value of a tag determining membership of datapoint
// to a group of points
func getMainTag(datapoint datapoint) string {
	var mainTag interface{}
	if v, ok := datapoint.Tags["path"]; ok {
		mainTag = v
	} else if v, ok := datapoint.Tags["interface"]; ok {
		mainTag = v
	} else if v, ok := datapoint.Tags["container_id"]; ok {
		mainTag = v
	}
	return tagToString(mainTag)
}

func init() {
	inputs.Add("dcos", func() telegraf.Input { return &Dcos{} })
}
