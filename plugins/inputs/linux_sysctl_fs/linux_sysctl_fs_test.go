package linux_sysctl_fs

import (
	"os"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestSysctlFSGather(t *testing.T) {
	td := t.TempDir()

	require.NoError(t, os.WriteFile(td+"/aio-nr", []byte("100\n"), 0644))
	require.NoError(t, os.WriteFile(td+"/aio-max-nr", []byte("101\n"), 0644))
	require.NoError(t, os.WriteFile(td+"/super-nr", []byte("102\n"), 0644))
	require.NoError(t, os.WriteFile(td+"/super-max", []byte("103\n"), 0644))
	require.NoError(t, os.WriteFile(td+"/file-nr", []byte("104\t0\t106\n"), 0644))
	require.NoError(t, os.WriteFile(td+"/inode-state", []byte("107\t108\t109\t0\t0\t0\t0\n"), 0644))

	sfs := &SysctlFS{
		path: td,
	}
	var acc testutil.Accumulator
	require.NoError(t, sfs.Gather(&acc))

	acc.AssertContainsFields(t, "linux_sysctl_fs", map[string]interface{}{
		"aio-nr":             uint64(100),
		"aio-max-nr":         uint64(101),
		"super-nr":           uint64(102),
		"super-max":          uint64(103),
		"file-nr":            uint64(104),
		"file-max":           uint64(106),
		"inode-nr":           uint64(107),
		"inode-free-nr":      uint64(108),
		"inode-preshrink-nr": uint64(109),
	})
}
