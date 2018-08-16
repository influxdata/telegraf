package testutil

import (
	"testing"

	"github.com/influxdata/telegraf"
	"github.com/stretchr/testify/require"
)

// MustEqual requires a and b to be identical.
func MustEqual(t *testing.T, got telegraf.Metric, want Metric) {
	require.Equal(t, want.Measurement, got.Name())
	require.Equal(t, want.Fields, got.Fields())
	require.Equal(t, want.Tags, got.Tags())
	require.Equal(t, want.Time, got.Time())
}
