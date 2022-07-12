package t128_graphql

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/Jeffail/gabs"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const (
	//DefaultRequestTimeout is the request timeout if none is configured
	DefaultRequestTimeout = 5 * time.Second

	//DefaultDeadline is the time for the graphQL server to build the response. Default is no deadline.
	DefaultDeadline = 0 * time.Second

	//MinTimeoutDeadlineDiff is the minimum value of plugin.Timeout - plugin.Deadline
	MinTimeoutDeadlineDiff = 1 * time.Second
)

//T128GraphQL is an input for metrics of a 128T router instance
type T128GraphQL struct {
	CollectorName   string            `toml:"collector_name"`
	BaseURL         string            `toml:"base_url"`
	UnixSocket      string            `toml:"unix_socket"`
	EntryPoint      string            `toml:"entry_point"`
	Fields          map[string]string `toml:"extract_fields"`
	Tags            map[string]string `toml:"extract_tags"`
	Timeout         config.Duration   `toml:"timeout"`
	RetryIfNotFound bool              `toml:"retry_if_not_found"`
	Deadline        config.Duration   `toml:"deadline"`

	Config           *Config
	Query            string
	requestBody      []byte
	client           *http.Client
	endpointNotFound bool
}

//SampleConfig returns the default configuration of the Input
func (*T128GraphQL) SampleConfig() string {
	return sampleConfig
}

//Description returns a one-sentence description on the Input
func (*T128GraphQL) Description() string {
	return "Make a 128T GraphQL query and return the data"
}

//Init sets up the input to be ready for action
func (plugin *T128GraphQL) Init() error {
	//check and load config
	err := plugin.checkConfig()
	if err != nil {
		return err
	}

	fieldsWithRelPath, fieldsWithAbsPath, err := validateAndSeparatePaths(plugin.Fields, plugin.EntryPoint)
	if err != nil {
		return err
	}

	tagsWithRelPath, tagsWithAbsPath, err := validateAndSeparatePaths(plugin.Tags, plugin.EntryPoint)
	if err != nil {
		return err
	}

	plugin.Config = LoadConfig(
		plugin.EntryPoint,
		fieldsWithRelPath,
		fieldsWithAbsPath,
		tagsWithRelPath,
		tagsWithAbsPath,
	)

	//build query, json path to data and request body
	plugin.Query = BuildQuery(plugin.Config)

	content := struct {
		Query string `json:"query,omitempty"`
	}{
		plugin.Query,
	}

	body, err := json.Marshal(content)
	if err != nil {
		return fmt.Errorf("failed to create request body for query '%s': %w", plugin.Query, err)
	}
	plugin.requestBody = body

	//setup client
	transport := http.DefaultTransport

	if plugin.UnixSocket != "" {
		transport = &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", plugin.UnixSocket)
			},
		}
	}

	plugin.client = &http.Client{Transport: transport, Timeout: time.Duration(plugin.Timeout)}

	plugin.endpointNotFound = false

	return nil
}

func (plugin *T128GraphQL) checkConfig() error {
	if plugin.CollectorName == "" {
		return fmt.Errorf("collector_name is a required configuration field")
	}

	if plugin.BaseURL == "" {
		return fmt.Errorf("base_url is a required configuration field")
	}

	if !strings.HasSuffix(plugin.BaseURL, "/") {
		plugin.BaseURL += "/"
	}

	if plugin.EntryPoint == "" {
		return fmt.Errorf("entry_point is a required configuration field")
	}

	if plugin.Fields == nil {
		return fmt.Errorf("extract_fields is a required configuration field")
	}

	if time.Duration(plugin.Deadline) != DefaultDeadline {
		timeoutDeadlineDiff := time.Duration(plugin.Timeout).Seconds() - time.Duration(plugin.Deadline).Seconds()
		if timeoutDeadlineDiff < MinTimeoutDeadlineDiff.Seconds() {
			return fmt.Errorf(
				"timeout must be at least %d seconds greater than deadline: currently %d seconds greater",
				int(MinTimeoutDeadlineDiff.Seconds()),
				int(timeoutDeadlineDiff),
			)
		}
	}

	return nil
}

