package smart

import (
	"bufio"
	"fmt"
	"log"
	"os/exec"
	"path"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

var (
	// Device Model:     APPLE SSD SM256E
	// Product:              HUH721212AL5204
	// Model Number: TS128GMTE850
	modelInfo = regexp.MustCompile("^(Device Model|Product|Model Number):\\s+(.*)$")
	// Serial Number:    S0X5NZBC422720
	serialInfo = regexp.MustCompile("^Serial Number:\\s+(.*)$")
	// LU WWN Device Id: 5 002538 655584d30
	wwnInfo = regexp.MustCompile("^LU WWN Device Id:\\s+(.*)$")
	// User Capacity:    251,000,193,024 bytes [251 GB]
	usercapacityInfo = regexp.MustCompile("^User Capacity:\\s+([0-9,]+)\\s+bytes.*$")
	// SMART support is: Enabled
	smartEnabledInfo = regexp.MustCompile("^SMART support is:\\s+(\\w+)$")
	// SMART overall-health self-assessment test result: PASSED
	// SMART Health Status: OK
	// PASSED, FAILED, UNKNOWN
	smartOverallHealth = regexp.MustCompile("^(SMART overall-health self-assessment test result|SMART Health Status):\\s+(\\w+).*$")

	// Accumulated start-stop cycles:  7
	sasStartStopAttr = regexp.MustCompile("^Accumulated start-stop cycles:\\s+(.*)$")
	// Accumulated load-unload cycles:  39
	sasLoadCycleAttr = regexp.MustCompile("^Accumulated load-unload cycles:\\s+(.*)$")
	// Current Drive Temperature:     34 C
	sasTempAttr = regexp.MustCompile("^Current Drive Temperature:\\s+(.*)\\s+C(.*)$")
	// Temperature: 38 Celsius
	nvmeTempAttr = regexp.MustCompile("^Temperature:\\s+(.*)\\s+(.*)$")
	// Power Cycles: 472
	nvmePowerCycleAttr = regexp.MustCompile("^Power Cycles:\\s+(.*)$")
	// Power On Hours: 6,038
	nvmePowerOnAttr = regexp.MustCompile("^Power On Hours:\\s+(.*)$")

	// ID# ATTRIBUTE_NAME          FLAGS    VALUE WORST THRESH FAIL RAW_VALUE
	//   1 Raw_Read_Error_Rate     -O-RC-   200   200   000    -    0
	//   5 Reallocated_Sector_Ct   PO--CK   100   100   000    -    0
	// 192 Power-Off_Retract_Count -O--C-   097   097   000    -    14716
	attribute = regexp.MustCompile("^\\s*([0-9]+)\\s(\\S+)\\s+([-P][-O][-S][-R][-C][-K])\\s+([0-9]+)\\s+([0-9]+)\\s+([0-9-]+)\\s+([-\\w]+)\\s+([\\w\\+\\.]+).*$")

	deviceFieldIds = map[string]string{
		"1":   "read_error_rate",
		"7":   "seek_error_rate",
		"190": "temp_c",
		"194": "temp_c",
		"199": "udma_crc_errors",
	}
)

type Smart struct {
	Path       string
	Nocheck    string
	Attributes bool
	Excludes   []string
	Devices    []string
	UseSudo    bool
}

var sampleConfig = `
  ## Optionally specify the path to the smartctl executable
  # path = "/usr/bin/smartctl"

  ## On most platforms smartctl requires root access.
  ## Setting 'use_sudo' to true will make use of sudo to run smartctl.
  ## Sudo must be configured to to allow the telegraf user to run smartctl
  ## without a password.
  # use_sudo = false

  ## Skip checking disks in this power mode. Defaults to
  ## "standby" to not wake up disks that have stoped rotating.
  ## See --nocheck in the man pages for smartctl.
  ## smartctl version 5.41 and 5.42 have faulty detection of
  ## power mode and might require changing this value to
  ## "never" depending on your disks.
  # nocheck = "standby"

  ## Gather detailed metrics for each SMART Attribute.
  # attributes = false

  ## Optionally specify devices to exclude from reporting.
  # excludes = [ "/dev/pass6" ]

  ## Optionally specify devices and device type, if unset
  ## a scan (smartctl --scan) for S.M.A.R.T. devices will
  ## done and all found will be included except for the
  ## excluded in excludes.
  # devices = [ "/dev/ada0 -d atacam" ]
`

func (m *Smart) SampleConfig() string {
	return sampleConfig
}

func (m *Smart) Description() string {
	return "Read metrics from storage devices supporting S.M.A.R.T."
}

func (m *Smart) Gather(acc telegraf.Accumulator) error {
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

	m.getAttributes(acc, devices)
	return nil
}

// Wrap with sudo
var runCmd = func(sudo bool, command string, args ...string) ([]byte, error) {
	cmd := exec.Command(command, args...)
	if sudo {
		cmd = exec.Command("sudo", append([]string{"-n", command}, args...)...)
	}
	return internal.CombinedOutputTimeout(cmd, time.Second*5)
}

// Scan for S.M.A.R.T. devices
func (m *Smart) scan() ([]string, error) {
	out, err := runCmd(m.UseSudo, m.Path, "--scan")
	if err != nil {
		return []string{}, fmt.Errorf("failed to run command '%s --scan': %s - %s", m.Path, err, string(out))
	}

	devices := []string{}
	for _, line := range strings.Split(string(out), "\n") {
		dev := strings.Split(line, " ")
		if len(dev) > 1 && !excludedDev(m.Excludes, strings.TrimSpace(dev[0])) {
			log.Printf("D! [inputs.smart] adding device: %+#v", dev)
			devices = append(devices, strings.TrimSpace(dev[0]))
		} else {
			log.Printf("D! [inputs.smart] skipping device: %+#v", dev)
		}
	}
	return devices, nil
}

func excludedDev(excludes []string, deviceLine string) bool {
	device := strings.Split(deviceLine, " ")
	if len(device) != 0 {
		for _, exclude := range excludes {
			if device[0] == exclude {
				return true
			}
		}
	}
	return false
}

// Get info and attributes for each S.M.A.R.T. device
func (m *Smart) getAttributes(acc telegraf.Accumulator, devices []string) {
	var wg sync.WaitGroup
	wg.Add(len(devices))

	for _, device := range devices {
		go gatherDisk(acc, m.UseSudo, m.Attributes, m.Path, m.Nocheck, device, &wg)
	}

	wg.Wait()
}

// Command line parse errors are denoted by the exit code having the 0 bit set.
// All other errors are drive/communication errors and should be ignored.
func exitStatus(err error) (int, error) {
	if exiterr, ok := err.(*exec.ExitError); ok {
		if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
			return status.ExitStatus(), nil
		}
	}
	return 0, err
}

