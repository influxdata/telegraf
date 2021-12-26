package azure_monitor

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type AzureMonitor struct {
	SubscriptionID       string                 `toml:"subscription_id"`
	ClientID             string                 `toml:"client_id"`
	ClientSecret         string                 `toml:"client_secret"`
	TenantID             string                 `toml:"tenant_id"`
	ResourceTargets      []*ResourceTarget      `toml:"resource_target"`
	ResourceGroupTargets []*ResourceGroupTarget `toml:"resource_group_target"`
	SubscriptionTargets  []*Resource            `toml:"subscription_target"`
	Log                  telegraf.Logger        `toml:"-"`

	azureClient *azureClient
}

type ResourceTarget struct {
	ResourceID   string   `toml:"resource_id"`
	Metrics      []string `toml:"metrics"`
	Aggregations []string `toml:"aggregations"`
}

type ResourceGroupTarget struct {
	ResourceGroup string      `toml:"resource_group"`
	Resources     []*Resource `toml:"resource"`
}

type Resource struct {
	ResourceType string   `toml:"resource_type"`
	Metrics      []string `toml:"metrics"`
	Aggregations []string `toml:"aggregations"`
}

type azureClient struct {
	client               *http.Client
	accessToken          string
	accessTokenExpiresOn time.Time
}

const (
	accessTokenURLGrantType = "client_credentials"
	accessTokenURLResource  = "https://management.azure.com/"
)

var sampleConfig = `
  # can be found under Overview->Essentials in the Azure portal for your application/service
  subscription_id = "<<SUBSCRIPTION_ID>>"
  # can be obtained by registering an application under Azure Active Directory
  client_id = "<<CLIENT_ID>>"
  # can be obtained by registering an application under Azure Active Directory
  client_secret = "<<CLIENT_SECRET>>"
  # can be found under Azure Active Directory->Properties
  tenant_id = "<<TENANT_ID>>"

  # resource target #1 to collect metrics from
  [[inputs.azure_monitor.resource_target]]
    # can be found undet Overview->Essentials->JSON View in the Azure portal for your application/service
    # must start with 'resourceGroups/...' ('/subscriptions/xxxxxxxx-xxxx-xxxx-xxx-xxxxxxxxxxxx'
    # must be removed from the beginning of Resource ID property value)
    resource_id = "<<RESOURCE_ID>>"
    # the metric names to collect
    # leave the array empty to use all metrics available to this resource
    metrics = [ "<<METRIC>>", "<<METRIC>>" ]
    # metrics aggregation type value to collect
    # can be 'Total', 'Count', 'Average', 'Minimum', 'Maximum'
    # leave the array empty to collect all aggregation types values for each metric
    aggregations = [ "<<AGGREGATION>>", "<<AGGREGATION>>" ]
    
  # resource target #2 to collect metrics from
  [[inputs.azure_monitor.resource_target]]
    resource_id = "<<RESOURCE_ID>>"
    metrics = [ "<<METRIC>>", "<<METRIC>>" ]
    aggregations = [ "<<AGGREGATION>>", "<<AGGREGATION>>" ]

  # resource group target #1 to collect metrics from resources under it with resource type
  [[inputs.azure_monitor.resource_group_target]]
    # the resource group name
    resource_group = "<<RESOURCE_GROUP_NAME>>"

    # defines the resources to collect metrics from
    [[inputs.azure_monitor.resource_group_target.resource]]
      # the resource type
      resource_type = "<<RESOURCE_TYPE>>"
      metrics = [ "<<METRIC>>", "<<METRIC>>" ]
      aggregations = [ "<<AGGREGATION>>", "<<AGGREGATION>>" ]
    
    # defines the resources to collect metrics from
    [[inputs.azure_monitor.resource_group_target.resource]]
      resource_type = "<<RESOURCE_TYPE>>"
      metrics = [ "<<METRIC>>", "<<METRIC>>" ]
      aggregations = [ "<<AGGREGATION>>", "<<AGGREGATION>>" ]
      
  # resource group target #2 to collect metrics from resources under it with resource type
  [[inputs.azure_monitor.resource_group_target]]
    resource_group = "<<RESOURCE_GROUP_NAME>>"

    [[inputs.azure_monitor.resource_group_target.resource]]
      resource_type = "<<RESOURCE_TYPE>>"
      metrics = [ "<<METRIC>>", "<<METRIC>>" ]
      aggregations = [ "<<AGGREGATION>>", "<<AGGREGATION>>" ]
  
  # subscription target #1 to collect metrics from resources under it with resource type    
  [[inputs.azure_monitor.subscription_target]]
    resource_type = "<<RESOURCE_TYPE>>"
    metrics = [ "<<METRIC>>", "<<METRIC>>" ]
    aggregations = [ "<<AGGREGATION>>", "<<AGGREGATION>>" ]
    
  # subscription target #2 to collect metrics from resources under it with resource type    
  [[inputs.azure_monitor.subscription_target]]
    resource_type = "<<RESOURCE_TYPE>>"
    metrics = [ "<<METRIC>>", "<<METRIC>>" ]
    aggregations = [ "<<AGGREGATION>>", "<<AGGREGATION>>" ]
`

