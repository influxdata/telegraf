package gcs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"cloud.google.com/go/storage"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"
)

const (
	emulatorHostEnv  = "STORAGE_EMULATOR_HOST"
	defaultOffSetKey = "offset-key.json"
)

type GCS struct {
	CredentialsFile string `toml:"credentials_file"`
	Bucket          string `toml:"bucket"`

	Prefix              string `toml:"key_prefix"`
	OffsetKey           string `toml:"offset_key"`
	ObjectsPerIteration int    `toml:"objects_per_iteration"`

	Log    telegraf.Logger
	offSet OffSet

	parser parsers.Parser
	client *storage.Client

	ctx context.Context
}

type OffSet struct {
	OffSet string `json:"offSet"`
}

func NewEmptyOffset() *OffSet {
	return &OffSet{OffSet: ""}
}

func NewOffset(offset string) *OffSet {
	return &OffSet{OffSet: offset}
}

func (offSet *OffSet) isPresent() bool {
	return offSet.OffSet != ""
}

func (gcs *GCS) SampleConfig() string {
	return sampleConfig
}

func (gcs *GCS) SetParser(parser parsers.Parser) {
	gcs.parser = parser
}

func (gcs *GCS) Description() string {
	return "Read metrics from Google Cloud Storage"
}

func (gcs *GCS) Gather(acc telegraf.Accumulator) error {
	query := gcs.createQuery()

	bucketName := gcs.Bucket
	bucket := gcs.client.Bucket(bucketName)
	it := bucket.Objects(gcs.ctx, &query)

	gcs.Log.Debugf("Reading Data from Bucket ", bucketName)

	processed := 0

	var name string
	for {
		attrs, err := it.Next()

		if err == iterator.Done {
			gcs.Log.Infof("Iterated all the keys")
			break
		}

		if err != nil {
			gcs.Log.Errorf("Error during iteration of keys", err)
			return err
		}

		name = attrs.Name

		if !gcs.shoudIgnore(name) {
			if err := gcs.processMeasurementsInObject(name, bucket, acc); err != nil {
				gcs.Log.Errorf("Could not process object: %v in bucket: %v", name, bucketName, err)
				acc.AddError(fmt.Errorf("Could not process object: %v in bucket: %v", name, err))
			}
		}

		processed++

		if gcs.reachedThreshlod(processed) {
			return gcs.updateOffset(bucket, name)
		}
	}

	return gcs.updateOffset(bucket, name)
}

func (gcs *GCS) createQuery() storage.Query {
	if gcs.offSet.isPresent() {
		return storage.Query{Prefix: gcs.Prefix, StartOffset: gcs.offSet.OffSet}
	}

	return storage.Query{Prefix: gcs.Prefix}
}

func (gcs *GCS) shoudIgnore(name string) bool {
	return gcs.offSet.OffSet == name
}

func (gcs *GCS) processMeasurementsInObject(name string, bucket *storage.BucketHandle, acc telegraf.Accumulator) error {
	gcs.Log.Debugf("Fetching key: %s", name)
	r, err := bucket.Object(name).NewReader(gcs.ctx)
	defer gcs.closeReader(r)

	if err != nil {
		return err
	}

	metrics, err := gcs.fetchedMetrics(r)

	if err != nil {
		return err
	}

	for _, metric := range metrics {
		acc.AddFields(metric.Name(), metric.Fields(), metric.Tags(), metric.Time())
		gcs.Log.Debug("the metrics", metric.Name(), metric.Fields(), metric.Tags(), metric.Time())
	}

	return nil
}

func (gcs *GCS) fetchedMetrics(r *storage.Reader) ([]telegraf.Metric, error) {
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(r); err != nil {
		return nil, err
	}

	return gcs.parser.Parse(buf.Bytes())
}

func (gcs *GCS) reachedThreshlod(processed int) bool {
	return gcs.ObjectsPerIteration != 0 && processed >= gcs.ObjectsPerIteration
}

