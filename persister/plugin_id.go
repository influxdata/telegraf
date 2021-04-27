package persister

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"

	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"

	"github.com/influxdata/telegraf"
)

func processTable(table *ast.Table) (map[string]string, error) {
	result := make(map[string]string)
	for key, value := range table.Fields {
		switch v := value.(type) {
		case *ast.KeyValue:
			result[key] = v.Value.Source()
		case *ast.Table:
			childs, err := processTable(v)
			if err != nil {
				return nil, fmt.Errorf("parsing table for %q failed: %v", key, err)
			}
			for k, v := range childs {
				result[key+"."+k] = v
			}
		case []*ast.Table:
			for i, t := range v {
				childs, err := processTable(t)
				if err != nil {
					return nil, fmt.Errorf("parsing table for %q #%d failed: %v", key, i, err)
				}
				for k, v := range childs {
					pk := fmt.Sprintf("%s#%d.%s", key, i, k)
					result[pk] = v
				}
			}
		default:
			fmt.Printf("ignoring unknown node type %T in key %q", value, key)
			continue
		}
	}
	return result, nil
}

func normalizePluginConfig(plugin interface{}) ([]telegraf.Tag, error) {
	// Convert the plugin configuration back to toml to reconstruct
	// the user-specified configuration.
	config, err := toml.Marshal(plugin)
	if err != nil {
		return nil, fmt.Errorf("unable to extract user-configuration from plugin: %v", err)
	}

	// We need to ensure that the marshalled TOML text is _always_ the same for the same input and does not change
	// for identical configuration (especially for maps this might be questionable).
	// So we restore the TOML configuration from the plugin values, parse the AST such that we get canonnical
	// entries for the configurations and sort them...
	root, err := toml.Parse(config)
	if err != nil {
		return nil, fmt.Errorf("unable to parse user-configuration from plugin: %v", err)
	}

	options, err := processTable(root)
	if err != nil {
		return nil, fmt.Errorf("unable to process AST for plugin: %v", err)
	}

	result := make([]telegraf.Tag, 0, len(options))
	for k, v := range options {
		result = append(result, telegraf.Tag{k, v})
	}
	sort.SliceStable(result, func(i, j int) bool { return result[i].Key < result[j].Key })

	// fmt.Println("-----------")
	// for i, x := range result {
	// 	fmt.Printf("%d: %q:%q\n", i, x.Key, x.Value)
	// }
	// fmt.Println("-----------")

	return result, nil
}

func generatePluginID(prefix string, plugin interface{}) (string, error) {
	if prefix == "" {
		return "", fmt.Errorf("empty prefix")
	}

	hash := sha256.New()

	// Prefix the ID with the name to prevent access to other plugin types
	if _, err := hash.Write(append([]byte(prefix), 0)); err != nil {
		return "", fmt.Errorf("hashing name failed: %v", err)
	}

	if generator, ok := plugin.(telegraf.StatefulPluginWithID); ok {
		// Add the plugin generated ID
		id := generator.GetPluginStateID()

		if _, err := hash.Write(append([]byte(id), 0)); err != nil {
			return "", fmt.Errorf("hashing state ID failed: %v", err)
		}
	} else {
		// Process the plugin to extract a normalized stable configuration.
		config, err := normalizePluginConfig(plugin)
		if err != nil {
			return "", err
		}

		// Add the config elements to the hash
		for _, kv := range config {
			if _, err := hash.Write([]byte(kv.Key + ":" + kv.Value)); err != nil {
				return "", fmt.Errorf("hashing configuration entry %q failed: %v", kv.Key, err)
			}
			if _, err := hash.Write([]byte{0}); err != nil {
				return "", fmt.Errorf("adding hash configuration-entry end marker failed: %v", err)
			}
		}
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}
