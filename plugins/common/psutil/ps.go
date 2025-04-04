package psutil

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
)

// PS is an interface that defines methods for gathering system statistics.
type PS interface {
	// CPUTimes returns the CPU times statistics.
	CPUTimes(perCPU, totalCPU bool) ([]cpu.TimesStat, error)
	// DiskUsage returns the disk usage statistics.
	DiskUsage(mountPointFilter []string, mountOptsExclude []string, fstypeExclude []string) ([]*disk.UsageStat, []*disk.PartitionStat, error)
	// NetIO returns network I/O statistics for every network interface installed on the system.
	NetIO() ([]net.IOCountersStat, error)
	// NetProto returns network statistics for the entire system.
	NetProto() ([]net.ProtoCountersStat, error)
	// DiskIO returns the disk I/O statistics.
	DiskIO(names []string) (map[string]disk.IOCountersStat, error)
	// VMStat returns the virtual memory statistics.
	VMStat() (*mem.VirtualMemoryStat, error)
	// SwapStat returns the swap memory statistics.
	SwapStat() (*mem.SwapMemoryStat, error)
	// NetConnections returns a list of network connections opened.
	NetConnections() ([]net.ConnectionStat, error)
	// NetConntrack returns more detailed info about the conntrack table.
	NetConntrack(perCPU bool) ([]net.ConntrackStat, error)
}

// PSDiskDeps is an interface that defines methods for gathering disk statistics.
type PSDiskDeps interface {
	// Partitions returns the disk partition statistics.
	Partitions(all bool) ([]disk.PartitionStat, error)
	// OSGetenv returns the value of the environment variable named by the key.
	OSGetenv(key string) string
	// OSStat returns the FileInfo structure describing the named file.
	OSStat(name string) (os.FileInfo, error)
	// PSDiskUsage returns a file system usage for the specified path.
	PSDiskUsage(path string) (*disk.UsageStat, error)
}

// SystemPS is a struct that implements the PS interface.
type SystemPS struct {
	PSDiskDeps
	Log telegraf.Logger `toml:"-"`
}

// SystemPSDisk is a struct that implements the PSDiskDeps interface.
type SystemPSDisk struct{}

// NewSystemPS creates a new instance of SystemPS.
func NewSystemPS() *SystemPS {
	return &SystemPS{PSDiskDeps: &SystemPSDisk{}}
}

// CPUTimes returns the CPU times statistics.
func (*SystemPS) CPUTimes(perCPU, totalCPU bool) ([]cpu.TimesStat, error) {
	var cpuTimes []cpu.TimesStat
	if perCPU {
		perCPUTimes, err := cpu.Times(true)
		if err != nil {
			return nil, err
		}
		cpuTimes = append(cpuTimes, perCPUTimes...)
	}
	if totalCPU {
		totalCPUTimes, err := cpu.Times(false)
		if err != nil {
			return nil, err
		}
		cpuTimes = append(cpuTimes, totalCPUTimes...)
	}
	return cpuTimes, nil
}

