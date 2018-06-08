package utility

import (
	"strconv"
	"testing"

	testutil "github.com/influxdata/telegraf/testutil"
	assert "github.com/stretchr/testify/assert"
)

//Test conversion of mdsdtime to filetime
func TestToFileTime(t *testing.T) {

	numOfSeconds := testutil.GetNumOfSeconds()
	numOfMicroSeconds := testutil.GetMicroSeconds()

	mdsdTime := MdsdTime{numOfSeconds, numOfMicroSeconds}

	actualFileTime, er := ToFileTime(mdsdTime)
	expectedFileTime := testutil.GetFileTime()
	assert.Nil(t, er)
	assert.Equal(t, expectedFileTime, actualFileTime)

	//test integer overflow
	mdsdTime.Seconds = int64(9223372036854775807)
	actualFileTime, er = ToFileTime(mdsdTime)
	assert.NotNil(t, er)
}

func TestToMdsdTime(t *testing.T) {

	fileTime := testutil.GetFileTime()
	actualMdsdTime := ToMdsdTime(fileTime)

	numOfSeconds := testutil.GetNumOfSeconds()
	numOfMicroSeconds := testutil.GetMicroSeconds()
	expectedMdsdTime := MdsdTime{Seconds: numOfSeconds, MicroSeconds: numOfMicroSeconds}

	assert.Equal(t, expectedMdsdTime, actualMdsdTime)
}
func TestGetIntervalISO8601(t *testing.T) {
	validPeriod := "3672s"
	actualPeriodStr, er := GetIntervalISO8601(validPeriod)
	expectedPeriodStr := "PT1H1M12S"
	assert.Nil(t, er)
	assert.Equal(t, expectedPeriodStr, actualPeriodStr)

	//invalid period
	invalidPeriod := "abcs"
	actualPeriodStr, er = GetIntervalISO8601(invalidPeriod)
	assert.NotNil(t, er)
}

func TestEncodeSpecialCharacterToUTF16(t *testing.T) {
	resourceId := "dummy/subscriptionId/resourceGroup/VMScaleset"
	actualEncodedResourceId := EncodeSpecialCharacterToUTF16(resourceId)
	expectedEncodedResourceId := "dummy:002FsubscriptionId:002FresourceGroup:002FVMScaleset"
	assert.Equal(t, expectedEncodedResourceId, actualEncodedResourceId)
}

func TestGetUTCTicks_DescendingOrder(t *testing.T) {
	timestamp := "19/04/2018 12:20:00 PM"
	actualTicks, er := GetUTCTicks_DescendingOrder(timestamp)
	expectedTicks := uint64(3140137571999999999)
	assert.Nil(t, er)
	assert.Equal(t, expectedTicks, actualTicks)

	//invalid timestamp
	timestamp = "3140137571999999999"
	actualTicks, er = GetUTCTicks_DescendingOrder(timestamp)
	assert.NotNil(t, er)
}

func TestGetPropsStr(t *testing.T) {
	props := make(map[string]interface{})
	props[MEAN] = float64(0)
	props[COUNTER_NAME] = "name"
	expectedPropStr := MEAN + " : " + strconv.FormatFloat(props[MEAN].(float64), 'E', -1, 64) + " , " + COUNTER_NAME + " : " + props[COUNTER_NAME].(string) + " , "
	actualPropsStr := GetPropsStr(props)
	assert.Equal(t, expectedPropStr, actualPropsStr)
}
