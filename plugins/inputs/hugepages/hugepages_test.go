//go:build linux
// +build linux

package hugepages

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/testutil"
)

func TestInit(t *testing.T) {
	t.Run("when no config is provided then all fields should be set to default values", func(t *testing.T) {
		h := Hugepages{}
		err := h.Init()

		require.NoError(t, err)
		require.True(t, h.gatherRoot)
		require.False(t, h.gatherPerNode)
		require.True(t, h.gatherMeminfo)
		require.Equal(t, rootHugepagePath, h.rootHugepagePath)
		require.Equal(t, numaNodePath, h.numaNodePath)
		require.Equal(t, meminfoPath, h.meminfoPath)
	})

	t.Run("when empty hugepages types is provided then plugin should fail to initialize", func(t *testing.T) {
		h := Hugepages{HugepagesTypes: []string{}}
		err := h.Init()

		require.Error(t, err)
		require.Contains(t, err.Error(), "plugin was configured with nothing to read")
	})

	t.Run("when valid hugepages types is provided then proper flags should be set", func(t *testing.T) {
		h := Hugepages{HugepagesTypes: []string{"root", "per_node", "meminfo"}}
		err := h.Init()

		require.NoError(t, err)
		require.True(t, h.gatherRoot)
		require.True(t, h.gatherPerNode)
		require.True(t, h.gatherMeminfo)
	})

	t.Run("when hugepages types contains not supported value then plugin should fail to initialize", func(t *testing.T) {
		h := Hugepages{HugepagesTypes: []string{"root", "per_node", "linux_hdd", "meminfo"}}
		err := h.Init()

		require.Error(t, err)
		require.Contains(t, err.Error(), "provided hugepages type")
	})
}

