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

package cbdatasource

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
	"runtime"
	"sync"
	"testing"

	"github.com/couchbase/go-couchbase"
	"github.com/couchbase/gomemcached"
	"github.com/couchbase/gomemcached/client"
)

// NOTE: Some of the tests are single-threaded, and will be skipped
// when run with GOMAXPROCS > 1.  To run them all, try using...
//
//    go test -cpu 1
//
func checkMaxProcs1(t *testing.T) bool {
	if runtime.GOMAXPROCS(-1) > 1 {
		t.Skip("skipping test; GOMAXPROCS > 1, see checkMaxProcs1()")
		return false
	}
	return true
}

type TestBucket struct {
	uuid     string
	vbsm     *couchbase.VBucketServerMap
	numClose int
}

func (bw *TestBucket) Close() {
	bw.numClose++
}

func (bw *TestBucket) GetUUID() string {
	return bw.uuid
}

func (bw *TestBucket) VBServerMap() *couchbase.VBucketServerMap {
	return bw.vbsm
}

type TestMutation struct {
	delete    bool
	vbucketID uint16
	key       []byte
	seq       uint64
}

type TestReceiver struct {
	onErrorCh       chan []string
	onGetMetaDataCh chan []string

	m    sync.Mutex
	errs []error
	muts []*TestMutation
	meta map[uint16][]byte

	numSnapshotStarts int
	numSetMetaDatas   int
	numGetMetaDatas   int
	numRollbacks      int

	testName string
}

func (r *TestReceiver) OnError(err error) {
	r.m.Lock()
	defer r.m.Unlock()

	if r.onErrorCh != nil {
		r.onErrorCh <- []string{"TestReceiver.OnError"}
	}

	// fmt.Printf("  testName: %s: %v\n", r.testName, err)
	r.errs = append(r.errs, err)
}

func (r *TestReceiver) DataUpdate(vbucketID uint16, key []byte, seq uint64,
	req *gomemcached.MCRequest) error {
	r.m.Lock()
	defer r.m.Unlock()

	r.muts = append(r.muts, &TestMutation{
		delete:    false,
		vbucketID: vbucketID,
		key:       key,
		seq:       seq,
	})
	return nil
}

func (r *TestReceiver) DataDelete(vbucketID uint16, key []byte, seq uint64,
	req *gomemcached.MCRequest) error {
	r.m.Lock()
	defer r.m.Unlock()

	r.muts = append(r.muts, &TestMutation{
		delete:    true,
		vbucketID: vbucketID,
		key:       key,
		seq:       seq,
	})
	return nil
}

func (r *TestReceiver) SnapshotStart(vbucketID uint16,
	snapStart, snapEnd uint64, snapType uint32) error {
	r.numSnapshotStarts++
	return nil
}

func (r *TestReceiver) SetMetaData(vbucketID uint16, value []byte) error {
	r.m.Lock()
	defer r.m.Unlock()

	r.numSetMetaDatas++
	if r.meta == nil {
		r.meta = make(map[uint16][]byte)
	}
	r.meta[vbucketID] = value
	return nil
}

func (r *TestReceiver) GetMetaData(vbucketID uint16) (value []byte, lastSeq uint64, err error) {
	r.m.Lock()
	defer r.m.Unlock()

	if r.onGetMetaDataCh != nil {
		r.onGetMetaDataCh <- []string{"TestReceiver.GetMetaData"}
	}

	r.numGetMetaDatas++
	rv := []byte(nil)
	if r.meta != nil {
		rv = r.meta[vbucketID]
	}
	for i := len(r.muts) - 1; i >= 0; i = i - 1 {
		if r.muts[i].vbucketID == vbucketID {
			return rv, r.muts[i].seq, nil
		}
	}
	return rv, 0, nil
}

func (r *TestReceiver) Rollback(vbucketID uint16, rollbackSeq uint64) error {
	r.numRollbacks++
	return fmt.Errorf("bad-rollback")
}

// Implements ReadWriteCloser interface for fake networking.
type TestRWC struct {
	name      string
	numReads  int
	numWrites int
	readCh    chan RWReq
	writeCh   chan RWReq
}

func (c *TestRWC) Read(p []byte) (n int, err error) {
	c.numReads++
	if c.readCh != nil {
		resCh := make(chan RWRes)
		c.readCh <- RWReq{op: "read", buf: p, resCh: resCh}
		res := <-resCh
		return res.n, res.err
	}
	return 0, fmt.Errorf("fake-read-err")
}

func (c *TestRWC) Write(p []byte) (n int, err error) {
	c.numWrites++
	if c.writeCh != nil {
		resCh := make(chan RWRes)
		c.writeCh <- RWReq{op: "write", buf: p, resCh: resCh}
		res := <-resCh
		return res.n, res.err
	}
	return 0, fmt.Errorf("fake-write-err")
}

