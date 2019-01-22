package activemq

import (
	"context"
	"fmt"
	"github.com/Azure/azure-storage-queue-go/azqueue"
	"log"
	"net/url"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type AzureStorageQueue struct {
	StorageAccountName string `toml:"azure_storage_account_name"`
	StorageAccountKey  string `toml:"azure_storage_account_key"`

	serviceURL *azqueue.ServiceURL
}

var sampleConfig = `
  ## Required Azure Storage Account name
  azure_storage_account_name = "TODO"

  ## Required Azure Storage Account access key
  azure_storage_account_key = "TODO"
  `

func (a *AzureStorageQueue) Description() string {
	return "Gather Azure Storage Queue metrics"
}

func (a *AzureStorageQueue) SampleConfig() string {
	return sampleConfig
}

func (a *AzureStorageQueue) ValidateConfiguration() error {
	if a.StorageAccountName == "" {
		return fmt.Errorf("azure_storage_account must be configured")
	}

	if a.StorageAccountKey == "" {
		return fmt.Errorf("azure_storage_account_key must be configured")
	}
	return nil
}

func (a *AzureStorageQueue) GetURL() (*url.URL, error) {
	err := a.ValidateConfiguration()
	if err != nil {
		return nil, err
	}

	return url.Parse("https://" + a.StorageAccountName + ".queue.core.windows.net")
}

func (a *AzureStorageQueue) GetServiceURL() (azqueue.ServiceURL, error) {
	if a.serviceURL == nil {
		_url, err := a.GetURL() // Will also validate StorageAccountKey
		if err != nil {
			return azqueue.ServiceURL{}, err
		}

		credential, err := azqueue.NewSharedKeyCredential(a.StorageAccountName, a.StorageAccountKey)
		if err != nil {
			return azqueue.ServiceURL{}, err
		}

		pipeline := azqueue.NewPipeline(credential, azqueue.PipelineOptions{})

		serviceURL := azqueue.NewServiceURL(*_url, pipeline)
		a.serviceURL = &serviceURL
	}
	return *a.serviceURL, nil
}

func (a *AzureStorageQueue) GatherQueueMetrics(acc telegraf.Accumulator, queueItem azqueue.QueueItem, properties *azqueue.QueueGetPropertiesResponse) {
	records := make(map[string]interface{})
	tags := make(map[string]string)
	tags["name"] = strings.TrimSpace(queueItem.Name)
	tags["storage_account"] = a.StorageAccountName
	records["size"] = properties.ApproximateMessagesCount()
	acc.AddFields("azure_storage_queues", records, tags)
}

func (a *AzureStorageQueue) Gather(acc telegraf.Accumulator) error {
	serviceURL, err := a.GetServiceURL()
	if err != nil {
		return err
	}

	ctx := context.TODO() // This example uses a never-expiring context

	for marker := (azqueue.Marker{}); marker.NotDone(); {
		queuesSegment, err := serviceURL.ListQueuesSegment(ctx, marker,
			azqueue.ListQueuesSegmentOptions{
				Detail: azqueue.ListQueuesSegmentDetails{Metadata: false},
			})
		if err != nil {
			// log.Fatal(err)
			return err
		}
		marker = queuesSegment.NextMarker

		for _, queueItem := range queuesSegment.QueueItems {
			queueURL := serviceURL.NewQueueURL(queueItem.Name)
			properties, err := queueURL.GetProperties(ctx)
			if err != nil {
				log.Fatal(err)
			} else {
				a.GatherQueueMetrics(acc, queueItem, properties)
			}
		}
	}
	return nil
}

func init() {
	inputs.Add("azure_storage_queue", func() telegraf.Input {
		return &AzureStorageQueue{}
	})
}
