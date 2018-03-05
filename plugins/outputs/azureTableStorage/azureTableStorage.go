package azureTableStorage

import (
	"fmt"
	"io"
	"os"

	"github.com/influxdata/telegraf"
	"github.com/Azure/azure-storage-go"
	"github.com/influxdata/telegraf/plugins/outputs"
)

// DefaultBaseURL is the domain name used for storage requests in the
// public cloud when a default client is created.
DefaultBaseURL = "core.windows.net"

// DefaultAPIVersion is the Azure Storage API version string used when a
// basic client is created.
DefaultAPIVersion = "2016-05-31"

defaultUseHTTPS      = true


type AzureTableStorage struct {
	string AccountName 
	string AccountKey
	string ResourceId
	string DeploymentId
    TableServiceClient * table
}

// NewBasicClient constructs a Client with given storage service name and
// key.
func NewBasicClient(accountName, accountKey string) (Client, error) {
	return NewClient(accountName, accountKey, DefaultBaseURL, DefaultAPIVersion, defaultUseHTTPS)
}

// getBasicClient returns a test client from storage credentials in the env
func getBasicClient(azureTableStorage) *Client {

	name := azureTableStorage.AccountName
	if name == "" {
		name = "ladextensionrgdiag526"
	}
	key := azureTableStorage.AccountKey
	if key == "" {
		key = "42WqyNltbP/S3rxbJizeelr4D35EUTU7en5QKgRotT6iWXZ7xtspB6j0/u5fs4kDaiheiIL8K9et0mdcBzcPig=="
	}
	cli, err := NewBasicClient(name, key)

	return &cli
}


var sampleConfig = `
  ## Files to write to, "stdout" is a specially handled file.
  files = ["stdout", "/tmp/metrics.out"]

  ## Data format to output.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "influx"
`

func (azureTableStorage *AzureTableStorage) Connect() error {
	//create a new client with NewClient() it will retuen a client object
	azureStorageClient := getBasicClient(azureTableStorage)
	// GetTableService returns TableServiceClient
	tableClient := azureStorageClient.GetTableService(azureTableStorage)
    azureTableStorage.TableName := "Sample1"//getTableName()
	azureTableStorage.table = tableClient.GetTableReference(azureTableStorage.TableName)
	er := azureStorage.table.Create(30, EmptyPayload, nil)
	if er != nil {
		fmt.Println("the table ", azureTableStorage.TableName, " already exists.")
	}
	return nil
}

func (azureTableStorage *azureTableStorage) SampleConfig() string {
	return sampleConfig
}

func (azureTableStorage *azureTableStorage) Description() string {
	return "Send telegraf metrics to file(s)"
}

func (azureTableStorage *azureTableStorage) Write(metrics []telegraf.Metric) error {
	rowKey := ""
	var entity *storage.Entity
	var props map[string]interface{}
	//TODO: generate partition key
	partitionKey:="CPU_metrics"
	// iterate over the list of metrics and create a new entity for each metrics and add to the table.
	for i, _ := range metrics {
		//TODO: generate row key
		rowKey = strconv.Itoa(i)
		//Get the reference of the entity by using Partition Key and row key. The combination needs to be unique for a new entity to be inserted.
		entity = azureStorage.table.GetEntityReference(partitionKey, rowKey)
		// add the map of metrics key n value to the entity in the field Properties.
		props = metrics[i].Fields()
		entity.Properties = props
		entity.Insert(FullMetadata, nil)
	}
	return nil
}

func init() {
	outputs.Add("azureTableStorage", func() telegraf.Output {
		return &AzureTableStorage{}
	})
}
