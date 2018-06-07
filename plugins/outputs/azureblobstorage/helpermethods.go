package azureblobstorage

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"math"
	"reflect"
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
func (azureBlobStorage *AzureBlobStorage) writeJsonBlockToBlob(jsonBlock string, blockBlobRef *storage.Blob) (string, error) {

	md5HashOfBlock, er := util.Getmd5Hash(jsonBlock)
	if er != nil {
		log.Println("E! ERROR while calculating md5 hash of content " + jsonBlock)
		return "", nil
	}
	log.Println("I! INFO mdd5 hash of block is " + md5HashOfBlock)
	blockId, ok := azureBlobStorage.getBlockId(md5HashOfBlock)
	if !ok {
		log.Println("I! invalid block id skipping this batch " + jsonBlock)
		return blockId, nil
	}
	log.Println("I! INFO attempting to write block with blockId: " + blockId)
	options := storage.PutBlockOptions{
		Timeout:    30,             // in seconds
		ContentMD5: md5HashOfBlock, // to ensure integrity of the content of block being sent to the storage
		RequestID:  md5HashOfBlock,
	}
	log.Println("I! INFO Request Id " + md5HashOfBlock)
	er = blockBlobRef.PutBlock(blockId, []byte(jsonBlock), &options)
	if er != nil {
		log.Println("E! ERROR while writing block to blob storage blockId,content" + blockId + jsonBlock + er.Error())
		return "", er
	}

	requestId, err := util.Getmd5Hash(base64.StdEncoding.EncodeToString([]byte(blockId)))
	log.Println("I! INFO Committing block with request Id " + requestId + "blockId:" + blockId)
	if err != nil {
		log.Println("E! ERROr while computing md5 hash of blockid " + blockId)
		return "", err
	}
	putBlockListOptions := storage.PutBlockListOptions{
		RequestID: requestId,
	}
	var blockList []storage.Block
	for i := range azureBlobStorage.blobInstanceProp.blockIds {
		blockList = append(blockList, storage.Block{azureBlobStorage.blobInstanceProp.blockIds[i], storage.BlockStatusCommitted})
	}
	blockList = append(blockList, storage.Block{blockId, storage.BlockStatusUncommitted})
	azureBlobStorage.blobInstanceProp.blockIds = append(azureBlobStorage.blobInstanceProp.blockIds, blockId)
	//blockList = []storage.Block{{blockId, storage.BlockStatusUncommitted}}
	er = blockBlobRef.PutBlockList(blockList, &putBlockListOptions)
	if er != nil {
		log.Println("E! ERROR while writing block to blob storage blockId,content" + blockId + jsonBlock + er.Error())
		return "", er
	}
	log.Println("I! Successfully written block with block id" + blockId)
	return blockId, nil
}

