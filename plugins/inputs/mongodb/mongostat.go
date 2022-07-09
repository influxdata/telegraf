/***
The code contained here came from https://github.com/mongodb/mongo-tools/blob/master/mongostat/stat_types.go
and contains modifications so that no other dependency from that project is needed. Other modifications included
removing unnecessary code specific to formatting the output and determine the current state of the database. It
is licensed under Apache Version 2.0, http://www.apache.org/licenses/LICENSE-2.0.html
***/

package mongodb

import (
	"sort"
	"strings"
	"time"
)

const (
	MongosProcess = "mongos"
)

// Flags to determine cases when to activate/deactivate columns for output.
const (
	Always   = 1 << iota // always activate the column
	Discover             // only active when mongostat is in discover mode
	Repl                 // only active if one of the nodes being monitored is in a replset
	Locks                // only active if node is capable of calculating lock info
	AllOnly              // only active if mongostat was run with --all option
	MMAPOnly             // only active if node has mmap-specific fields
	WTOnly               // only active if node has wiredtiger-specific fields
)

type MongoStatus struct {
	SampleTime    time.Time
	ServerStatus  *ServerStatus
	ReplSetStatus *ReplSetStatus
	ClusterStatus *ClusterStatus
	DbStats       *DbStats
	ColStats      *ColStats
	ShardStats    *ShardStats
	OplogStats    *OplogStats
	TopStats      *TopStats
}

type ServerStatus struct {
	SampleTime         time.Time              `bson:""`
	Flattened          map[string]interface{} `bson:""`
	Host               string                 `bson:"host"`
	Version            string                 `bson:"version"`
	Process            string                 `bson:"process"`
	Pid                int64                  `bson:"pid"`
	Uptime             int64                  `bson:"uptime"`
	UptimeMillis       int64                  `bson:"uptimeMillis"`
	UptimeEstimate     int64                  `bson:"uptimeEstimate"`
	LocalTime          time.Time              `bson:"localTime"`
	Asserts            *AssertsStats          `bson:"asserts"`
	BackgroundFlushing *FlushStats            `bson:"backgroundFlushing"`
	ExtraInfo          *ExtraInfo             `bson:"extra_info"`
	Connections        *ConnectionStats       `bson:"connections"`
	Dur                *DurStats              `bson:"dur"`
	GlobalLock         *GlobalLockStats       `bson:"globalLock"`
	Locks              map[string]LockStats   `bson:"locks,omitempty"`
	Network            *NetworkStats          `bson:"network"`
	Opcounters         *OpcountStats          `bson:"opcounters"`
	OpcountersRepl     *OpcountStats          `bson:"opcountersRepl"`
	OpLatencies        *OpLatenciesStats      `bson:"opLatencies"`
	RecordStats        *DBRecordStats         `bson:"recordStats"`
	Mem                *MemStats              `bson:"mem"`
	Repl               *ReplStatus            `bson:"repl"`
	ShardCursorType    map[string]interface{} `bson:"shardCursorType"`
	StorageEngine      *StorageEngine         `bson:"storageEngine"`
	WiredTiger         *WiredTiger            `bson:"wiredTiger"`
	Metrics            *MetricsStats          `bson:"metrics"`
	TCMallocStats      *TCMallocStats         `bson:"tcmalloc"`
}

// DbStats stores stats from all dbs
type DbStats struct {
	Dbs []Db
}

// Db represent a single DB
type Db struct {
	Name        string
	DbStatsData *DbStatsData
}

// DbStatsData stores stats from a db
type DbStatsData struct {
	Db          string      `bson:"db"`
	Collections int64       `bson:"collections"`
	Objects     int64       `bson:"objects"`
	AvgObjSize  float64     `bson:"avgObjSize"`
	DataSize    int64       `bson:"dataSize"`
	StorageSize int64       `bson:"storageSize"`
	NumExtents  int64       `bson:"numExtents"`
	Indexes     int64       `bson:"indexes"`
	IndexSize   int64       `bson:"indexSize"`
	Ok          int64       `bson:"ok"`
	GleStats    interface{} `bson:"gleStats"`
	FsUsedSize  int64       `bson:"fsUsedSize"`
	FsTotalSize int64       `bson:"fsTotalSize"`
}

type ColStats struct {
	Collections []Collection
}

type Collection struct {
	Name         string
	DbName       string
	ColStatsData *ColStatsData
}

type ColStatsData struct {
	Collection     string  `bson:"ns"`
	Count          int64   `bson:"count"`
	Size           int64   `bson:"size"`
	AvgObjSize     float64 `bson:"avgObjSize"`
	StorageSize    int64   `bson:"storageSize"`
	TotalIndexSize int64   `bson:"totalIndexSize"`
	Ok             int64   `bson:"ok"`
}

// ClusterStatus stores information related to the whole cluster
type ClusterStatus struct {
	JumboChunksCount int64
}

// ReplSetStatus stores information from replSetGetStatus
type ReplSetStatus struct {
	Members []ReplSetMember `bson:"members"`
	MyState int64           `bson:"myState"`
}

// OplogStatus stores information from getReplicationInfo
type OplogStats struct {
	TimeDiff int64
}

// ReplSetMember stores information related to a replica set member
type ReplSetMember struct {
	Name       string    `bson:"name"`
	State      int64     `bson:"state"`
	StateStr   string    `bson:"stateStr"`
	OptimeDate time.Time `bson:"optimeDate"`
}

// WiredTiger stores information related to the WiredTiger storage engine.
type WiredTiger struct {
	Transaction TransactionStats       `bson:"transaction"`
	Concurrent  ConcurrentTransactions `bson:"concurrentTransactions"`
	Cache       CacheStats             `bson:"cache"`
	Connection  WTConnectionStats      `bson:"connection"`
	DataHandle  DataHandleStats        `bson:"data-handle"`
}

// ShardStats stores information from shardConnPoolStats.
type ShardStats struct {
	ShardStatsData `bson:",inline"`
	Hosts          map[string]ShardHostStatsData `bson:"hosts"`
}

// ShardStatsData is the total Shard Stats from shardConnPoolStats database command.
type ShardStatsData struct {
	TotalInUse      int64 `bson:"totalInUse"`
	TotalAvailable  int64 `bson:"totalAvailable"`
	TotalCreated    int64 `bson:"totalCreated"`
	TotalRefreshing int64 `bson:"totalRefreshing"`
}

// ShardHostStatsData is the host-specific stats
// from shardConnPoolStats database command.
type ShardHostStatsData struct {
	InUse      int64 `bson:"inUse"`
	Available  int64 `bson:"available"`
	Created    int64 `bson:"created"`
	Refreshing int64 `bson:"refreshing"`
}

type TopStats struct {
	Totals map[string]TopStatCollection `bson:"totals"`
}

type TopStatCollection struct {
	Total     TopStatCollectionData `bson:"total"`
	ReadLock  TopStatCollectionData `bson:"readLock"`
	WriteLock TopStatCollectionData `bson:"writeLock"`
	Queries   TopStatCollectionData `bson:"queries"`
	GetMore   TopStatCollectionData `bson:"getmore"`
	Insert    TopStatCollectionData `bson:"insert"`
	Update    TopStatCollectionData `bson:"update"`
	Remove    TopStatCollectionData `bson:"remove"`
	Commands  TopStatCollectionData `bson:"commands"`
}

type TopStatCollectionData struct {
	Time  int64 `bson:"time"`
	Count int64 `bson:"count"`
}

type ConcurrentTransactions struct {
	Write ConcurrentTransStats `bson:"write"`
	Read  ConcurrentTransStats `bson:"read"`
}

type ConcurrentTransStats struct {
	Out          int64 `bson:"out"`
	Available    int64 `bson:"available"`
	TotalTickets int64 `bson:"totalTickets"`
}

// AssertsStats stores information related to assertions raised since the MongoDB process started
type AssertsStats struct {
	Regular   int64 `bson:"regular"`
	Warning   int64 `bson:"warning"`
	Msg       int64 `bson:"msg"`
	User      int64 `bson:"user"`
	Rollovers int64 `bson:"rollovers"`
}

