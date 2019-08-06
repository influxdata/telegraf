package marklogic

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	dac "github.com/xinsnake/go-http-digest-auth-client"
)

type Marklogic struct {
	URL      string   `toml:"url"`
	Hosts    []string `toml:"hosts"`
	Username string   `toml:"username"`
	Password string   `toml:"password"`

	// HTTP client & request
	client *http.Client
}

// NewMarkLogic return a new instance of MarkLogic with a default http client
func NewMarklogic() *Marklogic {
	tr := &http.Transport{ResponseHeaderTimeout: time.Duration(3 * time.Second)}
	client := &http.Client{
		Transport: tr,
		Timeout:   time.Duration(4 * time.Second),
	}
	return &Marklogic{client: client}
}

type MlPointInt struct {
	Value int `json:"value"`
}

type MlPointFloat struct {
	Value float64 `json:"value"`
}

type MlPointBool struct {
	Value bool `json:"value"`
}

// MarkLogic v2 management api endpoint for hosts status
const statsPath = "/manage/v2/hosts/"
const viewFormat = "?view=status&format=json"

type MlHost struct {
	HostStatus struct {
		ID               string `json:"id"`
		Name             string `json:"name"`
		StatusProperties struct {
			Online         MlPointBool `json:"online"`
			LoadProperties struct {
				TotalLoad MlPointFloat `json:"total-load"`
			} `json:"load-properties"`
			RateProperties struct {
				TotalRate MlPointFloat `json:"total-rate"`
			} `json:"rate-properties"`
			StatusDetail struct {
				Cpus                   MlPointInt   `json:"cpus"`
				Cores                  MlPointInt   `json:"cores"`
				TotalCPUStatUser       float64      `json:"total-cpu-stat-user"`
				TotalCPUStatSystem     float64      `json:"total-cpu-stat-system"`
				TotalCPUStatIdle       float64      `json:"total-cpu-stat-idle"`
				TotalCPUStatIowait     float64      `json:"total-cpu-stat-iowait"`
				MemoryProcessSize      MlPointInt   `json:"memory-process-size"`
				MemoryProcessRss       MlPointInt   `json:"memory-process-rss"`
				MemorySystemTotal      MlPointInt   `json:"memory-system-total"`
				MemorySystemFree       MlPointInt   `json:"memory-system-free"`
				MemorySize             MlPointInt   `json:"memory-size"`
				HostSize               MlPointInt   `json:"host-size"`
				DataDirSpace           MlPointInt   `json:"data-dir-space"`
				QueryReadBytes         MlPointInt   `json:"query-read-bytes"`
				QueryReadLoad          MlPointInt   `json:"query-read-load"`
				HTTPServerReceiveBytes MlPointInt `json:"http-server-receive-bytes"`
				HTTPServerSendBytes    MlPointInt `json:"http-server-send-bytes"`
			} `json:"status-detail"`
		} `json:"status-properties"`
	} `json:"host-status"`
}

func (c *Marklogic) Description() string {
	return "Gathers host health status data from a Marklogic Cluster"
}

var sampleConfig = `
	## Base URL of MarkLogic host for Management API endpoint.
  url = "http://localhost:8002"

	## List of specific hostnames in a cluster to retrieve information. At least (1) required.
  # hosts = ["hostname1", "hostname2"]

  ## Using HTTP Digest Authentication. This requires 'manage-user' role privileges
  # username = "telegraf"
  # password = "p@ssw0rd"
`

func (c *Marklogic) SampleConfig() string {
	return sampleConfig
}

