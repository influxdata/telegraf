package opcua_event_subscription

import (
	"fmt"
	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/ua"
	"github.com/influxdata/telegraf"
	"sync"
)

type NotificationHandler struct {
	Fields               []string
	Log                  telegraf.Logger
	Endpoint             string
	ClientHandleToNodeId *sync.Map
}

func (nh *NotificationHandler) HandleNotification(notification *opcua.PublishNotificationData, acc telegraf.Accumulator) {
	switch v := notification.Value.(type) {
	case *ua.EventNotificationList:
		nh.handleEventNotification(v, acc)
	default:
		nh.Log.Infof("Received unknown notification type: %T", v)
	}
}

func (nh *NotificationHandler) handleEventNotification(notification *ua.EventNotificationList, acc telegraf.Accumulator) {
	for _, event := range notification.Events {

		fields := make(map[string]interface{})
		for i, field := range event.EventFields {
			fieldName := nh.Fields[i]
			value := field.Value()

			if fieldName == "Message" {
				if localizedText, ok := value.(*ua.LocalizedText); ok {
					fields["Message"] = localizedText.Text
				} else {
					nh.Log.Warnf("Message field is not of type *ua.LocalizedText: %T", value)
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

		nodeId, ok := nh.ClientHandleToNodeId.Load(uint32(event.ClientHandle))
		if !ok {
			nh.Log.Warnf("NodeId not found for ClientHandle: %d", event.ClientHandle)
			nodeId = "unknown"
		}
		tags := map[string]string{
			"node_id":    nodeId.(string),
			"opcua_host": nh.Endpoint,
		}
		acc.AddFields("opcua_event_subscription", fields, tags)
	}
}
