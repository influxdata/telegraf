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
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	. "github.com/aerospike/aerospike-client-go/logger"

	. "github.com/aerospike/aerospike-client-go/types"
	. "github.com/aerospike/aerospike-client-go/types/atomic"
)

type partitionMap map[string][][]*Node

// String implements stringer interface for partitionMap
func (pm partitionMap) clone() partitionMap {
	// Make shallow copy of map.
	pmap := make(partitionMap, len(pm))
	for ns, replArr := range pm {
		newReplArr := make([][]*Node, len(replArr))
		for i, nArr := range replArr {
			newNArr := make([]*Node, len(nArr))
			copy(newNArr, nArr)
			newReplArr[i] = newNArr
		}
		pmap[ns] = newReplArr
	}
	return pmap
}

// String implements stringer interface for partitionMap
func (pm partitionMap) merge(partMap partitionMap) {
	// merge partitions; iterate over the new partition and update the old one
	for ns, replicaArray := range partMap {
		if pm[ns] == nil {
			pm[ns] = make([][]*Node, len(replicaArray))
		}

		for i, nodeArray := range replicaArray {
			if pm[ns][i] == nil {
				pm[ns][i] = make([]*Node, len(nodeArray))
			}

			for j, node := range nodeArray {
				if node != nil {
					pm[ns][i][j] = node
				}
			}
		}
	}
}

// String implements stringer interface for partitionMap
func (pm partitionMap) String() string {
	res := bytes.Buffer{}
	for ns, replicaArray := range pm {
		for i, nodeArray := range replicaArray {
			for j, node := range nodeArray {
				res.WriteString(ns)
				res.WriteString(",")
				res.WriteString(strconv.Itoa(i))
				res.WriteString(",")
				res.WriteString(strconv.Itoa(j))
				res.WriteString(",")
				if node != nil {
					res.WriteString(node.String())
				} else {
					res.WriteString("NIL")
				}
				res.WriteString("\n")
			}
		}
	}
	return res.String()
}

// Cluster encapsulates the aerospike cluster nodes and manages
// them.
type Cluster struct {
	// Initial host nodes specified by user.
	seeds *SyncVal //[]*Host

	// All aliases for all nodes in cluster.
	// Only accessed within cluster tend thread.
	aliases *SyncVal //map[Host]*Node

	// Map of active nodes in cluster.
	// Only accessed within cluster tend thread.
	nodesMap *SyncVal //map[string]*Node

	// Active nodes in cluster.
	nodes *SyncVal //[]*Node

	// Hints for best node for a partition
	partitionWriteMap    atomic.Value //partitionMap
	partitionUpdateMutex sync.Mutex

	clientPolicy ClientPolicy

	nodeIndex    uint64 // only used via atomic operations
	replicaIndex uint64 // only used via atomic operations

	wgTend      sync.WaitGroup
	tendChannel chan struct{}
	closed      AtomicBool

	// Aerospike v3.6.0+
	supportsFloat, supportsBatchIndex, supportsReplicasAll, supportsGeo *AtomicBool
	requestProleReplicas                                                *AtomicBool

	// User name in UTF-8 encoded bytes.
	user string

	// Password in hashed format in bytes.
	password *SyncVal // []byte
}

