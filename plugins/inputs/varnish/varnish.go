// +build !windows

package varnish

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
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
	Binary       string
	UseSudo      bool
	InstanceName string
	Timeout      internal.Duration

	filter filter.Filter
	run    runner
}

var defaultStats = []string{"MAIN.cache_hit", "MAIN.cache_miss", "MAIN.uptime"}
var defaultBinary = "/usr/bin/varnishstat"
var defaultTimeout = internal.Duration{Duration: time.Second}

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

// Shell out to varnish_stat and return the output
func varnishRunner(cmdName string, UseSudo bool, InstanceName string, Timeout internal.Duration) (*bytes.Buffer, error) {
	// Enable JSON output of stats.
	cmdArgs := []string{"-j"}

	cmdArgs = append(cmdArgs, []string{"-j"}...)

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

	out, err := s.run(s.Binary, s.UseSudo, s.InstanceName, s.Timeout)
	if err != nil {
		return fmt.Errorf("error gathering metrics: %s", err)
	}

	sectionMap, err := parseVarnishJSON(out.Bytes(), s.filter)
	if err != nil {
		return fmt.Errorf("error parsing JSON output: %s", err)
	}

	for metricType, section := range sectionMap {
		for section, fields := range section {
			if len(fields) == 0 {
				continue
			}
			tags := map[string]string{
				"section": section,
			}
			switch metricType {
			case "c":
				acc.AddCounter("varnish", fields, tags)
			case "g":
				acc.AddGauge("varnish", fields, tags)
			default:
				acc.AddFields("varnish", fields, tags)
			}
		}
	}

	return nil
}

func parseVarnishJSON(input []byte, f filter.Filter) (map[string]map[string]map[string]interface{}, error) {
	sectionMap := make(map[string]map[string]map[string]interface{})
	varnishJSON := make(map[string]interface{})

	err := json.Unmarshal(input, &varnishJSON)
	if err != nil {
		return sectionMap, err
	}

	for stat, v := range varnishJSON {
		if stat == "timestamp" {
			continue
		}

		if f != nil && !f.Match(stat) {
			continue
		}

		m := v.(map[string]interface{})
		metricValue := uint64(m["value"].(float64))
		metricType := m["flag"].(string)

		parts := strings.SplitN(stat, ".", 2)
		section := parts[0]
		field := parts[1]

		// Init the metricType if necessary
		if _, ok := sectionMap[metricType]; !ok {
			sectionMap[metricType] = make(map[string]map[string]interface{})
		}
		// Init the section if necessary
		if _, ok := sectionMap[metricType][section]; !ok {
			sectionMap[metricType][section] = make(map[string]interface{})
		}

		sectionMap[metricType][section][field] = metricValue
	}

	return sectionMap, nil
}

func init() {
	inputs.Add("varnish", func() telegraf.Input {
		return &Varnish{
			run:          varnishRunner,
			Stats:        defaultStats,
			Binary:       defaultBinary,
			UseSudo:      false,
			InstanceName: "",
			Timeout:      defaultTimeout,
		}
	})
}
