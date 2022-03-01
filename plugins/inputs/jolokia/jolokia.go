package jolokia

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// Default http timeouts
var DefaultResponseHeaderTimeout = config.Duration(3 * time.Second)
var DefaultClientTimeout = config.Duration(4 * time.Second)

type Server struct {
	Name     string
	Host     string
	Username string
	Password string
	Port     string
}

type Metric struct {
	Name      string
	Mbean     string
	Attribute string
	Path      string
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
	jClient   JolokiaClient
	Context   string
	Mode      string
	Servers   []Server
	Metrics   []Metric
	Proxy     Server
	Delimiter string

	ResponseHeaderTimeout config.Duration `toml:"response_header_timeout"`
	ClientTimeout         config.Duration `toml:"client_timeout"`
	Log                   telegraf.Logger `toml:"-"`
}

const sampleConfig = `
  # DEPRECATED: the jolokia plugin has been deprecated in favor of the
  # jolokia2 plugin
  # see https://github.com/influxdata/telegraf/tree/master/plugins/inputs/jolokia2

  ## This is the context root used to compose the jolokia url
  ## NOTE that Jolokia requires a trailing slash at the end of the context root
  ## NOTE that your jolokia security policy must allow for POST requests.
  context = "/jolokia/"

  ## This specifies the mode used
  # mode = "proxy"
  #
  ## When in proxy mode this section is used to specify further
  ## proxy address configurations.
  ## Remember to change host address to fit your environment.
  # [inputs.jolokia.proxy]
  #   host = "127.0.0.1"
  #   port = "8080"

  ## Optional http timeouts
  ##
  ## response_header_timeout, if non-zero, specifies the amount of time to wait
  ## for a server's response headers after fully writing the request.
  # response_header_timeout = "3s"
  ##
  ## client_timeout specifies a time limit for requests made by this client.
  ## Includes connection time, any redirects, and reading the response body.
  # client_timeout = "4s"

  ## Attribute delimiter
  ##
  ## When multiple attributes are returned for a single
  ## [inputs.jolokia.metrics], the field name is a concatenation of the metric
  ## name, and the attribute name, separated by the given delimiter.
  # delimiter = "_"

  ## List of servers exposing jolokia read service
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

func (j *Jolokia) SampleConfig() string {
	return sampleConfig
}

func (j *Jolokia) Description() string {
	return "Read JMX metrics through Jolokia"
}

func (j *Jolokia) doRequest(req *http.Request) ([]map[string]interface{}, error) {
	resp, err := j.jClient.MakeRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Process response
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("response from url \"%s\" has status code %d (%s), expected %d (%s)",
			req.RequestURI,
			resp.StatusCode,
			http.StatusText(resp.StatusCode),
			http.StatusOK,
			http.StatusText(http.StatusOK))
		return nil, err
	}

	// read body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Unmarshal json
	var jsonOut []map[string]interface{}
	if err = json.Unmarshal(body, &jsonOut); err != nil {
		return nil, fmt.Errorf("error decoding JSON response: %s: %s", err, body)
	}

	return jsonOut, nil
}

func (j *Jolokia) prepareRequest(server Server, metrics []Metric) (*http.Request, error) {
	var jolokiaURL *url.URL
	context := j.Context // Usually "/jolokia/"

	var bulkBodyContent []map[string]interface{}
	for _, metric := range metrics {
		// Create bodyContent
		bodyContent := map[string]interface{}{
			"type":  "read",
			"mbean": metric.Mbean,
		}

		if metric.Attribute != "" {
			bodyContent["attribute"] = metric.Attribute
			if metric.Path != "" {
				bodyContent["path"] = metric.Path
			}
		}

		// Add target, only in proxy mode
		if j.Mode == "proxy" {
			serviceURL := fmt.Sprintf("service:jmx:rmi:///jndi/rmi://%s:%s/jmxrmi",
				server.Host, server.Port)

			target := map[string]string{
				"url": serviceURL,
			}

			if server.Username != "" {
				target["user"] = server.Username
			}

			if server.Password != "" {
				target["password"] = server.Password
			}

			bodyContent["target"] = target

			proxy := j.Proxy

			// Prepare ProxyURL
			proxyURL, err := url.Parse("http://" + proxy.Host + ":" + proxy.Port + context)
			if err != nil {
				return nil, err
			}
			if proxy.Username != "" || proxy.Password != "" {
				proxyURL.User = url.UserPassword(proxy.Username, proxy.Password)
			}

			jolokiaURL = proxyURL
		} else {
			serverURL, err := url.Parse("http://" + server.Host + ":" + server.Port + context)
			if err != nil {
				return nil, err
			}
			if server.Username != "" || server.Password != "" {
				serverURL.User = url.UserPassword(server.Username, server.Password)
			}

			jolokiaURL = serverURL
		}

		bulkBodyContent = append(bulkBodyContent, bodyContent)
	}

	requestBody, err := json.Marshal(bulkBodyContent)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", jolokiaURL.String(), bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-type", "application/json")

	return req, nil
}

func (j *Jolokia) extractValues(measurement string, value interface{}, fields map[string]interface{}) {
	if mapValues, ok := value.(map[string]interface{}); ok {
		for k2, v2 := range mapValues {
			j.extractValues(measurement+j.Delimiter+k2, v2, fields)
		}
	} else {
		fields[measurement] = value
	}
}

func (j *Jolokia) Gather(acc telegraf.Accumulator) error {
	if j.jClient == nil {
		j.Log.Warn("DEPRECATED: the jolokia plugin has been deprecated " +
			"in favor of the jolokia2 plugin " +
			"(https://github.com/influxdata/telegraf/tree/master/plugins/inputs/jolokia2)")

		tr := &http.Transport{ResponseHeaderTimeout: time.Duration(j.ResponseHeaderTimeout)}
		j.jClient = &JolokiaClientImpl{&http.Client{
			Transport: tr,
			Timeout:   time.Duration(j.ClientTimeout),
		}}
	}

	servers := j.Servers
	metrics := j.Metrics
	tags := make(map[string]string)

	for _, server := range servers {
		tags["jolokia_name"] = server.Name
		tags["jolokia_port"] = server.Port
		tags["jolokia_host"] = server.Host
		fields := make(map[string]interface{})

		req, err := j.prepareRequest(server, metrics)
		if err != nil {
			acc.AddError(fmt.Errorf("unable to create request: %s", err))
			continue
		}
		out, err := j.doRequest(req)
		if err != nil {
			acc.AddError(fmt.Errorf("error performing request: %s", err))
			continue
		}

		if len(out) != len(metrics) {
			acc.AddError(fmt.Errorf("did not receive the correct number of metrics in response. expected %d, received %d", len(metrics), len(out)))
			continue
		}
		for i, resp := range out {
			if status, ok := resp["status"]; ok && status != float64(200) {
				acc.AddError(fmt.Errorf("not expected status value in response body (%s:%s mbean=\"%s\" attribute=\"%s\"): %3.f",
					server.Host, server.Port, metrics[i].Mbean, metrics[i].Attribute, status))
				continue
			} else if !ok {
				acc.AddError(fmt.Errorf("missing status in response body"))
				continue
			}

			if values, ok := resp["value"]; ok {
				j.extractValues(metrics[i].Name, values, fields)
			} else {
				acc.AddError(fmt.Errorf("missing key 'value' in output response"))
			}
		}

		acc.AddFields("jolokia", fields, tags)
	}

	return nil
}

func init() {
	inputs.Add("jolokia", func() telegraf.Input {
		return &Jolokia{
			ResponseHeaderTimeout: DefaultResponseHeaderTimeout,
			ClientTimeout:         DefaultClientTimeout,
			Delimiter:             "_",
		}
	})
}
