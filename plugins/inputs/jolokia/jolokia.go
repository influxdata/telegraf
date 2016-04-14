package jolokia

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
	"bytes"

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
	Mbean string
	Attribute string
	Path string
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
	Mode string
	Servers []Server
	Metrics []Metric
	Proxy Server
}

func (j *Jolokia) SampleConfig() string {
	return `
  # This is the context root used to compose the jolokia url
  context = "/jolokia"

  # This specifies the mode used
  # mode = "proxy"
  #
  # When in proxy mode this section is used to specify further proxy address configurations.
  # Remember to change servers addresses
  # [inputs.jolokia.proxy]
  # host = "127.0.0.1"
  # port = "8080"


  # List of servers exposing jolokia read service
  [[inputs.jolokia.servers]]
    name = "as-server-01"
    host = "127.0.0.1"
    port = "8080"
    # username = "myuser"
    # password = "mypassword"

  ## List of metrics collected on above servers
  ## Each metric consists in a name, a jmx path and either
  ## a pass or drop slice attribute.
  ## This collect all heap memory usage metrics.
  [[inputs.jolokia.metrics]]
    name = "heap_memory_usage"
    mbean  = "java.lang:type=Memory"
    attribute = "HeapMemoryUsage"

  ## This collect thread counts metrics.
  [[inputs.jolokia.metrics]]
    name = "thread_count"
    mbean  = "java.lang:type=Threading"
		attribute = "TotalStartedThreadCount,ThreadCount,DaemonThreadCount,PeakThreadCount"

  ## This collect number of class loaded/unloaded counts metrics.
  [[inputs.jolokia.metrics]]
    name = "class_count"
    mbean  = "java.lang:type=ClassLoading"
		attribute = "LoadedClassCount,UnloadedClassCount,TotalLoadedClassCount"
`
}

func (j *Jolokia) Description() string {
	return "Read JMX metrics through Jolokia"
}

func (j *Jolokia) doRequest(req *http.Request) (map[string]interface{}, error) {

	resp, err := j.jClient.MakeRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Process response
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("Response from url \"%s\" has status code %d (%s), expected %d (%s)",
			req.RequestURI,
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

	if status, ok := jsonOut["status"]; ok {
		if status != float64(200) {
			return nil, fmt.Errorf("Not expected status value in response body: %3.f", status)
		}
	} else {
		return nil, fmt.Errorf("Missing status in response body")
	}

	return jsonOut, nil
}

func (j *Jolokia) getAttr(requestUrl *url.URL) (map[string]interface{}, error) {
	// Create + send request
	req, err := http.NewRequest("GET", requestUrl.String(), nil)
	if err != nil {
		return nil, err
	}

	return j.doRequest(req)
}


func (j *Jolokia) collectMeasurement(measurement string, out map[string]interface{}, fields map[string]interface{}) {

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
		fmt.Printf("Missing key 'value' in output response\n")
	}

}


func (j *Jolokia) Gather(acc telegraf.Accumulator) error {
	context := j.Context // Usually "/jolokia"
	servers := j.Servers
	metrics := j.Metrics
	tags := make(map[string]string)
	mode := j.Mode

	if( mode == "agent" || mode == ""){

		for _, server := range servers {
			tags["server"] = server.Name
			tags["port"] = server.Port
			tags["host"] = server.Host
			fields := make(map[string]interface{})
			for _, metric := range metrics {

				measurement := metric.Name
				jmxPath := "/" + metric.Mbean
				if metric.Attribute != "" {
					jmxPath = jmxPath + "/" + metric.Attribute

					if metric.Path != "" {
						jmxPath = jmxPath + "/" + metric.Path
					}
				}

			// Prepare URL
				requestUrl, err := url.Parse("http://" + server.Host + ":" +
				server.Port + context + "/read" + jmxPath)
				if err != nil {
					return err
				}
				if server.Username != "" || server.Password != "" {
					requestUrl.User = url.UserPassword(server.Username, server.Password)
				}
				out, _ := j.getAttr(requestUrl)
				j.collectMeasurement(measurement, out, fields)
			}
			acc.AddFields("jolokia", fields, tags)
		}

	} else if ( mode == "proxy") {

		proxy := j.Proxy

		// Prepare ProxyURL
		proxyURL, err := url.Parse("http://" + proxy.Host + ":" +
		proxy.Port + context)
		if err != nil {
			return err
		}
		if proxy.Username != "" || proxy.Password != "" {
			proxyURL.User = url.UserPassword(proxy.Username, proxy.Password)
		}

		for _, server := range servers {
			tags["server"] = server.Name
			tags["port"] = server.Port
			tags["host"] = server.Host
			fields := make(map[string]interface{})
			for _, metric := range metrics {

				measurement := metric.Name
				// Prepare URL
				serviceUrl := fmt.Sprintf("service:jmx:rmi:///jndi/rmi://%s:%s/jmxrmi", server.Host, server.Port)

				target := map[string]string{
					"url": serviceUrl,
				}

				if server.Username != "" {
					target["user"] = server.Username
				}

				if server.Password != "" {
					target["password"] = server.Password
				}

				// Create + send request
				bodyContent := map[string]interface{}{
					"type": "read",
					"mbean": metric.Mbean,
					"target": target,
				}

				if metric.Attribute != "" {
					bodyContent["attribute"] = metric.Attribute
					if metric.Path != "" {
						bodyContent["path"] = metric.Path
					}
				}

				requestBody, err := json.Marshal(bodyContent)

				req, err := http.NewRequest("POST", proxyURL.String(), bytes.NewBuffer(requestBody))

				if err != nil {
					return err
				}

				req.Header.Add("Content-type", "application/json")

				out, err := j.doRequest(req)
				
				if err != nil {
					fmt.Printf("Error handling response: %s\n", err)
				}else {
					j.collectMeasurement(measurement, out, fields)
				}
			}
			acc.AddFields("jolokia", fields, tags)
		}

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
