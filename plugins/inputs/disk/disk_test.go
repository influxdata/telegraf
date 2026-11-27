package disk

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/shirou/gopsutil/v4/disk"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/common/psutil"
	"github.com/influxdata/telegraf/plugins/common/psutil/psutiltest"
	"github.com/influxdata/telegraf/testutil"
)

func TestDiskUsage(t *testing.T) {
	mck := &mock.Mock{}
	mps := psutiltest.MockPSDisk{SystemPS: &psutil.SystemPS{PSDiskDeps: &psutiltest.MockDiskUsage{Mock: mck}}, Mock: mck}
	defer mps.AssertExpectations(t)

	var acc testutil.Accumulator
	var err error

	psAll := []disk.PartitionStat{
		{
			Device:     "/dev/sda",
			Mountpoint: "/",
			Fstype:     "ext4",
			Opts:       []string{"ro", "noatime", "nodiratime"},
		},
		{
			Device:     "/dev/sdb",
			Mountpoint: "/home",
			Fstype:     "ext4",
			Opts:       []string{"rw", "noatime", "nodiratime", "errors=remount-ro"},
		},
		{
			Device:     "/dev/sda",
			Mountpoint: "/var/rootbind",
			Fstype:     "ext4",
			Opts:       []string{"ro", "noatime", "nodiratime", "bind"},
		},
	}
	duAll := []disk.UsageStat{
		{
			Path:        "/",
			Fstype:      "ext4",
			Total:       128,
			Free:        23,
			Used:        100,
			InodesTotal: 1234,
			InodesFree:  234,
			InodesUsed:  1000,
		},
		{
			Path:        "/home",
			Fstype:      "ext4",
			Total:       256,
			Free:        46,
			Used:        200,
			InodesTotal: 2468,
			InodesFree:  468,
			InodesUsed:  2000,
		},
		{
			Path:        "/var/rootbind",
			Fstype:      "ext4",
			Total:       128,
			Free:        23,
			Used:        100,
			InodesTotal: 1234,
			InodesFree:  234,
			InodesUsed:  1000,
		},
	}

	mps.On("Partitions", true).Return(psAll, nil)
	mps.On("OSGetenv", "HOST_MOUNT_PREFIX").Return("")
	mps.On("PSDiskUsage", "/").Return(&duAll[0], nil)
	mps.On("PSDiskUsage", "/home").Return(&duAll[1], nil)
	mps.On("PSDiskUsage", "/var/rootbind").Return(&duAll[2], nil)

	err = (&Disk{ps: mps}).Gather(&acc)
	require.NoError(t, err)

	numDiskMetrics := acc.NFields()
	expectedAllDiskMetrics := 24
	require.Equal(t, expectedAllDiskMetrics, numDiskMetrics)

	tags1 := map[string]string{
		"path":   string(os.PathSeparator),
		"fstype": "ext4",
		"device": "sda",
		"mode":   "ro",
	}
	tags2 := map[string]string{
		"path":   fmt.Sprintf("%chome", os.PathSeparator),
		"fstype": "ext4",
		"device": "sdb",
		"mode":   "rw",
	}
	tags3 := map[string]string{
		"path":   fmt.Sprintf("%cvar%crootbind", os.PathSeparator, os.PathSeparator),
		"fstype": "ext4",
		"device": "sda",
		"mode":   "ro",
	}

	fields1 := map[string]any{
		"total":               uint64(128),
		"used":                uint64(100),
		"free":                uint64(23),
		"inodes_total":        uint64(1234),
		"inodes_free":         uint64(234),
		"inodes_used":         uint64(1000),
		"used_percent":        float64(81.30081300813008),
		"inodes_used_percent": float64(81.03727714748784),
	}
	fields2 := map[string]any{
		"total":               uint64(256),
		"used":                uint64(200),
		"free":                uint64(46),
		"inodes_total":        uint64(2468),
		"inodes_free":         uint64(468),
		"inodes_used":         uint64(2000),
		"used_percent":        float64(81.30081300813008),
		"inodes_used_percent": float64(81.03727714748784),
	}
	fields3 := map[string]any{
		"total":               uint64(128),
		"used":                uint64(100),
		"free":                uint64(23),
		"inodes_total":        uint64(1234),
		"inodes_free":         uint64(234),
		"inodes_used":         uint64(1000),
		"used_percent":        float64(81.30081300813008),
		"inodes_used_percent": float64(81.03727714748784),
	}
	acc.AssertContainsTaggedFields(t, "disk", fields1, tags1)
	acc.AssertContainsTaggedFields(t, "disk", fields2, tags2)
	acc.AssertContainsTaggedFields(t, "disk", fields3, tags3)

	// We expect 7 more DiskMetrics to show up with an explicit match on "/"
	// and /home not matching the /dev in MountPoints
	err = (&Disk{ps: &mps, MountPoints: []string{"/", "/dev"}}).Gather(&acc)
	require.NoError(t, err)
	require.Equal(t, expectedAllDiskMetrics+8, acc.NFields())

	// We should see all the diskpoints as MountPoints includes both
	// /, /home, and /var/rootbind
	err = (&Disk{ps: &mps, MountPoints: []string{"/", "/home", "/var/rootbind"}}).Gather(&acc)
	require.NoError(t, err)
	require.Equal(t, expectedAllDiskMetrics+8*4, acc.NFields())

	// We should see all the mounts as MountPoints except the bind mound
	err = (&Disk{ps: &mps, IgnoreMountOpts: []string{"bind"}}).Gather(&acc)
	require.NoError(t, err)
	require.Equal(t, expectedAllDiskMetrics+8*6, acc.NFields())
}

