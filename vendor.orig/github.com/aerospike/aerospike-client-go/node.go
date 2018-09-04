// Copyright 2013-2017 Aerospike, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package aerospike

import (
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	. "github.com/aerospike/aerospike-client-go/logger"
	. "github.com/aerospike/aerospike-client-go/types"
	. "github.com/aerospike/aerospike-client-go/types/atomic"
)

const (
	_PARTITIONS = 4096
)

// Node represents an Aerospike Database Server Node
type Node struct {
	cluster *Cluster
	name    string
	host    *Host
	aliases atomic.Value //[]*Host

	// tendConn reserves a connection for tend so that it won't have to
	// wait in queue for connections, since that will cause starvation
	// and the node being dropped under load.
	tendConn     *Connection
	tendConnLock sync.Mutex // All uses of tend connection should be synchronized

	peersGeneration AtomicInt
	peersCount      AtomicInt

	connections     connectionQueue //AtomicQueue //ArrayBlockingQueue<*Connection>
	connectionCount AtomicInt
	health          AtomicInt //AtomicInteger

	partitionMap        partitionMap
	partitionGeneration AtomicInt
	referenceCount      AtomicInt
	failures            AtomicInt
	partitionChanged    AtomicBool

	active AtomicBool

	supportsFloat, supportsBatchIndex, supportsReplicasAll, supportsGeo, supportsPeers AtomicBool
}

// NewNode initializes a server node with connection parameters.
func newNode(cluster *Cluster, nv *nodeValidator) *Node {
	newNode := &Node{
		cluster: cluster,
		name:    nv.name,
		// address: nv.primaryAddress,
		host: nv.primaryHost,

		// Assign host to first IP alias because the server identifies nodes
		// by IP address (not hostname).
		connections:         *newConnectionQueue(cluster.clientPolicy.ConnectionQueueSize), //*NewAtomicQueue(cluster.clientPolicy.ConnectionQueueSize),
		connectionCount:     *NewAtomicInt(0),
		peersGeneration:     *NewAtomicInt(-1),
		partitionGeneration: *NewAtomicInt(-2),
		referenceCount:      *NewAtomicInt(0),
		failures:            *NewAtomicInt(0),
		active:              *NewAtomicBool(true),
		partitionChanged:    *NewAtomicBool(false),

		supportsFloat:       *NewAtomicBool(nv.supportsFloat),
		supportsBatchIndex:  *NewAtomicBool(nv.supportsBatchIndex),
		supportsReplicasAll: *NewAtomicBool(nv.supportsReplicasAll),
		supportsGeo:         *NewAtomicBool(nv.supportsGeo),
		supportsPeers:       *NewAtomicBool(nv.supportsPeers),
	}

	newNode.aliases.Store(nv.aliases)

	return newNode
}

// Refresh requests current status from server node, and updates node with the result.
func (nd *Node) Refresh(peers *peers) error {
	if !nd.active.Get() {
		return nil
	}

	// Close idleConnections
	defer nd.dropIdleConnections()

	nd.referenceCount.Set(0)

	if peers.usePeers.Get() {
		infoMap, err := nd.RequestInfo("node", "peers-generation", "partition-generation")
		if err != nil {
			nd.refreshFailed(err)
			return err
		}

		if err := nd.verifyNodeName(infoMap); err != nil {
			nd.refreshFailed(err)
			return err
		}

		if err := nd.verifyPeersGeneration(infoMap, peers); err != nil {
			nd.refreshFailed(err)
			return err
		}

		if err := nd.verifyPartitionGeneration(infoMap); err != nil {
			nd.refreshFailed(err)
			return err
		}
	} else {
		commands := []string{"node", "partition-generation", nd.cluster.clientPolicy.serviceString()}

		infoMap, err := nd.RequestInfo(commands...)
		if err != nil {
			nd.refreshFailed(err)
			return err
		}

		if err := nd.verifyNodeName(infoMap); err != nil {
			nd.refreshFailed(err)
			return err
		}

		if err = nd.verifyPartitionGeneration(infoMap); err != nil {
			nd.refreshFailed(err)
			return err
		}

		if err = nd.addFriends(infoMap, peers); err != nil {
			nd.refreshFailed(err)
			return err
		}
	}
	nd.failures.Set(0)
	peers.refreshCount.IncrementAndGet()
	nd.referenceCount.IncrementAndGet()

	return nil
}

