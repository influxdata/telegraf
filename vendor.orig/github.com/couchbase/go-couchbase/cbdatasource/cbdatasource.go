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

// Package cbdatasource streams data from a Couchbase cluster.  It is
// implemented using Couchbase DCP protocol and has auto-reconnecting
// and auto-restarting goroutines underneath the hood to provide a
// simple, high-level cluster-wide abstraction.  By using
// cbdatasource, your application does not need to worry about
// connections or reconnections to individual server nodes or cluster
// topology changes, rebalance & failovers.  The API starting point is
// NewBucketDataSource().
package cbdatasource

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/rand"
	"reflect"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/couchbase/go-couchbase"
	"github.com/couchbase/go-couchbase/trace"
	"github.com/couchbase/gomemcached"
	"github.com/couchbase/gomemcached/client"
)

const FlagOpenProducer = uint32(1)
const FlagOpenIncludeXattrs = uint32(4)
const FeatureEnabledDataType = uint16(0x01)
const FeatureEnabledXAttrs = uint16(0x06)
const FeatureEnabledXError = uint16(0x07)

var ErrXAttrsNotSupported = fmt.Errorf("xattrs not supported by server")

// BucketDataSource is the main control interface returned by
// NewBucketDataSource().
type BucketDataSource interface {
	// Use Start() to kickoff connectivity to a Couchbase cluster,
	// after which calls will be made to the Receiver's methods.
	Start() error

	// Asynchronously request a cluster map refresh.  A reason string
	// of "" is valid.
	Kick(reason string) error

	// Returns an immutable snapshot of stats.
	Stats(dest *BucketDataSourceStats) error

	// Stops the underlying goroutines.
	Close() error
}

// A Receiver interface is implemented by the application, or the
// receiver of data.  Calls to methods on this interface will be made
// by the BucketDataSource using multiple, concurrent goroutines, so
// the application should implement its own Receiver-side
// synchronizations if needed.
type Receiver interface {
	// Invoked in advisory fashion by the BucketDataSource when it
	// encounters an error.  The BucketDataSource will continue to try
	// to "heal" and restart connections, etc, as necessary.  The
	// Receiver has a recourse during these error notifications of
	// simply Close()'ing the BucketDataSource.
	OnError(error)

	// Invoked by the BucketDataSource when it has received a mutation
	// from the data source.  Receiver implementation is responsible
	// for making its own copies of the key and request.
	DataUpdate(vbucketID uint16, key []byte, seq uint64,
		r *gomemcached.MCRequest) error

	// Invoked by the BucketDataSource when it has received a deletion
	// or expiration from the data source.  Receiver implementation is
	// responsible for making its own copies of the key and request.
	DataDelete(vbucketID uint16, key []byte, seq uint64,
		r *gomemcached.MCRequest) error

	// An callback invoked by the BucketDataSource when it has
	// received a start snapshot message from the data source.  The
	// Receiver implementation, for example, might choose to optimize
	// persistence perhaps by preparing a batch write to
	// application-specific storage.
	SnapshotStart(vbucketID uint16, snapStart, snapEnd uint64, snapType uint32) error

	// The Receiver should persist the value parameter of
	// SetMetaData() for retrieval during some future call to
	// GetMetaData() by the BucketDataSource.  The metadata value
	// should be considered "in-stream", or as part of the sequence
	// history of mutations.  That is, a later Rollback() to some
	// previous sequence number for a particular vbucketID should
	// rollback both persisted metadata and regular data.
	SetMetaData(vbucketID uint16, value []byte) error

	// GetMetaData() should return the opaque value previously
	// provided by an earlier call to SetMetaData().  If there was no
	// previous call to SetMetaData(), such as in the case of a brand
	// new instance of a Receiver (as opposed to a restarted or
	// reloaded Receiver), the Receiver should return (nil, 0, nil)
	// for (value, lastSeq, err), respectively.  The lastSeq should be
	// the last sequence number received and persisted during calls to
	// the Receiver's DataUpdate() & DataDelete() methods.
	GetMetaData(vbucketID uint16) (value []byte, lastSeq uint64, err error)

	// Invoked by the BucketDataSource when the datasource signals a
	// rollback during stream initialization.  Note that both data and
	// metadata should be rolled back.
	Rollback(vbucketID uint16, rollbackSeq uint64) error
}

// A ReceiverEx interface is an advanced Receiver interface that's
// optionally implemented by the application, or the receiver of data.
// Calls to methods on this interface will be made by the
// BucketDataSource using multiple, concurrent goroutines, so the
// application should implement its own Receiver-side synchronizations
// if needed.
type ReceiverEx interface {
	Receiver

	// Invoked by the BucketDataSource when the datasource signals a
	// rollback during stream initialization.  Note that both data and
	// metadata should be rolled back.
	RollbackEx(vbucketID uint16, vbucketUUID uint64, rollbackSeq uint64) error
}

// BucketDataSourceOptions allows the application to provide
// configuration settings to NewBucketDataSource().
type BucketDataSourceOptions struct {
	// Optional - used during UPR_OPEN stream start.  If empty a
	// random name will be automatically generated.
	Name string

	// Factor (like 1.5) to increase sleep time between retries
	// in connecting to a cluster manager node.
	ClusterManagerBackoffFactor float32

	// Initial sleep time (millisecs) before first retry to cluster manager.
	ClusterManagerSleepInitMS int

	// Maximum sleep time (millisecs) between retries to cluster manager.
	ClusterManagerSleepMaxMS int

	// Factor (like 1.5) to increase sleep time between retries
	// in connecting to a data manager node.
	DataManagerBackoffFactor float32

	// Initial sleep time (millisecs) before first retry to data manager.
	DataManagerSleepInitMS int

	// Maximum sleep time (millisecs) between retries to data manager.
	DataManagerSleepMaxMS int

	// Buffer size in bytes provided for UPR flow control.
	FeedBufferSizeBytes uint32

	// Used for UPR flow control and buffer-ack messages when this
	// percentage of FeedBufferSizeBytes is reached.
	FeedBufferAckThreshold float32

	// Time interval in seconds of NO-OP messages for UPR flow control,
	// needs to be set to a non-zero value to enable no-ops.
	NoopTimeIntervalSecs uint32

	// Used for applications like backup which wish to control the
	// last sequence number provided.  Key is vbucketID, value is seqEnd.
	SeqEnd map[uint16]uint64

	// Optional function to connect to a couchbase cluster manager bucket.
	// Defaults to ConnectBucket() function in this package.
	ConnectBucket func(serverURL, poolName, bucketName string,
		auth couchbase.AuthHandler) (Bucket, error)

	// Optional function to connect to a couchbase data manager node.
	// Defaults to memcached.Connect().
	Connect func(protocol, dest string) (*memcached.Client, error)

	// Optional function for logging diagnostic messages.
	Logf func(fmt string, v ...interface{})

	// When true, message trace information will be captured and
	// reported via the Logf() callback.
	TraceCapacity int `json:"-"`

	// When there's been no send/receive activity for this many
	// milliseconds, then transmit a NOOP to the DCP source.  When 0,
	// the DefaultBucketDataSourceOptions.PingTimeoutMS is used.  Of
	// note, the NOOP itself counts as send/receive activity.
	PingTimeoutMS int

	// IncludeXAttrs is an optional flag which specifies whether
	// the clients are interested in the X Attributes values
	// during DCP connection set up.
	// Defaulted to false to keep it backward compatible.
	IncludeXAttrs bool
}

// AllServerURLsConnectBucketError is the error type passed to
// Receiver.OnError() when the BucketDataSource failed to connect to
// all the serverURL's provided as a parameter to
// NewBucketDataSource().  The application, for example, may choose to
// BucketDataSource.Close() based on this error.  Otherwise, the
// BucketDataSource will backoff and retry reconnecting to the
// serverURL's.
type AllServerURLsConnectBucketError struct {
	ServerURLs []string
}

func (e *AllServerURLsConnectBucketError) Error() string {
	return fmt.Sprintf("could not connect to any serverURL: %#v", e.ServerURLs)
}

// AuthFailError is the error type passed to Receiver.OnError() when there
// is an auth request error to the Couchbase cluster or server node.
type AuthFailError struct {
	ServerURL string
	User      string
}

func (e *AuthFailError) Error() string {
	return fmt.Sprintf("auth fail, serverURL: %#v, user: %s", e.ServerURL, e.User)
}

// A Bucket interface defines the set of methods that cbdatasource
// needs from an abstract couchbase.Bucket.  This separate interface
// allows for easier testability.
type Bucket interface {
	Close()
	GetUUID() string
	VBServerMap() *couchbase.VBucketServerMap
}

// DefaultBucketDataSourceOptions defines the default options that
// will be used if nil is provided to NewBucketDataSource().
var DefaultBucketDataSourceOptions = &BucketDataSourceOptions{
	ClusterManagerBackoffFactor: 1.5,
	ClusterManagerSleepInitMS:   100,
	ClusterManagerSleepMaxMS:    1000,

	DataManagerBackoffFactor: 1.5,
	DataManagerSleepInitMS:   100,
	DataManagerSleepMaxMS:    1000,

	FeedBufferSizeBytes:    20000000, // ~20MB; see UPR_CONTROL/connection_buffer_size.
	FeedBufferAckThreshold: 0.2,

	NoopTimeIntervalSecs: 120, // 120 seconds; see UPR_CONTROL/set_noop_interval

	TraceCapacity: 200,

	PingTimeoutMS: 30000,

	IncludeXAttrs: false,
}

