package inputs_cassandra

import (
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"

	"github.com/influxdata/telegraf/migrations"
	"github.com/influxdata/telegraf/migrations/common"
)

// Define "old" data structure
type cassandra struct {
	Context string   `toml:"context"`
	Servers []string `toml:"servers"`
	Metrics []string `toml:"metrics"`
	common.InputOptions
}

// Define "new" data structure(s)
type metricConfig struct {
	Name        string   `toml:"name"`
	Mbean       string   `toml:"mbean"`
	FieldPrefix *string  `toml:"field_prefix,omitempty"`
	TagKeys     []string `toml:"tag_keys,omitempty"`
}

type jolokiaAgent struct {
	URLs     []string       `toml:"urls"`
	Username string         `toml:"username,omitempty"`
	Password string         `toml:"password,omitempty"`
	Metrics  []metricConfig `toml:"metric"`

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

// Migration function
func migrate(tbl *ast.Table) ([]byte, string, error) {
	// Decode the old data structure
	var old cassandra
	if err := toml.UnmarshalTable(tbl, &old); err != nil {
		return nil, "", err
	}

	// Collect servers that use the same credentials
	endpoints := make(map[string]jolokiaAgent)
	for _, server := range old.Servers {
		u, err := url.Parse("http://" + server)
		if err != nil {
			return nil, "", fmt.Errorf("invalid url %q: %w", server, err)
		}
		if u.Path != "" {
			return nil, "", fmt.Errorf("unexpected path in %q: %w", server, err)
		}
		if u.Hostname() == "" {
			u.Host = "localhost:" + u.Port()
		}
		user := u.User.Username()
		passwd, _ := u.User.Password()
		key := user + ":" + passwd

		endpoint, found := endpoints[key]
		if !found {
			endpoint = jolokiaAgent{
				Username: user,
				Password: passwd,
			}
			endpoint.fillCommon(old.InputOptions)
		}
		u.User = nil
		endpoint.URLs = append(endpoint.URLs, u.String())
		endpoints[key] = endpoint
	}

	// Create new-style metrics according to the old config
	var javaMetrics []metricConfig
	var cassandraMetrics []metricConfig
	for _, metric := range old.Metrics {
		bean := strings.TrimPrefix(metric, "/")

		params := make(map[string]string)
		parts := strings.SplitN(bean, ":", 2)
		for _, p := range strings.Split(parts[1], ",") {
			x := strings.SplitN(p, "=", 2)
			params[x[0]] = x[1]
		}

		name, found := params["type"]
		if !found {
			return nil, "", fmt.Errorf("cannot determine name for metric %q", metric)
		}
		name = strings.SplitN(name, "/", 2)[0]

		var tagKeys []string
		var prefix *string
		for k := range params {
			switch k {
			case "name", "scope", "path", "keyspace":
				tagKeys = append(tagKeys, k)
			}
		}
		sort.Strings(tagKeys)
		for i, k := range tagKeys {
			if k == "name" {
				p := fmt.Sprintf("$%d_", i+1)
				prefix = &p
				break
			}
		}

		switch {
		case strings.HasPrefix(bean, "java.lang:"):
			javaMetrics = append(javaMetrics, metricConfig{
				Name:        name,
				Mbean:       bean,
				TagKeys:     tagKeys,
				FieldPrefix: prefix,
			})
		case strings.HasPrefix(bean, "org.apache.cassandra.metrics:"):
			cassandraMetrics = append(cassandraMetrics, metricConfig{
				Name:        name,
				Mbean:       bean,
				TagKeys:     tagKeys,
				FieldPrefix: prefix,
			})
		default:
			return nil, "", fmt.Errorf("unknown java metric %q", metric)
		}
	}

	// Create the corresponding metric configurations
	cfg := migrations.CreateTOMLStruct("inputs", "jolokia2_agent")
	for _, endpoint := range endpoints {
		if len(javaMetrics) > 0 {
			plugin := jolokiaAgent{
				URLs:     endpoint.URLs,
				Username: endpoint.Username,
				Password: endpoint.Password,
				Metrics:  javaMetrics,
			}
			plugin.fillCommon(old.InputOptions)
			plugin.NamePrefix = "java"
			cfg.Add("inputs", "jolokia2_agent", plugin)
		}
		if len(cassandraMetrics) > 0 {
			plugin := jolokiaAgent{
				URLs:     endpoint.URLs,
				Username: endpoint.Username,
				Password: endpoint.Password,
				Metrics:  cassandraMetrics,
			}
			plugin.fillCommon(old.InputOptions)
			plugin.NamePrefix = "cassandra"

			cfg.Add("inputs", "jolokia2_agent", plugin)
		}
	}

	// Marshal the new configuration
	buf, err := toml.Marshal(cfg)
	if err != nil {
		return nil, "", err
	}
	buf = append(buf, []byte("\n")...)

	// Create the new content to output
	return buf, "", nil
}

func (j *jolokiaAgent) fillCommon(o common.InputOptions) {
	o.Migrate()

	j.Interval = o.Interval
	j.Precision = o.Precision
	j.CollectionJitter = o.CollectionJitter
	j.CollectionOffset = o.CollectionOffset
	j.NamePrefix = o.NamePrefix
	j.NameSuffix = o.NameSuffix
	j.NameOverride = o.NameOverride
	j.Alias = o.Alias
	if len(o.Tags) > 0 {
		j.Tags = make(map[string]string, len(o.Tags))
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
	migrations.AddPluginMigration("inputs.cassandra", migrate)
}
