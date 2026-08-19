package azure_monitor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/monitor/armmonitor"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/influxdata/toml"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/testutil"
)

func TestInit_ResourceTargetsOnly(t *testing.T) {
	file, err := os.ReadFile("testdata/toml/init_resource_targets_only.toml")
	require.NoError(t, err)
	require.NotNil(t, file)
	require.NotEmpty(t, file)

	var am *AzureMonitor
	require.NoError(t, toml.Unmarshal(file, &am))

	am.Log = testutil.Logger{}
	am.factory = &mockClientFactory{}

	for index := 1; index <= 27; index++ {
		if index <= 10 {
			am.ResourceTargets[4].Metrics = append(am.ResourceTargets[4].Metrics, "metric1")
		} else if index <= 23 {
			am.ResourceTargets[4].Metrics = append(am.ResourceTargets[4].Metrics, "metric2")
		} else {
			am.ResourceTargets[4].Metrics = append(am.ResourceTargets[4].Metrics, "metric3")
		}
	}

	var expectedResourceMetrics []string
	for index := 1; index <= maxMetricsBatch; index++ {
		if index <= 10 {
			expectedResourceMetrics = append(expectedResourceMetrics, "metric1")
		} else {
			expectedResourceMetrics = append(expectedResourceMetrics, "metric2")
		}
	}

	require.NoError(t, am.Init())
	require.NoError(t, am.Start(nil))
	require.Len(t, am.receiver.resources, 8)

	for _, target := range am.receiver.resources {
		require.Contains(t, []string{
			"/subscriptions/subscriptionID/resourceGroups/resourceGroup1/providers/Microsoft.Test/type1/resource1",
			"/subscriptions/subscriptionID/resourceGroups/resourceGroup1/providers/Microsoft.Test/type2/resource2",
			"/subscriptions/subscriptionID/resourceGroups/resourceGroup2/providers/Microsoft.Test/type1/resource3"}, target.ResourceID)

		if target.ResourceID == "/subscriptions/subscriptionID/resourceGroups/resourceGroup1/providers/Microsoft.Test/type1/resource1" {
			require.Contains(t, []int{1, 2, 3, 4, maxMetricsBatch}, len(target.Metrics))

			if len(target.Metrics) == 1 {
				require.Equal(t, []string{"metric3"}, target.Metrics)
				require.Equal(t, []string{string(armmonitor.AggregationTypeEnumTotal), string(armmonitor.AggregationTypeEnumAverage)}, target.Aggregations)
			}
			if len(target.Metrics) == 2 {
				require.Equal(t, []string{"metric1", "metric2"}, target.Metrics)
				require.Equal(t, []string{string(armmonitor.AggregationTypeEnumTotal), string(armmonitor.AggregationTypeEnumAverage)}, target.Aggregations)
			}
			if len(target.Metrics) == 3 {
				require.Equal(t, []string{"metric2", "metric2", "metric2"}, target.Metrics)
				require.Equal(t, []string{
					string(armmonitor.AggregationTypeEnumAverage),
					string(armmonitor.AggregationTypeEnumCount),
					string(armmonitor.AggregationTypeEnumMaximum),
					string(armmonitor.AggregationTypeEnumMinimum),
					string(armmonitor.AggregationTypeEnumTotal),
				}, target.Aggregations)
			}
			if len(target.Metrics) == 4 {
				require.Equal(t, []string{"metric3", "metric3", "metric3", "metric3"}, target.Metrics)
				require.Equal(t, []string{
					string(armmonitor.AggregationTypeEnumAverage),
					string(armmonitor.AggregationTypeEnumCount),
					string(armmonitor.AggregationTypeEnumMaximum),
					string(armmonitor.AggregationTypeEnumMinimum),
					string(armmonitor.AggregationTypeEnumTotal),
				}, target.Aggregations)
			}
			if len(target.Metrics) == maxMetricsBatch {
				require.Equal(t, expectedResourceMetrics, target.Metrics)
				require.Equal(t, []string{
					string(armmonitor.AggregationTypeEnumAverage),
					string(armmonitor.AggregationTypeEnumCount),
					string(armmonitor.AggregationTypeEnumMaximum),
					string(armmonitor.AggregationTypeEnumMinimum),
					string(armmonitor.AggregationTypeEnumTotal),
				}, target.Aggregations)
			}
		}
		if target.ResourceID == "/subscriptions/subscriptionID/resourceGroups/resourceGroup1/providers/Microsoft.Test/type2/resource2" {
			require.Equal(t, []string{"metric1", "metric2"}, target.Metrics)
			require.Equal(t, []string{
				string(armmonitor.AggregationTypeEnumAverage),
				string(armmonitor.AggregationTypeEnumCount),
				string(armmonitor.AggregationTypeEnumMaximum),
				string(armmonitor.AggregationTypeEnumMinimum),
				string(armmonitor.AggregationTypeEnumTotal),
			}, target.Aggregations)
		}
		if target.ResourceID == "/subscriptions/subscriptionID/resourceGroups/resourceGroup2/providers/Microsoft.Test/type1/resource3" {
			require.Equal(t, []string{"metric1", "metric2", "metric3"}, target.Metrics)
			require.Equal(t, []string{string(armmonitor.AggregationTypeEnumTotal)}, target.Aggregations)
		}
	}
}

