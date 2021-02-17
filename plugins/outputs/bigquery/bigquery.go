package bigquery

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"

	"cloud.google.com/go/bigquery"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/outputs"
)

const timeStampFieldName = "timestamp"

var defaultTimeout = internal.Duration{Duration: 5 * time.Second}

const sampleConfig = `	
  ## Credentials File
  credentials_file = "/path/to/service/account/key.json"

  ## Google Cloud Platform Project
  project = "my-gcp-project"

  ## The namespace for the metric descriptor
  dataset = "telegraf"

  ## Timeout for BigQuery operations.
  # timeout = "5s"
`

type BigQuery struct {
	CredentialsFile string `toml:"credentials_file"`
	Project         string `toml:"project"`
	Dataset         string `toml:"dataset"`

	Timeout internal.Duration `toml:"timeout"`
	Log     telegraf.Logger   `toml: "-"`

	client *bigquery.Client
}

// SampleConfig returns the formatted sample configuration for the plugin.
func (s *BigQuery) SampleConfig() string {
	return sampleConfig
}

// Description returns the human-readable function definition of the plugin.
func (s *BigQuery) Description() string {
	return "Configuration for Google Cloud BigQuery to send entries"
}

func (b *BigQuery) Connect() error {
	if b.Project == "" {
		return fmt.Errorf("Project is a required field for BigQuery output")
	}

	if b.Dataset == "" {
		return fmt.Errorf("Dataset is a required field for BigQuery output")
	}

	if b.client == nil {
		return b.setUpDefaultClient()
	}

	return nil
}

func (b *BigQuery) setUpDefaultClient() error {
	var credentialsOption option.ClientOption

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, b.Timeout.Duration)
	defer cancel()

	if b.CredentialsFile != "" {
		credentialsOption = option.WithCredentialsFile(b.CredentialsFile)
	} else {
		creds, err := google.FindDefaultCredentials(ctx)
		if err != nil {
			return fmt.Errorf(
				"unable to find Google Cloud Platform Application Default Credentials: %v."+
					"Either set ADC or provide CredentialsFile config", err)
		}
		credentialsOption = option.WithCredentials(creds)
	}

	client, err := bigquery.NewClient(ctx, b.Project, credentialsOption)
	b.client = client
	return err
}

// Write the metrics to Google Cloud BigQuery.
func (b *BigQuery) Write(metrics []telegraf.Metric) error {
	groupedMetrics := b.groupByMetricName(metrics)

	var wg sync.WaitGroup

	for k, v := range groupedMetrics {
		wg.Add(1)
		go func(k string, v []bigquery.ValueSaver) {
			defer wg.Done()
			b.insertToTable(k, v)
		}(k, v)
	}

	wg.Wait()

	return nil
}

func (b *BigQuery) groupByMetricName(metrics []telegraf.Metric) map[string][]bigquery.ValueSaver {
	groupedMetrics := make(map[string][]bigquery.ValueSaver)

	for _, m := range metrics {
		bqm := newValuesSaver(m)
		groupedMetrics[m.Name()] = append(groupedMetrics[m.Name()], bqm)
	}

	return groupedMetrics
}

func newValuesSaver(m telegraf.Metric) *bigquery.ValuesSaver {
	s := make(bigquery.Schema, 0)
	r := make([]bigquery.Value, 0)
	timeSchema := timeStampFieldSchema()
	s = append(s, timeSchema)
	r = append(r, m.Time())

	s, r = tagsSchemaAndValues(m, s, r)
	s, r = valuesSchemaAndValues(m, s, r)

	return &bigquery.ValuesSaver{
		Schema: s.Relax(),
		Row:    r,
	}
}

func timeStampFieldSchema() *bigquery.FieldSchema {
	return &bigquery.FieldSchema{
		Name: timeStampFieldName,
		Type: bigquery.TimestampFieldType,
	}
}

func tagsSchemaAndValues(m telegraf.Metric, s bigquery.Schema, r []bigquery.Value) ([]*bigquery.FieldSchema, []bigquery.Value) {
	for _, t := range m.TagList() {
		s = append(s, tagFieldSchema(t))
		r = append(r, t.Value)
	}

	return s, r
}

func tagFieldSchema(t *telegraf.Tag) *bigquery.FieldSchema {
	return &bigquery.FieldSchema{
		Name: t.Key,
		Type: bigquery.StringFieldType,
	}
}

func valuesSchemaAndValues(m telegraf.Metric, s bigquery.Schema, r []bigquery.Value) ([]*bigquery.FieldSchema, []bigquery.Value) {
	for _, f := range m.FieldList() {
		s = append(s, valuesSchema(f))
		r = append(r, f.Value)
	}

	return s, r
}

func valuesSchema(f *telegraf.Field) *bigquery.FieldSchema {
	return &bigquery.FieldSchema{
		Name: f.Key,
		Type: valueToBqType(f.Value),
	}
}

func valueToBqType(v interface{}) bigquery.FieldType {
	switch reflect.ValueOf(v).Kind() {
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64:
		return bigquery.IntegerFieldType
	case reflect.Float32, reflect.Float64:
		return bigquery.FloatFieldType
	case reflect.Bool:
		return bigquery.BooleanFieldType
	default:
		return bigquery.StringFieldType
	}
}

func (b *BigQuery) insertToTable(metricName string, metrics []bigquery.ValueSaver) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, b.Timeout.Duration)
	defer cancel()

	table := b.client.DatasetInProject(b.Project, b.Dataset).Table(metricName)
	inserter := table.Inserter()

	if err := inserter.Put(ctx, metrics); err != nil {
		b.Log.Errorf("inserting metric %q failed: %v", metricName, err)
	}
}

func (b *BigQuery) tableForMetric(metricName string) *bigquery.Table {
	return b.client.DatasetInProject(b.Project, b.Dataset).Table(metricName)
}

// Close will terminate the session to the backend, returning error if an issue arises.
func (b *BigQuery) Close() error {
	return b.client.Close()
}

func init() {
	outputs.Add("bigquery", func() telegraf.Output {
		return &BigQuery{
			Timeout: defaultTimeout,
		}
	})
}
