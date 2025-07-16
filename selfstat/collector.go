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
	var cap int
	if tags != nil {
		cap += len(tags)
	}

	s := &Collector{
		tags:       make(map[string]string, cap),
		statistics: make(map[string]Stat),
	}

	// Collect all tags and add the plugin ID
	if tags != nil {
		maps.Copy(s.tags, tags)
	}

	return s
}

func (s *Collector) Register(measurement, field string, tags map[string]string) {
	cap := len(s.tags)
	if tags != nil {
		cap += len(tags)
	}

	t := make(map[string]string, cap)
	maps.Copy(t, s.tags)
	if tags != nil {
		maps.Copy(t, tags)
	}

	key := collectorKey(measurement, field, tags)
	s.statistics[key] = Register(measurement, field, t)
}

func (s *Collector) RegisterTiming(measurement, field string, tags map[string]string) {
	cap := len(s.tags)
	if tags != nil {
		cap += len(tags)
	}

	t := make(map[string]string, cap)
	maps.Copy(t, s.tags)
	if tags != nil {
		maps.Copy(t, tags)
	}

	key := collectorKey(measurement, field, tags)
	s.statistics[key] = RegisterTiming(measurement, field, t)
}

func (s *Collector) Unregister(measurement, field string, tags map[string]string) {
	cap := len(s.tags)
	if tags != nil {
		cap += len(tags)
	}

	t := make(map[string]string, cap)
	maps.Copy(t, s.tags)
	if tags != nil {
		maps.Copy(t, tags)
	}

	Unregister(measurement, field, t)
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
