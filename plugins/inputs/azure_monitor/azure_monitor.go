package azure_monitor

import (
	"fmt"
	"sync"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	receiver "github.com/logzio/azure-monitor-metrics-receiver"
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

	receiver     *receiver.AzureMonitorMetricsReceiver
	azureClients *receiver.AzureClients
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

const sampleConfig = `
  # can be found under Overview->Essentials in the Azure portal for your application/service
  subscription_id = "<<SUBSCRIPTION_ID>>"
  # can be obtained by registering an application under Azure Active Directory
  client_id = "<<CLIENT_ID>>"
  # can be obtained by registering an application under Azure Active Directory
  client_secret = "<<CLIENT_SECRET>>"
  # can be found under Azure Active Directory->Properties
  tenant_id = "<<TENANT_ID>>"

  # resource target #1 to collect metrics of
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
    
  # resource target #2 to collect metrics of
  [[inputs.azure_monitor.resource_target]]
    resource_id = "<<RESOURCE_ID>>"
    metrics = [ "<<METRIC>>", "<<METRIC>>" ]
    aggregations = [ "<<AGGREGATION>>", "<<AGGREGATION>>" ]

  # resource group target #1 to collect metrics of resources under it with resource type
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
      
  # resource group target #2 to collect metrics of resources under it with resource type
  [[inputs.azure_monitor.resource_group_target]]
    resource_group = "<<RESOURCE_GROUP_NAME>>"

    [[inputs.azure_monitor.resource_group_target.resource]]
      resource_type = "<<RESOURCE_TYPE>>"
      metrics = [ "<<METRIC>>", "<<METRIC>>" ]
      aggregations = [ "<<AGGREGATION>>", "<<AGGREGATION>>" ]
  
  # subscription target #1 to collect metrics of resources under it with resource type    
  [[inputs.azure_monitor.subscription_target]]
    resource_type = "<<RESOURCE_TYPE>>"
    metrics = [ "<<METRIC>>", "<<METRIC>>" ]
    aggregations = [ "<<AGGREGATION>>", "<<AGGREGATION>>" ]
    
  # subscription target #2 to collect metrics of resources under it with resource type    
  [[inputs.azure_monitor.subscription_target]]
    resource_type = "<<RESOURCE_TYPE>>"
    metrics = [ "<<METRIC>>", "<<METRIC>>" ]
    aggregations = [ "<<AGGREGATION>>", "<<AGGREGATION>>" ]
`

func (am *AzureMonitor) Description() string {
	return "Gather Azure resources metrics using Azure Monitor API"
}

func (am *AzureMonitor) SampleConfig() string {
	return sampleConfig
}

// Init is for setup, and validating config.
func (am *AzureMonitor) Init() error {
	if am.azureClients == nil {
		azureClients, err := receiver.CreateAzureClients(am.SubscriptionID, am.ClientID, am.ClientSecret, am.TenantID)
		if err != nil {
			return err
		}

		am.azureClients = azureClients
	}

	if err := am.setReceiver(); err != nil {
		return fmt.Errorf("error setting Azure Monitor receiver: %w", err)
	}

	if err := am.receiver.CreateResourceTargetsFromResourceGroupTargets(); err != nil {
		return fmt.Errorf("error creating resource targets from resource group targets: %w", err)
	}

	if err := am.receiver.CreateResourceTargetsFromSubscriptionTargets(); err != nil {
		return fmt.Errorf("error creating resource targets from subscription targets: %w", err)
	}

	if err := am.receiver.CheckResourceTargetsMetricsValidation(); err != nil {
		return fmt.Errorf("error checking resource targets metrics validation: %w", err)
	}

	if err := am.receiver.SetResourceTargetsMetrics(); err != nil {
		return fmt.Errorf("error setting resource targets metrics: %w", err)
	}

	if err := am.receiver.SplitResourceTargetsMetricsByMinTimeGrain(); err != nil {
		return fmt.Errorf("error spliting resource targets metrics by min time grain: %w", err)
	}

	am.receiver.SplitResourceTargetsWithMoreThanMaxMetrics()
	am.receiver.SetResourceTargetsAggregations()

	am.Log.Debug("Total resource targets: ", len(am.receiver.Targets.ResourceTargets))

	return nil
}

func (am *AzureMonitor) Gather(acc telegraf.Accumulator) error {
	var waitGroup sync.WaitGroup

	for _, target := range am.receiver.Targets.ResourceTargets {
		am.Log.Debug("Collecting metrics for resource target ", target.ResourceID)
		waitGroup.Add(1)

		go func(target *receiver.ResourceTarget) {
			defer waitGroup.Done()

			collectedMetrics, notCollectedMetrics, err := am.receiver.CollectResourceTargetMetrics(target)
			if err != nil {
				acc.AddError(err)
			}

			for _, collectedMetric := range collectedMetrics {
				acc.AddFields(collectedMetric.Name, collectedMetric.Fields, collectedMetric.Tags)
			}

			for _, notCollectedMetric := range notCollectedMetrics {
				am.Log.Info("Did not get any metric value from Azure Monitor API for the metric ID ", notCollectedMetric)
			}
		}(target)

		waitGroup.Wait()
	}

	return nil
}

func (am *AzureMonitor) setReceiver() error {
	var resourceTargets []*receiver.ResourceTarget
	var resourceGroupTargets []*receiver.ResourceGroupTarget
	var subscriptionTargets []*receiver.Resource

	for _, target := range am.ResourceTargets {
		resourceTargets = append(resourceTargets, receiver.NewResourceTarget(target.ResourceID, target.Metrics, target.Aggregations))
	}

	for _, target := range am.ResourceGroupTargets {
		var resources []*receiver.Resource
		for _, resource := range target.Resources {
			resources = append(resources, receiver.NewResource(resource.ResourceType, resource.Metrics, resource.Aggregations))
		}

		resourceGroupTargets = append(resourceGroupTargets, receiver.NewResourceGroupTarget(target.ResourceGroup, resources))
	}

	for _, target := range am.SubscriptionTargets {
		subscriptionTargets = append(subscriptionTargets, receiver.NewResource(target.ResourceType, target.Metrics, target.Aggregations))
	}

	targets := receiver.NewTargets(resourceTargets, resourceGroupTargets, subscriptionTargets)
	var err error
	am.receiver, err = receiver.NewAzureMonitorMetricsReceiver(am.SubscriptionID, am.ClientID, am.ClientSecret, am.TenantID, targets, am.azureClients)
	return err
}

func init() {
	inputs.Add("azure_monitor", func() telegraf.Input {
		return &AzureMonitor{}
	})
}
