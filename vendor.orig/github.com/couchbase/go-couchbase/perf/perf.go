package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/couchbase/go-couchbase"
	"log"
	"runtime"
	"sync"
	"time"
)

func maybeFatal(err error) {
	if err != nil {
		log.Fatalf("Error:  %v", err)
	}
}

var serverURL = flag.String("serverURL", "http://localhost:9000",
	"couchbase server URL")
var poolName = flag.String("poolName", "default",
	"pool name")
var bucketName = flag.String("bucketName", "default",
	"bucket name")
var set = flag.Bool("set", false, "create document mode")
var size = flag.Int("size", 1024, "document size")
var documents = flag.Int("documents", 2000000, "total documents")
var threads = flag.Int("threads", 10, "Number of threads")
var quantum = flag.Int("quantum", 1024, "Number of documents per bulkGet")
var jsonDoc = flag.Bool("generate-json", false, "generate json documents")

var wg sync.WaitGroup

func main() {
	flag.Parse()

	runtime.GOMAXPROCS(*threads)
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

	start := time.Now()
	if *set == true {
		var value []byte
		if *jsonDoc == false {
			value = generateRandomDoc(*size)
		}
		for i := 0; i < *threads; i++ {
			go doSetOps(cbbucket, i*(*documents / *threads), *documents / *threads, value)
			wg.Add(1)
		}
	} else {
		for i := 0; i < *threads; i++ {
			go doBulkGetOps(cbbucket, *documents / *threads, *quantum, i*(*documents / *threads))
			wg.Add(1)
		}
	}

	wg.Wait()

	finish := time.Now().Sub(start)
	fmt.Printf("**** Did %d ops in %s. Ops/sec %d\n",
		*documents, finish.String(), int(float64(*documents)/finish.Seconds()))

	if err != nil {
		log.Printf("Failed to connect to bucket %v", err)
		return
	}

}

func doBulkGetOps(b *couchbase.Bucket, total int, quantum int, startNum int) {

	defer wg.Done()
	start := time.Now()
	iter := total / quantum
	currentKeyNum := startNum
	for i := 0; i < iter; i++ {

		keylist := make([]string, quantum, quantum)
		for j := 0; j < quantum; j++ {
			key := fmt.Sprintf("test%d", currentKeyNum)
			keylist[j] = key
			currentKeyNum++

		}
		_, err := b.GetBulk(keylist)
		if err != nil {
			log.Printf(" Failed to get keys startnum %s to %d", keylist[0], quantum)
		}
	}
	fmt.Printf("Did %d ops in %s\n",
		total, time.Now().Sub(start).String())
}

func generateRandomDoc(size int) []byte {

	rb := make([]byte, size)
	_, err := rand.Read(rb)

	if err != nil {
		log.Fatal("Cannot generate data %v", err)
	}

	rs := base64.URLEncoding.EncodeToString(rb)
	data := map[string]interface{}{"data": rs}

	encode, _ := json.Marshal(data)
	return encode

}

func doSetOps(b *couchbase.Bucket, startNum int, total int, data []byte) {

	defer wg.Done()

	start := time.Now()

	var err error
	for i := 0; i < total; i++ {
		if data == nil {
			data, err = generateRandomDocument()
			if err != nil {
				log.Fatal(err)
			}
		}

		k := fmt.Sprintf("test%d", startNum+i)
		maybeFatal(b.SetRaw(k, 0, data))
	}
	fmt.Printf("Did %d ops in %s\n",
		total, time.Now().Sub(start).String())
}
