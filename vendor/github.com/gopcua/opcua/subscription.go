package opcua

import (
	"context"
	"time"

	"github.com/gopcua/opcua/debug"
	"github.com/gopcua/opcua/errors"
	"github.com/gopcua/opcua/id"
	"github.com/gopcua/opcua/ua"
	"github.com/gopcua/opcua/uasc"
)

const (
	DefaultSubscriptionMaxNotificationsPerPublish = 10000
	DefaultSubscriptionLifetimeCount              = 10000
	DefaultSubscriptionMaxKeepAliveCount          = 3000
	DefaultSubscriptionInterval                   = 100 * time.Millisecond
	DefaultSubscriptionPriority                   = 0
)

type Subscription struct {
	SubscriptionID            uint32
	RevisedPublishingInterval time.Duration
	RevisedLifetimeCount      uint32
	RevisedMaxKeepAliveCount  uint32
	Notifs                    chan *PublishNotificationData
	c                         *Client
}

type SubscriptionParameters struct {
	Interval                   time.Duration
	LifetimeCount              uint32
	MaxKeepAliveCount          uint32
	MaxNotificationsPerPublish uint32
	Priority                   uint8
}

func NewMonitoredItemCreateRequestWithDefaults(nodeID *ua.NodeID, attributeID ua.AttributeID, clientHandle uint32) *ua.MonitoredItemCreateRequest {
	if attributeID == 0 {
		attributeID = ua.AttributeIDValue
	}
	return &ua.MonitoredItemCreateRequest{
		ItemToMonitor: &ua.ReadValueID{
			NodeID:       nodeID,
			AttributeID:  attributeID,
			DataEncoding: &ua.QualifiedName{},
		},
		MonitoringMode: ua.MonitoringModeReporting,
		RequestedParameters: &ua.MonitoringParameters{
			ClientHandle:     clientHandle,
			DiscardOldest:    true,
			Filter:           nil,
			QueueSize:        10,
			SamplingInterval: 0.0,
		},
	}
}

type PublishNotificationData struct {
	SubscriptionID uint32
	Error          error
	Value          interface{}
}

// Cancel() deletes the Subscription from Server and makes the Client forget it so that publishing
// loops cannot deliver notifications to it anymore
func (s *Subscription) Cancel() error {
	s.c.forgetSubscription(s.SubscriptionID)

	req := &ua.DeleteSubscriptionsRequest{
		SubscriptionIDs: []uint32{s.SubscriptionID},
	}
	var res *ua.DeleteSubscriptionsResponse
	err := s.c.Send(req, func(v interface{}) error {
		return safeAssign(v, &res)
	})
	if err != nil {
		return err
	}
	if res.ResponseHeader.ServiceResult != ua.StatusOK {
		return res.ResponseHeader.ServiceResult
	}

	return nil
}

func (s *Subscription) Monitor(ts ua.TimestampsToReturn, items ...*ua.MonitoredItemCreateRequest) (*ua.CreateMonitoredItemsResponse, error) {
	// Part 4, 5.12.2.2 CreateMonitoredItems Service Parameters
	req := &ua.CreateMonitoredItemsRequest{
		SubscriptionID:     s.SubscriptionID,
		TimestampsToReturn: ts,
		ItemsToCreate:      items,
	}

	var res *ua.CreateMonitoredItemsResponse
	err := s.c.Send(req, func(v interface{}) error {
		return safeAssign(v, &res)
	})
	return res, err
}

func (s *Subscription) Unmonitor(monitoredItemIDs ...uint32) (*ua.DeleteMonitoredItemsResponse, error) {
	req := &ua.DeleteMonitoredItemsRequest{
		MonitoredItemIDs: monitoredItemIDs,
		SubscriptionID:   s.SubscriptionID,
	}
	var res *ua.DeleteMonitoredItemsResponse
	err := s.c.Send(req, func(v interface{}) error {
		return safeAssign(v, &res)
	})
	return res, err
}

// SetTriggering sends a request to the server to add and/or remove triggering links from a triggering item.
// To add links from a triggering item to an item to report provide the server assigned ID(s) in the `add` argument.
// To remove links from a triggering item to an item to report provide the server assigned ID(s) in the `remove` argument.
func (s *Subscription) SetTriggering(triggeringItemID uint32, add, remove []uint32) (*ua.SetTriggeringResponse, error) {
	// Part 4, 5.12.5.2 SetTriggering Service Parameters
	req := &ua.SetTriggeringRequest{
		SubscriptionID:   s.SubscriptionID,
		TriggeringItemID: triggeringItemID,
		LinksToAdd:       add,
		LinksToRemove:    remove,
	}

	var res *ua.SetTriggeringResponse
	err := s.c.Send(req, func(v interface{}) error {
		return safeAssign(v, &res)
	})
	return res, err
}

