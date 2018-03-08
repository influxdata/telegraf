package azureTableStorage

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf16"

	storage "github.com/Azure/azure-sdk-for-go/storage"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
)

const (
	EmptyPayload    storage.MetadataLevel = ""
	NoMetadata      storage.MetadataLevel = "application/json;odata=nometadata"
	MinimalMetadata storage.MetadataLevel = "application/json;odata=minimalmetadata"
	FullMetadata    storage.MetadataLevel = "application/json;odata=fullmetadata"
)

type AzureTableStorage struct {
	AccountName  string
	AccountKey   string
	ResourceId   string
	DeploymentId string
	TableName    string
	table        *storage.Table
}

// NewBasicClient constructs a Client with given storage service name and
// key.
func NewBasicClient(accountName, accountKey string) (storage.Client, error) {
	// DefaultBaseURL is the domain name used for storage requests in the
	// public cloud when a default client is created.
	DefaultBaseURL := "core.windows.net"

	// DefaultAPIVersion is the Azure Storage API version string used when a
	// basic client is created.
	DefaultAPIVersion := "2016-05-31"

	defaultUseHTTPS := true
	return storage.NewClient(accountName, accountKey, DefaultBaseURL, DefaultAPIVersion, defaultUseHTTPS)
}

// getBasicClient returns a test client from storage credentials in the env
func getBasicClient(azureTableStorage *AzureTableStorage) *storage.Client {

	name := azureTableStorage.AccountName
	if name == "" {
		name = "ladextensionrgdiag526"
	}
	key := azureTableStorage.AccountKey
	if key == "" {
		key = "42WqyNltbP/S3rxbJizeelr4D35EUTU7en5QKgRotT6iWXZ7xtspB6j0/u5fs4kDaiheiIL8K9et0mdcBzcPig=="
	}
	cli, _ := NewBasicClient(name, key)
	//fmt.Print(err.Error())
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
	tableClient := azureStorageClient.GetTableService()
	azureTableStorage.TableName = "Sample2" //getTableName()
	azureTableStorage.table = tableClient.GetTableReference(azureTableStorage.TableName)
	er := azureTableStorage.table.Create(30, EmptyPayload, nil)
	if er != nil {
		fmt.Println("the table ", azureTableStorage.TableName, " already exists.")
	}
	return nil
}

func (azureTableStorage *AzureTableStorage) SampleConfig() string {
	return sampleConfig
}

func (azureTableStorage *AzureTableStorage) Description() string {
	return "Send telegraf metrics to file(s)"
}

func encodeSpecialCharacterToUTF16(decodedStr string) string {
	_pkey := ""
	hex := ""
	var replacer = strings.NewReplacer("[", ":", "]", "")
	for _, c := range decodedStr {
		if !unicode.IsLetter(c) && !unicode.IsDigit(c) {
			hex = fmt.Sprintf("%04X", utf16.Encode([]rune(string(c))))
			_pkey = _pkey + replacer.Replace(hex)
		} else {
			_pkey = _pkey + string(c)
		}
	}
	return _pkey
}

func (azureTableStorage *AzureTableStorage) Write(metrics []telegraf.Metric) error {
	rowKey := ""
	var entity *storage.Entity
	var props map[string]interface{}
	//TODO: generate partition key
	partitionKey := encodeSpecialCharacterToUTF16(azureTableStorage.ResourceId)
	// iterate over the list of metrics and create a new entity for each metrics and add to the table.
	for i, _ := range metrics {
		//TODO: generate row key
		rowKey = strconv.Itoa(i)
		//Get the reference of the entity by using Partition Key and row key. The combination needs to be unique for a new entity to be inserted.
		entity = azureTableStorage.table.GetEntityReference(partitionKey, rowKey)
		// add the map of metrics key n value to the entity in the field Properties.
		props = metrics[i].Fields()
		props["DeploymentID"] = azureTableStorage.DeploymentId
		props["Host"], _ = os.Hostname()
		entity.Properties = props
		entity.Insert(FullMetadata, nil)
	}
	return nil
}
func (azureTableStorage *AzureTableStorage) Close() error {
	return nil
}
func init() {
	outputs.Add("azureTableStorage", func() telegraf.Output {
		return &AzureTableStorage{}
	})
}
