package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/couchbase/go-couchbase"
)

var poolName = flag.String("pool", "default", "Pool name")
var writeFlag = flag.Bool("write", false, "If true, will write a value to the key")

func main() {
	flag.Parse()

	if len(flag.Args()) < 3 {
		log.Fatalf("Usage: hello_getandtouch [-pool poolname] [-write] server bucket key")
	}

	c, err := couchbase.Connect(flag.Arg(0))
	if err != nil {
		log.Fatalf("Error connecting:  %v", err)
	}
	fmt.Printf("Connected to ver=%s\n", c.Info.ImplementationVersion)

	pool, err := c.GetPool(*poolName)
	if err != nil {
		log.Fatalf("Can't get pool %q:  %v", *poolName, err)
	}

	bucket, err := pool.GetBucket(flag.Arg(1))
	if err != nil {
		log.Fatalf("Can't get bucket %q:  %v", flag.Arg(1), err)
	}

	key := flag.Arg(2)

	// Write an initial value to the key, with expiry 2s
	if err = bucket.Set(key, 2, []string{"a", "b", "c"}); err != nil {
		log.Fatalf("Set returned error %v", err)
	}

	// Validate that expiry is extended when getAndTouch is called
	for i := 0; i < 10; i++ {
		result, _, err := bucket.GetAndTouchRaw(key, 3)
		if err != nil {
			log.Fatalf("GetAndTouchRaw returned error %v", err)
		}
		if len(result) == 0 {
			log.Fatalf("GetAndTouchRaw returned invalid content", err)
		}
		log.Printf("Successful retrieval via GetAndTouchRaw after %ds", i+1)
		time.Sleep(1 * time.Second)
	}

	// Validate failed retrieval post-expiry.  Use GetAndTouchRaw to shorten expiry,
	// then attempt standard retrieval via GetRaw
	bucket.GetAndTouchRaw(key, 1)
	time.Sleep(2 * time.Second)
	_, err = bucket.GetRaw(key)
	if err == nil {
		log.Fatalf("Retrieved document that should have expired")
	}

}
