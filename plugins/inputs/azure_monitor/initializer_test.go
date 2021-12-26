package azure_monitor

import (
	"strconv"
	"strings"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckConfigValidation_ResourceTargetsOnly(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: "subscriptionID",
		ClientID:       "clientID",
		ClientSecret:   "clientSecret",
		TenantID:       "tenantID",
		ResourceTargets: []*ResourceTarget{
			newResourceTarget("resource", []string{}, []string{}),
		},
		Log:         testutil.Logger{},
		azureClient: newAzureClient(),
	}

	err := am.checkConfigValidation()
	require.NoError(t, err)
}

func TestCheckConfigValidation_ResourceTargetWithNoResourceID(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: "subscriptionID",
		ClientID:       "clientID",
		ClientSecret:   "clientSecret",
		TenantID:       "tenantID",
		ResourceTargets: []*ResourceTarget{
			newResourceTarget("", []string{}, []string{}),
		},
		Log:         testutil.Logger{},
		azureClient: newAzureClient(),
	}

	err := am.checkConfigValidation()
	require.Error(t, err)
}

func TestCheckConfigValidation_ResourceTargetWithInvalidAggregation(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: "subscriptionID",
		ClientID:       "clientID",
		ClientSecret:   "clientSecret",
		TenantID:       "tenantID",
		ResourceTargets: []*ResourceTarget{
			newResourceTarget("resource", []string{}, []string{"invalidAggregation"}),
		},
		Log:         testutil.Logger{},
		azureClient: newAzureClient(),
	}

	err := am.checkConfigValidation()
	require.Error(t, err)
}

func TestCheckConfigValidation_ResourceGroupTargetsOnly(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: "subscriptionID",
		ClientID:       "clientID",
		ClientSecret:   "clientSecret",
		TenantID:       "tenantID",
		ResourceGroupTargets: []*ResourceGroupTarget{
			newResourceGroupTarget("resourceGroup",
				[]*Resource{
					{
						ResourceType: "resourceType",
						Metrics:      []string{},
						Aggregations: []string{},
					},
				},
			)},
		Log:         testutil.Logger{},
		azureClient: newAzureClient(),
	}

	err := am.checkConfigValidation()
	require.NoError(t, err)
}

func TestCheckConfigValidation_ResourceGroupTargetWithNoResource(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: "subscriptionID",
		ClientID:       "clientID",
		ClientSecret:   "clientSecret",
		TenantID:       "tenantID",
		ResourceGroupTargets: []*ResourceGroupTarget{
			newResourceGroupTarget("resourceGroup",
				[]*Resource{},
			)},
		Log:         testutil.Logger{},
		azureClient: newAzureClient(),
	}

	err := am.checkConfigValidation()
	require.Error(t, err)
}

func TestCheckConfigValidation_ResourceGroupTargetWithResourceWithoutResourceType(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: "subscriptionID",
		ClientID:       "clientID",
		ClientSecret:   "clientSecret",
		TenantID:       "tenantID",
		ResourceGroupTargets: []*ResourceGroupTarget{
			newResourceGroupTarget("resourceGroup",
				[]*Resource{
					{
						ResourceType: "",
						Metrics:      []string{},
						Aggregations: []string{},
					},
				},
			)},
		Log:         testutil.Logger{},
		azureClient: newAzureClient(),
	}

	err := am.checkConfigValidation()
	require.Error(t, err)
}

func TestCheckConfigValidation_ResourceGroupTargetWithInvalidAggregation(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: "subscriptionID",
		ClientID:       "clientID",
		ClientSecret:   "clientSecret",
		TenantID:       "tenantID",
		ResourceGroupTargets: []*ResourceGroupTarget{
			newResourceGroupTarget("resourceGroup",
				[]*Resource{
					{
						ResourceType: "resourceType",
						Metrics:      []string{},
						Aggregations: []string{"invalidAggregation"},
					},
				},
			)},
		Log:         testutil.Logger{},
		azureClient: newAzureClient(),
	}

	err := am.checkConfigValidation()
	require.Error(t, err)
}

