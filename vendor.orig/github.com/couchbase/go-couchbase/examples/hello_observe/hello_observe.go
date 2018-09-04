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
		log.Fatalf("Usage: hello_observe [-pool poolname] [-write] server bucket key")
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

	result, err := bucket.Observe(key)
	if err != nil {
		log.Fatalf("Observe returned error %v", err)
	}
	log.Printf("Observe result: %+v", result)

	if *writeFlag {
		log.Printf("Now writing to key %q with persistence...", key)
		start := time.Now()
		err = bucket.Write(key, 0, 0, "observe test", couchbase.Persist)
		if err != nil {
			log.Fatalf("Write returned error %v", err)
		}
		log.Printf("Write with persistence took %s", time.Since(start))
	}
}