func TestInit_ResourceGroupTargetsOnly(t *testing.T) {
	file, err := os.ReadFile("testdata/toml/init_resource_group_targets_only.toml")
	require.NoError(t, err)
	require.NotNil(t, file)
	require.NotEmpty(t, file)

	var am *AzureMonitor
	require.NoError(t, toml.Unmarshal(file, &am))

	am.Log = testutil.Logger{}
	am.factory = &mockClientFactory{}

	for index := 1; index <= 27; index++ {
		if index <= 10 {
			am.ResourceGroupTargets[0].Resources[1].Metrics = append(am.ResourceGroupTargets[0].Resources[1].Metrics, "metric1")
		} else if index <= 23 {
			am.ResourceGroupTargets[0].Resources[1].Metrics = append(am.ResourceGroupTargets[0].Resources[1].Metrics, "metric2")
		} else {
			am.ResourceGroupTargets[0].Resources[1].Metrics = append(am.ResourceGroupTargets[0].Resources[1].Metrics, "metric3")
		}
	}

	var expectedResourceMetrics []string
	for index := 1; index <= maxMetricsBatch; index++ {
		if index <= 10 {
			expectedResourceMetrics = append(expectedResourceMetrics, "metric1")
		} else {
			expectedResourceMetrics = append(expectedResourceMetrics, "metric2")
		}
	}

	require.NoError(t, am.Init())
	require.NoError(t, am.Start(nil))
	require.Len(t, am.receiver.resources, 9)

	for _, target := range am.receiver.resources {
		require.Contains(t, []string{
			"/subscriptions/subscriptionID/resourceGroups/resourceGroup1/providers/Microsoft.Test/type1/resource1",
			"/subscriptions/subscriptionID/resourceGroups/resourceGroup1/providers/Microsoft.Test/type2/resource2",
			"/subscriptions/subscriptionID/resourceGroups/resourceGroup2/providers/Microsoft.Test/type1/resource3"}, target.ResourceID)

		if target.ResourceID == "/subscriptions/subscriptionID/resourceGroups/resourceGroup1/providers/Microsoft.Test/type1/resource1" {
			require.Contains(t, []int{1, 2, 3, 4, maxMetricsBatch}, len(target.Metrics))

			if len(target.Metrics) == 1 {
				require.Equal(t, []string{"metric3"}, target.Metrics)
				require.Equal(t, []string{string(armmonitor.AggregationTypeEnumTotal), string(armmonitor.AggregationTypeEnumAverage)}, target.Aggregations)
			}
			if len(target.Metrics) == 2 {
				require.Equal(t, []string{"metric1", "metric2"}, target.Metrics)
				require.Equal(t, []string{string(armmonitor.AggregationTypeEnumTotal), string(armmonitor.AggregationTypeEnumAverage)}, target.Aggregations)
			}
			if len(target.Metrics) == 3 {
				require.Equal(t, []string{"metric2", "metric2", "metric2"}, target.Metrics)
				require.Equal(t, []string{
					string(armmonitor.AggregationTypeEnumAverage),
					string(armmonitor.AggregationTypeEnumCount),
					string(armmonitor.AggregationTypeEnumMaximum),
					string(armmonitor.AggregationTypeEnumMinimum),
					string(armmonitor.AggregationTypeEnumTotal),
				}, target.Aggregations)
			}
			if len(target.Metrics) == 4 {
				require.Equal(t, []string{"metric3", "metric3", "metric3", "metric3"}, target.Metrics)
				require.Equal(t, []string{
					string(armmonitor.AggregationTypeEnumAverage),
					string(armmonitor.AggregationTypeEnumCount),
					string(armmonitor.AggregationTypeEnumMaximum),
					string(armmonitor.AggregationTypeEnumMinimum),
					string(armmonitor.AggregationTypeEnumTotal),
				}, target.Aggregations)
			}
			if len(target.Metrics) == maxMetricsBatch {
				require.Equal(t, expectedResourceMetrics, target.Metrics)
				require.Equal(t, []string{
					string(armmonitor.AggregationTypeEnumAverage),
					string(armmonitor.AggregationTypeEnumCount),
					string(armmonitor.AggregationTypeEnumMaximum),
					string(armmonitor.AggregationTypeEnumMinimum),
					string(armmonitor.AggregationTypeEnumTotal),
				}, target.Aggregations)
			}
		}
		if target.ResourceID == "/subscriptions/subscriptionID/resourceGroups/resourceGroup1/providers/Microsoft.Test/type2/resource2" {
			require.Contains(t, []int{1, 2}, len(target.Metrics))

			if len(target.Metrics) == 1 {
				require.Equal(t, []string{"metric2"}, target.Metrics)
				require.Equal(t, []string{
					string(armmonitor.AggregationTypeEnumAverage),
					string(armmonitor.AggregationTypeEnumCount),
					string(armmonitor.AggregationTypeEnumMaximum),
					string(armmonitor.AggregationTypeEnumMinimum),
					string(armmonitor.AggregationTypeEnumTotal),
				}, target.Aggregations)
			}
			if len(target.Metrics) == 2 {
				require.Contains(t, []int{1, 5}, len(target.Aggregations))

				if len(target.Aggregations) == 1 {
					require.Equal(t, []string{"metric1", "metric2"}, target.Metrics)
					require.Equal(t, []string{string(armmonitor.AggregationTypeEnumAverage)}, target.Aggregations)
				}
				if len(target.Aggregations) == 5 {
					require.Equal(t, []string{"metric1", "metric2"}, target.Metrics)
					require.Equal(t, []string{
						string(armmonitor.AggregationTypeEnumAverage),
						string(armmonitor.AggregationTypeEnumCount),
						string(armmonitor.AggregationTypeEnumMaximum),
						string(armmonitor.AggregationTypeEnumMinimum),
						string(armmonitor.AggregationTypeEnumTotal),
					}, target.Aggregations)
				}
			}
		}
		if target.ResourceID == "/subscriptions/subscriptionID/resourceGroups/resourceGroup2/providers/Microsoft.Test/type1/resource3" {
			require.Equal(t, []string{"metric3"}, target.Metrics)
			require.Equal(t, []string{string(armmonitor.AggregationTypeEnumTotal), string(armmonitor.AggregationTypeEnumAverage)}, target.Aggregations)
		}
	}
}

