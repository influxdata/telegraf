package inputs_httpjson

import (
	"fmt"
	"net/url"

	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"

	"github.com/influxdata/telegraf/migrations"
	"github.com/influxdata/telegraf/migrations/common"
	"github.com/influxdata/telegraf/plugins/common/tls"
)

const msg = `
    Replacement 'inputs.http' will not report the 'response_time' field and the
    'server' tag is replaced by the 'url' tag. Please adapt your queries!
`

// Define "old" data structure
type httpJSON struct {
	Name            string `toml:"name"`
	Servers         []string
	Method          string
	TagKeys         []string
	ResponseTimeout string
	Parameters      map[string]string
	Headers         map[string]string
	tls.ClientConfig
	common.InputOptions
}

// Migration function
func migrate(tbl *ast.Table) ([]byte, string, error) {
	// Decode the old data structure
	var old httpJSON
	if err := toml.UnmarshalTable(tbl, &old); err != nil {
		return nil, "", err
	}

	// Fill common options
	plugin := make(map[string]interface{})
	old.InputOptions.Migrate()
	general, err := toml.Marshal(old.InputOptions)
	if err != nil {
		return nil, "", fmt.Errorf("marshalling general options failed: %w", err)
	}
	if err := toml.Unmarshal(general, &plugin); err != nil {
		return nil, "", fmt.Errorf("re-unmarshalling general options failed: %w", err)
	}

	// Use a map for the new plugin and fill in the data
	plugin["urls"] = old.Servers
	if old.Name != "" {
		if x, found := plugin["name_override"]; found && x != old.Name {
			return nil, "", fmt.Errorf("conflicting 'name' (%s) and 'name_override' (%s) setting", old.Name, old.NameOverride)
		}
		plugin["name_override"] = old.Name
	}
	if _, found := plugin["name_override"]; !found {
		plugin["name_override"] = "httpjson"
	}
	if old.Method != "" && old.Method != "GET" {
		plugin["method"] = old.Method
	}
	if len(old.TagKeys) > 0 {
		plugin["tag_keys"] = old.TagKeys
	}
	if old.ResponseTimeout != "" {
		plugin["timeout"] = old.ResponseTimeout
	}
	if len(old.Headers) > 0 {
		plugin["headers"] = old.Headers
	}
	if len(old.Parameters) > 0 {
		urls := make([]string, 0, len(old.Servers))
		for _, s := range old.Servers {
			u, err := url.Parse(s)
			if err != nil {
				return nil, "", fmt.Errorf("parsing server %q failed: %w", s, err)
			}
			q := u.Query()
			for k, v := range old.Parameters {
				q.Add(k, v)
			}
			u.RawQuery = q.Encode()
			urls = append(urls, u.String())
		}
		plugin["urls"] = urls
	}

	// Convert TLS parameters
	if old.TLSCA != "" {
		plugin["tls_ca"] = old.TLSCA
	}
	if old.TLSCert != "" {
		plugin["tls_cert"] = old.TLSCert
	}

	if old.TLSKey != "" {
		plugin["tls_key"] = old.TLSKey
	}
	if old.TLSKeyPwd != "" {
		plugin["tls_key_pwd"] = old.TLSKeyPwd
	}
	if old.TLSMinVersion != "" {
		plugin["tls_min_version"] = old.TLSMinVersion
	}
	if old.InsecureSkipVerify {
		plugin["insecure_skip_verify"] = true
	}
	if old.ServerName != "" {
		plugin["tls_server_name"] = old.ServerName
	}
	if old.RenegotiationMethod != "" {
		plugin["tls_renegotiation_method"] = old.RenegotiationMethod
	}
	if old.Enable != nil {
		plugin["tls_enable"] = *old.Enable
	}

	// Parser settings
	plugin["data_format"] = "json"

	// Create the corresponding metric configurations
	cfg := migrations.CreateTOMLStruct("inputs", "http")
	cfg.Add("inputs", "http", plugin)

	// Marshal the new configuration
	buf, err := toml.Marshal(cfg)
	if err != nil {
		return nil, "", err
	}
	buf = append(buf, []byte("\n")...)

	// Create the new content to output
	return buf, msg, nil
}

// Register the migration function for the plugin type
func init() {
	migrations.AddPluginMigration("inputs.httpjson", migrate)
}
