// +build linux

package mdstat

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
)

func TestFullMdstatProcFile(t *testing.T) {
	tmpfile := makeFakeMDStatFile([]byte(mdStatFileFull))
	defer os.Remove(tmpfile)
	k := MdstatConf{
		HostProc: "tmp",
	}
	acc := testutil.Accumulator{}
	err := k.Gather(&acc)
	assert.NoError(t, err)

	fields := map[string]interface{}{
		"BlocksSynced":           int64(10620027200),
		"BlocksSyncedFinishTime": float64(101.6),
		"BlocksSyncedPct":        float64(94.3),
		"BlocksSyncedSpeed":      float64(103517),
		"BlocksTotal":            int64(11251451904),
		"DisksActive":            int64(12),
		"DisksFailed":            int64(0),
		"DisksSpare":             int64(0),
		"DisksTotal":             int64(12),
	}
	acc.AssertContainsFields(t, "mdstat", fields)
}

func TestInvalidMdStatProcFile1(t *testing.T) {
	tmpfile := makeFakeMDStatFile([]byte(mdStatFileInvalid))
	defer os.Remove(tmpfile)

	k := MdstatConf{
		HostProc: "tmp",
	}

	acc := testutil.Accumulator{}
	err := k.Gather(&acc)
	assert.Error(t, err)
}

const mdStatFileFull = `
Personalities : [raid1] [raid10] [linear] [multipath] [raid0] [raid6] [raid5] [raid4]
md2 : active raid10 sde[2] sdl[9] sdf[3] sdk[8] sdh[5] sdd[1] sdg[4] sdn[11] sdm[10] sdj[7] sdc[0] sdi[6]
      11251451904 blocks super 1.2 512K chunks 2 near-copies [12/12] [UUUUUUUUUUUU]
      [==================>..]  check = 94.3% (10620027200/11251451904) finish=101.6min speed=103517K/sec
      bitmap: 35/84 pages [140KB], 65536KB chunk

md1 : active raid1 sdb2[2] sda2[0]
      5909504 blocks super 1.2 [2/2] [UU]

md0 : active raid1 sdb1[2] sda1[0]
      244005888 blocks super 1.2 [2/2] [UU]
      bitmap: 1/2 pages [4KB], 65536KB chunk

unused devices: <none>
`

/*
const mdStatFileEmpty = `
Personalities :
unused devices: <none>
`
*/
const mdStatFileInvalid = `
Personalities :

mdf1: testman actve 

md0 : active raid1 sdb1[2] sda1[0]
      244005888 blocks super 1.2 [2/2] [UU]
      bitmap: 1/2 pages [4KB], 65536KB chunk

unused devices: <none>
`

func makeFakeMDStatFile(content []byte) string {
	tmpfile, err := ioutil.TempFile("tmp", "mdstat")
	if err != nil {
		panic(err)
	}

	if _, err := tmpfile.Write(content); err != nil {
		panic(err)
	}
	if err := tmpfile.Close(); err != nil {
		panic(err)
	}

	return tmpfile.Name()
}