func TestDiskUsageHostMountPrefix(t *testing.T) {
	tests := []struct {
		name            string
		partitionStats  []disk.PartitionStat
		usageStats      []*disk.UsageStat
		hostMountPrefix string
		expectedTags    map[string]string
		expectedFields  map[string]any
	}{
		{
			name: "no host mount prefix",
			partitionStats: []disk.PartitionStat{
				{
					Device:     "/dev/sda",
					Mountpoint: "/",
					Fstype:     "ext4",
					Opts:       []string{"ro"},
				},
			},
			usageStats: []*disk.UsageStat{
				{
					Path:  "/",
					Total: 42,
				},
			},
			expectedTags: map[string]string{
				"path":   string(os.PathSeparator),
				"device": "sda",
				"fstype": "ext4",
				"mode":   "ro",
			},
			expectedFields: map[string]any{
				"total":               uint64(42),
				"used":                uint64(0),
				"free":                uint64(0),
				"inodes_total":        uint64(0),
				"inodes_free":         uint64(0),
				"inodes_used":         uint64(0),
				"used_percent":        float64(0),
				"inodes_used_percent": float64(0),
			},
		},
		{
			name: "host mount prefix",
			partitionStats: []disk.PartitionStat{
				{
					Device:     "/dev/sda",
					Mountpoint: "/hostfs/var",
					Fstype:     "ext4",
					Opts:       []string{"ro"},
				},
			},
			usageStats: []*disk.UsageStat{
				{
					Path:  "/hostfs/var",
					Total: 42,
				},
			},
			hostMountPrefix: "/hostfs",
			expectedTags: map[string]string{
				"path":   fmt.Sprintf("%cvar", os.PathSeparator),
				"device": "sda",
				"fstype": "ext4",
				"mode":   "ro",
			},
			expectedFields: map[string]any{
				"total":               uint64(42),
				"used":                uint64(0),
				"free":                uint64(0),
				"inodes_total":        uint64(0),
				"inodes_free":         uint64(0),
				"inodes_used":         uint64(0),
				"used_percent":        float64(0),
				"inodes_used_percent": float64(0),
			},
		},
		{
			name: "host mount prefix exact match",
			partitionStats: []disk.PartitionStat{
				{
					Device:     "/dev/sda",
					Mountpoint: "/hostfs",
					Fstype:     "ext4",
					Opts:       []string{"ro"},
				},
			},
			usageStats: []*disk.UsageStat{
				{
					Path:  "/hostfs",
					Total: 42,
				},
			},
			hostMountPrefix: "/hostfs",
			expectedTags: map[string]string{
				"path":   string(os.PathSeparator),
				"device": "sda",
				"fstype": "ext4",
				"mode":   "ro",
			},
			expectedFields: map[string]any{
				"total":               uint64(42),
				"used":                uint64(0),
				"free":                uint64(0),
				"inodes_total":        uint64(0),
				"inodes_free":         uint64(0),
				"inodes_used":         uint64(0),
				"used_percent":        float64(0),
				"inodes_used_percent": float64(0),
			},
		},
		{
			name: "virtual filesystem with none device",
			partitionStats: []disk.PartitionStat{
				{
					Device:     "none",
					Mountpoint: "/tmp",
					Fstype:     "tmpfs",
					Opts:       []string{"rw"},
				},
			},
			usageStats: []*disk.UsageStat{
				{
					Path:  "/tmp",
					Total: 42,
				},
			},
			expectedTags: map[string]string{
				"path":   fmt.Sprintf("%ctmp", os.PathSeparator),
				"device": "tmpfs",
				"fstype": "tmpfs",
				"mode":   "rw",
			},
			expectedFields: map[string]any{
				"total":               uint64(42),
				"used":                uint64(0),
				"free":                uint64(0),
				"inodes_total":        uint64(0),
				"inodes_free":         uint64(0),
				"inodes_used":         uint64(0),
				"used_percent":        float64(0),
				"inodes_used_percent": float64(0),
			},
		},
		{
			name: "virtual filesystem with empty device",
			partitionStats: []disk.PartitionStat{
				{
					Device:     "",
					Mountpoint: "/sys",
					Fstype:     "sysfs",
					Opts:       []string{"ro"},
				},
			},
			usageStats: []*disk.UsageStat{
				{
					Path:  "/sys",
					Total: 42,
				},
			},
			expectedTags: map[string]string{
				"path":   fmt.Sprintf("%csys", os.PathSeparator),
				"device": "sysfs",
				"fstype": "sysfs",
				"mode":   "ro",
			},
			expectedFields: map[string]any{
				"total":               uint64(42),
				"used":                uint64(0),
				"free":                uint64(0),
				"inodes_total":        uint64(0),
				"inodes_free":         uint64(0),
				"inodes_used":         uint64(0),
				"used_percent":        float64(0),
				"inodes_used_percent": float64(0),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mck := &mock.Mock{}
			mps := psutiltest.MockPSDisk{SystemPS: &psutil.SystemPS{PSDiskDeps: &psutiltest.MockDiskUsage{Mock: mck}}, Mock: mck}
			defer mps.AssertExpectations(t)

			var acc testutil.Accumulator
			var err error

			mps.On("Partitions", true).Return(tt.partitionStats, nil)

			for _, v := range tt.usageStats {
				mps.On("PSDiskUsage", v.Path).Return(v, nil)
			}

			mps.On("OSGetenv", "HOST_MOUNT_PREFIX").Return(tt.hostMountPrefix)

			err = (&Disk{ps: mps}).Gather(&acc)
			require.NoError(t, err)

			acc.AssertContainsTaggedFields(t, "disk", tt.expectedFields, tt.expectedTags)
		})
	}
}

