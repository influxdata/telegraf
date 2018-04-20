package azuretablestorage

import (
	"testing"
	"time"

	storage "github.com/Azure/azure-sdk-for-go/storage"
	assert "github.com/stretchr/testify/assert"
)

func getAzureTableStorageObj() AzureTableStorage {
	azureTableStorageObj := AzureTableStorage{
		AccountName:                "dummy_account_name",
		SasToken:                   "dummy_sas_token",
		ResourceId:                 "dummy/subscriptionId/resourceGroup/VMScaleset",
		DeploymentId:               "dummy_deployment_id",
		TableStorageEndPointSuffix: ".dummy.end.point.suffix",
		Periods:                    []string{"30s", "60s"},
	}
	return azureTableStorageObj
}
func getNumOfSeconds() int64 {
	return int64(1524115802)
}
func getMicroSeconds() int64 {
	return int64(0)
}
func getFileTime() int64 {
	return int64(131685894020000000)
}

//Test conversion of mdsdtime to filetime
func TestToFileTime(t *testing.T) {
	azureTableStorageObj := getAzureTableStorageObj()

	numOfSeconds := getNumOfSeconds()
	numOfMicroSeconds := getMicroSeconds()

	mdsdTime := MdsdTime{numOfSeconds, numOfMicroSeconds}

	actualFileTime, er := azureTableStorageObj.toFileTime(mdsdTime)
	expectedFileTime := getFileTime()
	assert.Nil(t, er)
	assert.Equal(t, expectedFileTime, actualFileTime)

	//test integer overflow
	mdsdTime.seconds = int64(9223372036854775807)
	actualFileTime, er = azureTableStorageObj.toFileTime(mdsdTime)
	assert.NotNil(t, er)
}

func TestToMdsdTime(t *testing.T) {
	azureTableStorageObj := getAzureTableStorageObj()

	fileTime := getFileTime()
	actualMdsdTime := azureTableStorageObj.toMdsdTime(fileTime)

	numOfSeconds := getNumOfSeconds()
	numOfMicroSeconds := getMicroSeconds()
	expectedMdsdTime := MdsdTime{seconds: numOfSeconds, microSeconds: numOfMicroSeconds}

	assert.Equal(t, expectedMdsdTime, actualMdsdTime)
}
func TestGetTableSuffix(t *testing.T) {
	azureTableStorageObj := getAzureTableStorageObj()
	numOfSeconds := getNumOfSeconds()
	actualSuffix, er := azureTableStorageObj.getTableDateSuffix(numOfSeconds)
	expectedSuffix := "20180415"
	assert.Equal(t, expectedSuffix, actualSuffix)
	assert.Nil(t, er)

	//invalid number of seconds
	invalidNumOfSeconds := int64(-123)
	actualSuffix, er = azureTableStorageObj.getTableDateSuffix(invalidNumOfSeconds)
	assert.NotNil(t, er)

	//tableSuffix should change by 10 days
	expectedSuffix = "20180425"
	time := time.Unix(numOfSeconds, 0)
	timeAfter12Days := time.AddDate(0, 0, 12)
	numOfSecondsAfter12Days := timeAfter12Days.Unix()
	actualSuffix, er = azureTableStorageObj.getTableDateSuffix(numOfSecondsAfter12Days)

	assert.Nil(t, er)
	assert.Equal(t, expectedSuffix, actualSuffix)

}
func TestGetPeriodStr(t *testing.T) {
	azureTableStorageObj := getAzureTableStorageObj()
	validPeriod := "3672s"
	actualPeriodStr, er := azureTableStorageObj.getPeriodStr(validPeriod)
	expectedPeriodStr := "PT1H1M12S"
	assert.Nil(t, er)
	assert.Equal(t, expectedPeriodStr, actualPeriodStr)

	//invalid period
	invalidPeriod := "abcs"
	actualPeriodStr, er = azureTableStorageObj.getPeriodStr(invalidPeriod)
	assert.NotNil(t, er)
}

func TestGetAzurePeriodVsTableNameVsTableRefMap(t *testing.T) {
	azureTableStorageObj := getAzureTableStorageObj()
	tableServiceClient := storage.TableServiceClient{}
	numberOfSeconds := getNumOfSeconds()

	actualMap, er := azureTableStorageObj.getAzurePeriodVsTableNameVsTableRefMap(numberOfSeconds, tableServiceClient)
	assert.Nil(t, er)

	expectedMap := map[string]TableNameVsTableRef{}

	period1 := azureTableStorageObj.Periods[0]
	tableName1 := "WADMetricsPT30SP10DV2520180415"
	table1 := tableServiceClient.GetTableReference(tableName1)
	tableRef1 := TableNameVsTableRef{tableName1, table1}

	period2 := azureTableStorageObj.Periods[1]
	tableName2 := "WADMetricsPT1MP10DV2520180415"
	table2 := tableServiceClient.GetTableReference(tableName2)
	tableRef2 := TableNameVsTableRef{tableName2, table2}
	expectedMap[period1] = tableRef1
	expectedMap[period2] = tableRef2
	assert.Equal(t, expectedMap, actualMap)
}

func TestEncodeSpecialCharacterToUTF16(t *testing.T) {
	azureTableStorageObj := getAzureTableStorageObj()
	resourceId := azureTableStorageObj.ResourceId
	actualEncodedResourceId := azureTableStorageObj.encodeSpecialCharacterToUTF16(resourceId)
	expectedEncodedResourceId := "dummy:002FsubscriptionId:002FresourceGroup:002FVMScaleset"
	assert.Equal(t, expectedEncodedResourceId, actualEncodedResourceId)
}

func TestGetUTCTicks_DescendingOrder(t *testing.T) {
	azureTableStorageObj := getAzureTableStorageObj()
	timestamp := "19/04/2018 12:20:00 PM"
	actualTicks, er := azureTableStorageObj.getUTCTicks_DescendingOrder(timestamp)
	expectedTicks := uint64(3140137571999999999)
	assert.Nil(t, er)
	assert.Equal(t, expectedTicks, actualTicks)

	//invalid timestamp
	timestamp = "3140137571999999999"
	actualTicks, er = azureTableStorageObj.getUTCTicks_DescendingOrder(timestamp)
	assert.NotNil(t, er)
}
