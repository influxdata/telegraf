package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/couchbase/go-couchbase"
)

func mf(err error, msg string) {
	if err != nil {
		log.Fatalf("%v: %v", msg, err)
	}
}

func main() {

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

	u, err := url.Parse(flag.Arg(0))
	mf(err, "parse")

	bucketInfo, err := couchbase.GetBucketList(u.String())
	fmt.Printf("List of buckets and password %v", bucketInfo)

	//connect to a gamesim-sample
	client, err := couchbase.Connect(u.String())
	if err != nil {
		fmt.Printf("Connect failed %v", err)
		return
	}

	cbpool, err := client.GetPool("default")
	if err != nil {
		fmt.Printf("Failed to connect to default pool %v", err)
		return
	}

	for _, bi := range bucketInfo {
		var cbbucket *couchbase.Bucket

		cbbucket, err = cbpool.GetBucketWithAuth(bi.Name, bi.Name, bi.Password)
		if err != nil {
			fmt.Printf("Failed to connect to bucket %s %v", bi.Name, err)
			return
		}

		err = cbbucket.Set("k1", 0, "value")
		if err != nil {
			fmt.Printf("set failed error %v", err)
			return
		}

	}

}