func TestDiskStats(t *testing.T) {
	var mps psutiltest.MockPS
	defer mps.AssertExpectations(t)
	var acc testutil.Accumulator
	var err error

	duAll := []*disk.UsageStat{
		{
			Path:        "/",
			Fstype:      "ext4",
			Total:       128,
			Free:        23,
			Used:        100,
			InodesTotal: 1234,
			InodesFree:  234,
			InodesUsed:  1000,
		},
		{
			Path:        "/home",
			Fstype:      "ext4",
			Total:       256,
			Free:        46,
			Used:        200,
			InodesTotal: 2468,
			InodesFree:  468,
			InodesUsed:  2000,
		},
		{
			Path:        "/var/rootbind",
			Fstype:      "ext4",
			Total:       128,
			Free:        23,
			Used:        100,
			InodesTotal: 1234,
			InodesFree:  234,
			InodesUsed:  1000,
		},
	}
	duMountFiltered := []*disk.UsageStat{
		{
			Path:        "/",
			Fstype:      "ext4",
			Total:       128,
			Free:        23,
			Used:        100,
			InodesTotal: 1234,
			InodesFree:  234,
			InodesUsed:  1000,
		},
	}
	duOptFiltered := []*disk.UsageStat{
		{
			Path:        "/",
			Fstype:      "ext4",
			Total:       128,
			Free:        23,
			Used:        100,
			InodesTotal: 1234,
			InodesFree:  234,
			InodesUsed:  1000,
		},
		{
			Path:        "/home",
			Fstype:      "ext4",
			Total:       256,
			Free:        46,
			Used:        200,
			InodesTotal: 2468,
			InodesFree:  468,
			InodesUsed:  2000,
		},
	}

	psAll := []*disk.PartitionStat{
		{
			Device:     "/dev/sda",
			Mountpoint: "/",
			Fstype:     "ext4",
			Opts:       []string{"ro", "noatime", "nodiratime"},
		},
		{
			Device:     "/dev/sdb",
			Mountpoint: "/home",
			Fstype:     "ext4",
			Opts:       []string{"rw", "noatime", "nodiratime", "errors=remount-ro"},
		},
		{
			Device:     "/dev/sda",
			Mountpoint: "/var/rootbind",
			Fstype:     "ext4",
			Opts:       []string{"ro", "noatime", "nodiratime", "bind"},
		},
	}

	psMountFiltered := []*disk.PartitionStat{
		{
			Device:     "/dev/sda",
			Mountpoint: "/",
			Fstype:     "ext4",
			Opts:       []string{"ro", "noatime", "nodiratime"},
		},
	}
	psOptFiltered := []*disk.PartitionStat{
		{
			Device:     "/dev/sda",
			Mountpoint: "/",
			Fstype:     "ext4",
			Opts:       []string{"ro", "noatime", "nodiratime"},
		},
		{
			Device:     "/dev/sdb",
			Mountpoint: "/home",
			Fstype:     "ext4",
			Opts:       []string{"rw", "noatime", "nodiratime", "errors=remount-ro"},
		},
	}

	mps.On("DiskUsage", []string(nil), []string(nil), []string(nil)).Return(duAll, psAll, nil)
	mps.On("DiskUsage", []string{"/", "/dev"}, []string(nil), []string(nil)).Return(duMountFiltered, psMountFiltered, nil)
	mps.On("DiskUsage", []string{"/", "/home", "/var/rootbind"}, []string(nil), []string(nil)).Return(duAll, psAll, nil)
	mps.On("DiskUsage", []string(nil), []string{"bind"}, []string(nil)).Return(duOptFiltered, psOptFiltered, nil)

	err = (&Disk{ps: &mps}).Gather(&acc)
	require.NoError(t, err)

	numDiskMetrics := acc.NFields()
	expectedAllDiskMetrics := 24
	require.Equal(t, expectedAllDiskMetrics, numDiskMetrics)

	tags1 := map[string]string{
		"path":   "/",
		"fstype": "ext4",
		"device": "sda",
		"mode":   "ro",
	}
	tags2 := map[string]string{
		"path":   "/home",
		"fstype": "ext4",
		"device": "sdb",
		"mode":   "rw",
	}

	fields1 := map[string]any{
		"total":               uint64(128),
		"used":                uint64(100),
		"free":                uint64(23),
		"inodes_total":        uint64(1234),
		"inodes_free":         uint64(234),
		"inodes_used":         uint64(1000),
		"used_percent":        float64(81.30081300813008),
		"inodes_used_percent": float64(81.03727714748784),
	}
	fields2 := map[string]any{
		"total":               uint64(256),
		"used":                uint64(200),
		"free":                uint64(46),
		"inodes_total":        uint64(2468),
		"inodes_free":         uint64(468),
		"inodes_used":         uint64(2000),
		"used_percent":        float64(81.30081300813008),
		"inodes_used_percent": float64(81.03727714748784),
	}
	acc.AssertContainsTaggedFields(t, "disk", fields1, tags1)
	acc.AssertContainsTaggedFields(t, "disk", fields2, tags2)

	// We expect 7 more DiskMetrics to show up with an explicit match on "/"
	// and /home and /var/rootbind not matching the /dev in MountPoints
	err = (&Disk{ps: &mps, MountPoints: []string{"/", "/dev"}}).Gather(&acc)
	require.NoError(t, err)
	require.Equal(t, expectedAllDiskMetrics+8, acc.NFields())

	// We should see all the diskpoints as MountPoints includes both
	// /, /home, and /var/rootbind
	err = (&Disk{ps: &mps, MountPoints: []string{"/", "/home", "/var/rootbind"}}).Gather(&acc)
	require.NoError(t, err)
	require.Equal(t, expectedAllDiskMetrics+8*4, acc.NFields())

	// We should see all the mounts as MountPoints except the bind mound
	err = (&Disk{ps: &mps, IgnoreMountOpts: []string{"bind"}}).Gather(&acc)
	require.NoError(t, err)
	require.Equal(t, expectedAllDiskMetrics+8*6, acc.NFields())
}

