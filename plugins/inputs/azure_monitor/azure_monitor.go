//go:generate ../../../tools/readme_config_includer/generator
package azure_monitor

import (
	_ "embed"
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
	azureManager azureClientsCreator
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

type azureClientsManager struct{}

type azureClientsCreator interface {
	createAzureClients(subscriptionID string, clientID string, clientSecret string, tenantID string) (*receiver.AzureClients, error)
}

//go:embed sample.conf
var sampleConfig string

func (am *AzureMonitor) SampleConfig() string {
	return sampleConfig
}

// Init is for setup, and validating config.
func (am *AzureMonitor) Init() error {
	var err error
	am.azureClients, err = am.azureManager.createAzureClients(am.SubscriptionID, am.ClientID, am.ClientSecret, am.TenantID)
	if err != nil {
		return err
	}

	if err = am.setReceiver(); err != nil {
		return fmt.Errorf("error setting Azure Monitor receiver: %w", err)
	}

	if err = am.receiver.CreateResourceTargetsFromResourceGroupTargets(); err != nil {
		return fmt.Errorf("error creating resource targets from resource group targets: %w", err)
	}

	if err = am.receiver.CreateResourceTargetsFromSubscriptionTargets(); err != nil {
		return fmt.Errorf("error creating resource targets from subscription targets: %w", err)
	}

	if err = am.receiver.CheckResourceTargetsMetricsValidation(); err != nil {
		return fmt.Errorf("error checking resource targets metrics validation: %w", err)
	}

	if err = am.receiver.SetResourceTargetsMetrics(); err != nil {
		return fmt.Errorf("error setting resource targets metrics: %w", err)
	}

	if err = am.receiver.SplitResourceTargetsMetricsByMinTimeGrain(); err != nil {
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
	}

	waitGroup.Wait()
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

func (acm *azureClientsManager) createAzureClients(
	subscriptionID string,
	clientID string,
	clientSecret string,
	tenantID string,
) (*receiver.AzureClients, error) {
	azureClients, err := receiver.CreateAzureClients(subscriptionID, clientID, clientSecret, tenantID)
	if err != nil {
		return nil, fmt.Errorf("error creating Azure clients: %w", err)
	}

	return azureClients, nil
}

func init() {
	inputs.Add("azure_monitor", func() telegraf.Input {
		return &AzureMonitor{
			azureManager: &azureClientsManager{},
		}
	})
}
