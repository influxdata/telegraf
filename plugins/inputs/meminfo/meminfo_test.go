package meminfo

import (
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestMemStats(t *testing.T) {
	var err error
	var acc testutil.Accumulator

	err = (&MemStats{Fields: make(map[string]interface{})}).Gather(&acc)
	require.NoError(t, err)

	acc.Metrics = nil
}
