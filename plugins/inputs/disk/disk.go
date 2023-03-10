//go:generate ../../../tools/readme_config_includer/generator
package disk

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/inputs/system"
	"github.com/shirou/gopsutil/v3/disk"
)

//go:embed sample.conf
var sampleConfig string

type DiskStats struct {
	ps system.PS

	LegacyMountPoints []string `toml:"mountpoints" deprecated:"0.10.2;2.0.0;use 'mount_points' instead"`

	MountPoints     []string `toml:"mount_points"`
	IgnoreFS        []string `toml:"ignore_fs"`
	IgnoreMountOpts []string `toml:"ignore_mount_opts"`

	Log telegraf.Logger `toml:"-"`
}

func (*DiskStats) SampleConfig() string {
	return sampleConfig
}

func (ds *DiskStats) Init() error {
	// Legacy support:
	if len(ds.LegacyMountPoints) != 0 {
		ds.MountPoints = ds.LegacyMountPoints
	}

	ps := system.NewSystemPS()
	ps.Log = ds.Log
	ds.ps = ps

	return nil
}

func (ds *DiskStats) Gather(acc telegraf.Accumulator) error {
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
		mountOpts := MountOptions(partitions[i].Opts)
		tags := map[string]string{
			"path":   du.Path,
			"device": strings.ReplaceAll(device, "/dev/", ""),
			"fstype": du.Fstype,
			"mode":   mountOpts.Mode(),
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

		fields := map[string]interface{}{
			"total":        du.Total,
			"free":         du.Free,
			"used":         du.Used,
			"used_percent": usedPercent,
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

func init() {
	inputs.Add("disk", func() telegraf.Input {
		return &DiskStats{}
	})
}