// BucketDataSourceStats is filled by the BucketDataSource.Stats()
// method.  All the metrics here prefixed with "Tot" are monotonic
// counters: they only increase.
type BucketDataSourceStats struct {
	TotStart uint64

	TotKick        uint64
	TotKickDeduped uint64
	TotKickOk      uint64

	TotRefreshCluster                              uint64
	TotRefreshClusterConnectBucket                 uint64
	TotRefreshClusterConnectBucketErr              uint64
	TotRefreshClusterConnectBucketOk               uint64
	TotRefreshClusterBucketUUIDErr                 uint64
	TotRefreshClusterVBMNilErr                     uint64
	TotRefreshClusterKickWorkers                   uint64
	TotRefreshClusterKickWorkersClosed             uint64
	TotRefreshClusterKickWorkersStopped            uint64
	TotRefreshClusterKickWorkersOk                 uint64
	TotRefreshClusterStopped                       uint64
	TotRefreshClusterAwokenClosed                  uint64
	TotRefreshClusterAwokenStopped                 uint64
	TotRefreshClusterAwokenRestart                 uint64
	TotRefreshClusterAwoken                        uint64
	TotRefreshClusterAllServerURLsConnectBucketErr uint64
	TotRefreshClusterDone                          uint64

	TotRefreshWorkers                uint64
	TotRefreshWorkersVBMNilErr       uint64
	TotRefreshWorkersVBucketIDErr    uint64
	TotRefreshWorkersServerIdxsErr   uint64
	TotRefreshWorkersMasterIdxErr    uint64
	TotRefreshWorkersMasterServerErr uint64
	TotRefreshWorkersRemoveWorker    uint64
	TotRefreshWorkersAddWorker       uint64
	TotRefreshWorkersKickWorker      uint64
	TotRefreshWorkersCloseWorker     uint64
	TotRefreshWorkersLoop            uint64
	TotRefreshWorkersLoopDone        uint64
	TotRefreshWorkersDone            uint64

	TotWorkerStart      uint64
	TotWorkerDone       uint64
	TotWorkerBody       uint64
	TotWorkerBodyKick   uint64
	TotWorkerConnect    uint64
	TotWorkerConnectErr uint64
	TotWorkerConnectOk  uint64
	TotWorkerAuth       uint64
	TotWorkerAuthErr    uint64
	TotWorkerAuthFail   uint64
	TotWorkerAuthOk     uint64
	TotWorkerUPROpenErr uint64
	TotWorkerUPROpenOk  uint64

	TotWorkerAuthenticateMemcachedConn    uint64
	TotWorkerAuthenticateMemcachedConnErr uint64
	TotWorkerAuthenticateMemcachedConnOk  uint64

	TotWorkerClientClose     uint64
	TotWorkerClientCloseDone uint64

	TotWorkerTransmitStart uint64
	TotWorkerTransmit      uint64
	TotWorkerTransmitErr   uint64
	TotWorkerTransmitOk    uint64
	TotWorkerTransmitDone  uint64

	TotWorkerReceiveStart uint64
	TotWorkerReceive      uint64
	TotWorkerReceiveErr   uint64
	TotWorkerReceiveOk    uint64
	TotWorkerReceiveDone  uint64

	TotWorkerSendEndCh uint64
	TotWorkerRecvEndCh uint64

	TotWorkerHandleRecv    uint64
	TotWorkerHandleRecvErr uint64
	TotWorkerHandleRecvOk  uint64

	TotWorkerCleanup     uint64
	TotWorkerCleanupDone uint64

	TotRefreshWorker     uint64
	TotRefreshWorkerDone uint64
	TotRefreshWorkerOk   uint64

	TotUPRDataChange                       uint64
	TotUPRDataChangeStateErr               uint64
	TotUPRDataChangeMutation               uint64
	TotUPRDataChangeDeletion               uint64
	TotUPRDataChangeExpiration             uint64
	TotUPRDataChangeErr                    uint64
	TotUPRDataChangeOk                     uint64
	TotUPRCloseStream                      uint64
	TotUPRCloseStreamRes                   uint64
	TotUPRCloseStreamResStateErr           uint64
	TotUPRCloseStreamResErr                uint64
	TotUPRCloseStreamResOk                 uint64
	TotUPRStreamReq                        uint64
	TotUPRStreamReqWant                    uint64
	TotUPRStreamReqRes                     uint64
	TotUPRStreamReqResStateErr             uint64
	TotUPRStreamReqResFail                 uint64
	TotUPRStreamReqResFailNotMyVBucket     uint64
	TotUPRStreamReqResFailERange           uint64
	TotUPRStreamReqResFailENoMem           uint64
	TotUPRStreamReqResRollback             uint64
	TotUPRStreamReqResRollbackStart        uint64
	TotUPRStreamReqResRollbackErr          uint64
	TotUPRStreamReqResWantAfterRollbackErr uint64
	TotUPRStreamReqResKick                 uint64
	TotUPRStreamReqResSuccess              uint64
	TotUPRStreamReqResSuccessOk            uint64
	TotUPRStreamReqResFLogErr              uint64
	TotUPRStreamEnd                        uint64
	TotUPRStreamEndStateErr                uint64
	TotUPRStreamEndKick                    uint64
	TotUPRSnapshot                         uint64
	TotUPRSnapshotStateErr                 uint64
	TotUPRSnapshotStart                    uint64
	TotUPRSnapshotStartErr                 uint64
	TotUPRSnapshotOk                       uint64
	TotUPRNoop                             uint64
	TotUPRControl                          uint64
	TotUPRControlErr                       uint64
	TotUPRBufferAck                        uint64

	TotWantCloseRequestedVBucketErr uint64
	TotWantClosingVBucketErr        uint64

	TotSelectBucketErr                uint64
	TotHandShakeErr                   uint64
	TotGetVBucketMetaData             uint64
	TotGetVBucketMetaDataUnmarshalErr uint64
	TotGetVBucketMetaDataErr          uint64
	TotGetVBucketMetaDataOk           uint64

	TotSetVBucketMetaData           uint64
	TotSetVBucketMetaDataMarshalErr uint64
	TotSetVBucketMetaDataErr        uint64
	TotSetVBucketMetaDataOk         uint64

	TotPingTimeout uint64
	TotPingReq     uint64
	TotPingReqDone uint64
}

// --------------------------------------------------------

// VBucketMetaData is an internal struct that is exposed to enable
// json marshaling.
type VBucketMetaData struct {
	SeqStart    uint64     `json:"seqStart"`
	SeqEnd      uint64     `json:"seqEnd"`
	SnapStart   uint64     `json:"snapStart"`
	SnapEnd     uint64     `json:"snapEnd"`
	FailOverLog [][]uint64 `json:"failOverLog"`
}

type bucketDataSource struct {
	serverURLs []string
	poolName   string
	bucketName string
	bucketUUID string
	vbucketIDs []uint16
	auth       couchbase.AuthHandler // Auth for couchbase.
	receiver   Receiver
	options    *BucketDataSourceOptions

	refreshClusterM       sync.Mutex // Protects the refreshClusterReasons field.
	refreshClusterReasons map[string]uint64

	// When refreshClusterReasons transitions from empty to non-empty,
	// then refreshClusterCh must be notified.

	stopCh           chan struct{}
	refreshClusterCh chan struct{}
	refreshWorkersCh chan string
	closedCh         chan bool

	stats BucketDataSourceStats

	m    sync.Mutex // Protects all the below fields.
	life string     // Valid life states: "" (unstarted); "running"; "closed".
	vbm  *couchbase.VBucketServerMap
}

// NewBucketDataSource is the main starting point for using the
// cbdatasource API.  The application must supply an array of 1 or
// more serverURLs (or "seed" URL's) to Couchbase Server
// cluster-manager REST URL endpoints, like "http://localhost:8091".
// The BucketDataSource (after Start()'ing) will try each serverURL,
// in turn, until it can get a successful cluster map.  Additionally,
// the application must supply a poolName & bucketName from where the
// BucketDataSource will retrieve data.  The optional bucketUUID is
// double-checked by the BucketDataSource to ensure we have the
// correct bucket, and a bucketUUID of "" means skip the bucketUUID
// validation.  An optional array of vbucketID numbers allows the
// application to specify which vbuckets to retrieve; and the
// vbucketIDs array can be nil which means all vbuckets are retrieved
// by the BucketDataSource.  The optional auth parameter can be nil.
// The application must supply its own implementation of the Receiver
// interface (see the example program as a sample).  The optional
// options parameter (which may be nil) allows the application to
// specify advanced parameters like backoff and retry-sleep values.
func NewBucketDataSource(
	serverURLs []string,
	poolName string,
	bucketName string,
	bucketUUID string,
	vbucketIDs []uint16,
	auth couchbase.AuthHandler,
	receiver Receiver,
	options *BucketDataSourceOptions) (BucketDataSource, error) {
	if len(serverURLs) < 1 {
		return nil, fmt.Errorf("missing at least 1 serverURL")
	}
	if poolName == "" {
		return nil, fmt.Errorf("missing poolName")
	}
	if bucketName == "" {
		return nil, fmt.Errorf("missing bucketName")
	}
	if receiver == nil {
		return nil, fmt.Errorf("missing receiver")
	}
	if options == nil {
		options = DefaultBucketDataSourceOptions
	}
	return &bucketDataSource{
		serverURLs: serverURLs,
		poolName:   poolName,
		bucketName: bucketName,
		bucketUUID: bucketUUID,
		vbucketIDs: vbucketIDs,
		auth:       auth,
		receiver:   receiver,
		options:    options,

		refreshClusterReasons: map[string]uint64{},

		stopCh:           make(chan struct{}),
		refreshClusterCh: make(chan struct{}),
		refreshWorkersCh: make(chan string, 1),
		closedCh:         make(chan bool),
	}, nil
}

