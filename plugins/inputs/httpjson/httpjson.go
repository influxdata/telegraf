package httpjson

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type HttpJson struct {
	Name       string
	Servers    []string
	Method     string
	TagKeys    []string
	Parameters map[string]string
	Headers    map[string]string
	client     HTTPClient
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
}

type RealHTTPClient struct {
	client *http.Client
}

func (c RealHTTPClient) MakeRequest(req *http.Request) (*http.Response, error) {
	return c.client.Do(req)
}

var sampleConfig = `
  # NOTE This plugin only reads numerical measurements, strings and booleans
  # will be ignored.

  # a name for the service being polled
  name = "webserver_stats"

  # URL of each server in the service's cluster
  servers = [
    "http://localhost:9999/stats/",
    "http://localhost:9998/stats/",
  ]

  # HTTP method to use (case-sensitive)
  method = "GET"

  # List of tag names to extract from top-level of JSON server response
  # tag_keys = [
  #   "my_tag_1",
  #   "my_tag_2"
  # ]

  # HTTP parameters (all values must be strings)
  [inputs.httpjson.parameters]
    event_type = "cpu_spike"
    threshold = "0.75"

  # HTTP Header parameters (all values must be strings)
  # [inputs.httpjson.headers]
  #   X-Auth-Token = "my-xauth-token"
  #   apiVersion = "v1"

`

func (h *HttpJson) SampleConfig() string {
	return sampleConfig
}

func (h *HttpJson) Description() string {
	return "Read flattened metrics from one or more JSON HTTP endpoints"
}

// Gathers data for all servers.
func (h *HttpJson) Gather(acc telegraf.Accumulator) error {
	var wg sync.WaitGroup

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
func (h *HttpJson) gatherServer(
	acc telegraf.Accumulator,
	serverURL string,
) error {
	resp, responseTime, err := h.sendRequest(serverURL)

	if err != nil {
		return err
	}

	var jsonOut map[string]interface{}
	if err = json.Unmarshal([]byte(resp), &jsonOut); err != nil {
		return errors.New("Error decoding JSON response")
	}

	tags := map[string]string{
		"server": serverURL,
	}

	for _, tag := range h.TagKeys {
		switch v := jsonOut[tag].(type) {
		case string:
			tags[tag] = v
		}
		delete(jsonOut, tag)
	}

	if responseTime >= 0 {
		jsonOut["response_time"] = responseTime
	}
	f := internal.JSONFlattener{}
	err = f.FlattenJSON("", jsonOut)
	if err != nil {
		return err
	}

	var msrmnt_name string
	if h.Name == "" {
		msrmnt_name = "httpjson"
	} else {
		msrmnt_name = "httpjson_" + h.Name
	}
	acc.AddFields(msrmnt_name, f.Fields, tags)
	return nil
}

// Sends an HTTP request to the server using the HttpJson object's HTTPClient
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

	params := url.Values{}
	for k, v := range h.Parameters {
		params.Add(k, v)
	}
	requestURL.RawQuery = params.Encode()

	// Create + send request
	req, err := http.NewRequest(h.Method, requestURL.String(), nil)
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
	inputs.Add("httpjson", func() telegraf.Input {
		return &HttpJson{client: RealHTTPClient{client: &http.Client{}}}
	})
}
