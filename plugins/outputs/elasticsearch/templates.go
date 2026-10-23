package elasticsearch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"text/template"
)

const telegrafTemplate = `
{
	{{ if (lt .Version 6) }}
	"template": "{{.TemplatePattern}}",
	{{ else }}
	"index_patterns" : [ "{{.TemplatePattern}}" ],
	{{ end }}
	"settings": {
		"index": {{.IndexTemplate}}
	},
	"mappings" : {
		{{ if (lt .Version 7) }}
		"metrics" : {
			{{ if (lt .Version 6) }}
			"_all": { "enabled": false },
			{{ end }}
		{{ end }}
		"properties" : {
			"@timestamp" : { "type" : "date" },
			"measurement_name" : { "type" : "keyword" }
		},
		"dynamic_templates": [
			{
				"tags": {
					"match_mapping_type": "string",
					"path_match": "tag.*",
					"mapping": {
						"ignore_above": 512,
						"type": "keyword"
					}
				}
			},
			{
				"metrics_long": {
					"match_mapping_type": "long",
					"mapping": {
						"type": "float",
						"index": false
					}
				}
			},
			{
				"metrics_double": {
					"match_mapping_type": "double",
					"mapping": {
						"type": "float",
						"index": false
					}
				}
			},
			{
				"text_fields": {
					"match": "*",
					"mapping": {
						"norms": false
					}
				}
			}
		]
		{{ if (lt .Version 7) }}
		}
		{{ end }}
	}
}`

const defaultTemplateIndexSettings = `
{
	"refresh_interval": "10s",
	"mapping.total_fields.limit": 5000,
	"auto_expand_replicas": "0-1",
	"codec": "best_compression"
}`

type templateData struct {
	TemplatePattern string
	Version         int
	IndexTemplate   string
}

func createNewTemplate(name string, pattern map[string]any, major int) (string, error) {
	// Create the template-pattern value by extracting the static prefix of the
	// template name
	if strings.Contains(name, "%") {
		name = name[0:strings.Index(name, "%")]
	}
	if strings.Contains(name, "{{") {
		name = name[0:strings.Index(name, "{{")]
	}

	// Allow overriding the index settings by user configuration settings
	indexTemplate := defaultTemplateIndexSettings
	if pattern != nil {
		p, err := json.Marshal(pattern)
		if err != nil {
			return "", fmt.Errorf("creating index settings for template failed: %w", err)
		}
		indexTemplate = string(p)
	}

	// Create the index settings from the `telegrafTemplate` and the extracted
	// variables
	data := templateData{
		TemplatePattern: name + "*",
		Version:         major,
		IndexTemplate:   indexTemplate,
	}

	var buf bytes.Buffer
	tmpl := template.Must(template.New("template").Parse(telegrafTemplate))
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("executing template failed: %w", err)
	}
	return buf.String(), nil
}
