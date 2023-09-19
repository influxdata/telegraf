//go:generate ../../../tools/readme_config_includer/generator
package exchange

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

const (
	currencyApiEndpoint = "https://api.freecurrencyapi.com"
	currencyApiResource = "/v1/latest"
)

type Exchange struct {
	APIKey         string `toml:"apikey"`
	BaseCurrency   string `toml:"base_currency"`
	TargetCurrency string `toml:"target_currency"`
}

func (*Exchange) SampleConfig() string {
	return sampleConfig
}

// Init is for setup, and validating config.
func (e *Exchange) Init() error {
	// We cannot access api without token
	if e.APIKey == "" {
		return fmt.Errorf("'api_token' cannot be blank")
	}

	return nil
}

func (e *Exchange) Gather(acc telegraf.Accumulator) error {
	res, err := e.makeApiRequest()
	if err != nil {
		acc.AddError(fmt.Errorf("%w", err))
		return err
	}

	result, err := parseResponse(res)
	if err != nil {
		acc.AddError(fmt.Errorf("%w", err))
		return err
	}

	acc.AddFields("state", map[string]interface{}{"value": result}, nil)

	return nil
}

func init() {
	inputs.Add("exchange", func() telegraf.Input { return &Exchange{} })
}

// ApiResponse struct that stores info about currency rate.
type ApiResponse struct {
	Data map[string]float32 `json:data`
}

// makeApiRequest makes request and return response.
func (e *Exchange) makeApiRequest() (*http.Response, error) {
	// Prepare http client
	client, req, err := e.prepareHttpClient()

	if err != nil {
		return nil, err
	}

	// Make request
	res, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (e *Exchange) prepareHttpClient() (*http.Client, *http.Request, error) {
	// Create HTTP client
	client := &http.Client{}

	// Prepare request params
	params := url.Values{}
	params.Add("base_currency", e.BaseCurrency)
	params.Add("currencies", e.TargetCurrency)
	params.Add("apikey", e.APIKey)

	// Prepare uri
	u, _ := url.ParseRequestURI(currencyApiEndpoint)
	u.Path = currencyApiResource
	u.RawQuery = params.Encode()

	// "http://example.com/path?param1=value1&param2=value2"
	urlStr := fmt.Sprintf("%v", u)
	req, err := http.NewRequest("GET", urlStr, nil)

	if err != nil {
		return nil, nil, err
	}

	return client, req, nil
}

// parseResponse parses response to the endpoint and returns ApiResponse struct.
func parseResponse(res *http.Response) (ApiResponse, error) {
	result := ApiResponse{}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return result, err
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return result, err
	}

	return result, nil
}
