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
## URI of MongoDB servers. Requried.
uri = "mongodb://admin:123456@localhost:27017"

## mongodb database. Requried.
db  = "testing"

## mongodb collection. Requried.
collection = "numbers"

## timeout for connections, default: 10s
timeout = 10

## Max size for connection pool, default: 10
max_pool_size = 10

## Min size for connection pool, default: 5
min_pool_size = 5

## Max idle time for connections, default: 30m
max_conn_idle_time = 30

## white list fields can insert to collection. If empty, all fields can be inserted.
fields = ["total", "count"]

## Data format to output.
## Each data format has its own unique set of configuration options, read
## more about them here:
## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_OUTPUT.md
data_format = "json"
`
var (
	defaultTimeout         uint64 = 10
	defaultMaxPoolSize     uint64 = 10
	defaultMinPoolSize     uint64 = 5
	defaultMaxConnIdleTime uint64 = 30
)

func (m *MongoDB) connect() error {
	if m.Timeout == 0 {
		m.Timeout = defaultTimeout
	}
	if m.MaxPoolSize == 0 {
		m.MaxPoolSize = defaultMaxPoolSize
	}
	if m.MinPoolSize == 0 {
		m.MinPoolSize = defaultMinPoolSize
	}
	if m.MaxConnIdleTime == 0 {
		m.MaxConnIdleTime = defaultMaxConnIdleTime
	}

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
			if len(m.Fields) == 0 {
				item[tag.Key] = tag.Value
				continue
			}
			if _, ok := m.FieldsCache[tag.Key]; ok {
				item[tag.Key] = tag.Value
			}
		}
		for _, field := range metric.FieldList() {
			if len(m.Fields) == 0 {
				item[field.Key] = field.Value
				continue
			}
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
