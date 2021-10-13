package mongodb

import (
	"context"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"testing"
	"time"
)

func TestConnectNoAuthAndInsertDocument(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	metric_database := "myMetricDatabase"
	metric_collection := "myMetricCollection"
	metric_granularity := "minutes"
	retention_policy := "15d"
	connection_string := "mongodb://localhost:27017"
	serverAPIOptions := options.ServerAPI(options.ServerAPIVersion1)
	clientOptions := options.Client().SetServerAPIOptions(serverAPIOptions)
	clientOptions = clientOptions.ApplyURI(connection_string)

	// test connect
	client, err := mongo.Connect(ctx, clientOptions)
	require.NoError(t, err)

	// test try to create time series collection. if it already exists as non time series
	// the collection would have to be dropped first
	collections := MongoDBGetCollections(metric_database, client, ctx)
	_, collectionExists := collections[metric_collection]
	if !collectionExists {
		MongoDBCreateTimeSeriesCollection(client, ctx, metric_database, metric_collection, metric_granularity, retention_policy)
	}

	// test insert
	mdb_json := "{\"measurement1\":50,\"measurement2\":\"value2\",\"timestamp\":ISODate(\"" + time.Now().UTC().Format(time.RFC3339) + "\"),\"tags\":{\"host\":\"myHostName\"}}"

	err = MongoDBInsert(metric_database, metric_collection, client, ctx, []byte(mdb_json))
	require.NoError(t, err)
}
