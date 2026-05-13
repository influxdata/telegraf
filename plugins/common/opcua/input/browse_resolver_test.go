package input

import (
	"testing"

	"github.com/gopcua/opcua/ua"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/plugins/common/opcua"
)

func TestResolveBrowsedNodesDispatchesByPattern(t *testing.T) {
	nodes := []*opcua.BrowsedNode{
		browsedVariable(t, "ns=2;s=Plant1.Device1.MV01", "MV01", "Plant1/Device1/MV01"),
		browsedVariable(t, "ns=2;s=Plant1.Device2.MV02", "MV02", "Plant1/Device2/MV02"),
		browsedVariable(t, "ns=2;s=Plant1.Device1.Temperature", "Temperature", "Plant1/Device1/Temperature"),
	}
	paths := compileBrowsePaths(t, []BrowsePathSettings{
		{Pattern: "Plant1/*/MV*", MetricName: "valves"},
		{Pattern: "Plant1/**/Temperature", MetricName: "temps"},
	})

	groups, total := ResolveBrowsedNodes(nodes, paths)
	require.Len(t, groups, 2)
	require.Equal(t, 3, total)

	require.Equal(t, "valves", groups[0].MetricName)
	require.Len(t, groups[0].Nodes, 2)
	require.Equal(t, "MV01", groups[0].Nodes[0].FieldName)
	require.Equal(t, "ns=2;s=Plant1.Device1.MV01", groups[0].Nodes[0].NodeIDStr)
	require.Equal(t, "MV02", groups[0].Nodes[1].FieldName)

	require.Equal(t, "temps", groups[1].MetricName)
	require.Len(t, groups[1].Nodes, 1)
	require.Equal(t, "Temperature", groups[1].Nodes[0].FieldName)
}

func TestResolveBrowsedNodesSkipsNonVariables(t *testing.T) {
	nodes := []*opcua.BrowsedNode{
		{
			NodeID:     mustParseNodeID(t, "ns=2;s=Plant1"),
			BrowseName: "Plant1",
			NodeClass:  ua.NodeClassObject,
			Path:       "Plant1",
		},
		browsedVariable(t, "ns=2;s=Plant1.MV01", "MV01", "Plant1/MV01"),
	}
	paths := compileBrowsePaths(t, []BrowsePathSettings{
		{Pattern: "Plant1/**", MetricName: "all"},
	})

	groups, total := ResolveBrowsedNodes(nodes, paths)
	require.Len(t, groups, 1)
	require.Equal(t, 1, total)
	require.Len(t, groups[0].Nodes, 1)
	require.Equal(t, "MV01", groups[0].Nodes[0].FieldName)
}

func TestResolveBrowsedNodesAppliesDefaultTags(t *testing.T) {
	nodes := []*opcua.BrowsedNode{
		browsedVariable(t, "ns=2;s=Plant1.MV01", "MV01", "Plant1/MV01"),
	}
	paths := compileBrowsePaths(t, []BrowsePathSettings{
		{
			Pattern:     "Plant1/MV*",
			MetricName:  "valves",
			DefaultTags: map[string]string{"plant": "plant1"},
		},
	})

	groups, _ := ResolveBrowsedNodes(nodes, paths)
	require.Equal(t, map[string]string{"plant": "plant1"}, groups[0].DefaultTags)
}

func TestResolveBrowsedNodesOmitsEmptyGroups(t *testing.T) {
	nodes := []*opcua.BrowsedNode{
		browsedVariable(t, "ns=2;s=Plant1.MV01", "MV01", "Plant1/MV01"),
	}
	paths := compileBrowsePaths(t, []BrowsePathSettings{
		{Pattern: "Plant2/**", MetricName: "other"},
	})

	groups, total := ResolveBrowsedNodes(nodes, paths)
	require.Empty(t, groups)
	require.Zero(t, total)
}

