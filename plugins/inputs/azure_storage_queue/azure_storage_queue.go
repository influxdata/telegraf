package azure_storage_queue

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/Azure/azure-storage-queue-go/azqueue"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type AzureStorageQueue struct {
	StorageAccountName   string `toml:"account_name"`
	StorageAccountKey    string `toml:"account_key"`
	PeekOldestMessageAge bool   `toml:"peek_oldest_message_age"`
	Log                  telegraf.Logger

	serviceURL *azqueue.ServiceURL
}

func (a *AzureStorageQueue) Init() error {
	if a.StorageAccountName == "" {
		return errors.New("account_name must be configured")
	}

	if a.StorageAccountKey == "" {
		return errors.New("account_key must be configured")
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
	tags["queue"] = strings.TrimSpace(queueItem.Name)
	tags["account"] = a.StorageAccountName
	fields["size"] = properties.ApproximateMessagesCount()
	if peekedMessage != nil {
		fields["oldest_message_age_ns"] = time.Now().UnixNano() - peekedMessage.InsertionTime.UnixNano()
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
		a.Log.Debugf("Listing queues of storage account '%s'", a.StorageAccountName)
		queuesSegment, err := serviceURL.ListQueuesSegment(ctx, marker,
			azqueue.ListQueuesSegmentOptions{
				Detail: azqueue.ListQueuesSegmentDetails{Metadata: false},
			})
		if err != nil {
			return err
		}
		marker = queuesSegment.NextMarker

		for _, queueItem := range queuesSegment.QueueItems {
			a.Log.Debugf("Processing queue '%s' of storage account '%s'", queueItem.Name, a.StorageAccountName)
			queueURL := serviceURL.NewQueueURL(queueItem.Name)
			properties, err := queueURL.GetProperties(ctx)
			if err != nil {
				a.Log.Errorf("Error getting properties for queue %s: %s", queueItem.Name, err.Error())
				continue
			}
			var peekedMessage *azqueue.PeekedMessage
			if a.PeekOldestMessageAge {
				messagesURL := queueURL.NewMessagesURL()
				messagesResponse, err := messagesURL.Peek(ctx, 1)
				if err != nil {
					a.Log.Errorf("Error peeking queue %s: %s", queueItem.Name, err.Error())
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
