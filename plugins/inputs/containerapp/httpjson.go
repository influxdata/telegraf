package containerapp

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/internal/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/tidwall/gjson"
)

var (
	utf8BOM  = []byte("\xef\xbb\xbf")
	maxponts = 100
)

// HttpJson struct
type HttpJson struct {
	Name            string
	Servers         []string
	Method          string
	TagKeys         []string
	ResponseTimeout internal.Duration
	Parameters      map[string]string
	Headers         map[string]string
	tls.ClientConfig
	transport *http.Transport
	client    HTTPClient
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

func (h *HttpJson) SampleConfig() string {
	return sampleConfig
}

func (h *HttpJson) Description() string {
	return "Read flattened metrics from one or more JSON HTTP endpoints"
}

// Gathers data for all servers.
func (h *HttpJson) Gather(acc telegraf.Accumulator) error {
	var wg sync.WaitGroup

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

type Parser struct {
	Method     string
	Parameters map[string]string
	Headers    map[string]string
	fields     map[string]interface{}
	tags       map[string]string
}

func (p *Parser) flatten(name string, i interface{}, tagKeys []string) {
	switch v := i.(type) {
	case bool:
		for _, tagName := range tagKeys {
			if name == tagName {
				p.tags[name] = strconv.FormatBool(v)
			}
		}
	case string:
		for _, tagName := range tagKeys {
			if name == tagName {
				p.tags[name] = v
			}
		}
	case float64:
		for _, tagName := range tagKeys {
			if name == tagName {
				p.tags[name] = strconv.FormatFloat(v, 'f', -1, 64)
				continue
			}
		}
		p.fields[name] = v

	case []interface{}:
		for k, vv := range v {
			prefix := ""
			if name != "" {
				prefix = name + "_"
			}
			p.flatten(prefix+strconv.Itoa(k), vv, tagKeys)
		}
	case map[string]interface{}:
		for k, vv := range v {
			prefix := ""
			if name != "" {
				prefix = name + "_"
			}
			p.flatten(prefix+k, vv, tagKeys)
		}
	}
}

func (p *Parser) flush(metricName string, acc telegraf.Accumulator, responseTime float64) {
	p.fields["response_time"] = responseTime
	count := 0
	sfields := map[string]interface{}{}
	for k, v := range p.fields {
		count++
		sfields[k] = v
		if count > maxponts {
			acc.AddFields(metricName, sfields, p.tags)
			count = 0
			sfields = map[string]interface{}{}
		}

	}
	if count != 0 {
		acc.AddFields(metricName, sfields, p.tags)
	}
}

func (p *Parser) Parse(
	metricName string,
	tagKeys []string,
	serverURL string,
	acc telegraf.Accumulator,
	resp string,
	responseTime float64,
) error {

	if !gjson.Valid(resp) {
		return errors.New("invalid json")
	}

	p.fields = make(map[string]interface{})
	p.tags = map[string]string{
		"server": serverURL,
	}

	pobj := gjson.Parse(resp)
	if pobj.IsArray() {
		for _, val := range pobj.Value().([]interface{}) {
			p.flatten("", val, tagKeys)
			p.flush(metricName, acc, responseTime)
		}
	} else {
		p.flatten("", pobj.Value(), tagKeys)
		p.flush(metricName, acc, responseTime)
	}

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
func (h *HttpJson) gatherServer(
	acc telegraf.Accumulator,
	serverURL string,
) error {

	var msrmnt_name string
	if h.Name == "" {
		msrmnt_name = "httpjson"
	} else {
		msrmnt_name = h.Name
	}

	p := Parser{
		Method:     h.Method,
		Parameters: h.Parameters,
		Headers:    h.Headers,
	}

	if h.client.HTTPClient() == nil {
		tlsCfg, err := h.ClientConfig.TLSConfig()
		if err != nil {
			return err
		}
		tr := &http.Transport{
			ResponseHeaderTimeout: h.ResponseTimeout.Duration,
			TLSClientConfig:       tlsCfg,
		}
		client := &http.Client{
			Transport: tr,
			Timeout:   h.ResponseTimeout.Duration,
		}
		h.client.SetHTTPClient(client)
	}

	resp, responseTime, err := h.sendRequest(serverURL)
	if err != nil {
		return err
	}

	resp = strings.Trim(resp, " ")
	if len(resp) == 0 {
		return nil
	}

	err = p.Parse(msrmnt_name, h.TagKeys, serverURL, acc, resp, responseTime)
	if err != nil {
		return err
	}
	return nil
}

// Sends an HTTP request to the server using the HttpJson object's HTTPClient.
// This request can be either a GET or a POST.
// Parameters:
//     serverURL: endpoint to send request to
//
// Returns:
//     string: body of the response
//     error : Any error that may have occurred
func (h *HttpJson) sendRequest(serverURL string) (string, float64, error) {
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
		log.Printf("E! Create request error to %s", serverURL)
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
	ctx, cancel := context.WithTimeout(req.Context(), 5*time.Second)
	defer cancel()
	req = req.WithContext(ctx)

	start := time.Now()

	resp, err := h.client.MakeRequest(req)
	if err != nil {
		log.Printf("E! Send request error to %s", serverURL)
		return "", -1, err
	}

	defer func() {
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
	}()
	responseTime := time.Since(start).Seconds()

	body, err := ioutil.ReadAll(resp.Body)
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

func NewHttpJson() *HttpJson {
	responseTimeout := internal.Duration{
		Duration: 5 * time.Second,
	}
	return &HttpJson{
		client: &RealHTTPClient{},
		transport: &http.Transport{
			ResponseHeaderTimeout: responseTimeout.Duration,
			DisableKeepAlives:     true,
		},
		ResponseTimeout: responseTimeout,
	}
}

func init() {
	inputs.Add("httpjson", func() telegraf.Input {
		return NewHttpJson()
	})
}
