package outputs_riemann_legacy

import (
	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"

	"github.com/influxdata/telegraf/migrations"
	"github.com/influxdata/telegraf/migrations/common"
)

// Define "old" data structure
type riemannLegacy struct {
	URL       string `toml:"url"`
	Transport string `toml:"transport"`
	Separator string `toml:"separator"`
	common.OutputOptions
}

// Define "new" data structure(s)
type riemann struct {
	URL       string `toml:"url"`
	Separator string `toml:"separator"`

	// Common options for outputs
	Alias          string              `toml:"alias,omitempty"`
	NamePass       []string            `toml:"namepass,omitempty"`
	NameDrop       []string            `toml:"namedrop,omitempty"`
	FieldInclude   []string            `toml:"fieldinclude,omitempty"`
	FieldExclude   []string            `toml:"fieldexclude,omitempty"`
	TagPassFilters map[string][]string `toml:"tagpass,omitempty"`
	TagDropFilters map[string][]string `toml:"tagdrop,omitempty"`
	TagExclude     []string            `toml:"tagexclude,omitempty"`
	TagInclude     []string            `toml:"taginclude,omitempty"`
	MetricPass     string              `toml:"metricpass,omitempty"`
}

// Migration function
func migrate(tbl *ast.Table) ([]byte, string, error) {
	// Decode the old data structure
	var old riemannLegacy
	if err := toml.UnmarshalTable(tbl, &old); err != nil {
		return nil, "", err
	}

	// Create new plugin configurations
	cfg := migrations.CreateTOMLStruct("outputs", "riemann")
	plugin := riemann{
		URL:       old.Transport + "://" + old.URL,
		Separator: old.Separator,
	}
	plugin.fillCommon(old.OutputOptions)
	cfg.Add("outputs", "riemann", plugin)

	// Marshal the new configuration
	buf, err := toml.Marshal(cfg)
	if err != nil {
		return nil, "", err
	}
	buf = append(buf, []byte("\n")...)

	// Create the new content to output
	return buf, "", nil
}

func (j *riemann) fillCommon(o common.OutputOptions) {
	o.Migrate()

	j.Alias = o.Alias

	if len(o.NamePass) > 0 {
		j.NamePass = append(j.NamePass, o.NamePass...)
	}
	if len(o.NameDrop) > 0 {
		j.NameDrop = append(j.NameDrop, o.NameDrop...)
	}
	if len(o.FieldInclude) > 0 {
		j.FieldInclude = append(j.FieldInclude, o.FieldInclude...)
	}
	if len(o.FieldExclude) > 0 {
		j.FieldExclude = append(j.FieldExclude, o.FieldExclude...)
	}
	if len(o.TagPassFilters) > 0 {
		j.TagPassFilters = make(map[string][]string, len(o.TagPassFilters))
		for k, v := range o.TagPassFilters {
			j.TagPassFilters[k] = v
		}
	}
	if len(o.TagDropFilters) > 0 {
		j.TagDropFilters = make(map[string][]string, len(o.TagDropFilters))
		for k, v := range o.TagDropFilters {
			j.TagDropFilters[k] = v
		}
	}
	if len(o.TagExclude) > 0 {
		j.TagExclude = append(j.TagExclude, o.TagExclude...)
	}
	if len(o.TagInclude) > 0 {
		j.TagInclude = append(j.TagInclude, o.TagInclude...)
	}
	j.MetricPass = o.MetricPass
}

// Register the migration function for the plugin type
func init() {
	migrations.AddPluginMigration("outputs.riemann_legacy", migrate)
}
