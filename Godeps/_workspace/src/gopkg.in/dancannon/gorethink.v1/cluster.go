package gorethink

import (
	"fmt"
	"math"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/cenkalti/backoff"
)

// A Cluster represents a connection to a RethinkDB cluster, a cluster is created
// by the Session and should rarely be created manually.
//
// The cluster keeps track of all nodes in the cluster and if requested can listen
// for cluster changes and start tracking a new node if one appears. Currently
// nodes are removed from the pool if they become unhealthy (100 failed queries).
// This should hopefully soon be replaced by a backoff system.
type Cluster struct {
	opts *ConnectOpts

	mu     sync.RWMutex
	seeds  []Host  // Initial host nodes specified by user.
	nodes  []*Node // Active nodes in cluster.
	closed bool

	nodeIndex int64
}

// NewCluster creates a new cluster by connecting to the given hosts.
func NewCluster(hosts []Host, opts *ConnectOpts) (*Cluster, error) {
	c := &Cluster{
		seeds: hosts,
		opts:  opts,
	}

	//Check that hosts in the ClusterConfig is not empty
	c.connectNodes(c.getSeeds())
	if !c.IsConnected() {
		return nil, ErrNoConnectionsStarted
	}

	if opts.DiscoverHosts {
		go c.discover()
	}

	return c, nil
}

// Query executes a ReQL query using the cluster to connect to the database
func (c *Cluster) Query(q Query) (cursor *Cursor, err error) {
	node, err := c.GetRandomNode()
	if err != nil {
		return nil, err
	}

	return node.Query(q)
}

// Exec executes a ReQL query using the cluster to connect to the database
func (c *Cluster) Exec(q Query) (err error) {
	node, err := c.GetRandomNode()
	if err != nil {
		return err
	}

	return node.Exec(q)
}

// SetMaxIdleConns sets the maximum number of connections in the idle
// connection pool.
func (c *Cluster) SetMaxIdleConns(n int) {
	for _, node := range c.GetNodes() {
		node.SetMaxIdleConns(n)
	}
}

// SetMaxOpenConns sets the maximum number of open connections to the database.
func (c *Cluster) SetMaxOpenConns(n int) {
	for _, node := range c.GetNodes() {
		node.SetMaxOpenConns(n)
	}
}

// Close closes the cluster
func (c *Cluster) Close(optArgs ...CloseOpts) error {
	if c.closed {
		return nil
	}

	for _, node := range c.GetNodes() {
		err := node.Close(optArgs...)
		if err != nil {
			return err
		}
	}

	c.closed = true

	return nil
}

// discover attempts to find new nodes in the cluster using the current nodes
func (c *Cluster) discover() {
	// Keep retrying with exponential backoff.
	b := backoff.NewExponentialBackOff()
	// Never finish retrying (max interval is still 60s)
	b.MaxElapsedTime = 0

	// Keep trying to discover new nodes
	for {
		backoff.RetryNotify(func() error {
			// If no hosts try seeding nodes
			if len(c.GetNodes()) == 0 {
				c.connectNodes(c.getSeeds())
			}

			return c.listenForNodeChanges()
		}, b, func(err error, wait time.Duration) {
			log.Debugf("Error discovering hosts %s, waiting %s", err, wait)
		})
	}
}

// listenForNodeChanges listens for changes to node status using change feeds.
// This function will block until the query fails
func (c *Cluster) listenForNodeChanges() error {
	// Start listening to changes from a random active node
	node, err := c.GetRandomNode()
	if err != nil {
		return err
	}

	cursor, err := node.Query(newQuery(
		DB("rethinkdb").Table("server_status").Changes(),
		map[string]interface{}{},
		c.opts,
	))
	if err != nil {
		return err
	}

	// Keep reading node status updates from changefeed
	var result struct {
		NewVal nodeStatus `gorethink:"new_val"`
		OldVal nodeStatus `gorethink:"old_val"`
	}
	for cursor.Next(&result) {
		addr := fmt.Sprintf("%s:%d", result.NewVal.Network.Hostname, result.NewVal.Network.ReqlPort)
		addr = strings.ToLower(addr)

		switch result.NewVal.Status {
		case "connected":
			// Connect to node using exponential backoff (give up after waiting 5s)
			// to give the node time to start-up.
			b := backoff.NewExponentialBackOff()
			b.MaxElapsedTime = time.Second * 5

			backoff.Retry(func() error {
				node, err := c.connectNodeWithStatus(result.NewVal)
				if err == nil {
					if !c.nodeExists(node) {
						c.addNode(node)

						log.WithFields(logrus.Fields{
							"id":   node.ID,
							"host": node.Host.String(),
						}).Debug("Connected to node")
					}
				}

				return err
			}, b)
		}
	}

	return cursor.Err()
}

