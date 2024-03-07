package prometheus_http

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSample(t *testing.T) {
	c := &PrometheusHttp{}
	output := c.SampleConfig()
	require.Equal(t, output, sampleConfig, "Sample config doesn't match")
}

func TestDescription(t *testing.T) {
	c := &PrometheusHttp{}
	output := c.Description()
	require.Equal(t, output, description, "Description output is not correct")
}
