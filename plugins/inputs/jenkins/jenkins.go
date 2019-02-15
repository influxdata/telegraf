package jenkins

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/internal/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// Jenkins plugin gathers information about the nodes and jobs running in a jenkins instance.
type Jenkins struct {
	URL      string
	Username string
	Password string
	// HTTP Timeout specified as a string - 3s, 1m, 1h
	ResponseTimeout internal.Duration

	tls.ClientConfig
	client *client

	MaxConnections    int               `toml:"max_connections"`
	MaxBuildAge       internal.Duration `toml:"max_build_age"`
	MaxSubJobDepth    int               `toml:"max_subjob_depth"`
	MaxSubJobPerLayer int               `toml:"max_subjob_per_layer"`
	JobExclude        []string          `toml:"job_exclude"`
	jobFilter         filter.Filter

	NodeExclude []string `toml:"node_exclude"`
	nodeFilter  filter.Filter

	semaphore chan struct{}
}

const sampleConfig = `
  ## The Jenkins URL
  url = "http://my-jenkins-instance:8080"
  # username = "admin"
  # password = "admin"

  ## Set response_timeout
  response_timeout = "5s"

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use SSL but skip chain & host verification
  # insecure_skip_verify = false

  ## Optional Max Job Build Age filter
  ## Default 1 hour, ignore builds older than max_build_age
  # max_build_age = "1h"

  ## Optional Sub Job Depth filter
  ## Jenkins can have unlimited layer of sub jobs
  ## This config will limit the layers of pulling, default value 0 means
  ## unlimited pulling until no more sub jobs
  # max_subjob_depth = 0

  ## Optional Sub Job Per Layer
  ## In workflow-multibranch-plugin, each branch will be created as a sub job.
  ## This config will limit to call only the lasted branches in each layer, 
  ## empty will use default value 10
  # max_subjob_per_layer = 10

  ## Jobs to exclude from gathering
  # job_exclude = [ "job1", "job2/subjob1/subjob2", "job3/*"]

  ## Nodes to exclude from gathering
  # node_exclude = [ "node1", "node2" ]

  ## Worker pool for jenkins plugin only
  ## Empty this field will use default value 5
  # max_connections = 5
`

// measurement
const (
	measurementNode = "jenkins_node"
	measurementJob  = "jenkins_job"
)

// SampleConfig implements telegraf.Input interface
func (j *Jenkins) SampleConfig() string {
	return sampleConfig
}

// Description implements telegraf.Input interface
func (j *Jenkins) Description() string {
	return "Read jobs and cluster metrics from Jenkins instances"
}

// Gather implements telegraf.Input interface
func (j *Jenkins) Gather(acc telegraf.Accumulator) error {
	if j.client == nil {
		client, err := j.newHTTPClient()
		if err != nil {
			return err
		}
		if err = j.initialize(client); err != nil {
			return err
		}
	}

	j.gatherNodesData(acc)
	j.gatherJobs(acc)

	return nil
}

func (j *Jenkins) newHTTPClient() (*http.Client, error) {
	tlsCfg, err := j.ClientConfig.TLSConfig()
	if err != nil {
		return nil, fmt.Errorf("error parse jenkins config[%s]: %v", j.URL, err)
	}
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsCfg,
			MaxIdleConns:    j.MaxConnections,
		},
		Timeout: j.ResponseTimeout.Duration,
	}, nil
}

// seperate the client as dependency to use httptest Client for mocking
func (j *Jenkins) initialize(client *http.Client) error {
	var err error

	// init job filter
	j.jobFilter, err = filter.Compile(j.JobExclude)
	if err != nil {
		return fmt.Errorf("error compile job filters[%s]: %v", j.URL, err)
	}

	// init node filter
	j.nodeFilter, err = filter.Compile(j.NodeExclude)
	if err != nil {
		return fmt.Errorf("error compile node filters[%s]: %v", j.URL, err)
	}

	// init tcp pool with default value
	if j.MaxConnections <= 0 {
		j.MaxConnections = 5
	}

	// default sub jobs can be acquired
	if j.MaxSubJobPerLayer <= 0 {
		j.MaxSubJobPerLayer = 10
	}

	j.semaphore = make(chan struct{}, j.MaxConnections)

	j.client = newClient(client, j.URL, j.Username, j.Password, j.MaxConnections)

	return j.client.init()
}

