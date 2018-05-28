package azuretablestorage

import (
	"errors"
	"log"
	"strings"
	"time"

	storage "github.com/Azure/azure-sdk-for-go/storage"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
	util "github.com/influxdata/telegraf/utility"
)

const (
	EmptyPayload    storage.MetadataLevel = ""
	NoMetadata      storage.MetadataLevel = "application/json;odata=nometadata"
	MinimalMetadata storage.MetadataLevel = "application/json;odata=minimalmetadata"
	FullMetadata    storage.MetadataLevel = "application/json;odata=fullmetadata"
)

type TableNameVsTableRef struct {
	TableName string         //TableName: name of Azure Table
	TableRef  *storage.Table //TableRef: reference of the Azure Table client object
}
type AzureTableStorage struct {
	AccountName                 string //azure storage account name
	SasToken                    string
	ResourceId                  string //resource id for the VM or VMSS
	DeploymentId                string
	Periods                     []string                       //this is the list of periods being configured for various aggregator instances.
	periodVsTableNameVsTableRef map[string]TableNameVsTableRef //Map of transfer period of metrics vs table name and table client ref.
	prevTableNameSuffix         string
	TableStorageEndPointSuffix  string
	HostName                    string
	columnsInTable              []string
	Protocol                    string
}

var sampleConfig = `
[[outputs.azuretablestorage]]
deployment_id = ""
resource_id = ""
account_name = ""
sas_token = ""
#periods is the list of period configured for each aggregator plugin
periods = ["30s","60s"] 
table_storage_end_point_suffix = ".table.core.windows.net"
host_name = ""
protocol = ""

`

//RETURNS:a map of <Period,{TableName, TableClientReference}>
func (azureTableStorage *AzureTableStorage) getAzureperiodVsTableNameVsTableRefMap(secondsElapsedTillNow int64,
	tableClient storage.TableServiceClient) (map[string]TableNameVsTableRef, error) {

	periodVsTableNameVsTableRef := map[string]TableNameVsTableRef{}

	tableNameSuffix, err := getTableDateSuffix(secondsElapsedTillNow)
	if err != nil {
		log.Println("Error while constructing suffix for azure table name")
		return periodVsTableNameVsTableRef, err
	}
	//Empty the map of tables every 10th day as they become obsolete now.
	if azureTableStorage.prevTableNameSuffix != tableNameSuffix {
		azureTableStorage.periodVsTableNameVsTableRef = map[string]TableNameVsTableRef{}
		azureTableStorage.prevTableNameSuffix = tableNameSuffix
	}

	for _, period := range azureTableStorage.Periods {
		periodStr, err := util.GetPeriodStr(period)
		if err != nil {
			log.Println("Error while parsing scheduled transfer period for metrics to the table: " + period)
			return periodVsTableNameVsTableRef, err
		}
		tableName := util.WAD_METRICS + periodStr + util.P10DV25 + tableNameSuffix
		table := tableClient.GetTableReference(tableName)
		if table.Name == "" {
			logMsg := "Error while getting table reference for table " + tableName
			log.Println(logMsg)
			return periodVsTableNameVsTableRef, errors.New(logMsg)
		}
		tableNameVsTableRefObj := TableNameVsTableRef{TableName: tableName, TableRef: table}
		periodVsTableNameVsTableRef[period] = tableNameVsTableRefObj
	}
	return periodVsTableNameVsTableRef, nil
}
func (azureTableStorage *AzureTableStorage) initDefaults() {
	azureTableStorage.Protocol = "https://"
	azureTableStorage.columnsInTable = []string{}
	azureTableStorage.columnsInTable = append(azureTableStorage.columnsInTable, util.MEAN)
	azureTableStorage.columnsInTable = append(azureTableStorage.columnsInTable, util.SAMPLE_COUNT)
	azureTableStorage.columnsInTable = append(azureTableStorage.columnsInTable, util.COUNTER_NAME)
	azureTableStorage.columnsInTable = append(azureTableStorage.columnsInTable, util.DEPLOYMENT_ID)
	azureTableStorage.columnsInTable = append(azureTableStorage.columnsInTable, util.HOST)
	azureTableStorage.columnsInTable = append(azureTableStorage.columnsInTable, util.LAST_SAMPLE)
	azureTableStorage.columnsInTable = append(azureTableStorage.columnsInTable, util.MAX_SAMPLE)
	azureTableStorage.columnsInTable = append(azureTableStorage.columnsInTable, util.MIN_SAMPLE)
	azureTableStorage.columnsInTable = append(azureTableStorage.columnsInTable, util.BEGIN_TIMESTAMP)
	azureTableStorage.columnsInTable = append(azureTableStorage.columnsInTable, util.END_TIMESTAMP)
	azureTableStorage.columnsInTable = append(azureTableStorage.columnsInTable, util.TOTAL)
}

