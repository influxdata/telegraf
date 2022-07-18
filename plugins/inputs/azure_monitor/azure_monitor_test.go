package azure_monitor

import (
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/monitor/armmonitor"
	"github.com/influxdata/telegraf/testutil"
	receiver "github.com/logzio/azure-monitor-metrics-receiver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInit_ResourceTargetsOnly(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: testSubscriptionID,
		ClientID:       testClientID,
		ClientSecret:   testClientSecret,
		TenantID:       testTenantID,
		ResourceTargets: []*ResourceTarget{
			{
				testResourceGroup1ResourceType1Resource1,
				[]string{testMetric1, testMetric2, testMetric3},
				[]string{string(armmonitor.AggregationTypeEnumTotal), string(armmonitor.AggregationTypeEnumAverage)},
			},
			{
				testResourceGroup1ResourceType2Resource2,
				[]string{testMetric1, testMetric2},
				[]string{},
			},
			{
				testResourceGroup2ResourceType1Resource3,
				[]string{},
				[]string{string(armmonitor.AggregationTypeEnumTotal)},
			},
			{
				testResourceGroup1ResourceType2Resource2,
				[]string{},
				[]string{},
			},
			{
				testResourceGroup1ResourceType1Resource1,
				[]string{},
				[]string{},
			},
		},
		Log:          testutil.Logger{},
		azureClients: setMockAzureClients(),
	}

	for index := 1; index <= 27; index++ {
		if index <= 10 {
			am.ResourceTargets[4].Metrics = append(am.ResourceTargets[4].Metrics, testMetric1)
		} else if index <= 23 {
			am.ResourceTargets[4].Metrics = append(am.ResourceTargets[4].Metrics, testMetric2)
		} else {
			am.ResourceTargets[4].Metrics = append(am.ResourceTargets[4].Metrics, testMetric3)
		}
	}

	expectedResourceMetrics := make([]string, 0)
	for index := 1; index <= receiver.MaxMetricsPerRequest; index++ {
		if index <= 10 {
			expectedResourceMetrics = append(expectedResourceMetrics, testMetric1)
		} else {
			expectedResourceMetrics = append(expectedResourceMetrics, testMetric2)
		}
	}

	err := am.Init()
	require.NoError(t, err)

	assert.Len(t, am.receiver.Targets.ResourceTargets, 8)

	for _, target := range am.receiver.Targets.ResourceTargets {
		assert.Contains(t, []string{testFullResourceGroup1ResourceType1Resource1, testFullResourceGroup1ResourceType2Resource2, testFullResourceGroup2ResourceType1Resource3}, target.ResourceID)

		if target.ResourceID == testFullResourceGroup1ResourceType1Resource1 {
			assert.Contains(t, []int{1, 2, 3, 4, receiver.MaxMetricsPerRequest}, len(target.Metrics))

			if len(target.Metrics) == 1 {
				assert.Equal(t, []string{testMetric3}, target.Metrics)
				assert.Equal(t, []string{string(armmonitor.AggregationTypeEnumTotal), string(armmonitor.AggregationTypeEnumAverage)}, target.Aggregations)
			}
			if len(target.Metrics) == 2 {
				assert.Equal(t, []string{testMetric1, testMetric2}, target.Metrics)
				assert.Equal(t, []string{string(armmonitor.AggregationTypeEnumTotal), string(armmonitor.AggregationTypeEnumAverage)}, target.Aggregations)
			}
			if len(target.Metrics) == 3 {
				assert.Equal(t, []string{testMetric2, testMetric2, testMetric2}, target.Metrics)
				assert.Equal(t, []string{string(armmonitor.AggregationTypeEnumAverage), string(armmonitor.AggregationTypeEnumCount), string(armmonitor.AggregationTypeEnumMaximum), string(armmonitor.AggregationTypeEnumMinimum), string(armmonitor.AggregationTypeEnumTotal)}, target.Aggregations)
			}
			if len(target.Metrics) == 4 {
				assert.Equal(t, []string{testMetric3, testMetric3, testMetric3, testMetric3}, target.Metrics)
				assert.Equal(t, []string{string(armmonitor.AggregationTypeEnumAverage), string(armmonitor.AggregationTypeEnumCount), string(armmonitor.AggregationTypeEnumMaximum), string(armmonitor.AggregationTypeEnumMinimum), string(armmonitor.AggregationTypeEnumTotal)}, target.Aggregations)
			}
			if len(target.Metrics) == receiver.MaxMetricsPerRequest {
				assert.Equal(t, expectedResourceMetrics, target.Metrics)
				assert.Equal(t, []string{string(armmonitor.AggregationTypeEnumAverage), string(armmonitor.AggregationTypeEnumCount), string(armmonitor.AggregationTypeEnumMaximum), string(armmonitor.AggregationTypeEnumMinimum), string(armmonitor.AggregationTypeEnumTotal)}, target.Aggregations)
			}
		}
		if target.ResourceID == testFullResourceGroup1ResourceType2Resource2 {
			assert.Equal(t, []string{testMetric1, testMetric2}, target.Metrics)
			assert.Equal(t, []string{string(armmonitor.AggregationTypeEnumAverage), string(armmonitor.AggregationTypeEnumCount), string(armmonitor.AggregationTypeEnumMaximum), string(armmonitor.AggregationTypeEnumMinimum), string(armmonitor.AggregationTypeEnumTotal)}, target.Aggregations)
		}
		if target.ResourceID == testFullResourceGroup2ResourceType1Resource3 {
			assert.Equal(t, []string{testMetric1, testMetric2, testMetric3}, target.Metrics)
			assert.Equal(t, []string{string(armmonitor.AggregationTypeEnumTotal)}, target.Aggregations)
		}
	}
}

