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
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/olivere/elastic"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/outputs"
)

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
	HealthCheckTimeout  config.Duration `toml:"health_check_timeout"`
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

	client             *elastic.Client
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
	transport := &http.Transport{
		TLSClientConfig: tlsCfg,
	}

	httpclient := &http.Client{
		Transport: transport,
		Timeout:   time.Duration(a.Timeout),
	}

	// Parse the initial URL to determine scheme
	elasticURL, err := url.Parse(a.URLs[0])
	if err != nil {
		return fmt.Errorf("parsing URL %q failed: %w", a.URLs[0], err)
	}

	// Setup the client options
	options := []elastic.ClientOptionFunc{
		elastic.SetHttpClient(httpclient),
		elastic.SetSniff(a.EnableSniffer),
		elastic.SetScheme(elasticURL.Scheme),
		elastic.SetURL(a.URLs...),
		elastic.SetHealthcheckInterval(time.Duration(a.HealthCheckInterval)),
		elastic.SetHealthcheckTimeout(time.Duration(a.HealthCheckTimeout)),
		elastic.SetGzip(a.EnableGzip),
	}

	if len(a.httpHeaders) > 0 {
		options = append(options, elastic.SetHeaders(a.httpHeaders))
	}

	authOptions, err := a.getAuthOptions()
	if err != nil {
		return fmt.Errorf("getting authentication options failed: %w", err)
	}
	options = append(options, authOptions...)

	if time.Duration(a.HealthCheckInterval) == 0 {
		a.Log.Debugf("Disabling health check")
		options = append(options, elastic.SetHealthcheck(false))
	}

	client, err := elastic.NewClient(options...)
	if err != nil {
		return fmt.Errorf("creating ES client failed: %w", err)
	}

	// Check for ES version on first node
	esVersion, err := client.ElasticsearchVersion(a.URLs[0])
	if err != nil {
		return fmt.Errorf("checking ES version failed: %w", err)
	}

	// quit if ES version is not supported
	majorReleaseNumber, err := strconv.Atoi(strings.Split(esVersion, ".")[0])
	if err != nil || majorReleaseNumber < 5 {
		return fmt.Errorf("elasticsearch version not supported: %s", esVersion)
	}

	a.Log.Infof("Elasticsearch version: %q", esVersion)

	a.client = client
	a.majorReleaseNumber = majorReleaseNumber

	if a.ManageTemplate {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(a.Timeout))
		defer cancel()

		if err := a.manageTemplate(ctx); err != nil {
			return fmt.Errorf("managing templates failed: %w", err)
		}
	}

	return nil
}

func (a *Elasticsearch) Close() error {
	a.client = nil
	return nil
}

func (a *Elasticsearch) Write(metrics []telegraf.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	bulkRequest := a.client.Bulk()
	for _, metric := range metrics {
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

		doc := map[string]any{
			"@timestamp":       metric.Time(),
			"measurement_name": name,
			"tag":              tags,
			name:               fields,
		}

		req := elastic.NewBulkIndexRequest().Index(indexName).Doc(doc)
		if a.UseOpTypeCreate {
			req.OpType("create")
		}
		if a.ForceDocumentID {
			id := getPointID(metric)
			req.Id(id)
		}
		if a.majorReleaseNumber <= 6 {
			req.Type("metrics")
		}
		if a.UsePipeline != "" {
			if pipelineName := a.getPipelineName(a.pipelineName, a.pipelineTagKeys, tags); pipelineName != "" {
				req.Pipeline(pipelineName)
			}
		}
		bulkRequest.Add(req)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(a.Timeout))
	defer cancel()

	res, err := bulkRequest.Do(ctx)
	if err != nil {
		return fmt.Errorf("error sending bulk request to Elasticsearch: %w", err)
	}

	if res.Errors {
		for id, err := range res.Failed() {
			a.Log.Errorf(
				"Elasticsearch indexing failure, id: %d, status: %d, error: %s, caused by: %s, %s",
				id,
				err.Status,
				err.Error.Reason,
				err.Error.CausedBy["reason"],
				err.Error.CausedBy["type"],
			)
			break
		}
		return fmt.Errorf("elasticsearch failed to index %d metrics", len(res.Failed()))
	}

	return nil
}

func (a *Elasticsearch) manageTemplate(ctx context.Context) error {
	templateExists, err := a.client.IndexTemplateExists(a.TemplateName).Do(ctx)
	if err != nil {
		return fmt.Errorf("checking existence of template %q failed: %w", a.TemplateName, err)
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

	if a.OverwriteTemplate || !templateExists || templatePattern != "" {
		// Create an index template
		tmpl, err := createNewTemplate(a.IndexName, a.IndexTemplate, a.majorReleaseNumber)
		if err != nil {
			return fmt.Errorf("creating index template %q failed: %w", a.TemplateName, err)
		}

		if _, err := a.client.IndexPutTemplate(a.TemplateName).BodyString(tmpl).Do(ctx); err != nil {
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

func (a *Elasticsearch) getAuthOptions() ([]elastic.ClientOptionFunc, error) {
	var fns []elastic.ClientOptionFunc

	if !a.Username.Empty() && !a.Password.Empty() {
		username, err := a.Username.Get()
		if err != nil {
			return nil, fmt.Errorf("getting username failed: %w", err)
		}
		password, err := a.Password.Get()
		if err != nil {
			username.Destroy()
			return nil, fmt.Errorf("getting password failed: %w", err)
		}
		fns = append(fns, elastic.SetBasicAuth(username.String(), password.String()))
		username.Destroy()
		password.Destroy()
	}

	if !a.AuthBearerToken.Empty() {
		token, err := a.AuthBearerToken.Get()
		if err != nil {
			return nil, fmt.Errorf("getting token failed: %w", err)
		}
		auth := []string{"Bearer " + token.String()}
		fns = append(fns, elastic.SetHeaders(http.Header{"Authorization": auth}))
		token.Destroy()
	}
	return fns, nil
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
			HealthCheckInterval: config.Duration(time.Second * 10),
			HealthCheckTimeout:  config.Duration(time.Second * 1),
		}
	})
}
