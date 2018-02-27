package lustre2

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

// Set config file variables to point to fake directory structure instead of /proc?

const obdfilterProcContents = `snapshot_time             1438693064.430544 secs.usecs
read_bytes                203238095 samples [bytes] 4096 1048576 78026117632000
write_bytes               71893382 samples [bytes] 1 1048576 15201500833981
get_info                  1182008495 samples [reqs]
set_info_async            2 samples [reqs]
connect                   1117 samples [reqs]
reconnect                 1160 samples [reqs]
disconnect                1084 samples [reqs]
statfs                    3575885 samples [reqs]
create                    698 samples [reqs]
destroy                   3190060 samples [reqs]
setattr                   605647 samples [reqs]
punch                     805187 samples [reqs]
sync                      6608753 samples [reqs]
preprw                    275131477 samples [reqs]
commitrw                  275131477 samples [reqs]
quotactl                  229231 samples [reqs]
ping                      78020757 samples [reqs]
`

const osdldiskfsProcContents = `snapshot_time             1438693135.640551 secs.usecs
get_page                  275132812 samples [usec] 0 3147 1320420955 22041662259
cache_access              19047063027 samples [pages] 1 1 19047063027
cache_hit                 7393729777 samples [pages] 1 1 7393729777
cache_miss                11653333250 samples [pages] 1 1 11653333250
`

const obdfilterJobStatsContents = `job_stats:
- job_id:          testjob1
  snapshot_time:   1461772761
  read_bytes:      { samples:           1, unit: bytes, min:    4096, max:    4096, sum:            4096 }
  write_bytes:     { samples:          25, unit: bytes, min: 1048576, max: 1048576, sum:        26214400 }
  getattr:         { samples:           0, unit:  reqs }
  setattr:         { samples:           0, unit:  reqs }
  punch:           { samples:           1, unit:  reqs }
  sync:            { samples:           0, unit:  reqs }
  destroy:         { samples:           0, unit:  reqs }
  create:          { samples:           0, unit:  reqs }
  statfs:          { samples:           0, unit:  reqs }
  get_info:        { samples:           0, unit:  reqs }
  set_info:        { samples:           0, unit:  reqs }
  quotactl:        { samples:           0, unit:  reqs }
`

const mdtProcContents = `snapshot_time             1438693238.20113 secs.usecs
open                      1024577037 samples [reqs]
close                     873243496 samples [reqs]
mknod                     349042 samples [reqs]
link                      445 samples [reqs]
unlink                    3549417 samples [reqs]
mkdir                     705499 samples [reqs]
rmdir                     227434 samples [reqs]
rename                    629196 samples [reqs]
getattr                   1503663097 samples [reqs]
setattr                   1898364 samples [reqs]
getxattr                  6145349681 samples [reqs]
setxattr                  83969 samples [reqs]
statfs                    2916320 samples [reqs]
sync                      434081 samples [reqs]
samedir_rename            259625 samples [reqs]
crossdir_rename           369571 samples [reqs]
`

const mdtJobStatsContents = `job_stats:
- job_id:          testjob1
  snapshot_time:   1461772761
  open:            { samples:           5, unit:  reqs }
  close:           { samples:           4, unit:  reqs }
  mknod:           { samples:           6, unit:  reqs }
  link:            { samples:           8, unit:  reqs }
  unlink:          { samples:          90, unit:  reqs }
  mkdir:           { samples:         521, unit:  reqs }
  rmdir:           { samples:         520, unit:  reqs }
  rename:          { samples:           9, unit:  reqs }
  getattr:         { samples:          11, unit:  reqs }
  setattr:         { samples:           1, unit:  reqs }
  getxattr:        { samples:           3, unit:  reqs }
  setxattr:        { samples:           4, unit:  reqs }
  statfs:          { samples:        1205, unit:  reqs }
  sync:            { samples:           2, unit:  reqs }
  samedir_rename:  { samples:         705, unit:  reqs }
  crossdir_rename: { samples:         200, unit:  reqs }
`

