package mongodb

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"time"

	"gopkg.in/mgo.v2/bson"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
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

func MongoDBInsert(database_name string, database_collection string, client *mongo.Client, ctx context.Context, json []byte) {
	var bdoc interface{}
	err := bson.UnmarshalJSON(json, &bdoc)
	collection := client.Database(database_name).Collection(database_collection)
	insertResult, err := collection.InsertOne(ctx, &bdoc)
	if err != nil {
		panic(err)
	}
	log.Println("Inserted a single document: ", insertResult.InsertedID)
}

type MongoDB struct {
	Connection_string   string          `toml:"connection_string"`
	Authentication_type string          `toml:"authentication_type"`
	Metric_database     string          `toml:"metric_database"`
	Metric_granularity  string          `toml:"metric_granularity"`
	Username            string          `toml:"username"`
	Password            string          `toml:"password"`
	CAFile              string          `toml:"cafile"`
	X509clientpem       string          `toml:"x509clientpem"`
	X509clientpempwd    string          `toml:"x509clientpempwd"`
	Allow_tls_insecure  bool            `toml:"allow_tls_insecure"`
	Retention_policy    string          `toml:"retention_policy"`
	Log                 telegraf.Logger `toml:"-"`
	client              *mongo.Client
	ctx                 context.Context
	collections         (map[string]bson.M)
	serializer          serializers.Serializer
}

func (s *MongoDB) Description() string {
	return "Configuration for sending metrics to MongoDB"
}

var sampleConfig = `
  connection_string = "mongodb://localhost:27017/admin"
  # connection_string = "mongodb://mongod1:27017,mongod2:27017,mongod3:27017/admin&replicaSet=myReplSet&w=1"
  authentication_type = "SCRAM" # SCRAM or X509
  username = "root" #username for SCRAM 
  password = "****" #password for SCRAM user or private key password if encrypted X509
  # x509clientpem = "client.pem"
  # x509clientpempwd = "****"
  # allow_tls_insecure = false
  # cafile = "ca.pem" #if using X509 authentication
  metric_database = "telegraf" #tells telegraf which database to write metrics to. collections are automatically created as time series collections
  metric_granularity = "seconds" #can be seconds, minutes, or hours
  retention_policy = "15d" #set a TTL on the collect. examples: 120m, 24h, or 15d
  data_format = "json" #always set to json for proper serialization
`

func (s *MongoDB) SampleConfig() string {
	return sampleConfig
}

// Init is for setup, and validating config.
func (s *MongoDB) Init() error {
	log.Println("connecting to " + s.Connection_string + " with username " + s.Username)
	return nil
}

func MongoDBCreateTimeSeriesCollection(s *MongoDB, database_collection string) {
	tso := options.TimeSeries()
	tso.SetTimeField("timestamp")
	tso.SetMetaField("tags")
	tso.SetGranularity(s.Metric_granularity)

	cco := options.CreateCollection()
	//check s,m,d
	if s.Retention_policy != "" {
		expiregranularity := s.Retention_policy[len(s.Retention_policy)-1:]
		expire_after_seconds, err := strconv.ParseInt(s.Retention_policy[0:len(s.Retention_policy)-1], 10, 64)
		if err != nil {
			log.Fatal(err)
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

	err := s.client.Database(s.Metric_database).CreateCollection(s.ctx, database_collection, cco)
	if err != nil {
		log.Panic(err)
	}
}

func DoesCollectionExist(s *MongoDB, database_collection string) bool {
	_, collectionExists := s.collections[database_collection]
	return collectionExists
}

func UpdateCollectionMap(s *MongoDB, database_collection string) {
	s.collections[database_collection] = bson.M{"bustedware": "llc"}
}

func (s *MongoDB) Connect() error {
	ctx := context.TODO()

	connection_string := s.Connection_string
	serverAPIOptions := options.ServerAPI(options.ServerAPIVersion1)
	clientOptions := options.Client().SetServerAPIOptions(serverAPIOptions)

	if s.Authentication_type == "SCRAM" {
		credential := options.Credential{
			AuthMechanism: "SCRAM-SHA-256",
			Username:      s.Username,
			Password:      s.Password,
		}
		clientOptions = clientOptions.SetAuth(credential)
	} else if s.Authentication_type == "X509" {
		//format connection string to include tls/x509 options
		new_connection_string, err := url.Parse(connection_string)
		if err != nil {
			log.Fatal(err)
		}
		q := new_connection_string.Query()
		q.Set("tls", "true")
		if s.Allow_tls_insecure {
			q.Set("tlsAllowInvalidCertificates", strconv.FormatBool(s.Allow_tls_insecure))
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
		s.collections = MongoDBGetCollections(s.Metric_database, s.client, s.ctx)
	}

	return err
}

func (s *MongoDB) Close() error {
	err := s.client.Disconnect(s.ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connection to MongoDB closed.")
	return err
}

// all metric fields are parent level of document
// metadata field is named "tags"
// converts native metric json timestamp to mongodb native timestamp
func NormalizeJSON(metric telegraf.Metric) string {
	bsonstr := "{"
	for k, v := range metric.Fields() {
		if _, ok := v.(string); ok {
			bsonstr = bsonstr + `"` + k + `":"` + v.(string) + `",`
		} else {
			tmpstr := fmt.Sprintf("%v", v)
			bsonstr = bsonstr + `"` + k + `":` + tmpstr + `,`
		}
	}
	bsonstr = bsonstr + "\"tags\":{"
	for k, v := range metric.Tags() {
		nonescapedstr := fmt.Sprintf("%#v", v)
		bsonstr = bsonstr + `"` + k + `":` + nonescapedstr + `,`
	}
	bsonstr = bsonstr[:len(bsonstr)-1] + "},\"timestamp\":ISODate(\"" + metric.Time().UTC().Format(time.RFC3339) + "\")}"
	return bsonstr
}

func (s *MongoDB) Write(metrics []telegraf.Metric) error {
	for _, metric := range metrics {
		// ensure collection gets created as time series collection.
		if !DoesCollectionExist(s, metric.Name()) {
			fmt.Println("creating time series collection for metric " + metric.Name() + "...")
			MongoDBCreateTimeSeriesCollection(s, metric.Name())
			UpdateCollectionMap(s, metric.Name())
		}

		mdb_bson := NormalizeJSON(metric)
		if mdb_bson != "" {
			fmt.Printf("%v\n", mdb_bson)
			MongoDBInsert(s.Metric_database, metric.Name(), s.client, s.ctx, []byte(mdb_bson))
		} else {
			fmt.Printf("null %v\n", mdb_bson)
		}
	}
	return nil
}

func (s *MongoDB) SetSerializer(serializer serializers.Serializer) {
	s.serializer = serializer
}

func init() {
	outputs.Add("mongodb", func() telegraf.Output { return &MongoDB{} })
}
