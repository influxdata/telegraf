package file

import (
	"context"

	driver "github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
)

type ArangoDb struct {
	Url            string          `toml:"url"`
	Username       string          `toml:"username"`
	Password       string          `toml:"password"`
	Database       string          `toml:"database"`
	Collection     string          `toml:"collection"`
	Log            telegraf.Logger `toml:"-"`
	UseBatchFormat bool            `toml:"use_batch_format"`

	db         driver.Database
	serializer serializers.Serializer
}

var sampleConfig = `
  ## ArangoDb URL to connect to
  url = "http://192.168.1.100:8529"

  ## Username to connect to ArangoDb
  username = "user"
  ## Password to connect to ArangoDb
  password = "password"

  ## Database to write logs to
  database = "logdb"

  ## Collection to write logs to
  collection = "log_data"
  
  ## Uses batch transactions to insert log records into ArangoDb
  use_batch_format = true
`

func (a *ArangoDb) SetSerializer(serializer serializers.Serializer) {
	a.serializer = serializer
}

func (a *ArangoDb) Connect() error {
	conn, err := http.NewConnection(http.ConnectionConfig{
		Endpoints: []string{a.Url},
	})

	if err != nil {
		a.Log.Errorf("Failed to create HTTP connection: %v", err)
		return err
	}

	c, err := driver.NewClient(driver.ClientConfig{
		Connection:     conn,
		Authentication: driver.BasicAuthentication(a.Username, a.Password),
	})

	exists, err := c.DatabaseExists(nil, a.Database)

	if err != nil {
		a.Log.Errorf("Could not check if database exists: %v", err)
		return err
	}

	var db driver.Database

	if exists == false {
		// Create database as it doesn't exist
		db, err = c.CreateDatabase(nil, a.Database, nil)
	} else {
		db, err = c.Database(nil, a.Database)
	}

	if err != nil {
		a.Log.Errorf("Could not connect to database: %v", err)
		return err
	}

	a.db = db

	if err != nil {
		a.Log.Errorf("Failed to create database: %v", err)
		return err
	}

	exists, err = db.CollectionExists(nil, a.Collection)

	if err != nil {
		a.Log.Errorf("Failed to check if collection exists: %v", err)
		return err
	}

	if exists == false {
		_, err := db.CreateCollection(nil, a.Collection, nil)
		if err != nil {
			a.Log.Errorf("Failed to create collection: %v", err)
			return err
		}
	}

	return nil
}

func (a *ArangoDb) Close() error {
	return nil
}

func (a *ArangoDb) SampleConfig() string {
	return sampleConfig
}

func (a *ArangoDb) Description() string {
	return "Send telegraf metrics to ArangoDb"
}

func (a *ArangoDb) Write(metrics []telegraf.Metric) error {
	var writeErr error = nil

	ctx := context.Background()
	col, err := a.db.Collection(nil, a.Collection)

	if err != nil {
		a.Log.Errorf("Could open collection: %v", err)
		return err
	}

	if a.UseBatchFormat {
		entries := make([]map[string]interface{}, len(metrics))

		for _, metric := range metrics {
			fields := metric.Fields()
			fields["measurement"] = metric.Name()

			entries = append(entries, fields)
		}
		_, _, err := col.CreateDocuments(ctx, entries)
		if err != nil {
			a.Log.Errorf("Could not write entries: %v", err)
			return err
		}
	} else {
		for _, metric := range metrics {
			fields := metric.Fields()
			fields["measurement"] = metric.Name()

			_, err := col.CreateDocument(ctx, fields)
			if err != nil {
				a.Log.Errorf("Could not write entry: %v", err)
			}
		}
	}

	return writeErr
}

func init() {
	outputs.Add("arangodb", func() telegraf.Output {
		return &ArangoDb{}
	})
}
