package elasticsearch

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

type templatePart struct {
	TemplatePattern string
	Version         int
	IndexTemplate   string
}
