package azure_monitor

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type AzureMonitor struct {
	azureClient *AzureClient

	SubscriptionID       string                 `toml:"subscription_id"`
	ClientID             string                 `toml:"client_id"`
	ClientSecret         string                 `toml:"client_secret"`
	TenantID             string                 `toml:"tenant_id"`
	ResourceTargets      []*ResourceTarget      `toml:"resource_target"`
	ResourceGroupTargets []*ResourceGroupTarget `toml:"resource_group_target"`
	SubscriptionTargets  []*Resource            `toml:"subscription_target"`

	Log telegraf.Logger `toml:"-"`
}

type ResourceTarget struct {
	ResourceID  string   `toml:"resource_id"`
	Metrics     []string `toml:"metrics"`
	Aggregation []string `toml:"aggregation"`
}

type ResourceGroupTarget struct {
	ResourceGroup string      `toml:"resource_group"`
	Resources     []*Resource `toml:"resource"`
}

type Resource struct {
	ResourceType string   `toml:"resource_type"`
	Metrics      []string `toml:"metrics"`
	Aggregation  []string `toml:"aggregation"`
}

type AzureClient struct {
	client               *http.Client
	accessToken          string
	accessTokenExpiresOn time.Time
}

type Metric struct {
	name   string
	fields map[string]interface{}
	tags   map[string]string
}