func TestInit_SubscriptionTargetsOnly(t *testing.T) {
	file, err := os.ReadFile("testdata/toml/init_subscription_targets_only.toml")
	require.NoError(t, err)
	require.NotNil(t, file)
	require.NotEmpty(t, file)

	var am *AzureMonitor
	require.NoError(t, toml.Unmarshal(file, &am))

	am.Log = testutil.Logger{}
	am.factory = &mockClientFactory{}

	for index := 1; index <= 27; index++ {
		if index <= 10 {
			am.SubscriptionTargets[4].Metrics = append(am.SubscriptionTargets[4].Metrics, "metric1")
		} else if index <= 23 {
			am.SubscriptionTargets[4].Metrics = append(am.SubscriptionTargets[4].Metrics, "metric2")
		} else {
			am.SubscriptionTargets[4].Metrics = append(am.SubscriptionTargets[4].Metrics, "metric3")
		}
	}

	var expectedResourceMetrics []string
	for index := 1; index <= maxMetricsBatch; index++ {
		if index <= 10 {
			expectedResourceMetrics = append(expectedResourceMetrics, "metric1")
		} else {
			expectedResourceMetrics = append(expectedResourceMetrics, "metric2")
		}
	}

	require.NoError(t, am.Init())
	require.NoError(t, am.Start(nil))
	require.Len(t, am.receiver.resources, 11)

	for _, target := range am.receiver.resources {
		require.Contains(t, []string{
			"/subscriptions/subscriptionID/resourceGroups/resourceGroup1/providers/Microsoft.Test/type1/resource1",
			"/subscriptions/subscriptionID/resourceGroups/resourceGroup1/providers/Microsoft.Test/type2/resource2",
			"/subscriptions/subscriptionID/resourceGroups/resourceGroup2/providers/Microsoft.Test/type1/resource3"}, target.ResourceID)

		if target.ResourceID == "/subscriptions/subscriptionID/resourceGroups/resourceGroup1/providers/Microsoft.Test/type1/resource1" {
			require.Contains(t, []int{1, 2, 3, 4, maxMetricsBatch}, len(target.Metrics))

			if len(target.Metrics) == 1 {
				require.Equal(t, []string{"metric3"}, target.Metrics)
				require.Equal(t, []string{
					string(armmonitor.AggregationTypeEnumTotal),
					string(armmonitor.AggregationTypeEnumAverage),
				}, target.Aggregations)
			}
			if len(target.Metrics) == 2 {
				require.Equal(t, []string{"metric1", "metric2"}, target.Metrics)
				require.Equal(t, []string{string(armmonitor.AggregationTypeEnumTotal), string(armmonitor.AggregationTypeEnumAverage)}, target.Aggregations)
			}
			if len(target.Metrics) == 3 {
				require.Equal(t, []string{"metric2", "metric2", "metric2"}, target.Metrics)
				require.Equal(t, []string{
					string(armmonitor.AggregationTypeEnumAverage),
					string(armmonitor.AggregationTypeEnumCount),
					string(armmonitor.AggregationTypeEnumMaximum),
					string(armmonitor.AggregationTypeEnumMinimum),
					string(armmonitor.AggregationTypeEnumTotal),
				}, target.Aggregations)
			}
			if len(target.Metrics) == 4 {
				require.Equal(t, []string{"metric3", "metric3", "metric3", "metric3"}, target.Metrics)
				require.Equal(t, []string{
					string(armmonitor.AggregationTypeEnumAverage),
					string(armmonitor.AggregationTypeEnumCount),
					string(armmonitor.AggregationTypeEnumMaximum),
					string(armmonitor.AggregationTypeEnumMinimum),
					string(armmonitor.AggregationTypeEnumTotal),
				}, target.Aggregations)
			}
			if len(target.Metrics) == maxMetricsBatch {
				require.Equal(t, expectedResourceMetrics, target.Metrics)
				require.Equal(t, []string{
					string(armmonitor.AggregationTypeEnumAverage),
					string(armmonitor.AggregationTypeEnumCount),
					string(armmonitor.AggregationTypeEnumMaximum),
					string(armmonitor.AggregationTypeEnumMinimum),
					string(armmonitor.AggregationTypeEnumTotal),
				}, target.Aggregations)
			}
		}
		if target.ResourceID == "/subscriptions/subscriptionID/resourceGroups/resourceGroup1/providers/Microsoft.Test/type2/resource2" {
			require.Contains(t, []int{1, 2}, len(target.Metrics))

			if len(target.Metrics) == 1 {
				require.Equal(t, []string{"metric2"}, target.Metrics)
				require.Equal(t, []string{
					string(armmonitor.AggregationTypeEnumAverage),
					string(armmonitor.AggregationTypeEnumCount),
					string(armmonitor.AggregationTypeEnumMaximum),
					string(armmonitor.AggregationTypeEnumMinimum),
					string(armmonitor.AggregationTypeEnumTotal),
				}, target.Aggregations)
			}
			if len(target.Metrics) == 2 {
				require.Contains(t, []int{1, 5}, len(target.Aggregations))

				if len(target.Aggregations) == 1 {
					require.Equal(t, []string{"metric1", "metric2"}, target.Metrics)
					require.Equal(t, []string{string(armmonitor.AggregationTypeEnumAverage)}, target.Aggregations)
				}
				if len(target.Aggregations) == 5 {
					require.Equal(t, []string{"metric1", "metric2"}, target.Metrics)
					require.Equal(t, []string{
						string(armmonitor.AggregationTypeEnumAverage),
						string(armmonitor.AggregationTypeEnumCount),
						string(armmonitor.AggregationTypeEnumMaximum),
						string(armmonitor.AggregationTypeEnumMinimum),
						string(armmonitor.AggregationTypeEnumTotal),
					}, target.Aggregations)
				}
			}
		}
		if target.ResourceID == "/subscriptions/subscriptionID/resourceGroups/resourceGroup2/providers/Microsoft.Test/type1/resource3" {
			require.Contains(t, []int{3, 7, maxMetricsBatch}, len(target.Metrics))

			if len(target.Metrics) == 3 {
				require.Equal(t, []string{"metric1", "metric2", "metric3"}, target.Metrics)
				require.Equal(t, []string{string(armmonitor.AggregationTypeEnumTotal), string(armmonitor.AggregationTypeEnumAverage)}, target.Aggregations)
			}
			if len(target.Metrics) == 7 {
				require.Equal(t, []string{"metric2", "metric2", "metric2", "metric3", "metric3", "metric3", "metric3"}, target.Metrics)
				require.Equal(t, []string{
					string(armmonitor.AggregationTypeEnumAverage),
					string(armmonitor.AggregationTypeEnumCount),
					string(armmonitor.AggregationTypeEnumMaximum),
					string(armmonitor.AggregationTypeEnumMinimum),
					string(armmonitor.AggregationTypeEnumTotal),
				}, target.Aggregations)
			}
			if len(target.Metrics) == maxMetricsBatch {
				require.Equal(t, expectedResourceMetrics, target.Metrics)
				require.Equal(t, []string{
					string(armmonitor.AggregationTypeEnumAverage),
					string(armmonitor.AggregationTypeEnumCount),
					string(armmonitor.AggregationTypeEnumMaximum),
					string(armmonitor.AggregationTypeEnumMinimum),
					string(armmonitor.AggregationTypeEnumTotal),
				}, target.Aggregations)
			}
		}
	}
}

