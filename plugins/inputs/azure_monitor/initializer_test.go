package azure_monitor

import (
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/monitor/armmonitor"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckConfigValidation_ResourceTargetsOnly(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: testSubscriptionID,
		ClientID:       testClientID,
		ClientSecret:   testClientSecret,
		TenantID:       testTenantID,
		ResourceTargets: []*ResourceTarget{
			newResourceTarget(testResourceGroup1ResourceType1Resource1, []string{}, []string{}),
		},
		Log:          testutil.Logger{},
		azureClients: setMockAzureClients(),
	}

	err := am.checkConfigValidation()
	require.NoError(t, err)
}

func TestCheckConfigValidation_ResourceTargetWithNoResourceID(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: testSubscriptionID,
		ClientID:       testClientID,
		ClientSecret:   testClientSecret,
		TenantID:       testTenantID,
		ResourceTargets: []*ResourceTarget{
			newResourceTarget("", []string{}, []string{}),
		},
		Log:          testutil.Logger{},
		azureClients: setMockAzureClients(),
	}

	err := am.checkConfigValidation()
	require.Error(t, err)
}

func TestCheckConfigValidation_ResourceTargetWithInvalidAggregation(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: testSubscriptionID,
		ClientID:       testClientID,
		ClientSecret:   testClientSecret,
		TenantID:       testTenantID,
		ResourceTargets: []*ResourceTarget{
			newResourceTarget(testResourceGroup1ResourceType1Resource1, []string{}, []string{testInvalidAggregation}),
		},
		Log:          testutil.Logger{},
		azureClients: setMockAzureClients(),
	}

	err := am.checkConfigValidation()
	require.Error(t, err)
}

func TestCheckConfigValidation_ResourceGroupTargetsOnly(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: testSubscriptionID,
		ClientID:       testClientID,
		ClientSecret:   testClientSecret,
		TenantID:       testTenantID,
		ResourceGroupTargets: []*ResourceGroupTarget{
			{
				ResourceGroup: testResourceGroup1,
				Resources: []*Resource{
					{
						ResourceType: testResourceType1,
						Metrics:      []string{},
						Aggregations: []string{},
					},
				},
			},
		},
		Log:          testutil.Logger{},
		azureClients: setMockAzureClients(),
	}

	err := am.checkConfigValidation()
	require.NoError(t, err)
}

func TestCheckConfigValidation_ResourceGroupTargetWithoutResourceGroup(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: testSubscriptionID,
		ClientID:       testClientID,
		ClientSecret:   testClientSecret,
		TenantID:       testTenantID,
		ResourceGroupTargets: []*ResourceGroupTarget{
			{
				ResourceGroup: "",
				Resources: []*Resource{
					{
						ResourceType: testResourceType1,
						Metrics:      []string{},
						Aggregations: []string{},
					},
				},
			},
		},
		Log:          testutil.Logger{},
		azureClients: setMockAzureClients(),
	}

	err := am.checkConfigValidation()
	require.Error(t, err)
}

func TestCheckConfigValidation_ResourceGroupTargetWithoutResources(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: testSubscriptionID,
		ClientID:       testClientID,
		ClientSecret:   testClientSecret,
		TenantID:       testTenantID,
		ResourceGroupTargets: []*ResourceGroupTarget{
			{
				ResourceGroup: testResourceGroup1,
				Resources:     []*Resource{},
			},
		},
		Log:          testutil.Logger{},
		azureClients: setMockAzureClients(),
	}

	err := am.checkConfigValidation()
	require.Error(t, err)
}

func TestCheckConfigValidation_ResourceGroupTargetWithResourceWithoutResourceType(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: testSubscriptionID,
		ClientID:       testClientID,
		ClientSecret:   testClientSecret,
		TenantID:       testTenantID,
		ResourceGroupTargets: []*ResourceGroupTarget{
			{
				ResourceGroup: testResourceGroup1,
				Resources: []*Resource{
					{
						ResourceType: "",
						Metrics:      []string{},
						Aggregations: []string{},
					},
				},
			},
		},
		Log:          testutil.Logger{},
		azureClients: setMockAzureClients(),
	}

	err := am.checkConfigValidation()
	require.Error(t, err)
}

