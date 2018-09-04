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

	u, err := url.Parse(flag.Arg(0))
	mf(err, "parse")

	if *bname == "" && u.User != nil {
		*bname = u.User.Username()
	}

	c, err := couchbase.Connect(u.String())
	mf(err, "connect - "+u.String())

	p, err := c.GetPool("default")
	mf(err, "pool")

	b, err := p.GetBucket(*bname)
	mf(err, "bucket")

	err = b.Set(",k", 90, map[string]interface{}{"x": 1})
	mf(err, "set")

	ob := map[string]interface{}{}
	err = b.Get(",k", &ob)
	mf(err, "get")
}
