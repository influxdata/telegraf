package simpleoutput

// simpleoutput.go

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/Azure/azure-kusto-go/kusto"
	"github.com/Azure/azure-kusto-go/kusto/ingest"
	"github.com/Azure/azure-kusto-go/kusto/unsafe"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
)

type AzureDataExplorer struct {
	Endpoint     string          `toml:"endpoint_url"`
	Database     string          `toml:"database"`
	Table        string          `toml:"table"`
	ClientId     string          `toml:"client_id"`
	ClientSecret string          `toml:"client_secret"`
	TenantId     string          `toml:"tenant_id"`
	Log          telegraf.Logger `toml:"-"`
	Client       *kusto.Client
	Ingesters    map[string]*ingest.Ingestion
	Ingester     *ingest.Ingestion
	Serializer   serializers.Serializer
}

const createTableCommand = `.create table ['%s']  (['fields']:dynamic, ['name']:string, ['tags']:dynamic, ['timestamp']:datetime)`

func (s *AzureDataExplorer) Description() string {
	return "Sends metrics to Azure Data Explorer"
}

func (s *AzureDataExplorer) SampleConfig() string {
	return `
  ## Azure Data Exlorer cluster endpoint
  ## ex: endpoint_url = "https://clustername.australiasoutheast.kusto.windows.net"
  # endpoint_url = ""
  
  ## The name of the database in Azure Data Explorer where the ingestion will happen
  # database = ""

  ## The name of the table in Azure Data Explorer where the ingestion will happen
  # table = ""

  ## The client ID of the Service Principal in Azure that has ingestion rights to the Azure Data Exploer Cluster
  # client_id = ""

  ## The client secret of the Service Principal in Azure that has ingestion rights to the Azure Data Exploer Cluster

  # client_secret = ""

  ## The tenant ID of the Azure Subsciption in which the Service Principal belongs to
  # tenant_id = ""
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
	s.Ingesters = make(map[string]*ingest.Ingestion)
	s.Ingester, _ = ingest.New(client, s.Database, s.Table)

	return nil
}

func (s *AzureDataExplorer) Close() error {

	s.Client = nil
	s.Ingesters = nil

	return nil
}

func (s *AzureDataExplorer) Write(metrics []telegraf.Metric) error {

	metricsPerNamespace := make(map[string][]byte)

	for _, m := range metrics {
		namespace := getNamespace(m)
		metricInBytes, err := s.Serializer.Serialize(m)
		if err != nil {
			return err
		}

		if existingBytes, ok := metricsPerNamespace[namespace]; ok {
			metricsPerNamespace[namespace] = append(existingBytes, metricInBytes...)
		} else {
			metricsPerNamespace[namespace] = metricInBytes
		}

		if _, ingestorExist := s.Ingesters[namespace]; !ingestorExist {
			//create a table for the namespace
			// err := createAzureDataExplorerTableForNamespace(s.Client, s.Database, namespace)
			// if err != nil {
			// 	return err
			// }

			//create a new ingestor client for the namespace
			s.Ingesters[namespace], err = ingest.New(s.Client, s.Database, namespace)
			if err != nil {
				return err
			}
		}
	}

	for key, mPerNamespace := range metricsPerNamespace {
		reader := bytes.NewReader(mPerNamespace)

		_, error := s.Ingesters[key].FromReader(context.TODO(), reader, ingest.FileFormat(ingest.JSON), ingest.IngestionMappingRef("metrics_mapping", ingest.JSON))
		if error != nil {
			s.Log.Errorf("error sending ingestion request to Azure Data Explorer: %v", error)
			return error
		}
	}
	return nil
}

func createAzureDataExplorerTableForNamespace(client *kusto.Client, database string, tableName string) error {

	stmt := kusto.NewStmt("", kusto.UnsafeStmt(unsafe.Stmt{Add: true}))
	stmt.UnsafeAdd(fmt.Sprintf(createTableCommand, tableName))
	_, err := client.Mgmt(context.TODO(), database, stmt)
	if err != nil {
		return err
	}
	return nil
}

func getNamespace(m telegraf.Metric) string {
	names := strings.SplitN(m.Name(), "-", 2)
	return names[0]
}

func init() {
	outputs.Add("azure_data_explorer", func() telegraf.Output {
		return &AzureDataExplorer{}
	})
}

func (s *AzureDataExplorer) SetSerializer(serializer serializers.Serializer) {
	s.Serializer = serializer
}
