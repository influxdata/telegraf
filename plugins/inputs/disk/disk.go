//go:generate ../../../tools/readme_config_includer/generator
package disk

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/shirou/gopsutil/v4/disk"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/inputs/system"
)

//go:embed sample.conf
var sampleConfig string

type Disk struct {
	MountPoints     []string        `toml:"mount_points"`
	IgnoreFS        []string        `toml:"ignore_fs"`
	IgnoreMountOpts []string        `toml:"ignore_mount_opts"`
	Log             telegraf.Logger `toml:"-"`

	ps system.PS
}

func (*Disk) SampleConfig() string {
	return sampleConfig
}

func (ds *Disk) Init() error {
	ps := system.NewSystemPS()
	ps.Log = ds.Log
	ds.ps = ps

	return nil
}

func (ds *Disk) Gather(acc telegraf.Accumulator) error {
	disks, partitions, err := ds.ps.DiskUsage(ds.MountPoints, ds.IgnoreMountOpts, ds.IgnoreFS)
	if err != nil {
		return fmt.Errorf("error getting disk usage info: %w", err)
	}
	for i, du := range disks {
		if du.Total == 0 {
			// Skip dummy filesystem (procfs, cgroupfs, ...)
			continue
		}

		device := partitions[i].Device
		mountOpts := mountOptions(partitions[i].Opts)
		tags := map[string]string{
			"path":   du.Path,
			"device": strings.ReplaceAll(device, "/dev/", ""),
			"fstype": du.Fstype,
			"mode":   mountOpts.mode(),
		}

		label, err := disk.Label(strings.TrimPrefix(device, "/dev/"))
		if err == nil && label != "" {
			tags["label"] = label
		}

		var usedPercent float64
		if du.Used+du.Free > 0 {
			usedPercent = float64(du.Used) /
				(float64(du.Used) + float64(du.Free)) * 100
		}

		var inodesUsedPercent float64
		if du.InodesUsed+du.InodesFree > 0 {
			inodesUsedPercent = float64(du.InodesUsed) /
				(float64(du.InodesUsed) + float64(du.InodesFree)) * 100
		}

		fields := map[string]interface{}{
			"total":               du.Total,
			"free":                du.Free,
			"used":                du.Used,
			"used_percent":        usedPercent,
			"inodes_total":        du.InodesTotal,
			"inodes_free":         du.InodesFree,
			"inodes_used":         du.InodesUsed,
			"inodes_used_percent": inodesUsedPercent,
		}
		acc.AddGauge("disk", fields, tags)
	}

	return nil
}

type mountOptions []string

func (opts mountOptions) mode() string {
	if opts.exists("rw") {
		return "rw"
	} else if opts.exists("ro") {
		return "ro"
	}
	return "unknown"
}

func (opts mountOptions) exists(opt string) bool {
	for _, o := range opts {
		if o == opt {
			return true
		}
	}
	return false
}

func init() {
	inputs.Add("disk", func() telegraf.Input {
		return &Disk{}
	})
}
