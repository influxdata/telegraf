// +build linux

package linux_mem

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

var meminfoSample = `MemTotal:         500132 kB
MemFree:           26020 kB
MemAvailable:     245104 kB
Buffers:           59832 kB
Cached:           166608 kB
SwapCached:          900 kB
Active:           218916 kB
Inactive:         169816 kB
Active(anon):      91756 kB
Inactive(anon):    99720 kB
Active(file):     127160 kB
Inactive(file):    70096 kB
Unevictable:           0 kB
Mlocked:               0 kB
SwapTotal:       1572860 kB
SwapFree:        1562852 kB
Dirty:                 8 kB
Writeback:             0 kB
AnonPages:        161540 kB
Mapped:            27648 kB
Shmem:             29184 kB
Slab:              61732 kB
SReclaimable:      34980 kB
SUnreclaim:        26752 kB
KernelStack:        2336 kB
PageTables:         5688 kB
NFS_Unstable:          0 kB
Bounce:                0 kB
WritebackTmp:          0 kB
CommitLimit:     1822924 kB
Committed_AS:     566032 kB
VmallocTotal:   34359738367 kB
VmallocUsed:        4100 kB
VmallocChunk:   34359731632 kB
HardwareCorrupted:     0 kB
AnonHugePages:         0 kB
HugePages_Total:       0
HugePages_Free:        0
HugePages_Rsvd:        0
HugePages_Surp:        0
Hugepagesize:       2048 kB
DirectMap4k:      108480 kB
DirectMap2M:      415744 kB
`

func TestMeminfoGather(t *testing.T) {
	tf, err := ioutil.TempFile("", "")
	require.NoError(t, err)
	defer os.Remove(tf.Name())
	defer func(p string) { meminfoPath = p }(meminfoPath)
	meminfoPath = tf.Name()

	_, err = tf.Write([]byte(meminfoSample))
	require.NoError(t, err)

	mi := Meminfo{}

	var acc testutil.Accumulator
	require.NoError(t, mi.Gather(&acc))

	memfields := map[string]interface{}{
		"MemTotal":          uint64(500132),
		"MemFree":           uint64(26020),
		"MemAvailable":      uint64(245104),
		"Buffers":           uint64(59832),
		"Cached":            uint64(166608),
		"SwapCached":        uint64(900),
		"Active":            uint64(218916),
		"Inactive":          uint64(169816),
		"Active_anon":       uint64(91756),
		"Inactive_anon":     uint64(99720),
		"Active_file":       uint64(127160),
		"Inactive_file":     uint64(70096),
		"Unevictable":       uint64(0),
		"Mlocked":           uint64(0),
		"SwapTotal":         uint64(1572860),
		"SwapFree":          uint64(1562852),
		"Dirty":             uint64(8),
		"Writeback":         uint64(0),
		"AnonPages":         uint64(161540),
		"Mapped":            uint64(27648),
		"Shmem":             uint64(29184),
		"Slab":              uint64(61732),
		"SReclaimable":      uint64(34980),
		"SUnreclaim":        uint64(26752),
		"KernelStack":       uint64(2336),
		"PageTables":        uint64(5688),
		"NFS_Unstable":      uint64(0),
		"Bounce":            uint64(0),
		"WritebackTmp":      uint64(0),
		"CommitLimit":       uint64(1822924),
		"Committed_AS":      uint64(566032),
		"VmallocTotal":      uint64(34359738367),
		"VmallocUsed":       uint64(4100),
		"VmallocChunk":      uint64(34359731632),
		"HardwareCorrupted": uint64(0),
		"AnonHugePages":     uint64(0),
		"HugePages_Total":   uint64(0),
		"HugePages_Free":    uint64(0),
		"HugePages_Rsvd":    uint64(0),
		"HugePages_Surp":    uint64(0),
		"Hugepagesize":      uint64(2048),
		"DirectMap4k":       uint64(108480),
		"DirectMap2M":       uint64(415744),
	}
	acc.AssertContainsFields(t, "mem", memfields)
}
