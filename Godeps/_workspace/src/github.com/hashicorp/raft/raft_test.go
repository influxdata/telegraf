package raft

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/hashicorp/go-msgpack/codec"
)

// MockFSM is an implementation of the FSM interface, and just stores
// the logs sequentially.
type MockFSM struct {
	sync.Mutex
	logs [][]byte
}

type MockSnapshot struct {
	logs     [][]byte
	maxIndex int
}

func (m *MockFSM) Apply(log *Log) interface{} {
	m.Lock()
	defer m.Unlock()
	m.logs = append(m.logs, log.Data)
	return len(m.logs)
}

func (m *MockFSM) Snapshot() (FSMSnapshot, error) {
	m.Lock()
	defer m.Unlock()
	return &MockSnapshot{m.logs, len(m.logs)}, nil
}

func (m *MockFSM) Restore(inp io.ReadCloser) error {
	m.Lock()
	defer m.Unlock()
	defer inp.Close()
	hd := codec.MsgpackHandle{}
	dec := codec.NewDecoder(inp, &hd)

	m.logs = nil
	return dec.Decode(&m.logs)
}

func (m *MockSnapshot) Persist(sink SnapshotSink) error {
	hd := codec.MsgpackHandle{}
	enc := codec.NewEncoder(sink, &hd)
	if err := enc.Encode(m.logs[:m.maxIndex]); err != nil {
		sink.Cancel()
		return err
	}
	sink.Close()
	return nil
}

func (m *MockSnapshot) Release() {
}

// Return configurations optimized for in-memory
func inmemConfig() *Config {
	conf := DefaultConfig()
	conf.HeartbeatTimeout = 50 * time.Millisecond
	conf.ElectionTimeout = 50 * time.Millisecond
	conf.LeaderLeaseTimeout = 50 * time.Millisecond
	conf.CommitTimeout = time.Millisecond
	return conf
}

type cluster struct {
	dirs   []string
	stores []*InmemStore
	fsms   []*MockFSM
	snaps  []*FileSnapshotStore
	trans  []*InmemTransport
	rafts  []*Raft
}

func (c *cluster) Merge(other *cluster) {
	c.dirs = append(c.dirs, other.dirs...)
	c.stores = append(c.stores, other.stores...)
	c.fsms = append(c.fsms, other.fsms...)
	c.snaps = append(c.snaps, other.snaps...)
	c.trans = append(c.trans, other.trans...)
	c.rafts = append(c.rafts, other.rafts...)
}

func (c *cluster) Close() {
	var futures []Future
	for _, r := range c.rafts {
		futures = append(futures, r.Shutdown())
	}

	// Wait for shutdown
	timer := time.AfterFunc(200*time.Millisecond, func() {
		panic("timed out waiting for shutdown")
	})

	for _, f := range futures {
		if err := f.Error(); err != nil {
			panic(fmt.Errorf("shutdown future err: %v", err))
		}
	}
	timer.Stop()

	for _, d := range c.dirs {
		os.RemoveAll(d)
	}
}

func (c *cluster) GetInState(s RaftState) []*Raft {
	in := make([]*Raft, 0, 1)
	for _, r := range c.rafts {
		if r.State() == s {
			in = append(in, r)
		}
	}
	return in
}

func (c *cluster) Leader() *Raft {
	timeout := time.AfterFunc(400*time.Millisecond, func() {
		panic("timeout waiting for leader")
	})
	defer timeout.Stop()

	for len(c.GetInState(Leader)) < 1 {
		time.Sleep(time.Millisecond)
	}
	leaders := c.GetInState(Leader)
	if len(leaders) != 1 {
		panic(fmt.Errorf("expected one leader: %v", leaders))
	}
	return leaders[0]
}

func (c *cluster) FullyConnect() {
	log.Printf("[WARN] Fully Connecting")
	for i, t1 := range c.trans {
		for j, t2 := range c.trans {
			if i != j {
				t1.Connect(t2.LocalAddr(), t2)
				t2.Connect(t1.LocalAddr(), t1)
			}
		}
	}
}

func (c *cluster) Disconnect(a string) {
	log.Printf("[WARN] Disconnecting %v", a)
	for _, t := range c.trans {
		if t.localAddr == a {
			t.DisconnectAll()
		} else {
			t.Disconnect(a)
		}
	}
}

func (c *cluster) EnsureLeader(t *testing.T, expect string) {
	limit := time.Now().Add(400 * time.Millisecond)
CHECK:
	for _, r := range c.rafts {
		leader := r.Leader()
		if expect == "" {
			if leader != "" {
				if time.Now().After(limit) {
					t.Fatalf("leader %v expected nil", leader)
				} else {
					goto WAIT
				}
			}
		} else {
			if leader == "" || leader != expect {
				if time.Now().After(limit) {
					t.Fatalf("leader %v expected %v", leader, expect)
				} else {
					goto WAIT
				}
			}
		}
	}

	return
WAIT:
	time.Sleep(10 * time.Millisecond)
	goto CHECK
}