func TestCheckConfigValidation_SubscriptionTargetsOnly(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: "subscriptionID",
		ClientID:       "clientID",
		ClientSecret:   "clientSecret",
		TenantID:       "tenantID",
		SubscriptionTargets: []*Resource{
			{
				ResourceType: "resourceType",
				Metrics:      []string{},
				Aggregations: []string{},
			},
		},
		Log:         testutil.Logger{},
		azureClient: newAzureClient(),
	}

	err := am.checkConfigValidation()
	require.NoError(t, err)
}

func TestCheckConfigValidation_SubscriptionTargetWithoutResourceType(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: "subscriptionID",
		ClientID:       "clientID",
		ClientSecret:   "clientSecret",
		TenantID:       "tenantID",
		SubscriptionTargets: []*Resource{
			{
				ResourceType: "",
				Metrics:      []string{},
				Aggregations: []string{},
			},
		},
		Log:         testutil.Logger{},
		azureClient: newAzureClient(),
	}

	err := am.checkConfigValidation()
	require.Error(t, err)
}

func TestCheckConfigValidation_SubscriptionTargetWithInvalidAggregation(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: "subscriptionID",
		ClientID:       "clientID",
		ClientSecret:   "clientSecret",
		TenantID:       "tenantID",
		SubscriptionTargets: []*Resource{
			{
				ResourceType: "resourceType",
				Metrics:      []string{},
				Aggregations: []string{"invalidAggregation"},
			},
		},
		Log:         testutil.Logger{},
		azureClient: newAzureClient(),
	}

	err := am.checkConfigValidation()
	require.Error(t, err)
}

func TestCheckConfigValidation_AllTargetTypes(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: "subscriptionID",
		ClientID:       "clientID",
		ClientSecret:   "clientSecret",
		TenantID:       "tenantID",
		ResourceTargets: []*ResourceTarget{
			newResourceTarget("resourceID", []string{}, []string{}),
		},
		ResourceGroupTargets: []*ResourceGroupTarget{
			newResourceGroupTarget("resourceGroup",
				[]*Resource{
					{
						ResourceType: "resourceType",
						Metrics:      []string{},
						Aggregations: []string{},
					},
				},
			)},
		SubscriptionTargets: []*Resource{
			{
				ResourceType: "resourceType",
				Metrics:      []string{},
				Aggregations: []string{},
			},
		},
		Log:         testutil.Logger{},
		azureClient: newAzureClient(),
	}

	err := am.checkConfigValidation()
	require.NoError(t, err)
}

func TestCheckConfigValidation_NoSubscriptionID(t *testing.T) {
	am := &AzureMonitor{
		ClientID:     "clientID",
		ClientSecret: "clientSecret",
		TenantID:     "tenantID",
		ResourceTargets: []*ResourceTarget{
			newResourceTarget("resourceID", []string{}, []string{}),
		},
		Log:         testutil.Logger{},
		azureClient: newAzureClient(),
	}

	err := am.checkConfigValidation()
	require.Error(t, err)
}

func TestCheckConfigValidation_NoClientID(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: "subscriptionID",
		ClientSecret:   "clientSecret",
		TenantID:       "tenantID",
		ResourceTargets: []*ResourceTarget{
			newResourceTarget("resourceID", []string{}, []string{}),
		},
		Log:         testutil.Logger{},
		azureClient: newAzureClient(),
	}

	err := am.checkConfigValidation()
	require.Error(t, err)
}

func TestCheckConfigValidation_NoClientSecret(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: "subscriptionID",
		ClientID:       "clientID",
		TenantID:       "tenantID",
		ResourceTargets: []*ResourceTarget{
			newResourceTarget("resourceID", []string{}, []string{}),
		},
		Log:         testutil.Logger{},
		azureClient: newAzureClient(),
	}

	err := am.checkConfigValidation()
	require.Error(t, err)
}

func TestCheckConfigValidation_NoTenantID(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: "subscriptionID",
		ClientID:       "clientID",
		ClientSecret:   "clientSecret",
		ResourceTargets: []*ResourceTarget{
			newResourceTarget("resourceID", []string{}, []string{}),
		},
		Log:         testutil.Logger{},
		azureClient: newAzureClient(),
	}

	err := am.checkConfigValidation()
	require.Error(t, err)
}

