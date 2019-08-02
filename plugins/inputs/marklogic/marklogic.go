package marklogic

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	dac "github.com/go-http-digest-auth-client"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Marklogic struct {

	Hosts         []string `toml:"hosts"`
	DigestUsername string   `toml:"digest_username"`
	DigestPassword string   `toml:"digest_password"`

	// HTTP client & request
	client *http.Client
}

// NewMarklogic return a new instance of Marklogic with a default http client
func NewMarklogic() *Marklogic {
	tr := &http.Transport{ResponseHeaderTimeout: time.Duration(3 * time.Second)}
	client := &http.Client{
		Transport: tr,
		Timeout:   time.Duration(4 * time.Second),
	}
	return &Marklogic{client: client}
}

type MLHosts struct {
	HostStatus struct {
		ID                  string `json:"id"`
		Name                string `json:"name"`
		Version             string `json:"version"`
		EffectiveVersion    int    `json:"effective-version"`
		HostMode            string `json:"host-mode"`
		HostModeDescription string `json:"host-mode-description"`
		Meta                struct {
			URI         string    `json:"uri"`
			CurrentTime time.Time `json:"current-time"`
			ElapsedTime struct {
				Units string  `json:"units"`
				Value float64 `json:"value"`
			} `json:"elapsed-time"`
		} `json:"meta"`
		StatusProperties struct {
			Online struct {
				Units string `json:"units"`
				Value bool   `json:"value"`
			} `json:"online"`
			LoadProperties struct {
				TotalLoad struct {
					Units string  `json:"units"`
					Value float64 `json:"value"`
				} `json:"total-load"`
			} `json:"load-properties"`
			StatusDetail struct {
				HostMode          struct {
					Units string `json:"units"`
					Value string `json:"value"`
				} `json:"host-mode"`
				Cpus struct {
					Units string `json:"units"`
					Value int    `json:"value"`
				} `json:"cpus"`
				Cores struct {
					Units string `json:"units"`
					Value int    `json:"value"`
				} `json:"cores"`
				CoreThreads struct {
					Units string `json:"units"`
					Value int    `json:"value"`
				} `json:"core-threads"`
				TotalCPUStatUser      float64 `json:"total-cpu-stat-user"`
				TotalCPUStatNice      float64  `json:"total-cpu-stat-nice"`
				TotalCPUStatSystem    float64 `json:"total-cpu-stat-system"`
				TotalCPUStatIdle      float64 `json:"total-cpu-stat-idle"`
				TotalCPUStatGuest     float64  `json:"total-cpu-stat-guest"`
				MemoryProcessSize     struct {
					Units string `json:"units"`
					Value float64 `json:"value"`
				} `json:"memory-process-size"`
				MemoryProcessRss struct {
					Units string `json:"units"`
					Value float64    `json:"value"`
				} `json:"memory-process-rss"`
				MemorySystemTotal struct {
					Units string `json:"units"`
					Value float64    `json:"value"`
				} `json:"memory-system-total"`
				MemorySystemFree struct {
					Units string `json:"units"`
					Value float64    `json:"value"`
				} `json:"memory-system-free"`
				DataDirSpace struct {
					Units string `json:"units"`
					Value float64    `json:"value"`
				} `json:"data-dir-space"`
				QueryReadBytes struct {
					Units string `json:"units"`
					Value float64    `json:"value"`
				} `json:"query-read-bytes"`
				QueryReadLoad struct {
					Units string `json:"units"`
					Value float64    `json:"value"`
				} `json:"query-read-load"`
				HTTPServerReceiveBytes struct {
					Units string `json:"units"`
					Value float64    `json:"value"`
				} `json:"http-server-receive-bytes"`
				HTTPServerSendBytes struct {
					Units string `json:"units"`
					Value float64    `json:"value"`
				} `json:"http-server-send-bytes"`
			} `json:"status-detail"`
		} `json:"status-properties"`
	} `json:"host-status"`
}

func (c *Marklogic) Description() string {
	return "Gathers host health status data from a Marklogic Cluster"
}

