//go:generate ../../../tools/readme_config_includer/generator
package opensearch

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

	"github.com/opensearch-project/opensearch-go/v2"
	"github.com/opensearch-project/opensearch-go/v2/opensearchapi"
	"github.com/opensearch-project/opensearch-go/v2/opensearchutil"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal/choice"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/outputs"
)

//go:embed sample.conf
var sampleConfig string

type Opensearch struct {
	Username            config.Secret   `toml:"username"`
	Password            config.Secret   `toml:"password"`
	AuthBearerToken     config.Secret   `toml:"auth_bearer_token"`
	EnableGzip          bool            `toml:"enable_gzip"`
	EnableSniffer       bool            `toml:"enable_sniffer"`
	FloatHandling       string          `toml:"float_handling"`
	FloatReplacement    float64         `toml:"float_replacement_value"`
	ForceDocumentID     bool            `toml:"force_document_id"`
	IndexName           string          `toml:"index_name"`
	TemplateName        string          `toml:"template_name"`
	ManageTemplate      bool            `toml:"manage_template"`
	OverwriteTemplate   bool            `toml:"overwrite_template"`
	DefaultPipeline     string          `toml:"default_pipeline"`
	UsePipeline         string          `toml:"use_pipeline"`
	Timeout             config.Duration `toml:"timeout"`
	HealthCheckInterval config.Duration `toml:"health_check_interval"`
	HealthCheckTimeout  config.Duration `toml:"health_check_timeout"`
	URLs                []string        `toml:"urls"`
	Log                 telegraf.Logger `toml:"-"`
	tls.ClientConfig

	indexTmpl    *template.Template
	pipelineTmpl *template.Template
	onSucc       func(context.Context, opensearchutil.BulkIndexerItem, opensearchutil.BulkIndexerResponseItem)
	onFail       func(context.Context, opensearchutil.BulkIndexerItem, opensearchutil.BulkIndexerResponseItem, error)
	osClient     *opensearch.Client
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
		return errors.New("opensearch urls or index_name is not defined")
	}

	// Determine if we should process NaN and inf values
	valOptions := []string{"", "none", "drop", "replace"}
	if err := choice.Check(o.FloatHandling, valOptions); err != nil {
		return fmt.Errorf("config float_handling type: %w", err)
	}

	if o.FloatHandling == "" {
		o.FloatHandling = "none"
	}

	indexTmpl, err := template.New("index").Parse(o.IndexName)
	if err != nil {
		return fmt.Errorf("error parsing index_name template: %w", err)
	}
	o.indexTmpl = indexTmpl

	pipelineTmpl, err := template.New("index").Parse(o.UsePipeline)
	if err != nil {
		return fmt.Errorf("error parsing use_pipeline template: %w", err)
	}
	o.pipelineTmpl = pipelineTmpl

	o.onSucc = func(_ context.Context, _ opensearchutil.BulkIndexerItem, res opensearchutil.BulkIndexerResponseItem) {
		o.Log.Debugf("Indexed to OpenSearch with status- [%d] Result- %s DocumentID- %s ", res.Status, res.Result, res.DocumentID)
	}
	o.onFail = func(_ context.Context, _ opensearchutil.BulkIndexerItem, res opensearchutil.BulkIndexerResponseItem, err error) {
		if err != nil {
			o.Log.Errorf("error while OpenSearch bulkIndexing: %v", err)
		} else {
			o.Log.Errorf("error while OpenSearch bulkIndexing: %s: %s", res.Error.Type, res.Error.Reason)
		}
	}

	if o.TemplateName == "" {
		return errors.New("template_name configuration not defined")
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
		o.Log.Errorf("error creating OpenSearch client: %v", err)
	}

	if o.ManageTemplate {
		err := o.manageTemplate(ctx)
		if err != nil {
			return err
		}
	}

	_, err = o.osClient.Ping()
	if err != nil {
		return fmt.Errorf("unable to ping OpenSearch server: %w", err)
	}

	return nil
}