// CacheStats stores cache statistics for WiredTiger.
type CacheStats struct {
	TrackedDirtyBytes         int64 `bson:"tracked dirty bytes in the cache"`
	CurrentCachedBytes        int64 `bson:"bytes currently in the cache"`
	MaxBytesConfigured        int64 `bson:"maximum bytes configured"`
	AppThreadsPageReadCount   int64 `bson:"application threads page read from disk to cache count"`
	AppThreadsPageReadTime    int64 `bson:"application threads page read from disk to cache time (usecs)"`
	AppThreadsPageWriteCount  int64 `bson:"application threads page write from cache to disk count"`
	AppThreadsPageWriteTime   int64 `bson:"application threads page write from cache to disk time (usecs)"`
	BytesWrittenFrom          int64 `bson:"bytes written from cache"`
	BytesReadInto             int64 `bson:"bytes read into cache"`
	PagesEvictedByAppThread   int64 `bson:"pages evicted by application threads"`
	PagesQueuedForEviction    int64 `bson:"pages queued for eviction"`
	PagesReadIntoCache        int64 `bson:"pages read into cache"`
	PagesWrittenFromCache     int64 `bson:"pages written from cache"`
	PagesRequestedFromCache   int64 `bson:"pages requested from the cache"`
	ServerEvictingPages       int64 `bson:"eviction server evicting pages"`
	WorkerThreadEvictingPages int64 `bson:"eviction worker thread evicting pages"`
	InternalPagesEvicted      int64 `bson:"internal pages evicted"`
	ModifiedPagesEvicted      int64 `bson:"modified pages evicted"`
	UnmodifiedPagesEvicted    int64 `bson:"unmodified pages evicted"`
}

type StorageEngine struct {
	Name string `bson:"name"`
}

// TransactionStats stores transaction checkpoints in WiredTiger.
type TransactionStats struct {
	TransCheckpointsTotalTimeMsecs int64 `bson:"transaction checkpoint total time (msecs)"`
	TransCheckpoints               int64 `bson:"transaction checkpoints"`
}

// WTConnectionStats stores statistices on wiredTiger connections
type WTConnectionStats struct {
	FilesCurrentlyOpen int64 `bson:"files currently open"`
}

// DataHandleStats stores statistics for wiredTiger data-handles
type DataHandleStats struct {
	DataHandlesCurrentlyActive int64 `bson:"connection data handles currently active"`
}

// ReplStatus stores data related to replica sets.
type ReplStatus struct {
	SetName           string      `bson:"setName"`
	IsWritablePrimary interface{} `bson:"isWritablePrimary"` // mongodb 5.x
	IsMaster          interface{} `bson:"ismaster"`
	Secondary         interface{} `bson:"secondary"`
	IsReplicaSet      interface{} `bson:"isreplicaset"`
	ArbiterOnly       interface{} `bson:"arbiterOnly"`
	Hosts             []string    `bson:"hosts"`
	Passives          []string    `bson:"passives"`
	Me                string      `bson:"me"`
}

// DBRecordStats stores data related to memory operations across databases.
type DBRecordStats struct {
	AccessesNotInMemory       int64                     `bson:"accessesNotInMemory"`
	PageFaultExceptionsThrown int64                     `bson:"pageFaultExceptionsThrown"`
	DBRecordAccesses          map[string]RecordAccesses `bson:",inline"`
}

// RecordAccesses stores data related to memory operations scoped to a database.
type RecordAccesses struct {
	AccessesNotInMemory       int64 `bson:"accessesNotInMemory"`
	PageFaultExceptionsThrown int64 `bson:"pageFaultExceptionsThrown"`
}

// MemStats stores data related to memory statistics.
type MemStats struct {
	Bits              int64       `bson:"bits"`
	Resident          int64       `bson:"resident"`
	Virtual           int64       `bson:"virtual"`
	Supported         interface{} `bson:"supported"`
	Mapped            int64       `bson:"mapped"`
	MappedWithJournal int64       `bson:"mappedWithJournal"`
}

// FlushStats stores information about memory flushes.
type FlushStats struct {
	Flushes      int64     `bson:"flushes"`
	TotalMs      int64     `bson:"total_ms"`
	AverageMs    float64   `bson:"average_ms"`
	LastMs       int64     `bson:"last_ms"`
	LastFinished time.Time `bson:"last_finished"`
}

// ConnectionStats stores information related to incoming database connections.
type ConnectionStats struct {
	Current      int64 `bson:"current"`
	Available    int64 `bson:"available"`
	TotalCreated int64 `bson:"totalCreated"`
}

// DurTiming stores information related to journaling.
type DurTiming struct {
	Dt               int64 `bson:"dt"`
	PrepLogBuffer    int64 `bson:"prepLogBuffer"`
	WriteToJournal   int64 `bson:"writeToJournal"`
	WriteToDataFiles int64 `bson:"writeToDataFiles"`
	RemapPrivateView int64 `bson:"remapPrivateView"`
}

// DurStats stores information related to journaling statistics.
type DurStats struct {
	Commits            int64 `bson:"commits"`
	JournaledMB        int64 `bson:"journaledMB"`
	WriteToDataFilesMB int64 `bson:"writeToDataFilesMB"`
	Compression        int64 `bson:"compression"`
	CommitsInWriteLock int64 `bson:"commitsInWriteLock"`
	EarlyCommits       int64 `bson:"earlyCommits"`
	TimeMs             DurTiming
}

// QueueStats stores the number of queued read/write operations.
type QueueStats struct {
	Total   int64 `bson:"total"`
	Readers int64 `bson:"readers"`
	Writers int64 `bson:"writers"`
}

// ClientStats stores the number of active read/write operations.
type ClientStats struct {
	Total   int64 `bson:"total"`
	Readers int64 `bson:"readers"`
	Writers int64 `bson:"writers"`
}

// GlobalLockStats stores information related locks in the MMAP storage engine.
type GlobalLockStats struct {
	TotalTime     int64        `bson:"totalTime"`
	LockTime      int64        `bson:"lockTime"`
	CurrentQueue  *QueueStats  `bson:"currentQueue"`
	ActiveClients *ClientStats `bson:"activeClients"`
}

// NetworkStats stores information related to network traffic.
type NetworkStats struct {
	BytesIn     int64 `bson:"bytesIn"`
	BytesOut    int64 `bson:"bytesOut"`
	NumRequests int64 `bson:"numRequests"`
}

// OpcountStats stores information related to commands and basic CRUD operations.
type OpcountStats struct {
	Insert  int64 `bson:"insert"`
	Query   int64 `bson:"query"`
	Update  int64 `bson:"update"`
	Delete  int64 `bson:"delete"`
	GetMore int64 `bson:"getmore"`
	Command int64 `bson:"command"`
}

// OpLatenciesStats stores information related to operation latencies for the database as a whole
type OpLatenciesStats struct {
	Reads    *LatencyStats `bson:"reads"`
	Writes   *LatencyStats `bson:"writes"`
	Commands *LatencyStats `bson:"commands"`
}

// LatencyStats lists total latency in microseconds and count of operations, enabling you to obtain an average
type LatencyStats struct {
	Latency int64 `bson:"latency"`
	Ops     int64 `bson:"ops"`
}

// MetricsStats stores information related to metrics
type MetricsStats struct {
	TTL           *TTLStats           `bson:"ttl"`
	Cursor        *CursorStats        `bson:"cursor"`
	Document      *DocumentStats      `bson:"document"`
	Commands      *CommandsStats      `bson:"commands"`
	Operation     *OperationStats     `bson:"operation"`
	QueryExecutor *QueryExecutorStats `bson:"queryExecutor"`
	Repl          *ReplStats          `bson:"repl"`
	Storage       *StorageStats       `bson:"storage"`
}

// TTLStats stores information related to documents with a ttl index.
type TTLStats struct {
	DeletedDocuments int64 `bson:"deletedDocuments"`
	Passes           int64 `bson:"passes"`
}

// CursorStats stores information related to cursor metrics.
type CursorStats struct {
	TimedOut int64            `bson:"timedOut"`
	Open     *OpenCursorStats `bson:"open"`
}

