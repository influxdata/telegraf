package system

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

type PS interface {
	CPUTimes(perCPU, totalCPU bool) ([]cpu.TimesStat, error)
	DiskUsage(mountPointFilter []string, fstypeExclude []string) ([]*disk.UsageStat, []*disk.PartitionStat, error)
	NetIO() ([]net.IOCountersStat, error)
	NetProto() ([]net.ProtoCountersStat, error)
	DiskIO(names []string) (map[string]disk.IOCountersStat, error)
	VMStat() (*mem.VirtualMemoryStat, error)
	SwapStat() (*mem.SwapMemoryStat, error)
	NetConnections() ([]net.ConnectionStat, error)
	Temperature() ([]host.TemperatureStat, error)
}

type PSDiskDeps interface {
	Partitions(all bool) ([]disk.PartitionStat, error)
	OSGetenv(key string) string
	OSStat(name string) (os.FileInfo, error)
	PSDiskUsage(path string) (*disk.UsageStat, error)
}

func NewSystemPS() *SystemPS {
	return &SystemPS{PSDiskDeps: &SystemPSDisk{}}
}

type SystemPS struct {
	PSDiskDeps
	Log telegraf.Logger `toml:"-"`
}

type SystemPSDisk struct{}

func (s *SystemPS) CPUTimes(perCPU, totalCPU bool) ([]cpu.TimesStat, error) {
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

func (s *SystemPS) DiskUsage(
	mountPointFilter []string,
	fstypeExclude []string,
) ([]*disk.UsageStat, []*disk.PartitionStat, error) {
	parts, err := s.Partitions(true)
	if err != nil {
		return nil, nil, err
	}

	mountPointFilterSet := newSet()
	for _, filter := range mountPointFilter {
		mountPointFilterSet.add(filter)
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

	for i := range parts {
		p := parts[i]

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

		du.Path = filepath.Join("/", strings.TrimPrefix(p.Mountpoint, hostMountPrefix))
		du.Fstype = p.Fstype
		usage = append(usage, du)
		partitions = append(partitions, &p)
	}

	return usage, partitions, nil
}

func (s *SystemPS) NetProto() ([]net.ProtoCountersStat, error) {
	return net.ProtoCounters(nil)
}

func (s *SystemPS) NetIO() ([]net.IOCountersStat, error) {
	return net.IOCounters(true)
}

func (s *SystemPS) NetConnections() ([]net.ConnectionStat, error) {
	return net.Connections("all")
}

func (s *SystemPS) DiskIO(names []string) (map[string]disk.IOCountersStat, error) {
	m, err := disk.IOCounters(names...)
	if err == internal.ErrorNotImplemented {
		return nil, nil
	}

	return m, err
}

func (s *SystemPS) VMStat() (*mem.VirtualMemoryStat, error) {
	return mem.VirtualMemory()
}

func (s *SystemPS) SwapStat() (*mem.SwapMemoryStat, error) {
	return mem.SwapMemory()
}

func (s *SystemPS) Temperature() ([]host.TemperatureStat, error) {
	temp, err := host.SensorsTemperatures()
	if err != nil {
		_, ok := err.(*host.Warnings)
		if !ok {
			return temp, err
		}
	}
	return temp, nil
}

func (s *SystemPSDisk) Partitions(all bool) ([]disk.PartitionStat, error) {
	return disk.Partitions(all)
}

func (s *SystemPSDisk) OSGetenv(key string) string {
	return os.Getenv(key)
}

func (s *SystemPSDisk) OSStat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

func (s *SystemPSDisk) PSDiskUsage(path string) (*disk.UsageStat, error) {
	return disk.Usage(path)
}
