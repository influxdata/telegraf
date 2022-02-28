package system

import (
	"testing"

	"github.com/stretchr/testify/require"
)

/*var hugepages = Hugepages{
	NUMANodePath: "./testdata/node",
	MeminfoPath:  "./testdata/meminfo",
}

func TestHugepagesStats(t *testing.T) {
	acc := &testutil.Accumulator{}
	require.NoError(t, hugepages.Gather(acc))

	require.True(t, acc.HasPoint("hugepages", map[string]string{"node": "node0", "hugepages_size": "2048kB"}, "free_hugepages", 123))
	require.True(t, acc.HasPoint("hugepages", map[string]string{"node": "node0", "hugepages_size": "2048kB"}, "nr_hugepages", 456))

	require.True(t, acc.HasPoint("hugepages", map[string]string{"name": "meminfo"}, "HugePages_Total", 666))
	require.True(t, acc.HasPoint("hugepages", map[string]string{"name": "meminfo"}, "HugePages_Free", 999))
}*/

func TestInit(t *testing.T) {
	t.Run("when no config is provided then all fields should be set to default values", func(t *testing.T) {
		h := Hugepages{}
		err := h.Init()

		require.NoError(t, err)
		require.True(t, h.gatherGlobal)
		require.False(t, h.gatherPerNode)
		require.True(t, h.gatherMeminfo)
		require.Equal(t, defaultGlobalHugepagePath, h.GlobalHugepagePath)
		require.Equal(t, defaultNumaNodePath, h.NUMANodePath)
		require.Equal(t, defaultMeminfoPath, h.MeminfoPath)
	})

	t.Run("when empty hugepages types is provided then plugin should fail to initialize", func(t *testing.T) {
		h := Hugepages{HugepagesTypes: []string{}}
		err := h.Init()

		require.Error(t, err)
		require.Contains(t, err.Error(), "plugin was configured with nothing to read")
	})

	t.Run("when valid hugepages types is provided then proper flags should be set", func(t *testing.T) {
		h := Hugepages{HugepagesTypes: []string{"global", "per_node", "meminfo"}}
		err := h.Init()

		require.NoError(t, err)
		require.True(t, h.gatherGlobal)
		require.True(t, h.gatherPerNode)
		require.True(t, h.gatherMeminfo)
	})

	t.Run("when valid hugepages types contains not supported value then plugin should fail to initialize", func(t *testing.T) {
		h := Hugepages{HugepagesTypes: []string{"global", "per_node", "linux_hdd", "meminfo"}}
		err := h.Init()

		require.Error(t, err)
		require.Contains(t, err.Error(), "provided hugepages type")
	})
}