func (c *cluster) EnsureSame(t *testing.T) {
	limit := time.Now().Add(400 * time.Millisecond)
	first := c.fsms[0]

CHECK:
	first.Lock()
	for i, fsm := range c.fsms {
		if i == 0 {
			continue
		}
		fsm.Lock()

		if len(first.logs) != len(fsm.logs) {
			fsm.Unlock()
			if time.Now().After(limit) {
				t.Fatalf("length mismatch: %d %d",
					len(first.logs), len(fsm.logs))
			} else {
				goto WAIT
			}
		}

		for idx := 0; idx < len(first.logs); idx++ {
			if bytes.Compare(first.logs[idx], fsm.logs[idx]) != 0 {
				fsm.Unlock()
				if time.Now().After(limit) {
					t.Fatalf("log mismatch at index %d", idx)
				} else {
					goto WAIT
				}
			}
		}
		fsm.Unlock()
	}

	first.Unlock()
	return

WAIT:
	first.Unlock()
	time.Sleep(20 * time.Millisecond)
	goto CHECK
}

func raftToPeerSet(r *Raft) map[string]struct{} {
	peers := make(map[string]struct{})
	peers[r.localAddr] = struct{}{}

	raftPeers, _ := r.peerStore.Peers()
	for _, p := range raftPeers {
		peers[p] = struct{}{}
	}
	return peers
}

func (c *cluster) EnsureSamePeers(t *testing.T) {
	limit := time.Now().Add(400 * time.Millisecond)
	peerSet := raftToPeerSet(c.rafts[0])

CHECK:
	for i, raft := range c.rafts {
		if i == 0 {
			continue
		}

		otherSet := raftToPeerSet(raft)
		if !reflect.DeepEqual(peerSet, otherSet) {
			if time.Now().After(limit) {
				t.Fatalf("peer mismatch: %v %v", peerSet, otherSet)
			} else {
				goto WAIT
			}
		}
	}
	return

WAIT:
	time.Sleep(20 * time.Millisecond)
	goto CHECK
}

func MakeCluster(n int, t *testing.T, conf *Config) *cluster {
	c := &cluster{}
	peers := make([]string, 0, n)

	// Setup the stores and transports
	for i := 0; i < n; i++ {
		dir, err := ioutil.TempDir("", "raft")
		if err != nil {
			t.Fatalf("err: %v ", err)
		}
		store := NewInmemStore()
		c.dirs = append(c.dirs, dir)
		c.stores = append(c.stores, store)
		c.fsms = append(c.fsms, &MockFSM{})

		dir2, snap := FileSnapTest(t)
		c.dirs = append(c.dirs, dir2)
		c.snaps = append(c.snaps, snap)

		addr, trans := NewInmemTransport()
		c.trans = append(c.trans, trans)
		peers = append(peers, addr)
	}

	// Wire the transports together
	c.FullyConnect()

	// Create all the rafts
	for i := 0; i < n; i++ {
		if conf == nil {
			conf = inmemConfig()
		}
		if n == 1 {
			conf.EnableSingleNode = true
		}

		logs := c.stores[i]
		store := c.stores[i]
		snap := c.snaps[i]
		trans := c.trans[i]
		peerStore := &StaticPeers{StaticPeers: peers}

		raft, err := NewRaft(conf, c.fsms[i], logs, store, snap, peerStore, trans)
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		c.rafts = append(c.rafts, raft)
	}

	return c
}

func MakeClusterNoPeers(n int, t *testing.T, conf *Config) *cluster {
	c := &cluster{}

	// Setup the stores and transports
	for i := 0; i < n; i++ {
		dir, err := ioutil.TempDir("", "raft")
		if err != nil {
			t.Fatalf("err: %v ", err)
		}
		store := NewInmemStore()
		c.dirs = append(c.dirs, dir)
		c.stores = append(c.stores, store)
		c.fsms = append(c.fsms, &MockFSM{})

		dir2, snap := FileSnapTest(t)
		c.dirs = append(c.dirs, dir2)
		c.snaps = append(c.snaps, snap)

		_, trans := NewInmemTransport()
		c.trans = append(c.trans, trans)
	}

	// Wire the transports together
	c.FullyConnect()

	// Create all the rafts
	for i := 0; i < n; i++ {
		if conf == nil {
			conf = inmemConfig()
		}

		logs := c.stores[i]
		store := c.stores[i]
		snap := c.snaps[i]
		trans := c.trans[i]
		peerStore := &StaticPeers{}

		raft, err := NewRaft(conf, c.fsms[i], logs, store, snap, peerStore, trans)
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		c.rafts = append(c.rafts, raft)
	}

	return c
}

func TestRaft_StartStop(t *testing.T) {
	c := MakeCluster(1, t, nil)
	c.Close()
}

