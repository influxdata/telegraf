package inputs_leofs

import (
	"fmt"

	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"

	"github.com/influxdata/telegraf/migrations"
	"github.com/influxdata/telegraf/migrations/common"
)

type leofs struct {
	Servers []string `toml:"servers"`
	common.InputOptions
}

func migrate(tbl *ast.Table) ([]byte, string, error) {
	var old leofs
	if err := toml.UnmarshalTable(tbl, &old); err != nil {
		return nil, "", err
	}

	plugin := make(map[string]interface{})
	old.InputOptions.Migrate()
	general, err := toml.Marshal(old.InputOptions)
	if err != nil {
		return nil, "", fmt.Errorf("marshalling general options failed: %w", err)
	}
	if err := toml.Unmarshal(general, &plugin); err != nil {
		return nil, "", fmt.Errorf("re-unmarshalling general options failed: %w", err)
	}

	plugin["agents"] = old.Servers

	cfg := migrations.CreateTOMLStruct("inputs", "snmp")
	cfg.Add("inputs", "snmp", plugin)

	buf, err := toml.Marshal(cfg)
	if err != nil {
		return nil, "", err
	}
	buf = append(buf, []byte("\n")...)

	return buf, "", nil
}

func init() {
	migrations.AddPluginMigration("inputs.leofs", migrate)
}