func (j *Jenkins) gatherNodeData(n node, acc telegraf.Accumulator) error {

	tags := map[string]string{}
	if n.DisplayName == "" {
		return fmt.Errorf("error empty node name")
	}

	tags["node_name"] = n.DisplayName
	// filter out excluded node_name
	if j.nodeFilter != nil && j.nodeFilter.Match(tags["node_name"]) {
		return nil
	}

	tags["arch"] = n.MonitorData.HudsonNodeMonitorsArchitectureMonitor

	tags["status"] = "online"
	if n.Offline {
		tags["status"] = "offline"
	}
	monitorData := n.MonitorData
	if monitorData.HudsonNodeMonitorsArchitectureMonitor == "" {
		return errors.New("empty monitor data, please check your permission")
	}
	tags["disk_path"] = monitorData.HudsonNodeMonitorsDiskSpaceMonitor.Path
	tags["temp_path"] = monitorData.HudsonNodeMonitorsTemporarySpaceMonitor.Path

	fields := map[string]interface{}{
		"response_time":    monitorData.HudsonNodeMonitorsResponseTimeMonitor.Average,
		"disk_available":   monitorData.HudsonNodeMonitorsDiskSpaceMonitor.Size,
		"temp_available":   monitorData.HudsonNodeMonitorsTemporarySpaceMonitor.Size,
		"swap_available":   monitorData.HudsonNodeMonitorsSwapSpaceMonitor.SwapAvailable,
		"memory_available": monitorData.HudsonNodeMonitorsSwapSpaceMonitor.MemoryAvailable,
		"swap_total":       monitorData.HudsonNodeMonitorsSwapSpaceMonitor.SwapTotal,
		"memory_total":     monitorData.HudsonNodeMonitorsSwapSpaceMonitor.MemoryTotal,
	}
	acc.AddFields(measurementNode, fields, tags)

	return nil
}

func (j *Jenkins) gatherNodesData(acc telegraf.Accumulator) {

	nodeResp, err := j.client.getAllNodes(context.Background())
	if err != nil {
		acc.AddError(err)
		return
	}
	// get node data
	for _, node := range nodeResp.Computers {
		err = j.gatherNodeData(node, acc)
		if err == nil {
			continue
		}
		acc.AddError(err)
	}
}

func (j *Jenkins) gatherJobs(acc telegraf.Accumulator) {
	js, err := j.client.getJobs(context.Background(), nil)
	if err != nil {
		acc.AddError(err)
		return
	}
	var wg sync.WaitGroup
	for _, job := range js.Jobs {
		wg.Add(1)
		go func(name string, wg *sync.WaitGroup, acc telegraf.Accumulator) {
			defer wg.Done()
			if err := j.getJobDetail(jobRequest{
				name:    name,
				parents: []string{},
				layer:   0,
			}, acc); err != nil {
				acc.AddError(err)
			}
		}(job.Name, &wg, acc)
	}
	wg.Wait()
}

// wrap the tcp request with doGet
// block tcp request if buffered channel is full
func (j *Jenkins) doGet(tcp func() error) error {
	j.semaphore <- struct{}{}
	if err := tcp(); err != nil {
		<-j.semaphore
		return err
	}
	<-j.semaphore
	return nil
}

func (j *Jenkins) getJobDetail(jr jobRequest, acc telegraf.Accumulator) error {
	if j.MaxSubJobDepth > 0 && jr.layer == j.MaxSubJobDepth {
		return nil
	}
	// filter out excluded job.
	if j.jobFilter != nil && j.jobFilter.Match(jr.hierarchyName()) {
		return nil
	}

	js, err := j.client.getJobs(context.Background(), &jr)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	for k, ij := range js.Jobs {
		if k < len(js.Jobs)-j.MaxSubJobPerLayer-1 {
			continue
		}
		wg.Add(1)
		// schedule tcp fetch for inner jobs
		go func(ij innerJob, jr jobRequest, acc telegraf.Accumulator) {
			defer wg.Done()
			if err := j.getJobDetail(jobRequest{
				name:    ij.Name,
				parents: jr.combined(),
				layer:   jr.layer + 1,
			}, acc); err != nil {
				acc.AddError(err)
			}
		}(ij, jr, acc)
	}
	wg.Wait()

	// collect build info
	number := js.LastBuild.Number
	if number < 1 {
		// no build info
		return nil
	}
	build, err := j.client.getBuild(context.Background(), jr, number)
	if err != nil {
		return err
	}

	if build.Building {
		log.Printf("D! Ignore running build on %s, build %v", jr.name, number)
		return nil
	}

	// stop if build is too old
	// Higher up in gatherJobs
	cutoff := time.Now().Add(-1 * j.MaxBuildAge.Duration)

	// Here we just test
	if build.GetTimestamp().Before(cutoff) {
		return nil
	}

	gatherJobBuild(jr, build, acc)
	return nil
}

