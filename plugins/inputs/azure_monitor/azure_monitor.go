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

	SubscriptionID string    `toml:"subscription_id"`
	ClientID       string    `toml:"client_id"`
	ClientSecret   string    `toml:"client_secret"`
	TenantID       string    `toml:"tenant_id"`
	Targets        []*Target `toml:"targets"`

	Log telegraf.Logger `toml:"-"`
}

type Target struct {
	ResourceID  string
	Metrics     []string
	Aggregation []string
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

var sampleConfig = `
# can be found under properties in the Azure portal for your application/service
subscription_id = "<<SUBSCRIPTION_ID>>"
# can be obtained by registering an application under Azure Active Directory
client_id = "<<CLIENT_ID>>"
# can be obtained by registering an application under Azure Active Directory
client_secret = "<<CLIENT_SECRET>>"
# can be found under Azure Active Directory properties
tenant_id = "<<TENANT_ID>>"

# represents a target to collect metrics from
[[inputs.azure_monitor.targets]]
# can be found under properties in the Azure portal for your application/service
# must start with 'resourceGroups/...' ('/subscriptions/xxxxxxxx-xxxx-xxxx-xxx-xxxxxxxxxxxx'
# must be removed from the beginning of Resource ID property value)
resource_id = "<<RESOURCE_ID>>"
# the metrics names to collect
# leave the array empty to use all metrics available to this resource
metrics = [ "<<METRIC>>", "<<METRIC>>" ]
# metrics aggregation type value to collect
# can be 'Total', 'Count', 'Average', 'Minimum', 'Maximum'
# leave the array empty to collect all aggregation types values for each metric (if available)
aggregation = [ "<<AGGREGATION>>", "<<AGGREGATION>>" ]
  
# represents a target to collect metrics from
[[inputs.azure_monitor.targets]]
resource_id = "<<RESOURCE_ID>>"
metrics = [ "<<METRIC>>", "<<METRIC>>" ]
filter = "<<FILTER>>"
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

	err = am.getAllTargetsMetricsNames()

	if err != nil {
		return err
	}

	am.getAllTargetsAggregation()

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

	if len(am.Targets) == 0 {
		return fmt.Errorf("targets is empty or missing. Please check your configuration")
	}

	for index, target := range am.Targets {
		if target.ResourceID == "" {
			return fmt.Errorf("target #%d resource_id is empty or missing. Please check your configuration", index+1)
		}
	}

	return nil
}

func (am *AzureMonitor) getAllTargetsMetricsNames() error {
	var waitGroup sync.WaitGroup

	errChan := make(chan error, 1)
	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	for index, target := range am.Targets {
		if len(target.Metrics) > 0 {
			continue
		}

		am.Log.Info("Getting metrics names for target #", index+1)
		waitGroup.Add(1)

		go func(target *Target) {
			defer waitGroup.Done()

			select {
			case <-ctx.Done():
				return
			default:
			}

			err := am.getTargetMetricsNames(target)

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

func (am *AzureMonitor) getTargetMetricsNames(target *Target) error {
	apiURL := am.buildMetricDefinitionsAPIURL(target)
	body, err := am.getTargetResponseBody(apiURL)

	if err != nil {
		return err
	}

	err = target.getTargetMetricsNames(body)

	if err != nil {
		return err
	}

	return nil
}

func (am *AzureMonitor) getAllTargetsAggregation() {
	for _, target := range am.Targets {
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

	err = am.collectAllTargetsMetrics(acc)

	if err != nil {
		return err
	}

	return nil
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
		"grant_type":    {"client_credentials"},
		"resource":      {"https://management.azure.com/"},
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

func (am *AzureMonitor) buildMetricDefinitionsAPIURL(target *Target) string {
	apiURL := fmt.Sprintf(
		"https://management.azure.com/subscriptions/%s/%s/providers/microsoft.insights/metricDefinitions?api-version=2018-01-01",
		am.SubscriptionID, target.ResourceID)

	return apiURL
}

func (am *AzureMonitor) buildMetricValuesAPIURL(target *Target) string {
	apiURL := fmt.Sprintf(
		"https://management.azure.com/subscriptions/%s/%s/providers/microsoft.insights/metrics?metricnames=%s&"+
			"aggregation=%s&api-version=2019-07-01",
		am.SubscriptionID, target.ResourceID, strings.Join(target.Metrics, ","), strings.Join(target.Aggregation, ","))

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

func (am *AzureMonitor) collectAllTargetsMetrics(acc telegraf.Accumulator) error {
	var waitGroup sync.WaitGroup

	errChan := make(chan error, 1)
	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	for index, target := range am.Targets {
		am.Log.Info("Collecting metrics for target #", index+1)
		waitGroup.Add(1)

		go func(target *Target) {
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

			metrics, err := collectTargetMetrics(body)

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

func (t *Target) getTargetMetricsNames(body []byte) error {
	bodyData, err := unmarshalJSON(body)

	if err != nil {
		return err
	}

	for _, value := range bodyData["value"].([]interface{}) {
		metricName := value.(map[string]interface{})["name"].(map[string]interface{})["value"].(string)
		t.Metrics = append(t.Metrics, metricName)
	}

	return nil
}

func (m *Metric) getMetricName(value map[string]interface{}) {
	resourceID := strings.Split(value["id"].(string), "/")
	replacer := strings.NewReplacer(".", "_", "/", "_")
	metricName := strings.ToLower(value["name"].(map[string]interface{})["localizedValue"].(string))

	m.name = fmt.Sprintf("azure_monitor_%s_%s",
		replacer.Replace(strings.ToLower(resourceID[6]+"_"+resourceID[7])),
		strings.Replace(metricName, " ", "_", -1),
	)
}

func (m *Metric) getMetricFields(data []interface{}) {
	for key, element := range data[len(data)-1].(map[string]interface{}) {
		m.fields[key] = element
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

func collectTargetMetrics(body []byte) ([]*Metric, error) {
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
