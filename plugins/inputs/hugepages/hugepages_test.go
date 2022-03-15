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
		h := Hugepages{Types: []string{}}
		err := h.Init()

		require.Error(t, err)
		require.Contains(t, err.Error(), "plugin was configured with nothing to read")
	})

	t.Run("when valid hugepages types is provided then proper flags should be set", func(t *testing.T) {
		h := Hugepages{Types: []string{"root", "per_node", "meminfo"}}
		err := h.Init()

		require.NoError(t, err)
		require.True(t, h.gatherRoot)
		require.True(t, h.gatherPerNode)
		require.True(t, h.gatherMeminfo)
	})

	t.Run("when hugepages types contains not supported value then plugin should fail to initialize", func(t *testing.T) {
		h := Hugepages{Types: []string{"root", "per_node", "linux_hdd", "meminfo"}}
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

		expectedFields := map[string]interface{}{
			"free":       883,
			"reserved":   0,
			"surplus":    0,
			"mempolicy":  2048,
			"total":      2048,
			"overcommit": 0,
		}
		acc.AssertContainsTaggedFields(t, "hugepages_root", expectedFields, map[string]string{"size_kb": "2048"})

		expectedFields = map[string]interface{}{
			"free":       0,
			"reserved":   0,
			"surplus":    0,
			"mempolicy":  8,
			"total":      8,
			"overcommit": 0,
		}
		acc.AssertContainsTaggedFields(t, "hugepages_root", expectedFields, map[string]string{"size_kb": "1048576"})
	})

	t.Run("when per node hugepages type is enabled then gather all per node metrics successfully", func(t *testing.T) {
		h := Hugepages{
			numaNodePath:  "./testdata/valid/node",
			gatherPerNode: true,
		}

		acc := &testutil.Accumulator{}
		require.NoError(t, h.Gather(acc))

		expectedFields := map[string]interface{}{
			"free":    434,
			"surplus": 0,
			"total":   1024,
		}
		acc.AssertContainsTaggedFields(t, "hugepages_per_node", expectedFields, map[string]string{"size_kb": "2048", "node": "0"})

		expectedFields = map[string]interface{}{
			"free":    449,
			"surplus": 0,
			"total":   1024,
		}
		acc.AssertContainsTaggedFields(t, "hugepages_per_node", expectedFields, map[string]string{"size_kb": "2048", "node": "1"})

		expectedFields = map[string]interface{}{
			"free":    0,
			"surplus": 0,
			"total":   4,
		}
		acc.AssertContainsTaggedFields(t, "hugepages_per_node", expectedFields, map[string]string{"size_kb": "1048576", "node": "0"})

		expectedFields = map[string]interface{}{
			"free":    0,
			"surplus": 0,
			"total":   4,
		}
		acc.AssertContainsTaggedFields(t, "hugepages_per_node", expectedFields, map[string]string{"size_kb": "1048576", "node": "1"})
	})

	t.Run("when meminfo hugepages type is enabled then gather all meminfo metrics successfully", func(t *testing.T) {
		h := Hugepages{
			meminfoPath:   "./testdata/valid/meminfo",
			gatherMeminfo: true,
		}

		acc := &testutil.Accumulator{}
		require.NoError(t, h.Gather(acc))

		expectedFields := map[string]interface{}{
			"anonymous_kb": 0,
			"shared_kb":    0,
			"file_kb":      0,
			"total":        2048,
			"free":         883,
			"reserved":     0,
			"surplus":      0,
			"size_kb":      2048,
			"tlb_kb":       12582912,
		}
		acc.AssertContainsFields(t, "hugepages_meminfo", expectedFields)
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
