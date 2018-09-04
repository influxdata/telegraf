package main

import (
	"github.com/couchbase/go-couchbase"
	"log"
)

const testURL = "http://localhost:9000"

func main() {
	// get a bucket and mc.Client connection
	bucket, err := getTestConnection("default")
	if err != nil {
		panic(err)
	}

	// Get failover log for a vbucket
	flogs, err := bucket.GetFailoverLogs([]uint16{0, 1, 2, 3, 4, 5, 6, 7})
	if err != nil {
		panic(err)
	}
	for vbno, flog := range flogs {
		log.Printf("Failover logs for vbucket %v: %v", vbno, flog)
	}
}

func getTestConnection(bucketname string) (*couchbase.Bucket, error) {
	couch, err := couchbase.Connect(testURL)
	if err != nil {
		log.Println("Make sure that couchbase is at", testURL)
		return nil, err
	}
	pool, err := couch.GetPool("default")
	if err != nil {
		return nil, err
	}
	bucket, err := pool.GetBucket(bucketname)
	return bucket, err
}
