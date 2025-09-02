package selfstat

import (
	"bytes"
	"maps"
	"slices"
)

type Collector struct {
	tags       map[string]string
	statistics map[string]Stat
}

func NewCollector(tags map[string]string) *Collector {
	var capacity int
	if tags != nil {
		capacity += len(tags)
	}

	s := &Collector{
		tags:       make(map[string]string, capacity),
		statistics: make(map[string]Stat),
	}

	// Collect all tags and add the plugin ID
	if tags != nil {
		maps.Copy(s.tags, tags)
	}

	return s
}

func (s *Collector) Register(measurement, field string, tags map[string]string) Stat {
	// Compute the stats-key and exit early if the stat was already registered
	key := collectorKey(measurement, field, tags)
	if stat, found := s.statistics[key]; found {
		return stat
	}

	// Merge the tags of the statistic with the collector tags
	capacity := len(s.tags)
	if tags != nil {
		capacity += len(tags)
	}

	t := make(map[string]string, capacity)
	maps.Copy(t, s.tags)
	if tags != nil {
		maps.Copy(t, tags)
	}

	// Register the new stat and return it
	stat := Register(measurement, field, t)
	s.statistics[key] = stat

	return stat
}

func (s *Collector) RegisterTiming(measurement, field string, tags map[string]string) Stat {
	// Compute the stats-key and exit early if the stat was already registered
	key := collectorKey(measurement, field, tags)
	if stat, found := s.statistics[key]; found {
		return stat
	}

	// Merge the tags of the statistic with the collector tags
	capacity := len(s.tags)
	if tags != nil {
		capacity += len(tags)
	}

	t := make(map[string]string, capacity)
	maps.Copy(t, s.tags)
	if tags != nil {
		maps.Copy(t, tags)
	}

	// Register the new stat and return it
	stat := RegisterTiming(measurement, field, t)
	s.statistics[key] = stat

	return stat
}

func (s *Collector) Unregister(measurement, field string, tags map[string]string) {
	capacity := len(s.tags)
	if tags != nil {
		capacity += len(tags)
	}

	t := make(map[string]string, capacity)
	maps.Copy(t, s.tags)
	if tags != nil {
		maps.Copy(t, tags)
	}

	Unregister(measurement, field, t)

	// Compute the stats-key and delete the entry
	key := collectorKey(measurement, field, tags)
	delete(s.statistics, key)
}

func (s *Collector) UnregisterAll() {
	for _, s := range s.statistics {
		s.Unregister()
	}
	clear(s.statistics)
}

func (s *Collector) Get(measurement, field string, tags map[string]string) Stat {
	key := collectorKey(measurement, field, tags)
	return s.statistics[key]
}

func (s *Collector) Reset(measurement, field string, tags map[string]string) {
	key := collectorKey(measurement, field, tags)
	if stats := s.statistics[key]; stats != nil {
		stats.Set(0)
	}
}

func collectorKey(measurement, field string, tags map[string]string) string {
	var buf bytes.Buffer
	buf.WriteString(measurement + "\n")
	buf.WriteString(field + "\n")

	if tags == nil {
		return buf.String()
	}

	for _, k := range slices.Sorted(maps.Keys(tags)) {
		v := tags[k]
		buf.WriteString(k + "=" + v + "\n")
	}
	return buf.String()
}
