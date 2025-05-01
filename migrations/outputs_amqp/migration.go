package outputs_amqp

import (
	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"

	"github.com/influxdata/telegraf/migrations"
)

type amqp struct {
	Headers         map[string]string `toml:"headers"`
	Database        string            `toml:"database"`
	RetentionPolicy string            `toml:"retention_policy"`

	Precision string `toml:"precision"`

	Brokers []string `toml:"brokers"`
	URL     string   `toml:"url"`
}

func migrate(tbl *ast.Table) ([]byte, string, error) {
	var old amqp
	var plugin map[string]interface{}
	if err := migrations.UnmarshalTableSkipMissing(tbl, &old); err != nil {
		return nil, "", err
	}
	if err := toml.UnmarshalTable(tbl, &plugin); err != nil {
		return nil, "", err
	}

	var applied bool

	// Only apply these fields if the headers array is empty, as was the previous behavior.
	// However, still remove these fields from the toml if they exist
	doHeaders := len(old.Headers) == 0
	if old.Database != "" {
		applied = true

		if doHeaders {
			if old.Headers == nil {
				old.Headers = make(map[string]string, 1)
			}
			old.Headers["database"] = old.Database
			plugin["headers"] = old.Headers
		}

		delete(plugin, "database")
	}

	if old.RetentionPolicy != "" {
		applied = true

		if doHeaders {
			if old.Headers == nil {
				old.Headers = make(map[string]string, 1)
			}
			old.Headers["retention_policy"] = old.RetentionPolicy
			plugin["headers"] = old.Headers
		}

		delete(plugin, "retention_policy")
	}

	if old.Precision != "" {
		applied = true
		delete(plugin, "precision")
	}

	if old.URL != "" {
		applied = true
		// Retains old behavior after this option was deprecated
		if len(old.Brokers) == 0 {
			plugin["brokers"] = []string{old.URL}
		}
		delete(plugin, "url")
	}

	// No options migrated so we can exit early
	if !applied {
		return nil, "", migrations.ErrNotApplicable
	}

	// Create the corresponding plugin configurations
	cfg := migrations.CreateTOMLStruct("outputs", "amqp")
	cfg.Add("outputs", "amqp", plugin)

	output, err := toml.Marshal(cfg)
	return output, "", err
}

// Register the migration function for the plugin type
func init() {
	migrations.AddPluginMigration("outputs.amqp", migrate)
}
