package azure_monitor

import (
	"fmt"
	"github.com/influxdata/telegraf/testutil"
	"strconv"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	resourceID1 = "resourceGroups/azure-rg1/providers/Microsoft.Storage/storageAccounts/azuresa1"
	resourceID2 = "resourceGroups/azure-rg1/providers/Microsoft.Storage/storageAccounts/azuresa2"
	resourceID3 = "resourceGroups/azure-rg2/providers/Microsoft.Storage/storageAccounts/azuresa3"
	resourceID4 = "resourceGroups/azure-rg2/providers/Microsoft.Storage/storageAccounts/azuresa4"
)

var subscriptionResourceGroupsBody = `
{
    "value": [
        {
            "id": "/subscriptions/subscription_id/resourceGroups/azure-rg1",
            "name": "azure-rg1",
            "location": "eastus",
            "properties": {
                "provisioningState": "Succeeded"
            }
        },
        {
            "id": "/subscriptions/subscription_id/resourceGroups/azure-rg2",
            "name": "azure-rg2",
            "location": "eastus",
            "properties": {
                "provisioningState": "Succeeded"
            }
        }
	]
}
`

// azure-rg1 resource group resources
var resourceGroup1ResourcesBody = `
{
    "value": [
        {
            "id": "/subscriptions/subscription_id/resourceGroups/azure-rg1/providers/Microsoft.Storage/storageAccounts/azuresa1",
            "name": "azuresa1",
            "type": "Microsoft.Storage/storageAccounts",
            "location": "eastus",
            "tags": {}
        },
        {
            "id": "/subscriptions/subscription_id/resourceGroups/azure-rg1/providers/Microsoft.Storage/storageAccounts/azuresa2",
            "name": "azuresa2",
            "type": "Microsoft.Storage/storageAccounts",
            "location": "eastus",
            "tags": {}
        }
	]
}
`

// azure-rg2 resource group resources
var resourceGroup2ResourcesBody = `
{
    "value": [
        {
            "id": "/subscriptions/subscription_id/resourceGroups/azure-rg2/providers/Microsoft.Storage/storageAccounts/azuresa3",
            "name": "azuresa3",
            "type": "Microsoft.Storage/storageAccounts",
            "location": "eastus",
            "tags": {}
        },
        {
            "id": "/subscriptions/subscription_id/resourceGroups/azure-rg2/providers/Microsoft.Storage/storageAccounts/azuresa4",
            "name": "azuresa4",
            "type": "Microsoft.Storage/storageAccounts",
            "location": "eastus",
            "tags": {}
        }
	]
}
`

var resourceTarget1MetricValues = `
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

var resourceTarget2MetricValues = `
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