const (
	maxMetricsPerRequest    int = 20
	minMetricsFields            = 2
	accessTokenURLGrantType     = "client_credentials"
	accessTokenURLResource      = "https://management.azure.com/"
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
    aggregation = [ "<<AGGREGATION>>", "<<AGGREGATION>>" ]
    
  # resource target #2 to collect metrics from
  [[inputs.azure_monitor.resource_target]]
    resource_id = "<<RESOURCE_ID>>"
    metrics = [ "<<METRIC>>", "<<METRIC>>" ]
    aggregation = [ "<<AGGREGATION>>", "<<AGGREGATION>>" ]

  # resource group target #1 to collect metrics from resources under it with resource type
  [[inputs.azure_monitor.resource_group_target]]
    # the resource group name
    resource_group = "<<RESOURCE_GROUP_NAME>>"

    # defines the resources to collect metrics from
    [[inputs.azure_monitor.resource_group_target.resource]]
      # the resource type
      resource_type = "<<RESOURCE_TYPE>>"
      metrics = [ "<<METRIC>>", "<<METRIC>>" ]
      aggregation = [ "<<AGGREGATION>>", "<<AGGREGATION>>" ]
    
    # defines the resources to collect metrics from
    [[inputs.azure_monitor.resource_group_target.resource]]
      resource_type = "<<RESOURCE_TYPE>>"
      metrics = [ "<<METRIC>>", "<<METRIC>>" ]
      aggregation = [ "<<AGGREGATION>>", "<<AGGREGATION>>" ]
      
  # resource group target #2 to collect metrics from resources under it with resource type
  [[inputs.azure_monitor.resource_group_target]]
    resource_group = "<<RESOURCE_GROUP_NAME>>"

    [[inputs.azure_monitor.resource_group_target.resource]]
      resource_type = "<<RESOURCE_TYPE>>"
      metrics = [ "<<METRIC>>", "<<METRIC>>" ]
      aggregation = [ "<<AGGREGATION>>", "<<AGGREGATION>>" ]
  
  # subscription target #1 to collect metrics from resources under it with resource type    
  [[inputs.azure_monitor.subscription_target]]
    resource_type = "<<RESOURCE_TYPE>>"
    metrics = [ "<<METRIC>>", "<<METRIC>>" ]
    aggregation = [ "<<AGGREGATION>>", "<<AGGREGATION>>" ]
    
  # subscription target #2 to collect metrics from resources under it with resource type    
  [[inputs.azure_monitor.subscription_target]]
    resource_type = "<<RESOURCE_TYPE>>"
    metrics = [ "<<METRIC>>", "<<METRIC>>" ]
    aggregation = [ "<<AGGREGATION>>", "<<AGGREGATION>>" ]
`

func (am *AzureMonitor) Description() string {
	return "Gather Azure resources metrics from Azure Monitor API"
}

func (am *AzureMonitor) SampleConfig() string {
	return sampleConfig
}

// Init is for setup, and validating config.
func (am *AzureMonitor) Init() error {
	err := am.checkConfigValidation()

	if err != nil {
		return err
	}

	err = am.getAccessToken()

	if err != nil {
		return err
	}

	err = am.createResourceGroupTargetsFromSubscriptionTargets()

	if err != nil {
		return err
	}

	err = am.createResourceTargetsFromResourceGroupTargets()

	if err != nil {
		return err
	}

	err = am.getResourceTargetsMetrics()

	if err != nil {
		return err
	}

	am.setResourceTargetsAggregation()
	am.checkTargetsMaximumMetrics()

	am.Log.Info("Total targets: ", len(am.ResourceTargets))

	return nil
}

func (am *AzureMonitor) checkConfigValidation() error {
	if am.SubscriptionID == "" {
		return fmt.Errorf("subscription_id is empty or missing. Please check your configuration")
	}

	if am.ClientID == "" {
		return fmt.Errorf("client_id is empty or missing. Please check your configuration")
	}

	if am.ClientSecret == "" {
		return fmt.Errorf("client_secret is empty or missing. Please check your configuration")
	}

	if am.TenantID == "" {
		return fmt.Errorf("tenant_id is empty or missing. Please check your configuration")
	}

	if len(am.ResourceTargets) == 0 && len(am.ResourceGroupTargets) == 0 && len(am.SubscriptionTargets) == 0 {
		return fmt.Errorf("there is no target to collect metrics from. Please check your configuration")
	}

	for index, resourceTarget := range am.ResourceTargets {
		if resourceTarget.ResourceID == "" {
			return fmt.Errorf(
				"resource target #%d resource_id is empty or missing. Please check your configuration", index+1)
		}
	}

	for resourceGroupIndex, resourceGroupTarget := range am.ResourceGroupTargets {
		if resourceGroupTarget.ResourceGroup == "" {
			return fmt.Errorf(
				"resource group target #%d resource_group is empty or missing. Please check your configuration",
				resourceGroupIndex+1)
		}

		for resourceIndex, resource := range resourceGroupTarget.Resources {
			if resource.ResourceType == "" {
				return fmt.Errorf(
					"resource group target #%d resource #%d resource_type is empty or missing. Please check your configuration",
					resourceGroupIndex+1, resourceIndex+1)
			}
		}
	}

	for index, target := range am.SubscriptionTargets {
		if target.ResourceType == "" {
			return fmt.Errorf(
				"subscription target #%d resource_type is empty or missing. Please check your configuration", index+1)
		}
	}

	return nil
}

func (am *AzureMonitor) createResourceGroupTargetsFromSubscriptionTargets() error {
	if len(am.SubscriptionTargets) == 0 {
		return nil
	}

	am.Log.Info("Creating resource group targets from subscription targets")

	apiURL := am.buildSubscriptionResourceGroupsAPIURL()
	body, err := am.getTargetResponseBody(apiURL)

	if err != nil {
		return err
	}

	bodyData, err := unmarshalJSON(body)

	if err != nil {
		return err
	}

	for _, value := range bodyData["value"].([]interface{}) {
		resourceGroup := value.(map[string]interface{})["name"].(string)
		resourceGroupTarget := NewResourceGroupTarget(resourceGroup, am.SubscriptionTargets)

		am.ResourceGroupTargets = append(am.ResourceGroupTargets, resourceGroupTarget)
	}

	return nil
}

func (am *AzureMonitor) createResourceTargetsFromResourceGroupTargets() error {
	var waitGroup sync.WaitGroup

	errChan := make(chan error, 1)
	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	for index, target := range am.ResourceGroupTargets {
		am.Log.Info("Creating resource targets from resource group target #", index+1)
		waitGroup.Add(1)

		go func(target *ResourceGroupTarget) {
			defer waitGroup.Done()

			select {
			case <-ctx.Done():
				return
			default:
			}

			err := am.createResourceTargetFromResourceGroupTarget(target)

			if err != nil {
				select {
				case errChan <- err:
				default:
				}

				cancel()
				return
			}
		}(target)
	}

	waitGroup.Wait()

	select {
	case err := <-errChan:
		return err
	default:
	}

	return nil
}

func (am *AzureMonitor) createResourceTargetFromResourceGroupTarget(target *ResourceGroupTarget) error {
	apiURL := am.buildResourceGroupResourcesAPIURL(target)
	body, err := am.getTargetResponseBody(apiURL)

	if err != nil {
		return err
	}

	bodyData, err := unmarshalJSON(body)

	if err != nil {
		return err
	}

	for _, value := range bodyData["value"].([]interface{}) {
		resourceType := value.(map[string]interface{})["type"].(string)
		resourceIndex := target.getResourceWithResourceTypeIndex(resourceType)

		if resourceIndex == -1 {
			continue
		}

		resourceName := value.(map[string]interface{})["name"].(string)
		resourceID := fmt.Sprintf("resourceGroups/%s/providers/%s/%s", target.ResourceGroup, resourceType, resourceName)
		resourceTarget := NewResourceTarget(resourceID, target.Resources[resourceIndex].Metrics, target.Resources[resourceIndex].Aggregation)

		am.ResourceTargets = append(am.ResourceTargets, resourceTarget)
	}

	return nil
}

func (am *AzureMonitor) getResourceTargetsMetrics() error {
	var waitGroup sync.WaitGroup

	errChan := make(chan error, 1)
	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	for index, target := range am.ResourceTargets {
		if len(target.Metrics) > 0 {
			continue
		}

		am.Log.Info("Getting metrics for target #", index+1)
		waitGroup.Add(1)

		go func(target *ResourceTarget) {
			defer waitGroup.Done()

			select {
			case <-ctx.Done():
				return
			default:
			}

			err := am.getResourceTargetMetrics(target)

			if err != nil {
				select {
				case errChan <- err:
				default:
				}

				cancel()
				return
			}
		}(target)
	}

	waitGroup.Wait()

	select {
	case err := <-errChan:
		return err
	default:
	}

	return nil
}

func (am *AzureMonitor) getResourceTargetMetrics(target *ResourceTarget) error {
	apiURL := am.buildMetricDefinitionsAPIURL(target)
	body, err := am.getTargetResponseBody(apiURL)

	if err != nil {
		return err
	}

	err = target.setResourceTargetMetrics(body)

	if err != nil {
		return err
	}

	return nil
}

func (am *AzureMonitor) checkTargetsMaximumMetrics() {
	for _, target := range am.ResourceTargets {
		if len(target.Metrics) <= maxMetricsPerRequest {
			continue
		}

		for start := maxMetricsPerRequest; start < len(target.Metrics); start += maxMetricsPerRequest {
			end := start + maxMetricsPerRequest

			if end > len(target.Metrics) {
				end = len(target.Metrics)
			}

			newTargetMetrics := target.Metrics[start:end]

			var newTargetAggregation []string

			for _, aggregation := range target.Aggregation {
				newTargetAggregation = append(newTargetAggregation, aggregation)
			}

			newTarget := NewResourceTarget(target.ResourceID, newTargetMetrics, newTargetAggregation)

			am.ResourceTargets = append(am.ResourceTargets, newTarget)
		}

		target.Metrics = target.Metrics[:maxMetricsPerRequest]
	}
}

func (am *AzureMonitor) setResourceTargetsAggregation() {
	for _, target := range am.ResourceTargets {
		if len(target.Aggregation) > 0 {
			continue
		}

		target.Aggregation = append(target.Aggregation, "Total", "Count", "Average", "Minimum", "Maximum")
	}
}

func (am *AzureMonitor) Gather(acc telegraf.Accumulator) error {
	err := am.refreshAccessToken()

	if err != nil {
		return err
	}

	err = am.collectResourceTargetsMetrics(acc)

	if err != nil {
		return err
	}

	return nil
}

func NewResourceTarget(
	resourceID string,
	metrics []string,
	aggregation []string,
) *ResourceTarget {
	return &ResourceTarget{
		ResourceID:  resourceID,
		Metrics:     metrics,
		Aggregation: aggregation,
	}
}

func NewResourceGroupTarget(resourceGroup string, resources []*Resource) *ResourceGroupTarget {
	return &ResourceGroupTarget{
		ResourceGroup: resourceGroup,
		Resources:     resources,
	}
}

func NewAzureClient() *AzureClient {
	return &AzureClient{
		client:               &http.Client{},
		accessToken:          "",
		accessTokenExpiresOn: time.Time{},
	}
}

func NewMetric() *Metric {
	return &Metric{
		name:   "",
		fields: make(map[string]interface{}),
		tags:   make(map[string]string),
	}
}

func (am *AzureMonitor) getAccessToken() error {
	var response *http.Response
	var err error

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
		return err
	}

	data, err := unmarshalJSON(body)

	if err != nil {
		return err
	}

	am.azureClient.accessToken = data["access_token"].(string)
	expiresOn, err := strconv.ParseInt(data["expires_on"].(string), 10, 64)

	if err != nil {
		return fmt.Errorf("error ParseInt of expires_on failed: %v", err)
	}

	am.azureClient.accessTokenExpiresOn = time.Unix(expiresOn, 0).UTC()

	return nil
}

func (am *AzureMonitor) refreshAccessToken() error {
	now := time.Now().UTC()
	refreshAt := am.azureClient.accessTokenExpiresOn.Add(-10 * time.Minute)

	if now.After(refreshAt) {
		err := am.getAccessToken()

		if err != nil {
			return fmt.Errorf("error refreshing access token: %v", err)
		}
	}

	return nil
}

func (am *AzureMonitor) buildMetricDefinitionsAPIURL(target *ResourceTarget) string {
	apiURL := fmt.Sprintf(
		"https://management.azure.com/subscriptions/%s/%s/providers/microsoft.insights/metricDefinitions?api-version=2018-01-01",
		am.SubscriptionID, target.ResourceID)

	return apiURL
}

func (am *AzureMonitor) buildMetricValuesAPIURL(target *ResourceTarget) string {
	metrics := strings.Join(target.Metrics, ",")
	metrics = strings.Replace(metrics, " ", "+", -1)
	apiURL := fmt.Sprintf(
		"https://management.azure.com/subscriptions/%s/%s/providers/microsoft.insights/metrics?metricnames=%s&"+
			"aggregation=%s&api-version=2019-07-01",
		am.SubscriptionID, target.ResourceID, metrics, strings.Join(target.Aggregation, ","))

	return apiURL
}

func (am *AzureMonitor) buildResourceGroupResourcesAPIURL(target *ResourceGroupTarget) string {
	apiURL := fmt.Sprintf("https://management.azure.com/subscriptions/%s/resourceGroups/%s/resources?api-version=2018-02-01",
		am.SubscriptionID, target.ResourceGroup)

	return apiURL
}

func (am *AzureMonitor) buildSubscriptionResourceGroupsAPIURL() string {
	apiURL := fmt.Sprintf("https://management.azure.com/subscriptions/%s/resourceGroups?api-version=2018-02-01",
		am.SubscriptionID)

	return apiURL
}

func (am *AzureMonitor) getTargetResponseBody(apiURL string) ([]byte, error) {
	request, err := http.NewRequest("GET", apiURL, nil)

	if err != nil {
		return nil, fmt.Errorf("error creating HTTP request: %v", err)
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Authorization", "Bearer "+am.azureClient.accessToken)

	response, err := am.azureClient.client.Do(request)

	if err != nil {
		return nil, fmt.Errorf("error getting response from Azure Monitor API: %v", err)
	}

	defer closeResponseBody(response.Body, &err)

	body, err := getResponseBody(response)

	return body, err
}

func (am *AzureMonitor) collectResourceTargetsMetrics(acc telegraf.Accumulator) error {
	var waitGroup sync.WaitGroup

	errChan := make(chan error, 1)
	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	for index, target := range am.ResourceTargets {
		am.Log.Info("Collecting metrics for target #", index+1)
		waitGroup.Add(1)

		go func(target *ResourceTarget) {
			defer waitGroup.Done()

			select {
			case <-ctx.Done():
				return
			default:
			}

			apiURL := am.buildMetricValuesAPIURL(target)
			body, err := am.getTargetResponseBody(apiURL)

			if err != nil {
				select {
				case errChan <- err:
				default:
				}

				cancel()
				return
			}

			metrics, err := am.collectResourceTargetMetrics(body)

			if err != nil {
				select {
				case errChan <- err:
				default:
				}

				cancel()
				return
			}

			for _, metric := range metrics {
				acc.AddFields(metric.name, metric.fields, metric.tags, time.Now())
			}
		}(target)

		waitGroup.Wait()

		select {
		case err := <-errChan:
			return err
		default:
		}
	}

	return nil
}

func (am *AzureMonitor) collectResourceTargetMetrics(body []byte) ([]*Metric, error) {
	bodyData, err := unmarshalJSON(body)

	if err != nil {
		return nil, err
	}

	var metrics []*Metric

	for _, value := range bodyData["value"].([]interface{}) {
		var metric *Metric

		timeSeries := value.(map[string]interface{})["timeseries"].([]interface{})[0]
		data := timeSeries.(map[string]interface{})["data"].([]interface{})

		if len(data) == 0 {
			metricName, fullResourceName, resourceType := getMetricWithNoValueDetails(value.(map[string]interface{}))

			am.Log.Info("There is no value to metric: ", metricName, " for resource: ", fullResourceName, " type: ", resourceType)
			continue
		}

		if !isMetricHaveValue(data) {
			metricName, fullResourceName, resourceType := getMetricWithNoValueDetails(value.(map[string]interface{}))

			am.Log.Info("There is no value to metric: ", metricName, " for resource: ", fullResourceName, " type: ", resourceType)
			continue
		}

		metric = NewMetric()

		metric.getMetricName(value.(map[string]interface{}))
		metric.getMetricFields(data)
		metric.getMetricTags(bodyData, value.(map[string]interface{}))

		metrics = append(metrics, metric)
	}

	return metrics, err
}

func (rt *ResourceTarget) setResourceTargetMetrics(body []byte) error {
	bodyData, err := unmarshalJSON(body)

	if err != nil {
		return err
	}

	for _, value := range bodyData["value"].([]interface{}) {
		metricName := value.(map[string]interface{})["name"].(map[string]interface{})["value"].(string)
		rt.Metrics = append(rt.Metrics, metricName)
	}

	return nil
}

func (rgt *ResourceGroupTarget) getResourceWithResourceTypeIndex(resourceType string) int {
	for index, resource := range rgt.Resources {
		if resource.ResourceType == resourceType {
			return index
		}
	}

	return -1
}

func (m *Metric) getMetricName(value map[string]interface{}) {
	resourceID := strings.Split(value["id"].(string), "/")
	replacer := strings.NewReplacer(".", "_", "/", "_", " ", "_", "(", "_", ")", "_")
	metricName := value["name"].(map[string]interface{})["localizedValue"].(string)

	m.name = fmt.Sprintf("azure_monitor_%s_%s",
		replacer.Replace(strings.ToLower(resourceID[6]+"_"+resourceID[7])),
		replacer.Replace(strings.ToLower(metricName)),
	)
}

func (m *Metric) getMetricFields(data []interface{}) {
	for index := len(data) - 1; index >= 0; index-- {
		if len(data[index].(map[string]interface{})) < minMetricsFields {
			continue
		}

		for key, element := range data[index].(map[string]interface{}) {
			m.fields[key] = element
		}

		return
	}
}

func (m *Metric) getMetricTags(bodyData map[string]interface{}, value map[string]interface{}) {
	resourceID := strings.Split(value["id"].(string), "/")

	m.tags["subscription_id"] = resourceID[2]
	m.tags["resource_group"] = resourceID[4]
	m.tags["namespace"] = resourceID[6] + "/" + resourceID[7]
	m.tags["resource_name"] = resourceID[8]
	m.tags["resource_region"] = bodyData["resourceregion"].(string)
	m.tags["unit"] = value["unit"].(string)
}

func isMetricHaveValue(data []interface{}) bool {
	for index := len(data) - 1; index >= 0; index-- {
		if len(data[index].(map[string]interface{})) >= minMetricsFields {
			return true
		}
	}

	return false
}

func getMetricWithNoValueDetails(value map[string]interface{}) (string, string, string) {
	metricName := value["name"].(map[string]interface{})["value"].(string)
	resourceID := strings.Split(value["id"].(string), "/")
	fullResourceName := resourceID[4] + "/" + resourceID[8]
	resourceType := resourceID[6] + "/" + resourceID[7]

	return metricName, fullResourceName, resourceType
}

func getResponseBody(response *http.Response) ([]byte, error) {
	if response.StatusCode != 200 {
		responseBytes, _ := ioutil.ReadAll(response.Body)

		return nil, fmt.Errorf("did not get status code 200, got: %d with body: %s", response.StatusCode, string(responseBytes))
	}

	body, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return nil, fmt.Errorf("error reading body of response: %v", err)
	}

	return body, err
}

func closeResponseBody(body io.ReadCloser, err *error) {
	closeError := body.Close()

	if closeError != nil {
		*err = fmt.Errorf("error closing body of response: %v", closeError)
	}
}

func unmarshalJSON(body []byte) (map[string]interface{}, error) {
	var data map[string]interface{}
	err := json.Unmarshal(body, &data)

	if err != nil {
		return nil, fmt.Errorf("error unmarshalling response body: %v", err)
	}

	return data, err
}

func init() {
	inputs.Add("azure_monitor", func() telegraf.Input {
		return &AzureMonitor{
			azureClient: NewAzureClient(),
		}
	})
}
