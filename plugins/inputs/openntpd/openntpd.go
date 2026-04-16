//go:generate ../../../tools/readme_config_includer/generator
package openntpd

import (
	"bufio"
	"bytes"
	_ "embed"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

var (
	defaultBinary  = "/usr/sbin/ntpctl"
	defaultTimeout = config.Duration(5 * time.Second)

	// Peer column index mappings
	peerTagI = map[string]int{
		"stratum": 2,
	}
	peerFloatI = map[string]int{
		"offset": 5,
		"delay":  6,
		"jitter": 7,
	}
	peerIntI = map[string]int{
		"wt":   0,
		"tl":   1,
		"next": 3,
		"poll": 4,
	}

	// Sensor column index mappings
	sensorIntI = map[string]int{
		"wt":   0,
		"gd":   1,
		"st":   2,
		"next": 3,
		"poll": 4,
	}
	sensorFloatI = map[string]int{
		"offset":     5,
		"correction": 6,
	}
)

type Openntpd struct {
	Binary  string          `toml:"binary"`
	Timeout config.Duration `toml:"timeout"`
	UseSudo bool            `toml:"use_sudo"`

	run runner
}

type runner func(cmdName string, timeout config.Duration, useSudo bool) (*bytes.Buffer, error)

func (*Openntpd) SampleConfig() string {
	return sampleConfig
}

func (n *Openntpd) Gather(acc telegraf.Accumulator) error {
	out, err := n.run(n.Binary, n.Timeout, n.UseSudo)
	if err != nil {
		return fmt.Errorf("error gathering metrics: %w", err)
	}

	// section tracks which part of the output we are in:
	// "" = before first section header (status line area)
	// "peer" = inside peer section
	// "sensor" = inside sensor section
	section := ""
	skipNext := false

	scanner := bufio.NewScanner(out)
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			continue
		}

		if trimmed == "peer" {
			section = "peer"
			skipNext = true
			continue
		}

		if trimmed == "sensor" {
			section = "sensor"
			skipNext = true
			continue
		}

		// Skip the column-header line that follows each section header
		if skipNext {
			skipNext = false
			continue
		}

		switch section {
		case "":
			parseStatusLine(trimmed, acc)
		case "peer":
			parsePeer(scanner, line, acc)
		case "sensor":
			parseSensor(scanner, line, acc)
		}
	}

	return nil
}

// parseStatusLine parses the summary line produced by `ntpctl -s all`, e.g.:
//
//	12/12 peers valid, 1/1 sensors valid, constraint offset -1s, clock synced, stratum 1
func parseStatusLine(line string, acc telegraf.Accumulator) {
	fields := make(map[string]interface{}, 7)

	parts := strings.Split(line, ", ")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		switch {
		case strings.HasSuffix(part, "peers valid"):
			fraction := strings.Fields(part)[0]
			if pv, pt, ok := parseFraction(fraction); ok {
				fields["peers_valid"] = pv
				fields["peers_total"] = pt
			} else {
				acc.AddError(fmt.Errorf("parsing peers fraction failed for %q", fraction))
			}
		case strings.HasSuffix(part, "sensors valid"):
			fraction := strings.Fields(part)[0]
			if sv, st, ok := parseFraction(fraction); ok {
				fields["sensors_valid"] = sv
				fields["sensors_total"] = st
			} else {
				acc.AddError(fmt.Errorf("parsing sensors fraction failed for %q", fraction))
			}
		case strings.HasPrefix(part, "constraint offset "):
			val := strings.TrimPrefix(part, "constraint offset ")
			val = strings.TrimSuffix(val, "s")
			if v, err := strconv.ParseInt(val, 10, 64); err == nil {
				fields["constraint_offset_s"] = v
			} else {
				acc.AddError(fmt.Errorf("integer value expected for constraint offset %q", val))
			}
		case part == "clock synced":
			fields["clock_synced"] = int64(1)
		case strings.HasPrefix(part, "stratum "):
			val := strings.TrimPrefix(part, "stratum ")
			if v, err := strconv.ParseInt(val, 10, 64); err == nil {
				fields["stratum"] = v
			} else {
				acc.AddError(fmt.Errorf("integer value expected for stratum %q", val))
			}
		}
	}

	// Make sure we always provide the clock_synced field
	if _, found := fields["clock_synced"]; !found {
		fields["clock_synced"] = int64(0)
	}

	acc.AddFields("openntpd_status", fields, nil)
}

// parseFraction splits "12/12" into (numerator, denominator, ok).
func parseFraction(s string) (num, den int64, ok bool) {
	n, d, found := strings.Cut(s, "/")
	if !found {
		return 0, 0, false
	}
	num, err1 := strconv.ParseInt(n, 10, 64)
	den, err2 := strconv.ParseInt(d, 10, 64)
	if err1 != nil || err2 != nil {
		return 0, 0, false
	}
	return num, den, true
}

