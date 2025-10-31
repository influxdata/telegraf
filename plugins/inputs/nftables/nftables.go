//go:generate ../../../tools/readme_config_includer/generator
//go:build linux

package nftables

import (
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

var (
	// defaultTables is the default nftables that we process
	defaultTables = []string{"filter"}
)

type Nftables struct {
	UseSudo bool     `toml:"use_sudo"`
	Binary  string   `toml:"binary"`
	Tables  []string `toml:"tables"`
}

func (*Nftables) SampleConfig() string {
	return sampleConfig
}

func (nft *Nftables) Init() error {
	if len(nft.Tables) == 0 {
		return errors.New("invalid configuration - expected a 'Tables' entry with list of nftables to monitor")
	}
	return nil
}

func (nft *Nftables) Gather(acc telegraf.Accumulator) error {
	for _, table := range nft.Tables {
		err := nft.getTableData(table, acc)
		if err != nil {
			acc.AddError(err) // Continue through all tables
		}
	}
	return nil
}

// List a specific table and add to Accumulator
func (nft *Nftables) getTableData(tableName string, acc telegraf.Accumulator) error {
	var binary string
	if nft.Binary != "" {
		binary = nft.Binary
	} else {
		binary = "nft"
	}
	nftablePath, err := exec.LookPath(binary)
	if err != nil {
		return errors.New("failed to find nft command ")
	}
	var args []string
	name := nftablePath
	if nft.UseSudo {
		name = "sudo"
		args = append(args, nftablePath)
	}
	args = append(args, "--json", "list", "table", tableName)
	c := exec.Command(name, args...)
	out, err := c.Output()
	if err != nil {
		return fmt.Errorf("error executing nft command: %w", err)
	}
	return parseNftableOutput(acc, out)
}

func parseNftableOutput(acc telegraf.Accumulator, out []byte) error {
	var nftable nftable
	err := json.Unmarshal(out, &nftable)
	if err != nil {
		return fmt.Errorf("error parsing: %s, Error: %w", out, err)
	}
	for _, rule := range nftable.Rules {
		// Rule must have a Counter and a Comment
		if rule.Counter != nil && len(rule.Comment) > 0 {
			fields := map[string]interface{}{"bytes": rule.Counter.Bytes, "pkts": rule.Counter.Packets}
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
