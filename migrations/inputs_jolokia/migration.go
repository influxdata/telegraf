package inputs_jolokia

import (
	"fmt"
	"strings"

	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"

	"github.com/influxdata/telegraf/migrations"
	"github.com/influxdata/telegraf/migrations/common"
)

// Define "old" data structure
type jolokia struct {
	Context               string   `toml:"context"`
	Mode                  string   `toml:"mode"`
	Servers               []server `toml:"servers"`
	Metrics               []metric `toml:"metrics"`
	Proxy                 server   `toml:"proxy"`
	Delimiter             string   `toml:"delimiter"`
	ResponseHeaderTimeout string   `toml:"response_header_timeout"`
	ClientTimeout         string   `toml:"client_timeout"`
	common.InputOptions
}
type server struct {
	Name     string `toml:"name"`
	Host     string `toml:"host"`
	Username string `toml:"username"`
	Password string `toml:"password"`
	Port     string `toml:"port"`
}

type metric struct {
	Name      string `toml:"name"`
	Mbean     string `toml:"mbean"`
	Attribute string `toml:"attribute"`
	Path      string `toml:"path"`
}

// Define "new" data structure(s)
type jolokiaAgent struct {
	URLs                  []string       `toml:"urls"`
	Username              string         `toml:"username,omitempty"`
	Password              string         `toml:"password,omitempty"`
	Metrics               []metricConfig `toml:"metric"`
	DefaultFieldSeparator string         `toml:"default_field_separator"`
	ResponseTimeout       string         `toml:"response_timeout,omitempty"`

	// Common options
	Interval         string            `toml:"interval,omitempty"`
	Precision        string            `toml:"precision,omitempty"`
	CollectionJitter string            `toml:"collection_jitter,omitempty"`
	CollectionOffset string            `toml:"collection_offset,omitempty"`
	NamePrefix       string            `toml:"name_prefix,omitempty"`
	NameSuffix       string            `toml:"name_suffix,omitempty"`
	NameOverride     string            `toml:"name_override,omitempty"`
	Alias            string            `toml:"alias,omitempty"`
	Tags             map[string]string `toml:"tags,omitempty"`

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

type jolokiaProxy struct {
	URL                   string         `toml:"url"`
	Username              string         `toml:"username,omitempty"`
	Password              string         `toml:"password,omitempty"`
	DefaultFieldSeparator string         `toml:"default_field_separator"`
	ResponseTimeout       string         `toml:"response_timeout,omitempty"`
	Targets               []targetConfig `toml:"target"`
	Metrics               []metricConfig `toml:"metric"`

	// Common options
	Interval         string            `toml:"interval,omitempty"`
	Precision        string            `toml:"precision,omitempty"`
	CollectionJitter string            `toml:"collection_jitter,omitempty"`
	CollectionOffset string            `toml:"collection_offset,omitempty"`
	NamePrefix       string            `toml:"name_prefix,omitempty"`
	NameSuffix       string            `toml:"name_suffix,omitempty"`
	NameOverride     string            `toml:"name_override,omitempty"`
	Alias            string            `toml:"alias,omitempty"`
	Tags             map[string]string `toml:"tags,omitempty"`

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

type targetConfig struct {
	URL      string `toml:"url"`
	Username string `toml:"username,omitempty"`
	Password string `toml:"password,omitempty"`
}

type metricConfig struct {
	Name        string   `toml:"name"`
	Mbean       string   `toml:"mbean"`
	Paths       []string `toml:"paths"`
	FieldPrefix *string  `toml:"field_prefix,omitempty"`
	TagKeys     []string `toml:"tag_keys,omitempty"`
}

// Migration function
func migrate(tbl *ast.Table) ([]byte, string, error) {
	var messages []string

	// Decode the old data structure
	var old jolokia
	if err := toml.UnmarshalTable(tbl, &old); err != nil {
		return nil, "", err
	}

	// Create new-style metrics according to the old config
	metrics := make([]metricConfig, 0, len(old.Metrics))
	for _, oldm := range old.Metrics {
		mbean := strings.SplitN(oldm.Mbean, "/", 2)
		m := metricConfig{
			Name:  oldm.Name,
			Mbean: mbean[0],
		}

		// Construct the new path from the old attribute/path setting
		contained := len(mbean) <= 1
		if oldm.Attribute != "" {
			attributes := strings.Split(oldm.Attribute, ",")
			for _, a := range attributes {
				if !contained && a == mbean[1] {
					contained = true
				}
				if oldm.Path != "" {
					m.Paths = append(m.Paths, a+"/"+oldm.Path)
				} else {
					m.Paths = append(m.Paths, a)
				}
			}
		} else if oldm.Path != "" {
			m.Paths = append(m.Paths, oldm.Path)
		}
		if !contained {
			m.Paths = append(m.Paths, mbean[1])
		}
		metrics = append(metrics, m)
	}

	// Setup the timeout if any
	var timeout string
	if old.ClientTimeout != "" && old.ResponseHeaderTimeout != "" {
		msg := "both 'client_timeout' and 'response_header_timeout' are specified using the former"
		messages = append(messages, msg)
		timeout = old.ClientTimeout
	} else if old.ClientTimeout != "" {
		timeout = old.ClientTimeout
	} else if old.ResponseHeaderTimeout != "" {
		timeout = old.ResponseHeaderTimeout
	}

	// Create the corresponding plugin configurations
	var newcfg interface{}
	if old.Mode == "proxy" {
		// Create a new proxy setup
		cfg := migrations.CreateTOMLStruct("inputs", "jolokia2_proxy")

		// Create a new agent setup
		for _, server := range old.Servers {
			proxy := "http://" + old.Proxy.Host + ":" + old.Proxy.Port + strings.TrimRight(old.Context, "/")
			plugin := jolokiaProxy{
				URL:      proxy,
				Username: old.Proxy.Username,
				Password: old.Proxy.Password,
				Targets: []targetConfig{{
					URL:      fmt.Sprintf("service:jmx:rmi:///jndi/rmi://%s:%s/jmxrmi", server.Host, server.Port),
					Username: server.Username,
					Password: server.Password,
				}},
				Metrics:               metrics,
				DefaultFieldSeparator: "_",
				ResponseTimeout:       timeout,
				NameOverride:          "jolokia",
				Tags: map[string]string{
					"jolokia_name": server.Name,
					"jolokia_port": server.Port,
					"jolokia_host": server.Host,
				},
			}
			plugin.fillCommon(old.InputOptions)
			cfg.Add("inputs", "jolokia2_proxy", plugin)
		}
		newcfg = cfg
	} else {
		cfg := migrations.CreateTOMLStruct("inputs", "jolokia2_agent")

		// Create a new agent setup
		for _, server := range old.Servers {
			endpoint := "http://" + server.Host + ":" + server.Port + strings.TrimRight(old.Context, "/")
			plugin := jolokiaAgent{
				URLs:                  []string{endpoint},
				Username:              server.Username,
				Password:              server.Password,
				Metrics:               metrics,
				DefaultFieldSeparator: "_",
				ResponseTimeout:       timeout,
				NameOverride:          "jolokia",
				Tags: map[string]string{
					"jolokia_name": server.Name,
					"jolokia_port": server.Port,
					"jolokia_host": server.Host,
				},
			}
			plugin.fillCommon(old.InputOptions)
			cfg.Add("inputs", "jolokia2_agent", plugin)
		}
		newcfg = cfg
	}

	// Marshal the new configuration
	buf, err := toml.Marshal(newcfg)
	if err != nil {
		return nil, "", err
	}
	buf = append(buf, []byte("\n")...)

	// Create the new content to output
	return buf, strings.Join(messages, ";"), nil
}

func (j *jolokiaAgent) fillCommon(o common.InputOptions) {
	o.Migrate()

	j.Interval = o.Interval
	j.Precision = o.Precision
	j.CollectionJitter = o.CollectionJitter
	j.CollectionOffset = o.CollectionOffset
	j.NamePrefix = o.NamePrefix
	j.NameSuffix = o.NameSuffix
	if o.NameOverride != "" {
		j.NameOverride = o.NameOverride
	}
	j.Alias = o.Alias
	if len(o.Tags) > 0 {
		if j.Tags == nil {
			j.Tags = make(map[string]string, len(o.Tags))
		}
		for k, v := range o.Tags {
			j.Tags[k] = v
		}
	}

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

func (j *jolokiaProxy) fillCommon(o common.InputOptions) {
	o.Migrate()

	j.Interval = o.Interval
	j.Precision = o.Precision
	j.CollectionJitter = o.CollectionJitter
	j.CollectionOffset = o.CollectionOffset
	j.NamePrefix = o.NamePrefix
	j.NameSuffix = o.NameSuffix
	if o.NameOverride != "" {
		j.NameOverride = o.NameOverride
	}
	j.Alias = o.Alias
	if len(o.Tags) > 0 {
		if j.Tags == nil {
			j.Tags = make(map[string]string, len(o.Tags))
		}
		for k, v := range o.Tags {
			j.Tags[k] = v
		}
	}

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
	migrations.AddPluginMigration("inputs.jolokia", migrate)
}
