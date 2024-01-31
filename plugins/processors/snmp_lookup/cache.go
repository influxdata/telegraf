package snmp_lookup

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/alitto/pond"
	"github.com/hashicorp/golang-lru/v2/expirable"
)

var ErrNotYetAvailable = errors.New("data not yet available")

type store struct {
	cache        *expirable.LRU[string, *tagMap]
	pool         *pond.WorkerPool
	inflight     sync.Map
	backlog      map[string]time.Time
	backlogTimer *time.Timer
	notify       func(string, *tagMap)
	update       func(string) *tagMap
	sync.Mutex
}

func newStore(size int, ttl time.Duration, workers int) *store {
	return &store{
		cache:   expirable.NewLRU[string, *tagMap](size, nil, ttl),
		pool:    pond.New(workers, 0, pond.MinWorkers(workers/2+1)),
		backlog: make(map[string]time.Time),
	}
}

func (s *store) addBacklog(agent string, earliest time.Time) {
	s.Lock()
	defer s.Unlock()
	fmt.Printf("  - adding backlog for agent %s\n", agent)
	t, found := s.backlog[agent]
	if !found || t.After(earliest) {
		s.backlog[agent] = earliest
		s.refreshTimer()
	}
}

func (s *store) removeBacklog(agent string) {
	s.Lock()
	defer s.Unlock()
	fmt.Printf("  - removing backlog for agent %s\n", agent)
	delete(s.backlog, agent)
	s.refreshTimer()
}

func (s *store) refreshTimer() {
	fmt.Println("  - refreshing timer")

	if s.backlogTimer != nil {
		s.backlogTimer.Stop()
	}
	if len(s.backlog) == 0 {
		return
	}
	var agent string
	var earliest time.Time
	for k, t := range s.backlog {
		if agent == "" || t.Before(earliest) {
			agent = k
			earliest = t
		}
	}
	s.backlogTimer = time.AfterFunc(time.Until(earliest), func() { s.enqueue(agent) })
}

func (s *store) enqueue(agent string) {
	fmt.Printf("  - enqueuing agent %s\n", agent)
	s.pool.Submit(func() {
		if _, inflight := s.inflight.LoadOrStore(agent, true); inflight {
			fmt.Println("  -> already in-flight...")
			return
		}
		tags := s.update(agent)
		fmt.Printf("  - received update for agent %s: %v\n", agent, tags)
		s.cache.Add(agent, tags)
		s.removeBacklog(agent)
		if s.notify != nil {
			fmt.Printf("  - sending notification %s\n", agent)
			s.notify(agent, tags)
		}
		s.inflight.Delete(agent)
	})
}

func (s *store) lookup(agent string, index string) (map[string]string, error) {
	fmt.Println("looking up ", agent, " index", index)
	entry, cached := s.cache.Peek(agent)
	if !cached {
		fmt.Println("  * not cached")
		// There is no cache at all, so we need to enqueue an update.
		s.enqueue(agent)
		return nil, ErrNotYetAvailable
	}

	value, found := entry.rows[index]
	if !found {
		// The index does not exist, therefore we need to update the
		// agent as it maybe appeared in the meantime
		if time.Since(entry.created) > minTimeBetweenUpdates {
			fmt.Println("  * not cached pause passed")
			// The minimum time between updates has passed so we are good to
			// directly update the cache.
			s.enqueue(agent)
			return nil, ErrNotYetAvailable
		}

		fmt.Println("  * not cached deferring")
		// The minimum time between updates has not yet passed so we
		// need to defer the agent update to later.
		s.addBacklog(agent, entry.created.Add(minTimeBetweenUpdates))
		return nil, ErrNotYetAvailable
	}
	fmt.Println("  => found", value)

	return value, nil
}

func (s *store) destroy() {
	s.pool.StopAndWait()
}

func (s *store) purge() {
	s.Lock()
	defer s.Unlock()
	s.cache.Purge()
}