// NewCluster generates a Cluster instance.
func NewCluster(policy *ClientPolicy, hosts []*Host) (*Cluster, error) {
	// Default TLS names when TLS enabled.
	newHosts := make([]*Host, 0, len(hosts))
	if policy.TlsConfig != nil && !policy.TlsConfig.InsecureSkipVerify {
		useClusterName := len(policy.ClusterName) > 0

		for _, host := range hosts {
			nh := *host
			if nh.TLSName == "" {
				if useClusterName {
					nh.TLSName = policy.ClusterName
				} else {
					nh.TLSName = host.Name
				}
			}
			newHosts = append(newHosts, &nh)
		}
		hosts = newHosts
	}

	newCluster := &Cluster{
		clientPolicy: *policy,
		tendChannel:  make(chan struct{}),

		seeds:    NewSyncVal(hosts),
		aliases:  NewSyncVal(make(map[Host]*Node)),
		nodesMap: NewSyncVal(make(map[string]*Node)),
		nodes:    NewSyncVal([]*Node{}),

		password: NewSyncVal(nil),

		supportsFloat:        NewAtomicBool(false),
		supportsBatchIndex:   NewAtomicBool(false),
		supportsReplicasAll:  NewAtomicBool(false),
		supportsGeo:          NewAtomicBool(false),
		requestProleReplicas: NewAtomicBool(policy.RequestProleReplicas),
	}

	newCluster.partitionWriteMap.Store(make(partitionMap))

	// setup auth info for cluster
	if policy.RequiresAuthentication() {
		newCluster.user = policy.User
		hashedPass, err := hashPassword(policy.Password)
		if err != nil {
			return nil, err
		}
		newCluster.password = NewSyncVal(hashedPass)
	}

	// try to seed connections for first use
	err := newCluster.waitTillStabilized()

	// apply policy rules
	if policy.FailIfNotConnected && !newCluster.IsConnected() {
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("Failed to connect to host(s): %v. The network connection(s) to cluster nodes may have timed out, or the cluster may be in a state of flux.", hosts)
	}

	// start up cluster maintenance go routine
	newCluster.wgTend.Add(1)
	go newCluster.clusterBoss(&newCluster.clientPolicy)

	Logger.Debug("New cluster initialized and ready to be used...")
	return newCluster, err
}

// String implements the stringer interface
func (clstr *Cluster) String() string {
	return fmt.Sprintf("%v", clstr.nodes)
}

// Maintains the cluster on intervals.
// All clean up code for cluster is here as well.
func (clstr *Cluster) clusterBoss(policy *ClientPolicy) {
	defer clstr.wgTend.Done()

	tendInterval := policy.TendInterval
	if tendInterval <= 10*time.Millisecond {
		tendInterval = 10 * time.Millisecond
	}

Loop:
	for {
		select {
		case <-clstr.tendChannel:
			// tend channel closed
			Logger.Debug("Tend channel closed. Shutting down the cluster...")
			break Loop
		case <-time.After(tendInterval):
			tm := time.Now()
			if err := clstr.tend(); err != nil {
				Logger.Warn(err.Error())
			}

			// Tending took longer than requested tend interval.
			// Tending is too slow for the cluster, and may be falling behind scheule.
			if tendDuration := time.Since(tm); tendDuration > clstr.clientPolicy.TendInterval {
				Logger.Warn("Tending took %s, while your requested ClientPolicy.TendInterval is %s. Tends are slower than the interval, and may be falling behind the changes in the cluster.", tendDuration, clstr.clientPolicy.TendInterval)
			}
		}
	}

	// cleanup code goes here
	clstr.closed.Set(true)

	// close the nodes
	nodeArray := clstr.GetNodes()
	for _, node := range nodeArray {
		node.Close()
	}
}

// AddSeeds adds new hosts to the cluster.
// They will be added to the cluster on next tend call.
func (clstr *Cluster) AddSeeds(hosts []*Host) {
	clstr.seeds.Update(func(val interface{}) (interface{}, error) {
		seeds := val.([]*Host)
		seeds = append(seeds, hosts...)
		return seeds, nil
	})
}

