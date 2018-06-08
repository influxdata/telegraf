package azureblobstorage

import (
	"encoding/base64"
	"log"
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
	MaxBlockSize              int
	//fileWithBlockIds file contains list of blockIds which are written to a particular blob path
	//this permanent persistence of blockIds in a blob is important for situations where telegraf crashes and
	//on its startup, if the blob path is same( can happen if telegraf crashes and restarts within 1 hr
	//or the value of configured interval). So, if the blob path is same then in order to keep the previously written blocks to the blob
	//the block ids of old blocks are to be sent while calling PutBlockList API.
	FileWithBlockIds  string
	blobInstanceProp  blobInstancePropType
	container         *storage.Container
	requiredFieldList map[string][]string
	requiredFieldSize int

	isBlobClientCreated bool
	isContainerCreated  bool
	blobClient          storage.BlobStorageClient
}

type blobInstancePropType struct {
	blobPath          string
	blockIds          []string
	blockIdEncodedLen int
	blockIdDecodedLen int
	oldBlocks         BlobPathBlockIds
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

var sampleConfig = `
[[outputs.azureblobstorage]]
 resource_id = "subscriptionId/resourceGroup/VMScaleset"
 account_name = "nirastoladdiag466"
 sas_token=""
  #periods is the list of period configured for each aggregator plugin
 interval = "3600s"
 blob_storage_end_point_suffix = ".blob.core.windows.net"
 protocol = "https://"
 event_name = "containerName2" #xmlCfg.xml -> eventname property
 event_version = "2" # xmlCfg.xml <MonitoringManagement eventVersion="2" namespace="" timestamp="2017-03-27T19:45:00.000" version="1.0">
 namespace = ""
 agent_identity_hash = "" # present in xmlCfg.xml this value is read from file /sys/class/dmi/id/product_uuid on the VM
 tenant=""
 role=""
 role_instance=""
 base_time="1527865026" #start time when first container is to be created.
 max_block_size=1000000 #1MB
 file_with_block_ids=""
`

func (azureBlobStorage *AzureBlobStorage) initializeProperties() error {
	var er error
	azureBlobStorage.isBlobClientCreated = false
	azureBlobStorage.isContainerCreated = false
	if azureBlobStorage.Interval == "" {
		azureBlobStorage.Interval = "3600s" //PT1H default
	}
	azureBlobStorage.intervalISO8601, er = util.GetIntervalISO8601(azureBlobStorage.Interval)
	if er != nil {
		log.Println("E! ERROR while Parsing interval to ISO8601 format " + azureBlobStorage.Interval + er.Error())
		return er
	}
	if azureBlobStorage.BaseTime == "" {
		azureBlobStorage.BaseTime = strconv.FormatInt(time.Now().Unix(), 10)
	}
	//so that only one blob is created in each interval
	azureBlobStorage.BaseTime, er = getBaseTimeMultipleOfInterval(azureBlobStorage.BaseTime, azureBlobStorage.Interval)
	if er != nil {
		log.Println("E! ERROR while converting base time as multiple of interval baseTime,interval " +
			azureBlobStorage.BaseTime + " " +
			azureBlobStorage.Interval + " " +
			er.Error())
		return er
	}
	azureBlobStorage.blobInstanceProp.blockIdEncodedLen = 60
	azureBlobStorage.blobInstanceProp.blockIdDecodedLen = base64.StdEncoding.DecodedLen(azureBlobStorage.blobInstanceProp.blockIdEncodedLen)
	azureBlobStorage.requiredFieldList, azureBlobStorage.requiredFieldSize = getRequiredFieldList()
	azureBlobStorage.blobInstanceProp.oldBlocks, er = azureBlobStorage.checkOldBlocksForBlob()
	if er != nil {
		log.Println("E! Error while reading in old blocks for this blob. If any prev blocks they might be lost" + er.Error())
	}
	return nil
}
func (azureBlobStorage *AzureBlobStorage) setBlobClient() error {
	var blobClient storage.BlobStorageClient
	blobServiceUrlEndpoint := azureBlobStorage.Protocol + azureBlobStorage.AccountName + azureBlobStorage.BlobStorageEndPointSuffix
	client, er := storage.NewAccountSASClientFromEndpointToken(blobServiceUrlEndpoint, azureBlobStorage.SasToken)
	if er != nil {
		log.Println("E! ERROR while getting client for blob storage " + er.Error())
		return er
	}

	blobClient = client.GetBlobService()
	log.Println("I! INFO validating blobClient")
	er = validateBlobClient(blobClient)
	if er != nil {
		log.Println("Error Invalid blob client " + er.Error())
		return er
	}
	log.Println("I! INFO successfully validated BlobClient")
	azureBlobStorage.blobClient = blobClient
	azureBlobStorage.isBlobClientCreated = true
	log.Println("I! INFO successfully created blob client")
	return nil
}

func (azureBlobStorage *AzureBlobStorage) Connect() error {
	log.Println("I! INFO initializing properties of azure blob storage ")
	er := azureBlobStorage.initializeProperties()
	if er != nil {
		log.Println("E! ERROR while initializing properties of blob storage plugin object " + er.Error())
		return er
	}
	log.Println("I! INFO successfully initialized")
	log.Println("I! INFO attempting to set blob storage client")
	er = azureBlobStorage.setBlobClient()
	if er != nil {
		log.Println("E! ERROR while creating Blob Client " + er.Error())
	}

	er = azureBlobStorage.createBlobContainer()
	if er != nil {
		log.Println("E! ERROR while creating container" + er.Error())
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
	return sampleConfig
}

// Write takes in group of points to be written to the Output
func (azureBlobStorage *AzureBlobStorage) Write(metrics []telegraf.Metric) error {
	var props map[string]interface{}
	var er error
	er = azureBlobStorage.checkBlobClientContainer()
	if er != nil {
		log.Println("E! ERROR while creating Blob Client and/or container " + er.Error())
		log.Println("E! ERROr skipping metrics ")
		return er
	}

	var jsonList string
	jsonList = "[\n"
	var i int
	for i, _ = range metrics {

		props = metrics[i].Fields()
		tags := metrics[i].Tags()

		props[util.COUNTER_NAME] = tags[util.INPUT_PLUGIN] + "/" + props[util.COUNTER_NAME].(string)

		er = azureBlobStorage.setCurrentBlobPath()
		if er != nil {
			//irrecoverable error, hence logging error and discarding writing blocks to it
			log.Println("E! ERROR while setting blobPath skipping writing metrics to this blobpath " + util.GetPropsStr(props) + er.Error())
			continue
		}

		jsonObject := azureBlobStorage.getJsonObject(props)
		//	isValidJsonRow := validateObject(jsonObject)
		//	log.Println(strconv.FormatBool(isValidJsonRow))
		jsonBlock, err := getJsonBlock(jsonObject)
		if err != nil {
			//irrecoverable error, hence logging error and discarding writing metric
			log.Println("E! ERROR while converting metrics fields to json, metric is not sent to blob storage " +
				azureBlobStorage.container.Name +
				util.GetPropsStr(props) +
				err.Error())
			continue
		}
		jsonList = jsonList + jsonBlock + ",\n"

		blockSize := len([]byte(jsonList))
		if blockSize >= azureBlobStorage.MaxBlockSize {
			jsonList = jsonList + "]"
			er = azureBlobStorage.flush(jsonList)
			if er != nil {
				log.Println("E! ERROR while writing block to blob storage " + er.Error())
				log.Println("I!" + jsonList)
				return er
			}
			jsonList = "[\n"
		} /*else {
			if i < len(metrics)-1 {
				jsonList = jsonList + ",\n"
			}
		}*/
	}
	jsonList = jsonList + "\n]"
	//if the flush() happened when last element of the metrics[] was getting processed then jsonList wil be equal to "[\n\n]"
	if i == len(metrics)-1 && len(strings.TrimSpace(jsonList)) > 4 {
		er = azureBlobStorage.flush(jsonList)
		if er != nil {
			log.Println("E! ERROR while writing block to blob storage " + er.Error())
			log.Println("I! " + jsonList)
			return er
		}
	}

	return nil
}

func (azureBlobStorage *AzureBlobStorage) flush(blockJsonListStr string) error {

	blockBlobRef := azureBlobStorage.container.GetBlobReference(azureBlobStorage.blobInstanceProp.blobPath)
	er := validateBlobRef(blockBlobRef)
	if er != nil {
		log.Println("E! ERROR invalid BlobReference for container,blob path " +
			azureBlobStorage.container.Name + " " +
			azureBlobStorage.blobInstanceProp.blobPath +
			er.Error())
		return er
	}
	blockId, er := azureBlobStorage.writeJsonBlockToBlob(blockJsonListStr, blockBlobRef)
	if er != nil {
		log.Println("!E ERROR while writing block to blob storage blockId,content" + blockId + blockJsonListStr + er.Error())
		return er
	} else {
		log.Println("I! INFO Success: Written block to storage" + blockId)
	}
	return nil
}

func init() {
	outputs.Add("azureblobstorage", func() telegraf.Output {
		return &AzureBlobStorage{}
	})
}
