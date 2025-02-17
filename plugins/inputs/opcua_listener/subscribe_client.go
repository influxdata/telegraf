package opcua_listener

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/id"
	"github.com/gopcua/opcua/ua"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	opcuaclient "github.com/influxdata/telegraf/plugins/common/opcua"
	"github.com/influxdata/telegraf/plugins/common/opcua/input"
)

type subscribeClientConfig struct {
	input.InputClientConfig
	SubscriptionInterval config.Duration `toml:"subscription_interval"`
	ConnectFailBehavior  string          `toml:"connect_fail_behavior"`
}

type subscribeClient struct {
	*input.OpcUAInputClient
	Config subscribeClientConfig

	sub                  *opcua.Subscription
	monitoredItemsReqs   []*ua.MonitoredItemCreateRequest
	eventItemsReqs       []*ua.MonitoredItemCreateRequest
	ClientHandleToNodeID sync.Map
	dataNotifications    chan *opcua.PublishNotificationData
	metrics              chan telegraf.Metric

	acc    telegraf.Accumulator
	ctx    context.Context
	cancel context.CancelFunc
}

func checkDataChangeFilterParameters(params *input.DataChangeFilter) error {
	switch {
	case params.Trigger != input.Status &&
		params.Trigger != input.StatusValue &&
		params.Trigger != input.StatusValueTimestamp:
		return fmt.Errorf("trigger '%s' not supported", params.Trigger)
	case params.DeadbandType != input.Absolute &&
		params.DeadbandType != input.Percent:
		return fmt.Errorf("deadband_type '%s' not supported", params.DeadbandType)
	case params.DeadbandValue == nil:
		return errors.New("deadband_value was not set")
	case *params.DeadbandValue < 0:
		return errors.New("negative deadband_value not supported")
	default:
		return nil
	}
}

func assignConfigValuesToRequest(req *ua.MonitoredItemCreateRequest, monParams *input.MonitoringParameters) error {
	req.RequestedParameters.SamplingInterval = float64(time.Duration(monParams.SamplingInterval) / time.Millisecond)

	if monParams.QueueSize != nil {
		req.RequestedParameters.QueueSize = *monParams.QueueSize
	}

	if monParams.DiscardOldest != nil {
		req.RequestedParameters.DiscardOldest = *monParams.DiscardOldest
	}

	if monParams.DataChangeFilter != nil {
		if err := checkDataChangeFilterParameters(monParams.DataChangeFilter); err != nil {
			return fmt.Errorf(err.Error()+", node '%s'", req.ItemToMonitor.NodeID)
		}

		req.RequestedParameters.Filter = ua.NewExtensionObject(
			&ua.DataChangeFilter{
				Trigger:       ua.DataChangeTriggerFromString(string(monParams.DataChangeFilter.Trigger)),
				DeadbandType:  uint32(ua.DeadbandTypeFromString(string(monParams.DataChangeFilter.DeadbandType))),
				DeadbandValue: *monParams.DataChangeFilter.DeadbandValue,
			},
		)
	}

	return nil
}