func TestInit_ResourceGroupTargetsOnly(t *testing.T) {
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
						Metrics:      []string{testMetric1, testMetric2, testMetric3},
						Aggregations: []string{string(armmonitor.AggregationTypeEnumTotal), string(armmonitor.AggregationTypeEnumAverage)},
					},
					{
						ResourceType: testResourceType1,
						Metrics:      []string{},
						Aggregations: []string{},
					},
					{
						ResourceType: testResourceType2,
						Metrics:      []string{},
						Aggregations: []string{string(armmonitor.AggregationTypeEnumAverage)},
					},
					{
						ResourceType: testResourceType2,
						Metrics:      []string{testMetric2},
						Aggregations: []string{},
					},
					{
						ResourceType: testResourceType2,
						Metrics:      []string{testMetric2},
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

	for index := 1; index <= 27; index++ {
		if index <= 10 {
			am.ResourceGroupTargets[0].Resources[1].Metrics = append(am.ResourceGroupTargets[0].Resources[1].Metrics, testMetric1)
		} else if index <= 23 {
			am.ResourceGroupTargets[0].Resources[1].Metrics = append(am.ResourceGroupTargets[0].Resources[1].Metrics, testMetric2)
		} else {
			am.ResourceGroupTargets[0].Resources[1].Metrics = append(am.ResourceGroupTargets[0].Resources[1].Metrics, testMetric3)
		}
	}

	expectedResourceMetrics := make([]string, 0)
	for index := 1; index <= receiver.MaxMetricsPerRequest; index++ {
		if index <= 10 {
			expectedResourceMetrics = append(expectedResourceMetrics, testMetric1)
		} else {
			expectedResourceMetrics = append(expectedResourceMetrics, testMetric2)
		}
	}

	err := am.Init()
	require.NoError(t, err)

	assert.Len(t, am.receiver.Targets.ResourceTargets, 9)

	for _, target := range am.receiver.Targets.ResourceTargets {
		assert.Contains(t, []string{testFullResourceGroup1ResourceType1Resource1, testFullResourceGroup1ResourceType2Resource2, testFullResourceGroup2ResourceType1Resource3}, target.ResourceID)

		if target.ResourceID == testFullResourceGroup1ResourceType1Resource1 {
			assert.Contains(t, []int{1, 2, 3, 4, receiver.MaxMetricsPerRequest}, len(target.Metrics))

			if len(target.Metrics) == 1 {
				assert.Equal(t, []string{testMetric3}, target.Metrics)
				assert.Equal(t, []string{string(armmonitor.AggregationTypeEnumTotal), string(armmonitor.AggregationTypeEnumAverage)}, target.Aggregations)
			}
			if len(target.Metrics) == 2 {
				assert.Equal(t, []string{testMetric1, testMetric2}, target.Metrics)
				assert.Equal(t, []string{string(armmonitor.AggregationTypeEnumTotal), string(armmonitor.AggregationTypeEnumAverage)}, target.Aggregations)
			}
			if len(target.Metrics) == 3 {
				assert.Equal(t, []string{testMetric2, testMetric2, testMetric2}, target.Metrics)
				assert.Equal(t, []string{string(armmonitor.AggregationTypeEnumAverage), string(armmonitor.AggregationTypeEnumCount), string(armmonitor.AggregationTypeEnumMaximum), string(armmonitor.AggregationTypeEnumMinimum), string(armmonitor.AggregationTypeEnumTotal)}, target.Aggregations)
			}
			if len(target.Metrics) == 4 {
				assert.Equal(t, []string{testMetric3, testMetric3, testMetric3, testMetric3}, target.Metrics)
				assert.Equal(t, []string{string(armmonitor.AggregationTypeEnumAverage), string(armmonitor.AggregationTypeEnumCount), string(armmonitor.AggregationTypeEnumMaximum), string(armmonitor.AggregationTypeEnumMinimum), string(armmonitor.AggregationTypeEnumTotal)}, target.Aggregations)
			}
			if len(target.Metrics) == receiver.MaxMetricsPerRequest {
				assert.Equal(t, expectedResourceMetrics, target.Metrics)
				assert.Equal(t, []string{string(armmonitor.AggregationTypeEnumAverage), string(armmonitor.AggregationTypeEnumCount), string(armmonitor.AggregationTypeEnumMaximum), string(armmonitor.AggregationTypeEnumMinimum), string(armmonitor.AggregationTypeEnumTotal)}, target.Aggregations)
			}
		}
		if target.ResourceID == testFullResourceGroup1ResourceType2Resource2 {
			assert.Contains(t, []int{1, 2}, len(target.Metrics))

			if len(target.Metrics) == 1 {
				assert.Equal(t, []string{testMetric2}, target.Metrics)
				assert.Equal(t, []string{string(armmonitor.AggregationTypeEnumAverage), string(armmonitor.AggregationTypeEnumCount), string(armmonitor.AggregationTypeEnumMaximum), string(armmonitor.AggregationTypeEnumMinimum), string(armmonitor.AggregationTypeEnumTotal)}, target.Aggregations)
			}
			if len(target.Metrics) == 2 {
				assert.Contains(t, []int{1, 5}, len(target.Aggregations))

				if len(target.Aggregations) == 1 {
					assert.Equal(t, []string{testMetric1, testMetric2}, target.Metrics)
					assert.Equal(t, []string{string(armmonitor.AggregationTypeEnumAverage)}, target.Aggregations)
				}
				if len(target.Aggregations) == 5 {
					assert.Equal(t, []string{testMetric1, testMetric2}, target.Metrics)
					assert.Equal(t, []string{string(armmonitor.AggregationTypeEnumAverage), string(armmonitor.AggregationTypeEnumCount), string(armmonitor.AggregationTypeEnumMaximum), string(armmonitor.AggregationTypeEnumMinimum), string(armmonitor.AggregationTypeEnumTotal)}, target.Aggregations)
				}
			}
		}
		if target.ResourceID == testFullResourceGroup2ResourceType1Resource3 {
			assert.Equal(t, []string{testMetric3}, target.Metrics)
			assert.Equal(t, []string{string(armmonitor.AggregationTypeEnumTotal), string(armmonitor.AggregationTypeEnumAverage)}, target.Aggregations)
		}
	}
}