func (c *TestRWC) Close() error {
	return nil
}

type RWReq struct {
	op    string
	buf   []byte
	resCh chan RWRes
}

type RWRes struct {
	n   int
	err error
}

// ------------------------------------------------------

func TestExponentialBackoffLoop(t *testing.T) {
	called := 0
	ExponentialBackoffLoop("test", func() int {
		called++
		return -1
	}, 0, 0, 0)
	if called != 1 {
		t.Errorf("expected 1 call")
	}

	called = 0
	ExponentialBackoffLoop("test", func() int {
		called++
		if called <= 1 {
			return 1
		}
		return -1
	}, 0, 0, 0)
	if called != 2 {
		t.Errorf("expected 2 calls")
	}

	called = 0
	ExponentialBackoffLoop("test", func() int {
		called++
		if called == 1 {
			return 1
		}
		if called == 2 {
			return 0
		}
		return -1
	}, 0, 0, 0)
	if called != 3 {
		t.Errorf("expected 2 calls")
	}

	called = 0
	ExponentialBackoffLoop("test", func() int {
		called++
		if called == 1 {
			return 1
		}
		if called == 2 {
			return 0
		}
		return -1
	}, 1, 100000.0, 1)
	if called != 3 {
		t.Errorf("expected 2 calls")
	}
}

// ------------------------------------------------------

func TestParseFailOverLog(t *testing.T) {
	f, err := ParseFailOverLog([]byte("hi"))
	if err == nil || f != nil {
		t.Errorf("expected ParseFailOverLog nil to fail")
	}
}

func TestNewBucketDataSource(t *testing.T) {
	serverURLs := []string(nil)
	bucketUUID := ""
	vbucketIDs := []uint16(nil)
	var auth couchbase.AuthHandler
	var receiver Receiver
	var options *BucketDataSourceOptions

	bds, err := NewBucketDataSource(serverURLs, "poolName", "bucketName", bucketUUID,
		vbucketIDs, auth, receiver, options)
	if err == nil || bds != nil {
		t.Errorf("expected err")
	}

	serverURLs = []string{"foo"}
	bucketUUID = ""
	vbucketIDs = []uint16(nil)
	bds, err = NewBucketDataSource(serverURLs, "poolName", "bucketName", bucketUUID,
		vbucketIDs, auth, receiver, options)
	if err == nil || bds != nil {
		t.Errorf("expected err")
	}

	poolName := ""
	bds, err = NewBucketDataSource(serverURLs, poolName, "bucketName", bucketUUID,
		vbucketIDs, auth, receiver, options)
	if err == nil || bds != nil {
		t.Errorf("expected err")
	}

	bucketName := ""
	bds, err = NewBucketDataSource(serverURLs, "poolName", bucketName, bucketUUID,
		vbucketIDs, auth, receiver, options)
	if err == nil || bds != nil {
		t.Errorf("expected err")
	}

	receiver = &TestReceiver{testName: "TestNewBucketDataSource"}
	bds, err = NewBucketDataSource(serverURLs, "poolName", "bucketName", bucketUUID,
		vbucketIDs, auth, receiver, options)
	if err != nil || bds == nil {
		t.Errorf("expected no err")
	}
	bdss := &BucketDataSourceStats{}
	err = bds.Stats(bdss)
	if err != nil {
		t.Errorf("expected no err on Stats()")
	}
	if !reflect.DeepEqual(bdss, &BucketDataSourceStats{}) {
		t.Errorf("expected zeroed Stats")
	}
}

func TestBucketDataSourceStatsAtomicCopyTo(t *testing.T) {
	b := &BucketDataSourceStats{
		TotSetVBucketMetaDataMarshalErr: 0x1f2f3f4f00001111,
	}
	x := &BucketDataSourceStats{}
	b.AtomicCopyTo(x, nil)
	if x.TotSetVBucketMetaDataMarshalErr != b.TotSetVBucketMetaDataMarshalErr {
		t.Errorf("expected copy to work, x: %#v vs b: %#v", x, b)
	}
	if x.TotWorkerReceiveOk != b.TotWorkerReceiveOk {
		t.Errorf("expected copy to work, x: %#v vs b: %#v", x, b)
	}
}

