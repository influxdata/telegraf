package input

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gopcua/opcua/id"
	"github.com/gopcua/opcua/ua"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/common/opcua"
)

type Trigger string

const (
	Status               Trigger = "Status"
	StatusValue          Trigger = "StatusValue"
	StatusValueTimestamp Trigger = "StatusValueTimestamp"
)

type DeadbandType string

const (
	Absolute DeadbandType = "Absolute"
	Percent  DeadbandType = "Percent"
)

type DataChangeFilter struct {
	Trigger       Trigger      `toml:"trigger"`
	DeadbandType  DeadbandType `toml:"deadband_type"`
	DeadbandValue *float64     `toml:"deadband_value"`
}

type MonitoringParameters struct {
	SamplingInterval config.Duration   `toml:"sampling_interval"`
	QueueSize        *uint32           `toml:"queue_size"`
	DiscardOldest    *bool             `toml:"discard_oldest"`
	DataChangeFilter *DataChangeFilter `toml:"data_change_filter"`
}

// NodeSettings describes how to map from a OPC UA node to a Metric
type NodeSettings struct {
	FieldName        string               `toml:"name"`
	Namespace        string               `toml:"namespace"`
	NamespaceURI     string               `toml:"namespace_uri"`
	IdentifierType   string               `toml:"identifier_type"`
	Identifier       string               `toml:"identifier"`
	DefaultTags      map[string]string    `toml:"default_tags"`
	MonitoringParams MonitoringParameters `toml:"monitoring_params"`
}

// NodeID returns the OPC UA node id
func (tag *NodeSettings) NodeID() string {
	if tag.NamespaceURI != "" {
		return "nsu=" + tag.NamespaceURI + ";" + tag.IdentifierType + "=" + tag.Identifier
	}
	return "ns=" + tag.Namespace + ";" + tag.IdentifierType + "=" + tag.Identifier
}

// NodeGroupSettings describes a mapping of group of nodes to Metrics
type NodeGroupSettings struct {
	MetricName       string            `toml:"name"`            // Overrides plugin's setting
	Namespace        string            `toml:"namespace"`       // Can be overridden by node setting
	NamespaceURI     string            `toml:"namespace_uri"`   // Can be overridden by node setting
	IdentifierType   string            `toml:"identifier_type"` // Can be overridden by node setting
	Nodes            []NodeSettings    `toml:"nodes"`
	DefaultTags      map[string]string `toml:"default_tags"`
	SamplingInterval config.Duration   `toml:"sampling_interval"` // Can be overridden by monitoring parameters
}

type EventNodeSettings struct {
	Namespace      string `toml:"namespace"`
	NamespaceURI   string `toml:"namespace_uri"`
	IdentifierType string `toml:"identifier_type"`
	Identifier     string `toml:"identifier"`
}

func (e *EventNodeSettings) NodeID() string {
	if e.NamespaceURI != "" {
		return "nsu=" + e.NamespaceURI + ";" + e.IdentifierType + "=" + e.Identifier
	}
	return "ns=" + e.Namespace + ";" + e.IdentifierType + "=" + e.Identifier
}

type EventGroupSettings struct {
	SamplingInterval config.Duration     `toml:"sampling_interval"`
	QueueSize        uint32              `toml:"queue_size"`
	EventTypeNode    EventNodeSettings   `toml:"event_type_node"`
	Namespace        string              `toml:"namespace"`
	NamespaceURI     string              `toml:"namespace_uri"`
	IdentifierType   string              `toml:"identifier_type"`
	NodeIDSettings   []EventNodeSettings `toml:"node_ids"`
	SourceNames      []string            `toml:"source_names"`
	Fields           []string            `toml:"fields"`
}

func (e *EventGroupSettings) UpdateNodeIDSettings() {
	for i := range e.NodeIDSettings {
		n := &e.NodeIDSettings[i]
		if n.Namespace == "" {
			n.Namespace = e.Namespace
		}
		if n.NamespaceURI == "" {
			n.NamespaceURI = e.NamespaceURI
		}
		if n.IdentifierType == "" {
			n.IdentifierType = e.IdentifierType
		}
	}
}

