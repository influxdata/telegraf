package gorethink

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/cenkalti/backoff"
	"github.com/hailocab/go-hostpool"
	"golang.org/x/net/context"
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
	seeds  []Host // Initial host nodes specified by user.
	hp     hostpool.HostPool
	nodes  map[string]*Node // Active nodes in cluster.
	closed bool

	nodeIndex int64
}

// NewCluster creates a new cluster by connecting to the given hosts.
func NewCluster(hosts []Host, opts *ConnectOpts) (*Cluster, error) {
	c := &Cluster{
		hp:    hostpool.NewEpsilonGreedy([]string{}, opts.HostDecayDuration, &hostpool.LinearEpsilonValueCalculator{}),
		seeds: hosts,
		opts:  opts,
	}

	// Attempt to connect to each host and discover any additional hosts if host
	// discovery is enabled
	if err := c.connectNodes(c.getSeeds()); err != nil {
		return nil, err
	}

	if !c.IsConnected() {
		return nil, ErrNoConnectionsStarted
	}

	if opts.DiscoverHosts {
		go c.discover()
	}

	return c, nil
}

// Query executes a ReQL query using the cluster to connect to the database
func (c *Cluster) Query(ctx context.Context, q Query) (cursor *Cursor, err error) {
	for i := 0; i < c.numRetries(); i++ {
		var node *Node
		var hpr hostpool.HostPoolResponse

		node, hpr, err = c.GetNextNode()
		if err != nil {
			return nil, err
		}

		cursor, err = node.Query(ctx, q)
		hpr.Mark(err)

		if !shouldRetryQuery(q, err) {
			break
		}
	}

	return cursor, err
}

// Exec executes a ReQL query using the cluster to connect to the database
func (c *Cluster) Exec(ctx context.Context, q Query) (err error) {
	for i := 0; i < c.numRetries(); i++ {
		var node *Node
		var hpr hostpool.HostPoolResponse

		node, hpr, err = c.GetNextNode()
		if err != nil {
			return err
		}

		err = node.Exec(ctx, q)
		hpr.Mark(err)

		if !shouldRetryQuery(q, err) {
			break
		}
	}

	return err
}

// Server returns the server name and server UUID being used by a connection.
func (c *Cluster) Server() (response ServerResponse, err error) {
	for i := 0; i < c.numRetries(); i++ {
		var node *Node
		var hpr hostpool.HostPoolResponse

		node, hpr, err = c.GetNextNode()
		if err != nil {
			return ServerResponse{}, err
		}

		response, err = node.Server()
		hpr.Mark(err)

		// This query should not fail so retry if any error is detected
		if err == nil {
			break
		}
	}

	return response, err
}

// SetInitialPoolCap sets the initial capacity of the connection pool.
func (c *Cluster) SetInitialPoolCap(n int) {
	for _, node := range c.GetNodes() {
		node.SetInitialPoolCap(n)
	}
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

	c.hp.Close()
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
			Log.Debugf("Error discovering hosts %s, waiting: %s", err, wait)
		})
	}
}

// listenForNodeChanges listens for changes to node status using change feeds.
// This function will block until the query fails
func (c *Cluster) listenForNodeChanges() error {
	// Start listening to changes from a random active node
	node, hpr, err := c.GetNextNode()
	if err != nil {
		return err
	}

	q, err := newQuery(
		DB("rethinkdb").Table("server_status").Changes(),
		map[string]interface{}{},
		c.opts,
	)
	if err != nil {
		return fmt.Errorf("Error building query: %s", err)
	}

	cursor, err := node.Query(context.Background(), q) // no need for timeout due to Changes()
	if err != nil {
		hpr.Mark(err)
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

						Log.WithFields(logrus.Fields{
							"id":   node.ID,
							"host": node.Host.String(),
						}).Debug("Connected to node")
					}
				}

				return err
			}, b)
		}
	}

	err = cursor.Err()
	hpr.Mark(err)
	return err
}

