package azureTableStorage

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
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
const layout = "02/01/2006 03:04:05 PM"

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

func encodeSpecialCharacterToUTF16(resourceId string) string {
	_pkey := ""
	hex := ""
	var replacer = strings.NewReplacer("[", ":", "]", "")
	for _, c := range resourceId {
		if !unicode.IsLetter(c) && !unicode.IsDigit(c) {
			hex = fmt.Sprintf("%04X", utf16.Encode([]rune(string(c))))
			_pkey = _pkey + replacer.Replace(hex)
		} else {
			_pkey = _pkey + string(c)
		}
	}
	return _pkey
}
func getPrimaryKey(resourceId string) string {
	return encodeSpecialCharacterToUTF16(resourceId)
}

func getUTCTicks_DescendingOrder(lastSampleTimestamp string) int64 {

	currentTime := time.Now().UTC()
	//maxValueDateTime := time.Date(9999, time.December, 31, 12, 59, 59, 59, time.UTC)
	//Ticks is the number of 100 nanoseconds from zero value of date
	maxValueDateTimeInTicks := int64(3155378975999999999)
	zeroTime := time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC)
	diff := currentTime.Sub(zeroTime)
	currentTimeInTicks := int64(diff.Nanoseconds()) / 100
	UTCTicks_DescendincurrentTimeOrder := maxValueDateTimeInTicks - currentTimeInTicks
	fmt.Println(UTCTicks_DescendincurrentTimeOrder)

	return UTCTicks_DescendincurrentTimeOrder
}

func getRowKeyComponents(lastSampleTimestamp string, counterName string) (string, string) {

	UTCTicks_DescendingOrder := getUTCTicks_DescendingOrder(lastSampleTimestamp)
	UTCTicks_DescendingOrderStr := strconv.FormatInt(UTCTicks_DescendingOrder, 10)
	encodedCounterName := encodeSpecialCharacterToUTF16(counterName)
	return UTCTicks_DescendingOrderStr, encodedCounterName
}

func (azureTableStorage *AzureTableStorage) Write(metrics []telegraf.Metric) error {
	var entity *storage.Entity
	var props map[string]interface{}
	partitionKey := getPrimaryKey(azureTableStorage.ResourceId)
	// iterate over the list of metrics and create a new entity for each metrics and add to the table.
	for i, _ := range metrics {
		props = metrics[i].Fields()
		UTCTicks_DescendingOrderStr, encodedCounterName := getRowKeyComponents(props["TIMESTAMP"].(string), props["CounterName"].(string))
		props["DeploymentId"] = azureTableStorage.DeploymentId
		props["Host"], _ = os.Hostname()

		rowKey1 := UTCTicks_DescendingOrderStr + "_" + encodedCounterName
		entity = azureTableStorage.table.GetEntityReference(partitionKey, rowKey1)
		entity.Properties = props
		entity.Insert(FullMetadata, nil)

		rowKey2 := encodedCounterName + "_" + UTCTicks_DescendingOrderStr
		entity = azureTableStorage.table.GetEntityReference(partitionKey, rowKey2)
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
