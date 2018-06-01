package azureblobstorage

import (
	"encoding/json"
	"log"
	"reflect"
	"strconv"
	"strings"
	"time"

	storage "github.com/Azure/azure-sdk-for-go/storage"
	telegraf "github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
	util "github.com/influxdata/telegraf/utility"
)

type AzureBlobStorage struct {
	AccountName               string //azure storage account name
	SasToken                  string
	ResourceId                string //resource id for the VM or VMSS//this is the list of periods being configured for various aggregator instances.
	BlobStorageEndPointSuffix string
	Protocol                  string
	Namespace                 string
	EventName                 string
	EventVersion              string
	AgentIdentityHash         string
	BaseTime                  string
	Interval                  string
	intervalISO8601           string
	Role                      string
	RoleInstance              string
	Tenant                    string
	blobPath                  string
	container                 *storage.Container
	requiredFieldList         map[string][]string
	requiredFieldSize         int
}

type Dimensions struct {
	Tenant       string `json:"Tenant"`
	Role         string `json:"Role"`
	RoleInstance string `json:"RoleInstance"`
}
type BlockObject struct {
	Time       string     `json:"time"`
	ResourceId string     `json:"resourceId"`
	Timegrain  string     `json:"timeGrain"`
	Dimensions Dimensions `json:"dimension"`
	MetricName string     `json:"metricName"`
	Total      float64    `json:"total"`
	Minimum    float64    `json:"minimum"`
	Maximum    float64    `json:"maximum"`
	Average    float64    `json:"average"`
	Count      float64    `json:"count"`
	Last       float64    `json:"last"`
}

func (azureBlobStorage *AzureBlobStorage) initializeProperties() error {
	var er error
	if azureBlobStorage.Interval == "" {
		azureBlobStorage.Interval = "3600s" //PT1H default
	}
	azureBlobStorage.intervalISO8601, er = util.GetIntervalISO8601(azureBlobStorage.Interval)
	if er != nil {
		log.Println("Error while Parsing interval to ISO8601 format " + azureBlobStorage.Interval + er.Error())
		return er
	}
	if azureBlobStorage.BaseTime == "" {
		azureBlobStorage.BaseTime = strconv.FormatInt(time.Now().Unix(), 10)
	}
	azureBlobStorage.BaseTime, er = getBaseTimeMultipleOfInterval(azureBlobStorage.BaseTime, azureBlobStorage.Interval)
	if er != nil {
		log.Println("Error while converting base time as multiple of interval baseTime,interval " +
			azureBlobStorage.BaseTime + " " +
			azureBlobStorage.Interval + " " +
			er.Error())
		return er
	}
	azureBlobStorage.requiredFieldList, azureBlobStorage.requiredFieldSize = getRequiredFieldList()

	return nil
}

func (azureBlobStorage *AzureBlobStorage) Connect() error {
	er := azureBlobStorage.initializeProperties()
	if er != nil {
		log.Println("Error while initializing properties of blob storage plugin object " + er.Error())
		return er
	}
	blobServiceUrlEndpoint := azureBlobStorage.Protocol + azureBlobStorage.AccountName + azureBlobStorage.BlobStorageEndPointSuffix
	client, er := storage.NewAccountSASClientFromEndpointToken(blobServiceUrlEndpoint, azureBlobStorage.SasToken)
	if er != nil {
		log.Println("error while getti ng client for blob storage " + er.Error())
		return er
	}

	blobClient := client.GetBlobService()
	er = validateBlobClient(blobClient)
	if er != nil {
		log.Println("Error Invalid blob client " + er.Error())
		return er
	}

	eventVersion, er := getEventVersionStr(azureBlobStorage.EventVersion)
	if er != nil {
		log.Println("Error while getting event version hence ignoring it while constructing name of container")
		eventVersion = ""
	}

	containerName := strings.ToLower(azureBlobStorage.Namespace + azureBlobStorage.EventName + eventVersion)
	azureBlobStorage.container = blobClient.GetContainerReference(containerName)
	er = validateContainerRef(azureBlobStorage.container)
	if er != nil {
		log.Println("Error Invalid container reference for " + containerName + er.Error())
		return er
	}

	options := storage.CreateContainerOptions{
		Access: storage.ContainerAccessTypeBlob,
	}
	isCreated, er := azureBlobStorage.container.CreateIfNotExists(&options)
	if er != nil {
		log.Println("Error while creating container " + containerName)
		log.Println(er.Error())
		return er
	}
	if isCreated {
		log.Println("Created container " + containerName)
	} else {
		log.Println("Container already exists " + containerName)
	}

	return nil
}

