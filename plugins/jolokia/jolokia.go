package jolokia

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	// "net/url"
	"strings"
	// "sync"

	"github.com/influxdb/telegraf/plugins"
)


type Server struct {
  Name string
  Host string
  Port string
}

type Metric struct {
  Name string
  Jmx string
  Pass []string
  Drop []string
}

type Jolokia struct {

  Context   string
  Servers   []Server
  Metrics   []Metric

}


func (j *Jolokia) SampleConfig() string {
	return `[jolokia]
  context = "/jolokia/read"

  [[jolokia.servers]]
        name = "stable"
        host = "192.168.103.2"
        port = "8180"

  [[jolokia.metrics]]
        name = "heap_memory_usage"
        jmx  = "/java.lang:type=Memory/HeapMemoryUsage"
        pass = ["used"]

  [[jolokia.metrics]]
        name = "memory_eden"
        jmx  = "/java.lang:type=MemoryPool,name=PS Eden Space/Usage"
        pass = ["used"]

  [[jolokia.metrics]]
        name = "heap_threads"
        jmx  = "/java.lang:type=Threading"
 #      drop = ["AllThread"]
        pass = ["CurrentThreadCpuTime","CurrentThreadUserTime","DaemonThreadCount","ThreadCount","TotalStartedThreadCount"]
`
}

func (j *Jolokia) Description() string {
	return "Read JMX metrics through Jolokia"
}



func getAttr(url string) (map[string]interface{}, error) {
  //make request
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

  // Process response
  if resp.StatusCode != http.StatusOK {
    err = fmt.Errorf("Response from url \"%s\" has status code %d (%s), expected %d (%s)",
      url,
      resp.StatusCode,
      http.StatusText(resp.StatusCode),
      http.StatusOK,
      http.StatusText(http.StatusOK))
    return nil, err
  }

  // read body
  body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

  // Unmarshal json
  var jsonOut map[string]interface{}
  if err = json.Unmarshal([]byte(body), &jsonOut); err != nil {
    return nil, errors.New("Error decoding JSON response")
  }

  return jsonOut, nil
}

func (m *Metric) shouldPass(field string) bool {

  if m.Pass != nil {

    for _, pass := range m.Pass{
      if strings.HasPrefix(field, pass) {
				return true
			}
    }

    return false
  }

  if m.Drop != nil {

    for _, drop := range m.Drop{
      if strings.HasPrefix(field, drop) {
				return false
			}
    }

    return true
  }

  return true
}

func (m *Metric) filterFields(fields map[string]interface{}) map[string]interface{} {

  for field, _ := range fields{
    if !m.shouldPass(field) {
			delete(fields, field)
		}
  }

  return fields
}


func (j *Jolokia) Gather(acc plugins.Accumulator) error {

  context := j.Context //"/jolokia/read"
  servers := j.Servers
  metrics := j.Metrics

  var tags = map[string]string{
    "group": "application_server",
  }

  for _, server := range servers {
    for _, metric := range metrics {

      measurement := metric.Name
      jmxPath := metric.Jmx

      tags["server"] = server.Name
			tags["port"] = server.Port
			tags["host"] = server.Host

      url := "http://" + server.Host + ":" + server.Port + context + jmxPath
      //fmt.Println(url)
      out, _ := getAttr(url)

      if values, ok := out["value"]; ok {
        switch values.(type) {
          case map[string]interface{}:
            acc.AddFields(measurement, metric.filterFields(values.(map[string]interface{})), tags)
          case interface{}:
            acc.Add(measurement, values.(interface{}), tags)
        }
      }else{
        fmt.Println("Missing key value")
      }
    }
  }

  return nil
}

func init() {
	plugins.Add("jolokia", func() plugins.Plugin {
		return &Jolokia{}
	})
}
