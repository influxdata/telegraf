package system

import (
	gonet "net"
	"os"
	"strings"

	dc "github.com/fsouza/go-dockerclient"
	"github.com/influxdb/telegraf/plugins"
	"github.com/shirou/gopsutil/common"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/docker"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
)

type DockerContainerStat struct {
	Id      string
	Name    string
	Command string
	Labels  map[string]string
	CPU     *cpu.CPUTimesStat
	Mem     *docker.CgroupMemStat
}

type PS interface {
	CPUTimes(perCPU, totalCPU bool) ([]cpu.CPUTimesStat, error)
	DiskUsage() ([]*disk.DiskUsageStat, error)
	NetIO() ([]net.NetIOCountersStat, error)
	DiskIO() (map[string]disk.DiskIOCountersStat, error)
	VMStat() (*mem.VirtualMemoryStat, error)
	SwapStat() (*mem.SwapMemoryStat, error)
	DockerStat() ([]*DockerContainerStat, error)
	NetConnections() ([]net.NetConnectionStat, error)
}

func add(acc plugins.Accumulator,
	name string, val float64, tags map[string]string) {
	if val >= 0 {
		acc.Add(name, val, tags)
	}
}

type systemPS struct {
	dockerClient *dc.Client
}

func (s *systemPS) CPUTimes(perCPU, totalCPU bool) ([]cpu.CPUTimesStat, error) {
	var cpuTimes []cpu.CPUTimesStat
	if perCPU {
		if perCPUTimes, err := cpu.CPUTimes(true); err == nil {
			cpuTimes = append(cpuTimes, perCPUTimes...)
		} else {
			return nil, err
		}
	}
	if totalCPU {
		if totalCPUTimes, err := cpu.CPUTimes(false); err == nil {
			cpuTimes = append(cpuTimes, totalCPUTimes...)
		} else {
			return nil, err
		}
	}
	return cpuTimes, nil
}

func (s *systemPS) DiskUsage() ([]*disk.DiskUsageStat, error) {
	parts, err := disk.DiskPartitions(true)
	if err != nil {
		return nil, err
	}

	var usage []*disk.DiskUsageStat

	for _, p := range parts {
		if _, err := os.Stat(p.Mountpoint); err == nil {
			du, err := disk.DiskUsage(p.Mountpoint)
			if err != nil {
				return nil, err
			}
			du.Fstype = p.Fstype
			usage = append(usage, du)
		}
	}

	return usage, nil
}

func (s *systemPS) NetIO() ([]net.NetIOCountersStat, error) {
	return net.NetIOCounters(true)
}

func (s *systemPS) NetConnections() ([]net.NetConnectionStat, error) {
	return net.NetConnections("all")
}

func (s *systemPS) DiskIO() (map[string]disk.DiskIOCountersStat, error) {
	m, err := disk.DiskIOCounters()
	if err == common.NotImplementedError {
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

func (s *systemPS) DockerStat() ([]*DockerContainerStat, error) {
	if s.dockerClient == nil {
		c, err := dc.NewClient("unix:///var/run/docker.sock")
		if err != nil {
			return nil, err
		}

		s.dockerClient = c
	}

	opts := dc.ListContainersOptions{}

	containers, err := s.dockerClient.ListContainers(opts)
	if err != nil {
		if _, ok := err.(*gonet.OpError); ok {
			return nil, nil
		}

		return nil, err
	}

	var stats []*DockerContainerStat

	for _, container := range containers {
		ctu, err := docker.CgroupCPUDocker(container.ID)
		if err != nil {
			return nil, err
		}

		mem, err := docker.CgroupMemDocker(container.ID)
		if err != nil {
			return nil, err
		}

		name := strings.Join(container.Names, " ")

		stats = append(stats, &DockerContainerStat{
			Id:      container.ID,
			Name:    name,
			Command: container.Command,
			Labels:  container.Labels,
			CPU:     ctu,
			Mem:     mem,
		})
	}

	return stats, nil
}
