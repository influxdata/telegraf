package system

import (
	gonet "net"
	"strings"

	dc "github.com/fsouza/go-dockerclient"
	"github.com/influxdb/telegraf/plugins"
	"github.com/influxdb/telegraf/plugins/system/ps/common"
	"github.com/influxdb/telegraf/plugins/system/ps/cpu"
	"github.com/influxdb/telegraf/plugins/system/ps/disk"
	"github.com/influxdb/telegraf/plugins/system/ps/docker"
	"github.com/influxdb/telegraf/plugins/system/ps/load"
	"github.com/influxdb/telegraf/plugins/system/ps/mem"
	"github.com/influxdb/telegraf/plugins/system/ps/net"
)

type DockerContainerStat struct {
	Id      string
	Name    string
	Command string
	CPU     *cpu.CPUTimesStat
	Mem     *docker.CgroupMemStat
}

type PS interface {
	LoadAvg() (*load.LoadAvgStat, error)
	CPUTimes() ([]cpu.CPUTimesStat, error)
	DiskUsage() ([]*disk.DiskUsageStat, error)
	NetIO() ([]net.NetIOCountersStat, error)
	DiskIO() (map[string]disk.DiskIOCountersStat, error)
	VMStat() (*mem.VirtualMemoryStat, error)
	SwapStat() (*mem.SwapMemoryStat, error)
	DockerStat() ([]*DockerContainerStat, error)
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

func (s *systemPS) LoadAvg() (*load.LoadAvgStat, error) {
	return load.LoadAvg()
}

func (s *systemPS) CPUTimes() ([]cpu.CPUTimesStat, error) {
	return cpu.CPUTimes(true)
}

func (s *systemPS) DiskUsage() ([]*disk.DiskUsageStat, error) {
	parts, err := disk.DiskPartitions(true)
	if err != nil {
		return nil, err
	}

	var usage []*disk.DiskUsageStat

	for _, p := range parts {
		du, err := disk.DiskUsage(p.Mountpoint)
		du.Fstype = p.Fstype
		if err != nil {
			return nil, err
		}

		usage = append(usage, du)
	}

	return usage, nil
}

func (s *systemPS) NetIO() ([]net.NetIOCountersStat, error) {
	return net.NetIOCounters(true)
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

	list, err := s.dockerClient.ListContainers(opts)
	if err != nil {
		if _, ok := err.(*gonet.OpError); ok {
			return nil, nil
		}

		return nil, err
	}

	var stats []*DockerContainerStat

	for _, cont := range list {
		ctu, err := docker.CgroupCPUDocker(cont.ID)
		if err != nil {
			return nil, err
		}

		mem, err := docker.CgroupMemDocker(cont.ID)
		if err != nil {
			return nil, err
		}

		name := strings.Join(cont.Names, " ")

		stats = append(stats, &DockerContainerStat{
			Id:      cont.ID,
			Name:    name,
			Command: cont.Command,
			CPU:     ctu,
			Mem:     mem,
		})
	}

	return stats, nil
}
