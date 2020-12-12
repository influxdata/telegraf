package mongodb

import (
	"context"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"gopkg.in/mgo.v2/bson"
)

type MongoDB struct {
	serializer serializers.Serializer

	URI             string   `toml:"uri"`
	DB              string   `toml:"db"`
	Collection      string   `toml:"collection"`
	Timeout         uint64   `toml:"timeout"`
	MaxPoolSize     uint64   `toml:"max_pool_size"`
	MinPoolSize     uint64   `toml:"min_pool_size"`
	MaxConnIdleTime uint64   `toml:"max_conn_idle_time"`
	Fields          []string `toml:"fields"`
	FieldsCache     map[string]bool

	client     *mongo.Client
	collection *mongo.Collection
}

var sampleConfig = `
## URLs of MongoDB servers
uri = "mongodb://admin:123456@localhost:27017"
db  = "testing"
collection = "numbers"
timeout = 15
max_pool_size = 10
min_pool_size = 5
max_conn_idle_time = 5
fields = ["total", "count"]

## Data format to output.
## Each data format has its own unique set of configuration options, read
## more about them here:
## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
data_format = "json"
`

func (m *MongoDB) connect() error {
	clientOptions := options.Client().
		ApplyURI(m.URI).
		SetMaxPoolSize(m.MaxPoolSize).
		SetMinPoolSize(m.MinPoolSize).
		SetMaxConnIdleTime(time.Duration(m.MaxConnIdleTime) * time.Minute)

	var err error

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(m.Timeout)*time.Second)
	defer cancel()

	m.client, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		return err
	}

	return m.client.Ping(ctx, readpref.Primary())
}

func (m *MongoDB) init() error {
	for _, field := range m.Fields {
		m.FieldsCache[field] = true
	}
	m.collection = m.client.Database(m.DB).Collection(m.Collection)
	return nil
}

func (m *MongoDB) Connect() error {
	var err error
	err = m.connect()
	if err != nil {
		return err
	}
	return m.init()
}

func (m *MongoDB) Close() error {
	if m.client != nil {
		m.client.Disconnect(context.Background())
		m.client = nil
	}
	return nil
}

func (m *MongoDB) write(data []interface{}) error {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(m.Timeout)*time.Second)
	_, err := m.collection.InsertMany(ctx, data)
	return err
}

func (m *MongoDB) Write(metrics []telegraf.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	var data []interface{}
	for _, metric := range metrics {
		_, err := m.serializer.Serialize(metric)
		if err != nil {
			return err
		}

		item := bson.M{}
		for _, tag := range metric.TagList() {
			if _, ok := m.FieldsCache[tag.Key]; ok {
				item[tag.Key] = tag.Value
			}
		}
		for _, field := range metric.FieldList() {
			if _, ok := m.FieldsCache[field.Key]; ok {
				item[field.Key] = field.Value
			}
		}
		data = append(data, item)
	}
	return m.write(data)
}

func (m *MongoDB) SetSerializer(serializer serializers.Serializer) {
	m.serializer = serializer
}

func (m *MongoDB) SampleConfig() string {
	return sampleConfig
}

func (m *MongoDB) Description() string {
	return "Configuration for MongoDB server to send metrics to"
}

func init() {
	outputs.Add("mongodb", func() telegraf.Output {
		return &MongoDB{}
	})
}
