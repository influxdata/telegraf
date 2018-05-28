package azuretablestorage

import (
	"testing"
	"time"

	storage "github.com/Azure/azure-sdk-for-go/storage"
	testutil "github.com/influxdata/telegraf/testutil"
	util "github.com/influxdata/telegraf/utility"
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

func TestGetTableSuffix(t *testing.T) {
	numOfSeconds := testutil.GetNumOfSeconds()
	actualSuffix, er := getTableDateSuffix(numOfSeconds)
	expectedSuffix := "20180415"
	assert.Equal(t, expectedSuffix, actualSuffix)
	assert.Nil(t, er)

	//invalid number of seconds
	invalidNumOfSeconds := int64(-123)
	actualSuffix, er = getTableDateSuffix(invalidNumOfSeconds)
	assert.NotNil(t, er)

	//tableSuffix should change by 10 days
	expectedSuffix = "20180425"
	time := time.Unix(numOfSeconds, 0)
	timeAfter12Days := time.AddDate(0, 0, 12)
	numOfSecondsAfter12Days := timeAfter12Days.Unix()
	actualSuffix, er = getTableDateSuffix(numOfSecondsAfter12Days)

	assert.Nil(t, er)
	assert.Equal(t, expectedSuffix, actualSuffix)

}

func TestGetAzurePeriodVsTableNameVsTableRefMap(t *testing.T) {
	azureTableStorageObj := getAzureTableStorageObj()
	tableServiceClient := storage.TableServiceClient{}
	numberOfSeconds := testutil.GetNumOfSeconds()

	actualMap, er := azureTableStorageObj.getAzureperiodVsTableNameVsTableRefMap(numberOfSeconds, tableServiceClient)
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

func TestValidateRow(t *testing.T) {
	columnsInTable := []string{}
	columnsInTable = append(columnsInTable, util.MEAN)
	columnsInTable = append(columnsInTable, util.SAMPLE_COUNT)
	columnsInTable = append(columnsInTable, util.COUNTER_NAME)
	columnsInTable = append(columnsInTable, util.DEPLOYMENT_ID)
	columnsInTable = append(columnsInTable, util.HOST)
	columnsInTable = append(columnsInTable, util.LAST_SAMPLE)
	columnsInTable = append(columnsInTable, util.MAX_SAMPLE)
	columnsInTable = append(columnsInTable, util.MIN_SAMPLE)
	columnsInTable = append(columnsInTable, util.BEGIN_TIMESTAMP)
	columnsInTable = append(columnsInTable, util.END_TIMESTAMP)
	columnsInTable = append(columnsInTable, util.TOTAL)

	//invalid: len of prop != len of valid columns, props has invalid column
	props := make(map[string]interface{})
	props["InValidColumn"] = "value"
	isValid := validateRow(columnsInTable, props)
	assert.False(t, isValid)

	//invalid: len of prop != len of valid columns, props has invalid column
	props1 := make(map[string]interface{})
	for i := range columnsInTable {
		props1[columnsInTable[i]+"Invalid"] = "invalid"
	}
	isValid = validateRow(columnsInTable, props1)
	assert.False(t, isValid)

	//invalid with empty value
	props2 := make(map[string]interface{})
	for i := range columnsInTable {
		props2[columnsInTable[i]] = ""
	}
	props2[columnsInTable[0]] = float64(0) // valid value
	isValid = validateRow(columnsInTable, props2)
	assert.False(t, isValid)

	//valid props
	props3 := make(map[string]interface{})
	for i := range columnsInTable {
		props3[columnsInTable[i]] = "valid"
	}
	props3[columnsInTable[0]] = float64(0) // valid value
	isValid = validateRow(columnsInTable, props3)
	assert.True(t, isValid)
}