var resourceMetricDefinitionsBody = `
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

var (
	am = &AzureMonitor{
		azureClient:    NewAzureClient(),
		SubscriptionID: "subscription_id",
		ClientID:       "client_id",
		ClientSecret:   "client_secret",
		TenantID:       "tenant_id",
		Log:            testutil.Logger{},
	}
	resource = &Resource{
		ResourceType: "Microsoft.Storage/storageAccounts",
		Metrics:      []string{"UsedCapacity"},
		Aggregation:  []string{"Total"},
	}
	subscriptionTargets  = getSubscriptionTargets()
	resourceGroupTargets = getResourceGroupTargets()
	resourceTarget1      = NewResourceTarget(
		"resourceGroups/azure-rg/providers/Microsoft.Storage/storageAccounts/azuresa1",
		[]string{"UsedCapacity", "Transactions"},
		[]string{"Total", "Average"},
	)
	resourceTarget2 = NewResourceTarget(
		"resourceGroups/azure-rg/providers/Microsoft.Storage/storageAccounts/azuresa2",
		[]string{},
		[]string{},
	)
	resourceTarget3 = NewResourceTarget(
		"resourceGroups/azure-rg/providers/Microsoft.Storage/storageAccounts/azuresa2",
		[]string{"Ingress"},
		[]string{"Minimum, Maximum"})
)

func resetAzureMonitor() {
	am.azureClient = NewAzureClient()
	am.ResourceTargets = make([]*ResourceTarget, 0)
	am.ResourceGroupTargets = make([]*ResourceGroupTarget, 0)
	am.SubscriptionTargets = make([]*Resource, 0)
}

func getSubscriptionTargets() []*Resource {
	subscriptionTargets := make([]*Resource, 0)
	subscriptionTargets = append(subscriptionTargets, resource)

	return subscriptionTargets
}

func getResourceGroupTargets() []*ResourceGroupTarget {
	resourceGroupTargets := make([]*ResourceGroupTarget, 0)
	resourceGroupTargets = append(resourceGroupTargets,
		NewResourceGroupTarget("azure-rg1", []*Resource{resource}),
		NewResourceGroupTarget("azure-rg2", []*Resource{resource}),
	)

	return resourceGroupTargets
}

func getResourceTarget1Metrics() []*Metric {
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

func getResourceTarget2Metrics() []*Metric {
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
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"POST",
		fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/token", am.TenantID),
		httpmock.NewStringResponder(200, `{"access_token": "abc123456789", "expires_on": "1636548796"}`))

	err := am.getAccessToken()

	require.NoError(t, err)
	assert.Equal(t, "abc123456789", am.azureClient.accessToken)

	expiresOn, err := strconv.ParseInt("1636548796", 10, 64)

	require.NoError(t, err)
	assert.Equal(t, time.Unix(expiresOn, 0).UTC(), am.azureClient.accessTokenExpiresOn)

	resetAzureMonitor()
}

func TestRefreshAccessToken_AccessTokenRefreshed(t *testing.T) {
	expiresOn, err := strconv.ParseInt("1636548796", 10, 64)

	require.NoError(t, err)

	am.azureClient.accessToken = "abc123456789"
	am.azureClient.accessTokenExpiresOn = time.Unix(expiresOn, 0).UTC()

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"POST",
		fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/token", am.TenantID),
		httpmock.NewStringResponder(200, `{"access_token": "123456789abc", "expires_on": "1736548796"}`))

	err = am.refreshAccessToken()

	require.NoError(t, err)
	assert.Equal(t, "123456789abc", am.azureClient.accessToken)

	expiresOn, err = strconv.ParseInt("1736548796", 10, 64)

	require.NoError(t, err)
	assert.Equal(t, time.Unix(expiresOn, 0).UTC(), am.azureClient.accessTokenExpiresOn)

	resetAzureMonitor()
}

func TestRefreshAccessToken_AccessTokenNotRefreshed(t *testing.T) {
	expiresOn, err := strconv.ParseInt("1736548796", 10, 64)

	require.NoError(t, err)

	am.azureClient.accessToken = "abc123456789"
	am.azureClient.accessTokenExpiresOn = time.Unix(expiresOn, 0).UTC()

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"POST",
		fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/token", am.TenantID),
		httpmock.NewStringResponder(200, `{"access_token": "123456789abc", "expires_on": "1836548796"}`))

	err = am.refreshAccessToken()

	require.NoError(t, err)
	assert.Equal(t, "abc123456789", am.azureClient.accessToken)

	resetAzureMonitor()
}

func TestCreateResourceGroupTargetsFromSubscriptionTargets_Success(t *testing.T) {
	am.SubscriptionTargets = subscriptionTargets

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		am.buildSubscriptionResourceGroupsAPIURL(),
		httpmock.NewStringResponder(200, subscriptionResourceGroupsBody))

	err := am.createResourceGroupTargetsFromSubscriptionTargets()

	require.NoError(t, err)

	assert.Equal(t, 2, len(am.ResourceGroupTargets))

	assert.Equal(t, "azure-rg1", am.ResourceGroupTargets[0].ResourceGroup)
	assert.Equal(t, 1, len(am.ResourceGroupTargets[0].Resources))
	assert.Equal(t, subscriptionTargets[0].ResourceType, am.ResourceGroupTargets[0].Resources[0].ResourceType)
	assert.Equal(t, subscriptionTargets[0].Metrics, am.ResourceGroupTargets[0].Resources[0].Metrics)
	assert.Equal(t, subscriptionTargets[0].Aggregation, am.ResourceGroupTargets[0].Resources[0].Aggregation)

	assert.Equal(t, "azure-rg2", am.ResourceGroupTargets[1].ResourceGroup)
	assert.Equal(t, 1, len(am.ResourceGroupTargets[1].Resources))
	assert.Equal(t, subscriptionTargets[0].ResourceType, am.ResourceGroupTargets[1].Resources[0].ResourceType)
	assert.Equal(t, subscriptionTargets[0].Metrics, am.ResourceGroupTargets[1].Resources[0].Metrics)
	assert.Equal(t, subscriptionTargets[0].Aggregation, am.ResourceGroupTargets[1].Resources[0].Aggregation)

	resetAzureMonitor()
}

func TestCreateResourceTargetsFromResourceGroupTargets_Success(t *testing.T) {
	am.ResourceGroupTargets = resourceGroupTargets

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		am.buildResourceGroupResourcesAPIURL(am.ResourceGroupTargets[0]),
		httpmock.NewStringResponder(200, resourceGroup1ResourcesBody))

	httpmock.RegisterResponder(
		"GET",
		am.buildResourceGroupResourcesAPIURL(am.ResourceGroupTargets[1]),
		httpmock.NewStringResponder(200, resourceGroup2ResourcesBody))

	err := am.createResourceTargetsFromResourceGroupTargets()

	require.NoError(t, err)

	assert.Equal(t, 4, len(am.ResourceTargets))

	for _, target := range am.ResourceTargets {
		if target.ResourceID == resourceID1 {
			assert.Equal(t, resourceGroupTargets[0].Resources[0].Metrics, target.Metrics)
			assert.Equal(t, resourceGroupTargets[0].Resources[0].Aggregation, target.Aggregation)
		} else if target.ResourceID == resourceID2 {
			assert.Equal(t, resourceGroupTargets[0].Resources[0].Metrics, target.Metrics)
			assert.Equal(t, resourceGroupTargets[0].Resources[0].Aggregation, target.Aggregation)
		} else if target.ResourceID == resourceID3 {
			assert.Equal(t, resourceGroupTargets[1].Resources[0].Metrics, target.Metrics)
			assert.Equal(t, resourceGroupTargets[1].Resources[0].Aggregation, target.Aggregation)
		} else if target.ResourceID == resourceID4 {
			assert.Equal(t, resourceGroupTargets[1].Resources[0].Metrics, target.Metrics)
			assert.Equal(t, resourceGroupTargets[1].Resources[0].Aggregation, target.Aggregation)
		} else {
			assert.FailNowf(t, "Did not get any expected resource id", "Test failed")
		}
	}

	resetAzureMonitor()
}

func TestGetResourceTargetsMetrics(t *testing.T) {
	am.ResourceTargets = append(am.ResourceTargets, resourceTarget1, resourceTarget2)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		am.buildMetricDefinitionsAPIURL(resourceTarget1),
		httpmock.NewStringResponder(200, resourceMetricDefinitionsBody))

	httpmock.RegisterResponder(
		"GET",
		am.buildMetricDefinitionsAPIURL(resourceTarget2),
		httpmock.NewStringResponder(200, resourceMetricDefinitionsBody))

	err := am.getResourceTargetsMetrics()

	require.NoError(t, err)

	assert.Equal(t, 2, len(am.ResourceTargets[0].Metrics))
	assert.Equal(t, resourceTarget1.Metrics, am.ResourceTargets[0].Metrics)

	assert.Equal(t, 3, len(am.ResourceTargets[1].Metrics))
	assert.Equal(t, []string{"UsedCapacity", "Transactions", "Ingress"}, am.ResourceTargets[1].Metrics)

	resetAzureMonitor()
}

func TestSetResourceTargetsAggregation_Success(t *testing.T) {
	am.ResourceTargets = append(am.ResourceTargets, resourceTarget1, resourceTarget2)
	am.setResourceTargetsAggregation()

	assert.Equal(t, 2, len(am.ResourceTargets[0].Aggregation))
	assert.Equal(t, resourceTarget1.Aggregation, am.ResourceTargets[0].Aggregation)

	assert.Equal(t, 5, len(am.ResourceTargets[1].Aggregation))
	assert.Equal(t, []string{"Total", "Count", "Average", "Minimum", "Maximum"}, am.ResourceTargets[1].Aggregation)

	resetAzureMonitor()
}

func TestInitOnlyResourceTargets_Success(t *testing.T) {
	am.ResourceTargets = append(am.ResourceTargets, resourceTarget1, resourceTarget2)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"POST",
		fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/token", am.TenantID),
		httpmock.NewStringResponder(200, `{"access_token": "abc123456789", "expires_on": "1636548796"}`))

	httpmock.RegisterResponder(
		"GET",
		am.buildMetricDefinitionsAPIURL(resourceTarget1),
		httpmock.NewStringResponder(200, resourceMetricDefinitionsBody))

	httpmock.RegisterResponder(
		"GET",
		am.buildMetricDefinitionsAPIURL(resourceTarget2),
		httpmock.NewStringResponder(200, resourceMetricDefinitionsBody))

	err := am.Init()

	require.NoError(t, err)
	assert.Equal(t, 2, len(am.ResourceTargets))

	resetAzureMonitor()
}

func TestInitOnlyResourceGroupTargets_Success(t *testing.T) {
	am.ResourceGroupTargets = resourceGroupTargets

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"POST",
		fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/token", am.TenantID),
		httpmock.NewStringResponder(200, `{"access_token": "abc123456789", "expires_on": "1636548796"}`))

	httpmock.RegisterResponder(
		"GET",
		am.buildResourceGroupResourcesAPIURL(am.ResourceGroupTargets[0]),
		httpmock.NewStringResponder(200, resourceGroup1ResourcesBody))

	httpmock.RegisterResponder(
		"GET",
		am.buildResourceGroupResourcesAPIURL(am.ResourceGroupTargets[1]),
		httpmock.NewStringResponder(200, resourceGroup2ResourcesBody))

	err := am.Init()

	require.NoError(t, err)
	assert.Equal(t, 4, len(am.ResourceTargets))

	resetAzureMonitor()
}

func TestInitOnlySubscriptionTargets_Success(t *testing.T) {
	am.SubscriptionTargets = subscriptionTargets

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"POST",
		fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/token", am.TenantID),
		httpmock.NewStringResponder(200, `{"access_token": "abc123456789", "expires_on": "1636548796"}`))

	httpmock.RegisterResponder(
		"GET",
		am.buildSubscriptionResourceGroupsAPIURL(),
		httpmock.NewStringResponder(200, subscriptionResourceGroupsBody))

	httpmock.RegisterResponder(
		"GET",
		am.buildResourceGroupResourcesAPIURL(resourceGroupTargets[0]),
		httpmock.NewStringResponder(200, resourceGroup1ResourcesBody))

	httpmock.RegisterResponder(
		"GET",
		am.buildResourceGroupResourcesAPIURL(resourceGroupTargets[1]),
		httpmock.NewStringResponder(200, resourceGroup2ResourcesBody))

	err := am.Init()

	require.NoError(t, err)
	assert.Equal(t, 4, len(am.ResourceTargets))

	resetAzureMonitor()
}

func TestInitAllTargetTypes_Success(t *testing.T) {
	am.SubscriptionTargets = subscriptionTargets
	am.ResourceGroupTargets = resourceGroupTargets
	am.ResourceTargets = append(am.ResourceTargets, resourceTarget1, resourceTarget2)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"POST",
		fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/token", am.TenantID),
		httpmock.NewStringResponder(200, `{"access_token": "abc123456789", "expires_on": "1636548796"}`))

	httpmock.RegisterResponder(
		"GET",
		am.buildSubscriptionResourceGroupsAPIURL(),
		httpmock.NewStringResponder(200, subscriptionResourceGroupsBody))

	httpmock.RegisterResponder(
		"GET",
		am.buildResourceGroupResourcesAPIURL(resourceGroupTargets[0]),
		httpmock.NewStringResponder(200, resourceGroup1ResourcesBody))

	httpmock.RegisterResponder(
		"GET",
		am.buildResourceGroupResourcesAPIURL(resourceGroupTargets[1]),
		httpmock.NewStringResponder(200, resourceGroup2ResourcesBody))

	httpmock.RegisterResponder(
		"GET",
		am.buildMetricDefinitionsAPIURL(resourceTarget2),
		httpmock.NewStringResponder(200, resourceMetricDefinitionsBody))

	err := am.Init()

	require.NoError(t, err)
	assert.Equal(t, 10, len(am.ResourceTargets))

	resetAzureMonitor()
}

func TestInit_NoSubscriptionID(t *testing.T) {
	am.SubscriptionID = ""
	err := am.Init()

	require.Error(t, err)

	resetAzureMonitor()
}

func TestInit_NoClientID(t *testing.T) {
	am.ClientID = ""
	err := am.Init()

	require.Error(t, err)

	resetAzureMonitor()
}

func TestInit_NoClientSecret(t *testing.T) {
	am.ClientSecret = ""
	err := am.Init()

	require.Error(t, err)

	resetAzureMonitor()
}

func TestInit_NoTenantID(t *testing.T) {
	am.TenantID = ""
	err := am.Init()

	require.Error(t, err)

	resetAzureMonitor()
}

func TestInit_NoTargets(t *testing.T) {
	err := am.Init()

	require.Error(t, err)

	resetAzureMonitor()
}

func TestInit_NoResourceTargetResourceID(t *testing.T) {
	am.ResourceTargets = append(am.ResourceTargets, NewResourceTarget("", []string{}, []string{}))
	err := am.Init()

	require.Error(t, err)

	resetAzureMonitor()
}

func TestInit_NoResourceGroupTargetResourceGroup(t *testing.T) {
	am.ResourceGroupTargets = append(am.ResourceGroupTargets, NewResourceGroupTarget("", []*Resource{}))
	err := am.Init()

	require.Error(t, err)

	resetAzureMonitor()
}

func TestInit_NoResourceGroupTargetResourceType(t *testing.T) {
	am.ResourceGroupTargets = append(am.ResourceGroupTargets, NewResourceGroupTarget("azure-rg1", []*Resource{}))
	err := am.Init()

	require.Error(t, err)

	resetAzureMonitor()
}

func TestInit_NoSubscriptionTargetResourceType(t *testing.T) {
	am.SubscriptionTargets = append(am.SubscriptionTargets, &Resource{})
	err := am.Init()

	require.Error(t, err)

	resetAzureMonitor()
}

func TestGather_Success(t *testing.T) {
	am.ResourceTargets = append(am.ResourceTargets, resourceTarget1, resourceTarget3)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"POST",
		fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/token", am.TenantID),
		httpmock.NewStringResponder(200, `{"access_token": "abc123456789", "expires_on": "1636548796"}`))

	httpmock.RegisterResponder(
		"GET",
		am.buildMetricValuesAPIURL(resourceTarget1),
		httpmock.NewStringResponder(200, resourceTarget1MetricValues))

	httpmock.RegisterResponder(
		"GET",
		am.buildMetricValuesAPIURL(resourceTarget3),
		httpmock.NewStringResponder(200, resourceTarget2MetricValues))

	acc := testutil.Accumulator{}
	err := acc.GatherError(am.Gather)

	require.NoError(t, err)

	resourceTarget1Metrics := getResourceTarget1Metrics()
	resourceTarget2Metrics := getResourceTarget2Metrics()

	acc.AssertContainsTaggedFields(t, resourceTarget1Metrics[0].name, resourceTarget1Metrics[0].fields, resourceTarget1Metrics[0].tags)
	acc.AssertContainsTaggedFields(t, resourceTarget1Metrics[1].name, resourceTarget1Metrics[1].fields, resourceTarget1Metrics[1].tags)

	acc.AssertContainsTaggedFields(t, resourceTarget2Metrics[0].name, resourceTarget2Metrics[0].fields, resourceTarget2Metrics[0].tags)

	resetAzureMonitor()
}
