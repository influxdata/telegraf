package kinesis

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kinesis"
	"github.com/aws/aws-sdk-go-v2/service/kinesis/types"
	"github.com/gofrs/uuid"
	"github.com/influxdata/telegraf"
	internalaws "github.com/influxdata/telegraf/config/aws"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
)

// Limit set by AWS (https://docs.aws.amazon.com/kinesis/latest/APIReference/API_PutRecords.html)
const maxRecordsPerRequest uint32 = 500

type (
	KinesisOutput struct {
		StreamName         string     `toml:"streamname"`
		PartitionKey       string     `toml:"partitionkey"`
		RandomPartitionKey bool       `toml:"use_random_partitionkey"`
		Partition          *Partition `toml:"partition"`
		Debug              bool       `toml:"debug"`

		Log        telegraf.Logger `toml:"-"`
		serializer serializers.Serializer
		svc        kinesisClient

		internalaws.CredentialConfig
	}

	Partition struct {
		Method  string `toml:"method"`
		Key     string `toml:"key"`
		Default string `toml:"default"`
	}
)

type kinesisClient interface {
	PutRecords(context.Context, *kinesis.PutRecordsInput, ...func(*kinesis.Options)) (*kinesis.PutRecordsOutput, error)
}

var sampleConfig = `
  ## Amazon REGION of kinesis endpoint.
  region = "ap-southeast-2"

  ## Amazon Credentials
  ## Credentials are loaded in the following order
  ## 1) Web identity provider credentials via STS if role_arn and web_identity_token_file are specified
  ## 2) Assumed credentials via STS if role_arn is specified
  ## 3) explicit credentials from 'access_key' and 'secret_key'
  ## 4) shared profile from 'profile'
  ## 5) environment variables
  ## 6) shared credentials file
  ## 7) EC2 Instance Profile
  #access_key = ""
  #secret_key = ""
  #token = ""
  #role_arn = ""
  #web_identity_token_file = ""
  #role_session_name = ""
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
  ## DEPRECATED: If set the partitionKey will be a random UUID on every put.
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
		k.Log.Error("Deprecated partitionkey configuration in use, please consider using outputs.kinesis.partition")
	}

	// We attempt first to create a session to Kinesis using an IAMS role, if that fails it will fall through to using
	// environment variables, and then Shared Credentials.
	if k.Debug {
		k.Log.Infof("Establishing a connection to Kinesis in %s", k.Region)
	}

	cfg, err := k.CredentialConfig.Credentials()
	if err != nil {
		return err
	}

	svc := kinesis.NewFromConfig(cfg)

	_, err = svc.DescribeStreamSummary(context.Background(), &kinesis.DescribeStreamSummaryInput{
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

func (k *KinesisOutput) writeKinesis(r []types.PutRecordsRequestEntry) time.Duration {
	start := time.Now()
	payload := &kinesis.PutRecordsInput{
		Records:    r,
		StreamName: aws.String(k.StreamName),
	}

	resp, err := k.svc.PutRecords(context.Background(), payload)
	if err != nil {
		k.Log.Errorf("Unable to write to Kinesis : %s", err.Error())
		return time.Since(start)
	}

	if k.Debug {
		k.Log.Infof("Wrote: '%+v'", resp)
	}

	failed := *resp.FailedRecordCount
	if failed > 0 {
		k.Log.Errorf("Unable to write %+v of %+v record(s) to Kinesis", failed, len(r))
	}

	return time.Since(start)
}

func (k *KinesisOutput) getPartitionKey(metric telegraf.Metric) string {
	if k.Partition != nil {
		switch k.Partition.Method {
		case "static":
			return k.Partition.Key
		case "random":
			u, err := uuid.NewV4()
			if err != nil {
				return k.Partition.Default
			}
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
			k.Log.Errorf("You have configured a Partition method of '%s' which is not supported", k.Partition.Method)
		}
	}
	if k.RandomPartitionKey {
		u, err := uuid.NewV4()
		if err != nil {
			return k.Partition.Default
		}
		return u.String()
	}
	return k.PartitionKey
}

func (k *KinesisOutput) Write(metrics []telegraf.Metric) error {
	var sz uint32

	if len(metrics) == 0 {
		return nil
	}

	r := []types.PutRecordsRequestEntry{}

	for _, metric := range metrics {
		sz++

		values, err := k.serializer.Serialize(metric)
		if err != nil {
			k.Log.Debugf("Could not serialize metric: %v", err)
			continue
		}

		partitionKey := k.getPartitionKey(metric)

		d := types.PutRecordsRequestEntry{
			Data:         values,
			PartitionKey: aws.String(partitionKey),
		}

		r = append(r, d)

		if sz == maxRecordsPerRequest {
			elapsed := k.writeKinesis(r)
			k.Log.Debugf("Wrote a %d point batch to Kinesis in %+v.", sz, elapsed)
			sz = 0
			r = nil
		}
	}
	if sz > 0 {
		elapsed := k.writeKinesis(r)
		k.Log.Debugf("Wrote a %d point batch to Kinesis in %+v.", sz, elapsed)
	}

	return nil
}

func init() {
	outputs.Add("kinesis", func() telegraf.Output {
		return &KinesisOutput{}
	})
}
