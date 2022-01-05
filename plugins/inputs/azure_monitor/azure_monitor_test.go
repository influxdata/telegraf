package azure_monitor

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/influxdata/telegraf/testutil"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getFileBody(filePath string) ([]byte, error) {
	jsonFile, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening file %s: %v", filePath, err)
	}

	defer closeFile(jsonFile, &err)

	body, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return nil, fmt.Errorf("error reading file %s: %v", filePath, err)
	}

	return body, nil
}

func closeFile(file *os.File, err *error) {
	if closeError := file.Close(); closeError != nil {
		*err = fmt.Errorf("error closing file: %v", err)
	}
}

func TestGetAccessToken_Success(t *testing.T) {
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

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"POST",
		fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/token", am.TenantID),
		httpmock.NewStringResponder(200, `{"access_token": "accessToken", "expires_on": "1636548796"}`))

	err := am.getAccessToken()
	require.NoError(t, err)

	assert.Equal(t, "accessToken", am.azureClient.accessToken)

	expiresOn, err := strconv.ParseInt("1636548796", 10, 64)
	require.NoError(t, err)
	require.NotNil(t, expiresOn)

	assert.Equal(t, time.Unix(expiresOn, 0).UTC(), am.azureClient.accessTokenExpiresOn)
}

func TestRefreshAccessToken_AccessTokenRefreshed(t *testing.T) {
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

	expiresOn, err := strconv.ParseInt("1636548796", 10, 64)
	require.NoError(t, err)
	require.NotNil(t, expiresOn)

	am.azureClient.accessToken = "accessToken"
	am.azureClient.accessTokenExpiresOn = time.Unix(expiresOn, 0).UTC()

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"POST",
		fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/token", am.TenantID),
		httpmock.NewStringResponder(200, `{"access_token": "newAccessToken", "expires_on": "1736548796"}`))

	err = am.refreshAccessToken()
	require.NoError(t, err)

	assert.Equal(t, "newAccessToken", am.azureClient.accessToken)

	expiresOn, err = strconv.ParseInt("1736548796", 10, 64)
	require.NoError(t, err)
	require.NotNil(t, expiresOn)

	assert.Equal(t, time.Unix(expiresOn, 0).UTC(), am.azureClient.accessTokenExpiresOn)
}

func TestRefreshAccessToken_AccessTokenNotRefreshed(t *testing.T) {
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

	expiresOn, err := strconv.ParseInt("1736548796", 10, 64)
	require.NoError(t, err)
	require.NotNil(t, expiresOn)

	am.azureClient.accessToken = "accessToken"
	am.azureClient.accessTokenExpiresOn = time.Unix(expiresOn, 0).UTC()

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"POST",
		fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/token", am.TenantID),
		httpmock.NewStringResponder(200, `{"access_token": "newAccessToken", "expires_on": "1836548796"}`))

	err = am.refreshAccessToken()
	require.NoError(t, err)

	assert.Equal(t, "accessToken", am.azureClient.accessToken)
	assert.Equal(t, time.Unix(expiresOn, 0).UTC(), am.azureClient.accessTokenExpiresOn)
}

