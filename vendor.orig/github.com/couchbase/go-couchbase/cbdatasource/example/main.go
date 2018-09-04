//  Copyright (c) 2014 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the
//  License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing,
//  software distributed under the License is distributed on an "AS
//  IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
//  express or implied. See the License for the specific language
//  governing permissions and limitations under the License.

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"reflect"
	"runtime"
	"runtime/pprof"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/couchbase/go-couchbase"
	"github.com/couchbase/go-couchbase/cbdatasource"
	"github.com/couchbase/gomemcached"
)

// Simple, memory-only sample program that uses the cbdatasource API's
// to get data from a couchbase cluster using DCP.

var verbose = flag.Int("verbose", 1,
	"verbose'ness of logging, where 0 is no logging")

var serverURL = flag.String("serverURL", "http://localhost:8091",
	"couchbase server URL")
var poolName = flag.String("poolName", "default",
	"pool name")
var bucketName = flag.String("bucketName", "default",
	"bucket name")
var bucketUUID = flag.String("bucketUUID", "",
	"bucket UUID")
var vbucketIds = flag.String("vbucketIds", "",
	"comma separated vbucket id numbers; defaults to all vbucket id's")
var authUser = flag.String("authUser", "",
	"auth user name (probably same as bucketName)")
var authPswd = flag.String("authPswd", "",
	"auth password")

var optionClusterManagerBackoffFactor = flag.Float64("optionClusterManagerBackoffFactor", 1.5,
	"factor to increase sleep time between retries to cluster manager")
var optionClusterManagerSleepInitMS = flag.Int("optionClusterManagerSleepInitMS", 100,
	"initial sleep time for retries to cluster manager")
var optionClusterManagerSleepMaxMS = flag.Int("optionClusterManagerSleepMaxMS", 1000,
	"max sleep time for retries to cluster manager")

var optionDataManagerBackoffFactor = flag.Float64("optionDataManagerBackoffFactor", 1.5,
	"factor to increase sleep time between retries to data manager")
var optionDataManagerSleepInitMS = flag.Int("optionDataManagerSleepInitMS", 100,
	"initial sleep time for retries to data manager")
var optionDataManagerSleepMaxMS = flag.Int("optionDataManagerSleepMaxMS", 1000,
	"max sleep time for retries to data manager")

var optionFeedBufferSizeBytes = flag.Int("optionFeedBufferSizeBytes", 20000000,
	"buffer size for flow control")
var optionFeedBufferAckThreshold = flag.Float64("optionFeedBufferAckThreshold", 0.2,
	"percent (0-to-1.0) of buffer size before sending a flow control buffer-ack")

var bds cbdatasource.BucketDataSource

func main() {
	flag.Parse()

	go dumpOnSignalForPlatform()

	if *verbose > 0 {
		log.Printf("%s started", os.Args[0])
		flag.VisitAll(func(f *flag.Flag) { log.Printf("  -%s=%s\n", f.Name, f.Value) })
		log.Printf("  GOMAXPROCS=%d", runtime.GOMAXPROCS(-1))
	}

	serverURLs := []string{*serverURL}

	vbucketIdsArr := []uint16(nil) // A nil means get all the vbuckets.
	if *vbucketIds != "" {
		vbucketIdsArr = []uint16{}
		for _, vbucketIdStr := range strings.Split(*vbucketIds, ",") {
			if vbucketIdStr != "" {
				vbucketId, err := strconv.Atoi(vbucketIdStr)
				if err != nil {
					log.Fatalf("error: could not parse vbucketId: %s", vbucketIdStr)
				}
				vbucketIdsArr = append(vbucketIdsArr, uint16(vbucketId))
			}
		}
		if len(vbucketIdsArr) <= 0 {
			vbucketIdsArr = nil
		}
	}

	if *optionFeedBufferSizeBytes < 0 {
		log.Fatalf("error: optionFeedBufferSizeBytes must be >= 0")
	}

	options := &cbdatasource.BucketDataSourceOptions{
		ClusterManagerBackoffFactor: float32(*optionClusterManagerBackoffFactor),
		ClusterManagerSleepInitMS:   *optionClusterManagerSleepInitMS,
		ClusterManagerSleepMaxMS:    *optionClusterManagerSleepMaxMS,

		DataManagerBackoffFactor: float32(*optionDataManagerBackoffFactor),
		DataManagerSleepInitMS:   *optionDataManagerSleepInitMS,
		DataManagerSleepMaxMS:    *optionDataManagerSleepMaxMS,

		FeedBufferSizeBytes:    uint32(*optionFeedBufferSizeBytes),
		FeedBufferAckThreshold: float32(*optionFeedBufferAckThreshold),
	}

	var auth couchbase.AuthHandler = nil
	if *authUser != "" {
		auth = &authUserPswd{}
	}

	receiver := &ExampleReceiver{}

	var err error

	bds, err = cbdatasource.NewBucketDataSource(serverURLs,
		*poolName, *bucketName, *bucketUUID, vbucketIdsArr, auth, receiver, options)
	if err != nil {
		log.Fatalf(fmt.Sprintf("error: NewBucketDataSource, err: %v", err))
	}

	if err = bds.Start(); err != nil {
		log.Fatalf(fmt.Sprintf("error: Start, err: %v", err))
	}

	if *verbose > 0 {
		log.Printf("started bucket data source: %v", bds)
	}

	for {
		time.Sleep(1000 * time.Millisecond)
		reportStats(bds, false)
	}
}

