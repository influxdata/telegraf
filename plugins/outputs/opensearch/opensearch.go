//go:generate ../../../tools/readme_config_includer/generator
package opensearch

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/tls"
	_ "embed"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/opensearch-project/opensearch-go/v2"
	"github.com/opensearch-project/opensearch-go/v2/opensearchapi"
	"github.com/opensearch-project/opensearch-go/v2/opensearchutil"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal/choice"
	httpconfig "github.com/influxdata/telegraf/plugins/common/http"
	"github.com/influxdata/telegraf/plugins/outputs"
)

//go:embed sample.conf
var sampleConfig string

type Opensearch struct {
	Username            config.Secret `toml:"username"`
	Password            config.Secret `toml:"password"`
	AuthBearerToken     config.Secret `toml:"auth_bearer_token"`
	EnableGzip          bool          `toml:"enable_gzip"`
	EnableSniffer       bool          `toml:"enable_sniffer"`
	FloatHandling       string        `toml:"float_handling"`
	FloatReplacement    float64       `toml:"float_replacement_value"`
	ForceDocumentID     bool          `toml:"force_document_id"`
	IndexName           string        `toml:"index_name"`
	TemplateName        string        `toml:"template_name"`
	ManageTemplate      bool          `toml:"manage_template"`
	OverwriteTemplate   bool          `toml:"overwrite_template"`
	pipelineName        string
	DefaultPipeline     string `toml:"default_pipeline"`
	UsePipeline         string `toml:"use_pipeline"`
	pipelineTagKeys     []string
	tagKeys             []string
	DefaultTagValue     string          `toml:"default_tag_value"`
	Timeout             config.Duration `toml:"timeout"`
	HealthCheckInterval config.Duration `toml:"health_check_interval"`
	HealthCheckTimeout  config.Duration `toml:"health_check_timeout"`
	URLs                []string        `toml:"urls"`
	Log                 telegraf.Logger `toml:"-"`
	onSucc              func(context.Context, opensearchutil.BulkIndexerItem, opensearchutil.BulkIndexerResponseItem)
	onFail              func(context.Context, opensearchutil.BulkIndexerItem, opensearchutil.BulkIndexerResponseItem, error)
	configOptions       httpconfig.HTTPClientConfig
	osClient            *opensearch.Client
}

//go:embed template.json
var indexTemplate string

type templatePart struct {
	TemplatePattern string
}

func (*Opensearch) SampleConfig() string {
	return sampleConfig
}

func (o *Opensearch) Init() error {
	if len(o.URLs) == 0 || o.IndexName == "" {
		return fmt.Errorf("opensearch urls or index_name is not defined")
	}

	// Determine if we should process NaN and inf values
	valOptions := []string{"", "none", "drop", "replace"}
	if err := choice.Check(o.FloatHandling, valOptions); err != nil {
		return fmt.Errorf("invalid float_handling type %q", o.FloatHandling)
	}

	if o.FloatHandling == "" {
		o.FloatHandling = "none"
	}

	o.IndexName, o.tagKeys = o.GetReplacementKeys(o.IndexName, ".Tag", "%s")
	o.pipelineName, o.pipelineTagKeys = o.GetReplacementKeys(o.UsePipeline, "", "%s")

	o.onSucc = func(ctx context.Context, item opensearchutil.BulkIndexerItem, res opensearchutil.BulkIndexerResponseItem) {
		o.Log.Debugf("Indexed to OpenSearch with status- [%d] Result- %s DocumentID- %s ", res.Status, res.Result, res.DocumentID)
	}
	o.onFail = func(ctx context.Context, item opensearchutil.BulkIndexerItem, res opensearchutil.BulkIndexerResponseItem, err error) {
		if err != nil {
			o.Log.Errorf("error while OpenSearch bulkIndexing: %w", err)
		} else {
			o.Log.Errorf("error while OpenSearch bulkIndexing: %s: %s", res.Error.Type, res.Error.Reason)
		}
	}

	if o.TemplateName == "" {
		return fmt.Errorf("template_name configuration not defined")
	}

	return nil
}