func TestInit_SubscriptionTargetsOnly(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: testSubscriptionID,
		ClientID:       testClientID,
		ClientSecret:   testClientSecret,
		TenantID:       testTenantID,
		SubscriptionTargets: []*Resource{
			{
				ResourceType: testResourceType1,
				Metrics:      []string{testMetric1, testMetric2, testMetric3},
				Aggregations: []string{string(armmonitor.AggregationTypeEnumTotal), string(armmonitor.AggregationTypeEnumAverage)},
			},
			{
				ResourceType: testResourceType2,
				Metrics:      []string{},
				Aggregations: []string{string(armmonitor.AggregationTypeEnumAverage)},
			},
			{
				ResourceType: testResourceType2,
				Metrics:      []string{testMetric2},
				Aggregations: []string{},
			},
			{
				ResourceType: testResourceType2,
				Metrics:      []string{},
				Aggregations: []string{},
			},
			{
				ResourceType: testResourceType1,
				Metrics:      []string{},
				Aggregations: []string{},
			},
		},
		Log:          testutil.Logger{},
		azureClients: setMockAzureClients(),
	}

	for index := 1; index <= 27; index++ {
		if index <= 10 {
			am.SubscriptionTargets[4].Metrics = append(am.SubscriptionTargets[4].Metrics, testMetric1)
		} else if index <= 23 {
			am.SubscriptionTargets[4].Metrics = append(am.SubscriptionTargets[4].Metrics, testMetric2)
		} else {
			am.SubscriptionTargets[4].Metrics = append(am.SubscriptionTargets[4].Metrics, testMetric3)
		}
	}

	expectedResourceMetrics := make([]string, 0)
	for index := 1; index <= receiver.MaxMetricsPerRequest; index++ {
		if index <= 10 {
			expectedResourceMetrics = append(expectedResourceMetrics, testMetric1)
		} else {
			expectedResourceMetrics = append(expectedResourceMetrics, testMetric2)
		}
	}

	err := am.Init()
	require.NoError(t, err)

	assert.Len(t, am.receiver.Targets.ResourceTargets, 11)

	for _, target := range am.receiver.Targets.ResourceTargets {
		assert.Contains(t, []string{testFullResourceGroup1ResourceType1Resource1, testFullResourceGroup1ResourceType2Resource2, testFullResourceGroup2ResourceType1Resource3}, target.ResourceID)

		if target.ResourceID == testFullResourceGroup1ResourceType1Resource1 {
			assert.Contains(t, []int{1, 2, 3, 4, receiver.MaxMetricsPerRequest}, len(target.Metrics))

			if len(target.Metrics) == 1 {
				assert.Equal(t, []string{testMetric3}, target.Metrics)
				assert.Equal(t, []string{string(armmonitor.AggregationTypeEnumTotal), string(armmonitor.AggregationTypeEnumAverage)}, target.Aggregations)
			}
			if len(target.Metrics) == 2 {
				assert.Equal(t, []string{testMetric1, testMetric2}, target.Metrics)
				assert.Equal(t, []string{string(armmonitor.AggregationTypeEnumTotal), string(armmonitor.AggregationTypeEnumAverage)}, target.Aggregations)
			}
			if len(target.Metrics) == 3 {
				assert.Equal(t, []string{testMetric2, testMetric2, testMetric2}, target.Metrics)
				assert.Equal(t, []string{string(armmonitor.AggregationTypeEnumAverage), string(armmonitor.AggregationTypeEnumCount), string(armmonitor.AggregationTypeEnumMaximum), string(armmonitor.AggregationTypeEnumMinimum), string(armmonitor.AggregationTypeEnumTotal)}, target.Aggregations)
			}
			if len(target.Metrics) == 4 {
				assert.Equal(t, []string{testMetric3, testMetric3, testMetric3, testMetric3}, target.Metrics)
				assert.Equal(t, []string{string(armmonitor.AggregationTypeEnumAverage), string(armmonitor.AggregationTypeEnumCount), string(armmonitor.AggregationTypeEnumMaximum), string(armmonitor.AggregationTypeEnumMinimum), string(armmonitor.AggregationTypeEnumTotal)}, target.Aggregations)
			}
			if len(target.Metrics) == receiver.MaxMetricsPerRequest {
				assert.Equal(t, expectedResourceMetrics, target.Metrics)
				assert.Equal(t, []string{string(armmonitor.AggregationTypeEnumAverage), string(armmonitor.AggregationTypeEnumCount), string(armmonitor.AggregationTypeEnumMaximum), string(armmonitor.AggregationTypeEnumMinimum), string(armmonitor.AggregationTypeEnumTotal)}, target.Aggregations)
			}
		}
		if target.ResourceID == testFullResourceGroup1ResourceType2Resource2 {
			assert.Contains(t, []int{1, 2}, len(target.Metrics))

			if len(target.Metrics) == 1 {
				assert.Equal(t, []string{testMetric2}, target.Metrics)
				assert.Equal(t, []string{string(armmonitor.AggregationTypeEnumAverage), string(armmonitor.AggregationTypeEnumCount), string(armmonitor.AggregationTypeEnumMaximum), string(armmonitor.AggregationTypeEnumMinimum), string(armmonitor.AggregationTypeEnumTotal)}, target.Aggregations)
			}
			if len(target.Metrics) == 2 {
				assert.Contains(t, []int{1, 5}, len(target.Aggregations))

				if len(target.Aggregations) == 1 {
					assert.Equal(t, []string{testMetric1, testMetric2}, target.Metrics)
					assert.Equal(t, []string{string(armmonitor.AggregationTypeEnumAverage)}, target.Aggregations)
				}
				if len(target.Aggregations) == 5 {
					assert.Equal(t, []string{testMetric1, testMetric2}, target.Metrics)
					assert.Equal(t, []string{string(armmonitor.AggregationTypeEnumAverage), string(armmonitor.AggregationTypeEnumCount), string(armmonitor.AggregationTypeEnumMaximum), string(armmonitor.AggregationTypeEnumMinimum), string(armmonitor.AggregationTypeEnumTotal)}, target.Aggregations)
				}
			}
		}
		if target.ResourceID == testFullResourceGroup2ResourceType1Resource3 {
			assert.Contains(t, []int{3, 7, receiver.MaxMetricsPerRequest}, len(target.Metrics))

			if len(target.Metrics) == 3 {
				assert.Equal(t, []string{testMetric1, testMetric2, testMetric3}, target.Metrics)
				assert.Equal(t, []string{string(armmonitor.AggregationTypeEnumTotal), string(armmonitor.AggregationTypeEnumAverage)}, target.Aggregations)
			}
			if len(target.Metrics) == 7 {
				assert.Equal(t, []string{testMetric2, testMetric2, testMetric2, testMetric3, testMetric3, testMetric3, testMetric3}, target.Metrics)
				assert.Equal(t, []string{string(armmonitor.AggregationTypeEnumAverage), string(armmonitor.AggregationTypeEnumCount), string(armmonitor.AggregationTypeEnumMaximum), string(armmonitor.AggregationTypeEnumMinimum), string(armmonitor.AggregationTypeEnumTotal)}, target.Aggregations)
			}
			if len(target.Metrics) == receiver.MaxMetricsPerRequest {
				assert.Equal(t, expectedResourceMetrics, target.Metrics)
				assert.Equal(t, []string{string(armmonitor.AggregationTypeEnumAverage), string(armmonitor.AggregationTypeEnumCount), string(armmonitor.AggregationTypeEnumMaximum), string(armmonitor.AggregationTypeEnumMinimum), string(armmonitor.AggregationTypeEnumTotal)}, target.Aggregations)
			}
		}
	}
}