func (e *EventGroupSettings) Validate() error {
	if err := e.EventTypeNode.validateEventNodeSettings(); err != nil {
		return fmt.Errorf("invalid event_type_node_settings: %w", err)
	}

	if len(e.NodeIDSettings) == 0 {
		return errors.New("at least one node_id must be specified")
	}

	for _, node := range e.NodeIDSettings {
		if err := node.validateEventNodeSettings(); err != nil {
			return fmt.Errorf("invalid node_id_settings: %w", err)
		}
	}

	if len(e.Fields) == 0 {
		return errors.New("at least one Field must be specified")
	}
	for _, field := range e.Fields {
		if field == "" {
			return errors.New("empty field name in fields stanza")
		}
	}
	return nil
}

func (e EventNodeSettings) validateEventNodeSettings() error {
	var defaultNodeSettings EventNodeSettings
	if e == defaultNodeSettings {
		return errors.New("node settings can't be empty")
	}
	if e.Identifier == "" {
		return errors.New("identifier must be set")
	}
	if e.IdentifierType == "" {
		return errors.New("identifier_type must be set")
	}

	// Validate namespace configuration
	hasNamespace := len(e.Namespace) > 0
	hasNamespaceURI := len(e.NamespaceURI) > 0

	if hasNamespace && hasNamespaceURI {
		return errors.New("cannot specify both 'namespace' and 'namespace_uri', use only one")
	}

	if !hasNamespace && !hasNamespaceURI {
		return errors.New("must specify either 'namespace' or 'namespace_uri'")
	}

	return nil
}

type TimestampSource string

const (
	TimestampSourceServer   TimestampSource = "server"
	TimestampSourceSource   TimestampSource = "source"
	TimestampSourceTelegraf TimestampSource = "gather"
)

// InputClientConfig a configuration for the input client
type InputClientConfig struct {
	opcua.OpcUAClientConfig
	MetricName      string               `toml:"name"`
	Timestamp       TimestampSource      `toml:"timestamp"`
	TimestampFormat string               `toml:"timestamp_format"`
	RootNodes       []NodeSettings       `toml:"nodes"`
	Groups          []NodeGroupSettings  `toml:"group"`
	EventGroups     []EventGroupSettings `toml:"events"`
}

func (o *InputClientConfig) Validate() error {
	if o.MetricName == "" {
		return errors.New("metric name is empty")
	}

	switch string(o.Timestamp) {
	case "", "gather", "server", "source":
		// Valid timestamp source
	default:
		return fmt.Errorf("unknown timestamp source %q, expected one of: [\"\", \"gather\", \"server\", \"source\"]", o.Timestamp)
	}

	if o.TimestampFormat == "" {
		o.TimestampFormat = time.RFC3339Nano
	}

	if len(o.Groups) == 0 && len(o.RootNodes) == 0 && o.EventGroups == nil {
		return errors.New("no groups, root nodes or events provided to gather from")
	}
	for _, group := range o.Groups {
		if len(group.Nodes) == 0 {
			return errors.New("group has no nodes to collect from")
		}
	}

	return nil
}

func (o *InputClientConfig) CreateInputClient(log telegraf.Logger) (*OpcUAInputClient, error) {
	if err := o.Validate(); err != nil {
		return nil, err
	}

	if o.EventGroups != nil {
		for _, eventGroup := range o.EventGroups {
			eventGroup.UpdateNodeIDSettings()
			if err := eventGroup.Validate(); err != nil {
				return nil, fmt.Errorf("invalid event_settings: %w", err)
			}
		}
	}

	log.Debug("Initialising OpcUAInputClient")
	opcClient, err := o.OpcUAClientConfig.CreateClient(log)
	if err != nil {
		return nil, err
	}

	c := &OpcUAInputClient{
		OpcUAClient: opcClient,
		Log:         log,
		Config:      *o,
		EventGroups: o.EventGroups,
	}

	log.Debug("Initialising node to metric mapping")
	if err := c.InitNodeMetricMapping(); err != nil {
		return nil, err
	}

	c.initLastReceivedValues()

	return c, nil
}

