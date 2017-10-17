package ipset

import (
	"bufio"
	"fmt"
	"os"
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
}

// Description returns a short description of the plugin
func (ipset *Ipset) Description() string {
	return "Gather packets and bytes counters from ipsets"
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

func (ips *Ipset) Gather(acc telegraf.Accumulator) error {

	// Is ipset installed ?
	ipsetPath, err := exec.LookPath("ipset")
	if err != nil {
		return fmt.Errorf("ipset is not installed.")
	}

	var args []string
	cmdName := ipsetPath
	if ips.UseSudo {
		cmdName = "sudo"
		args = append(args, ipsetPath)
	}
	args = append(args, "save")

	cmd := exec.Command(cmdName, args...)
	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error in dumping ipset data.")
		os.Exit(1)
	}

	scanner := bufio.NewScanner(cmdReader)
	go func() {
		for scanner.Scan() {
			data := strings.Split(scanner.Text(), " ")
			if data[0] == "add" && (data[4] != "0" || ips.ShowAllSets == true) {
				tags := map[string]string{
					"set":  data[1],
					"rule": data[2],
				}
				// Ignore sets created without the "counters" option
				nocomment := strings.Split(scanner.Text(), "\"")[0]
				if !(strings.Contains(nocomment, "packets") &&
					strings.Contains(nocomment, "bytes")) {
					continue
				}

				bytes_total, err := strconv.ParseInt(data[4], 10, 64)
				if err != nil {
					acc.AddError(err)
				}
				packets_total, err := strconv.ParseInt(data[6], 10, 64)
				if err != nil {
					acc.AddError(err)
				}
				fields := map[string]interface{}{
					"bytes_total":   bytes_total,
					"packets_total": packets_total,
				}
				acc.AddCounter("ipset", fields, tags)
			}
		}
	}()

	err = cmd.Start()
	if err != nil {
		acc.AddError(err)
		return nil
	}

	err = cmd.Wait()
	if err != nil {
		acc.AddError(err)
		return nil
	}

	return nil
}

func init() {
	inputs.Add("ipset", func() telegraf.Input {
		ips := new(Ipset)
		return ips
	})
}
