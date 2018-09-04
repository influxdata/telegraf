package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/couchbase/go-couchbase"
	"github.com/couchbase/gomemcached/client"
)

var poolName = flag.String("pool", "default", "Pool name")
var back = flag.Uint64("backfill", memcached.TapNoBackfill, "List historical values starting from here")
var dump = flag.Bool("dump", false, "Stop after backfill")
var raw = flag.Bool("raw", false, "Show raw event contents")
var ack = flag.Bool("ack", false, "Request ACKs from server")
var keysOnly = flag.Bool("keysOnly", false, "Send only keys, no values")
var checkpoint = flag.Bool("checkpoint", false, "Send checkpoint events")

func main() {
	flag.Parse()

	if len(flag.Args()) < 2 {
		log.Fatalf("Server URL and bucket name required")
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

	args := memcached.DefaultTapArguments()
	args.Backfill = uint64(*back)
	args.Dump = *dump
	args.SupportAck = *ack
	args.KeysOnly = *keysOnly
	args.Checkpoint = *checkpoint
	feed, err := bucket.StartTapFeed(&args)
	if err != nil {
		log.Fatalf("Error starting tap feed: %v", err)
	}
	for op := range feed.C {
		if *raw {
			log.Printf("Received %#v\n", op)
		} else {
			log.Printf("Received %s\n", op.String())
			if len(op.Value) > 0 && len(op.Value) < 500 {
				log.Printf("\tValue: %s", op.Value)
			}
		}
	}
}