func TestCheckConfigValidation_ResourceGroupTargetWithInvalidAggregation(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: testSubscriptionID,
		ClientID:       testClientID,
		ClientSecret:   testClientSecret,
		TenantID:       testTenantID,
		ResourceGroupTargets: []*ResourceGroupTarget{
			{
				ResourceGroup: testResourceGroup1,
				Resources: []*Resource{
					{
						ResourceType: testResourceType1,
						Metrics:      []string{},
						Aggregations: []string{testInvalidAggregation},
					},
				},
			},
		},
		Log:          testutil.Logger{},
		azureClients: setMockAzureClients(),
	}

	err := am.checkConfigValidation()
	require.Error(t, err)
}

func TestCheckConfigValidation_SubscriptionTargetsOnly(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: testSubscriptionID,
		ClientID:       testClientID,
		ClientSecret:   testClientSecret,
		TenantID:       testTenantID,
		SubscriptionTargets: []*Resource{
			{
				ResourceType: testResourceType1,
				Metrics:      []string{},
				Aggregations: []string{},
			},
		},
		Log:          testutil.Logger{},
		azureClients: setMockAzureClients(),
	}

	err := am.checkConfigValidation()
	require.NoError(t, err)
}

func TestCheckConfigValidation_SubscriptionTargetWithoutResourceType(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: testSubscriptionID,
		ClientID:       testClientID,
		ClientSecret:   testClientSecret,
		TenantID:       testTenantID,
		SubscriptionTargets: []*Resource{
			{
				ResourceType: "",
				Metrics:      []string{},
				Aggregations: []string{},
			},
		},
		Log:          testutil.Logger{},
		azureClients: setMockAzureClients(),
	}

	err := am.checkConfigValidation()
	require.Error(t, err)
}

func TestCheckConfigValidation_SubscriptionTargetWithInvalidAggregation(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: testSubscriptionID,
		ClientID:       testClientID,
		ClientSecret:   testClientSecret,
		TenantID:       testTenantID,
		SubscriptionTargets: []*Resource{
			{
				ResourceType: testResourceType1,
				Metrics:      []string{},
				Aggregations: []string{testInvalidAggregation},
			},
		},
		Log:          testutil.Logger{},
		azureClients: setMockAzureClients(),
	}

	err := am.checkConfigValidation()
	require.Error(t, err)
}

func TestCheckConfigValidation_AllTargetTypes(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: testSubscriptionID,
		ClientID:       testClientID,
		ClientSecret:   testClientSecret,
		TenantID:       testTenantID,
		ResourceTargets: []*ResourceTarget{
			newResourceTarget(testResourceGroup1ResourceType1Resource1, []string{}, []string{}),
		},
		ResourceGroupTargets: []*ResourceGroupTarget{
			{
				ResourceGroup: testResourceGroup1,
				Resources: []*Resource{
					{
						ResourceType: testResourceType1,
						Metrics:      []string{},
						Aggregations: []string{},
					},
				},
			}},
		SubscriptionTargets: []*Resource{
			{
				ResourceType: testResourceType1,
				Metrics:      []string{},
				Aggregations: []string{},
			},
		},
		Log:          testutil.Logger{},
		azureClients: setMockAzureClients(),
	}

	err := am.checkConfigValidation()
	require.NoError(t, err)
}

func TestCheckConfigValidation_NoSubscriptionID(t *testing.T) {
	am := &AzureMonitor{
		ClientID:     testClientID,
		ClientSecret: testClientSecret,
		TenantID:     testTenantID,
		ResourceTargets: []*ResourceTarget{
			newResourceTarget(testResourceGroup1ResourceType1Resource1, []string{}, []string{}),
		},
		Log:          testutil.Logger{},
		azureClients: setMockAzureClients(),
	}

	err := am.checkConfigValidation()
	require.Error(t, err)
}

func TestCheckConfigValidation_NoClientID(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: testSubscriptionID,
		ClientSecret:   testClientSecret,
		TenantID:       testTenantID,
		ResourceTargets: []*ResourceTarget{
			newResourceTarget(testResourceGroup1ResourceType1Resource1, []string{}, []string{}),
		},
		Log:          testutil.Logger{},
		azureClients: setMockAzureClients(),
	}

	err := am.checkConfigValidation()
	require.Error(t, err)
}

