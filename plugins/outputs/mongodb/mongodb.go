//go:generate ../../../tools/readme_config_includer/generator
package mongodb

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/outputs"
)

//go:embed sample.conf
var sampleConfig string

type MongoDB struct {
	Dsn                 string          `toml:"dsn"`
	AuthenticationType  string          `toml:"authentication"`
	MetricDatabase      string          `toml:"database"`
	MetricGranularity   string          `toml:"granularity"`
	Username            config.Secret   `toml:"username"`
	Password            config.Secret   `toml:"password"`
	WriteBatch          bool            `toml:"write_batch"`
	MetadataKeys        []string        `toml:"metadata_keys"`
	MetadataTagStrategy string          `toml:"metadata_tag_strategy"`
	ServerSelectTimeout config.Duration `toml:"timeout"`
	TTL                 config.Duration `toml:"ttl"`
	Log                 telegraf.Logger `toml:"-"`
	tls.ClientConfig

	client         *mongo.Client
	options        *options.ClientOptions
	collections    map[string]bool
	metadataFilter filter.Filter
}

func (*MongoDB) SampleConfig() string {
	return sampleConfig
}

func (s *MongoDB) Init() error {
	// Set defaults
	if s.MetricDatabase == "" {
		s.MetricDatabase = "telegraf"
	}
	switch s.MetricGranularity {
	case "":
		s.MetricGranularity = "seconds"
	case "seconds", "minutes", "hours":
	default:
		return errors.New("invalid time series collection granularity. please specify \"seconds\", \"minutes\", or \"hours\"")
	}

	switch s.MetadataTagStrategy {
	case "":
		s.MetadataTagStrategy = "keep"
	case "keep", "move", "clear":
	default:
		if len(s.MetadataKeys) > 0 {
			return fmt.Errorf("invalid 'metadata_tag_strategy' %q", s.MetadataTagStrategy)
		}
	}

	// Do some basic Dsn checks
	if !strings.HasPrefix(s.Dsn, "mongodb://") && !strings.HasPrefix(s.Dsn, "mongodb+srv://") {
		return errors.New("invalid connection string. expected mongodb://host:port/?{options} or mongodb+srv://host:port/?{options}")
	}
	if !strings.Contains(s.Dsn[strings.Index(s.Dsn, "://")+3:], "/") { // append '/' to Dsn if its missing
		s.Dsn = s.Dsn + "/"
	}

	serverAPIOptions := options.ServerAPI(options.ServerAPIVersion1) // use new mongodb versioned api
	s.options = options.Client().SetServerAPIOptions(serverAPIOptions)

	switch s.AuthenticationType {
	case "", "NONE":
		// No authentication
	case "SCRAM":
		if s.Username.Empty() {
			return errors.New("authentication for SCRAM must specify a username")
		}
		if s.Password.Empty() {
			return errors.New("authentication for SCRAM must specify a password")
		}
		username, err := s.Username.Get()
		if err != nil {
			return fmt.Errorf("getting username failed: %w", err)
		}
		password, err := s.Password.Get()
		if err != nil {
			username.Destroy()
			return fmt.Errorf("getting password failed: %w", err)
		}
		credential := options.Credential{
			AuthMechanism: "SCRAM-SHA-256",
			Username:      username.String(),
			Password:      password.String(),
		}
		username.Destroy()
		password.Destroy()
		s.options.SetAuth(credential)
	case "PLAIN":
		if s.Username.Empty() {
			return errors.New("authentication for PLAIN must specify a username")
		}
		if s.Password.Empty() {
			return errors.New("authentication for PLAIN must specify a password")
		}
		usernameRaw, err := s.Username.Get()
		if err != nil {
			return fmt.Errorf("getting username failed: %w", err)
		}
		username := usernameRaw.String()
		usernameRaw.Destroy()

		passwordRaw, err := s.Password.Get()
		if err != nil {
			return fmt.Errorf("getting password failed: %w", err)
		}
		password := passwordRaw.String()
		passwordRaw.Destroy()

		credential := options.Credential{
			AuthMechanism: "PLAIN",
			AuthSource:    "$external",
			Username:      username,
			Password:      password,
		}
		s.options.SetAuth(credential)

		// Check if TLS is enabled (via mongodb+srv:// or tls/ssl query params) and warn if not
		parsedDSN, err := url.Parse(s.Dsn)
		if err != nil {
			return fmt.Errorf("parsing DSN %q failed: %w", s.Dsn, err)
		}

		// mongodb+srv:// implies TLS, so only warn for mongodb:// without TLS
		if parsedDSN.Scheme != "mongodb+srv" {
			q := parsedDSN.Query()
			tlsEnabled := q.Get("tls") == "true" || q.Get("tls") == "1"
			sslEnabled := q.Get("ssl") == "true" || q.Get("ssl") == "1"

			if !tlsEnabled && !sslEnabled {
				s.Log.Warn("PLAIN authentication should be used with TLS enabled for security reasons!")
			}
		}
	case "X509":
		// format connection string to include tls/x509 options
		newConnectionString, err := url.Parse(s.Dsn)
		if err != nil {
			return err
		}
		q := newConnectionString.Query()
		q.Set("tls", "true")
		if s.InsecureSkipVerify {
			q.Set("tlsInsecure", strconv.FormatBool(s.InsecureSkipVerify))
		}
		if s.TLSCA != "" {
			q.Set("tlsCAFile", s.TLSCA)
		}
		q.Set("sslClientCertificateKeyFile", s.TLSKey)
		if s.TLSKeyPwd != "" {
			q.Set("sslClientCertificateKeyPassword", s.TLSKeyPwd)
		}
		newConnectionString.RawQuery = q.Encode()
		s.Dsn = newConnectionString.String()
		// always auth source $external
		credential := options.Credential{
			AuthSource:    "$external",
			AuthMechanism: "MONGODB-X509",
		}
		s.options.SetAuth(credential)
	default:
		return fmt.Errorf("unsupported authentication type %q", s.AuthenticationType)
	}

	if s.ServerSelectTimeout != 0 {
		s.options.SetServerSelectionTimeout(time.Duration(s.ServerSelectTimeout))
	}

	s.options.ApplyURI(s.Dsn)

	// Setup metadata filter if given
	if len(s.MetadataKeys) > 0 {
		f, err := filter.Compile(s.MetadataKeys)
		if err != nil {
			return fmt.Errorf("creating metadata filter failed: %w", err)
		}
		s.metadataFilter = f
	}

	return nil
}

