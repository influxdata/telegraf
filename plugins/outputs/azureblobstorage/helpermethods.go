package azureblobstorage

import (
	"log"
	"math"
	"strconv"
	"strings"
	"time"
)

func getEventVersionStr(eventVersion string) (string, error) {
	eventVerInt, er := strconv.Atoi(eventVersion)
	if er != nil {
		log.Println("error while parsing event version." + eventVersion + " Event version should be in the format eg 2")
		return "", er
	}
	eventVerStr := "ver" + strconv.Itoa(eventVerInt) + "v0"
	return eventVerStr, nil
}
func getBaseTimeStr(baseTime string) (string, error) {
	baseTimeInt, er := strconv.Atoi(baseTime) //TODO: use ParseInt
	if er != nil {
		log.Println("Error while Parsing baseTime" + baseTime)
		return "", er
	}
	date := time.Unix(int64(baseTimeInt), 0)
	baseTimeStr := "y=" + strconv.Itoa(date.Year()) + "/m=" + strconv.Itoa(int(date.Month())) + "/d=" + strconv.Itoa(date.Day()) + "/h=" + strconv.Itoa(date.Hour()) + "/m=" + strconv.Itoa(date.Minute())
	return baseTimeStr, nil
}
func getIntervalStr(interval string) (string, error) {
	intervalStr, er := GetPeriodStr(interval)
	if er != nil {
		log.Println("Error while Parsing interval" + interval)
		return "", er
	}
	return intervalStr, nil
}
func getBlobPath(resourceId string, identityHash string, baseTime string, interval string) (string, error) {
	intervalStr, er := getIntervalStr(interval)
	if er != nil {
		log.Println("Error while Parsing interval" + interval)
		return "", er
	}

	baseTimeStr, er := getBaseTimeStr(baseTime)
	if er != nil {
		log.Println("Error while Parsing baseTime" + baseTime)
		return "", er
	}
	blobPath := "resourceId=" + resourceId + "/i=" + identityHash + "/" + baseTimeStr + "/" + intervalStr + ".json"
	return blobPath, nil
}

func GetPeriodStr(period string) (string, error) {

	var periodStr string

	totalSeconds, err := strconv.Atoi(strings.Trim(period, "s"))

	if err != nil {
		log.Println(" Error while parsing period." + period)
		log.Print(err.Error())
		return "", err
	}

	hour := (int)(math.Floor(float64(totalSeconds) / 3600))
	min := int(math.Floor(float64(totalSeconds-(hour*3600)) / 60))
	sec := totalSeconds - (hour * 3600) - (min * 60)
	periodStr = "PT"
	if hour > 0 {
		periodStr += strconv.Itoa(hour) + "H"
	}
	if min > 0 {
		periodStr += strconv.Itoa(min) + "M"
	}
	if sec > 0 {
		periodStr += strconv.Itoa(sec) + "S"
	}
	return periodStr, nil
}