func (c *Cluster) connectNodes(hosts []Host) {
	// Add existing nodes to map
	nodeSet := map[string]*Node{}
	for _, node := range c.GetNodes() {
		nodeSet[node.ID] = node
	}

	// Attempt to connect to each seed host
	for _, host := range hosts {
		conn, err := NewConnection(host.String(), c.opts)
		if err != nil {
			log.Warnf("Error creating connection %s", err.Error())
			continue
		}
		defer conn.Close()

		_, cursor, err := conn.Query(newQuery(
			DB("rethinkdb").Table("server_status"),
			map[string]interface{}{},
			c.opts,
		))
		if err != nil {
			log.Warnf("Error fetching cluster status %s", err)
			continue
		}

		if c.opts.DiscoverHosts {
			var results []nodeStatus
			err = cursor.All(&results)
			if err != nil {
				continue
			}

			for _, result := range results {
				node, err := c.connectNodeWithStatus(result)
				if err == nil {
					if _, ok := nodeSet[node.ID]; !ok {
						log.WithFields(logrus.Fields{
							"id":   node.ID,
							"host": node.Host.String(),
						}).Debug("Connected to node")
						nodeSet[node.ID] = node
					}
				}
			}
		} else {
			node, err := c.connectNode(host.String(), []Host{host})
			if err == nil {
				if _, ok := nodeSet[node.ID]; !ok {
					log.WithFields(logrus.Fields{
						"id":   node.ID,
						"host": node.Host.String(),
					}).Debug("Connected to node")
					nodeSet[node.ID] = node
				}
			}
		}
	}

	nodes := []*Node{}
	for _, node := range nodeSet {
		nodes = append(nodes, node)
	}

	c.setNodes(nodes)
}

func (c *Cluster) connectNodeWithStatus(s nodeStatus) (*Node, error) {
	aliases := make([]Host, len(s.Network.CanonicalAddresses))
	for i, aliasAddress := range s.Network.CanonicalAddresses {
		aliases[i] = NewHost(aliasAddress.Host, int(s.Network.ReqlPort))
	}

	return c.connectNode(s.ID, aliases)
}

func (c *Cluster) connectNode(id string, aliases []Host) (*Node, error) {
	var pool *Pool
	var err error

	for len(aliases) > 0 {
		pool, err = NewPool(aliases[0], c.opts)
		if err != nil {
			aliases = aliases[1:]
			continue
		}

		err = pool.Ping()
		if err != nil {
			aliases = aliases[1:]
			continue
		}

		// Ping successful so break out of loop
		break
	}

	if err != nil {
		return nil, err
	}
	if len(aliases) == 0 {
		return nil, ErrInvalidNode
	}

	return newNode(id, aliases, c, pool), nil
}

// IsConnected returns true if cluster has nodes and is not already closed.
func (c *Cluster) IsConnected() bool {
	c.mu.RLock()
	closed := c.closed
	c.mu.RUnlock()

	return (len(c.GetNodes()) > 0) && !closed
}

// AddSeeds adds new seed hosts to the cluster.
func (c *Cluster) AddSeeds(hosts []Host) {
	c.mu.Lock()
	c.seeds = append(c.seeds, hosts...)
	c.mu.Unlock()
}

func (c *Cluster) getSeeds() []Host {
	c.mu.RLock()
	seeds := c.seeds
	c.mu.RUnlock()

	return seeds
}

// GetRandomNode returns a random node on the cluster
// TODO(dancannon) replace with hostpool
func (c *Cluster) GetRandomNode() (*Node, error) {
	if !c.IsConnected() {
		return nil, ErrClusterClosed
	}
	// Must copy array reference for copy on write semantics to work.
	nodeArray := c.GetNodes()
	length := len(nodeArray)
	for i := 0; i < length; i++ {
		// Must handle concurrency with other non-tending goroutines, so nodeIndex is consistent.
		index := int(math.Abs(float64(c.nextNodeIndex() % int64(length))))
		node := nodeArray[index]

		if !node.Closed() && node.IsHealthy() {
			return node, nil
		}
	}
	return nil, ErrNoConnections
}

// GetNodes returns a list of all nodes in the cluster
func (c *Cluster) GetNodes() []*Node {
	c.mu.RLock()
	nodes := c.nodes
	c.mu.RUnlock()

	return nodes
}

// GetHealthyNodes returns a list of all healthy nodes in the cluster
func (c *Cluster) GetHealthyNodes() []*Node {
	c.mu.RLock()
	nodes := []*Node{}
	for _, node := range c.nodes {
		if node.IsHealthy() {
			nodes = append(nodes, node)
		}
	}
	c.mu.RUnlock()

	return nodes
}

func (c *Cluster) nodeExists(search *Node) bool {
	for _, node := range c.GetNodes() {
		if node.ID == search.ID {
			return true
		}
	}
	return false
}

func (c *Cluster) addNode(node *Node) {
	c.mu.Lock()
	c.nodes = append(c.nodes, node)
	c.mu.Unlock()
}

func (c *Cluster) addNodes(nodesToAdd []*Node) {
	c.mu.Lock()
	c.nodes = append(c.nodes, nodesToAdd...)
	c.mu.Unlock()
}

func (c *Cluster) setNodes(nodes []*Node) {
	c.mu.Lock()
	c.nodes = nodes
	c.mu.Unlock()
}

func (c *Cluster) removeNode(nodeID string) {
	nodes := c.GetNodes()
	nodeArray := make([]*Node, len(nodes)-1)
	count := 0

	// Add nodes that are not in remove list.
	for _, n := range nodes {
		if n.ID != nodeID {
			nodeArray[count] = n
			count++
		}
	}

	// Do sanity check to make sure assumptions are correct.
	if count < len(nodeArray) {
		// Resize array.
		nodeArray2 := make([]*Node, count)
		copy(nodeArray2, nodeArray)
		nodeArray = nodeArray2
	}

	c.setNodes(nodeArray)
}

func (c *Cluster) nextNodeIndex() int64 {
	return atomic.AddInt64(&c.nodeIndex, 1)
}
