package azuretablestorage

import (
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf16"

	storage "github.com/Azure/azure-sdk-for-go/storage"
	"github.com/influxdata/telegraf"
	constants "github.com/influxdata/telegraf/plugins"
	"github.com/influxdata/telegraf/plugins/outputs"
	overflow "github.com/johncgriffin/overflow"
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
	AccountName                 string //azure storage account name
	SasToken                    string
	ResourceId                  string //resource id for the VM or VMSS
	DeploymentId                string
	Periods                     []string                       //this is the list of periods being configured for various aggregator instances.
	PeriodVsTableNameVsTableRef map[string]TableNameVsTableRef //Map of transfer period of metrics vs table name and table client ref.
	PrevTableNameSuffix         string
	TableStorageEndPointSuffix  string
}

var sampleConfig = `
[[outputs.azuretablestorage]]
deployment_id = ""
resource_id = ""
account_name = ""
sas_token=""
#periods is the list of period configured for each aggregator plugin
periods = ["30s","60s"] 
table_storage_end_point_suffix = ".table.core.windows.net"

`

type MdsdTime struct {
	seconds      int64
	microSeconds int64
}

func (azureTableStorage *AzureTableStorage) toFileTime(mdsdTime MdsdTime) (int64, error) {
	//check for int64 overflow
	fileTimeSeconds, ok := overflow.Add64(constants.EPOCH_DIFFERENCE, mdsdTime.seconds)
	if ok == false {
		erMsg := "integer64 overflow while computing fileTime"
		log.Print(erMsg)
		err := errors.New(erMsg)
		return int64(0), err
	}

	fileTimeTickPerSecond, ok := overflow.Mul64(fileTimeSeconds, constants.TICKS_PER_SECOND)
	if ok == false {
		erMsg := "integer64 overflow while computing fileTime"
		log.Print(erMsg)
		err := errors.New(erMsg)
		return int64(0), err
	}
	fileTime, ok := overflow.Add64(fileTimeTickPerSecond, mdsdTime.microSeconds*10)
	if ok == false {
		erMsg := "integer64 overflow while computing fileTime"
		log.Print(erMsg)
		err := errors.New(erMsg)
		return int64(0), err
	}

	return fileTime, nil
}

func (azureTableStorage *AzureTableStorage) toMdsdTime(fileTime int64) MdsdTime {
	mdsdTime := MdsdTime{0, 0}
	mdsdTime.microSeconds = (fileTime % constants.TICKS_PER_SECOND) / 10
	mdsdTime.seconds = (fileTime / constants.TICKS_PER_SECOND) - constants.EPOCH_DIFFERENCE
	return mdsdTime
}

//New tables are required to be created every 10th Day. And date suffix changes in the new table name.
//to maintain backward compatibilty of azure table schema with LAD 3.0 conversion of mdsd time to filetime and back is required.
//secondsElapsedTillNow: number of seconds elapsed till now from 1 Jan 1970.
//RETURNS: the date of the last day of the last 10 day interval.
func (azureTableStorage *AzureTableStorage) getTableDateSuffix(secondsElapsedTillNow int64) (string, error) {

	if secondsElapsedTillNow < 0 {
		erMsg := "Invalid time passed to getTableDateSuffix(): "
		log.Print(erMsg)
		err := errors.New(erMsg)
		return "", err
	}

	mdsdTime := MdsdTime{seconds: secondsElapsedTillNow, microSeconds: 0}

	//fileTime gives the number of 100ns elapsed till mdsdTime since 1601-01-01
	fileTime, er := azureTableStorage.toFileTime(mdsdTime)
	if er != nil {
		log.Print("Error occurred while converting mdsdtime to filetime mdsdTime = " + strconv.FormatInt(mdsdTime.seconds, 10) +
			":" + strconv.FormatInt(mdsdTime.microSeconds, 10))
		return "", er
	}
	//The “ten day” rollover is the time at which FILETIME mod (number of seconds in 10 days) is zero
	fileTime = fileTime - (fileTime % int64(10*24*60*60*constants.TICKS_PER_SECOND))

	//convert fileTime back to mdsd time.
	mdsdTime = azureTableStorage.toMdsdTime(fileTime)

	//get the date corresponding to the mdsdTime.
	suffixDate := time.Unix(int64(mdsdTime.seconds), 0)
	suffixDateStr := suffixDate.Format(constants.DATE_SUFFIX_FORMAT)
	suffixDateStr = strings.Replace(suffixDateStr, "/", "", -1)

	// the name of the table will have this date as its suffix.
	return suffixDateStr, nil
}

