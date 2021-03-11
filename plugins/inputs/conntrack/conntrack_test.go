// +build linux

package conntrack

import (
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
	"testing"

	"github.com/influxdata/telegraf/plugins/inputs/system"
	"github.com/influxdata/telegraf/testutil"
	"github.com/shirou/gopsutil/net"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	assert.EqualError(t, err, "Conntrack input failed to collect metrics. "+
		"Is the conntrack kernel module loaded?")
}

func TestDefaultsUsed(t *testing.T) {
	defer restoreDflts(dfltFiles, dfltDirs)
	tmpdir, err := ioutil.TempDir("", "tmp1")
	assert.NoError(t, err)
	defer os.Remove(tmpdir)

	tmpFile, err := ioutil.TempFile(tmpdir, "ip_conntrack_count")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	dfltDirs = []string{tmpdir}
	fname := path.Base(tmpFile.Name())
	dfltFiles = []string{fname}

	count := 1234321
	ioutil.WriteFile(tmpFile.Name(), []byte(strconv.Itoa(count)), 0660)
	c := &Conntrack{}
	acc := &testutil.Accumulator{}

	c.Gather(acc)
	acc.AssertContainsFields(t, inputName, map[string]interface{}{
		fname: float64(count)})
}

func TestConfigsUsed(t *testing.T) {
	defer restoreDflts(dfltFiles, dfltDirs)
	tmpdir, err := ioutil.TempDir("", "tmp1")
	assert.NoError(t, err)
	defer os.Remove(tmpdir)

	cntFile, err := ioutil.TempFile(tmpdir, "nf_conntrack_count")
	maxFile, err := ioutil.TempFile(tmpdir, "nf_conntrack_max")
	assert.NoError(t, err)
	defer os.Remove(cntFile.Name())
	defer os.Remove(maxFile.Name())

	dfltDirs = []string{tmpdir}
	cntFname := path.Base(cntFile.Name())
	maxFname := path.Base(maxFile.Name())
	dfltFiles = []string{cntFname, maxFname}

	count := 1234321
	max := 9999999
	ioutil.WriteFile(cntFile.Name(), []byte(strconv.Itoa(count)), 0660)
	ioutil.WriteFile(maxFile.Name(), []byte(strconv.Itoa(max)), 0660)
	c := &Conntrack{}
	acc := &testutil.Accumulator{}

	c.Gather(acc)

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

	c := NewConntrack(&mps)

	err := c.Gather(&acc)
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

	acc.AssertContainsTaggedFields(t, inputName, expectedFields, expectedTags)
}