func TestDiskUsageIssues(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Skipping due to Linux-only test-cases...")
	}

	tests := []struct {
		name     string
		prefix   string
		du       disk.UsageStat
		expected []telegraf.Metric
	}{
		{
			name:   "success",
			prefix: "",
			du: disk.UsageStat{
				Total:       256,
				Free:        46,
				Used:        200,
				InodesTotal: 2468,
				InodesFree:  468,
				InodesUsed:  2000,
			},
			expected: []telegraf.Metric{
				metric.New(
					"disk",
					map[string]string{
						"device": "tmpfs",
						"fstype": "tmpfs",
						"mode":   "rw",
						"path":   "/tmp",
					},
					map[string]any{
						"total":               uint64(256),
						"used":                uint64(200),
						"free":                uint64(46),
						"inodes_total":        uint64(2468),
						"inodes_free":         uint64(468),
						"inodes_used":         uint64(2000),
						"used_percent":        float64(81.30081300813008),
						"inodes_used_percent": float64(81.03727714748784),
					},
					time.Unix(0, 0),
					telegraf.Gauge,
				),
				metric.New(
					"disk",
					map[string]string{
						"device": "nvme0n1p4",
						"fstype": "ext4",
						"mode":   "rw",
						"path":   "/",
					},
					map[string]any{
						"total":               uint64(256),
						"used":                uint64(200),
						"free":                uint64(46),
						"inodes_total":        uint64(2468),
						"inodes_free":         uint64(468),
						"inodes_used":         uint64(2000),
						"used_percent":        float64(81.30081300813008),
						"inodes_used_percent": float64(81.03727714748784),
					},
					time.Unix(0, 0),
					telegraf.Gauge,
				),
			},
		},
		{
			name:   "issue 10297",
			prefix: "/host",
			du: disk.UsageStat{
				Total:       256,
				Free:        46,
				Used:        200,
				InodesTotal: 2468,
				InodesFree:  468,
				InodesUsed:  2000,
			},
			expected: []telegraf.Metric{
				metric.New(
					"disk",
					map[string]string{
						"device": "sda1",
						"fstype": "ext4",
						"label":  "root",
						"mode":   "rw",
						"path":   "/",
					},
					map[string]any{
						"total":               uint64(256),
						"used":                uint64(200),
						"free":                uint64(46),
						"inodes_total":        uint64(2468),
						"inodes_free":         uint64(468),
						"inodes_used":         uint64(2000),
						"used_percent":        float64(81.30081300813008),
						"inodes_used_percent": float64(81.03727714748784),
					},
					time.Unix(0, 0),
					telegraf.Gauge,
				),
				metric.New(
					"disk",
					map[string]string{
						"device": "sdb",
						"fstype": "ext4",
						"label":  "storage",
						"mode":   "rw",
						"path":   "/mnt/storage",
					},
					map[string]any{
						"total":               uint64(256),
						"used":                uint64(200),
						"free":                uint64(46),
						"inodes_total":        uint64(2468),
						"inodes_free":         uint64(468),
						"inodes_used":         uint64(2000),
						"used_percent":        float64(81.30081300813008),
						"inodes_used_percent": float64(81.03727714748784),
					},
					time.Unix(0, 0),
					telegraf.Gauge,
				),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup the environment
			hostMountPrefix := tt.prefix

			hostProcPrefix, err := filepath.Abs(filepath.Join("testdata", strings.ReplaceAll(tt.name, " ", "_"), "proc"))
			require.NoError(t, err)

			hostSysPrefix, err := filepath.Abs(filepath.Join("testdata", strings.ReplaceAll(tt.name, " ", "_"), "sys"))
			require.NoError(t, err)

			// Get the partitions in the test-case
			os.Clearenv()
			t.Setenv("HOST_PROC", hostProcPrefix)
			t.Setenv("HOST_SYS", hostSysPrefix)

			partitions, err := disk.Partitions(true)
			require.NoError(t, err)

			// Mock the disk usage
			mck := &mock.Mock{}
			mps := psutiltest.MockPSDisk{SystemPS: &psutil.SystemPS{PSDiskDeps: &psutiltest.MockDiskUsage{Mock: mck}}, Mock: mck}
			defer mps.AssertExpectations(t)

			mps.On("Partitions", true).Return(partitions, nil)

			for _, partition := range partitions {
				mountpoint := partition.Mountpoint
				if hostMountPrefix != "" {
					mountpoint = filepath.Join(hostMountPrefix, partition.Mountpoint)
				}
				diskUsage := tt.du
				diskUsage.Path = mountpoint
				diskUsage.Fstype = partition.Fstype
				mps.On("PSDiskUsage", mountpoint).Return(&diskUsage, nil)
			}
			mps.On("OSGetenv", "HOST_MOUNT_PREFIX").Return(hostMountPrefix)

			// Setup the plugin and run the test
			var acc testutil.Accumulator
			plugin := &Disk{ps: &mps}
			require.NoError(t, plugin.Gather(&acc))

			actual := acc.GetTelegrafMetrics()
			testutil.RequireMetricsEqual(t, tt.expected, actual, testutil.IgnoreTime(), testutil.SortMetrics())
		})
	}
	os.Clearenv()
}