func TestCheckConfigValidation_NoClientSecret(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: testSubscriptionID,
		ClientID:       testClientID,
		TenantID:       testTenantID,
		ResourceTargets: []*ResourceTarget{
			newResourceTarget(testResourceGroup1ResourceType1Resource1, []string{}, []string{}),
		},
		Log:          testutil.Logger{},
		azureClients: setMockAzureClients(),
	}

	err := am.checkConfigValidation()
	require.Error(t, err)
}

func TestCheckConfigValidation_NoTenantID(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: testSubscriptionID,
		ClientID:       testClientID,
		ClientSecret:   testClientSecret,
		ResourceTargets: []*ResourceTarget{
			newResourceTarget(testResourceGroup1ResourceType1Resource1, []string{}, []string{}),
		},
		Log:          testutil.Logger{},
		azureClients: setMockAzureClients(),
	}

	err := am.checkConfigValidation()
	require.Error(t, err)
}

func TestCheckConfigValidation_NoTargets(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: testSubscriptionID,
		ClientID:       testClientID,
		ClientSecret:   testClientSecret,
		TenantID:       testTenantID,
		Log:            testutil.Logger{},
		azureClients:   setMockAzureClients(),
	}

	err := am.checkConfigValidation()
	require.Error(t, err)
}

func TestAddPrefixToResourceTargetsResourceID_Success(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: testSubscriptionID,
		ClientID:       testClientID,
		ClientSecret:   testClientSecret,
		TenantID:       testTenantID,
		ResourceTargets: []*ResourceTarget{
			newResourceTarget(testResourceGroup1ResourceType1Resource1, []string{}, []string{}),
			newResourceTarget(testResourceGroup1ResourceType2Resource2, []string{}, []string{}),
		},
		Log:          testutil.Logger{},
		azureClients: setMockAzureClients(),
	}

	am.addPrefixToResourceTargetsResourceID()

	assert.Equal(t, testFullResourceGroup1ResourceType1Resource1, am.ResourceTargets[0].ResourceID)
	assert.Equal(t, testFullResourceGroup1ResourceType2Resource2, am.ResourceTargets[1].ResourceID)
}

func TestCreateResourceTargetsFromResourceGroupTargets_Success(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: testSubscriptionID,
		ClientID:       testClientID,
		ClientSecret:   testClientSecret,
		TenantID:       testTenantID,
		ResourceGroupTargets: []*ResourceGroupTarget{
			{
				ResourceGroup: testResourceGroup1,
				Resources: []*Resource{
					{
						ResourceType: testResourceType1,
						Metrics:      []string{testMetric1, testMetric2},
						Aggregations: []string{string(armmonitor.AggregationTypeEnumTotal)},
					},
					{
						ResourceType: testResourceType2,
						Metrics:      []string{testMetric1},
						Aggregations: []string{},
					},
				},
			},
			{
				ResourceGroup: testResourceGroup2,
				Resources: []*Resource{
					{
						ResourceType: testResourceType1,
						Metrics:      []string{testMetric3},
						Aggregations: []string{string(armmonitor.AggregationTypeEnumTotal), string(armmonitor.AggregationTypeEnumAverage)},
					},
				},
			},
		},
		Log:          testutil.Logger{},
		azureClients: setMockAzureClients(),
	}

	err := am.createResourceTargetsFromResourceGroupTargets()
	require.NoError(t, err)

	assert.Len(t, am.ResourceTargets, 3)

	for _, target := range am.ResourceTargets {
		assert.Contains(t, []string{testFullResourceGroup1ResourceType1Resource1, testFullResourceGroup1ResourceType2Resource2, testFullResourceGroup2ResourceType1Resource3}, target.ResourceID)

		if target.ResourceID == testFullResourceGroup1ResourceType1Resource1 {
			assert.Equal(t, am.ResourceGroupTargets[0].Resources[0].Metrics, target.Metrics)
			assert.Equal(t, am.ResourceGroupTargets[0].Resources[0].Aggregations, target.Aggregations)
		}
		if target.ResourceID == testFullResourceGroup1ResourceType2Resource2 {
			assert.Equal(t, am.ResourceGroupTargets[0].Resources[1].Metrics, target.Metrics)
			assert.Equal(t, am.ResourceGroupTargets[0].Resources[1].Aggregations, target.Aggregations)
		}
		if target.ResourceID == testFullResourceGroup2ResourceType1Resource3 {
			assert.Equal(t, am.ResourceGroupTargets[1].Resources[0].Metrics, target.Metrics)
			assert.Equal(t, am.ResourceGroupTargets[1].Resources[0].Aggregations, target.Aggregations)
		}
	}
}

