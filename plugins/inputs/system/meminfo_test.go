// +build linux

package system

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/influxdata/telegraf/testutil"

	"github.com/stretchr/testify/assert"
)

func TestFullMemInfoProcfile(t *testing.T) {
	tmpfile := makeFakeStatFile([]byte(memInfoFile_Full))
	defer os.Remove(tmpfile)

	k := Meminfo{
		statFile: tmpfile,
	}

	acc := testutil.Accumulator{}
	err := k.Gather(&acc)
	assert.NoError(t, err)

	fields := map[string]interface{}{
        "MemTotal":       int64(16416164),
        "MemFree":        int64(10021148),
        "MemAvailable":   int64(11400828),
        "Buffers":        int64(72572),
        "Cached":         int64(1586784),
        "SwapCached":     int64(65372),
        "Active":          int64(3306232),
        "Inactive":        int64(2549784),
        "Active_anon":    int64(2813764),
        "Inactive_anon":  int64(1462248),
        "Active_file":     int64(492468),
        "Inactive_file":  int64(1087536),
        "Unevictable":          int64(32),
        "Mlocked":              int64(32),
        "SwapTotal":       int64(2097148),
        "SwapFree":         int64(493224),
        "Dirty":              int64(2256),
        "Writeback":             int64(0),
        "AnonPages":       int64(4135280),
        "Mapped":          int64(1195320),
        "Shmem":            int64(666084),
        "Slab":             int64(242932),
        "SReclaimable":     int64(102768),
        "SUnreclaim":       int64(140164),
        "KernelStack":       int64(25248),
        "PageTables":       int64(106276),
        "NFS_Unstable":          int64(0),
        "Bounce":                int64(0),
        "WritebackTmp":          int64(0),
        "CommitLimit":    int64(10305228),
        "Committed_AS":   int64(18852196),
        "VmallocTotal":   int64(34359738367),
        "VmallocUsed":           int64(0),
        "VmallocChunk":          int64(0),
        "HardwareCorrupted":     int64(0),
        "AnonHugePages":         int64(0),
        "ShmemHugePages":        int64(0),
        "ShmemPmdMapped":        int64(0),
        "CmaTotal":              int64(0),
        "CmaFree":               int64(0),
        "HugePages_Total":       int64(0),
        "HugePages_Free":        int64(0),
        "HugePages_Rsvd":        int64(0),
        "HugePages_Surp":        int64(0),
        "Hugepagesize":       int64(2048),
        "DirectMap4k":     int64(1016636),
        "DirectMap2M":    int64(15751168),

	}
	acc.AssertContainsFields(t, "meminfo", fields)
}

func TestNoMemInfoProcfile(t *testing.T) {
	tmpfile := makeFakeStatFile([]byte(vmStatFile_Invalid))
	os.Remove(tmpfile)

	k := Meminfo{
		statFile: tmpfile,
	}

	acc := testutil.Accumulator{}
	err := k.Gather(&acc)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not exist")
}

const memInfoFile_Full = `MemTotal:       16416164 kB
MemFree:        10021148 kB
MemAvailable:   11400828 kB
Buffers:           72572 kB
Cached:          1586784 kB
SwapCached:        65372 kB
Active:          3306232 kB
Inactive:        2549784 kB
Active_anon:    2813764 kB
Inactive_anon:  1462248 kB
Active_file:     492468 kB
Inactive_file:  1087536 kB
Unevictable:          32 kB
Mlocked:              32 kB
SwapTotal:       2097148 kB
SwapFree:         493224 kB
Dirty:              2256 kB
Writeback:             0 kB
AnonPages:       4135280 kB
Mapped:          1195320 kB
Shmem:            666084 kB
Slab:             242932 kB
SReclaimable:     102768 kB
SUnreclaim:       140164 kB
KernelStack:       25248 kB
PageTables:       106276 kB
NFS_Unstable:          0 kB
Bounce:                0 kB
WritebackTmp:          0 kB
CommitLimit:    10305228 kB
Committed_AS:   18852196 kB
VmallocTotal:   34359738367 kB
VmallocUsed:           0 kB
VmallocChunk:          0 kB
HardwareCorrupted:     0 kB
AnonHugePages:         0 kB
ShmemHugePages:        0 kB
ShmemPmdMapped:        0 kB
CmaTotal:              0 kB
CmaFree:               0 kB
HugePages_Total:       0
HugePages_Free:        0
HugePages_Rsvd:        0
HugePages_Surp:        0
Hugepagesize:       2048 kB
DirectMap4k:     1016636 kB
DirectMap2M:    15751168 kB
`

func makeFakeMemInfoFile(content []byte) string {
	tmpfile, err := ioutil.TempFile("", "meminfo_test")
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
