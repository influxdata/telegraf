package bigquery

import (
	"context"
	"fmt"
	"os"

	"cloud.google.com/go/bigquery"
	bigqueryt "cloud.google.com/go/bigquery"
	"github.com/influxdata/telegraf"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
)

const (
	testingHostEnv   = "BIGQUERY_TESTING_HOST"
	defaultOffSetKey = "offset-key.json"
)

type BigQuery struct {
	CredentialsFile string `toml:"credentials_file"`
	Project         string `toml:"project"`
	Dataset         string `toml:"dataset"`

	client *bigqueryt.Client
}

type BigQueryMetric struct {
	metric telegraf.Metric
}

func (bm *BigQueryMetric) Save() (map[string]bigquery.Value, string, error) {
	mapValue := make(map[string]bigquery.Value)

	mapValue["timestamp"] = bm.metric.Time

	for _, tag := range bm.metric.TagList() {
		mapValue[tag.Key] = tag.Value
	}

	for _, field := range bm.metric.FieldList() {
		mapValue[field.Key] = field.Value
	}

	return mapValue, "on-purpose", nil
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

	if client, err := bigqueryt.NewClient(ctx, b.Project, noAuth, endpoints); err != nil {
		return err
	} else {
		b.client = client
		return nil
	}
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

	client, err := bigqueryt.NewClient(ctx, b.Project, credentialsOption)
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

func (b *BigQuery) groupByMetricName(metrics []telegraf.Metric) map[string][]BigQueryMetric {
	groupedMetrics := make(map[string][]BigQueryMetric)

	for _, m := range metrics {
		bqm := BigQueryMetric{metric: m}
		groupedMetrics[m.Name()] = append(groupedMetrics[m.Name()], bqm)
	}

	return groupedMetrics
}

func (b *BigQuery) insertToTable(metricName string, metrics []BigQueryMetric) error {
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