func TestInit_AllTargetTypes(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: testSubscriptionID,
		ClientID:       testClientID,
		ClientSecret:   testClientSecret,
		TenantID:       testTenantID,
		ResourceTargets: []*ResourceTarget{
			{
				testResourceGroup1ResourceType1Resource1,
				[]string{testMetric1, testMetric2, testMetric3},
				[]string{string(armmonitor.AggregationTypeEnumTotal), string(armmonitor.AggregationTypeEnumAverage)},
			},
			{
				testResourceGroup1ResourceType2Resource2,
				[]string{testMetric1, testMetric2},
				[]string{},
			},
			{
				testResourceGroup2ResourceType1Resource3,
				[]string{},
				[]string{string(armmonitor.AggregationTypeEnumTotal)},
			},
			{
				testResourceGroup1ResourceType2Resource2,
				[]string{},
				[]string{},
			},
			{
				testResourceGroup1ResourceType1Resource1,
				[]string{},
				[]string{},
			},
		},
		ResourceGroupTargets: []*ResourceGroupTarget{
			{
				ResourceGroup: testResourceGroup1,
				Resources: []*Resource{
					{
						ResourceType: testResourceType1,
						Metrics:      []string{testMetric1, testMetric2, testMetric3},
						Aggregations: []string{string(armmonitor.AggregationTypeEnumTotal), string(armmonitor.AggregationTypeEnumAverage)},
					},
					{
						ResourceType: testResourceType1,
						Metrics:      []string{},
						Aggregations: []string{},
					},
					{
						ResourceType: testResourceType2,
						Metrics:      []string{},
						Aggregations: []string{string(armmonitor.AggregationTypeEnumAverage)},
					},
					{
						ResourceType: testResourceType2,
						Metrics:      []string{testMetric2},
						Aggregations: []string{},
					},
					{
						ResourceType: testResourceType2,
						Metrics:      []string{testMetric2},
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
		SubscriptionTargets: []*Resource{
			{
				ResourceType: testResourceType1,
				Metrics:      []string{testMetric1, testMetric2, testMetric3},
				Aggregations: []string{string(armmonitor.AggregationTypeEnumTotal), string(armmonitor.AggregationTypeEnumAverage)},
			},
			{
				ResourceType: testResourceType2,
				Metrics:      []string{},
				Aggregations: []string{string(armmonitor.AggregationTypeEnumAverage)},
			},
			{
				ResourceType: testResourceType2,
				Metrics:      []string{testMetric2},
				Aggregations: []string{},
			},
			{
				ResourceType: testResourceType2,
				Metrics:      []string{},
				Aggregations: []string{},
			},
			{
				ResourceType: testResourceType1,
				Metrics:      []string{},
				Aggregations: []string{},
			},
		},
		Log:          testutil.Logger{},
		azureClients: setMockAzureClients(),
	}

	for index := 1; index <= 27; index++ {
		if index <= 10 {
			am.ResourceTargets[4].Metrics = append(am.ResourceTargets[4].Metrics, testMetric1)
			am.ResourceGroupTargets[0].Resources[1].Metrics = append(am.ResourceGroupTargets[0].Resources[1].Metrics, testMetric1)
			am.SubscriptionTargets[4].Metrics = append(am.SubscriptionTargets[4].Metrics, testMetric1)
		} else if index <= 23 {
			am.ResourceTargets[4].Metrics = append(am.ResourceTargets[4].Metrics, testMetric2)
			am.ResourceGroupTargets[0].Resources[1].Metrics = append(am.ResourceGroupTargets[0].Resources[1].Metrics, testMetric2)
			am.SubscriptionTargets[4].Metrics = append(am.SubscriptionTargets[4].Metrics, testMetric2)
		} else {
			am.ResourceTargets[4].Metrics = append(am.ResourceTargets[4].Metrics, testMetric3)
			am.ResourceGroupTargets[0].Resources[1].Metrics = append(am.ResourceGroupTargets[0].Resources[1].Metrics, testMetric3)
			am.SubscriptionTargets[4].Metrics = append(am.SubscriptionTargets[4].Metrics, testMetric3)
		}
	}

	err := am.Init()
	require.NoError(t, err)

	assert.Len(t, am.receiver.Targets.ResourceTargets, 28)
}

func TestInit_NoSubscriptionID(t *testing.T) {
	am := &AzureMonitor{
		ClientID:     testClientID,
		ClientSecret: testClientSecret,
		TenantID:     testTenantID,
		ResourceTargets: []*ResourceTarget{
			{
				testResourceGroup1ResourceType1Resource1,
				[]string{},
				[]string{},
			},
		},
		Log:          testutil.Logger{},
		azureClients: setMockAzureClients(),
	}

	err := am.Init()
	require.Error(t, err)
}

func TestInit_NoClientID(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: testSubscriptionID,
		ClientSecret:   testClientSecret,
		TenantID:       testTenantID,
		ResourceTargets: []*ResourceTarget{
			{
				testResourceGroup1ResourceType1Resource1,
				[]string{},
				[]string{},
			},
		},
		Log:          testutil.Logger{},
		azureClients: setMockAzureClients(),
	}

	err := am.Init()
	require.Error(t, err)
}

func TestInit_NoClientSecret(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: testSubscriptionID,
		ClientID:       testClientID,
		TenantID:       testTenantID,
		ResourceTargets: []*ResourceTarget{
			{
				testResourceGroup1ResourceType1Resource1,
				[]string{},
				[]string{},
			},
		},
		Log:          testutil.Logger{},
		azureClients: setMockAzureClients(),
	}

	err := am.Init()
	require.Error(t, err)
}

func TestInit_NoTenantID(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: testSubscriptionID,
		ClientID:       testClientID,
		ClientSecret:   testClientSecret,
		ResourceTargets: []*ResourceTarget{
			{
				testResourceGroup1ResourceType1Resource1,
				[]string{},
				[]string{},
			},
		},
		Log:          testutil.Logger{},
		azureClients: setMockAzureClients(),
	}

	err := am.Init()
	require.Error(t, err)
}

