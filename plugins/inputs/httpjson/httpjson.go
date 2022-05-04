package httpjson

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"
)

var (
	utf8BOM = []byte("\xef\xbb\xbf")
)

// HTTPJSON struct
type HTTPJSON struct {
	Name            string `toml:"name" deprecated:"1.3.0;use 'name_override', 'name_suffix', 'name_prefix' instead"`
	Servers         []string
	Method          string
	TagKeys         []string
	ResponseTimeout config.Duration
	Parameters      map[string]string
	Headers         map[string]string
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
func (h *HTTPJSON) Gather(acc telegraf.Accumulator) error {
	var wg sync.WaitGroup

	if h.client.HTTPClient() == nil {
		tlsCfg, err := h.ClientConfig.TLSConfig()
		if err != nil {
			return err
		}
		tr := &http.Transport{
			ResponseHeaderTimeout: time.Duration(h.ResponseTimeout),
			TLSClientConfig:       tlsCfg,
		}
		client := &http.Client{
			Transport: tr,
			Timeout:   time.Duration(h.ResponseTimeout),
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
func (h *HTTPJSON) gatherServer(
	acc telegraf.Accumulator,
	serverURL string,
) error {
	resp, responseTime, err := h.sendRequest(serverURL)
	if err != nil {
		return err
	}

	var msrmntName string
	if h.Name == "" {
		msrmntName = "httpjson"
	} else {
		msrmntName = "httpjson_" + h.Name
	}
	tags := map[string]string{
		"server": serverURL,
	}

	parser, err := parsers.NewParser(&parsers.Config{
		DataFormat:  "json",
		MetricName:  msrmntName,
		TagKeys:     h.TagKeys,
		DefaultTags: tags,
	})
	if err != nil {
		return err
	}

	metrics, err := parser.Parse([]byte(resp))
	if err != nil {
		return err
	}

	for _, metric := range metrics {
		fields := make(map[string]interface{})
		for k, v := range metric.Fields() {
			fields[k] = v
		}
		fields["response_time"] = responseTime
		acc.AddFields(metric.Name(), fields, metric.Tags())
	}
	return nil
}

// Sends an HTTP request to the server using the HTTPJSON object's HTTPClient.
// This request can be either a GET or a POST.
// Parameters:
//     serverURL: endpoint to send request to
//
// Returns:
//     string: body of the response
//     error : Any error that may have occurred
func (h *HTTPJSON) sendRequest(serverURL string) (string, float64, error) {
	// Prepare URL
	requestURL, err := url.Parse(serverURL)
	if err != nil {
		return "", -1, fmt.Errorf("Invalid server URL \"%s\"", serverURL)
	}

	data := url.Values{}
	switch {
	case h.Method == "GET":
		params := requestURL.Query()
		for k, v := range h.Parameters {
			params.Add(k, v)
		}
		requestURL.RawQuery = params.Encode()

	case h.Method == "POST":
		requestURL.RawQuery = ""
		for k, v := range h.Parameters {
			data.Add(k, v)
		}
	}

	// Create + send request
	req, err := http.NewRequest(h.Method, requestURL.String(),
		strings.NewReader(data.Encode()))
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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return string(body), responseTime, err
	}
	body = bytes.TrimPrefix(body, utf8BOM)

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
	inputs.Add("httpjson", func() telegraf.Input {
		return &HTTPJSON{
			client:          &RealHTTPClient{},
			ResponseTimeout: config.Duration(5 * time.Second),
		}
	})
}
