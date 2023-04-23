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

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal/choice"
	httpconfig "github.com/influxdata/telegraf/plugins/common/http"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/opensearch-project/opensearch-go/v2"
	"github.com/opensearch-project/opensearch-go/v2/opensearchapi"
	"github.com/opensearch-project/opensearch-go/v2/opensearchutil"
)

//go:embed sample.conf
var sampleConfig string

type Opensearch struct {
	AuthBearerToken     string          `toml:"auth_bearer_token"`
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
	ManageTemplate      bool            `toml:"manage_template"`
	OverwriteTemplate   bool            `toml:"overwrite_template"`
	Password            config.Secret   `toml:"password"`
	TemplateName        string          `toml:"template_name"`
	Timeout             config.Duration `toml:"timeout"`
	URLs                []string        `toml:"urls"`
	UsePipeline         string          `toml:"use_pipeline"`
	Username            config.Secret   `toml:"username"`
	Log                 telegraf.Logger `toml:"-"`
	pipelineName        string
	pipelineTagKeys     []string
	tagKeys             []string
	httpconfig.HTTPClientConfig

	osClient *opensearch.Client
}

//go:embed telegrafTemplate.json
var telegrafTemplate string

type templatePart struct {
	TemplatePattern string
}

func (*Opensearch) SampleConfig() string {
	return sampleConfig
}

func (o *Opensearch) Init() error {
	if o.URLs == nil || o.IndexName == "" {
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

	if o.InsecureSkipVerify {
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
	if o.AuthBearerToken != ""{
		header.Add("Authorization", "Bearer " + o.AuthBearerToken)
	}

	clientConfig.Header = header

	client, err := opensearch.NewClient(clientConfig)
	o.osClient = client

	return err
}

// getPointID generates a unique ID for a Metric Point
func getPointID(m telegraf.Metric) string {
	var buffer bytes.Buffer
	//Timestamp(ns),measurement name and Series Hash for compute the final SHA256 based hash ID

	buffer.WriteString(strconv.FormatInt(m.Time().Local().UnixNano(), 10))
	buffer.WriteString(m.Name())
	buffer.WriteString(strconv.FormatUint(m.HashID(), 10))

	return fmt.Sprintf("%x", sha256.Sum256(buffer.Bytes()))
}

func (o *Opensearch) Write(metrics []telegraf.Metric) error {
	
	bulkIndxr, err := opensearchutil.NewBulkIndexer(opensearchutil.BulkIndexerConfig{
		Client:     o.osClient,
		NumWorkers: 4,      // The number of worker goroutines (default: number of CPUs)
		FlushBytes: 5e+6,   // The flush threshold in bytes (default: 5M)
	})
	if err != nil {
		return fmt.Errorf("ERROR while intantiating Opensearch NewBulkIndexer: %s", err)
	}
	

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(o.Timeout))
	defer cancel()


	onSucc := func(ctx context.Context, item opensearchutil.BulkIndexerItem, res opensearchutil.BulkIndexerResponseItem) {
		fmt.Printf("Indexed to Opensearch with status- [%d] Result- %s DocumentID- %s \n", res.Status, res.Result, res.DocumentID)
	}
	onFail := func(ctx context.Context, item opensearchutil.BulkIndexerItem, res opensearchutil.BulkIndexerResponseItem, err error) {
		if err != nil {
			fmt.Printf("ERROR while Opensearch bulkIndexing: %s \n", err)
		} else {
			fmt.Printf("ERROR while Opensearch bulkIndexing: %s: %s \n", res.Error.Type, res.Error.Reason)
		}
	}

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
			Action: "index",
			Index: indexName,
			Body:      strings.NewReader(string(body)),
			OnSuccess: onSucc,
			OnFailure: onFail,
		}
		if o.ForceDocumentID {
			bulkIndxrItem.DocumentID = getPointID(metric)
		}

		if o.UsePipeline != "" {
			if pipelineName := o.getPipelineName(o.pipelineName, o.pipelineTagKeys, metric.Tags()); pipelineName != "" {
				//TODO BulkIndexer supports pipeline at config level not metric level
			}
		}

		bulkIndxr.Add(ctx, bulkIndxrItem)
	}

	
	if err := bulkIndxr.Close(ctx); err != nil {
		return fmt.Errorf("error sending bulk request to Opensearch: %s", err)
	}

	
	// Report the indexer statistics
	stats := bulkIndxr.Stats()
	if stats.NumAdded < uint64(len(metrics)) {
		return fmt.Errorf("Opensearch indexed [%d] documents with [%d] errors", stats.NumAdded, stats.NumFailed)
	} else {
		fmt.Printf("Opensearch successfully indexed [%d] documents", stats.NumFlushed)
	}

	return nil
}

func (o *Opensearch) manageTemplate(ctx context.Context) error {
	if o.TemplateName == "" {
		return fmt.Errorf("opensearch template_name configuration not defined")
	}

	tempReq := opensearchapi.CatTemplatesRequest{
		Name: o.TemplateName,
	}

	resp, errExists := tempReq.Do(ctx, o.osClient.Transport)
	if errExists != nil {
		return fmt.Errorf("opensearch template check failed, template name: %s, error: %s", o.TemplateName, errExists)
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

	if (o.OverwriteTemplate) || (!templateExists) || (templatePattern != "") {
		tp := templatePart{
			TemplatePattern: templatePattern + "*",
		}

		t := template.Must(template.New("template").Parse(telegrafTemplate))
		var tmpl bytes.Buffer

		if err := t.Execute(&tmpl, tp); err != nil {
			return err
		}

		indexTempReq := opensearchapi.IndicesPutTemplateRequest{
			Name: o.TemplateName,
			Body: strings.NewReader(tmpl.String()),
		}
		indexTempResp, errCreateTemplate := indexTempReq.Do(ctx, o.osClient.Transport)
		
		if errCreateTemplate != nil || indexTempResp.StatusCode != 200 {
			return fmt.Errorf("opensearch failed to create index template %s : %s", o.TemplateName, errCreateTemplate)
		}

		o.Log.Debugf("Template %s created or updated\n", o.TemplateName)
	} else {
		o.Log.Debug("Found existing Opensearch template. Skipping template management")
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
			startTag = -1
		} else {
			tagName := indexName[startTag+len(startKey) : endTag]

			var tagReplacer = strings.NewReplacer(
				startKey+tagName+"}}", replacement,
			)

			indexName = tagReplacer.Replace(indexName)
			tagKeys = append(tagKeys, strings.Trim(strings.TrimSpace(tagName), `"`))

			startTag = strings.Index(indexName, startKey)
		}
	}

	return indexName, tagKeys
}

func (o *Opensearch) GetIndexName(indexName string, eventTime time.Time, tagKeys []string, metricTags map[string]string) string {
	tagValues := []interface{}{}

	for _, key := range tagKeys {
		if value, ok := metricTags[key]; ok {
			tagValues = append(tagValues, value)
		} else {
			o.Log.Debugf("Tag '%s' not found, using '%s' on index name instead\n", key, o.DefaultTagValue)
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

	var tagValues []interface{}

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
