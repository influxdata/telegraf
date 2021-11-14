package azure_monitor

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/influxdata/telegraf/testutil"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

var resourceMetricsDefinitionsBody = `
{
  "value": [
    {
      "name": {
        "value": "UsedCapacity",
        "localizedValue": "Used capacity"
      }
    },
	{
      "name": {
        "value": "Transactions",
        "localizedValue": "Transactions"
      }
	},
	{
	  "name": {
        "value": "Ingress",
        "localizedValue": "Ingress"
      }
	}
  ]
}
`

var target1MetricsValues = `
{
  "cost": 0,
  "timespan": "2021-11-05T10:00:00Z/2021-11-05T11:01:00Z",
  "interval": "PT1H",
  "value": [
	{
      "id": "/subscriptions/subscription_id/resourceGroups/azure-rg/providers/Microsoft.Storage/storageAccounts/azuresa1/providers/Microsoft.Insights/metrics/UsedCapacity",
      "type": "Microsoft.Insights/metrics",
      "name": {
        "value": "UsedCapacity",
        "localizedValue": "Used capacity"
      },
      "displayDescription": "The amount of storage used by the storage account. For standard storage accounts, it's the sum of capacity used by blob, table, file, and queue. For premium storage accounts and Blob storage accounts, it is the same as BlobCapacity or FileCapacity.",
      "unit": "Bytes",
      "timeseries": [
        {
          "metadatavalues": [],
          "data": [
            {
              "timeStamp": "2021-11-05T10:00:00Z",
              "total": 9065573.0,
			  "average": 8501235.0
            }
          ]
        }
      ]
    },
    {
      "id": "/subscriptions/subscription_id/resourceGroups/azure-rg/providers/Microsoft.Storage/storageAccounts/azuresa1/providers/Microsoft.Insights/metrics/Transactions",
      "type": "Microsoft.Insights/metrics",
      "name": {
        "value": "Transactions",
        "localizedValue": "Transactions"
      },
      "unit": "Count",
      "timeseries": [
        {
          "metadatavalues": [],
          "data": [
            {
              "timeStamp": "2021-11-05T10:00:00Z",
              "total": 5.0,
    		  "average": 4.0
            }
          ]
        }
      ]
    }
  ],
  "namespace": "Microsoft.Storage/storageAccounts",
  "resourceregion": "eastus"
}
`

var target2MetricsValues = `
{
  "cost": 0,
  "timespan": "2021-11-05T10:00:00Z/2021-11-05T11:01:00Z",
  "interval": "PT1M",
  "value": [
	{
      "id": "/subscriptions/subscription_id/resourceGroups/azure-rg/providers/Microsoft.Storage/storageAccounts/azuresa2/providers/Microsoft.Insights/metrics/Ingress",
      "type": "Microsoft.Insights/metrics",
      "name": {
        "value": "Ingress",
        "localizedValue": "Ingress"
      },
      "displayDescription": "The amount of storage used by the storage account. For standard storage accounts, it's the sum of capacity used by blob, table, file, and queue. For premium storage accounts and Blob storage accounts, it is the same as BlobCapacity or FileCapacity.",
      "unit": "Bytes",
      "timeseries": [
        {
          "metadatavalues": [],
          "data": [
            {
              "timeStamp": "2021-11-05T10:00:00Z",
              "minimum": 200.0,
			  "maximum": 200.0
            },
			{
              "timeStamp": "2021-11-05T10:00:00Z",
              "minimum": 190.0,
			  "maximum": 210.0
            },
			{
              "timeStamp": "2021-11-05T10:00:00Z",
              "minimum": 180.0,
			  "maximum": 220.0
            },
			{
              "timeStamp": "2021-11-05T10:00:00Z",
              "minimum": 150.0,
			  "maximum": 225.0
            },
			{
              "timeStamp": "2021-11-05T10:00:00Z",
              "minimum": 125.0,
			  "maximum": 250.0
            }
          ]
        }
      ]
    }
  ],
  "namespace": "Microsoft.Storage/storageAccounts",
  "resourceregion": "eastus"
}
`

func getTarget(resourceID string, metrics []string, aggregation []string) *Target {
	return &Target{
		ResourceID:  resourceID,
		Metrics:     metrics,
		Aggregation: aggregation,
	}
}

func getAzureMonitor(subscriptionID string, clientID string, clientSecret string, tenantID string, targets []*Target) *AzureMonitor {
	return &AzureMonitor{
		azureClient:    NewAzureClient(),
		SubscriptionID: subscriptionID,
		ClientID:       clientID,
		ClientSecret:   clientSecret,
		TenantID:       tenantID,
		Targets:        targets,
		Log:            testutil.Logger{},
	}
}

