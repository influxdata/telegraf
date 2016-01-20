package system

import (
	"fmt"

	"github.com/influxdata/telegraf/plugins/inputs"
)

type DiskStats struct {
	ps PS

	// Legacy support
	Mountpoints []string

	MountPoints []string
}

func (_ *DiskStats) Description() string {
	return "Read metrics about disk usage by mount point"
}

var diskSampleConfig = `
  # By default, telegraf gather stats for all mountpoints.
  # Setting mountpoints will restrict the stats to the specified mountpoints.
  # mount_points = ["/"]
`

func (_ *DiskStats) SampleConfig() string {
	return diskSampleConfig
}

func (s *DiskStats) Gather(acc inputs.Accumulator) error {
	// Legacy support:
	if len(s.Mountpoints) != 0 {
		s.MountPoints = s.Mountpoints
	}

	disks, err := s.ps.DiskUsage(s.MountPoints)
	if err != nil {
		return fmt.Errorf("error getting disk usage info: %s", err)
	}

	for _, du := range disks {
		tags := map[string]string{
			"path":   du.Path,
			"fstype": du.Fstype,
		}
		fields := map[string]interface{}{
			"total":        du.Total,
			"free":         du.Free,
			"used":         du.Total - du.Free,
			"inodes_total": du.InodesTotal,
			"inodes_free":  du.InodesFree,
			"inodes_used":  du.InodesTotal - du.InodesFree,
		}
		acc.AddFields("disk", fields, tags)
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
  # devices = ["sda","sdb"]
  # Uncomment the following line if you do not need disk serial numbers.
  # skip_serial_number = true
`

func (_ *DiskIOStats) SampleConfig() string {
	return diskIoSampleConfig
}

func (s *DiskIOStats) Gather(acc inputs.Accumulator) error {
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
		tags["name"] = io.Name
		if !s.SkipSerialNumber {
			if len(io.SerialNumber) != 0 {
				tags["serial"] = io.SerialNumber
			} else {
				tags["serial"] = "unknown"
			}
		}

		fields := map[string]interface{}{
			"reads":       io.ReadCount,
			"writes":      io.WriteCount,
			"read_bytes":  io.ReadBytes,
			"write_bytes": io.WriteBytes,
			"read_time":   io.ReadTime,
			"write_time":  io.WriteTime,
			"io_time":     io.IoTime,
		}
		acc.AddFields("diskio", fields, tags)
	}

	return nil
}

func init() {
	inputs.Add("disk", func() inputs.Input {
		return &DiskStats{ps: &systemPS{}}
	})

	inputs.Add("diskio", func() inputs.Input {
		return &DiskIOStats{ps: &systemPS{}}
	})
}
