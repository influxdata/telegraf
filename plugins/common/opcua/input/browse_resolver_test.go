package input

import (
	"testing"

	"github.com/gopcua/opcua/ua"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/plugins/common/opcua"
)

func TestResolveBrowsedNodesDispatchesByPattern(t *testing.T) {
	nodes := []*opcua.BrowsedNode{
		browsedVariable(t, "ns=2;s=Plant1.Device1.MV01", "MV01", []string{"Plant1", "Device1", "MV01"}),
		browsedVariable(t, "ns=2;s=Plant1.Device2.MV02", "MV02", []string{"Plant1", "Device2", "MV02"}),
		browsedVariable(t, "ns=2;s=Plant1.Device1.Temperature", "Temperature", []string{"Plant1", "Device1", "Temperature"}),
	}
	paths := []BrowsePathSettings{
		{Pattern: "Plant1/*/MV*", MetricName: "valves"},
		{Pattern: "Plant1/**/Temperature", MetricName: "temps"},
	}

	groups, err := ResolveBrowsedNodes(nodes, paths)
	require.NoError(t, err)
	require.Len(t, groups, 2)

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
			NodeID:       mustParseNodeID(t, "ns=2;s=Plant1"),
			BrowseName:   "Plant1",
			NodeClass:    ua.NodeClassObject,
			PathSegments: []string{"Plant1"},
		},
		browsedVariable(t, "ns=2;s=Plant1.MV01", "MV01", []string{"Plant1", "MV01"}),
	}
	paths := []BrowsePathSettings{
		{Pattern: "Plant1/**", MetricName: "all"},
	}

	groups, err := ResolveBrowsedNodes(nodes, paths)
	require.NoError(t, err)
	require.Len(t, groups, 1)
	require.Len(t, groups[0].Nodes, 1)
	require.Equal(t, "MV01", groups[0].Nodes[0].FieldName)
}

func TestResolveBrowsedNodesAppliesDefaultTags(t *testing.T) {
	nodes := []*opcua.BrowsedNode{
		browsedVariable(t, "ns=2;s=Plant1.MV01", "MV01", []string{"Plant1", "MV01"}),
	}
	paths := []BrowsePathSettings{
		{
			Pattern:     "Plant1/MV*",
			MetricName:  "valves",
			DefaultTags: map[string]string{"plant": "plant1"},
		},
	}

	groups, err := ResolveBrowsedNodes(nodes, paths)
	require.NoError(t, err)
	require.Equal(t, map[string]string{"plant": "plant1"}, groups[0].DefaultTags)
}

func TestResolveBrowsedNodesEmptyResultPerPath(t *testing.T) {
	nodes := []*opcua.BrowsedNode{
		browsedVariable(t, "ns=2;s=Plant1.MV01", "MV01", []string{"Plant1", "MV01"}),
	}
	paths := []BrowsePathSettings{
		{Pattern: "Plant2/**", MetricName: "other"},
	}

	groups, err := ResolveBrowsedNodes(nodes, paths)
	require.NoError(t, err)
	require.Len(t, groups, 1)
	require.Empty(t, groups[0].Nodes)
}

func TestResolveBrowsedNodesNodeMatchesMultiplePatterns(t *testing.T) {
	nodes := []*opcua.BrowsedNode{
		browsedVariable(t, "ns=2;s=Plant1.MV01", "MV01", []string{"Plant1", "MV01"}),
	}
	paths := []BrowsePathSettings{
		{Pattern: "Plant1/MV*", MetricName: "by_prefix"},
		{Pattern: "**", MetricName: "everything"},
	}

	groups, err := ResolveBrowsedNodes(nodes, paths)
	require.NoError(t, err)
	require.Len(t, groups[0].Nodes, 1)
	require.Len(t, groups[1].Nodes, 1)
}

func TestResolveBrowsedNodesSkipsEmptyBrowseName(t *testing.T) {
	nodes := []*opcua.BrowsedNode{
		browsedVariable(t, "ns=2;s=Plant1.MV01", "MV01", []string{"Plant1", "MV01"}),
		{
			NodeID:       mustParseNodeID(t, "ns=2;s=Plant1.unnamed"),
			BrowseName:   "",
			NodeClass:    ua.NodeClassVariable,
			PathSegments: []string{"Plant1", ""},
		},
	}
	paths := []BrowsePathSettings{
		{Pattern: "Plant1/**", MetricName: "all"},
	}

	groups, err := ResolveBrowsedNodes(nodes, paths)
	require.NoError(t, err)
	require.Len(t, groups[0].Nodes, 1)
	require.Equal(t, "MV01", groups[0].Nodes[0].FieldName)
}

func TestResolveBrowsedNodesPatternCompileError(t *testing.T) {
	paths := []BrowsePathSettings{
		{Pattern: "Plant1/[bad", MetricName: "broken"},
	}

	_, err := ResolveBrowsedNodes(nil, paths)
	require.ErrorContains(t, err, "compiling browse pattern")
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

func browsedVariable(t *testing.T, nodeID, browseName string, segments []string) *opcua.BrowsedNode {
	t.Helper()
	return &opcua.BrowsedNode{
		NodeID:       mustParseNodeID(t, nodeID),
		BrowseName:   browseName,
		NodeClass:    ua.NodeClassVariable,
		PathSegments: segments,
	}
}

func mustParseNodeID(t *testing.T, s string) *ua.NodeID {
	t.Helper()
	nid, err := ua.ParseNodeID(s)
	require.NoError(t, err)
	return nid
}
