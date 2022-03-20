package azure_monitor

import (
	"context"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/monitor/armmonitor"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

type mockAzureResourcesClient struct{}

type mockAzureMetricDefinitionsClient struct{}

type mockAzureMetricsClient struct{}

const (
	testSubscriptionID = "subscriptionID"
	testClientID       = "clientID"
	testClientSecret   = "clientSecret"
	testTenantID       = "tenantID"

	testResourceGroup1 = "resourceGroup1"
	testResourceGroup2 = "resourceGroup2"

	testResourceType1 = "Microsoft.Test/type1"
	testResourceType2 = "Microsoft.Test/type2"
	testResourceType3 = "Microsoft.Test/type3"

	testResourceGroup1ResourceType1Resource1     = "resourceGroups/" + testResourceGroup1 + "/providers/" + testResourceType1 + "/resource1"
	testResourceGroup1ResourceType2Resource2     = "resourceGroups/" + testResourceGroup1 + "/providers/" + testResourceType2 + "/resource2"
	testResourceGroup2ResourceType1Resource3     = "resourceGroups/" + testResourceGroup2 + "/providers/" + testResourceType1 + "/resource3"
	testResourceGroup2ResourceType2Resource4     = "resourceGroups/" + testResourceGroup2 + "/providers/" + testResourceType2 + "/resource4"
	testResourceGroup2ResourceType2Resource5     = "resourceGroups/" + testResourceGroup2 + "/providers/" + testResourceType2 + "/resource5"
	testResourceGroup2ResourceType2Resource6     = "resourceGroups/" + testResourceGroup2 + "/providers/" + testResourceType2 + "/resource6"
	testFullResourceGroup1ResourceType1Resource1 = "/subscriptions/" + testSubscriptionID + "/" + testResourceGroup1ResourceType1Resource1
	testFullResourceGroup1ResourceType2Resource2 = "/subscriptions/" + testSubscriptionID + "/" + testResourceGroup1ResourceType2Resource2
	testFullResourceGroup2ResourceType1Resource3 = "/subscriptions/" + testSubscriptionID + "/" + testResourceGroup2ResourceType1Resource3
	testFullResourceGroup2ResourceType2Resource4 = "/subscriptions/" + testSubscriptionID + "/" + testResourceGroup2ResourceType2Resource4
	testFullResourceGroup2ResourceType2Resource5 = "/subscriptions/" + testSubscriptionID + "/" + testResourceGroup2ResourceType2Resource5
	testFullResourceGroup2ResourceType2Resource6 = "/subscriptions/" + testSubscriptionID + "/" + testResourceGroup2ResourceType2Resource6

	testMetric1             = "metric1"
	testMetric2             = "metric2"
	testMetric3             = "metric3"
	testMetric3WithComma    = ",metric3,"
	testMetric3ChangedComma = "%2metric3%2"
	testInvalidMetric       = "invalid"
	testInvalidAggregation  = "Invalid"

	testResourceRegion = "eastus"
)

func setMockAzureClients() *azureClients {
	return &azureClients{
		ctx:                     context.Background(),
		resourcesClient:         &mockAzureResourcesClient{},
		metricDefinitionsClient: &mockAzureMetricDefinitionsClient{},
		metricsClient:           &mockAzureMetricsClient{},
	}
}

func (marc *mockAzureResourcesClient) List(_ context.Context, _ *armresources.ClientListOptions) ([]*armresources.ClientListResponse, error) {
	responses := make([]*armresources.ClientListResponse, 0)
	resourceIDS := make([]string, 0)
	resourceTypes := make([]string, 0)
	resourceIDS = append(resourceIDS,
		testFullResourceGroup1ResourceType1Resource1,
		testFullResourceGroup1ResourceType2Resource2,
		testFullResourceGroup2ResourceType1Resource3)
	resourceTypes = append(resourceTypes, testResourceType1, testResourceType2)
	response := &armresources.ClientListResponse{
		ClientListResult: armresources.ClientListResult{
			ResourceListResult: armresources.ResourceListResult{
				Value: []*armresources.GenericResourceExpanded{
					{
						ID:   &resourceIDS[0],
						Type: &resourceTypes[0],
					},
					{
						ID:   &resourceIDS[1],
						Type: &resourceTypes[1],
					},
					{
						ID:   &resourceIDS[2],
						Type: &resourceTypes[0],
					},
				},
			},
		},
	}

	responses = append(responses, response)
	return responses, nil
}

func (marc *mockAzureResourcesClient) ListByResourceGroup(
	_ context.Context,
	resourceGroup string,
	_ *armresources.ClientListByResourceGroupOptions) ([]*armresources.ClientListByResourceGroupResponse, error) {
	responses := make([]*armresources.ClientListByResourceGroupResponse, 0)
	resourceIDS := make([]string, 0)
	resourceTypes := make([]string, 0)
	resourceIDS = append(resourceIDS,
		testFullResourceGroup1ResourceType1Resource1,
		testFullResourceGroup1ResourceType2Resource2,
		testFullResourceGroup2ResourceType1Resource3)
	resourceTypes = append(resourceTypes, testResourceType1, testResourceType2)

	if resourceGroup == testResourceGroup1 {
		response := &armresources.ClientListByResourceGroupResponse{
			ClientListByResourceGroupResult: armresources.ClientListByResourceGroupResult{
				ResourceListResult: armresources.ResourceListResult{
					Value: []*armresources.GenericResourceExpanded{
						{
							ID:   &resourceIDS[0],
							Type: &resourceTypes[0],
						},
						{
							ID:   &resourceIDS[1],
							Type: &resourceTypes[1],
						},
					},
				},
			},
		}

		responses = append(responses, response)
		return responses, nil
	}

	if resourceGroup == testResourceGroup2 {
		response := &armresources.ClientListByResourceGroupResponse{
			ClientListByResourceGroupResult: armresources.ClientListByResourceGroupResult{
				ResourceListResult: armresources.ResourceListResult{
					Value: []*armresources.GenericResourceExpanded{
						{
							ID:   &resourceIDS[2],
							Type: &resourceTypes[0],
						},
					},
				},
			},
		}

		responses = append(responses, response)
		return responses, nil
	}

	return nil, nil
}

func (mamdc *mockAzureMetricDefinitionsClient) List(
	_ context.Context,
	resourceID string,
	_ *armmonitor.MetricDefinitionsClientListOptions) (armmonitor.MetricDefinitionsClientListResponse, error) {
	metricNames := make([]string, 0)
	timeGrains := make([]string, 0)
	metricNames = append(metricNames, testMetric1, testMetric2, testMetric3)
	timeGrains = append(timeGrains, "PT1M", "PT5M")

	if resourceID == testFullResourceGroup1ResourceType1Resource1 {
		return armmonitor.MetricDefinitionsClientListResponse{
			MetricDefinitionsClientListResult: armmonitor.MetricDefinitionsClientListResult{
				MetricDefinitionCollection: armmonitor.MetricDefinitionCollection{
					Value: []*armmonitor.MetricDefinition{
						{
							ID: &resourceID,
							Name: &armmonitor.LocalizableString{
								Value: &metricNames[0],
							},
							MetricAvailabilities: []*armmonitor.MetricAvailability{
								{
									TimeGrain: &timeGrains[0],
								},
								{
									TimeGrain: &timeGrains[1],
								},
							},
						},
						{
							ID: &resourceID,
							Name: &armmonitor.LocalizableString{
								Value: &metricNames[1],
							},
							MetricAvailabilities: []*armmonitor.MetricAvailability{
								{
									TimeGrain: &timeGrains[0],
								},
								{
									TimeGrain: &timeGrains[1],
								},
							},
						},
						{
							ID: &resourceID,
							Name: &armmonitor.LocalizableString{
								Value: &metricNames[2],
							},
							MetricAvailabilities: []*armmonitor.MetricAvailability{
								{
									TimeGrain: &timeGrains[1],
								},
							},
						},
					},
				},
			},
		}, nil
	}

	if resourceID == testFullResourceGroup1ResourceType2Resource2 {
		return armmonitor.MetricDefinitionsClientListResponse{
			MetricDefinitionsClientListResult: armmonitor.MetricDefinitionsClientListResult{
				MetricDefinitionCollection: armmonitor.MetricDefinitionCollection{
					Value: []*armmonitor.MetricDefinition{
						{
							ID: &resourceID,
							Name: &armmonitor.LocalizableString{
								Value: &metricNames[0],
							},
							MetricAvailabilities: []*armmonitor.MetricAvailability{
								{
									TimeGrain: &timeGrains[0],
								},
								{
									TimeGrain: &timeGrains[1],
								},
							},
						},
						{
							ID: &resourceID,
							Name: &armmonitor.LocalizableString{
								Value: &metricNames[1],
							},
							MetricAvailabilities: []*armmonitor.MetricAvailability{
								{
									TimeGrain: &timeGrains[0],
								},
							},
						},
					},
				},
			},
		}, nil
	}

	if resourceID == testFullResourceGroup2ResourceType1Resource3 {
		return armmonitor.MetricDefinitionsClientListResponse{
			MetricDefinitionsClientListResult: armmonitor.MetricDefinitionsClientListResult{
				MetricDefinitionCollection: armmonitor.MetricDefinitionCollection{
					Value: []*armmonitor.MetricDefinition{
						{
							ID: &resourceID,
							Name: &armmonitor.LocalizableString{
								Value: &metricNames[0],
							},
							MetricAvailabilities: []*armmonitor.MetricAvailability{
								{
									TimeGrain: &timeGrains[0],
								},
								{
									TimeGrain: &timeGrains[1],
								},
							},
						},
						{
							ID: &resourceID,
							Name: &armmonitor.LocalizableString{
								Value: &metricNames[1],
							},
							MetricAvailabilities: []*armmonitor.MetricAvailability{
								{
									TimeGrain: &timeGrains[0],
								},
								{
									TimeGrain: &timeGrains[1],
								},
							},
						},
						{
							ID: &resourceID,
							Name: &armmonitor.LocalizableString{
								Value: &metricNames[2],
							},
							MetricAvailabilities: []*armmonitor.MetricAvailability{
								{
									TimeGrain: &timeGrains[0],
								},
								{
									TimeGrain: &timeGrains[1],
								},
							},
						},
					},
				},
			},
		}, nil
	}

	return armmonitor.MetricDefinitionsClientListResponse{}, nil
}

func (mamc *mockAzureMetricsClient) List(
	_ context.Context,
	resourceID string,
	_ *armmonitor.MetricsClientListOptions) (armmonitor.MetricsClientListResponse, error) {
	namespaces := make([]string, 0)
	metricIDS := make([]string, 0)
	metricNames := make([]string, 0)
	metricUnits := make([]armmonitor.MetricUnit, 0)
	metricTypes := make([]string, 0)
	timeStamps := make([]time.Time, 0)
	aggregationValues := make([]float64, 0)
	namespaces = append(namespaces, testResourceType1, testResourceType2)
	metricIDS = append(metricIDS,
		testFullResourceGroup1ResourceType1Resource1+"/providers/Microsoft.Insights/metrics/metric1",
		testFullResourceGroup1ResourceType1Resource1+"/providers/Microsoft.Insights/metrics/metric2",
		testFullResourceGroup1ResourceType2Resource2+"/providers/Microsoft.Insights/metrics/metric1",
		testFullResourceGroup2ResourceType1Resource3+"/providers/Microsoft.Insights/metrics/metric1",
		testFullResourceGroup2ResourceType2Resource4+"/providers/Microsoft.Insights/metrics/metric1",
		testFullResourceGroup2ResourceType2Resource5+"/providers/Microsoft.Insights/metrics/metric2",
		testFullResourceGroup2ResourceType2Resource6+"/providers/Microsoft.Insights/metrics/metric2")
	metricNames = append(metricNames, testMetric1, testMetric2, testMetric3)
	metricUnits = append(metricUnits, armmonitor.MetricUnitCount, armmonitor.MetricUnitBytes)
	metricTypes = append(metricTypes, testResourceType1, testResourceType2)
	timeStamps = append(timeStamps,
		time.Date(2022, 2, 22, 22, 0, 0, 0, time.UTC),
		time.Date(2022, 2, 22, 22, 1, 0, 0, time.UTC),
		time.Date(2022, 2, 22, 22, 2, 0, 0, time.UTC),
		time.Date(2022, 2, 22, 22, 58, 0, 0, time.UTC),
		time.Date(2022, 2, 22, 22, 59, 0, 0, time.UTC))
	aggregationValues = append(aggregationValues, 1.0, 2.0, 2.5, 3.0, 5.0)
	resourceRegion := testResourceRegion
	metricErrorCode := "Success"

	if resourceID == testFullResourceGroup1ResourceType1Resource1 {
		return armmonitor.MetricsClientListResponse{
			MetricsClientListResult: armmonitor.MetricsClientListResult{
				Response: armmonitor.Response{
					Namespace:      &namespaces[0],
					Resourceregion: &resourceRegion,
					Value: []*armmonitor.Metric{
						{
							ID: &metricIDS[0],
							Name: &armmonitor.LocalizableString{
								LocalizedValue: &metricNames[0],
							},
							Unit: &metricUnits[0],
							Timeseries: []*armmonitor.TimeSeriesElement{
								{
									Data: []*armmonitor.MetricValue{
										{
											TimeStamp: &timeStamps[0],
											Total:     &aggregationValues[0],
											Maximum:   &aggregationValues[0],
										},
										{
											TimeStamp: &timeStamps[1],
											Total:     &aggregationValues[1],
											Maximum:   &aggregationValues[1],
										},
										{
											TimeStamp: &timeStamps[2],
											Total:     &aggregationValues[2],
											Maximum:   &aggregationValues[2],
										},
										{
											TimeStamp: &timeStamps[3],
											Total:     &aggregationValues[1],
											Maximum:   &aggregationValues[2],
										},
										{
											TimeStamp: &timeStamps[4],
											Total:     &aggregationValues[4],
											Maximum:   &aggregationValues[4],
										},
									},
								},
							},
							ErrorCode: &metricErrorCode,
						},
						{
							ID: &metricIDS[1],
							Name: &armmonitor.LocalizableString{
								LocalizedValue: &metricNames[1],
							},
							Unit: &metricUnits[0],
							Timeseries: []*armmonitor.TimeSeriesElement{
								{
									Data: []*armmonitor.MetricValue{
										{
											TimeStamp: &timeStamps[0],
											Total:     &aggregationValues[1],
											Maximum:   &aggregationValues[1],
										},
										{
											TimeStamp: &timeStamps[1],
											Total:     &aggregationValues[0],
											Maximum:   &aggregationValues[1],
										},
										{
											TimeStamp: &timeStamps[2],
											Total:     &aggregationValues[0],
											Maximum:   &aggregationValues[1],
										},
										{
											TimeStamp: &timeStamps[3],
											Total:     &aggregationValues[2],
											Maximum:   &aggregationValues[2],
										},
										{
											TimeStamp: &timeStamps[4],
											Total:     &aggregationValues[2],
											Maximum:   &aggregationValues[2],
										},
									},
								},
							},
							ErrorCode: &metricErrorCode,
						},
					},
				},
			},
		}, nil
	}

	if resourceID == testFullResourceGroup1ResourceType2Resource2 {
		return armmonitor.MetricsClientListResponse{
			MetricsClientListResult: armmonitor.MetricsClientListResult{
				Response: armmonitor.Response{
					Namespace:      &namespaces[1],
					Resourceregion: &resourceRegion,
					Value: []*armmonitor.Metric{
						{
							ID: &metricIDS[2],
							Name: &armmonitor.LocalizableString{
								LocalizedValue: &metricNames[0],
							},
							Unit: &metricUnits[0],
							Timeseries: []*armmonitor.TimeSeriesElement{
								{
									Data: []*armmonitor.MetricValue{
										{
											TimeStamp: &timeStamps[0],
											Total:     &aggregationValues[4],
											Minimum:   &aggregationValues[4],
										},
										{
											TimeStamp: &timeStamps[1],
											Total:     &aggregationValues[3],
											Minimum:   &aggregationValues[3],
										},
										{
											TimeStamp: &timeStamps[2],
											Total:     &aggregationValues[4],
											Minimum:   &aggregationValues[3],
										},
										{
											TimeStamp: &timeStamps[3],
											Total:     &aggregationValues[2],
											Minimum:   &aggregationValues[2],
										},
										{
											TimeStamp: &timeStamps[4],
											Total:     &aggregationValues[4],
											Minimum:   &aggregationValues[2],
										},
									},
								},
							},
							ErrorCode: &metricErrorCode,
						},
					},
				},
			},
		}, nil
	}

	if resourceID == testFullResourceGroup2ResourceType1Resource3 {
		return armmonitor.MetricsClientListResponse{
			MetricsClientListResult: armmonitor.MetricsClientListResult{
				Response: armmonitor.Response{
					Namespace:      &namespaces[0],
					Resourceregion: &resourceRegion,
					Value: []*armmonitor.Metric{
						{
							ID: &metricIDS[3],
							Name: &armmonitor.LocalizableString{
								LocalizedValue: &metricNames[0],
							},
							Unit: &metricUnits[1],
							Timeseries: []*armmonitor.TimeSeriesElement{
								{
									Data: []*armmonitor.MetricValue{
										{
											TimeStamp: &timeStamps[0],
											Total:     &aggregationValues[4],
											Minimum:   &aggregationValues[4],
										},
										{
											TimeStamp: &timeStamps[1],
											Total:     &aggregationValues[3],
											Minimum:   &aggregationValues[3],
										},
										{
											TimeStamp: &timeStamps[2],
											Total:     &aggregationValues[4],
											Minimum:   &aggregationValues[3],
										},
										{
											TimeStamp: &timeStamps[3],
											Total:     &aggregationValues[2],
											Minimum:   &aggregationValues[2],
										},
										{
											TimeStamp: &timeStamps[4],
											Total:     nil,
											Minimum:   nil,
										},
									},
								},
							},
							ErrorCode: &metricErrorCode,
						},
					},
				},
			},
		}, nil
	}

	if resourceID == testFullResourceGroup2ResourceType2Resource4 {
		return armmonitor.MetricsClientListResponse{
			MetricsClientListResult: armmonitor.MetricsClientListResult{
				Response: armmonitor.Response{
					Namespace:      &namespaces[1],
					Resourceregion: &resourceRegion,
					Value: []*armmonitor.Metric{
						{
							ID: &metricIDS[4],
							Name: &armmonitor.LocalizableString{
								LocalizedValue: &metricNames[0],
							},
							Unit: &metricUnits[1],
							Timeseries: []*armmonitor.TimeSeriesElement{
								{
									Data: []*armmonitor.MetricValue{
										{
											TimeStamp: &timeStamps[0],
											Total:     nil,
											Maximum:   nil,
										},
										{
											TimeStamp: &timeStamps[1],
											Total:     nil,
											Maximum:   nil,
										},
										{
											TimeStamp: &timeStamps[2],
											Total:     nil,
											Maximum:   nil,
										},
										{
											TimeStamp: &timeStamps[3],
											Total:     nil,
											Maximum:   nil,
										},
										{
											TimeStamp: &timeStamps[4],
											Total:     nil,
											Maximum:   nil,
										},
									},
								},
							},
							ErrorCode: &metricErrorCode,
						},
					},
				},
			},
		}, nil
	}

	if resourceID == testFullResourceGroup2ResourceType2Resource5 {
		return armmonitor.MetricsClientListResponse{
			MetricsClientListResult: armmonitor.MetricsClientListResult{
				Response: armmonitor.Response{
					Namespace:      &namespaces[1],
					Resourceregion: &resourceRegion,
					Value: []*armmonitor.Metric{
						{
							ID: &metricIDS[5],
							Name: &armmonitor.LocalizableString{
								LocalizedValue: &metricNames[0],
							},
							Unit: &metricUnits[1],
							Timeseries: []*armmonitor.TimeSeriesElement{
								{
									Data: []*armmonitor.MetricValue{},
								},
							},
							ErrorCode: &metricErrorCode,
						},
					},
				},
			},
		}, nil
	}

	if resourceID == testFullResourceGroup2ResourceType2Resource6 {
		return armmonitor.MetricsClientListResponse{
			MetricsClientListResult: armmonitor.MetricsClientListResult{
				Response: armmonitor.Response{
					Namespace:      &namespaces[1],
					Resourceregion: &resourceRegion,
					Value: []*armmonitor.Metric{
						{
							ID: &metricIDS[6],
							Name: &armmonitor.LocalizableString{
								LocalizedValue: &metricNames[0],
							},
							Unit:       &metricUnits[1],
							Timeseries: []*armmonitor.TimeSeriesElement{},
							ErrorCode:  &metricErrorCode,
						},
					},
				},
			},
		}, nil
	}

	return armmonitor.MetricsClientListResponse{}, nil
}