func TestCreateResourceTargetsFromResourceGroupTargets_NoResourceFound(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: testSubscriptionID,
		ClientID:       testClientID,
		ClientSecret:   testClientSecret,
		TenantID:       testTenantID,
		ResourceGroupTargets: []*ResourceGroupTarget{
			{
				testResourceGroup2,
				[]*Resource{
					{
						ResourceType: testResourceType2,
						Metrics:      []string{testMetric1, testMetric2},
						Aggregations: []string{string(armmonitor.AggregationTypeEnumTotal)},
					},
				},
			},
		},
		Log:          testutil.Logger{},
		azureClients: setMockAzureClients(),
	}

	err := am.createResourceTargetsFromResourceGroupTargets()
	require.Error(t, err)
}

func TestCreateResourceTargetsFromSubscriptionTargets_Success(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: testSubscriptionID,
		ClientID:       testClientID,
		ClientSecret:   testClientSecret,
		TenantID:       testTenantID,
		SubscriptionTargets: []*Resource{
			{
				ResourceType: testResourceType1,
				Metrics:      []string{testMetric1, testMetric2},
				Aggregations: []string{string(armmonitor.AggregationTypeEnumTotal)},
			},
			{
				ResourceType: testResourceType2,
				Metrics:      []string{testMetric3},
				Aggregations: []string{string(armmonitor.AggregationTypeEnumTotal), string(armmonitor.AggregationTypeAverage)},
			},
		},
		Log:          testutil.Logger{},
		azureClients: setMockAzureClients(),
	}

	err := am.createResourceTargetsFromSubscriptionTargets()
	require.NoError(t, err)

	assert.Len(t, am.ResourceTargets, 3)

	for _, target := range am.ResourceTargets {
		assert.Contains(t, []string{testFullResourceGroup1ResourceType1Resource1, testFullResourceGroup1ResourceType2Resource2, testFullResourceGroup2ResourceType1Resource3}, target.ResourceID)

		if target.ResourceID == testFullResourceGroup1ResourceType1Resource1 {
			assert.Equal(t, am.SubscriptionTargets[0].Metrics, target.Metrics)
			assert.Equal(t, am.SubscriptionTargets[0].Aggregations, target.Aggregations)
		}
		if target.ResourceID == testFullResourceGroup1ResourceType2Resource2 {
			assert.Equal(t, am.SubscriptionTargets[1].Metrics, target.Metrics)
			assert.Equal(t, am.SubscriptionTargets[1].Aggregations, target.Aggregations)
		}
		if target.ResourceID == testFullResourceGroup2ResourceType1Resource3 {
			assert.Equal(t, am.SubscriptionTargets[0].Metrics, target.Metrics)
			assert.Equal(t, am.SubscriptionTargets[0].Aggregations, target.Aggregations)
		}
	}
}

func TestCreateResourceTargetsFromSubscriptionTargets_NoResourceFound(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: testSubscriptionID,
		ClientID:       testClientID,
		ClientSecret:   testClientSecret,
		TenantID:       testTenantID,
		SubscriptionTargets: []*Resource{
			{
				ResourceType: testResourceType3,
				Metrics:      []string{testMetric1, testMetric2},
				Aggregations: []string{string(armmonitor.AggregationTypeEnumTotal)},
			},
		},
		Log:          testutil.Logger{},
		azureClients: setMockAzureClients(),
	}

	err := am.createResourceTargetsFromSubscriptionTargets()
	require.Error(t, err)
}

func TestCheckResourceTargetsMetricsValidation_Success(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: testSubscriptionID,
		ClientID:       testClientID,
		ClientSecret:   testClientSecret,
		TenantID:       testTenantID,
		ResourceTargets: []*ResourceTarget{
			newResourceTarget(testFullResourceGroup1ResourceType1Resource1, []string{}, []string{}),
			newResourceTarget(testFullResourceGroup1ResourceType2Resource2, []string{testMetric1, testMetric2}, []string{}),
			newResourceTarget(testFullResourceGroup2ResourceType1Resource3, []string{testMetric1}, []string{string(armmonitor.AggregationTypeEnumTotal)}),
		},
		Log:          testutil.Logger{},
		azureClients: setMockAzureClients(),
	}

	err := am.checkResourceTargetsMetricsValidation()
	require.NoError(t, err)
}

