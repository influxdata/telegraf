package main

import (
	"fmt"
	atomic "github.com/couchbase/go-couchbase/platform"
)

func main() {

	var someval atomic.AlignedInt64

	atomic.StoreInt64(&someval, int64(512))
	fmt.Printf(" Value of someval %v", someval)

	rval := atomic.LoadInt64(&someval)

	fmt.Printf(" Returned val %v", rval)
}