func TestInit_ResourceTargetsOnly(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: "subscriptionID",
		ClientID:       "clientID",
		ClientSecret:   "clientSecret",
		TenantID:       "tenantID",
		ResourceTargets: []*ResourceTarget{
			newResourceTarget("resource1", []string{"metric1", "metric2", "metric3"}, []string{"Total", "Average"}),
			newResourceTarget("resource2", []string{"metric1", "metric2"}, []string{"Total"}),
			newResourceTarget("resource3", []string{}, []string{"Total"}),
			newResourceTarget("resource4", []string{"metric1", "metric2"}, []string{}),
			newResourceTarget("resource5", []string{}, []string{}),
			newResourceTarget("resource6", []string{}, []string{"Average"}),
			newResourceTarget("resource7", []string{}, []string{"Average"}),
		},
		Log:         testutil.Logger{},
		azureClient: newAzureClient(),
	}

	for index := 1; index <= 25; index++ {
		if index <= 10 {
			am.ResourceTargets[5].Metrics = append(am.ResourceTargets[5].Metrics, "metric1")
			am.ResourceTargets[6].Metrics = append(am.ResourceTargets[6].Metrics, "metric1")
		} else if index <= 22 {
			am.ResourceTargets[5].Metrics = append(am.ResourceTargets[5].Metrics, "metric2")
			am.ResourceTargets[6].Metrics = append(am.ResourceTargets[6].Metrics, "metric2")
		} else {
			am.ResourceTargets[5].Metrics = append(am.ResourceTargets[5].Metrics, "metric2")
			am.ResourceTargets[6].Metrics = append(am.ResourceTargets[6].Metrics, "metric3")
		}
	}

	expectedResource6Metrics := make([]string, 0)
	expectedResource7Metrics := make([]string, 0)
	for index := 1; index <= maxMetricsPerRequest; index++ {
		if index <= 10 {
			expectedResource6Metrics = append(expectedResource6Metrics, "metric1")
			expectedResource7Metrics = append(expectedResource7Metrics, "metric1")
		} else {
			expectedResource6Metrics = append(expectedResource6Metrics, "metric2")
			expectedResource7Metrics = append(expectedResource7Metrics, "metric2")
		}
	}

	resourceMetricDefinitionsBody, err := getFileBody("testdata/metric_definitions_body.json")
	require.NoError(t, err)
	require.NotNil(t, resourceMetricDefinitionsBody)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"POST",
		fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/token", am.TenantID),
		httpmock.NewStringResponder(200, `{"access_token": "accessToken", "expires_on": "1636548796"}`))

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

	httpmock.RegisterResponder(
		"GET",
		am.buildMetricDefinitionsAPIURL(am.ResourceTargets[4].ResourceID),
		httpmock.NewBytesResponder(200, resourceMetricDefinitionsBody))

	httpmock.RegisterResponder(
		"GET",
		am.buildMetricDefinitionsAPIURL(am.ResourceTargets[5].ResourceID),
		httpmock.NewBytesResponder(200, resourceMetricDefinitionsBody))

	httpmock.RegisterResponder(
		"GET",
		am.buildMetricDefinitionsAPIURL(am.ResourceTargets[6].ResourceID),
		httpmock.NewBytesResponder(200, resourceMetricDefinitionsBody))

	err = am.Init()
	require.NoError(t, err)

	assert.Equal(t, 13, len(am.ResourceTargets))

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
				assert.Equal(t, []string{"Total"}, target.Aggregations)
			} else if len(target.Metrics) == 1 {
				assert.Equal(t, []string{"metric3"}, target.Metrics)
				assert.Equal(t, []string{"Total"}, target.Aggregations)
			} else {
				assert.FailNowf(t, "resource3 has no any expected metrics size", "Test failed")
			}
		} else if target.ResourceID == "resource4" {
			if len(target.Metrics) == 2 {
				assert.Equal(t, []string{"metric1", "metric2"}, target.Metrics)
				assert.Equal(t, []string{"Total", "Count", "Average", "Minimum", "Maximum"}, target.Aggregations)
			} else if len(target.Metrics) == 1 {
				assert.Equal(t, []string{"metric3"}, target.Metrics)
				assert.Equal(t, []string{"Total", "Count", "Average", "Minimum", "Maximum"}, target.Aggregations)
			} else {
				assert.FailNowf(t, "resource4 has no any expected metrics size", "Test failed")
			}
		} else if target.ResourceID == "resource5" {
			if len(target.Metrics) == 2 {
				assert.Equal(t, []string{"metric1", "metric2"}, target.Metrics)
				assert.Equal(t, []string{"Total", "Count", "Average", "Minimum", "Maximum"}, target.Aggregations)
			} else if len(target.Metrics) == 1 {
				assert.Equal(t, []string{"metric3"}, target.Metrics)
				assert.Equal(t, []string{"Total", "Count", "Average", "Minimum", "Maximum"}, target.Aggregations)
			} else {
				assert.FailNowf(t, "resource5 has no any expected metrics size", "Test failed")
			}
		} else if target.ResourceID == "resource6" {
			if len(target.Metrics) == maxMetricsPerRequest {
				assert.Equal(t, expectedResource6Metrics, target.Metrics)
				assert.Equal(t, []string{"Average"}, target.Aggregations)
			} else if len(target.Metrics) == 5 {
				assert.Equal(t, []string{"metric2", "metric2", "metric2", "metric2", "metric2"}, target.Metrics)
				assert.Equal(t, []string{"Average"}, target.Aggregations)
			} else {
				assert.FailNowf(t, "resource6 has no any expected metrics size", "Test failed")
			}
		} else if target.ResourceID == "resource7" {
			if len(target.Metrics) == maxMetricsPerRequest {
				assert.Equal(t, expectedResource7Metrics, target.Metrics)
				assert.Equal(t, []string{"Average"}, target.Aggregations)
			} else if len(target.Metrics) == 2 {
				assert.Equal(t, []string{"metric2", "metric2"}, target.Metrics)
				assert.Equal(t, []string{"Average"}, target.Aggregations)
			} else if len(target.Metrics) == 3 {
				assert.Equal(t, []string{"metric3", "metric3", "metric3"}, target.Metrics)
				assert.Equal(t, []string{"Average"}, target.Aggregations)
			} else {
				assert.FailNowf(t, "resource7 has no any expected metrics size", "Test failed")
			}
		} else {
			assert.FailNowf(t, "Did not get any expected resource ID", "Test failed")
		}
	}
}