// DiskUsage returns the disk usage statistics.
func (s *SystemPS) DiskUsage(mountPointFilter, mountOptsExclude, fstypeExclude []string) ([]*disk.UsageStat, []*disk.PartitionStat, error) {
	parts, err := s.Partitions(true)
	if err != nil {
		return nil, nil, err
	}

	mountPointFilterSet := newSet()
	for _, filter := range mountPointFilter {
		mountPointFilterSet.add(filter)
	}
	mountOptFilterSet := newSet()
	for _, filter := range mountOptsExclude {
		mountOptFilterSet.add(filter)
	}
	fstypeExcludeSet := newSet()
	for _, filter := range fstypeExclude {
		fstypeExcludeSet.add(filter)
	}
	paths := newSet()
	for _, part := range parts {
		paths.add(part.Mountpoint)
	}

	// Autofs mounts indicate a potential mount, the partition will also be
	// listed with the actual filesystem when mounted.  Ignore the autofs
	// partition to avoid triggering a mount.
	fstypeExcludeSet.add("autofs")

	var usage []*disk.UsageStat
	var partitions []*disk.PartitionStat
	hostMountPrefix := s.OSGetenv("HOST_MOUNT_PREFIX")

partitionRange:
	for i := range parts {
		p := parts[i]

		for _, o := range p.Opts {
			if !mountOptFilterSet.empty() && mountOptFilterSet.has(o) {
				continue partitionRange
			}
		}
		// If there is a filter set and if the mount point is not a
		// member of the filter set, don't gather info on it.
		if !mountPointFilterSet.empty() && !mountPointFilterSet.has(p.Mountpoint) {
			continue
		}

		// If the mount point is a member of the exclude set,
		// don't gather info on it.
		if fstypeExcludeSet.has(p.Fstype) {
			continue
		}

		// If there's a host mount prefix use it as newer gopsutil version check for
		// the init's mountpoints usually pointing to the host-mountpoint but in the
		// container. This won't work for checking the disk-usage as the disks are
		// mounted at HOST_MOUNT_PREFIX...
		mountpoint := p.Mountpoint
		if hostMountPrefix != "" && !strings.HasPrefix(p.Mountpoint, hostMountPrefix) {
			mountpoint = filepath.Join(hostMountPrefix, p.Mountpoint)
			// Exclude conflicting paths
			if paths.has(mountpoint) {
				if s.Log != nil {
					s.Log.Debugf("[SystemPS] => dropped by mount prefix (%q): %q", mountpoint, hostMountPrefix)
				}
				continue
			}
		}

		du, err := s.PSDiskUsage(mountpoint)
		if err != nil {
			if s.Log != nil {
				s.Log.Debugf("[SystemPS] => unable to get disk usage (%q): %v", mountpoint, err)
			}
			continue
		}

		du.Path = filepath.Join(string(os.PathSeparator), strings.TrimPrefix(p.Mountpoint, hostMountPrefix))
		du.Fstype = p.Fstype
		usage = append(usage, du)
		partitions = append(partitions, &p)
	}

	return usage, partitions, nil
}

// NetProto returns network statistics for the entire system.
func (*SystemPS) NetProto() ([]net.ProtoCountersStat, error) {
	return net.ProtoCounters(nil)
}

// NetIO returns network I/O statistics for every network interface installed on the system.
func (*SystemPS) NetIO() ([]net.IOCountersStat, error) {
	return net.IOCounters(true)
}

// NetConnections returns a list of network connections opened.
func (*SystemPS) NetConnections() ([]net.ConnectionStat, error) {
	return net.Connections("all")
}

// NetConntrack returns more detailed info about the conntrack table.
func (*SystemPS) NetConntrack(perCPU bool) ([]net.ConntrackStat, error) {
	return net.ConntrackStats(perCPU)
}

// DiskIO returns the disk I/O statistics.
func (*SystemPS) DiskIO(names []string) (map[string]disk.IOCountersStat, error) {
	m, err := disk.IOCounters(names...)
	if errors.Is(err, internal.ErrNotImplemented) {
		return nil, nil
	}

	return m, err
}

// VMStat returns the virtual memory statistics.
func (*SystemPS) VMStat() (*mem.VirtualMemoryStat, error) {
	return mem.VirtualMemory()
}

// SwapStat returns the swap memory statistics.
func (*SystemPS) SwapStat() (*mem.SwapMemoryStat, error) {
	return mem.SwapMemory()
}

// Partitions returns the disk partition statistics.
func (*SystemPSDisk) Partitions(all bool) ([]disk.PartitionStat, error) {
	return disk.Partitions(all)
}

// OSGetenv returns the value of the environment variable named by the key.
func (*SystemPSDisk) OSGetenv(key string) string {
	return os.Getenv(key)
}

// OSStat returns the FileInfo structure describing the named file.
func (*SystemPSDisk) OSStat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

// PSDiskUsage returns a file system usage for the specified path.
func (*SystemPSDisk) PSDiskUsage(path string) (*disk.UsageStat, error) {
	return disk.Usage(path)
}

type set struct {
	m map[string]struct{}
}

func (s *set) empty() bool {
	return len(s.m) == 0
}

func (s *set) add(key string) {
	s.m[key] = struct{}{}
}

func (s *set) has(key string) bool {
	var ok bool
	_, ok = s.m[key]
	return ok
}

func newSet() *set {
	s := &set{
		m: make(map[string]struct{}),
	}
	return s
}