func (azureBlobStorage *AzureBlobStorage) Close() error {
	return nil
}

// Description returns a one-sentence description on the Output
func (azureBlobStorage *AzureBlobStorage) Description() string {
	return ""
}

// SampleConfig returns the default configuration of the Output
func (azureBlobStorage *AzureBlobStorage) SampleConfig() string {
	return ""
}
func (azureBlobStorage *AzureBlobStorage) setCurrentBlobPath() error {
	var er error
	isIntervalOver, er := checkIsIntervalOver(azureBlobStorage.BaseTime, azureBlobStorage.Interval)
	if er != nil {
		log.Println("Error while checking if new bolb is to be constructed baseTime, interval " +
			azureBlobStorage.BaseTime + " " +
			azureBlobStorage.Interval + " " +
			er.Error())
		return er
	}
	if isIntervalOver == true {
		newBaseTime, er := getNewBaseTime(azureBlobStorage.BaseTime, azureBlobStorage.Interval)
		if er != nil {
			log.Println("Error while setting new base time by adding interval to basetime " +
				azureBlobStorage.BaseTime + " " +
				azureBlobStorage.Interval + " " +
				er.Error())
			return er
		}
		azureBlobStorage.BaseTime = newBaseTime
		azureBlobStorage.blobPath, er = getBlobPath(azureBlobStorage.ResourceId, azureBlobStorage.AgentIdentityHash, azureBlobStorage.BaseTime, azureBlobStorage.intervalISO8601)
		if er != nil {
			log.Println("Error while constructing BlobPath" + azureBlobStorage.ResourceId + " " +
				azureBlobStorage.AgentIdentityHash + " " +
				azureBlobStorage.BaseTime + " " +
				azureBlobStorage.intervalISO8601 + " " +
				er.Error())
			return er
		}
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

// Write takes in group of points to be written to the Output
func (azureBlobStorage *AzureBlobStorage) Write(metrics []telegraf.Metric) error {
	var props map[string]interface{}
	var er error

	for i, _ := range metrics {

		props = metrics[i].Fields()
		tags := metrics[i].Tags()

		props[util.COUNTER_NAME] = tags[util.INPUT_PLUGIN] + "/" + props[util.COUNTER_NAME].(string)

		er = azureBlobStorage.setCurrentBlobPath()
		if er != nil {
			//irrecoverable error, hence logging error and discarding writing blocks to it
			log.Println("Error while setting blobPath skipping writing metrics to this blobpath " + util.GetPropsStr(props) + er.Error())
			continue
		}
		jsonObject := azureBlobStorage.getJsonObject(props)
		//	isValidJsonRow := validateObject(jsonObject)
		//	log.Println(strconv.FormatBool(isValidJsonRow))
		jsonBlock, err := getJsonBlock(jsonObject)
		if err != nil {
			//irrecoverable error, hence logging error and discarding writing metric
			log.Println("Error while converting metrics fields to json, metric is not sent to blob storage " +
				azureBlobStorage.container.Name +
				util.GetPropsStr(props) +
				err.Error())
			continue
		}

		blockBlobRef := azureBlobStorage.container.GetBlobReference(azureBlobStorage.blobPath)
		er = validateBlobRef(blockBlobRef)
		if er != nil {
			log.Println("Error invalid BlobReference for container,blob path " +
				azureBlobStorage.container.Name + " " +
				azureBlobStorage.blobPath +
				er.Error())
			return er
		}
		//isValidJson := validateJsonRow(jsonObject, jsonBlock)
		blockId, er := writeJsonBlockToBlob(azureBlobStorage.requiredFieldList,
			azureBlobStorage.requiredFieldSize, jsonBlock, blockBlobRef)
		if er != nil {
			log.Println("Error while writing block to blob storage blockId,content" + blockId + jsonBlock + er.Error())
			return er
		} else {
			log.Println("Success: Written block to storage" + blockId)
		}
	}
	return nil
}
func init() {
	outputs.Add("azureblobstorage", func() telegraf.Output {
		return &AzureBlobStorage{}
	})
}
