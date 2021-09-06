//go:build integration
// +build integration

package mongodb

import (
	"context"
	"log"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/influxdata/telegraf/testutil"
)

var server *Server

func testSetup(_ *testing.M) {
	connectionString := os.Getenv("MONGODB_URL")
	if connectionString == "" {
		connectionString = "mongodb://127.0.0.1:27017"
	}

	m := &MongoDB{
		Log:     testutil.Logger{},
		Servers: []string{connectionString},
	}
	err := m.Init()
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	server = m.clients[0]
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
