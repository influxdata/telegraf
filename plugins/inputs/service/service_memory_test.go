package service

import (
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestGather(t *testing.T) {
	var acc testutil.Accumulator

	sm := MemoryStats{ps: &servicePs{}}
	sm.ProcessNames = []string{"coreaudiod"}

	err := sm.Gather(&acc)

	require.NoError(t, err)
}