//Gather takes in an accumulator and adds the metrics that the Input gathers
func (plugin *T128GraphQL) Gather(acc telegraf.Accumulator) error {
	if !plugin.RetryIfNotFound && plugin.endpointNotFound {
		return nil
	}

	request, err := plugin.createRequest()
	if err != nil {
		acc.AddError(fmt.Errorf("failed to create a request for query %s: %w", plugin.Query, err))
		return nil
	}

	response, err := plugin.client.Do(request)
	if err != nil {
		acc.AddError(fmt.Errorf("failed to make graphQL request for collector %s: %w", plugin.CollectorName, err))
		return nil
	}
	defer response.Body.Close()

	message, err := ioutil.ReadAll(response.Body)
	if err != nil {
		message = []byte("")
	}

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		template := fmt.Sprintf("status code %d not OK for collector ", response.StatusCode) + plugin.CollectorName + ": %s"
		for _, err := range decodeAndReportJSONErrors(message, template) {
			acc.AddError(err)
		}
		return nil
	}

	//decode json
	jsonParsed, err := gabs.ParseJSON(message)
	if err != nil {
		acc.AddError(fmt.Errorf("invalid json response for collector %s: %w", plugin.CollectorName, err))
		return nil
	}

	//look for other errors in response
	exists := jsonParsed.Exists("errors")
	if exists {
		template := fmt.Sprintf("found errors in response for collector %s", plugin.CollectorName) + ": %s"
		for _, err := range decodeAndReportJSONErrors(message, template) {
			acc.AddError(err)

			if strings.Contains(err.Error(), "returned a 404") {
				plugin.endpointNotFound = true

				if !plugin.RetryIfNotFound {
					acc.AddError(errors.New("collector configured to not retry when endpoint not found (404), stopping queries"))
				}
			}
		}
	}

	//look for empty response
	dataExists := jsonParsed.Exists("data")
	if !dataExists {
		acc.AddError(fmt.Errorf("no data found in response for collector %s", plugin.CollectorName))
		return nil
	}

	processedResponses, err := ProcessResponse(jsonParsed, plugin.CollectorName, plugin.Config.Fields, plugin.Config.Tags)
	if err != nil {
		acc.AddError(err)
		return nil
	}

	for _, processedResponse := range processedResponses {
		acc.AddFields(
			plugin.CollectorName,
			processedResponse.Fields,
			processedResponse.Tags,
		)
	}
	return nil
}

func (plugin *T128GraphQL) createRequest() (*http.Request, error) {
	request, err := http.NewRequest("POST", plugin.BaseURL, bytes.NewReader(plugin.requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request for query '%s': %w", plugin.Query, err)
	}

	request.Header.Add("Content-Type", "application/json")

	if time.Duration(plugin.Deadline) != DefaultDeadline {
		deadline := int(time.Duration(plugin.Deadline).Truncate(time.Second))
		request.Header.Add("deadline", fmt.Sprintf("%d", deadline))
	}

	return request, nil
}

func decodeAndReportJSONErrors(response []byte, template string) []error {
	var errors []error

	parsedJSON, err := gabs.ParseJSON(response)
	if err != nil {
		errors = append(errors, fmt.Errorf(template, response))
		return errors
	}

	jsonObj, err := parsedJSON.JSONPointer("/errors")
	if err != nil {
		errors = append(errors, fmt.Errorf(template, parsedJSON.String()))
		return errors
	}

	jsonChildren, err := jsonObj.Children()
	if err != nil {
		errors = append(errors, fmt.Errorf(template, parsedJSON.String()))
		return errors
	}

	for _, child := range jsonChildren {
		errorNode := child.Data().(map[string]interface{})
		message := fmt.Sprintf("%v", errorNode["message"])
		errors = append(errors, fmt.Errorf(template, message))
	}
	return errors
}

func validateAndSeparatePaths(data map[string]string, entryPoint string) (map[string]string, map[string]string, error) {
	predicateRegex := regexp.MustCompile(`\(.*?\)`)
	cleanEntryPoint := predicateRegex.ReplaceAllString(entryPoint, "")
	dataWithRelPath := make(map[string]string)
	dataWithAbsPath := make(map[string]string)

	for name, path := range data {
		if predicateRegex.MatchString(path) {
			return nil, nil, fmt.Errorf("absolute path %s on tag can not contain graphQL arguments", path)
		}

		leafIndex := strings.LastIndex(path, "/")
		pathToLeaf := path
		if leafIndex != -1 {
			pathToLeaf = pathToLeaf[:leafIndex]
		}

		if !strings.HasPrefix(cleanEntryPoint, pathToLeaf) {
			dataWithRelPath[name] = path
			continue
		}

		dataWithAbsPath[name] = path
	}

	return dataWithRelPath, dataWithAbsPath, nil
}

func init() {
	inputs.Add("t128_graphql", func() telegraf.Input {
		return &T128GraphQL{
			Timeout:  config.Duration(DefaultRequestTimeout),
			Deadline: config.Duration(DefaultDeadline),
		}
	})
}
