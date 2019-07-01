package kinesis

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/serializers"
	uuid "github.com/satori/go.uuid"
)

const (
	// MaxOutputRecords is the maximum number of records that we can send in a single send to Kinesis.
	maxOutputRecords = 5
	// maxRecordSizeBytes is the maximum size for a record when sending to Kinesis.
	// 1020KB for the payload and 4KB for the partition key.
	maxRecordSizeBytes = 1020 * 1024
	randomPartitionKey = "-random-"
)

var (
	// gzipCompressionLevel sets the compression level. Tests indicate that 7 gives the best trade off
	// between speed and compression.
	gzipCompressionLevel = 7
)

type slug struct {
	metricsBytes []byte
	size         int
}

type putRecordsHandler struct {
	rawMetrics         map[string][]telegraf.Metric
	slugs              map[string][][]byte
	maxOutputRecords   int
	randomPartitionKey string
	serializer         serializers.Serializer
	readyToSendLock    bool
}

func newPutRecordsHandler() *putRecordsHandler {
	handler := &putRecordsHandler{
		maxOutputRecords:   maxOutputRecords,
		randomPartitionKey: randomPartitionKey,
	}
	handler.init()

	return handler
}

func (handler *putRecordsHandler) init() {
	handler.rawMetrics = make(map[string][]telegraf.Metric)
	handler.slugs = make(map[string][][]byte)
}

func (handler *putRecordsHandler) setSerializer(serializer serializers.Serializer) {
	handler.serializer = serializer
}

func (handler *putRecordsHandler) addMetric(partition string, metric telegraf.Metric) error {
	if handler.readyToSendLock {
		return fmt.Errorf("Already packaged current metrics. Send first then add more")
	}
	if _, ok := handler.rawMetrics[partition]; !ok {
		handler.rawMetrics[partition] = make([]telegraf.Metric, 0)
	}

	handler.rawMetrics[partition] = append(handler.rawMetrics[partition], metric)
	return nil
}

func (handler *putRecordsHandler) addSlugs(partitionKey string, slugs ...[]byte) {
	if _, ok := handler.slugs[partitionKey]; !ok {
		handler.slugs[partitionKey] = make([][]byte, 0)
	}
	// Add each new slug into the current slice of []bytes
	for _, slug := range slugs {
		handler.slugs[partitionKey] = append(handler.slugs[partitionKey], slug)
	}
}

