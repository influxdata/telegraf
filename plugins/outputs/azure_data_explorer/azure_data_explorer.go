package simpleoutput

// simpleoutput.go

import (
	"bytes"
	"context"

	"github.com/Azure/azure-kusto-go/kusto"
	"github.com/Azure/azure-kusto-go/kusto/ingest"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
)

type AzureDataExplorer struct {
	Ok           bool            `toml:"ok"`
	Endpoint     string          `toml:"endpoint_url"`
	Database     string          `toml:"database"`
	Table        string          `toml:"table"`
	ClientId     string          `toml:"client_id"`
	ClientSecret string          `toml:"client_secret"`
	TenantId     string          `toml:"tenant_id"`
	Log          telegraf.Logger `toml:"-"`
	Client       *kusto.Client
	Ingester     *ingest.Ingestion
	Serializer   serializers.Serializer
}

func (s *AzureDataExplorer) Description() string {
	return "Sends metrics to Azure Data Explorer"
}

func (s *AzureDataExplorer) SampleConfig() string {
	return `
  ok = true
`
}

func (s *AzureDataExplorer) Connect() error {
	// Make any connection required here
	authorizer := kusto.Authorization{
		Config: auth.NewClientCredentialsConfig(s.ClientId, s.ClientSecret, s.TenantId),
	}

	client, err := kusto.New(s.Endpoint, authorizer)

	if err != nil {
		return err
	}
	s.Client = client
	s.Ingester, err = ingest.New(client, s.Database, s.Table)

	if err != nil {
		return err
	}

	return nil
}

func (s *AzureDataExplorer) Close() error {
	// Close any connections here.
	// Write will not be called once Close is called, so there is no need to synchronize.
	return nil
}

func (s *AzureDataExplorer) Write(metrics []telegraf.Metric) error {
	result, err := s.Serializer.SerializeBatch(metrics)

	if err != nil {
		return err
	}

	reader := bytes.NewReader(result)
	s.Ingester.FromReader(context.TODO(), reader)

	// for _, metric := range metrics {
	// 	// write `metric` to the output sink here
	// }
	return nil
}

func init() {
	outputs.Add("azure_data_explorer", func() telegraf.Output {
		return &AzureDataExplorer{}
	})
}
