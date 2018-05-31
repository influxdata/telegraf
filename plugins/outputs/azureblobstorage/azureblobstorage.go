package azureblobstorage

import (
	"log"
	"strings"

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
	container                 *storage.Container
	AgentIdentityHash         string
	blobPath                  string
	BaseTime                  string
	Interval                  string
	intervalISO8601           string
	Role                      string
	RoleInstance              string
	Tenant                    string
}
func (azureBlobStorage *AzureBlobStorage) initialize() {
	var er error
	azureBlobStorage.intervalISO8601, er = getIntervalISO8601(azureBlobStorage.Interval)
	if er != nil {
		//log this error and stop as this is not a transient error which will get fixed on retries.
		log.Fatal("Error while Parsing interval to ISO8601 format " + azureBlobStorage.Interval + er.Error())
	}
}
func (azureBlobStorage *AzureBlobStorage) Connect() error {
	azureBlobStorage.initialize()
	blobServiceUrlEndpoint := azureBlobStorage.Protocol + azureBlobStorage.AccountName + azureBlobStorage.BlobStorageEndPointSuffix
	client, er := storage.NewAccountSASClientFromEndpointToken(blobServiceUrlEndpoint, azureBlobStorage.SasToken)
	if er != nil {
		log.Println("error while getti ng client for blob storage " + er.Error())
		return er
	}
	//TODO: validate client
	blobClient := client.GetBlobService()

	containerName := strings.ToLower(azureBlobStorage.Namespace + azureBlobStorage.EventName) //getEventVerStr(azureBlobStorage.EventVersion))
	//TODO: validate container ref
	azureBlobStorage.container = blobClient.GetContainerReference(containerName)
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

// Write takes in group of points to be written to the Output
func (azureBlobStorage *AzureBlobStorage) Write(metrics []telegraf.Metric) error {
	var props map[string]interface{}
	var er error
	azureBlobStorage.blobPath, er = getBlobPath(azureBlobStorage.ResourceId, azureBlobStorage.AgentIdentityHash, azureBlobStorage.BaseTime, azureBlobStorage.intervalISO8601)
	if er != nil {
		log.Println("Error while constructing BlobPath" + er.Error())
		return nil
	}

	for i, _ := range metrics {
		props = metrics[i].Fields()

		tags := metrics[i].Tags()
		props[util.COUNTER_NAME] = tags[util.INPUT_PLUGIN] + "/" + props[util.COUNTER_NAME].(string)
		jsonBlock, err := getJsonBlock(props, azureBlobStorage)

		if err != nil {
			//discarding the metric as this error is not recoverable.
			log.Println("Error while converting metrics fields to json metric is not sent to blob storage " + azureBlobStorage.container.Name + util.GetPropsStr(props) + err.Error())
			continue
		}
		//TODO: validate BlobRef
		blockBlobRef := azureBlobStorage.container.GetBlobReference(azureBlobStorage.blobPath)
		blockId, er := writeJsonBlockToBlob(jsonBlock, blockBlobRef)
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
