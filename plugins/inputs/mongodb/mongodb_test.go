// +build integration

package mongodb

import (
	"context"
	"log"
	"math/rand"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/influxdata/telegraf/testutil"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var server *Server

func testSetup(_ *testing.M) {
	connectionString := os.Getenv("MONGODB_URL")
	if connectionString == "" {
		connectionString = "mongodb://127.0.0.1:27017"
	}

	u, err := url.Parse(connectionString)
	if err != nil {
		log.Fatalf("Unable to parse URL: %v", err)
	}

	opts := options.Client().ApplyURI(connectionString).SetReadPreference(readpref.Nearest())
	client, err := mongo.Connect(context.Background(), opts)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	server = &Server{
		client:   client,
		hostname: u.Host,
		Log:      testutil.Logger{},
	}
}

func testTeardown(_ *testing.M) {
	err := server.client.Disconnect(context.Background())
	if err != nil {
		log.Fatalf("failed to disconnect: %v", err)
	}
}

func TestMain(m *testing.M) {
	// seed randomness for use with tests
	rand.Seed(time.Now().UTC().UnixNano())

	testSetup(m)
	res := m.Run()
	testTeardown(m)

	os.Exit(res)
}
