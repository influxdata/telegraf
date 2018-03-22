package azuretablestorage

import (
	"fmt"
	"math"
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

type TableNameVsTableRef struct {
	TableName string         //TableName: name of Azure Table
	TableRef  *storage.Table //TableRef: reference of the Azure Table client object
}
type AzureTableStorage struct {
	AccountName                 string //azure storage account name.
	SasUrl                      string //endpoint for table service
	SasToken                    string
	ResourceId                  string //resource id for the VM or VMSS
	DeploymentId                string
	Periods                     []string                       //this is the list of periods being configured for various aggregator instances.
	PeriodVsTableNameVsTableRef map[string]TableNameVsTableRef //Map of transfer period of metrics vs table name and table client ref.
	PrevTableNameSuffix         string
}

var sampleConfig = `
# Configuration for azuretablestorage output plugin to write metrics to azure table
[[outputs.azuretablestorage]]
 deploymentId = "deploymentId"
 resourceId = "subscriptionId/resourceGroup/VMScaleset"
 accountName = "ladextensionrgdiag526"
 sasURL = "https://ladextensionrgdiag526.table.core.windows.net"
 sasToken="sv=2017-07-29&ss=bt&srt=sco&sp=rwdlacu&se=2019-03-20T19:34:18Z&st=2018-03-19T11:34:18Z&spr=https&sig=tw%2BfX8RJw%2FLd7%2Fv5K1w4b2bOJwBAPcqkUsFqBB7LllQ%3D"
 #periods is the list of period configured for each aggregator plugin
 periods = ["30s","60s"] 

`

//New tables are required to be created every 10th Day. And date suffix changes in the new table name.
//RETURNS: the date of the last day of the last 10 day interval.
func getTableDateSuffix() string {

	//get number of seconds elapsed till now from 1 Jan 1970.
	secondsElapsedTillNow := time.Now().Unix()

	//get the number of seconds as multiple of number of seconds in 10 days
	secondsIn10Day := int64(10 * 24 * 60 * 60)
	secondsElapsedTillNow10DayMulTiple := secondsElapsedTillNow - (secondsElapsedTillNow % secondsIn10Day)

	//get the date from the value of number of seconds obtained by above equation. This date
	// will be the last day of the previous 10 day period.
	suffixDate := time.Unix(secondsElapsedTillNow10DayMulTiple, 0)
	suffixDateStr := suffixDate.Format(DATE_SUFFIX_FORMAT)
	suffixDateStr = strings.Replace(suffixDateStr, "/", "", -1)

	// the name of the table will have this date as its suffix.
	return suffixDateStr
}

//period is in the format "60s"
//RETURNS: period in the format "PT1M"
func getPeriodStr(period string) string {

	var periodStr string

	totalSeconds, _ := strconv.Atoi(strings.Trim(period, "s"))
	hour := (int)(math.Floor(float64(totalSeconds) / 3600))
	min := int(math.Floor(float64(totalSeconds-(hour*3600)) / 60))
	sec := totalSeconds - (hour * 3600) - (min * 60)
	periodStr = PT
	if hour > 0 {
		periodStr += strconv.Itoa(hour) + H
	}
	if min > 0 {
		periodStr += strconv.Itoa(min) + M
	}
	if sec > 0 {
		periodStr += strconv.Itoa(sec) + S
	}
	return periodStr
}

//RETURNS:a map of <Period,{TableName, TableClientReference}>
func getAzurePeriodVsTableNameVsTableRefMap(azureTableStorage *AzureTableStorage,
	tableClient storage.TableServiceClient) map[string]TableNameVsTableRef {

	periodVsTableNameVsTableRef := map[string]TableNameVsTableRef{}

	//Empty the list of tables every 10th day as they become obsolete now.
	tableNameSuffix := getTableDateSuffix()
	if azureTableStorage.PrevTableNameSuffix != tableNameSuffix {
		azureTableStorage.PeriodVsTableNameVsTableRef = map[string]TableNameVsTableRef{}
		azureTableStorage.PrevTableNameSuffix = tableNameSuffix
	}

	for _, period := range azureTableStorage.Periods {
		periodStr := getPeriodStr(period)
		tableName := WADMetrics + periodStr + P10DV25 + tableNameSuffix
		table := tableClient.GetTableReference(tableName)
		tableNameVsTableRefObj := TableNameVsTableRef{TableName: tableName, TableRef: table}
		periodVsTableNameVsTableRef[period] = tableNameVsTableRefObj
	}
	return periodVsTableNameVsTableRef
}

func (azureTableStorage *AzureTableStorage) Connect() error {
	client, er := storage.NewAccountSASClientFromEndpointToken(azureTableStorage.SasUrl, azureTableStorage.SasToken)
	if er != nil {
		return er
	}
	tableClient := client.GetTableService()
	azureTableStorage.PeriodVsTableNameVsTableRef =
		getAzurePeriodVsTableNameVsTableRefMap(azureTableStorage, tableClient)

	for _, tableVsTableRef := range azureTableStorage.PeriodVsTableNameVsTableRef {
		er := tableVsTableRef.TableRef.Create(30, FullMetadata, nil)
		if strings.Contains(er.Error(), "TableAlreadyExists") {
			fmt.Println("the table ", tableVsTableRef.TableName, " already exists.")
		} else {
			return er
		}
	}
	return nil
}

func (azureTableStorage *AzureTableStorage) SampleConfig() string {
	return sampleConfig
}

func (azureTableStorage *AzureTableStorage) Description() string {
	return "Sends telegraf metrics to azure storage tables"
}

//RETURNS: resourceId encoded by converting any letter other than alphanumerics to unicode as per UTF-16
func encodeSpecialCharacterToUTF16(resourceId string) string {
	encodedStr := ""
	hex := ""
	var replacer = strings.NewReplacer("[", ":", "]", "")
	for _, c := range resourceId {
		if !unicode.IsLetter(c) && !unicode.IsDigit(c) {
			hex = fmt.Sprintf("%04X", utf16.Encode([]rune(string(c))))
			encodedStr = encodedStr + replacer.Replace(hex)
		} else {
			encodedStr = encodedStr + string(c)
		}
	}
	return encodedStr
}

//RETURNS: primary key for the azure table.
func getPrimaryKey(resourceId string) string {
	return encodeSpecialCharacterToUTF16(resourceId)
}

//RETURNS: difference of max value that can be held by time and number of 100 ns in current time.
func getUTCTicks_DescendingOrder(lastSampleTimestamp string) uint64 {

	currentTime, _ := time.Parse(layout, lastSampleTimestamp)
	//maxValueDateTime := time.Date(9999, time.December, 31, 12, 59, 59, 59, time.UTC)
	//Ticks is the number of 100 nanoseconds from zero value of date
	//this value is copied from mdsd code.
	maxValueDateTimeInTicks := uint64(3155378975999999999)

	//zero time is taken to be 1 Jan,1970 instead of 1 Jan, 1 to avoid integer overflow.
	//The Sub() returns int64 and hence it can hold ony nanoseconds corresponding to 290yrs.
	zeroTime := time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC)
	diff := uint64(currentTime.Sub(zeroTime))
	currentTimeInTicks := diff / 100
	UTCTicks_DescendincurrentTimeOrder := maxValueDateTimeInTicks - currentTimeInTicks
	fmt.Println(UTCTicks_DescendincurrentTimeOrder)

	return UTCTicks_DescendincurrentTimeOrder
}

