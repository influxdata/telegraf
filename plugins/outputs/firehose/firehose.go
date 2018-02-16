// Package firehose provides an output plugin for telegraf to write to Kinesis Firehose
package firehose

import (
	"bytes"
	"compress/gzip"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/service/firehose"
	"github.com/aws/aws-sdk-go/service/firehose/firehoseiface"
	"github.com/influxdata/telegraf"
	internalaws "github.com/influxdata/telegraf/internal/config/aws"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
)

type (
	errorEntry struct {
		submitAttemptCount int64
		batch              [][]byte // list of serialized metric strings
	}
)

type (
	FirehoseOutput struct {
		Region    string `toml:"region"`
		AccessKey string `toml:"access_key"`
		SecretKey string `toml:"secret_key"`
		RoleARN   string `toml:"role_arn"`
		Profile   string `toml:"profile"`
		Filename  string `toml:"shared_credential_file"`
		Token     string `toml:"token"`

		DeliveryStreamName    string `toml:"delivery_stream_name"`
		Debug                 bool   `toml:"debug"`
		MaxSubmitAttempts     int64  `toml:"max_submit_attempts"`
		EnableGzipCompression bool   `toml:"enable_gzip_compression"`

		svc             firehoseiface.FirehoseAPI
		errorBuffer     []*errorEntry
		totalErrorCount int64

		serializer serializers.Serializer
	}
)

var sampleConfig = `
  ## Output to AWS Firehose.

  ## NOTE: Keep in mind that firehose has a max batch size of 1MB.
  ## Please set telegraf's metric_batch_size and/or flush_interval appropriately so that
  ## this plugin does not try to push more than 1MB of metrics at a time.

  ## Amazon REGION of the AWS firehose endpoint.
  region = "us-east-2"

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

  ## Firehose StreamName must exist prior to starting telegraf.
  delivery_stream_name = "FirehoseName"

  ## The maximum number of times to attempt resubmitting a metric batch
  max_submit_attempts = 10

  ## Whether or not to enable gzip compression. Compression rates are usually
  ## as good as 90% space savings. However, the destination must decompress the data 
  ## before it goes to influx.
  ## The good news here, gzip it designed to be concat'ed together. We will compress 
  ## all metrics received during a telegrate write and send them to firehose. Firehose will
  ## then concat all this data together- but a single decompress will get all lines back in
  ## a single pass.  All metric lines will have a newline character seperating them.
  enable_gzip_compression = false

  ## Data format to output.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "influx"

  ## debug will show upstream aws messages.
  debug = false
`

func (f *FirehoseOutput) SampleConfig() string {
	return sampleConfig
}

func (f *FirehoseOutput) Description() string {
	return "Configuration for the AWS Firehose output."
}

func (f *FirehoseOutput) Connect() error {
	// We attempt first to create a session to Firehose using an IAMS role, if that fails it will fall through to using
	// environment variables, and then Shared Credentials.
	if f.Debug {
		log.Printf("E! firehose: Building a session for connection to Firehose in %+v", f.Region)
	}

	credentialConfig := &internalaws.CredentialConfig{
		Region:    f.Region,
		AccessKey: f.AccessKey,
		SecretKey: f.SecretKey,
		RoleARN:   f.RoleARN,
		Profile:   f.Profile,
		Filename:  f.Filename,
		Token:     f.Token,
	}
	configProvider := credentialConfig.Credentials()

	// we simply create the skeleton here. AWS doesn't attempt
	// any connection yet, so we don't know if an error will happen.
	svc := firehose.New(configProvider)
	f.svc = svc
	return nil
}

func (f *FirehoseOutput) Close() error {
	return nil
}

func (f *FirehoseOutput) SetSerializer(serializer serializers.Serializer) {
	f.serializer = serializer
}