// Updates cluster state
func (clstr *Cluster) tend() error {

	nodes := clstr.GetNodes()
	nodeCountBeforeTend := len(nodes)

	// All node additions/deletions are performed in tend goroutine.
	// If active nodes don't exist, seed cluster.
	if len(nodes) == 0 {
		Logger.Info("No connections available; seeding...")
		if newNodesFound, err := clstr.seedNodes(); !newNodesFound {
			return err
		}

		// refresh nodes list after seeding
		nodes = clstr.GetNodes()
	}

	peers := newPeers(len(nodes)+16, 16)

	floatSupport := true
	batchIndexSupport := true
	replicasAllSupport := true
	geoSupport := true

	for _, node := range nodes {
		// Clear node reference counts.
		node.referenceCount.Set(0)
		node.partitionChanged.Set(false)
		if !node.supportsPeers.Get() {
			peers.usePeers.Set(false)
		}
	}

	wg := sync.WaitGroup{}
	wg.Add(len(nodes))
	for _, node := range nodes {
		go func(node *Node) {
			defer wg.Done()
			if err := node.Refresh(peers); err != nil {
				Logger.Debug("Error occured while refreshing node: %s", node.String())
			}
		}(node)
	}
	wg.Wait()

	// Refresh peers when necessary.
	if peers.usePeers.Get() && (peers.genChanged.Get() || len(peers.peers()) != nodeCountBeforeTend) {
		// Refresh peers for all nodes that responded the first time even if only one node's peers changed.
		peers.refreshCount.Set(0)

		wg.Add(len(nodes))
		for _, node := range nodes {
			go func(node *Node) {
				defer wg.Done()
				node.refreshPeers(peers)
			}(node)
		}
		wg.Wait()
	}

	// find the first host that connects
	for _, _peer := range peers.peers() {
		if clstr.peerExists(peers, _peer.nodeName) {
			// Node already exists. Do not even try to connect to hosts.
			continue
		}

		wg.Add(1)
		go func(__peer *peer) {
			defer wg.Done()
			for _, host := range __peer.hosts {
				// attempt connection to the host
				nv := nodeValidator{}
				if err := nv.validateNode(clstr, host); err != nil {
					Logger.Warn("Add node `%s` failed: `%s`", host, err)
					continue
				}

				// Must look for new node name in the unlikely event that node names do not agree.
				if __peer.nodeName != nv.name {
					Logger.Warn("Peer node `%s` is different than actual node `%s` for host `%s`", __peer.nodeName, nv.name, host)
				}

				if clstr.peerExists(peers, nv.name) {
					// Node already exists. Do not even try to connect to hosts.
					break
				}

				// Create new node.
				node := clstr.createNode(&nv)
				peers.addNode(nv.name, node)
				node.refreshPartitions(peers)
				break
			}
		}(_peer)
	}

	// Refresh partition map when necessary.
	wg.Add(len(nodes))
	for _, node := range nodes {
		go func(node *Node) {
			defer wg.Done()
			if node.partitionChanged.Get() {
				node.refreshPartitions(peers)
			}
		}(node)
	}

	// This waits for the both steps above
	wg.Wait()

	if peers.genChanged.Get() || !peers.usePeers.Get() {
		// Handle nodes changes determined from refreshes.
		removeList := clstr.findNodesToRemove(peers.refreshCount.Get())

		// Remove nodes in a batch.
		if len(removeList) > 0 {
			for _, n := range removeList {
				Logger.Debug("The following nodes will be removed: %s", n)
			}
			clstr.removeNodes(removeList)
		}
	}

	// Add nodes in a batch.
	if len(peers.nodes()) > 0 {
		clstr.addNodes(peers.nodes())
	}

	if !floatSupport {
		Logger.Warn("Some cluster nodes do not support float type. Disabling native float support in the client library...")
	}

	// Disable prole requests if some nodes don't support it.
	if clstr.clientPolicy.RequestProleReplicas && !replicasAllSupport {
		Logger.Warn("Some nodes don't support 'replicas-all'. Will use 'replicas-master' for all nodes.")
	}

	// set the cluster supported features
	clstr.supportsFloat.Set(floatSupport)
	clstr.supportsBatchIndex.Set(batchIndexSupport)
	clstr.supportsReplicasAll.Set(replicasAllSupport)
	clstr.requestProleReplicas.Set(clstr.clientPolicy.RequestProleReplicas && replicasAllSupport)
	clstr.supportsGeo.Set(geoSupport)

	// update all partitions in one go
	var partitionMap partitionMap
	for _, node := range clstr.GetNodes() {
		if node.partitionChanged.Get() {
			if partitionMap == nil {
				partitionMap = clstr.getPartitions().clone()
			}

			partitionMap.merge(node.partitionMap)
		}
	}

	if partitionMap != nil {
		clstr.setPartitions(partitionMap)
	}

	// only log if node count is changed
	if nodeCountBeforeTend != len(clstr.GetNodes()) {
		Logger.Info("Tend finished. Live node count changes from %d to %d", nodeCountBeforeTend, len(clstr.GetNodes()))
	}
	return nil
}

