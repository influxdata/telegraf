package smart

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

var (
	execCommand = exec.Command // execCommand is used to mock commands in tests.

	deviceInScan = regexp.MustCompile("^(/dev/\\w*)\\s+.*")
	// Device Model:     APPLE SSD SM256E
	modelInInfo = regexp.MustCompile("^Device Model:\\s+(.*)$")
	// Serial Number:    S0X5NZBC422720
	serialInInfo = regexp.MustCompile("^Serial Number:\\s+(.*)$")
	// User Capacity:    251,000,193,024 bytes [251 GB]
	usercapacityInInfo = regexp.MustCompile("^User Capacity:\\s+([0-9,]+)\\s+bytes.*$")
	// SMART support is: Enabled
	smartEnabledInInfo = regexp.MustCompile("^SMART support is:\\s+(\\w+)$")

	// ID# ATTRIBUTE_NAME          FLAGS    VALUE WORST THRESH FAIL RAW_VALUE
	//   1 Raw_Read_Error_Rate     -O-RC-   200   200   000    -    0
	//   5 Reallocated_Sector_Ct   PO--CK   100   100   000    -    0
	// 192 Power-Off_Retract_Count -O--C-   097   097   000    -    14716
	attribute = regexp.MustCompile("^\\s*([0-9]+)\\s(\\S+)\\s+([-P][-O][-S][-R][-C][-K])\\s+([0-9]+)\\s+([0-9]+)\\s+([0-9]+)\\s+([-\\w]+)\\s+([\\w\\+\\.]+).*$")
)

type Smart struct {
	Path     string
	Excludes []string
	Devices  []string
}

var sampleConfig = `
  ## optionally specify the path to the smartctl executable
  # path = "/usr/bin/smartctl"
  #
  ## optionally specify devices to exclude from reporting.
  # excludes = [ "/dev/pass6" ]
  #
  ## optionally specify devices, if unset a scan (smartctl --scan)
  ## for S.M.A.R.T. devices will done and all found will be included.
  # devices = [ "/dev/ada0" ]
`

func (m *Smart) SampleConfig() string {
	return sampleConfig
}

func (m *Smart) Description() string {
	return "Read metrics from storage devices supporting S.M.A.R.T."
}

func (m *Smart) Gather(acc telegraf.Accumulator) error {
	fmt.Printf("Config: %v\n", m)
	if len(m.Path) == 0 {
		return fmt.Errorf("smartctl not found: verify that smartctl is installed and that smartctl is in your PATH")
	}

	devices := m.Devices
	if len(devices) == 0 {
		var err error
		devices, err = m.scan()
		if err != nil {
			return err
		}
	}

	err := m.getAttributes(acc, devices)
	if err != nil {
		return err
	}

	return nil
}

// Scan for S.M.A.R.T. devices
func (m *Smart) scan() ([]string, error) {

	cmd := execCommand(m.Path, "--scan")
	out, err := internal.CombinedOutputTimeout(cmd, time.Second*5)
	if err != nil {
		return []string{}, fmt.Errorf("failed to run command %s: %s - %s", strings.Join(cmd.Args, " "), err, string(out))
	}

	devices := []string{}
	for _, line := range strings.Split(string(out), "\n") {
		dev := deviceInScan.FindStringSubmatch(line)
		if len(dev) == 2 && !excludedDev(m.Excludes, dev[1]) {
			devices = append(devices, dev[1])
		}
	}
	return devices, nil
}

func excludedDev(excludes []string, device string) bool {
	fmt.Printf("DEBUG: %s in %v?\n", device, excludes)
	for _, exclude := range excludes {
		if device == exclude {
			fmt.Printf("DEBUG: filtered: %s\n", device)
			return true
		}
	}
	return false
}

// Get info and attributes for each S.M.A.R.T. device
func (m *Smart) getAttributes(acc telegraf.Accumulator, devices []string) error {

	for _, device := range devices {
		cmd := execCommand(m.Path, "--info", "--attributes", "--tolerance=verypermissive", "--nocheck=standby", "--format=brief", device)
		out, err := internal.CombinedOutputTimeout(cmd, time.Second*5)
		if err != nil {
			return fmt.Errorf("failed to run command %s: %s - %s", strings.Join(cmd.Args, " "), err, string(out))
		}

		device_tags := map[string]string{}
		device_tags["device"] = device

		for _, line := range strings.Split(string(out), "\n") {
			model := modelInInfo.FindStringSubmatch(line)
			if len(model) > 1 {
				device_tags["device_model"] = model[1]
			}

			serial := serialInInfo.FindStringSubmatch(line)
			if len(serial) > 1 {
				device_tags["serial_no"] = serial[1]
			}

			capacity := usercapacityInInfo.FindStringSubmatch(line)
			if len(capacity) > 1 {
				device_tags["capacity"] = strings.Replace(capacity[1], ",", "", -1)
			}

			enabled := smartEnabledInInfo.FindStringSubmatch(line)
			if len(enabled) > 1 {
				device_tags["enabled"] = enabled[1]
			}

			attr := attribute.FindStringSubmatch(line)

			if len(attr) > 1 {
				tags := map[string]string{}
				for k, v := range device_tags {
					tags[k] = v
				}
				fields := make(map[string]interface{})

				tags["id"] = attr[1]
				tags["name"] = attr[2]
				tags["flags"] = attr[3]

				if i, err := strconv.Atoi(attr[4]); err == nil {
					fields["value"] = i
				}
				if i, err := strconv.Atoi(attr[5]); err == nil {
					fields["worst"] = i
				}
				if i, err := strconv.Atoi(attr[6]); err == nil {
					fields["threshold"] = i
				}

				tags["fail"] = attr[7]
				if val, err := parseRawValue(attr[8]); err == nil {
					fields["raw_value"] = val
				}

				acc.AddFields("smart", fields, tags)
			}
		}
	}
	return nil
}

func parseRawValue(rawVal string) (int, error) {

	// Integer
	if i, err := strconv.Atoi(rawVal); err == nil {
		return i, nil
	}

	// Duration: 65h+33m+09.259s
	unit := regexp.MustCompile("^(.*)([hms])$")
	parts := strings.Split(rawVal, "+")
	if len(parts) == 0 {
		return 0, fmt.Errorf("Couldn't parse RAW_VALUE '%s'", rawVal)
	}

	duration := 0
	for _, part := range parts {
		timePart := unit.FindStringSubmatch(part)
		if len(timePart) == 0 {
			continue
		}
		switch timePart[2] {
		case "h":
			duration += atoi(timePart[1]) * 3600
		case "m":
			duration += atoi(timePart[1]) * 60
		case "s":
			// drop fractions of seconds
			duration += atoi(strings.Split(timePart[1], ".")[0])
		default:
			// Unknown, ignore
		}
	}
	return duration, nil
}

func atoi(str string) int {
	if i, err := strconv.Atoi(str); err == nil {
		return i
	}
	return 0
}

func init() {
	m := Smart{}
	path, _ := exec.LookPath("smartctl")
	if len(path) > 0 {
		m.Path = path
	}
	inputs.Add("smart", func() telegraf.Input {
		return &m
	})
}