func (am *AzureMonitor) Description() string {
	return "Gather Azure resources metrics from Azure Monitor API"
}

func (am *AzureMonitor) SampleConfig() string {
	return sampleConfig
}

// Init is for setup, and validating config.
func (am *AzureMonitor) Init() error {
	am.azureClient = newAzureClient()

	if err := am.checkConfigValidation(); err != nil {
		return fmt.Errorf("config is not valid: %v", err)
	}

	if err := am.getAccessToken(); err != nil {
		return fmt.Errorf("error getting access token: %v", err)
	}

	if err := am.createResourceGroupTargetsFromSubscriptionTargets(); err != nil {
		return fmt.Errorf("error creating group targets from subscription targets: %v", err)
	}

	if err := am.createResourceTargetsFromResourceGroupTargets(); err != nil {
		return fmt.Errorf("error creating resource targets from resource group targets: %v", err)
	}

	if err := am.checkResourceTargetsMetricsValidation(); err != nil {
		return fmt.Errorf("error checking resource targets metrics validation: %v", err)
	}

	if err := am.setResourceTargetsMetrics(); err != nil {
		return fmt.Errorf("error setting resource targets metrics: %v", err)
	}

	if err := am.checkResourceTargetsMetricsMinTimeGrain(); err != nil {
		return fmt.Errorf("error checking resource targets metrics min time grain: %v", err)
	}

	am.checkResourceTargetsMaxMetrics()
	am.setResourceTargetsAggregations()

	am.Log.Debug("Total resource targets: ", len(am.ResourceTargets))

	if len(am.ResourceTargets) == 0 {
		return fmt.Errorf("no resource target was created. Please check your resource group targets and " +
			"subscription targets in your configuration")
	}
	fmt.Println(am.azureClient.accessToken)
	return nil
}

func (am *AzureMonitor) Gather(acc telegraf.Accumulator) error {
	// access token has expiration date. Must check every gather if access token has expired and create a new one
	if err := am.refreshAccessToken(); err != nil {
		return fmt.Errorf("error refreshing access token: %v", err)
	}

	am.collectResourceTargetsMetrics(acc)

	return nil
}

func newAzureClient() *azureClient {
	return &azureClient{
		client:               &http.Client{},
		accessToken:          "",
		accessTokenExpiresOn: time.Time{},
	}
}

func (am *AzureMonitor) getAccessToken() error {
	var response *http.Response
	var err error

	am.Log.Debug("Getting access token")

	target := fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/token", am.TenantID)
	form := url.Values{
		"grant_type":    {accessTokenURLGrantType},
		"resource":      {accessTokenURLResource},
		"client_id":     {am.ClientID},
		"client_secret": {am.ClientSecret},
	}
	response, err = am.azureClient.client.PostForm(target, form)
	if err != nil {
		return fmt.Errorf("error authenticating against Azure API: %v", err)
	}

	defer closeResponseBody(response.Body, &err)

	body, err := getResponseBody(response)
	if err != nil {
		return fmt.Errorf("error getting access token response body: %v", err)
	}

	accessToken, ok := body["access_token"].(string)
	if !ok {
		return fmt.Errorf("access_token key is missing in access token response body")
	}

	am.azureClient.accessToken = accessToken

	expiresOnStr, ok := body["expires_on"].(string)
	if !ok {
		return fmt.Errorf("expires_on key is missing in access token response body")
	}

	expiresOn, err := strconv.ParseInt(expiresOnStr, 10, 64)
	if err != nil {
		return fmt.Errorf("error ParseInt of expires_on: %v", err)
	}

	am.azureClient.accessTokenExpiresOn = time.Unix(expiresOn, 0).UTC()

	return nil
}

func (am *AzureMonitor) refreshAccessToken() error {
	now := time.Now().UTC()
	refreshAt := am.azureClient.accessTokenExpiresOn.Add(-10 * time.Minute)

	if now.After(refreshAt) {
		am.Log.Debug("Refreshing access token")

		if err := am.getAccessToken(); err != nil {
			return fmt.Errorf("error refreshing access token: %v", err)
		}
	}

	return nil
}

func (am *AzureMonitor) getAPIResponseBody(apiURL string) (map[string]interface{}, error) {
	request, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating HTTP request: %v", err)
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Authorization", "Bearer "+am.azureClient.accessToken)

	response, err := am.azureClient.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("error getting response from API: %v", err)
	}

	defer closeResponseBody(response.Body, &err)

	body, err := getResponseBody(response)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func getResponseBody(response *http.Response) (map[string]interface{}, error) {
	if response.StatusCode != 200 {
		responseBytes, _ := ioutil.ReadAll(response.Body)
		return nil, fmt.Errorf("did not get status code 200, got: %d with body: %s", response.StatusCode, string(responseBytes))
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading body of response: %v", err)
	}

	var data map[string]interface{}
	if err = json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("error unmarshalling response body: %v", err)
	}

	return data, err
}

func closeResponseBody(body io.ReadCloser, err *error) {
	if closeError := body.Close(); closeError != nil {
		*err = fmt.Errorf("error closing body of response: %v", closeError)
	}
}

func init() {
	inputs.Add("azure_monitor", func() telegraf.Input {
		return &AzureMonitor{}
	})
}
