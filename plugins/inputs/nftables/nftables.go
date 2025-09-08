package nftables

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"os/exec"
)

const measurement = "nftables"

type NFTables struct {
	Tables []string
}

var NFTableConfig = `
  ## Configuration for nftables
  tables = [ "filter" ]
`

type NFTable struct {
	Metainfo *Metainfo
	Rules    []*Rule
}

// The nftable JSON output is designed in a generic way so
// just handle it via some custom unmarshalling to keep our
// interface clean
func (nftable *NFTable) UnmarshalJSON(b []byte) error {
	var aTable map[string][]map[string]json.RawMessage
	if err := json.Unmarshal(b, &aTable); err != nil {
		return fmt.Errorf("Unable to Unmarshal: %s", b)
	}
	// []map[string]interface
	nfthings := aTable["nftables"]
	for _, nfthing := range nfthings {
		hasKey := func(key string) bool { _, ok := nfthing[key]; return ok }
		switch {
		case hasKey("metainfo"):
			var mi Metainfo
			if err := json.Unmarshal(nfthing["metainfo"], &mi); err == nil {
				nftable.Metainfo = &mi
			} else {
				return fmt.Errorf("Unable to parse Metadata: %v", err)
			}
		case hasKey("rule"):
			var r Rule
			if err := json.Unmarshal(nfthing["rule"], &r); err == nil {
				nftable.Rules = append(nftable.Rules, &r)
			} else {
				return fmt.Errorf("Unable to parse Rule: %v", err)
			}
		default:
			// something we are not parsing
		}
	}
	return nil
}

type Metainfo struct {
	Version           string `json:"version"`
	ReleaseName       string `json:"release_name"`
	JsonSchemaVersion int    `json:"json_schema_version"`
}

type Rule struct {
	Family  string
	Table   string
	Chain   string
	Comment string
	Counter *Counter
}

func (rule *Rule) UnmarshalJSON(b []byte) error {
	var raw struct {
		Family  string                       `json:"family"`
		Table   string                       `json:"table"`
		Chain   string                       `json:"chain"`
		Comment string                       `json:comment"`
		Exprs   []map[string]json.RawMessage `json:"expr"`
	}
	if err := json.Unmarshal(b, &raw); err != nil {
		return fmt.Errorf("Unable to Unmarshal: %s", b)
	}
	rule.Family = raw.Family
	rule.Table = raw.Table
	rule.Chain = raw.Chain
	rule.Comment = raw.Comment

	for _, expr := range raw.Exprs {
		hasKey := func(key string) bool { _, ok := expr[key]; return ok }
		switch {
		case hasKey("counter"):
			rule.Counter = &Counter{}
			if err := json.Unmarshal(expr["counter"], rule.Counter); err != nil {
				return fmt.Errorf("Unable to parse Metadata: %v", err)
			}
		}
	}
	return nil
}

type Counter struct {
	Packets int64 `json:"packets"`
	Bytes   int64 `json:"bytes"`
}

func (s *NFTables) SampleConfig() string {
	return NFTableConfig
}

func (s *NFTables) Description() string {
	return "Gather chain data from an nftable table"
}

func (self *NFTables) Gather(acc telegraf.Accumulator) error {
	if len(self.Tables) == 0 {
		return errors.New("Invalid Configuration. Expected a `Tables` entry with list of nftables to monitor")
	}
	for _, table := range self.Tables {
		err := getTableData(table, acc)
		if err != nil {
			return err
		}
	}
	return nil
}

// List a specific table and add to Accumulator
func getTableData(tableName string, acc telegraf.Accumulator) error {
	nftablePath, err := exec.LookPath("nft")
	if err != nil {
		return errors.New("failed to find nft command ")
	}
	var args []string
	name := "sudo"
	args = append(args, nftablePath, "--json")
	args = append(args, "list", "table", tableName)
	c := exec.Command(name, args...)
	out, err := c.Output()
	if err != nil {
		return errors.New(fmt.Sprintf("Error Executing %s, error: %s", c, err))
	}
	return parseNFTableOutput(out, acc)
}

func parseNFTableOutput(out []byte, acc telegraf.Accumulator) error {
	var nftable NFTable
	err := json.Unmarshal(out, &nftable)
	if err != nil {
		return errors.New(fmt.Sprintf("Error Parsing: %s, Error: %v", out, err))
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
		return &NFTables{}
	})
}