// packageMetrics is responsible to get the metrics split into payloads that are no larger than 1020kb.
// Each partition key will have metrics that need to be split into payloads.
// If the partition key is random then it will create payloads ready to be split between as many shards
// that you have available.
// packageMetrics can't be called again until init is called.
// Really it is designed to be used once and then thrown away.
func (handler *putRecordsHandler) packageMetrics(shards int64) error {
	if handler.readyToSendLock {
		return fmt.Errorf("Waiting to send data, can't accept more metrics currently")
	}

	splitIntoBlocks := func(howManyBlocks int, partitionKey string, metrics []telegraf.Metric) [][]telegraf.Metric {
		blocks := make([][]telegraf.Metric, howManyBlocks)
		for index := range blocks {
			blocks[index] = make([]telegraf.Metric, 0)
		}

		currentBlock := 0
		for _, metric := range metrics {
			blocks[currentBlock] = append(blocks[currentBlock], metric)
			currentBlock++
			if currentBlock == len(blocks) {
				currentBlock = 0
			}
		}

		return blocks
	}

	requiredBlocks := func(currentSize int) int {
		return (currentSize / maxRecordSizeBytes) + 1
	}

	// At this point we need to know if the metrics will fit in a single push to kinesis
	// if not we need to start splitting it.
	// We start with a go for gold dash and bulk serialize.
	// If that doesn't work we will then know how many block we would need.
	// Split again into x blocks, serialize and return.
	for partitionKey, metrics := range handler.rawMetrics {
		if partitionKey == randomPartitionKey {
			// for Random partition keys we need to split the data first then check that it
			// will fit into the payloads. If not we need to split it again, but we know how many
			// blocks to make.

			// Make as many blocks as there is shards
			blocks := int(shards)
			// If we have less metrics than shards, we reduce the block count to 1
			// It will be faster to send one block in this case.
			if int64(len(metrics)) < shards {
				blocks = 1
			}
			safeBlocks := make([][]byte, 0)
			splitBlocks := splitIntoBlocks(blocks, partitionKey, metrics)

			for _, block := range splitBlocks {
				metricsEncoded, err := handler.serializer.SerializeBatch(block)
				if err != nil {
					return err
				}

				blocksNeeded := requiredBlocks(len(metricsEncoded))
				if blocksNeeded == 1 {
					safeBlocks = append(safeBlocks, metricsEncoded)
				} else {
					newBlocks := splitIntoBlocks(blocksNeeded, partitionKey, block)
					for _, newBlock := range newBlocks {
						metricsEncoded, err := handler.serializer.SerializeBatch(newBlock)
						if err != nil {
							return err
						}
						safeBlocks = append(safeBlocks, metricsEncoded)
					}
				}
			}
			handler.slugs[randomPartitionKey] = safeBlocks
			// clear splitBlocks because we don't need it
			splitBlocks = nil

			// Now we need to move the data into its own partition keys
			for _, metricBytes := range handler.slugs[randomPartitionKey] {
				key := uuid.NewV4().String()
				handler.addSlugs(key, metricBytes)
			}
			// We are done now so we need to clear out the random key map value
			delete(handler.slugs, randomPartitionKey)
			continue
		}

		// Try one for static keys
		tryOne, err := handler.serializer.SerializeBatch(metrics)
		if err != nil {
			return err
		}

		// We always need a single block.
		// (len(tryOne) / maxRecordSizeBytes) will give a int due to maxRecordSizeBytes being a const
		// If tryOne is smaller than maxRecordSizeBytes we get zero.
		// or we get how many blocks we need + the starting 1.
		blocksNeeded := requiredBlocks(len(tryOne))

		if blocksNeeded == 1 {
			// The single block is large enough to carry all the metrics.
			handler.addSlugs(partitionKey, tryOne)
			continue
		}

		// We now know how many blocks we need, but need to redistribute the metrics into the blocks
		blocks := splitIntoBlocks(blocksNeeded, partitionKey, metrics)
		for _, metrics := range blocks {
			metricsBytes, err := handler.serializer.SerializeBatch(metrics)
			if err != nil {
				return err
			}
			handler.addSlugs(partitionKey, metricsBytes)
		}
	}

	return nil
}

func newGzipEncoder() (*internal.GzipEncoder, error) {
	// Grab the Gzip encoder directly because we need to set the level.
	gz, err := internal.NewGzipEncoder()
	if err != nil {
		return nil, err
	}
	err = gz.SetLevel(gzipCompressionLevel)
	if err != nil {
		return nil, err
	}

	return gz, nil
}

func (handler *putRecordsHandler) encodeSlugs(encoder internal.ContentEncoder) error {
	for partitionKey, slugs := range handler.slugs {
		for index, slug := range slugs {
			// snappy doesn't return errors
			encodedBytes, err := encoder.Encode(slug)
			if err != nil {
				return err
			}
			handler.slugs[partitionKey][index] = encodedBytes
		}
	}
	return nil
}

// convertToKinesisPutRequests will return a slice that contains a []*kinesis.PutRecordsRequestEntry
// sized to fit into a PutRecords calls. The number of of outer slices is how many times you would
// need to call kinesis.PutRecords.
// The Inner slices adheres to the following rules. No more than 500 records at once and no more than
// 5MB of data including the partition keys.
func (handler *putRecordsHandler) convertToKinesisPutRequests() [][]*kinesis.PutRecordsRequestEntry {
	putRequests := make([][]*kinesis.PutRecordsRequestEntry, 0)
	// We need to seed it with the first one.
	putRequests = append(putRequests, make([]*kinesis.PutRecordsRequestEntry, 0))

	currentIndex := 0
	currentSize := 0
	for partitionKey, metricBytesSlice := range handler.slugs {
		for _, metricBytes := range metricBytesSlice {
			// We need to see if the current data will fit in this put request
			payloadSize := len(partitionKey) + len(metricBytes)
			if currentSize+payloadSize > maxRecordSizeBytes {
				currentIndex++
				putRequests = append(putRequests, make([]*kinesis.PutRecordsRequestEntry, 0))
				currentSize = 0
			}

			currentSize = currentSize + payloadSize

			putRequests[currentIndex] = append(
				putRequests[currentIndex],
				&kinesis.PutRecordsRequestEntry{
					Data:         metricBytes,
					PartitionKey: aws.String(partitionKey),
				},
			)
		}
	}

	return putRequests
}