func TestInit_ResourceGroupTargetsOnly(t *testing.T) {
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
						Metrics:      []string{"metric1", "metric2", "metric3"},
						Aggregations: []string{"Total"},
					},
					{
						ResourceType: "Microsoft/type2",
						Metrics:      []string{"metric1", "metric2"},
						Aggregations: []string{"Average"},
					},
					{
						ResourceType: "Microsoft/type2",
						Metrics:      []string{},
						Aggregations: []string{},
					},
					{
						ResourceType: "Microsoft/type2",
						Metrics:      []string{},
						Aggregations: []string{"Average"},
					},
					{
						ResourceType: "Microsoft/type2",
						Metrics:      []string{},
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

	for index := 1; index <= 25; index++ {
		if index <= 10 {
			am.ResourceGroupTargets[0].Resources[3].Metrics = append(am.ResourceGroupTargets[0].Resources[3].Metrics, "metric1")
			am.ResourceGroupTargets[0].Resources[4].Metrics = append(am.ResourceGroupTargets[0].Resources[4].Metrics, "metric1")
		} else if index <= 22 {
			am.ResourceGroupTargets[0].Resources[3].Metrics = append(am.ResourceGroupTargets[0].Resources[3].Metrics, "metric2")
			am.ResourceGroupTargets[0].Resources[4].Metrics = append(am.ResourceGroupTargets[0].Resources[4].Metrics, "metric2")
		} else {
			am.ResourceGroupTargets[0].Resources[3].Metrics = append(am.ResourceGroupTargets[0].Resources[3].Metrics, "metric2")
			am.ResourceGroupTargets[0].Resources[4].Metrics = append(am.ResourceGroupTargets[0].Resources[4].Metrics, "metric3")
		}
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

	resourceMetricDefinitionsBody, err := getFileBody("testdata/metric_definitions_body.json")
	require.NoError(t, err)
	require.NotNil(t, resourceMetricDefinitionsBody)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"POST",
		fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/token", am.TenantID),
		httpmock.NewStringResponder(200, `{"access_token": "accessToken", "expires_on": "1636548796"}`))

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

	httpmock.RegisterResponder(
		"GET",
		am.buildMetricDefinitionsAPIURL("resourceGroups/resourceGroup1/providers/Microsoft/type1/resourceGroup1Resource1"),
		httpmock.NewBytesResponder(200, resourceMetricDefinitionsBody))

	httpmock.RegisterResponder(
		"GET",
		am.buildMetricDefinitionsAPIURL("resourceGroups/resourceGroup1/providers/Microsoft/type1/resourceGroup1Resource2"),
		httpmock.NewBytesResponder(200, resourceMetricDefinitionsBody))

	httpmock.RegisterResponder(
		"GET",
		am.buildMetricDefinitionsAPIURL("resourceGroups/resourceGroup1/providers/Microsoft/type2/resourceGroup1Resource3"),
		httpmock.NewBytesResponder(200, resourceMetricDefinitionsBody))

	httpmock.RegisterResponder(
		"GET",
		am.buildMetricDefinitionsAPIURL("resourceGroups/resourceGroup2/providers/Microsoft/type1/resourceGroup2Resource1"),
		httpmock.NewBytesResponder(200, resourceMetricDefinitionsBody))

	httpmock.RegisterResponder(
		"GET",
		am.buildMetricDefinitionsAPIURL("resourceGroups/resourceGroup2/providers/Microsoft/type3/resourceGroup2Resource2"),
		httpmock.NewBytesResponder(200, resourceMetricDefinitionsBody))

	httpmock.RegisterResponder(
		"GET",
		am.buildMetricDefinitionsAPIURL("resourceGroups/resourceGroup2/providers/Microsoft/type3/resourceGroup2Resource3"),
		httpmock.NewBytesResponder(200, resourceMetricDefinitionsBody))

	err = am.Init()
	require.NoError(t, err)

	assert.Equal(t, 13, len(am.ResourceTargets))
}