func getTarget1Metrics() []*Metric {
	var metrics []*Metric

	metric1 := NewMetric()

	metric1.name = "azure_monitor_microsoft_storage_storageaccounts_used_capacity"
	metric1.fields["timeStamp"] = "2021-11-05T10:00:00Z"
	metric1.fields["total"] = 9065573.0
	metric1.fields["average"] = 8501235.0
	metric1.tags["subscription_id"] = "subscription_id"
	metric1.tags["resource_group"] = "azure-rg"
	metric1.tags["namespace"] = "Microsoft.Storage/storageAccounts"
	metric1.tags["resource_name"] = "azuresa1"
	metric1.tags["resource_region"] = "eastus"
	metric1.tags["unit"] = "Bytes"

	metric2 := NewMetric()

	metric2.name = "azure_monitor_microsoft_storage_storageaccounts_transactions"
	metric2.fields["timeStamp"] = "2021-11-05T10:00:00Z"
	metric2.fields["total"] = 5.0
	metric2.fields["average"] = 4.0
	metric2.tags["subscription_id"] = "subscription_id"
	metric2.tags["resource_group"] = "azure-rg"
	metric2.tags["namespace"] = "Microsoft.Storage/storageAccounts"
	metric2.tags["resource_name"] = "azuresa1"
	metric2.tags["resource_region"] = "eastus"
	metric2.tags["unit"] = "Count"

	metrics = append(metrics, metric1, metric2)

	return metrics
}

func getTarget2Metrics() []*Metric {
	var metrics []*Metric

	metric := NewMetric()

	metric.name = "azure_monitor_microsoft_storage_storageaccounts_ingress"
	metric.fields["timeStamp"] = "2021-11-05T10:00:00Z"
	metric.fields["minimum"] = 125.0
	metric.fields["maximum"] = 250.0
	metric.tags["subscription_id"] = "subscription_id"
	metric.tags["resource_group"] = "azure-rg"
	metric.tags["namespace"] = "Microsoft.Storage/storageAccounts"
	metric.tags["resource_name"] = "azuresa2"
	metric.tags["resource_region"] = "eastus"
	metric.tags["unit"] = "Bytes"

	metrics = append(metrics, metric)

	return metrics
}

func TestGetAccessToken_Success(t *testing.T) {
	am := getAzureMonitor(
		"subscription_id",
		"client_id",
		"client_secret",
		"tenant_id",
		[]*Target{})

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"POST",
		fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/token", am.TenantID),
		httpmock.NewStringResponder(200, `{"access_token": "abc123456789", "expires_on": "1636548796"}`))

	err := am.getAccessToken()

	if err != nil {
		assert.Fail(t, "Got an error", err)
	}

	assert.Equal(t, "abc123456789", am.azureClient.accessToken)

	expiresOn, err := strconv.ParseInt("1636548796", 10, 64)

	if err != nil {
		assert.Fail(t, "Got an error", err)
	}

	assert.Equal(t, time.Unix(expiresOn, 0).UTC(), am.azureClient.accessTokenExpiresOn)
}

func TestRefreshAccessToken_AccessTokenRefreshed(t *testing.T) {
	am := getAzureMonitor(
		"subscription_id",
		"client_id",
		"client_secret",
		"tenant_id",
		[]*Target{})

	expiresOn, err := strconv.ParseInt("1636548796", 10, 64)

	if err != nil {
		assert.Fail(t, "Got an error", err)
	}

	am.azureClient.accessToken = "abc123456789"
	am.azureClient.accessTokenExpiresOn = time.Unix(expiresOn, 0).UTC()

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"POST",
		fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/token", am.TenantID),
		httpmock.NewStringResponder(200, `{"access_token": "123456789abc", "expires_on": "1736548796"}`))

	err = am.refreshAccessToken()

	if err != nil {
		assert.Fail(t, "Got an error", err)
	}

	assert.Equal(t, "123456789abc", am.azureClient.accessToken)

	expiresOn, err = strconv.ParseInt("1736548796", 10, 64)

	if err != nil {
		assert.Fail(t, "Got an error", err)
	}

	assert.Equal(t, time.Unix(expiresOn, 0).UTC(), am.azureClient.accessTokenExpiresOn)
}

