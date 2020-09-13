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

const dataArray = `
<Document>
  <Extra value="1771">extra_tag</Extra>
  <Data>
    <Host_1>
      <Name>Host_1</Name>
      <Uptime>1000</Uptime>
      <Connections>
        <Total>15</Total>
        <Current>2</Current>
      </Connections>
    </Host_1>
    <Host_2>
      <Name>Host_2</Name>
      <Uptime>1240</Uptime>
      <Connections>
        <Total>33</Total>
        <Current>4</Current>
      </Connections>
    </Host_2>
  </Data>
</Document>
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
		Query:      "//VHost/*",
		DetectType: true,
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
		AttrPrefix: "_",
		DetectType: true,
		Query:      "//VHost/*",
		TagKeys:    []string{"Host_1_Name", "Host_2_Name"},
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
		"Host_1_Name":   "Host",
	})
	require.Equal(t, metrics[1].Tags(), map[string]string{
		"xml_node_name": "Host_2",
		"Host_2_Name":   "Server",
	})

	require.Equal(t, metrics[0].Fields(), map[string]interface{}{
		"Host_1_AvgCPU":   float64(13.3),
		"Host_1_FQDN":     string("host.local"),
		"Host_1_IsMaster": bool(true),
	})
	require.Equal(t, metrics[1].Fields(), map[string]interface{}{
		"Host_2_ConnectionsCurrent": int64(5),
		"Host_2_ConnectionsTotal":   int64(18),
	})
}

// Must return two metrics - one per selected top-level node
// With extra tags and fields
func TestArrayParsing(t *testing.T) {
	p := XMLParser{
		MetricName: "xml_test",
		ParseArray: true,
		TagNode:    true,
		DetectType: true,
		Query:      "//Data/*",
		TagKeys:    []string{"Name"},
	}

	metrics, err := p.Parse([]byte(dataArray))
	require.NoError(t, err)
	require.Len(t, metrics, 2)

	require.Len(t, metrics[0].Tags(), 2)
	require.Len(t, metrics[1].Tags(), 2)
	require.Len(t, metrics[0].Fields(), 3)
	require.Len(t, metrics[1].Fields(), 3)

	require.Equal(t, metrics[0].Name(), "xml_test")
	require.Equal(t, metrics[1].Name(), "xml_test")

	require.Equal(t, metrics[0].Tags(), map[string]string{
		"xml_node_name": "Host_1",
		"Name":          "Host_1",
	})
	require.Equal(t, metrics[1].Tags(), map[string]string{
		"xml_node_name": "Host_2",
		"Name":          "Host_2",
	})

	require.Equal(t, metrics[0].Fields(), map[string]interface{}{
		"Uptime":              int64(1000),
		"Connections_Total":   int64(15),
		"Connections_Current": int64(2),
	})
	require.Equal(t, metrics[1].Fields(), map[string]interface{}{
		"Uptime":              int64(1240),
		"Connections_Total":   int64(33),
		"Connections_Current": int64(4),
	})
}
