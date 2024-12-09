package config

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"

	"github.com/influxdata/toml/ast"
)

type keyValuePair struct {
	Key   string
	Value string
}

func processTable(parent string, table *ast.Table) ([]keyValuePair, error) {
	var prefix string
	var options []keyValuePair

	if parent != "" {
		prefix = parent + "."
	}

	for k, value := range table.Fields {
		switch v := value.(type) {
		case *ast.KeyValue:
			key := prefix + k
			options = append(options, keyValuePair{
				Key:   key,
				Value: v.Value.Source(),
			})
		case *ast.Table:
			key := prefix + k
			children, err := processTable(key, v)
			if err != nil {
				return nil, fmt.Errorf("parsing table for %q failed: %w", key, err)
			}
			options = append(options, children...)
		case []*ast.Table:
			for i, t := range v {
				key := fmt.Sprintf("%s#%d.%s", prefix, i, k)
				children, err := processTable(key, t)
				if err != nil {
					return nil, fmt.Errorf("parsing table for %q #%d failed: %w", key, i, err)
				}
				options = append(options, children...)
			}
		default:
			return nil, fmt.Errorf("unknown node type %T in key %q", value, prefix+k)
		}
	}
	return options, nil
}

func generatePluginID(prefix string, table *ast.Table) (string, error) {
	// We need to ensure that identically configured plugins _always_
	// result in the same ID no matter which order the options are specified.
	// This is even more relevant as Golang does _not_ give any guarantee
	// on the ordering of maps.
	// So we flatten out the configuration options (also for nested objects)
	// and then sort the resulting array by the canonical key-name.
	cfg, err := processTable("", table)
	if err != nil {
		return "", fmt.Errorf("processing AST failed: %w", err)
	}
	sort.SliceStable(cfg, func(i, j int) bool { return cfg[i].Key < cfg[j].Key })

	// Hash the config options to get the ID. We also prefix the ID with
	// the plugin name to prevent overlap with other plugin types.
	hash := sha256.New()
	hash.Write(append([]byte(prefix), 0))
	for _, kv := range cfg {
		hash.Write([]byte(kv.Key + ":" + kv.Value))
		hash.Write([]byte{0})
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}