func TestCheckConfigValidation_NoTargets(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: "subscriptionID",
		ClientID:       "clientID",
		ClientSecret:   "clientSecret",
		TenantID:       "tenantID",
		Log:            testutil.Logger{},
		azureClient:    newAzureClient(),
	}

	err := am.checkConfigValidation()
	require.Error(t, err)
}

func TestCreateResourceGroupTargetsFromSubscriptionTargets_Success(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: "subscriptionID",
		ClientID:       "clientID",
		ClientSecret:   "clientSecret",
		TenantID:       "tenantID",
		SubscriptionTargets: []*Resource{
			{
				ResourceType: "Microsoft/type1",
				Metrics:      []string{"metric1", "metric2"},
				Aggregations: []string{"Total"},
			},
			{
				ResourceType: "Microsoft/type2",
				Metrics:      []string{"metric3"},
				Aggregations: []string{"Total", "Average"},
			},
		},
		Log:         testutil.Logger{},
		azureClient: newAzureClient(),
	}

	subscriptionResourceGroupsBody, err := getFileBody("testdata/subscription_resource_groups_body.json")
	require.NoError(t, err)
	require.NotNil(t, subscriptionResourceGroupsBody)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		am.buildSubscriptionResourceGroupsAPIURL(),
		httpmock.NewBytesResponder(200, subscriptionResourceGroupsBody))

	err = am.createResourceGroupTargetsFromSubscriptionTargets()
	require.NoError(t, err)

	assert.Equal(t, 2, len(am.ResourceGroupTargets))

	assert.Equal(t, "resourceGroup1", am.ResourceGroupTargets[0].ResourceGroup)
	assert.Equal(t, 2, len(am.ResourceGroupTargets[0].Resources))
	assert.Equal(t, am.SubscriptionTargets[0].ResourceType, am.ResourceGroupTargets[0].Resources[0].ResourceType)
	assert.Equal(t, am.SubscriptionTargets[0].Metrics, am.ResourceGroupTargets[0].Resources[0].Metrics)
	assert.Equal(t, am.SubscriptionTargets[0].Aggregations, am.ResourceGroupTargets[0].Resources[0].Aggregations)
	assert.Equal(t, am.SubscriptionTargets[1].ResourceType, am.ResourceGroupTargets[0].Resources[1].ResourceType)
	assert.Equal(t, am.SubscriptionTargets[1].Metrics, am.ResourceGroupTargets[0].Resources[1].Metrics)
	assert.Equal(t, am.SubscriptionTargets[1].Aggregations, am.ResourceGroupTargets[0].Resources[1].Aggregations)

	assert.Equal(t, "resourceGroup2", am.ResourceGroupTargets[1].ResourceGroup)
	assert.Equal(t, 2, len(am.ResourceGroupTargets[1].Resources))
	assert.Equal(t, am.SubscriptionTargets[0].ResourceType, am.ResourceGroupTargets[1].Resources[0].ResourceType)
	assert.Equal(t, am.SubscriptionTargets[0].Metrics, am.ResourceGroupTargets[1].Resources[0].Metrics)
	assert.Equal(t, am.SubscriptionTargets[0].Aggregations, am.ResourceGroupTargets[1].Resources[0].Aggregations)
	assert.Equal(t, am.SubscriptionTargets[1].ResourceType, am.ResourceGroupTargets[1].Resources[1].ResourceType)
	assert.Equal(t, am.SubscriptionTargets[1].Metrics, am.ResourceGroupTargets[1].Resources[1].Metrics)
	assert.Equal(t, am.SubscriptionTargets[1].Aggregations, am.ResourceGroupTargets[1].Resources[1].Aggregations)
}

