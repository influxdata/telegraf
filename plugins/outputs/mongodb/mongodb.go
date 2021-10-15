package mongodb

import (
	"context"
	"log"
	"net/url"
	"strconv"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
)

func MongoDBGetCollections(database_name string, client *mongo.Client, ctx context.Context) map[string]bson.M {
	ret := map[string]bson.M{}
	collections, _ := client.Database(database_name).ListCollections(ctx, bson.M{})
	for collections.Next(ctx) {
		var collection bson.M
		if err := collections.Decode(&collection); err != nil {
			log.Fatal(err)
		}
		ret[collection["name"].(string)] = collection
	}
	return ret
}

func (s *MongoDB) MongoDBInsert(database_collection string, bson bson.D) error {
	collection := s.client.Database(s.MetricDatabase).Collection(database_collection)
	// insertResult, err := collection.InsertOne(ctx, &bdoc)
	_, err := collection.InsertOne(s.ctx, &bson)
	if err != nil {
		log.Fatal(err)
	}
	// log.Println("Inserted a single document: ", insertResult.InsertedID)
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
	ctx                context.Context
	collections        (map[string]bson.M)
}

func (s *MongoDB) Description() string {
	return "Configuration for sending metrics to MongoDB"
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

// Init is for setup, and validating config.
func (s *MongoDB) Init() error {
	log.Println("connecting to " + s.Dsn + " with username " + s.Username)
	return nil
}

func (s *MongoDB) MongoDBCreateTimeSeriesCollection(database_collection string) error {
	tso := options.TimeSeries()
	tso.SetTimeField("timestamp")
	tso.SetMetaField("tags")
	tso.SetGranularity(s.MetricGranularity)

	cco := options.CreateCollection()
	//check s,m,d
	if s.TTL != "" {
		expiregranularity := s.TTL[len(s.TTL)-1:]
		expire_after_seconds, err := strconv.ParseInt(s.TTL[0:len(s.TTL)-1], 10, 64)
		if err != nil {
			log.Fatal(err)
			return err
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

	err := s.client.Database(s.MetricDatabase).CreateCollection(s.ctx, database_collection, cco)
	if err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}

func (s *MongoDB) DoesCollectionExist(database_collection string) bool {
	_, collectionExists := s.collections[database_collection]
	return collectionExists
}

func (s *MongoDB) UpdateCollectionMap(database_collection string) {
	s.collections[database_collection] = bson.M{"bustedware": "llc"}
}

func (s *MongoDB) Connect() error {
	ctx := context.TODO()

	connection_string := s.Dsn
	serverAPIOptions := options.ServerAPI(options.ServerAPIVersion1)
	clientOptions := options.Client().SetServerAPIOptions(serverAPIOptions)

	if s.AuthenticationType == "SCRAM" {
		credential := options.Credential{
			AuthMechanism: "SCRAM-SHA-256",
			Username:      s.Username,
			Password:      s.Password,
		}
		clientOptions = clientOptions.SetAuth(credential)
	} else if s.AuthenticationType == "X509" {
		//format connection string to include tls/x509 options
		new_connection_string, err := url.Parse(connection_string)
		if err != nil {
			log.Fatal(err)
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
		connection_string = new_connection_string.String()
		// always auth source $external
		credential := options.Credential{
			AuthSource:    "$external",
			AuthMechanism: "MONGODB-X509",
		}
		clientOptions = clientOptions.SetAuth(credential)
	}
	//TODO
	//https://github.com/mongodb/mongo-go-driver/blob/master/mongo/client_examples_test.go
	//these would only be for enterprise mongodb since community does not support LDAP/KERBEROS

	// LDAP
	// "mongodb://ldap-user:ldap-pwd@localhost:27017/?authMechanism=PLAIN".
	// else if s.Authentication_type == "LDAP" {
	// credential := options.Credential{
	//     AuthMechanism: "PLAIN",
	//     Username:      "ldap-user",
	//     Password:      "ldap-pwd",
	// }
	// }

	// KERBEROS
	// "mongodb://drivers%40KERBEROS.EXAMPLE.COM@mongo-server.example.com:27017/?authMechanism=GSSAPI".
	// else if s.Authentication_type == "KERBEROS" {
	// credential := options.Credential{
	//     AuthMechanism: "GSSAPI",
	//     Username:      "drivers@KERBEROS.EXAMPLE.COM",
	// }
	// }

	// AWS / ATLAS
	//https://github.com/mongodb/mongo-go-driver/blob/master/mongo/client_examples_test.go#L249

	clientOptions = clientOptions.ApplyURI(connection_string)
	client, err := mongo.Connect(ctx, clientOptions)
	s.client = client
	s.ctx = ctx
	if err != nil {
		log.Print("unable to connect")
		log.Fatal(err)
	} else {
		log.Println("connected!")
		s.collections = MongoDBGetCollections(s.MetricDatabase, s.client, s.ctx)
	}

	return err
}

func (s *MongoDB) Close() error {
	err := s.client.Disconnect(s.ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connection to MongoDB closed.")
	return err
}

// all metric/measurement fields are parent level of document
// metadata field is named "tags"
// mongodb stores timestamp as UTC. conversion should be performed during reads in app or in aggregation pipeline
func MarshalMetric(metric telegraf.Metric) bson.D {
	var bdoc bson.D
	for k, v := range metric.Fields() {
		bdoc = append(bdoc, bson.E{k, v})
	}
	var tags bson.D
	for k, v := range metric.Tags() {
		tags = append(tags, bson.E{k, v})
	}
	bdoc = append(bdoc, bson.E{"tags", tags})
	bdoc = append(bdoc, bson.E{"timestamp", metric.Time()})
	return bdoc
}

func (s *MongoDB) Write(metrics []telegraf.Metric) error {
	for _, metric := range metrics {
		// ensure collection gets created as time series collection.
		if !s.DoesCollectionExist(metric.Name()) {
			log.Println("creating time series collection for metric " + metric.Name() + "...")
			s.MongoDBCreateTimeSeriesCollection(metric.Name())
			s.UpdateCollectionMap(metric.Name())
		}

		bson := MarshalMetric(metric)
		err := s.MongoDBInsert(metric.Name(), bson)
		if err != nil {
			log.Fatal(err)
		}
	}
	return nil
}

func init() {
	outputs.Add("mongodb", func() telegraf.Output { return &MongoDB{} })
}
