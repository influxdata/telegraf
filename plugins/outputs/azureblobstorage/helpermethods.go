package azureblobstorage

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"strconv"
	"time"

	storage "github.com/Azure/azure-sdk-for-go/storage"
	util "github.com/influxdata/telegraf/utility"
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
func getIntervalISO8601(interval string) (string, error) {
	intervalStr, er := util.GetPeriodStr(interval)
	if er != nil {
		log.Println("Error while Parsing interval" + interval)
		return "", er
	}
	return intervalStr, nil
}
func getBlobPath(resourceId string, identityHash string, baseTime string, intervalISO8601 string) (string, error) {
	baseTimeStr, er := getBaseTimeStr(baseTime)
	if er != nil {
		log.Println("Error while Parsing baseTime" + baseTime)
		return "", er
	}
	blobPath := "resourceId=" + resourceId + "/i=" + identityHash + "/" + baseTimeStr + "/" + intervalISO8601 + ".json"
	return blobPath, nil
}

/*
Sample

    { "time" : "2018-04-23T01:00:06.4265660Z",
  "resourceId" : "/subscriptions/20ff167c-9f4b-4a73-9fd6-0dbe93fa778a/resourceGroups/nirasto_lad/providers/Microsoft.Compute/virtualMachines/rhel72metric",
  "timeGrain" : "PT1H",
  "dimensions": {
     "Tenant": "",
     "Role": "",
     "RoleInstance": ""
  },
  "metricName": "/builtin/network/totaltxerrors",
  "total": 0,
  "minimum": 0,
  "maximum": 0,
  "average": 0,
  "count": 4,
  "last": 0
}
*/
func validateJsonBlock(jsonBlock string) bool {
	return true
}
func writeJsonBlockToBlob(jsonBlock string, blockBlobRef *storage.Blob) (string, error) {
	isJsonBlockValid := validateJsonBlock(jsonBlock)
	if isJsonBlockValid == false {
		log.Println("Invalid block skipped to write it to blob storage " + jsonBlock)
		return "", nil
	}
	blockId := base64.StdEncoding.EncodeToString([]byte(jsonBlock[1:60]))

	log.Println("Writing block to storage" + blockId)
	md5HashOfBlock, er := util.Getmd5Hash(jsonBlock)
	if er != nil {
		log.Println("Error while calculating md5 hash of content")
		return "", nil
	}
	options := storage.PutBlockOptions{
		Timeout:    30,             // in seconds
		ContentMD5: md5HashOfBlock, // to ensure integrity of the content of block being sent to the storage
		RequestID:  md5HashOfBlock, // TODO: generate RequestID
	}

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

type Dimensions struct {
	Tenant       string
	Role         string
	RoleInstance string
}
type BlockFields struct {
	time       string
	resourceId string
	timegrain  string
	dimensions Dimensions
	metricName string
	total      float64
	minimum    float64
	maximum    float64
	average    float64
	count      float64
	last       float64
}

//TODO: rewrite after writing validate method.
func getJsonBlock(props map[string]interface{}, azureBlobStorage *AzureBlobStorage) (string, error) {
	jsonBlock := ""
	propToBlockColumnsMap := getColumnConversionMap()
	blockProps := make(map[string]interface{})
	blockProps[util.BLOCK_JSON_KEY_COUNTER_NAME] = props[propToBlockColumnsMap[util.BLOCK_JSON_KEY_COUNTER_NAME]]
	blockProps[util.BLOCK_JSON_KEY_END_TIMESTAMP] = props[propToBlockColumnsMap[util.BLOCK_JSON_KEY_END_TIMESTAMP]]
	blockProps[util.BLOCK_JSON_KEY_LAST_SAMPLE] = props[propToBlockColumnsMap[util.BLOCK_JSON_KEY_LAST_SAMPLE]]
	blockProps[util.BLOCK_JSON_KEY_MAX_SAMPLE] = props[propToBlockColumnsMap[util.BLOCK_JSON_KEY_MAX_SAMPLE]]
	blockProps[util.BLOCK_JSON_KEY_MEAN] = props[propToBlockColumnsMap[util.BLOCK_JSON_KEY_MEAN]]
	blockProps[util.BLOCK_JSON_KEY_MIN_SAMPLE] = props[propToBlockColumnsMap[util.BLOCK_JSON_KEY_MIN_SAMPLE]]
	blockProps[util.BLOCK_JSON_KEY_SAMPLE_COUNT] = props[propToBlockColumnsMap[util.BLOCK_JSON_KEY_SAMPLE_COUNT]]
	blockProps[util.BLOCK_JSON_KEY_TOTAL] = props[propToBlockColumnsMap[util.BLOCK_JSON_KEY_TOTAL]]

	blockProps[util.BLOCK_JSON_KEY_RESOURCE_ID] = azureBlobStorage.ResourceId
	blockProps[util.BLOCK_JSON_KEY_TIME_GRAIN] = azureBlobStorage.intervalISO8601

	dimensionMap := make(map[string]string)
	dimensionMap[util.BLOCK_JSON_KEY_TENANT] = azureBlobStorage.Tenant
	dimensionMap[util.BLOCK_JSON_KEY_ROLE] = azureBlobStorage.Role
	dimensionMap[util.BLOCK_JSON_KEY_ROLEINSTANCE] = azureBlobStorage.RoleInstance

	blockProps[util.BLOCK_JSON_KEY_DIMENSIONS] = dimensionMap

	var blockByteArray []byte
	blockByteArray, er := json.Marshal(blockProps)
	if er != nil {
		log.Println("Error while parsing metrics properties to json " + util.GetPropsStr(blockProps) + er.Error())
		return "", er
	}

	jsonBlock = string(blockByteArray[:])
	return jsonBlock, nil
}