func TestInit_AllTargetTypes(t *testing.T) {
	file, err := os.ReadFile("testdata/toml/init_all_target_types.toml")
	require.NoError(t, err)
	require.NotNil(t, file)
	require.NotEmpty(t, file)

	var am *AzureMonitor
	require.NoError(t, toml.Unmarshal(file, &am))

	am.Log = testutil.Logger{}
	am.factory = &mockClientFactory{}

	for index := 1; index <= 27; index++ {
		if index <= 10 {
			am.ResourceTargets[4].Metrics = append(am.ResourceTargets[4].Metrics, "metric1")
			am.ResourceGroupTargets[0].Resources[1].Metrics = append(am.ResourceGroupTargets[0].Resources[1].Metrics, "metric1")
			am.SubscriptionTargets[4].Metrics = append(am.SubscriptionTargets[4].Metrics, "metric1")
		} else if index <= 23 {
			am.ResourceTargets[4].Metrics = append(am.ResourceTargets[4].Metrics, "metric2")
			am.ResourceGroupTargets[0].Resources[1].Metrics = append(am.ResourceGroupTargets[0].Resources[1].Metrics, "metric2")
			am.SubscriptionTargets[4].Metrics = append(am.SubscriptionTargets[4].Metrics, "metric2")
		} else {
			am.ResourceTargets[4].Metrics = append(am.ResourceTargets[4].Metrics, "metric3")
			am.ResourceGroupTargets[0].Resources[1].Metrics = append(am.ResourceGroupTargets[0].Resources[1].Metrics, "metric3")
			am.SubscriptionTargets[4].Metrics = append(am.SubscriptionTargets[4].Metrics, "metric3")
		}
	}

	require.NoError(t, am.Init())
	require.NoError(t, am.Start(nil))
	require.Len(t, am.receiver.resources, 28)
}

func TestInit_NoSubscriptionID(t *testing.T) {
	file, err := os.ReadFile("testdata/toml/init_no_subscription_id.toml")
	require.NoError(t, err)
	require.NotNil(t, file)
	require.NotEmpty(t, file)

	var am *AzureMonitor
	require.NoError(t, toml.Unmarshal(file, &am))

	am.Log = testutil.Logger{}
	am.factory = &mockClientFactory{}

	require.Error(t, am.Init())
}

func TestInit_NoClientID(t *testing.T) {
	file, err := os.ReadFile("testdata/toml/init_no_client_id.toml")
	require.NoError(t, err)
	require.NotNil(t, file)
	require.NotEmpty(t, file)

	var am *AzureMonitor
	require.NoError(t, toml.Unmarshal(file, &am))

	am.Log = testutil.Logger{}
	am.factory = &mockClientFactory{}

	require.Error(t, am.Init())
}

func TestInit_NoTenantID(t *testing.T) {
	file, err := os.ReadFile("testdata/toml/init_no_tenant_id.toml")
	require.NoError(t, err)
	require.NotNil(t, file)
	require.NotEmpty(t, file)

	var am *AzureMonitor
	require.NoError(t, toml.Unmarshal(file, &am))

	am.Log = testutil.Logger{}
	am.factory = &mockClientFactory{}

	require.Error(t, am.Init())
}

