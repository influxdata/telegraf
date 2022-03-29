package elasticsearch

import (
	"bytes"
	"context"
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
		"index": {
			"refresh_interval": "10s",
			"mapping.total_fields.limit": 5000,
			"auto_expand_replicas" : "0-1",
			"codec" : "best_compression"
		}
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

func (a *Elasticsearch) manageTemplate(ctx context.Context) error {
	if a.TemplateName == "" {
		return fmt.Errorf("elasticsearch template_name configuration not defined")
	}

	templateExists, errExists := a.Client.IndexTemplateExists(a.TemplateName).Do(ctx)

	if errExists != nil {
		return fmt.Errorf("elasticsearch template check failed, template name: %s, error: %s", a.TemplateName, errExists)
	}

	templatePattern := a.IndexName

	if strings.Contains(templatePattern, "%") {
		templatePattern = templatePattern[0:strings.Index(templatePattern, "%")]
	}

	if strings.Contains(templatePattern, "{{") {
		templatePattern = templatePattern[0:strings.Index(templatePattern, "{{")]
	}

	if templatePattern == "" {
		return fmt.Errorf("template cannot be created for dynamic index names without an index prefix")
	}

	if (a.OverwriteTemplate) || (!templateExists) || (templatePattern != "") {
		tp := templatePart{
			TemplatePattern: templatePattern + "*",
			Version:         a.majorReleaseNumber,
		}

		t := template.Must(template.New("template").Parse(telegrafTemplate))
		var tmpl bytes.Buffer

		if err := t.Execute(&tmpl, tp); err != nil {
			return err
		}
		_, errCreateTemplate := a.Client.IndexPutTemplate(a.TemplateName).BodyString(tmpl.String()).Do(ctx)

		if errCreateTemplate != nil {
			return fmt.Errorf("elasticsearch failed to create index template %s : %s", a.TemplateName, errCreateTemplate)
		}

		a.Log.Debugf("Template %s created or updated\n", a.TemplateName)
	} else {
		a.Log.Debug("Found existing Elasticsearch template. Skipping template management")
	}
	return nil
}

func (a *Elasticsearch) GetTagKeys(indexName string) (string, []string) {
	tagKeys := []string{}
	startTag := strings.Index(indexName, "{{")

	for startTag >= 0 {
		endTag := strings.Index(indexName, "}}")

		if endTag < 0 {
			startTag = -1
		} else {
			tagName := indexName[startTag+2 : endTag]

			var tagReplacer = strings.NewReplacer(
				"{{"+tagName+"}}", "%s",
			)

			indexName = tagReplacer.Replace(indexName)
			tagKeys = append(tagKeys, strings.TrimSpace(tagName))

			startTag = strings.Index(indexName, "{{")
		}
	}

	return indexName, tagKeys
}
