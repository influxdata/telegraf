package libvirt

import (
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestLibvirt(t *testing.T) {
	var acc testutil.Accumulator

	lv := &Libvirt{Uri: "test:///default"}

	err := lv.Gather(&acc)
	require.NoError(t, err)

	require.True(t, acc.HasMeasurement("libvirt"), "libvirt measurement exists")
	require.True(t, acc.HasTag("libvirt", "domain"), "is tagged with domain")
	require.True(t, acc.HasUIntField("libvirt", "cpu_time"), "has cpu_time field")
	require.True(t, acc.HasUIntField("libvirt", "memory"), "has memory field")
	require.True(t, acc.HasUIntField("libvirt", "max_mem"), "has max_mem field")
	require.True(t, acc.HasUIntField("libvirt", "nr_virt_cpu"), "has nr_virt_cpu field")
}
