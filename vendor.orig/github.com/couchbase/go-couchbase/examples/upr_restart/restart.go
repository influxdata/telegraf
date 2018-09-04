package main

import (
	"fmt"
	"log"
	"time"

	"github.com/couchbase/go-couchbase"
	"github.com/couchbase/gomemcached"
	"github.com/couchbase/gomemcached/client"
)

var vbcount = 2

const TESTURL = "http://localhost:9000"

// Flush the bucket before trying this program
func main() {
	// get a bucket and mc.Client connection
	bucket, err := getTestConnection("default")
	if err != nil {
		panic(err)
	}

	// start upr feed
	feed, err := bucket.StartUprFeed("index" /*name*/, 0)
	if err != nil {
		panic(err)
	}

	for i := 0; i < vbcount; i++ {
		err := feed.UprRequestStream(
			uint16(i) /*vbno*/, uint16(0) /*opaque*/, 0 /*flag*/, 0, /*vbuuid*/
			0 /*seqStart*/, 0xFFFFFFFFFFFFFFFF /*seqEnd*/, 0 /*snaps*/, 0)
		if err != nil {
			fmt.Printf("%s", err.Error())
		}
	}

	vbseqNo := receiveMutations(feed, 20000)

	vbList := make([]uint16, 0)
	for i := 0; i < vbcount; i++ {
		vbList = append(vbList, uint16(i))
	}
	failoverlogMap, err := bucket.GetFailoverLogs(vbList)
	if err != nil {
		log.Printf(" error in failover log request %s", err.Error())

	}

	// get a bucket and mc.Client connection
	bucket1, err := getTestConnection("default")
	if err != nil {
		panic(err)
	}

	// add mutations to the bucket
	var mutationCount = 5000
	addKVset(bucket1, mutationCount)

	log.Println("Restarting ....")
	feed, err = bucket.StartUprFeed("index" /*name*/, 0)
	if err != nil {
		panic(err)
	}

	for i := 0; i < vbcount; i++ {
		log.Printf("Vbucket %d High sequence number %d, Snapshot end sequence %d", i, vbseqNo[i][0], vbseqNo[i][1])
		failoverLog := failoverlogMap[uint16(i)]
		err := feed.UprRequestStream(
			uint16(i) /*vbno*/, uint16(0) /*opaque*/, 0, /*flag*/
			failoverLog[0][0],                              /*vbuuid*/
			vbseqNo[i][0] /*seqStart*/, 0xFFFFFFFFFFFFFFFF, /*seqEnd*/
			0 /*snaps*/, vbseqNo[i][1])
		if err != nil {
			fmt.Printf("%s", err.Error())
		}
	}

	var e, f *memcached.UprEvent
	var mutations int
loop:
	for {
		select {
		case f = <-feed.C:
		case <-time.After(time.Second):
			break loop
		}

		if f.Opcode == gomemcached.UPR_MUTATION {
			vbseqNo[f.VBucket][0] = f.Seqno
			e = f
			mutations += 1
		}
	}

	log.Printf(" got %d mutations", mutations)

	exptSeq := vbseqNo[e.VBucket][0] + 1

	if e.Seqno != exptSeq {
		fmt.Printf("Expected seqno %v, received %v", exptSeq+1, e.Seqno)
		//panic(err)
	}
	feed.Close()
}

func addKVset(b *couchbase.Bucket, count int) {
	for i := 0; i < count; i++ {
		key := fmt.Sprintf("key%v", i)
		value := fmt.Sprintf("Hello world%v", i)
		if err := b.Set(key, 0, value); err != nil {
			panic(err)
		}
	}
}

func receiveMutations(feed *couchbase.UprFeed, breakAfter int) [][2]uint64 {
	var vbseqNo = make([][2]uint64, vbcount)
	var mutations = 0
	var ssMarkers = 0
	var e *memcached.UprEvent
loop:
	for {
		select {
		case e = <-feed.C:
		case <-time.After(time.Second):
			break loop
		}

		if e.Opcode == gomemcached.UPR_MUTATION {
			vbseqNo[e.VBucket][0] = e.Seqno
			mutations += 1
		}

		if e.Opcode == gomemcached.UPR_MUTATION {
			vbseqNo[e.VBucket][1] = e.SnapendSeq
			ssMarkers += 1
		}
		if mutations == breakAfter {
			break loop
		}
	}

	log.Printf(" Mutation count %d, Snapshot markers %d", mutations, ssMarkers)

	return vbseqNo
}

func getTestConnection(bucketname string) (*couchbase.Bucket, error) {
	couch, err := couchbase.Connect(TESTURL)
	if err != nil {
		fmt.Println("Make sure that couchbase is at", TESTURL)
		return nil, err
	}
	pool, err := couch.GetPool("default")
	if err != nil {
		return nil, err
	}
	bucket, err := pool.GetBucket(bucketname)
	return bucket, err
}
