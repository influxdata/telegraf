package marklogic

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// Marklogic configuration toml
type Marklogic struct {
	URL      string   `toml:"url"`
	Hosts    []string `toml:"hosts"`
	Username string   `toml:"username"`
	Password string   `toml:"password"`
	Sources  []string

	tls.ClientConfig

	client *http.Client
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

// MarkLogic v2 management api endpoints for hosts status
const statsPath = "/manage/v2/hosts/"
const viewFormat = "view=status&format=json"

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
				Cpus                   MlPointInt `json:"cpus"`
				Cores                  MlPointInt `json:"cores"`
				TotalCPUStatUser       float64    `json:"total-cpu-stat-user"`
				TotalCPUStatSystem     float64    `json:"total-cpu-stat-system"`
				TotalCPUStatIdle       float64    `json:"total-cpu-stat-idle"`
				TotalCPUStatIowait     float64    `json:"total-cpu-stat-iowait"`
				MemoryProcessSize      MlPointInt `json:"memory-process-size"`
				MemoryProcessRss       MlPointInt `json:"memory-process-rss"`
				MemorySystemTotal      MlPointInt `json:"memory-system-total"`
				MemorySystemFree       MlPointInt `json:"memory-system-free"`
				MemoryProcessSwapSize  MlPointInt `json:"memory-process-swap-size"`
				MemorySize             MlPointInt `json:"memory-size"`
				HostSize               MlPointInt `json:"host-size"`
				LogDeviceSpace         MlPointInt `json:"log-device-space"`
				DataDirSpace           MlPointInt `json:"data-dir-space"`
				QueryReadBytes         MlPointInt `json:"query-read-bytes"`
				QueryReadLoad          MlPointInt `json:"query-read-load"`
				MergeReadLoad          MlPointInt `json:"merge-read-load"`
				MergeWriteLoad         MlPointInt `json:"merge-write-load"`
				HTTPServerReceiveBytes MlPointInt `json:"http-server-receive-bytes"`
				HTTPServerSendBytes    MlPointInt `json:"http-server-send-bytes"`
			} `json:"status-detail"`
		} `json:"status-properties"`
	} `json:"host-status"`
}

// Init parse all source URLs and place on the Marklogic struct
func (c *Marklogic) Init() error {
	if len(c.URL) == 0 {
		c.URL = "http://localhost:8002/"
	}

	for _, u := range c.Hosts {
		base, err := url.Parse(c.URL)
		if err != nil {
			return err
		}

		base.Path = path.Join(base.Path, statsPath, u)
		addr := base.ResolveReference(base)

		addr.RawQuery = viewFormat
		u := addr.String()
		c.Sources = append(c.Sources, u)
	}
	return nil
}

// Gather metrics from HTTP Server.
func (c *Marklogic) Gather(accumulator telegraf.Accumulator) error {
	var wg sync.WaitGroup

	if c.client == nil {
		client, err := c.createHTTPClient()

		if err != nil {
			return err
		}
		c.client = client
	}

	// Range over all source URL's appended to the struct
	for _, serv := range c.Sources {
		//fmt.Printf("Encoded URL is %q\n", serv)
		wg.Add(1)
		go func(serv string) {
			defer wg.Done()
			if err := c.fetchAndInsertData(accumulator, serv); err != nil {
				accumulator.AddError(fmt.Errorf("[host=%s]: %s", serv, err))
			}
		}(serv)
	}

	wg.Wait()

	return nil
}

func (c *Marklogic) fetchAndInsertData(acc telegraf.Accumulator, address string) error {
	ml := &MlHost{}
	if err := c.gatherJSONData(address, ml); err != nil {
		return err
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
		"memory_process_swap_size":  ml.HostStatus.StatusProperties.StatusDetail.MemoryProcessSwapSize.Value,
		"memory_size":               ml.HostStatus.StatusProperties.StatusDetail.MemorySize.Value,
		"host_size":                 ml.HostStatus.StatusProperties.StatusDetail.HostSize.Value,
		"log_device_space":          ml.HostStatus.StatusProperties.StatusDetail.LogDeviceSpace.Value,
		"data_dir_space":            ml.HostStatus.StatusProperties.StatusDetail.DataDirSpace.Value,
		"query_read_bytes":          ml.HostStatus.StatusProperties.StatusDetail.QueryReadBytes.Value,
		"query_read_load":           ml.HostStatus.StatusProperties.StatusDetail.QueryReadLoad.Value,
		"merge_read_load":           ml.HostStatus.StatusProperties.StatusDetail.MergeReadLoad.Value,
		"merge_write_load":          ml.HostStatus.StatusProperties.StatusDetail.MergeWriteLoad.Value,
		"http_server_receive_bytes": ml.HostStatus.StatusProperties.StatusDetail.HTTPServerReceiveBytes.Value,
		"http_server_send_bytes":    ml.HostStatus.StatusProperties.StatusDetail.HTTPServerSendBytes.Value,
	}

	// Accumulate the tags and values
	acc.AddFields("marklogic", fields, tags)

	return nil
}

func (c *Marklogic) createHTTPClient() (*http.Client, error) {
	tlsCfg, err := c.ClientConfig.TLSConfig()
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsCfg,
		},
		Timeout: 5 * time.Second,
	}

	return client, nil
}

func (c *Marklogic) gatherJSONData(address string, v interface{}) error {
	req, err := http.NewRequest("GET", address, nil)
	if err != nil {
		return err
	}

	if c.Username != "" || c.Password != "" {
		req.SetBasicAuth(c.Username, c.Password)
	}

	response, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("marklogic: API responded with status-code %d, expected %d",
			response.StatusCode, http.StatusOK)
	}

	return json.NewDecoder(response.Body).Decode(v)
}

func init() {
	inputs.Add("marklogic", func() telegraf.Input {
		return &Marklogic{}
	})
}
