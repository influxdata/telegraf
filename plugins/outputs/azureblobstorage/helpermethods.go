package azureblobstorage

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"strconv"
	"strings"
	"time"

	storage "github.com/Azure/azure-sdk-for-go/storage"
	util "github.com/influxdata/telegraf/utility"
	overflow "github.com/johncgriffin/overflow"
)

//decides if its time to write new blob
func checkIsIntervalOver(baseTime string, interval string) (bool, error) {
	isIntervalOver := false
	baseTimeInt64, er := strconv.ParseInt(baseTime, 10, 64)
	if er != nil {
		log.Println("Error while parsing baseTime " + baseTime + er.Error())
		return isIntervalOver, er
	}

	intervalInt64, err := strconv.ParseInt(strings.Trim(interval, "s"), 10, 64)

	if err != nil {
		log.Println(" Error while parsing period." + interval)
		log.Print(er.Error())
		return isIntervalOver, er
	}

	baseTimeDate := time.Unix(baseTimeInt64, 0)
	currentDate := time.Now()
	timeDiffInt64 := int64(currentDate.Sub(baseTimeDate))

	if intervalInt64 <= timeDiffInt64 {
		isIntervalOver = true
	}

	return isIntervalOver, nil
}

//return number of seconds obtained by adding interval to current base time
func getNewBaseTime(baseTime string, interval string) (string, error) {
	baseTimeInt64, er := strconv.ParseInt(baseTime, 10, 64)
	if er != nil {
		log.Println("Error while parsing baseTime " + baseTime + er.Error())
		return "", er
	}

	intervalInt64, err := strconv.ParseInt(strings.Trim(interval, "s"), 10, 64)
	if err != nil {
		log.Println(" Error while parsing period." + interval)
		log.Print(er.Error())
		return "", er
	}

	newBaseTimeInt64, ok := overflow.Add64(baseTimeInt64, intervalInt64)
	if ok == false {
		msg := "Error Integer overflow occurred while adding baseTime and interval " +
			strconv.FormatInt(baseTimeInt64, 10) + " " +
			strconv.FormatInt(intervalInt64, 10) + " "
		log.Println(msg)
		er = errors.New(msg)
		return "", er
	}
	return strconv.FormatInt(newBaseTimeInt64, 10), nil
}
func validateBlobRef(blobRef *storage.Blob) error {
	if blobRef.Name == "" {
		erMsg := "Error: invalid Blob Reference, blobReference.Name is empty "
		log.Println(erMsg)
		er := errors.New(erMsg)
		return er
	}
	return nil
}

//if container name is not set in the reference the container reference is invalid
func validateContainerRef(containerRef *storage.Container) error {
	if containerRef.Name == "" {
		erMsg := "Error while getting container reference for container "
		log.Println(erMsg)
		er := errors.New(erMsg)
		return er
	}
	return nil
}

//if error occurs while getting properties of blob client that means the client is not constructed correctly
func validateBlobClient(blobClient storage.BlobStorageClient) error {
	_, er := blobClient.GetServiceProperties()
	if er != nil {
		log.Println("Error while validating blob client by calling GetServiceProperties()" + er.Error())
		return er
	}
	return nil
}

func getJsonBlock(jsonObj BlockObject) (string, error) {
	jsonBlock, er := json.Marshal(&jsonObj)
	if er != nil {
		return "", er
	}
	return string(jsonBlock), nil
}

//if event version = 2 then result is ver2v0
func getEventVersionStr(eventVersion string) (string, error) {
	eventVerInt, er := strconv.Atoi(eventVersion)
	if er != nil {
		log.Println("error while parsing event version." + eventVersion + " Event version should be in the format eg 2")
		return "", er
	}
	eventVerStr := "ver" + strconv.Itoa(eventVerInt) + "v0"
	return eventVerStr, nil
}

//takes in number of seconds and converts them into y=2015/m=05/d=03/h=00/m=00 format
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

func getBaseTimeMultipleOfInterval(baseTime string, interval string) (string, error) {

	baseTimeInt64, er := strconv.ParseInt(baseTime, 10, 64)
	if er != nil {
		log.Println("Error while parsing baseTime " + baseTime + er.Error())
		return "", er
	}

	intervalInt64, err := strconv.ParseInt(strings.Trim(interval, "s"), 10, 64)
	if err != nil {
		log.Println(" Error while parsing period." + interval)
		log.Print(er.Error())
		return "", er
	}

	//get baseTime in multiples of interval
	baseTimeInt64 = baseTimeInt64 - (baseTimeInt64 % intervalInt64)
	return strconv.FormatInt(baseTimeInt64, 10), nil
}

//blob path is a combination of resource id, identity hash and the base time
//eg: Blob path: resourceId=<test_resource_id>/i=<agentIdentityHash>/y=2015/m=05/d=03/h=00/m=00/name=PT1H.json
func getBlobPath(resourceId string, identityHash string, baseTime string, intervalISO8601 string) (string, error) {
	baseTimeStr, er := getBaseTimeStr(baseTime)
	if er != nil {
		log.Println("Error while Parsing baseTime" + baseTime)
		return "", er
	}
	blobPath := "resourceId=" + resourceId + "/i=" + identityHash + "/" + baseTimeStr + "/" + intervalISO8601 + ".json"
	return blobPath, nil
}