func TestRaft_AfterShutdown(t *testing.T) {
	c := MakeCluster(1, t, nil)
	c.Close()
	raft := c.rafts[0]

	// Everything should fail now
	if f := raft.Apply(nil, 0); f.Error() != ErrRaftShutdown {
		t.Fatalf("should be shutdown: %v", f.Error())
	}
	if f := raft.AddPeer(NewInmemAddr()); f.Error() != ErrRaftShutdown {
		t.Fatalf("should be shutdown: %v", f.Error())
	}
	if f := raft.RemovePeer(NewInmemAddr()); f.Error() != ErrRaftShutdown {
		t.Fatalf("should be shutdown: %v", f.Error())
	}
	if f := raft.Snapshot(); f.Error() != ErrRaftShutdown {
		t.Fatalf("should be shutdown: %v", f.Error())
	}

	// Should be idempotent
	raft.Shutdown()
}

func TestRaft_SingleNode(t *testing.T) {
	conf := inmemConfig()
	c := MakeCluster(1, t, conf)
	defer c.Close()
	raft := c.rafts[0]

	// Watch leaderCh for change
	select {
	case v := <-raft.LeaderCh():
		if !v {
			t.Fatalf("should become leader")
		}
	case <-time.After(conf.HeartbeatTimeout * 3):
		t.Fatalf("timeout becoming leader")
	}

	// Should be leader
	if s := raft.State(); s != Leader {
		t.Fatalf("expected leader: %v", s)
	}

	// Should be able to apply
	future := raft.Apply([]byte("test"), time.Millisecond)
	if err := future.Error(); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Check the response
	if future.Response().(int) != 1 {
		t.Fatalf("bad response: %v", future.Response())
	}

	// Check the index
	if idx := future.Index(); idx == 0 {
		t.Fatalf("bad index: %d", idx)
	}

	// Check that it is applied to the FSM
	if len(c.fsms[0].logs) != 1 {
		t.Fatalf("did not apply to FSM!")
	}
}

func TestRaft_TripleNode(t *testing.T) {
	// Make the cluster
	c := MakeCluster(3, t, nil)
	defer c.Close()

	// Should be one leader
	leader := c.Leader()
	c.EnsureLeader(t, leader.localAddr)

	// Should be able to apply
	future := leader.Apply([]byte("test"), time.Millisecond)
	if err := future.Error(); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Wait for replication
	time.Sleep(30 * time.Millisecond)

	// Check that it is applied to the FSM
	for _, fsm := range c.fsms {
		fsm.Lock()
		num := len(fsm.logs)
		fsm.Unlock()
		if num != 1 {
			t.Fatalf("did not apply to FSM!")
		}
	}
}

func TestRaft_LeaderFail(t *testing.T) {
	// Make the cluster
	c := MakeCluster(3, t, nil)
	defer c.Close()

	// Should be one leader
	leader := c.Leader()

	// Should be able to apply
	future := leader.Apply([]byte("test"), time.Millisecond)
	if err := future.Error(); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Wait for replication
	time.Sleep(30 * time.Millisecond)

	// Disconnect the leader now
	log.Printf("[INFO] Disconnecting %v", leader)
	c.Disconnect(leader.localAddr)

	// Wait for new leader
	limit := time.Now().Add(200 * time.Millisecond)
	var newLead *Raft
	for time.Now().Before(limit) && newLead == nil {
		time.Sleep(10 * time.Millisecond)
		leaders := c.GetInState(Leader)
		if len(leaders) == 1 && leaders[0] != leader {
			newLead = leaders[0]
		}
	}
	if newLead == nil {
		t.Fatalf("expected new leader")
	}

	// Ensure the term is greater
	if newLead.getCurrentTerm() <= leader.getCurrentTerm() {
		t.Fatalf("expected newer term! %d %d", newLead.getCurrentTerm(), leader.getCurrentTerm())
	}

	// Apply should work not work on old leader
	future1 := leader.Apply([]byte("fail"), time.Millisecond)

	// Apply should work on newer leader
	future2 := newLead.Apply([]byte("apply"), time.Millisecond)

	// Future2 should work
	if err := future2.Error(); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Reconnect the networks
	log.Printf("[INFO] Reconnecting %v", leader)
	c.FullyConnect()

	// Future1 should fail
	if err := future1.Error(); err != ErrLeadershipLost && err != ErrNotLeader {
		t.Fatalf("err: %v", err)
	}

	// Wait for log replication
	c.EnsureSame(t)

	// Check two entries are applied to the FSM
	for _, fsm := range c.fsms {
		fsm.Lock()
		if len(fsm.logs) != 2 {
			t.Fatalf("did not apply both to FSM! %v", fsm.logs)
		}
		if bytes.Compare(fsm.logs[0], []byte("test")) != 0 {
			t.Fatalf("first entry should be 'test'")
		}
		if bytes.Compare(fsm.logs[1], []byte("apply")) != 0 {
			t.Fatalf("second entry should be 'apply'")
		}
		fsm.Unlock()
	}
}

