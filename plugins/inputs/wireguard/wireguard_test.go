package wireguard

import (
	"testing"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/assert"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
)

var _ telegraf.Input = &Wireguard{}

func TestGather(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	m := &Wireguard{
		Interfaces: []string{""},
	}
	var acc testutil.Accumulator
	err := m.Gather(&acc)
	require.NoError(t, err)
	assert.True(t, acc.HasMeasurement("wireguard"))
}
