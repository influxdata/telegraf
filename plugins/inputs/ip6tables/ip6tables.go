// +build linux

package ip6tables

import (
	"errors"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// Iptables is a telegraf plugin to gather packets and bytes throughput from Linux's ip6tables packet filter.
type Iptables struct {
	UseSudo bool
	UseLock bool
	Table   string
	Chains  []string
	lister  chainLister
}

// Description returns a short description of the plugin.
func (ipt *Iptables) Description() string {
	return "Gather packets and bytes throughput from ip6tables"
}

// SampleConfig returns sample configuration options.
func (ipt *Iptables) SampleConfig() string {
	return `
  ## ip6tables require root access on most systems.
  ## Setting 'use_sudo' to true will make use of sudo to run ip6tables.
  ## Users must configure sudo to allow telegraf user to run ip6tables with no password.
  ## ip6tables can be restricted to only list command "ip6tables -nvL".
  use_sudo = false
  ## Setting 'use_lock' to true runs ip6tables with the "-w" option.
  ## Adjust your sudo settings appropriately if using this option ("ip6tables -wnvl")
  use_lock = false
  ## defines the table to monitor:
  table = "filter"
  ## defines the chains to monitor.
  ## NOTE: ip6tables rules without a comment will not be monitored.
  ## Read the plugin documentation for more information.
  chains = [ "INPUT" ]
`
}

// Gather gathers ip6tables packets and bytes throughput from the configured tables and chains.
func (ipt *Iptables) Gather(acc telegraf.Accumulator) error {
	if ipt.Table == "" || len(ipt.Chains) == 0 {
		return nil
	}
	// best effort : we continue through the chains even if an error is encountered,
	// but we keep track of the last error.
	for _, chain := range ipt.Chains {
		data, e := ipt.lister(ipt.Table, chain)
		if e != nil {
			acc.AddError(e)
			continue
		}
		e = ipt.parseAndGather(data, acc)
		if e != nil {
			acc.AddError(e)
			continue
		}
	}
	return nil
}

func (ipt *Iptables) chainList(table, chain string) (string, error) {
	iptablePath, err := exec.LookPath("ip6tables")
	if err != nil {
		return "", err
	}
	var args []string
	name := iptablePath
	if ipt.UseSudo {
		name = "sudo"
		args = append(args, iptablePath)
	}
	ip6tablesBaseArgs := "-nvL"
	if ipt.UseLock {
		ip6tablesBaseArgs = "-wnvL"
	}
	args = append(args, ip6tablesBaseArgs, chain, "-t", table, "-x")
	c := exec.Command(name, args...)
	out, err := c.Output()
	return string(out), err
}

const measurement = "ip6tables"

var errParse = errors.New("Cannot parse ip6tables list information")
var chainNameRe = regexp.MustCompile(`^Chain\s+(\S+)`)
var fieldsHeaderRe = regexp.MustCompile(`^\s*pkts\s+bytes\s+`)
var valuesRe = regexp.MustCompile(`^\s*(\d+)\s+(\d+)\s+.*?/\*\s*(.+?)\s*\*/\s*`)

func (ipt *Iptables) parseAndGather(data string, acc telegraf.Accumulator) error {
	lines := strings.Split(data, "\n")
	if len(lines) < 3 {
		return nil
	}
	mchain := chainNameRe.FindStringSubmatch(lines[0])
	if mchain == nil {
		return errParse
	}
	if !fieldsHeaderRe.MatchString(lines[1]) {
		return errParse
	}
	for _, line := range lines[2:] {
		matches := valuesRe.FindStringSubmatch(line)
		if len(matches) != 4 {
			continue
		}

		pkts := matches[1]
		bytes := matches[2]
		comment := matches[3]

		tags := map[string]string{"table": ipt.Table, "chain": mchain[1], "ruleid": comment}
		fields := make(map[string]interface{})

		var err error
		fields["pkts"], err = strconv.ParseUint(pkts, 10, 64)
		if err != nil {
			continue
		}
		fields["bytes"], err = strconv.ParseUint(bytes, 10, 64)
		if err != nil {
			continue
		}
		acc.AddFields(measurement, fields, tags)
	}
	return nil
}

type chainLister func(table, chain string) (string, error)

func init() {
	inputs.Add("ip6tables", func() telegraf.Input {
		ipt := new(Iptables)
		ipt.lister = ipt.chainList
		return ipt
	})
}