// DocumentStats stores information related to document metrics.
type DocumentStats struct {
	Deleted  int64 `bson:"deleted"`
	Inserted int64 `bson:"inserted"`
	Returned int64 `bson:"returned"`
	Updated  int64 `bson:"updated"`
}

// CommandsStats stores information related to document metrics.
type CommandsStats struct {
	Aggregate     *CommandsStatsValue `bson:"aggregate"`
	Count         *CommandsStatsValue `bson:"count"`
	Delete        *CommandsStatsValue `bson:"delete"`
	Distinct      *CommandsStatsValue `bson:"distinct"`
	Find          *CommandsStatsValue `bson:"find"`
	FindAndModify *CommandsStatsValue `bson:"findAndModify"`
	GetMore       *CommandsStatsValue `bson:"getMore"`
	Insert        *CommandsStatsValue `bson:"insert"`
	Update        *CommandsStatsValue `bson:"update"`
}

type CommandsStatsValue struct {
	Failed int64 `bson:"failed"`
	Total  int64 `bson:"total"`
}

// OpenCursorStats stores information related to open cursor metrics
type OpenCursorStats struct {
	NoTimeout int64 `bson:"noTimeout"`
	Pinned    int64 `bson:"pinned"`
	Total     int64 `bson:"total"`
}

// OperationStats stores information related to query operations
// using special operation types
type OperationStats struct {
	ScanAndOrder   int64 `bson:"scanAndOrder"`
	WriteConflicts int64 `bson:"writeConflicts"`
}

// QueryExecutorStats stores information related to query execution
type QueryExecutorStats struct {
	Scanned        int64 `bson:"scanned"`
	ScannedObjects int64 `bson:"scannedObjects"`
}

// ReplStats stores information related to replication process
type ReplStats struct {
	Apply    *ReplApplyStats    `bson:"apply"`
	Buffer   *ReplBufferStats   `bson:"buffer"`
	Executor *ReplExecutorStats `bson:"executor,omitempty"`
	Network  *ReplNetworkStats  `bson:"network"`
}

// ReplApplyStats stores information related to oplog application process
type ReplApplyStats struct {
	Batches *BasicStats `bson:"batches"`
	Ops     int64       `bson:"ops"`
}

// ReplBufferStats stores information related to oplog buffer
type ReplBufferStats struct {
	Count     int64 `bson:"count"`
	SizeBytes int64 `bson:"sizeBytes"`
}

// ReplExecutorStats stores information related to replication executor
type ReplExecutorStats struct {
	Pool             map[string]int64 `bson:"pool"`
	Queues           map[string]int64 `bson:"queues"`
	UnsignaledEvents int64            `bson:"unsignaledEvents"`
}

// ReplNetworkStats stores information related to network usage by replication process
type ReplNetworkStats struct {
	Bytes    int64       `bson:"bytes"`
	GetMores *BasicStats `bson:"getmores"`
	Ops      int64       `bson:"ops"`
}

// BasicStats stores information about an operation
type BasicStats struct {
	Num         int64 `bson:"num"`
	TotalMillis int64 `bson:"totalMillis"`
}

// ReadWriteLockTimes stores time spent holding read/write locks.
type ReadWriteLockTimes struct {
	Read       int64 `bson:"R"`
	Write      int64 `bson:"W"`
	ReadLower  int64 `bson:"r"`
	WriteLower int64 `bson:"w"`
}

// LockStats stores information related to time spent acquiring/holding locks
// for a given database.
type LockStats struct {
	TimeLockedMicros    ReadWriteLockTimes `bson:"timeLockedMicros"`
	TimeAcquiringMicros ReadWriteLockTimes `bson:"timeAcquiringMicros"`

	// AcquireCount and AcquireWaitCount are new fields of the lock stats only populated on 3.0 or newer.
	// Typed as a pointer so that if it is nil, mongostat can assume the field is not populated
	// with real namespace data.
	AcquireCount     *ReadWriteLockTimes `bson:"acquireCount,omitempty"`
	AcquireWaitCount *ReadWriteLockTimes `bson:"acquireWaitCount,omitempty"`
}

// ExtraInfo stores additional platform specific information.
type ExtraInfo struct {
	PageFaults *int64 `bson:"page_faults"`
}

// TCMallocStats stores information related to TCMalloc memory allocator metrics
type TCMallocStats struct {
	Generic  *GenericTCMAllocStats  `bson:"generic"`
	TCMalloc *DetailedTCMallocStats `bson:"tcmalloc"`
}

// GenericTCMAllocStats stores generic TCMalloc memory allocator metrics
type GenericTCMAllocStats struct {
	CurrentAllocatedBytes int64 `bson:"current_allocated_bytes"`
	HeapSize              int64 `bson:"heap_size"`
}

// DetailedTCMallocStats stores detailed TCMalloc memory allocator metrics
type DetailedTCMallocStats struct {
	PageheapFreeBytes            int64 `bson:"pageheap_free_bytes"`
	PageheapUnmappedBytes        int64 `bson:"pageheap_unmapped_bytes"`
	MaxTotalThreadCacheBytes     int64 `bson:"max_total_thread_cache_bytes"`
	CurrentTotalThreadCacheBytes int64 `bson:"current_total_thread_cache_bytes"`
	TotalFreeBytes               int64 `bson:"total_free_bytes"`
	CentralCacheFreeBytes        int64 `bson:"central_cache_free_bytes"`
	TransferCacheFreeBytes       int64 `bson:"transfer_cache_free_bytes"`
	ThreadCacheFreeBytes         int64 `bson:"thread_cache_free_bytes"`
	PageheapComittedBytes        int64 `bson:"pageheap_committed_bytes"`
	PageheapScavengeCount        int64 `bson:"pageheap_scavenge_count"`
	PageheapCommitCount          int64 `bson:"pageheap_commit_count"`
	PageheapTotalCommitBytes     int64 `bson:"pageheap_total_commit_bytes"`
	PageheapDecommitCount        int64 `bson:"pageheap_decommit_count"`
	PageheapTotalDecommitBytes   int64 `bson:"pageheap_total_decommit_bytes"`
	PageheapReserveCount         int64 `bson:"pageheap_reserve_count"`
	PageheapTotalReserveBytes    int64 `bson:"pageheap_total_reserve_bytes"`
	SpinLockTotalDelayNanos      int64 `bson:"spinlock_total_delay_ns"`
}

// StorageStats stores information related to record allocations
type StorageStats struct {
	FreelistSearchBucketExhausted int64 `bson:"freelist.search.bucketExhausted"`
	FreelistSearchRequests        int64 `bson:"freelist.search.requests"`
	FreelistSearchScanned         int64 `bson:"freelist.search.scanned"`
}

// StatHeader describes a single column for mongostat's terminal output,
// its formatting, and in which modes it should be displayed.
type StatHeader struct {
	// The text to appear in the column's header cell
	HeaderText string

	// Bitmask containing flags to determine if this header is active or not
	ActivateFlags int
}

// StatHeaders are the complete set of data metrics supported by mongostat.
var StatHeaders = []StatHeader{
	{"", Always}, // placeholder for hostname column (blank header text)
	{"insert", Always},
	{"query", Always},
	{"update", Always},
	{"delete", Always},
	{"getmore", Always},
	{"command", Always},
	{"% dirty", WTOnly},
	{"% used", WTOnly},
	{"flushes", Always},
	{"mapped", MMAPOnly},
	{"vsize", Always},
	{"res", Always},
	{"non-mapped", MMAPOnly | AllOnly},
	{"faults", MMAPOnly},
	{"lr|lw %", MMAPOnly | AllOnly},
	{"lrt|lwt", MMAPOnly | AllOnly},
	{"    locked db", Locks},
	{"qr|qw", Always},
	{"ar|aw", Always},
	{"netIn", Always},
	{"netOut", Always},
	{"conn", Always},
	{"set", Repl},
	{"repl", Repl},
	{"time", Always},
}

// NamespacedLocks stores information on the LockStatus of namespaces.
type NamespacedLocks map[string]LockStatus

// LockUsage stores information related to a namespace's lock usage.
type LockUsage struct {
	Namespace string
	Reads     int64
	Writes    int64
}

type lockUsages []LockUsage

