// +build linux

package ipset

import (
	"os/exec"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// Ipsets is a telegraf plugin to gather packets and bytes counters from ipset
type Ipset struct {
	ShowAllSets bool
	UseSudo     bool
	lister      setLister
}

// Description returns a short description of the plugin
func (ipset *Ipset) Description() string {
	return "Gather packets and bytes counters from Linux ipsets"
}

// SampleConfig returns sample configuration options.
func (ipset *Ipset) SampleConfig() string {
	return `
  ## By default, we only show sets which have already matched at least 1 packet.
  ## set show_all_sets = true to gather them all.
  show_all_sets = false
  ## Adjust your sudo settings appropriately if using this option ("sudo ipset save")
  ## TODO: can we replace this with systemd privileges ? CAP_NET_ADMIN should DTRT
  use_sudo = false
`
}

const measurement = "ipset"

func (ips *Ipset) Gather(acc telegraf.Accumulator) error {
	list, e := ips.lister()
	if e != nil {
		acc.AddError(e)
	}

	lines := strings.Split(list, "\n")
	for _, line := range lines {
		// Ignore sets created without the "counters" option
		nocomment := strings.Split(line, "\"")[0]
		if !(strings.Contains(nocomment, "packets") &&
			strings.Contains(nocomment, "bytes")) {
			continue
		}

		data := strings.Split(line, " ")
		if data[0] == "add" && (data[4] != "0" || ips.ShowAllSets == true) {
			tags := map[string]string{
				"set":  data[1],
				"rule": data[2],
			}
			packets_total, err := strconv.ParseInt(data[4], 10, 64)
			if err != nil {
				acc.AddError(err)
			}
			bytes_total, err := strconv.ParseInt(data[6], 10, 64)
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

func (ips *Ipset) setList() (string, error) {
	// Is ipset installed ?
	ipsetPath, err := exec.LookPath("ipset")
	if err != nil {
		return "", err
	}
	var args []string
	cmdName := ipsetPath
	if ips.UseSudo {
		cmdName = "sudo"
		args = append(args, ipsetPath)
	}
	args = append(args, "save")

	cmd := exec.Command(cmdName, args...)
	out, err := cmd.Output()
	return string(out), err
}

type setLister func() (string, error)

func init() {
	inputs.Add("ipset", func() telegraf.Input {
		ips := new(Ipset)
		ips.lister = ips.setList
		return ips
	})
}