func (nd *Node) verifyNodeName(infoMap map[string]string) error {
	infoName, exists := infoMap["node"]

	if !exists || len(infoName) == 0 {
		return NewAerospikeError(INVALID_NODE_ERROR, "Node name is empty")
	}

	if !(nd.name == infoName) {
		// Set node to inactive immediately.
		nd.active.Set(false)
		return NewAerospikeError(INVALID_NODE_ERROR, "Node name has changed. Old="+nd.name+" New="+infoName)
	}
	return nil
}

func (nd *Node) verifyPeersGeneration(infoMap map[string]string, peers *peers) error {
	genString := infoMap["peers-generation"]
	if len(genString) == 0 {
		return NewAerospikeError(PARSE_ERROR, "peers-generation is empty")
	}

	gen, err := strconv.Atoi(genString)
	if err != nil {
		return NewAerospikeError(PARSE_ERROR, "peers-generation is not a number: "+genString)
	}

	peers.genChanged.Or(nd.peersGeneration.Get() != gen)
	return nil
}

func (nd *Node) verifyPartitionGeneration(infoMap map[string]string) error {
	genString := infoMap["partition-generation"]

	if len(genString) == 0 {
		return NewAerospikeError(PARSE_ERROR, "partition-generation is empty")
	}

	gen, err := strconv.Atoi(genString)
	if err != nil {
		return NewAerospikeError(PARSE_ERROR, "partition-generation is not a number:"+genString)
	}

	if nd.partitionGeneration.Get() != gen {
		nd.partitionChanged.Set(true)
	}
	return nil
}

func (nd *Node) addFriends(infoMap map[string]string, peers *peers) error {
	friendString, exists := infoMap[nd.cluster.clientPolicy.serviceString()]

	if !exists || len(friendString) == 0 {
		nd.peersCount.Set(0)
		return nil
	}

	friendNames := strings.Split(friendString, ";")
	nd.peersCount.Set(len(friendNames))

	for _, friend := range friendNames {
		friendInfo := strings.Split(friend, ":")

		if len(friendInfo) != 2 {
			Logger.Error("Node info from asinfo:services is malformed. Expected HOST:PORT, but got `%s`", friend)
			continue
		}

		hostName := friendInfo[0]
		port, _ := strconv.Atoi(friendInfo[1])

		if nd.cluster.clientPolicy.IpMap != nil {
			if alternativeHost, ok := nd.cluster.clientPolicy.IpMap[hostName]; ok {
				hostName = alternativeHost
			}
		}

		host := NewHost(hostName, port)
		node := nd.cluster.findAlias(host)

		if node != nil {
			node.referenceCount.IncrementAndGet()
		} else {
			if !peers.hostExists(*host) {
				nd.prepareFriend(host, peers)
			}
		}
	}

	return nil
}

func (nd *Node) prepareFriend(host *Host, peers *peers) bool {
	nv := &nodeValidator{}
	if err := nv.validateNode(nd.cluster, host); err != nil {
		Logger.Warn("Adding node `%s` failed: ", host, err)
		return false
	}

	node := peers.nodeByName(nv.name)

	if node != nil {
		// Duplicate node name found.  This usually occurs when the server
		// services list contains both internal and external IP addresses
		// for the same node.
		peers.addHost(*host)
		node.addAlias(host)
		return true
	}

	// Check for duplicate nodes in cluster.
	node = nd.cluster.nodesMap.Get().(map[string]*Node)[nv.name]

	if node != nil {
		peers.addHost(*host)
		node.addAlias(host)
		node.referenceCount.IncrementAndGet()
		nd.cluster.addAlias(host, node)
		return true
	}

	node = nd.cluster.createNode(nv)
	peers.addHost(*host)
	peers.addNode(nv.name, node)
	return true
}

func (nd *Node) refreshPeers(peers *peers) {
	// Do not refresh peers when node connection has already failed during this cluster tend iteration.
	if nd.failures.Get() > 0 || !nd.active.Get() {
		return
	}

	peerParser, err := parsePeers(nd.cluster, nd)
	if err != nil {
		Logger.Debug("Parsing peers failed: %s", err)
		nd.refreshFailed(err)
		return
	}

	peers.appendPeers(peerParser.peers)
	nd.peersGeneration.Set(int(peerParser.generation()))
	nd.peersCount.Set(len(peers.peers()))
	peers.refreshCount.IncrementAndGet()
}