func (c *Cluster) connectNodes(hosts []Host) error {
	// Add existing nodes to map
	nodeSet := map[string]*Node{}
	for _, node := range c.GetNodes() {
		nodeSet[node.ID] = node
	}

	var attemptErr error

	// Attempt to connect to each seed host
	for _, host := range hosts {
		conn, err := NewConnection(host.String(), c.opts)
		if err != nil {
			attemptErr = err
			Log.Warnf("Error creating connection: %s", err.Error())
			continue
		}
		defer conn.Close()

		if c.opts.DiscoverHosts {
			q, err := newQuery(
				DB("rethinkdb").Table("server_status"),
				map[string]interface{}{},
				c.opts,
			)
			if err != nil {
				Log.Warnf("Error building query: %s", err)
				continue
			}

			_, cursor, err := conn.Query(nil, q) // nil = connection opts' timeout
			if err != nil {
				attemptErr = err
				Log.Warnf("Error fetching cluster status: %s", err)
				continue
			}

			var results []nodeStatus
			err = cursor.All(&results)
			if err != nil {
				attemptErr = err
				continue
			}

			for _, result := range results {
				node, err := c.connectNodeWithStatus(result)
				if err == nil {
					if _, ok := nodeSet[node.ID]; !ok {
						Log.WithFields(logrus.Fields{
							"id":   node.ID,
							"host": node.Host.String(),
						}).Debug("Connected to node")
						nodeSet[node.ID] = node
					}
				} else {
					attemptErr = err
					Log.Warnf("Error connecting to node: %s", err)
				}
			}
		} else {
			svrRsp, err := conn.Server()
			if err != nil {
				attemptErr = err
				Log.Warnf("Error fetching server ID: %s", err)
				continue
			}

			node, err := c.connectNode(svrRsp.ID, []Host{host})
			if err == nil {
				if _, ok := nodeSet[node.ID]; !ok {
					Log.WithFields(logrus.Fields{
						"id":   node.ID,
						"host": node.Host.String(),
					}).Debug("Connected to node")

					nodeSet[node.ID] = node
				}
			} else {
				attemptErr = err
				Log.Warnf("Error connecting to node: %s", err)
			}
		}
	}

	// If no nodes were contactable then return the last error, this does not
	// include driver errors such as if there was an issue building the
	// query
	if len(nodeSet) == 0 {
		return attemptErr
	}

	nodes := []*Node{}
	for _, node := range nodeSet {
		nodes = append(nodes, node)
	}
	c.setNodes(nodes)

	return nil
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

// GetNextNode returns a random node on the cluster
func (c *Cluster) GetNextNode() (*Node, hostpool.HostPoolResponse, error) {
	if !c.IsConnected() {
		return nil, nil, ErrNoConnections
	}
	c.mu.RLock()
	defer c.mu.RUnlock()

	nodes := c.nodes
	hpr := c.hp.Get()
	if n, ok := nodes[hpr.Host()]; ok {
		if !n.Closed() {
			return n, hpr, nil
		}
	}

	return nil, nil, ErrNoConnections
}

// GetNodes returns a list of all nodes in the cluster
func (c *Cluster) GetNodes() []*Node {
	c.mu.RLock()
	nodes := make([]*Node, 0, len(c.nodes))
	for _, n := range c.nodes {
		nodes = append(nodes, n)
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
	c.mu.RLock()
	nodes := append(c.GetNodes(), node)
	c.mu.RUnlock()

	c.setNodes(nodes)
}

func (c *Cluster) addNodes(nodesToAdd []*Node) {
	c.mu.RLock()
	nodes := append(c.GetNodes(), nodesToAdd...)
	c.mu.RUnlock()

	c.setNodes(nodes)
}

func (c *Cluster) setNodes(nodes []*Node) {
	nodesMap := make(map[string]*Node, len(nodes))
	hosts := make([]string, len(nodes))
	for i, node := range nodes {
		host := node.Host.String()

		nodesMap[host] = node
		hosts[i] = host
	}

	c.mu.Lock()
	c.nodes = nodesMap
	c.hp.SetHosts(hosts)
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

func (c *Cluster) numRetries() int {
	if n := c.opts.NumRetries; n > 0 {
		return n
	}

	return 3
}
