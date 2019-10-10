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

	serializer           serializers.Serializer
	metricsCache         []telegraf.Metric
	timeOfPreviousFlush  time.Time
	blobService          azblob.ServiceURL
	blobContainerService azblob.ContainerURL
}

// Connect to the Output
func (s *AzureBlob) Connect() error {
	fmt.Printf("Initializing Azure Blob output plugin for BlobAccount %s, BlobContainerName %s, MachineName %s and FlushInterval %d seconds\n",
		s.BlobAccount, s.BlobContainerName, s.MachineName, s.FlushInterval)
	// authenticate and create a pipeline
	c, err := azblob.NewSharedKeyCredential(s.BlobAccount, s.BlobAccountKey)
	if err != nil {
		return err
	}
	p := azblob.NewPipeline(c, azblob.PipelineOptions{})
	u, err := url.Parse(fmt.Sprintf(blobFormatString, s.BlobAccount))
	if err != nil {
		return err
	}
	// create Azure Blob services
	s.blobService = azblob.NewServiceURL(*u, p)
	s.blobContainerService = s.blobService.NewContainerURL(s.BlobContainerName)

	err = s.createContainerIfNotExists(context.TODO())
	if err != nil {
		return err
	}

	// setting time to an initial value
	s.timeOfPreviousFlush = time.Now()
	// setting hostname to an initial value
	if s.MachineName == "" {
		hostname, err := os.Hostname()
		if err != nil {
			return err
		}
		s.MachineName = hostname
	}
	// setting format to an initial value
	return nil
}

func (s *AzureBlob) SetSerializer(serializer serializers.Serializer) {
	s.serializer = serializer
}

// Close any connections to the Output
func (s *AzureBlob) Close() error {
	fmt.Println("Close message received. Synchronously flushing saved events")
	err := s.flushMetricsCacheToAzureBlob()
	return err
}

// Write takes in group of points to be written to the Output
func (s *AzureBlob) Write(metrics []telegraf.Metric) error {
	fmt.Printf("%d metrics cached\n", len(metrics))

	s.metricsCache = append(s.metricsCache, metrics...)

	// check if it is time to flush all cached metrics
	if time.Now().Sub(s.timeOfPreviousFlush) < time.Duration(s.FlushInterval)*time.Second {
		return nil
	}

	err := s.flushMetricsCacheToAzureBlob()
	s.timeOfPreviousFlush = time.Now()
	if err != nil {
		return err
	}
	// we'll empty the cache only if the createBlob operation succeeded
	s.metricsCache = nil
	return nil
}

func (s *AzureBlob) flushMetricsCacheToAzureBlob() error {
	var sb strings.Builder
	for _, metric := range s.metricsCache {
		bytes, err := s.serializer.Serialize(metric)
		if err != nil {
			return fmt.Errorf("cannot serialize because of %s", err)
		}
		sb.Write(bytes)
	}
	//bytes, err := s.serializer.SerializeBatch(s.metricsCache)
	bytes := []byte(sb.String())
	if len(s.metricsCache) == 0 {
		fmt.Println("0 items in cache - will not write anything")
		return nil
	}

	// format is endDateTime-startDateTime-machineName
	blobName := fmt.Sprintf("%s-%s-%s", s.metricsCache[len(s.metricsCache)-1].Time().Format(timeFormatString),
		s.metricsCache[0].Time().Format(timeFormatString), s.MachineName)

	_, err := s.createBlockBlob(context.TODO(), blobName, bytes)
	if err != nil {
		return fmt.Errorf("error creating blob: %s", err)
	}

	fmt.Printf("blob '%s' written successfully\n", blobName)
	return nil

}

func init() {
	outputs.Add("azure_blob", func() telegraf.Output { return &AzureBlob{} })
}

func (s *AzureBlob) getBlockBlobURL(ctx context.Context, blobName string) azblob.BlockBlobURL {
	blob := s.blobContainerService.NewBlockBlobURL(blobName)
	return blob
}

func (s *AzureBlob) createBlockBlob(ctx context.Context, blobName string, byteString []byte) (azblob.BlockBlobURL, error) {
	b := s.getBlockBlobURL(ctx, blobName+".zip")                              // .zip is the file that will be created
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
	## Data format to output.
	## Each data format has its own unique set of configuration options, read
	## more about them here:
	## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
	data_format = "json"
	`
}

func (s *AzureBlob) createContainerIfNotExists(ctx context.Context) error {
	exists, err := s.checkIfContainerExists(ctx)
	if err != nil {
		return err
	}
	if !exists {
		resp, err := s.blobContainerService.Create(ctx, azblob.Metadata{}, azblob.PublicAccessNone)
		if err != nil {
			return err
		}
		fmt.Printf("Container %s creation status code %d\n", s.BlobContainerName, resp.StatusCode())
	}
	return nil
}

func (s *AzureBlob) checkIfContainerExists(ctx context.Context) (bool, error) {
	resp, err := s.blobService.ListContainersSegment(ctx, azblob.Marker{}, azblob.ListContainersSegmentOptions{
		Prefix: s.BlobContainerName,
	})
	if err != nil {
		return false, err
	}
	for _, containerItem := range resp.ContainerItems {
		if containerItem.Name == s.BlobContainerName {
			return true, nil
		}
	}
	return false, nil
}