func TestParseFstab(t *testing.T) {
	input := `
# /etc/fstab: static file system information.
   # indented comment

UUID=root-uuid /               ext4    errors=remount-ro 0       1
/dev/sdb1      /mnt/data/      xfs     rw,noatime        0       2
/dev/sdc1      /mnt/my\040dir  ext4    ro
/dev/sdd1      none            swap    sw                0       0
proc           /proc           proc    defaults          0       0
sysfs          /sys            sysfs   defaults          0       0
devpts         /dev/pts        devpts  gid=5,mode=620    0       0
/dev/sr0       /media/cdrom    iso9660 ro,noauto         0       0
malformed_line
nas:/export    /mnt/nfs        nfs4    rw,_netdev        0       0
`
	entries, err := parseFstab(strings.NewReader(input))
	require.NoError(t, err)

	expected := []fstabEntry{
		{
			device:     "UUID=root-uuid",
			mountpoint: "/",
			fstype:     "ext4",
			opts:       []string{"errors=remount-ro"},
		},
		{
			device:     "/dev/sdb1",
			mountpoint: "/mnt/data",
			fstype:     "xfs",
			opts:       []string{"rw", "noatime"},
		},
		{
			device:     "/dev/sdc1",
			mountpoint: "/mnt/my dir",
			fstype:     "ext4",
			opts:       []string{"ro"},
		},
		{
			device:     "nas:/export",
			mountpoint: "/mnt/nfs",
			fstype:     "nfs4",
			opts:       []string{"rw", "_netdev"},
		},
	}
	require.Equal(t, expected, entries)
}