// writeToFirehose uses the AWS GO SDK to write batched records to firehose and
// queue the failed writes to the errorBuffer of f
func writeToFirehose(f *FirehoseOutput, b [][]byte, submitAttemptCount int64) {
	start := time.Now()
	dataSize := 0 // track the size of the record in bytes
	var values []byte

	// take metrics and flatten into a record

	if f.EnableGzipCompression == true {
		var buf bytes.Buffer
		zw := gzip.NewWriter(&buf)

		for _, line := range b {
			_, err := zw.Write(line)
			if err != nil {
				log.Fatal(err)
			}
		}

		if err := zw.Close(); err != nil {
			log.Fatal(err)
		}

		values = buf.Bytes()
		dataSize = len(values)
	} else {
		for _, line := range b {
			values = append(values, line...)
			dataSize = dataSize + len(line)
		}
	}

	// todo.. check that we are not exceeding the max record size of 1MB to firehose
	// though.. we'd be really impressed if any single telegraf metric write is >1MB. That's a lot.
	// for now, i've updated the docstring to ask that users adjust metric_batch_size and flush_interval
	// if they run into this.  A proper solution here would need to predict and (if needed) move lines
	// around to get as close to 1MB as possible.. the level of effort is high enough to postpone for now.

	d := firehose.Record{
		Data: values,
	}

	r := []*firehose.Record{}
	r = append(r, &d)

	batchInput := &firehose.PutRecordBatchInput{}
	batchInput.SetDeliveryStreamName(f.DeliveryStreamName)
	batchInput.SetRecords(r)

	// attempt to send data to firehose
	resp, err := f.svc.PutRecordBatch(batchInput)

	// if we had a failure, log it and enqueue the request for next time
	if (err != nil) || (*resp.FailedPutCount > 0) {
		log.Printf("E! firehose: Unable to write to Firehose : %+v \n", err.Error())
		newErrorEntry := errorEntry{submitAttemptCount: submitAttemptCount + 1, batch: b}
		f.errorBuffer = append(f.errorBuffer, &newErrorEntry)
		f.totalErrorCount += 1
		return
	}
	if f.Debug {
		log.Printf("E! %+v \n", resp)
	}

	log.Printf("I! firehose: successfully sent %d metric batches in %+v. Total size: %d bytes.\n", len(r), time.Since(start), dataSize)

}

// Write function is responsible for first writing the batch entries held in
// the errorBuffer of f and then writing metrics in batches of 500.
func (f *FirehoseOutput) Write(metrics []telegraf.Metric) error {

	if len(metrics) == 0 {
		return nil
	}

	// if we have any failures from last write, we will attempt them
	// again here.  Note: we only try up to the max number of retry attempts
	// as specified by the configuration file.
	// We'll loop over the errorBuffer and copy out the messages that need resending.
	// We then clear the errorBuffer. We do it this way because we will be appending
	// any future errors to the end, and we need to ensure messages are dropped if unneeded.
	// push/pop of a slice it not completely efficient, so we figured an allocate/deallocate was better.
	if len(f.errorBuffer) > 0 {
		log.Printf("I! firehose: processing %d batches with previous errors queued for retry.", len(f.errorBuffer))
		tempErrorBuffer := []*errorEntry{}

		for i, entry := range f.errorBuffer {
			if entry.submitAttemptCount <= f.MaxSubmitAttempts {
				tempErrorBuffer = append(tempErrorBuffer, entry)
			} else {
				log.Printf("I! firehose: -> discarded failed metric batch %d of %d. Attempt #%d > MaxSubmitAttempts (%d).\n",
					i+1, len(f.errorBuffer), entry.submitAttemptCount, f.MaxSubmitAttempts)
			}
		}

		// clear the errorBuffer (as we're processing it now)
		f.errorBuffer = []*errorEntry{}

		// attempt resubmit
		for i, entry := range tempErrorBuffer {
			log.Printf("I! firehose: -> resending failed metric batch %d of %d.\n", i+1, len(tempErrorBuffer))
			writeToFirehose(f, entry.batch, entry.submitAttemptCount)
		}
		log.Printf("I! firehose: done resending previous batches")
	}

	prepared_batch := make([][]byte, len(metrics))
	for _, metric := range metrics {

		values, err := f.serializer.Serialize(metric)
		if err != nil {
			return err
		}

		prepared_batch = append(prepared_batch, values)
	}

	writeToFirehose(f, prepared_batch, 0)
	return nil
}

func init() {
	outputs.Add("firehose", func() telegraf.Output {
		return &FirehoseOutput{}
	})
}
