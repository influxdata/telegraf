//go:generate ../../../tools/readme_config_includer/generator
package elasticsearch

import (
	"bytes"
	"context"
	"crypto/sha256"
	_ "embed"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/outputs"
)

type client interface {
	close()
	bulk(context.Context, []document) error
	templateExists(context.Context, string) (bool, error)
	putTemplate(context.Context, string, string) error
}

//go:embed sample.conf
var sampleConfig string

type Elasticsearch struct {
	URLs                []string        `toml:"urls"`
	Username            config.Secret   `toml:"username"`
	Password            config.Secret   `toml:"password"`
	AuthBearerToken     config.Secret   `toml:"auth_bearer_token"`
	DefaultPipeline     string          `toml:"default_pipeline"`
	DefaultTagValue     string          `toml:"default_tag_value"`
	EnableGzip          bool            `toml:"enable_gzip"`
	EnableSniffer       bool            `toml:"enable_sniffer"`
	FloatHandling       string          `toml:"float_handling"`
	FloatReplacement    float64         `toml:"float_replacement_value"`
	ForceDocumentID     bool            `toml:"force_document_id"`
	HealthCheckInterval config.Duration `toml:"health_check_interval"`
	HealthCheckTimeout  config.Duration `toml:"health_check_timeout" deprecated:"1.41.0;option is ignored, use 'timeout' instead"`
	IndexName           string          `toml:"index_name"`
	IndexTemplate       map[string]any  `toml:"template_index_settings"`
	ManageTemplate      bool            `toml:"manage_template"`
	OverwriteTemplate   bool            `toml:"overwrite_template"`
	UseOpTypeCreate     bool            `toml:"use_optype_create"`
	TemplateName        string          `toml:"template_name"`
	Timeout             config.Duration `toml:"timeout"`
	UsePipeline         string          `toml:"use_pipeline"`
	Headers             map[string]any  `toml:"headers"`
	Log                 telegraf.Logger `toml:"-"`
	tls.ClientConfig

	client             client
	httpHeaders        http.Header
	majorReleaseNumber int
	pipelineName       string
	pipelineTagKeys    []string
	tagKeys            []string
}

func (*Elasticsearch) SampleConfig() string {
	return sampleConfig
}

func (a *Elasticsearch) Init() error {
	// Check settings
	if len(a.URLs) == 0 {
		return errors.New("no 'urls' defined")
	}
	if a.IndexName == "" {
		return errors.New("no 'index_name' defined")
	}

	// Determine if we should process NaN and inf values
	switch a.FloatHandling {
	case "":
		a.FloatHandling = "none"
	case "none", "drop", "replace":
		// Do nothing, those are valid
	default:
		return fmt.Errorf("invalid 'float_handling' type %q", a.FloatHandling)
	}

	// Check template manipulation settings
	if a.ManageTemplate {
		if a.TemplateName == "" {
			return errors.New("no 'template_name' defined")
		}
		if strings.HasPrefix(a.IndexName, "%") || strings.HasPrefix(a.IndexName, "{{") {
			return errors.New("dynamic 'index_name' must have a static prefix")
		}
	}

	// Setup index, tags and pipelines
	a.IndexName, a.tagKeys = getTagKeys(a.IndexName)
	a.pipelineName, a.pipelineTagKeys = getTagKeys(a.UsePipeline)

	// Prepare the HTTP headers
	a.httpHeaders = make(http.Header, len(a.Headers))
	for key, value := range a.Headers {
		switch v := value.(type) {
		case string:
			// Single string value - split on comma for backward compatibility
			config.PrintOptionValueDeprecationNotice("outputs.elasticsearch", "headers."+key, v, telegraf.DeprecationInfo{
				Since:     "1.32.0",
				RemovalIn: "1.45.0",
				Notice:    "Use array syntax instead: [\"value1\", \"value2\"]",
			})
			for headerValue := range strings.SplitSeq(v, ",") {
				a.httpHeaders.Add(key, strings.TrimSpace(headerValue))
			}
		case []any:
			// TOML might parse arrays as []any
			for _, headerValue := range v {
				strVal, ok := headerValue.(string)
				if !ok {
					return fmt.Errorf("header %q contains non-string value in array: %v with type %T", key, headerValue, headerValue)
				}
				a.httpHeaders.Add(key, strings.TrimSpace(strVal))
			}
		default:
			return fmt.Errorf("header %q has invalid type %T, expected string or []string", key, value)
		}
	}

	return nil
}