type authUserPswd struct{}

func (a authUserPswd) GetCredentials() (string, string, string) {
	return *authUser, *authPswd, ""
}

// ----------------------------------------------------------------

type ExampleReceiver struct {
	m sync.Mutex

	seqs map[uint16]uint64 // To track max seq #'s we received per vbucketId.
	meta map[uint16][]byte // To track metadata blob's per vbucketId.
}

func (r *ExampleReceiver) OnError(err error) {
	if *verbose > 0 {
		log.Printf("error: %v", err)
	}
	reportStats(bds, true)
}

func (r *ExampleReceiver) DataUpdate(vbucketId uint16, key []byte, seq uint64,
	req *gomemcached.MCRequest) error {
	if *verbose > 1 {
		log.Printf("data-update: vbucketId: %d, key: %s, seq: %x, req: %#v",
			vbucketId, key, seq, req)
	}
	r.updateSeq(vbucketId, seq)
	return nil
}

func (r *ExampleReceiver) DataDelete(vbucketId uint16, key []byte, seq uint64,
	req *gomemcached.MCRequest) error {
	if *verbose > 1 {
		log.Printf("data-delete: vbucketId: %d, key: %s, seq: %x, req: %#v",
			vbucketId, key, seq, req)
	}
	r.updateSeq(vbucketId, seq)
	return nil
}

func (r *ExampleReceiver) SnapshotStart(vbucketId uint16,
	snapStart, snapEnd uint64, snapType uint32) error {
	if *verbose > 1 {
		log.Printf("snapshot-start: vbucketId: %d, snapStart: %x, snapEnd: %x, snapType: %x",
			vbucketId, snapStart, snapEnd, snapType)
	}
	return nil
}

func (r *ExampleReceiver) SetMetaData(vbucketId uint16, value []byte) error {
	if *verbose > 1 {
		log.Printf("set-metadata: vbucketId: %d, value: %s", vbucketId, value)
	}

	r.m.Lock()
	defer r.m.Unlock()

	if r.meta == nil {
		r.meta = make(map[uint16][]byte)
	}
	r.meta[vbucketId] = value

	return nil
}

func (r *ExampleReceiver) GetMetaData(vbucketId uint16) (
	value []byte, lastSeq uint64, err error) {
	if *verbose > 1 {
		log.Printf("get-metadata: vbucketId: %d", vbucketId)
	}

	r.m.Lock()
	defer r.m.Unlock()

	value = []byte(nil)
	if r.meta != nil {
		value = r.meta[vbucketId]
	}

	if r.seqs != nil {
		lastSeq = r.seqs[vbucketId]
	}

	return value, lastSeq, nil
}

func (r *ExampleReceiver) Rollback(vbucketId uint16, rollbackSeq uint64) error {
	if *verbose > 0 {
		log.Printf("rollback: vbucketId: %d, rollbackSeq: %x", vbucketId, rollbackSeq)
	}

	return fmt.Errorf("unimpl-rollback")
}

// ----------------------------------------------------------------

func (r *ExampleReceiver) updateSeq(vbucketId uint16, seq uint64) {
	r.m.Lock()
	defer r.m.Unlock()

	if r.seqs == nil {
		r.seqs = make(map[uint16]uint64)
	}
	if r.seqs[vbucketId] < seq {
		r.seqs[vbucketId] = seq // Remember the max seq for GetMetaData().
	}
}

// ----------------------------------------------------------------

var mutexStats sync.Mutex
var lastStats = &cbdatasource.BucketDataSourceStats{}
var currStats = &cbdatasource.BucketDataSourceStats{}

func reportStats(b cbdatasource.BucketDataSource, force bool) {
	if *verbose <= 0 {
		return
	}

	mutexStats.Lock()
	defer mutexStats.Unlock()

	b.Stats(currStats)
	if force || !reflect.DeepEqual(lastStats, currStats) {
		buf, err := json.Marshal(currStats)
		if err == nil {
			log.Printf("%s", string(buf))
		}
		lastStats, currStats = currStats, lastStats
	}
}

func dumpOnSignal(signals ...os.Signal) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, signals...)
	for _ = range c {
		reportStats(bds, true)

		log.Printf("dump: goroutine...")
		pprof.Lookup("goroutine").WriteTo(os.Stderr, 1)
		log.Printf("dump: heap...")
		pprof.Lookup("heap").WriteTo(os.Stderr, 1)
	}
}