func TestInit_SubscriptionTargetsOnly(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: "subscriptionID",
		ClientID:       "clientID",
		ClientSecret:   "clientSecret",
		TenantID:       "tenantID",
		SubscriptionTargets: []*Resource{
			{
				ResourceType: "Microsoft/type1",
				Metrics:      []string{"metric1", "metric2", "metric3"},
				Aggregations: []string{"Total"},
			},
			{
				ResourceType: "Microsoft/type2",
				Metrics:      []string{"metric3"},
				Aggregations: []string{"Total", "Average"},
			},
			{
				ResourceType: "Microsoft/type3",
				Metrics:      []string{},
				Aggregations: []string{},
			},
			{
				ResourceType: "Microsoft/type3",
				Metrics:      []string{},
				Aggregations: []string{},
			},
			{
				ResourceType: "Microsoft/type3",
				Metrics:      []string{},
				Aggregations: []string{},
			},
		},
		Log:         testutil.Logger{},
		azureClient: newAzureClient(),
	}

	for index := 1; index <= 25; index++ {
		if index <= 10 {
			am.SubscriptionTargets[3].Metrics = append(am.SubscriptionTargets[3].Metrics, "metric1")
			am.SubscriptionTargets[4].Metrics = append(am.SubscriptionTargets[4].Metrics, "metric1")
		} else if index <= 22 {
			am.SubscriptionTargets[3].Metrics = append(am.SubscriptionTargets[3].Metrics, "metric2")
			am.SubscriptionTargets[4].Metrics = append(am.SubscriptionTargets[4].Metrics, "metric2")
		} else {
			am.SubscriptionTargets[3].Metrics = append(am.SubscriptionTargets[3].Metrics, "metric2")
			am.SubscriptionTargets[4].Metrics = append(am.SubscriptionTargets[4].Metrics, "metric3")
		}
	}

	subscriptionResourceGroupsBody, err := getFileBody("testdata/subscription_resource_groups_body.json")
	require.NoError(t, err)
	require.NotNil(t, subscriptionResourceGroupsBody)

	resourceGroup1Body, err := getFileBody("testdata/resource_group_1_resources_body.json")
	require.NoError(t, err)
	require.NotNil(t, resourceGroup1Body)

	resourceGroup2Body, err := getFileBody("testdata/resource_group_2_resources_body.json")
	require.NoError(t, err)
	require.NotNil(t, resourceGroup2Body)

	resourceMetricDefinitionsBody, err := getFileBody("testdata/metric_definitions_body.json")
	require.NoError(t, err)
	require.NotNil(t, resourceMetricDefinitionsBody)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"POST",
		fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/token", am.TenantID),
		httpmock.NewStringResponder(200, `{"access_token": "accessToken", "expires_on": "1636548796"}`))

	httpmock.RegisterResponder(
		"GET",
		am.buildSubscriptionResourceGroupsAPIURL(),
		httpmock.NewBytesResponder(200, subscriptionResourceGroupsBody))

	httpmock.RegisterResponder(
		"GET",
		am.buildResourceGroupResourcesAPIURL("resourceGroup1"),
		httpmock.NewBytesResponder(200, resourceGroup1Body))

	httpmock.RegisterResponder(
		"GET",
		am.buildResourceGroupResourcesAPIURL("resourceGroup2"),
		httpmock.NewBytesResponder(200, resourceGroup2Body))

	httpmock.RegisterResponder(
		"GET",
		am.buildMetricDefinitionsAPIURL("resourceGroups/resourceGroup1/providers/Microsoft/type1/resourceGroup1Resource1"),
		httpmock.NewBytesResponder(200, resourceMetricDefinitionsBody))

	httpmock.RegisterResponder(
		"GET",
		am.buildMetricDefinitionsAPIURL("resourceGroups/resourceGroup1/providers/Microsoft/type1/resourceGroup1Resource2"),
		httpmock.NewBytesResponder(200, resourceMetricDefinitionsBody))

	httpmock.RegisterResponder(
		"GET",
		am.buildMetricDefinitionsAPIURL("resourceGroups/resourceGroup1/providers/Microsoft/type2/resourceGroup1Resource3"),
		httpmock.NewBytesResponder(200, resourceMetricDefinitionsBody))

	httpmock.RegisterResponder(
		"GET",
		am.buildMetricDefinitionsAPIURL("resourceGroups/resourceGroup2/providers/Microsoft/type1/resourceGroup2Resource1"),
		httpmock.NewBytesResponder(200, resourceMetricDefinitionsBody))

	httpmock.RegisterResponder(
		"GET",
		am.buildMetricDefinitionsAPIURL("resourceGroups/resourceGroup2/providers/Microsoft/type3/resourceGroup2Resource2"),
		httpmock.NewBytesResponder(200, resourceMetricDefinitionsBody))

	httpmock.RegisterResponder(
		"GET",
		am.buildMetricDefinitionsAPIURL("resourceGroups/resourceGroup2/providers/Microsoft/type3/resourceGroup2Resource3"),
		httpmock.NewBytesResponder(200, resourceMetricDefinitionsBody))

	err = am.Init()
	require.NoError(t, err)

	assert.Equal(t, 21, len(am.ResourceTargets))
	assert.Equal(t, 2, len(am.ResourceGroupTargets))

	for _, target := range am.ResourceGroupTargets {
		if target.ResourceGroup == "resourceGroup1" || target.ResourceGroup == "resourceGroup2" {
			assert.Equal(t, 5, len(target.Resources))
			assert.Equal(t, am.SubscriptionTargets, target.Resources)
		}
	}
}

