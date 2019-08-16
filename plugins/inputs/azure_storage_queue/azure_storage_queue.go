package activemq

import (
	"context"
	"errors"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/Azure/azure-storage-queue-go/azqueue"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type AzureStorageQueue struct {
	StorageAccountName   string `toml:"azure_storage_account_name"`
	StorageAccountKey    string `toml:"azure_storage_account_key"`
	PeekOldestMessageAge bool   `toml:"peek_oldest_message_age"`

	serviceURL *azqueue.ServiceURL
}

var sampleConfig = `
  ## Required Azure Storage Account name
  azure_storage_account_name = "mystorageaccount"

  ## Required Azure Storage Account access key
  azure_storage_account_key = "storageaccountaccesskey"

  ## Set to false to disable peeking age of oldest message (executes faster)
  # peek_oldest_message_age = true
  `

func (a *AzureStorageQueue) Description() string {
	return "Gather Azure Storage Queue metrics"
}

func (a *AzureStorageQueue) SampleConfig() string {
	return sampleConfig
}

func (a *AzureStorageQueue) Init() error {
	if a.StorageAccountName == "" {
		return errors.New("azure_storage_account must be configured")
	}

	if a.StorageAccountKey == "" {
		return errors.New("azure_storage_account_key must be configured")
	}
	return nil
}

func (a *AzureStorageQueue) GetServiceURL() (azqueue.ServiceURL, error) {
	if a.serviceURL == nil {
		_url, err := url.Parse("https://" + a.StorageAccountName + ".queue.core.windows.net")
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

func (a *AzureStorageQueue) GatherQueueMetrics(acc telegraf.Accumulator, queueItem azqueue.QueueItem, properties *azqueue.QueueGetPropertiesResponse, peekedMessage *azqueue.PeekedMessage) {
	fields := make(map[string]interface{})
	tags := make(map[string]string)
	tags["name"] = strings.TrimSpace(queueItem.Name)
	tags["storage_account"] = a.StorageAccountName
	fields["size"] = properties.ApproximateMessagesCount()
	if peekedMessage != nil {
		fields["oldest_message_age"] = time.Now().Unix() - peekedMessage.InsertionTime.Unix()
	}
	acc.AddFields("azure_storage_queues", fields, tags)
}

func (a *AzureStorageQueue) Gather(acc telegraf.Accumulator) error {
	serviceURL, err := a.GetServiceURL()
	if err != nil {
		return err
	}

	ctx := context.TODO()

	for marker := (azqueue.Marker{}); marker.NotDone(); {
		queuesSegment, err := serviceURL.ListQueuesSegment(ctx, marker,
			azqueue.ListQueuesSegmentOptions{
				Detail: azqueue.ListQueuesSegmentDetails{Metadata: false},
			})
		if err != nil {
			return err
		}
		marker = queuesSegment.NextMarker

		for _, queueItem := range queuesSegment.QueueItems {
			queueURL := serviceURL.NewQueueURL(queueItem.Name)
			properties, err := queueURL.GetProperties(ctx)
			if err != nil {
				log.Printf("E! [inputs.azure_storage_queue] Error getting properties for queue %s: %s", queueItem.Name, err.Error())
				continue
			}
			var peekedMessage *azqueue.PeekedMessage
			if a.PeekOldestMessageAge {
				messagesURL := queueURL.NewMessagesURL()
				messagesResponse, err := messagesURL.Peek(ctx, 1)
				if err != nil {
					log.Printf("E! [inputs.azure_storage_queue] Error peeking queue %s: %s", queueItem.Name, err.Error())
				} else if messagesResponse.NumMessages() > 0 {
					peekedMessage = messagesResponse.Message(0)
				}
			}

			a.GatherQueueMetrics(acc, queueItem, properties, peekedMessage)
		}
	}
	return nil
}

func init() {
	inputs.Add("azure_storage_queue", func() telegraf.Input {
		return &AzureStorageQueue{PeekOldestMessageAge: true}
	})
}
