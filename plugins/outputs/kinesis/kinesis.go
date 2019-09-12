package kinesis

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	internalaws "github.com/influxdata/telegraf/internal/config/aws"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
	"github.com/influxdata/telegraf/selfstat"
)

const (
	putUnitBytes = 25 * 1024
)

var sampleConfig = `
  ## Amazon REGION of kinesis endpoint.
  region = "ap-southeast-2"

  ## Amazon Credentials
  ## Credentials are loaded in the following order
  ## 1) Assumed credentials via STS if role_arn is specified
  ## 2) explicit credentials from 'access_key' and 'secret_key'
  ## 3) shared profile from 'profile'
  ## 4) environment variables
  ## 5) shared credentials file
  ## 6) EC2 Instance Profile
  # access_key = ""
  # secret_key = ""
  # token = ""
  # role_arn = ""
  # profile = ""
  # shared_credential_file = ""

  ## Endpoint to make request against, the correct endpoint is automatically
  ## determined and this option should only be set if you wish to override the
  ## default.
  ##   ex: endpoint_url = "http://localhost:8000"
  # endpoint_url = ""

  ## Kinesis StreamName must exist prior to starting telegraf.
  streamname = "StreamName"

  ## DEPRECATED: PartitionKey as used for sharding data.
  # partitionkey = "PartitionKey"
  ## DEPRECATED: If set the paritionKey will be a random UUID on every put.
  ## This allows for scaling across multiple shards in a stream.
  ## This will cause issues with ordering.
  # use_random_partitionkey = false

  ## Write multiple metrics per record in batch serialization format.  Enabling
  ## this option is recommended for the best throughput and lowest cost.
  # use_batch_format = false

  ## Content encoding for message payloads, can be set to "gzip" to or
  ## "identity" to apply no encoding.
  ##
  ## As compression is performed per Kinesis record, it is recommended to use
  ## compression with "use_batch_format = true".
  # content_encoding = "identity"

  ## Data format to output.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "influx"

  ## The partition key can be calculated using one of several methods:
  ##
  ## Use a static value for all writes:
  [outputs.kinesis.partition]
    method = "static"
    key = "PartitionKey"

  ## Use a random partition key on each write:
  #  [outputs.kinesis.partition]
  #    method = "random"
  #    max_partitions = 5
  #
  ## Use the measurement name as the partition key:
  #  [outputs.kinesis.partition]
  #    method = "measurement"
  #
  ## Use the value of a tag for all writes, if the tag is not set the empty
  ## default option will be used. When no default, defaults to "telegraf"
  #  [outputs.kinesis.partition]
  #    method = "tag"
  #    key = "host"
  #    default = "mykey"
`

const (
	// Maximum number of records that can be sent in a single call to
	// PutRecords.
	maxPutRecords = 500
)

type Partitioner interface {
	// Partition partitions metrics and returns an array where each item should
	// be written in the same record.
	Partition(metrics []telegraf.Metric) []Partition
}

type Client interface {
	DescribeStreamSummary(input *kinesis.DescribeStreamSummaryInput) (*kinesis.DescribeStreamSummaryOutput, error)
	PutRecords(input *kinesis.PutRecordsInput) (*kinesis.PutRecordsOutput, error)
}

type Kinesis struct {
	Region             string           `toml:"region"`
	AccessKey          string           `toml:"access_key"`
	SecretKey          string           `toml:"secret_key"`
	RoleARN            string           `toml:"role_arn"`
	Profile            string           `toml:"profile"`
	Filename           string           `toml:"shared_credential_file"`
	Token              string           `toml:"token"`
	EndpointURL        string           `toml:"endpoint_url"`
	StreamName         string           `toml:"streamname"`
	Partition          *PartitionConfig `toml:"partition"`
	UseBatchFormat     bool             `toml:"use_batch_format"`
	ContentEncoding    string           `toml:"content_encoding"`
	PartitionKey       string           `toml:"partitionkey"`            // deprecated in 1.5; use partition
	RandomPartitionKey bool             `toml:"use_random_partitionkey"` // deprecated in 1.5; use partition
	Debug              bool             `toml:"debug"`                   // deprecated in 1.13;

	Log         telegraf.Logger `toml:"-"`
	client      Client
	partitioner Partitioner
	serializer  serializers.Serializer
	encoder     internal.ContentEncoder

	putBytes        selfstat.Stat
	putPayloadUnits selfstat.Stat
	putRecords      selfstat.Stat
}

type PartitionConfig struct {
	Method  string `toml:"method"`
	Key     string `toml:"key"`
	Default string `toml:"default"`
}

