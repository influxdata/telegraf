package azureblobstorage

import (
	"encoding/base64"
	"log"
	"strings"

	storage "github.com/Azure/azure-sdk-for-go/storage"
	telegraf "github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
)

type AzureBlobStorage struct {
	AccountName               string //azure storage account name
	SasToken                  string
	ResourceId                string //resource id for the VM or VMSS
	DeploymentId              string
	Periods                   []string //this is the list of periods being configured for various aggregator instances.
	BlobStorageEndPointSuffix string
	HostName                  string
	columnsInTable            []string
	Protocol                  string
	Namespace                 string
	EventName                 string
	EventVersion              string
	container                 *storage.Container
	AgentIdentityHash         string
	blobPath                  string
	BaseTime                  string
	Interval                  string
}

func (azureBlobStorage *AzureBlobStorage) Connect() error {

	blobServiceUrlEndpoint := azureBlobStorage.Protocol + azureBlobStorage.AccountName + azureBlobStorage.BlobStorageEndPointSuffix
	client, er := storage.NewAccountSASClientFromEndpointToken(blobServiceUrlEndpoint, azureBlobStorage.SasToken)
	if er != nil {
		log.Println("error while getti ng client for blob storage " + er.Error())
		return er
	}

	blobClient := client.GetBlobService()

	containerName := strings.ToLower(azureBlobStorage.Namespace + azureBlobStorage.EventName + azureBlobStorage.EventVersion) //getEventVerStr(azureBlobStorage.EventVersion)
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
	azureBlobStorage.blobPath, er = getBlobPath(azureBlobStorage.ResourceId, azureBlobStorage.AgentIdentityHash, azureBlobStorage.BaseTime, azureBlobStorage.Interval)
	if er != nil {
		log.Println("Error while constructing BlobPath" + er.Error())
		return nil
	}
	for i, _ := range metrics {
		props = metrics[i].Fields()
		props["DeploymentId"] = azureBlobStorage.DeploymentId

		//tags := metrics[i].Tags()
		//props["CounterName"] = tags["InputPlugin"] + "/" + props["CounterName"].(string)
		jsonBlock := getJsonBlock(props)

		blockID := base64.StdEncoding.EncodeToString([]byte(jsonBlock))
		blockBlobRef := azureBlobStorage.container.GetBlobReference(azureBlobStorage.blobPath)

	}
	/*// iterate over the list of metrics and create a new entity for each metrics and add to the table.
	for i, _ := range metrics {
		props = metrics[i].Fields()
		props[util.DEPLOYMENT_ID] = azureBlobStorage.DeploymentId
		props[util.HOST] = azureTableStorage.HostName

		tags := metrics[i].Tags()
		props[util.COUNTER_NAME] = tags[util.INPUT_PLUGIN] + "/" + props[util.COUNTER_NAME].(string)

		UTCTicks_DescendingOrderStr, encodedCounterName, er := getRowKeyComponents(props[util.END_TIMESTAMP].(string),
			props[util.COUNTER_NAME].(string))
		if er != nil {
			log.Println("Error: Unable to get valid row key components. Since, this cannot be corrected even on retries hence skipping this row." + util.GetPropsStr(props))
			continue
		}

		periodStr := tags[util.PERIOD]
		table := azureTableStorage.periodVsTableNameVsTableRef[periodStr].TableRef

		//don't write incomplete rows to the table storage
		isValidRow := validateRow(azureTableStorage.columnsInTable, props)
		if isValidRow == false {
			logMsg := "Invalid Row hence not writing it to the table. Row values : " + util.GetPropsStr(props)
			log.Println(logMsg)
			continue
		}
		//two rows are written for each metric as Azure table has optimized prefix search only and no index.
		rowKey1 := UTCTicks_DescendingOrderStr + "_" + encodedCounterName
		rowKey2 := encodedCounterName + "_" + UTCTicks_DescendingOrderStr
		er = writeEntitiesToTable(partitionKey, rowKey1, rowKey2, props, table)
		if er != nil {
			log.Println("Error occured while writing entities to the table")
			return er
		}
	}*/
	return nil
}
func init() {
	outputs.Add("azureblobstorage", func() telegraf.Output {
		return &AzureBlobStorage{}
	})
}
