package system

import (
	"fmt"

	"github.com/influxdb/tivan/plugins"
	"github.com/influxdb/tivan/plugins/system/ps/common"
	"github.com/influxdb/tivan/plugins/system/ps/cpu"
	"github.com/influxdb/tivan/plugins/system/ps/disk"
	"github.com/influxdb/tivan/plugins/system/ps/load"
	"github.com/influxdb/tivan/plugins/system/ps/net"
)

type PS interface {
	LoadAvg() (*load.LoadAvgStat, error)
	CPUTimes() ([]cpu.CPUTimesStat, error)
	DiskUsage() ([]*disk.DiskUsageStat, error)
	NetIO() ([]net.NetIOCountersStat, error)
	DiskIO() (map[string]disk.DiskIOCountersStat, error)
}

type SystemStats struct {
	ps PS
}

func (s *SystemStats) add(acc plugins.Accumulator, name string, val float64) {
	if val >= 0 {
		acc.Add(name, val, nil)
	}
}

func (s *SystemStats) Gather(acc plugins.Accumulator) error {
	lv, err := s.ps.LoadAvg()
	if err != nil {
		return err
	}

	acc.Add("load1", lv.Load1, nil)
	acc.Add("load5", lv.Load5, nil)
	acc.Add("load15", lv.Load15, nil)

	times, err := s.ps.CPUTimes()
	if err != nil {
		return fmt.Errorf("error getting CPU info: %s", err)
	}

	for _, cts := range times {
		s.add(acc, cts.CPU+".user", cts.User)
		s.add(acc, cts.CPU+".system", cts.System)
		s.add(acc, cts.CPU+".idle", cts.Idle)
		s.add(acc, cts.CPU+".nice", cts.Nice)
		s.add(acc, cts.CPU+".iowait", cts.Iowait)
		s.add(acc, cts.CPU+".irq", cts.Irq)
		s.add(acc, cts.CPU+".softirq", cts.Softirq)
		s.add(acc, cts.CPU+".steal", cts.Steal)
		s.add(acc, cts.CPU+".guest", cts.Guest)
		s.add(acc, cts.CPU+".guestNice", cts.GuestNice)
		s.add(acc, cts.CPU+".stolen", cts.Stolen)
	}

	disks, err := s.ps.DiskUsage()
	if err != nil {
		return err
	}

	for _, du := range disks {
		tags := map[string]string{
			"path": du.Path,
		}

		acc.Add("total", du.Total, tags)
		acc.Add("free", du.Free, tags)
		acc.Add("used", du.Total-du.Free, tags)
		acc.Add("inodes_total", du.InodesTotal, tags)
		acc.Add("inodes_free", du.InodesFree, tags)
		acc.Add("inodes_used", du.InodesTotal-du.InodesFree, tags)
	}

	diskio, err := s.ps.DiskIO()
	if err != nil {
		return err
	}

	for _, io := range diskio {
		tags := map[string]string{
			"name":   io.Name,
			"serial": io.SerialNumber,
		}

		acc.Add("reads", io.ReadCount, tags)
		acc.Add("writes", io.WriteCount, tags)
		acc.Add("read_bytes", io.ReadBytes, tags)
		acc.Add("write_bytes", io.WriteBytes, tags)
		acc.Add("read_time", io.ReadTime, tags)
		acc.Add("write_time", io.WriteTime, tags)
		acc.Add("io_time", io.IoTime, tags)
	}

	netio, err := s.ps.NetIO()
	if err != nil {
		return err
	}

	for _, io := range netio {
		tags := map[string]string{
			"interface": io.Name,
		}

		acc.Add("bytes_sent", io.BytesSent, tags)
		acc.Add("bytes_recv", io.BytesRecv, tags)
		acc.Add("packets_sent", io.PacketsSent, tags)
		acc.Add("packets_recv", io.PacketsRecv, tags)
		acc.Add("err_in", io.Errin, tags)
		acc.Add("err_out", io.Errout, tags)
		acc.Add("drop_in", io.Dropin, tags)
		acc.Add("drop_out", io.Dropout, tags)
	}

	return nil
}

type systemPS struct{}

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

func init() {
	plugins.Add("system", func() plugins.Plugin {
		return &SystemStats{ps: &systemPS{}}
	})
}