func TestInit_NoTargets(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: testSubscriptionID,
		ClientID:       testClientID,
		ClientSecret:   testClientSecret,
		TenantID:       testTenantID,
		Log:            testutil.Logger{},
		azureClients:   setMockAzureClients(),
	}

	err := am.Init()
	require.Error(t, err)
}

func TestInit_ResourceTargetWithoutResourceID(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: testSubscriptionID,
		ClientID:       testClientID,
		ClientSecret:   testClientSecret,
		TenantID:       testTenantID,
		ResourceTargets: []*ResourceTarget{
			{
				"",
				[]string{},
				[]string{},
			},
		},
		Log:          testutil.Logger{},
		azureClients: setMockAzureClients(),
	}

	err := am.Init()
	require.Error(t, err)
}

func TestInit_ResourceTargetWithInvalidAggregation(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: testSubscriptionID,
		ClientID:       testClientID,
		ClientSecret:   testClientSecret,
		TenantID:       testTenantID,
		ResourceTargets: []*ResourceTarget{
			{
				testResourceGroup1ResourceType1Resource1,
				[]string{},
				[]string{testInvalidAggregation},
			},
		},
		Log:          testutil.Logger{},
		azureClients: setMockAzureClients(),
	}

	err := am.Init()
	require.Error(t, err)
}