// parsePeer parses a two-line peer entry. headerLine is the first line
// (peer address / DNS name); the second line is read from scanner.
func parsePeer(scanner *bufio.Scanner, headerLine string, acc telegraf.Accumulator) {
	headerFields := strings.Fields(headerLine)
	if len(headerFields) == 0 {
		return
	}

	mFields := make(map[string]interface{}, 7)
	tags := make(map[string]string, 3)

	// DNS resolution error → keep DNS name as remote
	if headerFields[0] != "not" {
		tags["remote"] = headerFields[0]
	} else {
		tags["remote"] = headerFields[len(headerFields)-1]
	}

	if !scanner.Scan() {
		return
	}
	statsLine := scanner.Text()
	statsFields := strings.Fields(statsLine)
	if len(statsFields) == 0 {
		return
	}

	// Optional state prefix (e.g. "*")
	if strings.Contains(statsFields[0], "*") {
		tags["state_prefix"] = statsFields[0]
		statsFields = statsFields[1:]
	}

	for key, index := range peerTagI {
		if index < len(statsFields) {
			tags[key] = statsFields[index]
		}
	}

	for key, index := range peerIntI {
		if index >= len(statsFields) || statsFields[index] == "-" {
			continue
		}
		raw := statsFields[index]
		if key == "next" || key == "poll" {
			raw = strings.TrimSuffix(raw, "s")
		}
		v, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			acc.AddError(fmt.Errorf("value %q is not an integer", statsFields[index]))
			continue
		}
		mFields[key] = v
	}

	for key, index := range peerFloatI {
		if index >= len(statsFields) {
			continue
		}
		raw := statsFields[index]
		if raw == "-" || raw == "----" || raw == "peer" ||
			raw == "not" || raw == "valid" {
			continue
		}
		if key == "offset" || key == "delay" || key == "jitter" {
			raw = strings.TrimSuffix(raw, "ms")
		}
		v, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			acc.AddError(fmt.Errorf("value %q is not a float", statsFields[index]))
			continue
		}
		mFields[key] = v
	}

	acc.AddFields("openntpd", mFields, tags)
}

// parseSensor parses a two-line sensor entry. headerLine is the first line
// (sensor name and refid); the second line is read from scanner.
func parseSensor(scanner *bufio.Scanner, headerLine string, acc telegraf.Accumulator) {
	headerFields := strings.Fields(headerLine)
	if len(headerFields) == 0 {
		return
	}

	tags := make(map[string]string, 3)
	tags["sensor"] = headerFields[0]
	if len(headerFields) >= 2 {
		tags["refid"] = headerFields[1]
	}

	if !scanner.Scan() {
		return
	}
	statsLine := scanner.Text()
	statsFields := strings.Fields(statsLine)
	if len(statsFields) == 0 {
		return
	}

	// Optional state prefix (e.g. "*")
	if strings.Contains(statsFields[0], "*") {
		tags["state_prefix"] = statsFields[0]
		statsFields = statsFields[1:]
	}

	mFields := make(map[string]interface{}, 7)

	for key, index := range sensorIntI {
		if index >= len(statsFields) || statsFields[index] == "-" {
			continue
		}
		raw := statsFields[index]
		if key == "next" || key == "poll" {
			raw = strings.TrimSuffix(raw, "s")
		}
		v, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			acc.AddError(fmt.Errorf("value %q is not an integer", statsFields[index]))
			continue
		}
		mFields[key] = v
	}

	for key, index := range sensorFloatI {
		if index >= len(statsFields) || statsFields[index] == "-" {
			continue
		}
		raw := strings.TrimSuffix(statsFields[index], "ms")
		v, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			acc.AddError(fmt.Errorf("value %q is not a float", statsFields[index]))
			continue
		}
		mFields[key] = v
	}

	acc.AddFields("openntpd_sensors", mFields, tags)
}

// Shell out to ntpctl and return the output
func openntpdRunner(cmdName string, timeout config.Duration, useSudo bool) (*bytes.Buffer, error) {
	cmdArgs := []string{"-s", "all"}

	cmd := exec.Command(cmdName, cmdArgs...)

	if useSudo {
		cmdArgs = append([]string{cmdName}, cmdArgs...)
		cmd = exec.Command("sudo", cmdArgs...)
	}

	var out bytes.Buffer
	cmd.Stdout = &out
	err := internal.RunTimeout(cmd, time.Duration(timeout))
	if err != nil {
		return &out, fmt.Errorf("error running ntpctl: %w", err)
	}

	return &out, nil
}

func init() {
	inputs.Add("openntpd", func() telegraf.Input {
		return &Openntpd{
			run:     openntpdRunner,
			Binary:  defaultBinary,
			Timeout: defaultTimeout,
			UseSudo: false,
		}
	})
}
