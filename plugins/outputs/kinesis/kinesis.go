package kinesis

import (
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/satori/go.uuid"

	"github.com/influxdata/telegraf"
	internalaws "github.com/influxdata/telegraf/internal/config/aws"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
)

type (
	KinesisOutput struct {
		Region      string `toml:"region"`
		AccessKey   string `toml:"access_key"`
		SecretKey   string `toml:"secret_key"`
		RoleARN     string `toml:"role_arn"`
		Profile     string `toml:"profile"`
		Filename    string `toml:"shared_credential_file"`
		Token       string `toml:"token"`
		EndpointURL string `toml:"endpoint_url"`

		StreamName         string     `toml:"streamname"`
		PartitionKey       string     `toml:"partitionkey"`
		RandomPartitionKey bool       `toml:"use_random_partitionkey"`
		Partition          *Partition `toml:"partition"`
		Debug              bool       `toml:"debug"`
		svc                *kinesis.Kinesis

		serializer serializers.Serializer
	}

	Partition struct {
		Method  string `toml:"method"`
		Key     string `toml:"key"`
		Default string `toml:"default"`
	}
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
  #access_key = ""
  #secret_key = ""
  #token = ""
  #role_arn = ""
  #profile = ""
  #shared_credential_file = ""

  ## Endpoint to make request against, the correct endpoint is automatically
  ## determined and this option should only be set if you wish to override the
  ## default.
  ##   ex: endpoint_url = "http://localhost:8000"
  # endpoint_url = ""

  ## Kinesis StreamName must exist prior to starting telegraf.
  streamname = "StreamName"
  ## DEPRECATED: PartitionKey as used for sharding data.
  partitionkey = "PartitionKey"
  ## DEPRECATED: If set the paritionKey will be a random UUID on every put.
  ## This allows for scaling across multiple shards in a stream.
  ## This will cause issues with ordering.
  use_random_partitionkey = false
  ## The partition key can be calculated using one of several methods:
  ##
  ## Use a static value for all writes:
  #  [outputs.kinesis.partition]
  #    method = "static"
  #    key = "howdy"
  #
  ## Use a random partition key on each write:
  #  [outputs.kinesis.partition]
  #    method = "random"
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


  ## Data format to output.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "influx"

  ## debug will show upstream aws messages.
  debug = false
`

func (k *KinesisOutput) SampleConfig() string {
	return sampleConfig
}

func (k *KinesisOutput) Description() string {
	return "Configuration for the AWS Kinesis output."
}

func (k *KinesisOutput) Connect() error {
	if k.Partition == nil {
		log.Print("E! kinesis : Deprecated paritionkey configuration in use, please consider using outputs.kinesis.partition")
	}

	// We attempt first to create a session to Kinesis using an IAMS role, if that fails it will fall through to using
	// environment variables, and then Shared Credentials.
	if k.Debug {
		log.Printf("I! kinesis: Establishing a connection to Kinesis in %s", k.Region)
	}

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
	svc := kinesis.New(configProvider)

	_, err := svc.DescribeStreamSummary(&kinesis.DescribeStreamSummaryInput{
		StreamName: aws.String(k.StreamName),
	})
	k.svc = svc
	return err
}

func (k *KinesisOutput) Close() error {
	return nil
}

func (k *KinesisOutput) SetSerializer(serializer serializers.Serializer) {
	k.serializer = serializer
}

func writekinesis(k *KinesisOutput, r []*kinesis.PutRecordsRequestEntry) time.Duration {
	start := time.Now()
	payload := &kinesis.PutRecordsInput{
		Records:    r,
		StreamName: aws.String(k.StreamName),
	}

	if k.Debug {
		resp, err := k.svc.PutRecords(payload)
		if err != nil {
			log.Printf("E! kinesis: Unable to write to Kinesis : %s", err.Error())
		}
		log.Printf("I! Wrote: '%+v'", resp)

	} else {
		_, err := k.svc.PutRecords(payload)
		if err != nil {
			log.Printf("E! kinesis: Unable to write to Kinesis : %s", err.Error())
		}
	}
	return time.Since(start)
}

func (k *KinesisOutput) getPartitionKey(metric telegraf.Metric) string {
	if k.Partition != nil {
		switch k.Partition.Method {
		case "static":
			return k.Partition.Key
		case "random":
			u := uuid.NewV4()
			return u.String()
		case "measurement":
			return metric.Name()
		case "tag":
			if t, ok := metric.GetTag(k.Partition.Key); ok {
				return t
			} else if len(k.Partition.Default) > 0 {
				return k.Partition.Default
			}
			// Default partition name if default is not set
			return "telegraf"
		default:
			log.Printf("E! kinesis : You have configured a Partition method of '%s' which is not supported", k.Partition.Method)
		}
	}
	if k.RandomPartitionKey {
		u := uuid.NewV4()
		return u.String()
	}
	return k.PartitionKey
}

func (k *KinesisOutput) Write(metrics []telegraf.Metric) error {
	var sz uint32

	if len(metrics) == 0 {
		return nil
	}

	r := []*kinesis.PutRecordsRequestEntry{}

	for _, metric := range metrics {
		sz++

		values, err := k.serializer.Serialize(metric)
		if err != nil {
			return err
		}

		partitionKey := k.getPartitionKey(metric)

		d := kinesis.PutRecordsRequestEntry{
			Data:         values,
			PartitionKey: aws.String(partitionKey),
		}

		r = append(r, &d)

		if sz == 500 {
			// Max Messages Per PutRecordRequest is 500
			elapsed := writekinesis(k, r)
			log.Printf("D! Wrote a %d point batch to Kinesis in %+v.", sz, elapsed)
			sz = 0
			r = nil
		}

	}
	if sz > 0 {
		elapsed := writekinesis(k, r)
		log.Printf("D! Wrote a %d point batch to Kinesis in %+v.", sz, elapsed)
	}

	return nil
}

func init() {
	outputs.Add("kinesis", func() telegraf.Output {
		return &KinesisOutput{}
	})
}
