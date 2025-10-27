package heartbeat

import (
	"sync"
	"time"
)

type statistics struct {
	metrics          uint64
	logErrors        uint64
	logWarnings      uint64
	lastUpdate       time.Time
	lastUpdateFailed bool

	sync.RWMutex
}

func (s *statistics) snapshot() *statistics {
	s.RLock()
	defer s.RUnlock()

	return &statistics{
		metrics:          s.metrics,
		logErrors:        s.logErrors,
		logWarnings:      s.logWarnings,
		lastUpdate:       s.lastUpdate,
		lastUpdateFailed: s.lastUpdateFailed,
	}
}

func (s *statistics) remove(snap *statistics, ts time.Time) {
	s.Lock()
	defer s.Unlock()

	s.metrics -= snap.metrics
	s.logErrors -= snap.logErrors
	s.logWarnings -= snap.logWarnings
	s.lastUpdate = ts
	s.lastUpdateFailed = false
}

func (s *statistics) variables() map[string]interface{} {
	s.RLock()
	defer s.RUnlock()

	vars := map[string]interface{}{
		"metrics":      s.metrics,
		"log_errors":   s.logErrors,
		"log_warnings": s.logWarnings,
		"last_update":  s.lastUpdate,
	}

	return vars
}