var sampleConfig = `
  ## List URLs of Marklogic hosts using Management API endpoint.
  # hosts = ["http://localhost:8002/manage/v2/hosts/${hostname}?view=status&format=json"]

	# Using HTTP Digest Authentication.
	# digest_username = "telegraf"
	# digest_password = "p@ssw0rd"
`

func (c *Marklogic) SampleConfig() string {
	return sampleConfig
}

// Gather read stats from all hosts configured in cluster.
func (c *Marklogic) Gather(accumulator telegraf.Accumulator) error {
	var wg sync.WaitGroup

	if len(c.Hosts) == 0 {
		c.Hosts = []string{"http://localhost:8002/manage/v2/hosts/ml1.local?view=status&format=json"}
	}

	// Range over all ML servers, gathering stats. Returns early in case of any error.
	for _, u := range c.Hosts {
		wg.Add(1)
		go func(host string) {
			defer wg.Done()
			if err := c.fetchAndInsertData(accumulator, host); err != nil {
				accumulator.AddError(fmt.Errorf("[host=%s]: %s", host, err))
			}
		}(u)
	}

	wg.Wait()

	return nil
}

func (c *Marklogic) fetchAndInsertData(acc telegraf.Accumulator, host string) error {
	if c.client == nil {
		c.client = &http.Client{
			Transport: &http.Transport{
				ResponseHeaderTimeout: time.Duration(3 * time.Second),
			},
			Timeout: time.Duration(4 * time.Second),
		}
	}

t := dac.NewTransport(c.DigestUsername, c.DigestPassword)
req, err := http.NewRequest("GET", host, nil)
if err != nil {
	return err
}

	response, err := t.RoundTrip(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return fmt.Errorf("Failed to get status from Marklogic Management API: HTTP responded %d", response.StatusCode)
	}

	// Decode the response JSON into a new lights struct
	stats := &MLHosts{}
	if err := json.NewDecoder(response.Body).Decode(stats); err != nil {
		return fmt.Errorf("Unable to decode MLHosts{} object from Management API response: %s", err)
	}

	// Build a map of tags
	tags := map[string]string{
    //"server":   hs.Host,
		"ml_hostname":  stats.HostStatus.Name,
		"id":  stats.HostStatus.ID,

	}

	// Build a map of field values
	fields := map[string]interface{}{

		"online":                             stats.HostStatus.StatusProperties.Online.Value,
    "total_cpu_stat_user":                stats.HostStatus.StatusProperties.StatusDetail.TotalCPUStatUser,
		"total_cpu_stat_system":              stats.HostStatus.StatusProperties.StatusDetail.TotalCPUStatSystem,
		"memory_process_size":                stats.HostStatus.StatusProperties.StatusDetail.MemoryProcessSize.Value,
		"memory_process_rss":                 stats.HostStatus.StatusProperties.StatusDetail.MemoryProcessRss.Value,
		"memory_system_total":                stats.HostStatus.StatusProperties.StatusDetail.MemorySystemTotal.Value,
		"memory_system_free":                 stats.HostStatus.StatusProperties.StatusDetail.MemorySystemFree.Value,
		"num_cores":                          stats.HostStatus.StatusProperties.StatusDetail.Cores.Value,
		"total_load":                         stats.HostStatus.StatusProperties.LoadProperties.TotalLoad.Value,
		"data_dir_space":                     stats.HostStatus.StatusProperties.StatusDetail.DataDirSpace.Value,
		"query_read_bytes":                   stats.HostStatus.StatusProperties.StatusDetail.QueryReadBytes.Value,
		"query_read_load":                    stats.HostStatus.StatusProperties.StatusDetail.QueryReadLoad.Value,
		"http_server_receive_bytes":          stats.HostStatus.StatusProperties.StatusDetail.HTTPServerReceiveBytes.Value,
	  "http_server_send_bytes":             stats.HostStatus.StatusProperties.StatusDetail.HTTPServerSendBytes.Value,
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