func (d *bucketDataSource) Start() error {
	atomic.AddUint64(&d.stats.TotStart, 1)

	d.m.Lock()
	if d.life != "" {
		d.m.Unlock()
		return fmt.Errorf("call to Start() in wrong state: %s", d.life)
	}
	d.life = "running"
	d.m.Unlock()

	backoffFactor := d.options.ClusterManagerBackoffFactor
	if backoffFactor <= 0.0 {
		backoffFactor = DefaultBucketDataSourceOptions.ClusterManagerBackoffFactor
	}
	sleepInitMS := d.options.ClusterManagerSleepInitMS
	if sleepInitMS <= 0 {
		sleepInitMS = DefaultBucketDataSourceOptions.ClusterManagerSleepInitMS
	}
	sleepMaxMS := d.options.ClusterManagerSleepMaxMS
	if sleepMaxMS <= 0 {
		sleepMaxMS = DefaultBucketDataSourceOptions.ClusterManagerSleepMaxMS
	}

	go func() {
		ExponentialBackoffLoop("cbdatasource.refreshCluster",
			func() int { return d.refreshCluster() },
			sleepInitMS, backoffFactor, sleepMaxMS)

		// We reach here when we need to shutdown.
		close(d.refreshWorkersCh)
		atomic.AddUint64(&d.stats.TotRefreshClusterDone, 1)
	}()

	go d.refreshWorkers()

	return nil
}

func (d *bucketDataSource) isRunning() bool {
	d.m.Lock()
	life := d.life
	d.m.Unlock()
	return life == "running"
}

func (d *bucketDataSource) refreshCluster() int {
	atomic.AddUint64(&d.stats.TotRefreshCluster, 1)

	if !d.isRunning() {
		return -1
	}

	for _, serverURL := range d.serverURLs {
		atomic.AddUint64(&d.stats.TotRefreshClusterConnectBucket, 1)

		connectBucket := d.options.ConnectBucket
		if connectBucket == nil {
			connectBucket = ConnectBucket
		}

		bucket, err := connectBucket(serverURL, d.poolName, d.bucketName, d.auth)
		if err != nil {
			atomic.AddUint64(&d.stats.TotRefreshClusterConnectBucketErr, 1)
			d.receiver.OnError(err)
			continue // Try another serverURL.
		}
		atomic.AddUint64(&d.stats.TotRefreshClusterConnectBucketOk, 1)

		if d.bucketUUID != "" && d.bucketUUID != bucket.GetUUID() {
			bucket.Close()
			atomic.AddUint64(&d.stats.TotRefreshClusterBucketUUIDErr, 1)
			d.receiver.OnError(fmt.Errorf("mismatched bucket uuid,"+
				" serverURL: %s, bucketName: %s, bucketUUID: %s, bucket.UUID: %s",
				serverURL, d.bucketName, d.bucketUUID, bucket.GetUUID()))
			continue // Try another serverURL.
		}

		vbm := bucket.VBServerMap()
		if vbm == nil {
			bucket.Close()
			atomic.AddUint64(&d.stats.TotRefreshClusterVBMNilErr, 1)
			d.receiver.OnError(fmt.Errorf("refreshCluster got no vbm,"+
				" serverURL: %s, bucketName: %s, bucketUUID: %s, bucket.UUID: %s",
				serverURL, d.bucketName, d.bucketUUID, bucket.GetUUID()))
			continue // Try another serverURL.
		}

		bucket.Close()

		d.m.Lock()
		d.vbm = vbm
		d.m.Unlock()

		for {
			atomic.AddUint64(&d.stats.TotRefreshClusterKickWorkers, 1)
			select {
			case <-d.stopCh:
				atomic.AddUint64(&d.stats.TotRefreshClusterKickWorkersStopped, 1)
				return -1

			case d.refreshWorkersCh <- "new-vbm": // Kick workers to refresh.
				// NO-OP.
			}
			atomic.AddUint64(&d.stats.TotRefreshClusterKickWorkersOk, 1)

			// Wait for refreshCluster kick.
			var refreshClusterReasons map[string]uint64

			for {
				d.refreshClusterM.Lock()
				if len(d.refreshClusterReasons) > 0 {
					refreshClusterReasons = d.refreshClusterReasons
					d.refreshClusterReasons = map[string]uint64{}
				}
				d.refreshClusterM.Unlock()

				if len(refreshClusterReasons) > 0 {
					break
				}

				select {
				case <-d.stopCh:
					atomic.AddUint64(&d.stats.TotRefreshClusterStopped, 1)
					return -1

				case _, refreshAlive := <-d.refreshClusterCh:
					if !refreshAlive {
						atomic.AddUint64(&d.stats.TotRefreshClusterAwokenClosed, 1)
						return -1
					}
				}
			}

			if !d.isRunning() {
				atomic.AddUint64(&d.stats.TotRefreshClusterAwokenStopped, 1)
				return -1
			}

			// If it's only that new workers have appeared, then we
			// can keep with this inner loop and not have to restart
			// all the way at the top / retrieve a new cluster map, etc.
			wasNewWorkerOnly :=
				len(refreshClusterReasons) == 1 &&
					refreshClusterReasons["new-worker"] > 0
			if !wasNewWorkerOnly {
				atomic.AddUint64(&d.stats.TotRefreshClusterAwokenRestart, 1)
				return 1 // Assume progress, so restart at first serverURL.
			}

			atomic.AddUint64(&d.stats.TotRefreshClusterAwoken, 1)
		}
	}

	// Notify Receiver in case it wants to Close() down this
	// BucketDataSource after enough attempts.  The typed interfaces
	// allow Receiver to have better error handling logic.
	atomic.AddUint64(&d.stats.TotRefreshClusterAllServerURLsConnectBucketErr, 1)
	d.receiver.OnError(&AllServerURLsConnectBucketError{ServerURLs: d.serverURLs})

	return 0 // Ran through all the serverURLs, so no progress.
}

func (d *bucketDataSource) refreshWorkers() {
	// Keyed by server, value is chan of array of vbucketID's that the
	// worker needs to provide.
	workers := make(map[string]chan []uint16)

OUTER_LOOP:
	for _ = range d.refreshWorkersCh { // Wait for a refresh kick.
		atomic.AddUint64(&d.stats.TotRefreshWorkers, 1)

		d.m.Lock()
		vbm := d.vbm
		d.m.Unlock()

		if vbm == nil {
			atomic.AddUint64(&d.stats.TotRefreshWorkersVBMNilErr, 1)
			continue
		}

		// If nil vbucketIDs, then default to all vbucketIDs.
		vbucketIDs := d.vbucketIDs
		if vbucketIDs == nil {
			vbucketIDs = make([]uint16, len(vbm.VBucketMap))
			for i := 0; i < len(vbucketIDs); i++ {
				vbucketIDs[i] = uint16(i)
			}
		}

		// Group the wanted vbucketIDs by server.
		vbucketIDsByServer := make(map[string][]uint16)

		for _, vbucketID := range vbucketIDs {
			if int(vbucketID) >= len(vbm.VBucketMap) {
				atomic.AddUint64(&d.stats.TotRefreshWorkersVBucketIDErr, 1)
				d.receiver.OnError(fmt.Errorf("refreshWorkers"+
					" saw bad vbucketID: %d",
					vbucketID))
				d.Kick("bad-vbm")
				continue
			}
			serverIdxs := vbm.VBucketMap[vbucketID]
			if serverIdxs == nil || len(serverIdxs) <= 0 {
				atomic.AddUint64(&d.stats.TotRefreshWorkersServerIdxsErr, 1)
				d.receiver.OnError(fmt.Errorf("refreshWorkers"+
					" no serverIdxs for vbucketID: %d",
					vbucketID))
				continue
			}
			masterIdx := serverIdxs[0]
			if masterIdx < 0 || int(masterIdx) >= len(vbm.ServerList) {
				atomic.AddUint64(&d.stats.TotRefreshWorkersMasterIdxErr, 1)
				d.receiver.OnError(fmt.Errorf("refreshWorkers"+
					" bad masterIdx: %d, vbucketID: %d",
					masterIdx, vbucketID))
				continue
			}
			masterServer := vbm.ServerList[masterIdx]
			if masterServer == "" {
				atomic.AddUint64(&d.stats.TotRefreshWorkersMasterServerErr, 1)
				d.receiver.OnError(fmt.Errorf("refreshWorkers"+
					" no masterServer for vbucketID: %d",
					vbucketID))
				continue
			}
			v, exists := vbucketIDsByServer[masterServer]
			if !exists || v == nil {
				v = []uint16{}
			}
			vbucketIDsByServer[masterServer] = append(v, vbucketID)
		}

		// Remove any extraneous workers.
		for server, workerCh := range workers {
			if _, exists := vbucketIDsByServer[server]; !exists {
				atomic.AddUint64(&d.stats.TotRefreshWorkersRemoveWorker, 1)
				delete(workers, server)
				close(workerCh)
			}
		}

		// Add any missing workers and update workers with their
		// latest vbucketIDs.
		for server, serverVBucketIDs := range vbucketIDsByServer {
			workerCh, exists := workers[server]
			if !exists || workerCh == nil {
				atomic.AddUint64(&d.stats.TotRefreshWorkersAddWorker, 1)
				workerCh = make(chan []uint16, 1)
				workers[server] = workerCh
				d.workerStart(server, workerCh)
			}

			select {
			case <-d.stopCh:
				break OUTER_LOOP
			case workerCh <- serverVBucketIDs:
				// NOOP.
			}

			atomic.AddUint64(&d.stats.TotRefreshWorkersKickWorker, 1)
		}

		atomic.AddUint64(&d.stats.TotRefreshWorkersLoop, 1)
	}

	atomic.AddUint64(&d.stats.TotRefreshWorkersLoopDone, 1)

	// We reach here when we need to shutdown.
	for _, workerCh := range workers {
		atomic.AddUint64(&d.stats.TotRefreshWorkersCloseWorker, 1)
		close(workerCh)
	}

	close(d.closedCh)
	atomic.AddUint64(&d.stats.TotRefreshWorkersDone, 1)
}