func TestGather(t *testing.T) {
	t.Run("when root hugepages type is enabled then gather all root metrics successfully", func(t *testing.T) {
		h := Hugepages{
			rootHugepagePath: "./testdata/valid/mm/hugepages",
			gatherRoot:       true,
		}

		acc := &testutil.Accumulator{}
		require.NoError(t, h.Gather(acc))

		require.True(t, acc.HasPoint("hugepages_root", map[string]string{"hugepages_size_kb": "2048"}, "free_hugepages", 883))
		require.True(t, acc.HasPoint("hugepages_root", map[string]string{"hugepages_size_kb": "2048"}, "resv_hugepages", 0))
		require.True(t, acc.HasPoint("hugepages_root", map[string]string{"hugepages_size_kb": "2048"}, "surplus_hugepages", 0))
		require.True(t, acc.HasPoint("hugepages_root", map[string]string{"hugepages_size_kb": "2048"}, "nr_hugepages_mempolicy", 2048))
		require.True(t, acc.HasPoint("hugepages_root", map[string]string{"hugepages_size_kb": "2048"}, "nr_hugepages", 2048))
		require.True(t, acc.HasPoint("hugepages_root", map[string]string{"hugepages_size_kb": "2048"}, "nr_overcommit_hugepages", 0))

		require.True(t, acc.HasPoint("hugepages_root", map[string]string{"hugepages_size_kb": "1048576"}, "free_hugepages", 0))
		require.True(t, acc.HasPoint("hugepages_root", map[string]string{"hugepages_size_kb": "1048576"}, "resv_hugepages", 0))
		require.True(t, acc.HasPoint("hugepages_root", map[string]string{"hugepages_size_kb": "1048576"}, "surplus_hugepages", 0))
		require.True(t, acc.HasPoint("hugepages_root", map[string]string{"hugepages_size_kb": "1048576"}, "nr_hugepages_mempolicy", 8))
		require.True(t, acc.HasPoint("hugepages_root", map[string]string{"hugepages_size_kb": "1048576"}, "nr_hugepages", 8))
		require.True(t, acc.HasPoint("hugepages_root", map[string]string{"hugepages_size_kb": "1048576"}, "nr_overcommit_hugepages", 0))
	})

	t.Run("when per node hugepages type is enabled then gather all per node metrics successfully", func(t *testing.T) {
		h := Hugepages{
			numaNodePath:  "./testdata/valid/node",
			gatherPerNode: true,
		}

		acc := &testutil.Accumulator{}
		require.NoError(t, h.Gather(acc))

		require.True(t, acc.HasPoint("hugepages_per_node", map[string]string{"hugepages_size_kb": "2048", "node": "0"}, "free_hugepages", 434))
		require.True(t, acc.HasPoint("hugepages_per_node", map[string]string{"hugepages_size_kb": "2048", "node": "0"}, "surplus_hugepages", 0))
		require.True(t, acc.HasPoint("hugepages_per_node", map[string]string{"hugepages_size_kb": "2048", "node": "0"}, "nr_hugepages", 1024))

		require.True(t, acc.HasPoint("hugepages_per_node", map[string]string{"hugepages_size_kb": "2048", "node": "1"}, "free_hugepages", 449))
		require.True(t, acc.HasPoint("hugepages_per_node", map[string]string{"hugepages_size_kb": "2048", "node": "1"}, "surplus_hugepages", 0))
		require.True(t, acc.HasPoint("hugepages_per_node", map[string]string{"hugepages_size_kb": "2048", "node": "1"}, "nr_hugepages", 1024))

		require.True(t, acc.HasPoint("hugepages_per_node", map[string]string{"hugepages_size_kb": "1048576", "node": "0"}, "free_hugepages", 0))
		require.True(t, acc.HasPoint("hugepages_per_node", map[string]string{"hugepages_size_kb": "1048576", "node": "0"}, "surplus_hugepages", 0))
		require.True(t, acc.HasPoint("hugepages_per_node", map[string]string{"hugepages_size_kb": "1048576", "node": "0"}, "nr_hugepages", 4))

		require.True(t, acc.HasPoint("hugepages_per_node", map[string]string{"hugepages_size_kb": "1048576", "node": "1"}, "free_hugepages", 0))
		require.True(t, acc.HasPoint("hugepages_per_node", map[string]string{"hugepages_size_kb": "1048576", "node": "1"}, "surplus_hugepages", 0))
		require.True(t, acc.HasPoint("hugepages_per_node", map[string]string{"hugepages_size_kb": "1048576", "node": "1"}, "nr_hugepages", 4))
	})

	t.Run("when meminfo hugepages type is enabled then gather all meminfo metrics successfully", func(t *testing.T) {
		h := Hugepages{
			meminfoPath:   "./testdata/valid/meminfo",
			gatherMeminfo: true,
		}

		acc := &testutil.Accumulator{}
		require.NoError(t, h.Gather(acc))

		require.True(t, acc.HasPoint("hugepages_meminfo", map[string]string{}, "AnonHugePages_kb", 0))
		require.True(t, acc.HasPoint("hugepages_meminfo", map[string]string{}, "ShmemHugePages_kb", 0))
		require.True(t, acc.HasPoint("hugepages_meminfo", map[string]string{}, "FileHugePages_kb", 0))
		require.True(t, acc.HasPoint("hugepages_meminfo", map[string]string{}, "HugePages_Total", 2048))
		require.True(t, acc.HasPoint("hugepages_meminfo", map[string]string{}, "HugePages_Free", 883))
		require.True(t, acc.HasPoint("hugepages_meminfo", map[string]string{}, "HugePages_Rsvd", 0))
		require.True(t, acc.HasPoint("hugepages_meminfo", map[string]string{}, "HugePages_Surp", 0))
		require.True(t, acc.HasPoint("hugepages_meminfo", map[string]string{}, "Hugepagesize_kb", 2048))
		require.True(t, acc.HasPoint("hugepages_meminfo", map[string]string{}, "Hugetlb_kb", 12582912))
	})

	t.Run("when root hugepages type is enabled but path is invalid then return error", func(t *testing.T) {
		h := Hugepages{
			rootHugepagePath: "./testdata/not_existing_path",
			gatherRoot:       true,
		}

		acc := &testutil.Accumulator{}
		require.Error(t, h.Gather(acc))
	})

	t.Run("when root hugepages type is enabled but files/directories don't have proper naming then gather no metrics", func(t *testing.T) {
		h := Hugepages{
			rootHugepagePath: "./testdata/invalid/1/node0/hugepages",
			gatherRoot:       true,
		}

		acc := &testutil.Accumulator{}
		require.NoError(t, h.Gather(acc))
		require.Nil(t, acc.Metrics)
	})

	t.Run("when root hugepages type is enabled but metric file doesn't contain number then return error", func(t *testing.T) {
		h := Hugepages{
			rootHugepagePath: "./testdata/invalid/2/node1/hugepages",
			gatherRoot:       true,
		}

		acc := &testutil.Accumulator{}
		require.Error(t, h.Gather(acc))
	})

	t.Run("when per node hugepages type is enabled but path is invalid then return error", func(t *testing.T) {
		h := Hugepages{
			numaNodePath:  "./testdata/not_existing_path",
			gatherPerNode: true,
		}

		acc := &testutil.Accumulator{}
		require.Error(t, h.Gather(acc))
	})

	t.Run("when per node hugepages type is enabled but files/directories don't have proper naming then gather no metrics", func(t *testing.T) {
		h := Hugepages{
			numaNodePath:  "./testdata/invalid/1",
			gatherPerNode: true,
		}

		acc := &testutil.Accumulator{}
		require.NoError(t, h.Gather(acc))
		require.Nil(t, acc.Metrics)
	})

	t.Run("when per node hugepages type is enabled but metric file doesn't contain number then return error", func(t *testing.T) {
		h := Hugepages{
			numaNodePath:  "./testdata/invalid/2/",
			gatherPerNode: true,
		}

		acc := &testutil.Accumulator{}
		require.Error(t, h.Gather(acc))
	})

	t.Run("when meminfo hugepages type is enabled but path is invalid then return error", func(t *testing.T) {
		h := Hugepages{
			meminfoPath:   "./testdata/not_existing_path",
			gatherMeminfo: true,
		}

		acc := &testutil.Accumulator{}
		require.Error(t, h.Gather(acc))
	})

	t.Run("when per node hugepages type is enabled but any metric doesn't contain number then return error", func(t *testing.T) {
		h := Hugepages{
			meminfoPath:   "./testdata/invalid/meminfo",
			gatherMeminfo: true,
		}

		acc := &testutil.Accumulator{}
		require.Error(t, h.Gather(acc))
	})
}