// NodeMetricMapping mapping from a single node to a metric
type NodeMetricMapping struct {
	Tag        NodeSettings
	idStr      string
	metricName string
	MetricTags map[string]string
}

// NewNodeMetricMapping builds a new NodeMetricMapping from the given argument
func NewNodeMetricMapping(metricName string, node NodeSettings, groupTags map[string]string) (*NodeMetricMapping, error) {
	mergedTags := make(map[string]string)
	maps.Copy(mergedTags, groupTags)
	maps.Copy(mergedTags, node.DefaultTags)

	return &NodeMetricMapping{
		Tag:        node,
		idStr:      node.NodeID(),
		metricName: metricName,
		MetricTags: mergedTags,
	}, nil
}

type EventNodeMetricMapping struct {
	NodeID           *ua.NodeID
	SamplingInterval *config.Duration
	QueueSize        *uint32
	EventTypeNode    *ua.NodeID
	SourceNames      []string
	Fields           []string
}

// NodeValue The received value for a node
type NodeValue struct {
	TagName    string
	Value      interface{}
	Quality    ua.StatusCode
	ServerTime time.Time
	SourceTime time.Time
	DataType   ua.TypeID
	IsArray    bool
}

// OpcUAInputClient can receive data from an OPC UA server and map it to Metrics. This type does not contain
// logic for actually retrieving data from the server, but is used by other types like ReadClient and
// OpcUAInputSubscribeClient to store data needed to convert node ids to the corresponding metrics.
type OpcUAInputClient struct {
	*opcua.OpcUAClient
	Config InputClientConfig
	Log    telegraf.Logger

	NodeMetricMapping      []NodeMetricMapping
	NodeIDs                []*ua.NodeID
	LastReceivedData       []NodeValue
	EventGroups            []EventGroupSettings
	EventNodeMetricMapping []EventNodeMetricMapping
}

// Stop the connection to the client
func (o *OpcUAInputClient) Stop(ctx context.Context) <-chan struct{} {
	ch := make(chan struct{})
	defer close(ch)
	err := o.Disconnect(ctx)
	if err != nil {
		o.Log.Warn("Disconnecting from server failed with error ", err)
	}

	return ch
}

// metricParts is only used to ensure no duplicate metrics are created
type metricParts struct {
	metricName string
	fieldName  string
	tags       string // sorted by tag name and in format tag1=value1, tag2=value2
}

