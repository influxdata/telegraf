package system

import (
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

var hugepages = Hugepages{
	NUMANodePath: "./testdata/node",
	MeminfoPath:  "./testdata/meminfo",
}

func TestHugepagesStats(t *testing.T) {
	acc := &testutil.Accumulator{}
	require.NoError(t, hugepages.Gather(acc))

	require.True(t, acc.HasPoint("hugepages", map[string]string{"node": "node0"}, "free", int(123)))
	require.True(t, acc.HasPoint("hugepages", map[string]string{"node": "node0"}, "nr", int(456)))

	require.True(t, acc.HasPoint("hugepages", map[string]string{"name": "meminfo"}, "HugePages_Total", int(666)))
	require.True(t, acc.HasPoint("hugepages", map[string]string{"name": "meminfo"}, "HugePages_Free", int(999)))
}