func gatherDisk(acc telegraf.Accumulator, usesudo, collectAttributes bool, smartctl, nocheck, device string, wg *sync.WaitGroup) {
	defer wg.Done()
	// smartctl 5.41 & 5.42 have are broken regarding handling of --nocheck/-n
	args := []string{"--info", "--health", "--attributes", "--tolerance=verypermissive", "-n", nocheck, "--format=brief"}
	args = append(args, strings.Split(device, " ")...)
	out, e := runCmd(usesudo, smartctl, args...)
	outStr := string(out)

	// Ignore all exit statuses except if it is a command line parse error
	exitStatus, er := exitStatus(e)
	if er != nil {
		acc.AddError(fmt.Errorf("failed to run command '%s %s': %s - %s", smartctl, strings.Join(args, " "), e, outStr))
		return
	}

	deviceTags := map[string]string{}
	deviceNode := strings.Split(device, " ")[0]
	deviceTags["device"] = path.Base(deviceNode)
	deviceFields := make(map[string]interface{})
	deviceFields["exit_status"] = exitStatus

	scanner := bufio.NewScanner(strings.NewReader(outStr))

	for scanner.Scan() {
		line := scanner.Text()

		model := modelInfo.FindStringSubmatch(line)
		if len(model) > 2 {
			deviceTags["model"] = model[2]
		}

		serial := serialInfo.FindStringSubmatch(line)
		if len(serial) > 1 {
			deviceTags["serial_no"] = serial[1]
		}

		wwn := wwnInfo.FindStringSubmatch(line)
		if len(wwn) > 1 {
			deviceTags["wwn"] = strings.Replace(wwn[1], " ", "", -1)
		}

		capacity := usercapacityInfo.FindStringSubmatch(line)
		if len(capacity) > 1 {
			deviceTags["capacity"] = strings.Replace(capacity[1], ",", "", -1)
		}

		enabled := smartEnabledInfo.FindStringSubmatch(line)
		if len(enabled) > 1 {
			deviceTags["enabled"] = enabled[1]
		}

		health := smartOverallHealth.FindStringSubmatch(line)
		if len(health) > 2 {
			deviceFields["health_ok"] = (health[2] == "PASSED" || health[2] == "OK")
		}

		tags := map[string]string{}
		fields := make(map[string]interface{})

		if collectAttributes {
			deviceNode := strings.Split(device, " ")[0]
			tags["device"] = path.Base(deviceNode)

			if serial, ok := deviceTags["serial_no"]; ok {
				tags["serial_no"] = serial
			}
			if wwn, ok := deviceTags["wwn"]; ok {
				tags["wwn"] = wwn
			}
		}

		attr := attribute.FindStringSubmatch(line)
		if len(attr) > 1 {
			if collectAttributes {
				tags["id"] = attr[1]
				tags["name"] = attr[2]
				tags["flags"] = attr[3]

				fields["exit_status"] = exitStatus
				if i, err := strconv.ParseInt(attr[4], 10, 64); err == nil {
					fields["value"] = i
				}
				if i, err := strconv.ParseInt(attr[5], 10, 64); err == nil {
					fields["worst"] = i
				}
				if i, err := strconv.ParseInt(attr[6], 10, 64); err == nil {
					fields["threshold"] = i
				}

				tags["fail"] = attr[7]
				if val, err := parseRawValue(attr[8]); err == nil {
					fields["raw_value"] = val
				}

				acc.AddFields("smart_attribute", fields, tags)
			}

			// If the attribute matches on the one in deviceFieldIds
			// save the raw value to a field.
			if field, ok := deviceFieldIds[attr[1]]; ok {
				if val, err := parseRawValue(attr[8]); err == nil {
					deviceFields[field] = val
				}
			}
		} else {
			if collectAttributes {
				if startStop := sasStartStopAttr.FindStringSubmatch(line); len(startStop) > 1 {
					tags["id"] = "4"
					tags["name"] = "Start_Stop_Count"
					i, err := strconv.ParseInt(strings.Replace(startStop[1], ",", "", -1), 10, 64)
					if err != nil {
						continue
					}
					fields["raw_value"] = i

					acc.AddFields("smart_attribute", fields, tags)
					continue
				}

				if powerCycle := nvmePowerCycleAttr.FindStringSubmatch(line); len(powerCycle) > 1 {
					tags["id"] = "12"
					tags["name"] = "Power_Cycle_Count"
					i, err := strconv.ParseInt(strings.Replace(powerCycle[1], ",", "", -1), 10, 64)
					if err != nil {
						continue
					}
					fields["raw_value"] = i

					acc.AddFields("smart_attribute", fields, tags)
					continue
				}

				if powerOn := nvmePowerOnAttr.FindStringSubmatch(line); len(powerOn) > 1 {
					tags["id"] = "9"
					tags["name"] = "Power_On_Hours"
					i, err := strconv.ParseInt(strings.Replace(powerOn[1], ",", "", -1), 10, 64)
					if err != nil {
						continue
					}
					fields["raw_value"] = i

					acc.AddFields("smart_attribute", fields, tags)
					continue
				}

				if loadCycle := sasLoadCycleAttr.FindStringSubmatch(line); len(loadCycle) > 1 {
					tags["id"] = "193"
					tags["name"] = "Load_Cycle_Count"
					i, err := strconv.ParseInt(strings.Replace(loadCycle[1], ",", "", -1), 10, 64)
					if err != nil {
						continue
					}
					fields["raw_value"] = i

					acc.AddFields("smart_attribute", fields, tags)
					continue
				}

				if temp := sasTempAttr.FindStringSubmatch(line); len(temp) > 1 {
					tags["id"] = "194"
					tags["name"] = "Temperature_Celsius"
					tempC, err := strconv.ParseInt(temp[1], 10, 64)
					if err != nil {
						continue
					}
					fields["raw_value"] = tempC
					deviceFields["temp_c"] = tempC

					acc.AddFields("smart_attribute", fields, tags)
				}

				if temp := nvmeTempAttr.FindStringSubmatch(line); len(temp) > 1 {
					tags["id"] = "194"
					tags["name"] = "Temperature_Celsius"
					tempC, err := strconv.ParseInt(temp[1], 10, 64)
					if err != nil {
						continue
					}
					fields["raw_value"] = tempC
					deviceFields["temp_c"] = tempC

					acc.AddFields("smart_attribute", fields, tags)
				}
			}
		}
	}
	acc.AddFields("smart_device", deviceFields, deviceTags)
}

func parseRawValue(rawVal string) (int64, error) {
	// Integer
	if i, err := strconv.ParseInt(rawVal, 10, 64); err == nil {
		return i, nil
	}

	// Duration: 65h+33m+09.259s
	unit := regexp.MustCompile("^(.*)([hms])$")
	parts := strings.Split(rawVal, "+")
	if len(parts) == 0 {
		return 0, fmt.Errorf("Couldn't parse RAW_VALUE '%s'", rawVal)
	}

	duration := int64(0)
	for _, part := range parts {
		timePart := unit.FindStringSubmatch(part)
		if len(timePart) == 0 {
			continue
		}
		switch timePart[2] {
		case "h":
			duration += parseInt(timePart[1]) * int64(3600)
		case "m":
			duration += parseInt(timePart[1]) * int64(60)
		case "s":
			// drop fractions of seconds
			duration += parseInt(strings.Split(timePart[1], ".")[0])
		default:
			// Unknown, ignore
		}
	}
	return duration, nil
}

func parseInt(str string) int64 {
	if i, err := strconv.ParseInt(str, 10, 64); err == nil {
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
	m.Nocheck = "standby"

	inputs.Add("smart", func() telegraf.Input {
		return &m
	})
}
