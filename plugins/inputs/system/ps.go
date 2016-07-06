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
	DiskUsage(mountPointFilter []string, fstypeExclude []string) ([]*disk.UsageStat, error)
	NetIO() ([]net.IOCountersStat, error)
	NetProto() ([]net.ProtoCountersStat, error)
	DiskIO() (map[string]disk.IOCountersStat, error)
	VMStat() (*mem.VirtualMemoryStat, error)
	SwapStat() (*mem.SwapMemoryStat, error)
	NetConnections() ([]net.ConnectionStat, error)
}

func add(acc telegraf.Accumulator,
	name string, val float64, tags map[string]string) {
	if val >= 0 {
		acc.Add(name, val, tags)
	}
}

type systemPS struct{}

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
) ([]*disk.UsageStat, error) {
	parts, err := disk.Partitions(true)
	if err != nil {
		return nil, err
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

	var usage []*disk.UsageStat

	for _, p := range parts {
		if len(mountPointFilter) > 0 {
			// If the mount point is not a member of the filter set,
			// don't gather info on it.
			_, ok := mountPointFilterSet[p.Mountpoint]
			if !ok {
				continue
			}
		}
		mountpoint := os.Getenv("HOST_MOUNT_PREFIX") + p.Mountpoint
		if _, err := os.Stat(mountpoint); err == nil {
			du, err := disk.Usage(mountpoint)
			du.Path = p.Mountpoint
			if err != nil {
				return nil, err
			}
			// If the mount point is a member of the exclude set,
			// don't gather info on it.
			_, ok := fstypeExcludeSet[p.Fstype]
			if ok {
				continue
			}
			du.Fstype = p.Fstype
			usage = append(usage, du)
		}
	}

	return usage, nil
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

func (s *systemPS) DiskIO() (map[string]disk.IOCountersStat, error) {
	m, err := disk.IOCounters()
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
