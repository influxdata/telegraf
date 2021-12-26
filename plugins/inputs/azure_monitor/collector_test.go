package azure_monitor

import (
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetMetricName_Success(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: "subscriptionID",
		ClientID:       "clientID",
		ClientSecret:   "clientSecret",
		TenantID:       "tenantID",
		ResourceTargets: []*ResourceTarget{
			newResourceTarget("resource1", []string{"metric1"}, []string{"Total", "Maximum"}),
		},
		Log:         testutil.Logger{},
		azureClient: newAzureClient(),
	}

	resourceMetricValuesBody, err := getFileBody("testData/resource_1_metric_values_body.json")
	require.NoError(t, err)
	require.NotNil(t, resourceMetricValuesBody)

	apiURL := am.buildMetricValuesAPIURL(am.ResourceTargets[0])
	require.NotNil(t, apiURL)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		apiURL,
		httpmock.NewBytesResponder(200, resourceMetricValuesBody))

	body, err := am.getAPIResponseBody(apiURL)
	require.NoError(t, err)
	require.NotNil(t, body)

	values, ok := body["value"].([]interface{})
	require.True(t, ok)
	require.NotNil(t, values)

	metricName, err := getMetricName(body, values[0].(map[string]interface{}))
	require.NoError(t, err)
	require.NotNil(t, metricName)

	assert.Equal(t, "azure_monitor_microsoft_type1_metric1", *metricName)
}

func TestGetMetricFields_AllTimeSeriesWithData(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: "subscriptionID",
		ClientID:       "clientID",
		ClientSecret:   "clientSecret",
		TenantID:       "tenantID",
		ResourceTargets: []*ResourceTarget{
			newResourceTarget("resource1", []string{"metric1"}, []string{"Total", "Maximum"}),
		},
		Log:         testutil.Logger{},
		azureClient: newAzureClient(),
	}

	expectedMetricFields := make(map[string]interface{}, 0)
	expectedMetricFields["timeStamp"] = "2021-11-05T10:59:00Z"
	expectedMetricFields["total"] = 5.0
	expectedMetricFields["maximum"] = 5.0

	apiURL := am.buildMetricValuesAPIURL(am.ResourceTargets[0])
	require.NotNil(t, apiURL)

	resourceMetricValuesBody, err := getFileBody("testData/resource_1_metric_values_body.json")
	require.NoError(t, err)
	require.NotNil(t, resourceMetricValuesBody)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		apiURL,
		httpmock.NewBytesResponder(200, resourceMetricValuesBody))

	body, err := am.getAPIResponseBody(apiURL)
	require.NoError(t, err)
	require.NotNil(t, body)

	values, ok := body["value"].([]interface{})
	require.True(t, ok)
	require.NotNil(t, values)

	timesSeries, ok := values[0].(map[string]interface{})["timeseries"].([]interface{})
	require.True(t, ok)
	require.NotNil(t, timesSeries)

	data, ok := timesSeries[0].(map[string]interface{})["data"].([]interface{})
	require.True(t, ok)
	require.NotNil(t, data)

	metricFields := getMetricFields(data)
	require.NotNil(t, metricFields)

	assert.Equal(t, expectedMetricFields, metricFields)
}