func TestResolveBrowsedNodesNodeMatchesMultiplePatterns(t *testing.T) {
	nodes := []*opcua.BrowsedNode{
		browsedVariable(t, "ns=2;s=Plant1.MV01", "MV01", "Plant1/MV01"),
	}
	paths := compileBrowsePaths(t, []BrowsePathSettings{
		{Pattern: "Plant1/MV*", MetricName: "by_prefix"},
		{Pattern: "**", MetricName: "everything"},
	})

	groups, total := ResolveBrowsedNodes(nodes, paths)
	require.Len(t, groups, 2)
	require.Equal(t, 2, total)
	require.Len(t, groups[0].Nodes, 1)
	require.Len(t, groups[1].Nodes, 1)
}

func TestResolveBrowsedNodesSkipsEmptyBrowseName(t *testing.T) {
	nodes := []*opcua.BrowsedNode{
		browsedVariable(t, "ns=2;s=Plant1.MV01", "MV01", "Plant1/MV01"),
		{
			NodeID:     mustParseNodeID(t, "ns=2;s=Plant1.unnamed"),
			BrowseName: "",
			NodeClass:  ua.NodeClassVariable,
			Path:       "Plant1/",
		},
	}
	paths := compileBrowsePaths(t, []BrowsePathSettings{
		{Pattern: "Plant1/**", MetricName: "all"},
	})

	groups, _ := ResolveBrowsedNodes(nodes, paths)
	require.Len(t, groups[0].Nodes, 1)
	require.Equal(t, "MV01", groups[0].Nodes[0].FieldName)
}

func TestValidateBrowseConfigCompilesPatterns(t *testing.T) {
	cfg := InputClientConfig{
		MetricName: "opcua",
		Browse: BrowseConfig{
			Paths: []BrowsePathSettings{
				{Pattern: "Plant1/[bad", MetricName: "x"},
			},
		},
	}
	require.ErrorContains(t, cfg.Validate(), "invalid browse pattern")
}

func TestValidateBrowseConfigDefaultsRoot(t *testing.T) {
	cfg := InputClientConfig{
		MetricName: "opcua",
		Browse: BrowseConfig{
			Paths: []BrowsePathSettings{
				{Pattern: "Plant1/**", MetricName: "x"},
			},
		},
	}
	require.NoError(t, cfg.Validate())
	require.Equal(t, "ns=0;i=85", cfg.Browse.Root)
}

func TestValidateBrowseConfigRejectsBadRoot(t *testing.T) {
	cfg := InputClientConfig{
		MetricName: "opcua",
		Browse: BrowseConfig{
			Root: "ns=abc;i=1",
			Paths: []BrowsePathSettings{
				{Pattern: "Plant1/**", MetricName: "x"},
			},
		},
	}
	require.ErrorContains(t, cfg.Validate(), "invalid browse root")
}

func TestValidateBrowseConfigEmptyPattern(t *testing.T) {
	cfg := InputClientConfig{
		MetricName: "opcua",
		Browse: BrowseConfig{
			Paths: []BrowsePathSettings{
				{Pattern: "", MetricName: "x"},
			},
		},
	}
	require.ErrorContains(t, cfg.Validate(), "empty pattern")
}

func TestValidateBrowseAloneIsSufficient(t *testing.T) {
	cfg := InputClientConfig{
		MetricName: "opcua",
		Browse: BrowseConfig{
			Paths: []BrowsePathSettings{
				{Pattern: "**", MetricName: "all"},
			},
		},
	}
	require.NoError(t, cfg.Validate())
}

func browsedVariable(t *testing.T, nodeID, browseName, path string) *opcua.BrowsedNode {
	t.Helper()
	return &opcua.BrowsedNode{
		NodeID:     mustParseNodeID(t, nodeID),
		BrowseName: browseName,
		NodeClass:  ua.NodeClassVariable,
		Path:       path,
	}
}

func mustParseNodeID(t *testing.T, s string) *ua.NodeID {
	t.Helper()
	nid, err := ua.ParseNodeID(s)
	require.NoError(t, err)
	return nid
}

func compileBrowsePaths(t *testing.T, paths []BrowsePathSettings) []BrowsePathSettings {
	t.Helper()
	for i := range paths {
		f, err := filter.Compile([]string{paths[i].Pattern}, '/')
		require.NoError(t, err)
		paths[i].compiled = f
	}
	return paths
}