func percentageInt64(value, outOf int64) float64 {
	if value == 0 || outOf == 0 {
		return 0
	}
	return 100 * (float64(value) / float64(outOf))
}

func averageInt64(value, outOf int64) int64 {
	if value == 0 || outOf == 0 {
		return 0
	}
	return value / outOf
}

func (slice lockUsages) Len() int {
	return len(slice)
}

func (slice lockUsages) Less(i, j int) bool {
	return slice[i].Reads+slice[i].Writes < slice[j].Reads+slice[j].Writes
}

func (slice lockUsages) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

// CollectionLockStatus stores a collection's lock statistics.
type CollectionLockStatus struct {
	ReadAcquireWaitsPercentage  float64
	WriteAcquireWaitsPercentage float64
	ReadAcquireTimeMicros       int64
	WriteAcquireTimeMicros      int64
}

// LockStatus stores a database's lock statistics.
type LockStatus struct {
	DBName     string
	Percentage float64
	Global     bool
}

// StatLine is a wrapper for all metrics reported by mongostat for monitored hosts.
type StatLine struct {
	Key string
	// What storage engine is being used for the node with this stat line
	StorageEngine string

	Error    error
	IsMongos bool
	Host     string
	Version  string

	UptimeNanos int64

	// The time at which this StatLine was generated.
	Time time.Time

	// The last time at which this StatLine was printed to output.
	LastPrinted time.Time

	// Opcounter fields
	Insert, InsertCnt   int64
	Query, QueryCnt     int64
	Update, UpdateCnt   int64
	Delete, DeleteCnt   int64
	GetMore, GetMoreCnt int64
	Command, CommandCnt int64

	// Asserts fields
	Regular   int64
	Warning   int64
	Msg       int64
	User      int64
	Rollovers int64

	// OpLatency fields
	WriteOpsCnt    int64
	WriteLatency   int64
	ReadOpsCnt     int64
	ReadLatency    int64
	CommandOpsCnt  int64
	CommandLatency int64

	// TTL fields
	Passes, PassesCnt                     int64
	DeletedDocuments, DeletedDocumentsCnt int64

	// Cursor fields
	TimedOutC, TimedOutCCnt   int64
	NoTimeoutC, NoTimeoutCCnt int64
	PinnedC, PinnedCCnt       int64
	TotalC, TotalCCnt         int64

	// Document fields
	DeletedD, InsertedD, ReturnedD, UpdatedD int64

	//Commands fields
	AggregateCommandTotal, AggregateCommandFailed         int64
	CountCommandTotal, CountCommandFailed                 int64
	DeleteCommandTotal, DeleteCommandFailed               int64
	DistinctCommandTotal, DistinctCommandFailed           int64
	FindCommandTotal, FindCommandFailed                   int64
	FindAndModifyCommandTotal, FindAndModifyCommandFailed int64
	GetMoreCommandTotal, GetMoreCommandFailed             int64
	InsertCommandTotal, InsertCommandFailed               int64
	UpdateCommandTotal, UpdateCommandFailed               int64

	// Operation fields
	ScanAndOrderOp, WriteConflictsOp int64

	// Query Executor fields
	TotalKeysScanned, TotalObjectsScanned int64

	// Connection fields
	CurrentC, AvailableC, TotalCreatedC int64

	// Collection locks (3.0 mmap only)
	CollectionLocks *CollectionLockStatus

	// Cache utilization (wiredtiger only)
	CacheDirtyPercent float64
	CacheUsedPercent  float64

	// Cache utilization extended (wiredtiger only)
	TrackedDirtyBytes         int64
	CurrentCachedBytes        int64
	MaxBytesConfigured        int64
	AppThreadsPageReadCount   int64
	AppThreadsPageReadTime    int64
	AppThreadsPageWriteCount  int64
	BytesWrittenFrom          int64
	BytesReadInto             int64
	PagesEvictedByAppThread   int64
	PagesQueuedForEviction    int64
	PagesReadIntoCache        int64
	PagesWrittenFromCache     int64
	PagesRequestedFromCache   int64
	ServerEvictingPages       int64
	WorkerThreadEvictingPages int64
	InternalPagesEvicted      int64
	ModifiedPagesEvicted      int64
	UnmodifiedPagesEvicted    int64

	// Connection statistics (wiredtiger only)
	FilesCurrentlyOpen int64

	// Data handles statistics (wiredtiger only)
	DataHandlesCurrentlyActive int64

	// Replicated Opcounter fields
	InsertR, InsertRCnt                      int64
	QueryR, QueryRCnt                        int64
	UpdateR, UpdateRCnt                      int64
	DeleteR, DeleteRCnt                      int64
	GetMoreR, GetMoreRCnt                    int64
	CommandR, CommandRCnt                    int64
	ReplLag                                  int64
	OplogStats                               *OplogStats
	Flushes, FlushesCnt                      int64
	FlushesTotalTime                         int64
	Mapped, Virtual, Resident, NonMapped     int64
	Faults, FaultsCnt                        int64
	HighestLocked                            *LockStatus
	QueuedReaders, QueuedWriters             int64
	ActiveReaders, ActiveWriters             int64
	AvailableReaders, AvailableWriters       int64
	TotalTicketsReaders, TotalTicketsWriters int64
	NetIn, NetInCnt                          int64
	NetOut, NetOutCnt                        int64
	NumConnections                           int64
	ReplSetName                              string
	NodeType                                 string
	NodeState                                string
	NodeStateInt                             int64

	// Replicated Metrics fields
	ReplNetworkBytes                    int64
	ReplNetworkGetmoresNum              int64
	ReplNetworkGetmoresTotalMillis      int64
	ReplNetworkOps                      int64
	ReplBufferCount                     int64
	ReplBufferSizeBytes                 int64
	ReplApplyBatchesNum                 int64
	ReplApplyBatchesTotalMillis         int64
	ReplApplyOps                        int64
	ReplExecutorPoolInProgressCount     int64
	ReplExecutorQueuesNetworkInProgress int64
	ReplExecutorQueuesSleepers          int64
	ReplExecutorUnsignaledEvents        int64

	// Cluster fields
	JumboChunksCount int64

	// DB stats field
	DbStatsLines []DbStatLine

	// Col Stats field
	ColStatsLines []ColStatLine

	// Shard stats
	TotalInUse, TotalAvailable, TotalCreated, TotalRefreshing int64

	// Shard Hosts stats field
	ShardHostStatsLines map[string]ShardHostStatLine

	TopStatLines []TopStatLine

	// TCMalloc stats field
	TCMallocCurrentAllocatedBytes        int64
	TCMallocHeapSize                     int64
	TCMallocCentralCacheFreeBytes        int64
	TCMallocCurrentTotalThreadCacheBytes int64
	TCMallocMaxTotalThreadCacheBytes     int64
	TCMallocTotalFreeBytes               int64
	TCMallocTransferCacheFreeBytes       int64
	TCMallocThreadCacheFreeBytes         int64
	TCMallocSpinLockTotalDelayNanos      int64
	TCMallocPageheapFreeBytes            int64
	TCMallocPageheapUnmappedBytes        int64
	TCMallocPageheapComittedBytes        int64
	TCMallocPageheapScavengeCount        int64
	TCMallocPageheapCommitCount          int64
	TCMallocPageheapTotalCommitBytes     int64
	TCMallocPageheapDecommitCount        int64
	TCMallocPageheapTotalDecommitBytes   int64
	TCMallocPageheapReserveCount         int64
	TCMallocPageheapTotalReserveBytes    int64

	// Storage stats field
	StorageFreelistSearchBucketExhausted int64
	StorageFreelistSearchRequests        int64
	StorageFreelistSearchScanned         int64
}

type DbStatLine struct {
	Name        string
	Collections int64
	Objects     int64
	AvgObjSize  float64
	DataSize    int64
	StorageSize int64
	NumExtents  int64
	Indexes     int64
	IndexSize   int64
	Ok          int64
	FsUsedSize  int64
	FsTotalSize int64
}
type ColStatLine struct {
	Name           string
	DbName         string
	Count          int64
	Size           int64
	AvgObjSize     float64
	StorageSize    int64
	TotalIndexSize int64
	Ok             int64
}