// A worker connects to one data manager server.
func (d *bucketDataSource) workerStart(server string, workerCh chan []uint16) {
	backoffFactor := d.options.DataManagerBackoffFactor
	if backoffFactor <= 0.0 {
		backoffFactor = DefaultBucketDataSourceOptions.DataManagerBackoffFactor
	}
	sleepInitMS := d.options.DataManagerSleepInitMS
	if sleepInitMS <= 0 {
		sleepInitMS = DefaultBucketDataSourceOptions.DataManagerSleepInitMS
	}
	sleepMaxMS := d.options.DataManagerSleepMaxMS
	if sleepMaxMS <= 0 {
		sleepMaxMS = DefaultBucketDataSourceOptions.DataManagerSleepMaxMS
	}

	// Use exponential backoff loop to handle connect retries to the server.
	go func() {
		atomic.AddUint64(&d.stats.TotWorkerStart, 1)

		ExponentialBackoffLoop("cbdatasource.worker-"+server,
			func() int { return d.worker(server, workerCh) },
			sleepInitMS, backoffFactor, sleepMaxMS)

		atomic.AddUint64(&d.stats.TotWorkerDone, 1)
	}()
}

type VBucketState struct {
	// Valid values for state: "" (dead/closed/unknown), "requested",
	// "running", "closing".
	State     string
	SnapStart uint64
	SnapEnd   uint64
	SnapSaved bool // True when the snapStart/snapEnd have been persisted.
}

