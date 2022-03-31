package mongodb

import (
	"context"
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
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/outputs"
)

func (s *MongoDB) getCollections(ctx context.Context) error {
	s.collections = map[string]bson.M{}
	collections, err := s.client.Database(s.MetricDatabase).ListCollections(ctx, bson.M{})
	if err != nil {
		return fmt.Errorf("unable to execute ListCollections: %v", err)
	}
	for collections.Next(ctx) {
		var collection bson.M
		if err := collections.Decode(&collection); err != nil {
			return fmt.Errorf("unable to decode ListCollections: %v", err)
		}
		name, ok := collection["name"].(string)
		if !ok {
			return fmt.Errorf("non-string name in %v", collection)
		}
		s.collections[name] = collection
	}
	return nil
}

func (s *MongoDB) insertDocument(ctx context.Context, databaseCollection string, bdoc bson.D) error {
	collection := s.client.Database(s.MetricDatabase).Collection(databaseCollection)
	_, err := collection.InsertOne(ctx, &bdoc)
	return err
}

type MongoDB struct {
	Dsn                 string          `toml:"dsn"`
	AuthenticationType  string          `toml:"authentication"`
	MetricDatabase      string          `toml:"database"`
	MetricGranularity   string          `toml:"granularity"`
	Username            string          `toml:"username"`
	Password            string          `toml:"password"`
	ServerSelectTimeout config.Duration `toml:"timeout"`
	TTL                 config.Duration `toml:"ttl"`
	Log                 telegraf.Logger `toml:"-"`
	client              *mongo.Client
	clientOptions       *options.ClientOptions
	collections         map[string]bson.M
	tls.ClientConfig
}

func (s *MongoDB) Init() error {
	if s.MetricDatabase == "" {
		s.MetricDatabase = "telegraf"
	}
	switch s.MetricGranularity {
	case "":
		s.MetricGranularity = "seconds"
	case "seconds", "minutes", "hours":
	default:
		return fmt.Errorf("invalid time series collection granularity. please specify \"seconds\", \"minutes\", or \"hours\"")
	}

	// do some basic Dsn checks
	if !strings.HasPrefix(s.Dsn, "mongodb://") && !strings.HasPrefix(s.Dsn, "mongodb+srv://") {
		return fmt.Errorf("invalid connection string. expected mongodb://host:port/?{options} or mongodb+srv://host:port/?{options}")
	}
	if !strings.Contains(s.Dsn[strings.Index(s.Dsn, "://")+3:], "/") { //append '/' to Dsn if its missing
		s.Dsn = s.Dsn + "/"
	}

	serverAPIOptions := options.ServerAPI(options.ServerAPIVersion1) //use new mongodb versioned api
	s.clientOptions = options.Client().SetServerAPIOptions(serverAPIOptions)

	switch s.AuthenticationType {
	case "SCRAM":
		if s.Username == "" {
			return fmt.Errorf("SCRAM authentication must specify a username")
		}
		if s.Password == "" {
			return fmt.Errorf("SCRAM authentication must specify a password")
		}
		credential := options.Credential{
			AuthMechanism: "SCRAM-SHA-256",
			Username:      s.Username,
			Password:      s.Password,
		}
		s.clientOptions.SetAuth(credential)
	case "X509":
		//format connection string to include tls/x509 options
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
		s.clientOptions.SetAuth(credential)
	}

	if s.ServerSelectTimeout != 0 {
		s.clientOptions.SetServerSelectionTimeout(time.Duration(s.ServerSelectTimeout))
	}

	s.clientOptions.ApplyURI(s.Dsn)
	return nil
}

func (s *MongoDB) createTimeSeriesCollection(databaseCollection string) error {
	_, collectionExists := s.collections[databaseCollection]
	if !collectionExists {
		ctx := context.Background()
		tso := options.TimeSeries()
		tso.SetTimeField("timestamp")
		tso.SetMetaField("tags")
		tso.SetGranularity(s.MetricGranularity)
		cco := options.CreateCollection()
		if s.TTL != 0 {
			cco.SetExpireAfterSeconds(int64(time.Duration(s.TTL).Seconds()))
		}
		cco.SetTimeSeriesOptions(tso)
		err := s.client.Database(s.MetricDatabase).CreateCollection(ctx, databaseCollection, cco)
		if err != nil {
			return fmt.Errorf("unable to create time series collection: %v", err)
		}
		s.collections[databaseCollection] = bson.M{}
	}
	return nil
}

func (s *MongoDB) Connect() error {
	ctx := context.Background()
	client, err := mongo.Connect(ctx, s.clientOptions)
	if err != nil {
		return fmt.Errorf("unable to connect: %v", err)
	}
	s.client = client
	if err := s.getCollections(ctx); err != nil {
		return fmt.Errorf("unable to get collections from specified metric database: %v", err)
	}
	return nil
}

func (s *MongoDB) Close() error {
	ctx := context.Background()
	return s.client.Disconnect(ctx)
}

// all metric/measurement fields are parent level of document
// metadata field is named "tags"
// mongodb stores timestamp as UTC. conversion should be performed during reads in app or in aggregation pipeline
func marshalMetric(metric telegraf.Metric) bson.D {
	var bdoc bson.D
	for k, v := range metric.Fields() {
		bdoc = append(bdoc, primitive.E{Key: k, Value: v})
	}
	var tags bson.D
	for k, v := range metric.Tags() {
		tags = append(tags, primitive.E{Key: k, Value: v})
	}
	bdoc = append(bdoc, primitive.E{Key: "tags", Value: tags})
	bdoc = append(bdoc, primitive.E{Key: "timestamp", Value: metric.Time()})
	return bdoc
}

func (s *MongoDB) Write(metrics []telegraf.Metric) error {
	ctx := context.Background()
	for _, metric := range metrics {
		if err := s.createTimeSeriesCollection(metric.Name()); err != nil {
			return err
		}
		bdoc := marshalMetric(metric)
		if err := s.insertDocument(ctx, metric.Name(), bdoc); err != nil {
			return err
		}
	}
	return nil
}

func init() {
	outputs.Add("mongodb", func() telegraf.Output { return &MongoDB{} })
}
