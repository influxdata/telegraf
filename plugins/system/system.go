package system

import (
	"fmt"

	"github.com/influxdb/tivan/plugins"
	"github.com/influxdb/tivan/plugins/system/ps/cpu"
	"github.com/influxdb/tivan/plugins/system/ps/disk"
	"github.com/influxdb/tivan/plugins/system/ps/load"
)

type PS interface {
	LoadAvg() (*load.LoadAvgStat, error)
	CPUTimes() ([]cpu.CPUTimesStat, error)
	DiskUsage() ([]*disk.DiskUsageStat, error)
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

func init() {
	plugins.Add("system", func() plugins.Plugin {
		return &SystemStats{ps: &systemPS{}}
	})
}
