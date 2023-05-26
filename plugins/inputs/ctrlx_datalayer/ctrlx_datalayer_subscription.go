package ctrlx_datalayer

import (
	"strings"
	"time"

	"github.com/influxdata/telegraf/config"
)

// A subscription can be used to watch multiple ctrlX Data Layer nodes for changes.
// Additional configuration settings can be given to tune the sampling and monitoring behaviour of the nodes.
// All nodes in a subscription share the same configuration.
// The plugin is able to create and manage multiple subscriptions.

// The allowed values of the subscription property 'QueueBehaviour'
var queueBehaviours = []string{"DiscardOldest", "DiscardNewest"}

// The allowed values of the subscription property 'ValueChange'
var valueChanges = []string{"Status", "StatusValue", "StatusValueTimestamp"}

// The default subscription settings
const (
	defaultKeepaliveInterval = config.Duration(60 * time.Second)
	defaultErrorInterval     = config.Duration(10 * time.Second)
	defaultReconnectInterval = config.Duration(10 * time.Second)
	defaultPublishInterval   = config.Duration(1 * time.Second)
	defaultSamplingInterval  = config.Duration(1 * time.Second)
	defaultQueueSize         = 10
	defaultQueueBehaviour    = "DiscardOldest"
	defaultValueChange       = "StatusValue"
	defaultMeasurementName   = "ctrlx"
	subscriptionPath         = "/automation/api/v2/events"
)

// Node contains all properties of a node configuration
type Node struct {
	Name    string            `toml:"name"`
	Address string            `toml:"address"`
	Tags    map[string]string `toml:"tags"`
}

// Subscription contains all properties of a subscription configuration
type Subscription struct {
	index             int
	Nodes             []Node            `toml:"nodes"`
	Tags              map[string]string `toml:"tags"`
	Measurement       string            `toml:"measurement"`
	PublishInterval   config.Duration   `toml:"publish_interval"`
	KeepaliveInterval config.Duration   `toml:"keep_alive_interval"`
	ErrorInterval     config.Duration   `toml:"error_interval"`
	SamplingInterval  config.Duration   `toml:"sampling_interval"`
	QueueSize         uint              `toml:"queue_size"`
	QueueBehaviour    string            `toml:"queue_behaviour"`
	DeadBandValue     float64           `toml:"dead_band_value"`
	ValueChange       string            `toml:"value_change"`
	OutputJSONString  bool              `toml:"output_json_string"`
}

// Rule can be used to override default rule settings.
type Rule struct {
	RuleType string      `json:"rule_type"`
	Rule     interface{} `json:"rule"`
}

// Sampling can be used to override default sampling settings.
type Sampling struct {
	SamplingInterval uint64 `json:"samplingInterval"`
}

// Queueing can be used to override default queuing settings.
type Queueing struct {
	QueueSize uint   `json:"queueSize"`
	Behaviour string `json:"behaviour"`
}

// DataChangeFilter can be used to override default data change filter settings.
type DataChangeFilter struct {
	DeadBandValue float64 `json:"deadBandValue"`
}

// ChangeEvents can be used to override default change events settings.
type ChangeEvents struct {
	ValueChange      string `json:"valueChange"`
	BrowselistChange bool   `json:"browselistChange"`
	MetadataChange   bool   `json:"metadataChange"`
}

// SubscriptionProperties can be used to override default subscription settings.
type SubscriptionProperties struct {
	KeepaliveInterval int64  `json:"keepaliveInterval"`
	Rules             []Rule `json:"rules"`
	ID                string `json:"id"`
	PublishInterval   int64  `json:"publishInterval"`
	ErrorInterval     int64  `json:"errorInterval"`
}

// SubscriptionRequest can be used to to create a sse subscription at the ctrlX Data Layer.
type SubscriptionRequest struct {
	Properties SubscriptionProperties `json:"properties"`
	Nodes      []string               `json:"nodes"`
}

// applyDefaultSettings applies the default settings if they are not configured in the config file.
func (s *Subscription) applyDefaultSettings() {
	if s.Measurement == "" {
		s.Measurement = defaultMeasurementName
	}
	if s.PublishInterval == 0 {
		s.PublishInterval = defaultPublishInterval
	}
	if s.KeepaliveInterval == 0 {
		s.KeepaliveInterval = defaultKeepaliveInterval
	}
	if s.ErrorInterval == 0 {
		s.ErrorInterval = defaultErrorInterval
	}
	if s.SamplingInterval == 0 {
		s.SamplingInterval = defaultSamplingInterval
	}
	if s.QueueSize == 0 {
		s.QueueSize = defaultQueueSize
	}
	if s.QueueBehaviour == "" {
		s.QueueBehaviour = defaultQueueBehaviour
	}
	if s.ValueChange == "" {
		s.ValueChange = defaultValueChange
	}
}

// createRequestBody builds the request body for the sse subscription, based on the subscription configuration.
// The request body can be send to the server to create a new subscription.
func (s *Subscription) createRequest(id string) SubscriptionRequest {
	pl := SubscriptionRequest{
		Properties: SubscriptionProperties{
			Rules: []Rule{
				{"Sampling", Sampling{uint64(time.Duration(s.SamplingInterval).Microseconds())}},
				{"Queueing", Queueing{s.QueueSize, s.QueueBehaviour}},
				{"DataChangeFilter", DataChangeFilter{s.DeadBandValue}},
				{"ChangeEvents", ChangeEvents{s.ValueChange, false, false}},
			},
			ID:                id,
			KeepaliveInterval: time.Duration(s.KeepaliveInterval).Milliseconds(),
			PublishInterval:   time.Duration(s.PublishInterval).Milliseconds(),
			ErrorInterval:     time.Duration(s.ErrorInterval).Milliseconds(),
		},
		Nodes: s.addressList(),
	}

	return pl
}

// addressList lists all configured node addresses
func (s *Subscription) addressList() []string {
	addressList := []string{}
	for _, node := range s.Nodes {
		addressList = append(addressList, node.Address)
	}
	return addressList
}

// node finds the node according the node address
func (s *Subscription) node(address string) *Node {
	for _, node := range s.Nodes {
		if address == node.Address {
			return &node
		}
	}
	return nil
}

// fieldKey determines the field key out of node name or address
func (n *Node) fieldKey() string {
	if n.Name != "" {
		// return user defined node name as field key
		return n.Name
	}

	// fallback: field key is extracted from mandatory node address
	i := strings.LastIndex(n.Address, "/")
	if i > 0 {
		// return last part of node address as field key
		return n.Address[i+1:]
	}

	// return full node address as field key
	return n.Address
}