func TestImmediateStartCloseMAXPROCS1(t *testing.T) {
	if !checkMaxProcs1(t) {
		return
	}

	connectBucket := func(serverURL, poolName, bucketName string,
		auth couchbase.AuthHandler) (Bucket, error) {
		return nil, fmt.Errorf("fake connectBucket err")
	}

	connect := func(protocol, dest string) (*memcached.Client, error) {
		if protocol != "tcp" || dest != "serverA" {
			t.Errorf("unexpected connect, protocol: %s, dest: %s", protocol, dest)
		}
		return nil, fmt.Errorf("fake connect err")
	}

	serverURLs := []string{"serverA"}
	bucketUUID := ""
	vbucketIDs := []uint16(nil)
	var auth couchbase.AuthHandler
	receiver := &TestReceiver{testName: "TestImmediateStartClose"}
	options := &BucketDataSourceOptions{
		ConnectBucket: connectBucket,
		Connect:       connect,
	}

	bds, err := NewBucketDataSource(serverURLs, "poolName", "bucketName", bucketUUID,
		vbucketIDs, auth, receiver, options)
	if err != nil || bds == nil {
		t.Errorf("expected no err, got err: %v", err)
	}

	err = bds.Close()
	if err == nil {
		t.Errorf("expected err on Close before Start")
	}

	// ------------------------------------------------------------
	err = bds.Start()
	if err != nil {
		t.Errorf("expected no err on Start")
	}
	bdss := &BucketDataSourceStats{}
	err = bds.Stats(bdss)
	if err != nil {
		t.Errorf("expected no err on Stats()")
	}
	if !reflect.DeepEqual(bdss, &BucketDataSourceStats{
		TotStart: 1,
	}) {
		t.Errorf("expected same stats")
	}

	// ------------------------------------------------------------
	err = bds.Start()
	if err == nil {
		t.Errorf("expected err on re-Start")
	}
	err = bds.Stats(bdss)
	if err != nil {
		t.Errorf("expected no err on Stats()")
	}
	if !reflect.DeepEqual(bdss, &BucketDataSourceStats{
		TotStart: 2,
	}) {
		t.Errorf("expected same stats")
	}

	// ------------------------------------------------------------
	err = bds.Close()
	if err != nil {
		t.Errorf("expected no err on Close")
	}
	err = bds.Stats(bdss)
	if err != nil {
		t.Errorf("expected no err on Stats()")
	}
	if !reflect.DeepEqual(bdss, &BucketDataSourceStats{
		TotStart:                  2,
		TotRefreshCluster:         1,
		TotRefreshClusterDone:     1,
		TotRefreshWorkersDone:     1,
		TotRefreshWorkersLoopDone: 1,
	}) {
		t.Errorf("expected same stats, got: %#v", bdss)
	}

	// ------------------------------------------------------------
	err = bds.Close()
	if err == nil {
		t.Errorf("expected err on re-Close")
	}
}

func TestErrOnConnectBucketMAXPROCS1(t *testing.T) {
	if !checkMaxProcs1(t) {
		return
	}

	theErr := fmt.Errorf("fake connectBucket err")

	connectBucket := func(serverURL, poolName, bucketName string,
		auth couchbase.AuthHandler) (Bucket, error) {
		return nil, theErr
	}

	connect := func(protocol, dest string) (*memcached.Client, error) {
		if protocol != "tcp" || dest != "serverA" {
			t.Errorf("unexpected connect, protocol: %s, dest: %s", protocol, dest)
		}
		return nil, fmt.Errorf("fake connect err")
	}

	serverURLs := []string{"serverA"}
	bucketUUID := ""
	vbucketIDs := []uint16(nil)
	var auth couchbase.AuthHandler
	receiver := &TestReceiver{testName: "TestImmediateStartClose"}
	options := &BucketDataSourceOptions{
		ConnectBucket: connectBucket,
		Connect:       connect,
	}

	bds, err := NewBucketDataSource(serverURLs, "poolName", "bucketName", bucketUUID,
		vbucketIDs, auth, receiver, options)
	if err != nil || bds == nil {
		t.Errorf("expected no err, got err: %v", err)
	}

	err = bds.Start()
	if err != nil {
		t.Errorf("expected no err on Start")
	}

	runtime.Gosched()

	err = bds.Close()
	if err != nil {
		t.Errorf("expected no err on Close")
	}

	if len(receiver.errs) < 1 {
		t.Errorf("expected at least 1 err due to err on connectBucket")
	}
	if receiver.errs[0] != theErr {
		t.Errorf("expected first err due to err on connectBucket")
	}
}

func TestWrongBucketUUIDMAXPROCS1(t *testing.T) {
	if !checkMaxProcs1(t) {
		return
	}

	connectBucket := func(serverURL, poolName, bucketName string,
		auth couchbase.AuthHandler) (Bucket, error) {
		return &TestBucket{vbsm: &couchbase.VBucketServerMap{}}, nil
	}

	connect := func(protocol, dest string) (*memcached.Client, error) {
		if protocol != "tcp" || dest != "serverA" {
			t.Errorf("unexpected connect, protocol: %s, dest: %s", protocol, dest)
		}
		return nil, fmt.Errorf("fake connect err")
	}

	serverURLs := []string{"serverA"}
	bucketUUID := "not-a-good-uuid"
	vbucketIDs := []uint16(nil)
	var auth couchbase.AuthHandler
	receiver := &TestReceiver{testName: "TestImmediateStartClose"}
	options := &BucketDataSourceOptions{
		ConnectBucket: connectBucket,
		Connect:       connect,
	}

	bds, err := NewBucketDataSource(serverURLs, "poolName", "bucketName", bucketUUID,
		vbucketIDs, auth, receiver, options)
	if err != nil || bds == nil {
		t.Errorf("expected no err, got err: %v", err)
	}

	err = bds.Start()
	if err != nil {
		t.Errorf("expected no err on Start")
	}

	runtime.Gosched()

	err = bds.Close()
	if err != nil {
		t.Errorf("expected no err on Close")
	}

	if len(receiver.errs) < 1 {
		t.Errorf("expected at least 1 err due to mixmatched bucketUUID")
	}
}

