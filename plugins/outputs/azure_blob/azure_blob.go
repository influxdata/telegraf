package azure_blob

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"

	"github.com/Azure/azure-storage-blob-go/azblob"
)

var metricsMemoryCache []telegraf.Metric
var timeOfPreviousFlush time.Time = time.Now()

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
	fmt.Println("AzureBlob Connect")
	return nil
}

// Close any connections to the Output
func (s *AzureBlob) Close() error {
	fmt.Println("Close")
	return nil
}

// Description returns a one-sentence description on the Output
func (s *AzureBlob) Description() string {
	fmt.Println("Description")
	return "Azure Blob"
}

// SampleConfig returns the default configuration of the Output
func (s *AzureBlob) SampleConfig() string {
	fmt.Println("SampleConfig")
	return "sample config"
}

// Write takes in group of points to be written to the Output
func (s *AzureBlob) Write(metrics []telegraf.Metric) error {
	fmt.Printf("Write %d\n", len(metrics))

	metricsMemoryCache = append(metricsMemoryCache, metrics...)

	if time.Now().Sub(timeOfPreviousFlush) < time.Duration(s.FlushInterval)*time.Second {
		return nil
	}

	tempCache := make([]telegraf.Metric, len(metricsMemoryCache))
	copy(tempCache, metricsMemoryCache)
	go s.flushMetricsMemoryCacheToAzureBlob(tempCache)
	metricsMemoryCache = nil

	timeOfPreviousFlush = time.Now()

	return nil
}

func (s *AzureBlob) flushMetricsMemoryCacheToAzureBlob(cache []telegraf.Metric) {
	fmt.Printf("Flush %d\n", len(cache))

	// TODO: check if blob container exists

	var str strings.Builder
	for _, metric := range cache {
		_, err := str.WriteString(fmt.Sprintf("%v\n", metric))
		if err != nil {
			fmt.Printf("error creating string in flush: %s", err)
		}
	}

	if len(cache) == 0 {
		fmt.Println("0 items in cache - will not write anything")
		return
	}

	// format is endDate-startDate-machineName.txt
	blobName := fmt.Sprintf("%s-%s-%s.txt", cache[len(cache)-1].Time().Format(timeFormatString),
		cache[0].Time().Format(timeFormatString), s.MachineName)

	_, err := s.createBlockBlob(context.TODO(), blobName, str.String())
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("Blob named %s written successfully\n", blobName)
	}
}

func init() {
	outputs.Add("azure_blob", func() telegraf.Output { return &AzureBlob{} })
}

func (s *AzureBlob) getBlockBlobURL(ctx context.Context, blobName string) azblob.BlockBlobURL {
	container := s.getContainerURL(ctx)
	blob := container.NewBlockBlobURL(blobName)
	return blob
}

func (s *AzureBlob) createBlockBlob(ctx context.Context, blobName string, data string) (azblob.BlockBlobURL, error) {
	b := s.getBlockBlobURL(ctx, blobName)

	_, err := b.Upload(
		ctx,
		strings.NewReader(data),
		azblob.BlobHTTPHeaders{
			ContentType: "text/plain",
		},
		azblob.Metadata{},
		azblob.BlobAccessConditions{},
	)

	return b, err
}

func (s *AzureBlob) getContainerURL(ctx context.Context) azblob.ContainerURL {
	c, _ := azblob.NewSharedKeyCredential(s.BlobAccount, s.BlobAccountKey)
	p := azblob.NewPipeline(c, azblob.PipelineOptions{})
	u, _ := url.Parse(fmt.Sprintf(blobFormatString, s.BlobAccount))
	service := azblob.NewServiceURL(*u, p)
	container := service.NewContainerURL(s.BlobContainerName)
	return container
}