// writes blocks to blob storage
// https://docs.microsoft.com/en-us/rest/api/storageservices/understanding-block-blobs--append-blobs--and-page-blobs#about-block-blobs
func writeJsonBlockToBlob(requiredFields map[string][]string, requiredFieldCount int, jsonBlock string, blockBlobRef *storage.Blob) (string, error) {

	//block Id the length of string pre encoding should be less than 64
	blockId := base64.StdEncoding.EncodeToString([]byte(jsonBlock[1:60]))
	md5HashOfBlock, er := util.Getmd5Hash(jsonBlock)
	if er != nil {
		log.Println("Error while calculating md5 hash of content " + jsonBlock)
		return "", nil
	}

	options := storage.PutBlockOptions{
		Timeout:    30,             // in seconds
		ContentMD5: md5HashOfBlock, // to ensure integrity of the content of block being sent to the storage
		RequestID:  md5HashOfBlock,
	}
	log.Println("Request Id " + md5HashOfBlock)
	log.Println("Writing block to storage " + blockId)
	er = blockBlobRef.PutBlock(blockId, []byte(jsonBlock), &options)
	if er != nil {
		log.Println("Error while writing block to blob storage blockId,content" + blockId + jsonBlock + er.Error())
		return "", er
	}
	blockList := []storage.Block{{blockId, storage.BlockStatusUncommitted}}
	requestId, err := util.Getmd5Hash(base64.StdEncoding.EncodeToString([]byte(blockId)))
	if err != nil {
		log.Println("Error while computing md5 hash of blockid " + blockId)
		return "", err
	}
	log.Println("Committing block (RequestId, BlockID) " +
		requestId + " " +
		blockId)
	putBlockListOptions := storage.PutBlockListOptions{
		RequestID: requestId,
	}
	er = blockBlobRef.PutBlockList(blockList, &putBlockListOptions)
	if er != nil {
		log.Println("Error while writing block to blob storage blockId,content" + blockId + jsonBlock + er.Error())
		return "", er
	}
	return blockId, nil
}

func getColumnConversionMap() map[string]string {
	columnConversionMap := make(map[string]string)
	columnConversionMap[util.BLOCK_JSON_KEY_COUNTER_NAME] = util.COUNTER_NAME
	columnConversionMap[util.BLOCK_JSON_KEY_END_TIMESTAMP] = util.END_TIMESTAMP
	columnConversionMap[util.BLOCK_JSON_KEY_LAST_SAMPLE] = util.LAST_SAMPLE
	columnConversionMap[util.BLOCK_JSON_KEY_MAX_SAMPLE] = util.LAST_SAMPLE
	columnConversionMap[util.BLOCK_JSON_KEY_MIN_SAMPLE] = util.MIN_SAMPLE
	columnConversionMap[util.BLOCK_JSON_KEY_MEAN] = util.MEAN
	columnConversionMap[util.BLOCK_JSON_KEY_SAMPLE_COUNT] = util.SAMPLE_COUNT
	columnConversionMap[util.BLOCK_JSON_KEY_TOTAL] = util.TOTAL
	return columnConversionMap
}

func getRequiredFieldList() (map[string][]string, int) {
	requiredFieldMap := make(map[string][]string)
	requiredFieldMap[util.BLOCK_JSON_KEY_END_TIMESTAMP] = nil
	requiredFieldMap[util.BLOCK_JSON_KEY_COUNTER_NAME] = nil
	requiredFieldMap[util.BLOCK_JSON_KEY_RESOURCE_ID] = nil
	requiredFieldMap[util.BLOCK_JSON_KEY_TIME_GRAIN] = nil
	requiredFieldMap[util.BLOCK_JSON_KEY_TOTAL] = nil
	requiredFieldMap[util.BLOCK_JSON_KEY_TOTAL] = nil
	requiredFieldMap[util.BLOCK_JSON_KEY_MAX_SAMPLE] = nil
	requiredFieldMap[util.BLOCK_JSON_KEY_MEAN] = nil
	requiredFieldMap[util.BLOCK_JSON_KEY_MIN_SAMPLE] = nil
	requiredFieldMap[util.BLOCK_JSON_KEY_SAMPLE_COUNT] = nil
	requiredFieldMap[util.BLOCK_JSON_KEY_LAST_SAMPLE] = nil
	var dimensions []string
	dimensions = append(dimensions, util.BLOCK_JSON_KEY_ROLE)
	dimensions = append(dimensions, util.BLOCK_JSON_KEY_ROLE_INSTANCE)
	dimensions = append(dimensions, util.BLOCK_JSON_KEY_TENANT)
	requiredFieldMap[util.BLOCK_JSON_KEY_DIMENSIONS] = dimensions

	return requiredFieldMap, 14
}
