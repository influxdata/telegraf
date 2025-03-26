//go:generate ../../../tools/readme_config_includer/generator
package azure_storage_queue

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azqueue"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type AzureStorageQueue struct {
	EndpointURL          string `toml:"endpoint"`
	StorageAccountName   string `toml:"account_name"`
	StorageAccountKey    string `toml:"account_key"`
	PeekOldestMessageAge bool   `toml:"peek_oldest_message_age"`
	Log                  telegraf.Logger

	client *azqueue.ServiceClient
}

func (*AzureStorageQueue) SampleConfig() string {
	return sampleConfig
}

func (a *AzureStorageQueue) Init() error {
	// Check settings
	if a.StorageAccountName == "" {
		return errors.New("account_name must be configured")
	}

	if a.StorageAccountKey == "" {
		return errors.New("account_key must be configured")
	}

	// Prepare the client
	if a.EndpointURL == "" {
		a.EndpointURL = "https://" + a.StorageAccountName + ".queue.core.windows.net"
	}
	credentials, err := azqueue.NewSharedKeyCredential(a.StorageAccountName, a.StorageAccountKey)
	if err != nil {
		return fmt.Errorf("creating shared-key credentials failed: %w", err)
	}

	client, err := azqueue.NewServiceClientWithSharedKeyCredential(a.EndpointURL, credentials, nil)
	if err != nil {
		return fmt.Errorf("creating client failed: %w", err)
	}
	a.client = client

	return nil
}

func (a *AzureStorageQueue) Gather(acc telegraf.Accumulator) error {
	ctx := context.Background()

	a.Log.Debugf("Listing queues of storage account %q", a.StorageAccountName)

	// Iterate through the queues and generate metrics
	pages := a.client.NewListQueuesPager(nil)
	for pages.More() {
		response, err := pages.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("getting next page failed: %w", err)
		}

		// Get the properties and the message properties for each of the queues
		for _, queue := range response.Queues {
			if queue.Name == nil {
				continue
			}
			name := strings.TrimSpace(*queue.Name)

			// Access the queue and get the properties
			c := a.client.NewQueueClient(*queue.Name)
			props, err := c.GetProperties(ctx, nil)
			if err != nil {
				acc.AddError(fmt.Errorf("getting properties for queue %q failed: %w", name, err))
				continue
			}
			if props.ApproximateMessagesCount == nil {
				acc.AddError(fmt.Errorf("unset message count for queue %q", name))
				continue
			}

			// Setup the metric elements
			tags := map[string]string{
				"account": a.StorageAccountName,
				"queue":   strings.TrimSpace(name),
			}
			fields := map[string]interface{}{
				"size": *props.ApproximateMessagesCount,
			}
			now := time.Now()
			if a.PeekOldestMessageAge {
				if r, err := c.PeekMessage(ctx, nil); err != nil {
					acc.AddError(fmt.Errorf("peeking message for queue %q failed: %w", name, err))
				} else if len(r.Messages) > 0 && r.Messages[0] != nil && r.Messages[0].InsertionTime != nil {
					msg := r.Messages[0]
					fields["oldest_message_age_ns"] = now.Sub(*msg.InsertionTime).Nanoseconds()
				}
			}
			acc.AddFields("azure_storage_queues", fields, tags, now)
		}
	}

	return nil
}

func init() {
	inputs.Add("azure_storage_queue", func() telegraf.Input {
		return &AzureStorageQueue{PeekOldestMessageAge: true}
	})
}
