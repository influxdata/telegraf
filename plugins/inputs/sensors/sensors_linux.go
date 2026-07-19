//go:build linux

package sensors

import (
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
)

var (
	execCommand = exec.Command // execCommand is used to mock commands in tests.
	numberRegp  = regexp.MustCompile("[0-9]+")
)

const cmd = "sensors"

func (s *Sensors) Init() error {
	switch s.MetricVersion {
	case 0:
		s.MetricVersion = 1
	case 1, 2:
	default:
		return fmt.Errorf("invalid metric_version %d, please use 1 or 2", s.MetricVersion)
	}

	if s.path == "" {
		path, err := exec.LookPath(cmd)
		if err != nil {
			return fmt.Errorf("looking up %q failed: %w", cmd, err)
		}
		s.path = path
	}

	if s.path == "" {
		return fmt.Errorf("no path specified for %q", cmd)
	}

	return nil
}

func (s *Sensors) Gather(acc telegraf.Accumulator) error {
	if len(s.path) == 0 {
		return errors.New("sensors not found: verify that lm-sensors package is installed and that sensors is in your PATH")
	}

	return s.parse(acc)
}

// parse forks the command:
//
//	sensors -u -A
//
// and parses the output to add it to the telegraf.Accumulator.
func (s *Sensors) parse(acc telegraf.Accumulator) error {
	tags := make(map[string]string)
	fields := make(map[string]interface{})
	chip := ""
	cmd := execCommand(s.path, "-A", "-u")
	out, err := internal.StdOutputTimeout(cmd, time.Duration(s.Timeout))
	if err != nil {
		return fmt.Errorf("failed to run command %q: %w - %s", strings.Join(cmd.Args, " "), err, string(out))
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, line := range lines {
		if len(line) == 0 {
			acc.AddFields(measurement, fields, tags)
			chip = ""
			tags = make(map[string]string)
			fields = make(map[string]interface{})
			continue
		}
		if len(chip) == 0 {
			chip = line
			s.setLinuxDeviceTag(tags, chip)
			continue
		}
		if !strings.HasPrefix(line, "  ") {
			if len(tags) > 1 {
				acc.AddFields(measurement, fields, tags)
			}
			fields = make(map[string]interface{})
			tags = s.linuxTags(chip, strings.TrimRight(snake(line), ":"))
		} else {
			splitted := strings.Split(line, ":")
			fieldName := strings.TrimSpace(splitted[0])
			if s.RemoveNumbers {
				fieldName = numberRegp.ReplaceAllString(fieldName, "")
			}
			fieldValue, err := strconv.ParseFloat(strings.TrimSpace(splitted[1]), 64)
			if err != nil {
				return err
			}
			fields[fieldName] = fieldValue
		}
	}
	acc.AddFields(measurement, fields, tags)
	return nil
}

func (s *Sensors) setLinuxDeviceTag(tags map[string]string, chip string) {
	if s.MetricVersion == 2 {
		tags["device"] = chip
		return
	}
	tags["chip"] = chip
}

func (s *Sensors) linuxTags(chip, feature string) map[string]string {
	if s.MetricVersion == 2 {
		return map[string]string{
			"device": chip,
			"sensor": feature,
			"type":   strings.TrimRightFunc(feature, unicode.IsDigit),
		}
	}
	return map[string]string{
		"chip":    chip,
		"feature": feature,
	}
}

// snake converts string to snake case
func snake(input string) string {
	return strings.ToLower(strings.ReplaceAll(strings.TrimSpace(input), " ", "_"))
}