func (o *Opensearch) newClient() error {
	username, err := o.Username.Get()
	if err != nil {
		return fmt.Errorf("getting username failed: %w", err)
	}
	defer username.Destroy()

	password, err := o.Password.Get()
	if err != nil {
		return fmt.Errorf("getting password failed: %w", err)
	}
	defer password.Destroy()

	tlsConfig, err := o.ClientConfig.TLSConfig()
	if err != nil {
		return fmt.Errorf("creating TLS config failed: %w", err)
	}
	clientConfig := opensearch.Config{
		Addresses: o.URLs,
		Username:  username.String(),
		Password:  password.String(),
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
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
		header.Add("Authorization", "Bearer "+token.String())
		defer token.Destroy()
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
		return errors.New("failed to instantiate OpenSearch bulkindexer")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(o.Timeout))
	defer cancel()

	for _, metric := range metrics {
		var name = metric.Name()

		// index name has to be re-evaluated each time for telegraf
		// to send the metric to the correct time-based index
		indexName, err := o.GetIndexName(metric)
		if err != nil {
			return fmt.Errorf("generating indexname failed: %w", err)
		}

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
			pipelineName, err := o.getPipelineName(metric)
			if err != nil {
				return fmt.Errorf("failed to evaluate pipeline name: %w", err)
			}

			if pipelineName != "" {
				if indexers[pipelineName] != nil {
					if err := indexers[pipelineName].Add(ctx, bulkIndxrItem); err != nil {
						o.Log.Errorf("error adding metric entry to OpenSearch bulkIndexer: %v for pipeline %s", err, pipelineName)
					}
					continue
				}
			}
		}

		if err := indexers["default"].Add(ctx, bulkIndxrItem); err != nil {
			o.Log.Errorf("error adding metric entry to OpenSearch default bulkIndexer: %v", err)
		}
	}

	for _, bulkIndxr := range indexers {
		if err := bulkIndxr.Close(ctx); err != nil {
			return fmt.Errorf("error sending bulk request to OpenSearch: %w", err)
		}

		// Report the indexer statistics
		stats := bulkIndxr.Stats()
		if stats.NumFailed > 0 {
			return fmt.Errorf("failed to index [%d] documents", stats.NumFailed)
		}

		o.Log.Debugf("Successfully indexed [%d] documents", stats.NumAdded)
	}

	return nil
}

// BulkIndexer supports pipeline at config level so separate indexer instance for each unique pipeline
func getTargetIndexers(metrics []telegraf.Metric, osInst *Opensearch) map[string]opensearchutil.BulkIndexer {
	var indexers = make(map[string]opensearchutil.BulkIndexer)

	if osInst.UsePipeline != "" {
		for _, metric := range metrics {
			pipelineName, err := osInst.getPipelineName(metric)
			if err != nil {
				osInst.Log.Errorf("error while evaluating pipeline name: %v for pipeline %s", err, pipelineName)
			}

			if pipelineName != "" {
				// BulkIndexer supports pipeline at config level not metric level
				if _, ok := indexers[pipelineName]; ok {
					continue
				}
				bulkIndxr, err := createBulkIndexer(osInst, pipelineName)
				if err != nil {
					osInst.Log.Errorf("error while instantiating OpenSearch NewBulkIndexer: %v for pipeline: %s", err, pipelineName)
				} else {
					indexers[pipelineName] = bulkIndxr
				}
			}
		}
	}

	bulkIndxr, err := createBulkIndexer(osInst, "")
	if err != nil {
		osInst.Log.Errorf("error while instantiating OpenSearch NewBulkIndexer: %v for default pipeline", err)
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

func (o *Opensearch) GetIndexName(metric telegraf.Metric) (string, error) {
	var buf bytes.Buffer
	err := o.indexTmpl.Execute(&buf, metric)

	if err != nil {
		return "", fmt.Errorf("creating index name failed: %w", err)
	}
	var indexName = buf.String()
	if strings.Contains(indexName, "{{") {
		return "", fmt.Errorf("failed to evaluate valid indexname: %s", indexName)
	}
	return indexName, nil
}

func (o *Opensearch) getPipelineName(metric telegraf.Metric) (string, error) {
	if o.UsePipeline == "" || !strings.Contains(o.UsePipeline, "{{") {
		return o.UsePipeline, nil
	}

	var buf bytes.Buffer
	err := o.pipelineTmpl.Execute(&buf, metric)
	if err != nil {
		return "", fmt.Errorf("creating pipeline name failed: %w", err)
	}
	var pipelineName = buf.String()
	if strings.Contains(pipelineName, "{{") {
		return "", fmt.Errorf("failed to evaluate valid pipelineName: %s", pipelineName)
	}
	o.Log.Debugf("PipelineTemplate- %s", pipelineName)

	if pipelineName == "" {
		pipelineName = o.DefaultPipeline
	}
	return pipelineName, nil
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

	if strings.Contains(templatePattern, "{{") {
		templatePattern = templatePattern[0:strings.Index(templatePattern, "{{")]
	}

	if templatePattern == "" {
		return errors.New("template cannot be created for dynamic index names without an index prefix")
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

func (o *Opensearch) Close() error {
	o.osClient = nil
	return nil
}
