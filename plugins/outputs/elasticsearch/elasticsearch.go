//go:generate ../../../tools/readme_config_includer/generator
package elasticsearch

import (
	"bytes"
	"context"
	"crypto/sha256"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"text/template"
	"time"

	elasticsearch "github.com/elastic/go-elasticsearch/v9"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/outputs"
)

//go:embed sample.conf
var sampleConfig string

type Elasticsearch struct {
	AuthBearerToken     config.Secret          `toml:"auth_bearer_token"`
	DefaultPipeline     string                 `toml:"default_pipeline"`
	DefaultTagValue     string                 `toml:"default_tag_value"`
	EnableGzip          bool                   `toml:"enable_gzip"`
	EnableSniffer       bool                   `toml:"enable_sniffer"`
	FloatHandling       string                 `toml:"float_handling"`
	FloatReplacement    float64                `toml:"float_replacement_value"`
	ForceDocumentID     bool                   `toml:"force_document_id"`
	HealthCheckInterval config.Duration        `toml:"health_check_interval"`
	HealthCheckTimeout  config.Duration        `toml:"health_check_timeout"`
	IndexName           string                 `toml:"index_name"`
	IndexTemplate       map[string]interface{} `toml:"template_index_settings"`
	ManageTemplate      bool                   `toml:"manage_template"`
	OverwriteTemplate   bool                   `toml:"overwrite_template"`
	UseOpTypeCreate     bool                   `toml:"use_optype_create"`
	Username            config.Secret          `toml:"username"`
	Password            config.Secret          `toml:"password"`
	TemplateName        string                 `toml:"template_name"`
	Timeout             config.Duration        `toml:"timeout"`
	URLs                []string               `toml:"urls"`
	UsePipeline         string                 `toml:"use_pipeline"`
	Headers             map[string]interface{} `toml:"headers"`
	Log                 telegraf.Logger        `toml:"-"`
	majorReleaseNumber  int
	pipelineName        string
	pipelineTagKeys     []string
	tagKeys             []string
	tls.ClientConfig

	Client *elasticsearch.Client
}

// ecsVersion is the ECS schema version stamped into every indexed document.
const ecsVersion = "9.3.0"