//RETURNS: counter name being unicode encoded and decreasing time diff.
func getRowKeyComponents(lastSampleTimestamp string, counterName string) (string, string) {

	UTCTicks_DescendingOrder := getUTCTicks_DescendingOrder(lastSampleTimestamp)
	UTCTicks_DescendingOrderStr := strconv.FormatInt(int64(UTCTicks_DescendingOrder), 10)
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
		UTCTicks_DescendingOrderStr, encodedCounterName := getRowKeyComponents(props[TIMESTAMP].(string), props[CounterName].(string))
		props[DeploymentId] = azureTableStorage.DeploymentId
		props[Host], _ = os.Hostname()

		//period is the period which decides when to transfer the aggregated metrics.Its in format "60s"
		periodStr := metrics[i].Tags()[Period]
		table := azureTableStorage.PeriodVsTableNameVsTableRef[periodStr].TableRef

		//two rows are written for each metric as Azure table has optimized prefix search only and no index.
		rowKey1 := UTCTicks_DescendingOrderStr + "_" + encodedCounterName
		entity = table.GetEntityReference(partitionKey, rowKey1)
		entity.Properties = props
		entity.Insert(FullMetadata, nil)

		rowKey2 := encodedCounterName + "_" + UTCTicks_DescendingOrderStr
		entity = table.GetEntityReference(partitionKey, rowKey2)
		entity.Properties = props
		entity.Insert(FullMetadata, nil)

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
