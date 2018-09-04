package main

import (
	"fmt"
	"github.com/couchbase/go-couchbase"
	"log"
	"runtime"
	"time"
)

func IncrementCounter(bucket *couchbase.Bucket, key string, amount uint64) {

	fmt.Printf("Trying to increment counter, key=%s, amount=%d\n", key, amount)

	value, err := bucket.Incr(key, amount, amount, 0)

	if err != nil {
		fmt.Printf("Error happened while incrementing %s\n", err)
	} else {

		fmt.Printf("Incremented counter, new value=%d\n", value)
	}

	value, err = bucket.Decr(key, amount, amount, 0)
	if err != nil {
		fmt.Printf("Error happened while decrementing %s\n", err)
	} else {

		fmt.Printf("Decremented counter, new value=%d\n", value)
	}

}

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())

	c, err := couchbase.Connect("http://localhost:9000/")
	if err != nil {
		log.Fatalf("Error connecting:  %v", err)
	}

	pool, err := c.GetPool("default")
	if err != nil {
		log.Fatalf("Error getting pool:  %v", err)
	}

	bucket, err := pool.GetBucket("default")
	if err != nil {
		log.Fatalf("Error getting bucket:  %v", err)
	}

	bucket.Delete("12345")

	go IncrementCounter(bucket, "12345", 2)
	go IncrementCounter(bucket, "12345", 2)
	go IncrementCounter(bucket, "12345", 2)

	time.Sleep(10000000)
}