//period is in the format "60s"
//RETURNS: period in the format "PT1M"
func (azureTableStorage *AzureTableStorage) getPeriodStr(period string) (string, error) {

	var periodStr string

	totalSeconds, err := strconv.Atoi(strings.Trim(period, "s"))

	if err != nil {
		log.Println("Period is not in the format of '60s'")
		log.Print(err)
		return "", err
	}

	hour := (int)(math.Floor(float64(totalSeconds) / 3600))
	min := int(math.Floor(float64(totalSeconds-(hour*3600)) / 60))
	sec := totalSeconds - (hour * 3600) - (min * 60)
	periodStr = constants.PT
	if hour > 0 {
		periodStr += strconv.Itoa(hour) + constants.H
	}
	if min > 0 {
		periodStr += strconv.Itoa(min) + constants.M
	}
	if sec > 0 {
		periodStr += strconv.Itoa(sec) + constants.S
	}
	return periodStr, nil
}

//RETURNS:a map of <Period,{TableName, TableClientReference}>
func (azureTableStorage *AzureTableStorage) getAzurePeriodVsTableNameVsTableRefMap(secondsElapsedTillNow int64,
	tableClient storage.TableServiceClient) (map[string]TableNameVsTableRef, error) {

	periodVsTableNameVsTableRef := map[string]TableNameVsTableRef{}

	tableNameSuffix, err := azureTableStorage.getTableDateSuffix(secondsElapsedTillNow)
	if err != nil {
		log.Println("Error while constructing suffix for azure table name")
		return periodVsTableNameVsTableRef, err
	}
	//Empty the map of tables every 10th day as they become obsolete now.
	if azureTableStorage.PrevTableNameSuffix != tableNameSuffix {
		azureTableStorage.PeriodVsTableNameVsTableRef = map[string]TableNameVsTableRef{}
		azureTableStorage.PrevTableNameSuffix = tableNameSuffix
	}

	for _, period := range azureTableStorage.Periods {
		periodStr, err := azureTableStorage.getPeriodStr(period)
		if err != nil {
			log.Println("Error while parsing acheduled transfer period for metrics to the table: " + period)
			return periodVsTableNameVsTableRef, err
		}
		tableName := constants.WAD_METRICS + periodStr + constants.P10DV25 + tableNameSuffix
		table := tableClient.GetTableReference(tableName)
		tableNameVsTableRefObj := TableNameVsTableRef{TableName: tableName, TableRef: table}
		periodVsTableNameVsTableRef[period] = tableNameVsTableRefObj
	}
	return periodVsTableNameVsTableRef, nil
}

