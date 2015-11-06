package system

import (
	"fmt"

	"github.com/influxdb/telegraf/plugins"
)

type DiskStats struct {
	ps PS

	Mountpoints []string
}

func (_ *DiskStats) Description() string {
	return "Read metrics about disk usage by mount point"
}

var diskSampleConfig = `
  # By default, telegraf gather stats for all mountpoints.
  # Setting mountpoints will restrict the stats to the specified mountpoints.
  # Mountpoints=["/"]
`

func (_ *DiskStats) SampleConfig() string {
	return diskSampleConfig
}

func (s *DiskStats) Gather(acc plugins.Accumulator) error {
	disks, err := s.ps.DiskUsage()
	if err != nil {
		return fmt.Errorf("error getting disk usage info: %s", err)
	}

	var restrictMpoints bool
	mPoints := make(map[string]bool)
	if len(s.Mountpoints) != 0 {
		restrictMpoints = true
		for _, mp := range s.Mountpoints {
			mPoints[mp] = true
		}
	}

	for _, du := range disks {
		_, member := mPoints[du.Path]
		if restrictMpoints && !member {
			continue
		}
		tags := map[string]string{
			"path":   du.Path,
			"fstype": du.Fstype,
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

type DiskIOStats struct {
	ps PS

	Devices          []string
	SkipSerialNumber bool
}

func (_ *DiskIOStats) Description() string {
	return "Read metrics about disk IO by device"
}

var diskIoSampleConfig = `
  # By default, telegraf will gather stats for all devices including 
  # disk partitions.
  # Setting devices will restrict the stats to the specified devcies.
  # Devices=["sda","sdb"]
  # Uncomment the following line if you do not need disk serial numbers.
  # SkipSerialNumber = true
`

func (_ *DiskIOStats) SampleConfig() string {
	return diskIoSampleConfig
}

func (s *DiskIOStats) Gather(acc plugins.Accumulator) error {
	diskio, err := s.ps.DiskIO()
	if err != nil {
		return fmt.Errorf("error getting disk io info: %s", err)
	}

	var restrictDevices bool
	devices := make(map[string]bool)
	if len(s.Devices) != 0 {
		restrictDevices = true
		for _, dev := range s.Devices {
			devices[dev] = true
		}
	}

	for _, io := range diskio {
		_, member := devices[io.Name]
		if restrictDevices && !member {
			continue
		}
		tags := map[string]string{}
		if len(io.Name) != 0 {
			tags["name"] = io.Name
		}
		if len(io.SerialNumber) != 0 && !s.SkipSerialNumber {
			tags["serial"] = io.SerialNumber
		}

		acc.Add("reads", io.ReadCount, tags)
		acc.Add("writes", io.WriteCount, tags)
		acc.Add("read_bytes", io.ReadBytes, tags)
		acc.Add("write_bytes", io.WriteBytes, tags)
		acc.Add("read_time", io.ReadTime, tags)
		acc.Add("write_time", io.WriteTime, tags)
		acc.Add("io_time", io.IoTime, tags)
	}

	return nil
}

func init() {
	plugins.Add("disk", func() plugins.Plugin {
		return &DiskStats{ps: &systemPS{}}
	})

	plugins.Add("io", func() plugins.Plugin {
		return &DiskIOStats{ps: &systemPS{}}
	})
}
