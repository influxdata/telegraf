package azure_monitor

import (
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/monitor/armmonitor"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetMetricName_Success(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: testSubscriptionID,
		ClientID:       testClientID,
		ClientSecret:   testClientSecret,
		TenantID:       testTenantID,
		ResourceTargets: []*ResourceTarget{
			newResourceTarget(testFullResourceGroup1ResourceType1Resource1, []string{testMetric1}, []string{string(armmonitor.AggregationTypeEnumTotal), string(armmonitor.AggregationTypeEnumAverage)}),
		},
		Log:          testutil.Logger{},
		azureClients: setMockAzureClients(),
	}

	response, err := am.azureClients.metricsClient.List(am.azureClients.ctx, am.ResourceTargets[0].ResourceID, nil)
	assert.NoError(t, err)

	metricName, err := createMetricName(response.Value[0], &response)
	require.NoError(t, err)
	require.NotNil(t, metricName)

	assert.Equal(t, "azure_monitor_microsoft_test_type1_metric1", *metricName)
}

func TestGetMetricFields_AllTimeseriesWithData(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: testSubscriptionID,
		ClientID:       testClientID,
		ClientSecret:   testClientSecret,
		TenantID:       testTenantID,
		ResourceTargets: []*ResourceTarget{
			newResourceTarget(testFullResourceGroup1ResourceType1Resource1, []string{testMetric1}, []string{string(armmonitor.AggregationTypeEnumTotal), string(armmonitor.AggregationTypeEnumMaximum)}),
		},
		Log:          testutil.Logger{},
		azureClients: setMockAzureClients(),
	}

	response, err := am.azureClients.metricsClient.List(am.azureClients.ctx, am.ResourceTargets[0].ResourceID, nil)
	assert.NoError(t, err)

	metricFields := getMetricFields(response.Value[0].Timeseries[0].Data)
	require.NotNil(t, metricFields)

	assert.Len(t, metricFields, 3)

	assert.Equal(t, "2022-02-22T22:59:00Z", metricFields["timeStamp"])
	assert.Equal(t, 5.0, metricFields["total"])
	assert.Equal(t, 5.0, metricFields["maximum"])
}

func TestGetMetricFields_LastTimeseriesWithoutData(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: testSubscriptionID,
		ClientID:       testClientID,
		ClientSecret:   testClientSecret,
		TenantID:       testTenantID,
		ResourceTargets: []*ResourceTarget{
			newResourceTarget(testFullResourceGroup2ResourceType1Resource3, []string{testMetric1}, []string{string(armmonitor.AggregationTypeEnumTotal), string(armmonitor.AggregationTypeEnumAverage)}),
		},
		Log:          testutil.Logger{},
		azureClients: setMockAzureClients(),
	}

	response, err := am.azureClients.metricsClient.List(am.azureClients.ctx, am.ResourceTargets[0].ResourceID, nil)
	assert.NoError(t, err)

	metricFields := getMetricFields(response.Value[0].Timeseries[0].Data)
	require.NotNil(t, metricFields)

	assert.Len(t, metricFields, 3)

	assert.Equal(t, "2022-02-22T22:58:00Z", metricFields["timeStamp"])
	assert.Equal(t, 2.5, metricFields["total"])
	assert.Equal(t, 2.5, metricFields["minimum"])
}

func TestGetMetricFields_AllTimeseriesWithoutData(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: testSubscriptionID,
		ClientID:       testClientID,
		ClientSecret:   testClientSecret,
		TenantID:       testTenantID,
		ResourceTargets: []*ResourceTarget{
			newResourceTarget(testFullResourceGroup2ResourceType2Resource4, []string{testMetric1}, []string{string(armmonitor.AggregationTypeEnumTotal), string(armmonitor.AggregationTypeEnumMaximum)}),
		},
		Log:          testutil.Logger{},
		azureClients: setMockAzureClients(),
	}

	response, err := am.azureClients.metricsClient.List(am.azureClients.ctx, am.ResourceTargets[0].ResourceID, nil)
	assert.NoError(t, err)

	metricFields := getMetricFields(response.Value[0].Timeseries[0].Data)
	require.Nil(t, metricFields)
}

func TestGetMetricFields_NoTimeseriesData(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: testSubscriptionID,
		ClientID:       testClientID,
		ClientSecret:   testClientSecret,
		TenantID:       testTenantID,
		ResourceTargets: []*ResourceTarget{
			newResourceTarget(testFullResourceGroup2ResourceType2Resource5, []string{testMetric1}, []string{string(armmonitor.AggregationTypeEnumTotal), string(armmonitor.AggregationTypeEnumMaximum)}),
		},
		Log:          testutil.Logger{},
		azureClients: setMockAzureClients(),
	}

	response, err := am.azureClients.metricsClient.List(am.azureClients.ctx, am.ResourceTargets[0].ResourceID, nil)
	assert.NoError(t, err)

	metricFields := getMetricFields(response.Value[0].Timeseries[0].Data)
	require.Nil(t, metricFields)
}

func TestGetMetricTags_Success(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: testSubscriptionID,
		ClientID:       testClientID,
		ClientSecret:   testClientSecret,
		TenantID:       testTenantID,
		ResourceTargets: []*ResourceTarget{
			newResourceTarget(testFullResourceGroup1ResourceType1Resource1, []string{testMetric1}, []string{string(armmonitor.AggregationTypeEnumTotal), string(armmonitor.AggregationTypeEnumMaximum)}),
		},
		Log:          testutil.Logger{},
		azureClients: &azureClients{metricsClient: &mockAzureMetricsClient{}},
	}

	response, err := am.azureClients.metricsClient.List(am.azureClients.ctx, am.ResourceTargets[0].ResourceID, nil)
	assert.NoError(t, err)

	metricTags, err := getMetricTags(response.Value[0], &response)
	require.NoError(t, err)
	require.NotNil(t, metricTags)

	assert.Len(t, metricTags, 6)

	assert.Equal(t, testSubscriptionID, metricTags[metricTagSubscriptionID])
	assert.Equal(t, testResourceGroup1, metricTags[metricTagResourceGroup])
	assert.Equal(t, "resource1", metricTags[metricTagResourceName])
	assert.Equal(t, testResourceType1, metricTags[metricTagNamespace])
	assert.Equal(t, testResourceRegion, metricTags[metricTagResourceRegion])
	assert.Equal(t, string(armmonitor.MetricUnitCount), metricTags[metricTagUnit])
}
