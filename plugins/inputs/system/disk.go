package system

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type DiskStats struct {
	ps PS

	// Legacy support
	Mountpoints []string

	MountPoints []string
	IgnoreFS    []string `toml:"ignore_fs"`
}

func (_ *DiskStats) Description() string {
	return "Read metrics about disk usage by mount point"
}

var diskSampleConfig = `
  ## By default, telegraf gather stats for all mountpoints.
  ## Setting mountpoints will restrict the stats to the specified mountpoints.
  # mount_points = ["/"]

  ## Ignore some mountpoints by filesystem type. For example (dev)tmpfs (usually
  ## present on /run, /var/run, /dev/shm or /dev).
  ignore_fs = ["tmpfs", "devtmpfs", "devfs"]
`

func (_ *DiskStats) SampleConfig() string {
	return diskSampleConfig
}

func (s *DiskStats) Gather(acc telegraf.Accumulator) error {
	// Legacy support:
	if len(s.Mountpoints) != 0 {
		s.MountPoints = s.Mountpoints
	}

	disks, partitions, err := s.ps.DiskUsage(s.MountPoints, s.IgnoreFS)
	if err != nil {
		return fmt.Errorf("error getting disk usage info: %s", err)
	}

	for i, du := range disks {
		if du.Total == 0 {
			// Skip dummy filesystem (procfs, cgroupfs, ...)
			continue
		}
		tags := map[string]string{
			"path":   du.Path,
			"device": strings.Replace(partitions[i].Device, "/dev/", "", -1),
			"fstype": du.Fstype,
		}
		var used_percent float64
		if du.Used+du.Free > 0 {
			used_percent = float64(du.Used) /
				(float64(du.Used) + float64(du.Free)) * 100
		}

		var inodes_used_percent float64
		if du.InodesUsed+du.InodesFree > 0 {
			inodes_used_percent = float64(du.InodesUsed) / (float64(du.InodesUsed) + float64(du.InodesFree)) * 100
		}

		fields := map[string]interface{}{
			"total":               du.Total,
			"free":                du.Free,
			"used":                du.Used,
			"used_percent":        used_percent,
			"inodes_total":        du.InodesTotal,
			"inodes_free":         du.InodesFree,
			"inodes_used":         du.InodesUsed,
			"inodes_used_percent": inodes_used_percent,
		}
		acc.AddGauge("disk", fields, tags)
	}

	return nil
}

type DiskIOStats struct {
	ps PS

	Devices          []string
	DeviceTags       []string
	NameTemplates    []string
	SkipSerialNumber bool

	infoCache map[string]diskInfoCache
}

func (_ *DiskIOStats) Description() string {
	return "Read metrics about disk IO by device"
}

var diskIoSampleConfig = `
  ## By default, telegraf will gather stats for all devices including
  ## disk partitions.
  ## Setting devices will restrict the stats to the specified devices.
  # devices = ["sda", "sdb"]
  ## Uncomment the following line if you need disk serial numbers.
  # skip_serial_number = false
  #
  ## On systems which support it, device metadata can be added in the form of
  ## tags.
  ## Currently only Linux is supported via udev properties. You can view
  ## available properties for a device by running:
  ## 'udevadm info -q property -n /dev/sda'
  # device_tags = ["ID_FS_TYPE", "ID_FS_USAGE"]
  #
  ## Using the same metadata source as device_tags, you can also customize the
  ## name of the device via templates.
  ## The 'name_templates' parameter is a list of templates to try and apply to
  ## the device. The template may contain variables in the form of '$PROPERTY' or
  ## '${PROPERTY}'. The first template which does not contain any variables not
  ## present for the device is used as the device name tag.
  ## The typical use case is for LVM volumes, to get the VG/LV name instead of
  ## the near-meaningless DM-0 name.
  # name_templates = ["$ID_FS_LABEL","$DM_VG_NAME/$DM_LV_NAME"]
`

func (_ *DiskIOStats) SampleConfig() string {
	return diskIoSampleConfig
}

func (s *DiskIOStats) Gather(acc telegraf.Accumulator) error {
	diskio, err := s.ps.DiskIO(s.Devices)
	if err != nil {
		return fmt.Errorf("error getting disk io info: %s", err)
	}

	for _, io := range diskio {
		tags := map[string]string{}
		tags["name"] = s.diskName(io.Name)
		for t, v := range s.diskTags(io.Name) {
			tags[t] = v
		}
		if !s.SkipSerialNumber {
			if len(io.SerialNumber) != 0 {
				tags["serial"] = io.SerialNumber
			} else {
				tags["serial"] = "unknown"
			}
		}

		fields := map[string]interface{}{
			"reads":            io.ReadCount,
			"writes":           io.WriteCount,
			"read_bytes":       io.ReadBytes,
			"write_bytes":      io.WriteBytes,
			"read_time":        io.ReadTime,
			"write_time":       io.WriteTime,
			"io_time":          io.IoTime,
			"iops_in_progress": io.IopsInProgress,
		}
		acc.AddCounter("diskio", fields, tags)
	}

	return nil
}

var varRegex = regexp.MustCompile(`\$(?:\w+|\{\w+\})`)

func (s *DiskIOStats) diskName(devName string) string {
	di, err := s.diskInfo(devName)
	if err != nil {
		// discard error :-(
		// We can't return error because it's non-fatal to the Gather().
		// And we have no logger, so we can't log it.
		return devName
	}
	if di == nil {
		return devName
	}

	for _, nt := range s.NameTemplates {
		miss := false
		name := varRegex.ReplaceAllStringFunc(nt, func(sub string) string {
			sub = sub[1:] // strip leading '$'
			if sub[0] == '{' {
				sub = sub[1 : len(sub)-1] // strip leading & trailing '{' '}'
			}
			if v, ok := di[sub]; ok {
				return v
			}
			miss = true
			return ""
		})

		if !miss {
			return name
		}
	}

	return devName
}

func (s *DiskIOStats) diskTags(devName string) map[string]string {
	di, err := s.diskInfo(devName)
	if err != nil {
		// discard error :-(
		// We can't return error because it's non-fatal to the Gather().
		// And we have no logger, so we can't log it.
		return nil
	}
	if di == nil {
		return nil
	}

	tags := map[string]string{}
	for _, dt := range s.DeviceTags {
		if v, ok := di[dt]; ok {
			tags[dt] = v
		}
	}

	return tags
}

func init() {
	inputs.Add("disk", func() telegraf.Input {
		return &DiskStats{ps: &systemPS{}}
	})

	inputs.Add("diskio", func() telegraf.Input {
		return &DiskIOStats{ps: &systemPS{}, SkipSerialNumber: true}
	})
}