//TODO: if Connect() fails on multiple retries then log message, skip this plugin and continue with rest of the sinks.
func (azureTableStorage *AzureTableStorage) Connect() error {
	azureTableStorage.initDefaults()
	sasUrl := azureTableStorage.Protocol + azureTableStorage.AccountName + azureTableStorage.TableStorageEndPointSuffix
	client, er := storage.NewAccountSASClientFromEndpointToken(sasUrl, azureTableStorage.SasToken)

	if er != nil {
		log.Println("Error in getting table storage client ")
		return er
	}

	tableClient := client.GetTableService()
	er = validateTableClient(tableClient)
	if er != nil {
		log.Println("Error while creating table client ")
		return er
	}

	//secondsElapsedTillNow: number of seconds elapsed till now from 1 Jan 1970.
	secondsElapsedTillNow := time.Now().Unix()
	azureTableStorage.periodVsTableNameVsTableRef, er =
		azureTableStorage.getAzureperiodVsTableNameVsTableRefMap(secondsElapsedTillNow, tableClient)

	if er != nil {
		log.Println("Error while constructing map of <period , <tableName, tableClient>>")
		return er
	}

	//create all the tables
	for _, tableVsTableRef := range azureTableStorage.periodVsTableNameVsTableRef {
		er := tableVsTableRef.TableRef.Create(30, FullMetadata, nil)
		if er != nil && strings.Contains(er.Error(), "TableAlreadyExists") {
			log.Println("the table ", tableVsTableRef.TableName, " already exists.")
		} else if er != nil {
			log.Println("Error while creating table " + tableVsTableRef.TableName)
			log.Println(er.Error())
			return er
		}
	}
	return nil
}

func (azureTableStorage *AzureTableStorage) SampleConfig() string {
	return sampleConfig
}

func (azureTableStorage *AzureTableStorage) Description() string {
	return "Sends metrics collected by input plugin to azure storage tables"
}

func (azureTableStorage *AzureTableStorage) Write(metrics []telegraf.Metric) error {

	var props map[string]interface{}
	partitionKey := getPartitionKey(azureTableStorage.ResourceId)

	// iterate over the list of metrics and create a new entity for each metrics and add to the table.
	for i, _ := range metrics {
		props = metrics[i].Fields()
		props[util.DEPLOYMENT_ID] = azureTableStorage.DeploymentId
		props[util.HOST] = azureTableStorage.HostName

		tags := metrics[i].Tags()
		props[util.COUNTER_NAME] = tags[util.INPUT_PLUGIN] + "/" + props[util.COUNTER_NAME].(string)

		UTCTicks_DescendingOrderStr, encodedCounterName, er := getRowKeyComponents(props[util.END_TIMESTAMP].(string),
			props[util.COUNTER_NAME].(string))
		if er != nil {
			log.Println("Error: Unable to get valid row key components. Since, this cannot be corrected even on retries hence skipping this row." + util.GetPropsStr(props))
			continue
		}

		periodStr := tags[util.PERIOD]
		table := azureTableStorage.periodVsTableNameVsTableRef[periodStr].TableRef

		//don't write incomplete rows to the table storage
		isValidRow := validateRow(azureTableStorage.columnsInTable, props)
		if isValidRow == false {
			logMsg := "Invalid Row hence not writing it to the table. Row values : " + util.GetPropsStr(props)
			log.Println(logMsg)
			continue
		}
		//two rows are written for each metric as Azure table has optimized prefix search only and no index.
		rowKey1 := UTCTicks_DescendingOrderStr + "_" + encodedCounterName
		rowKey2 := encodedCounterName + "_" + UTCTicks_DescendingOrderStr
		er = writeEntitiesToTable(partitionKey, rowKey1, rowKey2, props, table)
		if er != nil {
			log.Println("Error occured while writing entities to the table")
			return er
		}
	}
	return nil
}

func (azureTableStorage *AzureTableStorage) Close() error {
	return nil
}

func init() {
	outputs.Add("azuretablestorage", func() telegraf.Output {
		return &AzureTableStorage{}
	})
}
