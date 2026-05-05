package input

import (
	"fmt"

	"github.com/gopcua/opcua/ua"

	"github.com/influxdata/telegraf/plugins/common/opcua"
)

// ResolveBrowsedNodes converts discovered Variable nodes that match one or
// more browse-path patterns into NodeGroupSettings, one group per configured
// path. A node may match multiple patterns and appear in multiple groups;
// downstream duplicate detection enforces uniqueness on (metric_name,
// field_name, tags).
//
// Non-Variable nodes are skipped. Pattern compile errors are returned as
// configuration errors.
func ResolveBrowsedNodes(nodes []*opcua.BrowsedNode, paths []BrowsePathSettings) ([]NodeGroupSettings, error) {
	groups := make([]NodeGroupSettings, len(paths))
	compiled := make([]*opcua.PathPattern, len(paths))
	for i, p := range paths {
		c, err := opcua.CompilePathPattern(p.Pattern)
		if err != nil {
			return nil, fmt.Errorf("compiling browse pattern at index %d: %w", i, err)
		}
		compiled[i] = c
		groups[i] = NodeGroupSettings{
			MetricName:  p.MetricName,
			DefaultTags: p.DefaultTags,
		}
	}

	for _, node := range nodes {
		if node.NodeClass != ua.NodeClassVariable {
			continue
		}
		for i, c := range compiled {
			if c.Match(node.PathSegments) {
				groups[i].Nodes = append(groups[i].Nodes, NodeSettings{
					FieldName: node.BrowseName,
					NodeIDStr: node.NodeID.String(),
				})
			}
		}
	}

	return groups, nil
}