// Connect once to the server and work the UPR stream.  If anything
// goes wrong, return our level of progress in order to let our caller
// control any potential retries.
func (d *bucketDataSource) worker(server string, workerCh chan []uint16) int {
	atomic.AddUint64(&d.stats.TotWorkerBody, 1)

	if !d.isRunning() {
		return -1
	}

	atomic.AddUint64(&d.stats.TotWorkerConnect, 1)
	connect := d.options.Connect
	if connect == nil {
		connect = memcached.Connect
	}

	emptyWorkerCh := func() {
		for {
			select {
			case _, ok := <-workerCh:
				if !ok {
					return
				}

				// Else, keep looping to consume workerCh.

			default:
				return // Stop loop when workerCh is empty.
			}
		}
	}

	client, err := connect("tcp", server)
	if err != nil {
		atomic.AddUint64(&d.stats.TotWorkerConnectErr, 1)
		d.receiver.OnError(fmt.Errorf("worker connect, server: %s, err: %v",
			server, err))

		// If we can't connect, then maybe a node was rebalanced out
		// or failed-over, so consume the workerCh so that the
		// refresh-cluster goroutine will be unblocked and can receive
		// our kick.
		emptyWorkerCh()

		d.Kick("worker-connect-err")

		return 0
	}
	atomic.AddUint64(&d.stats.TotWorkerConnectOk, 1)

	defer func() {
		atomic.AddUint64(&d.stats.TotWorkerClientClose, 1)
		client.Close()
		atomic.AddUint64(&d.stats.TotWorkerClientCloseDone, 1)
	}()

	var user, pswd string

	if d.auth != nil {
		if auth, ok := d.auth.(couchbase.GenericMcdAuthHandler); ok {
			atomic.AddUint64(&d.stats.TotWorkerAuthenticateMemcachedConn, 1)
			err = auth.AuthenticateMemcachedConn(server, client)
			if err != nil {
				atomic.AddUint64(&d.stats.TotWorkerAuthenticateMemcachedConnErr, 1)
				d.receiver.OnError(fmt.Errorf("worker auth,"+
					" AuthenticateMemcachedConn, server: %s, err: %v",
					server, err))

				// If we can't authenticate, then maybe a node was
				// rebalanced out, so consume the workerCh so that the
				// refresh-cluster goroutine will be unblocked and can
				// receive our kick.
				emptyWorkerCh()

				d.Kick("worker-auth-AuthenticateMemcachedConn")

				return 0
			}
			atomic.AddUint64(&d.stats.TotWorkerAuthenticateMemcachedConnOk, 1)
		} else if auth, ok := d.auth.(couchbase.AuthWithSaslHandler); ok {
			user, pswd = auth.GetSaslCredentials()
		} else {
			user, pswd, _ = d.auth.GetCredentials()
		}

		if user != "" {
			atomic.AddUint64(&d.stats.TotWorkerAuth, 1)
			res, err := client.Auth(user, pswd)
			if err != nil {
				atomic.AddUint64(&d.stats.TotWorkerAuthErr, 1)
				d.receiver.OnError(fmt.Errorf("worker auth, server: %s,"+
					" user: %s, err: %v", server, user, err))
				return 0
			}

			if res.Status != gomemcached.SUCCESS {
				atomic.AddUint64(&d.stats.TotWorkerAuthFail, 1)
				d.receiver.OnError(&AuthFailError{ServerURL: server, User: user})
				return 0
			}
		}
	}

	uprOpenName := d.options.Name
	if uprOpenName == "" {
		uprOpenName = fmt.Sprintf("cbdatasource-%x", rand.Int63())
	}

	// Server handshakes.
	var openUprFlags uint32
	openUprFlags, err = serverHandShake(client, d)
	if err != nil {
		// Just bump the stats and continue with errors logged
		// since even at feature mismatch the openConn can proceed.
		atomic.AddUint64(&d.stats.TotHandShakeErr, 1)
		if d.options.Logf != nil {
			d.options.Logf("cbdatasource: serverHandShake, err: %v", err)
		}
	}

	// Select the bucket for this connection.
	if user != d.bucketName {
		err = selectBucket(client, d.bucketName)
		if err != nil {
			atomic.AddUint64(&d.stats.TotSelectBucketErr, 1)
			d.receiver.OnError(err)
			return 0
		}
	}

	err = UPROpen(client, uprOpenName, d.options, openUprFlags)
	if err != nil {
		atomic.AddUint64(&d.stats.TotWorkerUPROpenErr, 1)
		d.receiver.OnError(err)
		return 0
	}
	atomic.AddUint64(&d.stats.TotWorkerUPROpenOk, 1)

	ackBytes :=
		uint32(d.options.FeedBufferAckThreshold * float32(d.options.FeedBufferSizeBytes))

	sendCh := make(chan interface{}, 1)
	sendEndCh := make(chan struct{})
	recvEndCh := make(chan struct{})

	cleanup := func(progress int, err error) int {
		atomic.AddUint64(&d.stats.TotWorkerCleanup, 1)

		if err != nil {
			d.receiver.OnError(err)
		}

		go func() {
			<-recvEndCh
			close(sendCh)
		}()

		atomic.AddUint64(&d.stats.TotWorkerCleanupDone, 1)
		return progress
	}

	currVBuckets := make(map[uint16]*VBucketState)
	currVBucketsMutex := sync.Mutex{} // Protects currVBuckets.

	go func() { // Sender goroutine.
		defer func() {
			close(sendEndCh)
			atomic.AddUint64(&d.stats.TotWorkerTransmitDone, 1)
		}()

		atomic.AddUint64(&d.stats.TotWorkerTransmitStart, 1)
		for msg := range sendCh {
			atomic.AddUint64(&d.stats.TotWorkerTransmit, 1)
			mcPkt, ok := msg.(*gomemcached.MCRequest)
			if ok {
				// Transmit a request.
				err := client.Transmit(mcPkt)
				if err != nil {
					atomic.AddUint64(&d.stats.TotWorkerTransmitErr, 1)
					d.receiver.OnError(fmt.Errorf("client.Transmit, err: %v", err))
					return
				}
			} else {
				mcPkt, ok := msg.(*gomemcached.MCResponse)
				if ok {
					// Transmit a response.
					err := client.TransmitResponse(mcPkt)
					if err != nil {
						atomic.AddUint64(&d.stats.TotWorkerTransmitErr, 1)
						d.receiver.OnError(fmt.Errorf("client.TransmitResponse, err: %v", err))
						return
					}
				} else {
					// Unknown packet.
					d.receiver.OnError(fmt.Errorf("Unidentified packet to transmit!"))
					return
				}
			}
			atomic.AddUint64(&d.stats.TotWorkerTransmitOk, 1)
		}
	}()

	go func() { // Receiver goroutine.
		defer func() {
			close(recvEndCh)
			atomic.AddUint64(&d.stats.TotWorkerReceiveDone, 1)
		}()

		traceCapacity := d.options.TraceCapacity
		if traceCapacity == 0 {
			traceCapacity = DefaultBucketDataSourceOptions.TraceCapacity
		}
		if traceCapacity < 0 {
			traceCapacity = 0
		}

		trs := map[uint16]*trace.RingBuffer{} // Keyed by vbucket.

		if d.options.Logf != nil {
			d.options.Logf("cbdatasource: receiver tracing,"+
				" server: %s, name: %s, capacity: %d",
				server, uprOpenName, traceCapacity)

			defer func() {
				vbids := make(sort.IntSlice, 0, len(trs))
				for vbid, tr := range trs {
					if tr != nil {
						vbids = append(vbids, int(vbid))
					}
				}
				sort.Sort(vbids)

				var buf bytes.Buffer
				for _, vbid := range vbids {
					tr := trs[uint16(vbid)]
					if tr != nil {
						fmt.Fprintf(&buf, " vb: %d => %s;", vbid,
							trace.MsgsToString(tr.Msgs(), ",", " "))
					}
				}

				d.options.Logf("cbdatasource: receiver closed,"+
					" server: %s, name: %s, traces:%s",
					server, uprOpenName, buf.String())
			}()
		}

		atomic.AddUint64(&d.stats.TotWorkerReceiveStart, 1)

		var hdr [gomemcached.HDR_LEN]byte
		var pkt gomemcached.MCRequest
		var res gomemcached.MCResponse

		// Track received bytes in case we need to buffer-ack.
		recvBytesTotal := uint32(0)

		conn := client.Hijack()

		for {
			// TODO: memory allocation here.
			atomic.AddUint64(&d.stats.TotWorkerReceive, 1)
			_, err := pkt.Receive(conn, hdr[:])
			if err != nil {
				atomic.AddUint64(&d.stats.TotWorkerReceiveErr, 1)
				d.receiver.OnError(fmt.Errorf("pkt.Receive, err: %v", err))
				return
			}
			atomic.AddUint64(&d.stats.TotWorkerReceiveOk, 1)

			tr := trs[pkt.VBucket]
			if tr == nil {
				tr = trace.NewRingBuffer(traceCapacity, trace.ConsolidateByTitle)
				trs[pkt.VBucket] = tr
			}

			if pkt.Opcode == gomemcached.UPR_MUTATION ||
				pkt.Opcode == gomemcached.UPR_DELETION ||
				pkt.Opcode == gomemcached.UPR_EXPIRATION {
				tr.Add("md", nil)

				atomic.AddUint64(&d.stats.TotUPRDataChange, 1)

				vbucketID := pkt.VBucket

				currVBucketsMutex.Lock()

				vbucketState := currVBuckets[vbucketID]
				if vbucketState == nil || vbucketState.State != "running" {
					currVBucketsMutex.Unlock()
					atomic.AddUint64(&d.stats.TotUPRDataChangeStateErr, 1)
					d.receiver.OnError(fmt.Errorf("error: DataChange,"+
						" wrong vbucketState: %#v, err: %v", vbucketState, err))
					return
				}

				if !vbucketState.SnapSaved {
					// NOTE: Following the ep-engine's approach, we
					// wait to persist SnapStart/SnapEnd until we see
					// the first mutation/deletion in the new snapshot
					// range.  That reduces a race window where if we
					// kill and restart this process right now after a
					// setVBucketMetaData() and before the next,
					// first-mutation-in-snapshot, then a restarted
					// stream-req using this just-saved
					// SnapStart/SnapEnd might have a lastSeq number <
					// SnapStart, where Couchbase Server will respond
					// to the stream-req with an ERANGE error code.
					v, _, err := d.getVBucketMetaData(vbucketID)
					if err != nil || v == nil {
						currVBucketsMutex.Unlock()
						d.receiver.OnError(fmt.Errorf("error: DataChange,"+
							" getVBucketMetaData, vbucketID: %d, err: %v",
							vbucketID, err))
						return
					}

					v.SnapStart = vbucketState.SnapStart
					v.SnapEnd = vbucketState.SnapEnd

					err = d.setVBucketMetaData(vbucketID, v)
					if err != nil {
						currVBucketsMutex.Unlock()
						d.receiver.OnError(fmt.Errorf("error: DataChange,"+
							" getVBucketMetaData, vbucketID: %d, err: %v",
							vbucketID, err))
						return
					}

					vbucketState.SnapSaved = true
				}

				currVBucketsMutex.Unlock()

				seq := binary.BigEndian.Uint64(pkt.Extras[:8])

				if pkt.Opcode == gomemcached.UPR_MUTATION {
					atomic.AddUint64(&d.stats.TotUPRDataChangeMutation, 1)
					err = d.receiver.DataUpdate(vbucketID, pkt.Key, seq, &pkt)
				} else {
					if pkt.Opcode == gomemcached.UPR_DELETION {
						atomic.AddUint64(&d.stats.TotUPRDataChangeDeletion, 1)
					} else {
						atomic.AddUint64(&d.stats.TotUPRDataChangeExpiration, 1)
					}
					err = d.receiver.DataDelete(vbucketID, pkt.Key, seq, &pkt)
				}

				if err != nil {
					atomic.AddUint64(&d.stats.TotUPRDataChangeErr, 1)
					d.receiver.OnError(fmt.Errorf("error: DataChange, err: %v", err))
					return
				}

				atomic.AddUint64(&d.stats.TotUPRDataChangeOk, 1)
			} else {
				tr.Add(fmt.Sprintf("%d", pkt.Opcode), nil)

				res.Opcode = pkt.Opcode
				res.Opaque = pkt.Opaque
				res.Status = gomemcached.Status(pkt.VBucket)
				res.Extras = pkt.Extras
				res.Cas = pkt.Cas
				res.Key = pkt.Key
				res.Body = pkt.Body

				atomic.AddUint64(&d.stats.TotWorkerHandleRecv, 1)
				currVBucketsMutex.Lock()
				err := d.handleRecv(sendCh, currVBuckets, &res)
				currVBucketsMutex.Unlock()
				if err != nil {
					atomic.AddUint64(&d.stats.TotWorkerHandleRecvErr, 1)
					d.receiver.OnError(fmt.Errorf("error: HandleRecv, err: %v", err))
					return
				}
				atomic.AddUint64(&d.stats.TotWorkerHandleRecvOk, 1)
			}

			recvBytesTotal +=
				uint32(gomemcached.HDR_LEN) +
					uint32(len(pkt.Key)+len(pkt.Extras)+len(pkt.Body))
			if ackBytes > 0 && recvBytesTotal > ackBytes {
				atomic.AddUint64(&d.stats.TotUPRBufferAck, 1)
				ack := &gomemcached.MCRequest{Opcode: gomemcached.UPR_BUFFERACK}
				ack.Extras = make([]byte, 4) // TODO: Memory mgmt.
				binary.BigEndian.PutUint32(ack.Extras, uint32(recvBytesTotal))
				sendCh <- ack
				recvBytesTotal = 0
			}
		}
	}()

	atomic.AddUint64(&d.stats.TotWorkerBodyKick, 1)
	d.Kick("new-worker")

	pingTimeoutMS := d.options.PingTimeoutMS
	if pingTimeoutMS <= 0 {
		pingTimeoutMS = DefaultBucketDataSourceOptions.PingTimeoutMS
	}

	var prevActivity uint64

	for {
		currVBucketsMutex.Lock()
		d.logVBucketStates(server, uprOpenName, "worker, looping beg",
			currVBuckets, nil)
		currVBucketsMutex.Unlock()

		select {
		case <-sendEndCh:
			atomic.AddUint64(&d.stats.TotWorkerSendEndCh, 1)
			return cleanup(0, nil)

		case <-recvEndCh:
			// If we lost a connection, then maybe a node was rebalanced out,
			// or failed over, so ask for a cluster refresh just in case.
			d.Kick("recvEndCh")

			atomic.AddUint64(&d.stats.TotWorkerRecvEndCh, 1)
			return cleanup(0, nil)

		case wantVBucketIDs, alive := <-workerCh:
			atomic.AddUint64(&d.stats.TotRefreshWorker, 1)

			if !alive {
				atomic.AddUint64(&d.stats.TotRefreshWorkerDone, 1)
				return cleanup(-1, nil) // We've been asked to shutdown.
			}

			currVBucketsMutex.Lock()
			d.logVBucketStates(server, uprOpenName, "refreshWorker-prior",
				currVBuckets, nil)
			err := d.refreshWorker(sendCh, currVBuckets, wantVBucketIDs)
			d.logVBucketStates(server, uprOpenName, "refreshWorker-after",
				currVBuckets, err)
			currVBucketsMutex.Unlock()
			if err != nil {
				return cleanup(0, err)
			}

			atomic.AddUint64(&d.stats.TotRefreshWorkerOk, 1)

		case <-time.After(time.Duration(pingTimeoutMS) * time.Millisecond):
			atomic.AddUint64(&d.stats.TotPingTimeout, 1)

			currActivity := atomic.LoadUint64(&d.stats.TotWorkerTransmit) +
				atomic.LoadUint64(&d.stats.TotWorkerReceive)
			if currActivity == prevActivity { // If no activity, then ping.
				atomic.AddUint64(&d.stats.TotPingReq, 1)
				sendCh <- &gomemcached.MCRequest{Opcode: gomemcached.NOOP}
				atomic.AddUint64(&d.stats.TotPingReqDone, 1)
			}

			prevActivity = currActivity
		}
	}

	return cleanup(-1, nil) // Unreached.
}