func (clstr *Cluster) peerExists(peers *peers, nodeName string) bool {
	node := clstr.findNodeByName(nodeName)
	if node != nil {
		node.referenceCount.IncrementAndGet()
		return true
	}

	node = peers.nodeByName(nodeName)
	if node != nil {
		node.referenceCount.IncrementAndGet()
		return true
	}

	return false
}

// Tend the cluster until it has stabilized and return control.
// This helps avoid initial database request timeout issues when
// a large number of threads are initiated at client startup.
//
// If the cluster has not stabilized by the timeout, return
// control as well.  Do not return an error since future
// database requests may still succeed.
func (clstr *Cluster) waitTillStabilized() error {
	count := -1

	doneCh := make(chan error, 10)

	// will run until the cluster is stabilized
	go func() {
		var err error
		for {
			if err = clstr.tend(); err != nil {
				if aerr, ok := err.(AerospikeError); ok {
					switch aerr.ResultCode() {
					case NOT_AUTHENTICATED, CLUSTER_NAME_MISMATCH_ERROR:
						doneCh <- err
						return
					}
				}
				Logger.Warn(err.Error())
			}

			// Check to see if cluster has changed since the last Tend().
			// If not, assume cluster has stabilized and return.
			if count == len(clstr.GetNodes()) {
				break
			}

			time.Sleep(time.Millisecond)

			count = len(clstr.GetNodes())
		}
		doneCh <- err
	}()

	select {
	case <-time.After(clstr.clientPolicy.Timeout):
		return errors.New("Connecting to the cluster timed out.")
	case err := <-doneCh:
		return err
	}
}

func (clstr *Cluster) findAlias(alias *Host) *Node {
	res, _ := clstr.aliases.GetSyncedVia(func(val interface{}) (interface{}, error) {
		aliases := val.(map[Host]*Node)
		return aliases[*alias], nil
	})

	return res.(*Node)
}

func (clstr *Cluster) setPartitions(partMap partitionMap) {
	clstr.partitionWriteMap.Store(partMap)
}

func (clstr *Cluster) getPartitions() partitionMap {
	return clstr.partitionWriteMap.Load().(partitionMap)
}

// Adds seeds to the cluster
func (clstr *Cluster) seedNodes() (bool, error) {
	// Must copy array reference for copy on write semantics to work.
	seedArrayIfc, _ := clstr.seeds.GetSyncedVia(func(val interface{}) (interface{}, error) {
		seeds := val.([]*Host)
		seeds_copy := make([]*Host, len(seeds))
		copy(seeds_copy, seeds)

		return seeds_copy, nil
	})
	seedArray := seedArrayIfc.([]*Host)

	successChan := make(chan struct{}, len(seedArray))
	errChan := make(chan error, len(seedArray))

	Logger.Info("Seeding the cluster. Seeds count: %d", len(seedArray))

	// Add all nodes at once to avoid copying entire array multiple times.
	var wg sync.WaitGroup
	wg.Add(len(seedArray))
	for i, seed := range seedArray {
		go func(index int, seed *Host) {
			defer wg.Done()

			nodesToAdd := &nodesToAddT{nodesToAdd: map[string]*Node{}}
			nv := nodeValidator{}
			err := nv.seedNodes(clstr, seed, nodesToAdd)
			if err != nil {
				Logger.Warn("Seed %s failed: %s", seed.String(), err.Error())
				errChan <- err
				return
			}
			clstr.addNodes(nodesToAdd.nodesToAdd)
			successChan <- struct{}{}
		}(i, seed)
	}

	errorList := make([]error, 0, len(seedArray))
	seedCount := len(seedArray)
L:
	for {
		select {
		case err := <-errChan:
			errorList = append(errorList, err)
			seedCount--
			if seedCount <= 0 {
				break L
			}
		case <-successChan:
			// even one seed is enough
			return true, nil
		case <-time.After(clstr.clientPolicy.Timeout):
			// time is up, no seeds found
			wg.Wait()
			break L
		}
	}

	var errStrs []string
	for _, err := range errorList {
		if err != nil {
			if aerr, ok := err.(AerospikeError); ok {
				switch aerr.ResultCode() {
				case NOT_AUTHENTICATED:
					return false, NewAerospikeError(NOT_AUTHENTICATED)
				case CLUSTER_NAME_MISMATCH_ERROR:
					return false, aerr
				}
			}
			errStrs = append(errStrs, err.Error())
		}
	}

	return false, NewAerospikeError(INVALID_NODE_ERROR, "Failed to connect to hosts:"+strings.Join(errStrs, "\n"))
}