func (nd *Node) refreshPartitions(peers *peers) {
	// Do not refresh peers when node connection has already failed during this cluster tend iteration.
	// Also, avoid "split cluster" case where this node thinks it's a 1-node cluster.
	// Unchecked, such a node can dominate the partition map and cause all other
	// nodes to be dropped.
	if nd.failures.Get() > 0 || !nd.active.Get() || (nd.peersCount.Get() == 0 && peers.refreshCount.Get() > 1) {
		return
	}

	parser, err := newPartitionParser(nd, _PARTITIONS, nd.cluster.clientPolicy.RequestProleReplicas)
	if err != nil {
		nd.refreshFailed(err)
		return
	}

	if parser.generation != nd.partitionGeneration.Get() {
		Logger.Info("Node %s partition generation %d changed to %d", nd.GetName(), nd.partitionGeneration.Get(), parser.getGeneration())
		nd.partitionMap = parser.getPartitionMap()
		nd.partitionChanged.Set(true)
		nd.partitionGeneration.Set(parser.getGeneration())
	}
}

func (nd *Node) refreshFailed(e error) {
	nd.failures.IncrementAndGet()

	// Only log message if cluster is still active.
	if nd.cluster.IsConnected() {
		Logger.Warn("Node `%s` refresh failed: `%s`", nd, e)
	}
}

// dropIdleConnections picks a connection from the head of the connection pool queue
// if that connection is idle, it drops it and takes the next one until it picks
// a fresh connection or exhaust the queue.
func (nd *Node) dropIdleConnections() {
	nd.connections.DropIdle()
}

// GetConnection gets a connection to the node.
// If no pooled connection is available, a new connection will be created, unless
// ClientPolicy.MaxQueueSize number of connections are already created.
// This method will retry to retrieve a connection in case the connection pool
// is empty, until timeout is reached.
func (nd *Node) GetConnection(timeout time.Duration) (conn *Connection, err error) {
	deadline := time.Now().Add(timeout)
	if timeout == 0 {
		deadline = time.Now().Add(time.Second)
	}

CL:
	// try to acquire a connection; if the connection pool is empty, retry until
	// timeout occures. If no timeout is set, will retry indefinitely.
	conn, err = nd.getConnection(timeout)
	if err != nil {
		if err == ErrConnectionPoolEmpty && nd.IsActive() && time.Now().Before(deadline) {
			// give the scheduler time to breath; affects latency minimally, but throughput drastically
			time.Sleep(time.Microsecond)
			goto CL
		}

		return nil, err
	}

	return conn, nil
}

// getConnection gets a connection to the node.
// If no pooled connection is available, a new connection will be created.
// This method does not include logic to retry in case the connection pool is empty
func (nd *Node) getConnection(timeout time.Duration) (conn *Connection, err error) {
	return nd.getConnectionWithHint(timeout, 0)
}

// getConnectionWithHint gets a connection to the node.
// If no pooled connection is available, a new connection will be created.
// This method does not include logic to retry in case the connection pool is empty
func (nd *Node) getConnectionWithHint(timeout time.Duration, hint byte) (conn *Connection, err error) {
	// try to get a valid connection from the connection pool
	for t := nd.connections.Poll(hint); t != nil; t = nd.connections.Poll(hint) {
		conn = t //.(*Connection)
		if conn.IsConnected() {
			break
		}
		conn.Close()
		conn = nil
	}

	if conn == nil {
		cc := nd.connectionCount.IncrementAndGet()

		// if connection count is limited and enough connections are already created, don't create a new one
		if nd.cluster.clientPolicy.LimitConnectionsToQueueSize && cc > nd.cluster.clientPolicy.ConnectionQueueSize {
			nd.connectionCount.DecrementAndGet()
			return nil, ErrConnectionPoolEmpty
		}

		if conn, err = NewSecureConnection(&nd.cluster.clientPolicy, nd.host); err != nil {
			nd.connectionCount.DecrementAndGet()
			return nil, err
		}
		conn.node = nd

		// need to authenticate
		if err = conn.Authenticate(nd.cluster.user, nd.cluster.Password()); err != nil {
			// Socket not authenticated. Do not put back into pool.
			conn.Close()
			return nil, err
		}
	}

	if err = conn.SetTimeout(timeout); err != nil {
		// Do not put back into pool.
		conn.Close()
		return nil, err
	}

	conn.setIdleTimeout(nd.cluster.clientPolicy.IdleTimeout)
	conn.refresh()

	return conn, nil
}

// PutConnection puts back a connection to the pool.
// If connection pool is full, the connection will be
// closed and discarded.
func (nd *Node) putConnectionWithHint(conn *Connection, hint byte) {
	conn.refresh()
	if !nd.active.Get() || !nd.connections.Offer(conn, hint) {
		conn.Close()
	}
}

