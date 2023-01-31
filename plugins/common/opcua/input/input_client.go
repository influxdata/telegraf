package input

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gopcua/opcua/ua"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal/choice"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/common/opcua"
)

// NodeSettings describes how to map from a OPC UA node to a Metric
type NodeSettings struct {
	FieldName      string            `toml:"name"`
	Namespace      string            `toml:"namespace"`
	IdentifierType string            `toml:"identifier_type"`
	Identifier     string            `toml:"identifier"`
	DataType       string            `toml:"data_type" deprecated:"1.17.0;option is ignored"`
	Description    string            `toml:"description" deprecated:"1.17.0;option is ignored"`
	TagsSlice      [][]string        `toml:"tags" deprecated:"1.25.0;use 'default_tags' instead"`
	DefaultTags    map[string]string `toml:"default_tags"`
}

// NodeID returns the OPC UA node id
func (tag *NodeSettings) NodeID() string {
	return "ns=" + tag.Namespace + ";" + tag.IdentifierType + "=" + tag.Identifier
}

// NodeGroupSettings describes a mapping of group of nodes to Metrics
type NodeGroupSettings struct {
	MetricName     string            `toml:"name"`            // Overrides plugin's setting
	Namespace      string            `toml:"namespace"`       // Can be overridden by node setting
	IdentifierType string            `toml:"identifier_type"` // Can be overridden by node setting
	Nodes          []NodeSettings    `toml:"nodes"`
	TagsSlice      [][]string        `toml:"tags" deprecated:"1.26.0;use default_tags"`
	DefaultTags    map[string]string `toml:"default_tags"`
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
	MetricName      string              `toml:"name"`
	Timestamp       TimestampSource     `toml:"timestamp"`
	TimestampFormat string              `toml:"timestamp_format"`
	RootNodes       []NodeSettings      `toml:"nodes"`
	Groups          []NodeGroupSettings `toml:"group"`
}

func (o *InputClientConfig) Validate() error {
	if o.MetricName == "" {
		return fmt.Errorf("metric name is empty")
	}

	err := choice.Check(string(o.Timestamp), []string{"", "gather", "server", "source"})
	if err != nil {
		return err
	}

	if o.TimestampFormat == "" {
		o.TimestampFormat = time.RFC3339Nano
	}

	return nil
}

func (o *InputClientConfig) CreateInputClient(log telegraf.Logger) (*OpcUAInputClient, error) {
	err := o.Validate()
	if err != nil {
		return nil, err
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
	}

	log.Debug("Initialising node to metric mapping")
	err = c.InitNodeMetricMapping()
	if err != nil {
		return nil, err
	}

	c.initLastReceivedValues()

	err = c.initNodeIDs()
	return c, err
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
	for n, t := range groupTags {
		mergedTags[n] = t
	}

	nodeTags := make(map[string]string)
	if len(node.DefaultTags) > 0 {
		nodeTags = node.DefaultTags
	} else if len(node.TagsSlice) > 0 {
		// fixme: once the TagsSlice has been removed (after deprecation), remove this if else logic
		var err error
		nodeTags, err = tagsSliceToMap(node.TagsSlice)
		if err != nil {
			return nil, err
		}
	}

	for n, t := range nodeTags {
		mergedTags[n] = t
	}

	return &NodeMetricMapping{
		Tag:        node,
		idStr:      node.NodeID(),
		metricName: metricName,
		MetricTags: mergedTags,
	}, nil
}

// NodeValue The received value for a node
type NodeValue struct {
	TagName    string
	Value      interface{}
	Quality    ua.StatusCode
	ServerTime time.Time
	SourceTime time.Time
	DataType   ua.TypeID
}

// OpcUAInputClient can receive data from an OPC UA server and map it to Metrics. This type does not contain
// logic for actually retrieving data from the server, but is used by other types like ReadClient and
// OpcUAInputSubscribeClient to store data needed to convert node ids to the corresponding metrics.
type OpcUAInputClient struct {
	*opcua.OpcUAClient
	Config InputClientConfig
	Log    telegraf.Logger

	NodeMetricMapping []NodeMetricMapping
	NodeIDs           []*ua.NodeID
	LastReceivedData  []NodeValue
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
			// Writes to a string-builder will always succeed
			//nolint:errcheck,revive
			sb.WriteString(", ")
		}
		// Writes to a string-builder will always succeed
		//nolint:errcheck,revive
		sb.WriteString(key)
		// Writes to a string-builder will always succeed
		//nolint:errcheck,revive
		sb.WriteString("=")
		// Writes to a string-builder will always succeed
		//nolint:errcheck,revive
		sb.WriteString(n.MetricTags[key])
	}
	x := metricParts{
		metricName: n.metricName,
		fieldName:  n.Tag.FieldName,
		tags:       sb.String(),
	}
	return x
}