func TestCreateResourceTargetsFromResourceGroupTargets_Success(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: "subscriptionID",
		ClientID:       "clientID",
		ClientSecret:   "clientSecret",
		TenantID:       "tenantID",
		ResourceGroupTargets: []*ResourceGroupTarget{
			newResourceGroupTarget("resourceGroup1",
				[]*Resource{
					{
						ResourceType: "Microsoft/type1",
						Metrics:      []string{"metric1", "metric2"},
						Aggregations: []string{"Total"},
					},
					{
						ResourceType: "Microsoft/type2",
						Metrics:      []string{"metric1", "metric2"},
						Aggregations: []string{"Average"},
					},
				},
			),
			newResourceGroupTarget("resourceGroup2",
				[]*Resource{
					{
						ResourceType: "Microsoft/type1",
						Metrics:      []string{"metric1"},
						Aggregations: []string{"Total", "Average"},
					},
					{
						ResourceType: "Microsoft/type2",
						Metrics:      []string{"metric1", "metric2"},
						Aggregations: []string{"Total"},
					},
				},
			),
			newResourceGroupTarget("resourceGroup3",
				[]*Resource{
					{
						ResourceType: "Microsoft/type2",
						Metrics:      []string{},
						Aggregations: []string{},
					},
				},
			),
		},
		Log:         testutil.Logger{},
		azureClient: newAzureClient(),
	}

	resourceGroup1ResourcesBody, err := getFileBody("testdata/resource_group_1_resources_body.json")
	require.NoError(t, err)
	require.NotNil(t, resourceGroup1ResourcesBody)

	resourceGroup2ResourcesBody, err := getFileBody("testdata/resource_group_2_resources_body.json")
	require.NoError(t, err)
	require.NotNil(t, resourceGroup2ResourcesBody)

	resourceGroup3ResourcesBody, err := getFileBody("testdata/resource_group_3_resources_body.json")
	require.NoError(t, err)
	require.NotNil(t, resourceGroup3ResourcesBody)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		am.buildResourceGroupResourcesAPIURL(am.ResourceGroupTargets[0].ResourceGroup),
		httpmock.NewBytesResponder(200, resourceGroup1ResourcesBody))

	httpmock.RegisterResponder(
		"GET",
		am.buildResourceGroupResourcesAPIURL(am.ResourceGroupTargets[1].ResourceGroup),
		httpmock.NewBytesResponder(200, resourceGroup2ResourcesBody))

	httpmock.RegisterResponder(
		"GET",
		am.buildResourceGroupResourcesAPIURL(am.ResourceGroupTargets[2].ResourceGroup),
		httpmock.NewBytesResponder(200, resourceGroup3ResourcesBody))

	err = am.createResourceTargetsFromResourceGroupTargets()
	require.NoError(t, err)

	assert.Equal(t, 4, len(am.ResourceTargets))

	for _, target := range am.ResourceTargets {
		if strings.HasSuffix(target.ResourceID, "resourceGroup1Resource1") {
			assert.Equal(t, am.ResourceGroupTargets[0].Resources[0].Metrics, target.Metrics)
			assert.Equal(t, am.ResourceGroupTargets[0].Resources[0].Aggregations, target.Aggregations)
		} else if strings.HasSuffix(target.ResourceID, "resourceGroup1Resource2") {
			assert.Equal(t, am.ResourceGroupTargets[0].Resources[0].Metrics, target.Metrics)
			assert.Equal(t, am.ResourceGroupTargets[0].Resources[0].Aggregations, target.Aggregations)
		} else if strings.HasSuffix(target.ResourceID, "resourceGroup1Resource3") {
			assert.Equal(t, am.ResourceGroupTargets[0].Resources[1].Metrics, target.Metrics)
			assert.Equal(t, am.ResourceGroupTargets[0].Resources[1].Aggregations, target.Aggregations)
		} else if strings.HasSuffix(target.ResourceID, "resourceGroup2Resource1") {
			assert.Equal(t, am.ResourceGroupTargets[1].Resources[0].Metrics, target.Metrics)
			assert.Equal(t, am.ResourceGroupTargets[1].Resources[0].Aggregations, target.Aggregations)
		} else {
			assert.FailNowf(t, "Did not get any expected resource ID", "Test failed")
		}
	}
}

