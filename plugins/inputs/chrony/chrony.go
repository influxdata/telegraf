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
	// Used to mock commands in tests
	execCommand  = exec.Command
	execLookPath = exec.LookPath
)

type Chrony struct {
	DNSLookup bool     `toml:"dns_lookup"`
	UseSudo   bool     `toml:"use_sudo"`
	Binary    string   `toml:"binary"`
	Commands  []string `toml:"commands"`

	path      string
	arguments []string
}

var defaultBinary = "chronyc"
var defaultCommands = []string{"tracking"}

func (*Chrony) Description() string {
	return "Get standard chrony metrics, requires chronyc executable."
}

func (*Chrony) SampleConfig() string {
	return `
  ## If true, chronyc tries to perform a DNS lookup for the time server.
  # dns_lookup = false

  ## Run chronyc binary with sudo.
  ## Required if chronyd cmdport is disabled, or when running the "serverstats" command.
  ## Sudo must be configured to allow the telegraf user to run chronyc without a password.
  ## That configuration may look something like:
  ## telegraf ALL=(ALL:ALL) NOPASSWD:/usr/bin/chronyc
  # use_sudo = false

  ## Location of the chronyc binary, if not in PATH
  # binary = "chronyc"

  ## "tracking" is the default. Some commands with useful fields are "serverstats", "smoothing", "rtcdata", and "ntpdata".
  ## To run "ntpdata" command for a specific source, append the Name or IP of the source after a space, eg. "ntpdata 10.1.2.3"
  ## To run "ntpdata" command multiple times for different sources, use multiple instances of this input plugin to allow for adding different tags.
  # commands = ["tracking"]
  `
}

func (c *Chrony) Init() error {
	var err error
	c.path, err = execLookPath(c.Binary)
	if err != nil {
		return errors.New("chronyc binary not found: verify that chrony is installed and that chronyc is in your PATH, or that the binary option contains a valid path.")
	}

	if c.UseSudo {
		// Change the exec path to the path to sudo, and add the path to chronyc as the first argument

		c.arguments = append(c.arguments, c.path)

		c.path, err = execLookPath("sudo")
		if err != nil {
			return errors.New("sudo not found: verify that sudo is installed and in your PATH")
		}
	}

	// -m is required to pass multiple commands
	c.arguments = append(c.arguments, "-m")

	if !c.DNSLookup {
		c.arguments = append(c.arguments, "-n")
	}

	if len(c.Commands) == 0 {
		return errors.New("No commands specified - try the default [\"tracking\"]")
	}

	c.arguments = append(c.arguments, c.Commands...)

	return nil
}

func (c *Chrony) Gather(acc telegraf.Accumulator) error {
	cmd := execCommand(c.path, c.arguments...)
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
		if len(strings.TrimSpace(line)) == 0 {
			// When running multiple commands, there are sometimes empty lines between output
			continue
		}
		stats := strings.Split(line, ":")
		if len(stats) < 2 {
			return nil, nil, fmt.Errorf("unexpected output from chronyc, expected ':' in %s", out)
		}
		name := strings.ToLower(strings.Replace(strings.TrimSpace(stats[0]), " ", "_", -1))
		// ignore reference time - different keys for "tracking" vs "ntpdata" commands
		if strings.Contains(name, "ref_time") || strings.Contains(name, "reference_time") {
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
	inputs.Add("chrony", func() telegraf.Input {
		return &Chrony{
			Binary:   defaultBinary,
			Commands: defaultCommands,
		}
	})
}
