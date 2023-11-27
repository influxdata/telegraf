//go:generate ../../../tools/readme_config_includer/generator
package bigquery

import (
	"context"
	_ "embed"
	"errors"
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
	"github.com/influxdata/telegraf/plugins/serializers/json"
)

//go:embed sample.conf
var sampleConfig string

const timeStampFieldName = "timestamp"

var defaultTimeout = config.Duration(5 * time.Second)

type BigQuery struct {
	CredentialsFile string `toml:"credentials_file"`
	Project         string `toml:"project"`
	Dataset         string `toml:"dataset"`

	Timeout          config.Duration `toml:"timeout"`
	ReplaceHyphenTo  string          `toml:"replace_hyphen_to"`
	CompactTable     bool            `toml:"compact_table"`
	CompactTableName string          `toml:"compact_table_name"`

	Log telegraf.Logger `toml:"-"`

	client     *bigquery.Client
	serializer json.Serializer

	warnedOnHyphens map[string]bool
}

func (*BigQuery) SampleConfig() string {
	return sampleConfig
}

func (s *BigQuery) Init() error {
	if s.Project == "" {
		s.Project = bigquery.DetectProjectID
	}

	if s.Dataset == "" {
		return errors.New(`"dataset" is required`)
	}

	if s.CompactTable && s.CompactTableName == "" {
		return errors.New(`"compact_table_name" is required`)
	}

	s.warnedOnHyphens = make(map[string]bool)

	return s.serializer.Init()
}

func (s *BigQuery) Connect() error {
	if s.client == nil {
		if err := s.setUpDefaultClient(); err != nil {
			return err
		}
	}

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Duration(s.Timeout))
	defer cancel()

	// Check if the compact table exists
	_, err := s.client.DatasetInProject(s.Project, s.Dataset).Table(s.CompactTableName).Metadata(ctx)
	if s.CompactTable && err != nil {
		return fmt.Errorf("compact table: %w", err)
	}
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
				"unable to find Google Cloud Platform Application Default Credentials: %w. "+
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
	if s.CompactTable {
		return s.writeCompact(metrics)
	}

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

func (s *BigQuery) writeCompact(metrics []telegraf.Metric) error {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Duration(s.Timeout))
	defer cancel()

	inserter := s.client.DatasetInProject(s.Project, s.Dataset).Table(s.CompactTableName).Inserter()

	compactValues := make([]*bigquery.ValuesSaver, len(metrics))
	for i, m := range metrics {
		compactValues[i] = s.newCompactValuesSaver(m)
	}
	return inserter.Put(ctx, compactValues)
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

func (s *BigQuery) newCompactValuesSaver(m telegraf.Metric) *bigquery.ValuesSaver {
	s.serializer.Transformation = "tags"
	tags, err := s.serializer.Serialize(m)
	if err != nil {
		s.Log.Warnf("serializing tags: %v", err)
	}

	s.serializer.Transformation = "fields"
	fields, err := s.serializer.Serialize(m)
	if err != nil {
		s.Log.Warnf("serializing fields: %v", err)
	}

	return &bigquery.ValuesSaver{
		Schema: bigquery.Schema{
			timeStampFieldSchema(),
			newStringFieldSchema("name"),
			newJSONFieldSchema("tags"),
			newJSONFieldSchema("fields"),
		},
		Row: []bigquery.Value{
			m.Time(),
			m.Name(),
			string(tags),
			string(fields),
		},
	}
}

func timeStampFieldSchema() *bigquery.FieldSchema {
	return &bigquery.FieldSchema{
		Name: timeStampFieldName,
		Type: bigquery.TimestampFieldType,
	}
}

func newStringFieldSchema(name string) *bigquery.FieldSchema {
	return &bigquery.FieldSchema{
		Name: name,
		Type: bigquery.StringFieldType,
	}
}

func newJSONFieldSchema(name string) *bigquery.FieldSchema {
	return &bigquery.FieldSchema{
		Name: name,
		Type: bigquery.JSONFieldType,
	}
}

func tagsSchemaAndValues(m telegraf.Metric, s bigquery.Schema, r []bigquery.Value) ([]*bigquery.FieldSchema, []bigquery.Value) {
	for _, t := range m.TagList() {
		s = append(s, newStringFieldSchema(t.Key))
		r = append(r, t.Value)
	}

	return s, r
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
