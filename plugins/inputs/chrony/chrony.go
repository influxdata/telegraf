package chrony

import (
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

var (
	execCommand = exec.Command // execCommand is used to mock commands in tests.
)

type Chrony struct {
	DNSLookup bool `toml:"dns_lookup"`
	path      string
}

func (*Chrony) Description() string {
	return "Get standard chrony metrics, requires chronyc executable."
}

func (*Chrony) SampleConfig() string {
	return `
  ## If true, chronyc tries to perform a DNS lookup for the time server.
  # dns_lookup = false
  `
}

func (c *Chrony) Gather(acc telegraf.Accumulator) error {
	if len(c.path) == 0 {
		return errors.New("chronyc not found: verify that chrony is installed and that chronyc is in your PATH")
	}

	flags := []string{}
	if !c.DNSLookup {
		flags = append(flags, "-n")
	}
	flags = append(flags, "tracking")

	cmd := execCommand(c.path, flags...)
	out, err := internal.CombinedOutputTimeout(cmd, time.Second*5)
	if err != nil {
		return fmt.Errorf("failed to run command %s: %s - %s", strings.Join(cmd.Args, " "), err, string(out))
	}
	fields, tags, err := processChronycOutput(string(out))
	if err != nil {
		return err
	}
	acc.AddFields("chrony", fields, tags)
	return nil
}

// processChronycOutput takes in a string output from the chronyc command, like:
//
//     Reference ID    : 192.168.1.22 (ntp.example.com)
//     Stratum         : 3
//     Ref time (UTC)  : Thu May 12 14:27:07 2016
//     System time     : 0.000020390 seconds fast of NTP time
//     Last offset     : +0.000012651 seconds
//     RMS offset      : 0.000025577 seconds
//     Frequency       : 16.001 ppm slow
//     Residual freq   : -0.000 ppm
//     Skew            : 0.006 ppm
//     Root delay      : 0.001655 seconds
//     Root dispersion : 0.003307 seconds
//     Update interval : 507.2 seconds
//     Leap status     : Normal
//
// The value on the left side of the colon is used as field name, if the first field on
// the right side is a float. If it cannot be parsed as float, it is a tag name.
//
// Ref time is ignored and all names are converted to snake case.
//
// It returns (<fields>, <tags>)
func processChronycOutput(out string) (map[string]interface{}, map[string]string, error) {
	tags := map[string]string{}
	fields := map[string]interface{}{}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	for _, line := range lines {
		stats := strings.Split(line, ":")
		if len(stats) < 2 {
			return nil, nil, fmt.Errorf("unexpected output from chronyc, expected ':' in %s", out)
		}
		name := strings.ToLower(strings.Replace(strings.TrimSpace(stats[0]), " ", "_", -1))
		// ignore reference time
		if strings.Contains(name, "ref_time") {
			continue
		}
		valueFields := strings.Fields(stats[1])
		if len(valueFields) == 0 {
			return nil, nil, fmt.Errorf("unexpected output from chronyc: %s", out)
		}
		if strings.Contains(strings.ToLower(name), "stratum") {
			tags["stratum"] = valueFields[0]
			continue
		}
		if strings.Contains(strings.ToLower(name), "reference_id") {
			tags["reference_id"] = valueFields[0]
			continue
		}
		value, err := strconv.ParseFloat(valueFields[0], 64)
		if err != nil {
			tags[name] = strings.ToLower(strings.Join(valueFields, " "))
			continue
		}
		if strings.Contains(stats[1], "slow") {
			value = -value
		}
		fields[name] = value
	}

	return fields, tags, nil
}

func init() {
	c := Chrony{}
	path, _ := exec.LookPath("chronyc")
	if len(path) > 0 {
		c.path = path
	}
	inputs.Add("chrony", func() telegraf.Input {
		return &c
	})
}