func init() {
	outputs.Add("opensearch", func() telegraf.Output {
		return &Opensearch{
			Timeout:             config.Duration(time.Second * 5),
			HealthCheckInterval: config.Duration(time.Second * 10),
			HealthCheckTimeout:  config.Duration(time.Second * 1),
		}
	})
}

func (o *Opensearch) Connect() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(o.Timeout))
	defer cancel()

	err := o.newClient()
	if err != nil {
		o.Log.Errorf("error creating OpenSearch client: %w", err)
	}

	if o.ManageTemplate {
		err := o.manageTemplate(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (o *Opensearch) newClient() error {
	username, err := o.Username.Get()
	if err != nil {
		return fmt.Errorf("getting username failed: %w", err)
	}
	defer config.ReleaseSecret(username)

	password, err := o.Password.Get()
	if err != nil {
		return fmt.Errorf("getting password failed: %w", err)
	}
	defer config.ReleaseSecret(password)

	clientConfig := opensearch.Config{
		Addresses: o.URLs,
		Username:  string(username),
		Password:  string(password),
	}

	if o.configOptions.InsecureSkipVerify {
		clientConfig.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	header := http.Header{}
	if o.EnableGzip {
		header.Add("Content-Encoding", "gzip")
		header.Add("Content-Type", "application/json")
		header.Add("Accept-Encoding", "gzip")
	}

	if !o.AuthBearerToken.Empty() {
		token, err := o.AuthBearerToken.Get()
		if err != nil {
			return fmt.Errorf("getting token failed: %w", err)
		}
		if string(token) != "" {
			header.Add("Authorization", "Bearer "+string(token))
		}
	}
	clientConfig.Header = header

	client, err := opensearch.NewClient(clientConfig)
	o.osClient = client

	return err
}

// getPointID generates a unique ID for a Metric Point
// Timestamp(ns),measurement name and Series Hash for compute the final
// SHA256 based hash ID
func getPointID(m telegraf.Metric) string {
	var buffer bytes.Buffer
	buffer.WriteString(strconv.FormatInt(m.Time().Local().UnixNano(), 10))
	buffer.WriteString(m.Name())
	buffer.WriteString(strconv.FormatUint(m.HashID(), 10))

	return fmt.Sprintf("%x", sha256.Sum256(buffer.Bytes()))
}

func (o *Opensearch) Write(metrics []telegraf.Metric) error {
	// get indexers based on unique pipeline values
	indexers := getTargetIndexers(metrics, o)
	if len(indexers) == 0 {
		return fmt.Errorf("failed to instantiate OpenSearch bulkindexer")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(o.Timeout))
	defer cancel()

	for _, metric := range metrics {
		var name = metric.Name()

		// index name has to be re-evaluated each time for telegraf
		// to send the metric to the correct time-based index
		indexName := o.GetIndexName(o.IndexName, metric.Time(), o.tagKeys, metric.Tags())

		// Handle NaN and inf field-values
		fields := make(map[string]interface{})
		for k, value := range metric.Fields() {
			v, ok := value.(float64)
			if !ok || o.FloatHandling == "none" || !(math.IsNaN(v) || math.IsInf(v, 0)) {
				fields[k] = value
				continue
			}
			if o.FloatHandling == "drop" {
				continue
			}

			if math.IsNaN(v) || math.IsInf(v, 1) {
				fields[k] = o.FloatReplacement
			} else {
				fields[k] = -o.FloatReplacement
			}
		}

		m := make(map[string]interface{})

		m["@timestamp"] = metric.Time()
		m["measurement_name"] = name
		m["tag"] = metric.Tags()
		m[name] = fields

		body, err := json.Marshal(m)
		if err != nil {
			return fmt.Errorf("failed to marshal body: %w", err)
		}

		bulkIndxrItem := opensearchutil.BulkIndexerItem{
			Action:    "index",
			Index:     indexName,
			Body:      strings.NewReader(string(body)),
			OnSuccess: o.onSucc,
			OnFailure: o.onFail,
		}
		if o.ForceDocumentID {
			bulkIndxrItem.DocumentID = getPointID(metric)
		}

		if o.UsePipeline != "" {
			if pipelineName := o.getPipelineName(o.pipelineName, o.pipelineTagKeys, metric.Tags()); pipelineName != "" {
				if indexers[pipelineName] != nil {
					if err := indexers[pipelineName].Add(ctx, bulkIndxrItem); err != nil {
						o.Log.Errorf("error adding metric entry to OpenSearch bulkIndexer: %w for pipeline %s", err, pipelineName)
					}
					continue
				}
			}
		}

		if err := indexers["default"].Add(ctx, bulkIndxrItem); err != nil {
			o.Log.Errorf("error adding metric entry to OpenSearch default bulkIndexer: %w", err)
		}
	}

	for _, bulkIndxr := range indexers {
		if err := bulkIndxr.Close(ctx); err != nil {
			return fmt.Errorf("error sending bulk request to OpenSearch: %w", err)
		}

		// Report the indexer statistics
		stats := bulkIndxr.Stats()
		if stats.NumAdded < uint64(len(metrics)) {
			return fmt.Errorf("indexed [%d] documents with [%d] errors", stats.NumAdded, stats.NumFailed)
		}
		o.Log.Debugf("Successfully indexed [%d] documents", stats.NumAdded)
	}

	return nil
}

// BulkIndexer supports pipeline at config level so seperate indexer instance for each unique pipeline
func getTargetIndexers(metrics []telegraf.Metric, osInst *Opensearch) map[string]opensearchutil.BulkIndexer {
	var indexers = make(map[string]opensearchutil.BulkIndexer)

	if osInst.UsePipeline != "" {
		for _, metric := range metrics {
			if pipelineName := osInst.getPipelineName(osInst.pipelineName, osInst.pipelineTagKeys, metric.Tags()); pipelineName != "" {
				// BulkIndexer supports pipeline at config level not metric level
				if _, ok := indexers[osInst.pipelineName]; ok {
					continue
				}
				bulkIndxr, err := createBulkIndexer(osInst, pipelineName)
				if err != nil {
					osInst.Log.Errorf("error while intantiating OpenSearch NewBulkIndexer: %w for pipeline: %s", err, pipelineName)
				} else {
					indexers[pipelineName] = bulkIndxr
				}
			}
		}
	}

	bulkIndxr, err := createBulkIndexer(osInst, "")
	if err != nil {
		osInst.Log.Errorf("error while intantiating OpenSearch NewBulkIndexer: %w for default pipeline", err)
	} else {
		indexers["default"] = bulkIndxr
	}
	return indexers
}

func createBulkIndexer(osInst *Opensearch, pipelineName string) (opensearchutil.BulkIndexer, error) {
	var bulkIndexerConfig = opensearchutil.BulkIndexerConfig{
		Client:     osInst.osClient,
		NumWorkers: 4,    // The number of worker goroutines (default: number of CPUs)
		FlushBytes: 5e+6, // The flush threshold in bytes (default: 5M)
	}
	if pipelineName != "" {
		bulkIndexerConfig.Pipeline = pipelineName
	}

	return opensearchutil.NewBulkIndexer(bulkIndexerConfig)
}

func (o *Opensearch) manageTemplate(ctx context.Context) error {
	tempReq := opensearchapi.CatTemplatesRequest{
		Name: o.TemplateName,
	}

	resp, err := tempReq.Do(ctx, o.osClient.Transport)
	if err != nil {
		return fmt.Errorf("template check failed, template name: %s, error: %w", o.TemplateName, err)
	}

	templateExists := resp.Body != http.NoBody
	templatePattern := o.IndexName

	if strings.Contains(templatePattern, "%") {
		templatePattern = templatePattern[0:strings.Index(templatePattern, "%")]
	}

	if strings.Contains(templatePattern, "{{") {
		templatePattern = templatePattern[0:strings.Index(templatePattern, "{{")]
	}

	if templatePattern == "" {
		return fmt.Errorf("template cannot be created for dynamic index names without an index prefix")
	}

	if o.OverwriteTemplate || !templateExists || templatePattern != "" {
		tp := templatePart{
			TemplatePattern: templatePattern + "*",
		}

		t := template.Must(template.New("template").Parse(indexTemplate))
		var tmpl bytes.Buffer

		if err := t.Execute(&tmpl, tp); err != nil {
			return err
		}

		indexTempReq := opensearchapi.IndicesPutTemplateRequest{
			Name: o.TemplateName,
			Body: strings.NewReader(tmpl.String()),
		}
		indexTempResp, err := indexTempReq.Do(ctx, o.osClient.Transport)

		if err != nil || indexTempResp.StatusCode != 200 {
			return fmt.Errorf("creating index template %q failed: %w", o.TemplateName, err)
		}

		o.Log.Debugf("Template %s created or updated", o.TemplateName)
	} else {
		o.Log.Debug("Found existing OpenSearch template. Skipping template management")
	}
	return nil
}

func (o *Opensearch) GetReplacementKeys(indexName string, key string, replacement string) (string, []string) {
	tagKeys := []string{}
	startKey := "{{" + key
	startTag := strings.Index(indexName, startKey)

	for startTag >= 0 {
		endTag := startTag + strings.Index(indexName[startTag:], "}}")

		if endTag < 0 {
			break
		}
		tagName := indexName[startTag+len(startKey) : endTag]
		var tagReplacer = strings.NewReplacer(startKey+tagName+"}}", replacement)

		indexName = tagReplacer.Replace(indexName)
		tagKeys = append(tagKeys, strings.Trim(strings.TrimSpace(tagName), `"`))

		startTag = strings.Index(indexName, startKey)
	}

	return indexName, tagKeys
}

func (o *Opensearch) GetIndexName(indexName string, eventTime time.Time, tagKeys []string, metricTags map[string]string) string {
	tagValues := []interface{}{}

	for _, key := range tagKeys {
		if value, ok := metricTags[key]; ok {
			tagValues = append(tagValues, value)
		} else {
			o.Log.Debugf("Tag %q not found, using %q on index name instead", key, o.DefaultTagValue)
			tagValues = append(tagValues, o.DefaultTagValue)
		}
	}

	indexName = fmt.Sprintf(indexName, tagValues...)

	if strings.Contains(indexName, "{{.Time.Format") {
		updatedIndexName, dateStrArr := o.GetReplacementKeys(indexName, ".Time.Format", "%DATE%")
		indexName = strings.Replace(updatedIndexName, "%DATE%", eventTime.UTC().Format(dateStrArr[0]), 1)
	}

	return indexName
}

func (o *Opensearch) getPipelineName(pipelineInput string, tagKeys []string, metricTags map[string]string) string {
	if !strings.Contains(pipelineInput, "%") || len(tagKeys) == 0 {
		return pipelineInput
	}

	tagValues := make([]interface{}, 0, len(tagKeys))

	for _, key := range tagKeys {
		if value, ok := metricTags[key]; ok {
			tagValues = append(tagValues, value)
			continue
		}
		o.Log.Debugf("Tag %s not found, reverting to default pipeline instead.", key)
		return o.DefaultPipeline
	}
	return fmt.Sprintf(pipelineInput, tagValues...)
}

func (o *Opensearch) Close() error {
	o.osClient = nil
	return nil
}
