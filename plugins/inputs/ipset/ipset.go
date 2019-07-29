package ipset

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// Ipsets is a telegraf plugin to gather packets and bytes counters from ipset
type Ipset struct {
	IncludeUnmatchedSets bool
	UseSudo              bool
	Timeout              internal.Duration
	lister               setLister
}

type setLister func(Timeout internal.Duration, UseSudo bool) (*bytes.Buffer, error)

const measurement = "ipset"

var defaultTimeout = internal.Duration{Duration: time.Second}

// Description returns a short description of the plugin
func (ipset *Ipset) Description() string {
	return "Gather packets and bytes counters from Linux ipsets"
}

// SampleConfig returns sample configuration options.
func (ipset *Ipset) SampleConfig() string {
	return `
  ## By default, we only show sets which have already matched at least 1 packet.
  ## set include_unmatched_sets = true to gather them all.
  include_unmatched_sets = false
  ## Adjust your sudo settings appropriately if using this option ("sudo ipset save")
  use_sudo = false
  ## The default timeout of 1s for ipset execution can be overridden here:
  # timeout = "1s"
`
}

func (ips *Ipset) Gather(acc telegraf.Accumulator) error {
	out, e := ips.lister(ips.Timeout, ips.UseSudo)
	if e != nil {
		acc.AddError(e)
	}

	scanner := bufio.NewScanner(out)
	for scanner.Scan() {
		line := scanner.Text()
		// Ignore sets created without the "counters" option
		nocomment := strings.Split(line, "\"")[0]
		if !(strings.Contains(nocomment, "packets") &&
			strings.Contains(nocomment, "bytes")) {
			continue
		}

		data := strings.Fields(line)
		if len(data) < 7 {
			acc.AddError(fmt.Errorf("Error parsing line (expected at least 7 fields): %s", line))
			continue
		}
		if data[0] == "add" && (data[4] != "0" || ips.IncludeUnmatchedSets) {
			tags := map[string]string{
				"set":  data[1],
				"rule": data[2],
			}
			packets_total, err := strconv.ParseUint(data[4], 10, 64)
			if err != nil {
				acc.AddError(err)
			}
			bytes_total, err := strconv.ParseUint(data[6], 10, 64)
			if err != nil {
				acc.AddError(err)
			}
			fields := map[string]interface{}{
				"packets_total": packets_total,
				"bytes_total":   bytes_total,
			}
			acc.AddCounter(measurement, fields, tags)
		}
	}
	return nil
}

func setList(Timeout internal.Duration, UseSudo bool) (*bytes.Buffer, error) {
	// Is ipset installed ?
	ipsetPath, err := exec.LookPath("ipset")
	if err != nil {
		return nil, err
	}
	var args []string
	cmdName := ipsetPath
	if UseSudo {
		cmdName = "sudo"
		args = append(args, ipsetPath)
	}
	args = append(args, "save")

	cmd := exec.Command(cmdName, args...)

	var out bytes.Buffer
	cmd.Stdout = &out
	err = internal.RunTimeout(cmd, Timeout.Duration)
	if err != nil {
		return &out, fmt.Errorf("error running ipset save: %s", err)
	}

	return &out, nil
}

func init() {
	inputs.Add("ipset", func() telegraf.Input {
		return &Ipset{
			lister:  setList,
			Timeout: defaultTimeout,
		}
	})
}
