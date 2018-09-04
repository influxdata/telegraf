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

	couchbase.SetConnectionPoolParams(256, 16)
	couchbase.SetTcpKeepalive(true, 30)

	go performOp(cbbucket)

	errCh := make(chan error)

	cbbucket.RunBucketUpdater(func(bucket string, err error) {
		log.Printf(" Updated retured err %v", err)
		errCh <- err
	})

	<-errCh

}

func performOp(b *couchbase.Bucket) {

	i := 512
	key := fmt.Sprintf("k%d", i)
	value := fmt.Sprintf("value%d", i)
	err := b.Set(key, len(value), value)
	if err != nil {
		log.Printf("set failed error %v", err)
		return
	}
	var rv interface{}
	var cas uint64
	var mt *couchbase.MutationToken
	// get the CAS value for this key
	err = b.Gets(key, &rv, &cas)
	if err != nil {
		log.Printf("Gets failed. error %v", err)
		return
	}

	for {
		value = fmt.Sprintf("value%d", i)
		cas, mt, err = b.CasWithMeta(key, 0, 10, cas, value)
		if err != nil {
			log.Printf(" Cas2 operation failed. error %v", err)
			return
		}
		log.Printf(" Got new cas value %v mutation token %v", cas, mt)
		var flags, expiry int
		var seqNo uint64

		err = b.GetMeta(key, &flags, &expiry, &cas, &seqNo)
		if err != nil {
			log.Printf(" Failed to get meta . Error %v", err)
			return
		}

		log.Printf(" meta values for key %s. Flags %d, Expiry %v, Cas %d, Sequence %d", key, flags, time.Unix(int64(expiry), 0), cas, seqNo)

		<-time.After(1 * time.Second)
		i++
	}

}
