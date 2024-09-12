package migrations

import (
	"errors"
	"fmt"

	"github.com/influxdata/toml/ast"
)

var ErrNotApplicable = errors.New("no migration applied")

type PluginMigrationFunc func(*ast.Table) ([]byte, string, error)

var PluginMigrations = make(map[string]PluginMigrationFunc)

func AddPluginMigration(name string, f PluginMigrationFunc) {
	if _, found := PluginMigrations[name]; found {
		panic(fmt.Errorf("plugin migration function already registered for %q", name))
	}
	PluginMigrations[name] = f
}

type PluginOptionMigrationFunc PluginMigrationFunc

var PluginOptionMigrations = make(map[string]PluginOptionMigrationFunc)

func AddPluginOptionMigration(name string, f PluginOptionMigrationFunc) {
	if _, found := PluginOptionMigrations[name]; found {
		panic(fmt.Errorf("plugin option migration function already registered for %q", name))
	}
	PluginOptionMigrations[name] = f
}

type GeneralMigrationFunc func(string, string, *ast.Table) ([]byte, string, error)

var GeneralMigrations []GeneralMigrationFunc

func AddGeneralMigration(f GeneralMigrationFunc) {
	GeneralMigrations = append(GeneralMigrations, f)
}

type GlobalMigrationFunc func(string, *ast.Table) ([]byte, string, error)

var GlobalMigrations []GlobalMigrationFunc

func AddGlobalMigration(f GlobalMigrationFunc) {
	GlobalMigrations = append(GlobalMigrations, f)
}