type nodeResponse struct {
	Computers []node `json:"computer"`
}

type node struct {
	DisplayName string      `json:"displayName"`
	Offline     bool        `json:"offline"`
	MonitorData monitorData `json:"monitorData"`
}

type monitorData struct {
	HudsonNodeMonitorsArchitectureMonitor string           `json:"hudson.node_monitors.ArchitectureMonitor"`
	HudsonNodeMonitorsDiskSpaceMonitor    nodeSpaceMonitor `json:"hudson.node_monitors.DiskSpaceMonitor"`
	HudsonNodeMonitorsResponseTimeMonitor struct {
		Average int64 `json:"average"`
	} `json:"hudson.node_monitors.ResponseTimeMonitor"`
	HudsonNodeMonitorsSwapSpaceMonitor struct {
		SwapAvailable   float64 `json:"availableSwapSpace"`
		SwapTotal       float64 `json:"totalSwapSpace"`
		MemoryAvailable float64 `json:"availablePhysicalMemory"`
		MemoryTotal     float64 `json:"totalPhysicalMemory"`
	} `json:"hudson.node_monitors.SwapSpaceMonitor"`
	HudsonNodeMonitorsTemporarySpaceMonitor nodeSpaceMonitor `json:"hudson.node_monitors.TemporarySpaceMonitor"`
}

type nodeSpaceMonitor struct {
	Path string  `json:"path"`
	Size float64 `json:"size"`
}

type jobResponse struct {
	LastBuild jobBuild   `json:"lastBuild"`
	Jobs      []innerJob `json:"jobs"`
	Name      string     `json:"name"`
}

type innerJob struct {
	Name  string `json:"name"`
	URL   string `json:"url"`
	Color string `json:"color"`
}

type jobBuild struct {
	Number int64
	URL    string
}

type buildResponse struct {
	Building  bool   `json:"building"`
	Duration  int64  `json:"duration"`
	Result    string `json:"result"`
	Timestamp int64  `json:"timestamp"`
}

func (b *buildResponse) GetTimestamp() time.Time {
	return time.Unix(0, int64(b.Timestamp)*int64(time.Millisecond))
}

const (
	nodePath = "/computer/api/json"
	jobPath  = "/api/json"
)

type jobRequest struct {
	name    string
	parents []string
	layer   int
}

func (jr jobRequest) combined() []string {
	return append(jr.parents, jr.name)
}

func (jr jobRequest) URL() string {
	return "/job/" + strings.Join(jr.combined(), "/job/") + jobPath
}

func (jr jobRequest) buildURL(number int64) string {
	return "/job/" + strings.Join(jr.combined(), "/job/") + "/" + strconv.Itoa(int(number)) + jobPath
}

func (jr jobRequest) hierarchyName() string {
	return strings.Join(jr.combined(), "/")
}

func (jr jobRequest) parentsString() string {
	return strings.Join(jr.parents, "/")
}

func gatherJobBuild(jr jobRequest, b *buildResponse, acc telegraf.Accumulator) {
	tags := map[string]string{"name": jr.name, "parents": jr.parentsString(), "result": b.Result}
	fields := make(map[string]interface{})
	fields["duration"] = b.Duration
	fields["result_code"] = mapResultCode(b.Result)

	acc.AddFields(measurementJob, fields, tags, b.GetTimestamp())
}

// perform status mapping
func mapResultCode(s string) int {
	switch strings.ToLower(s) {
	case "success":
		return 0
	case "failure":
		return 1
	case "not_built":
		return 2
	case "unstable":
		return 3
	case "aborted":
		return 4
	}
	return -1
}

func init() {
	inputs.Add("jenkins", func() telegraf.Input {
		return &Jenkins{
			MaxBuildAge:       internal.Duration{Duration: time.Duration(time.Hour)},
			MaxConnections:    5,
			MaxSubJobPerLayer: 10,
		}
	})
}
