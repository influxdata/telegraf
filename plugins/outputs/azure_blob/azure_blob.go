package azure_blob

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"

	"github.com/Azure/azure-storage-blob-go/azblob"
)

var metricsCache []telegraf.Metric
var timeOfPreviousFlush time.Time
var blobService azblob.ServiceURL
var blobContainerService azblob.ContainerURL

const (
	blobFormatString = `https://%s.blob.core.windows.net`
	timeFormatString = "20060102150405" // YYYYMMDDHHMMSS
)

// AzureBlob allows publishing of metrics to Azure Blob Storage
type AzureBlob struct {
	BlobAccount       string `toml:"blobAccount"`
	BlobAccountKey    string `toml:"blobAccountKey"`
	BlobContainerName string `toml:"blobContainerName"`
	FlushInterval     int    `toml:"flushInterval"`
	MachineName       string `toml:"machineName"`
}

// Connect to the Output
func (s *AzureBlob) Connect() error {
	c, err := azblob.NewSharedKeyCredential(s.BlobAccount, s.BlobAccountKey)
	if err != nil {
		return err
	}
	p := azblob.NewPipeline(c, azblob.PipelineOptions{})
	u, err := url.Parse(fmt.Sprintf(blobFormatString, s.BlobAccount))
	if err != nil {
		return err
	}
	blobService = azblob.NewServiceURL(*u, p)
	blobContainerService = blobService.NewContainerURL(s.BlobContainerName)

	err = createContainerIfNotExists(context.TODO(), s.BlobContainerName)
	if err != nil {
		return err
	}

	// setting time to an initial value
	timeOfPreviousFlush = time.Now()
	return nil
}

// Close any connections to the Output
func (s *AzureBlob) Close() error {
	fmt.Println("Close message received. Synchronously flushing saved events")

	tempCache := make([]telegraf.Metric, len(metricsCache))
	copy(tempCache, metricsCache)
	s.flushMetricsMemoryCacheToAzureBlob(tempCache)

	return nil
}

// Write takes in group of points to be written to the Output
func (s *AzureBlob) Write(metrics []telegraf.Metric) error {
	fmt.Printf("Write %d\n", len(metrics))

	metricsCache = append(metricsCache, metrics...)

	if time.Now().Sub(timeOfPreviousFlush) < time.Duration(s.FlushInterval)*time.Second {
		return nil
	}

	tempCache := make([]telegraf.Metric, len(metricsCache))
	copy(tempCache, metricsCache)
	s.flushMetricsMemoryCacheToAzureBlob(tempCache)
	metricsCache = nil

	timeOfPreviousFlush = time.Now()

	return nil
}

func (s *AzureBlob) flushMetricsCacheToAzureBlob(cache []telegraf.Metric) {

}

func (s *AzureBlob) flushMetricsMemoryCacheToAzureBlob(cache []telegraf.Metric) {
	fmt.Printf("Flush %d\n", len(cache))

	var str strings.Builder
	for _, metric := range cache {
		_, err := str.WriteString(fmt.Sprintf("%v\n", metric))
		if err != nil {
			fmt.Printf("error creating string in flush because: %s. Recovering...\n", err)
		}
	}

	if len(cache) == 0 {
		fmt.Println("0 items in cache - will not write anything")
		return
	}

	// format is endDate-startDate-machineName.txt
	blobName := fmt.Sprintf("%s-%s-%s.zip", cache[len(cache)-1].Time().Format(timeFormatString),
		cache[0].Time().Format(timeFormatString), s.MachineName)

	_, err := s.createBlockBlob(context.TODO(), blobName, str.String())
	if err != nil {
		fmt.Printf("error writing to blob because of %s\n", err)
	} else {
		fmt.Printf("Blob named %s written successfully\n", blobName)
	}
}

func init() {
	outputs.Add("azure_blob", func() telegraf.Output { return &AzureBlob{} })
}

func (s *AzureBlob) getBlockBlobURL(ctx context.Context, blobName string) azblob.BlockBlobURL {
	blob := blobContainerService.NewBlockBlobURL(blobName)
	return blob
}

func (s *AzureBlob) createBlockBlob(ctx context.Context, blobName string, data string) (azblob.BlockBlobURL, error) {
	b := s.getBlockBlobURL(ctx, blobName)

	compressedData, err := compressString(data)

	if err != nil {
		return azblob.BlockBlobURL{}, err
	}

	_, err = b.Upload(
		ctx,
		//strings.NewReader(data),
		bytes.NewReader(compressedData),
		azblob.BlobHTTPHeaders{
			//ContentType: "text/plain",
			ContentType: "application/octet-stream",
		},
		azblob.Metadata{},
		azblob.BlobAccessConditions{},
	)

	return b, err
}

func compressString(data string) ([]byte, error) {
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	if _, err := gz.Write([]byte(data)); err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// Description returns a one-sentence description on the Output
func (s *AzureBlob) Description() string {
	return "Azure Blob plugin sends telegraf data to a Azure Blob"
}

// SampleConfig returns the default configuration of the Output
func (s *AzureBlob) SampleConfig() string {
	return `
	## Azure Blob account
	# blobAccount = "myblobaccount"
	## Azure Blob account key
	# blobAccountKey = "myblobaccountkey"
	## Azure Blob container name
	# blobContainerName = "telegrafcontainer"
	## Flush interval in seconds
	# flushInterval = 300
	## Machine name that is sending the data
	# machineName = "myhostname"
	`
}

func createContainerIfNotExists(ctx context.Context, blobContainerName string) error {
	exists, err := checkIfContainerExists(ctx, blobContainerName)
	if err != nil {
		return err
	}
	if !exists {
		resp, err := blobContainerService.Create(ctx, azblob.Metadata{}, azblob.PublicAccessNone)
		if err != nil {
			return err
		}
		fmt.Printf("Container %s creation status code %d\n", blobContainerName, resp.StatusCode())
	}
	return nil
}

func checkIfContainerExists(ctx context.Context, blobContainerName string) (bool, error) {
	resp, err := blobService.ListContainersSegment(ctx, azblob.Marker{}, azblob.ListContainersSegmentOptions{
		Prefix: blobContainerName,
	})
	if err != nil {
		return false, err
	}
	for _, containerItem := range resp.ContainerItems {
		if containerItem.Name == blobContainerName {
			return true, nil
		}
	}
	return false, nil
}
