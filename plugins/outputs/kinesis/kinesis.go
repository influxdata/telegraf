package kinesis

import (
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/satori/go.uuid"

	"github.com/influxdata/telegraf"
	internalaws "github.com/influxdata/telegraf/internal/config/aws"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
)

type KinesisOutput struct {
	Region    string `toml:"region"`
	AccessKey string `toml:"access_key"`
	SecretKey string `toml:"secret_key"`
	RoleARN   string `toml:"role_arn"`
	Profile   string `toml:"profile"`
	Filename  string `toml:"shared_credential_file"`
	Token     string `toml:"token"`

	StreamName         string `toml:"streamname"`
	PartitionKey       string `toml:"partitionkey"`
	RandomPartitionKey bool   `toml:"use_random_partitionkey"`
	Debug              bool   `toml:"debug"`
	svc                *kinesis.Kinesis

	serializer serializers.Serializer
}

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

  ## Kinesis StreamName must exist prior to starting telegraf.
  streamname = "StreamName"
  ## PartitionKey as used for sharding data.
  partitionkey = "PartitionKey"
  ## If set the paritionKey will be a random UUID on every put.
  ## This allows for scaling across multiple shards in a stream.
  ## This will cause issues with ordering.
  use_random_partitionkey = false


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

func checkstream(l []*string, s string) bool {
	// Check if the StreamName exists in the slice returned from the ListStreams API request.
	for _, stream := range l {
		if *stream == s {
			return true
		}
	}
	return false
}

func (k *KinesisOutput) Connect() error {
	// We attempt first to create a session to Kinesis using an IAMS role, if that fails it will fall through to using
	// environment variables, and then Shared Credentials.
	if k.Debug {
		log.Printf("E! kinesis: Establishing a connection to Kinesis in %+v", k.Region)
	}

	credentialConfig := &internalaws.CredentialConfig{
		Region:    k.Region,
		AccessKey: k.AccessKey,
		SecretKey: k.SecretKey,
		RoleARN:   k.RoleARN,
		Profile:   k.Profile,
		Filename:  k.Filename,
		Token:     k.Token,
	}
	configProvider := credentialConfig.Credentials()
	svc := kinesis.New(configProvider)

	KinesisParams := &kinesis.ListStreamsInput{
		Limit: aws.Int64(100),
	}

	resp, err := svc.ListStreams(KinesisParams)

	if err != nil {
		log.Printf("E! kinesis: Error in ListSteams API call : %+v \n", err)
	}

	if checkstream(resp.StreamNames, k.StreamName) {
		if k.Debug {
			log.Printf("E! kinesis: Stream Exists")
		}
		k.svc = svc
		return nil
	} else {
		log.Printf("E! kinesis : You have configured a StreamName %+v which does not exist. exiting.", k.StreamName)
		os.Exit(1)
	}
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
			log.Printf("E! kinesis: Unable to write to Kinesis : %+v \n", err.Error())
		}
		log.Printf("E! %+v \n", resp)

	} else {
		_, err := k.svc.PutRecords(payload)
		if err != nil {
			log.Printf("E! kinesis: Unable to write to Kinesis : %+v \n", err.Error())
		}
	}
	return time.Since(start)
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

		partitionKey := k.PartitionKey
		if k.RandomPartitionKey {
			u := uuid.NewV4()
			partitionKey = u.String()
		}

		d := kinesis.PutRecordsRequestEntry{
			Data:         values,
			PartitionKey: aws.String(partitionKey),
		}

		r = append(r, &d)

		if sz == 500 {
			// Max Messages Per PutRecordRequest is 500
			elapsed := writekinesis(k, r)
			log.Printf("E! Wrote a %+v point batch to Kinesis in %+v.\n", sz, elapsed)
			sz = 0
			r = nil
		}

	}

	writekinesis(k, r)

	return nil
}

func init() {
	outputs.Add("kinesis", func() telegraf.Output {
		return &KinesisOutput{}
	})
}