func TestCheckResourceTargetsMetricsValidation_Success(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: "subscriptionID",
		ClientID:       "clientID",
		ClientSecret:   "clientSecret",
		TenantID:       "tenantID",
		ResourceTargets: []*ResourceTarget{
			newResourceTarget("resource1", []string{}, []string{}),
			newResourceTarget("resource2", []string{"metric1", "metric2"}, []string{"Total"}),
		},
		Log:         testutil.Logger{},
		azureClient: newAzureClient(),
	}

	resourceMetricDefinitionsBody, err := getFileBody("testdata/metric_definitions_body.json")
	require.NoError(t, err)
	require.NotNil(t, resourceMetricDefinitionsBody)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		am.buildMetricDefinitionsAPIURL(am.ResourceTargets[0].ResourceID),
		httpmock.NewBytesResponder(200, resourceMetricDefinitionsBody))

	httpmock.RegisterResponder(
		"GET",
		am.buildMetricDefinitionsAPIURL(am.ResourceTargets[1].ResourceID),
		httpmock.NewBytesResponder(200, resourceMetricDefinitionsBody))

	err = am.checkResourceTargetsMetricsValidation()
	require.NoError(t, err)
}

func TestCheckResourceTargetsMetricsValidation_WithResourceTargetWithInvalidMetric(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: "subscriptionID",
		ClientID:       "clientID",
		ClientSecret:   "clientSecret",
		TenantID:       "tenantID",
		ResourceTargets: []*ResourceTarget{
			newResourceTarget("resource1", []string{}, []string{}),
			newResourceTarget("resource2", []string{"metric1", "metric2"}, []string{"Total"}),
			newResourceTarget("resource3", []string{"invalidMetric"}, []string{"Total"}),
		},
		Log:         testutil.Logger{},
		azureClient: newAzureClient(),
	}

	resourceMetricDefinitionsBody, err := getFileBody("testdata/metric_definitions_body.json")
	require.NoError(t, err)
	require.NotNil(t, resourceMetricDefinitionsBody)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		am.buildMetricDefinitionsAPIURL(am.ResourceTargets[0].ResourceID),
		httpmock.NewBytesResponder(200, resourceMetricDefinitionsBody))

	httpmock.RegisterResponder(
		"GET",
		am.buildMetricDefinitionsAPIURL(am.ResourceTargets[1].ResourceID),
		httpmock.NewBytesResponder(200, resourceMetricDefinitionsBody))

	httpmock.RegisterResponder(
		"GET",
		am.buildMetricDefinitionsAPIURL(am.ResourceTargets[2].ResourceID),
		httpmock.NewBytesResponder(200, resourceMetricDefinitionsBody))

	err = am.checkResourceTargetsMetricsValidation()
	require.Error(t, err)
}

func TestSetResourceTargetsMetrics_Success(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: "subscriptionID",
		ClientID:       "clientID",
		ClientSecret:   "clientSecret",
		TenantID:       "tenantID",
		ResourceTargets: []*ResourceTarget{
			newResourceTarget("resource1", []string{}, []string{}),
			newResourceTarget("resource2", []string{"metric1", "metric2"}, []string{"Total"}),
		},
		Log:         testutil.Logger{},
		azureClient: newAzureClient(),
	}

	resourceMetricDefinitionsBody, err := getFileBody("testdata/metric_definitions_body.json")
	require.NoError(t, err)
	require.NotNil(t, resourceMetricDefinitionsBody)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		am.buildMetricDefinitionsAPIURL(am.ResourceTargets[0].ResourceID),
		httpmock.NewBytesResponder(200, resourceMetricDefinitionsBody))

	httpmock.RegisterResponder(
		"GET",
		am.buildMetricDefinitionsAPIURL(am.ResourceTargets[1].ResourceID),
		httpmock.NewBytesResponder(200, resourceMetricDefinitionsBody))

	err = am.setResourceTargetsMetrics()
	require.NoError(t, err)

	assert.Equal(t, 2, len(am.ResourceTargets))

	assert.Equal(t, 3, len(am.ResourceTargets[0].Metrics))
	assert.Equal(t, []string{"metric1", "metric2", "metric3"}, am.ResourceTargets[0].Metrics)

	assert.Equal(t, 2, len(am.ResourceTargets[1].Metrics))
	assert.Equal(t, []string{"metric1", "metric2"}, am.ResourceTargets[1].Metrics)
}