func TestBucketDataSourceStartNilVBSM(t *testing.T) {
	var connectBucketResult Bucket
	var connectBucketErr error
	var connectBucketCh chan []string
	var connectCh chan []string

	connectBucket := func(serverURL, poolName, bucketName string,
		auth couchbase.AuthHandler) (Bucket, error) {
		connectBucketCh <- []string{serverURL, poolName, bucketName}
		return connectBucketResult, connectBucketErr
	}

	connect := func(protocol, dest string) (*memcached.Client, error) {
		if protocol != "tcp" || dest != "serverA" {
			t.Errorf("unexpected connect, protocol: %s, dest: %s", protocol, dest)
		}
		connectCh <- []string{protocol, dest}
		return nil, fmt.Errorf("fake connect err")
	}

	serverURLs := []string{"serverA"}
	bucketUUID := ""
	vbucketIDs := []uint16(nil)
	var auth couchbase.AuthHandler
	receiver := &TestReceiver{testName: "TestNewBucketDataSource"}
	options := &BucketDataSourceOptions{
		ConnectBucket: connectBucket,
		Connect:       connect,
	}

	connectBucketResult = &TestBucket{
		uuid: bucketUUID,
		vbsm: nil,
	}
	connectBucketErr = nil
	connectBucketCh = make(chan []string)

	bds, err := NewBucketDataSource(serverURLs, "poolName", "bucketName", bucketUUID,
		vbucketIDs, auth, receiver, options)
	if err != nil || bds == nil {
		t.Errorf("expected no err, got err: %v", err)
	}
	err = bds.Start()
	if err != nil {
		t.Errorf("expected no-err on Start()")
	}
	c := <-connectBucketCh
	if !reflect.DeepEqual(c, []string{"serverA", "poolName", "bucketName"}) {
		t.Errorf("expected connectBucket params")
	}
	select {
	case c := <-connectCh:
		t.Errorf("expected no connect due to nil vbsm, got: %#v", c)
	default:
	}
	err = bds.Close()
	if err != nil {
		t.Errorf("expected clean Close(), got err: %v", err)
	}
	if len(receiver.errs) < 1 {
		t.Errorf("expected connect err")
	}
}

func TestConnectError(t *testing.T) {
	var connectBucketResult Bucket
	var connectBucketErr error
	var connectBucketCh chan []string
	var connectCh chan []string

	connectBucket := func(serverURL, poolName, bucketName string,
		auth couchbase.AuthHandler) (Bucket, error) {
		connectBucketCh <- []string{serverURL, poolName, bucketName}
		return connectBucketResult, connectBucketErr
	}

	connect := func(protocol, dest string) (*memcached.Client, error) {
		if protocol != "tcp" || dest != "serverA" {
			t.Errorf("unexpected connect, protocol: %s, dest: %s", protocol, dest)
		}
		connectCh <- []string{protocol, dest}
		return nil, fmt.Errorf("fake-connect-error, protocol: %s, dest: %s",
			protocol, dest)
	}

	serverURLs := []string{"serverA"}
	bucketUUID := ""
	vbucketIDs := []uint16{0, 1, 2, 3}
	var auth couchbase.AuthHandler
	receiver := &TestReceiver{
		testName:  "TestBucketDataSourceStartVBSM",
		onErrorCh: make(chan []string),
	}
	options := &BucketDataSourceOptions{
		ConnectBucket: connectBucket,
		Connect:       connect,
	}

	connectBucketResult = &TestBucket{
		uuid: bucketUUID,
		vbsm: &couchbase.VBucketServerMap{
			ServerList: []string{"serverA"},
			VBucketMap: [][]int{
				[]int{0},
				[]int{0},
				[]int{0},
				[]int{0},
			},
		},
	}
	connectBucketErr = nil
	connectBucketCh = make(chan []string)
	connectCh = make(chan []string)

	bds, err := NewBucketDataSource(serverURLs, "poolName", "bucketName", bucketUUID,
		vbucketIDs, auth, receiver, options)
	if err != nil || bds == nil {
		t.Errorf("expected no err, got err: %v", err)
	}
	err = bds.Start()
	if err != nil {
		t.Errorf("expected no-err on Start()")
	}
	c := <-connectBucketCh
	if !reflect.DeepEqual(c, []string{"serverA", "poolName", "bucketName"}) {
		t.Errorf("expected connectBucket params, got: %#v", c)
	}
	c = <-connectCh
	if !reflect.DeepEqual(c, []string{"tcp", "serverA"}) {
		t.Errorf("expected connect params, got: %#v", c)
	}
	e, eok := <-receiver.onErrorCh
	if e == nil || !eok {
		t.Errorf("expected receiver.onErrorCh")
	}
	err = bds.Close()
	if err != nil {
		t.Errorf("expected clean Close(), got err: %v", err)
	}

	receiver.m.Lock()
	defer receiver.m.Unlock()

	if len(receiver.errs) != 1 {
		t.Errorf("expected connect err, got: %d", len(receiver.errs))
	}
}