func (gcs *GCS) updateOffset(bucket *storage.BucketHandle, name string) error {
	offsetModel := NewOffset(name)
	marshalled, err := json.Marshal(offsetModel)

	if err != nil {
		return err
	}

	offsetKey := bucket.Object(gcs.OffsetKey)
	writer := offsetKey.NewWriter(gcs.ctx)
	writer.ContentType = "application/json"
	defer writer.Close()

	if _, err := writer.Write(marshalled); err != nil {
		return err
	}

	gcs.offSet = *offsetModel

	return nil
}

func (gcs *GCS) Init() error {
	gcs.ctx = context.Background()
	err := gcs.setUpClient()
	if err != nil {
		gcs.Log.Error("Could not create client", err)
		return err
	}

	return gcs.setOffset()
}

func (gcs *GCS) setUpClient() error {
	if endpoint, present := os.LookupEnv(emulatorHostEnv); present {
		return gcs.setUpLocalClient(endpoint)
	}

	return gcs.setUpDefaultClient()
}

func (gcs *GCS) setUpLocalClient(endpoint string) error {
	noAuth := option.WithoutAuthentication()
	endpoints := option.WithEndpoint("http://" + endpoint)
	client, err := storage.NewClient(gcs.ctx, noAuth, endpoints)

	if err != nil {
		return err
	}

	gcs.client = client
	return nil
}

func (gcs *GCS) setUpDefaultClient() error {
	var credentialsOption option.ClientOption

	if gcs.CredentialsFile != "" {
		credentialsOption = option.WithCredentialsFile(gcs.CredentialsFile)
	} else {
		creds, err := google.FindDefaultCredentials(gcs.ctx, storage.ScopeReadOnly)
		if err != nil {
			return fmt.Errorf(
				"unable to find GCP Application Default Credentials: %v."+
					"Either set ADC or provide CredentialsFile config", err)
		}
		credentialsOption = option.WithCredentials(creds)
	}

	client, err := storage.NewClient(gcs.ctx, credentialsOption)
	gcs.client = client
	return err
}

func (gcs *GCS) setOffset() error {
	if gcs.client == nil {
		return fmt.Errorf("Cannot set offset if client is not set")
	}

	if gcs.OffsetKey != "" {
		gcs.OffsetKey = gcs.Prefix + gcs.OffsetKey
	} else {
		gcs.OffsetKey = gcs.Prefix + defaultOffSetKey
	}

	btk := gcs.client.Bucket(gcs.Bucket)
	obj := btk.Object(gcs.OffsetKey)

	var offSet OffSet

	if r, err := obj.NewReader(gcs.ctx); err == nil {
		defer gcs.closeReader(r)
		buf := new(bytes.Buffer)

		if _, err := io.Copy(buf, r); err == nil {
			if marshalError := json.Unmarshal(buf.Bytes(), &offSet); marshalError != nil {
				return marshalError
			}
		}
	} else {
		offSet = *NewEmptyOffset()
	}

	gcs.offSet = offSet

	return nil
}

func init() {
	inputs.Add("google_cloud_storage", func() telegraf.Input {
		gcs := &GCS{}
		return gcs
	})
}

func (gcs *GCS) closeReader(r *storage.Reader) {
	err := r.Close()
	gcs.Log.Errorf("Could not close reader", err)
}

const sampleConfig = `
  ## Required. Name of Cloud Storage bucket to ingest metrics from.
  bucket = "my-bucket"

  ## Optional. Prefix of Cloud Storage bucket keys to list metrics from.
  # key_prefix = "my-bucket"

  ## Key that will store the offsets in order to pick up where the ingestion was left.
  offset_key = "offset_key"

  ## Key that will store the offsets in order to pick up where the ingestion was left.
  objects_per_iteration = 10

  ## Required. Data format to consume.
  ## Each data format has its own unique set of configuration options.
  ## Read more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "influx"

  ## Optional. Filepath for GCP credentials JSON file to authorize calls to
  ## Google Cloud Storage APIs. If not set explicitly, Telegraf will attempt to use
  ## Application Default Credentials, which is preferred.
  # credentials_file = "path/to/my/creds.json"
`
