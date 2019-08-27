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
	// The maximum size of a data blob (the data payload before Base64-encoding) within one record is 1 megabyte (MB).
	// The data blob includes the partition key
	// 1020KB for the payload and 4KB for the partition key.
	maxRecordSizeBytes = (1024 * 1024) - 4096
	randomPartitionKey = "-random-"
	// gzipCompressionLevel sets the compression level. Tests indicate that 7 gives the best trade off
	// between speed and compression.
	gzipCompressionLevel = 7
)

// splitIntoPayloads will take a []telegraf.Metric and spread them as evenly as possible into
// as many []telegraf.Metrics as specified by 'howManyPayloads'.
// It will then return these buckets of metrics in a [][]telegraf.Metrics
func splitIntoPayloads(howManyPayloads int, metrics []telegraf.Metric) [][]telegraf.Metric {
	payloads := make([][]telegraf.Metric, howManyPayloads)
	for index := range payloads {
		payloads[index] = make([]telegraf.Metric, 0)
	}

	currentBlock := 0
	for _, metric := range metrics {
		payloads[currentBlock] = append(payloads[currentBlock], metric)
		currentBlock++
		if currentBlock == len(payloads) {
			currentBlock = 0
		}
	}

	return payloads
}

// requiredPayloads will take a int and see how many payloads you would need to fit in into
// a kinesis request.
func requiredPayloads(blobSize int) int {
	return (blobSize / maxRecordSizeBytes) + 1
}

type putRecordsHandler struct {
	rawMetrics         map[string][]telegraf.Metric
	payloads           map[string][][]byte
	maxOutputRecords   int
	randomPartitionKey string
	serializer         serializers.Serializer
	readyToSendLock    bool
}

func newPutRecordsHandler(serializer serializers.Serializer) *putRecordsHandler {
	return &putRecordsHandler{
		maxOutputRecords:   maxOutputRecords,
		randomPartitionKey: randomPartitionKey,
		serializer:         serializer,
		rawMetrics:         make(map[string][]telegraf.Metric),
		payloads:           make(map[string][][]byte),
	}
}

func (handler *putRecordsHandler) addRawMetric(partition string, metric telegraf.Metric) error {
	if handler.readyToSendLock {
		return fmt.Errorf("Already packaged current metrics. Send first then add more")
	}

	if _, ok := handler.rawMetrics[partition]; !ok {
		handler.rawMetrics[partition] = make([]telegraf.Metric, 0)
	}

	handler.rawMetrics[partition] = append(handler.rawMetrics[partition], metric)
	return nil
}

func (handler *putRecordsHandler) addPayload(partitionKey string, payloadBody ...[]byte) {
	if _, ok := handler.payloads[partitionKey]; !ok {
		handler.payloads[partitionKey] = make([][]byte, 0)
	}
	// Add each new slug into the current slice of []bytes
	for _, body := range payloadBody {
		handler.payloads[partitionKey] = append(handler.payloads[partitionKey], body)
	}
}

// packageMetrics is responsible for splitting metrics into payloads that are no larger than 1020kb.
// Each partition key will have metrics that need to be split into payloads.
// If the partition key is random then it will create payloads ready to be split between as many shards
// that you have available.
// packageMetrics can only be called once.
func (handler *putRecordsHandler) packageMetrics(shards int64) error {
	if handler.readyToSendLock {
		return fmt.Errorf("Waiting to send data, can't accept more metrics currently")
	}

	// We need to know if the metrics will fit in a single put request body for kinesis if not we need to split them.
	// We start with a go for gold dash and bulk serialize. We then measure the size and see if it will fit.
	// If that doesn't fit we can now use the serialized data to measure how many bytes we have and how many payloads we would need.
	// We then split into x payloads, serialize the smaller batches and put these into the converted payloads.
	// Payloads need to be sent using the correct partition keys. So we need to make sure we resect the partition key.
	for partitionKey, metrics := range handler.rawMetrics {
		switch partitionKey {
		case randomPartitionKey:
			// for Random partition keys we need to split the data first then check that it
			// will fit into the payloads. If not we need to split it again, but we know how many
			// payloads to make.

			// Make as many payloads as there is shards
			shardCount := int(shards)
			// If we have less metrics than shards, we reduce the payload count to 1
			// It will be faster to send one payload in this case.
			if int64(len(metrics)) < shards {
				shardCount = 1
			}
			safePayloads := make([][]byte, 0)

			shardPayloads := splitIntoPayloads(shardCount, metrics)

			for _, payload := range shardPayloads {
				metricsEncoded, err := handler.serializer.SerializeBatch(payload)
				if err != nil {
					return err
				}

				payloadsNeeded := requiredPayloads(len(metricsEncoded))
				if payloadsNeeded == 1 {
					safePayloads = append(safePayloads, metricsEncoded)
				} else {
					newBlocks := splitIntoPayloads(payloadsNeeded, payload)
					for _, newBlock := range newBlocks {
						metricsEncoded, err := handler.serializer.SerializeBatch(newBlock)
						if err != nil {
							return err
						}
						safePayloads = append(safePayloads, metricsEncoded)
					}
				}
			}
			handler.payloads[randomPartitionKey] = safePayloads

			// Now we need to move the data into its own partition keys
			for _, metricBytes := range handler.payloads[randomPartitionKey] {
				handler.addPayload(uuid.NewV4().String(), metricBytes)
			}

			// clean up to release memory.
			// We are done now so we need to clear out the random key map value
			delete(handler.payloads, randomPartitionKey)
			// clear splitPayloads because we don't need it anymore
			shardPayloads = nil
			continue
		default:
			// Try one for static keys
			EncodedMetrics, err := handler.serializer.SerializeBatch(metrics)
			if err != nil {
				return err
			}

			PayloadsNeeded := requiredPayloads(len(EncodedMetrics))

			if PayloadsNeeded == 1 {
				// The single block is large enough to carry all the metrics.
				handler.addPayload(partitionKey, EncodedMetrics)
				continue
			}

			// We need to split the metrics into payloads that will fit.
			// But we know how many we need now.
			for _, metrics := range splitIntoPayloads(PayloadsNeeded, metrics) {
				EncodedMetrics, err := handler.serializer.SerializeBatch(metrics)
				if err != nil {
					return err
				}
				handler.addPayload(partitionKey, EncodedMetrics)
			}
		}
		// clean up the data in the handler to stop it from storing double the data
		delete(handler.rawMetrics, partitionKey)
	}

	return nil
}

func (handler *putRecordsHandler) encodePayloadBodies(encoder internal.ContentEncoder) error {
	for partitionKey, payloads := range handler.payloads {
		for index, payload := range payloads {
			encodedBytes, err := encoder.Encode(payload)
			if err != nil {
				return err
			}
			handler.payloads[partitionKey][index] = encodedBytes
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
	for partitionKey, metricBytesSlice := range handler.payloads {
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
