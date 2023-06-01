package inputs_cassandra

import (
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"

	"github.com/influxdata/telegraf/migrations"
)

// Define "old" data structure
type cassandra struct {
	Context string   `toml:"context"`
	Servers []string `toml:"servers"`
	Metrics []string `toml:"metrics"`
}

// Define "new" data structure(s)
type metricConfig struct {
	Name        string   `toml:"name"`
	Mbean       string   `toml:"mbean"`
	FieldPrefix *string  `toml:"field_prefix,omitempty"`
	TagKeys     []string `toml:"tag_keys,omitempty"`
}

type jolokiaAgent struct {
	URLs       []string `toml:"urls"`
	Username   string   `toml:"username,omitempty"`
	Password   string   `toml:"password,omitempty"`
	NamePrefix string   `toml:"name_prefix"`

	Metrics []metricConfig `toml:"metric"`
}

// Migration function
func migrate(tbl *ast.Table) ([]byte, error) {
	// Decode the old data structure
	var old cassandra
	if err := toml.UnmarshalTable(tbl, &old); err != nil {
		return nil, err
	}

	// Collect servers that use the same credentials
	endpoints := make(map[string]jolokiaAgent)
	for _, server := range old.Servers {
		u, err := url.Parse("http://" + server)
		if err != nil {
			return nil, fmt.Errorf("invalid url %q: %w", server, err)
		}
		if u.Path != "" {
			return nil, fmt.Errorf("unexpected path in %q: %w", server, err)
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
			return nil, fmt.Errorf("cannot determine name for metric %q", metric)
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
			return nil, fmt.Errorf("unknown java metric %q", metric)
		}
	}

	// Create the corresponding metric configurations
	cfg := migrations.CreateTOMLStruct("inputs", "jolokia2_agent")
	for _, endpoint := range endpoints {
		if len(javaMetrics) > 0 {
			plugin := jolokiaAgent{
				URLs:       endpoint.URLs,
				Username:   endpoint.Username,
				Password:   endpoint.Password,
				Metrics:    javaMetrics,
				NamePrefix: "java",
			}
			cfg.Add("inputs", "jolokia2_agent", plugin)
		}
		if len(cassandraMetrics) > 0 {
			plugin := jolokiaAgent{
				URLs:       endpoint.URLs,
				Username:   endpoint.Username,
				Password:   endpoint.Password,
				Metrics:    cassandraMetrics,
				NamePrefix: "cassandra",
			}
			cfg.Add("inputs", "jolokia2_agent", plugin)
		}
	}
	buf, err := toml.Marshal(cfg)
	if err != nil {
		return nil, err
	}
	buf = append(buf, []byte("\n")...)

	// Create the new content to output
	return buf, nil
}

// Register the migration function for the plugin type
func init() {
	migrations.AddPluginMigration("inputs.cassandra", migrate)
}
