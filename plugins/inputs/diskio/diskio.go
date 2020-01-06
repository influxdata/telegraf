package diskio

import (
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"

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

	last map[string]map[string]interface{}
}

func (_ *DiskIO) Description() string {
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

func (_ *DiskIO) SampleConfig() string {
	return diskIOsampleConfig
}

// hasMeta reports whether s contains any special glob characters.
func hasMeta(s string) bool {
	return strings.IndexAny(s, "*?[") >= 0
}

func (s *DiskIO) init() error {
	for _, device := range s.Devices {
		if hasMeta(device) {
			filter, err := filter.Compile(s.Devices)
			if err != nil {
				return fmt.Errorf("error compiling device pattern: %s", err.Error())
			}
			s.deviceFilter = filter
		}
	}

	s.last = make(map[string]map[string]interface{})
	s.initialized = true

	return nil
}

func (s *DiskIO) Gather(acc telegraf.Accumulator) error {
	if !s.initialized {
		err := s.init()
		if err != nil {
			return err
		}
	}

	devices := []string{}
	if s.deviceFilter == nil {
		devices = s.Devices
	}

	diskio, err := s.ps.DiskIO(devices)
	if err != nil {
		return fmt.Errorf("error getting disk io info: %s", err.Error())
	}

	now := time.Now()

	for _, io := range diskio {

		match := false
		if s.deviceFilter != nil && s.deviceFilter.Match(io.Name) {
			match = true
		}

		tags := map[string]string{}
		var devLinks []string
		tags["name"], devLinks = s.diskName(io.Name)

		if s.deviceFilter != nil && !match {
			for _, devLink := range devLinks {
				if s.deviceFilter.Match(devLink) {
					match = true
					break
				}
			}
			if !match {
				continue
			}
		}

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
			"weighted_io_time": io.WeightedIO,
			"iops_in_progress": io.IopsInProgress,
			"merged_reads":     io.MergedReadCount,
			"merged_writes":    io.MergedWriteCount,
			"read_time_avg":    float32(0),
			"write_time_avg":   float32(0),
		}

		_, ok := s.last[io.Name]
		if ok {
			last := s.last[io.Name]
			fields["read_time_avg"] = calcTimeAvg(
				fields["reads"].(uint64),
				last["reads"].(uint64),
				fields["read_time"].(uint64),
				last["read_time"].(uint64),
			)
			fields["write_time_avg"] = calcTimeAvg(
				fields["writes"].(uint64),
				last["writes"].(uint64),
				fields["write_time"].(uint64),
				last["write_time"].(uint64),
			)
		}

		acc.AddCounter("diskio", fields, tags, now)
		s.last[io.Name] = fields
	}

	return nil
}

func (s *DiskIO) diskName(devName string) (string, []string) {
	di, err := s.diskInfo(devName)
	devLinks := strings.Split(di["DEVLINKS"], " ")
	for i, devLink := range devLinks {
		devLinks[i] = strings.TrimPrefix(devLink, "/dev/")
	}

	if len(s.NameTemplates) == 0 {
		return devName, devLinks
	}

	if err != nil {
		s.Log.Warnf("Error gathering disk info: %s", err)
		return devName, devLinks
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
			return name, devLinks
		}
	}

	return devName, devLinks
}

func (s *DiskIO) diskTags(devName string) map[string]string {
	if len(s.DeviceTags) == 0 {
		return nil
	}

	di, err := s.diskInfo(devName)
	if err != nil {
		s.Log.Warnf("Error gathering disk info: %s", err)
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

func calcTimeAvg(ops uint64, opsLast uint64, time uint64, timeLast uint64) float32 {
	var opsDiff uint64
	var timeDiff uint64
	if ops > opsLast {
		opsDiff = ops - opsLast
	} else {
		opsDiff = 1 + ops + (math.MaxUint64 - opsLast)
	}
	if time > timeLast {
		timeDiff = time - timeLast
	} else {
		timeDiff = 1 + time + (math.MaxUint64 - timeLast)
	}
	if opsDiff == 0 || timeDiff == 0 {
		return 0
	}
	return float32(timeDiff) / float32(opsDiff)
}

func init() {
	ps := system.NewSystemPS()
	inputs.Add("diskio", func() telegraf.Input {
		return &DiskIO{ps: ps, SkipSerialNumber: true}
	})
}
