//go:generate ../../../tools/readme_config_includer/generator
//go:build linux

package nftables

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

var (
	// defaultTables is the default nftables that we process
	defaultTables = []string{"filter"}
)

type Nftables struct {
	UseSudo bool     `toml:"use_sudo"`
	Binary  string   `toml:"binary"`
	Tables  []string `toml:"tables"`
	args    []string
}

func (*Nftables) SampleConfig() string {
	return sampleConfig
}

func (nft *Nftables) Init() error {
	// Set defaults
	if len(nft.Tables) == 0 {
		nft.Tables = []string{"filter"}
	}
	if nft.Binary == "" {
		nft.Binary = "nft"
	}

	// Construct the command
	nft.args = make([]string, 0, 3)
	if nft.UseSudo {
		nft.args = append(nft.args, nft.Binary)
		nft.Binary = "sudo"
	}
	nft.args = append(nft.args, "--json", "list", "table")
	return nil
}

func (nft *Nftables) Gather(acc telegraf.Accumulator) error {
	for _, table := range nft.Tables {
		acc.AddError(nft.getTableData(table, acc))
	}
	return nil
}

// List a specific table and add to Accumulator
func (nft *Nftables) getTableData(tableName string, acc telegraf.Accumulator) error {
	args := append(nft.args, tableName)
	c := exec.Command(nft.Binary, args...)
	out, err := c.Output()
	if err != nil {
		return fmt.Errorf("error executing nft command: %w", err)
	}
	return parseNftableOutput(acc, out)
}

func parseNftableOutput(acc telegraf.Accumulator, out []byte) error {
	var nftable table
	if err := json.Unmarshal(out, &nftable); err != nil {
		return fmt.Errorf("parsing command output failed: %w", err)
	}
	for _, rule := range nftable.Rules {
		if len(rule.Comment) == 0 {
			continue
		}
		for _, expr := range rule.Exprs {
			if expr.Cntr == nil {
				continue
			}
			fields := map[string]interface{}{"bytes": expr.Cntr.Bytes, "pkts": expr.Cntr.Packets}
			tags := map[string]string{"table": rule.Table, "chain": rule.Chain, "ruleid": rule.Comment}
			acc.AddFields("nftables", fields, tags)
		}
	}
	return nil
}

func init() {
	inputs.Add("nftables", func() telegraf.Input {
		return &Nftables{
			Tables: defaultTables,
		}
	})
}
