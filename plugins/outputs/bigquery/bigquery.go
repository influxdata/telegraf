package bigquery

import (
	"context"
	"fmt"
	"os"
	"reflect"

	"cloud.google.com/go/bigquery"
	"github.com/influxdata/telegraf"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
)

const (
	testingHostEnv     = "BIGQUERY_TESTING_HOST"
	defaultOffSetKey   = "offset-key.json"
	timeStampFieldName = "timestamp"
)

type BigQuery struct {
	CredentialsFile string `toml:"credentials_file"`
	Project         string `toml:"project"`
	Dataset         string `toml:"dataset"`

	client *bigquery.Client
}

var sampleConfig = `	
  ## Credentials File
  credentials_file = "/path/to/service/account/key.json"
  ## GCP Project
  project = "my-gcp-project"

  ## The namespace for the metric descriptor
  dataset = "telegraf"
`

func (b *BigQuery) Connect() error {
	if b.Project == "" {
		return fmt.Errorf("Project is a required field for BigQuery output")
	}

	if b.Dataset == "" {
		return fmt.Errorf("Dataset is a required field for BigQuery output")
	}

	b.setUpClient()

	return nil
}

func (b *BigQuery) setUpClient() error {
	if endpoint, present := os.LookupEnv(testingHostEnv); present {
		return b.setUpTestClient(endpoint)
	}

	return b.setUpDefaultClient()
}

func (b *BigQuery) setUpTestClient(endpoint string) error {
	noAuth := option.WithoutAuthentication()
	endpoints := option.WithEndpoint("http://" + endpoint)

	ctx := context.Background()

	c, err := bigquery.NewClient(ctx, b.Project, noAuth, endpoints)

	if err != nil {
		return err
	}

	b.client = c

	return nil
}

func (b *BigQuery) setUpDefaultClient() error {
	var credentialsOption option.ClientOption

	ctx := context.Background()

	if b.CredentialsFile != "" {
		credentialsOption = option.WithCredentialsFile(b.CredentialsFile)
	} else {
		creds, err := google.FindDefaultCredentials(ctx)
		if err != nil {
			return fmt.Errorf(
				"unable to find GCP Application Default Credentials: %v."+
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

	for k, v := range groupedMetrics {
		if err := b.insertToTable(k, v); err != nil {
			return err
		}
	}

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

func (b *BigQuery) insertToTable(metricName string, metrics []bigquery.ValueSaver) error {
	ctx := context.Background()

	table := b.client.DatasetInProject(b.Project, b.Dataset).Table(metricName)
	inserter := table.Inserter()

	return inserter.Put(ctx, metrics)
}

func (b *BigQuery) tableForMetric(metricName string) *bigquery.Table {
	return b.client.DatasetInProject(b.Project, b.Dataset).Table(metricName)
}

// Close will terminate the session to the backend, returning error if an issue arises.
func (b *BigQuery) Close() error {
	return b.client.Close()
}

// SampleConfig returns the formatted sample configuration for the plugin.
func (s *BigQuery) SampleConfig() string {
	return sampleConfig
}

// Description returns the human-readable function definition of the plugin.
func (s *BigQuery) Description() string {
	return "Configuration for Google Cloud BigQuery to send entries"
}
