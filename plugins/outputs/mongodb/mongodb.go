package mongodb

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoDB struct {
	URI             string   `toml:"uri"`
	DB              string   `toml:"db"`
	Collection      string   `toml:"collection"`
	Timeout         int64    `toml:"timeout"`
	MaxPoolSize     int64    `toml:"max_pool_size"`
	MinPoolSize     int64    `toml:"min_pool_size"`
	MaxConnIdleTime int64    `toml:"max_conn_idle_time"`
	Fields          []string `toml:"fields"`
	FieldsCache     map[string]bool
	client          *mongo.Client
	serializer      serializers.Serializer
}

func (m *MongoDB) connect() error {
	return nil
}

func (m *MongoDB) Connect() error {
	return m.connect()
}

func (m *MongoDB) Close() error {
	return nil
}

func (m *MongoDB) Write(metrics []telegraf.Metric) error {
	return nil
}

func (m *MongoDB) SetSerializer(serializer serializers.Serializer) {
	m.serializer = serializer
}

func (m *MongoDB) SampleConfig() string {
	return "sampleConfig"
}

func (m *MongoDB) Description() string {
	return "Configuration for MongoDB server to send metrics to"
}

func init() {
	outputs.Add("mongodb", func() telegraf.Output {
		return &MongoDB{}
	})
}
