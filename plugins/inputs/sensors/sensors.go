// +build linux

package sensors

import (
	"errors"
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

	// Main Regex, which is used to parse each line of the sensors output
	// SAMPLE: VCore:     +0.90 V  (min =  +0.60 V, max =  +1.49 V)
	readingsRe = regexp.MustCompile(
		`(.*)` + // Match the feature, group 1
			`:\s*.*?` + // Match everything after the reading, including + signs
			`([0-9.]+)` + // The actual reading, group 2
			`\s*.*?` + // Optional whitespace or C sign (°)
			`([a-zA-Z]+)` + // Unit of reading, could be C, V, RPM group 3
			`\s*` + // Optional whitespace
			`(\(.*?\))?`, // The options part after the reading, i.e (min = +0.60 V, max = +1.49 V), group 4
	)

	// Regex used for the options part, matches key/value paris
	// SAMPLE: min = +0.60 V
	optionsRe = regexp.MustCompile(
		`(` +
			`[^=\s\(\)]*?)` + // Option name, group 1
			`\s*=\s*.*?` + // Match everything until the value, including + signs
			`([0-9.]+)` + // Actual value of the option, group 2
			`.?` + // Whitespace or C sign ((°)
			`([a-zA-Z]*)`, // Unit of reading, could be C, V, RPM
	)
)

type Sensors struct {
	path string
}

func (*Sensors) Description() string {
	return "Monitor sensors, requires lm-sensors package"
}

func (*Sensors) SampleConfig() string {
	return ``
}

func (s *Sensors) Gather(acc telegraf.Accumulator) error {
	if len(s.path) == 0 {
		return errors.New("sensors not found: verify that lm-sensors package is installed and that sensors is in your PATH")
	}

	return s.parse(acc)
}

// parse forks the command:
//     sensors -A
// and parses the output to add it to the telegraf.Accumulator.
func (s *Sensors) parse(acc telegraf.Accumulator) error {
	tags := map[string]string{}
	fields := map[string]interface{}{}
	chip := ""

	cmd := execCommand(s.path, "-A")
	out, err := internal.CombinedOutputTimeout(cmd, time.Second*5)
	if err != nil {
		return fmt.Errorf("failed to run command %s: %s - %s", strings.Join(cmd.Args, " "), err, string(out))
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, line := range lines {
		// Next adapter reached
		if len(line) == 0 {
			chip = ""
			tags = map[string]string{}
			fields = map[string]interface{}{}
			continue
		}

		// First line is the chip line
		if len(chip) == 0 {
			chip = line
			tags["chip"] = chip
			continue
		}

		// Group 1 => feature
		// Group 2 => reading
		// Group 3 => unit
		// Group 4 => options
		parsed := readingsRe.FindAllStringSubmatch(line, -1)
		if parsed == nil {
			continue
		}

		fields = map[string]interface{}{}
		tags = map[string]string{
			"chip":    chip,
			"feature": strings.TrimSpace(parsed[0][1]),
		}

		fields["reading"], err = strconv.ParseFloat(parsed[0][2], 64)
		if err != nil {
			return err
		}

		fields["unit"] = parsed[0][3]

		if len(parsed[0][4]) > 0 {
			// Group 1 => option name
			// Group 2 => option value
			parsed = optionsRe.FindAllStringSubmatch(parsed[0][4], -1)
			for _, item := range parsed {
				fields[item[1]], err = strconv.ParseFloat(item[2], 64)
				if err != nil {
					return err
				}
			}
		}

		acc.AddFields("sensors", fields, tags)
	}

	acc.AddFields("sensors", fields, tags)
	return nil
}

func init() {
	s := Sensors{}
	path, _ := exec.LookPath("sensors")
	if len(path) > 0 {
		s.path = path
	}

	inputs.Add("sensors", func() telegraf.Input {
		return &s
	})
}
