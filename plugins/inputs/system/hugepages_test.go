package system

import (
	"testing"

	"github.com/influxdata/telegraf/testutil"
)

var hugepages = Hugepages{
	NUMANodePath: "./testdata/node",
	MeminfoPath:  "./testdata/meminfo",
}

func init() {
	hugepages.loadPaths()
}

func TestHugepagesStatsFromMeminfo(t *testing.T) {
	acc := &testutil.Accumulator{}
	require.NoError(t, hugepages.GatherStatsFromMeminfo(acc))

	fields := map[string]interface{}{
		"HugePages_Total": int(666),
		"HugePages_Free":  int(999),
	}
	acc.AssertContainsFields(t, "hugepages", fields)
}

func TestHugepagesStatsPerNode(t *testing.T) {
	acc := &testutil.Accumulator{}
	err := hugepages.GatherStatsPerNode(acc)
	if err != nil {
		t.Fatal(err)
	}
	fields := map[string]interface{}{
		"free": int(123),
		"nr":   int(456),
	}
	acc.AssertContainsFields(t, "hugepages", fields)
}