func newMP(n *NodeMetricMapping) metricParts {
	keys := make([]string, 0, len(n.MetricTags))
	for key := range n.MetricTags {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var sb strings.Builder
	for i, key := range keys {
		if i != 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(key)
		sb.WriteString("=")
		sb.WriteString(n.MetricTags[key])
	}
	x := metricParts{
		metricName: n.metricName,
		fieldName:  n.Tag.FieldName,
		tags:       sb.String(),
	}
	return x
}

func validateNodeToAdd(existing map[metricParts]struct{}, nmm *NodeMetricMapping) error {
	if nmm.Tag.FieldName == "" {
		return fmt.Errorf("empty name in %q", nmm.Tag.FieldName)
	}

	// Validate namespace configuration
	hasNamespace := len(nmm.Tag.Namespace) > 0
	hasNamespaceURI := len(nmm.Tag.NamespaceURI) > 0

	if hasNamespace && hasNamespaceURI {
		return fmt.Errorf("node %q: cannot specify both 'namespace' and 'namespace_uri', use only one", nmm.Tag.FieldName)
	}

	if !hasNamespace && !hasNamespaceURI {
		return fmt.Errorf("node %q: must specify either 'namespace' or 'namespace_uri'", nmm.Tag.FieldName)
	}

	if len(nmm.Tag.Identifier) == 0 {
		return errors.New("empty node identifier not allowed")
	}

	for k, v := range nmm.MetricTags {
		if k == "" {
			return fmt.Errorf("empty tag name in tags for %q", nmm.Tag.FieldName)
		}
		if v == "" {
			return fmt.Errorf("empty tag value for tag %q in %q", k, nmm.Tag.FieldName)
		}
	}

	mp := newMP(nmm)
	if _, exists := existing[mp]; exists {
		return fmt.Errorf("name %q is duplicated (metric name %q, tags %q)",
			mp.fieldName, mp.metricName, mp.tags)
	}

	switch nmm.Tag.IdentifierType {
	case "i":
		if _, err := strconv.Atoi(nmm.Tag.Identifier); err != nil {
			return fmt.Errorf("identifier type %q does not match the type of identifier %q", nmm.Tag.IdentifierType, nmm.Tag.Identifier)
		}
	case "s", "g", "b":
		// Valid identifier type - do nothing.
	default:
		return fmt.Errorf("invalid identifier type %q in %q", nmm.Tag.IdentifierType, nmm.Tag.FieldName)
	}

	existing[mp] = struct{}{}
	return nil
}

// InitNodeMetricMapping builds nodes from the configuration
func (o *OpcUAInputClient) InitNodeMetricMapping() error {
	existing := make(map[metricParts]struct{}, len(o.Config.RootNodes))
	for _, node := range o.Config.RootNodes {
		nmm, err := NewNodeMetricMapping(o.Config.MetricName, node, make(map[string]string))
		if err != nil {
			return err
		}

		if err := validateNodeToAdd(existing, nmm); err != nil {
			return err
		}
		o.NodeMetricMapping = append(o.NodeMetricMapping, *nmm)
	}

	for _, group := range o.Config.Groups {
		if group.MetricName == "" {
			group.MetricName = o.Config.MetricName
		}

		for _, node := range group.Nodes {
			if node.Namespace == "" {
				node.Namespace = group.Namespace
			}
			if node.NamespaceURI == "" {
				node.NamespaceURI = group.NamespaceURI
			}
			if node.IdentifierType == "" {
				node.IdentifierType = group.IdentifierType
			}
			if node.MonitoringParams.SamplingInterval == 0 {
				node.MonitoringParams.SamplingInterval = group.SamplingInterval
			}

			nmm, err := NewNodeMetricMapping(group.MetricName, node, group.DefaultTags)
			if err != nil {
				return err
			}

			if err := validateNodeToAdd(existing, nmm); err != nil {
				return err
			}
			o.NodeMetricMapping = append(o.NodeMetricMapping, *nmm)
		}
	}

	return nil
}

func (o *OpcUAInputClient) InitNodeIDs() error {
	o.NodeIDs = make([]*ua.NodeID, 0, len(o.NodeMetricMapping))
	namespaceArray := o.NamespaceArray()

	for _, node := range o.NodeMetricMapping {
		nodeIDStr := node.Tag.NodeID()

		// Check if this uses namespace URI (nsu=) format
		if strings.HasPrefix(nodeIDStr, "nsu=") {
			// Namespace URI format requires namespace array
			if len(namespaceArray) == 0 {
				return fmt.Errorf("node ID %q uses namespace URI (nsu=) but namespace array is not available - connection to server may be required", nodeIDStr)
			}
			// Use ParseExpandedNodeID for namespace URI support
			expandedNodeID, err := ua.ParseExpandedNodeID(nodeIDStr, namespaceArray)
			if err != nil {
				return fmt.Errorf("failed to parse node ID %q: %w", nodeIDStr, err)
			}
			o.NodeIDs = append(o.NodeIDs, expandedNodeID.NodeID)
		} else {
			// Use ParseNodeID for namespace index (ns=) format
			nid, err := ua.ParseNodeID(nodeIDStr)
			if err != nil {
				return fmt.Errorf("failed to parse node ID %q: %w", nodeIDStr, err)
			}
			o.NodeIDs = append(o.NodeIDs, nid)
		}
	}

	return nil
}

func (o *OpcUAInputClient) InitEventNodeIDs() error {
	namespaceArray := o.NamespaceArray()

	for _, eventSetting := range o.EventGroups {
		eventTypeNodeIDStr := eventSetting.EventTypeNode.NodeID()
		var eid *ua.NodeID

		// Parse event type node ID
		if strings.HasPrefix(eventTypeNodeIDStr, "nsu=") {
			if len(namespaceArray) == 0 {
				return fmt.Errorf(
					"event type node ID %q uses namespace URI (nsu=) but namespace array is not available - "+
						"connection to server may be required",
					eventTypeNodeIDStr,
				)
			}
			expandedNodeID, err := ua.ParseExpandedNodeID(eventTypeNodeIDStr, namespaceArray)
			if err != nil {
				return fmt.Errorf("failed to parse event type node ID %q: %w", eventTypeNodeIDStr, err)
			}
			eid = expandedNodeID.NodeID
		} else {
			parsedID, err := ua.ParseNodeID(eventTypeNodeIDStr)
			if err != nil {
				return fmt.Errorf("failed to parse event type node ID %q: %w", eventTypeNodeIDStr, err)
			}
			eid = parsedID
		}

		for _, node := range eventSetting.NodeIDSettings {
			nodeIDStr := node.NodeID()
			var nid *ua.NodeID

			// Parse node ID
			if strings.HasPrefix(nodeIDStr, "nsu=") {
				if len(namespaceArray) == 0 {
					return fmt.Errorf(
						"event node ID %q uses namespace URI (nsu=) but namespace array is not available - "+
							"connection to server may be required",
						nodeIDStr,
					)
				}
				expandedNodeID, err := ua.ParseExpandedNodeID(nodeIDStr, namespaceArray)
				if err != nil {
					return fmt.Errorf("failed to parse node ID %q: %w", nodeIDStr, err)
				}
				nid = expandedNodeID.NodeID
			} else {
				parsedID, err := ua.ParseNodeID(nodeIDStr)
				if err != nil {
					return fmt.Errorf("failed to parse node ID %q: %w", nodeIDStr, err)
				}
				nid = parsedID
			}

			nmm := EventNodeMetricMapping{
				NodeID:           nid,
				SamplingInterval: &eventSetting.SamplingInterval,
				QueueSize:        &eventSetting.QueueSize,
				EventTypeNode:    eid,
				SourceNames:      eventSetting.SourceNames,
				Fields:           eventSetting.Fields,
			}
			o.EventNodeMetricMapping = append(o.EventNodeMetricMapping, nmm)
		}
	}

	return nil
}

func (o *OpcUAInputClient) initLastReceivedValues() {
	o.LastReceivedData = make([]NodeValue, len(o.NodeMetricMapping))
	for nodeIdx, nmm := range o.NodeMetricMapping {
		o.LastReceivedData[nodeIdx].TagName = nmm.Tag.FieldName
	}
}

func (o *OpcUAInputClient) UpdateNodeValue(nodeIdx int, d *ua.DataValue) {
	o.LastReceivedData[nodeIdx].Quality = d.Status
	if !o.StatusCodeOK(d.Status) {
		// Verify NodeIDs array has been built before trying to get item; otherwise show '?' for node id
		if len(o.NodeIDs) > nodeIdx {
			o.Log.Errorf("status not OK for node %v (%v): %v", o.NodeMetricMapping[nodeIdx].Tag.FieldName, o.NodeIDs[nodeIdx].String(), d.Status)
		} else {
			o.Log.Errorf("status not OK for node %v (%v): %v", o.NodeMetricMapping[nodeIdx].Tag.FieldName, '?', d.Status)
		}

		return
	}

	if d.Value != nil {
		o.LastReceivedData[nodeIdx].DataType = d.Value.Type()
		o.LastReceivedData[nodeIdx].IsArray = d.Value.Has(ua.VariantArrayValues)

		o.LastReceivedData[nodeIdx].Value = d.Value.Value()
		if o.LastReceivedData[nodeIdx].DataType == ua.TypeIDDateTime {
			if t, ok := d.Value.Value().(time.Time); ok {
				o.LastReceivedData[nodeIdx].Value = t.Format(o.Config.TimestampFormat)
			}
		}
	}
	o.LastReceivedData[nodeIdx].ServerTime = d.ServerTimestamp
	o.LastReceivedData[nodeIdx].SourceTime = d.SourceTimestamp
}

func (o *OpcUAInputClient) MetricForNode(nodeIdx int) telegraf.Metric {
	nmm := &o.NodeMetricMapping[nodeIdx]
	tags := map[string]string{
		"id": nmm.idStr,
	}
	for k, v := range nmm.MetricTags {
		tags[k] = v
	}

	fields := make(map[string]interface{})
	if o.LastReceivedData[nodeIdx].Value != nil {
		// Simple scalar types can be stored directly under the field name while
		// arrays (see 5.2.5) and structures (see 5.2.6) must be unpacked.
		// Note: Structures and arrays of structures are currently not supported.
		if o.LastReceivedData[nodeIdx].IsArray {
			switch typedValue := o.LastReceivedData[nodeIdx].Value.(type) {
			case []uint8:
				fields = unpack(nmm.Tag.FieldName, typedValue)
			case []uint16:
				fields = unpack(nmm.Tag.FieldName, typedValue)
			case []uint32:
				fields = unpack(nmm.Tag.FieldName, typedValue)
			case []uint64:
				fields = unpack(nmm.Tag.FieldName, typedValue)
			case []int8:
				fields = unpack(nmm.Tag.FieldName, typedValue)
			case []int16:
				fields = unpack(nmm.Tag.FieldName, typedValue)
			case []int32:
				fields = unpack(nmm.Tag.FieldName, typedValue)
			case []int64:
				fields = unpack(nmm.Tag.FieldName, typedValue)
			case []float32:
				fields = unpack(nmm.Tag.FieldName, typedValue)
			case []float64:
				fields = unpack(nmm.Tag.FieldName, typedValue)
			case []string:
				fields = unpack(nmm.Tag.FieldName, typedValue)
			case []bool:
				fields = unpack(nmm.Tag.FieldName, typedValue)
			default:
				o.Log.Errorf("could not unpack variant array of type: %T", typedValue)
			}
		} else {
			fields = map[string]interface{}{
				nmm.Tag.FieldName: o.LastReceivedData[nodeIdx].Value,
			}
		}
	}

	fields["Quality"] = strings.TrimSpace(o.LastReceivedData[nodeIdx].Quality.Error())
	for _, field := range o.Config.OptionalFields {
		if field == "DataType" {
			fields["DataType"] = strings.Replace(o.LastReceivedData[nodeIdx].DataType.String(), "TypeID", "", 1)
			break
		}
	}
	if !o.StatusCodeOK(o.LastReceivedData[nodeIdx].Quality) {
		mp := newMP(nmm)
		o.Log.Debugf("status not OK for node %q(metric name %q, tags %q)",
			mp.fieldName, mp.metricName, mp.tags)
	}

	var t time.Time
	switch o.Config.Timestamp {
	case TimestampSourceServer:
		t = o.LastReceivedData[nodeIdx].ServerTime
	case TimestampSourceSource:
		t = o.LastReceivedData[nodeIdx].SourceTime
	default:
		t = time.Now()
	}

	return metric.New(nmm.metricName, tags, fields, t)
}

func unpack[Slice ~[]E, E any](prefix string, value Slice) map[string]interface{} {
	fields := make(map[string]interface{}, len(value))
	for i, v := range value {
		key := fmt.Sprintf("%s[%d]", prefix, i)
		fields[key] = v
	}
	return fields
}

func (o *OpcUAInputClient) MetricForEvent(nodeIdx int, event *ua.EventFieldList) telegraf.Metric {
	node := o.EventNodeMetricMapping[nodeIdx]
	fields := make(map[string]interface{}, len(event.EventFields))
	for i, field := range event.EventFields {
		name := node.Fields[i]
		value := field.Value()

		if value == nil {
			o.Log.Warnf("Field %s has no value", name)
			continue
		}

		switch v := value.(type) {
		case *ua.LocalizedText:
			fields[name] = v.Text
		case time.Time:
			fields[name] = v.Format(time.RFC3339)
		default:
			fields[name] = v
		}
	}
	tags := map[string]string{
		"node_id": node.NodeID.String(),
		"source":  o.Config.Endpoint,
	}
	var t time.Time
	switch o.Config.Timestamp {
	case TimestampSourceServer:
		t = o.LastReceivedData[nodeIdx].ServerTime
	case TimestampSourceSource:
		t = o.LastReceivedData[nodeIdx].SourceTime
	default:
		t = time.Now()
	}

	return metric.New("opcua_event", tags, fields, t)
}

// CreateEventFilter creates a new event filter for event streaming
func (node *EventNodeMetricMapping) CreateEventFilter() (*ua.ExtensionObject, error) {
	selects, err := node.createSelectClauses()
	if err != nil {
		return nil, err
	}
	wheres, err := node.createWhereClauses()
	if err != nil {
		return nil, err
	}
	return &ua.ExtensionObject{
		EncodingMask: ua.ExtensionObjectBinary,
		TypeID:       &ua.ExpandedNodeID{NodeID: ua.NewNumericNodeID(0, id.EventFilter_Encoding_DefaultBinary)},
		Value: ua.EventFilter{
			SelectClauses: selects,
			WhereClause:   wheres,
		},
	}, nil
}

func (node *EventNodeMetricMapping) createSelectClauses() ([]*ua.SimpleAttributeOperand, error) {
	selects := make([]*ua.SimpleAttributeOperand, len(node.Fields))
	typeDefinition, err := node.determineNodeIDType()
	if err != nil {
		return nil, err
	}
	for i, name := range node.Fields {
		selects[i] = &ua.SimpleAttributeOperand{
			TypeDefinitionID: typeDefinition,
			BrowsePath:       []*ua.QualifiedName{{NamespaceIndex: 0, Name: name}},
			AttributeID:      ua.AttributeIDValue,
		}
	}
	return selects, nil
}

func (node *EventNodeMetricMapping) createWhereClauses() (*ua.ContentFilter, error) {
	if len(node.SourceNames) == 0 {
		return &ua.ContentFilter{
			Elements: make([]*ua.ContentFilterElement, 0),
		}, nil
	}
	operands := make([]*ua.ExtensionObject, 0)
	for _, sourceName := range node.SourceNames {
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

	typeDefinition, err := node.determineNodeIDType()
	if err != nil {
		return nil, err
	}

	attributeOperand := &ua.ExtensionObject{
		EncodingMask: ua.ExtensionObjectBinary,
		TypeID: &ua.ExpandedNodeID{
			NodeID: ua.NewNumericNodeID(0, id.SimpleAttributeOperand_Encoding_DefaultBinary),
		},
		Value: &ua.SimpleAttributeOperand{
			TypeDefinitionID: typeDefinition,
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

	return wheres, nil
}

func (node *EventNodeMetricMapping) determineNodeIDType() (*ua.NodeID, error) {
	switch node.EventTypeNode.Type() {
	case ua.NodeIDTypeGUID:
		return ua.NewGUIDNodeID(node.EventTypeNode.Namespace(), node.EventTypeNode.StringID()), nil
	case ua.NodeIDTypeString:
		return ua.NewStringNodeID(node.EventTypeNode.Namespace(), node.EventTypeNode.StringID()), nil
	case ua.NodeIDTypeByteString:
		return ua.NewByteStringNodeID(node.EventTypeNode.Namespace(), []byte(node.EventTypeNode.StringID())), nil
	case ua.NodeIDTypeTwoByte:
		nodeID := node.EventTypeNode.IntID()
		if nodeID > 255 {
			return nil, fmt.Errorf("twoByte EventType requires a value in the range 0-255, got %d", nodeID)
		}
		return ua.NewTwoByteNodeID(uint8(node.EventTypeNode.IntID())), nil
	case ua.NodeIDTypeFourByte:
		return ua.NewFourByteNodeID(uint8(node.EventTypeNode.Namespace()), uint16(node.EventTypeNode.IntID())), nil
	case ua.NodeIDTypeNumeric:
		return ua.NewNumericNodeID(node.EventTypeNode.Namespace(), node.EventTypeNode.IntID()), nil
	default:
		return nil, fmt.Errorf("unsupported NodeID type: %v", node.EventTypeNode.String())
	}
}
