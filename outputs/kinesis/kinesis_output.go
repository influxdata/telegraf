package kinesis_output

import (
	"errors"
	"log"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"

	"github.com/influxdb/influxdb/client/v2"
	"github.com/influxdb/telegraf/outputs"
)

type KinesisOutput struct {
	Region       string `toml:"region"`
	StreamName   string `toml:"streamname"`
	PartitionKey string `toml:"partitionkey"`
	Format       string `toml:"format"`
	Debug        bool   `toml:"debug"`
	svc          *kinesis.Kinesis
}

var sampleConfig = `
  # Amazon REGION of kinesis endpoint.
  # Most AWS services have a region specific endpoint this will be used by
  # telegraf to output data.
  # Authentication is provided by an IAMS role, SharedCredentials or Environment
  # Variables.
  region = "ap-southeast-2"
  # Kinesis StreamName must exist prior to starting telegraf.
  streamname = "StreamName"
  # PartitionKey as used for sharding data.
  partitionkey = "PartitionKey"
  # format of the Data payload in the kinesis PutRecord, supported
  # String and Custom.
  format = "string"
  # debug will show upstream aws messages.
  debug = false
`

func (k *KinesisOutput) SampleConfig() string {
	return sampleConfig
}

func (k *KinesisOutput) Description() string {
	return "Configuration for the AWS Kinesis output."
}

func (k *KinesisOutput) Connect() error {
	// We attempt first to create a session to Kinesis using an IAMS role, if that fails it will fall through to using
	// environment variables, and then Shared Credentials.
	if k.Debug {
		log.Printf("Establishing a connection to Kinesis in %+v", k.Region)
	}
	Config := &aws.Config{
		Region: aws.String(k.Region),
		Credentials: credentials.NewChainCredentials(
			[]credentials.Provider{
				&ec2rolecreds.EC2RoleProvider{Client: ec2metadata.New(session.New())},
				&credentials.EnvProvider{},
				&credentials.SharedCredentialsProvider{},
			}),
	}
	svc := kinesis.New(session.New(Config))

	k.svc = svc
	return nil
}

func (k *KinesisOutput) Close() error {
	return errors.New("Error")
}

func formatmetric(k *KinesisOutput, point *client.Point) (string, error) {
	if k.Format == "string" {
		return point.String(), nil
	} else {
		m := fmt.Sprintf("%+v,%+v,%+v %+v",
			point.Name(),
			point.Tags(),
			point.String(),
			point.Time(),
		)
		return m, nil
	}
}

func (k *KinesisOutput) Write(points []*client.Point) error {
	if len(points) == 0 {
		return nil
	}

	r := []*kinesis.PutRecordsRequestEntry{}

	for _, p := range points {
		metric, _ := formatmetric(k, p)
		d := kinesis.PutRecordsRequestEntry{
			Data:         []byte(metric),
			PartitionKey: aws.String(k.PartitionKey),
		}
		r = append(r, &d)
	}

	payload := &kinesis.PutRecordsInput{
		Records:    r,
		StreamName: aws.String(k.StreamName),
	}

	if k.Debug {
		resp, err := k.svc.PutRecords(payload)
		if err != nil {
			log.Printf("Unable to write to Kinesis : %+v \n", err.Error())
		}
		log.Printf("%+v \n", resp)
	} else {
		_, err := k.svc.PutRecords(payload)
		if err != nil {
			log.Printf("Unable to write to Kinesis : %+v \n", err.Error())
		}
	}

	return nil
}

func init() {
	outputs.Add("kinesis_output", func() outputs.Output {
		return &KinesisOutput{}
	})
}
