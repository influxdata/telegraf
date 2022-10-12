//go:build linux

package conntrack

import (
	"os"
	"path"
	"strconv"
	"strings"
	"testing"

	"github.com/shirou/gopsutil/v3/net"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/plugins/inputs/system"
	"github.com/influxdata/telegraf/testutil"
)

func restoreDflts(savedFiles, savedDirs []string) {
	dfltFiles = savedFiles
	dfltDirs = savedDirs
}

func TestNoFilesFound(t *testing.T) {
	defer restoreDflts(dfltFiles, dfltDirs)

	dfltFiles = []string{"baz.txt"}
	dfltDirs = []string{"./foo/bar"}
	c := &Conntrack{}
	acc := &testutil.Accumulator{}
	err := c.Gather(acc)

	require.EqualError(t, err, "Conntrack input failed to collect metrics. "+
		"Is the conntrack kernel module loaded?")
}

func TestDefaultsUsed(t *testing.T) {
	defer restoreDflts(dfltFiles, dfltDirs)
	tmpdir, err := os.MkdirTemp("", "tmp1")
	require.NoError(t, err)
	defer os.Remove(tmpdir)

	tmpFile, err := os.CreateTemp(tmpdir, "ip_conntrack_count")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	dfltDirs = []string{tmpdir}
	fname := path.Base(tmpFile.Name())
	dfltFiles = []string{fname}

	count := 1234321
	require.NoError(t, os.WriteFile(tmpFile.Name(), []byte(strconv.Itoa(count)), 0660))
	c := &Conntrack{}
	acc := &testutil.Accumulator{}

	require.NoError(t, c.Gather(acc))
	acc.AssertContainsFields(t, inputName, map[string]interface{}{
		fname: float64(count)})
}

func TestConfigsUsed(t *testing.T) {
	defer restoreDflts(dfltFiles, dfltDirs)
	tmpdir, err := os.MkdirTemp("", "tmp1")
	require.NoError(t, err)
	defer os.Remove(tmpdir)

	cntFile, err := os.CreateTemp(tmpdir, "nf_conntrack_count")
	require.NoError(t, err)
	maxFile, err := os.CreateTemp(tmpdir, "nf_conntrack_max")
	require.NoError(t, err)
	defer os.Remove(cntFile.Name())
	defer os.Remove(maxFile.Name())

	dfltDirs = []string{tmpdir}
	cntFname := path.Base(cntFile.Name())
	maxFname := path.Base(maxFile.Name())
	dfltFiles = []string{cntFname, maxFname}

	count := 1234321
	max := 9999999
	require.NoError(t, os.WriteFile(cntFile.Name(), []byte(strconv.Itoa(count)), 0660))
	require.NoError(t, os.WriteFile(maxFile.Name(), []byte(strconv.Itoa(max)), 0660))
	c := &Conntrack{}
	acc := &testutil.Accumulator{}

	require.NoError(t, c.Gather(acc))

	fix := func(s string) string {
		return strings.Replace(s, "nf_", "ip_", 1)
	}

	acc.AssertContainsFields(t, inputName,
		map[string]interface{}{
			fix(cntFname): float64(count),
			fix(maxFname): float64(max),
		})
}

func TestCollectStats(t *testing.T) {
	var mps system.MockPS
	defer mps.AssertExpectations(t)
	var acc testutil.Accumulator

	sts := net.ConntrackStat{
		Entries:       1234,
		Searched:      10,
		Found:         1,
		New:           5,
		Invalid:       43,
		Ignore:        13,
		Delete:        3,
		DeleteList:    5,
		Insert:        9,
		InsertFailed:  20,
		Drop:          49,
		EarlyDrop:     7,
		IcmpError:     21,
		ExpectNew:     12,
		ExpectCreate:  44,
		ExpectDelete:  53,
		SearchRestart: 31,
	}

	mps.On("NetConntrack", false).Return([]net.ConntrackStat{sts}, nil)
	cs := &Conntrack{
		ps: &mps,
	}
	cs.Collect = []string{"all"}

	err := cs.Gather(&acc)
	require.NoError(t, err)

	expectedTags := map[string]string{
		"cpu": "all",
	}

	expectedFields := map[string]interface{}{
		"entries":        uint32(1234),
		"searched":       uint32(10),
		"found":          uint32(1),
		"new":            uint32(5),
		"invalid":        uint32(43),
		"ignore":         uint32(13),
		"delete":         uint32(3),
		"delete_list":    uint32(5),
		"insert":         uint32(9),
		"insert_failed":  uint32(20),
		"drop":           uint32(49),
		"early_drop":     uint32(7),
		"icmp_error":     uint32(21),
		"expect_new":     uint32(12),
		"expect_create":  uint32(44),
		"expect_delete":  uint32(53),
		"search_restart": uint32(31),
	}

	acc.AssertContainsFields(t, inputName, expectedFields)
	acc.AssertContainsTaggedFields(t, inputName, expectedFields, expectedTags)

	require.Equal(t, 19, acc.NFields())
}

func TestCollectStatsPerCpu(t *testing.T) {
	var mps system.MockPS
	defer mps.AssertExpectations(t)
	var acc testutil.Accumulator

	sts := []net.ConntrackStat{
		{
			Entries:       59,
			Searched:      10,
			Found:         1,
			New:           5,
			Invalid:       43,
			Ignore:        13,
			Delete:        3,
			DeleteList:    5,
			Insert:        9,
			InsertFailed:  20,
			Drop:          49,
			EarlyDrop:     7,
			IcmpError:     21,
			ExpectNew:     12,
			ExpectCreate:  44,
			ExpectDelete:  53,
			SearchRestart: 31,
		},
		{
			Entries:       79,
			Searched:      10,
			Found:         1,
			New:           5,
			Invalid:       43,
			Ignore:        13,
			Delete:        3,
			DeleteList:    5,
			Insert:        9,
			InsertFailed:  10,
			Drop:          49,
			EarlyDrop:     7,
			IcmpError:     21,
			ExpectNew:     12,
			ExpectCreate:  44,
			ExpectDelete:  53,
			SearchRestart: 31,
		},
	}

	mps.On("NetConntrack", true).Return(sts, nil)

	cs := &Conntrack{
		ps: &mps,
	}
	cs.Collect = []string{"all", "percpu"}

	err := cs.Gather(&acc)
	require.NoError(t, err)

	//cpu0
	expectedFields := map[string]interface{}{
		"entries":        uint32(59),
		"searched":       uint32(10),
		"found":          uint32(1),
		"new":            uint32(5),
		"invalid":        uint32(43),
		"ignore":         uint32(13),
		"delete":         uint32(3),
		"delete_list":    uint32(5),
		"insert":         uint32(9),
		"insert_failed":  uint32(20),
		"drop":           uint32(49),
		"early_drop":     uint32(7),
		"icmp_error":     uint32(21),
		"expect_new":     uint32(12),
		"expect_create":  uint32(44),
		"expect_delete":  uint32(53),
		"search_restart": uint32(31),
	}

	acc.AssertContainsFields(t, inputName, expectedFields)
	acc.AssertContainsTaggedFields(t, inputName, expectedFields,
		map[string]string{
			"cpu": "cpu0",
		})

	//TODO: check cpu1 fields

	require.Equal(t, 36, acc.NFields())
}
