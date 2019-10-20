package azure_blob

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"

	"github.com/Azure/azure-storage-blob-go/azblob"
)

type loginMethod int

const (
	invalidLogin loginMethod = iota
	accountLogin
	sasLogin
)

const (
	emptyCredentials string = "You need to have either a account name/account key combination or a valid SAS url in order to login"
)

const verbose = true

const (
	blobFormatString     = `https://%s.blob.core.windows.net`
	timeFormatString     = "20060102150405" // YYYYMMDDHHMMSS
	defaultContainerName = "metrics"
)

// AzureBlob allows publishing of metrics to Azure Blob Storage
type AzureBlob struct {
	BlobAccount       string `toml:"blobAccount"`
	BlobAccountKey    string `toml:"blobAccountKey"`
	BlobContainerName string `toml:"blobContainerName"`
	BlobAccountSasURL string `toml:"blobAccountSasURL"`
	FlushInterval     int    `toml:"flushInterval"`
	MachineName       string `toml:"machineName"`

	serializer           serializers.Serializer
	metricsCache         []telegraf.Metric
	timeOfPreviousFlush  time.Time
	blobService          azblob.ServiceURL
	blobContainerService azblob.ContainerURL
}

// Connect initializes the connection to Azure Storage
func (a *AzureBlob) Connect() error {
	login := a.getLoginMethod()

	var err error
	if login == accountLogin {
		err = a.initializeAccountConnection()
	} else if login == sasLogin {
		err = a.initializeSasConnection()
	} else {
		return fmt.Errorf(emptyCredentials)
	}

	if err != nil {
		return err
	}

	// setting hostname to current host's hostname if one was not provided by the user
	if a.MachineName == "" {
		hostname, err := os.Hostname()
		if err != nil {
			return err
		}
		a.MachineName = hostname
	}

	// setting timeOfPreviousFlush to an initial value
	a.timeOfPreviousFlush = time.Now()
	return nil
}

func (a *AzureBlob) initializeAccountConnection() error {
	log(fmt.Sprintf("Initializing Azure Blob output plugin for BlobAccount %s, BlobContainerName %s, MachineName %s and FlushInterval %d seconds\n",
		a.BlobAccount, a.BlobContainerName, a.MachineName, a.FlushInterval))
	c, err := azblob.NewSharedKeyCredential(a.BlobAccount, a.BlobAccountKey)
	if err != nil {
		return err
	}
	u, err := url.Parse(fmt.Sprintf(blobFormatString, a.BlobAccount))
	if err != nil {
		return err
	}
	// create the pipeline
	p := azblob.NewPipeline(c, azblob.PipelineOptions{})
	// inialize Blob Container Service
	a.blobService = azblob.NewServiceURL(*u, p)
	a.blobContainerService = a.blobService.NewContainerURL(a.BlobContainerName)
	// if user hasn't provided a Blob Container Name, set a default
	if a.BlobContainerName == "" {
		a.BlobContainerName = defaultContainerName
	}
	err = a.createContainerIfNotExists(context.Background())
	if err != nil {
		return err
	}
	return nil
}

func (a *AzureBlob) initializeSasConnection() error {
	log(fmt.Sprintf("Initializing Azure Blob output plugin with SAS for URL %s, MachineName %s and FlushInterval %d seconds\n", strings.Split(a.BlobAccountSasURL, "?")[0], a.MachineName, a.FlushInterval))
	c := azblob.NewAnonymousCredential()
	u, err := url.Parse(a.BlobAccountSasURL)
	if err != nil {
		return err
	}
	// create the pipeline
	p := azblob.NewPipeline(c, azblob.PipelineOptions{})
	// recall that the container url *MUST* contain the container name
	a.blobContainerService = azblob.NewContainerURL(*u, p)
	return nil
}

// Close will send whatever is in memory to Azure Storage
func (a *AzureBlob) Close() error {
	log(fmt.Sprintf("Close message received. Synchronously flushing saved events"))

	err := a.flushMetricsCacheToAzureBlob()
	return err
}

// Write will receive metrics and cache them
// if the flushInterval has passed, it will call
func (a *AzureBlob) Write(metrics []telegraf.Metric) error {
	log(fmt.Sprintf("%d metrics cached\n", len(metrics)))

	a.metricsCache = append(a.metricsCache, metrics...)

	// check if it is time to flush all cached metrics
	if time.Now().Sub(a.timeOfPreviousFlush) < time.Duration(a.FlushInterval)*time.Second {
		return nil
	}

	err := a.flushMetricsCacheToAzureBlob()

	if err != nil {
		return err
	}

	a.timeOfPreviousFlush = time.Now()

	// we'll empty the cache only if the createBlob operation succeeded
	a.metricsCache = nil
	return nil
}