func TestCheckResourceTargetsMetricsMinTimeGrain_Success(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: "subscriptionID",
		ClientID:       "clientID",
		ClientSecret:   "clientSecret",
		TenantID:       "tenantID",
		ResourceTargets: []*ResourceTarget{
			newResourceTarget("resource1", []string{"metric1", "metric2", "metric3"}, []string{"Total", "Average"}),
			newResourceTarget("resource2", []string{"metric1", "metric2"}, []string{"Total"}),
			newResourceTarget("resource3", []string{"metric1", "metric2", "metric3"}, []string{}),
			newResourceTarget("resource4", []string{"metric1", "metric2"}, []string{}),
		},
		Log:         testutil.Logger{},
		azureClient: newAzureClient(),
	}

	resourceMetricDefinitionsBody, err := getFileBody("testdata/metric_definitions_body.json")
	require.NoError(t, err)
	require.NotNil(t, resourceMetricDefinitionsBody)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		am.buildMetricDefinitionsAPIURL(am.ResourceTargets[0].ResourceID),
		httpmock.NewBytesResponder(200, resourceMetricDefinitionsBody))

	httpmock.RegisterResponder(
		"GET",
		am.buildMetricDefinitionsAPIURL(am.ResourceTargets[1].ResourceID),
		httpmock.NewBytesResponder(200, resourceMetricDefinitionsBody))

	httpmock.RegisterResponder(
		"GET",
		am.buildMetricDefinitionsAPIURL(am.ResourceTargets[2].ResourceID),
		httpmock.NewBytesResponder(200, resourceMetricDefinitionsBody))

	httpmock.RegisterResponder(
		"GET",
		am.buildMetricDefinitionsAPIURL(am.ResourceTargets[3].ResourceID),
		httpmock.NewBytesResponder(200, resourceMetricDefinitionsBody))

	err = am.checkResourceTargetsMetricsMinTimeGrain()
	require.NoError(t, err)

	assert.Equal(t, 6, len(am.ResourceTargets))

	for _, target := range am.ResourceTargets {
		if target.ResourceID == "resource1" {
			if len(target.Metrics) == 2 {
				assert.Equal(t, []string{"metric1", "metric2"}, target.Metrics)
				assert.Equal(t, []string{"Total", "Average"}, target.Aggregations)
			} else if len(target.Metrics) == 1 {
				assert.Equal(t, []string{"metric3"}, target.Metrics)
				assert.Equal(t, []string{"Total", "Average"}, target.Aggregations)
			} else {
				assert.FailNowf(t, "resource1 has no any expected metrics size", "Test failed")
			}
		} else if target.ResourceID == "resource2" {
			assert.Equal(t, []string{"metric1", "metric2"}, target.Metrics)
			assert.Equal(t, []string{"Total"}, target.Aggregations)
		} else if target.ResourceID == "resource3" {
			if len(target.Metrics) == 2 {
				assert.Equal(t, []string{"metric1", "metric2"}, target.Metrics)
				assert.Equal(t, []string{}, target.Aggregations)
			} else if len(target.Metrics) == 1 {
				assert.Equal(t, []string{"metric3"}, target.Metrics)
				assert.Equal(t, []string{}, target.Aggregations)
			} else {
				assert.FailNowf(t, "resource3 has no any expected metrics size", "Test failed")
			}
		} else if target.ResourceID == "resource4" {
			assert.Equal(t, []string{"metric1", "metric2"}, target.Metrics)
			assert.Equal(t, []string{}, target.Aggregations)
		} else {
			assert.FailNowf(t, "Did not get any expected resource ID", "Test failed")
		}
	}
}