func (azureBlobStorage *AzureBlobStorage) getBlockId(md5HashOfBlock string) (string, bool) {
	blockId := ""
	//block Id: the length of string pre encoding should be less than 64
	//using md5 hash of block content instead of block content as the firs 60 characters of the content are usually same
	//hence resulting in same block id for different blocks which will override previous block
	//each time new block with same block id is written to blob.

	blockIdValue := md5HashOfBlock + strconv.Itoa(util.GetRand().Intn(9999))
	if len(blockIdValue) > azureBlobStorage.blobInstanceProp.blockIdDecodedLen {
		blockId = base64.StdEncoding.EncodeToString([]byte(blockIdValue[(len(blockIdValue) - azureBlobStorage.blobInstanceProp.blockIdDecodedLen):]))
	} else {
		diffLen := float64(azureBlobStorage.blobInstanceProp.blockIdDecodedLen - len(blockIdValue))
		//	minRange := int64(math.Pow10(diffLen - 1))
		maxRange := math.Pow(10, diffLen) - 1
		//	randeDiff := maxRange - minRange
		blockIdValue = blockIdValue + strconv.FormatInt(int64(math.Floor(util.GetRand().Float64()*maxRange)), 10)
		blockId = base64.StdEncoding.EncodeToString([]byte(blockIdValue))
		if len(blockId) != azureBlobStorage.blobInstanceProp.blockIdEncodedLen {
			log.Println("I! the length of block id is not " + strconv.Itoa(len(blockId)) + blockId + strconv.Itoa(azureBlobStorage.blobInstanceProp.blockIdEncodedLen))
			return blockId, false
		}
	}
	return blockId, true
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

func (azureBlobStorage *AzureBlobStorage) checkBlobClientContainer() error {
	if azureBlobStorage.isBlobClientCreated == false {
		er := azureBlobStorage.setBlobClient()
		if er != nil {
			log.Println("E! Error while creating Blob Client " + er.Error())
			return er
		} else {
			azureBlobStorage.isBlobClientCreated = true
			log.Println("I! INFO blob client created successfully")
		}
	}
	if azureBlobStorage.isContainerCreated == false {
		er := azureBlobStorage.createBlobContainer()
		if er != nil {
			log.Println("E! Error while creating Blob Container " + er.Error())
			return er
		} else {
			azureBlobStorage.isContainerCreated = true
			log.Println("I! INFO container created successfully")
		}
	}
	return nil
}

func (azureBlobStorage *AzureBlobStorage) setCurrentBlobPath() error {
	var er error
	isIntervalOver, er := checkIsIntervalOver(azureBlobStorage.BaseTime, azureBlobStorage.Interval)
	if er != nil {
		log.Println("E! ERROR while checking if new bolb is to be constructed baseTime, interval " +
			azureBlobStorage.BaseTime + " " +
			azureBlobStorage.Interval + " " +
			er.Error())
		return er
	}
	if isIntervalOver == true {
		newBaseTime, er := getNewBaseTime(azureBlobStorage.BaseTime, azureBlobStorage.Interval)
		if er != nil {
			log.Println("E! ERROR while setting new base time by adding interval to basetime " +
				azureBlobStorage.BaseTime + " " +
				azureBlobStorage.Interval + " " +
				er.Error())
			return er
		}
		azureBlobStorage.BaseTime = newBaseTime
		azureBlobStorage.blobInstanceProp.blobPath, er = getBlobPath(azureBlobStorage.ResourceId, azureBlobStorage.AgentIdentityHash, azureBlobStorage.BaseTime, azureBlobStorage.intervalISO8601)
		if er != nil {
			log.Println("E! ERROR while constructing BlobPath" + azureBlobStorage.ResourceId + " " +
				azureBlobStorage.AgentIdentityHash + " " +
				azureBlobStorage.BaseTime + " " +
				azureBlobStorage.intervalISO8601 + " " +
				er.Error())
			return er
		}
	}
	return nil
}

func (azureBlobStorage *AzureBlobStorage) getJsonObject(props map[string]interface{}) BlockObject {
	jsonObject := BlockObject{}
	jsonObject.Total = props[util.TOTAL].(float64)
	jsonObject.Timegrain = azureBlobStorage.intervalISO8601
	jsonObject.Time = props[util.END_TIMESTAMP].(string)
	jsonObject.ResourceId = azureBlobStorage.ResourceId
	jsonObject.Minimum = props[util.MIN_SAMPLE].(float64)
	jsonObject.Maximum = props[util.MAX_SAMPLE].(float64)
	jsonObject.MetricName = props[util.COUNTER_NAME].(string)
	jsonObject.Last = props[util.LAST_SAMPLE].(float64)
	jsonObject.Count = props[util.SAMPLE_COUNT].(float64)
	jsonObject.Average = props[util.MEAN].(float64)

	dimensionObj := Dimensions{}
	dimensionObj.Role = azureBlobStorage.Role
	dimensionObj.RoleInstance = azureBlobStorage.RoleInstance
	dimensionObj.Tenant = azureBlobStorage.Tenant

	jsonObject.Dimensions = dimensionObj
	return jsonObject
}
func checkIsValueValid(value interface{}) bool {
	isValid := true
	switch typeOfValue := value.(type) {
	case string:
		if value.(string) == "" {
			isValid = false
		}
		break
	case float64:
		break
	case Dimensions:
		isValid = validateObject1(value.(Dimensions))
		_ = typeOfValue
	}
	return isValid
}
func validateObject1(object Dimensions) bool {
	isValid := true
	s := reflect.ValueOf(&object).Elem()
	_ = s.Type()
	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		isValid = checkIsValueValid(f.Interface())
		if isValid == false {
			break
		}
	}
	return isValid
}
func validateObject(object BlockObject) bool {
	isValid := true
	s := reflect.ValueOf(&object).Elem()
	_ = s.Type()
	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		isValid = checkIsValueValid(f.Interface())
		if isValid == false {
			break
		}
	}
	return isValid
}

func compareBlockObject(obj1 BlockObject, obj2 BlockObject) bool {
	return true
}
func validateJsonRow(jsonObject BlockObject, jsonBlock string) bool {
	var jsonBlockObject BlockObject
	er := json.Unmarshal([]byte(jsonBlock), &jsonBlockObject)
	if er != nil {

	}
	isValid := true
	isValid = compareBlockObject(jsonObject, jsonBlockObject)
	return isValid
}

func (azureBlobStorage *AzureBlobStorage) createBlobContainer() error {

	eventVersion, er := getEventVersionStr(azureBlobStorage.EventVersion)
	if er != nil {
		log.Println("E! ERROR while getting event version hence ignoring it while constructing name of container" + er.Error())
		eventVersion = ""
	}
	containerName := strings.ToLower(azureBlobStorage.Namespace + azureBlobStorage.EventName + eventVersion)
	azureBlobStorage.container = azureBlobStorage.blobClient.GetContainerReference(containerName)
	er = validateContainerRef(azureBlobStorage.container)
	if er != nil {
		log.Println("E! ERROR Invalid container reference for " + containerName + er.Error())
		return er
	}

	options := storage.CreateContainerOptions{
		Access: storage.ContainerAccessTypeBlob,
	}
	log.Println("I! INFO attempting to create container " + containerName)
	isCreated, er := azureBlobStorage.container.CreateIfNotExists(&options)
	if er != nil {
		log.Println("E! ERROR while creating container " + containerName)
		log.Println(er.Error())
		return er
	}
	if isCreated {
		log.Println("I! INFO Created container " + containerName)
	} else {
		log.Println("I! INFO Container already exists " + containerName)
	}
	azureBlobStorage.isContainerCreated = true
	return nil
}