func TestInit_NoTargets(t *testing.T) {
	file, err := os.ReadFile("testdata/toml/init_no_targets.toml")
	require.NoError(t, err)
	require.NotNil(t, file)
	require.NotEmpty(t, file)

	var am *AzureMonitor
	require.NoError(t, toml.Unmarshal(file, &am))

	am.Log = testutil.Logger{}
	am.factory = &mockClientFactory{}

	require.Error(t, am.Init())
}

func TestInit_ResourceTargetWithoutResourceID(t *testing.T) {
	file, err := os.ReadFile("testdata/toml/init_resource_target_without_resource_id.toml")
	require.NoError(t, err)
	require.NotNil(t, file)
	require.NotEmpty(t, file)

	var am *AzureMonitor
	require.NoError(t, toml.Unmarshal(file, &am))

	am.Log = testutil.Logger{}
	am.factory = &mockClientFactory{}

	require.Error(t, am.Init())
}

func TestInit_ResourceTargetWithInvalidResourceID(t *testing.T) {
	file, err := os.ReadFile("testdata/toml/init_resource_target_with_invalid_resource_id.toml")
	require.NoError(t, err)
	require.NotNil(t, file)
	require.NotEmpty(t, file)

	var am *AzureMonitor
	require.NoError(t, toml.Unmarshal(file, &am))

	am.Log = testutil.Logger{}
	am.factory = &mockClientFactory{}

	require.NoError(t, am.Init())
	require.Error(t, am.Start(nil))
}

func TestInit_ResourceTargetWithInvalidMetric(t *testing.T) {
	file, err := os.ReadFile("testdata/toml/init_resource_target_with_invalid_metric.toml")
	require.NoError(t, err)
	require.NotNil(t, file)
	require.NotEmpty(t, file)

	var am *AzureMonitor
	require.NoError(t, toml.Unmarshal(file, &am))

	am.Log = testutil.Logger{}
	am.factory = &mockClientFactory{}

	require.NoError(t, am.Init())
	require.Error(t, am.Start(nil))
}

func TestInit_ResourceTargetWithInvalidAggregation(t *testing.T) {
	file, err := os.ReadFile("testdata/toml/init_resource_target_with_invalid_aggregation.toml")
	require.NoError(t, err)
	require.NotNil(t, file)
	require.NotEmpty(t, file)

	var am *AzureMonitor
	require.NoError(t, toml.Unmarshal(file, &am))

	am.Log = testutil.Logger{}
	am.factory = &mockClientFactory{}

	require.NoError(t, am.Init())
	require.Error(t, am.Start(nil))
}

func TestInit_ResourceGroupTargetWithoutResourceGroup(t *testing.T) {
	file, err := os.ReadFile("testdata/toml/init_resource_group_target_without_resource_group.toml")
	require.NoError(t, err)
	require.NotNil(t, file)
	require.NotEmpty(t, file)

	var am *AzureMonitor
	require.NoError(t, toml.Unmarshal(file, &am))

	am.Log = testutil.Logger{}
	am.factory = &mockClientFactory{}

	require.Error(t, am.Init())
}

func TestInit_ResourceGroupTargetWithResourceWithoutResourceType(t *testing.T) {
	file, err := os.ReadFile("testdata/toml/init_resource_group_target_with_resource_without_resource_type.toml")
	require.NoError(t, err)
	require.NotNil(t, file)
	require.NotEmpty(t, file)

	var am *AzureMonitor
	require.NoError(t, toml.Unmarshal(file, &am))

	am.Log = testutil.Logger{}
	am.factory = &mockClientFactory{}

	require.Error(t, am.Init())
}

func TestInit_ResourceGroupTargetWithInvalidResourceGroup(t *testing.T) {
	file, err := os.ReadFile("testdata/toml/init_resource_group_target_with_invalid_resource_group.toml")
	require.NoError(t, err)
	require.NotNil(t, file)
	require.NotEmpty(t, file)

	var am *AzureMonitor
	require.NoError(t, toml.Unmarshal(file, &am))

	am.Log = testutil.Logger{}
	am.factory = &mockClientFactory{}

	require.NoError(t, am.Init())
	require.Error(t, am.Start(nil))
}

func TestInit_ResourceGroupTargetWithInvalidResourceType(t *testing.T) {
	file, err := os.ReadFile("testdata/toml/init_resource_group_target_with_invalid_resource_type.toml")
	require.NoError(t, err)
	require.NotNil(t, file)
	require.NotEmpty(t, file)

	var am *AzureMonitor
	require.NoError(t, toml.Unmarshal(file, &am))

	am.Log = testutil.Logger{}
	am.factory = &mockClientFactory{}

	require.NoError(t, am.Init())
	require.Error(t, am.Start(nil))
}

func TestInit_ResourceGroupTargetWithInvalidMetric(t *testing.T) {
	file, err := os.ReadFile("testdata/toml/init_resource_group_target_with_invalid_metric.toml")
	require.NoError(t, err)
	require.NotNil(t, file)
	require.NotEmpty(t, file)

	var am *AzureMonitor
	require.NoError(t, toml.Unmarshal(file, &am))

	am.Log = testutil.Logger{}
	am.factory = &mockClientFactory{}

	require.NoError(t, am.Init())
	require.Error(t, am.Start(nil))
}

func TestInit_ResourceGroupTargetWithInvalidAggregation(t *testing.T) {
	file, err := os.ReadFile("testdata/toml/init_resource_group_target_with_invalid_aggregation.toml")
	require.NoError(t, err)
	require.NotNil(t, file)
	require.NotEmpty(t, file)

	var am *AzureMonitor
	require.NoError(t, toml.Unmarshal(file, &am))

	am.Log = testutil.Logger{}
	am.factory = &mockClientFactory{}

	require.NoError(t, am.Init())
	require.Error(t, am.Start(nil))
}