func (k *Kinesis) SampleConfig() string {
	return sampleConfig
}

func (k *Kinesis) Description() string {
	return "Write metrics to a Kinesis data stream"
}

func (k *Kinesis) Init() error {
	// Handle deprecated partition options
	if k.Partition == nil {
		if k.RandomPartitionKey {
			k.Log.Warn("The use_random_partitionkey is deprecated, use the partition table as a replacement")
			k.Partition = &PartitionConfig{
				Method: "random",
			}
		} else {
			k.Log.Warn("The partitionkey is deprecated, use the partition table as a replacement")
			k.Partition = &PartitionConfig{
				Method: "static",
				Key:    k.PartitionKey,
			}
		}
	}

	// Metrics are partitioned based on the partition key method and if
	// use_batch_format is enabled.
	if k.UseBatchFormat {
		switch k.Partition.Method {
		case "static", "measurement", "tag":
			k.partitioner = &FixedBatchPartitioner{k.Partition}
		case "random":
			k.partitioner = &RandomBatchPartitioner{k.Partition}
		default:
			return fmt.Errorf("unsupported partition method")
		}
	} else {
		k.partitioner = &SingleRecordPartitioner{k.Partition}
	}

	var err error
	k.encoder, err = internal.NewContentEncoder(k.ContentEncoding)
	if err != nil {
		return err
	}

	tags := map[string]string{}
	k.putBytes = selfstat.Register("kinesis", "put_bytes", tags)
	k.putPayloadUnits = selfstat.Register("kinesis", "put_payload_units", tags)
	k.putRecords = selfstat.Register("kinesis", "put_records", tags)

	return nil
}

func (k *Kinesis) Connect() error {
	credentialConfig := &internalaws.CredentialConfig{
		Region:      k.Region,
		AccessKey:   k.AccessKey,
		SecretKey:   k.SecretKey,
		RoleARN:     k.RoleARN,
		Profile:     k.Profile,
		Filename:    k.Filename,
		Token:       k.Token,
		EndpointURL: k.EndpointURL,
	}
	configProvider := credentialConfig.Credentials()
	k.client = kinesis.New(configProvider)

	_, err := k.client.DescribeStreamSummary(&kinesis.DescribeStreamSummaryInput{
		StreamName: aws.String(k.StreamName),
	})
	return err
}

func (k *Kinesis) Close() error {
	return nil
}

func (k *Kinesis) SetSerializer(serializer serializers.Serializer) {
	k.serializer = serializer
}

func (k *Kinesis) Write(metrics []telegraf.Metric) error {
	partitions := k.partitioner.Partition(metrics)

	var records []*kinesis.PutRecordsRequestEntry
	for _, part := range partitions {
		octets, err := k.serialize(part.Metrics)
		if err != nil {
			return err
		}

		octets, err = k.encoder.Encode(octets)
		if err != nil {
			return err
		}

		entry := &kinesis.PutRecordsRequestEntry{
			Data:         octets,
			PartitionKey: aws.String(part.Key),
		}

		k.updatePutStats(int64(len(octets)))

		records = append(records, entry)
		if len(records) == maxPutRecords {
			err := k.writeRecords(records)
			if err != nil {
				return fmt.Errorf("error sending put record: %v", err)
			}
			records = records[:0]
		}
	}

	if len(records) != 0 {
		err := k.writeRecords(records)
		if err != nil {
			return fmt.Errorf("error sending put record: %v", err)
		}
	}

	return nil
}

func (k *Kinesis) writeRecords(records []*kinesis.PutRecordsRequestEntry) error {
	payload := &kinesis.PutRecordsInput{
		Records:    records,
		StreamName: aws.String(k.StreamName),
	}
	_, err := k.client.PutRecords(payload)
	return err
}

func (k *Kinesis) serialize(metrics []telegraf.Metric) ([]byte, error) {
	if k.UseBatchFormat {
		return k.serializer.SerializeBatch(metrics)
	} else {
		return k.serializer.Serialize(metrics[0])
	}
}

func (k *Kinesis) updatePutStats(bytes int64) {
	k.putBytes.Incr(bytes)
	units := mround(bytes, putUnitBytes) / putUnitBytes
	k.putPayloadUnits.Incr(units)
	k.putRecords.Incr(1)
}

// mround rounds up to the next multiple of a value
func mround(value, multiple int64) int64 {
	if multiple == 0 {
		return value
	}

	remainder := value % multiple
	if remainder == 0 {
		return value
	}

	return value + multiple - remainder
}

func init() {
	outputs.Add("kinesis", func() telegraf.Output {
		return &Kinesis{}
	})
}