type ShardHostStatLine struct {
	InUse      int64
	Available  int64
	Created    int64
	Refreshing int64
}

type TopStatLine struct {
	CollectionName                string
	TotalTime, TotalCount         int64
	ReadLockTime, ReadLockCount   int64
	WriteLockTime, WriteLockCount int64
	QueriesTime, QueriesCount     int64
	GetMoreTime, GetMoreCount     int64
	InsertTime, InsertCount       int64
	UpdateTime, UpdateCount       int64
	RemoveTime, RemoveCount       int64
	CommandsTime, CommandsCount   int64
}

func parseLocks(stat ServerStatus) map[string]LockUsage {
	returnVal := map[string]LockUsage{}
	for namespace, lockInfo := range stat.Locks {
		returnVal[namespace] = LockUsage{
			namespace,
			lockInfo.TimeLockedMicros.Read + lockInfo.TimeLockedMicros.ReadLower,
			lockInfo.TimeLockedMicros.Write + lockInfo.TimeLockedMicros.WriteLower,
		}
	}
	return returnVal
}

func computeLockDiffs(prevLocks, curLocks map[string]LockUsage) []LockUsage {
	lockUsages := lockUsages(make([]LockUsage, 0, len(curLocks)))
	for namespace, curUsage := range curLocks {
		prevUsage, hasKey := prevLocks[namespace]
		if !hasKey {
			// This namespace didn't appear in the previous batch of lock info,
			// so we can't compute a diff for it - skip it.
			continue
		}
		// Calculate diff of lock usage for this namespace and add to the list
		lockUsages = append(lockUsages,
			LockUsage{
				namespace,
				curUsage.Reads - prevUsage.Reads,
				curUsage.Writes - prevUsage.Writes,
			})
	}
	// Sort the array in order of least to most locked
	sort.Sort(lockUsages)
	return lockUsages
}

func diff(newVal, oldVal, sampleTime int64) (avg int64, newValue int64) {
	d := newVal - oldVal
	if d < 0 {
		d = newVal
	}
	return d / sampleTime, newVal
}

