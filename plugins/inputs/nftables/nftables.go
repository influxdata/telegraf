//go:generate ../../../tools/readme_config_includer/generator
//go:build linux

package nftables

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type Nftables struct {
	UseSudo bool     `toml:"use_sudo"`
	Binary  string   `toml:"binary"`
	Tables  []string `toml:"tables"`

	args []string
}

func (*Nftables) SampleConfig() string {
	return sampleConfig
}

func (n *Nftables) Init() error {
	// Set defaults
	if len(n.Tables) == 0 {
		n.Tables = []string{"filter"}
	}
	if n.Binary == "" {
		n.Binary = "nft"
	}

	// Construct the command
	n.args = make([]string, 0, 3)
	if n.UseSudo {
		n.args = append(n.args, n.Binary)
		n.Binary = "sudo"
	}
	n.args = append(n.args, "--json", "list", "table")
	return nil
}

func (n *Nftables) Gather(acc telegraf.Accumulator) error {
	for _, table := range n.Tables {
		acc.AddError(n.gatherTable(acc, table))
	}
	return nil
}

func (n *Nftables) gatherTable(acc telegraf.Accumulator, name string) error {
	// Run the nft command
	args := append(n.args, name)
	c := exec.Command(n.Binary, args...)
	out, err := c.Output()
	if err != nil {
		var oserr *exec.ExitError
		if errors.As(err, &oserr) {
			buf, _, _ := bytes.Cut(oserr.Stderr, []byte("\n"))
			msg := string(bytes.TrimSpace(buf))
			if msg == "Error: No such file or directory" {
				return fmt.Errorf("table %q does not exist", name)
			}
			return fmt.Errorf("error executing nft command: %w (%s)", err, msg)
		}
		return fmt.Errorf("error executing nft command: %w", err)
	}

	// Parse the result into metrics and add them to the accumulator
	var nftable table
	if err := json.Unmarshal(out, &nftable); err != nil {
		return fmt.Errorf("parsing command output failed: %w", err)
	}
	for _, rule := range nftable.Rules {
		if len(rule.Comment) == 0 {
			continue
		}
		for _, expr := range rule.Exprs {
			if expr.Cntr == nil || expr.Cntr.isNamedRef {
				continue
			}
			fields := map[string]interface{}{
				"bytes": expr.Cntr.Bytes,
				"pkts":  expr.Cntr.Packets,
			}
			tags := map[string]string{
				"table": rule.Table,
				"chain": rule.Chain,
				"rule":  rule.Comment,
			}
			acc.AddFields("nftables", fields, tags)
		}
	}
	for _, set := range nftable.Sets {
		fields := map[string]interface{}{
			"count": len(set.Elem),
		}
		tags := map[string]string{
			"table": set.Table,
			"set":   set.Name,
		}
		acc.AddFields("nftables", fields, tags)
	}
	return nil
}

func init() {
	inputs.Add("nftables", func() telegraf.Input {
		return &Nftables{}
	})
}