func TestInit_ResourceGroupTargetWithoutResources(t *testing.T) {
	file, err := os.ReadFile("testdata/toml/init_resource_group_target_without_resources.toml")
	require.NoError(t, err)
	require.NotNil(t, file)
	require.NotEmpty(t, file)

	var am *AzureMonitor
	require.NoError(t, toml.Unmarshal(file, &am))

	am.Log = testutil.Logger{}
	am.factory = &mockClientFactory{}

	require.Error(t, am.Init())
}

func TestInit_ResourceGroupTargetNoResourceFound(t *testing.T) {
	file, err := os.ReadFile("testdata/toml/init_resource_group_target_no_resource_found.toml")
	require.NoError(t, err)
	require.NotNil(t, file)
	require.NotEmpty(t, file)

	var am *AzureMonitor
	require.NoError(t, toml.Unmarshal(file, &am))

	am.Log = testutil.Logger{}
	am.factory = &mockClientFactory{}

	require.NoError(t, am.Init())
	require.Error(t, am.Start(nil))
}

func TestInit_SubscriptionTargetWithoutResourceType(t *testing.T) {
	file, err := os.ReadFile("testdata/toml/init_subscription_target_without_resource_type.toml")
	require.NoError(t, err)
	require.NotNil(t, file)
	require.NotEmpty(t, file)

	var am *AzureMonitor
	require.NoError(t, toml.Unmarshal(file, &am))

	am.Log = testutil.Logger{}
	am.factory = &mockClientFactory{}

	require.Error(t, am.Init())
}

func TestInit_SubscriptionTargetWithInvalidResourceType(t *testing.T) {
	file, err := os.ReadFile("testdata/toml/init_subscription_target_with_invalid_resource_type.toml")
	require.NoError(t, err)
	require.NotNil(t, file)
	require.NotEmpty(t, file)

	var am *AzureMonitor
	require.NoError(t, toml.Unmarshal(file, &am))

	am.Log = testutil.Logger{}
	am.factory = &mockClientFactory{}

	require.NoError(t, am.Init())
	require.Error(t, am.Start(nil))
}

func TestInit_SubscriptionTargetWithInvalidMetric(t *testing.T) {
	file, err := os.ReadFile("testdata/toml/init_subscription_target_with_invalid_metric.toml")
	require.NoError(t, err)
	require.NotNil(t, file)
	require.NotEmpty(t, file)

	var am *AzureMonitor
	require.NoError(t, toml.Unmarshal(file, &am))

	am.Log = testutil.Logger{}
	am.factory = &mockClientFactory{}

	require.NoError(t, am.Init())
	require.Error(t, am.Start(nil))
}

func TestInit_SubscriptionTargetWithInvalidAggregation(t *testing.T) {
	file, err := os.ReadFile("testdata/toml/init_subscription_target_with_invalid_aggregation.toml")
	require.NoError(t, err)
	require.NotNil(t, file)
	require.NotEmpty(t, file)

	var am *AzureMonitor
	require.NoError(t, toml.Unmarshal(file, &am))

	am.Log = testutil.Logger{}
	am.factory = &mockClientFactory{}

	require.NoError(t, am.Init())
	require.Error(t, am.Start(nil))
}

func TestInit_SubscriptionTargetNoResourceFound(t *testing.T) {
	file, err := os.ReadFile("testdata/toml/init_subscription_target_no_resource_found.toml")
	require.NoError(t, err)
	require.NotNil(t, file)
	require.NotEmpty(t, file)

	var am *AzureMonitor
	require.NoError(t, toml.Unmarshal(file, &am))

	am.Log = testutil.Logger{}
	am.factory = &mockClientFactory{}

	require.NoError(t, am.Init())
	require.Error(t, am.Start(nil))
}

func TestStartBadCredentials(t *testing.T) {
	file, err := os.ReadFile("testdata/toml/init_bad_credentials.toml")
	require.NoError(t, err)
	require.NotNil(t, file)
	require.NotEmpty(t, file)

	var am *AzureMonitor
	require.NoError(t, toml.Unmarshal(file, &am))

	am.Log = testutil.Logger{}
	am.factory = &azureFactory{}
	require.NoError(t, am.Init())
	require.Error(t, am.Start(nil))
}