// PutConnection puts back a connection to the pool.
// If connection pool is full, the connection will be
// closed and discarded.
func (nd *Node) PutConnection(conn *Connection) {
	nd.putConnectionWithHint(conn, 0)
}

// InvalidateConnection closes and discards a connection from the pool.
func (nd *Node) InvalidateConnection(conn *Connection) {
	conn.Close()
}

// GetHost retrieves host for the node.
func (nd *Node) GetHost() *Host {
	return nd.host
}

// IsActive Checks if the node is active.
func (nd *Node) IsActive() bool {
	return nd != nil && nd.active.Get() && nd.partitionGeneration.Get() >= -1
}

// GetName returns node name.
func (nd *Node) GetName() string {
	return nd.name
}

// GetAliases returns node aliases.
func (nd *Node) GetAliases() []*Host {
	return nd.aliases.Load().([]*Host)
}

// Sets node aliases
func (nd *Node) setAliases(aliases []*Host) {
	nd.aliases.Store(aliases)
}

// AddAlias adds an alias for the node
func (nd *Node) addAlias(aliasToAdd *Host) {
	// Aliases are only referenced in the cluster tend goroutine,
	// so synchronization is not necessary.
	aliases := nd.GetAliases()
	if aliases == nil {
		aliases = []*Host{}
	}

	aliases = append(aliases, aliasToAdd)
	nd.setAliases(aliases)
}

// Close marks node as inactive and closes all of its pooled connections.
func (nd *Node) Close() {
	nd.active.Set(false)
	nd.closeConnections()
}

// String implements stringer interface
func (nd *Node) String() string {
	return nd.name + " " + nd.host.String()
}

func (nd *Node) closeConnections() {
	for conn := nd.connections.Poll(0); conn != nil; conn = nd.connections.Poll(0) {
		// conn.(*Connection).Close()
		conn.Close()
	}
}

// Equals compares equality of two nodes based on their names.
func (nd *Node) Equals(other *Node) bool {
	return nd != nil && other != nil && (nd == other || nd.name == other.name)
}

// MigrationInProgress determines if the node is participating in a data migration
func (nd *Node) MigrationInProgress() (bool, error) {
	values, err := RequestNodeStats(nd)
	if err != nil {
		return false, err
	}

	// if the migration_progress_send exists and is not `0`, then migration is in progress
	if migration, exists := values["migrate_progress_send"]; exists && migration != "0" {
		return true, nil
	}

	// migration not in progress
	return false, nil
}

// WaitUntillMigrationIsFinished will block until migration operations are finished.
func (nd *Node) WaitUntillMigrationIsFinished(timeout time.Duration) (err error) {
	if timeout <= 0 {
		timeout = _NO_TIMEOUT
	}
	done := make(chan error)

	go func() {
		// this function is guaranteed to return after timeout
		// no go routines will be leaked
		for {
			if res, err := nd.MigrationInProgress(); err != nil || !res {
				done <- err
				return
			}
		}
	}()

	dealine := time.After(timeout)
	select {
	case <-dealine:
		return NewAerospikeError(TIMEOUT)
	case err = <-done:
		return err
	}
}

// initTendConn sets up a connection to be used for info requests.
// The same connection will be used for tend.
func (nd *Node) initTendConn(timeout time.Duration) error {
	if nd.tendConn == nil || !nd.tendConn.IsConnected() {
		// Tend connection required a long timeout
		tendConn, err := nd.GetConnection(timeout)
		if err != nil {
			return err
		}

		nd.tendConn = tendConn
	}

	// Set timeout for tend conn
	return nd.tendConn.SetTimeout(timeout)
}

// RequestInfo gets info values by name from the specified database server node.
func (nd *Node) RequestInfo(name ...string) (map[string]string, error) {
	nd.tendConnLock.Lock()
	defer nd.tendConnLock.Unlock()

	if err := nd.initTendConn(nd.cluster.clientPolicy.Timeout); err != nil {
		return nil, err
	}

	response, err := RequestInfo(nd.tendConn, name...)
	if err != nil {
		nd.tendConn.Close()
		return nil, err
	}
	return response, nil
}

// requestRawInfo gets info values by name from the specified database server node.
// It won't parse the results.
func (nd *Node) requestRawInfo(name ...string) (*info, error) {
	nd.tendConnLock.Lock()
	defer nd.tendConnLock.Unlock()

	if err := nd.initTendConn(nd.cluster.clientPolicy.Timeout); err != nil {
		return nil, err
	}

	response, err := newInfo(nd.tendConn, name...)
	if err != nil {
		nd.tendConn.Close()
		return nil, err
	}
	return response, nil
}
