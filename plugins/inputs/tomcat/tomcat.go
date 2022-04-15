package tomcat

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type TomcatStatus struct {
	TomcatJvm        TomcatJvm         `xml:"jvm"`
	TomcatConnectors []TomcatConnector `xml:"connector"`
}

type TomcatJvm struct {
	JvmMemory      JvmMemoryStat       `xml:"memory"`
	JvmMemoryPools []JvmMemoryPoolStat `xml:"memorypool"`
}

type JvmMemoryStat struct {
	Free  int64 `xml:"free,attr"`
	Total int64 `xml:"total,attr"`
	Max   int64 `xml:"max,attr"`
}

type JvmMemoryPoolStat struct {
	Name           string `xml:"name,attr"`
	Type           string `xml:"type,attr"`
	UsageInit      int64  `xml:"usageInit,attr"`
	UsageCommitted int64  `xml:"usageCommitted,attr"`
	UsageMax       int64  `xml:"usageMax,attr"`
	UsageUsed      int64  `xml:"usageUsed,attr"`
}

type TomcatConnector struct {
	Name        string      `xml:"name,attr"`
	ThreadInfo  ThreadInfo  `xml:"threadInfo"`
	RequestInfo RequestInfo `xml:"requestInfo"`
}

type ThreadInfo struct {
	MaxThreads         int64 `xml:"maxThreads,attr"`
	CurrentThreadCount int64 `xml:"currentThreadCount,attr"`
	CurrentThreadsBusy int64 `xml:"currentThreadsBusy,attr"`
}
type RequestInfo struct {
	MaxTime        int   `xml:"maxTime,attr"`
	ProcessingTime int   `xml:"processingTime,attr"`
	RequestCount   int   `xml:"requestCount,attr"`
	ErrorCount     int   `xml:"errorCount,attr"`
	BytesReceived  int64 `xml:"bytesReceived,attr"`
	BytesSent      int64 `xml:"bytesSent,attr"`
}

type Tomcat struct {
	URL      string
	Username string
	Password string
	Timeout  config.Duration
	tls.ClientConfig

	client  *http.Client
	request *http.Request
}

func (s *Tomcat) Gather(acc telegraf.Accumulator) error {
	if s.client == nil {
		client, err := s.createHTTPClient()
		if err != nil {
			return err
		}
		s.client = client
	}

	if s.request == nil {
		_, err := url.Parse(s.URL)
		if err != nil {
			return err
		}
		request, err := http.NewRequest("GET", s.URL, nil)
		if err != nil {
			return err
		}
		request.SetBasicAuth(s.Username, s.Password)
		s.request = request
	}

	resp, err := s.client.Do(s.request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received HTTP status code %d from %q; expected 200",
			resp.StatusCode, s.URL)
	}

	var status TomcatStatus
	if err := xml.NewDecoder(resp.Body).Decode(&status); err != nil {
		return err
	}

	// add tomcat_jvm_memory measurements
	tcm := map[string]interface{}{
		"free":  status.TomcatJvm.JvmMemory.Free,
		"total": status.TomcatJvm.JvmMemory.Total,
		"max":   status.TomcatJvm.JvmMemory.Max,
	}
	acc.AddFields("tomcat_jvm_memory", tcm, nil)

	// add tomcat_jvm_memorypool measurements
	for _, mp := range status.TomcatJvm.JvmMemoryPools {
		tcmpTags := map[string]string{
			"name": mp.Name,
			"type": mp.Type,
		}

		tcmpFields := map[string]interface{}{
			"init":      mp.UsageInit,
			"committed": mp.UsageCommitted,
			"max":       mp.UsageMax,
			"used":      mp.UsageUsed,
		}

		acc.AddFields("tomcat_jvm_memorypool", tcmpFields, tcmpTags)
	}

	// add tomcat_connector measurements
	for _, c := range status.TomcatConnectors {
		name, err := strconv.Unquote(c.Name)
		if err != nil {
			name = c.Name
		}

		tccTags := map[string]string{
			"name": name,
		}

		tccFields := map[string]interface{}{
			"max_threads":          c.ThreadInfo.MaxThreads,
			"current_thread_count": c.ThreadInfo.CurrentThreadCount,
			"current_threads_busy": c.ThreadInfo.CurrentThreadsBusy,
			"max_time":             c.RequestInfo.MaxTime,
			"processing_time":      c.RequestInfo.ProcessingTime,
			"request_count":        c.RequestInfo.RequestCount,
			"error_count":          c.RequestInfo.ErrorCount,
			"bytes_received":       c.RequestInfo.BytesReceived,
			"bytes_sent":           c.RequestInfo.BytesSent,
		}

		acc.AddFields("tomcat_connector", tccFields, tccTags)
	}

	return nil
}

func (s *Tomcat) createHTTPClient() (*http.Client, error) {
	tlsConfig, err := s.ClientConfig.TLSConfig()
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
		Timeout: time.Duration(s.Timeout),
	}

	return client, nil
}

func init() {
	inputs.Add("tomcat", func() telegraf.Input {
		return &Tomcat{
			URL:      "http://127.0.0.1:8080/manager/status/all?XML=true",
			Username: "tomcat",
			Password: "s3cret",
			Timeout:  config.Duration(5 * time.Second),
		}
	})
}