func TestRaft_BehindFollower(t *testing.T) {
	// Make the cluster
	c := MakeCluster(3, t, nil)
	defer c.Close()

	// Disconnect one follower
	leader := c.Leader()
	followers := c.GetInState(Follower)
	behind := followers[0]
	c.Disconnect(behind.localAddr)

	// Commit a lot of things
	var future Future
	for i := 0; i < 100; i++ {
		future = leader.Apply([]byte(fmt.Sprintf("test%d", i)), 0)
	}

	// Wait for the last future to apply
	if err := future.Error(); err != nil {
		t.Fatalf("err: %v", err)
	} else {
		log.Printf("[INFO] Finished apply without behind follower")
	}

	// Check that we have a non zero last contact
	if behind.LastContact().IsZero() {
		t.Fatalf("expected previous contact")
	}

	// Reconnect the behind node
	c.FullyConnect()

	// Ensure all the logs are the same
	c.EnsureSame(t)

	// Ensure one leader
	leader = c.Leader()
	c.EnsureLeader(t, leader.localAddr)
}

func TestRaft_ApplyNonLeader(t *testing.T) {
	// Make the cluster
	c := MakeCluster(3, t, nil)
	defer c.Close()

	// Wait for a leader
	c.Leader()
	time.Sleep(10 * time.Millisecond)

	// Try to apply to them
	followers := c.GetInState(Follower)
	if len(followers) != 2 {
		t.Fatalf("Expected 2 followers")
	}
	follower := followers[0]

	// Try to apply
	future := follower.Apply([]byte("test"), time.Millisecond)

	if future.Error() != ErrNotLeader {
		t.Fatalf("should not apply on follower")
	}

	// Should be cached
	if future.Error() != ErrNotLeader {
		t.Fatalf("should not apply on follower")
	}
}

func TestRaft_ApplyConcurrent(t *testing.T) {
	// Make the cluster
	conf := inmemConfig()
	conf.HeartbeatTimeout = 80 * time.Millisecond
	conf.ElectionTimeout = 80 * time.Millisecond
	c := MakeCluster(3, t, conf)
	defer c.Close()

	// Wait for a leader
	leader := c.Leader()

	// Create a wait group
	var group sync.WaitGroup
	group.Add(100)

	applyF := func(i int) {
		defer group.Done()
		future := leader.Apply([]byte(fmt.Sprintf("test%d", i)), 0)
		if err := future.Error(); err != nil {
			t.Fatalf("err: %v", err)
		}
	}

	// Concurrently apply
	for i := 0; i < 100; i++ {
		go applyF(i)
	}

	// Wait to finish
	doneCh := make(chan struct{})
	go func() {
		group.Wait()
		close(doneCh)
	}()
	select {
	case <-doneCh:
	case <-time.After(time.Second):
		t.Fatalf("timeout")
	}

	// Check the FSMs
	c.EnsureSame(t)
}

func TestRaft_ApplyConcurrent_Timeout(t *testing.T) {
	// Make the cluster
	conf := inmemConfig()
	conf.HeartbeatTimeout = 80 * time.Millisecond
	conf.ElectionTimeout = 80 * time.Millisecond
	c := MakeCluster(1, t, conf)
	defer c.Close()

	// Wait for a leader
	leader := c.Leader()

	// Enough enqueues should cause at least one timeout...
	var didTimeout int32 = 0
	for i := 0; i < 200; i++ {
		go func(i int) {
			future := leader.Apply([]byte(fmt.Sprintf("test%d", i)), time.Microsecond)
			if future.Error() == ErrEnqueueTimeout {
				atomic.StoreInt32(&didTimeout, 1)
			}
		}(i)
	}

	// Wait
	time.Sleep(20 * time.Millisecond)

	// Some should have failed
	if atomic.LoadInt32(&didTimeout) == 0 {
		t.Fatalf("expected a timeout")
	}
}

func TestRaft_JoinNode(t *testing.T) {
	// Make a cluster
	c := MakeCluster(2, t, nil)
	defer c.Close()

	// Apply a log to this cluster to ensure it is 'newer'
	var future Future
	leader := c.Leader()
	future = leader.Apply([]byte("first"), 0)
	if err := future.Error(); err != nil {
		t.Fatalf("err: %v", err)
	} else {
		log.Printf("[INFO] Applied log")
	}

	// Make a new cluster of 1
	c1 := MakeCluster(1, t, nil)

	// Merge clusters
	c.Merge(c1)
	c.FullyConnect()

	// Wait until we have 2 leaders
	limit := time.Now().Add(200 * time.Millisecond)
	var leaders []*Raft
	for time.Now().Before(limit) && len(leaders) != 2 {
		time.Sleep(10 * time.Millisecond)
		leaders = c.GetInState(Leader)
	}
	if len(leaders) != 2 {
		t.Fatalf("expected two leader: %v", leaders)
	}

	// Join the new node in
	future = leader.AddPeer(c1.rafts[0].localAddr)
	if err := future.Error(); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Wait until we have 2 followers
	limit = time.Now().Add(200 * time.Millisecond)
	var followers []*Raft
	for time.Now().Before(limit) && len(followers) != 2 {
		time.Sleep(10 * time.Millisecond)
		followers = c.GetInState(Follower)
	}
	if len(followers) != 2 {
		t.Fatalf("expected two followers: %v", followers)
	}

	// Check the FSMs
	c.EnsureSame(t)

	// Check the peers
	c.EnsureSamePeers(t)

	// Ensure one leader
	leader = c.Leader()
	c.EnsureLeader(t, leader.localAddr)
}

