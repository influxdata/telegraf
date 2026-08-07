package input

import (
	"github.com/gopcua/opcua/ua"

	"github.com/influxdata/telegraf/plugins/common/opcua"
)

// ResolveBrowsedNodes converts discovered Variable nodes that match one or
// more browse-path patterns into NodeGroupSettings, one group per configured
// path with matches. Each path's compiled filter is populated in
// InputClientConfig.Validate with "/" as the segment separator, so patterns
// support *, **, ?, [abc], {a,b,c}, and \ escapes.
//
// A node may match multiple patterns and appear in multiple groups; downstream
// duplicate detection enforces uniqueness on (metric_name, field_name, tags).
// Non-Variable nodes and nodes with an empty BrowseName are skipped. Groups
// with no matching nodes are omitted from the result. The second return value
// is the total count of node/pattern matches across all groups.
func ResolveBrowsedNodes(nodes []*opcua.BrowsedNode, paths []BrowsePathSettings) ([]NodeGroupSettings, int) {
	groups := make([]NodeGroupSettings, 0, len(paths))
	total := 0
	for _, p := range paths {
		g := NodeGroupSettings{
			MetricName:  p.MetricName,
			DefaultTags: p.DefaultTags,
		}
		for _, node := range nodes {
			// Skip nodes the resolver cannot turn into a valid NodeSettings.
			// validateNodeToAdd downstream rejects empty FieldName and would
			// otherwise abort the whole mapping for one malformed server entry.
			if node.NodeClass != ua.NodeClassVariable || node.BrowseName == "" {
				continue
			}
			if p.compiled.Match(node.Path) {
				g.Nodes = append(g.Nodes, NodeSettings{
					FieldName: node.BrowseName,
					NodeIDStr: node.NodeID.String(),
				})
			}
		}
		if len(g.Nodes) > 0 {
			total += len(g.Nodes)
			groups = append(groups, g)
		}
	}
	return groups, total
}