func (d *bucketDataSource) refreshWorker(sendCh chan interface{},
	currVBuckets map[uint16]*VBucketState, wantVBucketIDsArr []uint16) error {
	// Convert to map for faster lookup.
	wantVBucketIDs := map[uint16]bool{}
	for _, wantVBucketID := range wantVBucketIDsArr {
		wantVBucketIDs[wantVBucketID] = true
	}

	for currVBucketID, state := range currVBuckets {
		if !wantVBucketIDs[currVBucketID] {
			if state != nil {
				if state.State == "requested" {
					// A UPR_STREAMREQ request is already on the wire, so
					// error rather than have complex compensation logic.
					atomic.AddUint64(&d.stats.TotWantCloseRequestedVBucketErr, 1)
					return fmt.Errorf("want close requested vbucketID: %d", currVBucketID)
				}
				if state.State == "running" {
					state.State = "closing"
					atomic.AddUint64(&d.stats.TotUPRCloseStream, 1)
					sendCh <- &gomemcached.MCRequest{
						Opcode:  gomemcached.UPR_CLOSESTREAM,
						VBucket: currVBucketID,
						Opaque:  uint32(currVBucketID),
					}
				} // Else, state.State of "" or "closing", so no-op.
			} // Else state of nil, so no-op.
		}
	}

	for wantVBucketID := range wantVBucketIDs {
		state := currVBuckets[wantVBucketID]
		if state != nil && state.State == "closing" {
			// A UPR_CLOSESTREAM request is already on the wire, so
			// error rather than have complex compensation logic.
			atomic.AddUint64(&d.stats.TotWantClosingVBucketErr, 1)
			return fmt.Errorf("want closing vbucketID: %d", wantVBucketID)
		}
		if state == nil || state.State == "" {
			currVBuckets[wantVBucketID] = &VBucketState{State: "requested"}
			atomic.AddUint64(&d.stats.TotUPRStreamReqWant, 1)
			err := d.sendStreamReq(sendCh, wantVBucketID)
			if err != nil {
				return err
			}
		} // Else, state.State of "requested" or "running", so no-op.
	}

	return nil
}

func (d *bucketDataSource) handleRecv(sendCh chan interface{},
	currVBuckets map[uint16]*VBucketState, res *gomemcached.MCResponse) error {
	switch res.Opcode {
	case gomemcached.UPR_NOOP:
		atomic.AddUint64(&d.stats.TotUPRNoop, 1)
		sendCh <- &gomemcached.MCResponse{
			Opcode: gomemcached.UPR_NOOP,
			Opaque: res.Opaque,
		}

	case gomemcached.UPR_STREAMREQ:
		atomic.AddUint64(&d.stats.TotUPRStreamReqRes, 1)

		vbucketID := uint16(res.Opaque)
		vbucketState := currVBuckets[vbucketID]

		delete(currVBuckets, vbucketID)

		if vbucketState == nil || vbucketState.State != "requested" {
			atomic.AddUint64(&d.stats.TotUPRStreamReqResStateErr, 1)
			return fmt.Errorf("streamreq non-requested,"+
				" vbucketID: %d, vbucketState: %#v, res: %#v",
				vbucketID, vbucketState, res)
		}

		if res.Status != gomemcached.SUCCESS {
			atomic.AddUint64(&d.stats.TotUPRStreamReqResFail, 1)

			if res.Status == gomemcached.ROLLBACK ||
				res.Status == gomemcached.ERANGE {
				rollbackSeq := uint64(0)
				vbucketUUID := uint64(0)

				if res.Status == gomemcached.ROLLBACK {
					atomic.AddUint64(&d.stats.TotUPRStreamReqResRollback, 1)

					if len(res.Body) < 8 {
						return fmt.Errorf("bad rollback body: %#v", res)
					}

					rollbackSeq = binary.BigEndian.Uint64(res.Body)
					if len(res.Extras) >= 32 {
						vbucketUUID = binary.BigEndian.Uint64(res.Extras[24:32])
					}
				} else {
					// NOTE: Not sure what else to do here on ERANGE
					// error response besides rollback to zero.
					atomic.AddUint64(&d.stats.TotUPRStreamReqResFailERange, 1)
				}

				atomic.AddUint64(&d.stats.TotUPRStreamReqResRollbackStart, 1)
				var err error
				if receiverEx, ok := d.receiver.(ReceiverEx); ok {
					err = receiverEx.RollbackEx(vbucketID, vbucketUUID, rollbackSeq)
				} else {
					err = d.receiver.Rollback(vbucketID, rollbackSeq)
				}

				if err != nil {
					atomic.AddUint64(&d.stats.TotUPRStreamReqResRollbackErr, 1)
					return err
				}

				currVBuckets[vbucketID] = &VBucketState{State: "requested"}
				atomic.AddUint64(&d.stats.TotUPRStreamReqResWantAfterRollbackErr, 1)
				err = d.sendStreamReq(sendCh, vbucketID)
				if err != nil {
					return err
				}
			} else {
				if res.Status == gomemcached.NOT_MY_VBUCKET {
					atomic.AddUint64(&d.stats.TotUPRStreamReqResFailNotMyVBucket, 1)
				} else if res.Status == gomemcached.ENOMEM {
					atomic.AddUint64(&d.stats.TotUPRStreamReqResFailENoMem, 1)
				}

				// Maybe the vbucket moved, so kick off a cluster refresh.
				atomic.AddUint64(&d.stats.TotUPRStreamReqResKick, 1)
				d.Kick("stream-req-error")
			}
		} else { // SUCCESS case.
			atomic.AddUint64(&d.stats.TotUPRStreamReqResSuccess, 1)

			flog, err := ParseFailOverLog(res.Body[:])
			if err != nil {
				atomic.AddUint64(&d.stats.TotUPRStreamReqResFLogErr, 1)
				return err
			}
			v, _, err := d.getVBucketMetaData(vbucketID)
			if err != nil {
				return err
			}

			v.FailOverLog = flog

			err = d.setVBucketMetaData(vbucketID, v)
			if err != nil {
				return err
			}

			currVBuckets[vbucketID] = &VBucketState{State: "running"}
			atomic.AddUint64(&d.stats.TotUPRStreamReqResSuccessOk, 1)
		}

	case gomemcached.UPR_STREAMEND:
		atomic.AddUint64(&d.stats.TotUPRStreamEnd, 1)

		vbucketID := uint16(res.Status)
		vbucketState := currVBuckets[vbucketID]

		delete(currVBuckets, vbucketID)

		if vbucketState == nil ||
			(vbucketState.State != "running" && vbucketState.State != "closing") {
			atomic.AddUint64(&d.stats.TotUPRStreamEndStateErr, 1)
			return fmt.Errorf("stream-end bad state,"+
				" vbucketID: %d, vbucketState: %#v, res: %#v",
				vbucketID, vbucketState, res)
		}

		// We should not normally see a stream-end, unless we were
		// trying to close.  Maybe the vbucket moved, though, so kick
		// off a cluster refresh.
		if vbucketState.State != "closing" {
			atomic.AddUint64(&d.stats.TotUPRStreamEndKick, 1)
			d.Kick("stream-end")
		}

	case gomemcached.UPR_CLOSESTREAM:
		atomic.AddUint64(&d.stats.TotUPRCloseStreamRes, 1)

		vbucketID := uint16(res.Opaque)
		vbucketState := currVBuckets[vbucketID]

		if vbucketState == nil || vbucketState.State != "closing" {
			atomic.AddUint64(&d.stats.TotUPRCloseStreamResStateErr, 1)
			return fmt.Errorf("close-stream bad state,"+
				" vbucketID: %d, vbucketState: %#v, res: %#v",
				vbucketID, vbucketState, res)
		}

		if res.Status != gomemcached.SUCCESS {
			atomic.AddUint64(&d.stats.TotUPRCloseStreamResErr, 1)
			return fmt.Errorf("close-stream failed,"+
				" vbucketID: %d, vbucketState: %#v, res: %#v",
				vbucketID, vbucketState, res)
		}

		// At this point, we can ignore this success response to our
		// close-stream request, as the server will send a stream-end
		// afterwards.
		atomic.AddUint64(&d.stats.TotUPRCloseStreamResOk, 1)

	case gomemcached.UPR_SNAPSHOT:
		atomic.AddUint64(&d.stats.TotUPRSnapshot, 1)

		vbucketID := uint16(res.Status)
		vbucketState := currVBuckets[vbucketID]

		if vbucketState == nil || vbucketState.State != "running" {
			atomic.AddUint64(&d.stats.TotUPRSnapshotStateErr, 1)
			return fmt.Errorf("snapshot non-running,"+
				" vbucketID: %d, vbucketState: %#v, res: %#v",
				vbucketID, vbucketState, res)
		}

		if len(res.Extras) < 20 {
			return fmt.Errorf("bad snapshot extras, res: %#v", res)
		}

		vbucketState.SnapStart = binary.BigEndian.Uint64(res.Extras[0:8])
		vbucketState.SnapEnd = binary.BigEndian.Uint64(res.Extras[8:16])
		vbucketState.SnapSaved = false

		snapType := binary.BigEndian.Uint32(res.Extras[16:20])

		// NOTE: We should never see a snapType with SNAP_ACK flag of
		// true, as that's only used during takeovers, so that's why
		// we're not implementing SNAP_ACK handling here.

		atomic.AddUint64(&d.stats.TotUPRSnapshotStart, 1)
		err := d.receiver.SnapshotStart(vbucketID,
			vbucketState.SnapStart, vbucketState.SnapEnd, snapType)
		if err != nil {
			atomic.AddUint64(&d.stats.TotUPRSnapshotStartErr, 1)
			return err
		}

		atomic.AddUint64(&d.stats.TotUPRSnapshotOk, 1)

	case gomemcached.UPR_CONTROL:
		atomic.AddUint64(&d.stats.TotUPRControl, 1)
		if res.Status != gomemcached.SUCCESS {
			atomic.AddUint64(&d.stats.TotUPRControlErr, 1)
			return fmt.Errorf("failed control: %#v", res)
		}

	case gomemcached.UPR_OPEN:
		// Opening was long ago, so we should not see UPR_OPEN responses.
		return fmt.Errorf("unexpected upr_open, res: %#v", res)

	case gomemcached.UPR_ADDSTREAM:
		// This normally comes from ns-server / dcp-migrator.
		return fmt.Errorf("unexpected upr_addstream, res: %#v", res)

	case gomemcached.UPR_BUFFERACK:
		// We should be emitting buffer-ack's, not receiving them.
		return fmt.Errorf("unexpected buffer-ack, res: %#v", res)

	case gomemcached.UPR_MUTATION, gomemcached.UPR_DELETION, gomemcached.UPR_EXPIRATION:
		// This should have been handled already in receiver goroutine.
		return fmt.Errorf("unexpected data change, res: %#v", res)

	case gomemcached.NOOP:
		// We use NOOP for "keep alive" ping messages.
		return nil

	default:
		return fmt.Errorf("unknown opcode, res: %#v", res)
	}

	return nil
}

