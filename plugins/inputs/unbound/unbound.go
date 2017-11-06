// +build !windows

package unbound

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type runner func(cmdName string, UseSudo bool) (*bytes.Buffer, error)

// Unbound is used to store configuration values
type Unbound struct {
	Stats   []string
	Binary  string
	UseSudo bool

	filter filter.Filter
	run    runner
}

var defaultStats = []string{"total.*", "num.*", "time.up", "mem.*"}
var defaultBinary = "/usr/sbin/unbound-control"

var sampleConfig = `
  ## If running as a restricted user you can prepend sudo for additional access:
  #use_sudo = false

  ## The default location of the unbound-control binary can be overridden with:
  binary = "/usr/sbin/unbound-control"

  ## By default, telegraf gather stats for 3 metric points.
  ## Setting stats will override the defaults shown below.
  ## Glob matching can be used, ie, stats = ["total.*"]
  ## stats may also be set to ["*"], which will collect all stats
  stats = ["total.*", "num.*","time.up", "mem.*"]
`

func (s *Unbound) Description() string {
	return "A plugin to collect stats from Unbound - a validating, recursive, and caching DNS resolver "
}

// SampleConfig displays configuration instructions
func (s *Unbound) SampleConfig() string {
	return sampleConfig
}

// Shell out to unbound_stat and return the output
func unboundRunner(cmdName string, UseSudo bool) (*bytes.Buffer, error) {
	cmdArgs := []string{"stats"}

	cmd := exec.Command(cmdName, cmdArgs...)

	if UseSudo {
		cmdArgs = append([]string{cmdName}, cmdArgs...)
		cmd = exec.Command("sudo", cmdArgs...)
	}

	var out bytes.Buffer
	cmd.Stdout = &out
	err := internal.RunTimeout(cmd, time.Millisecond*200)
	if err != nil {
		return &out, fmt.Errorf("error running unbound-control: %s", err)
	}

	return &out, nil
}

// Gather collects the configured stats from unbound_stat and adds them to the
// Accumulator
//
// The prefix of each stat (eg MAIN, MEMPOOL, LCK, etc) will be used as a
// 'section' tag and all stats that share that prefix will be reported as fields
// with that tag
func (s *Unbound) Gather(acc telegraf.Accumulator) error {
	if s.filter == nil {
		var err error
		if len(s.Stats) == 0 {
			s.filter, err = filter.Compile(defaultStats)
		} else {
			// legacy support, change "all" -> "*":
			if s.Stats[0] == "all" {
				s.Stats[0] = "*"
			}
			s.filter, err = filter.Compile(s.Stats)
		}
		if err != nil {
			return err
		}
	}

	out, err := s.run(s.Binary, s.UseSudo)
	if err != nil {
		return fmt.Errorf("error gathering metrics: %s", err)
	}

	sectionMap := make(map[string]map[string]interface{})
	scanner := bufio.NewScanner(out)
	for scanner.Scan() {

		cols := strings.Split(scanner.Text(), "=")

		stat := cols[0]
		value := cols[1]

		if s.filter != nil && !s.filter.Match(stat) {
			continue
		}

		parts := strings.SplitN(stat, ".", 2)

		section := parts[0]
		field := parts[1]

		// Init the section if necessary
		if _, ok := sectionMap[section]; !ok {
			sectionMap[section] = make(map[string]interface{})
		}

		sectionMap[section][field], err = strconv.ParseUint(value, 10, 64)
		if err != nil {
			sectionMap[section][field], err = strconv.ParseFloat(value, 64)
			if err != nil {
				acc.AddError(fmt.Errorf("Expected a numeric or a float value for %s = %v\n",
					stat, value))
			}
		}

	}

	for section, fields := range sectionMap {
		tags := map[string]string{
			"section": section,
		}
		if len(fields) == 0 {
			continue
		}
		acc.AddFields("unbound", fields, tags)
	}

	return nil
}

func init() {
	inputs.Add("unbound", func() telegraf.Input {
		return &Unbound{
			run:     unboundRunner,
			Stats:   defaultStats,
			Binary:  defaultBinary,
			UseSudo: false,
		}
	})
}
