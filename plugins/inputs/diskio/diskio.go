//go:generate ../../../tools/readme_config_includer/generator
package diskio

import (
	_ "embed"
	"fmt"
	"regexp"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/inputs/system"
)

//go:embed sample.conf
var sampleConfig string

var (
	varRegex = regexp.MustCompile(`\$(?:\w+|\{\w+\})`)
)

// hasMeta reports whether s contains any special glob characters.
func hasMeta(s string) bool {
	return strings.ContainsAny(s, "*?[")
}

type DiskIO struct {
	Devices          []string        `toml:"devices"`
	DeviceTags       []string        `toml:"device_tags"`
	NameTemplates    []string        `toml:"name_templates"`
	SkipSerialNumber bool            `toml:"skip_serial_number"`
	Log              telegraf.Logger `toml:"-"`

	ps           system.PS
	infoCache    map[string]diskInfoCache
	deviceFilter filter.Filter
	warnDiskName map[string]bool
	warnDiskTags map[string]bool
}

func (*DiskIO) SampleConfig() string {
	return sampleConfig
}

func (d *DiskIO) Init() error {
	for _, device := range d.Devices {
		if hasMeta(device) {
			deviceFilter, err := filter.Compile(d.Devices)
			if err != nil {
				return fmt.Errorf("error compiling device pattern: %w", err)
			}
			d.deviceFilter = deviceFilter
		}
	}

	d.infoCache = make(map[string]diskInfoCache)
	d.warnDiskName = make(map[string]bool)
	d.warnDiskTags = make(map[string]bool)

	return nil
}

func (d *DiskIO) Gather(acc telegraf.Accumulator) error {
	var devices []string
	if d.deviceFilter == nil {
		for _, dev := range d.Devices {
			devices = append(devices, resolveName(dev))
		}
	}

	diskio, err := d.ps.DiskIO(devices)
	if err != nil {
		return fmt.Errorf("error getting disk io info: %w", err)
	}

	for _, io := range diskio {
		match := false
		if d.deviceFilter != nil && d.deviceFilter.Match(io.Name) {
			match = true
		}

		tags := map[string]string{}
		var devLinks []string
		tags["name"], devLinks = d.diskName(io.Name)

		if wwid := getDeviceWWID(io.Name); wwid != "" {
			tags["wwid"] = wwid
		}

		if d.deviceFilter != nil && !match {
			for _, devLink := range devLinks {
				if d.deviceFilter.Match(devLink) {
					match = true
					break
				}
			}
			if !match {
				continue
			}
		}

		for t, v := range d.diskTags(io.Name) {
			tags[t] = v
		}

		if !d.SkipSerialNumber {
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
			"weighted_io_time": io.WeightedIO,
			"iops_in_progress": io.IopsInProgress,
			"merged_reads":     io.MergedReadCount,
			"merged_writes":    io.MergedWriteCount,
		}
		acc.AddCounter("diskio", fields, tags)
	}

	return nil
}

func (d *DiskIO) diskName(devName string) (string, []string) {
	di, err := d.diskInfo(devName)
	devLinks := strings.Split(di["DEVLINKS"], " ")
	for i, devLink := range devLinks {
		devLinks[i] = strings.TrimPrefix(devLink, "/dev/")
	}
	// Return error after attempting to process some of the devlinks.
	// These could exist if we got further along the diskInfo call.
	if err != nil {
		if ok := d.warnDiskName[devName]; !ok {
			d.warnDiskName[devName] = true
			d.Log.Warnf("Unable to gather disk name for %q: %s", devName, err)
		}
		return devName, devLinks
	}

	if len(d.NameTemplates) == 0 {
		return devName, devLinks
	}

	for _, nt := range d.NameTemplates {
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
			return name, devLinks
		}
	}

	return devName, devLinks
}

func (d *DiskIO) diskTags(devName string) map[string]string {
	if len(d.DeviceTags) == 0 {
		return nil
	}

	di, err := d.diskInfo(devName)
	if err != nil {
		if ok := d.warnDiskTags[devName]; !ok {
			d.warnDiskTags[devName] = true
			d.Log.Warnf("Unable to gather disk tags for %q: %s", devName, err)
		}
		return nil
	}

	tags := map[string]string{}
	for _, dt := range d.DeviceTags {
		if v, ok := di[dt]; ok {
			tags[dt] = v
		}
	}

	return tags
}

func init() {
	ps := system.NewSystemPS()
	inputs.Add("diskio", func() telegraf.Input {
		return &DiskIO{ps: ps, SkipSerialNumber: true}
	})
}