func (a *Elasticsearch) Connect() error {
	// Create a custom HTTP client for connecting to the ES cluster
	tlsCfg, err := a.ClientConfig.TLSConfig()
	if err != nil {
		return fmt.Errorf("creating TLS configuration failed: %w", err)
	}
	httpClient := &http.Client{
		Transport: &http.Transport{TLSClientConfig: tlsCfg},
		Timeout:   time.Duration(a.Timeout),
	}
	transport := roundTripper{
		urls:              a.URLs,
		headers:           a.httpHeaders,
		enableGzip:        a.EnableGzip,
		enableSniffer:     a.EnableSniffer,
		discoveryInterval: time.Duration(a.HealthCheckInterval),
		httpClient:        httpClient,
		log:               a.Log,
	}
	if err := transport.configureAuth(a.Username, a.Password, a.AuthBearerToken); err != nil {
		httpClient.CloseIdleConnections()
		return fmt.Errorf("getting authentication options failed: %w", err)
	}

	version, major, err := transport.probeVersion(context.Background())
	if err != nil {
		httpClient.CloseIdleConnections()
		return err
	}

	var c client
	switch major {
	case 5:
		if transport.enableSniffer {
			a.Log.Warn("'enable_sniffer' is not supported for Elasticsearch 5.x and will be ignored")
		}
		c, err = newClientV5(transport)
	case 6:
		c, err = newClientV6(transport)
	case 7:
		c, err = newClientV7(transport)
	default:
		httpClient.CloseIdleConnections()
		return fmt.Errorf("elasticsearch version not supported: %s", version)
	}
	if err != nil {
		httpClient.CloseIdleConnections()
		return fmt.Errorf("creating ES client failed: %w", err)
	}
	a.client = c
	a.majorReleaseNumber = major
	a.Log.Infof("Elasticsearch version: %q", version)

	if a.ManageTemplate {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(a.Timeout))
		defer cancel()

		if err := a.manageTemplate(ctx); err != nil {
			a.Close()
			return fmt.Errorf("managing templates failed: %w", err)
		}
	}

	return nil
}

func (a *Elasticsearch) Close() error {
	if a.client != nil {
		a.client.close()
		a.client = nil
	}
	return nil
}

func (a *Elasticsearch) Write(metrics []telegraf.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	documents := make([]document, 0, len(metrics))
	for _, metric := range metrics {
		documents = append(documents, a.buildDocument(metric))
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(a.Timeout))
	defer cancel()
	return a.client.bulk(ctx, documents)
}

func (a *Elasticsearch) buildDocument(metric telegraf.Metric) document {
	name := metric.Name()

	// Determine the name of the index based on the given template and the
	// actual metric data
	tags := metric.Tags()
	indexName := a.getIndexName(a.IndexName, metric.Time(), a.tagKeys, tags)

	// Handle NaN and inf field-values
	fields := make(map[string]any, len(metric.FieldList()))
	for _, f := range metric.FieldList() {
		v, ok := f.Value.(float64)
		if !ok || a.FloatHandling == "none" || !(math.IsNaN(v) || math.IsInf(v, 0)) {
			fields[f.Key] = f.Value
			continue
		}
		if a.FloatHandling == "drop" {
			continue
		}

		if math.IsNaN(v) || math.IsInf(v, 1) {
			fields[f.Key] = a.FloatReplacement
		} else {
			fields[f.Key] = -a.FloatReplacement
		}
	}

	doc := document{
		index:  indexName,
		opType: "index",
		body: map[string]any{
			"@timestamp":       metric.Time(),
			"measurement_name": name,
			"tag":              tags,
			name:               fields,
		},
	}
	if a.UseOpTypeCreate {
		doc.opType = "create"
	}
	if a.ForceDocumentID {
		doc.id = getPointID(metric)
	}
	if a.UsePipeline != "" {
		doc.pipeline = a.getPipelineName(a.pipelineName, a.pipelineTagKeys, tags)
	}
	return doc
}