func (clstr *Cluster) createNode(nv *nodeValidator) *Node {
	return newNode(clstr, nv)
}

// Finds a node by name in a list of nodes
func (clstr *Cluster) findNodeName(list []*Node, name string) bool {
	for _, node := range list {
		if node.GetName() == name {
			return true
		}
	}
	return false
}

func (clstr *Cluster) addAlias(host *Host, node *Node) {
	if host != nil && node != nil {
		clstr.aliases.Update(func(val interface{}) (interface{}, error) {
			aliases := val.(map[Host]*Node)
			aliases[*host] = node
			return aliases, nil
		})
	}
}

func (clstr *Cluster) removeAlias(alias *Host) {
	if alias != nil {
		clstr.aliases.Update(func(val interface{}) (interface{}, error) {
			aliases := val.(map[Host]*Node)
			delete(aliases, *alias)
			return aliases, nil
		})
	}
}

func (clstr *Cluster) findNodesToRemove(refreshCount int) []*Node {
	nodes := clstr.GetNodes()

	removeList := []*Node{}

	for _, node := range nodes {
		if !node.IsActive() {
			// Inactive nodes must be removed.
			removeList = append(removeList, node)
			continue
		}

		switch len(nodes) {
		case 1:
			// Single node clusters rely on whether it responded to info requests.
			if node.failures.Get() >= 5 {
				// Remove node.  Seeds will be tried in next cluster tend iteration.
				removeList = append(removeList, node)
			}

		case 2:
			// Two node clusters require at least one successful refresh before removing.
			if refreshCount == 1 && node.referenceCount.Get() == 0 && node.failures.Get() > 0 {
				// Node is not referenced nor did it respond.
				removeList = append(removeList, node)
			}

		default:
			// Multi-node clusters require at least one successful refresh before removing
			// or alternatively, if connection to the whle cluster has been cut.
			if (refreshCount >= 1 && node.referenceCount.Get() == 0) || (refreshCount == 0 && node.failures.Get() > 5) {
				// Node is not referenced by other nodes.
				// Check if node responded to info request.
				if node.failures.Get() == 0 {
					// Node is alive, but not referenced by other nodes.  Check if mapped.
					if !clstr.findNodeInPartitionMap(node) {
						// Node doesn't have any partitions mapped to it.
						// There is no point in keeping it in the cluster.
						removeList = append(removeList, node)
					}
				} else {
					// Node not responding. Remove it.
					removeList = append(removeList, node)
				}
			}
		}
	}
	return removeList
}

func (clstr *Cluster) findNodeInPartitionMap(filter *Node) bool {
	partitions := clstr.getPartitions()

	for _, replicaArray := range partitions {
		for _, nodeArray := range replicaArray {
			for _, node := range nodeArray {
				// Use reference equality for performance.
				if node == filter {
					return true
				}
			}
		}
	}
	return false
}

func (clstr *Cluster) addNodes(nodesToAdd map[string]*Node) {
	clstr.nodes.Update(func(val interface{}) (interface{}, error) {
		nodes := val.([]*Node)
		for _, node := range nodesToAdd {
			if node != nil && !clstr.findNodeName(nodes, node.name) {
				Logger.Debug("Adding node %s (%s) to the cluster.", node.name, node.host.String())
				nodes = append(nodes, node)
			}
		}

		nodesMap := make(map[string]*Node, len(nodes))
		nodesAliases := make(map[Host]*Node, len(nodes))
		for i := range nodes {
			nodesMap[nodes[i].name] = nodes[i]

			for _, alias := range nodes[i].GetAliases() {
				nodesAliases[*alias] = nodes[i]
			}
		}

		clstr.nodesMap.Set(nodesMap)
		clstr.aliases.Set(nodesAliases)

		return nodes, nil
	})
}

