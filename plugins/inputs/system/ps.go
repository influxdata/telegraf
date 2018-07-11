package system

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
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
}

type PSDiskDeps interface {
	Partitions(all bool) ([]disk.PartitionStat, error)
	OSGetenv(key string) string
	OSStat(name string) (os.FileInfo, error)
	PSDiskUsage(path string) (*disk.UsageStat, error)
}

func add(acc telegraf.Accumulator,
	name string, val float64, tags map[string]string) {
	if val >= 0 {
		acc.AddFields(name, map[string]interface{}{"value": val}, tags)
	}
}

func NewSystemPS() *SystemPS {
	return &SystemPS{&SystemPSDisk{}}
}

type SystemPS struct {
	PSDiskDeps
}

type SystemPSDisk struct{}

func (s *SystemPS) CPUTimes(perCPU, totalCPU bool) ([]cpu.TimesStat, error) {
	var cpuTimes []cpu.TimesStat
	if perCPU {
		if perCPUTimes, err := cpu.Times(true); err == nil {
			cpuTimes = append(cpuTimes, perCPUTimes...)
		} else {
			return nil, err
		}
	}
	if totalCPU {
		if totalCPUTimes, err := cpu.Times(false); err == nil {
			cpuTimes = append(cpuTimes, totalCPUTimes...)
		} else {
			return nil, err
		}
	}
	return cpuTimes, nil
}

func (s *SystemPS) DiskUsage(
	mountPointFilter []string,
	fstypeExclude []string,
) ([]*disk.UsageStat, []*disk.PartitionStat, error) {
	parts, err := s.Partitions(true)
	if err != nil {
		return nil, nil, err
	}

	// Make a "set" out of the filter slice
	mountPointFilterSet := make(map[string]bool)
	for _, filter := range mountPointFilter {
		mountPointFilterSet[filter] = true
	}
	fstypeExcludeSet := make(map[string]bool)
	for _, filter := range fstypeExclude {
		fstypeExcludeSet[filter] = true
	}
	paths := make(map[string]bool)
	for _, part := range parts {
		paths[part.Mountpoint] = true
	}

	// Autofs mounts indicate a potential mount, the partition will also be
	// listed with the actual filesystem when mounted.  Ignore the autofs
	// partition to avoid triggering a mount.
	fstypeExcludeSet["autofs"] = true

	var usage []*disk.UsageStat
	var partitions []*disk.PartitionStat
	hostMountPrefix := s.OSGetenv("HOST_MOUNT_PREFIX")

	for i := range parts {
		p := parts[i]

		if len(mountPointFilter) > 0 {
			// If the mount point is not a member of the filter set,
			// don't gather info on it.
			if _, ok := mountPointFilterSet[p.Mountpoint]; !ok {
				continue
			}
		}

		// If the mount point is a member of the exclude set,
		// don't gather info on it.
		if _, ok := fstypeExcludeSet[p.Fstype]; ok {
			continue
		}

		// If there's a host mount prefix, exclude any paths which conflict
		// with the prefix.
		if len(hostMountPrefix) > 0 &&
			!strings.HasPrefix(p.Mountpoint, hostMountPrefix) &&
			paths[hostMountPrefix+p.Mountpoint] {
			continue
		}

		du, err := s.PSDiskUsage(p.Mountpoint)
		if err != nil {
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
	if err == internal.NotImplementedError {
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
