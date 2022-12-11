//go:generate ../../../tools/readme_config_includer/generator
package opensearch

import (
	"bytes"
	"context"
	"crypto/sha256"
	_ "embed"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/olivere/elastic"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal/choice"
	httpconfig "github.com/influxdata/telegraf/plugins/common/http"
	"github.com/influxdata/telegraf/plugins/outputs"
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
	Password            string          `toml:"password"`
	TemplateName        string          `toml:"template_name"`
	Timeout             config.Duration `toml:"timeout"`
	URLs                []string        `toml:"urls"`
	UsePipeline         string          `toml:"use_pipeline"`
	Username            string          `toml:"username"`
	Log                 telegraf.Logger `toml:"-"`
	pipelineName        string
	pipelineTagKeys     []string
	tagKeys             []string
	httpconfig.HTTPClientConfig

	client *elastic.Client
}

//go:embed telegrafTemplate.json
var telegrafTemplate string

type templatePart struct {
	TemplatePattern string
}

func (*Opensearch) SampleConfig() string {
	return sampleConfig
}

func (a *Opensearch) Connect() error {

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(a.Timeout))
	defer cancel()

	var clientOptions []elastic.ClientOptionFunc

	ctxt := context.Background()
	httpclient, err := a.HTTPClientConfig.CreateClient(ctxt, a.Log)
	if err != nil {
		return err
	}

	osURL, err := url.Parse(a.URLs[0])
	if err != nil {
		return fmt.Errorf("parsing URL failed: %v", err)
	}

	clientOptions = append(clientOptions,
		elastic.SetHttpClient(httpclient),
		elastic.SetSniff(a.EnableSniffer),
		elastic.SetScheme(osURL.Scheme),
		elastic.SetURL(a.URLs...),
		elastic.SetHealthcheckInterval(time.Duration(a.HealthCheckInterval)),
		elastic.SetHealthcheckTimeout(time.Duration(a.HealthCheckTimeout)),
		elastic.SetGzip(a.EnableGzip),
	)

	if a.Username != "" && a.Password != "" {
		clientOptions = append(clientOptions,
			elastic.SetBasicAuth(a.Username, a.Password),
		)
	}

	if a.AuthBearerToken != "" {
		clientOptions = append(clientOptions,
			elastic.SetHeaders(http.Header{
				"Authorization": []string{fmt.Sprintf("Bearer %s", a.AuthBearerToken)},
			}),
		)
	}

	if time.Duration(a.HealthCheckInterval) == 0 {
		clientOptions = append(clientOptions,
			elastic.SetHealthcheck(false),
		)
		a.Log.Debug("Disabling health check")
	}

	a.client, err = elastic.NewClient(clientOptions...)
	if err != nil {
		return err
	}

	if a.ManageTemplate {
		err := a.manageTemplate(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

// getPointID generates a unique ID for a Metric Point
func getPointID(m telegraf.Metric) string {
	var buffer bytes.Buffer
	//Timestamp(ns),measurement name and Series Hash for compute the final SHA256 based hash ID

	buffer.WriteString(strconv.FormatInt(m.Time().Local().UnixNano(), 10)) //nolint:revive // from buffer.go: "err is always nil"
	buffer.WriteString(m.Name())                                           //nolint:revive // from buffer.go: "err is always nil"
	buffer.WriteString(strconv.FormatUint(m.HashID(), 10))                 //nolint:revive // from buffer.go: "err is always nil"

	return fmt.Sprintf("%x", sha256.Sum256(buffer.Bytes()))
}

func (a *Opensearch) Write(metrics []telegraf.Metric) error {
	bulkRequest := a.client.Bulk()

	for _, metric := range metrics {
		var name = metric.Name()

		// index name has to be re-evaluated each time for telegraf
		// to send the metric to the correct time-based index
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

		m := make(map[string]interface{})

		m["@timestamp"] = metric.Time()
		m["measurement_name"] = name
		m["tag"] = metric.Tags()
		m[name] = fields

		br := elastic.NewBulkIndexRequest().Index(indexName).Doc(m)

		if a.ForceDocumentID {
			id := getPointID(metric)
			br.Id(id)
		}

		if a.UsePipeline != "" {
			if pipelineName := a.getPipelineName(a.pipelineName, a.pipelineTagKeys, metric.Tags()); pipelineName != "" {
				br.Pipeline(pipelineName)
			}
		}

		bulkRequest.Add(br)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(a.Timeout))
	defer cancel()

	res, err := bulkRequest.Do(ctx)

	if err != nil {
		return fmt.Errorf("error sending bulk request to Opensearch: %s", err)
	}

	if res.Errors {
		for id, err := range res.Failed() {
			a.Log.Errorf("Opensearch indexing failure, id: %d, error: %s, caused by: %s, %s", id, err.Error.Reason, err.Error.CausedBy["reason"], err.Error.CausedBy["type"])
			break
		}
		return fmt.Errorf("opensearch failed to index %d metrics", len(res.Failed()))
	}

	return nil
}

func (a *Opensearch) manageTemplate(ctx context.Context) error {
	if a.TemplateName == "" {
		return fmt.Errorf("opensearch template_name configuration not defined")
	}

	templateExists, errExists := a.client.IndexTemplateExists(a.TemplateName).Do(ctx)

	if errExists != nil {
		return fmt.Errorf("opensearch template check failed, template name: %s, error: %s", a.TemplateName, errExists)
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
		}

		t := template.Must(template.New("template").Parse(telegrafTemplate))
		var tmpl bytes.Buffer

		if err := t.Execute(&tmpl, tp); err != nil {
			return err
		}
		_, errCreateTemplate := a.client.IndexPutTemplate(a.TemplateName).BodyString(tmpl.String()).Do(ctx)

		if errCreateTemplate != nil {
			return fmt.Errorf("opensearch failed to create index template %s : %s", a.TemplateName, errCreateTemplate)
		}

		a.Log.Debugf("Template %s created or updated\n", a.TemplateName)
	} else {
		a.Log.Debug("Found existing Opensearch template. Skipping template management")
	}
	return nil
}

func (a *Opensearch) GetReplacementKeys(indexName string, key string, replacement string) (string, []string) {
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

func (a *Opensearch) GetIndexName(indexName string, eventTime time.Time, tagKeys []string, metricTags map[string]string) string {
	tagValues := []interface{}{}

	for _, key := range tagKeys {
		if value, ok := metricTags[key]; ok {
			tagValues = append(tagValues, value)
		} else {
			a.Log.Debugf("Tag '%s' not found, using '%s' on index name instead\n", key, a.DefaultTagValue)
			tagValues = append(tagValues, a.DefaultTagValue)
		}
	}

	indexName = fmt.Sprintf(indexName, tagValues...)

	if strings.Contains(indexName, "{{.Time.Format") {
		updatedIndexName, dateStrArr := a.GetReplacementKeys(indexName, ".Time.Format", "%DATE%")
		indexName = strings.Replace(updatedIndexName, "%DATE%", eventTime.UTC().Format(dateStrArr[0]), 1)
	}

	return indexName
}

func (a *Opensearch) getPipelineName(pipelineInput string, tagKeys []string, metricTags map[string]string) string {
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

func (a *Opensearch) Close() error {
	a.client = nil
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

func (a *Opensearch) Init() error {
	if a.URLs == nil || a.IndexName == "" {
		return fmt.Errorf("opensearch urls or index_name is not defined")
	}

	// Determine if we should process NaN and inf values
	valOptions := []string{"", "none", "drop", "replace"}
	if err := choice.Check(a.FloatHandling, valOptions); err != nil {
		return fmt.Errorf("invalid float_handling type %q", a.FloatHandling)
	}

	if a.FloatHandling == "" {
		a.FloatHandling = "none"
	}

	a.IndexName, a.tagKeys = a.GetReplacementKeys(a.IndexName, ".Tag", "%s")
	a.pipelineName, a.pipelineTagKeys = a.GetReplacementKeys(a.UsePipeline, "", "%s")

	return nil
}
