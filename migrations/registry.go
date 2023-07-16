package migrations

import (
	"fmt"

	"github.com/influxdata/toml/ast"
)

type PluginMigrationFunc func(*ast.Table) ([]byte, string, error)

var PluginMigrations = make(map[string]PluginMigrationFunc)

func AddPluginMigration(name string, f PluginMigrationFunc) {
	if _, found := PluginMigrations[name]; found {
		panic(fmt.Errorf("plugin migration function already registered for %q", name))
	}
	PluginMigrations[name] = f
}

type pluginTOMLStruct map[string]map[string][]interface{}

func CreateTOMLStruct(category, name string) pluginTOMLStruct {
	return map[string]map[string][]interface{}{
		category: {
			name: make([]interface{}, 0),
		},
	}
}

func (p *pluginTOMLStruct) Add(category, name string, plugin interface{}) {
	cfg := map[string]map[string][]interface{}(*p)
	cfg[category][name] = append(cfg[category][name], plugin)
}
