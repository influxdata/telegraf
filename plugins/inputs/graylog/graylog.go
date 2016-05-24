package graylog

import (
	"errors"
	"fmt"
	"strconv"
	"bytes"
	"regexp"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
	"encoding/json"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"
)

type GrayLog struct {
	Name       string
	Servers    []string
	TagKeys    []string
	Metrics	   []string
	Headers    map[string]string

	// Path to CA file
	SSLCA string `toml:"ssl_ca"`
	// Path to host cert file
	SSLCert string `toml:"ssl_cert"`
	// Path to cert key file
	SSLKey string `toml:"ssl_key"`
	// Use SSL but skip chain & host verification
	InsecureSkipVerify bool

	client HTTPClient
}

type HTTPClient interface {
	// Returns the result of an http request
	//
	// Parameters:
	// req: HTTP request object
	//
	// Returns:
	// http.Response:  HTTP respons object
	// error        :  Any error that may have occurred
	MakeRequest(req *http.Request) (*http.Response, error)

	SetHTTPClient(client *http.Client)
	HTTPClient() *http.Client
}

type Messagebody struct {
    Metrics  []string `json:"metrics"`
}

type RealHTTPClient struct {
	client *http.Client
}

func (c *RealHTTPClient) MakeRequest(req *http.Request) (*http.Response, error) {
	return c.client.Do(req)
}

func (c *RealHTTPClient) SetHTTPClient(client *http.Client) {
	c.client = client
}

func (c *RealHTTPClient) HTTPClient() *http.Client {
	return c.client
}

var sampleConfig = `
  ## NOTE This plugin only reads numerical measurements, strings and booleans
  ## will be ignored.

  ## a name for the service being polled
  name = "webserver_stats"

  ## URL of each server in the service's cluster
  servers = [
    "http://localhost:9999/stats/",
    "http://localhost:9998/stats/",
  ]


  ## List of tag names to extract from top-level of JSON server response
  # tag_keys = [
  #   "my_tag_1",
  #   "my_tag_2"
  # ]
  metrics = [
    "jvm.cl.loaded",
    "jvm.memory.pools.Metaspace.committed"
  ]

  ## HTTP Header parameters (all values must be strings)
  # [inputs.graylog.headers]
  #   X-Auth-Token = "my-xauth-token"
  #   apiVersion = "v1"

  ## Optional SSL Config
  # ssl_ca = "/etc/telegraf/ca.pem"
  # ssl_cert = "/etc/telegraf/cert.pem"
  # ssl_key = "/etc/telegraf/key.pem"
  ## Use SSL but skip chain & host verification
  # insecure_skip_verify = false
`

func (h *GrayLog) SampleConfig() string {
	return sampleConfig
}

func (h *GrayLog) Description() string {
	return "Read flattened metrics from one or more JSON HTTP endpoints"
}

// Gathers data for all servers.
func (h *GrayLog) Gather(acc telegraf.Accumulator) error {
	var wg sync.WaitGroup

	if h.client.HTTPClient() == nil {
		tlsCfg, err := internal.GetTLSConfig(
			h.SSLCert, h.SSLKey, h.SSLCA, h.InsecureSkipVerify)
		if err != nil {
			return err
		}
		tr := &http.Transport{
			ResponseHeaderTimeout: time.Duration(3 * time.Second),
			TLSClientConfig:       tlsCfg,
		}
		client := &http.Client{
			Transport: tr,
			Timeout:   time.Duration(4 * time.Second),
		}
		h.client.SetHTTPClient(client)
	}

	errorChannel := make(chan error, len(h.Servers))

	for _, server := range h.Servers {
		wg.Add(1)
		go func(server string) {
			defer wg.Done()
			if err := h.gatherServer(acc, server); err != nil {
				errorChannel <- err
			}
		}(server)
	}

	wg.Wait()
	close(errorChannel)

	// Get all errors and return them as one giant error
	errorStrings := []string{}
	for err := range errorChannel {
		errorStrings = append(errorStrings, err.Error())
	}

	if len(errorStrings) == 0 {
		return nil
	}
	return errors.New(strings.Join(errorStrings, "\n"))
}

// Gathers data from a particular server
// Parameters:
//     acc      : The telegraf Accumulator to use
//     serverURL: endpoint to send request to
//     service  : the service being queried
//
// Returns:
//     error: Any error that may have occurred
func (h *GrayLog) gatherServer(
	acc telegraf.Accumulator,
	serverURL string,
) error {
	resp, responseTime, err := h.sendRequest(serverURL)
	var dat map[string]interface{}
	var name_list []string

	if err != nil {
		return err
	}

	var msrmnt_name string
	if h.Name == "" {
		msrmnt_name = "graylog"
	} else {
		msrmnt_name = "graylog_" + h.Name
	}
	tags := map[string]string{
		"server": serverURL,
	}

	parser, err := parsers.NewJSONParser(msrmnt_name, h.TagKeys, tags)
	if err != nil {
		return err
	}
	if err := json.Unmarshal([]byte(resp), &dat); err != nil {
		panic(err)
	}
	if rec, ok := dat["metrics"].([]interface{}); ok {
		for _, metric := range rec {
			if m, ok := metric.(map[string]interface{}); ok {
				str, _ := m["full_name"].(string)
				name_list = append(name_list, str)
			}
		}
	}
	metrics, err := parser.Parse([]byte(resp))
	if err != nil {
		return err
	}

	for _, metric := range metrics {
		fields := make(map[string]interface{})
		for k, v := range metric.Fields() {
			re, _ := regexp.Compile(`metrics_([0-9]+)`)
			match := re.FindAllStringSubmatch(k, -1)
			if(match != nil) {
				i, _ := strconv.Atoi(match[0][1])
				fields[name_list[i]] = v
			} 
		}
		fields["response_time"] = responseTime
		acc.AddFields(metric.Name(), fields, metric.Tags())
	}
	return nil
}

// Sends an HTTP request to the server using the GrayLog object's HTTPClient.
// Parameters:
//     serverURL: endpoint to send request to
//
// Returns:
//     string: body of the response
//     error : Any error that may have occurred
func (h *GrayLog) sendRequest(serverURL string) (string, float64, error) {
	// Prepare URL
	requestURL, err := url.Parse(serverURL)
	if err != nil {
		return "", -1, fmt.Errorf("Invalid server URL \"%s\"", serverURL)
	}
	m := &Messagebody{Metrics: h.Metrics}
	http_body, err := json.Marshal(m)
	if err != nil {
		return "", -1, fmt.Errorf("Invalid list of Metrics %s", h.Metrics)
	}
	req, err := http.NewRequest("POST", requestURL.String(), bytes.NewBuffer(http_body))
	if err != nil {
		return "", -1, err
	}

	// Add header parameters
	for k, v := range h.Headers {
		if strings.ToLower(k) == "host" {
			req.Host = v
		} else {
			req.Header.Add(k, v)
		}
	}

	start := time.Now()
	resp, err := h.client.MakeRequest(req)
	if err != nil {
		return "", -1, err
	}

	defer resp.Body.Close()
	responseTime := time.Since(start).Seconds()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return string(body), responseTime, err
	}

	// Process response
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("Response from url \"%s\" has status code %d (%s), expected %d (%s)",
			requestURL.String(),
			resp.StatusCode,
			http.StatusText(resp.StatusCode),
			http.StatusOK,
			http.StatusText(http.StatusOK))
		return string(body), responseTime, err
	}
	return string(body), responseTime, err
}

func init() {
	inputs.Add("graylog", func() telegraf.Input {
		return &GrayLog{
			client: &RealHTTPClient{},
		}
	})
}
