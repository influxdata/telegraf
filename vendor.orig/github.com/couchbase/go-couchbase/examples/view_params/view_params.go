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

var updateInterval = flag.Int("updateInterval", 5000,
	"min update interval ms (int)")
var updateMinChanges = flag.Int("updateMinChanges", 5000,
	"min update changes (int)")

func main() {

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr,
			"%v [flags] http://user:pass@host:8091/\n\n",
			os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExample : ./view_params -updateInterval=7 -updateMinChanges=4000 http://Administrator:asdasd@localhost:9000\n")
		os.Exit(64)
	}

	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
	}

	u, err := url.Parse(flag.Arg(0))
	mf(err, "parse")

	params := map[string]interface{}{"updateInterval": *updateInterval, "updateMinChanges": *updateMinChanges}

	viewParams, err := couchbase.SetViewUpdateParams(u.String(), params)
	if err != nil {
		log.Fatal(" Failed ", err)
	}

	log.Printf("Returned view params %v", viewParams)

}