func (d *bucketDataSource) getVBucketMetaData(vbucketID uint16) (
	*VBucketMetaData, uint64, error) {
	atomic.AddUint64(&d.stats.TotGetVBucketMetaData, 1)

	buf, lastSeq, err := d.receiver.GetMetaData(vbucketID)
	if err != nil {
		atomic.AddUint64(&d.stats.TotGetVBucketMetaDataErr, 1)
		return nil, 0, err
	}

	vbucketMetaData := &VBucketMetaData{}
	if len(buf) > 0 {
		if err = json.Unmarshal(buf, vbucketMetaData); err != nil {
			atomic.AddUint64(&d.stats.TotGetVBucketMetaDataUnmarshalErr, 1)
			return nil, 0, fmt.Errorf("could not parse vbucketMetaData,"+
				" buf: %q, err: %v", buf, err)
		}
	}

	atomic.AddUint64(&d.stats.TotGetVBucketMetaDataOk, 1)
	return vbucketMetaData, lastSeq, nil
}

func (d *bucketDataSource) setVBucketMetaData(vbucketID uint16,
	v *VBucketMetaData) error {
	atomic.AddUint64(&d.stats.TotSetVBucketMetaData, 1)

	buf, err := json.Marshal(v)
	if err != nil {
		atomic.AddUint64(&d.stats.TotSetVBucketMetaDataMarshalErr, 1)
		return err
	}

	err = d.receiver.SetMetaData(vbucketID, buf)
	if err != nil {
		atomic.AddUint64(&d.stats.TotSetVBucketMetaDataErr, 1)
		return err
	}

	atomic.AddUint64(&d.stats.TotSetVBucketMetaDataOk, 1)
	return nil
}

func (d *bucketDataSource) sendStreamReq(sendCh chan interface{},
	vbucketID uint16) error {
	vbucketMetaData, lastSeq, err := d.getVBucketMetaData(vbucketID)
	if err != nil {
		return fmt.Errorf("sendStreamReq, err: %v", err)
	}

	vbucketUUID := uint64(0)
	if len(vbucketMetaData.FailOverLog) >= 1 {
		smax := uint64(0)
		for _, pair := range vbucketMetaData.FailOverLog {
			if smax <= pair[1] {
				smax = pair[1]
				vbucketUUID = pair[0]
			}
		}
	}

	seqStart := lastSeq

	seqEnd := uint64(0xffffffffffffffff)
	if d.options.SeqEnd != nil { // Allow apps like backup to control the seqEnd.
		if s, exists := d.options.SeqEnd[vbucketID]; exists {
			seqEnd = s
		}
	}

	flags := uint32(0) // Flags mostly used for takeovers, etc, which we don't use.

	req := &gomemcached.MCRequest{
		Opcode:  gomemcached.UPR_STREAMREQ,
		VBucket: vbucketID,
		Opaque:  uint32(vbucketID),
		Extras:  make([]byte, 48),
	}
	binary.BigEndian.PutUint32(req.Extras[:4], flags)
	binary.BigEndian.PutUint32(req.Extras[4:8], uint32(0)) // Reserved.
	binary.BigEndian.PutUint64(req.Extras[8:16], seqStart)
	binary.BigEndian.PutUint64(req.Extras[16:24], seqEnd)
	binary.BigEndian.PutUint64(req.Extras[24:32], vbucketUUID)
	binary.BigEndian.PutUint64(req.Extras[32:40], vbucketMetaData.SnapStart)
	binary.BigEndian.PutUint64(req.Extras[40:48], vbucketMetaData.SnapEnd)

	atomic.AddUint64(&d.stats.TotUPRStreamReq, 1)
	sendCh <- req

	return nil
}

func (d *bucketDataSource) Stats(dest *BucketDataSourceStats) error {
	d.stats.AtomicCopyTo(dest, nil)
	return nil
}

func (d *bucketDataSource) Close() error {
	d.m.Lock()
	if d.life != "running" {
		d.m.Unlock()
		return fmt.Errorf("call to Close() when not running state: %s", d.life)
	}
	d.life = "closed"
	d.m.Unlock()

	// Stopping the refreshClusterCh's goroutine will end
	// refreshWorkersCh's goroutine, which closes every workerCh and
	// then finally closes the closedCh.
	close(d.stopCh)

	<-d.closedCh

	// TODO: By this point, worker goroutines may still be going, but
	// should end soon.  Instead Close() should be 100% synchronous.
	return nil
}

func (d *bucketDataSource) Kick(reason string) error {
	if d.isRunning() {
		atomic.AddUint64(&d.stats.TotKick, 1)

		var refreshClusterCh chan struct{}

		d.refreshClusterM.Lock()
		if len(d.refreshClusterReasons) <= 0 {
			// Transitioning from empty to non-empty, so need to notify.
			refreshClusterCh = d.refreshClusterCh
		}
		d.refreshClusterReasons[reason] += 1
		d.refreshClusterM.Unlock()

		if refreshClusterCh == nil {
			atomic.AddUint64(&d.stats.TotKickDeduped, 1)
		} else {
			go func() {
				refreshClusterCh <- struct{}{}

				atomic.AddUint64(&d.stats.TotKickOk, 1)
			}()
		}
	}

	return nil
}

// --------------------------------------------------------------

func (d *bucketDataSource) logVBucketStates(server, uprOpenName, prefix string,
	vbucketStates map[uint16]*VBucketState, errToLog error) {
	if d.options.Logf == nil {
		return
	}

	if len(vbucketStates) <= 0 {
		d.options.Logf("cbdatasource: server: %s, uprOpenName: %s, %s,"+
			" vbucketStates empty", server, uprOpenName, prefix)
	}

	// Key is VBucketState.State (ex: "", "running", "closing"),
	// and value is array of vbid's.
	byState := map[string]sort.IntSlice{}
	for vbid, vbucketState := range vbucketStates {
		byState[vbucketState.State] =
			append(byState[vbucketState.State], int(vbid))
	}

	for state, vbids := range byState {
		var buf bytes.Buffer

		if len(vbids) > 0 {
			first := true
			emitRange := func(beg, end int) {
				if !first {
					fmt.Fprint(&buf, ", ")
				}
				if beg < end {
					fmt.Fprintf(&buf, "%d-%d", beg, end)
				} else {
					fmt.Fprintf(&buf, "%d", beg)
				}
				first = false
			}

			sort.Sort(vbids)

			vbidRangeBeg := vbids[0]
			vbidRangeEnd := vbids[0]

			for i := 1; i < len(vbids); i++ {
				vbid := vbids[i]
				if vbid > vbidRangeEnd+1 {
					emitRange(vbidRangeBeg, vbidRangeEnd)
					vbidRangeBeg = vbid
				}
				vbidRangeEnd = vbid
			}

			emitRange(vbidRangeBeg, vbidRangeEnd)
		}

		d.options.Logf("cbdatasource: server: %s, uprOpenName: %s, %s,"+
			" vbucketState: %q (has %d vbuckets), %s",
			server, uprOpenName, prefix, state, len(vbids), buf.String())
	}
}

// --------------------------------------------------------------

type bucketWrapper struct {
	b *couchbase.Bucket
}

func (bw *bucketWrapper) Close() {
	bw.b.Close()
}

func (bw *bucketWrapper) GetUUID() string {
	return bw.b.UUID
}

func (bw *bucketWrapper) VBServerMap() *couchbase.VBucketServerMap {
	return bw.b.VBServerMap()
}

