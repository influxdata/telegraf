// +build linux

package sysstat

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

var (
	firstTimestamp time.Time
	execCommand    = exec.Command // execCommand is used to mock commands in tests.
	dfltActivities = []string{"DISK"}
)

const parseInterval = 1 // parseInterval is the interval (in seconds) where the parsing of the binary file takes place.

type Sysstat struct {
	// Sadc represents the path to the sadc collector utility.
	Sadc string `toml:"sadc_path"`

	// Force the execution time of sadc
	SadcInterval internal.Duration `toml:"sadc_interval"`

	// Sadf represents the path to the sadf cmd.
	Sadf string `toml:"sadf_path"`

	// Activities is a list of activities that are passed as argument to the
	// collector utility (e.g: DISK, SNMP etc...)
	// The more activities that are added, the more data is collected.
	Activities []string

	// Options is a map of options.
	//
	// The key represents the actual option that the Sadf command is called with and
	// the value represents the description for that option.
	//
	// For example, if you have the following options map:
	//    map[string]string{"-C": "cpu", "-d": "disk"}
	// The Sadf command is run with the options -C and -d to extract cpu and
	// disk metrics from the collected binary file.
	//
	// If Group is false (see below), each metric will be prefixed with the corresponding description
	// and represents itself a measurement.
	//
	// If Group is true, metrics are grouped to a single measurement with the corresponding description as name.
	Options map[string]string

	// Group determines if metrics are grouped or not.
	Group bool

	// DeviceTags adds the possibility to add additional tags for devices.
	DeviceTags map[string][]map[string]string `toml:"device_tags"`
	tmpFile    string
	interval   int
}

func (*Sysstat) Description() string {
	return "Sysstat metrics collector"
}

var sampleConfig = `
  ## Path to the sadc command.
  #
  ## Common Defaults:
  ##   Debian/Ubuntu: /usr/lib/sysstat/sadc
  ##   Arch:          /usr/lib/sa/sadc
  ##   RHEL/CentOS:   /usr/lib64/sa/sadc
  sadc_path = "/usr/lib/sa/sadc" # required
  #
  #
  ## Path to the sadf command, if it is not in PATH
  # sadf_path = "/usr/bin/sadf"
  #
  #
  ## Activities is a list of activities, that are passed as argument to the
  ## sadc collector utility (e.g: DISK, SNMP etc...)
  ## The more activities that are added, the more data is collected.
  # activities = ["DISK"]
  #
  #
  ## Group metrics to measurements.
  ##
  ## If group is false each metric will be prefixed with a description
  ## and represents itself a measurement.
  ##
  ## If Group is true, corresponding metrics are grouped to a single measurement.
  # group = true
  #
  #
  ## Options for the sadf command. The values on the left represent the sadf
  ## options and the values on the right their description (which are used for
  ## grouping and prefixing metrics).
  ##
  ## Run 'sar -h' or 'man sar' to find out the supported options for your
  ## sysstat version.
  [inputs.sysstat.options]
    -C = "cpu"
    -B = "paging"
    -b = "io"
    -d = "disk"             # requires DISK activity
    "-n ALL" = "network"
    "-P ALL" = "per_cpu"
    -q = "queue"
    -R = "mem"
    -r = "mem_util"
    -S = "swap_util"
    -u = "cpu_util"
    -v = "inode"
    -W = "swap"
    -w = "task"
  #  -H = "hugepages"        # only available for newer linux distributions
  #  "-I ALL" = "interrupts" # requires INT activity
  #
  #
  ## Device tags can be used to add additional tags for devices.
  ## For example the configuration below adds a tag vg with value rootvg for
  ## all metrics with sda devices.
  # [[inputs.sysstat.device_tags.sda]]
  #  vg = "rootvg"
`

func (*Sysstat) SampleConfig() string {
	return sampleConfig
}

func (s *Sysstat) Gather(acc telegraf.Accumulator) error {
	if s.SadcInterval.Duration != 0 {
		// Collect interval is calculated as interval - parseInterval
		s.interval = int(s.SadcInterval.Duration.Seconds()) + parseInterval
	}

	if s.interval == 0 {
		if firstTimestamp.IsZero() {
			firstTimestamp = time.Now()
		} else {
			s.interval = int(time.Since(firstTimestamp).Seconds() + 0.5)
		}
	}
	ts := time.Now().Add(time.Duration(s.interval) * time.Second)
	if err := s.collect(); err != nil {
		return err
	}
	var wg sync.WaitGroup
	for option := range s.Options {
		wg.Add(1)
		go func(acc telegraf.Accumulator, option string) {
			defer wg.Done()
			acc.AddError(s.parse(acc, option, ts))
		}(acc, option)
	}
	wg.Wait()

	if _, err := os.Stat(s.tmpFile); err == nil {
		acc.AddError(os.Remove(s.tmpFile))
	}

	return nil
}