func TestRefreshAccessToken_AccessTokenNotRefreshed(t *testing.T) {
	am := getAzureMonitor(
		"subscription_id",
		"client_id",
		"client_secret",
		"tenant_id",
		[]*Target{})

	expiresOn, err := strconv.ParseInt("1736548796", 10, 64)

	if err != nil {
		assert.Fail(t, "Got an error", err)
	}

	am.azureClient.accessToken = "abc123456789"
	am.azureClient.accessTokenExpiresOn = time.Unix(expiresOn, 0).UTC()

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"POST",
		fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/token", am.TenantID),
		httpmock.NewStringResponder(200, `{"access_token": "123456789abc", "expires_on": "1836548796"}`))

	err = am.refreshAccessToken()

	if err != nil {
		assert.Fail(t, "Got an error", err)
	}

	assert.Equal(t, "abc123456789", am.azureClient.accessToken)
}

func TestGetAllTargetsMetricsNames_Success(t *testing.T) {
	target1 := getTarget(
		"resourceGroups/azure-rg/providers/Microsoft.Storage/storageAccounts/azuresa1",
		[]string{"UsedCapacity", "Transactions"},
		[]string{"Total", "Average"})

	target2 := getTarget(
		"resourceGroups/azure-rg/providers/Microsoft.Storage/storageAccounts/azuresa2",
		[]string{},
		[]string{"Total", "Average"})

	am := getAzureMonitor(
		"subscription_id",
		"client_id",
		"client_secret",
		"tenant_id",
		[]*Target{target1, target2})

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		am.buildMetricDefinitionsApiURL(target1),
		httpmock.NewStringResponder(200, ``))

	httpmock.RegisterResponder(
		"GET",
		am.buildMetricDefinitionsApiURL(target2),
		httpmock.NewStringResponder(200, resourceMetricsDefinitionsBody))

	err := am.getAllTargetsMetricsNames()

	if err != nil {
		assert.Fail(t, "Got an error", err)
	}

	assert.Equal(t, 2, len(am.Targets[0].Metrics))
	assert.Equal(t, target1.Metrics, am.Targets[0].Metrics)

	assert.Equal(t, 3, len(am.Targets[1].Metrics))
	assert.Equal(t, []string{"UsedCapacity", "Transactions", "Ingress"}, am.Targets[1].Metrics)
}

func TestGetAllTargetsAggregation_Success(t *testing.T) {
	target1 := getTarget(
		"resourceGroups/azure-rg/providers/Microsoft.Storage/storageAccounts/azuresa1",
		[]string{"UsedCapacity", "Transactions"},
		[]string{"Total", "Average"})

	target2 := getTarget(
		"resourceGroups/azure-rg/providers/Microsoft.Storage/storageAccounts/azuresa2",
		[]string{"Ingress"},
		[]string{})

	am := getAzureMonitor(
		"subscription_id",
		"client_id",
		"client_secret",
		"tenant_id",
		[]*Target{target1, target2})

	am.getAllTargetsAggregation()

	assert.Equal(t, 2, len(am.Targets[0].Aggregation))
	assert.Equal(t, target1.Aggregation, am.Targets[0].Aggregation)

	assert.Equal(t, 5, len(am.Targets[1].Aggregation))
	assert.Equal(t, []string{"Total", "Count", "Average", "Minimum", "Maximum"}, am.Targets[1].Aggregation)
}

func TestInit_Success(t *testing.T) {
	target1 := getTarget(
		"resourceGroups/azure-rg/providers/Microsoft.Storage/storageAccounts/azuresa1",
		[]string{"UsedCapacity", "Transactions"},
		[]string{"Total", "Average"})

	target2 := getTarget(
		"resourceGroups/azure-rg/providers/Microsoft.Storage/storageAccounts/azuresa2",
		[]string{},
		[]string{})

	am := getAzureMonitor(
		"subscription_id",
		"client_id",
		"client_secret",
		"tenant_id",
		[]*Target{target1, target2})

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"POST",
		fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/token", am.TenantID),
		httpmock.NewStringResponder(200, `{"access_token": "abc123456789", "expires_on": "1636548796"}`))

	httpmock.RegisterResponder(
		"GET",
		am.buildMetricDefinitionsApiURL(target1),
		httpmock.NewStringResponder(200, ``))

	httpmock.RegisterResponder(
		"GET",
		am.buildMetricDefinitionsApiURL(target2),
		httpmock.NewStringResponder(200, resourceMetricsDefinitionsBody))

	err := am.Init()

	if err != nil {
		assert.Fail(t, "Got an error", err)
	}
}

