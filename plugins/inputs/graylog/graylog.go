package graylog

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type ResponseMetrics struct {
	Metrics []Metric `json:"metrics"`
}

type Metric struct {
	FullName string                 `json:"full_name"`
	Name     string                 `json:"name"`
	Type     string                 `json:"type"`
	Fields   map[string]interface{} `json:"metric"`
}

type GrayLog struct {
	Servers  []string        `toml:"servers"`
	Metrics  []string        `toml:"metrics"`
	Username string          `toml:"username"`
	Password string          `toml:"password"`
	Timeout  config.Duration `toml:"timeout"`

	tls.ClientConfig
	client HTTPClient
}

type HTTPClient interface {
	// Returns the result of an http request
	//
	// Parameters:
	// req: HTTP request object
	//
	// Returns:
	// http.Response:  HTTP response object
	// error        :  Any error that may have occurred
	MakeRequest(req *http.Request) (*http.Response, error)

	SetHTTPClient(client *http.Client)
	HTTPClient() *http.Client
}

type Messagebody struct {
	Metrics []string `json:"metrics"`
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

// Gathers data for all servers.
func (h *GrayLog) Gather(acc telegraf.Accumulator) error {
	var wg sync.WaitGroup

	if h.client.HTTPClient() == nil {
		tlsCfg, err := h.ClientConfig.TLSConfig()
		if err != nil {
			return err
		}
		tr := &http.Transport{
			ResponseHeaderTimeout: time.Duration(h.Timeout),
			TLSClientConfig:       tlsCfg,
		}
		client := &http.Client{
			Transport: tr,
			Timeout:   time.Duration(h.Timeout),
		}
		h.client.SetHTTPClient(client)
	}

	for _, server := range h.Servers {
		wg.Add(1)
		go func(server string) {
			defer wg.Done()
			acc.AddError(h.gatherServer(acc, server))
		}(server)
	}

	wg.Wait()

	return nil
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
	resp, _, err := h.sendRequest(serverURL)
	if err != nil {
		return err
	}
	requestURL, err := url.Parse(serverURL)
	if err != nil {
		return fmt.Errorf("unable to parse address '%s': %s", serverURL, err)
	}

	host, port, _ := net.SplitHostPort(requestURL.Host)
	var dat ResponseMetrics
	if err := json.Unmarshal([]byte(resp), &dat); err != nil {
		return err
	}
	for _, mItem := range dat.Metrics {
		fields := make(map[string]interface{})
		tags := map[string]string{
			"server": host,
			"port":   port,
			"name":   mItem.Name,
			"type":   mItem.Type,
		}
		h.flatten(mItem.Fields, fields, "")
		acc.AddFields(mItem.FullName, fields, tags)
	}
	return nil
}

// Flatten JSON hierarchy to produce field name and field value
// Parameters:
//    item: Item map to flatten
//    fields: Map to store generated fields.
//    id: Prefix for top level metric (empty string "")
// Returns:
//    void
func (h *GrayLog) flatten(item map[string]interface{}, fields map[string]interface{}, id string) {
	if id != "" {
		id = id + "_"
	}
	for k, i := range item {
		switch i := i.(type) {
		case int:
			fields[id+k] = float64(i)
		case float64:
			fields[id+k] = i
		case map[string]interface{}:
			h.flatten(i, fields, id+k)
		default:
		}
	}
}

// Sends an HTTP request to the server using the GrayLog object's HTTPClient.
// Parameters:
//     serverURL: endpoint to send request to
//
// Returns:
//     string: body of the response
//     error : Any error that may have occurred
func (h *GrayLog) sendRequest(serverURL string) (string, float64, error) {
	headers := map[string]string{
		"Content-Type": "application/json",
		"Accept":       "application/json",
	}
	method := "GET"
	content := bytes.NewBufferString("")
	headers["Authorization"] = "Basic " + base64.URLEncoding.EncodeToString([]byte(h.Username+":"+h.Password))
	// Prepare URL
	requestURL, err := url.Parse(serverURL)
	if err != nil {
		return "", -1, fmt.Errorf("invalid server URL \"%s\"", serverURL)
	}
	// Add X-Requested-By header
	headers["X-Requested-By"] = "Telegraf"

	if strings.Contains(requestURL.String(), "multiple") {
		m := &Messagebody{Metrics: h.Metrics}
		httpBody, err := json.Marshal(m)
		if err != nil {
			return "", -1, fmt.Errorf("invalid list of Metrics %s", h.Metrics)
		}
		method = "POST"
		content = bytes.NewBuffer(httpBody)
	}
	req, err := http.NewRequest(method, requestURL.String(), content)
	if err != nil {
		return "", -1, err
	}
	// Add header parameters
	for k, v := range headers {
		req.Header.Add(k, v)
	}
	start := time.Now()
	resp, err := h.client.MakeRequest(req)
	if err != nil {
		return "", -1, err
	}

	defer resp.Body.Close()
	responseTime := time.Since(start).Seconds()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return string(body), responseTime, err
	}

	// Process response
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("response from url \"%s\" has status code %d (%s), expected %d (%s)",
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
			client:  &RealHTTPClient{},
			Timeout: config.Duration(5 * time.Second),
		}
	})
}