func TestInit_ResourceGroupTargetWithoutResourceGroup(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: testSubscriptionID,
		ClientID:       testClientID,
		ClientSecret:   testClientSecret,
		TenantID:       testTenantID,
		ResourceGroupTargets: []*ResourceGroupTarget{
			{
				"",
				[]*Resource{
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

	err := am.Init()
	require.Error(t, err)
}

func TestInit_ResourceGroupTargetWithResourceWithoutResourceType(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: testSubscriptionID,
		ClientID:       testClientID,
		ClientSecret:   testClientSecret,
		TenantID:       testTenantID,
		ResourceGroupTargets: []*ResourceGroupTarget{
			{
				testResourceGroup1,
				[]*Resource{
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

	err := am.Init()
	require.Error(t, err)
}

func TestInit_ResourceGroupTargetWithInvalidMetric(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: testSubscriptionID,
		ClientID:       testClientID,
		ClientSecret:   testClientSecret,
		TenantID:       testTenantID,
		ResourceGroupTargets: []*ResourceGroupTarget{
			{
				testResourceGroup1,
				[]*Resource{
					{
						ResourceType: testResourceType1,
						Metrics:      []string{testInvalidMetric},
						Aggregations: []string{},
					},
				},
			},
		},
		Log:          testutil.Logger{},
		azureClients: setMockAzureClients(),
	}

	err := am.Init()
	require.Error(t, err)
}

func TestInit_ResourceGroupTargetWithInvalidAggregation(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: testSubscriptionID,
		ClientID:       testClientID,
		ClientSecret:   testClientSecret,
		TenantID:       testTenantID,
		ResourceGroupTargets: []*ResourceGroupTarget{
			{
				testResourceGroup1,
				[]*Resource{
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

	err := am.Init()
	require.Error(t, err)
}

func TestInit_ResourceGroupTargetWithoutResources(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: testSubscriptionID,
		ClientID:       testClientID,
		ClientSecret:   testClientSecret,
		TenantID:       testTenantID,
		ResourceGroupTargets: []*ResourceGroupTarget{
			{
				testResourceGroup1,
				[]*Resource{},
			},
		},
		Log:          testutil.Logger{},
		azureClients: setMockAzureClients(),
	}

	err := am.Init()
	require.Error(t, err)
}

func TestInit_ResourceGroupTargetNoResourceFound(t *testing.T) {
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

	err := am.Init()
	require.Error(t, err)
}

func TestInit_SubscriptionTargetWithoutResourceType(t *testing.T) {
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

	err := am.Init()
	require.Error(t, err)
}

func TestInit_SubscriptionTargetWithInvalidMetric(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: testSubscriptionID,
		ClientID:       testClientID,
		ClientSecret:   testClientSecret,
		TenantID:       testTenantID,
		SubscriptionTargets: []*Resource{
			{
				ResourceType: testResourceType1,
				Metrics:      []string{testInvalidMetric},
				Aggregations: []string{},
			},
		},
		Log:          testutil.Logger{},
		azureClients: setMockAzureClients(),
	}

	err := am.Init()
	require.Error(t, err)
}

func TestInit_SubscriptionTargetWithInvalidAggregation(t *testing.T) {
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

	err := am.Init()
	require.Error(t, err)
}

func TestInit_SubscriptionTargetNoResourceFound(t *testing.T) {
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

	err := am.Init()
	require.Error(t, err)
}

func TestInit_BadCredentials(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: testSubscriptionID,
		ClientID:       testClientID,
		ClientSecret:   testClientSecret,
		TenantID:       testTenantID,
		ResourceTargets: []*ResourceTarget{
			{
				testResourceGroup1ResourceType1Resource1,
				[]string{},
				[]string{},
			},
		},
		Log: testutil.Logger{},
	}

	err := am.Init()
	require.Error(t, err)
}

func TestGather_Success(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: testSubscriptionID,
		ClientID:       testClientID,
		ClientSecret:   testClientSecret,
		TenantID:       testTenantID,
		ResourceTargets: []*ResourceTarget{
			{
				testResourceGroup1ResourceType1Resource1,
				[]string{testMetric1, testMetric2},
				[]string{string(armmonitor.AggregationTypeEnumTotal), string(armmonitor.AggregationTypeEnumMaximum)},
			},
			{
				testResourceGroup1ResourceType2Resource2,
				[]string{testMetric1},
				[]string{string(armmonitor.AggregationTypeEnumTotal), string(armmonitor.AggregationTypeEnumMinimum)},
			},
			{
				testResourceGroup2ResourceType1Resource3,
				[]string{testMetric1},
				[]string{string(armmonitor.AggregationTypeEnumTotal), string(armmonitor.AggregationTypeEnumMinimum)},
			},
			{
				testResourceGroup2ResourceType2Resource4,
				[]string{testMetric1},
				[]string{string(armmonitor.AggregationTypeEnumTotal), string(armmonitor.AggregationTypeEnumMaximum)},
			},
			{
				testResourceGroup2ResourceType2Resource5,
				[]string{testMetric2},
				[]string{string(armmonitor.AggregationTypeEnumTotal)},
			},
			{
				testResourceGroup2ResourceType2Resource6,
				[]string{testMetric2},
				[]string{string(armmonitor.AggregationTypeEnumAverage)},
			},
		},
		Log:          testutil.Logger{},
		azureClients: setMockAzureClients(),
	}

	resourceTargets := make([]*receiver.ResourceTarget, 0)

	for _, target := range am.ResourceTargets {
		resourceTargets = append(resourceTargets, receiver.NewResourceTarget(target.ResourceID, target.Metrics, target.Aggregations))
	}

	rec, err := receiver.NewAzureMonitorMetricsReceiver(
		am.SubscriptionID,
		am.ClientID,
		am.ClientSecret,
		am.TenantID,
		receiver.NewTargets(resourceTargets, []*receiver.ResourceGroupTarget{}, []*receiver.Resource{}),
		am.azureClients,
	)
	require.NoError(t, err)
	require.NotNil(t, rec)

	am.receiver = rec

	expectedResource1Metric1Name := "azure_monitor_microsoft_test_type1_metric1"
	expectedResource1Metric1MetricFields := make(map[string]interface{})
	expectedResource1Metric1MetricFields[receiver.MetricFieldTimeStamp] = "2022-02-22T22:59:00Z"
	expectedResource1Metric1MetricFields[receiver.MetricFieldTotal] = 5.0
	expectedResource1Metric1MetricFields[receiver.MetricFieldMaximum] = 5.0
	expectedResource1Metric2Name := "azure_monitor_microsoft_test_type1_metric2"
	expectedResource1Metric2MetricFields := make(map[string]interface{})
	expectedResource1Metric2MetricFields[receiver.MetricFieldTimeStamp] = "2022-02-22T22:59:00Z"
	expectedResource1Metric2MetricFields[receiver.MetricFieldTotal] = 2.5
	expectedResource1Metric2MetricFields[receiver.MetricFieldMaximum] = 2.5
	expectedResource1MetricsTags := make(map[string]string)
	expectedResource1MetricsTags[receiver.MetricTagSubscriptionID] = testSubscriptionID
	expectedResource1MetricsTags[receiver.MetricTagResourceGroup] = testResourceGroup1
	expectedResource1MetricsTags[receiver.MetricTagNamespace] = testResourceType1
	expectedResource1MetricsTags[receiver.MetricTagResourceName] = testResource1Name
	expectedResource1MetricsTags[receiver.MetricTagResourceRegion] = testResourceRegion
	expectedResource1MetricsTags[receiver.MetricTagUnit] = string(armmonitor.MetricUnitCount)

	expectedResource2Metric1Name := "azure_monitor_microsoft_test_type2_metric1"
	expectedResource2Metric1MetricFields := make(map[string]interface{})
	expectedResource2Metric1MetricFields[receiver.MetricFieldTimeStamp] = "2022-02-22T22:59:00Z"
	expectedResource2Metric1MetricFields[receiver.MetricFieldTotal] = 5.0
	expectedResource2Metric1MetricFields[receiver.MetricFieldMinimum] = 2.5
	expectedResource2MetricsTags := make(map[string]string)
	expectedResource2MetricsTags[receiver.MetricTagSubscriptionID] = testSubscriptionID
	expectedResource2MetricsTags[receiver.MetricTagResourceGroup] = testResourceGroup1
	expectedResource2MetricsTags[receiver.MetricTagNamespace] = testResourceType2
	expectedResource2MetricsTags[receiver.MetricTagResourceName] = testResource2Name
	expectedResource2MetricsTags[receiver.MetricTagResourceRegion] = testResourceRegion
	expectedResource2MetricsTags[receiver.MetricTagUnit] = string(armmonitor.MetricUnitCount)

	expectedResource3Metric1Name := "azure_monitor_microsoft_test_type1_metric1"
	expectedResource3Metric1MetricFields := make(map[string]interface{})
	expectedResource3Metric1MetricFields[receiver.MetricFieldTimeStamp] = "2022-02-22T22:58:00Z"
	expectedResource3Metric1MetricFields[receiver.MetricFieldTotal] = 2.5
	expectedResource3Metric1MetricFields[receiver.MetricFieldMinimum] = 2.5
	expectedResource3MetricsTags := make(map[string]string)
	expectedResource3MetricsTags[receiver.MetricTagSubscriptionID] = testSubscriptionID
	expectedResource3MetricsTags[receiver.MetricTagResourceGroup] = testResourceGroup2
	expectedResource3MetricsTags[receiver.MetricTagNamespace] = testResourceType1
	expectedResource3MetricsTags[receiver.MetricTagResourceName] = testResource3Name
	expectedResource3MetricsTags[receiver.MetricTagResourceRegion] = testResourceRegion
	expectedResource3MetricsTags[receiver.MetricTagUnit] = string(armmonitor.MetricUnitBytes)

	acc := testutil.Accumulator{}
	err = acc.GatherError(am.Gather)
	require.NoError(t, err)

	assert.Len(t, acc.Metrics, 4)

	acc.AssertContainsTaggedFields(t, expectedResource1Metric1Name, expectedResource1Metric1MetricFields, expectedResource1MetricsTags)
	acc.AssertContainsTaggedFields(t, expectedResource1Metric2Name, expectedResource1Metric2MetricFields, expectedResource1MetricsTags)
	acc.AssertContainsTaggedFields(t, expectedResource2Metric1Name, expectedResource2Metric1MetricFields, expectedResource2MetricsTags)
	acc.AssertContainsTaggedFields(t, expectedResource3Metric1Name, expectedResource3Metric1MetricFields, expectedResource3MetricsTags)
}