// collect collects sysstat data with the collector utility sadc.
// It runs the following command:
//     Sadc -S <Activity1> -S <Activity2> ... <collectInterval> 2 tmpFile
// The above command collects system metrics during <collectInterval> and
// saves it in binary form to tmpFile.
func (s *Sysstat) collect() error {
	options := []string{}
	for _, act := range s.Activities {
		options = append(options, "-S", act)
	}
	s.tmpFile = path.Join("/tmp", fmt.Sprintf("sysstat-%d", time.Now().Unix()))
	// collectInterval has to be smaller than the telegraf data collection interval
	collectInterval := s.interval - parseInterval

	// If true, interval is not defined yet and Gather is run for the first time.
	if collectInterval < 0 {
		collectInterval = 1 // In that case we only collect for 1 second.
	}

	options = append(options, strconv.Itoa(collectInterval), "2", s.tmpFile)
	cmd := execCommand(s.Sadc, options...)
	out, err := internal.CombinedOutputTimeout(cmd, time.Second*time.Duration(collectInterval+parseInterval))
	if err != nil {
		if err := os.Remove(s.tmpFile); err != nil {
			log.Printf("E! failed to remove tmp file after %s command: %s", strings.Join(cmd.Args, " "), err)
		}
		return fmt.Errorf("failed to run command %s: %s - %s", strings.Join(cmd.Args, " "), err, string(out))
	}
	return nil
}

func filterEnviron(env []string, prefix string) []string {
	newenv := env[:0]
	for _, envvar := range env {
		if !strings.HasPrefix(envvar, prefix) {
			newenv = append(newenv, envvar)
		}
	}
	return newenv
}

// Return the Cmd with its environment configured to use the C locale
func withCLocale(cmd *exec.Cmd) *exec.Cmd {
	var env []string
	if cmd.Env != nil {
		env = cmd.Env
	} else {
		env = os.Environ()
	}
	env = filterEnviron(env, "LANG")
	env = filterEnviron(env, "LC_")
	env = append(env, "LANG=C")
	cmd.Env = env
	return cmd
}

// parse runs Sadf on the previously saved tmpFile:
//    Sadf -p -- -p <option> tmpFile
// and parses the output to add it to the telegraf.Accumulator acc.
func (s *Sysstat) parse(acc telegraf.Accumulator, option string, ts time.Time) error {
	cmd := execCommand(s.Sadf, s.sadfOptions(option)...)
	cmd = withCLocale(cmd)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("running command '%s' failed: %s", strings.Join(cmd.Args, " "), err)
	}

	r := bufio.NewReader(stdout)
	csv := csv.NewReader(r)
	csv.Comma = '\t'
	csv.FieldsPerRecord = 6
	var measurement string
	// groupData to accumulate data when Group=true
	type groupData struct {
		tags   map[string]string
		fields map[string]interface{}
	}
	m := make(map[string]groupData)
	for {
		record, err := csv.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		device := record[3]
		value, err := strconv.ParseFloat(record[5], 64)
		if err != nil {
			return err
		}

		tags := map[string]string{}
		if device != "-" {
			tags["device"] = device
			if addTags, ok := s.DeviceTags[device]; ok {
				for _, tag := range addTags {
					for k, v := range tag {
						tags[k] = v
					}
				}

			}
		}

		if s.Group {
			measurement = s.Options[option]
			if _, ok := m[device]; !ok {
				m[device] = groupData{
					fields: make(map[string]interface{}),
					tags:   make(map[string]string),
				}
			}
			g, _ := m[device]
			if len(g.tags) == 0 {
				for k, v := range tags {
					g.tags[k] = v
				}
			}
			g.fields[escape(record[4])] = value
		} else {
			measurement = s.Options[option] + "_" + escape(record[4])
			fields := map[string]interface{}{
				"value": value,
			}
			acc.AddFields(measurement, fields, tags, ts)
		}

	}
	if s.Group {
		for _, v := range m {
			acc.AddFields(measurement, v.fields, v.tags, ts)
		}
	}
	if err := internal.WaitTimeout(cmd, time.Second*5); err != nil {
		return fmt.Errorf("command %s failed with %s",
			strings.Join(cmd.Args, " "), err)
	}
	return nil
}

// sadfOptions creates the correct options for the sadf utility.
func (s *Sysstat) sadfOptions(activityOption string) []string {
	options := []string{
		"-p",
		"--",
		"-p",
	}

	opts := strings.Split(activityOption, " ")
	options = append(options, opts...)
	options = append(options, s.tmpFile)

	return options
}

// escape removes % and / chars in field names
func escape(dirty string) string {
	var fieldEscaper = strings.NewReplacer(
		`%`, "pct_",
		`/`, "_per_",
	)
	return fieldEscaper.Replace(dirty)
}

func init() {
	s := Sysstat{
		Group:      true,
		Activities: dfltActivities,
	}
	sadf, _ := exec.LookPath("sadf")
	if len(sadf) > 0 {
		s.Sadf = sadf
	}
	inputs.Add("sysstat", func() telegraf.Input {
		return &s
	})
}