func (azureTableStorage *AzureTableStorage) Connect() error {

	sasUrl := "https://" + azureTableStorage.AccountName + azureTableStorage.TableStorageEndPointSuffix
	client, er := storage.NewAccountSASClientFromEndpointToken(sasUrl, azureTableStorage.SasToken)

	if er != nil {
		log.Println("Error in getting table storage client ")
		return er
	}
	//secondsElapsedTillNow: number of seconds elapsed till now from 1 Jan 1970.
	secondsElapsedTillNow := time.Now().Unix()

	tableClient := client.GetTableService()
	azureTableStorage.PeriodVsTableNameVsTableRef, er =
		azureTableStorage.getAzurePeriodVsTableNameVsTableRefMap(secondsElapsedTillNow, tableClient)

	if er != nil {
		log.Println("Error while constructing map of <period , <tableName, tableClient>>")
		return er
	}

	for _, tableVsTableRef := range azureTableStorage.PeriodVsTableNameVsTableRef {
		er := tableVsTableRef.TableRef.Create(30, FullMetadata, nil)
		if er != nil && strings.Contains(er.Error(), "TableAlreadyExists") {
			log.Println("the table ", tableVsTableRef.TableName, " already exists.")
		} else if er != nil {
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

//RETURNS: resourceId encoded by converting any letter other than alphanumerics to unicode as per UTF-16
func (azureTableStorage *AzureTableStorage) encodeSpecialCharacterToUTF16(resourceId string) string {
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
func (azureTableStorage *AzureTableStorage) getPrimaryKey(resourceId string) string {
	return azureTableStorage.encodeSpecialCharacterToUTF16(resourceId)
}

//RETURNS: difference of max value that can be held by time and number of 100 ns in current time.
func (azureTableStorage *AzureTableStorage) getUTCTicks_DescendingOrder(lastSampleTimestamp string) (uint64, error) {

	currentTime, err := time.Parse(layout, lastSampleTimestamp)
	if err != nil {
		log.Println("Error while parsing timestamp " + lastSampleTimestamp + "in the layout " + layout)
		log.Print(err)
		return 0, err
	}
	//maxValureDateTime := time.Date(9999, time.December, 31, 12, 59, 59, 59, time.UTC)
	//Ticks is the number of 100 nanoseconds from zero value of date
	//this value is copied from mdsd code.
	maxValueDateTimeInTicks := uint64(3155378975999999999)

	//zero time is taken to be 1 Jan,1970 instead of 1 Jan, 1 to avoid integer overflow.
	//The Sub() returns int64 and hence it can hold ony nanoseconds corresponding to 290yrs.
	zeroTime := time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC)
	diff := uint64(currentTime.Sub(zeroTime))
	currentTimeInTicks := diff / 100
	UTCTicks_DescendincurrentTimeOrder := maxValueDateTimeInTicks - currentTimeInTicks

	return UTCTicks_DescendincurrentTimeOrder, nil
}

//RETURNS: counter name being unicode encoded and decreasing time diff.
func (azureTableStorage *AzureTableStorage) getRowKeyComponents(lastSampleTimestamp string, counterName string) (string, string, error) {

	UTCTicks_DescendingOrder, err := azureTableStorage.getUTCTicks_DescendingOrder(lastSampleTimestamp)
	if err != nil {
		log.Println("Error while computing UTCTicks_DescendingOrder")
		return "", "", err
	}
	UTCTicks_DescendingOrderStr := strconv.FormatInt(int64(UTCTicks_DescendingOrder), 10)
	encodedCounterName := azureTableStorage.encodeSpecialCharacterToUTF16(counterName)
	return UTCTicks_DescendingOrderStr, encodedCounterName, nil
}

func (azureTableStorage *AzureTableStorage) Write(metrics []telegraf.Metric) error {
	var entity *storage.Entity
	var props map[string]interface{}
	partitionKey := azureTableStorage.getPrimaryKey(azureTableStorage.ResourceId)

	// iterate over the list of metrics and create a new entity for each metrics and add to the table.
	for i, _ := range metrics {
		props = metrics[i].Fields()
		props[constants.DEPLOYMENT_ID] = azureTableStorage.DeploymentId
		var err error
		props[constants.HOST], err = os.Hostname()
		if err != nil {
			log.Println("Error while getting hostname from os")
			log.Print(err)
			return err
		}
		//period decides when to transfer the aggregated metrics.Its in format "60s"
		tags := metrics[i].Tags()
		props[constants.COUNTER_NAME] = tags[constants.INPUT_PLUGIN] + "/" + props[constants.COUNTER_NAME].(string)

		UTCTicks_DescendingOrderStr, encodedCounterName, er := azureTableStorage.getRowKeyComponents(props[constants.END_TIMESTAMP].(string), props[constants.COUNTER_NAME].(string))
		if er != nil {
			return er
		}

		periodStr := tags[constants.PERIOD]
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
