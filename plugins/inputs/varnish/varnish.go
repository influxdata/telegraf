// +build !windows

package varnish

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
	"time"
)

const (
	kwAll = "all"
)

// Varnish is used to store configuration values
type Varnish struct {
	Stats  []string
	Binary string
}

var defaultStats = []string{"MAIN.cache_hit", "MAIN.cache_miss", "MAIN.uptime"}
var defaultBinary = "/usr/bin/varnishstat"

var varnishSampleConfig = `
  ## The default location of the varnishstat binary can be overridden with:
  binary = "/usr/bin/varnishstat"

  ## By default, telegraf gather stats for 3 metric points.
  ## Setting stats will override the defaults shown below.
  ## stats may also be set to ["all"], which will collect all stats
  stats = ["MAIN.cache_hit", "MAIN.cache_miss", "MAIN.uptime"]
`

func (s *Varnish) Description() string {
	return "A plugin to collect stats from Varnish HTTP Cache"
}

// SampleConfig displays configuration instructions
func (s *Varnish) SampleConfig() string {
	return fmt.Sprintf(varnishSampleConfig, strings.Join(defaultStats, "\",\""))
}

func (s *Varnish) setDefaults() {
	if len(s.Stats) == 0 {
		s.Stats = defaultStats
	}

	if s.Binary == "" {
		s.Binary = defaultBinary
	}
}

// Builds a filter function that will indicate whether a given stat should
// be reported
func (s *Varnish) statsFilter() func(string) bool {
	s.setDefaults()

	// Build a set for constant-time lookup of whether stats should be included
	filter := make(map[string]struct{})
	for _, s := range s.Stats {
		filter[s] = struct{}{}
	}

	// Create a function that respects the kwAll by always returning true
	// if it is set
	return func(stat string) bool {
		if s.Stats[0] == kwAll {
			return true
		}

		_, found := filter[stat]
		return found
	}
}

// Shell out to varnish_stat and return the output
var varnishStat = func(cmdName string) (*bytes.Buffer, error) {
	cmdArgs := []string{"-1"}

	cmd := exec.Command(cmdName, cmdArgs...)
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
	s.setDefaults()
	out, err := varnishStat(s.Binary)
	if err != nil {
		return fmt.Errorf("error gathering metrics: %s", err)
	}

	statsFilter := s.statsFilter()
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

		if !statsFilter(stat) {
			continue
		}

		parts := strings.SplitN(stat, ".", 2)
		section := parts[0]
		field := parts[1]

		// Init the section if necessary
		if _, ok := sectionMap[section]; !ok {
			sectionMap[section] = make(map[string]interface{})
		}

		sectionMap[section][field], err = strconv.Atoi(value)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Expected a numeric value for %s = %v\n",
				stat, value)
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
	inputs.Add("varnish", func() telegraf.Input { return &Varnish{} })
}