func TestRaft_RemoveFollower(t *testing.T) {
	// Make a cluster
	c := MakeCluster(3, t, nil)
	defer c.Close()

	// Get the leader
	leader := c.Leader()

	// Wait until we have 2 followers
	limit := time.Now().Add(200 * time.Millisecond)
	var followers []*Raft
	for time.Now().Before(limit) && len(followers) != 2 {
		time.Sleep(10 * time.Millisecond)
		followers = c.GetInState(Follower)
	}
	if len(followers) != 2 {
		t.Fatalf("expected two followers: %v", followers)
	}

	// Remove a follower
	follower := followers[0]
	future := leader.RemovePeer(follower.localAddr)
	if err := future.Error(); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Wait a while
	time.Sleep(20 * time.Millisecond)

	// Other nodes should have fewer peers
	if peers, _ := leader.peerStore.Peers(); len(peers) != 2 {
		t.Fatalf("too many peers")
	}
	if peers, _ := followers[1].peerStore.Peers(); len(peers) != 2 {
		t.Fatalf("too many peers")
	}
}

func TestRaft_RemoveLeader(t *testing.T) {
	// Make a cluster
	c := MakeCluster(3, t, nil)
	defer c.Close()

	// Get the leader
	leader := c.Leader()

	// Wait until we have 2 followers
	limit := time.Now().Add(200 * time.Millisecond)
	var followers []*Raft
	for time.Now().Before(limit) && len(followers) != 2 {
		time.Sleep(10 * time.Millisecond)
		followers = c.GetInState(Follower)
	}
	if len(followers) != 2 {
		t.Fatalf("expected two followers: %v", followers)
	}

	// Remove the leader
	leader.RemovePeer(leader.localAddr)

	// Wait a while
	time.Sleep(20 * time.Millisecond)

	// Should have a new leader
	newLeader := c.Leader()

	// Wait a bit for log application
	time.Sleep(20 * time.Millisecond)

	// Other nodes should have fewer peers
	if peers, _ := newLeader.peerStore.Peers(); len(peers) != 2 {
		t.Fatalf("too many peers")
	}

	// Old leader should be shutdown
	if leader.State() != Shutdown {
		t.Fatalf("leader should be shutdown")
	}

	// Old leader should have no peers
	if peers, _ := leader.peerStore.Peers(); len(peers) != 1 {
		t.Fatalf("leader should have no peers")
	}
}

func TestRaft_RemoveLeader_NoShutdown(t *testing.T) {
	// Make a cluster
	conf := inmemConfig()
	conf.ShutdownOnRemove = false
	c := MakeCluster(3, t, conf)
	defer c.Close()

	// Get the leader
	leader := c.Leader()

	// Wait until we have 2 followers
	limit := time.Now().Add(200 * time.Millisecond)
	var followers []*Raft
	for time.Now().Before(limit) && len(followers) != 2 {
		time.Sleep(10 * time.Millisecond)
		followers = c.GetInState(Follower)
	}
	if len(followers) != 2 {
		t.Fatalf("expected two followers: %v", followers)
	}

	// Remove the leader
	leader.RemovePeer(leader.localAddr)

	// Wait a while
	time.Sleep(20 * time.Millisecond)

	// Should have a new leader
	newLeader := c.Leader()

	// Wait a bit for log application
	time.Sleep(20 * time.Millisecond)

	// Other nodes should have fewer peers
	if peers, _ := newLeader.peerStore.Peers(); len(peers) != 2 {
		t.Fatalf("too many peers")
	}

	// Old leader should be a follower
	if leader.State() != Follower {
		t.Fatalf("leader should be shutdown")
	}

	// Old leader should have no peers
	if peers, _ := leader.peerStore.Peers(); len(peers) != 1 {
		t.Fatalf("leader should have no peers")
	}
}