func (a *AzureBlob) flushMetricsCacheToAzureBlob() error {
	var sb strings.Builder
	for _, metric := range a.metricsCache {
		bytes, err := a.serializer.Serialize(metric)
		if err != nil {
			return fmt.Errorf("cannot serialize because of %s", err)
		}
		sb.Write(bytes)
	}

	bytes := []byte(sb.String())
	if len(a.metricsCache) == 0 {
		fmt.Println("0 items in cache - will not write anything")
		return nil
	}

	loc, _ := time.LoadLocation("UTC")
	// format is machineName-startDateTime-endDateTime (in UTC format)
	blobName := fmt.Sprintf("%s-%s-%s",
		a.MachineName,
		a.metricsCache[0].Time().In(loc).Format(timeFormatString),
		a.metricsCache[len(a.metricsCache)-1].Time().In(loc).Format(timeFormatString))

	_, err := a.createBlockBlob(context.Background(), blobName, bytes)
	if err != nil {
		return fmt.Errorf("error creating blob: %s", err)
	}

	log(fmt.Sprintf("blob '%s.zip' created successfully\n", blobName))

	return nil
}

func init() {
	outputs.Add("azure_blob", func() telegraf.Output { return &AzureBlob{} })
}

func (a *AzureBlob) getBlockBlobURL(ctx context.Context, blobName string) azblob.BlockBlobURL {
	blob := a.blobContainerService.NewBlockBlobURL(blobName)
	return blob
}

func (a *AzureBlob) createBlockBlob(ctx context.Context, blobName string, byteString []byte) (azblob.BlockBlobURL, error) {
	b := a.getBlockBlobURL(ctx, blobName+".zip")                              // .zip is the file that will be created
	compressedData, err := compressBytesIntoFile(blobName+".txt", byteString) // .txt is the file that will be contained in the zip
	if err != nil {
		return azblob.BlockBlobURL{}, err
	}

	_, err = b.Upload(
		ctx,
		bytes.NewReader(compressedData),
		azblob.BlobHTTPHeaders{
			ContentType: "application/octet-stream",
		},
		azblob.Metadata{},
		azblob.BlobAccessConditions{},
	)

	return b, err
}

func compressBytesIntoFile(filename string, data []byte) ([]byte, error) {
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)
	zipFile, err := zipWriter.Create(filename)
	if err != nil {
		return nil, err
	}
	_, err = zipFile.Write(data)
	if err != nil {
		return nil, err
	}
	err = zipWriter.Close()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Description returns a one-sentence description on the Output
func (AzureBlob) Description() string {
	return "Azure Blob plugin periodically sends zipped telegraf data to a specified Azure Blob container"
}

// SampleConfig returns the default configuration of the Output
func (AzureBlob) SampleConfig() string {
	return `
	## You need to have either an accountName/accountKey combination or a SAS URL
	## SAS URL should contain the Blob Container Name and have appropriate permissions (create and write)
	## Azure Blob account
	# blobAccount = "myblobaccount"
	## Azure Blob account key
	# blobAccountKey = "myblobaccountkey"
	## Azure Blob container name. Used only when authenticating via accountName. If omitted, "metrics" is used
	# blobContainerName = "telegrafcontainer"
	## Azure Blob Container SAS URL
	# blobAccountSasURL = "YOUR_SAS_URL"
	## Flush interval in seconds
	# flushInterval = 300
	## Machine name that is sending the data
	# machineName = "myhostname"
	## Data format to output.
	## Each data format has its own unique set of configuration options, read
	## more about them here:
	## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
	# data_format = "json"
	`
}

func (a *AzureBlob) getLoginMethod() loginMethod {
	if a.BlobAccount != "" && a.BlobAccountKey != "" {
		return accountLogin
	}
	if a.BlobAccountSasURL != "" {
		return sasLogin
	}
	return invalidLogin
}

func (a *AzureBlob) createContainerIfNotExists(ctx context.Context) error {
	exists, err := a.checkIfContainerExists(ctx)
	if err != nil {
		return err
	}
	if !exists {
		resp, err := a.blobContainerService.Create(ctx, azblob.Metadata{}, azblob.PublicAccessNone)
		if err != nil {
			return err
		}
		log(fmt.Sprintf("Container %s creation status code is %d\n", a.BlobContainerName, resp.StatusCode()))
	}
	return nil
}

// SetSerializer sets the data export format
func (a *AzureBlob) SetSerializer(serializer serializers.Serializer) {
	a.serializer = serializer
}

func (a *AzureBlob) checkIfContainerExists(ctx context.Context) (bool, error) {
	resp, err := a.blobService.ListContainersSegment(ctx, azblob.Marker{}, azblob.ListContainersSegmentOptions{
		Prefix: a.BlobContainerName,
	})
	if err != nil {
		return false, err
	}
	for _, containerItem := range resp.ContainerItems {
		if containerItem.Name == a.BlobContainerName {
			return true, nil
		}
	}
	return false, nil
}

func log(msg string) {
	if verbose {
		fmt.Print(msg)
	}
}
