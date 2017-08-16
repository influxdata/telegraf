package tomcat

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"

	"github.com/influxdata/telegraf"
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
}

var sampleconfig = `
  ## A Tomcat status URI to gather stats.
  ## Default is "http://127.0.0.1:8080/manager/status/all?XML=true".
  url = "http://127.0.0.1:8080/manager/status/all?XML=true"
  ## Credentials for status URI.
  ## Default is tomcat/s3cret.
  username = "tomcat"
  password = "s3cret"
`

func (s *Tomcat) Description() string {
	return "A Telegraf plugin to collect tomcat metrics."
}

func (s *Tomcat) SampleConfig() string {
	return sampleconfig
}

func (s *Tomcat) Gather(acc telegraf.Accumulator) error {

	if s.URL == "" {
		s.URL = "http://127.0.0.1:8080/manager/status/all?XML=true"
	}

	if s.Username == "" {
		s.Username = "tomcat"
	}

	if s.Password == "" {
		s.Password = "s3cret"
	}

	_, err := url.Parse(s.URL)
	if err != nil {
		return fmt.Errorf("Unable to parse address '%s': %s", s.URL, err)
	}

	req, err := http.NewRequest("GET", s.URL, nil)
	req.SetBasicAuth(s.Username, s.Password)
	cli := &http.Client{}
	resp, err := cli.Do(req)
	if err != nil {
		return fmt.Errorf("Unable to call URL '%s': %s", s.URL, err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	var status TomcatStatus
	xml.Unmarshal(body, &status)

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
			return fmt.Errorf("Unable to unquote name '%s': %s", c.Name, err)
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

func init() {
	inputs.Add("tomcat", func() telegraf.Input { return &Tomcat{} })
}