// Gather read stats from all hosts configured in cluster.
func (c *Marklogic) Gather(accumulator telegraf.Accumulator) error {
	var wg sync.WaitGroup
	var url string

	if len(c.URL) == 0 {
		c.URL = string("http://localhost:8002")
	}

	// Range over all ML hostnames, gathering stats. Returns early in case of any error.
	for _, u := range c.Hosts {
		wg.Add(1)
		go func(host string) {
			defer wg.Done()

			url = string(c.URL + statsPath + host + viewFormat)
			if err := c.fetchAndInsertData(accumulator, url); err != nil {
				accumulator.AddError(fmt.Errorf("[host=%s]: %s", url, err))
			}
		}(u)
	}

	wg.Wait()

	return nil
}

func (c *Marklogic) fetchAndInsertData(acc telegraf.Accumulator, url string) error {
	if c.client == nil {
		c.client = &http.Client{
			Transport: &http.Transport{
				ResponseHeaderTimeout: time.Duration(3 * time.Second),
			},
			Timeout: time.Duration(4 * time.Second),
		}
	}

	t := dac.NewTransport(c.Username, c.Password)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	response, err := t.RoundTrip(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return fmt.Errorf("failed to get status from MarkLogic management api: HTTP responded %d", response.StatusCode)
	}

	// Decode the response JSON into a new lights struct
	ml := &MlHost{}
	if err := json.NewDecoder(response.Body).Decode(ml); err != nil {
		return fmt.Errorf("unable to decode MlHost{} object from management api response: %s", err)
	}

	// Build a map of tags
	tags := map[string]string{
		"source": ml.HostStatus.Name,
		"id":     ml.HostStatus.ID,
	}

	// Build a map of field values
	fields := map[string]interface{}{
		"online":                    ml.HostStatus.StatusProperties.Online.Value,
		"total_load":                ml.HostStatus.StatusProperties.LoadProperties.TotalLoad.Value,
		"total_rate":                ml.HostStatus.StatusProperties.RateProperties.TotalRate.Value,
		"ncpus":                     ml.HostStatus.StatusProperties.StatusDetail.Cpus.Value,
		"ncores":                    ml.HostStatus.StatusProperties.StatusDetail.Cores.Value,
		"total_cpu_stat_user":       ml.HostStatus.StatusProperties.StatusDetail.TotalCPUStatUser,
		"total_cpu_stat_system":     ml.HostStatus.StatusProperties.StatusDetail.TotalCPUStatSystem,
		"total_cpu_stat_idle":       ml.HostStatus.StatusProperties.StatusDetail.TotalCPUStatIdle,
		"total_cpu_stat_iowait":     ml.HostStatus.StatusProperties.StatusDetail.TotalCPUStatIowait,
		"memory_process_size":       ml.HostStatus.StatusProperties.StatusDetail.MemoryProcessSize.Value,
		"memory_process_rss":        ml.HostStatus.StatusProperties.StatusDetail.MemoryProcessRss.Value,
		"memory_system_total":       ml.HostStatus.StatusProperties.StatusDetail.MemorySystemTotal.Value,
		"memory_system_free":        ml.HostStatus.StatusProperties.StatusDetail.MemorySystemFree.Value,
		"memory_size":               ml.HostStatus.StatusProperties.StatusDetail.MemorySize.Value,
		"host_size":                 ml.HostStatus.StatusProperties.StatusDetail.HostSize.Value,
		"data_dir_space":            ml.HostStatus.StatusProperties.StatusDetail.DataDirSpace.Value,
		"query_read_bytes":          ml.HostStatus.StatusProperties.StatusDetail.QueryReadBytes.Value,
		"query_read_load":           ml.HostStatus.StatusProperties.StatusDetail.QueryReadLoad.Value,
		"http_server_receive_bytes": ml.HostStatus.StatusProperties.StatusDetail.HTTPServerReceiveBytes.Value,
		"http_server_send_bytes":    ml.HostStatus.StatusProperties.StatusDetail.HTTPServerSendBytes.Value,
	}

	// Accumulate the tags and values
	acc.AddFields("marklogic", fields, tags)

	return nil
}

func init() {
	inputs.Add("marklogic", func() telegraf.Input {
		return NewMarklogic()
	})
}
