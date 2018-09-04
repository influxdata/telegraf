package main

import (
	"flag"
	"fmt"
	"github.com/couchbase/cbauth"
	"github.com/couchbase/go-couchbase"
	"log"
	"net/url"
)

var serverURL = flag.String("serverURL", "http://localhost:9000",
	"couchbase server URL")
var poolName = flag.String("poolName", "default",
	"pool name")
var bucketName = flag.String("bucketName", "default",
	"bucket name")
var authUser = flag.String("authUser", "",
	"auth user name (probably same as bucketName)")
var authPswd = flag.String("authPswd", "",
	"auth password")

func main() {

	flag.Parse()
	/*
	   NOTE. This example requires the following environment variables to be set.

	   CBAUTH_REVRPC_URL

	   e.g

	   CBAUTH_REVRPC_URL="http://Administrator:asdasd@127.0.0.1:9000/_cbauth"

	*/

	url, err := url.Parse(*serverURL)
	if err != nil {
		log.Printf("Failed to parse url %v", err)
		return
	}

	hostPort := url.Host

	user, bucket_password, err := cbauth.GetHTTPServiceAuth(hostPort)
	if err != nil {
		log.Printf("Failed %v", err)
		return
	}

	log.Printf(" HTTP Servce username %s password %s", user, bucket_password)

	client, err := couchbase.ConnectWithAuthCreds(*serverURL, user, bucket_password)
	if err != nil {
		log.Printf("Connect failed %v", err)
		return
	}

	cbpool, err := client.GetPool("default")
	if err != nil {
		log.Printf("Failed to connect to default pool %v", err)
		return
	}

	mUser, mPassword, err := cbauth.GetMemcachedServiceAuth(hostPort)
	if err != nil {
		log.Printf(" failed %v", err)
		return
	}

	var cbbucket *couchbase.Bucket
	cbbucket, err = cbpool.GetBucketWithAuth(*bucketName, mUser, mPassword)

	if err != nil {
		log.Printf("Failed to connect to bucket %v", err)
		return
	}

	log.Printf(" Bucket name %s Bucket %v", *bucketName, cbbucket)

	err = cbbucket.Set("k1", 5, "value")
	if err != nil {
		log.Printf("set failed error %v", err)
		return
	}

	if *authUser != "" {
		creds, err := cbauth.Auth(*authUser, *authPswd)
		if err != nil {
			log.Printf(" failed %v", err)
			return
		}

		permission := fmt.Sprintf("cluster.bucket[%s].data!read", *bucketName)
		canAccess, err := creds.IsAllowed(permission)
		if err != nil {
			log.Printf(" error %v checking permission %v", err, permission)
		} else {
			log.Printf(" result of checking permission %v : %v", permission, canAccess)
		}
	}

}
