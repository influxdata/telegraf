package global_agent

import (
	"errors"

	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"

	"github.com/influxdata/telegraf/migrations"
)

// Migration function
func migrate(name string, tbl *ast.Table) ([]byte, string, error) {
	// Migrate the agent section only...
	if name != "agent" {
		return nil, "", migrations.ErrNotApplicable
	}

	// Decode the old data structure
	var agent map[string]interface{}
	if err := toml.UnmarshalTable(tbl, &agent); err != nil {
		return nil, "", err
	}

	// Check for deprecated option(s) and migrate them
	var applied bool

	// Migrate log settings
	var logtarget string
	var logtargetFound bool
	if raw, found := agent["logtarget"]; found {
		if v, ok := raw.(string); ok {
			logtarget = v
			logtargetFound = true
		}
	}

	var logformat string
	var logformatFound bool
	if raw, found := agent["logformat"]; found {
		if v, ok := raw.(string); ok {
			logformat = v
			logformatFound = true
		}
	}

	if logtargetFound {
		switch logtarget {
		case "stderr":
			delete(agent, "logfile")
		case "file":
		case "eventlog":
			if logformatFound && logformat != "eventlog" {
				return nil, "", errors.New("contradicting setting for 'logtarget' and 'logformat'")
			}
			agent["logformat"] = "eventlog"
			delete(agent, "logfile")
		}
		applied = true
		delete(agent, "logtarget")
	}

	// No options migrated so we can exit early
	if !applied {
		return nil, "", migrations.ErrNotApplicable
	}

	output, err := toml.Marshal(map[string]map[string]interface{}{"agent": agent})
	return output, "", err
}

// Register the migration function for the plugin type
func init() {
	migrations.AddGlobalMigration(migrate)
}