func TestInit_AllTargetTypes(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: "subscriptionID",
		ClientID:       "clientID",
		ClientSecret:   "clientSecret",
		TenantID:       "tenantID",
		ResourceTargets: []*ResourceTarget{
			newResourceTarget("resource", []string{"metric1"}, []string{"Total"}),
			newResourceTarget("resource", []string{}, []string{}),
		},
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
		SubscriptionTargets: []*Resource{
			{
				ResourceType: "Microsoft/type1",
				Metrics:      []string{"metric1", "metric2", "metric3"},
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

	resourceGroup1ResourcesBody, err := getFileBody("testdata/resource_group_1_resources_body.json")
	require.NoError(t, err)
	require.NotNil(t, resourceGroup1ResourcesBody)

	resourceGroup2ResourcesBody, err := getFileBody("testdata/resource_group_2_resources_body.json")
	require.NoError(t, err)
	require.NotNil(t, resourceGroup2ResourcesBody)

	resourceGroup3ResourcesBody, err := getFileBody("testdata/resource_group_3_resources_body.json")
	require.NoError(t, err)
	require.NotNil(t, resourceGroup3ResourcesBody)

	resourceMetricDefinitionsBody, err := getFileBody("testdata/metric_definitions_body.json")
	require.NoError(t, err)
	require.NotNil(t, resourceMetricDefinitionsBody)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"POST",
		fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/token", am.TenantID),
		httpmock.NewStringResponder(200, `{"access_token": "accessToken", "expires_on": "1636548796"}`))

	httpmock.RegisterResponder(
		"GET",
		am.buildSubscriptionResourceGroupsAPIURL(),
		httpmock.NewBytesResponder(200, subscriptionResourceGroupsBody))

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

	httpmock.RegisterResponder(
		"GET",
		am.buildMetricDefinitionsAPIURL("resourceGroups/resourceGroup1/providers/Microsoft/type1/resourceGroup1Resource1"),
		httpmock.NewBytesResponder(200, resourceMetricDefinitionsBody))

	httpmock.RegisterResponder(
		"GET",
		am.buildMetricDefinitionsAPIURL("resourceGroups/resourceGroup1/providers/Microsoft/type1/resourceGroup1Resource2"),
		httpmock.NewBytesResponder(200, resourceMetricDefinitionsBody))

	httpmock.RegisterResponder(
		"GET",
		am.buildMetricDefinitionsAPIURL("resourceGroups/resourceGroup1/providers/Microsoft/type2/resourceGroup1Resource3"),
		httpmock.NewBytesResponder(200, resourceMetricDefinitionsBody))

	httpmock.RegisterResponder(
		"GET",
		am.buildMetricDefinitionsAPIURL("resourceGroups/resourceGroup2/providers/Microsoft/type1/resourceGroup2Resource1"),
		httpmock.NewBytesResponder(200, resourceMetricDefinitionsBody))

	httpmock.RegisterResponder(
		"GET",
		am.buildMetricDefinitionsAPIURL(am.ResourceTargets[0].ResourceID),
		httpmock.NewBytesResponder(200, resourceMetricDefinitionsBody))

	httpmock.RegisterResponder(
		"GET",
		am.buildMetricDefinitionsAPIURL(am.ResourceTargets[1].ResourceID),
		httpmock.NewBytesResponder(200, resourceMetricDefinitionsBody))

	err = am.Init()
	require.NoError(t, err)

	assert.Equal(t, 14, len(am.ResourceTargets))
	assert.Equal(t, 5, len(am.ResourceGroupTargets))
}

func TestInit_ResourceGroupTargetsOnlyNoResourceTargetsCreated(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: "subscriptionID",
		ClientID:       "clientID",
		ClientSecret:   "clientSecret",
		TenantID:       "tenantID",
		ResourceGroupTargets: []*ResourceGroupTarget{
			newResourceGroupTarget("resourceGroup1",
				[]*Resource{
					{
						ResourceType: "Microsoft/type3",
						Metrics:      []string{"metric1", "metric2"},
						Aggregations: []string{"Total"},
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

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"POST",
		fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/token", am.TenantID),
		httpmock.NewStringResponder(200, `{"access_token": "accessToken", "expires_on": "1636548796"}`))

	httpmock.RegisterResponder(
		"GET",
		am.buildResourceGroupResourcesAPIURL(am.ResourceGroupTargets[0].ResourceGroup),
		httpmock.NewBytesResponder(200, resourceGroup1ResourcesBody))

	err = am.Init()
	require.Error(t, err)
}

func TestInit_SubscriptionTargetsOnlyNoResourceGroupTargetsCreated(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: "subscriptionID",
		ClientID:       "clientID",
		ClientSecret:   "clientSecret",
		TenantID:       "tenantID",
		SubscriptionTargets: []*Resource{
			{
				ResourceType: "Microsoft/type1",
				Metrics:      []string{"metric1", "metric2", "metric3"},
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

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"POST",
		fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/token", am.TenantID),
		httpmock.NewStringResponder(200, `{"access_token": "accessToken", "expires_on": "1636548796"}`))

	httpmock.RegisterResponder(
		"GET",
		am.buildSubscriptionResourceGroupsAPIURL(),
		httpmock.NewStringResponder(200, `{"value": []}`))

	err := am.Init()
	require.Error(t, err)
}

func TestInit_NoSubscriptionID(t *testing.T) {
	am := &AzureMonitor{
		ClientID:     "clientID",
		ClientSecret: "clientSecret",
		TenantID:     "tenantID",
		ResourceTargets: []*ResourceTarget{
			newResourceTarget("resource", []string{}, []string{}),
		},
		Log:         testutil.Logger{},
		azureClient: newAzureClient(),
	}

	err := am.Init()
	require.Error(t, err)
}

func TestInit_NoClientID(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: "subscriptionID",
		ClientSecret:   "clientSecret",
		TenantID:       "tenantID",
		ResourceTargets: []*ResourceTarget{
			newResourceTarget("resource", []string{}, []string{}),
		},
		Log:         testutil.Logger{},
		azureClient: newAzureClient(),
	}

	err := am.Init()
	require.Error(t, err)
}

func TestInit_NoClientSecret(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: "subscriptionID",
		ClientID:       "clientID",
		TenantID:       "tenantID",
		ResourceTargets: []*ResourceTarget{
			newResourceTarget("resource", []string{}, []string{}),
		},
		Log:         testutil.Logger{},
		azureClient: newAzureClient(),
	}

	err := am.Init()
	require.Error(t, err)
}

func TestInit_NoTenantID(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: "subscriptionID",
		ClientID:       "clientID",
		ClientSecret:   "clientSecret",
		ResourceTargets: []*ResourceTarget{
			newResourceTarget("resource", []string{}, []string{}),
		},
		Log:         testutil.Logger{},
		azureClient: newAzureClient(),
	}

	err := am.Init()
	require.Error(t, err)
}

func TestInit_NoTargets(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: "subscriptionID",
		ClientID:       "clientID",
		ClientSecret:   "clientSecret",
		TenantID:       "tenantID",
		Log:            testutil.Logger{},
		azureClient:    newAzureClient(),
	}

	err := am.Init()
	require.Error(t, err)
}

func TestInit_ResourceTargetWithoutResourceID(t *testing.T) {
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

	err := am.Init()
	require.Error(t, err)
}

func TestInit_ResourceTargetWithInvalidAggregation(t *testing.T) {
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

	err := am.Init()
	require.Error(t, err)
}

func TestInit_ResourceGroupTargetWithoutResourceGroup(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: "subscriptionID",
		ClientID:       "clientID",
		ClientSecret:   "clientSecret",
		TenantID:       "tenantID",
		ResourceGroupTargets: []*ResourceGroupTarget{
			newResourceGroupTarget("",
				[]*Resource{
					{
						ResourceType: "Microsoft/type",
						Metrics:      []string{},
						Aggregations: []string{},
					},
				},
			),
		},
		Log:         testutil.Logger{},
		azureClient: newAzureClient(),
	}

	err := am.Init()
	require.Error(t, err)
}

