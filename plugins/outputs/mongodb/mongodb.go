package mongodb

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
)

func (s *MongoDB) MongoDBGetCollections(ctx context.Context) error {
	s.collections = map[string]bson.M{}
	collections, _ := s.client.Database(s.MetricDatabase).ListCollections(ctx, bson.M{})
	for collections.Next(ctx) {
		var collection bson.M
		if err := collections.Decode(&collection); err != nil {
			s.Log.Error(err)
			return fmt.Errorf("unable to decode ListCollections: %v", err)
		}
		s.collections[collection["name"].(string)] = collection
	}
	return nil
}

func (s *MongoDB) MongoDBInsert(ctx context.Context, database_collection string, bson bson.D) error {
	collection := s.client.Database(s.MetricDatabase).Collection(database_collection)
	_, err := collection.InsertOne(ctx, &bson)
	if err != nil {
		s.Log.Error(err)
	}
	return err
}

type MongoDB struct {
	Dsn                string          `toml:"dsn"`
	AuthenticationType string          `toml:"authentication"`
	MetricDatabase     string          `toml:"database"`
	MetricGranularity  string          `toml:"granularity"`
	Username           string          `toml:"username"`
	Password           string          `toml:"password"`
	CAFile             string          `toml:"cafile"`
	X509clientpem      string          `toml:"x509clientpem"`
	X509clientpempwd   string          `toml:"x509clientpempwd"`
	AllowTLSInsecure   bool            `toml:"allow_tls_insecure"`
	TTL                string          `toml:"ttl"`
	Log                telegraf.Logger `toml:"-"`
	client             *mongo.Client
	clientOptions      *options.ClientOptions
	collections        map[string]bson.M
}

func (s *MongoDB) Description() string {
	return "Sends metrics to MongoDB"
}

var sampleConfig = `
  dsn = "mongodb://localhost:27017/admin"
  # dsn = "mongodb://mongod1:27017,mongod2:27017,mongod3:27017/admin&replicaSet=myReplSet&w=1"
  authentication = "SCRAM" # NONE, SCRAM, or X509
  username = "root" #username for SCRAM 
  password = "****" #password for SCRAM user or private key password if encrypted X509
  # x509clientpem = "client.pem"
  # x509clientpempwd = "****"
  # allow_tls_insecure = false
  # cafile = "ca.pem" #if using X509 authentication
  database = "telegraf" #tells telegraf which database to write metrics to. collections are automatically created as time series collections
  granularity = "seconds" #can be seconds, minutes, or hours
  ttl = "15d" #set a TTL on the collect. examples: 120m, 24h, or 15d
`

func (s *MongoDB) SampleConfig() string {
	return sampleConfig
}

func (s *MongoDB) Init() error {
	serverAPIOptions := options.ServerAPI(options.ServerAPIVersion1) //use new mongodb versioned api
	s.clientOptions = options.Client().SetServerAPIOptions(serverAPIOptions)

	if s.AuthenticationType == "SCRAM" {
		credential := options.Credential{
			AuthMechanism: "SCRAM-SHA-256",
			Username:      s.Username,
			Password:      s.Password,
		}
		s.clientOptions = s.clientOptions.SetAuth(credential)
	} else if s.AuthenticationType == "X509" {
		//format connection string to include tls/x509 options
		new_connection_string, err := url.Parse(s.Dsn)
		if err != nil {
			s.Log.Error(err)
		}
		q := new_connection_string.Query()
		q.Set("tls", "true")
		if s.AllowTLSInsecure {
			q.Set("tlsAllowInvalidCertificates", strconv.FormatBool(s.AllowTLSInsecure))
		}
		q.Set("tlsCAFile", s.CAFile)
		q.Set("sslClientCertificateKeyFile", s.X509clientpem)
		if s.X509clientpempwd != "" {
			q.Set("sslClientCertificateKeyPassword", s.X509clientpempwd)
		}
		new_connection_string.RawQuery = q.Encode()
		s.Dsn = new_connection_string.String()
		// always auth source $external
		credential := options.Credential{
			AuthSource:    "$external",
			AuthMechanism: "MONGODB-X509",
		}
		s.clientOptions = s.clientOptions.SetAuth(credential)
	}

	s.clientOptions = s.clientOptions.ApplyURI(s.Dsn)
	return nil
}

func (s *MongoDB) MongoDBCreateTimeSeriesCollection(database_collection string) error {
	ctx := context.Background()
	tso := options.TimeSeries()
	tso.SetTimeField("timestamp")
	tso.SetMetaField("tags")
	tso.SetGranularity(s.MetricGranularity)

	cco := options.CreateCollection()
	if s.TTL != "" {
		expiregranularity := s.TTL[len(s.TTL)-1:]
		expire_after_seconds, err := strconv.ParseInt(s.TTL[0:len(s.TTL)-1], 10, 64)
		if err != nil {
			s.Log.Error(err)
			return fmt.Errorf("unable to parse ttl: %v", err)
		}
		if expiregranularity == "m" {
			expire_after_seconds = expire_after_seconds * 60
		} else if expiregranularity == "h" {
			expire_after_seconds = expire_after_seconds * 60 * 60
		} else if expiregranularity == "d" {
			expire_after_seconds = expire_after_seconds * 24 * 60 * 60
		}
		cco.SetExpireAfterSeconds(expire_after_seconds)
	}
	cco.SetTimeSeriesOptions(tso)

	err := s.client.Database(s.MetricDatabase).CreateCollection(ctx, database_collection, cco)
	if err != nil {
		s.Log.Error(err)
	}
	return err
}

func (s *MongoDB) DoesCollectionExist(database_collection string) bool {
	_, collectionExists := s.collections[database_collection]
	return collectionExists
}

func (s *MongoDB) UpdateCollectionMap(database_collection string) {
	s.collections[database_collection] = bson.M{"bustedware": "llc"}
}

func (s *MongoDB) Connect() error {
	ctx := context.Background()

	s.Log.Debugf("connecting to " + s.Dsn)
	client, err := mongo.Connect(ctx, s.clientOptions)
	s.client = client
	if err != nil {
		return fmt.Errorf("unable to connect: %v", err)
	}
	s.Log.Debugf("connected!")
	err = s.MongoDBGetCollections(ctx)
	return err
}

func (s *MongoDB) Close() error {
	ctx := context.Background()
	err := s.client.Disconnect(ctx)
	if err != nil {
		s.Log.Error(err)
	}
	s.Log.Debugf("Connection to MongoDB closed.")
	return err
}

// all metric/measurement fields are parent level of document
// metadata field is named "tags"
// mongodb stores timestamp as UTC. conversion should be performed during reads in app or in aggregation pipeline
func MarshalMetric(metric telegraf.Metric) bson.D {
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
		// ensure collection gets created as time series collection.
		if !s.DoesCollectionExist(metric.Name()) {
			s.Log.Debugf("creating time series collection for metric " + metric.Name() + "...")
			s.MongoDBCreateTimeSeriesCollection(metric.Name())
			s.UpdateCollectionMap(metric.Name())
		}

		bson := MarshalMetric(metric)
		err := s.MongoDBInsert(ctx, metric.Name(), bson)
		if err != nil {
			s.Log.Error(err)
		}
	}
	return nil
}

func init() {
	outputs.Add("mongodb", func() telegraf.Output { return &MongoDB{} })
}