func TestConnThatAlwaysErrors(t *testing.T) {
	var lastRWCM sync.Mutex
	var lastRWC *TestRWC

	rwcWriteCh := make(chan RWReq)

	newFakeConn := func(dest string) io.ReadWriteCloser {
		lastRWCM.Lock()
		defer lastRWCM.Unlock()

		lastRWC = &TestRWC{
			name:    dest,
			writeCh: rwcWriteCh,
		}
		return lastRWC
	}

	var connectBucketResult Bucket
	var connectBucketErr error
	var connectBucketCh chan []string
	var connectCh chan []string

	connectBucket := func(serverURL, poolName, bucketName string,
		auth couchbase.AuthHandler) (Bucket, error) {
		connectBucketCh <- []string{serverURL, poolName, bucketName}
		return connectBucketResult, connectBucketErr
	}

	connect := func(protocol, dest string) (*memcached.Client, error) {
		if protocol != "tcp" || dest != "serverA" {
			t.Errorf("unexpected connect, protocol: %s, dest: %s", protocol, dest)
		}
		rv, err := memcached.Wrap(newFakeConn(dest))
		connectCh <- []string{protocol, dest}
		return rv, err
	}

	serverURLs := []string{"serverA"}
	bucketUUID := ""
	vbucketIDs := []uint16{0, 1, 2, 3}
	var auth couchbase.AuthHandler
	receiver := &TestReceiver{testName: "TestBucketDataSourceStartVBSM"}
	options := &BucketDataSourceOptions{
		ConnectBucket: connectBucket,
		Connect:       connect,
	}

	connectBucketResult = &TestBucket{
		uuid: bucketUUID,
		vbsm: &couchbase.VBucketServerMap{
			ServerList: []string{"serverA"},
			VBucketMap: [][]int{
				[]int{0},
				[]int{0},
				[]int{0},
				[]int{0},
			},
		},
	}
	connectBucketErr = nil
	connectBucketCh = make(chan []string)
	connectCh = make(chan []string)

	bds, err := NewBucketDataSource(serverURLs, "poolName", "bucketName", bucketUUID,
		vbucketIDs, auth, receiver, options)
	if err != nil || bds == nil {
		t.Errorf("expected no err, got err: %v", err)
	}
	err = bds.Start()
	if err != nil {
		t.Errorf("expected no-err on Start()")
	}
	c := <-connectBucketCh
	if !reflect.DeepEqual(c, []string{"serverA", "poolName", "bucketName"}) {
		t.Errorf("expected connectBucket params, got: %#v", c)
	}
	c = <-connectCh
	if !reflect.DeepEqual(c, []string{"tcp", "serverA"}) {
		t.Errorf("expected connect params, got: %#v", c)
	}
	<-rwcWriteCh
	err = bds.Close()
	if err != nil {
		t.Errorf("expected clean Close(), got err: %v", err)
	}

	receiver.m.Lock()
	defer receiver.m.Unlock()

	// NOTE: This test passes on dev box (golang 1.3), but not on CI
	// system (drone.io, golang 1.2).
	// if len(receiver.errs) < 1 {
	//    t.Errorf("expected connect err")
	// }

	lastRWCM.Lock()
	rwc := lastRWC
	lastRWCM.Unlock()

	if rwc == nil {
		t.Errorf("expected a lastRWC")
	}
	if rwc.numReads != 0 {
		t.Errorf("expected a lastRWC with 0 reads, %#v", rwc)
	}
	if rwc.numWrites != 1 {
		t.Errorf("expected a lastRWC with 1 write, %#v", rwc)
	}
	if receiver.numSetMetaDatas != 0 {
		t.Errorf("expected 1 set-meta-data, %#v", receiver)
	}
	if receiver.numGetMetaDatas != 0 {
		t.Errorf("expected 1 get-meta-data, %#v", receiver)
	}
}