func TestMountStatus(t *testing.T) {
	fstabContent := `
# Expected mounts
/dev/sda1       /               ext4    ro,noatime     0 1
UUID=1234-5678  /mnt/backup     ext4    rw             0 2
nas:/export     /mnt/nfs        nfs4    defaults       0 0
/dev/sdc1       /mnt/my\040disk xfs     ro,nofail      0 0
/dev/sdd1       /mnt/bind       none    bind           0 0
/dev/sde1       /mnt/ignored    tmpfs   rw             0 0
/dev/sdf1       /mnt/dummy      tmpfs   rw             0 0
/dev/sdg1       none            swap    sw             0 0
/dev/sr0        /mnt/cdrom      iso9660 ro,noauto      0 0
proc            /proc           proc    defaults       0 0
# Duplicate mount point should only emit once
/dev/sdh1       /mnt/backup     ext4    rw             0 2
`
	fstabPath := filepath.Join(t.TempDir(), "fstab")
	require.NoError(t, os.WriteFile(fstabPath, []byte(fstabContent), 0o600))

	partitions := []disk.PartitionStat{
		{
			Device:     "/dev/sda1",
			Mountpoint: "/",
			Fstype:     "ext4",
			Opts:       []string{"ro", "noatime"},
		},
		{
			Device:     "tmpfs",
			Mountpoint: "/mnt/dummy",
			Fstype:     "tmpfs",
			Opts:       []string{"rw"},
		},
	}
	rootUsage := disk.UsageStat{
		Path:        "/",
		Fstype:      "ext4",
		Total:       1000,
		Free:        400,
		Used:        600,
		InodesTotal: 100,
		InodesFree:  40,
		InodesUsed:  60,
	}
	dummyUsage := disk.UsageStat{
		Path:   "/mnt/dummy",
		Fstype: "tmpfs",
		Total:  0,
	}

	setupMock := func(t *testing.T) *psutiltest.MockPSDisk {
		t.Helper()
		mck := &mock.Mock{}
		mps := &psutiltest.MockPSDisk{
			SystemPS: &psutil.SystemPS{PSDiskDeps: &psutiltest.MockDiskUsage{Mock: mck}},
			Mock:     mck,
		}
		mps.On("Partitions", true).Return(partitions, nil)
		mps.On("OSGetenv", "HOST_MOUNT_PREFIX").Return("")
		mps.On("PSDiskUsage", "/").Return(&rootUsage, nil).Maybe()
		mps.On("PSDiskUsage", "/mnt/dummy").Return(&dummyUsage, nil).Maybe()
		t.Cleanup(func() {
			mps.AssertExpectations(t)
		})
		return mps
	}

	t.Run("mounted and unmounted entries", func(t *testing.T) {
		mps := setupMock(t)
		var acc testutil.Accumulator
		plugin := &Disk{
			ps:                mps,
			IgnoreFS:          []string{"tmpfs"},
			IgnoreMountOpts:   []string{"bind"},
			ReportMountStatus: true,
			FstabPath:         fstabPath,
		}
		require.NoError(t, plugin.Gather(&acc))

		expected := []telegraf.Metric{
			metric.New(
				"disk",
				map[string]string{
					"path":   string(os.PathSeparator),
					"device": "sda1",
					"fstype": "ext4",
					"mode":   "ro",
				},
				map[string]any{
					"total":               uint64(1000),
					"free":                uint64(400),
					"used":                uint64(600),
					"used_percent":        float64(60),
					"inodes_total":        uint64(100),
					"inodes_free":         uint64(40),
					"inodes_used":         uint64(60),
					"inodes_used_percent": float64(60),
					"mount_status":        "mounted",
				},
				time.Unix(0, 0),
				telegraf.Gauge,
			),
			metric.New(
				"disk",
				map[string]string{
					"path":   filepath.FromSlash("/mnt/backup"),
					"device": "UUID=1234-5678",
					"fstype": "ext4",
					"mode":   "rw",
				},
				map[string]any{
					"mount_status": "unmounted",
				},
				time.Unix(0, 0),
				telegraf.Gauge,
			),
			metric.New(
				"disk",
				map[string]string{
					"path":   filepath.FromSlash("/mnt/nfs"),
					"device": "nas:/export",
					"fstype": "nfs4",
					"mode":   "unknown",
				},
				map[string]any{
					"mount_status": "unmounted",
				},
				time.Unix(0, 0),
				telegraf.Gauge,
			),
			metric.New(
				"disk",
				map[string]string{
					"path":   filepath.FromSlash("/mnt/my disk"),
					"device": "sdc1",
					"fstype": "xfs",
					"mode":   "ro",
				},
				map[string]any{
					"mount_status": "unmounted",
				},
				time.Unix(0, 0),
				telegraf.Gauge,
			),
		}
		testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime(), testutil.SortMetrics())
	})

	t.Run("mount_points filter applies to expected mounts", func(t *testing.T) {
		mps := setupMock(t)
		var acc testutil.Accumulator
		plugin := &Disk{
			ps:                mps,
			MountPoints:       []string{"/", "/mnt/nfs/"},
			ReportMountStatus: true,
			FstabPath:         fstabPath,
		}
		require.NoError(t, plugin.Gather(&acc))

		expected := []telegraf.Metric{
			metric.New(
				"disk",
				map[string]string{
					"path":   string(os.PathSeparator),
					"device": "sda1",
					"fstype": "ext4",
					"mode":   "ro",
				},
				map[string]any{
					"total":               uint64(1000),
					"free":                uint64(400),
					"used":                uint64(600),
					"used_percent":        float64(60),
					"inodes_total":        uint64(100),
					"inodes_free":         uint64(40),
					"inodes_used":         uint64(60),
					"inodes_used_percent": float64(60),
					"mount_status":        "mounted",
				},
				time.Unix(0, 0),
				telegraf.Gauge,
			),
			metric.New(
				"disk",
				map[string]string{
					"path":   filepath.FromSlash("/mnt/nfs"),
					"device": "nas:/export",
					"fstype": "nfs4",
					"mode":   "unknown",
				},
				map[string]any{
					"mount_status": "unmounted",
				},
				time.Unix(0, 0),
				telegraf.Gauge,
			),
		}
		testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime(), testutil.SortMetrics())
	})

	t.Run("mounted zero-total filesystem is not reported unmounted", func(t *testing.T) {
		mps := setupMock(t)
		var acc testutil.Accumulator
		plugin := &Disk{
			ps:                mps,
			MountPoints:       []string{"/mnt/dummy"},
			ReportMountStatus: true,
			FstabPath:         fstabPath,
		}
		require.NoError(t, plugin.Gather(&acc))
		require.Empty(t, acc.GetTelegrafMetrics())
	})

	t.Run("disabled by default leaves behavior unchanged", func(t *testing.T) {
		mps := setupMock(t)
		var acc testutil.Accumulator
		plugin := &Disk{
			ps:        mps,
			FstabPath: filepath.Join(t.TempDir(), "nonexistent"),
		}
		require.NoError(t, plugin.Gather(&acc))

		actual := acc.GetTelegrafMetrics()
		require.Len(t, actual, 1)
		_, hasStatus := actual[0].GetField("mount_status")
		require.False(t, hasStatus)
	})

	t.Run("missing fstab returns error when enabled", func(t *testing.T) {
		mps := setupMock(t)
		var acc testutil.Accumulator
		plugin := &Disk{
			ps:                mps,
			ReportMountStatus: true,
			FstabPath:         filepath.Join(t.TempDir(), "missing-fstab"),
		}
		require.Error(t, plugin.Gather(&acc))
		require.Empty(t, acc.GetTelegrafMetrics())
	})
}
