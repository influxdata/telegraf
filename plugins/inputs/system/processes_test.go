package system

import (
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcesses(t *testing.T) {
	processes := &Processes{}
	var acc testutil.Accumulator

	err := processes.Gather(&acc)
	require.NoError(t, err)

	assert.True(t, acc.HasUIntField("processes", "running"))
	assert.True(t, acc.HasUIntField("processes", "sleeping"))
	assert.True(t, acc.HasUIntField("processes", "stopped"))
}