func TestCheckResourceTargetsMetricsValidation_WithResourceTargetWithInvalidMetric(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: testSubscriptionID,
		ClientID:       testClientID,
		ClientSecret:   testClientSecret,
		TenantID:       testTenantID,
		ResourceTargets: []*ResourceTarget{
			newResourceTarget(testFullResourceGroup1ResourceType1Resource1, []string{}, []string{}),
			newResourceTarget(testFullResourceGroup1ResourceType2Resource2, []string{testMetric1, testMetric2}, []string{}),
			newResourceTarget(testFullResourceGroup2ResourceType1Resource3, []string{testInvalidMetric}, []string{string(armmonitor.AggregationTypeEnumTotal)}),
		},
		Log:          testutil.Logger{},
		azureClients: setMockAzureClients(),
	}

	err := am.checkResourceTargetsMetricsValidation()
	require.Error(t, err)
}

func TestSetResourceTargetsMetrics_Success(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: testSubscriptionID,
		ClientID:       testClientID,
		ClientSecret:   testClientSecret,
		TenantID:       testTenantID,
		ResourceTargets: []*ResourceTarget{
			newResourceTarget(testFullResourceGroup1ResourceType1Resource1, []string{}, []string{}),
			newResourceTarget(testFullResourceGroup1ResourceType2Resource2, []string{}, []string{string(armmonitor.AggregationTypeAverage)}),
			newResourceTarget(testFullResourceGroup2ResourceType1Resource3, []string{testMetric1}, []string{string(armmonitor.AggregationTypeEnumTotal)}),
		},
		Log:          testutil.Logger{},
		azureClients: setMockAzureClients(),
	}

	err := am.setResourceTargetsMetrics()
	require.NoError(t, err)

	assert.Len(t, am.ResourceTargets, 3)

	for _, target := range am.ResourceTargets {
		assert.Contains(t, []string{testFullResourceGroup1ResourceType1Resource1, testFullResourceGroup1ResourceType2Resource2, testFullResourceGroup2ResourceType1Resource3}, target.ResourceID)

		if target.ResourceID == testFullResourceGroup1ResourceType1Resource1 {
			assert.Len(t, target.Metrics, 3)
			assert.Equal(t, []string{testMetric1, testMetric2, testMetric3}, target.Metrics)
		}
		if target.ResourceID == testFullResourceGroup1ResourceType2Resource2 {
			assert.Len(t, target.Metrics, 2)
			assert.Equal(t, []string{testMetric1, testMetric2}, target.Metrics)
		}
		if target.ResourceID == testFullResourceGroup2ResourceType1Resource3 {
			assert.Len(t, target.Metrics, 1)
			assert.Equal(t, []string{testMetric1}, target.Metrics)
		}
	}
}

func TestCheckResourceTargetsMetricsMinTimeGrain_Success(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: testSubscriptionID,
		ClientID:       testClientID,
		ClientSecret:   testClientSecret,
		TenantID:       testTenantID,
		ResourceTargets: []*ResourceTarget{
			newResourceTarget(testFullResourceGroup1ResourceType1Resource1, []string{testMetric1, testMetric2, testMetric3}, []string{string(armmonitor.AggregationTypeEnumTotal), string(armmonitor.AggregationTypeEnumAverage)}),
			newResourceTarget(testFullResourceGroup1ResourceType2Resource2, []string{testMetric1, testMetric2}, []string{string(armmonitor.AggregationTypeEnumAverage)}),
			newResourceTarget(testFullResourceGroup2ResourceType1Resource3, []string{testMetric1}, []string{string(armmonitor.AggregationTypeEnumTotal)}),
		},
		Log:          testutil.Logger{},
		azureClients: setMockAzureClients(),
	}

	err := am.checkResourceTargetsMetricsMinTimeGrain()
	require.NoError(t, err)

	assert.Len(t, am.ResourceTargets, 4)

	for _, target := range am.ResourceTargets {
		if target.ResourceID == testFullResourceGroup1ResourceType1Resource1 {
			assert.Contains(t, []int{1, 2}, len(target.Metrics))

			if len(target.Metrics) == 1 {
				assert.Equal(t, []string{testMetric3}, target.Metrics)
				assert.Equal(t, []string{string(armmonitor.AggregationTypeEnumTotal), string(armmonitor.AggregationTypeEnumAverage)}, target.Aggregations)
			}
			if len(target.Metrics) == 2 {
				assert.Equal(t, []string{testMetric1, testMetric2}, target.Metrics)
				assert.Equal(t, []string{string(armmonitor.AggregationTypeEnumTotal), string(armmonitor.AggregationTypeEnumAverage)}, target.Aggregations)
			}
		}
		if target.ResourceID == testFullResourceGroup1ResourceType2Resource2 {
			assert.Equal(t, []string{testMetric1, testMetric2}, target.Metrics)
			assert.Equal(t, []string{string(armmonitor.AggregationTypeEnumAverage)}, target.Aggregations)
		}
		if target.ResourceID == testFullResourceGroup2ResourceType1Resource3 {
			assert.Equal(t, []string{testMetric1}, target.Metrics)
			assert.Equal(t, []string{string(armmonitor.AggregationTypeEnumTotal)}, target.Aggregations)
		}
	}
}

