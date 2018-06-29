/***
The code contained here came from https://github.com/mongodb/mongo-tools/blob/master/mongostat/stat_types.go
and contains modifications so that no other dependency from that project is needed. Other modifications included
removing uneccessary code specific to formatting the output and determine the current state of the database. It
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
	ShardStats    *ShardStats
	OplogStats    *OplogStats
}

type ServerStatus struct {
	Host               string                 `bson:"host"`
	Version            string                 `bson:"version"`
	Process            string                 `bson:"process"`
	Pid                int64                  `bson:"pid"`
	Uptime             int64                  `bson:"uptime"`
	UptimeMillis       int64                  `bson:"uptimeMillis"`
	UptimeEstimate     int64                  `bson:"uptimeEstimate"`
	LocalTime          time.Time              `bson:"localTime"`
	Asserts            map[string]int64       `bson:"asserts"`
	BackgroundFlushing *FlushStats            `bson:"backgroundFlushing"`
	ExtraInfo          *ExtraInfo             `bson:"extra_info"`
	Connections        *ConnectionStats       `bson:"connections"`
	Dur                *DurStats              `bson:"dur"`
	GlobalLock         *GlobalLockStats       `bson:"globalLock"`
	Locks              map[string]LockStats   `bson:"locks,omitempty"`
	Network            *NetworkStats          `bson:"network"`
	Opcounters         *OpcountStats          `bson:"opcounters"`
	OpcountersRepl     *OpcountStats          `bson:"opcountersRepl"`
	RecordStats        *DBRecordStats         `bson:"recordStats"`
	Mem                *MemStats              `bson:"mem"`
	Repl               *ReplStatus            `bson:"repl"`
	ShardCursorType    map[string]interface{} `bson:"shardCursorType"`
	StorageEngine      map[string]string      `bson:"storageEngine"`
	WiredTiger         *WiredTiger            `bson:"wiredTiger"`
	Metrics            *MetricsStats          `bson:"metrics"`
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

type ConcurrentTransactions struct {
	Write ConcurrentTransStats `bson:"write"`
	Read  ConcurrentTransStats `bson:"read"`
}

type ConcurrentTransStats struct {
	Out int64 `bson:"out"`
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
	ServerEvictingPages       int64 `bson:"eviction server evicting pages"`
	WorkerThreadEvictingPages int64 `bson:"eviction worker thread evicting pages"`
}

// TransactionStats stores transaction checkpoints in WiredTiger.
type TransactionStats struct {
	TransCheckpoints int64 `bson:"transaction checkpoints"`
}

// ReplStatus stores data related to replica sets.
type ReplStatus struct {
	SetName      interface{} `bson:"setName"`
	IsMaster     interface{} `bson:"ismaster"`
	Secondary    interface{} `bson:"secondary"`
	IsReplicaSet interface{} `bson:"isreplicaset"`
	ArbiterOnly  interface{} `bson:"arbiterOnly"`
	Hosts        []string    `bson:"hosts"`
	Passives     []string    `bson:"passives"`
	Me           string      `bson:"me"`
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
	TotalCreated int64 `bson:"total_created"`
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

// OpcountStats stores information related to comamnds and basic CRUD operations.
type OpcountStats struct {
	Insert  int64 `bson:"insert"`
	Query   int64 `bson:"query"`
	Update  int64 `bson:"update"`
	Delete  int64 `bson:"delete"`
	GetMore int64 `bson:"getmore"`
	Command int64 `bson:"command"`
}

// MetricsStats stores information related to metrics
type MetricsStats struct {
	TTL      *TTLStats      `bson:"ttl"`
	Cursor   *CursorStats   `bson:"cursor"`
	Document *DocumentStats `bson:"document"`
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

// OpenCursorStats stores information related to open cursor metrics
type OpenCursorStats struct {
	NoTimeout int64 `bson:"noTimeout"`
	Pinned    int64 `bson:"pinned"`
	Total     int64 `bson:"total"`
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

	// The time at which this StatLine was generated.
	Time time.Time

	// The last time at which this StatLine was printed to output.
	LastPrinted time.Time

	// Opcounter fields
	Insert, Query, Update, Delete, GetMore, Command int64

	// TTL fields
	Passes, DeletedDocuments int64

	// Cursor fields
	TimedOutC                   int64
	NoTimeoutC, PinnedC, TotalC int64

	// Document fields
	DeletedD, InsertedD, ReturnedD, UpdatedD int64

	// Connection fields
	CurrentC, AvailableC, TotalCreatedC int64

	// Collection locks (3.0 mmap only)
	CollectionLocks *CollectionLockStatus

	// Cache utilization (wiredtiger only)
	CacheDirtyPercent float64
	CacheUsedPercent  float64

	// Cache ultilization extended (wiredtiger only)
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
	ServerEvictingPages       int64
	WorkerThreadEvictingPages int64

	// Replicated Opcounter fields
	InsertR, QueryR, UpdateR, DeleteR, GetMoreR, CommandR int64
	ReplLag                                               int64
	OplogTimeDiff                                         int64
	Flushes                                               int64
	Mapped, Virtual, Resident, NonMapped                  int64
	Faults                                                int64
	HighestLocked                                         *LockStatus
	QueuedReaders, QueuedWriters                          int64
	ActiveReaders, ActiveWriters                          int64
	NetIn, NetOut                                         int64
	NumConnections                                        int64
	ReplSetName                                           string
	NodeType                                              string
	NodeState                                             string

	// Cluster fields
	JumboChunksCount int64

	// DB stats field
	DbStatsLines []DbStatLine

	// Shard stats
	TotalInUse, TotalAvailable, TotalCreated, TotalRefreshing int64

	// Shard Hosts stats field
	ShardHostStatsLines map[string]ShardHostStatLine
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
}

type ShardHostStatLine struct {
	InUse      int64
	Available  int64
	Created    int64
	Refreshing int64
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

func diff(newVal, oldVal, sampleTime int64) int64 {
	d := newVal - oldVal
	if d < 0 {
		d = newVal
	}
	return d / sampleTime
}

// NewStatLine constructs a StatLine object from two MongoStatus objects.
func NewStatLine(oldMongo, newMongo MongoStatus, key string, all bool, sampleSecs int64) *StatLine {
	oldStat := *oldMongo.ServerStatus
	newStat := *newMongo.ServerStatus

	returnVal := &StatLine{
		Key:       key,
		Host:      newStat.Host,
		Mapped:    -1,
		Virtual:   -1,
		Resident:  -1,
		NonMapped: -1,
		Faults:    -1,
	}

	// set the storage engine appropriately
	if newStat.StorageEngine != nil && newStat.StorageEngine["name"] != "" {
		returnVal.StorageEngine = newStat.StorageEngine["name"]
	} else {
		returnVal.StorageEngine = "mmapv1"
	}

	if newStat.Opcounters != nil && oldStat.Opcounters != nil {
		returnVal.Insert = diff(newStat.Opcounters.Insert, oldStat.Opcounters.Insert, sampleSecs)
		returnVal.Query = diff(newStat.Opcounters.Query, oldStat.Opcounters.Query, sampleSecs)
		returnVal.Update = diff(newStat.Opcounters.Update, oldStat.Opcounters.Update, sampleSecs)
		returnVal.Delete = diff(newStat.Opcounters.Delete, oldStat.Opcounters.Delete, sampleSecs)
		returnVal.GetMore = diff(newStat.Opcounters.GetMore, oldStat.Opcounters.GetMore, sampleSecs)
		returnVal.Command = diff(newStat.Opcounters.Command, oldStat.Opcounters.Command, sampleSecs)
	}

	if newStat.Metrics != nil && oldStat.Metrics != nil {
		if newStat.Metrics.TTL != nil && oldStat.Metrics.TTL != nil {
			returnVal.Passes = diff(newStat.Metrics.TTL.Passes, oldStat.Metrics.TTL.Passes, sampleSecs)
			returnVal.DeletedDocuments = diff(newStat.Metrics.TTL.DeletedDocuments, oldStat.Metrics.TTL.DeletedDocuments, sampleSecs)
		}
		if newStat.Metrics.Cursor != nil && oldStat.Metrics.Cursor != nil {
			returnVal.TimedOutC = diff(newStat.Metrics.Cursor.TimedOut, oldStat.Metrics.Cursor.TimedOut, sampleSecs)
			if newStat.Metrics.Cursor.Open != nil && oldStat.Metrics.Cursor.Open != nil {
				returnVal.NoTimeoutC = diff(newStat.Metrics.Cursor.Open.NoTimeout, oldStat.Metrics.Cursor.Open.NoTimeout, sampleSecs)
				returnVal.PinnedC = diff(newStat.Metrics.Cursor.Open.Pinned, oldStat.Metrics.Cursor.Open.Pinned, sampleSecs)
				returnVal.TotalC = diff(newStat.Metrics.Cursor.Open.Total, oldStat.Metrics.Cursor.Open.Total, sampleSecs)
			}
		}
	}

	if newStat.OpcountersRepl != nil && oldStat.OpcountersRepl != nil {
		returnVal.InsertR = diff(newStat.OpcountersRepl.Insert, oldStat.OpcountersRepl.Insert, sampleSecs)
		returnVal.QueryR = diff(newStat.OpcountersRepl.Query, oldStat.OpcountersRepl.Query, sampleSecs)
		returnVal.UpdateR = diff(newStat.OpcountersRepl.Update, oldStat.OpcountersRepl.Update, sampleSecs)
		returnVal.DeleteR = diff(newStat.OpcountersRepl.Delete, oldStat.OpcountersRepl.Delete, sampleSecs)
		returnVal.GetMoreR = diff(newStat.OpcountersRepl.GetMore, oldStat.OpcountersRepl.GetMore, sampleSecs)
		returnVal.CommandR = diff(newStat.OpcountersRepl.Command, oldStat.OpcountersRepl.Command, sampleSecs)
	}

	returnVal.CacheDirtyPercent = -1
	returnVal.CacheUsedPercent = -1
	if newStat.WiredTiger != nil && oldStat.WiredTiger != nil {
		returnVal.Flushes = newStat.WiredTiger.Transaction.TransCheckpoints - oldStat.WiredTiger.Transaction.TransCheckpoints
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
		returnVal.ServerEvictingPages = newStat.WiredTiger.Cache.ServerEvictingPages
		returnVal.WorkerThreadEvictingPages = newStat.WiredTiger.Cache.WorkerThreadEvictingPages
	} else if newStat.BackgroundFlushing != nil && oldStat.BackgroundFlushing != nil {
		returnVal.Flushes = newStat.BackgroundFlushing.Flushes - oldStat.BackgroundFlushing.Flushes
	}

	returnVal.Time = newMongo.SampleTime
	returnVal.IsMongos =
		(newStat.ShardCursorType != nil || strings.HasPrefix(newStat.Process, MongosProcess))

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
		setName, isReplSet := newStat.Repl.SetName.(string)
		if isReplSet {
			returnVal.ReplSetName = setName
		}
		// BEGIN code modification
		if newStat.Repl.IsMaster.(bool) {
			returnVal.NodeType = "PRI"
		} else if newStat.Repl.Secondary != nil && newStat.Repl.Secondary.(bool) {
			returnVal.NodeType = "SEC"
		} else if newStat.Repl.ArbiterOnly != nil && newStat.Repl.ArbiterOnly.(bool) {
			returnVal.NodeType = "ARB"
		} else {
			returnVal.NodeType = "UNK"
		}
		// END code modification
	} else if returnVal.IsMongos {
		returnVal.NodeType = "RTR"
	}

	if oldStat.ExtraInfo != nil && newStat.ExtraInfo != nil &&
		oldStat.ExtraInfo.PageFaults != nil && newStat.ExtraInfo.PageFaults != nil {
		returnVal.Faults = diff(*(newStat.ExtraInfo.PageFaults), *(oldStat.ExtraInfo.PageFaults), sampleSecs)
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

				var timeDiffMillis int64
				timeDiffMillis = newStat.UptimeMillis - oldStat.UptimeMillis

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
		hasWT := (newStat.WiredTiger != nil && oldStat.WiredTiger != nil)
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
		} else if newStat.GlobalLock.ActiveClients != nil {
			returnVal.ActiveReaders = newStat.GlobalLock.ActiveClients.Readers
			returnVal.ActiveWriters = newStat.GlobalLock.ActiveClients.Writers
		}
	}

	if oldStat.Network != nil && newStat.Network != nil {
		returnVal.NetIn = diff(newStat.Network.BytesIn, oldStat.Network.BytesIn, sampleSecs)
		returnVal.NetOut = diff(newStat.Network.BytesOut, oldStat.Network.BytesOut, sampleSecs)
	}

	if newStat.Connections != nil {
		returnVal.NumConnections = newStat.Connections.Current
	}

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
				if member.State == 1 {
					// I'm the master
					returnVal.ReplLag = 0
					break
				} else {
					// I'm secondary
					me = member
				}
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

	newClusterStat := *newMongo.ClusterStatus
	returnVal.JumboChunksCount = newClusterStat.JumboChunksCount
	returnVal.OplogTimeDiff = newMongo.OplogStats.TimeDiff

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
		}
		returnVal.DbStatsLines = append(returnVal.DbStatsLines, *dbStatLine)
	}

	// Set shard stats
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

	return returnVal
}
