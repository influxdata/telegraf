//go:generate ../../../tools/readme_config_includer/generator
package disk

import (
	_ "embed"
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/shirou/gopsutil/v4/disk"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/common/psutil"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type Disk struct {
	MountPoints       []string        `toml:"mount_points"`
	IgnoreFS          []string        `toml:"ignore_fs"`
	IgnoreMountOpts   []string        `toml:"ignore_mount_opts"`
	ReportMountStatus bool            `toml:"report_mount_status"`
	FstabPath         string          `toml:"fstab_path"`
	Log               telegraf.Logger `toml:"-"`

	ps          psutil.PS
	fstabReader func() ([]fstabEntry, error)
}

func (*Disk) SampleConfig() string {
	return sampleConfig
}

func (ds *Disk) Init() error {
	ps := psutil.NewSystemPS()
	ps.Log = ds.Log
	ds.ps = ps

	return nil
}

func (ds *Disk) Gather(acc telegraf.Accumulator) error {
	disks, partitions, err := ds.ps.DiskUsage(ds.MountPoints, ds.IgnoreMountOpts, ds.IgnoreFS)
	if err != nil {
		return fmt.Errorf("error getting disk usage info: %w", err)
	}

	var expectedMounts []fstabEntry
	var mountedPaths map[string]struct{}
	if ds.ReportMountStatus {
		expectedMounts, err = ds.readExpectedMounts()
		if err != nil {
			return fmt.Errorf("error reading fstab: %w", err)
		}
		mountedPaths = make(map[string]struct{}, len(disks))
	}

	for i, du := range disks {
		if ds.ReportMountStatus {
			mountedPaths[normalizeMountPath(du.Path)] = struct{}{}
		}
		if du.Total == 0 {
			// Skip dummy filesystem (procfs, cgroupfs, ...)
			continue
		}

		device := partitions[i].Device
		if device == "none" || device == "" {
			device = partitions[i].Fstype
		}
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

		fields := map[string]any{
			"total":               du.Total,
			"free":                du.Free,
			"used":                du.Used,
			"used_percent":        usedPercent,
			"inodes_total":        du.InodesTotal,
			"inodes_free":         du.InodesFree,
			"inodes_used":         du.InodesUsed,
			"inodes_used_percent": inodesUsedPercent,
		}
		if ds.ReportMountStatus {
			fields["mount_status"] = "mounted"
		}
		acc.AddGauge("disk", fields, tags)
	}

	if ds.ReportMountStatus {
		ds.gatherUnmounted(acc, expectedMounts, mountedPaths)
	}

	return nil
}

func (ds *Disk) readExpectedMounts() ([]fstabEntry, error) {
	if ds.fstabReader != nil {
		return ds.fstabReader()
	}
	fstabPath := ds.FstabPath
	if fstabPath == "" {
		fstabPath = defaultFstabPath
	}
	return readFstabFile(fstabPath)
}

func (ds *Disk) gatherUnmounted(acc telegraf.Accumulator, expected []fstabEntry, mountedPaths map[string]struct{}) {
	mountPointFilter := make(map[string]struct{}, len(ds.MountPoints))
	for _, mp := range ds.MountPoints {
		mountPointFilter[normalizeMountPath(mp)] = struct{}{}
	}

	ignoreFS := make(map[string]struct{}, len(ds.IgnoreFS)+1)
	ignoreFS["autofs"] = struct{}{}
	for _, fs := range ds.IgnoreFS {
		ignoreFS[fs] = struct{}{}
	}

	ignoreOpts := make(map[string]struct{}, len(ds.IgnoreMountOpts))
	for _, opt := range ds.IgnoreMountOpts {
		ignoreOpts[opt] = struct{}{}
	}

	seen := make(map[string]struct{}, len(expected))
	for _, entry := range expected {
		if _, ok := seen[entry.mountpoint]; ok {
			continue
		}
		seen[entry.mountpoint] = struct{}{}

		if _, ok := mountedPaths[entry.mountpoint]; ok {
			continue
		}
		if len(mountPointFilter) > 0 {
			if _, ok := mountPointFilter[entry.mountpoint]; !ok {
				continue
			}
		}
		if _, ok := ignoreFS[entry.fstype]; ok {
			continue
		}
		if hasAnyOpt(entry.opts, ignoreOpts) {
			continue
		}

		device := entry.device
		if device == "none" || device == "" {
			device = entry.fstype
		}
		tags := map[string]string{
			"path":   filepath.FromSlash(entry.mountpoint),
			"device": strings.ReplaceAll(device, "/dev/", ""),
			"fstype": entry.fstype,
			"mode":   mountOptions(entry.opts).mode(),
		}
		fields := map[string]any{
			"mount_status": "unmounted",
		}
		acc.AddGauge("disk", fields, tags)
	}
}

func hasAnyOpt(opts []string, ignoreOpts map[string]struct{}) bool {
	if len(ignoreOpts) == 0 {
		return false
	}
	for _, opt := range opts {
		if _, ok := ignoreOpts[opt]; ok {
			return true
		}
	}
	return false
}

type mountOptions []string

func (opts mountOptions) mode() string {
	if slices.Contains(opts, "rw") {
		return "rw"
	} else if slices.Contains(opts, "ro") {
		return "ro"
	}
	return "unknown"
}

func init() {
	inputs.Add("disk", func() telegraf.Input {
		return &Disk{}
	})
}
