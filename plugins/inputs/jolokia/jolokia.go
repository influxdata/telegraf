package jolokia

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Server struct {
	Name     string
	Url      string
	Headertimeout  internal.Duration
        Requesttimeout  internal.Duration
	Username string
	Password string
	// Path to CA file
	Ca    string
	// Path to host cert file
	Cert  string
	// Path to cert key file
	Key   string
	// Use SSL but skip chain & host verification
	Insecureverify bool
}

type Metric struct {
	Name      string
	Mbean     string
	Attribute string
	Path      string
}

type Jolokia struct {
	Context         string
	Mode            string
	Proxy           Server
	Servers         []Server
	Metrics	        []Metric

}

const sampleConfig = `
  ## This is the context root used to compose the jolokia url
  ## NOTE that your jolokia security policy must allow for POST requests.
  context = "/jolokia"

  ## This specifies the mode used
  # mode = "proxy"
  #
  ## When in proxy mode this section is used to specify further
  ## proxy address configurations.
  ## Remember to change host address to fit your environment.
  # [inputs.jolokia.proxy]
  #   url = "localhost:8080"
  #   headertimeout = 30
  #   requesttimeout = 30
  #   username = "myuser"
  #   password = "mypassword"
  ## Optional SSL Config
  #   ca = "/etc/telegraf/ca.pem"
  #   cert = "/etc/telegraf/cert.pem"
  #   key = "/etc/telegraf/key.pem"
  ## Use SSL but skip chain & host verification
  #   insecureverify = false

  ## List of servers exposing jolokia read service
  [[inputs.jolokia.servers]]
    name = "as-server-01"
    url = "http://as-server-01:8080"
    # headertimeout = 30
    # requesttimeout = 30
    # username = "myuser"
    # password = "mypassword"
    ## Optional SSL Config
    # ca = "/etc/telegraf/ca.pem"
    # cert = "/etc/telegraf/cert.pem"
    # key = "/etc/telegraf/key.pem"
    ## Use SSL but skip chain & host verification
    # insecureverify = false

  ## List of metrics collected on above servers
  ## Each metric consists in a name, a jmx path and either
  ## a pass or drop slice attribute.
  ## This collect all heap memory usage metrics.
  [[inputs.jolokia.metrics]]
    name = "heap_memory_usage"
    mbean  = "java.lang:type=Memory"
    attribute = "HeapMemoryUsage"

  ## This collect thread counts metrics.
  [[inputs.jolokia.metrics]]
    name = "thread_count"
    mbean  = "java.lang:type=Threading"
    attribute = "TotalStartedThreadCount,ThreadCount,DaemonThreadCount,PeakThreadCount"

  ## This collect number of class loaded/unloaded counts metrics.
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

func (j *Jolokia) createHttpClient(server Server) (*http.Client, error) {
	var tr *http.Transport

	if server.Headertimeout.Duration < time.Second {
		server.Headertimeout.Duration = time.Second * 5
	}

        if server.Requesttimeout.Duration < time.Second {
                server.Requesttimeout.Duration = time.Second * 5
        }

	serverUrl, err := url.Parse(server.Url)
	if err != nil || serverUrl.String() == "" {
		return nil, err
        }

	if serverUrl.Scheme == "https" {
		if server.Cert == "" && server.Key == "" && server.Ca == "" {
			err = fmt.Errorf("No SSL configuration provided")
			return nil,err
		}
		tlsCfg, err := internal.GetTLSConfig(
			server.Cert, server.Key, server.Ca, server.Insecureverify)
		if err != nil {
			return nil, err
		}
		tr = &http.Transport{
			ResponseHeaderTimeout: server.Headertimeout.Duration,
			TLSClientConfig:       tlsCfg,
		}
	} else {
		tr = &http.Transport{
			ResponseHeaderTimeout: server.Headertimeout.Duration,
		}
	}

	client := &http.Client{
		Transport: tr,
		Timeout:   server.Requesttimeout.Duration,
	}

	return client, nil
}

func (j *Jolokia) doRequest(server Server,metric Metric) (map[string]interface{}, error) {
	var client *http.Client

        req, err := j.prepareRequest(server, metric)
	if err != nil || req.URL.String() == "" {
		return nil, err
	}

	if j.Mode == "proxy" {
		client, err = j.createHttpClient(j.Proxy)
		if err != nil {
			return nil, err
		}
	} else {
		client, err = j.createHttpClient(server)
		if err != nil {
			return nil, err
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Process response
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("Response from url \"%s\" has status code %d (%s), expected %d (%s)",
			req.URL.String(),
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
			return nil, fmt.Errorf("Not expected status value in response body: %3.f",
				status)
		}
	} else {
		return nil, fmt.Errorf("Missing status in response body")
	}

	return jsonOut, nil
}

func (j *Jolokia) prepareRequest(server Server, metric Metric) (*http.Request, error) {
	var jolokiaUrl *url.URL

	context := j.Context // Usually "/jolokia"

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

        serverUrl, err := url.Parse(server.Url + context)
        if err != nil || serverUrl.String() == "" {
                return nil, err
        }

	// Add target, only in proxy mode
	if j.Mode == "proxy" {
		serviceUrl := fmt.Sprintf("service:jmx:rmi:///jndi/rmi://%s/jmxrmi",
			serverUrl.Host)

		target := map[string]string{
			"url": serviceUrl,
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
		proxyUrl, err := url.Parse(proxy.Url + context)
		if err != nil || proxyUrl.String() == "" {
			return nil, err
		}
		if proxy.Username != "" || proxy.Password != "" {
			proxyUrl.User = url.UserPassword(proxy.Username, proxy.Password)
		}

		jolokiaUrl = proxyUrl

	} else {
		if server.Username != "" || server.Password != "" {
			serverUrl.User = url.UserPassword(server.Username, server.Password)
		}

		jolokiaUrl = serverUrl
	}

	requestBody, err := json.Marshal(bodyContent)

	req, err := http.NewRequest("POST", jolokiaUrl.String(), bytes.NewBuffer(requestBody))

	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-type", "application/json")

	return req, nil
}

func (j *Jolokia) Gather(acc telegraf.Accumulator) error {
	servers := j.Servers
	metrics := j.Metrics
	tags := make(map[string]string)


	for _, server := range servers {
		tags["jolokia_name"] = server.Name
		tags["jolokia_url"] = server.Url
		fields := make(map[string]interface{})

		for _, metric := range metrics {
			measurement := metric.Name

			out, err := j.doRequest(server,metric)

			if err != nil {
				fmt.Printf("Error handling response: %s\n", err)
			} else {

				if values, ok := out["value"]; ok {
					switch t := values.(type) {
					case map[string]interface{}:
						for k, v := range t {
							switch t2 := v.(type) {
							case map[string]interface{}:
								for k2, v2 := range t2 {
									fields[measurement+"_"+k+"_"+k2] = v2
								}
							case interface{}:
								fields[measurement+"_"+k] = t2
							}
						}
					case interface{}:
						fields[measurement] = t
					}
				} else {
					fmt.Printf("Missing key 'value' in output response\n")
				}

			}
		}

		acc.AddFields("jolokia", fields, tags)
	}

	return nil
}

func init() {
	inputs.Add("jolokia", func() telegraf.Input {
		return &Jolokia{}
	})
}
