package azure_monitor

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/monitor/armmonitor"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type AzureMonitor struct {
	SubscriptionID       string                 `toml:"subscription_id"`
	ClientID             string                 `toml:"client_id"`
	ClientSecret         string                 `toml:"client_secret"`
	TenantID             string                 `toml:"tenant_id"`
	ResourceTargets      []*ResourceTarget      `toml:"resource_target"`
	ResourceGroupTargets []*ResourceGroupTarget `toml:"resource_group_target"`
	SubscriptionTargets  []*Resource            `toml:"subscription_target"`
	Log                  telegraf.Logger        `toml:"-"`

	azureClients *azureClients
}

type ResourceTarget struct {
	ResourceID   string   `toml:"resource_id"`
	Metrics      []string `toml:"metrics"`
	Aggregations []string `toml:"aggregations"`
}

type ResourceGroupTarget struct {
	ResourceGroup string      `toml:"resource_group"`
	Resources     []*Resource `toml:"resource"`
}

type Resource struct {
	ResourceType string   `toml:"resource_type"`
	Metrics      []string `toml:"metrics"`
	Aggregations []string `toml:"aggregations"`
}

type azureClients struct {
	ctx                     context.Context
	resourcesClient         resourcesClient
	metricDefinitionsClient metricDefinitionsClient
	metricsClient           metricsClient
}

type azureResourcesClient struct {
	client *armresources.Client
}

type resourcesClient interface {
	List(context.Context, *armresources.ClientListOptions) ([]*armresources.ClientListResponse, error)
	ListByResourceGroup(context.Context, string, *armresources.ClientListByResourceGroupOptions) ([]*armresources.ClientListByResourceGroupResponse, error)
}

type metricDefinitionsClient interface {
	List(context.Context, string, *armmonitor.MetricDefinitionsClientListOptions) (armmonitor.MetricDefinitionsClientListResponse, error)
}

type metricsClient interface {
	List(context.Context, string, *armmonitor.MetricsClientListOptions) (armmonitor.MetricsClientListResponse, error)
}

var sampleConfig = `
  # can be found under Overview->Essentials in the Azure portal for your application/service
  subscription_id = "<<SUBSCRIPTION_ID>>"
  # can be obtained by registering an application under Azure Active Directory
  client_id = "<<CLIENT_ID>>"
  # can be obtained by registering an application under Azure Active Directory
  client_secret = "<<CLIENT_SECRET>>"
  # can be found under Azure Active Directory->Properties
  tenant_id = "<<TENANT_ID>>"

  # resource target #1 to collect metrics from
  [[inputs.azure_monitor.resource_target]]
    # can be found undet Overview->Essentials->JSON View in the Azure portal for your application/service
    # must start with 'resourceGroups/...' ('/subscriptions/xxxxxxxx-xxxx-xxxx-xxx-xxxxxxxxxxxx'
    # must be removed from the beginning of Resource ID property value)
    resource_id = "<<RESOURCE_ID>>"
    # the metric names to collect
    # leave the array empty to use all metrics available to this resource
    metrics = [ "<<METRIC>>", "<<METRIC>>" ]
    # metrics aggregation type value to collect
    # can be 'Total', 'Count', 'Average', 'Minimum', 'Maximum'
    # leave the array empty to collect all aggregation types values for each metric
    aggregations = [ "<<AGGREGATION>>", "<<AGGREGATION>>" ]
    
  # resource target #2 to collect metrics from
  [[inputs.azure_monitor.resource_target]]
    resource_id = "<<RESOURCE_ID>>"
    metrics = [ "<<METRIC>>", "<<METRIC>>" ]
    aggregations = [ "<<AGGREGATION>>", "<<AGGREGATION>>" ]

  # resource group target #1 to collect metrics from resources under it with resource type
  [[inputs.azure_monitor.resource_group_target]]
    # the resource group name
    resource_group = "<<RESOURCE_GROUP_NAME>>"

    # defines the resources to collect metrics from
    [[inputs.azure_monitor.resource_group_target.resource]]
      # the resource type
      resource_type = "<<RESOURCE_TYPE>>"
      metrics = [ "<<METRIC>>", "<<METRIC>>" ]
      aggregations = [ "<<AGGREGATION>>", "<<AGGREGATION>>" ]
    
    # defines the resources to collect metrics from
    [[inputs.azure_monitor.resource_group_target.resource]]
      resource_type = "<<RESOURCE_TYPE>>"
      metrics = [ "<<METRIC>>", "<<METRIC>>" ]
      aggregations = [ "<<AGGREGATION>>", "<<AGGREGATION>>" ]
      
  # resource group target #2 to collect metrics from resources under it with resource type
  [[inputs.azure_monitor.resource_group_target]]
    resource_group = "<<RESOURCE_GROUP_NAME>>"

    [[inputs.azure_monitor.resource_group_target.resource]]
      resource_type = "<<RESOURCE_TYPE>>"
      metrics = [ "<<METRIC>>", "<<METRIC>>" ]
      aggregations = [ "<<AGGREGATION>>", "<<AGGREGATION>>" ]
  
  # subscription target #1 to collect metrics from resources under it with resource type    
  [[inputs.azure_monitor.subscription_target]]
    resource_type = "<<RESOURCE_TYPE>>"
    metrics = [ "<<METRIC>>", "<<METRIC>>" ]
    aggregations = [ "<<AGGREGATION>>", "<<AGGREGATION>>" ]
    
  # subscription target #2 to collect metrics from resources under it with resource type    
  [[inputs.azure_monitor.subscription_target]]
    resource_type = "<<RESOURCE_TYPE>>"
    metrics = [ "<<METRIC>>", "<<METRIC>>" ]
    aggregations = [ "<<AGGREGATION>>", "<<AGGREGATION>>" ]
`

