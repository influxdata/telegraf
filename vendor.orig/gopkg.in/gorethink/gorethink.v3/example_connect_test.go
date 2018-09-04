package gorethink_test

import (
	"log"
	"os"

	r "gopkg.in/gorethink/gorethink.v3"
)

var session *r.Session
var url string

func init() {
	// If the test is being run by wercker look for the rethink url
	url = os.Getenv("RETHINKDB_URL")
	if url == "" {
		url = "localhost:28015"
	}
}

func ExampleConnect() {
	var err error

	session, err = r.Connect(r.ConnectOpts{
		Address: url,
	})
	if err != nil {
		log.Fatalln(err.Error())
	}
}

func ExampleConnect_connectionPool() {
	var err error

	session, err = r.Connect(r.ConnectOpts{
		Address:    url,
		InitialCap: 10,
		MaxOpen:    10,
	})
	if err != nil {
		log.Fatalln(err.Error())
	}
}

func ExampleConnect_cluster() {
	var err error

	session, err = r.Connect(r.ConnectOpts{
		Addresses: []string{url},
		//  Addresses: []string{url1, url2, url3, ...},
	})
	if err != nil {
		log.Fatalln(err.Error())
	}
}