func (a *Elasticsearch) manageTemplate(ctx context.Context) error {
	templateExists, err := a.client.templateExists(ctx, a.TemplateName)
	if err != nil {
		return fmt.Errorf("checking existence of template %q failed: %w", a.TemplateName, err)
	}

	if a.OverwriteTemplate || !templateExists {
		// Create an index template
		tmpl, err := createNewTemplate(a.IndexName, a.IndexTemplate, a.majorReleaseNumber)
		if err != nil {
			return fmt.Errorf("creating index template %q failed: %w", a.TemplateName, err)
		}

		if err := a.client.putTemplate(ctx, a.TemplateName, tmpl); err != nil {
			return fmt.Errorf("sending index template %q failed: %w", a.TemplateName, err)
		}
		a.Log.Debugf("Template %s created or updated", a.TemplateName)
	} else {
		a.Log.Debug("Found existing Elasticsearch template. Skipping template management")
	}
	return nil
}

func (a *Elasticsearch) getIndexName(indexName string, eventTime time.Time, tagKeys []string, metricTags map[string]string) string {
	if strings.Contains(indexName, "%") {
		dateReplacer := strings.NewReplacer(
			"%Y", eventTime.UTC().Format("2006"),
			"%y", eventTime.UTC().Format("06"),
			"%m", eventTime.UTC().Format("01"),
			"%d", eventTime.UTC().Format("02"),
			"%H", eventTime.UTC().Format("15"),
			"%V", getISOWeek(eventTime.UTC()),
		)
		indexName = dateReplacer.Replace(indexName)
	}

	tagValues := make([]any, 0, len(tagKeys))
	for _, key := range tagKeys {
		if value, ok := metricTags[key]; ok {
			tagValues = append(tagValues, value)
		} else {
			a.Log.Debugf("Tag %q not found, using %q on index name instead", key, a.DefaultTagValue)
			tagValues = append(tagValues, a.DefaultTagValue)
		}
	}

	return fmt.Sprintf(indexName, tagValues...)
}

func (a *Elasticsearch) getPipelineName(pipelineInput string, tagKeys []string, metricTags map[string]string) string {
	if !strings.Contains(pipelineInput, "%") || len(tagKeys) == 0 {
		return pipelineInput
	}

	tagValues := make([]any, 0, len(tagKeys))
	for _, key := range tagKeys {
		if value, ok := metricTags[key]; ok {
			tagValues = append(tagValues, value)
			continue
		}
		a.Log.Debugf("Tag %s not found, reverting to default pipeline instead", key)
		return a.DefaultPipeline
	}

	return fmt.Sprintf(pipelineInput, tagValues...)
}

func getPointID(m telegraf.Metric) string {
	var buffer bytes.Buffer
	// Timestamp(ns),measurement name and Series Hash for compute the final SHA256 based hash ID

	buffer.WriteString(strconv.FormatInt(m.Time().Local().UnixNano(), 10))
	buffer.WriteString(m.Name())
	buffer.WriteString(strconv.FormatUint(m.HashID(), 10))

	return fmt.Sprintf("%x", sha256.Sum256(buffer.Bytes()))
}

func getTagKeys(indexName string) (string, []string) {
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

func getISOWeek(eventTime time.Time) string {
	_, week := eventTime.ISOWeek()
	return strconv.Itoa(week)
}

func init() {
	outputs.Add("elasticsearch", func() telegraf.Output {
		return &Elasticsearch{
			Timeout:             config.Duration(time.Second * 5),
			HealthCheckInterval: config.Duration(15 * time.Minute),
		}
	})
}