// telegrafTemplate is the composable index template for ES 8+.
const telegrafTemplate = `{
	"index_patterns": ["{{.TemplatePattern}}"],
	"priority": 100,
	"template": {
		"settings": {
			"index": {{.IndexTemplate}}
		},
		"mappings": {
			"properties": {
				"@timestamp": { "type": "date" },
				"ecs": {
					"properties": {
						"version": { "type": "keyword", "ignore_above": 1024 }
					}
				},
				"event": {
					"properties": {
						"dataset": { "type": "keyword", "ignore_above": 1024 }
					}
				}
			},
			"dynamic_templates": [
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
		}
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
	IndexTemplate   string
}

func (*Elasticsearch) SampleConfig() string {
	return sampleConfig
}

func (a *Elasticsearch) Connect() error {
	if a.URLs == nil || a.IndexName == "" {
		return errors.New("elasticsearch urls or index_name is not defined")
	}

	switch a.FloatHandling {
	case "", "none":
		a.FloatHandling = "none"
	case "drop", "replace":
	default:
		return fmt.Errorf("invalid float_handling type %q", a.FloatHandling)
	}

	if a.EnableSniffer {
		a.Log.Warn("enable_sniffer is not supported with the go-elasticsearch client; ignoring")
	}
	if time.Duration(a.HealthCheckInterval) != 0 || time.Duration(a.HealthCheckTimeout) != 0 {
		a.Log.Warn("health_check_interval and health_check_timeout are not supported with the go-elasticsearch client; ignoring")
	}

	tlsCfg, err := a.ClientConfig.TLSConfig()
	if err != nil {
		return err
	}
	tr := &http.Transport{
		TLSClientConfig: tlsCfg,
	}

	cfg := elasticsearch.Config{
		Addresses:           a.URLs,
		Transport:           tr,
		CompressRequestBody: a.EnableGzip,
		Header:              a.processHeaders(),
	}

	if err := a.applyAuth(&cfg); err != nil {
		return err
	}

	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(a.Timeout))
	defer cancel()

	res, err := client.Info(client.Info.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("elasticsearch version check failed: %w", err)
	}
	defer res.Body.Close()
	if res.IsError() {
		return fmt.Errorf("elasticsearch version check returned error: %s", res.String())
	}

	var info struct {
		Version struct {
			Number string `json:"number"`
		} `json:"version"`
	}
	if err := json.NewDecoder(res.Body).Decode(&info); err != nil {
		return fmt.Errorf("elasticsearch version parse failed: %w", err)
	}

	majorReleaseNumber, err := strconv.Atoi(strings.Split(info.Version.Number, ".")[0])
	if err != nil || majorReleaseNumber < 9 {
		return fmt.Errorf("elasticsearch version not supported: %s (must be >= 9)", info.Version.Number)
	}

	a.Log.Infof("Elasticsearch version: %q", info.Version.Number)
	a.Client = client
	a.majorReleaseNumber = majorReleaseNumber

	if a.ManageTemplate {
		if err := a.manageTemplate(ctx); err != nil {
			return err
		}
	}

	a.IndexName, a.tagKeys = GetTagKeys(a.IndexName)
	a.pipelineName, a.pipelineTagKeys = GetTagKeys(a.UsePipeline)

	return nil
}

// applyAuth sets authentication fields on the elasticsearch.Config.
func (a *Elasticsearch) applyAuth(cfg *elasticsearch.Config) error {
	if !a.Username.Empty() && !a.Password.Empty() {
		username, err := a.Username.Get()
		if err != nil {
			return fmt.Errorf("getting username failed: %w", err)
		}
		defer username.Destroy()

		password, err := a.Password.Get()
		if err != nil {
			return fmt.Errorf("getting password failed: %w", err)
		}
		defer password.Destroy()

		cfg.Username = username.String()
		cfg.Password = password.String()
	}

	if !a.AuthBearerToken.Empty() {
		token, err := a.AuthBearerToken.Get()
		if err != nil {
			return fmt.Errorf("getting token failed: %w", err)
		}
		defer token.Destroy()

		if cfg.Header == nil {
			cfg.Header = http.Header{}
		}
		cfg.Header.Set("Authorization", "Bearer "+token.String())
	}

	return nil
}

func (a *Elasticsearch) processHeaders() http.Header {
	headers := http.Header{}

	if len(a.Headers) == 0 {
		return headers
	}

	for key, value := range a.Headers {
		switch v := value.(type) {
		case string:
			// Single string value - split on comma for backward compatibility
			config.PrintOptionValueDeprecationNotice("outputs.elasticsearch", "headers."+key, v, telegraf.DeprecationInfo{
				Since:     "1.32.0",
				RemovalIn: "1.45.0",
				Notice:    "Use array syntax instead: [\"value1\", \"value2\"]",
			})
			for _, headerValue := range strings.Split(v, ",") {
				headers.Add(key, strings.TrimSpace(headerValue))
			}
		case []interface{}:
			// TOML might parse arrays as []interface{}
			for _, headerValue := range v {
				if strVal, ok := headerValue.(string); ok {
					headers.Add(key, strings.TrimSpace(strVal))
				} else {
					a.Log.Errorf("Header %q contains non-string value in array: %v (type: %T)", key, headerValue, headerValue)
				}
			}
		default:
			a.Log.Errorf("Header %q has invalid type %T, expected string or []string", key, value)
		}
	}

	return headers
}

// GetPointID generates a unique ID for a Metric Point
func GetPointID(m telegraf.Metric) string {
	var buffer bytes.Buffer
	// Timestamp(ns),measurement name and Series Hash for compute the final SHA256 based hash ID

	buffer.WriteString(strconv.FormatInt(m.Time().Local().UnixNano(), 10))
	buffer.WriteString(m.Name())
	buffer.WriteString(strconv.FormatUint(m.HashID(), 10))

	return fmt.Sprintf("%x", sha256.Sum256(buffer.Bytes()))
}

// expandDotKeys converts a flat map of string tags into a nested map by splitting keys on ".".
// Non-dot keys are inserted first so they take precedence over dot-notation paths that share the same root.
func expandDotKeys(tags map[string]string) map[string]interface{} {
	result := make(map[string]interface{})

	// First pass: flat keys (no dots) — these take priority
	for k, v := range tags {
		if !strings.Contains(k, ".") {
			result[k] = v
		}
	}

	// Second pass: dot-notation keys → nested maps
	for k, v := range tags {
		if strings.Contains(k, ".") {
			setNestedValue(result, strings.Split(k, "."), v)
		}
	}

	return result
}

// setNestedValue recursively walks parts to set value at the leaf, creating intermediate maps as needed.
// If an intermediate key already holds a scalar (not a map), the nested insertion is silently skipped.
func setNestedValue(m map[string]interface{}, parts []string, value string) {
	key := parts[0]
	if len(parts) == 1 {
		m[key] = value
		return
	}

	child, exists := m[key]
	if !exists {
		child = make(map[string]interface{})
		m[key] = child
	}

	if childMap, ok := child.(map[string]interface{}); ok {
		setNestedValue(childMap, parts[1:], value)
	}
	// If child is a scalar, the nested key is silently skipped (scalar wins).
}

// buildECSDocument constructs an ECS-compliant document from a Telegraf metric.
// Tags with dot-notation keys are expanded into nested JSON objects.
// Metric fields are nested under the measurement name.
func (a *Elasticsearch) buildECSDocument(metric telegraf.Metric, fields map[string]interface{}) map[string]interface{} {
	m := make(map[string]interface{})
	m["@timestamp"] = metric.Time()
	m["ecs"] = map[string]interface{}{"version": ecsVersion}
	m["event"] = map[string]interface{}{"dataset": metric.Name()}

	for k, v := range expandDotKeys(metric.Tags()) {
		m[k] = v
	}

	if len(fields) > 0 {
		m[metric.Name()] = fields
	}

	return m
}

func (a *Elasticsearch) Write(metrics []telegraf.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)

	for _, metric := range metrics {
		indexName := a.GetIndexName(a.IndexName, metric.Time(), a.tagKeys, metric.Tags())

		// Handle NaN and inf field-values
		fields := make(map[string]interface{})
		for k, value := range metric.Fields() {
			v, ok := value.(float64)
			if !ok || a.FloatHandling == "none" || !(math.IsNaN(v) || math.IsInf(v, 0)) {
				fields[k] = value
				continue
			}
			if a.FloatHandling == "drop" {
				continue
			}

			if math.IsNaN(v) || math.IsInf(v, 1) {
				fields[k] = a.FloatReplacement
			} else {
				fields[k] = -a.FloatReplacement
			}
		}

		// Build bulk action metadata
		actionMeta := map[string]interface{}{"_index": indexName}
		if a.ForceDocumentID {
			actionMeta["_id"] = GetPointID(metric)
		}
		if a.UsePipeline != "" {
			if pipelineName := a.getPipelineName(a.pipelineName, a.pipelineTagKeys, metric.Tags()); pipelineName != "" {
				actionMeta["pipeline"] = pipelineName
			}
		}

		actionKey := "index"
		if a.UseOpTypeCreate {
			actionKey = "create"
		}

		action := map[string]interface{}{actionKey: actionMeta}
		if err := enc.Encode(action); err != nil {
			return fmt.Errorf("failed to encode bulk action: %w", err)
		}

		doc := a.buildECSDocument(metric, fields)
		if err := enc.Encode(doc); err != nil {
			return fmt.Errorf("error sending bulk request to Elasticsearch: %w", err)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(a.Timeout))
	defer cancel()

	res, err := a.Client.Bulk(bytes.NewReader(buf.Bytes()), a.Client.Bulk.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("error sending bulk request to Elasticsearch: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error sending bulk request to Elasticsearch: %s", res.String())
	}

	var bulkRes struct {
		Errors bool `json:"errors"`
		Items  []map[string]struct {
			Status int `json:"status"`
			Error  struct {
				Type   string `json:"type"`
				Reason string `json:"reason"`
			} `json:"error"`
		} `json:"items"`
	}
	if err := json.NewDecoder(res.Body).Decode(&bulkRes); err != nil {
		return fmt.Errorf("failed to parse bulk response: %w", err)
	}

	if bulkRes.Errors {
		var failCount int
		for _, item := range bulkRes.Items {
			for action, result := range item {
				if result.Status >= 400 {
					failCount++
					a.Log.Errorf(
						"Elasticsearch indexing failure, action: %s, status: %d, type: %s, reason: %s",
						action, result.Status, result.Error.Type, result.Error.Reason,
					)
					break
				}
			}
		}
		return fmt.Errorf("elasticsearch failed to index %d metrics", failCount)
	}

	return nil
}

func (a *Elasticsearch) manageTemplate(ctx context.Context) error {
	if a.TemplateName == "" {
		return errors.New("elasticsearch template_name configuration not defined")
	}

	templatePattern := a.IndexName

	if strings.Contains(templatePattern, "%") {
		templatePattern = templatePattern[0:strings.Index(templatePattern, "%")]
	}

	if strings.Contains(templatePattern, "{{") {
		templatePattern = templatePattern[0:strings.Index(templatePattern, "{{")]
	}

	if templatePattern == "" {
		return errors.New("template cannot be created for dynamic index names without an index prefix")
	}

	// Check using the composable index template API (_index_template)
	existsRes, err := a.Client.Indices.ExistsIndexTemplate(
		a.TemplateName,
		a.Client.Indices.ExistsIndexTemplate.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("elasticsearch template check failed, template name: %s, error: %w", a.TemplateName, err)
	}
	existsRes.Body.Close()
	templateExists := existsRes.StatusCode == http.StatusOK

	if a.OverwriteTemplate || !templateExists {
		data, err := a.createNewTemplate(templatePattern)
		if err != nil {
			return err
		}

		putRes, err := a.Client.Indices.PutIndexTemplate(
			a.TemplateName,
			bytes.NewReader(data.Bytes()),
			a.Client.Indices.PutIndexTemplate.WithContext(ctx),
		)
		if err != nil {
			return fmt.Errorf("elasticsearch failed to create index template %s: %w", a.TemplateName, err)
		}
		defer putRes.Body.Close()
		if putRes.IsError() {
			return fmt.Errorf("elasticsearch failed to create index template %s: %s", a.TemplateName, putRes.String())
		}

		a.Log.Debugf("Template %s created or updated\n", a.TemplateName)
	} else {
		a.Log.Debug("Found existing Elasticsearch template. Skipping template management")
	}
	return nil
}

func (a *Elasticsearch) createNewTemplate(templatePattern string) (*bytes.Buffer, error) {
	var indexTemplate string
	if a.IndexTemplate != nil {
		data, err := json.Marshal(&a.IndexTemplate)
		if err != nil {
			return nil, fmt.Errorf("elasticsearch failed to create index settings for template %s: %w", a.TemplateName, err)
		}
		indexTemplate = string(data)
	} else {
		indexTemplate = defaultTemplateIndexSettings
	}

	tp := templatePart{
		TemplatePattern: templatePattern + "*",
		IndexTemplate:   indexTemplate,
	}

	t := template.Must(template.New("template").Parse(telegrafTemplate))
	var tmpl bytes.Buffer

	if err := t.Execute(&tmpl, tp); err != nil {
		return nil, err
	}
	return &tmpl, nil
}

func GetTagKeys(indexName string) (string, []string) {
	tagKeys := make([]string, 0)
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

func (a *Elasticsearch) GetIndexName(indexName string, eventTime time.Time, tagKeys []string, metricTags map[string]string) string {
	if strings.Contains(indexName, "%") {
		var dateReplacer = strings.NewReplacer(
			"%Y", eventTime.UTC().Format("2006"),
			"%y", eventTime.UTC().Format("06"),
			"%m", eventTime.UTC().Format("01"),
			"%d", eventTime.UTC().Format("02"),
			"%H", eventTime.UTC().Format("15"),
			"%V", getISOWeek(eventTime.UTC()),
		)

		indexName = dateReplacer.Replace(indexName)
	}

	tagValues := make([]interface{}, 0, len(tagKeys))
	for _, key := range tagKeys {
		if value, ok := metricTags[key]; ok {
			tagValues = append(tagValues, value)
		} else {
			a.Log.Debugf("Tag %q not found, using %q on index name instead\n", key, a.DefaultTagValue)
			tagValues = append(tagValues, a.DefaultTagValue)
		}
	}

	return fmt.Sprintf(indexName, tagValues...)
}

func (a *Elasticsearch) getPipelineName(pipelineInput string, tagKeys []string, metricTags map[string]string) string {
	if !strings.Contains(pipelineInput, "%") || len(tagKeys) == 0 {
		return pipelineInput
	}

	var tagValues []interface{}

	for _, key := range tagKeys {
		if value, ok := metricTags[key]; ok {
			tagValues = append(tagValues, value)
			continue
		}
		a.Log.Debugf("Tag %s not found, reverting to default pipeline instead.", key)
		return a.DefaultPipeline
	}
	return fmt.Sprintf(pipelineInput, tagValues...)
}

func getISOWeek(eventTime time.Time) string {
	_, week := eventTime.ISOWeek()
	return strconv.Itoa(week)
}

func (a *Elasticsearch) Close() error {
	a.Client = nil
	return nil
}

func init() {
	outputs.Add("elasticsearch", func() telegraf.Output {
		return &Elasticsearch{
			Timeout:             config.Duration(time.Second * 5),
			HealthCheckInterval: config.Duration(time.Second * 10),
			HealthCheckTimeout:  config.Duration(time.Second * 1),
		}
	})
}
