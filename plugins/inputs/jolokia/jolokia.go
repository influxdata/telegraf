package jolokia

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Server struct {
	Name     string
	Host     string
	Username string
	Password string
	Port     string
}

type Metric struct {
	Name string
	Jmx  string
}

type JolokiaClient interface {
	MakeRequest(req *http.Request) (*http.Response, error)
}

type JolokiaClientImpl struct {
	client *http.Client
}

func (c JolokiaClientImpl) MakeRequest(req *http.Request) (*http.Response, error) {
	return c.client.Do(req)
}

type Jolokia struct {
	jClient JolokiaClient
	Context string
	Servers []Server
	Metrics []Metric
}

func (j *Jolokia) SampleConfig() string {
	return `
  ## This is the context root used to compose the jolokia url
  context = "/jolokia/read"

  ## List of servers exposing jolokia read service
  [[inputs.jolokia.servers]]
    name = "stable"
    host = "192.168.103.2"
    port = "8180"
    # username = "myuser"
    # password = "mypassword"

  ## List of metrics collected on above servers
  ## Each metric consists in a name, a jmx path and either
  ## a pass or drop slice attribute.
  ## This collect all heap memory usage metrics.
  [[inputs.jolokia.metrics]]
    name = "heap_memory_usage"
    jmx  = "/java.lang:type=Memory/HeapMemoryUsage"
`
}

func (j *Jolokia) Description() string {
	return "Read JMX metrics through Jolokia"
}

func (j *Jolokia) getAttr(requestUrl *url.URL) (map[string]interface{}, error) {
	// Create + send request
	req, err := http.NewRequest("GET", requestUrl.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := j.jClient.MakeRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Process response
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("Response from url \"%s\" has status code %d (%s), expected %d (%s)",
			requestUrl,
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

func (j *Jolokia) Gather(acc telegraf.Accumulator) error {
	context := j.Context //"/jolokia/read"
	servers := j.Servers
	metrics := j.Metrics
	tags := make(map[string]string)

	for _, server := range servers {
		tags["server"] = server.Name
		tags["port"] = server.Port
		tags["host"] = server.Host
		fields := make(map[string]interface{})
		for _, metric := range metrics {

			measurement := metric.Name
			jmxPath := metric.Jmx

			// Prepare URL
			requestUrl, err := url.Parse("http://" + server.Host + ":" +
				server.Port + context + jmxPath)
			if err != nil {
				return err
			}
			if server.Username != "" || server.Password != "" {
				requestUrl.User = url.UserPassword(server.Username, server.Password)
			}

			out, _ := j.getAttr(requestUrl)

			if values, ok := out["value"]; ok {
				switch t := values.(type) {
				case map[string]interface{}:
					for k, v := range t {
						fields[measurement+"_"+k] = v
					}
				case interface{}:
					fields[measurement] = t
				}
			} else {
				fmt.Printf("Missing key 'value' in '%s' output response\n",
					requestUrl.String())
			}
		}
		acc.AddFields("jolokia", fields, tags)
	}

	return nil
}

func init() {
	inputs.Add("jolokia", func() telegraf.Input {
		tr := &http.Transport{ResponseHeaderTimeout: time.Duration(3 * time.Second)}
		client := &http.Client{
			Transport: tr,
			Timeout:   time.Duration(4 * time.Second),
		}
		return &Jolokia{jClient: &JolokiaClientImpl{client: client}}
	})
}