func (s *MongoDB) Connect() error {
	// Connect to the database
	ctx := context.Background()
	client, err := mongo.Connect(ctx, s.options)
	if err != nil {
		return fmt.Errorf("connecting to server failed: %w", err)
	}
	s.client = client

	// Cache the existing collections to prevent recreating those during write
	collections, err := s.client.Database(s.MetricDatabase).ListCollections(ctx, bson.M{})
	if err != nil {
		return fmt.Errorf("listing collections failed: %w", err)
	}

	s.collections = make(map[string]bool, collections.RemainingBatchLength())
	for collections.Next(ctx) {
		var collection bson.M
		if err = collections.Decode(&collection); err != nil {
			return fmt.Errorf("decoding collections failed: %w", err)
		}

		raw, found := collection["name"]
		if !found {
			return fmt.Errorf("name does not exist in collection %+v", collection)
		}
		name, ok := raw.(string)
		if !ok {
			return fmt.Errorf("non-string name %v (%T) in collection", raw, raw)
		}
		s.collections[name] = true
	}

	return nil
}

func (s *MongoDB) Write(metrics []telegraf.Metric) error {
	ctx := context.Background()

	if s.WriteBatch {
		return s.writeBatch(ctx, metrics)
	}

	return s.writeIndividual(ctx, metrics)
}

func (s *MongoDB) Close() error {
	ctx := context.Background()
	return s.client.Disconnect(ctx)
}

