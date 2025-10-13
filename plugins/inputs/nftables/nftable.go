package nftables

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

type nftable struct {
	Metainfo          *Metainfo
	Rules             []*Rule
	JSONSchemaVersion int `json:"json_schema_version"`
}

// UnmarshalJSON handles custom parsing of the nftables JSON output which uses
// a generic structure incompatible that standard JSON unmarshaling.
func (nftable *nftable) UnmarshalJSON(b []byte) error {
	var rawTable map[string][]map[string]json.RawMessage
	if err := json.Unmarshal(b, &rawTable); err != nil {
		return fmt.Errorf("unable to unmarshal: %s", b)
	}
	// []map[string]interface
	nfthings := rawTable["nftables"]
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
	Version     string `json:"version"`
	ReleaseName string `json:"release_name"`
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
			// we can return early since we are not looking for
			// any further data. From testing, multiple counters
			// were never seen attached to a single rule.
			return nil
		}
	}
	return nil
}

type Counter struct {
	Packets int64 `json:"packets"`
	Bytes   int64 `json:"bytes"`
}
