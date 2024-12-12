package opcua_event_subscription

import (
	"context"
	"errors"
	"fmt"

	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/id"
	"github.com/gopcua/opcua/ua"
	"github.com/influxdata/telegraf"
	"sync"
	"time"
)

type SubscriptionManager struct {
	Client               *opcua.Client
	NodeIDs              []NodeIDWrapper
	Fields               []string
	EventType            NodeIDWrapper
	SourceNames          []string
	NotifyChannels       []chan *opcua.PublishNotificationData
	subscriptions        []*opcua.Subscription
	Log                  telegraf.Logger
	Interval             time.Duration
	ClientHandleToNodeID *sync.Map
}

func (sm *SubscriptionManager) CreateSubscription(ctx context.Context, notifyCh chan *opcua.PublishNotificationData) error {
	if len(sm.subscriptions) == 0 {
		if ctx == nil {
			return errors.New("context is nil")
		}
		if notifyCh == nil {
			return errors.New("notification channel is nil")
		}
		sm.NotifyChannels = append(sm.NotifyChannels, notifyCh)

		sub, err := sm.Client.Subscribe(ctx, &opcua.SubscriptionParameters{
			Interval: sm.Interval,
		}, notifyCh)
		if err != nil {
			return fmt.Errorf("failed to create subscription: %w", err)
		}
		sm.subscriptions = append(sm.subscriptions, sub)
	}
	return nil
}

func (sm *SubscriptionManager) Subscribe(ctx context.Context) error {
	filter := ua.EventFilter{
		SelectClauses: sm.createSelectClauses(),
		WhereClause:   sm.createWhereClauses(),
	}

	filterExtObj := ua.ExtensionObject{
		EncodingMask: ua.ExtensionObjectBinary,
		TypeID: &ua.ExpandedNodeID{
			NodeID: ua.NewNumericNodeID(0, id.EventFilter_Encoding_DefaultBinary),
		},
		Value: filter,
	}

	for i, nodeID := range sm.NodeIDs {
		miCreateRequest := &ua.MonitoredItemCreateRequest{
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
		sm.ClientHandleToNodeID.Store(uint32(i), nodeID.ID.String())
		res, err := sm.subscriptions[0].Monitor(ctx, ua.TimestampsToReturnBoth, miCreateRequest)
		if err != nil || res.Results[0].StatusCode != ua.StatusOK {
			sm.Log.Debug("failed to create monitored item for nodeID: %s", nodeID.ID.String())
			return fmt.Errorf("failed to create monitored item: %w", err)
		}
	}
	sm.Log.Info("Subscribed successfully")
	return nil
}

func (sm *SubscriptionManager) createSelectClauses() []*ua.SimpleAttributeOperand {
	selects := make([]*ua.SimpleAttributeOperand, len(sm.Fields))
	for i, name := range sm.Fields {
		selects[i] = &ua.SimpleAttributeOperand{
			TypeDefinitionID: ua.NewNumericNodeID(sm.EventType.ID.Namespace(), sm.EventType.ID.IntID()),
			BrowsePath:       []*ua.QualifiedName{{NamespaceIndex: 0, Name: name}},
			AttributeID:      ua.AttributeIDValue,
		}
	}
	return selects
}

func (sm *SubscriptionManager) createWhereClauses() *ua.ContentFilter {
	if len(sm.SourceNames) == 0 {
		return &ua.ContentFilter{
			Elements: make([]*ua.ContentFilterElement, 0),
		}
	}
	operands := make([]*ua.ExtensionObject, 0)
	for _, sourceName := range sm.SourceNames {
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
			TypeDefinitionID: ua.NewNumericNodeID(sm.EventType.ID.Namespace(), sm.EventType.ID.IntID()),
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

	wheres := &ua.ContentFilter{
		Elements: []*ua.ContentFilterElement{filterElement},
	}

	return wheres
}