func TestLustre2GeneratesMetrics(t *testing.T) {

	tempdir := os.TempDir() + "/telegraf/proc/fs/lustre/"
	ost_name := "OST0001"

	mdtdir := tempdir + "/mdt/"
	err := os.MkdirAll(mdtdir+"/"+ost_name, 0755)
	require.NoError(t, err)

	osddir := tempdir + "/osd-ldiskfs/"
	err = os.MkdirAll(osddir+"/"+ost_name, 0755)
	require.NoError(t, err)

	obddir := tempdir + "/obdfilter/"
	err = os.MkdirAll(obddir+"/"+ost_name, 0755)
	require.NoError(t, err)

	err = ioutil.WriteFile(mdtdir+"/"+ost_name+"/md_stats", []byte(mdtProcContents), 0644)
	require.NoError(t, err)

	err = ioutil.WriteFile(osddir+"/"+ost_name+"/stats", []byte(osdldiskfsProcContents), 0644)
	require.NoError(t, err)

	err = ioutil.WriteFile(obddir+"/"+ost_name+"/stats", []byte(obdfilterProcContents), 0644)
	require.NoError(t, err)

	// Begin by testing standard Lustre stats
	m := &Lustre2{
		Ost_procfiles: []string{obddir + "/*/stats", osddir + "/*/stats"},
		Mds_procfiles: []string{mdtdir + "/*/md_stats"},
	}

	var acc testutil.Accumulator

	err = m.Gather(&acc)
	require.NoError(t, err)

	tags := map[string]string{
		"name": ost_name,
	}

	fields := map[string]interface{}{
		"cache_access":    uint64(19047063027),
		"cache_hit":       uint64(7393729777),
		"cache_miss":      uint64(11653333250),
		"close":           uint64(873243496),
		"crossdir_rename": uint64(369571),
		"getattr":         uint64(1503663097),
		"getxattr":        uint64(6145349681),
		"link":            uint64(445),
		"mkdir":           uint64(705499),
		"mknod":           uint64(349042),
		"open":            uint64(1024577037),
		"read_bytes":      uint64(78026117632000),
		"read_calls":      uint64(203238095),
		"rename":          uint64(629196),
		"rmdir":           uint64(227434),
		"samedir_rename":  uint64(259625),
		"setattr":         uint64(1898364),
		"setxattr":        uint64(83969),
		"statfs":          uint64(2916320),
		"sync":            uint64(434081),
		"unlink":          uint64(3549417),
		"write_bytes":     uint64(15201500833981),
		"write_calls":     uint64(71893382),
	}

	acc.AssertContainsTaggedFields(t, "lustre2", fields, tags)

	err = os.RemoveAll(os.TempDir() + "/telegraf")
	require.NoError(t, err)
}

func TestLustre2GeneratesJobstatsMetrics(t *testing.T) {

	tempdir := os.TempDir() + "/telegraf/proc/fs/lustre/"
	ost_name := "OST0001"
	job_name := "testjob1"

	mdtdir := tempdir + "/mdt/"
	err := os.MkdirAll(mdtdir+"/"+ost_name, 0755)
	require.NoError(t, err)

	obddir := tempdir + "/obdfilter/"
	err = os.MkdirAll(obddir+"/"+ost_name, 0755)
	require.NoError(t, err)

	err = ioutil.WriteFile(mdtdir+"/"+ost_name+"/job_stats", []byte(mdtJobStatsContents), 0644)
	require.NoError(t, err)

	err = ioutil.WriteFile(obddir+"/"+ost_name+"/job_stats", []byte(obdfilterJobStatsContents), 0644)
	require.NoError(t, err)

	// Test Lustre Jobstats
	m := &Lustre2{
		Ost_procfiles: []string{obddir + "/*/job_stats"},
		Mds_procfiles: []string{mdtdir + "/*/job_stats"},
	}

	var acc testutil.Accumulator

	err = m.Gather(&acc)
	require.NoError(t, err)

	tags := map[string]string{
		"name":  ost_name,
		"jobid": job_name,
	}

	fields := map[string]interface{}{
		"jobstats_read_calls":      uint64(1),
		"jobstats_read_min_size":   uint64(4096),
		"jobstats_read_max_size":   uint64(4096),
		"jobstats_read_bytes":      uint64(4096),
		"jobstats_write_calls":     uint64(25),
		"jobstats_write_min_size":  uint64(1048576),
		"jobstats_write_max_size":  uint64(1048576),
		"jobstats_write_bytes":     uint64(26214400),
		"jobstats_ost_getattr":     uint64(0),
		"jobstats_ost_setattr":     uint64(0),
		"jobstats_punch":           uint64(1),
		"jobstats_ost_sync":        uint64(0),
		"jobstats_destroy":         uint64(0),
		"jobstats_create":          uint64(0),
		"jobstats_ost_statfs":      uint64(0),
		"jobstats_get_info":        uint64(0),
		"jobstats_set_info":        uint64(0),
		"jobstats_quotactl":        uint64(0),
		"jobstats_open":            uint64(5),
		"jobstats_close":           uint64(4),
		"jobstats_mknod":           uint64(6),
		"jobstats_link":            uint64(8),
		"jobstats_unlink":          uint64(90),
		"jobstats_mkdir":           uint64(521),
		"jobstats_rmdir":           uint64(520),
		"jobstats_rename":          uint64(9),
		"jobstats_getattr":         uint64(11),
		"jobstats_setattr":         uint64(1),
		"jobstats_getxattr":        uint64(3),
		"jobstats_setxattr":        uint64(4),
		"jobstats_statfs":          uint64(1205),
		"jobstats_sync":            uint64(2),
		"jobstats_samedir_rename":  uint64(705),
		"jobstats_crossdir_rename": uint64(200),
	}

	acc.AssertContainsTaggedFields(t, "lustre2", fields, tags)

	err = os.RemoveAll(os.TempDir() + "/telegraf")
	require.NoError(t, err)
}
