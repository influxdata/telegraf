package main

import (
	"flag"
	"log"

	"github.com/couchbase/gomemcached/client"
)

var prot = flag.String("prot", "tcp", "Layer 3 protocol (tcp, tcp4, tcp6)")
var dest = flag.String("dest", "localhost:11210", "Host:port to connect to")
var u = flag.String("user", "", "SASL plain username")
var p = flag.String("pass", "", "SASL plain password")
var back = flag.Uint64("backfill", memcached.TapNoBackfill,
	"List historical values starting from here")
var dump = flag.Bool("dump", false, "Stop after backfill")
var raw = flag.Bool("raw", false, "Show raw event contents")
var ack = flag.Bool("ack", false, "Request ACKs from server")
var keysOnly = flag.Bool("keysOnly", false, "Send only keys, no values")
var checkpoint = flag.Bool("checkpoint", false, "Send checkpoint events")

func main() {
	flag.Parse()
	log.Printf("Connecting to %s/%s", *prot, *dest)

	client, err := memcached.Connect(*prot, *dest)
	if err != nil {
		log.Fatalf("Error connecting: %v", err)
	}

	if *u != "" {
		resp, err := client.Auth(*u, *p)
		if err != nil {
			log.Fatalf("auth error: %v", err)
		}
		log.Printf("Auth response = %v", resp)
	}

	args := memcached.DefaultTapArguments()
	args.Backfill = uint64(*back)
	args.Dump = *dump
	args.SupportAck = *ack
	args.KeysOnly = *keysOnly
	args.Checkpoint = *checkpoint
	feed, err := client.StartTapFeed(args)
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
	log.Printf("Tap feed closed; err = %v.", feed.Error)
}
