package xml

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const emptyNode = `
<VHost>
  <ConnectionsCurrent></ConnectionsCurrent>
</VHost>
`

const dataOnlyInNodes = `
<VHost>
  <ConnectionsCurrent>4</ConnectionsCurrent>
  <ConnectionsTotal>17</ConnectionsTotal>
</VHost>
`

const dataInAttrs = `
<VHost>
  <Host_1 Name="Host" AvgCPU="13.3" FQDN="host.local" IsMaster="true" />
  <Host_2 Name="Server" ConnectionsCurrent="5" ConnectionsTotal="18" />
</VHost>
`

// Must return no metrics
func TestWrongQuery(t *testing.T) {
	p := XMLParser{
		MetricName: "xml_test",
		MergeNodes: true,
		Query:      "//Node/*",
	}

	metrics, err := p.Parse([]byte(dataOnlyInNodes))
	require.NoError(t, err)
	require.Len(t, metrics, 0)
}

// Must return an empty metric
func TestEmptyNode(t *testing.T) {
	p := XMLParser{
		MetricName: "xml_test",
		MergeNodes: true,
	}

	metrics, err := p.Parse([]byte(emptyNode))
	require.NoError(t, err)
	require.Len(t, metrics, 1)

	require.Equal(t, metrics[0].Name(), "xml_test")
	require.Len(t, metrics[0].Fields(), 0)
	require.Len(t, metrics[0].Tags(), 0)
}

// Must return one metric with two fields
func TestMergeNodes(t *testing.T) {
	p := XMLParser{
		MetricName: "xml_test",
		MergeNodes: true,
	}

	metrics, err := p.Parse([]byte(dataOnlyInNodes))
	require.NoError(t, err)
	require.Len(t, metrics, 1)

	require.Equal(t, metrics[0].Name(), "xml_test")
	require.Len(t, metrics[0].Fields(), 2)
	require.Equal(t, metrics[0].Fields(), map[string]interface{}{
		"ConnectionsCurrent": int64(4),
		"ConnectionsTotal":   int64(17),
	})
}

// Must return two metrics - one per node
// With "Name" and "node_name" tags
// Field conversion is also checked
func TestMultiplueNodes(t *testing.T) {
	p := XMLParser{
		MetricName: "xml_test",
		MergeNodes: false,
		TagNode:    true,
		Query:      "//VHost/*",
		TagKeys:    []string{"Name"},
	}

	metrics, err := p.Parse([]byte(dataInAttrs))
	require.NoError(t, err)
	require.Len(t, metrics, 2)

	require.Len(t, metrics[0].Tags(), 2)
	require.Len(t, metrics[1].Tags(), 2)
	require.Len(t, metrics[0].Fields(), 3)
	require.Len(t, metrics[1].Fields(), 2)

	require.Equal(t, metrics[0].Tags(), map[string]string{
		"xml_node_name": "Host_1",
		"Name":          "Host",
	})
	require.Equal(t, metrics[1].Tags(), map[string]string{
		"xml_node_name": "Host_2",
		"Name":          "Server",
	})

	require.Equal(t, metrics[0].Fields(), map[string]interface{}{
		"AvgCPU":   float64(13.3),
		"FQDN":     string("host.local"),
		"IsMaster": bool(true),
	})
	require.Equal(t, metrics[1].Fields(), map[string]interface{}{
		"ConnectionsCurrent": int64(5),
		"ConnectionsTotal":   int64(18),
	})
}
