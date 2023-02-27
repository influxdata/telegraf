package grafana_dashboard

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSample(t *testing.T) {
	c := &GrafanaDashboard{}
	output := c.SampleConfig()
	require.Equal(t, output, sampleConfig, "Sample config doesn't match")
}

func TestDescription(t *testing.T) {
	c := &GrafanaDashboard{}
	output := c.Description()
	require.Equal(t, output, description, "Description output is not correct")
}