func TestGather_Success(t *testing.T) {
	file, err := os.ReadFile("testdata/toml/gather_success.toml")
	require.NoError(t, err)
	require.NotNil(t, file)
	require.NotEmpty(t, file)

	var am *AzureMonitor
	require.NoError(t, toml.Unmarshal(file, &am))

	am.Log = testutil.Logger{}
	am.factory = &mockClientFactory{}

	require.NoError(t, am.Init())
	require.NoError(t, am.Start(nil))

	expectedResource1Metric1Name := "azure_monitor_microsoft_test_type1_metric1"
	expectedResource1Metric1MetricFields := make(map[string]interface{})
	expectedResource1Metric1MetricFields["timeStamp"] = "2022-02-22T22:59:00Z"
	expectedResource1Metric1MetricFields["total"] = 5.0
	expectedResource1Metric1MetricFields["maximum"] = 5.0
	expectedResource1Metric2Name := "azure_monitor_microsoft_test_type1_metric2"
	expectedResource1Metric2MetricFields := make(map[string]interface{})
	expectedResource1Metric2MetricFields["timeStamp"] = "2022-02-22T22:59:00Z"
	expectedResource1Metric2MetricFields["total"] = 2.5
	expectedResource1Metric2MetricFields["maximum"] = 2.5
	expectedResource1MetricsTags := make(map[string]string)
	expectedResource1MetricsTags["subscription_id"] = "subscriptionID"
	expectedResource1MetricsTags["resource_group"] = "resourceGroup1"
	expectedResource1MetricsTags["namespace"] = "Microsoft.Test/type1"
	expectedResource1MetricsTags["resource_name"] = "resource1"
	expectedResource1MetricsTags["resource_region"] = "eastus"
	expectedResource1MetricsTags["unit"] = string(armmonitor.MetricUnitCount)

	expectedResource2Metric1Name := "azure_monitor_microsoft_test_type2_metric1"
	expectedResource2Metric1MetricFields := make(map[string]interface{})
	expectedResource2Metric1MetricFields["timeStamp"] = "2022-02-22T22:59:00Z"
	expectedResource2Metric1MetricFields["total"] = 5.0
	expectedResource2Metric1MetricFields["minimum"] = 2.5
	expectedResource2MetricsTags := make(map[string]string)
	expectedResource2MetricsTags["subscription_id"] = "subscriptionID"
	expectedResource2MetricsTags["resource_group"] = "resourceGroup1"
	expectedResource2MetricsTags["namespace"] = "Microsoft.Test/type2"
	expectedResource2MetricsTags["resource_name"] = "resource2"
	expectedResource2MetricsTags["resource_region"] = "eastus"
	expectedResource2MetricsTags["unit"] = string(armmonitor.MetricUnitCount)

	expectedResource3Metric1Name := "azure_monitor_microsoft_test_type1_metric1"
	expectedResource3Metric1MetricFields := make(map[string]interface{})
	expectedResource3Metric1MetricFields["timeStamp"] = "2022-02-22T22:58:00Z"
	expectedResource3Metric1MetricFields["total"] = 2.5
	expectedResource3Metric1MetricFields["minimum"] = 2.5
	expectedResource3MetricsTags := make(map[string]string)
	expectedResource3MetricsTags["subscription_id"] = "subscriptionID"
	expectedResource3MetricsTags["resource_group"] = "resourceGroup2"
	expectedResource3MetricsTags["namespace"] = "Microsoft.Test/type1"
	expectedResource3MetricsTags["resource_name"] = "resource3"
	expectedResource3MetricsTags["resource_region"] = "eastus"
	expectedResource3MetricsTags["unit"] = string(armmonitor.MetricUnitBytes)

	acc := testutil.Accumulator{}

	require.NoError(t, acc.GatherError(am.Gather))
	require.Len(t, acc.Metrics, 4)

	acc.AssertContainsTaggedFields(t, expectedResource1Metric1Name, expectedResource1Metric1MetricFields, expectedResource1MetricsTags)
	acc.AssertContainsTaggedFields(t, expectedResource1Metric2Name, expectedResource1Metric2MetricFields, expectedResource1MetricsTags)
	acc.AssertContainsTaggedFields(t, expectedResource2Metric1Name, expectedResource2Metric1MetricFields, expectedResource2MetricsTags)
	acc.AssertContainsTaggedFields(t, expectedResource3Metric1Name, expectedResource3Metric1MetricFields, expectedResource3MetricsTags)
}

func TestGather_China_Success(t *testing.T) {
	file, err := os.ReadFile("testdata/toml/gather_success_cloud_option_china.toml")
	require.NoError(t, err)
	require.NotNil(t, file)
	require.NotEmpty(t, file)

	var am *AzureMonitor
	require.NoError(t, toml.Unmarshal(file, &am))

	am.Log = testutil.Logger{}
	am.factory = &mockClientFactory{}

	require.NoError(t, am.Init())
	require.NoError(t, am.Start(nil))
}

func TestGather_Government_Success(t *testing.T) {
	file, err := os.ReadFile("testdata/toml/gather_success_cloud_option_government.toml")
	require.NoError(t, err)
	require.NotNil(t, file)
	require.NotEmpty(t, file)

	var am *AzureMonitor
	require.NoError(t, toml.Unmarshal(file, &am))

	am.Log = testutil.Logger{}
	am.factory = &mockClientFactory{}

	require.NoError(t, am.Init())
	require.NoError(t, am.Start(nil))
}

func TestGather_Public_Success(t *testing.T) {
	file, err := os.ReadFile("testdata/toml/gather_success_cloud_option_public.toml")
	require.NoError(t, err)
	require.NotNil(t, file)
	require.NotEmpty(t, file)

	var am *AzureMonitor
	require.NoError(t, toml.Unmarshal(file, &am))

	am.Log = testutil.Logger{}
	am.factory = &mockClientFactory{}

	require.NoError(t, am.Init())
	require.NoError(t, am.Start(nil))
}

type mockClientFactory struct{}

func (*mockClientFactory) createClient(_, _, _, _ string, _ azcore.ClientOptions) (client, error) {
	return &mockClient{}, nil
}

type mockClient struct{}

func (*mockClient) ResourcesList(context.Context, *armresources.ClientListOptions) ([]*armresources.ClientListResponse, error) {
	file, err := os.ReadFile("testdata/json/azure_resources_response.json")
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	var genericResourcesExpanded []*armresources.GenericResourceExpanded
	if err := json.Unmarshal(file, &genericResourcesExpanded); err != nil {
		return nil, err
	}

	response := &armresources.ClientListResponse{
		ResourceListResult: armresources.ResourceListResult{
			Value: genericResourcesExpanded,
		},
	}

	return []*armresources.ClientListResponse{response}, nil
}

