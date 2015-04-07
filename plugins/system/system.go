package system

import (
	"fmt"
	gonet "net"
	"strings"

	dc "github.com/fsouza/go-dockerclient"
	"github.com/influxdb/tivan/plugins"
	"github.com/influxdb/tivan/plugins/system/ps/common"
	"github.com/influxdb/tivan/plugins/system/ps/cpu"
	"github.com/influxdb/tivan/plugins/system/ps/disk"
	"github.com/influxdb/tivan/plugins/system/ps/docker"
	"github.com/influxdb/tivan/plugins/system/ps/load"
	"github.com/influxdb/tivan/plugins/system/ps/mem"
	"github.com/influxdb/tivan/plugins/system/ps/net"
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

type SystemStats struct {
	ps PS
}

func (s *SystemStats) add(acc plugins.Accumulator,
	name string, val float64, tags map[string]string) {
	if val >= 0 {
		acc.Add(name, val, tags)
	}
}

func (s *SystemStats) Gather(acc plugins.Accumulator) error {
	lv, err := s.ps.LoadAvg()
	if err != nil {
		return err
	}

	acc.Add("system_load1", lv.Load1, nil)
	acc.Add("system_load5", lv.Load5, nil)
	acc.Add("system_load15", lv.Load15, nil)

	times, err := s.ps.CPUTimes()
	if err != nil {
		return fmt.Errorf("error getting CPU info: %s", err)
	}

	for _, cts := range times {
		tags := map[string]string{
			"cpu": cts.CPU,
		}

		s.add(acc, "cpu_user", cts.User, tags)
		s.add(acc, "cpu_system", cts.System, tags)
		s.add(acc, "cpu_idle", cts.Idle, tags)
		s.add(acc, "cpu_nice", cts.Nice, tags)
		s.add(acc, "cpu_iowait", cts.Iowait, tags)
		s.add(acc, "cpu_irq", cts.Irq, tags)
		s.add(acc, "cpu_softirq", cts.Softirq, tags)
		s.add(acc, "cpu_steal", cts.Steal, tags)
		s.add(acc, "cpu_guest", cts.Guest, tags)
		s.add(acc, "cpu_guestNice", cts.GuestNice, tags)
		s.add(acc, "cpu_stolen", cts.Stolen, tags)
	}

	disks, err := s.ps.DiskUsage()
	if err != nil {
		return fmt.Errorf("error getting disk usage info: %s", err)
	}

	for _, du := range disks {
		tags := map[string]string{
			"path": du.Path,
		}

		acc.Add("disk_total", du.Total, tags)
		acc.Add("disk_free", du.Free, tags)
		acc.Add("disk_used", du.Total-du.Free, tags)
		acc.Add("disk_inodes_total", du.InodesTotal, tags)
		acc.Add("disk_inodes_free", du.InodesFree, tags)
		acc.Add("disk_inodes_used", du.InodesTotal-du.InodesFree, tags)
	}

	diskio, err := s.ps.DiskIO()
	if err != nil {
		return fmt.Errorf("error getting disk io info: %s", err)
	}

	for _, io := range diskio {
		tags := map[string]string{
			"name":   io.Name,
			"serial": io.SerialNumber,
		}

		acc.Add("io_reads", io.ReadCount, tags)
		acc.Add("io_writes", io.WriteCount, tags)
		acc.Add("io_read_bytes", io.ReadBytes, tags)
		acc.Add("io_write_bytes", io.WriteBytes, tags)
		acc.Add("io_read_time", io.ReadTime, tags)
		acc.Add("io_write_time", io.WriteTime, tags)
		acc.Add("io_io_time", io.IoTime, tags)
	}

	netio, err := s.ps.NetIO()
	if err != nil {
		return fmt.Errorf("error getting net io info: %s", err)
	}

	for _, io := range netio {
		tags := map[string]string{
			"interface": io.Name,
		}

		acc.Add("net_bytes_sent", io.BytesSent, tags)
		acc.Add("net_bytes_recv", io.BytesRecv, tags)
		acc.Add("net_packets_sent", io.PacketsSent, tags)
		acc.Add("net_packets_recv", io.PacketsRecv, tags)
		acc.Add("net_err_in", io.Errin, tags)
		acc.Add("net_err_out", io.Errout, tags)
		acc.Add("net_drop_in", io.Dropin, tags)
		acc.Add("net_drop_out", io.Dropout, tags)
	}

	vm, err := s.ps.VMStat()
	if err != nil {
		return fmt.Errorf("error getting virtual memory info: %s", err)
	}

	vmtags := map[string]string(nil)

	acc.Add("mem_total", vm.Total, vmtags)
	acc.Add("mem_available", vm.Available, vmtags)
	acc.Add("mem_used", vm.Used, vmtags)
	acc.Add("mem_used_prec", vm.UsedPercent, vmtags)
	acc.Add("mem_free", vm.Free, vmtags)
	acc.Add("mem_active", vm.Active, vmtags)
	acc.Add("mem_inactive", vm.Inactive, vmtags)
	acc.Add("mem_buffers", vm.Buffers, vmtags)
	acc.Add("mem_cached", vm.Cached, vmtags)
	acc.Add("mem_wired", vm.Wired, vmtags)
	acc.Add("mem_shared", vm.Shared, vmtags)

	swap, err := s.ps.SwapStat()
	if err != nil {
		return fmt.Errorf("error getting swap memory info: %s", err)
	}

	swaptags := map[string]string(nil)

	acc.Add("swap_total", swap.Total, swaptags)
	acc.Add("swap_used", swap.Used, swaptags)
	acc.Add("swap_free", swap.Free, swaptags)
	acc.Add("swap_used_perc", swap.UsedPercent, swaptags)
	acc.Add("swap_in", swap.Sin, swaptags)
	acc.Add("swap_out", swap.Sout, swaptags)

	containers, err := s.ps.DockerStat()
	if err != nil {
		return fmt.Errorf("error getting docker info: %s", err)
	}

	for _, cont := range containers {
		tags := map[string]string{
			"id":      cont.Id,
			"name":    cont.Name,
			"command": cont.Command,
		}

		cts := cont.CPU

		acc.Add("docker_user", cts.User, tags)
		acc.Add("docker_system", cts.System, tags)
		acc.Add("docker_idle", cts.Idle, tags)
		acc.Add("docker_nice", cts.Nice, tags)
		acc.Add("docker_iowait", cts.Iowait, tags)
		acc.Add("docker_irq", cts.Irq, tags)
		acc.Add("docker_softirq", cts.Softirq, tags)
		acc.Add("docker_steal", cts.Steal, tags)
		acc.Add("docker_guest", cts.Guest, tags)
		acc.Add("docker_guestNice", cts.GuestNice, tags)
		acc.Add("docker_stolen", cts.Stolen, tags)

		acc.Add("docker_cache", cont.Mem.Cache, tags)
		acc.Add("docker_rss", cont.Mem.RSS, tags)
		acc.Add("docker_rss_huge", cont.Mem.RSSHuge, tags)
		acc.Add("docker_mapped_file", cont.Mem.MappedFile, tags)
		acc.Add("docker_swap_in", cont.Mem.Pgpgin, tags)
		acc.Add("docker_swap_out", cont.Mem.Pgpgout, tags)
		acc.Add("docker_page_fault", cont.Mem.Pgfault, tags)
		acc.Add("docker_page_major_fault", cont.Mem.Pgmajfault, tags)
		acc.Add("docker_inactive_anon", cont.Mem.InactiveAnon, tags)
		acc.Add("docker_active_anon", cont.Mem.ActiveAnon, tags)
		acc.Add("docker_inactive_file", cont.Mem.InactiveFile, tags)
		acc.Add("docker_active_file", cont.Mem.ActiveFile, tags)
		acc.Add("docker_unevictable", cont.Mem.Unevictable, tags)
		acc.Add("docker_memory_limit", cont.Mem.HierarchicalMemoryLimit, tags)
		acc.Add("docker_total_cache", cont.Mem.TotalCache, tags)
		acc.Add("docker_total_rss", cont.Mem.TotalRSS, tags)
		acc.Add("docker_total_rss_huge", cont.Mem.TotalRSSHuge, tags)
		acc.Add("docker_total_mapped_file", cont.Mem.TotalMappedFile, tags)
		acc.Add("docker_total_swap_in", cont.Mem.TotalPgpgIn, tags)
		acc.Add("docker_total_swap_out", cont.Mem.TotalPgpgOut, tags)
		acc.Add("docker_total_page_fault", cont.Mem.TotalPgFault, tags)
		acc.Add("docker_total_page_major_fault", cont.Mem.TotalPgMajFault, tags)
		acc.Add("docker_total_inactive_anon", cont.Mem.TotalInactiveAnon, tags)
		acc.Add("docker_total_active_anon", cont.Mem.TotalActiveAnon, tags)
		acc.Add("docker_total_inactive_file", cont.Mem.TotalInactiveFile, tags)
		acc.Add("docker_total_active_file", cont.Mem.TotalActiveFile, tags)
		acc.Add("docker_total_unevictable", cont.Mem.TotalUnevictable, tags)
	}

	return nil
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

func init() {
	plugins.Add("system", func() plugins.Plugin {
		return &SystemStats{ps: &systemPS{}}
	})
}