func TestCheckResourceTargetsMaxMetrics_Success(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: testSubscriptionID,
		ClientID:       testClientID,
		ClientSecret:   testClientSecret,
		TenantID:       testTenantID,
		ResourceTargets: []*ResourceTarget{
			newResourceTarget(testFullResourceGroup1ResourceType1Resource1, []string{}, []string{string(armmonitor.AggregationTypeEnumTotal), string(armmonitor.AggregationTypeEnumAverage)}),
			newResourceTarget(testFullResourceGroup1ResourceType2Resource2, []string{testMetric1, testMetric2}, []string{string(armmonitor.AggregationTypeEnumAverage)}),
			newResourceTarget(testFullResourceGroup2ResourceType1Resource3, []string{testMetric1}, []string{string(armmonitor.AggregationTypeEnumTotal)}),
		},
		Log:          testutil.Logger{},
		azureClients: setMockAzureClients(),
	}

	for index := 1; index <= 25; index++ {
		am.ResourceTargets[0].Metrics = append(am.ResourceTargets[0].Metrics, testMetric1)
	}

	expectedResource1Metrics := make([]string, 0)
	for index := 1; index <= maxMetricsPerRequest; index++ {
		expectedResource1Metrics = append(expectedResource1Metrics, testMetric1)
	}

	am.checkResourceTargetsMaxMetrics()

	assert.Len(t, am.ResourceTargets, 4)

	for _, target := range am.ResourceTargets {
		assert.Contains(t, []string{testFullResourceGroup1ResourceType1Resource1, testFullResourceGroup1ResourceType2Resource2, testFullResourceGroup2ResourceType1Resource3}, target.ResourceID)

		if target.ResourceID == testFullResourceGroup1ResourceType1Resource1 {
			assert.Contains(t, []int{5, maxMetricsPerRequest}, len(target.Metrics))

			if len(target.Metrics) == maxMetricsPerRequest {
				assert.Equal(t, expectedResource1Metrics, target.Metrics)
				assert.Equal(t, []string{string(armmonitor.AggregationTypeEnumTotal), string(armmonitor.AggregationTypeEnumAverage)}, target.Aggregations)
			}
			if len(target.Metrics) == 5 {
				assert.Equal(t, []string{testMetric1, testMetric1, testMetric1, testMetric1, testMetric1}, target.Metrics)
				assert.Equal(t, []string{string(armmonitor.AggregationTypeEnumTotal), string(armmonitor.AggregationTypeEnumAverage)}, target.Aggregations)
			}
		}
		if target.ResourceID == testFullResourceGroup1ResourceType2Resource2 {
			assert.Equal(t, []string{testMetric1, testMetric2}, target.Metrics)
			assert.Equal(t, []string{string(armmonitor.AggregationTypeEnumAverage)}, target.Aggregations)
		}
		if target.ResourceID == testFullResourceGroup2ResourceType1Resource3 {
			assert.Equal(t, []string{testMetric1}, target.Metrics)
			assert.Equal(t, []string{string(armmonitor.AggregationTypeEnumTotal)}, target.Aggregations)
		}
	}
}

