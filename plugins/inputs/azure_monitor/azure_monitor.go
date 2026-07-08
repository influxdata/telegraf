//go:generate ../../../tools/readme_config_includer/generator
package azure_monitor

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"sync"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

const maxMetricsBatch = 20

type AzureMonitor struct {
	SubscriptionID       string                 `toml:"subscription_id"`
	ClientID             string                 `toml:"client_id"`
	ClientSecret         config.Secret          `toml:"client_secret"`
	TenantID             string                 `toml:"tenant_id"`
	CloudOption          string                 `toml:"cloud_option,omitempty"`
	ResourceTargets      []*resourceTarget      `toml:"resource_target"`
	ResourceGroupTargets []*resourceGroupTarget `toml:"resource_group_target"`
	SubscriptionTargets  []*resource            `toml:"subscription_target"`
	Log                  telegraf.Logger        `toml:"-"`

	client   client
	factory  clientFactory
	receiver *metricReceiver
}

type resourceTarget struct {
	ResourceID   string   `toml:"resource_id"`
	Metrics      []string `toml:"metrics"`
	Aggregations []string `toml:"aggregations"`
}

type resourceGroupTarget struct {
	ResourceGroup string      `toml:"resource_group"`
	Resources     []*resource `toml:"resource"`
}

type resource struct {
	ResourceType string   `toml:"resource_type"`
	Metrics      []string `toml:"metrics"`
	Aggregations []string `toml:"aggregations"`
}

func (*AzureMonitor) SampleConfig() string {
	return sampleConfig
}

func (am *AzureMonitor) Init() error {
	// Validate settings
	if am.SubscriptionID == "" {
		return errors.New("subscription_id is required")
	}

	if len(am.ResourceTargets) == 0 && len(am.ResourceGroupTargets) == 0 && len(am.SubscriptionTargets) == 0 {
		return errors.New("no target to collect metrics from")
	}

	// Validate the targets
	if err := am.validateTargets(); err != nil {
		return err
	}

	// Canonicalize resource target IDs
	for _, target := range am.ResourceTargets {
		target.ResourceID = "/subscriptions/" + am.SubscriptionID + "/" + target.ResourceID
	}

	// Setup client options
	var clientOptions azcore.ClientOptions
	switch am.CloudOption {
	case "AzureChina":
		clientOptions.Cloud = cloud.AzureChina
	case "AzureGovernment":
		clientOptions.Cloud = cloud.AzureGovernment
	case "", "AzurePublic":
		clientOptions.Cloud = cloud.AzurePublic
	default:
		return fmt.Errorf("unknown cloud option: %s", am.CloudOption)
	}

	var clientSecret string
	if !am.ClientSecret.Empty() {
		if am.ClientID == "" {
			return errors.New("client_id is required when client_secret is set")
		}
		if am.TenantID == "" {
			return errors.New("tenant_id is required when client_secret is set")
		}
		secret, err := am.ClientSecret.Get()
		if err != nil {
			return fmt.Errorf("getting client secret failed: %w", err)
		}
		clientSecret = secret.String()
		secret.Destroy()
	}

	// Create a new client. This does not connect to the Azure API.
	client, err := am.factory.createClient(am.SubscriptionID, am.ClientID, clientSecret, am.TenantID, clientOptions)
	if err != nil {
		return fmt.Errorf("creating client failed: %w", err)
	}
	am.client = client

	return nil
}

func (am *AzureMonitor) Start(telegraf.Accumulator) error {
	receiver, err := newReceiver(
		context.Background(), am.client, am.SubscriptionID, am.ResourceTargets, am.ResourceGroupTargets, am.SubscriptionTargets,
	)
	if err != nil {
		return fmt.Errorf("creating receiver failed: %w", err)
	}
	am.receiver = receiver
	am.Log.Debug("Total resource targets: ", len(am.receiver.resources))

	return nil
}

func (am *AzureMonitor) Stop() {
	am.receiver = nil
}

func (am *AzureMonitor) Gather(acc telegraf.Accumulator) error {
	var wg sync.WaitGroup

	ctx := context.Background()
	for _, target := range am.receiver.resources {
		am.Log.Debug("Collecting metrics for resource target ", target.ResourceID)

		wg.Add(1)
		go func(target *resourceTarget) {
			defer wg.Done()

			am.receiver.collectMetrics(ctx, acc, target, am.Log)
		}(target)
	}
	wg.Wait()

	return nil
}

func (am *AzureMonitor) validateTargets() error {
	// Validate resource targets
	for index, target := range am.ResourceTargets {
		if target.ResourceID == "" {
			return fmt.Errorf("missing resource ID in resource target #%d", index+1)
		}
	}

	// Validate resource group targets
	for index, target := range am.ResourceGroupTargets {
		if target.ResourceGroup == "" {
			return fmt.Errorf("missing resource group in resource group target #%d", index+1)
		}

		if len(target.Resources) == 0 {
			return fmt.Errorf("no resources in resource group target #%d", index+1)
		}

		for resourceIndex, resource := range target.Resources {
			if resource.ResourceType == "" {
				return fmt.Errorf("no resource type for resource group target #%d resource #%d", index+1, resourceIndex+1)
			}
		}
	}

	// Validate subscription targets
	for index, target := range am.SubscriptionTargets {
		if target.ResourceType == "" {
			return fmt.Errorf("missing resource type in subscription target #%d", index+1)
		}
	}

	return nil
}

func init() {
	inputs.Add("azure_monitor", func() telegraf.Input {
		return &AzureMonitor{
			factory: &azureFactory{},
		}
	})
}