func (s *MongoDB) writeIndividual(ctx context.Context, metrics []telegraf.Metric) error {
	// Write one metric at a time
	for _, metric := range metrics {
		name := metric.Name()
		// Create a new collection if it doesn't exist
		if !s.collections[name] {
			if err := s.createCollection(ctx, name); err != nil {
				return fmt.Errorf("creating time series collection %q failed: %w", name, err)
			}
		}
		doc := s.marshal(metric)

		collection := s.client.Database(s.MetricDatabase).Collection(name)
		if _, err := collection.InsertOne(ctx, &doc); err != nil {
			return fmt.Errorf("inserting metric into collection %q failed: %w", name, err)
		}
	}
	return nil
}

func (s *MongoDB) writeBatch(ctx context.Context, metrics []telegraf.Metric) error {
	// Collect metrics by name
	batches := make(map[string][]interface{})
	for _, m := range metrics {
		name := m.Name()
		batches[name] = append(batches[name], s.marshal(m))
	}

	// Write all metrics of a collection at a time
	for name, batch := range batches {
		// Create a new collection if it doesn't exist
		if !s.collections[name] {
			if err := s.createCollection(ctx, name); err != nil {
				return fmt.Errorf("creating time series collection %q failed: %w", name, err)
			}
		}
		collection := s.client.Database(s.MetricDatabase).Collection(name)

		// Write the batch at once
		if _, err := collection.InsertMany(ctx, batch); err != nil {
			return fmt.Errorf("inserting metrics into collection %q failed: %w", name, err)
		}
	}
	return nil
}

func (s *MongoDB) createCollection(ctx context.Context, name string) error {
	// Setup a new timeseries collection for the given metric name
	series := options.TimeSeries()
	series.SetTimeField("timestamp")
	if s.metadataFilter != nil {
		series.SetMetaField("metadata")
	} else {
		series.SetMetaField("tags")
	}
	series.SetGranularity(s.MetricGranularity)

	collection := options.CreateCollection()
	if s.TTL != 0 {
		collection.SetExpireAfterSeconds(int64(time.Duration(s.TTL).Seconds()))
	}
	collection.SetTimeSeriesOptions(series)

	// Create the new collection
	if err := s.client.Database(s.MetricDatabase).CreateCollection(ctx, name, collection); err != nil {
		return err
	}
	s.collections[name] = true

	return nil
}

// Convert a metric into a MongoDB document with all fields being parent level
// of document. Metadata and/or tags will be added as subdocument. MongoDB
// stores timestamp as UTC so conversion should be performed on the query or
// aggregation side.
func (s *MongoDB) marshal(metric telegraf.Metric) bson.D {
	doc := make(bson.D, 0, len(metric.FieldList())+2)
	doc = append(doc, primitive.E{Key: "timestamp", Value: metric.Time()})

	tags := make(bson.D, 0, len(metric.TagList()))
	metadata := make(bson.D, 0, len(s.MetadataKeys))
	for _, t := range metric.TagList() {
		// Add metadata if specified any
		if s.metadataFilter != nil && s.metadataFilter.Match(t.Key) {
			metadata = append(metadata, primitive.E{Key: t.Key, Value: t.Value})
			if s.MetadataTagStrategy == "keep" {
				tags = append(tags, primitive.E{Key: t.Key, Value: t.Value})
			}
		} else if s.MetadataTagStrategy != "clear" {
			tags = append(tags, primitive.E{Key: t.Key, Value: t.Value})
		}
	}

	if s.metadataFilter != nil {
		doc = append(doc, primitive.E{Key: "metadata", Value: metadata})
	}

	if s.metadataFilter == nil || s.MetadataTagStrategy != "clear" {
		doc = append(doc, primitive.E{Key: "tags", Value: tags})
	}

	for _, f := range metric.FieldList() {
		doc = append(doc, primitive.E{Key: f.Key, Value: f.Value})
	}

	return doc
}

func init() {
	outputs.Add("mongodb", func() telegraf.Output { return &MongoDB{} })
}
