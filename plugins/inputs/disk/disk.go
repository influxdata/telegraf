package disk

import (
	"fmt"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/inputs/system"
)

type DiskStats struct {
	ps system.PS

	// Legacy support
	Mountpoints []string

	MountPoints []string
	IgnoreFS    []string `toml:"ignore_fs"`
}

func (_ *DiskStats) Description() string {
	return "Read metrics about disk usage by mount point"
}

var diskSampleConfig = `
  ## By default stats will be gathered for all mount points.
  ## Set mount_points will restrict the stats to only the specified mount points.
  # mount_points = ["/"]

  ## Ignore mount points by filesystem type.
  ignore_fs = ["tmpfs", "devtmpfs", "devfs", "overlay", "aufs", "squashfs"]
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
		mountOpts := parseOptions(partitions[i].Opts)
		tags := map[string]string{
			"path":   du.Path,
			"device": strings.Replace(partitions[i].Device, "/dev/", "", -1),
			"fstype": du.Fstype,
			"mode":   mountOpts.Mode(),
		}
		var used_percent float64
		if du.Used+du.Free > 0 {
			used_percent = float64(du.Used) /
				(float64(du.Used) + float64(du.Free)) * 100
		}

		fields := map[string]interface{}{
			"total":        du.Total,
			"free":         du.Free,
			"used":         du.Used,
			"used_percent": used_percent,
			"inodes_total": du.InodesTotal,
			"inodes_free":  du.InodesFree,
			"inodes_used":  du.InodesUsed,
		}
		acc.AddGauge("disk", fields, tags)
	}

	return nil
}

type MountOptions []string

func (opts MountOptions) Mode() string {
	if opts.exists("rw") {
		return "rw"
	} else if opts.exists("ro") {
		return "ro"
	} else {
		return "unknown"
	}
}

func (opts MountOptions) exists(opt string) bool {
	for _, o := range opts {
		if o == opt {
			return true
		}
	}
	return false
}

func parseOptions(opts string) MountOptions {
	return strings.Split(opts, ",")
}

func init() {
	ps := system.NewSystemPS()
	inputs.Add("disk", func() telegraf.Input {
		return &DiskStats{ps: ps}
	})
}