func TestCheckResourceTargetsMaxMetrics_Success(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: "subscriptionID",
		ClientID:       "clientID",
		ClientSecret:   "clientSecret",
		TenantID:       "tenantID",
		ResourceTargets: []*ResourceTarget{
			newResourceTarget("resource1", []string{}, []string{"Total", "Average"}),
			newResourceTarget("resource2", []string{"metric1", "metric2"}, []string{"Total"}),
			newResourceTarget("resource3", []string{}, []string{}),
			newResourceTarget("resource4", []string{"metric1", "metric2"}, []string{}),
		},
		Log:         testutil.Logger{},
		azureClient: newAzureClient(),
	}

	for index := 1; index <= 25; index++ {
		am.ResourceTargets[0].Metrics = append(am.ResourceTargets[0].Metrics, "metric"+strconv.Itoa(index))
		am.ResourceTargets[2].Metrics = append(am.ResourceTargets[2].Metrics, "metric"+strconv.Itoa(index))
	}

	expectedResource1Metrics := make([]string, 0)
	expectedResource3Metrics := make([]string, 0)
	for index := 1; index <= maxMetricsPerRequest; index++ {
		expectedResource1Metrics = append(expectedResource1Metrics, "metric"+strconv.Itoa(index))
		expectedResource3Metrics = append(expectedResource3Metrics, "metric"+strconv.Itoa(index))
	}

	am.checkResourceTargetsMaxMetrics()

	assert.Equal(t, 6, len(am.ResourceTargets))

	for _, target := range am.ResourceTargets {
		if target.ResourceID == "resource1" {
			if len(target.Metrics) == maxMetricsPerRequest {
				assert.Equal(t, expectedResource1Metrics, target.Metrics)
				assert.Equal(t, []string{"Total", "Average"}, target.Aggregations)
			} else if len(target.Metrics) == 5 {
				assert.Equal(t, []string{"metric21", "metric22", "metric23", "metric24", "metric25"}, target.Metrics)
				assert.Equal(t, []string{"Total", "Average"}, target.Aggregations)
			} else {
				assert.FailNowf(t, "resource1 has no any expected metrics size", "Test failed")
			}
		} else if target.ResourceID == "resource2" {
			assert.Equal(t, []string{"metric1", "metric2"}, target.Metrics)
			assert.Equal(t, []string{"Total"}, target.Aggregations)
		} else if target.ResourceID == "resource3" {
			if len(target.Metrics) == maxMetricsPerRequest {
				assert.Equal(t, expectedResource3Metrics, target.Metrics)
				assert.Equal(t, []string{}, target.Aggregations)
			} else if len(target.Metrics) == 5 {
				assert.Equal(t, []string{"metric21", "metric22", "metric23", "metric24", "metric25"}, target.Metrics)
				assert.Equal(t, []string{}, target.Aggregations)
			} else {
				assert.FailNowf(t, "resource1 has no any expected metrics size", "Test failed")
			}
		} else if target.ResourceID == "resource4" {
			assert.Equal(t, []string{"metric1", "metric2"}, target.Metrics)
			assert.Equal(t, []string{}, target.Aggregations)
		} else {
			assert.FailNowf(t, "Did not get any expected resource ID", "Test failed")
		}
	}
}

func TestSetResourceTargetsAggregations_Success(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: "subscriptionID",
		ClientID:       "clientID",
		ClientSecret:   "clientSecret",
		TenantID:       "tenantID",
		ResourceTargets: []*ResourceTarget{
			newResourceTarget("resource1", []string{"metric1", "metric2", "metric3"}, []string{}),
			newResourceTarget("resource2", []string{"metric1", "metric2"}, []string{"Total"}),
		},
		Log:         testutil.Logger{},
		azureClient: newAzureClient(),
	}

	am.setResourceTargetsAggregations()

	assert.Equal(t, 5, len(am.ResourceTargets[0].Aggregations))
	assert.Equal(t, []string{"Total", "Count", "Average", "Minimum", "Maximum"}, am.ResourceTargets[0].Aggregations)

	assert.Equal(t, 1, len(am.ResourceTargets[1].Aggregations))
	assert.Equal(t, []string{"Total"}, am.ResourceTargets[1].Aggregations)
}
