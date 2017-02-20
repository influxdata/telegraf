package smart

import (
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

var (
	execCommand = exec.Command // execCommand is used to mock commands in tests.

	// Device Model:     APPLE SSD SM256E
	modelInInfo = regexp.MustCompile("^Device Model:\\s+(.*)$")
	// Serial Number:    S0X5NZBC422720
	serialInInfo = regexp.MustCompile("^Serial Number:\\s+(.*)$")
	// User Capacity:    251,000,193,024 bytes [251 GB]
	usercapacityInInfo = regexp.MustCompile("^User Capacity:\\s+([0-9,]+)\\s+bytes.*$")
	// SMART support is: Enabled
	smartEnabledInInfo = regexp.MustCompile("^SMART support is:\\s+(\\w+)$")
	// SMART overall-health self-assessment test result: PASSED
	smartOverallHealth = regexp.MustCompile("^SMART overall-health self-assessment test result:\\s+(\\w+).*$")

	// ID# ATTRIBUTE_NAME          FLAGS    VALUE WORST THRESH FAIL RAW_VALUE
	//   1 Raw_Read_Error_Rate     -O-RC-   200   200   000    -    0
	//   5 Reallocated_Sector_Ct   PO--CK   100   100   000    -    0
	// 192 Power-Off_Retract_Count -O--C-   097   097   000    -    14716
	attribute = regexp.MustCompile("^\\s*([0-9]+)\\s(\\S+)\\s+([-P][-O][-S][-R][-C][-K])\\s+([0-9]+)\\s+([0-9]+)\\s+([0-9]+)\\s+([-\\w]+)\\s+([\\w\\+\\.]+).*$")
)

type Smart struct {
	Path     string
	Nocheck  string
	Excludes []string
	Devices  []string
}

var sampleConfig = `
  ## Optionally specify the path to the smartctl executable
  # path = "/usr/bin/smartctl"
  #
  ## Skip checking disks in this power mode. Defaults to
  ## "standby" to not wake up disks that have stoped rotating.
  ## See --nockeck in the man pages for smartctl.
  ## smartctl version 5.41 and 5.42 have faulty detection of
  ## power mode and might require changing this value to
  ## "never" depending on your disks.
  # nocheck = "standby"
  #
  ## Optionally specify devices to exclude from reporting.
  # excludes = [ "/dev/pass6" ]
  #
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

	errs := m.getAttributes(acc, devices)
	if len(errs) > 0 {
		var errStrs []string
		for _, e := range errs {
			errStrs = append(errStrs, e.Error())
		}
		return errors.New(strings.Join(errStrs, ", "))
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
		dev := strings.Split(line, "#")
		if len(dev) > 1 && !excludedDev(m.Excludes, strings.TrimSpace(dev[0])) {
			devices = append(devices, strings.TrimSpace(dev[0]))
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
func (m *Smart) getAttributes(acc telegraf.Accumulator, devices []string) []error {

	errchan := make(chan error)
	for _, device := range devices {
		go gatherDisk(acc, m.Path, m.Nocheck, device, errchan)
	}

	var errors []error
	for i := 0; i < len(devices); i++ {
		err := <-errchan
		if err != nil {
			errors = append(errors, err)
		}
	}

	return errors
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

func gatherDisk(acc telegraf.Accumulator, path, nockeck, device string, err chan error) {

	// smartctl 5.41 & 5.42 have are broken regarding handling of --nocheck/-n
	args := []string{"--info", "--health", "--attributes", "--tolerance=verypermissive", "-n", nockeck, "--format=brief"}
	args = append(args, strings.Split(device, " ")...)
	cmd := execCommand(path, args...)
	out, e := internal.CombinedOutputTimeout(cmd, time.Second*5)
	outStr := string(out)

	// Ignore all exit statuses except if it is a command line parse error
	exitStatus, er := exitStatus(e)
	if er != nil {
		err <- fmt.Errorf("failed to run command %s: %s - %s", strings.Join(cmd.Args, " "), e, outStr)
		return
	}

	device_tags := map[string]string{}
	device_tags["device"] = strings.Split(device, " ")[0]
	device_fields := make(map[string]interface{})
	device_fields["exit_status"] = exitStatus

	for _, line := range strings.Split(outStr, "\n") {

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

		health := smartOverallHealth.FindStringSubmatch(line)
		if len(health) > 1 {
			device_tags["health"] = health[1]
		}

		attr := attribute.FindStringSubmatch(line)

		if len(attr) > 1 {
			tags := map[string]string{}
			fields := make(map[string]interface{})

			tags["device"] = strings.Split(device, " ")[0]
			tags["id"] = attr[1]
			tags["name"] = attr[2]
			tags["flags"] = attr[3]

			fields["exit_status"] = exitStatus
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

			acc.AddFields("smart_attribute", fields, tags)
		}
	}
	acc.AddFields("smart_device", device_fields, device_tags)

	err <- nil
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
	m.Nocheck = "standby"

	inputs.Add("smart", func() telegraf.Input {
		return &m
	})
}
