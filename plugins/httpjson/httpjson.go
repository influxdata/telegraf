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

	"github.com/influxdb/telegraf/plugins"
)

type HttpJson struct {
	Services []Service
	client   HTTPClient
}

type Service struct {
	Name       string
	Servers    []string
	Method     string
	TagKeys    []string
	Parameters map[string]string
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
  # Specify services via an array of tables
  [[httpjson.services]]

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
    # 	"my_tag_1",
    # 	"my_tag_2"
    # ]

    # HTTP parameters (all values must be strings)
    [httpjson.services.parameters]
      event_type = "cpu_spike"
      threshold = "0.75"
`

func (h *HttpJson) SampleConfig() string {
	return sampleConfig
}

func (h *HttpJson) Description() string {
	return "Read flattened metrics from one or more JSON HTTP endpoints"
}

// Gathers data for all servers.
func (h *HttpJson) Gather(acc plugins.Accumulator) error {
	var wg sync.WaitGroup

	totalServers := 0
	for _, service := range h.Services {
		totalServers += len(service.Servers)
	}
	errorChannel := make(chan error, totalServers)

	for _, service := range h.Services {
		for _, server := range service.Servers {
			wg.Add(1)
			go func(service Service, server string) {
				defer wg.Done()
				if err := h.gatherServer(acc, service, server); err != nil {
					errorChannel <- err
				}
			}(service, server)
		}
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
	acc plugins.Accumulator,
	service Service,
	serverURL string,
) error {
	resp, err := h.sendRequest(service, serverURL)
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

	for _, tag := range service.TagKeys {
		switch v := jsonOut[tag].(type) {
		case string:
			tags[tag] = v
		}
		delete(jsonOut, tag)
	}

	processResponse(acc, service.Name, tags, jsonOut)
	return nil
}

// Sends an HTTP request to the server using the HttpJson object's HTTPClient
// Parameters:
//     serverURL: endpoint to send request to
//
// Returns:
//     string: body of the response
//     error : Any error that may have occurred
func (h *HttpJson) sendRequest(service Service, serverURL string) (string, error) {
	// Prepare URL
	requestURL, err := url.Parse(serverURL)
	if err != nil {
		return "", fmt.Errorf("Invalid server URL \"%s\"", serverURL)
	}

	params := url.Values{}
	for k, v := range service.Parameters {
		params.Add(k, v)
	}
	requestURL.RawQuery = params.Encode()

	// Create + send request
	req, err := http.NewRequest(service.Method, requestURL.String(), nil)
	if err != nil {
		return "", err
	}

	resp, err := h.client.MakeRequest(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return string(body), err
	}

	// Process response
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("Response from url \"%s\" has status code %d (%s), expected %d (%s)",
			requestURL.String(),
			resp.StatusCode,
			http.StatusText(resp.StatusCode),
			http.StatusOK,
			http.StatusText(http.StatusOK))
		return string(body), err
	}

	return string(body), err
}

// Flattens the map generated from the JSON object and stores its float values using a
// plugins.Accumulator. It ignores any non-float values.
// Parameters:
//     acc: the Accumulator to use
//     prefix: What the name of the measurement name should be prefixed by.
//     tags: telegraf tags to
func processResponse(acc plugins.Accumulator, prefix string, tags map[string]string, v interface{}) {
	switch t := v.(type) {
	case map[string]interface{}:
		for k, v := range t {
			processResponse(acc, prefix+"_"+k, tags, v)
		}
	case float64:
		acc.Add(prefix, v, tags)
	}
}

func init() {
	plugins.Add("httpjson", func() plugins.Plugin {
		return &HttpJson{client: RealHTTPClient{client: &http.Client{}}}
	})
}
