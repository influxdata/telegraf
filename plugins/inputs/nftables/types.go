//go:build linux

package nftables

import (
	"encoding/json"
	"fmt"
)

type table struct {
	Metainfo          *metainfo
	Rules             []*rule
	Sets              []*set
	JSONSchemaVersion int `json:"json_schema_version"`
}

// UnmarshalJSON handles custom parsing of the nftables JSON output which uses
// a generic structure incompatible that standard JSON unmarshaling.
func (nftable *table) UnmarshalJSON(b []byte) error {
	var rawTable map[string][]map[string]json.RawMessage
	if err := json.Unmarshal(b, &rawTable); err != nil {
		return fmt.Errorf("unable to unmarshal: %s", b)
	}

	// Get the top-level structure which should be of type []map[string]interface
	nfthings := rawTable["nftables"]
	for _, nfthing := range nfthings {
		if _, found := nfthing["metainfo"]; found {
			var mi metainfo
			if err := json.Unmarshal(nfthing["metainfo"], &mi); err != nil {
				return fmt.Errorf("unable to parse metadata: %w", err)
			}
			nftable.Metainfo = &mi
		} else if _, found := nfthing["rule"]; found {
			var r rule
			if err := json.Unmarshal(nfthing["rule"], &r); err != nil {
				return fmt.Errorf("unable to parse rule: %w", err)
			}
			nftable.Rules = append(nftable.Rules, &r)
		} else if _, found := nfthing["set"]; found {
			var s set
			if err := json.Unmarshal(nfthing["set"], &s); err != nil {
				return fmt.Errorf("unable to parse set: %w", err)
			}
			nftable.Sets = append(nftable.Sets, &s)
		}
	}
	return nil
}

type metainfo struct {
	Version     string `json:"version"`
	ReleaseName string `json:"release_name"`
}

type rule struct {
	Family  string `json:"family"`
	Table   string `json:"table"`
	Chain   string `json:"chain"`
	Comment string `json:"comment"`
	Exprs   []expr `json:"expr"`
}

type expr struct {
	Cntr *counter `json:"counter,omitempty"`
}

type counter struct {
	Packets    int64 `json:"packets"`
	Bytes      int64 `json:"bytes"`
	isNamedRef bool
}

// UnmarshalJSON handles both anonymous counters (objects with packets/bytes)
// and named counter references (strings). Named references are marked with
// isNamedRef flag since they don't contain inline statistics.
func (c *counter) UnmarshalJSON(b []byte) error {
	if len(b) > 0 && b[0] == '"' {
		// Named counter reference - mark it and return
		c.isNamedRef = true
		return nil
	}
	// Anonymous counter - parse the object using type alias to avoid
	// infinite recursion (alias has no methods, so json.Unmarshal uses
	// default struct unmarshaling instead of calling this method again)
	type counterAlias counter
	return json.Unmarshal(b, (*counterAlias)(c))
}

type set struct {
	Family string `json:"family"`
	Name   string `json:"name"`
	Table  string `json:"table"`
	Elem   []elem `json:"elem,omitempty"`
}

type elem struct{}
