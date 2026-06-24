package azure_monitor

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/monitor/armmonitor"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

type client interface {
	ResourcesList(
		context.Context,
		*armresources.ClientListOptions,
	) ([]*armresources.ClientListResponse, error)
	ResourcesListByResourceGroup(
		context.Context,
		string,
		*armresources.ClientListByResourceGroupOptions,
	) ([]*armresources.ClientListByResourceGroupResponse, error)
	MetricDefinitionsList(
		context.Context,
		string,
		*armmonitor.MetricDefinitionsClientListOptions,
	) (armmonitor.MetricDefinitionsClientListResponse, error)
	MetricsList(
		context.Context,
		string,
		*armmonitor.MetricsClientListOptions,
	) (armmonitor.MetricsClientListResponse, error)
}

type clientFactory interface {
	createClient(subscriptionID, clientID, clientSecret, tenantID string, clientOptions azcore.ClientOptions) (client, error)
}

type azureFactory struct{}

func (*azureFactory) createClient(subscriptionID, clientID, clientSecret, tenantID string, clientOptions azcore.ClientOptions) (client, error) {
	var credentials azcore.TokenCredential

	if clientSecret != "" {
		secret, err := azidentity.NewClientSecretCredential(
			tenantID,
			clientID,
			clientSecret,
			&azidentity.ClientSecretCredentialOptions{ClientOptions: clientOptions},
		)
		if err != nil {
			return nil, fmt.Errorf("error creating Azure client credential: %w", err)
		}
		credentials = secret
	} else {
		token, err := azidentity.NewDefaultAzureCredential(
			&azidentity.DefaultAzureCredentialOptions{
				TenantID:      tenantID,
				ClientOptions: clientOptions,
			})
		if err != nil {
			return nil, fmt.Errorf("error creating Azure token: %w", err)
		}
		credentials = token
	}

	options := &arm.ClientOptions{ClientOptions: clientOptions}
	metricClient, err := armmonitor.NewMetricsClient(subscriptionID, credentials, options)
	if err != nil {
		return nil, fmt.Errorf("error creating Azure metric client: %w", err)
	}
	defClient, err := armmonitor.NewMetricDefinitionsClient(subscriptionID, credentials, options)
	if err != nil {
		return nil, fmt.Errorf("error creating Azure definitions client: %w", err)
	}

	resClient, err := armresources.NewClient(subscriptionID, credentials, options)
	if err != nil {
		return nil, fmt.Errorf("error creating Azure resources client: %w", err)
	}

	return &azureClient{
		resourceClient:   resClient,
		definitionClient: defClient,
		metricsClient:    metricClient,
	}, nil
}

type azureClient struct {
	resourceClient   *armresources.Client
	definitionClient *armmonitor.MetricDefinitionsClient
	metricsClient    *armmonitor.MetricsClient
}

// ResourcesList lists the resources
func (c *azureClient) ResourcesList(
	ctx context.Context,
	options *armresources.ClientListOptions,
) ([]*armresources.ClientListResponse, error) {
	responses := make([]*armresources.ClientListResponse, 0)
	pager := c.resourceClient.NewListPager(options)

	for pager.More() {
		response, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		responses = append(responses, &response)
	}

	return responses, nil
}

// ResourcesListByResourceGroup lists the resources by group
func (c *azureClient) ResourcesListByResourceGroup(
	ctx context.Context,
	resourceGroup string,
	options *armresources.ClientListByResourceGroupOptions,
) ([]*armresources.ClientListByResourceGroupResponse, error) {
	responses := make([]*armresources.ClientListByResourceGroupResponse, 0)
	pager := c.resourceClient.NewListByResourceGroupPager(resourceGroup, options)

	for pager.More() {
		response, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		responses = append(responses, &response)
	}

	return responses, nil
}

// MetricDefinitionsList lists the metric definitions for a given resource
func (c *azureClient) MetricDefinitionsList(
	ctx context.Context,
	resourceID string,
	options *armmonitor.MetricDefinitionsClientListOptions,
) (armmonitor.MetricDefinitionsClientListResponse, error) {
	var response armmonitor.MetricDefinitionsClientListResponse

	pager := c.definitionClient.NewListPager(resourceID, options)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return armmonitor.MetricDefinitionsClientListResponse{}, err
		}
		response.MetricDefinitionCollection.Value = append(response.MetricDefinitionCollection.Value, page.Value...)
	}
	return response, nil
}

// MetricsList collects the metric for a given resource
func (c *azureClient) MetricsList(
	ctx context.Context,
	resource string,
	option *armmonitor.MetricsClientListOptions,
) (armmonitor.MetricsClientListResponse, error) {
	return c.metricsClient.List(ctx, resource, option)
}
