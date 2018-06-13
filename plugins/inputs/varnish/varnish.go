// +build !windows

package varnish

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

type runner func(cmdName string, UseSudo bool, InstanceName string) (*bytes.Buffer, error)

// Varnish is used to store configuration values
type Varnish struct {
	Stats        []string
	Binary       string
	UseSudo      bool
	InstanceName string

	filter filter.Filter
	run    runner
}

var defaultStats = []string{"MAIN.cache_hit", "MAIN.cache_miss", "MAIN.uptime"}
var defaultBinary = "/usr/bin/varnishstat"

var sampleConfig = `
  ## If running as a restricted user you can prepend sudo for additional access:
  #use_sudo = false

  ## The default location of the varnishstat binary can be overridden with:
  binary = "/usr/bin/varnishstat"

  ## By default, telegraf gather stats for 3 metric points.
  ## Setting stats will override the defaults shown below.
  ## Glob matching can be used, ie, stats = ["MAIN.*"]
  ## stats may also be set to ["*"], which will collect all stats
  stats = ["MAIN.cache_hit", "MAIN.cache_miss", "MAIN.uptime"]

  ## Optional name for the varnish instance (or working directory) to query
  ## Usually appened after -n in varnish cli
  # instance_name = instanceName
`

func (s *Varnish) Description() string {
	return "A plugin to collect stats from Varnish HTTP Cache"
}

// SampleConfig displays configuration instructions
func (s *Varnish) SampleConfig() string {
	return sampleConfig
}

// Shell out to varnish_stat and return the output
func varnishRunner(cmdName string, UseSudo bool, InstanceName string) (*bytes.Buffer, error) {
	cmdArgs := []string{"-1"}

	if InstanceName != "" {
		cmdArgs = append(cmdArgs, []string{"-n", InstanceName}...)
	}

	cmd := exec.Command(cmdName, cmdArgs...)

	if UseSudo {
		cmdArgs = append([]string{cmdName}, cmdArgs...)
		cmdArgs = append([]string{"-n"}, cmdArgs...)
		cmd = exec.Command("sudo", cmdArgs...)
	}

	var out bytes.Buffer
	cmd.Stdout = &out
	err := internal.RunTimeout(cmd, time.Millisecond*200)
	if err != nil {
		return &out, fmt.Errorf("error running varnishstat: %s", err)
	}

	return &out, nil
}

// Gather collects the configured stats from varnish_stat and adds them to the
// Accumulator
//
// The prefix of each stat (eg MAIN, MEMPOOL, LCK, etc) will be used as a
// 'section' tag and all stats that share that prefix will be reported as fields
// with that tag
func (s *Varnish) Gather(acc telegraf.Accumulator) error {
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

	out, err := s.run(s.Binary, s.UseSudo, s.InstanceName)
	if err != nil {
		return fmt.Errorf("error gathering metrics: %s", err)
	}

	sectionMap := make(map[string]map[string]interface{})
	scanner := bufio.NewScanner(out)
	for scanner.Scan() {
		cols := strings.Fields(scanner.Text())
		if len(cols) < 2 {
			continue
		}
		if !strings.Contains(cols[0], ".") {
			continue
		}

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
			acc.AddError(fmt.Errorf("Expected a numeric value for %s = %v\n",
				stat, value))
		}
	}

	for section, fields := range sectionMap {
		tags := map[string]string{
			"section": section,
		}
		if len(fields) == 0 {
			continue
		}

		acc.AddFields("varnish", fields, tags)
	}

	return nil
}

func init() {
	inputs.Add("varnish", func() telegraf.Input {
		return &Varnish{
			run:          varnishRunner,
			Stats:        defaultStats,
			Binary:       defaultBinary,
			UseSudo:      false,
			InstanceName: "",
		}
	})
}