func TestGetMetricFields_LastTimeSeriesWithoutData(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: "subscriptionID",
		ClientID:       "clientID",
		ClientSecret:   "clientSecret",
		TenantID:       "tenantID",
		ResourceTargets: []*ResourceTarget{
			newResourceTarget("resource1", []string{"metric2"}, []string{"Total", "Maximum"}),
		},
		Log:         testutil.Logger{},
		azureClient: newAzureClient(),
	}

	expectedMetricFields := make(map[string]interface{}, 0)
	expectedMetricFields["timeStamp"] = "2021-11-05T10:57:00Z"
	expectedMetricFields["total"] = 4.0
	expectedMetricFields["maximum"] = 9.0

	apiURL := am.buildMetricValuesAPIURL(am.ResourceTargets[0])
	require.NotNil(t, apiURL)

	resourceMetricValuesBody, err := getFileBody("testData/resource_1_metric_values_body.json")
	require.NoError(t, err)
	require.NotNil(t, resourceMetricValuesBody)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		apiURL,
		httpmock.NewBytesResponder(200, resourceMetricValuesBody))

	body, err := am.getAPIResponseBody(apiURL)
	require.NoError(t, err)
	require.NotNil(t, body)

	values, ok := body["value"].([]interface{})
	require.True(t, ok)
	require.NotNil(t, values)

	timesSeries, ok := values[1].(map[string]interface{})["timeseries"].([]interface{})
	require.True(t, ok)
	require.NotNil(t, timesSeries)

	data, ok := timesSeries[0].(map[string]interface{})["data"].([]interface{})
	require.True(t, ok)
	require.NotNil(t, data)

	metricFields := getMetricFields(data)
	require.NotNil(t, metricFields)

	assert.Equal(t, expectedMetricFields, metricFields)
}

func TestGetMetricFields_AllTimeSeriesWithoutData(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: "subscriptionID",
		ClientID:       "clientID",
		ClientSecret:   "clientSecret",
		TenantID:       "tenantID",
		ResourceTargets: []*ResourceTarget{
			newResourceTarget("resource2", []string{"metric3"}, []string{"Total", "Maximum"}),
		},
		Log:         testutil.Logger{},
		azureClient: newAzureClient(),
	}

	apiURL := am.buildMetricValuesAPIURL(am.ResourceTargets[0])
	require.NotNil(t, apiURL)

	resourceMetricValuesBody, err := getFileBody("testData/resource_2_metric_values_body.json")
	require.NoError(t, err)
	require.NotNil(t, resourceMetricValuesBody)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		apiURL,
		httpmock.NewBytesResponder(200, resourceMetricValuesBody))

	body, err := am.getAPIResponseBody(apiURL)
	require.NoError(t, err)
	require.NotNil(t, body)

	values, ok := body["value"].([]interface{})
	require.True(t, ok)
	require.NotNil(t, values)

	timesSeries, ok := values[2].(map[string]interface{})["timeseries"].([]interface{})
	require.True(t, ok)
	require.NotNil(t, timesSeries)

	data, ok := timesSeries[0].(map[string]interface{})["data"].([]interface{})
	require.True(t, ok)
	require.NotNil(t, data)

	metricFields := getMetricFields(data)
	require.Nil(t, metricFields)
}

func TestGetMetricTags_Success(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: "subscriptionID",
		ClientID:       "clientID",
		ClientSecret:   "clientSecret",
		TenantID:       "tenantID",
		ResourceTargets: []*ResourceTarget{
			newResourceTarget("resource1", []string{"metric1"}, []string{"Total", "Maximum"}),
		},
		Log:         testutil.Logger{},
		azureClient: newAzureClient(),
	}

	expectedMetricTags := make(map[string]string, 0)
	expectedMetricTags["subscription_id"] = "subscriptionID"
	expectedMetricTags["resource_group"] = "resourceGroup"
	expectedMetricTags["namespace"] = "Microsoft/type1"
	expectedMetricTags["resource_name"] = "resource1"
	expectedMetricTags["resource_region"] = "eastus"
	expectedMetricTags["unit"] = "Count"

	apiURL := am.buildMetricValuesAPIURL(am.ResourceTargets[0])
	require.NotNil(t, apiURL)

	resourceMetricValuesBody, err := getFileBody("testData/resource_1_metric_values_body.json")
	require.NoError(t, err)
	require.NotNil(t, resourceMetricValuesBody)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		apiURL,
		httpmock.NewBytesResponder(200, resourceMetricValuesBody))

	body, err := am.getAPIResponseBody(apiURL)
	require.NoError(t, err)
	require.NotNil(t, body)

	values, ok := body["value"].([]interface{})
	require.True(t, ok)
	require.NotNil(t, values)

	metricTags, err := getMetricTags(body, values[0].(map[string]interface{}))
	require.NoError(t, err)
	require.NotNil(t, metricTags)

	assert.Equal(t, expectedMetricTags, metricTags)
}