// ConnectBucket is the default function used by BucketDataSource
// to connect to a Couchbase cluster to retrieve Bucket information.
// It is exposed for testability and to allow applications to
// override or wrap via BucketDataSourceOptions.
func ConnectBucket(serverURL, poolName, bucketName string,
	auth couchbase.AuthHandler) (Bucket, error) {
	var bucket *couchbase.Bucket
	var err error
	var client couchbase.Client
	var pool couchbase.Pool

	if auth != nil {
		client, err = couchbase.ConnectWithAuth(serverURL, auth)
		if err != nil {
			return nil, err
		}

		pool, err = client.GetPool(poolName)
		if err != nil {
			return nil, err
		}

		bucket, err = pool.GetBucket(bucketName)
	} else {
		bucket, err = couchbase.GetBucket(serverURL, poolName, bucketName)
	}
	if err != nil {
		return nil, err
	}
	if bucket == nil {
		return nil, fmt.Errorf("unknown bucket,"+
			" serverURL: %s, bucketName: %s", serverURL, bucketName)
	}

	return &bucketWrapper{b: bucket}, nil
}

// Select bucket for this connection.
func selectBucket(mc *memcached.Client, bucketName string) error {
	_, err := mc.Send(&gomemcached.MCRequest{
		Opcode: gomemcached.SELECT_BUCKET,
		Key:    []byte(bucketName),
	})
	return err
}

func serverHandShake(mc *memcached.Client, d *bucketDataSource) (uint32, error) {
	if d.options.IncludeXAttrs {
		// If the server supports xattrs, then need to include the
		// xattrs flag in open dcp request.
		err := xAttrsSupported(sendHelo(mc, d))
		if err != nil {
			return 0, err
		}
		return FlagOpenIncludeXattrs, nil
	}
	_, err := sendHelo(mc, d)
	return 0, err
}

func sendHelo(mc *memcached.Client, d *bucketDataSource) (*gomemcached.MCResponse, error) {
	size := 4
	if d.options.IncludeXAttrs {
		size = 6
	}
	payload := make([]byte, size)
	binary.BigEndian.PutUint16(payload[0:2], FeatureEnabledDataType)
	binary.BigEndian.PutUint16(payload[2:4], FeatureEnabledXError)
	if d.options.IncludeXAttrs {
		binary.BigEndian.PutUint16(payload[4:6], FeatureEnabledXAttrs)
	}
	return mc.Send(&gomemcached.MCRequest{
		Opcode: gomemcached.HELLO,
		Key:    []byte(fmt.Sprintf("cbdatasource-%s", d.options.Name)),
		Body:   payload,
	})
}

func xAttrsSupported(res *gomemcached.MCResponse, err error) error {
	if err != nil {
		return err
	}
	// Can't assume any ordering of response bytes since we may be
	// connecting to an older server with partial feature support.
	if len(res.Body) < 2 {
		return ErrXAttrsNotSupported
	}
	feature1 := binary.BigEndian.Uint16(res.Body[0:2])
	if feature1 == FeatureEnabledXAttrs {
		return nil
	}
	if len(res.Body) < 4 {
		return ErrXAttrsNotSupported
	}
	feature2 := binary.BigEndian.Uint16(res.Body[2:4])
	if feature2 == FeatureEnabledXAttrs {
		return nil
	}
	return ErrXAttrsNotSupported
}

// UPROpen starts a UPR_OPEN stream on a memcached client connection.
// It is exposed for testability.
func UPROpen(mc *memcached.Client, name string,
	option *BucketDataSourceOptions, openFlags uint32) error {
	rq := &gomemcached.MCRequest{
		Opcode: gomemcached.UPR_OPEN,
		Key:    []byte(name),
		Opaque: 0xf00d1234,
		Extras: make([]byte, 8),
	}
	bufSize := option.FeedBufferSizeBytes
	noopInterval := option.NoopTimeIntervalSecs

	binary.BigEndian.PutUint32(rq.Extras[:4], 0) // First 4 bytes are reserved.
	flags := FlagOpenProducer | openFlags        // NOTE: 1 for producer, 0 for consumer.
	binary.BigEndian.PutUint32(rq.Extras[4:], flags)

	if err := mc.Transmit(rq); err != nil {
		return fmt.Errorf("UPROpen transmit, err: %v", err)
	}

	res, err := mc.Receive()
	if err != nil {
		return fmt.Errorf("UPROpen receive, err: %v", err)
	}
	if res.Opcode != gomemcached.UPR_OPEN {
		return fmt.Errorf("UPROpen unexpected #opcode %v", res.Opcode)
	}
	if res.Opaque != rq.Opaque {
		return fmt.Errorf("UPROpen opaque mismatch, %v over %v", res.Opaque, res.Opaque)
	}
	if res.Status != gomemcached.SUCCESS {
		return fmt.Errorf("UPROpen failed, status: %v, %#v", res.Status, res)
	}

	if bufSize > 0 {
		rq := &gomemcached.MCRequest{
			Opcode: gomemcached.UPR_CONTROL,
			Key:    []byte("connection_buffer_size"),
			Body:   []byte(strconv.Itoa(int(bufSize))),
		}
		if err = mc.Transmit(rq); err != nil {
			return fmt.Errorf("UPROpen transmit UPR_CONTROL"+
				" (connection_buffer_size), err: %v", err)
		}
	}

	if noopInterval > 0 {
		rq := &gomemcached.MCRequest{
			Opcode: gomemcached.UPR_CONTROL,
			Key:    []byte("enable_noop"),
			Body:   []byte("true"),
		}
		if err = mc.Transmit(rq); err != nil {
			return fmt.Errorf("UPROpen transmit UPR_CONTROL"+
				" (enable_noop), err: %v", err)
		}

		rq = &gomemcached.MCRequest{
			Opcode: gomemcached.UPR_CONTROL,
			Key:    []byte("set_noop_interval"),
			Body:   []byte(strconv.Itoa(int(noopInterval))),
		}
		if err = mc.Transmit(rq); err != nil {
			return fmt.Errorf("UPROpen transmit UPR_CONTROL"+
				" (set_noop_interval), err: %v", err)
		}
	}

	return nil
}

// ParseFailOverLog parses a byte array to an array of [vbucketUUID,
// seqNum] pairs.  It is exposed for testability.
func ParseFailOverLog(body []byte) ([][]uint64, error) {
	if len(body)%16 != 0 {
		return nil, fmt.Errorf("invalid body length %v, in failover-log", len(body))
	}
	flog := make([][]uint64, len(body)/16)
	for i, j := 0, 0; i < len(body); i += 16 {
		uuid := binary.BigEndian.Uint64(body[i : i+8])
		seqn := binary.BigEndian.Uint64(body[i+8 : i+16])
		flog[j] = []uint64{uuid, seqn}
		j++
	}
	return flog, nil
}

// --------------------------------------------------------------

// AtomicCopyTo copies metrics from s to r (or, from source to
// result), and also applies an optional fn function.  The fn is
// invoked with metrics from s and r, and can be used to compute
// additions, subtractions, negations, etc.  When fn is nil,
// AtomicCopyTo behaves as a straight copier.
func (s *BucketDataSourceStats) AtomicCopyTo(r *BucketDataSourceStats,
	fn func(sv uint64, rv uint64) uint64) {
	// Using reflection rather than a whole slew of explicit
	// invocations of atomic.LoadUint64()/StoreUint64()'s.
	if fn == nil {
		fn = func(sv uint64, rv uint64) uint64 { return sv }
	}
	rve := reflect.ValueOf(r).Elem()
	sve := reflect.ValueOf(s).Elem()
	svet := sve.Type()
	for i := 0; i < svet.NumField(); i++ {
		rvef := rve.Field(i)
		svef := sve.Field(i)
		if rvef.CanAddr() && svef.CanAddr() {
			rvefp := rvef.Addr().Interface()
			svefp := svef.Addr().Interface()
			rv := atomic.LoadUint64(rvefp.(*uint64))
			sv := atomic.LoadUint64(svefp.(*uint64))
			atomic.StoreUint64(rvefp.(*uint64), fn(sv, rv))
		}
	}
}

// --------------------------------------------------------------

// ExponentialBackoffLoop invokes f() in a loop, sleeping an
// exponentially growing number of milliseconds in between invocations
// if needed.  The provided f() function should return < 0 to stop the
// loop; >= 0 to continue the loop, where > 0 means there was progress
// which allows an immediate retry of f() with no sleeping.  A return
// of < 0 is useful when f() will never make any future progress.
// Repeated attempts with no progress will have exponentially growing
// backoff sleep times.
func ExponentialBackoffLoop(name string,
	f func() int,
	startSleepMS int,
	backoffFactor float32,
	maxSleepMS int) {
	nextSleepMS := startSleepMS
	for {
		progress := f()
		if progress < 0 {
			return
		}
		if progress > 0 {
			// When there was some progress, we can reset nextSleepMS.
			nextSleepMS = startSleepMS
		} else {
			// If zero progress was made this cycle, then sleep.
			time.Sleep(time.Duration(nextSleepMS) * time.Millisecond)

			// Increase nextSleepMS in case next time also has 0 progress.
			nextSleepMS = int(float32(nextSleepMS) * backoffFactor)
			if nextSleepMS > maxSleepMS {
				nextSleepMS = maxSleepMS
			}
		}
	}
}
