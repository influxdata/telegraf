package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/couchbase/go-couchbase"
)

func maybeFatal(e error, f string, args ...interface{}) {
	if e != nil {
		fmt.Fprintf(os.Stderr, f, args...)
		os.Exit(64)
	}
}

func main() {
	cbServ := flag.String("couchbase", "http://localhost:8091/",
		"URL to couchbase")
	cbBucket := flag.String("bucket", "default", "couchbase bucket")
	includeExt := flag.Bool("includeExt", false, "include file extension in document ID")
	verbose := flag.Bool("v", false, "verbose output")

	flag.Parse()

	b, err := couchbase.GetBucket(*cbServ, "default", *cbBucket)
	maybeFatal(err, "Error connecting to couchbase: %v\n", err)

	for _, filename := range flag.Args() {
		key := pathToID(filename, *includeExt)
		bytes, err := ioutil.ReadFile(filename)
		maybeFatal(err, "Error reading file contents: %v\n", err)
		b.SetRaw(key, 0, bytes)
		if *verbose {
			fmt.Printf("Loaded %s to key %s\n", filename, key)
		}
	}
	if *verbose {
		fmt.Printf("Loaded %d documents into bucket %s\n", len(flag.Args()), *cbBucket)
	}
}

func pathToID(p string, includeExt bool) string {
	_, file := path.Split(p)
	if includeExt {
		return file
	}
	ext := path.Ext(file)
	return file[0 : len(file)-len(ext)]
}