func TestInit_ResourceGroupTargetWithResourceWithoutResourceType(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: "subscriptionID",
		ClientID:       "clientID",
		ClientSecret:   "clientSecret",
		TenantID:       "tenantID",
		ResourceGroupTargets: []*ResourceGroupTarget{
			newResourceGroupTarget("",
				[]*Resource{
					{
						ResourceType: "",
						Metrics:      []string{},
						Aggregations: []string{},
					},
				},
			),
		},
		Log:         testutil.Logger{},
		azureClient: newAzureClient(),
	}

	err := am.Init()
	require.Error(t, err)
}

func TestInit_ResourceGroupTargetWithInvalidAggregation(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: "subscriptionID",
		ClientID:       "clientID",
		ClientSecret:   "clientSecret",
		TenantID:       "tenantID",
		ResourceGroupTargets: []*ResourceGroupTarget{
			newResourceGroupTarget("",
				[]*Resource{
					{
						ResourceType: "Microsoft/type",
						Metrics:      []string{},
						Aggregations: []string{"invalidAggregation"},
					},
				},
			),
		},
		Log:         testutil.Logger{},
		azureClient: newAzureClient(),
	}

	err := am.Init()
	require.Error(t, err)
}

func TestInit_ResourceGroupTargetWithoutResources(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: "subscriptionID",
		ClientID:       "clientID",
		ClientSecret:   "clientSecret",
		TenantID:       "tenantID",
		ResourceGroupTargets: []*ResourceGroupTarget{
			newResourceGroupTarget("",
				[]*Resource{},
			),
		},
		Log:         testutil.Logger{},
		azureClient: newAzureClient(),
	}

	err := am.Init()
	require.Error(t, err)
}

