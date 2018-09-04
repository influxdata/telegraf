package main

import (
	"flag"
	"fmt"
	"github.com/couchbase/go-couchbase"
	"github.com/couchbase/gomemcached"
	"github.com/couchbase/gomemcached/client"
	"log"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"net/url"
	"os"
	"runtime/pprof"
	"time"
)

var vbcount = 64
var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
var memprofile = flag.String("memprofile", "", "write memory profile to this file")

func mf(err error, msg string) {
	if err != nil {
		log.Fatalf("%v: %v", msg, err)
	}
}

// Flush the bucket before trying this program
func main() {

	//runtime.GOMAXPROCS(4)

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	bname := flag.String("bucket", "",
		"bucket to connect to (defaults to username)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr,
			"%v [flags] http://user:pass@host:8091/\n\nFlags:\n",
			os.Args[0])
		flag.PrintDefaults()
		os.Exit(64)
	}

	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
	}

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal(err)
		}
		defer pprof.WriteHeapProfile(f)
		defer f.Close()
	}

	u, err := url.Parse(flag.Arg(0))
	mf(err, "parse")

	if *bname == "" && u.User != nil {
		*bname = u.User.Username()
	}

	c, err := couchbase.Connect(u.String())
	mf(err, "connect - "+u.String())

	p, err := c.GetPool("default")
	mf(err, "pool")

	bucket, err := p.GetBucket(*bname)
	mf(err, "bucket")

	// start upr feed
	name := fmt.Sprintf("%v", time.Now().UnixNano())
	feed, err := bucket.StartUprFeed(name, 0)
	if err != nil {
		log.Print(" Failed to start stream ", err)
		return
	}

	// request stream for all vbuckets
	for i := 0; i < vbcount; i++ {
		err := feed.UprRequestStream(
			uint16(i) /*vbno*/, uint16(0) /*opaque*/, 0 /*flag*/, 0, /*vbuuid*/
			0 /*seqStart*/, 0xFFFFFFFFFFFFFFFF /*seqEnd*/, 0 /*snaps*/, 0)
		if err != nil {
			fmt.Printf("%s", err.Error())
		}
	}

	// observe the mutations from the channel.
	var event *memcached.UprEvent
	var mutations = 0
	//var callOnce bool
loop:
	for {
		select {
		case e, ok := <-feed.C:
			if !ok {
				break loop
			} else {
				event = e
			}
		case <-time.After(time.Second):
			break loop
		}
		if event.Opcode == gomemcached.UPR_MUTATION {
			//log.Printf(" got mutation %s", e.Value)
			mutations += 1
		}

		if event.Opcode == gomemcached.UPR_STREAMEND {
			log.Printf(" Received Stream end for vbucket %d", event.VBucket)
		}

		if mutations%100000 == 0 && mutations != 0 {
			log.Printf(" received %d mutations ", mutations)
		}

		//e.Release()
	}

	// close stream for all vbuckets
	for i := 0; i < vbcount; i++ {
		err := feed.UprCloseStream(
			uint16(i) /*vbno*/, uint16(0))
		if err != nil {
			fmt.Printf("%s", err.Error())
		}
	}

	feed.Close()
	log.Printf("Mutation count %d", mutations)

}

func addKVset(b *couchbase.Bucket, count int) {
	for i := 0; i < count; i++ {
		key := fmt.Sprintf("key%v", i+1000000)
		val_len := rand.Intn(10*1024) + rand.Intn(10*1024)
		value := fmt.Sprintf("This is a test key %d", val_len)
		err := b.Set(key, 0, value)
		if err != nil {
			panic(err)
		}

		if i%1000000 == 0 {
			fmt.Printf("\n Added %d keys", i)
		}
	}
}