func (clstr *Cluster) removeNodes(nodesToRemove []*Node) {

	// There is no need to delete nodes from partitionWriteMap because the nodes
	// have already been set to inactive.

	// Cleanup node resources.
	for _, node := range nodesToRemove {
		// Remove node's aliases from cluster alias set.
		// Aliases are only used in tend goroutine, so synchronization is not necessary.
		clstr.aliases.Update(func(val interface{}) (interface{}, error) {
			aliases := val.(map[Host]*Node)
			for _, alias := range node.GetAliases() {
				delete(aliases, *alias)
			}
			return aliases, nil
		})

		clstr.nodesMap.Update(func(val interface{}) (interface{}, error) {
			nodesMap := val.(map[string]*Node)
			delete(nodesMap, node.name)
			return nodesMap, nil
		})

		node.Close()
	}

	// Remove all nodes at once to avoid copying entire array multiple times.
	clstr.nodes.Update(func(val interface{}) (interface{}, error) {
		nodes := val.([]*Node)
		nlist := make([]*Node, 0, len(nodes))
		nlist = append(nlist, nodes...)
		for i, n := range nlist {
			for _, ntr := range nodesToRemove {
				if ntr.Equals(n) {
					nlist[i] = nil
				}
			}
		}

		newNodes := make([]*Node, 0, len(nlist))
		for i := range nlist {
			if nlist[i] != nil {
				newNodes = append(newNodes, nlist[i])
			}
		}

		return newNodes, nil
	})

}

// IsConnected returns true if cluster has nodes and is not already closed.
func (clstr *Cluster) IsConnected() bool {
	// Must copy array reference for copy on write semantics to work.
	nodeArray := clstr.GetNodes()
	return (len(nodeArray) > 0) && !clstr.closed.Get()
}

func (clstr *Cluster) getReadNode(partition *Partition, replica ReplicaPolicy) (*Node, error) {
	switch replica {
	case MASTER:
		return clstr.getMasterNode(partition)
	case MASTER_PROLES:
		return clstr.getMasterProleNode(partition)
	default:
		// includes case RANDOM:
		return clstr.GetRandomNode()
	}
}

func (clstr *Cluster) getMasterNode(partition *Partition) (*Node, error) {
	pmap := clstr.getPartitions()
	replicaArray := pmap[partition.Namespace]

	if replicaArray != nil {
		node := replicaArray[0][partition.PartitionId]
		if node != nil && node.IsActive() {
			return node, nil
		}
	}

	return clstr.GetRandomNode()
}

func (clstr *Cluster) getMasterProleNode(partition *Partition) (*Node, error) {
	pmap := clstr.getPartitions()
	replicaArray := pmap[partition.Namespace]

	if replicaArray != nil {
		for range replicaArray {
			index := int(atomic.AddUint64(&clstr.replicaIndex, 1) % uint64(len(replicaArray)))
			node := replicaArray[index][partition.PartitionId]
			if node != nil && node.IsActive() {
				return node, nil
			}
		}
	}

	return clstr.GetRandomNode()
}

// GetRandomNode returns a random node on the cluster
func (clstr *Cluster) GetRandomNode() (*Node, error) {
	// Must copy array reference for copy on write semantics to work.
	nodeArray := clstr.GetNodes()
	length := len(nodeArray)
	for i := 0; i < length; i++ {
		// Must handle concurrency with other non-tending goroutines, so nodeIndex is consistent.
		index := int(atomic.AddUint64(&clstr.nodeIndex, 1) % uint64(length))
		node := nodeArray[index]

		if node != nil && node.IsActive() {
			// Logger.Debug("Node `%s` is active. index=%d", node, index)
			return node, nil
		}
	}
	return nil, NewAerospikeError(INVALID_NODE_ERROR)
}

