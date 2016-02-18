package kinesis

import (
	"fmt"
	"log"
	"os"
	"sync/atomic"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
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
  ## Amazon REGION of kinesis endpoint.
  region = "ap-southeast-2"
  ## Kinesis StreamName must exist prior to starting telegraf.
  streamname = "StreamName"
  ## PartitionKey as used for sharding data.
  partitionkey = "PartitionKey"
  ## format of the Data payload in the kinesis PutRecord, supported
  ## String and Custom.
  format = "string"
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
		log.Printf("kinesis: Establishing a connection to Kinesis in %+v", k.Region)
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

	KinesisParams := &kinesis.ListStreamsInput{
		Limit: aws.Int64(100),
	}

	resp, err := svc.ListStreams(KinesisParams)

	if err != nil {
		log.Printf("kinesis: Error in ListSteams API call : %+v \n", err)
	}

	if checkstream(resp.StreamNames, k.StreamName) {
		if k.Debug {
			log.Printf("kinesis: Stream Exists")
		}
		k.svc = svc
		return nil
	} else {
		log.Printf("kinesis : You have configured a StreamName %+v which does not exist. exiting.", k.StreamName)
		os.Exit(1)
	}
	return err
}

func (k *KinesisOutput) Close() error {
	return nil
}

func FormatMetric(k *KinesisOutput, point telegraf.Metric) (string, error) {
	if k.Format == "string" {
		return point.String(), nil
	} else {
		m := fmt.Sprintf("%+v,%+v,%+v",
			point.Name(),
			point.Tags(),
			point.String())
		return m, nil
	}
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
			log.Printf("kinesis: Unable to write to Kinesis : %+v \n", err.Error())
		}
		log.Printf("%+v \n", resp)

	} else {
		_, err := k.svc.PutRecords(payload)
		if err != nil {
			log.Printf("kinesis: Unable to write to Kinesis : %+v \n", err.Error())
		}
	}
	return time.Since(start)
}

func (k *KinesisOutput) Write(metrics []telegraf.Metric) error {
	var sz uint32 = 0

	if len(metrics) == 0 {
		return nil
	}

	r := []*kinesis.PutRecordsRequestEntry{}

	for _, p := range metrics {
		atomic.AddUint32(&sz, 1)

		metric, _ := FormatMetric(k, p)
		d := kinesis.PutRecordsRequestEntry{
			Data:         []byte(metric),
			PartitionKey: aws.String(k.PartitionKey),
		}
		r = append(r, &d)

		if sz == 500 {
			// Max Messages Per PutRecordRequest is 500
			elapsed := writekinesis(k, r)
			log.Printf("Wrote a %+v point batch to Kinesis in %+v.\n", sz, elapsed)
			atomic.StoreUint32(&sz, 0)
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