func TestInit_NoSubscriptionID(t *testing.T) {
	target := getTarget(
		"resourceGroups/azure-rg/providers/Microsoft.Storage/storageAccounts/azuresa",
		[]string{"UsedCapacity", "Transactions"},
		[]string{"Total", "Average"})

	am := getAzureMonitor(
		"",
		"client_id",
		"client_secret",
		"tenant_id",
		[]*Target{target})

	err := am.Init()

	if err == nil {
		assert.Fail(t, "Did not get an error")
	}
}

func TestInit_NoClientID(t *testing.T) {
	target := getTarget(
		"resourceGroups/azure-rg/providers/Microsoft.Storage/storageAccounts/azuresa",
		[]string{"UsedCapacity", "Transactions"},
		[]string{"Total", "Average"})

	am := getAzureMonitor(
		"subscription_id",
		"",
		"client_secret",
		"tenant_id",
		[]*Target{target})

	err := am.Init()

	if err == nil {
		assert.Fail(t, "Did not get an error")
	}
}

func TestInit_NoClientSecret(t *testing.T) {
	target := getTarget(
		"resourceGroups/azure-rg/providers/Microsoft.Storage/storageAccounts/azuresa",
		[]string{"UsedCapacity", "Transactions"},
		[]string{"Total", "Average"})

	am := getAzureMonitor(
		"subscription_id",
		"client_id",
		"",
		"tenant_id",
		[]*Target{target})

	err := am.Init()

	if err == nil {
		assert.Fail(t, "Did not get an error")
	}
}

func TestInit_NoTenantID(t *testing.T) {
	target := getTarget(
		"resourceGroups/azure-rg/providers/Microsoft.Storage/storageAccounts/azuresa",
		[]string{"UsedCapacity", "Transactions"},
		[]string{"Total", "Average"})

	am := getAzureMonitor(
		"subscription_id",
		"client_id",
		"client_secret",
		"",
		[]*Target{target})

	err := am.Init()

	if err == nil {
		assert.Fail(t, "Did not get an error")
	}
}

func TestInit_NoTargets(t *testing.T) {
	am := getAzureMonitor(
		"subscription_id",
		"client_id",
		"client_secret",
		"tenant_id",
		[]*Target{})

	err := am.Init()

	if err == nil {
		assert.Fail(t, "Did not get an error")
	}
}

func TestInit_NoTargetResourceID(t *testing.T) {
	target := getTarget(
		"",
		[]string{"UsedCapacity", "Transactions"},
		[]string{"Total", "Average"})

	am := getAzureMonitor(
		"subscription_id",
		"client_id",
		"client_secret",
		"tenant_id",
		[]*Target{target})

	err := am.Init()

	if err == nil {
		assert.Fail(t, "Did not get an error")
	}
}

func TestGather_Success(t *testing.T) {
	target1 := getTarget(
		"resourceGroups/azure-rg/providers/Microsoft.Storage/storageAccounts/azuresa1",
		[]string{"UsedCapacity", "Transactions"},
		[]string{"Total", "Average"})

	target2 := getTarget(
		"resourceGroups/azure-rg/providers/Microsoft.Storage/storageAccounts/azuresa2",
		[]string{"Ingress"},
		[]string{"Minimum, Maximum"})

	am := getAzureMonitor(
		"subscription_id",
		"client_id",
		"client_secret",
		"tenant_id",
		[]*Target{target1, target2})

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"POST",
		fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/token", am.TenantID),
		httpmock.NewStringResponder(200, `{"access_token": "abc123456789", "expires_on": "1636548796"}`))

	httpmock.RegisterResponder(
		"GET",
		am.buildMetricValuesApiURL(target1),
		httpmock.NewStringResponder(200, target1MetricsValues))

	httpmock.RegisterResponder(
		"GET",
		am.buildMetricValuesApiURL(target2),
		httpmock.NewStringResponder(200, target2MetricsValues))

	acc := testutil.Accumulator{}
	err := acc.GatherError(am.Gather)

	if err != nil {
		assert.Fail(t, "Got an error", err)
	}

	target1Metrics := getTarget1Metrics()
	target2Metrics := getTarget2Metrics()

	acc.AssertContainsTaggedFields(t, target1Metrics[0].name, target1Metrics[0].fields, target1Metrics[0].tags)
	acc.AssertContainsTaggedFields(t, target1Metrics[1].name, target1Metrics[1].fields, target1Metrics[1].tags)

	acc.AssertContainsTaggedFields(t, target2Metrics[0].name, target2Metrics[0].fields, target2Metrics[0].tags)
}
