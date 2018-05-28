package azuretablestorage

import (
	"errors"
	"log"
	"reflect"
	"strconv"
	"strings"
	"time"

	storage "github.com/Azure/azure-sdk-for-go/storage"
	util "github.com/influxdata/telegraf/utility"
)

//RETURNS: primary key for the azure table.
func getPartitionKey(resourceId string) string {
	return util.EncodeSpecialCharacterToUTF16(resourceId)
}

//RETURNS: counter name being unicode encoded and decreasing time diff.
func getRowKeyComponents(lastSampleTimestamp string, counterName string) (string, string, error) {

	UTCTicks_DescendingOrder, err := util.GetUTCTicks_DescendingOrder(lastSampleTimestamp)
	if err != nil {
		log.Println("Error while computing UTCTicks_DescendingOrder")
		return "", "", err
	}
	UTCTicks_DescendingOrderStr := strconv.FormatInt(int64(UTCTicks_DescendingOrder), 10)
	encodedCounterName := util.EncodeSpecialCharacterToUTF16(counterName)
	return UTCTicks_DescendingOrderStr, encodedCounterName, nil
}
func contains(obj string, array []string) bool {
	present := false
	for _, v := range array {
		if v == obj {
			present = true
			break
		}
	}
	return present
}

func validateRow(columnlist []string, props map[string]interface{}) bool {
	isValidRow := true
	count := 0
	for key, value := range props {
		if !contains(key, columnlist) {
			isValidRow = false
			break
		}
		if reflect.TypeOf(value).String() == "string" && value.(string) == "" {
			isValidRow = false
			break
		}
		count++
	}
	if count != len(columnlist) {
		isValidRow = false
	}
	return isValidRow
}

func validateTableClient(tableClient storage.TableServiceClient) error {
	var er error
	er = nil

	//the tableClient.GetServiceProperties() is called to check if tableClient is created correctly
	_, er = tableClient.GetServiceProperties()
	return er
}

//New tables are required to be created every 10th Day. And date suffix changes in the new table name.
//to maintain backward compatibilty of azure table schema with LAD 3.0 conversion of mdsd time to filetime and back is required.
//secondsElapsedTillNow: number of seconds elapsed till now from 1 Jan 1970.
//RETURNS: the date of the last day of the last 10 day interval.
func getTableDateSuffix(secondsElapsedTillNow int64) (string, error) {

	if secondsElapsedTillNow < 0 {
		erMsg := "Invalid time passed to getTableDateSuffix(): "
		log.Print(erMsg)
		err := errors.New(erMsg)
		return "", err
	}

	mdsdTime := util.MdsdTime{Seconds: secondsElapsedTillNow, MicroSeconds: 0}

	//fileTime gives the number of 100ns elapsed till mdsdTime since 1601-01-01
	fileTime, er := util.ToFileTime(mdsdTime)
	if er != nil {
		log.Print("Error occurred while converting mdsdtime to filetime mdsdTime = " +
			strconv.FormatInt(mdsdTime.Seconds, 10) +
			":" + strconv.FormatInt(mdsdTime.MicroSeconds, 10))
		return "", er
	}
	//The “ten day” rollover is the time at which FILETIME mod (number of seconds in 10 days) is zero
	fileTime = fileTime - (fileTime % int64(10*24*60*60*util.TICKS_PER_SECOND))

	//convert fileTime back to mdsd time.
	mdsdTime = util.ToMdsdTime(fileTime)

	//get the date corresponding to the mdsdTime.
	suffixDate := time.Unix(int64(mdsdTime.Seconds), 0)
	suffixDateStr := suffixDate.Format(util.DATE_SUFFIX_FORMAT)
	suffixDateStr = strings.Replace(suffixDateStr, "/", "", -1)

	// the name of the table will have this date as its suffix.
	return suffixDateStr, nil
}

func writeEntitiesToTable(
	partitionKey string,
	rowKey1 string,
	rowKey2 string,
	props map[string]interface{},
	table *storage.Table) error {

	var er error
	var entity *storage.Entity
	entity = table.GetEntityReference(partitionKey, rowKey1)
	if entity.PartitionKey == "" {
		logMsg := "Error: invalid entity reference for entity : " + util.GetPropsStr(props)
		log.Println(logMsg)
		return errors.New(logMsg)
	}
	entity.Properties = props
	er = entity.Insert(FullMetadata, nil)
	if er != nil {
		log.Println("Error: Failed to write entity to the table " + entity.Table.Name +
			" entity value: " + util.GetPropsStr(entity.Properties))
		return er
	}

	entity = table.GetEntityReference(partitionKey, rowKey2)
	if entity.PartitionKey == "" {
		logMsg := "Error: invalid entity reference for entity : " + util.GetPropsStr(props)
		log.Println(logMsg)
		return errors.New(logMsg)
	}
	entity.Properties = props
	er = entity.Insert(FullMetadata, nil)
	if er != nil {
		log.Println("Failed to write entity to the table " + entity.Table.Name +
			" entity value: " + util.GetPropsStr(entity.Properties))
		return er
	}
	return er
}