func TestRaft_RemoveLeader_SplitCluster(t *testing.T) {
	// Enable operation after a remove
	conf := inmemConfig()
	conf.EnableSingleNode = true
	conf.ShutdownOnRemove = false
	conf.DisableBootstrapAfterElect = false

	// Make a cluster
	c := MakeCluster(3, t, conf)
	defer c.Close()

	// Get the leader
	leader := c.Leader()

	// Remove the leader
	leader.RemovePeer(leader.localAddr)

	// Wait until we have 2 leaders
	limit := time.Now().Add(200 * time.Millisecond)
	var leaders []*Raft
	for time.Now().Before(limit) && len(leaders) != 2 {
		time.Sleep(10 * time.Millisecond)
		leaders = c.GetInState(Leader)
	}
	if len(leaders) != 2 {
		t.Fatalf("expected two leader: %v", leaders)
	}

	// Old leader should have no peers
	if len(leader.peers) != 0 {
		t.Fatalf("leader should have no peers")
	}
}

func TestRaft_AddKnownPeer(t *testing.T) {
	// Make a cluster
	c := MakeCluster(3, t, nil)
	defer c.Close()

	// Get the leader
	leader := c.Leader()
	followers := c.GetInState(Follower)

	// Add a follower
	future := leader.AddPeer(followers[0].localAddr)

	// Should be already added
	if err := future.Error(); err != ErrKnownPeer {
		t.Fatalf("err: %v", err)
	}
}

func TestRaft_RemoveUnknownPeer(t *testing.T) {
	// Make a cluster
	c := MakeCluster(3, t, nil)
	defer c.Close()

	// Get the leader
	leader := c.Leader()

	// Remove unknown
	future := leader.RemovePeer(NewInmemAddr())

	// Should be already added
	if err := future.Error(); err != ErrUnknownPeer {
		t.Fatalf("err: %v", err)
	}
}

func TestRaft_SnapshotRestore(t *testing.T) {
	// Make the cluster
	conf := inmemConfig()
	conf.TrailingLogs = 10
	c := MakeCluster(1, t, conf)
	defer c.Close()

	// Commit a lot of things
	leader := c.Leader()
	var future Future
	for i := 0; i < 100; i++ {
		future = leader.Apply([]byte(fmt.Sprintf("test%d", i)), 0)
	}

	// Wait for the last future to apply
	if err := future.Error(); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Take a snapshot
	snapFuture := leader.Snapshot()
	if err := snapFuture.Error(); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Check for snapshot
	if snaps, _ := leader.snapshots.List(); len(snaps) != 1 {
		t.Fatalf("should have a snapshot")
	}

	// Logs should be trimmed
	if idx, _ := leader.logs.FirstIndex(); idx != 92 {
		t.Fatalf("should trim logs to 92: %d", idx)
	}

	// Shutdown
	shutdown := leader.Shutdown()
	if err := shutdown.Error(); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Restart the Raft
	r := leader
	r, err := NewRaft(r.conf, r.fsm, r.logs, r.stable,
		r.snapshots, r.peerStore, r.trans)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	c.rafts[0] = r

	// We should have restored from the snapshot!
	if last := r.getLastApplied(); last != 101 {
		t.Fatalf("bad last: %v", last)
	}
}

func TestRaft_SnapshotRestore_PeerChange(t *testing.T) {
	// Make the cluster
	conf := inmemConfig()
	conf.TrailingLogs = 10
	c := MakeCluster(3, t, conf)
	defer c.Close()

	// Commit a lot of things
	leader := c.Leader()
	var future Future
	for i := 0; i < 100; i++ {
		future = leader.Apply([]byte(fmt.Sprintf("test%d", i)), 0)
	}

	// Wait for the last future to apply
	if err := future.Error(); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Take a snapshot
	snapFuture := leader.Snapshot()
	if err := snapFuture.Error(); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Shutdown
	shutdown := leader.Shutdown()
	if err := shutdown.Error(); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Make a separate cluster
	c2 := MakeClusterNoPeers(2, t, conf)
	defer c2.Close()

	// Kill the old cluster
	for _, sec := range c.rafts {
		if sec != leader {
			sec.Shutdown()
		}
	}

	// Change the peer addresses
	peers := []string{leader.trans.LocalAddr()}
	for _, sec := range c2.rafts {
		peers = append(peers, sec.trans.LocalAddr())
	}

	// Restart the Raft with new peers
	r := leader
	peerStore := &StaticPeers{StaticPeers: peers}
	r, err := NewRaft(r.conf, r.fsm, r.logs, r.stable,
		r.snapshots, peerStore, r.trans)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	c.rafts[0] = r
	c2.rafts = append(c2.rafts, r)
	c2.trans = append(c2.trans, r.trans.(*InmemTransport))
	c2.fsms = append(c2.fsms, r.fsm.(*MockFSM))
	c2.FullyConnect()

	// Wait a while
	time.Sleep(50 * time.Millisecond)

	// Ensure we elect a leader, and that we replicate
	// to our new followers
	c2.EnsureSame(t)

	// We should have restored from the snapshot!
	if last := r.getLastApplied(); last != 102 {
		t.Fatalf("bad last: %v", last)
	}
}

