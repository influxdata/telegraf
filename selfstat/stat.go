package selfstat

import (
	"maps"
	"sync/atomic"
)

type stat struct {
	v           atomic.Int64
	measurement string
	field       string
	tags        map[string]string
}

func (s *stat) Incr(v int64) {
	s.v.Add(v)
}

func (s *stat) Set(v int64) {
	s.v.Store(v)
}

func (s *stat) Get() int64 {
	return s.v.Load()
}

func (s *stat) Name() string {
	return s.measurement
}

func (s *stat) FieldName() string {
	return s.field
}

// Tags returns a copy of the stat's tags.
// NOTE this allocates a new map every time it is called.
func (s *stat) Tags() map[string]string {
	m := make(map[string]string, len(s.tags))
	maps.Copy(m, s.tags)
	return m
}

// Unregister removes this stat from the registry only
func (s *stat) Unregister() {
	registry.remove(s.measurement, s.field, s.tags)
}
