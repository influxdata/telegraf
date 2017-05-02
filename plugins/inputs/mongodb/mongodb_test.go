// +build integration

package mongodb

import (
	"log"
	"math/rand"
	"net/url"
	"os"
	"testing"
	"time"

	"gopkg.in/mgo.v2"
)

var connect_url string
var server *Server

func init() {
	connect_url = os.Getenv("MONGODB_URL")
	if connect_url == "" {
		connect_url = "127.0.0.1:27017"
		server = &Server{Url: &url.URL{Host: connect_url}}
	} else {
		full_url, err := url.Parse(connect_url)
		if err != nil {
			log.Fatalf("Unable to parse URL (%s), %s\n", full_url, err.Error())
		}
		server = &Server{Url: full_url}
	}
}

func testSetup(m *testing.M) {
	var err error
	var dialAddrs []string
	if server.Url.User != nil {
		dialAddrs = []string{server.Url.String()}
	} else {
		dialAddrs = []string{server.Url.Host}
	}
	dialInfo, err := mgo.ParseURL(dialAddrs[0])
	if err != nil {
		log.Fatalf("Unable to parse URL (%s), %s\n", dialAddrs[0], err.Error())
	}
	dialInfo.Direct = true
	dialInfo.Timeout = 5 * time.Second
	sess, err := mgo.DialWithInfo(dialInfo)
	if err != nil {
		log.Fatalf("Unable to connect to MongoDB, %s\n", err.Error())
	}
	server.Session = sess
	server.Session, _ = mgo.Dial(server.Url.Host)
	if err != nil {
		log.Fatalln(err.Error())
	}
}

func testTeardown(m *testing.M) {
	server.Session.Close()
}

func TestMain(m *testing.M) {
	// seed randomness for use with tests
	rand.Seed(time.Now().UTC().UnixNano())

	testSetup(m)
	res := m.Run()
	testTeardown(m)

	os.Exit(res)
}