func (sc *subscribeClientConfig) createSubscribeClient(log telegraf.Logger) (*subscribeClient, error) {
	client, err := sc.InputClientConfig.CreateInputClient(log)
	if err != nil {
		return nil, err
	}

	if err := client.InitNodeIDs(); err != nil {
		return nil, err
	}

	processingCtx, processingCancel := context.WithCancel(context.Background())
	subClient := &subscribeClient{
		OpcUAInputClient:   client,
		Config:             *sc,
		monitoredItemsReqs: make([]*ua.MonitoredItemCreateRequest, len(client.NodeIDs)),
		eventItemsReqs:     make([]*ua.MonitoredItemCreateRequest, len(client.EventStreamingInput.NodeIDs)),
		// 100 was chosen to make sure that the channels will not block when multiple changes come in at the same time.
		// The channel size should be increased if reports come in on Telegraf blocking when many changes come in at
		// the same time. It could be made dependent on the number of nodes subscribed to and the subscription interval.
		dataNotifications: make(chan *opcua.PublishNotificationData, 100),
		metrics:           make(chan telegraf.Metric, 100),
		ctx:               processingCtx,
		cancel:            processingCancel,
	}

	log.Debugf("Creating monitored items")
	for i, nodeID := range client.NodeIDs {
		// The node id index (i) is used as the handle for the monitored item
		req := opcua.NewMonitoredItemCreateRequestWithDefaults(nodeID, ua.AttributeIDValue, uint32(i))
		if err := assignConfigValuesToRequest(req, &client.NodeMetricMapping[i].Tag.MonitoringParams); err != nil {
			return nil, err
		}
		subClient.monitoredItemsReqs[i] = req
	}

	selects := make([]*ua.SimpleAttributeOperand, len(client.EventStreamingInput.Fields))
	for i, name := range client.EventStreamingInput.Fields {
		selects[i] = &ua.SimpleAttributeOperand{
			TypeDefinitionID: ua.NewNumericNodeID(client.EventStreamingInput.EventType.ID.Namespace(), client.EventStreamingInput.EventType.ID.IntID()),
			BrowsePath:       []*ua.QualifiedName{{NamespaceIndex: 0, Name: name}},
			AttributeID:      ua.AttributeIDValue,
		}
	}
	wheres := &ua.ContentFilter{
		Elements: make([]*ua.ContentFilterElement, 0),
	}
	operands := make([]*ua.ExtensionObject, 0)
	for _, sourceName := range client.EventStreamingInput.SourceNames {
		literalOperand := &ua.ExtensionObject{
			EncodingMask: 1,
			TypeID: &ua.ExpandedNodeID{
				NodeID: ua.NewNumericNodeID(0, id.LiteralOperand_Encoding_DefaultBinary),
			},
			Value: ua.LiteralOperand{
				Value: ua.MustVariant(sourceName),
			},
		}
		operands = append(operands, literalOperand)
	}

	attributeOperand := &ua.ExtensionObject{
		EncodingMask: ua.ExtensionObjectBinary,
		TypeID: &ua.ExpandedNodeID{
			NodeID: ua.NewNumericNodeID(0, id.SimpleAttributeOperand_Encoding_DefaultBinary),
		},
		Value: &ua.SimpleAttributeOperand{
			TypeDefinitionID: ua.NewNumericNodeID(client.EventStreamingInput.EventType.ID.Namespace(), client.EventStreamingInput.EventType.ID.IntID()),
			BrowsePath: []*ua.QualifiedName{
				{NamespaceIndex: 0, Name: "SourceName"},
			},
			AttributeID: ua.AttributeIDValue,
		},
	}

	filterElement := &ua.ContentFilterElement{
		FilterOperator: ua.FilterOperatorInList,
		FilterOperands: append([]*ua.ExtensionObject{attributeOperand}, operands...),
	}
	if len(filterElement.FilterOperands) > 0 {
		wheres = &ua.ContentFilter{
			Elements: []*ua.ContentFilterElement{filterElement},
		}
	}

	filter := ua.EventFilter{
		SelectClauses: selects,
		WhereClause:   wheres,
	}

	filterExtObj := ua.ExtensionObject{
		EncodingMask: ua.ExtensionObjectBinary,
		TypeID: &ua.ExpandedNodeID{
			NodeID: ua.NewNumericNodeID(0, id.EventFilter_Encoding_DefaultBinary),
		},
		Value: filter,
	}

	for i, nodeID := range client.EventStreamingInput.NodeIDs {
		eventStreamingReq := &ua.MonitoredItemCreateRequest{
			ItemToMonitor: &ua.ReadValueID{
				NodeID:       nodeID.ID,
				AttributeID:  ua.AttributeIDEventNotifier,
				DataEncoding: &ua.QualifiedName{},
			},
			MonitoringMode: ua.MonitoringModeReporting,
			RequestedParameters: &ua.MonitoringParameters{
				ClientHandle:     uint32(i),
				SamplingInterval: 10000.0, // 10 seconds
				QueueSize:        10,
				DiscardOldest:    true,
				Filter:           &filterExtObj,
			},
		}
		subClient.ClientHandleToNodeID.Store(uint32(i), nodeID.ID.String())
		subClient.eventItemsReqs[i] = eventStreamingReq
	}

	return subClient, nil
}

func (o *subscribeClient) connect() error {
	err := o.OpcUAClient.Connect(o.ctx)
	if err != nil {
		return err
	}

	o.Log.Debugf("Creating OPC UA subscription")
	o.sub, err = o.Client.Subscribe(o.ctx, &opcua.SubscriptionParameters{
		Interval: time.Duration(o.Config.SubscriptionInterval),
	}, o.dataNotifications)
	if err != nil {
		o.Log.Error("Failed to create subscription")
		return err
	}

	o.Log.Debugf("Subscribed with subscription ID %d", o.sub.SubscriptionID)
	return nil
}

func (o *subscribeClient) stop(ctx context.Context) <-chan struct{} {
	o.Log.Debugf("Stopping OPC subscription...")
	if o.State() != opcuaclient.Connected {
		return nil
	}
	if o.sub != nil {
		if err := o.sub.Cancel(ctx); err != nil {
			o.Log.Warn("Cancelling OPC UA subscription failed with error ", err)
		}
	}
	closing := o.OpcUAInputClient.Stop(ctx)
	o.cancel()
	return closing
}

