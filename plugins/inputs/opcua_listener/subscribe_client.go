package opcua_listener

import (
	"context"
	"fmt"
	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/ua"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/opcua/input"
	"reflect"
	"time"
)

type SubscribeClientConfig struct {
	input.InputClientConfig
	SubscriptionInterval config.Duration `toml:"subscription_interval"`
}

type SubscribeClient struct {
	*input.OpcUAInputClient
	Config SubscribeClientConfig

	sub                *opcua.Subscription
	monitoredItemsReqs []*ua.MonitoredItemCreateRequest
	dataNotifications  chan *opcua.PublishNotificationData
	metrics            chan telegraf.Metric

	processingCtx    context.Context
	processingCancel context.CancelFunc
}

func (sc *SubscribeClientConfig) CreateSubscribeClient(log telegraf.Logger) (*SubscribeClient, error) {
	client, err := sc.InputClientConfig.CreateInputClient(log)
	if err != nil {
		return nil, err
	}

	subClient := &SubscribeClient{
		OpcUAInputClient:   client,
		Config:             *sc,
		monitoredItemsReqs: make([]*ua.MonitoredItemCreateRequest, len(client.NodeIDs)),
		dataNotifications:  make(chan *opcua.PublishNotificationData, 100), // TODO: Make this a setting or a calculation, maybe?,
		metrics:            make(chan telegraf.Metric, 100),                // TODO: Make this a setting or a calculation, maybe?,
	}

	log.Debugf("Creating monitored items")
	for i, nodeID := range client.NodeIDs {
		// The node id index (i) is used as the handle for the monitored item
		req := opcua.NewMonitoredItemCreateRequestWithDefaults(nodeID, ua.AttributeIDValue, uint32(i))
		subClient.monitoredItemsReqs[i] = req
	}

	return subClient, nil
}

func (o *SubscribeClient) Connect() error {
	err := o.OpcUAClient.Connect()
	if err != nil {
		return err
	}

	o.Log.Debugf("Creating OPC UA subscription")
	o.sub, err = o.Client.Subscribe(&opcua.SubscriptionParameters{
		Interval: time.Duration(o.Config.SubscriptionInterval),
	}, o.dataNotifications)
	if err != nil {
		o.Log.Error("Failed to create subscription")
		return err
	}

	o.Log.Debugf("Subscribed with subscription ID %d", o.sub.SubscriptionID)
	return nil
}

func (o *SubscribeClient) Stop(ctx context.Context) <-chan struct{} {
	o.Log.Debugf("Opc Subscribe Stopped")
	err := o.sub.Cancel(ctx)
	if err != nil {
		o.Log.Warn("Cancelling OPC UA subscription failed with error ", err)
	}
	closing := o.OpcUAInputClient.Stop(ctx)
	o.processingCancel()
	return closing
}

func (o *SubscribeClient) CurrentValues() ([]telegraf.Metric, error) {
	return []telegraf.Metric{}, nil
}

func (o *SubscribeClient) StartStreamValues(ctx context.Context) (<-chan telegraf.Metric, error) {
	err := o.Connect()
	if err != nil {
		return nil, err
	}

	resp, err := o.sub.MonitorWithContext(ctx, ua.TimestampsToReturnBoth, o.monitoredItemsReqs...)
	if err != nil {
		o.Log.Error("Failed to create monitored items ", err)
		return nil, fmt.Errorf("failed to start monitoring items %s", err)
	}
	o.Log.Debug("Monitoring items")

	for _, res := range resp.Results {
		if !o.StatusCodeOK(res.StatusCode) {
			return nil, fmt.Errorf("creating monitored item failed with status code %d", res.StatusCode)
		}
	}

	o.processingCtx, o.processingCancel = context.WithCancel(context.Background())
	go o.processReceivedNotifications()

	return o.metrics, nil
}

func (o *SubscribeClient) processReceivedNotifications() {
	for {
		select {
		case <-o.processingCtx.Done():
			o.Log.Debug("Processing received notifications stopped")
			return

		case res, ok := <-o.dataNotifications:
			if !ok {
				o.Log.Debugf("Data notification channel closed. Processing of received notifications stopped")
				return
			}
			if res.Error != nil {
				o.Log.Error(res.Error)
				continue
			}

			switch notif := res.Value.(type) {
			case *ua.DataChangeNotification:
				o.Log.Debugf("Received data change notification with %d items", len(notif.MonitoredItems))
				// It is assumed the notifications are ordered chronologically
				for _, monitoredItemNotif := range notif.MonitoredItems {
					i := int(monitoredItemNotif.ClientHandle)
					oldValue := o.LastReceivedData[i].Value
					o.UpdateNodeValue(i, monitoredItemNotif.Value)
					o.Log.Debugf("Data change notification: node '%s' value changed from %f to %f",
						o.NodeIDs[i].String(), oldValue, o.LastReceivedData[i].Value)
					o.metrics <- o.MetricForNode(i)
				}

			default:
				o.Log.Warnf("Received notification has unexpected type %s", reflect.TypeOf(res.Value))
			}
		}
	}
}
