// Tool receives raw events from go-couchbase UPR client.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/couchbase/go-couchbase"
	mcd "github.com/couchbase/gomemcached"
	mc "github.com/couchbase/gomemcached/client"
)

var options struct {
	buckets    []string // buckets to connect with
	maxVbno    int      // maximum number of vbuckets
	stats      int      // periodic timeout(ms) to print stats, 0 will disable
	duration   int
	printflogs bool
}

var done = make(chan bool, 16)
var rch = make(chan []interface{}, 10000)

func argParse() string {
	var buckets string

	flag.StringVar(&buckets, "buckets", "default",
		"buckets to listen")
	flag.IntVar(&options.maxVbno, "maxvb", 1024,
		"maximum number of vbuckets")
	flag.IntVar(&options.stats, "stats", 1000,
		"periodic timeout in mS, to print statistics, `0` will disable stats")
	flag.IntVar(&options.duration, "duration", 3000,
		"receive mutations till duration milliseconds.")
	flag.BoolVar(&options.printflogs, "flogs", false,
		"display failover logs")

	flag.Parse()

	options.buckets = strings.Split(buckets, ",")
	args := flag.Args()
	if len(args) < 1 {
		usage()
		os.Exit(1)
	}
	return args[0]
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage : %s [OPTIONS] <cluster-addr> \n", os.Args[0])
	flag.PrintDefaults()
}

func main() {
	cluster := argParse()
	ch := make(chan *couchbase.UprFeed, 10)
	for _, bucket := range options.buckets {
		go startBucket(cluster, bucket, ch)
	}
	receive(ch)
}

func startBucket(cluster, bucketn string, ch chan *couchbase.UprFeed) int {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("%s:\n%s\n", r, debug.Stack())
		}
	}()

	log.Printf("Connecting with %q\n", bucketn)
	b, err := ConnectBucket(cluster, "default", bucketn)
	mf(err, "bucket")

	uprFeed, err := b.StartUprFeed("rawupr", uint32(0))
	mf(err, "- upr")

	vbnos := listOfVbnos(options.maxVbno)

	flogs, err := b.GetFailoverLogs(vbnos)
	mf(err, "- upr failoverlogs")

	if options.printflogs {
		printFlogs(vbnos, flogs)
	}

	ch <- uprFeed

	go startUpr(uprFeed, flogs)

	for {
		e, ok := <-uprFeed.C
		if ok == false {
			log.Printf("Closing for bucket %q\n", bucketn)
		}
		rch <- []interface{}{bucketn, e}
	}
}

func startUpr(uprFeed *couchbase.UprFeed, flogs couchbase.FailoverLog) {
	start, end := uint64(0), uint64(0xFFFFFFFFFFFFFFFF)
	snapStart, snapEnd := uint64(0), uint64(0)
	for vbno, flog := range flogs {
		x := flog[len(flog)-1] // map[uint16][][2]uint64
		opaque, flags, vbuuid := uint16(0), uint32(0), x[0]
		err := uprFeed.UprRequestStream(
			vbno, opaque, flags, vbuuid, start, end, snapStart, snapEnd)
		mf(err, fmt.Sprintf("stream-req for %v failed", vbno))
	}
}

func endUpr(uprFeed *couchbase.UprFeed, vbnos []uint16) error {
	for _, vbno := range vbnos {
		if err := uprFeed.UprCloseStream(vbno, uint16(0)); err != nil {
			mf(err, "- UprCloseStream()")
			return err
		}
	}
	return nil
}

func mf(err error, msg string) {
	if err != nil {
		log.Fatalf("%v: %v", msg, err)
	}
}

func receive(ch chan *couchbase.UprFeed) {
	// bucket -> Opcode -> #count
	counts := make(map[string]map[mcd.CommandCode]int)

	var tick <-chan time.Time
	if options.stats > 0 {
		tick = time.Tick(time.Millisecond * time.Duration(options.stats))
	}

	finTimeout := time.After(time.Millisecond * time.Duration(options.duration))
	uprFeeds := make([]*couchbase.UprFeed, 0)
loop:
	for {
		select {
		case uprFeed := <-ch:
			uprFeeds = append(uprFeeds, uprFeed)

		case msg, ok := <-rch:
			if ok == false {
				break loop
			}
			bucket, e := msg[0].(string), msg[1].(*mc.UprEvent)
			if _, ok := counts[bucket]; !ok {
				counts[bucket] = make(map[mcd.CommandCode]int)
			}
			if _, ok := counts[bucket][e.Opcode]; !ok {
				counts[bucket][e.Opcode] = 0
			}
			counts[bucket][e.Opcode]++

		case <-tick:
			for bucket, m := range counts {
				log.Printf("%q %s\n", bucket, sprintCounts(m))
			}

		case <-finTimeout:
			for _, uprFeed := range uprFeeds {
				endUpr(uprFeed, listOfVbnos(options.maxVbno))
			}
			break loop
		}
	}
	fmt.Println("sleep wait ....")
	time.Sleep(10000 * time.Millisecond)
}

func sprintCounts(counts map[mcd.CommandCode]int) string {
	line := ""
	for i := 0; i < 256; i++ {
		opcode := mcd.CommandCode(i)
		if n, ok := counts[opcode]; ok {
			line += fmt.Sprintf("%s:%v ", mcd.CommandNames[opcode], n)
		}
	}
	return strings.TrimRight(line, " ")
}

func listOfVbnos(maxVbno int) []uint16 {
	// list of vbuckets
	vbnos := make([]uint16, 0, maxVbno)
	for i := 0; i < maxVbno; i++ {
		vbnos = append(vbnos, uint16(i))
	}
	return vbnos
}

func printFlogs(vbnos []uint16, flogs couchbase.FailoverLog) {
	for i, vbno := range vbnos {
		log.Printf("Failover log for vbucket %v\n", vbno)
		log.Printf("   %#v\n", flogs[uint16(i)])
	}
}

func ConnectBucket(cluster, pooln, bucketn string) (*couchbase.Bucket, error) {
	couch, err := couchbase.Connect("http://" + cluster)
	if err != nil {
		return nil, err
	}
	pool, err := couch.GetPool(pooln)
	if err != nil {
		return nil, err
	}
	bucket, err := pool.GetBucket(bucketn)
	if err != nil {
		return nil, err
	}
	return bucket, err
}
