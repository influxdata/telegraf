package opcua_event_subscription

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/opcua"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type OpcuaEventSubscription struct {
	Endpoint       string          `toml:"endpoint"`
	Interval       config.Duration `toml:"interval"`
	EventType      NodeIDWrapper   `toml:"event_type"`
	NodeIDs        []NodeIDWrapper `toml:"node_ids"`
	SourceNames    []string        `toml:"source_names"`
	Fields         []string        `toml:"fields"`
	SecurityMode   string          `toml:"security_mode"`
	SecurityPolicy string          `toml:"security_policy"`
	Certificate    string          `toml:"certificate"`
	PrivateKey     string          `toml:"private_key"`

	Log                  telegraf.Logger
	ClientHandleToNodeId sync.Map

	client      *opcua.Client
	subscription *opcua.Subscription
	cancel      context.CancelFunc
}

func (o *OpcuaEventSubscription) SampleConfig() string {
	return `
        ## OPC UA Server Endpoint
        endpoint = "opc.tcp://opcua.demo-this.com:62544/Quickstarts/AlarmConditionServer"

        ## Polling interval
        interval = "10s"

        ## Event Type Filter
        event_type = "ns=0;i=2041"

        ## Node IDs to subscribe to
        node_ids = ["ns=2;s=0:East/Blue"]

        ## Source Name Filter (optional)
        source_names = ["SourceName1", "SourceName2"]

        ## Fields to be returned
        fields = ["Severity", "Message"]

        ## Security mode and policy (optional)
        security_mode = "None"
        security_policy = "None"

        ## Client certificate and key (optional)
        certificate = ""
        private_key = ""
    `
}

func (o *OpcuaEventSubscription) Start(acc telegraf.Accumulator) error {
	o.Log.Info("******************START******************")

	// Validate required fields
	if o.Endpoint == "" {
		return fmt.Errorf("missing mandatory field: endpoint")
	}
	if o.Interval <= 0 {
		return fmt.Errorf("missing or invalid mandatory field: interval")
	}
	if len(o.NodeIDs) == 0 {
		return fmt.Errorf("missing mandatory field: node_ids")
	}
	if o.EventType.ID == nil {
		return fmt.Errorf("missing mandatory field: event_type")
	}
	if len(o.Fields) == 0 {
		return fmt.Errorf("missing mandatory field: fields")
	}

	// Initialize the Telegraf OPC UA Client
	clientConfig := &opcua.ClientConfig{
		Endpoint:       o.Endpoint,
		SecurityMode:   o.SecurityMode,
		SecurityPolicy: o.SecurityPolicy,
		Certificate:    o.Certificate,
		PrivateKey:     o.PrivateKey,
	}
	o.client = opcua.NewClient(clientConfig, o.Log)

	if err := o.client.Connect(); err != nil {
		return fmt.Errorf("failed to connect to OPC UA server: %v", err)
	}

	// Create subscription
	ctx, cancel := context.WithCancel(context.Background())
	o.cancel = cancel

	sub, err := o.client.CreateSubscription(ctx, time.Duration(o.Interval))
	if err != nil {
		return fmt.Errorf("failed to create subscription: %v", err)
	}
	o.subscription = sub

	// Subscribe to Node IDs
	for _, nodeID := range o.NodeIDs {
		if err := sub.MonitorEvent(nodeID.String(), o.EventType.String(), func(event *opcua.EventNotification) {
			o.handleEvent(event, acc)
		}); err != nil {
			return fmt.Errorf("failed to monitor event for node ID %s: %v", nodeID, err)
		}
	}

	return nil
}

func (o *OpcuaEventSubscription) handleEvent(event *opcua.EventNotification, acc telegraf.Accumulator) {
	fields := map[string]interface{}{}
	tags := map[string]string{"endpoint": o.Endpoint}

	for _, field := range o.Fields {
		if value, ok := event.Fields[field]; ok {
			fields[field] = value
		}
	}

	acc.AddFields("opcua_event", fields, tags, event.Timestamp)
}

func (o *OpcuaEventSubscription) Gather(acc telegraf.Accumulator) error {
	// No need to gather manually, subscription handles data push
	return nil
}

func (o *OpcuaEventSubscription) Stop() {
	o.Log.Info("******************STOP******************")
	if o.cancel != nil {
		o.cancel()
	}
	if o.subscription != nil {
		o.subscription.Cancel(context.Background())
	}
	if o.client != nil {
		o.client.Close()
	}
}

func init() {
	inputs.Add("opcua_event_subscription", func() telegraf.Input {
		return &OpcuaEventSubscription{}
	})
}