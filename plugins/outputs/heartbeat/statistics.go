package heartbeat

import (
	"maps"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf/selfstat"
)

type statistics struct {
	metrics          uint64
	logErrors        uint64
	logWarnings      uint64
	lastUpdate       time.Time
	lastUpdateFailed bool

	lastAgent      map[string]interface{}
	lastInputs     map[string][]map[string]interface{}
	lastOutputs    map[string][]map[string]interface{}
	currentAgent   map[string]interface{}
	currentInputs  map[string][]map[string]interface{}
	currentOutputs map[string][]map[string]interface{}

	sync.RWMutex
}

func (s *statistics) snapshot() *statistics {
	s.RLock()
	defer s.RUnlock()

	out := &statistics{
		metrics:          s.metrics,
		logErrors:        s.logErrors,
		logWarnings:      s.logWarnings,
		lastUpdate:       s.lastUpdate,
		lastUpdateFailed: s.lastUpdateFailed,

		lastAgent:      s.currentAgent,
		lastInputs:     s.currentInputs,
		lastOutputs:    s.currentOutputs,
		currentAgent:   make(map[string]interface{}),
		currentInputs:  make(map[string][]map[string]interface{}),
		currentOutputs: make(map[string][]map[string]interface{}),
	}

	// Add internal statistics
	for _, m := range selfstat.Metrics() {
		statsType := strings.TrimPrefix(m.Name(), "internal_")

		switch statsType {
		case "gather":
			tags := m.Tags()

			// Create the entry
			entry := m.Fields()
			var name string
			for k, v := range tags {
				switch k {
				case "input":
					name = v
				case "_id":
					entry["id"] = v
				default:
					entry[k] = v
				}
			}
			out.currentInputs[name] = append(out.currentInputs[name], entry)
		case "write":
			tags := m.Tags()

			// Create the entry
			entry := m.Fields()
			var name string
			for k, v := range tags {
				switch k {
				case "output":
					name = v
				case "_id":
					entry["id"] = v
				default:
					entry[k] = v
				}
			}
			entry["buffer_fullness"] = float64(entry["buffer_size"].(int64)) / float64(entry["buffer_limit"].(int64))
			out.currentOutputs[name] = append(out.currentOutputs[name], entry)
		case "agent":
			out.currentAgent = m.Fields()
		}
	}

	return out
}

func (s *statistics) remove(snap *statistics, ts time.Time) {
	s.Lock()
	defer s.Unlock()

	s.metrics -= snap.metrics
	s.logErrors -= snap.logErrors
	s.logWarnings -= snap.logWarnings
	s.lastUpdate = ts
	s.lastUpdateFailed = false

	s.currentAgent = snap.currentAgent
	s.currentInputs = snap.currentInputs
	s.currentOutputs = snap.currentOutputs
}

func (s *statistics) variables() map[string]interface{} {
	s.RLock()
	defer s.RUnlock()

	// Add the raw statistucs
	vars := map[string]interface{}{
		"metrics":      s.metrics,
		"log_errors":   s.logErrors,
		"log_warnings": s.logWarnings,
	}
	if s.lastUpdate.IsZero() {
		vars["last_update"] = nil
	} else {
		vars["last_update"] = s.lastUpdate
	}

	// Calculate diff
	agent := maps.Clone(s.currentAgent)
	for k, v := range s.lastAgent {
		agent[k] = agent[k].(int64) - v.(int64)
	}
	vars["agent"] = agent

	inputs := maps.Clone(s.currentInputs)
	for name, entries := range inputs {
		lastInput, found := s.lastInputs[name]
		if !found {
			continue
		}

		ids := make(map[string]int, len(entries))
		for i, e := range entries {
			ids[e["id"].(string)] = i
		}

		for _, old := range lastInput {
			id := old["id"].(string)
			index := ids[id]

			for k, raw := range old {
				// Ignore known non-accumulated fields
				if k == "gather_time_ns" {
					continue
				}

				v, ok := raw.(int64)
				if !ok {
					continue
				}
				entries[index][k] = entries[index][k].(int64) - v
			}
		}
		inputs[name] = entries
	}
	vars["inputs"] = inputs

	outputs := maps.Clone(s.currentOutputs)
	for name, entries := range outputs {
		lastOutput, found := s.lastOutputs[name]
		if !found {
			continue
		}

		ids := make(map[string]int, len(entries))
		for i, e := range entries {
			ids[e["id"].(string)] = i
		}

		for _, old := range lastOutput {
			id := old["id"].(string)
			index := ids[id]

			for k, raw := range old {
				// Ignore known non-accumulated fields
				switch k {
				case "write_time_ns", "buffer_size", "buffer_limit", "buffer_fullness":
					continue
				}

				v, ok := raw.(int64)
				if !ok {
					continue
				}
				entries[index][k] = entries[index][k].(int64) - v
			}
		}
		outputs[name] = entries
	}
	vars["outputs"] = outputs

	return vars
}
