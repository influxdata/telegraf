package simpleoutput

// simpleoutput.go

import (
	"bytes"
	"context"
	"fmt"

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
	ClientId     string          `toml:"client_id"`
	ClientSecret string          `toml:"client_secret"`
	TenantId     string          `toml:"tenant_id"`
	DataFormat   string          `toml:"data_format"`
	Log          telegraf.Logger `toml:"-"`
	Client       *kusto.Client
	Ingesters    map[string]*ingest.Ingestion
	Serializer   serializers.Serializer
}

const createTableCommand = `.create-merge table ['%s']  (['fields']:dynamic, ['name']:string, ['tags']:dynamic, ['timestamp']:datetime);`
const createTableMappingCommand = `.create-or-alter table ['%s'] ingestion json mapping '%s_mapping' '[{"column":"fields", "Properties":{"Path":"$[\'fields\']"}},{"column":"name", "Properties":{"Path":"$[\'name\']"}},{"column":"tags", "Properties":{"Path":"$[\'tags\']"}},{"column":"timestamp", "Properties":{"Path":"$[\'timestamp\']"}}]'`

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
		namespace := m.Name() // getNamespace(m)
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
			err := createAzureDataExplorerTableForNamespace(s.Client, s.Database, namespace)
			if err != nil {
				return err
			}

			//create a new ingestor client for the namespace
			s.Ingesters[namespace], err = ingest.New(s.Client, s.Database, namespace)
			if err != nil {
				return err
			}
		}
	}

	for key, mPerNamespace := range metricsPerNamespace {
		reader := bytes.NewReader(mPerNamespace)

		_, error := s.Ingesters[key].FromReader(context.TODO(), reader, ingest.FileFormat(ingest.JSON), ingest.IngestionMappingRef(fmt.Sprintf("%s_mapping", key), ingest.JSON))
		if error != nil {
			s.Log.Errorf("error sending ingestion request to Azure Data Explorer for metric %s: %v", key, error)
		}
	}
	return nil
}

func createAzureDataExplorerTableForNamespace(client *kusto.Client, database string, tableName string) error {

	// Create a database
	createStmt := kusto.NewStmt("", kusto.UnsafeStmt(unsafe.Stmt{Add: true})).UnsafeAdd(fmt.Sprintf(createTableCommand, tableName))
	_, errCreatingTable := client.Mgmt(context.TODO(), database, createStmt)
	if errCreatingTable != nil {
		return errCreatingTable
	}

	createTableMappingstmt := kusto.NewStmt("", kusto.UnsafeStmt(unsafe.Stmt{Add: true})).UnsafeAdd(fmt.Sprintf(createTableMappingCommand, tableName, tableName))
	_, errCreatingTableMapping := client.Mgmt(context.TODO(), database, createTableMappingstmt)
	if errCreatingTableMapping != nil {
		return errCreatingTableMapping
	}

	return nil
}

// // This is to group metrics based on the convention of having a hyphen in the metric name. It complies with Azure Monitor way of metric categorization.
// func getNamespace(m telegraf.Metric) string {
// 	names := strings.SplitN(m.Name(), "-", 2)
// 	return names[0]
// }

func (s *AzureDataExplorer) Init() error {
	if s.DataFormat != "json" {
		return fmt.Errorf("the azure data explorer supports json data format only, pleaes make sure to add the 'data_format=\"json\"' in the output configuration")
	}
	return nil
}

func init() {
	outputs.Add("azure_data_explorer", func() telegraf.Output {
		return &AzureDataExplorer{}
	})
}

func (s *AzureDataExplorer) SetSerializer(serializer serializers.Serializer) {
	s.Serializer = serializer
}
