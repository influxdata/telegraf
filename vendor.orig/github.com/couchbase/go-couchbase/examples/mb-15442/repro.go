package main

import (
	"database/sql"
	"flag"
	"fmt"
	"github.com/couchbase/go-couchbase"
	_ "github.com/couchbaselabs/go_n1ql"
	"log"
)

var serverURL = flag.String("serverURL", "http://localhost:9000",
	"couchbase server URL")
var poolName = flag.String("poolName", "default",
	"pool name")
var bucketName = flag.String("bucketName", "default",
	"bucket name")

func main() {

	flag.Parse()

	client, err := couchbase.Connect(*serverURL)
	if err != nil {
		log.Printf("Connect failed %v", err)
		return
	}

	cbpool, err := client.GetPool("default")
	if err != nil {
		log.Printf("Failed to connect to default pool %v", err)
		return
	}

	var cbbucket *couchbase.Bucket
	cbbucket, err = cbpool.GetBucket(*bucketName)

	if err != nil {
		log.Printf("Failed to connect to bucket %v", err)
		return
	}

	performOp(cbbucket)

}

func performOp(b *couchbase.Bucket) {

	key := fmt.Sprintf("odwalla-juice1")
	odwalla1 := map[string]interface{}{"type": "juice"}
	log.Printf(" setting key %v value %v", key, odwalla1)
	err := b.SetWithMeta(key, 0x1000001, 0, odwalla1)
	if err != nil {
		log.Printf("set failed error %v", err)
		return
	}

	_, flags, _, err := b.GetsRaw("odwalla-juice1")
	if err != nil {
		log.Fatal(err)
	}

	if flags != 0x1000001 {
		log.Fatal("Flag mismatch %v", flags)
	}

	n1ql, err := sql.Open("n1ql", "localhost:8093")
	if err != nil {
		log.Fatal(err)
	}

	result, err := n1ql.Exec("UPDATE default USE KEYS \"odwalla-juice1\" SET type=\"product-juice\" RETURNING default.type")

	if err != nil {
		log.Fatal(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Rows affected %d", rowsAffected)

	_, flags, _, err = b.GetsRaw("odwalla-juice1")
	if err != nil {
		log.Fatal(err)
	}

	if flags != 0x1000001 {
		log.Fatal("Flag mismatch %v", flags)
	}
}