// fixme: once the TagsSlice has been removed (after deprecation), remove this
// tagsSliceToMap takes an array of pairs of strings and creates a map from it
func tagsSliceToMap(tags [][]string) (map[string]string, error) {
	m := make(map[string]string)
	for i, tag := range tags {
		if len(tag) != 2 {
			return nil, fmt.Errorf("tag %d needs 2 values, has %d: %v", i+1, len(tag), tag)
		}
		if tag[0] == "" {
			return nil, fmt.Errorf("tag %d has empty name", i+1)
		}
		if tag[1] == "" {
			return nil, fmt.Errorf("tag %d has empty value", i+1)
		}
		if _, ok := m[tag[0]]; ok {
			return nil, fmt.Errorf("tag %d has duplicate key: %v", i+1, tag[0])
		}
		m[tag[0]] = tag[1]
	}
	return m, nil
}

func validateNodeToAdd(existing map[metricParts]struct{}, nmm *NodeMetricMapping) error {
	if nmm.Tag.FieldName == "" {
		return fmt.Errorf("empty name in '%s'", nmm.Tag.FieldName)
	}

	if len(nmm.Tag.Namespace) == 0 {
		return fmt.Errorf("empty node namespace not allowed")
	}

	if len(nmm.Tag.Identifier) == 0 {
		return fmt.Errorf("empty node identifier not allowed")
	}

	mp := newMP(nmm)
	if _, exists := existing[mp]; exists {
		return fmt.Errorf("name '%s' is duplicated (metric name '%s', tags '%s')",
			mp.fieldName, mp.metricName, mp.tags)
	}

	switch nmm.Tag.IdentifierType {
	case "i":
		if _, err := strconv.Atoi(nmm.Tag.Identifier); err != nil {
			return fmt.Errorf("identifier type '%s' does not match the type of identifier '%s'", nmm.Tag.IdentifierType, nmm.Tag.Identifier)
		}
	case "s", "g", "b":
		// Valid identifier type - do nothing.
	default:
		return fmt.Errorf("invalid identifier type '%s' in '%s'", nmm.Tag.IdentifierType, nmm.Tag.FieldName)
	}

	existing[mp] = struct{}{}
	return nil
}

// InitNodeMetricMapping builds nodes from the configuration
func (o *OpcUAInputClient) InitNodeMetricMapping() error {
	existing := map[metricParts]struct{}{}
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

		if len(group.DefaultTags) > 0 && len(group.TagsSlice) > 0 {
			o.Log.Warn("Tags found in both `tags` and `default_tags`, only using tags defined in `default_tags`")
		}

		groupTags := make(map[string]string)
		if len(group.DefaultTags) > 0 {
			groupTags = group.DefaultTags
		} else if len(group.TagsSlice) > 0 {
			// fixme: once the TagsSlice has been removed (after deprecation), remove this if else logic
			var err error
			groupTags, err = tagsSliceToMap(group.TagsSlice)
			if err != nil {
				return err
			}
		}

		for _, node := range group.Nodes {
			if node.Namespace == "" {
				node.Namespace = group.Namespace
			}
			if node.IdentifierType == "" {
				node.IdentifierType = group.IdentifierType
			}

			nmm, err := NewNodeMetricMapping(group.MetricName, node, groupTags)
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

func (o *OpcUAInputClient) initNodeIDs() error {
	o.NodeIDs = make([]*ua.NodeID, len(o.NodeMetricMapping))
	for i, node := range o.NodeMetricMapping {
		nid, err := ua.ParseNodeID(node.Tag.NodeID())
		if err != nil {
			return err
		}
		o.NodeIDs[i] = nid
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
		o.Log.Errorf("status not OK for node %v: %v", o.NodeMetricMapping[nodeIdx].Tag.FieldName, d.Status)
		return
	}

	if d.Value != nil {
		o.LastReceivedData[nodeIdx].DataType = d.Value.Type()

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
	fields := make(map[string]interface{})
	tags := map[string]string{
		"id": nmm.idStr,
	}
	for k, v := range nmm.MetricTags {
		tags[k] = v
	}

	fields[nmm.Tag.FieldName] = o.LastReceivedData[nodeIdx].Value
	fields["Quality"] = strings.TrimSpace(fmt.Sprint(o.LastReceivedData[nodeIdx].Quality))
	if !o.StatusCodeOK(o.LastReceivedData[nodeIdx].Quality) {
		mp := newMP(nmm)
		o.Log.Debugf("status not OK for node '%s'(metric name '%s', tags '%s')",
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