func (o *subscribeClient) startStreamValues(ctx context.Context) (<-chan telegraf.Metric, error) {
	err := o.connect()
	if err != nil {
		switch o.Config.ConnectFailBehavior {
		case "retry":
			o.Log.Warnf("Failed to connect to OPC UA server %s. Will attempt to connect again at the next interval: %s", o.Config.Endpoint, err)
			return nil, nil
		case "ignore":
			o.Log.Errorf("Failed to connect to OPC UA server %s. Will not retry: %s", o.Config.Endpoint, err)
			return nil, nil
		}
		return nil, err
	}

	resp, err := o.sub.Monitor(ctx, ua.TimestampsToReturnBoth, o.monitoredItemsReqs...)
	if err != nil {
		return nil, fmt.Errorf("failed to start monitoring items: %w", err)
	}
	o.Log.Debug("Monitoring items")

	for idx, res := range resp.Results {
		if !o.StatusCodeOK(res.StatusCode) {
			// Verify NodeIDs array has been built before trying to get item; otherwise show '?' for node id
			if len(o.OpcUAInputClient.NodeIDs) > idx {
				o.Log.Debugf("Failed to create monitored item for node %v (%v)",
					o.OpcUAInputClient.NodeMetricMapping[idx].Tag.FieldName, o.OpcUAInputClient.NodeIDs[idx].String())
			} else {
				o.Log.Debugf("Failed to create monitored item for node %v (%v)", o.OpcUAInputClient.NodeMetricMapping[idx].Tag.FieldName, '?')
			}

			return nil, fmt.Errorf("creating monitored item failed with status code: %w", res.StatusCode)
		}
	}

	go o.processReceivedNotifications()

	return o.metrics, nil
}

func (o *subscribeClient) startStreamEvents(ctx context.Context) error {
	err := o.connect()
	if err != nil {
		switch o.Config.ConnectFailBehavior {
		case "retry":
			o.Log.Warnf("Failed to connect to OPC UA server %s. Will attempt to connect again at the next interval: %s", o.Config.Endpoint, err)
			return nil
		case "ignore":
			o.Log.Errorf("Failed to connect to OPC UA server %s. Will not retry: %s", o.Config.Endpoint, err)
			return nil
		}
		return err
	}

	resp, err := o.sub.Monitor(ctx, ua.TimestampsToReturnBoth, o.eventItemsReqs...)
	if err != nil {
		return fmt.Errorf("failed to start monitoring event stream: %w", err)
	}
	o.Log.Debug("Started monitoring event streaming")

	for idx, res := range resp.Results {
		if !o.StatusCodeOK(res.StatusCode) {
			// Verify NodeIDs array has been built before trying to get item; otherwise show '?' for node id
			if len(o.OpcUAInputClient.NodeIDs) > idx {
				o.Log.Debugf("Failed to create monitored item for node %v (%v)",
					o.OpcUAInputClient.NodeMetricMapping[idx].Tag.FieldName, o.OpcUAInputClient.NodeIDs[idx].String())
			} else {
				o.Log.Debugf("Failed to create monitored item for node %v (%v)", o.OpcUAInputClient.NodeMetricMapping[idx].Tag.FieldName, '?')
			}

			return fmt.Errorf("creating monitored item failed with status code: %w", res.StatusCode)
		}
	}

	go o.processReceivedNotifications()

	return nil
}

func (o *subscribeClient) processReceivedNotifications() {
	for {
		select {
		case <-o.ctx.Done():
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
				o.handleDataChange(notif)
			case *ua.EventNotificationList:
				o.handleEventNotification(notif)
			default:
				o.Log.Warnf("Received notification has unexpected type %s", reflect.TypeOf(res.Value))
			}
		}
	}
}

func (o *subscribeClient) handleDataChange(notif *ua.DataChangeNotification) {
	o.Log.Debugf("Received data change notification with %d items", len(notif.MonitoredItems))
	// It is assumed the notifications are ordered chronologically
	for _, monitoredItemNotif := range notif.MonitoredItems {
		i := int(monitoredItemNotif.ClientHandle)
		oldValue := o.LastReceivedData[i].Value
		o.UpdateNodeValue(i, monitoredItemNotif.Value)
		o.Log.Debugf("Data change notification: node %q value changed from %v to %v",
			o.NodeIDs[i].String(), oldValue, o.LastReceivedData[i].Value)
		o.metrics <- o.MetricForNode(i)
	}
}

func (o *subscribeClient) handleEventNotification(notif *ua.EventNotificationList) {
	for _, event := range notif.Events {
		fields := make(map[string]interface{})
		for i, field := range event.EventFields {
			fieldName := o.OpcUAInputClient.EventStreamingInput.Fields[i]
			value := field.Value()

			if fieldName == "Message" {
				if localizedText, ok := value.(*ua.LocalizedText); ok {
					fields["Message"] = localizedText.Text
				} else {
					o.Log.Warnf("Message field is not of type *ua.LocalizedText: %T", value)
				}
				continue
			}
			var stringValue string
			switch v := value.(type) {
			case string:
				stringValue = v
			case fmt.Stringer:
				stringValue = v.String()
			case nil:
				stringValue = "null"
			default:
				stringValue = fmt.Sprintf("%v", v)
			}
			fields[fieldName] = stringValue
		}

		nodeID, ok := o.ClientHandleToNodeID.Load(event.ClientHandle)
		if !ok {
			o.Log.Warnf("NodeId not found for ClientHandle: %d", event.ClientHandle)
			nodeID = "unknown"
		}
		tags := map[string]string{
			"node_id":    nodeID.(string),
			"opcua_host": o.Config.Endpoint,
		}
		o.acc.AddFields("opcua_event_subscription", fields, tags)
	}
}