func TestRaft_AutoSnapshot(t *testing.T) {
	// Make the cluster
	conf := inmemConfig()
	conf.SnapshotInterval = 5 * time.Millisecond
	conf.SnapshotThreshold = 50
	conf.TrailingLogs = 10
	c := MakeCluster(1, t, conf)
	defer c.Close()

	// Commit a lot of things
	leader := c.Leader()
	var future Future
	for i := 0; i < 100; i++ {
		future = leader.Apply([]byte(fmt.Sprintf("test%d", i)), 0)
	}

	// Wait for the last future to apply
	if err := future.Error(); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Wait for a snapshot to happen
	time.Sleep(50 * time.Millisecond)

	// Check for snapshot
	if snaps, _ := leader.snapshots.List(); len(snaps) == 0 {
		t.Fatalf("should have a snapshot")
	}
}

func TestRaft_SendSnapshotFollower(t *testing.T) {
	// Make the cluster
	conf := inmemConfig()
	conf.TrailingLogs = 10
	c := MakeCluster(3, t, conf)
	defer c.Close()

	// Disconnect one follower
	followers := c.GetInState(Follower)
	behind := followers[0]
	c.Disconnect(behind.localAddr)

	// Commit a lot of things
	leader := c.Leader()
	var future Future
	for i := 0; i < 100; i++ {
		future = leader.Apply([]byte(fmt.Sprintf("test%d", i)), 0)
	}

	// Wait for the last future to apply
	if err := future.Error(); err != nil {
		t.Fatalf("err: %v", err)
	} else {
		log.Printf("[INFO] Finished apply without behind follower")
	}

	// Snapshot, this will truncate logs!
	for _, r := range c.rafts {
		future = r.Snapshot()
		if err := future.Error(); err != nil {
			t.Fatalf("err: %v", err)
		}
	}

	// Reconnect the behind node
	c.FullyConnect()

	// Ensure all the logs are the same
	c.EnsureSame(t)
}

func TestRaft_ReJoinFollower(t *testing.T) {
	// Enable operation after a remove
	conf := inmemConfig()
	conf.ShutdownOnRemove = false

	// Make a cluster
	c := MakeCluster(3, t, conf)
	defer c.Close()

	// Get the leader
	leader := c.Leader()

	// Wait until we have 2 followers
	limit := time.Now().Add(200 * time.Millisecond)
	var followers []*Raft
	for time.Now().Before(limit) && len(followers) != 2 {
		time.Sleep(10 * time.Millisecond)
		followers = c.GetInState(Follower)
	}
	if len(followers) != 2 {
		t.Fatalf("expected two followers: %v", followers)
	}

	// Remove a follower
	follower := followers[0]
	future := leader.RemovePeer(follower.localAddr)
	if err := future.Error(); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Wait a while
	time.Sleep(20 * time.Millisecond)

	// Other nodes should have fewer peers
	if peers, _ := leader.peerStore.Peers(); len(peers) != 2 {
		t.Fatalf("too many peers: %v", peers)
	}
	if peers, _ := followers[1].peerStore.Peers(); len(peers) != 2 {
		t.Fatalf("too many peers: %v", peers)
	}

	// Get the leader
	time.Sleep(20 * time.Millisecond)
	leader = c.Leader()

	// Rejoin. The follower will have a higher term than the leader,
	// this will cause the leader to step down, and a new round of elections
	// to take place. We should eventually re-stabilize.
	future = leader.AddPeer(follower.localAddr)
	if err := future.Error(); err != nil && err != ErrLeadershipLost {
		t.Fatalf("err: %v", err)
	}

	// Wait a while
	time.Sleep(40 * time.Millisecond)

	// Other nodes should have fewer peers
	if peers, _ := leader.peerStore.Peers(); len(peers) != 3 {
		t.Fatalf("missing peers: %v", peers)
	}
	if peers, _ := followers[1].peerStore.Peers(); len(peers) != 3 {
		t.Fatalf("missing peers: %v", peers)
	}

	// Should be a follower now
	if follower.State() != Follower {
		t.Fatalf("bad state: %v", follower.State())
	}
}

func TestRaft_LeaderLeaseExpire(t *testing.T) {
	// Make a cluster
	conf := inmemConfig()
	c := MakeCluster(2, t, conf)
	defer c.Close()

	// Get the leader
	leader := c.Leader()

	// Wait until we have a followers
	limit := time.Now().Add(200 * time.Millisecond)
	var followers []*Raft
	for time.Now().Before(limit) && len(followers) != 1 {
		time.Sleep(10 * time.Millisecond)
		followers = c.GetInState(Follower)
	}
	if len(followers) != 1 {
		t.Fatalf("expected a followers: %v", followers)
	}

	// Disconnect the follower now
	follower := followers[0]
	log.Printf("[INFO] Disconnecting %v", follower)
	c.Disconnect(follower.localAddr)

	// Watch the leaderCh
	select {
	case v := <-leader.LeaderCh():
		if v {
			t.Fatalf("should step down as leader")
		}
	case <-time.After(conf.LeaderLeaseTimeout * 2):
		t.Fatalf("timeout stepping down as leader")
	}

	// Should be no leaders
	if len(c.GetInState(Leader)) != 0 {
		t.Fatalf("expected step down")
	}

	// Verify no further contact
	last := follower.LastContact()
	time.Sleep(50 * time.Millisecond)

	// Check that last contact has not changed
	if last != follower.LastContact() {
		t.Fatalf("unexpected further contact")
	}

	// Ensure both have cleared their leader
	if l := leader.Leader(); l != "" {
		t.Fatalf("bad: %v", l)
	}
	if l := follower.Leader(); l != "" {
		t.Fatalf("bad: %v", l)
	}
}