func TestUPROpenStreamReqMAXPROCS1(t *testing.T) {
	if !checkMaxProcs1(t) {
		return
	}

	var lastRWCM sync.Mutex
	var lastRWC *TestRWC

	newFakeConn := func(dest string) io.ReadWriteCloser {
		lastRWCM.Lock()
		defer lastRWCM.Unlock()

		lastRWC = &TestRWC{
			name:    dest,
			readCh:  make(chan RWReq),
			writeCh: make(chan RWReq),
		}
		return lastRWC
	}

	var connectBucketResult Bucket
	var connectBucketErr error
	var connectBucketCh chan []string
	var connectCh chan []string

	connectBucket := func(serverURL, poolName, bucketName string,
		auth couchbase.AuthHandler) (Bucket, error) {
		connectBucketCh <- []string{serverURL, poolName, bucketName}
		return connectBucketResult, connectBucketErr
	}

	connect := func(protocol, dest string) (*memcached.Client, error) {
		if protocol != "tcp" || dest != "serverA" {
			t.Errorf("unexpected connect, protocol: %s, dest: %s", protocol, dest)
		}
		rv, err := memcached.Wrap(newFakeConn(dest))
		connectCh <- []string{protocol, dest}
		return rv, err
	}

	serverURLs := []string{"serverA"}
	bucketUUID := ""
	vbucketIDs := []uint16{2}
	var auth couchbase.AuthHandler
	receiver := &TestReceiver{
		testName:        "TestBucketDataSourceStartVBSM",
		onGetMetaDataCh: make(chan []string),
	}
	options := &BucketDataSourceOptions{
		ConnectBucket: connectBucket,
		Connect:       connect,
	}

	connectBucketResult = &TestBucket{
		uuid: bucketUUID,
		vbsm: &couchbase.VBucketServerMap{
			ServerList: []string{"serverA"},
			VBucketMap: [][]int{
				[]int{0},
				[]int{0},
				[]int{0},
				[]int{0},
			},
		},
	}
	connectBucketErr = nil
	connectBucketCh = make(chan []string)
	connectCh = make(chan []string)

	bds, err := NewBucketDataSource(serverURLs, "poolName", "bucketName", bucketUUID,
		vbucketIDs, auth, receiver, options)
	if err != nil || bds == nil {
		t.Errorf("expected no err, got err: %v", err)
	}
	err = bds.Start()
	if err != nil {
		t.Errorf("expected no-err on Start()")
	}
	c := <-connectBucketCh
	if !reflect.DeepEqual(c, []string{"serverA", "poolName", "bucketName"}) {
		t.Errorf("expected connectBucket params, got: %#v", c)
	}
	c = <-connectCh
	if !reflect.DeepEqual(c, []string{"tcp", "serverA"}) {
		t.Errorf("expected connect params, got: %#v", c)
	}

	lastRWCM.Lock()
	rwc := lastRWC
	lastRWCM.Unlock()
	if rwc == nil {
		t.Errorf("expected a rwc")
	}

	// ------------------------------------------------------------
	reqW := <-rwc.writeCh
	req := &gomemcached.MCRequest{}
	n, err := req.Receive(bytes.NewReader(reqW.buf), nil)
	if err != nil || n < 24 {
		t.Errorf("expected read req to work, err: %v", err)
	}
	if req.Opcode != gomemcached.UPR_OPEN {
		t.Errorf("expected upr-open, got: %#v", req)
	}
	reqW.resCh <- RWRes{n: len(reqW.buf), err: nil}

	res := &gomemcached.MCResponse{
		Opcode: req.Opcode,
		Opaque: req.Opaque,
	}
	reqR := <-rwc.readCh
	copy(reqR.buf, res.HeaderBytes())
	reqR.resCh <- RWRes{n: len(reqR.buf), err: nil}

	// ------------------------------------------------------------
	e, eok := <-receiver.onGetMetaDataCh
	if e == nil || !eok {
		t.Errorf("expected receiver.onGetMetaDataCh")
	}

	// ------------------------------------------------------------
	receiver.m.Lock()
	if len(receiver.errs) != 0 {
		t.Errorf("expected 0 errs, got: %v", receiver.errs)
	}
	if receiver.numSetMetaDatas != 0 {
		t.Errorf("expected 0 numSetMetaDatas, got: %#v", receiver)
	}
	if receiver.numGetMetaDatas != 1 {
		t.Errorf("expected 1 numGetMetaDatas, got: %#v", receiver)
	}
	if receiver.meta[2] != nil {
		t.Errorf("expected nil meta for vbucket 2")
	}
	receiver.m.Unlock()

	reqW = <-rwc.writeCh
	req = &gomemcached.MCRequest{}
	n, err = req.Receive(bytes.NewReader(reqW.buf), nil)
	if err != nil || n < 24 {
		t.Errorf("expected read req to work, err: %v", err)
	}
	if req.Opcode != gomemcached.UPR_STREAMREQ {
		t.Errorf("expected upr-streamreq, got: %#v", req)
	}
	if req.VBucket != 2 {
		t.Errorf("expected vbucketID 2, got: %#v", req)
	}
	if req.Opaque != 2 {
		t.Errorf("expected opaque 2, got: %#v", req)
	}
	reqW.resCh <- RWRes{n: len(reqW.buf), err: nil}

	receiver.m.Lock()
	if len(receiver.errs) != 0 {
		t.Errorf("expected 0 errs")
	}
	if receiver.numSetMetaDatas != 0 {
		t.Errorf("expected 0 numSetMetaDatas")
	}
	if receiver.numGetMetaDatas != 1 {
		t.Errorf("expected 1 numGetMetaDatas")
	}
	receiver.m.Unlock()

	res = &gomemcached.MCResponse{
		Opcode: req.Opcode,
		Opaque: req.Opaque,
		Body:   make([]byte, 16),
	}
	binary.BigEndian.PutUint64(res.Body[:8], 102030)   // VB-UUID
	binary.BigEndian.PutUint64(res.Body[8:16], 302010) // VB-SEQ
	reqR = <-rwc.readCh
	copy(reqR.buf, res.HeaderBytes())
	reqR.resCh <- RWRes{n: len(reqR.buf), err: nil}

	reqR = <-rwc.readCh
	copy(reqR.buf, res.Body)
	reqR.resCh <- RWRes{n: len(reqR.buf), err: nil}

	// ------------------------------------------------------------
	e, eok = <-receiver.onGetMetaDataCh
	if e == nil || !eok {
		t.Errorf("expected receiver.onGetMetaDataCh")
	}

	// ------------------------------------------------------------
	req = &gomemcached.MCRequest{
		Opcode: gomemcached.UPR_CONTROL,
	}
	reqR = <-rwc.readCh
	copy(reqR.buf, req.HeaderBytes())
	reqR.resCh <- RWRes{n: len(reqR.buf), err: nil}

	// ------------------------------------------------------------
	receiver.m.Lock()
	if len(receiver.muts) != 0 {
		t.Errorf("expected 0 muts")
	}
	if len(receiver.errs) != 0 {
		t.Errorf("expected 0 errs, got: %v", receiver.errs)
	}
	if receiver.numSetMetaDatas != 1 {
		t.Errorf("expected 1 numSetMetaDatas, got: %#v", receiver)
	}
	if receiver.numGetMetaDatas != 2 {
		t.Errorf("expected 2 numGetMetaDatas, got: %#v", receiver)
	}
	if receiver.meta[2] == nil {
		t.Errorf("expected meta for vbucket 2")
	}
	receiver.m.Unlock()

	req = &gomemcached.MCRequest{
		Opcode:  gomemcached.UPR_SNAPSHOT,
		VBucket: 2,
		Extras:  make([]byte, 20),
	}
	binary.BigEndian.PutUint64(req.Extras[0:8], 102034)  // snapshot-start
	binary.BigEndian.PutUint64(req.Extras[8:16], 103000) // snapshot-end
	binary.BigEndian.PutUint32(req.Extras[16:20], 0)     // snapshot-type

	hb := req.HeaderBytes()
	bb := make([]byte, len(hb)-24+len(res.Body))
	copy(bb, hb[24:])
	copy(bb[len(hb)-24:], res.Body)

	reqR = <-rwc.readCh
	copy(reqR.buf, hb[:24])
	reqR.resCh <- RWRes{n: len(hb), err: nil}

	reqR = <-rwc.readCh
	copy(reqR.buf, bb)
	reqR.resCh <- RWRes{n: len(reqR.buf), err: nil}

	runtime.Gosched()

	// ------------------------------------------------------------
	receiver.m.Lock()
	if len(receiver.muts) != 0 {
		t.Errorf("expected 0 muts")
	}
	if len(receiver.errs) != 0 {
		t.Errorf("expected 0 errs, got: %v", receiver.errs)
	}
	if receiver.numSetMetaDatas != 1 {
		t.Errorf("expected 1 numSetMetaDatas, got: %#v", receiver)
	}
	if receiver.numGetMetaDatas != 2 {
		t.Errorf("expected 2 numGetMetaDatas, got: %#v", receiver)
	}
	if receiver.meta[2] == nil {
		t.Errorf("expected meta for vbucket 2")
	}
	receiver.m.Unlock()

	// ------------------------------------------------------------

	req = &gomemcached.MCRequest{
		Opcode:  gomemcached.UPR_MUTATION,
		VBucket: 2,
		Extras:  make([]byte, 8),
		Key:     []byte("hello"),
		Body:    []byte("world"),
	}
	binary.BigEndian.PutUint64(req.Extras, 102034)

	hb = req.HeaderBytes()
	bb = make([]byte, len(hb)-24+len(res.Body))
	copy(bb, hb[24:])
	copy(bb[len(hb)-24:], res.Body)

	reqR = <-rwc.readCh
	copy(reqR.buf, hb[:24])
	reqR.resCh <- RWRes{n: len(hb), err: nil}

	reqR = <-rwc.readCh
	copy(reqR.buf, bb)
	reqR.resCh <- RWRes{n: len(reqR.buf), err: nil}

	runtime.Gosched()

	// ------------------------------------------------------------
	e, eok = <-receiver.onGetMetaDataCh
	if e == nil || !eok {
		t.Errorf("expected receiver.onGetMetaDataCh")
	}

	// ------------------------------------------------------------

	receiver.m.Lock()
	if len(receiver.muts) != 1 {
		t.Errorf("expected 1 muts")
	}
	if len(receiver.errs) != 0 {
		t.Errorf("expected 0 errs, got: %v", receiver.errs)
	}
	if receiver.numSetMetaDatas != 2 {
		t.Errorf("expected 1 numSetMetaDatas, got: %#v", receiver)
	}
	if receiver.numGetMetaDatas != 3 {
		t.Errorf("expected 2 numGetMetaDatas, got: %#v", receiver)
	}
	if receiver.meta[2] == nil {
		t.Errorf("expected meta for vbucket 2")
	}
	receiver.m.Unlock()

	// ------------------------------------------------------------
	req = &gomemcached.MCRequest{
		Opcode:  gomemcached.UPR_DELETION,
		VBucket: 2,
		Key:     []byte("goodbye"),
		Extras:  make([]byte, 8),
	}
	binary.BigEndian.PutUint64(req.Extras, 102035)

	hb = req.HeaderBytes()
	bb = make([]byte, len(hb)-24+len(res.Body))
	copy(bb, hb[24:])
	copy(bb[len(hb)-24:], res.Body)

	reqR = <-rwc.readCh
	copy(reqR.buf, hb[:24])
	reqR.resCh <- RWRes{n: len(hb), err: nil}

	reqR = <-rwc.readCh
	copy(reqR.buf, bb)
	reqR.resCh <- RWRes{n: len(reqR.buf), err: nil}

	runtime.Gosched()

	receiver.m.Lock()
	if len(receiver.muts) != 2 {
		t.Errorf("expected 2 muts")
	}
	receiver.m.Unlock()

	// ------------------------------------------------------------
	req = &gomemcached.MCRequest{
		Opcode:  gomemcached.UPR_EXPIRATION,
		VBucket: 2,
		Key:     []byte("overdue"),
		Extras:  make([]byte, 8),
	}
	binary.BigEndian.PutUint64(req.Extras, 102036)

	hb = req.HeaderBytes()
	bb = make([]byte, len(hb)-24+len(res.Body))
	copy(bb, hb[24:])
	copy(bb[len(hb)-24:], res.Body)

	reqR = <-rwc.readCh
	copy(reqR.buf, hb[:24])
	reqR.resCh <- RWRes{n: len(hb), err: nil}

	reqR = <-rwc.readCh
	copy(reqR.buf, bb)
	reqR.resCh <- RWRes{n: len(reqR.buf), err: nil}

	runtime.Gosched()

	receiver.m.Lock()
	if len(receiver.muts) != 3 {
		t.Errorf("expected 3 muts")
	}
	receiver.m.Unlock()

	// ------------------------------------------------------------
	err = bds.Close()
	if err != nil {
		t.Errorf("expected clean Close(), got err: %v", err)
	}

	receiver.m.Lock()
	if len(receiver.errs) != 0 {
		t.Errorf("expected 0 errs, got: %v", receiver.errs)
	}
	if receiver.numSetMetaDatas != 2 {
		t.Errorf("expected 2 numSetMetaDatas, got: %#v", receiver)
	}
	if receiver.numGetMetaDatas != 3 {
		t.Errorf("expected 3 numGetMetaDatas, got: %#v", receiver)
	}
	if receiver.meta[2] == nil {
		t.Errorf("expected meta for vbucket 2")
	}

	close(receiver.onGetMetaDataCh)
	receiver.onGetMetaDataCh = nil

	receiver.m.Unlock()

	vbmd, lastSeq, err := bds.(*bucketDataSource).getVBucketMetaData(2)
	if err != nil || vbmd == nil {
		t.Errorf("expected gvbmd to work")
	}
	if lastSeq != 102036 {
		t.Errorf("expected lastseq of 102036, got %d", lastSeq)
	}
	if len(vbmd.FailOverLog) != 1 ||
		len(vbmd.FailOverLog[0]) != 2 ||
		vbmd.FailOverLog[0][0] != 102030 ||
		vbmd.FailOverLog[0][1] != 302010 {
		t.Errorf("mismatch failoverlog, got: %#v", vbmd.FailOverLog)
	}
}
