package main

import (
	"flag"
	"fmt"
	"github.com/couchbase/go-couchbase"
	"log"
	"time"
)

var serverURL = flag.String("serverURL", "http://localhost:9000",
	"couchbase server URL")
var poolName = flag.String("poolName", "default",
	"pool name")
var bucketName = flag.String("bucketName", "default",
	"bucket name")

func main() {

	flag.Parse()

	couchbase.EnableMutationToken = true
	client, err := couchbase.Connect(*serverURL)
	if err != nil {
		log.Printf("Connect failed %v", err)
		return
	}

	cbpool, err := client.GetPool("default")
	if err != nil {
		log.Printf("Failed to connect to default pool %v", err)
		return
	}

	var cbbucket *couchbase.Bucket
	cbbucket, err = cbpool.GetBucket(*bucketName)

	if err != nil {
		log.Printf("Failed to connect to bucket %v", err)
		return
	}

	go cbbucket.StartOPPollers(4)
	dsSet := false

	err = cbbucket.SetObserveAndPersist(couchbase.PersistMaster, couchbase.ObserveReplicateTwo)
	if err != nil {
		log.Printf("Not supported %v", err)
	} else {
		dsSet = true
	}

	if dsSet == false {
		err = cbbucket.SetObserveAndPersist(couchbase.PersistMaster, couchbase.ObserveReplicateOne)
		if err != nil {
			log.Printf("Not supported %v", err)
		} else {
			dsSet = true
		}
	}

	if dsSet == false {
		err = cbbucket.SetObserveAndPersist(couchbase.PersistMaster, couchbase.ObserveNone)
		if err != nil {
			log.Fatal(err)
		}
	}

	i := 512
	var mt *couchbase.MutationToken
	var failover bool

	for {
		key := fmt.Sprintf("key%d", i)
		value := fmt.Sprintf("value%d", i)
		mt, err = cbbucket.SetWithMeta(key, 0, 10, value)
		if err != nil {
			log.Printf(" Set operation failed for key %v. error %v", key, err)
			goto skip_mutation
		}
		log.Printf(" Got mutation token %v", mt)

		//observe persist this mutation
		err, failover = cbbucket.ObserveAndPersistPoll(mt.VBid, mt.Guard, mt.Value)
		if err != nil {
			log.Printf("Failure in Observe / Persist %v", err)
		}
		if failover == true {
			log.Printf(" Hard failover, cannot meet durablity requirements")
		}

	skip_mutation:
		<-time.After(1 * time.Second)
		i++
	}
}