func (s *Subscription) publish(acks []*ua.SubscriptionAcknowledgement) (*ua.PublishResponse, error) {
	if acks == nil {
		acks = []*ua.SubscriptionAcknowledgement{}
	}
	req := &ua.PublishRequest{
		SubscriptionAcknowledgements: acks,
	}

	var res *ua.PublishResponse
	err := s.c.sendWithTimeout(req, s.publishTimeout(), func(v interface{}) error {
		return safeAssign(v, &res)
	})
	return res, err
}

func (s *Subscription) publishTimeout() time.Duration {
	timeout := time.Duration(s.RevisedMaxKeepAliveCount) * s.RevisedPublishingInterval // expected keepalive interval
	if timeout > uasc.MaxTimeout {
		return uasc.MaxTimeout
	}
	if timeout < s.c.cfg.RequestTimeout {
		return s.c.cfg.RequestTimeout
	}
	return timeout
}

// Run() starts an infinite loop that sends PublishRequests and delivers received
// notifications to registered Subscriptions.
// It is the responsibility of the user to stop no longer needed Run() loops by cancelling ctx
// Note that Run() may return before ctx is cancelled in case of an irrecoverable communication error
func (s *Subscription) Run(ctx context.Context) {
	var acks []*ua.SubscriptionAcknowledgement

	for {
		select {
		case <-ctx.Done():
			return
		default:
			// send the next publish request
			// note that res contains data even if an error was returned
			res, err := s.publish(acks)
			switch {
			case err == ua.StatusBadSequenceNumberUnknown:
				// At least one ack has been submitted repeatedly
				// Ignore the error. Acks will be cleared below
			case err == ua.StatusBadTimeout:
				// ignore and continue the loop
			case err == ua.StatusBadNoSubscription:
				// All subscriptions have been deleted, but the publishing loop is still running
				// The user will stop the loop or create subscriptions at his discretion
			case err != nil:
				// irrecoverable error
				s.c.notifySubscriptionsOfError(ctx, res, err)
				debug.Printf("subscription %v Run loop stopped", s.SubscriptionID)
				return
			}

			if res != nil {
				// Prepare SubscriptionAcknowledgement for next PublishRequest
				acks = make([]*ua.SubscriptionAcknowledgement, 0)
				if res.AvailableSequenceNumbers != nil {
					for _, i := range res.AvailableSequenceNumbers {
						ack := &ua.SubscriptionAcknowledgement{
							SubscriptionID: res.SubscriptionID,
							SequenceNumber: i,
						}
						acks = append(acks, ack)
					}
				}
			}

			if err == nil {
				s.c.notifySubscription(ctx, res)
			}
		}
	}
}

func (s *Subscription) sendNotification(ctx context.Context, data *PublishNotificationData) {
	select {
	case <-ctx.Done():
		return
	case s.Notifs <- data:
	}

}

// Stats returns a diagnostic struct with metadata about the current subscription
func (s *Subscription) Stats() (*ua.SubscriptionDiagnosticsDataType, error) {
	// TODO(kung-foo): once browsing feature is merged, attempt to get direct access to the
	// diagnostics node. for example, Prosys lists them like:
	// i=2290/ns=1;g=918ee6f4-2d25-4506-980d-e659441c166d
	// maybe cache the nodeid to speed up future stats queries
	node := s.c.Node(ua.NewNumericNodeID(0, id.Server_ServerDiagnostics_SubscriptionDiagnosticsArray))
	v, err := node.Value()
	if err != nil {
		return nil, err
	}

	for _, eo := range v.Value().([]*ua.ExtensionObject) {
		stat := eo.Value.(*ua.SubscriptionDiagnosticsDataType)

		if stat.SubscriptionID == s.SubscriptionID {
			return stat, nil
		}
	}

	return nil, errors.Errorf("unable to find SubscriptionDiagnostics for sub=%d", s.SubscriptionID)
}

func (p *SubscriptionParameters) setDefaults() {
	if p.MaxNotificationsPerPublish == 0 {
		p.MaxNotificationsPerPublish = DefaultSubscriptionMaxNotificationsPerPublish
	}
	if p.LifetimeCount == 0 {
		p.LifetimeCount = DefaultSubscriptionLifetimeCount
	}
	if p.MaxKeepAliveCount == 0 {
		p.MaxKeepAliveCount = DefaultSubscriptionMaxKeepAliveCount
	}
	if p.Interval == 0 {
		p.Interval = DefaultSubscriptionInterval
	}
	if p.Priority == 0 {
		// DefaultSubscriptionPriority is 0 at the time of writing, so this redundant assignment is
		// made only to allow for a one-liner change of default priority should a need arise
		// and to explicitly expose the default priority as a constant
		p.Priority = DefaultSubscriptionPriority
	}
}
