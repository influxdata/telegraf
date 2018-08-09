package meminfo

import (
	"testing"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestMemStats(t *testing.T) {
	var err error
	var acc testutil.Accumulator

	err = (&MemStats{Fields: make(map[string]interface{})}).Gather(&acc)
	require.NoError(t, err)

	acc.Metrics = nil
}
