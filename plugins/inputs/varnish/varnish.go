// +build !windows

package varnish

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type runner func(cmdName string, UseSudo bool, InstanceName string, Timeout internal.Duration) (*bytes.Buffer, error)

// Varnish is used to store configuration values
type Varnish struct {
	Stats        []string
	StatBinary   string
	AdmBinary    string
	UseSudo      bool
	InstanceName string
	Timeout      internal.Duration

	filter  filter.Filter
	statrun runner
	admrun  runner
}

var defaultStats = []string{"MAIN.cache_hit", "MAIN.cache_miss", "MAIN.uptime"}
var defaultStatBinary = "/usr/bin/varnishstat"
var defaultAdmBinary = "/usr/bin/varnishadm"
var defaultTimeout = internal.Duration{Duration: time.Second}

var sampleConfig = `
  ## If running as a restricted user you can prepend sudo for additional access:
  #use_sudo = false

  ## The default location of the varnishstat binary can be overridden with:
  varnishstat_binary = "/usr/bin/varnishstat"

  ## The default location of the varnishadm binary can be overridden with:
  varnishsadm_binary = "/usr/bin/varnishstat"

  ## By default, telegraf gather stats for 3 metric points.
  ## Setting stats will override the defaults shown below.
  ## Glob matching can be used, ie, stats = ["MAIN.*"]
  ## stats may also be set to ["*"], which will collect all stats
  stats = ["MAIN.cache_hit", "MAIN.cache_miss", "MAIN.uptime"]

  ## Optional name for the varnish instance (or working directory) to query
  ## Usually appened after -n in varnish cli
  # instance_name = instanceName

  ## Timeout for varnishstat command
  # timeout = "1s"
`

func (s *Varnish) Description() string {
	return "A plugin to collect stats from Varnish HTTP Cache"
}

// SampleConfig displays configuration instructions
func (s *Varnish) SampleConfig() string {
	return sampleConfig
}

// Shell out to varnish_adm and return the output
func varnishAdmRunner(cmdName string, UseSudo bool, InstanceName string, Timeout internal.Duration) (*bytes.Buffer, error) {
	cmdArgs := []string{"vcl.list"}

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

	err := internal.RunTimeout(cmd, Timeout.Duration)
	if err != nil {
		return &out, fmt.Errorf("error running varnishadm: %s", err)
	}

	return &out, nil
}

// Shell out to varnish_stat and return the output
func varnishStatRunner(cmdName string, UseSudo bool, InstanceName string, Timeout internal.Duration) (*bytes.Buffer, error) {
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

	err := internal.RunTimeout(cmd, Timeout.Duration)
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

	admout, err := s.admrun(s.AdmBinary, s.UseSudo, s.InstanceName, s.Timeout)
	if err != nil {
		return fmt.Errorf("error gathering vcl list from varnishadm: %s", err)
	}

	admscanner := bufio.NewScanner(admout)
	vclactive := ""
	for admscanner.Scan() {
		vcllist := strings.Fields(admscanner.Text())
		if len(vcllist) == 0 {
			continue
		}

		status := vcllist[0]
		vclname := vcllist[3]

		if status == "active" {
			vclactive = vclname
			continue
		}
	}

	statout, err := s.statrun(s.StatBinary, s.UseSudo, s.InstanceName, s.Timeout)
	if err != nil {
		return fmt.Errorf("error gathering metrics: %s", err)
	}

	sectionMap := make(map[string]map[string]interface{})
	scanner := bufio.NewScanner(statout)
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

		if strings.HasPrefix(stat, "VBE.") {
			if !strings.Contains(stat, vclactive) {
				continue
			}
			if strings.Contains(stat, vclactive) {
				stat = strings.Replace(stat, "."+vclactive, "", 1)
			}
			if strings.HasPrefix(stat, "VBE.goto") {
				re := regexp.MustCompile("^VBE\\.goto\\.[0-9a-f]+(.+)$")
				match := re.FindStringSubmatch(stat)
				stat = ("VBE.goto" + match[1])
			}
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
			"vcl":     vclactive,
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
			statrun:      varnishStatRunner,
			admrun:       varnishAdmRunner,
			Stats:        defaultStats,
			StatBinary:   defaultStatBinary,
			AdmBinary:    defaultAdmBinary,
			UseSudo:      false,
			InstanceName: "",
			Timeout:      defaultTimeout,
		}
	})
}