func TestRaft_Barrier(t *testing.T) {
	// Make the cluster
	c := MakeCluster(3, t, nil)
	defer c.Close()

	// Get the leader
	leader := c.Leader()

	// Commit a lot of things
	for i := 0; i < 100; i++ {
		leader.Apply([]byte(fmt.Sprintf("test%d", i)), 0)
	}

	// Wait for a barrier complete
	barrier := leader.Barrier(0)

	// Wait for the barrier future to apply
	if err := barrier.Error(); err != nil {
		t.Fatalf("err: %v", err)
	}

	// Ensure all the logs are the same
	c.EnsureSame(t)
	if len(c.fsms[0].logs) != 100 {
		t.Fatalf("Bad log length")
	}
}

func TestRaft_VerifyLeader(t *testing.T) {
	// Make the cluster
	c := MakeCluster(3, t, nil)
	defer c.Close()

	// Get the leader
	leader := c.Leader()

	// Verify we are leader
	verify := leader.VerifyLeader()

	// Wait for the verify to apply
	if err := verify.Error(); err != nil {
		t.Fatalf("err: %v", err)
	}
}

func TestRaft_VerifyLeader_Single(t *testing.T) {
	// Make the cluster
	c := MakeCluster(1, t, nil)
	defer c.Close()

	// Get the leader
	leader := c.Leader()

	// Verify we are leader
	verify := leader.VerifyLeader()

	// Wait for the verify to apply
	if err := verify.Error(); err != nil {
		t.Fatalf("err: %v", err)
	}
}

func TestRaft_VerifyLeader_Fail(t *testing.T) {
	// Make a cluster
	conf := inmemConfig()
	c := MakeCluster(2, t, conf)
	defer c.Close()

	// Get the leader
	leader := c.Leader()

	// Wait until we have a followers
	limit := time.Now().Add(200 * time.Millisecond)
	var followers []*Raft
	for time.Now().Before(limit) && len(followers) != 1 {
		time.Sleep(10 * time.Millisecond)
		followers = c.GetInState(Follower)
	}
	if len(followers) != 1 {
		t.Fatalf("expected a followers: %v", followers)
	}

	// Force follower to different term
	follower := followers[0]
	follower.setCurrentTerm(follower.getCurrentTerm() + 1)

	// Verify we are leader
	verify := leader.VerifyLeader()

	// Wait for the leader to step down
	if err := verify.Error(); err != ErrNotLeader && err != ErrLeadershipLost {
		t.Fatalf("err: %v", err)
	}

	// Ensure the known leader is cleared
	if l := leader.Leader(); l != "" {
		t.Fatalf("bad: %v", l)
	}
}

func TestRaft_VerifyLeader_ParitalConnect(t *testing.T) {
	// Make a cluster
	conf := inmemConfig()
	c := MakeCluster(3, t, conf)
	defer c.Close()

	// Get the leader
	leader := c.Leader()

	// Wait until we have a followers
	limit := time.Now().Add(200 * time.Millisecond)
	var followers []*Raft
	for time.Now().Before(limit) && len(followers) != 2 {
		time.Sleep(10 * time.Millisecond)
		followers = c.GetInState(Follower)
	}
	if len(followers) != 2 {
		t.Fatalf("expected a followers: %v", followers)
	}

	// Force partial disconnect
	follower := followers[0]
	log.Printf("[INFO] Disconnecting %v", follower)
	c.Disconnect(follower.localAddr)

	// Verify we are leader
	verify := leader.VerifyLeader()

	// Wait for the leader to step down
	if err := verify.Error(); err != nil {
		t.Fatalf("err: %v", err)
	}
}

func TestRaft_SettingPeers(t *testing.T) {
	// Make the cluster
	c := MakeClusterNoPeers(3, t, nil)
	defer c.Close()

	peers := make([]string, 0)
	for _, v := range c.rafts {
		peers = append(peers, v.localAddr)
	}

	for _, v := range c.rafts {
		future := v.SetPeers(peers)
		if err := future.Error(); err != nil {
			t.Fatalf("error setting peers: %v", err)
		}
	}

	// Wait a while
	time.Sleep(20 * time.Millisecond)

	// Should have a new leader
	if leader := c.Leader(); leader == nil {
		t.Fatalf("no leader?")
	}
}
