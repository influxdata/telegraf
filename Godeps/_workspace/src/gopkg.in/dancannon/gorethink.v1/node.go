package gorethink

import (
	"sync"
	"sync/atomic"
	"time"

	p "github.com/dancannon/gorethink/ql2"
)

const (
	maxNodeHealth = 100
)

// Node represents a database server in the cluster
type Node struct {
	ID      string
	Host    Host
	aliases []Host

	cluster         *Cluster
	pool            *Pool
	refreshDoneChan chan struct{}

	mu     sync.RWMutex
	closed bool
	health int64
}

func newNode(id string, aliases []Host, cluster *Cluster, pool *Pool) *Node {
	node := &Node{
		ID:              id,
		Host:            aliases[0],
		aliases:         aliases,
		cluster:         cluster,
		pool:            pool,
		health:          maxNodeHealth,
		refreshDoneChan: make(chan struct{}),
	}
	// Start node refresh loop
	refreshInterval := cluster.opts.NodeRefreshInterval
	if refreshInterval <= 0 {
		// Default to refresh every 30 seconds
		refreshInterval = time.Second * 30
	}

	go func() {
		refreshTicker := time.NewTicker(refreshInterval)
		for {
			select {
			case <-refreshTicker.C:
				node.Refresh()
			case <-node.refreshDoneChan:
				return
			}
		}
	}()

	return node
}

// Closed returns true if the node is closed
func (n *Node) Closed() bool {
	n.mu.RLock()
	defer n.mu.RUnlock()

	return n.closed
}

// Close closes the session
func (n *Node) Close(optArgs ...CloseOpts) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.closed {
		return nil
	}

	if len(optArgs) >= 1 {
		if optArgs[0].NoReplyWait {
			n.NoReplyWait()
		}
	}

	n.refreshDoneChan <- struct{}{}
	if n.pool != nil {
		n.pool.Close()
	}
	n.pool = nil
	n.closed = true

	return nil
}

// SetMaxIdleConns sets the maximum number of connections in the idle
// connection pool.
func (n *Node) SetMaxIdleConns(idleConns int) {
	n.pool.SetMaxIdleConns(idleConns)
}

// SetMaxOpenConns sets the maximum number of open connections to the database.
func (n *Node) SetMaxOpenConns(openConns int) {
	n.pool.SetMaxOpenConns(openConns)
}

// NoReplyWait ensures that previous queries with the noreply flag have been
// processed by the server. Note that this guarantee only applies to queries
// run on the given connection
func (n *Node) NoReplyWait() error {
	return n.pool.Exec(Query{
		Type: p.Query_NOREPLY_WAIT,
	})
}

// Query executes a ReQL query using this nodes connection pool.
func (n *Node) Query(q Query) (cursor *Cursor, err error) {
	if n.Closed() {
		return nil, ErrInvalidNode
	}

	cursor, err = n.pool.Query(q)
	if err != nil {
		n.DecrementHealth()
	}

	return cursor, err
}

// Exec executes a ReQL query using this nodes connection pool.
func (n *Node) Exec(q Query) (err error) {
	if n.Closed() {
		return ErrInvalidNode
	}

	err = n.pool.Exec(q)
	if err != nil {
		n.DecrementHealth()
	}

	return err
}

// Refresh attempts to connect to the node and check that it is still connected
// to the cluster.
//
// If an error occurred or the node is no longer connected then
// the nodes health is decrease, if there were no issues then the node is marked
// as being healthy.
func (n *Node) Refresh() {
	if n.cluster.opts.DiscoverHosts {
		// If host discovery is enabled then check the servers status
		cursor, err := n.pool.Query(newQuery(
			DB("rethinkdb").Table("server_status").Get(n.ID),
			map[string]interface{}{},
			n.cluster.opts,
		))
		if err != nil {
			n.DecrementHealth()
			return
		}
		defer cursor.Close()

		var status nodeStatus
		err = cursor.One(&status)
		if err != nil {
			return
		}

		if status.Status != "connected" {
			n.DecrementHealth()
			return
		}
	} else {
		// If host discovery is disabled just execute a simple ping query
		cursor, err := n.pool.Query(newQuery(
			Expr("OK"),
			map[string]interface{}{},
			n.cluster.opts,
		))
		if err != nil {
			n.DecrementHealth()
			return
		}
		defer cursor.Close()

		var status string
		err = cursor.One(&status)
		if err != nil {
			return
		}

		if status != "OK" {
			n.DecrementHealth()
			return
		}
	}

	// If status check was successful reset health
	n.ResetHealth()
}

// DecrementHealth decreases the nodes health by 1 (the nodes health starts at maxNodeHealth)
func (n *Node) DecrementHealth() {
	atomic.AddInt64(&n.health, -1)
}

// ResetHealth sets the nodes health back to maxNodeHealth (fully healthy)
func (n *Node) ResetHealth() {
	atomic.StoreInt64(&n.health, maxNodeHealth)
}

// IsHealthy checks the nodes health by ensuring that the health counter is above 0.
func (n *Node) IsHealthy() bool {
	return n.health > 0
}

type nodeStatus struct {
	ID      string `gorethink:"id"`
	Name    string `gorethink:"name"`
	Status  string `gorethink:"status"`
	Network struct {
		Hostname           string `gorethink:"hostname"`
		ClusterPort        int64  `gorethink:"cluster_port"`
		ReqlPort           int64  `gorethink:"reql_port"`
		CanonicalAddresses []struct {
			Host string `gorethink:"host"`
			Port int64  `gorethink:"port"`
		} `gorethink:"canonical_addresses"`
	} `gorethink:"network"`
}
