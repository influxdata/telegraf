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

const measurement = "nftables"

type Nftables struct {
	UseSudo bool   `toml:"use_sudo"`
	Binary  string `toml:"binary"`
	Tables  []string
}

//go:embed sample.conf
var sampleConfig string

type Nftable struct {
	Metainfo          *Metainfo
	Rules             []*Rule
	JSONSchemaVersion int `json:"json_schema_version"`
}

// UnmarshalJSON handles custom parsing of the nftable output which is
// designed in a generic that is not compatible with the generic parser.
func (nftable *Nftable) UnmarshalJSON(b []byte) error {
	var atable map[string][]map[string]json.RawMessage
	if err := json.Unmarshal(b, &atable); err != nil {
		return fmt.Errorf("unable to unmarshal: %s", b)
	}
	// []map[string]interface
	nfthings := atable["nftables"]
	for _, nfthing := range nfthings {
		hasKey := func(key string) bool { _, ok := nfthing[key]; return ok }
		switch {
		case hasKey("metainfo"):
			var mi Metainfo
			err := json.Unmarshal(nfthing["metainfo"], &mi)
			if err != nil {
				return fmt.Errorf("unable to parse metadata: %w", err)
			}
			nftable.Metainfo = &mi
		case hasKey("rule"):
			var r Rule
			err := json.Unmarshal(nfthing["rule"], &r)
			if err != nil {
				return fmt.Errorf("unable to parse rule: %w", err)
			}
			nftable.Rules = append(nftable.Rules, &r)
		default:
			// something we are not parsing
		}
	}
	return nil
}

type Metainfo struct {
	Version           string `json:"version"`
	ReleaseName       string `json:"release_name"`
	JSONSchemaVersion int    `json:"json_schema_version"`
}

type Rule struct {
	Family  string
	Table   string
	Chain   string
	Comment string
	Counter *Counter
}

// UnmarshalJSON handles properly extracting the counter expression from
// the Exprs array input
func (rule *Rule) UnmarshalJSON(b []byte) error {
	var raw struct {
		Family  string                       `json:"family"`
		Table   string                       `json:"table"`
		Chain   string                       `json:"chain"`
		Comment string                       `json:"comment"`
		Exprs   []map[string]json.RawMessage `json:"expr"`
	}
	if err := json.Unmarshal(b, &raw); err != nil {
		return fmt.Errorf("unable to unmarshal: %s", b)
	}
	rule.Family = raw.Family
	rule.Table = raw.Table
	rule.Chain = raw.Chain
	rule.Comment = raw.Comment

	// iterate rule expressions looking for the single counter
	for _, expr := range raw.Exprs {
		hasKey := func(key string) bool { _, ok := expr[key]; return ok }
		if hasKey("counter") {
			rule.Counter = &Counter{}
			if err := json.Unmarshal(expr["counter"], rule.Counter); err != nil {
				return fmt.Errorf("unable to parse counter: %w", err)
			}
			// we can return early since we are not looking for anything else
			return nil
		}
	}
	return nil
}

type Counter struct {
	Packets int64 `json:"packets"`
	Bytes   int64 `json:"bytes"`
}

func (*Nftables) SampleConfig() string {
	return sampleConfig
}

func (*Nftables) Init() error {
	return nil
}

func (nft *Nftables) Gather(acc telegraf.Accumulator) error {
	if len(nft.Tables) == 0 {
		return errors.New("invalid configuration - expected a 'Tables' entry with list of nftables to monitor")
	}
	for _, table := range nft.Tables {
		err := nft.getTableData(table, acc)
		if err != nil {
			return err
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
		return fmt.Errorf("error executing %s, error: %w", c, err)
	}
	return parseNftableOutput(out, acc)
}

func parseNftableOutput(out []byte, acc telegraf.Accumulator) error {
	var nftable Nftable
	err := json.Unmarshal(out, &nftable)
	if err != nil {
		return fmt.Errorf("error parsing: %s, Error: %w", out, err)
	}
	for _, rule := range nftable.Rules {
		// Rule must have a Counter and a Comment
		if rule.Counter != nil && len(rule.Comment) > 0 {
			fields := map[string]interface{}{"bytes": rule.Counter.Bytes, "pkts": rule.Counter.Packets}
			tags := map[string]string{"table": rule.Table, "chain": rule.Chain, "ruleid": rule.Comment}
			acc.AddFields(measurement, fields, tags)
		}
	}
	return nil
}

func init() {
	inputs.Add("nftables", func() telegraf.Input {
		return &Nftables{}
	})
}
