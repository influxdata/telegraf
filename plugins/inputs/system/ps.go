package system

import (
	"os"

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

func newSystemPS() *systemPS {
	return &systemPS{&systemPSDisk{}}
}

type systemPS struct {
	PSDiskDeps
}

type systemPSDisk struct{}

func (s *systemPS) CPUTimes(perCPU, totalCPU bool) ([]cpu.TimesStat, error) {
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

func (s *systemPS) DiskUsage(
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

	// Autofs mounts indicate a potential mount, the partition will also be
	// listed with the actual filesystem when mounted.  Ignore the autofs
	// partition to avoid triggering a mount.
	fstypeExcludeSet["autofs"] = true

	var usage []*disk.UsageStat
	var partitions []*disk.PartitionStat

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

		mountpoint := s.OSGetenv("HOST_MOUNT_PREFIX") + p.Mountpoint
		if _, err := s.OSStat(mountpoint); err != nil {
			continue
		}
		du, err := s.PSDiskUsage(mountpoint)
		if err != nil {
			continue
		}
		du.Path = p.Mountpoint
		du.Fstype = p.Fstype
		usage = append(usage, du)
		partitions = append(partitions, &p)
	}

	return usage, partitions, nil
}

func (s *systemPS) NetProto() ([]net.ProtoCountersStat, error) {
	return net.ProtoCounters(nil)
}

func (s *systemPS) NetIO() ([]net.IOCountersStat, error) {
	return net.IOCounters(true)
}

func (s *systemPS) NetConnections() ([]net.ConnectionStat, error) {
	return net.Connections("all")
}

func (s *systemPS) DiskIO(names []string) (map[string]disk.IOCountersStat, error) {
	m, err := disk.IOCounters(names...)
	if err == internal.NotImplementedError {
		return nil, nil
	}

	return m, err
}

func (s *systemPS) VMStat() (*mem.VirtualMemoryStat, error) {
	return mem.VirtualMemory()
}

func (s *systemPS) SwapStat() (*mem.SwapMemoryStat, error) {
	return mem.SwapMemory()
}

func (s *systemPSDisk) Partitions(all bool) ([]disk.PartitionStat, error) {
	return disk.Partitions(all)
}

func (s *systemPSDisk) OSGetenv(key string) string {
	return os.Getenv(key)
}

func (s *systemPSDisk) OSStat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

func (s *systemPSDisk) PSDiskUsage(path string) (*disk.UsageStat, error) {
	return disk.Usage(path)
}