func TestChangeResourceTargetsMetricsWithComma(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: testSubscriptionID,
		ClientID:       testClientID,
		ClientSecret:   testClientSecret,
		TenantID:       testTenantID,
		ResourceTargets: []*ResourceTarget{
			newResourceTarget(testFullResourceGroup1ResourceType1Resource1, []string{testMetric1, testMetric2, testMetric3WithComma}, []string{string(armmonitor.AggregationTypeEnumTotal), string(armmonitor.AggregationTypeEnumAverage)}),
			newResourceTarget(testFullResourceGroup1ResourceType2Resource2, []string{testMetric1, testMetric2}, []string{string(armmonitor.AggregationTypeEnumAverage)}),
			newResourceTarget(testFullResourceGroup2ResourceType1Resource3, []string{testMetric1}, []string{string(armmonitor.AggregationTypeEnumTotal)}),
		},
		Log:          testutil.Logger{},
		azureClients: setMockAzureClients(),
	}

	am.changeResourceTargetsMetricsWithComma()

	assert.Len(t, am.ResourceTargets, 3)

	for _, target := range am.ResourceTargets {
		assert.Contains(t, []string{testFullResourceGroup1ResourceType1Resource1, testFullResourceGroup1ResourceType2Resource2, testFullResourceGroup2ResourceType1Resource3}, target.ResourceID)

		if target.ResourceID == testFullResourceGroup1ResourceType1Resource1 {
			assert.Equal(t, []string{testMetric1, testMetric2, testMetric3ChangedComma}, target.Metrics)
			assert.Equal(t, []string{string(armmonitor.AggregationTypeEnumTotal), string(armmonitor.AggregationTypeEnumAverage)}, target.Aggregations)
		}
		if target.ResourceID == testFullResourceGroup1ResourceType2Resource2 {
			assert.Equal(t, []string{testMetric1, testMetric2}, target.Metrics)
			assert.Equal(t, []string{string(armmonitor.AggregationTypeEnumAverage)}, target.Aggregations)
		}
		if target.ResourceID == testFullResourceGroup2ResourceType1Resource3 {
			assert.Equal(t, []string{testMetric1}, target.Metrics)
			assert.Equal(t, []string{string(armmonitor.AggregationTypeEnumTotal)}, target.Aggregations)
		}
	}
}

func TestSetResourceTargetsAggregations_Success(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: testSubscriptionID,
		ClientID:       testClientID,
		ClientSecret:   testClientSecret,
		TenantID:       testTenantID,
		ResourceTargets: []*ResourceTarget{
			newResourceTarget(testFullResourceGroup1ResourceType1Resource1, []string{testMetric1, testMetric2, testMetric3}, []string{}),
			newResourceTarget(testFullResourceGroup1ResourceType2Resource2, []string{testMetric1, testMetric2}, []string{string(armmonitor.AggregationTypeEnumAverage)}),
			newResourceTarget(testFullResourceGroup2ResourceType1Resource3, []string{testMetric1}, []string{string(armmonitor.AggregationTypeEnumTotal), string(armmonitor.AggregationTypeEnumAverage)}),
		},
		Log:          testutil.Logger{},
		azureClients: setMockAzureClients(),
	}

	am.setResourceTargetsAggregations()

	for _, target := range am.ResourceTargets {
		assert.Contains(t, []string{testFullResourceGroup1ResourceType1Resource1, testFullResourceGroup1ResourceType2Resource2, testFullResourceGroup2ResourceType1Resource3}, target.ResourceID)

		if target.ResourceID == testFullResourceGroup1ResourceType1Resource1 {
			assert.Equal(t, []string{testMetric1, testMetric2, testMetric3}, target.Metrics)
			assert.Equal(t, []string{string(armmonitor.AggregationTypeEnumAverage), string(armmonitor.AggregationTypeEnumCount), string(armmonitor.AggregationTypeEnumMaximum), string(armmonitor.AggregationTypeEnumMinimum), string(armmonitor.AggregationTypeEnumTotal)}, target.Aggregations)
		}
		if target.ResourceID == testFullResourceGroup1ResourceType2Resource2 {
			assert.Equal(t, []string{testMetric1, testMetric2}, target.Metrics)
			assert.Equal(t, []string{string(armmonitor.AggregationTypeEnumAverage)}, target.Aggregations)
		}
		if target.ResourceID == testFullResourceGroup2ResourceType1Resource3 {
			assert.Equal(t, []string{testMetric1}, target.Metrics)
			assert.Equal(t, []string{string(armmonitor.AggregationTypeEnumTotal), string(armmonitor.AggregationTypeEnumAverage)}, target.Aggregations)
		}
	}
}