func (am *AzureMonitor) Description() string {
	return "Gather Azure resources metrics from Azure Monitor"
}

func (am *AzureMonitor) SampleConfig() string {
	return sampleConfig
}

// Init is for setup, and validating config.
func (am *AzureMonitor) Init() error {
	if err := am.setAzureClients(); err != nil {
		return fmt.Errorf("error setting azure clients: %v", err)
	}

	if err := am.checkConfigValidation(); err != nil {
		return fmt.Errorf("config is not valid: %v", err)
	}

	am.addPrefixToResourceTargetsResourceID()

	if err := am.createResourceTargetsFromResourceGroupTargets(); err != nil {
		return fmt.Errorf("error creating resource targets from resource group targets: %v", err)
	}

	if err := am.createResourceTargetsFromSubscriptionTargets(); err != nil {
		return fmt.Errorf("error creating resource targets from subscription targets: %v", err)
	}

	if err := am.checkResourceTargetsMetricsValidation(); err != nil {
		return fmt.Errorf("error checking resource targets metrics validation: %v", err)
	}

	if err := am.setResourceTargetsMetrics(); err != nil {
		return fmt.Errorf("error setting resource targets metrics: %v", err)
	}

	if err := am.checkResourceTargetsMetricsMinTimeGrain(); err != nil {
		return fmt.Errorf("error checking resource targets metrics min time grain: %v", err)
	}

	am.checkResourceTargetsMaxMetrics()
	am.changeResourceTargetsMetricsWithComma()
	am.setResourceTargetsAggregations()

	am.Log.Debug("Total resource targets: ", len(am.ResourceTargets))

	return nil
}

func (am *AzureMonitor) Gather(acc telegraf.Accumulator) error {
	am.collectResourceTargetsMetrics(acc)

	return nil
}

func (am *AzureMonitor) setAzureClients() error {
	if am.azureClients != nil {
		return nil
	}

	credential, err := azidentity.NewClientSecretCredential(am.TenantID, am.ClientID, am.ClientSecret, nil)
	if err != nil {
		return fmt.Errorf("error creating Azure client credentials: %v", err)
	}

	am.azureClients = &azureClients{
		ctx:                     context.Background(),
		resourcesClient:         newAzureResourcesClient(am.SubscriptionID, credential),
		metricsClient:           armmonitor.NewMetricsClient(credential, nil),
		metricDefinitionsClient: armmonitor.NewMetricDefinitionsClient(credential, nil),
	}

	return nil
}

func init() {
	inputs.Add("azure_monitor", func() telegraf.Input {
		return &AzureMonitor{}
	})
}
