package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/couchbase/go-couchbase"
)

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "%v [flags] ddocname\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(64)
	}
}

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
	objName := flag.String("objname", "designDoc",
		"Name of the variable to create")
	flag.Parse()

	ddocName := flag.Arg(0)
	if ddocName == "" {
		fmt.Fprintf(os.Stderr, "No ddoc given\n")
		flag.Usage()
	}

	b, err := couchbase.GetBucket(*cbServ, "default", *cbBucket)
	maybeFatal(err, "Error connecting to couchbase: %v\n", err)

	j := json.RawMessage{}
	err = b.GetDDoc(ddocName, &j)
	maybeFatal(err, "Error getting ddoc: %v\n", err)

	buf := &bytes.Buffer{}
	err = json.Indent(buf, []byte(j), "", "  ")

	fmt.Printf("const %s = `%s`\n", *objName,
		strings.Replace(buf.String(), "`", "` + \"`\" + `", 0))
}