func TestInit_SubscriptionTargetWithoutResourceType(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: "subscriptionID",
		ClientID:       "clientID",
		ClientSecret:   "clientSecret",
		TenantID:       "tenantID",
		ResourceGroupTargets: []*ResourceGroupTarget{
			newResourceGroupTarget("",
				[]*Resource{
					{
						ResourceType: "",
						Metrics:      []string{},
						Aggregations: []string{},
					},
				},
			),
		},
		Log:         testutil.Logger{},
		azureClient: newAzureClient(),
	}

	err := am.Init()
	require.Error(t, err)
}

func TestInit_SubscriptionTargetWithInvalidAggregation(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: "subscriptionID",
		ClientID:       "clientID",
		ClientSecret:   "clientSecret",
		TenantID:       "tenantID",
		ResourceGroupTargets: []*ResourceGroupTarget{
			newResourceGroupTarget("",
				[]*Resource{
					{
						ResourceType: "Microsoft/type",
						Metrics:      []string{},
						Aggregations: []string{"invalidAggregation"},
					},
				},
			),
		},
		Log:         testutil.Logger{},
		azureClient: newAzureClient(),
	}

	err := am.Init()
	require.Error(t, err)
}

func TestGather_Success(t *testing.T) {
	am := &AzureMonitor{
		SubscriptionID: "subscriptionID",
		ClientID:       "clientID",
		ClientSecret:   "clientSecret",
		TenantID:       "tenantID",
		ResourceTargets: []*ResourceTarget{
			newResourceTarget("resource1", []string{"metric1", "metric2", "metric3"}, []string{"Total", "Maximum"}),
			newResourceTarget("resource2", []string{"metric1", "metric2", "metric3"}, []string{"Total", "Maximum"}),
		},
		Log:         testutil.Logger{},
		azureClient: newAzureClient(),
	}

	expectedResource1Metric1Name := "azure_monitor_microsoft_type1_metric1"
	expectedResource1Metric1MetricFields := make(map[string]interface{})
	expectedResource1Metric1MetricFields["timeStamp"] = "2021-11-05T10:59:00Z"
	expectedResource1Metric1MetricFields["total"] = 5.0
	expectedResource1Metric1MetricFields["maximum"] = 5.0

	expectedResource1Metric2Name := "azure_monitor_microsoft_type1_metric2"
	expectedResource1Metric2MetricFields := make(map[string]interface{})
	expectedResource1Metric2MetricFields["timeStamp"] = "2021-11-05T10:57:00Z"
	expectedResource1Metric2MetricFields["total"] = 4.0
	expectedResource1Metric2MetricFields["maximum"] = 9.0

	expectedResource1MetricsTags := make(map[string]string)
	expectedResource1MetricsTags["subscription_id"] = "subscriptionID"
	expectedResource1MetricsTags["resource_group"] = "resourceGroup"
	expectedResource1MetricsTags["namespace"] = "Microsoft/type1"
	expectedResource1MetricsTags["resource_name"] = "resource1"
	expectedResource1MetricsTags["resource_region"] = "eastus"
	expectedResource1MetricsTags["unit"] = "Count"

	expectedResource2Metric1Name := "azure_monitor_microsoft_type2_metric1"
	expectedResource2Metric1MetricFields := make(map[string]interface{})
	expectedResource2Metric1MetricFields["timeStamp"] = "2021-11-05T10:59:00Z"
	expectedResource2Metric1MetricFields["total"] = 10.0
	expectedResource2Metric1MetricFields["maximum"] = 10.0

	expectedResource2MetricsTags := make(map[string]string)
	expectedResource2MetricsTags["subscription_id"] = "subscriptionID"
	expectedResource2MetricsTags["resource_group"] = "resourceGroup"
	expectedResource2MetricsTags["namespace"] = "Microsoft/type2"
	expectedResource2MetricsTags["resource_name"] = "resource2"
	expectedResource2MetricsTags["resource_region"] = "eastus"
	expectedResource2MetricsTags["unit"] = "Count"

	resource1MetricValuesBody, err := getFileBody("testdata/resource_1_metric_values_body.json")
	require.NoError(t, err)
	require.NotNil(t, resource1MetricValuesBody)

	resource2MetricValuesBody, err := getFileBody("testdata/resource_2_metric_values_body.json")
	require.NoError(t, err)
	require.NotNil(t, resource2MetricValuesBody)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"POST",
		fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/token", am.TenantID),
		httpmock.NewStringResponder(200, `{"access_token": "accessToken", "expires_on": "1636548796"}`))

	httpmock.RegisterResponder(
		"GET",
		am.buildMetricValuesAPIURL(am.ResourceTargets[0]),
		httpmock.NewBytesResponder(200, resource1MetricValuesBody))

	httpmock.RegisterResponder(
		"GET",
		am.buildMetricValuesAPIURL(am.ResourceTargets[1]),
		httpmock.NewBytesResponder(200, resource2MetricValuesBody))

	acc := testutil.Accumulator{}
	err = acc.GatherError(am.Gather)
	require.NoError(t, err)

	assert.Equal(t, 3, len(acc.Metrics))

	acc.AssertContainsTaggedFields(t, expectedResource1Metric1Name, expectedResource1Metric1MetricFields, expectedResource1MetricsTags)
	acc.AssertContainsTaggedFields(t, expectedResource1Metric2Name, expectedResource1Metric2MetricFields, expectedResource1MetricsTags)
	acc.AssertContainsTaggedFields(t, expectedResource2Metric1Name, expectedResource2Metric1MetricFields, expectedResource2MetricsTags)
}