// NewStatLine constructs a StatLine object from two MongoStatus objects.
func NewStatLine(oldMongo, newMongo MongoStatus, key string, all bool, sampleSecs int64) *StatLine {
	oldStat := *oldMongo.ServerStatus
	newStat := *newMongo.ServerStatus

	returnVal := &StatLine{
		Key:       key,
		Host:      newStat.Host,
		Version:   newStat.Version,
		Mapped:    -1,
		Virtual:   -1,
		Resident:  -1,
		NonMapped: -1,
		Faults:    -1,
	}

	returnVal.UptimeNanos = 1000 * 1000 * newStat.UptimeMillis

	// set connection info
	returnVal.CurrentC = newStat.Connections.Current
	returnVal.AvailableC = newStat.Connections.Available
	returnVal.TotalCreatedC = newStat.Connections.TotalCreated

	// set the storage engine appropriately
	if newStat.StorageEngine != nil && newStat.StorageEngine.Name != "" {
		returnVal.StorageEngine = newStat.StorageEngine.Name
	} else {
		returnVal.StorageEngine = "mmapv1"
	}

	if newStat.Opcounters != nil && oldStat.Opcounters != nil {
		returnVal.Insert, returnVal.InsertCnt = diff(newStat.Opcounters.Insert, oldStat.Opcounters.Insert, sampleSecs)
		returnVal.Query, returnVal.QueryCnt = diff(newStat.Opcounters.Query, oldStat.Opcounters.Query, sampleSecs)
		returnVal.Update, returnVal.UpdateCnt = diff(newStat.Opcounters.Update, oldStat.Opcounters.Update, sampleSecs)
		returnVal.Delete, returnVal.DeleteCnt = diff(newStat.Opcounters.Delete, oldStat.Opcounters.Delete, sampleSecs)
		returnVal.GetMore, returnVal.GetMoreCnt = diff(newStat.Opcounters.GetMore, oldStat.Opcounters.GetMore, sampleSecs)
		returnVal.Command, returnVal.CommandCnt = diff(newStat.Opcounters.Command, oldStat.Opcounters.Command, sampleSecs)
	}

	if newStat.OpLatencies != nil {
		if newStat.OpLatencies.Reads != nil {
			returnVal.ReadOpsCnt = newStat.OpLatencies.Reads.Ops
			returnVal.ReadLatency = newStat.OpLatencies.Reads.Latency
		}
		if newStat.OpLatencies.Writes != nil {
			returnVal.WriteOpsCnt = newStat.OpLatencies.Writes.Ops
			returnVal.WriteLatency = newStat.OpLatencies.Writes.Latency
		}
		if newStat.OpLatencies.Commands != nil {
			returnVal.CommandOpsCnt = newStat.OpLatencies.Commands.Ops
			returnVal.CommandLatency = newStat.OpLatencies.Commands.Latency
		}
	}

	if newStat.Asserts != nil {
		returnVal.Regular = newStat.Asserts.Regular
		returnVal.Warning = newStat.Asserts.Warning
		returnVal.Msg = newStat.Asserts.Msg
		returnVal.User = newStat.Asserts.User
		returnVal.Rollovers = newStat.Asserts.Rollovers
	}

	if newStat.TCMallocStats != nil {
		if newStat.TCMallocStats.Generic != nil {
			returnVal.TCMallocCurrentAllocatedBytes = newStat.TCMallocStats.Generic.CurrentAllocatedBytes
			returnVal.TCMallocHeapSize = newStat.TCMallocStats.Generic.HeapSize
		}
		if newStat.TCMallocStats.TCMalloc != nil {
			returnVal.TCMallocCentralCacheFreeBytes = newStat.TCMallocStats.TCMalloc.CentralCacheFreeBytes
			returnVal.TCMallocCurrentTotalThreadCacheBytes = newStat.TCMallocStats.TCMalloc.CurrentTotalThreadCacheBytes
			returnVal.TCMallocMaxTotalThreadCacheBytes = newStat.TCMallocStats.TCMalloc.MaxTotalThreadCacheBytes
			returnVal.TCMallocTransferCacheFreeBytes = newStat.TCMallocStats.TCMalloc.TransferCacheFreeBytes
			returnVal.TCMallocThreadCacheFreeBytes = newStat.TCMallocStats.TCMalloc.ThreadCacheFreeBytes
			returnVal.TCMallocTotalFreeBytes = newStat.TCMallocStats.TCMalloc.TotalFreeBytes
			returnVal.TCMallocSpinLockTotalDelayNanos = newStat.TCMallocStats.TCMalloc.SpinLockTotalDelayNanos

			returnVal.TCMallocPageheapFreeBytes = newStat.TCMallocStats.TCMalloc.PageheapFreeBytes
			returnVal.TCMallocPageheapUnmappedBytes = newStat.TCMallocStats.TCMalloc.PageheapUnmappedBytes
			returnVal.TCMallocPageheapComittedBytes = newStat.TCMallocStats.TCMalloc.PageheapComittedBytes
			returnVal.TCMallocPageheapScavengeCount = newStat.TCMallocStats.TCMalloc.PageheapScavengeCount
			returnVal.TCMallocPageheapCommitCount = newStat.TCMallocStats.TCMalloc.PageheapCommitCount
			returnVal.TCMallocPageheapTotalCommitBytes = newStat.TCMallocStats.TCMalloc.PageheapTotalCommitBytes
			returnVal.TCMallocPageheapDecommitCount = newStat.TCMallocStats.TCMalloc.PageheapDecommitCount
			returnVal.TCMallocPageheapTotalDecommitBytes = newStat.TCMallocStats.TCMalloc.PageheapTotalDecommitBytes
			returnVal.TCMallocPageheapReserveCount = newStat.TCMallocStats.TCMalloc.PageheapReserveCount
			returnVal.TCMallocPageheapTotalReserveBytes = newStat.TCMallocStats.TCMalloc.PageheapTotalReserveBytes
		}
	}

	if newStat.Metrics != nil && oldStat.Metrics != nil {
		if newStat.Metrics.TTL != nil && oldStat.Metrics.TTL != nil {
			returnVal.Passes, returnVal.PassesCnt = diff(newStat.Metrics.TTL.Passes, oldStat.Metrics.TTL.Passes, sampleSecs)
			returnVal.DeletedDocuments, returnVal.DeletedDocumentsCnt = diff(newStat.Metrics.TTL.DeletedDocuments, oldStat.Metrics.TTL.DeletedDocuments, sampleSecs)
		}
		if newStat.Metrics.Cursor != nil && oldStat.Metrics.Cursor != nil {
			returnVal.TimedOutC, returnVal.TimedOutCCnt = diff(newStat.Metrics.Cursor.TimedOut, oldStat.Metrics.Cursor.TimedOut, sampleSecs)
			if newStat.Metrics.Cursor.Open != nil && oldStat.Metrics.Cursor.Open != nil {
				returnVal.NoTimeoutC, returnVal.NoTimeoutCCnt = diff(newStat.Metrics.Cursor.Open.NoTimeout, oldStat.Metrics.Cursor.Open.NoTimeout, sampleSecs)
				returnVal.PinnedC, returnVal.PinnedCCnt = diff(newStat.Metrics.Cursor.Open.Pinned, oldStat.Metrics.Cursor.Open.Pinned, sampleSecs)
				returnVal.TotalC, returnVal.TotalCCnt = diff(newStat.Metrics.Cursor.Open.Total, oldStat.Metrics.Cursor.Open.Total, sampleSecs)
			}
		}
		if newStat.Metrics.Document != nil {
			returnVal.DeletedD = newStat.Metrics.Document.Deleted
			returnVal.InsertedD = newStat.Metrics.Document.Inserted
			returnVal.ReturnedD = newStat.Metrics.Document.Returned
			returnVal.UpdatedD = newStat.Metrics.Document.Updated
		}

		if newStat.Metrics.Commands != nil {
			if newStat.Metrics.Commands.Aggregate != nil {
				returnVal.AggregateCommandTotal = newStat.Metrics.Commands.Aggregate.Total
				returnVal.AggregateCommandFailed = newStat.Metrics.Commands.Aggregate.Failed
			}
			if newStat.Metrics.Commands.Count != nil {
				returnVal.CountCommandTotal = newStat.Metrics.Commands.Count.Total
				returnVal.CountCommandFailed = newStat.Metrics.Commands.Count.Failed
			}
			if newStat.Metrics.Commands.Delete != nil {
				returnVal.DeleteCommandTotal = newStat.Metrics.Commands.Delete.Total
				returnVal.DeleteCommandFailed = newStat.Metrics.Commands.Delete.Failed
			}
			if newStat.Metrics.Commands.Distinct != nil {
				returnVal.DistinctCommandTotal = newStat.Metrics.Commands.Distinct.Total
				returnVal.DistinctCommandFailed = newStat.Metrics.Commands.Distinct.Failed
			}
			if newStat.Metrics.Commands.Find != nil {
				returnVal.FindCommandTotal = newStat.Metrics.Commands.Find.Total
				returnVal.FindCommandFailed = newStat.Metrics.Commands.Find.Failed
			}
			if newStat.Metrics.Commands.FindAndModify != nil {
				returnVal.FindAndModifyCommandTotal = newStat.Metrics.Commands.FindAndModify.Total
				returnVal.FindAndModifyCommandFailed = newStat.Metrics.Commands.FindAndModify.Failed
			}
			if newStat.Metrics.Commands.GetMore != nil {
				returnVal.GetMoreCommandTotal = newStat.Metrics.Commands.GetMore.Total
				returnVal.GetMoreCommandFailed = newStat.Metrics.Commands.GetMore.Failed
			}
			if newStat.Metrics.Commands.Insert != nil {
				returnVal.InsertCommandTotal = newStat.Metrics.Commands.Insert.Total
				returnVal.InsertCommandFailed = newStat.Metrics.Commands.Insert.Failed
			}
			if newStat.Metrics.Commands.Update != nil {
				returnVal.UpdateCommandTotal = newStat.Metrics.Commands.Update.Total
				returnVal.UpdateCommandFailed = newStat.Metrics.Commands.Update.Failed
			}
		}

		if newStat.Metrics.Operation != nil {
			returnVal.ScanAndOrderOp = newStat.Metrics.Operation.ScanAndOrder
			returnVal.WriteConflictsOp = newStat.Metrics.Operation.WriteConflicts
		}

		if newStat.Metrics.QueryExecutor != nil {
			returnVal.TotalKeysScanned = newStat.Metrics.QueryExecutor.Scanned
			returnVal.TotalObjectsScanned = newStat.Metrics.QueryExecutor.ScannedObjects
		}

		if newStat.Metrics.Repl != nil {
			if newStat.Metrics.Repl.Apply != nil {
				returnVal.ReplApplyBatchesNum = newStat.Metrics.Repl.Apply.Batches.Num
				returnVal.ReplApplyBatchesTotalMillis = newStat.Metrics.Repl.Apply.Batches.TotalMillis
				returnVal.ReplApplyOps = newStat.Metrics.Repl.Apply.Ops
			}
			if newStat.Metrics.Repl.Buffer != nil {
				returnVal.ReplBufferCount = newStat.Metrics.Repl.Buffer.Count
				returnVal.ReplBufferSizeBytes = newStat.Metrics.Repl.Buffer.SizeBytes
			}
			if newStat.Metrics.Repl.Executor != nil {
				returnVal.ReplExecutorPoolInProgressCount = newStat.Metrics.Repl.Executor.Pool["inProgressCount"]
				returnVal.ReplExecutorQueuesNetworkInProgress = newStat.Metrics.Repl.Executor.Queues["networkInProgress"]
				returnVal.ReplExecutorQueuesSleepers = newStat.Metrics.Repl.Executor.Queues["sleepers"]
				returnVal.ReplExecutorUnsignaledEvents = newStat.Metrics.Repl.Executor.UnsignaledEvents
			}
			if newStat.Metrics.Repl.Network != nil {
				returnVal.ReplNetworkBytes = newStat.Metrics.Repl.Network.Bytes
				if newStat.Metrics.Repl.Network.GetMores != nil {
					returnVal.ReplNetworkGetmoresNum = newStat.Metrics.Repl.Network.GetMores.Num
					returnVal.ReplNetworkGetmoresTotalMillis = newStat.Metrics.Repl.Network.GetMores.TotalMillis
				}
				returnVal.ReplNetworkOps = newStat.Metrics.Repl.Network.Ops
			}
		}

		if newStat.Metrics.Storage != nil {
			returnVal.StorageFreelistSearchBucketExhausted = newStat.Metrics.Storage.FreelistSearchBucketExhausted
			returnVal.StorageFreelistSearchRequests = newStat.Metrics.Storage.FreelistSearchRequests
			returnVal.StorageFreelistSearchScanned = newStat.Metrics.Storage.FreelistSearchScanned
		}
	}

	if newStat.OpcountersRepl != nil && oldStat.OpcountersRepl != nil {
		returnVal.InsertR, returnVal.InsertRCnt = diff(newStat.OpcountersRepl.Insert, oldStat.OpcountersRepl.Insert, sampleSecs)
		returnVal.QueryR, returnVal.QueryRCnt = diff(newStat.OpcountersRepl.Query, oldStat.OpcountersRepl.Query, sampleSecs)
		returnVal.UpdateR, returnVal.UpdateRCnt = diff(newStat.OpcountersRepl.Update, oldStat.OpcountersRepl.Update, sampleSecs)
		returnVal.DeleteR, returnVal.DeleteRCnt = diff(newStat.OpcountersRepl.Delete, oldStat.OpcountersRepl.Delete, sampleSecs)
		returnVal.GetMoreR, returnVal.GetMoreRCnt = diff(newStat.OpcountersRepl.GetMore, oldStat.OpcountersRepl.GetMore, sampleSecs)
		returnVal.CommandR, returnVal.CommandRCnt = diff(newStat.OpcountersRepl.Command, oldStat.OpcountersRepl.Command, sampleSecs)
	}

	returnVal.CacheDirtyPercent = -1
	returnVal.CacheUsedPercent = -1
	if newStat.WiredTiger != nil {
		returnVal.CacheDirtyPercent = float64(newStat.WiredTiger.Cache.TrackedDirtyBytes) / float64(newStat.WiredTiger.Cache.MaxBytesConfigured)
		returnVal.CacheUsedPercent = float64(newStat.WiredTiger.Cache.CurrentCachedBytes) / float64(newStat.WiredTiger.Cache.MaxBytesConfigured)

		returnVal.TrackedDirtyBytes = newStat.WiredTiger.Cache.TrackedDirtyBytes
		returnVal.CurrentCachedBytes = newStat.WiredTiger.Cache.CurrentCachedBytes
		returnVal.MaxBytesConfigured = newStat.WiredTiger.Cache.MaxBytesConfigured
		returnVal.AppThreadsPageReadCount = newStat.WiredTiger.Cache.AppThreadsPageReadCount
		returnVal.AppThreadsPageReadTime = newStat.WiredTiger.Cache.AppThreadsPageReadTime
		returnVal.AppThreadsPageWriteCount = newStat.WiredTiger.Cache.AppThreadsPageWriteCount
		returnVal.BytesWrittenFrom = newStat.WiredTiger.Cache.BytesWrittenFrom
		returnVal.BytesReadInto = newStat.WiredTiger.Cache.BytesReadInto
		returnVal.PagesEvictedByAppThread = newStat.WiredTiger.Cache.PagesEvictedByAppThread
		returnVal.PagesQueuedForEviction = newStat.WiredTiger.Cache.PagesQueuedForEviction
		returnVal.PagesReadIntoCache = newStat.WiredTiger.Cache.PagesReadIntoCache
		returnVal.PagesWrittenFromCache = newStat.WiredTiger.Cache.PagesWrittenFromCache
		returnVal.PagesRequestedFromCache = newStat.WiredTiger.Cache.PagesRequestedFromCache
		returnVal.ServerEvictingPages = newStat.WiredTiger.Cache.ServerEvictingPages
		returnVal.WorkerThreadEvictingPages = newStat.WiredTiger.Cache.WorkerThreadEvictingPages

		returnVal.InternalPagesEvicted = newStat.WiredTiger.Cache.InternalPagesEvicted
		returnVal.ModifiedPagesEvicted = newStat.WiredTiger.Cache.ModifiedPagesEvicted
		returnVal.UnmodifiedPagesEvicted = newStat.WiredTiger.Cache.UnmodifiedPagesEvicted

		returnVal.FlushesTotalTime = newStat.WiredTiger.Transaction.TransCheckpointsTotalTimeMsecs * int64(time.Millisecond)

		returnVal.FilesCurrentlyOpen = newStat.WiredTiger.Connection.FilesCurrentlyOpen

		returnVal.DataHandlesCurrentlyActive = newStat.WiredTiger.DataHandle.DataHandlesCurrentlyActive
	}
	if newStat.WiredTiger != nil && oldStat.WiredTiger != nil {
		returnVal.Flushes, returnVal.FlushesCnt = diff(newStat.WiredTiger.Transaction.TransCheckpoints, oldStat.WiredTiger.Transaction.TransCheckpoints, sampleSecs)
	} else if newStat.BackgroundFlushing != nil && oldStat.BackgroundFlushing != nil {
		returnVal.Flushes, returnVal.FlushesCnt = diff(newStat.BackgroundFlushing.Flushes, oldStat.BackgroundFlushing.Flushes, sampleSecs)
	}

	returnVal.Time = newMongo.SampleTime
	returnVal.IsMongos =
		newStat.ShardCursorType != nil || strings.HasPrefix(newStat.Process, MongosProcess)

	// BEGIN code modification
	if oldStat.Mem.Supported.(bool) {
		// END code modification
		if !returnVal.IsMongos {
			returnVal.Mapped = newStat.Mem.Mapped
		}
		returnVal.Virtual = newStat.Mem.Virtual
		returnVal.Resident = newStat.Mem.Resident

		if !returnVal.IsMongos && all {
			returnVal.NonMapped = newStat.Mem.Virtual - newStat.Mem.Mapped
		}
	}

	if newStat.Repl != nil {
		returnVal.ReplSetName = newStat.Repl.SetName
		// BEGIN code modification
		if val, ok := newStat.Repl.IsMaster.(bool); ok && val {
			returnVal.NodeType = "PRI"
		} else if val, ok := newStat.Repl.IsWritablePrimary.(bool); ok && val {
			returnVal.NodeType = "PRI"
		} else if val, ok := newStat.Repl.Secondary.(bool); ok && val {
			returnVal.NodeType = "SEC"
		} else if val, ok := newStat.Repl.ArbiterOnly.(bool); ok && val {
			returnVal.NodeType = "ARB"
		} else {
			returnVal.NodeType = "UNK"
		} // END code modification
	} else if returnVal.IsMongos {
		returnVal.NodeType = "RTR"
	}

	if oldStat.ExtraInfo != nil && newStat.ExtraInfo != nil &&
		oldStat.ExtraInfo.PageFaults != nil && newStat.ExtraInfo.PageFaults != nil {
		returnVal.Faults, returnVal.FaultsCnt = diff(*(newStat.ExtraInfo.PageFaults), *(oldStat.ExtraInfo.PageFaults), sampleSecs)
	}
	if !returnVal.IsMongos && oldStat.Locks != nil {
		globalCheck, hasGlobal := oldStat.Locks["Global"]
		if hasGlobal && globalCheck.AcquireCount != nil {
			// This appears to be a 3.0+ server so the data in these fields do *not* refer to
			// actual namespaces and thus we can't compute lock %.
			returnVal.HighestLocked = nil

			// Check if it's a 3.0+ MMAP server so we can still compute collection locks
			collectionCheck, hasCollection := oldStat.Locks["Collection"]
			if hasCollection && collectionCheck.AcquireWaitCount != nil {
				readWaitCountDiff := newStat.Locks["Collection"].AcquireWaitCount.Read - oldStat.Locks["Collection"].AcquireWaitCount.Read
				readTotalCountDiff := newStat.Locks["Collection"].AcquireCount.Read - oldStat.Locks["Collection"].AcquireCount.Read
				writeWaitCountDiff := newStat.Locks["Collection"].AcquireWaitCount.Write - oldStat.Locks["Collection"].AcquireWaitCount.Write
				writeTotalCountDiff := newStat.Locks["Collection"].AcquireCount.Write - oldStat.Locks["Collection"].AcquireCount.Write
				readAcquireTimeDiff := newStat.Locks["Collection"].TimeAcquiringMicros.Read - oldStat.Locks["Collection"].TimeAcquiringMicros.Read
				writeAcquireTimeDiff := newStat.Locks["Collection"].TimeAcquiringMicros.Write - oldStat.Locks["Collection"].TimeAcquiringMicros.Write
				returnVal.CollectionLocks = &CollectionLockStatus{
					ReadAcquireWaitsPercentage:  percentageInt64(readWaitCountDiff, readTotalCountDiff),
					WriteAcquireWaitsPercentage: percentageInt64(writeWaitCountDiff, writeTotalCountDiff),
					ReadAcquireTimeMicros:       averageInt64(readAcquireTimeDiff, readWaitCountDiff),
					WriteAcquireTimeMicros:      averageInt64(writeAcquireTimeDiff, writeWaitCountDiff),
				}
			}
		} else {
			prevLocks := parseLocks(oldStat)
			curLocks := parseLocks(newStat)
			lockdiffs := computeLockDiffs(prevLocks, curLocks)
			if len(lockdiffs) == 0 {
				if newStat.GlobalLock != nil {
					returnVal.HighestLocked = &LockStatus{
						DBName:     "",
						Percentage: percentageInt64(newStat.GlobalLock.LockTime, newStat.GlobalLock.TotalTime),
						Global:     true,
					}
				}
			} else {
				// Get the entry with the highest lock
				highestLocked := lockdiffs[len(lockdiffs)-1]

				timeDiffMillis := newStat.UptimeMillis - oldStat.UptimeMillis
				lockToReport := highestLocked.Writes

				// if the highest locked namespace is not '.'
				if highestLocked.Namespace != "." {
					for _, namespaceLockInfo := range lockdiffs {
						if namespaceLockInfo.Namespace == "." {
							lockToReport += namespaceLockInfo.Writes
						}
					}
				}

				// lock data is in microseconds and uptime is in milliseconds - so
				// divide by 1000 so that they units match
				lockToReport /= 1000

				returnVal.HighestLocked = &LockStatus{
					DBName:     highestLocked.Namespace,
					Percentage: percentageInt64(lockToReport, timeDiffMillis),
					Global:     false,
				}
			}
		}
	} else {
		returnVal.HighestLocked = nil
	}

	if newStat.GlobalLock != nil {
		hasWT := newStat.WiredTiger != nil && oldStat.WiredTiger != nil
		//If we have wiredtiger stats, use those instead
		if newStat.GlobalLock.CurrentQueue != nil {
			if hasWT {
				returnVal.QueuedReaders = newStat.GlobalLock.CurrentQueue.Readers + newStat.GlobalLock.ActiveClients.Readers - newStat.WiredTiger.Concurrent.Read.Out
				returnVal.QueuedWriters = newStat.GlobalLock.CurrentQueue.Writers + newStat.GlobalLock.ActiveClients.Writers - newStat.WiredTiger.Concurrent.Write.Out
				if returnVal.QueuedReaders < 0 {
					returnVal.QueuedReaders = 0
				}
				if returnVal.QueuedWriters < 0 {
					returnVal.QueuedWriters = 0
				}
			} else {
				returnVal.QueuedReaders = newStat.GlobalLock.CurrentQueue.Readers
				returnVal.QueuedWriters = newStat.GlobalLock.CurrentQueue.Writers
			}
		}

		if hasWT {
			returnVal.ActiveReaders = newStat.WiredTiger.Concurrent.Read.Out
			returnVal.ActiveWriters = newStat.WiredTiger.Concurrent.Write.Out
			returnVal.AvailableReaders = newStat.WiredTiger.Concurrent.Read.Available
			returnVal.AvailableWriters = newStat.WiredTiger.Concurrent.Write.Available
			returnVal.TotalTicketsReaders = newStat.WiredTiger.Concurrent.Read.TotalTickets
			returnVal.TotalTicketsWriters = newStat.WiredTiger.Concurrent.Write.TotalTickets
		} else if newStat.GlobalLock.ActiveClients != nil {
			returnVal.ActiveReaders = newStat.GlobalLock.ActiveClients.Readers
			returnVal.ActiveWriters = newStat.GlobalLock.ActiveClients.Writers
		}
	}

	if oldStat.Network != nil && newStat.Network != nil {
		returnVal.NetIn, returnVal.NetInCnt = diff(newStat.Network.BytesIn, oldStat.Network.BytesIn, sampleSecs)
		returnVal.NetOut, returnVal.NetOutCnt = diff(newStat.Network.BytesOut, oldStat.Network.BytesOut, sampleSecs)
	}

	if newStat.Connections != nil {
		returnVal.NumConnections = newStat.Connections.Current
	}

	if newMongo.ReplSetStatus != nil {
		newReplStat := *newMongo.ReplSetStatus

		if newReplStat.Members != nil {
			myName := newStat.Repl.Me
			// Find the master and myself
			master := ReplSetMember{}
			me := ReplSetMember{}
			for _, member := range newReplStat.Members {
				if member.Name == myName {
					// Store my state string
					returnVal.NodeState = member.StateStr
					// Store my state integer
					returnVal.NodeStateInt = member.State

					if member.State == 1 {
						// I'm the master
						returnVal.ReplLag = 0
						break
					}

					// I'm secondary
					me = member
				} else if member.State == 1 {
					// Master found
					master = member
				}
			}

			if me.State == 2 {
				// OptimeDate.Unix() type is int64
				lag := master.OptimeDate.Unix() - me.OptimeDate.Unix()
				if lag < 0 {
					returnVal.ReplLag = 0
				} else {
					returnVal.ReplLag = lag
				}
			}
		}
	}

	if newMongo.ClusterStatus != nil {
		newClusterStat := *newMongo.ClusterStatus
		returnVal.JumboChunksCount = newClusterStat.JumboChunksCount
	}

	if newMongo.OplogStats != nil {
		returnVal.OplogStats = newMongo.OplogStats
	}

	if newMongo.DbStats != nil {
		newDbStats := *newMongo.DbStats
		for _, db := range newDbStats.Dbs {
			dbStatsData := db.DbStatsData
			// mongos doesn't have the db key, so setting the db name
			if dbStatsData.Db == "" {
				dbStatsData.Db = db.Name
			}
			dbStatLine := &DbStatLine{
				Name:        dbStatsData.Db,
				Collections: dbStatsData.Collections,
				Objects:     dbStatsData.Objects,
				AvgObjSize:  dbStatsData.AvgObjSize,
				DataSize:    dbStatsData.DataSize,
				StorageSize: dbStatsData.StorageSize,
				NumExtents:  dbStatsData.NumExtents,
				Indexes:     dbStatsData.Indexes,
				IndexSize:   dbStatsData.IndexSize,
				Ok:          dbStatsData.Ok,
				FsTotalSize: dbStatsData.FsTotalSize,
				FsUsedSize:  dbStatsData.FsUsedSize,
			}
			returnVal.DbStatsLines = append(returnVal.DbStatsLines, *dbStatLine)
		}
	}

	if newMongo.ColStats != nil {
		for _, col := range newMongo.ColStats.Collections {
			colStatsData := col.ColStatsData
			// mongos doesn't have the db key, so setting the db name
			if colStatsData.Collection == "" {
				colStatsData.Collection = col.Name
			}
			colStatLine := &ColStatLine{
				Name:           colStatsData.Collection,
				DbName:         col.DbName,
				Count:          colStatsData.Count,
				Size:           colStatsData.Size,
				AvgObjSize:     colStatsData.AvgObjSize,
				StorageSize:    colStatsData.StorageSize,
				TotalIndexSize: colStatsData.TotalIndexSize,
				Ok:             colStatsData.Ok,
			}
			returnVal.ColStatsLines = append(returnVal.ColStatsLines, *colStatLine)
		}
	}

	// Set shard stats
	if newMongo.ShardStats != nil {
		newShardStats := *newMongo.ShardStats
		returnVal.TotalInUse = newShardStats.TotalInUse
		returnVal.TotalAvailable = newShardStats.TotalAvailable
		returnVal.TotalCreated = newShardStats.TotalCreated
		returnVal.TotalRefreshing = newShardStats.TotalRefreshing
		returnVal.ShardHostStatsLines = map[string]ShardHostStatLine{}
		for host, stats := range newShardStats.Hosts {
			shardStatLine := &ShardHostStatLine{
				InUse:      stats.InUse,
				Available:  stats.Available,
				Created:    stats.Created,
				Refreshing: stats.Refreshing,
			}

			returnVal.ShardHostStatsLines[host] = *shardStatLine
		}
	}

	if newMongo.TopStats != nil {
		for collection, data := range newMongo.TopStats.Totals {
			topStatDataLine := &TopStatLine{
				CollectionName: collection,
				TotalTime:      data.Total.Time,
				TotalCount:     data.Total.Count,
				ReadLockTime:   data.ReadLock.Time,
				ReadLockCount:  data.ReadLock.Count,
				WriteLockTime:  data.WriteLock.Time,
				WriteLockCount: data.WriteLock.Count,
				QueriesTime:    data.Queries.Time,
				QueriesCount:   data.Queries.Count,
				GetMoreTime:    data.GetMore.Time,
				GetMoreCount:   data.GetMore.Count,
				InsertTime:     data.Insert.Time,
				InsertCount:    data.Insert.Count,
				UpdateTime:     data.Update.Time,
				UpdateCount:    data.Update.Count,
				RemoveTime:     data.Remove.Time,
				RemoveCount:    data.Remove.Count,
				CommandsTime:   data.Commands.Time,
				CommandsCount:  data.Commands.Count,
			}
			returnVal.TopStatLines = append(returnVal.TopStatLines, *topStatDataLine)
		}
	}

	return returnVal
}