func (*mockClient) ResourcesListByResourceGroup(
	_ context.Context,
	resourceGroup string,
	_ *armresources.ClientListByResourceGroupOptions) ([]*armresources.ClientListByResourceGroupResponse, error) {
	var responses []*armresources.ClientListByResourceGroupResponse

	file, err := os.ReadFile("testdata/json/azure_resources_response.json")
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	var genericResourcesExpanded []*armresources.GenericResourceExpanded
	if err := json.Unmarshal(file, &genericResourcesExpanded); err != nil {
		return nil, err
	}

	if resourceGroup == "resourceGroup1" {
		response := &armresources.ClientListByResourceGroupResponse{
			ResourceListResult: armresources.ResourceListResult{
				Value: []*armresources.GenericResourceExpanded{
					genericResourcesExpanded[0],
					genericResourcesExpanded[1],
				},
			},
		}

		responses = append(responses, response)
		return responses, nil
	}

	if resourceGroup == "resourceGroup2" {
		response := &armresources.ClientListByResourceGroupResponse{
			ResourceListResult: armresources.ResourceListResult{
				Value: []*armresources.GenericResourceExpanded{
					genericResourcesExpanded[2],
				},
			},
		}

		responses = append(responses, response)
		return responses, nil
	}

	return nil, errors.New("resource group was not found")
}

func (*mockClient) MetricDefinitionsList(
	_ context.Context,
	resourceID string,
	_ *armmonitor.MetricDefinitionsClientListOptions) (armmonitor.MetricDefinitionsClientListResponse, error) {
	file, err := os.ReadFile("testdata/json/azure_metric_definitions_responses.json")
	if err != nil {
		return armmonitor.MetricDefinitionsClientListResponse{}, fmt.Errorf("error reading file: %w", err)
	}

	var metricDefinitions [][]*armmonitor.MetricDefinition
	if err := json.Unmarshal(file, &metricDefinitions); err != nil {
		return armmonitor.MetricDefinitionsClientListResponse{}, err
	}

	switch resourceID {
	case "/subscriptions/subscriptionID/resourceGroups/resourceGroup1/providers/Microsoft.Test/type1/resource1":
		return armmonitor.MetricDefinitionsClientListResponse{
			MetricDefinitionCollection: armmonitor.MetricDefinitionCollection{
				Value: metricDefinitions[0],
			},
		}, nil
	case "/subscriptions/subscriptionID/resourceGroups/resourceGroup1/providers/Microsoft.Test/type2/resource2",
		"/subscriptions/subscriptionID/resourceGroups/resourceGroup2/providers/Microsoft.Test/type2/resource4",
		"/subscriptions/subscriptionID/resourceGroups/resourceGroup2/providers/Microsoft.Test/type2/resource5",
		"/subscriptions/subscriptionID/resourceGroups/resourceGroup2/providers/Microsoft.Test/type2/resource6":
		return armmonitor.MetricDefinitionsClientListResponse{
			MetricDefinitionCollection: armmonitor.MetricDefinitionCollection{
				Value: metricDefinitions[1],
			},
		}, nil
	case "/subscriptions/subscriptionID/resourceGroups/resourceGroup2/providers/Microsoft.Test/type1/resource3":
		return armmonitor.MetricDefinitionsClientListResponse{
			MetricDefinitionCollection: armmonitor.MetricDefinitionCollection{
				Value: metricDefinitions[2],
			},
		}, nil
	}

	return armmonitor.MetricDefinitionsClientListResponse{}, errors.New("resource ID was not found")
}

func (*mockClient) MetricsList(
	_ context.Context,
	resourceID string,
	_ *armmonitor.MetricsClientListOptions) (armmonitor.MetricsClientListResponse, error) {
	file, err := os.ReadFile("testdata/json/azure_metrics_responses.json")
	if err != nil {
		return armmonitor.MetricsClientListResponse{}, fmt.Errorf("error reading file: %w", err)
	}

	var metricResponses []armmonitor.Response
	if err := json.Unmarshal(file, &metricResponses); err != nil {
		return armmonitor.MetricsClientListResponse{}, err
	}

	if resourceID == "/subscriptions/subscriptionID/resourceGroups/resourceGroup1/providers/Microsoft.Test/type1/resource1" {
		return armmonitor.MetricsClientListResponse{
			Response: metricResponses[0],
		}, nil
	}

	if resourceID == "/subscriptions/subscriptionID/resourceGroups/resourceGroup1/providers/Microsoft.Test/type2/resource2" {
		return armmonitor.MetricsClientListResponse{
			Response: metricResponses[1],
		}, nil
	}

	if resourceID == "/subscriptions/subscriptionID/resourceGroups/resourceGroup2/providers/Microsoft.Test/type1/resource3" {
		return armmonitor.MetricsClientListResponse{
			Response: metricResponses[2],
		}, nil
	}

	if resourceID == "/subscriptions/subscriptionID/resourceGroups/resourceGroup2/providers/Microsoft.Test/type2/resource4" {
		return armmonitor.MetricsClientListResponse{
			Response: metricResponses[3],
		}, nil
	}

	if resourceID == "/subscriptions/subscriptionID/resourceGroups/resourceGroup2/providers/Microsoft.Test/type2/resource5" {
		return armmonitor.MetricsClientListResponse{
			Response: metricResponses[4],
		}, nil
	}

	if resourceID == "/subscriptions/subscriptionID/resourceGroups/resourceGroup2/providers/Microsoft.Test/type2/resource6" {
		return armmonitor.MetricsClientListResponse{
			Response: metricResponses[5],
		}, nil
	}

	return armmonitor.MetricsClientListResponse{}, errors.New("resource ID was not found")
}
