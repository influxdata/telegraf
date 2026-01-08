//go:build linux

package nftables

import (
	"encoding/json"
	"fmt"
)

type table struct {
	Metainfo          *metainfo
	Rules             []*rule
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

// UnmarshalJSON handles both anonymous counters (objects with packets/bytes)
// and named counter references (strings). Named counters are skipped as they
// don't contain inline statistics.
func (e *expr) UnmarshalJSON(b []byte) error {
	var raw struct {
		Counter json.RawMessage `json:"counter,omitempty"`
	}

	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}

	// Handle counter field if present
	if len(raw.Counter) > 0 && string(raw.Counter) != "null" {
		// Check if it's a string (named counter reference) or an object
		// (anonymous counter with stats). Named counters start with a quote.
		if raw.Counter[0] == '"' {
			// Named counter reference - skip it (no inline stats)
			return nil
		}
		// Anonymous counter - parse the object
		var c counter
		if err := json.Unmarshal(raw.Counter, &c); err != nil {
			return err
		}
		e.Cntr = &c
	}

	return nil
}

type counter struct {
	Packets int64 `json:"packets"`
	Bytes   int64 `json:"bytes"`
}