// GetNodes returns a list of all nodes in the cluster
func (clstr *Cluster) GetNodes() []*Node {
	// Must copy array reference for copy on write semantics to work.
	return clstr.nodes.Get().([]*Node)
}

// GetSeeds returns a list of all seed nodes in the cluster
func (clstr *Cluster) GetSeeds() []Host {
	res, _ := clstr.seeds.GetSyncedVia(func(val interface{}) (interface{}, error) {
		seeds := val.([]*Host)
		res := make([]Host, 0, len(seeds))
		for _, seed := range seeds {
			res = append(res, *seed)
		}

		return res, nil
	})

	return res.([]Host)
}

// GetAliases returns a list of all node aliases in the cluster
func (clstr *Cluster) GetAliases() map[Host]*Node {
	res, _ := clstr.aliases.GetSyncedVia(func(val interface{}) (interface{}, error) {
		aliases := val.(map[Host]*Node)
		res := make(map[Host]*Node, len(aliases))
		for h, n := range aliases {
			res[h] = n
		}

		return res, nil
	})

	return res.(map[Host]*Node)
}

// GetNodeByName finds a node by name and returns an
// error if the node is not found.
func (clstr *Cluster) GetNodeByName(nodeName string) (*Node, error) {
	node := clstr.findNodeByName(nodeName)

	if node == nil {
		return nil, NewAerospikeError(INVALID_NODE_ERROR)
	}
	return node, nil
}

func (clstr *Cluster) findNodeByName(nodeName string) *Node {
	// Must copy array reference for copy on write semantics to work.
	for _, node := range clstr.GetNodes() {
		if node.GetName() == nodeName {
			return node
		}
	}
	return nil
}

// Close closes all cached connections to the cluster nodes
// and stops the tend goroutine.
func (clstr *Cluster) Close() {
	if !clstr.closed.Get() {
		// send close signal to maintenance channel
		close(clstr.tendChannel)

		// wait until tend is over
		clstr.wgTend.Wait()
	}
}

// MigrationInProgress determines if any node in the cluster
// is participating in a data migration
func (clstr *Cluster) MigrationInProgress(timeout time.Duration) (res bool, err error) {
	if timeout <= 0 {
		timeout = _DEFAULT_TIMEOUT
	}

	done := make(chan bool, 1)

	go func() {
		// this function is guaranteed to return after _DEFAULT_TIMEOUT
		nodes := clstr.GetNodes()
		for _, node := range nodes {
			if node.IsActive() {
				if res, err = node.MigrationInProgress(); res || err != nil {
					done <- true
					return
				}
			}
		}

		res, err = false, nil
		done <- false
	}()

	dealine := time.After(timeout)
	for {
		select {
		case <-dealine:
			return false, NewAerospikeError(TIMEOUT)
		case <-done:
			return res, err
		}
	}
}

// WaitUntillMigrationIsFinished will block until all
// migration operations in the cluster all finished.
func (clstr *Cluster) WaitUntillMigrationIsFinished(timeout time.Duration) (err error) {
	if timeout <= 0 {
		timeout = _NO_TIMEOUT
	}
	done := make(chan error, 1)

	go func() {
		// this function is guaranteed to return after timeout
		// no go routines will be leaked
		for {
			if res, err := clstr.MigrationInProgress(timeout); err != nil || !res {
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

// Password returns the password that is currently used with the cluster.
func (clstr *Cluster) Password() (res []byte) {
	pass := clstr.password.Get()
	if pass != nil {
		return pass.([]byte)
	}
	return nil
}

func (clstr *Cluster) changePassword(user string, password string, hash []byte) {
	// change password ONLY if the user is the same
	if clstr.user == user {
		clstr.clientPolicy.Password = password
		clstr.password.Set(hash)
	}
}

// ClientPolicy returns the client policy that is currently used with the cluster.
func (clstr *Cluster) ClientPolicy() (res ClientPolicy) {
	return clstr.clientPolicy
}
