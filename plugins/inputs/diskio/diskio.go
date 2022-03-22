package diskio

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/inputs/system"
)

var (
	varRegex = regexp.MustCompile(`\$(?:\w+|\{\w+\})`)
)

type DiskIO struct {
	ps system.PS

	Devices          []string
	DeviceTags       []string
	NameTemplates    []string
	SkipSerialNumber bool

	Log telegraf.Logger

	infoCache    map[string]diskInfoCache
	deviceFilter filter.Filter
	initialized  bool
}

func (d *DiskIO) Description() string {
	return "Read metrics about disk IO by device"
}

var diskIOsampleConfig = `
  ## By default, telegraf will gather stats for all devices including
  ## disk partitions.
  ## Setting devices will restrict the stats to the specified devices.
  # devices = ["sda", "sdb", "vd*"]
  ## Uncomment the following line if you need disk serial numbers.
  # skip_serial_number = false
  #
  ## On systems which support it, device metadata can be added in the form of
  ## tags.
  ## Currently only Linux is supported via udev properties. You can view
  ## available properties for a device by running:
  ## 'udevadm info -q property -n /dev/sda'
  ## Note: Most, but not all, udev properties can be accessed this way. Properties
  ## that are currently inaccessible include DEVTYPE, DEVNAME, and DEVPATH.
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

func (d *DiskIO) SampleConfig() string {
	return diskIOsampleConfig
}

// hasMeta reports whether s contains any special glob characters.
func hasMeta(s string) bool {
	return strings.ContainsAny(s, "*?[")
}

func (d *DiskIO) init() error {
	for _, device := range d.Devices {
		if hasMeta(device) {
			deviceFilter, err := filter.Compile(d.Devices)
			if err != nil {
				return fmt.Errorf("error compiling device pattern: %s", err.Error())
			}
			d.deviceFilter = deviceFilter
		}
	}
	d.initialized = true
	return nil
}

func (d *DiskIO) Gather(acc telegraf.Accumulator) error {
	if !d.initialized {
		err := d.init()
		if err != nil {
			return err
		}
	}

	devices := []string{}
	if d.deviceFilter == nil {
		devices = d.Devices
	}

	diskio, err := d.ps.DiskIO(devices)
	if err != nil {
		return fmt.Errorf("error getting disk io info: %s", err.Error())
	}

	for _, io := range diskio {
		match := false
		if d.deviceFilter != nil && d.deviceFilter.Match(io.Name) {
			match = true
		}

		tags := map[string]string{}
		var devLinks []string
		tags["name"], devLinks = d.diskName(io.Name)

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

	if len(d.NameTemplates) == 0 {
		return devName, devLinks
	}

	if err != nil {
		d.Log.Warnf("Error gathering disk info: %s", err)
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
		d.Log.Warnf("Error gathering disk info: %s", err)
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
	// Backwards compatible alias
	inputs.Add("io", func() telegraf.Input {
		return &DiskIO{ps: ps, SkipSerialNumber: true}
	})
}
