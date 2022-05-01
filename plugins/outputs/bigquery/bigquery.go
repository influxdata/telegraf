package bigquery

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"cloud.google.com/go/bigquery"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/outputs"
)

const timeStampFieldName = "timestamp"

var defaultTimeout = config.Duration(5 * time.Second)

type BigQuery struct {
	CredentialsFile string `toml:"credentials_file"`
	Project         string `toml:"project"`
	Dataset         string `toml:"dataset"`

	Timeout         config.Duration `toml:"timeout"`
	ReplaceHyphenTo string          `toml:"replace_hyphen_to"`

	Log telegraf.Logger `toml:"-"`

	client *bigquery.Client

	warnedOnHyphens map[string]bool
}

func (s *BigQuery) Connect() error {
	if s.Project == "" {
		return fmt.Errorf("Project is a required field for BigQuery output")
	}

	if s.Dataset == "" {
		return fmt.Errorf("Dataset is a required field for BigQuery output")
	}

	if s.client == nil {
		return s.setUpDefaultClient()
	}

	s.warnedOnHyphens = make(map[string]bool)

	return nil
}

func (s *BigQuery) setUpDefaultClient() error {
	var credentialsOption option.ClientOption

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Duration(s.Timeout))
	defer cancel()

	if s.CredentialsFile != "" {
		credentialsOption = option.WithCredentialsFile(s.CredentialsFile)
	} else {
		creds, err := google.FindDefaultCredentials(ctx)
		if err != nil {
			return fmt.Errorf(
				"unable to find Google Cloud Platform Application Default Credentials: %v. "+
					"Either set ADC or provide CredentialsFile config", err)
		}
		credentialsOption = option.WithCredentials(creds)
	}

	client, err := bigquery.NewClient(ctx, s.Project, credentialsOption)
	s.client = client
	return err
}

// Write the metrics to Google Cloud BigQuery.
func (s *BigQuery) Write(metrics []telegraf.Metric) error {
	groupedMetrics := s.groupByMetricName(metrics)

	var wg sync.WaitGroup

	for k, v := range groupedMetrics {
		wg.Add(1)
		go func(k string, v []bigquery.ValueSaver) {
			defer wg.Done()
			s.insertToTable(k, v)
		}(k, v)
	}

	wg.Wait()

	return nil
}

func (s *BigQuery) groupByMetricName(metrics []telegraf.Metric) map[string][]bigquery.ValueSaver {
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

func (s *BigQuery) insertToTable(metricName string, metrics []bigquery.ValueSaver) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Duration(s.Timeout))
	defer cancel()

	tableName := s.metricToTable(metricName)
	table := s.client.DatasetInProject(s.Project, s.Dataset).Table(tableName)
	inserter := table.Inserter()

	if err := inserter.Put(ctx, metrics); err != nil {
		s.Log.Errorf("inserting metric %q failed: %v", metricName, err)
	}
}

func (s *BigQuery) metricToTable(metricName string) string {
	if !strings.Contains(metricName, "-") {
		return metricName
	}

	dhm := strings.ReplaceAll(metricName, "-", s.ReplaceHyphenTo)

	if warned := s.warnedOnHyphens[metricName]; !warned {
		s.Log.Warnf("Metric %q contains hyphens please consider using the rename processor plugin, falling back to %q", metricName, dhm)
		s.warnedOnHyphens[metricName] = true
	}

	return dhm
}

// Close will terminate the session to the backend, returning error if an issue arises.
func (s *BigQuery) Close() error {
	return s.client.Close()
}

func init() {
	outputs.Add("bigquery", func() telegraf.Output {
		return &BigQuery{
			Timeout:         defaultTimeout,
			ReplaceHyphenTo: "_",
		}
	})
}
