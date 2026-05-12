package input

import (
	"fmt"

	"github.com/gopcua/opcua/ua"

	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/plugins/common/opcua"
)

// ResolveBrowsedNodes converts discovered Variable nodes that match one or
// more browse-path patterns into NodeGroupSettings, one group per configured
// path. Patterns are compiled with filter.Compile using "/" as the segment
// separator, so they support *, **, ?, [abc], {a,b,c}, and \ escapes.
//
// A node may match multiple patterns and appear in multiple groups; downstream
// duplicate detection enforces uniqueness on (metric_name, field_name, tags).
// Non-Variable nodes and nodes with an empty BrowseName are skipped.
func ResolveBrowsedNodes(nodes []*opcua.BrowsedNode, paths []BrowsePathSettings) ([]NodeGroupSettings, error) {
	groups := make([]NodeGroupSettings, len(paths))
	filters := make([]filter.Filter, len(paths))
	for i, p := range paths {
		f, err := filter.Compile([]string{p.Pattern}, '/')
		if err != nil {
			return nil, fmt.Errorf("compiling browse pattern at index %d: %w", i, err)
		}
		filters[i] = f
		groups[i] = NodeGroupSettings{
			MetricName:  p.MetricName,
			DefaultTags: p.DefaultTags,
		}
	}

	for _, node := range nodes {
		// Skip nodes the resolver cannot turn into a valid NodeSettings.
		// validateNodeToAdd downstream rejects empty FieldName and would
		// otherwise abort the whole mapping for one malformed server entry.
		if node.NodeClass != ua.NodeClassVariable || node.BrowseName == "" {
			continue
		}
		for i, f := range filters {
			if f.Match(node.Path) {
				groups[i].Nodes = append(groups[i].Nodes, NodeSettings{
					FieldName: node.BrowseName,
					NodeIDStr: node.NodeID.String(),
				})
			}
		}
	}

	return groups, nil
}
