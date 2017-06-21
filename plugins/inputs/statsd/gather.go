package statsd

import (
	"fmt"
	"log"
	"time"
)

func (s *Statsd) gatherTimings(now time.Time) []*accumulator {
	log.Println("D! Statsd.Gather() Starting range of s.timings")
	t, ok := <-s.timings
	if !ok {
		panic("chan cachedtimings is closed")
	}
	log.Printf("D! Statsd.Gather().gatherTimings len=%d", len(t))
	a := make([]*accumulator, 0, len(t))
	for key := range t {
		// Defining a template to parse field names for timers allows us to split
		// out multiple fields per timer. In this case we prefix each stat with the
		// field name and store these all in a single measurement.
		fields := make(map[string]interface{})
		for fieldName, stats := range t[key].fields {
			var prefix string
			if fieldName != defaultFieldName {
				prefix = fieldName + "_"
			}
			fields[prefix+"mean"] = stats.Mean()
			fields[prefix+"stddev"] = stats.Stddev()
			fields[prefix+"upper"] = stats.Upper()
			fields[prefix+"lower"] = stats.Lower()
			fields[prefix+"count"] = stats.Count()
			fmt.Println(s.Percentiles)
			for _, percentile := range s.Percentiles {
				name := fmt.Sprintf("%s%d_percentile", prefix, percentile)
				fields[name] = stats.Percentile(percentile)
			}
		}
		a = append(a, &accumulator{
			measurement: t[key].name,
			fields:      fields,
			tags:        t[key].tags,
			t:           []time.Time{now},
		})
	}
	if s.DeleteTimings {
		s.timingsReset <- struct{}{}
	}
	return a
}

func (s *Statsd) gatherGauges(now time.Time) []*accumulator {
	log.Println("D! Statsd.Gather() Starting range of s.gauges")
	g, ok := <-s.gauges
	if !ok {
		panic("chan cachedgauges is closed")
	}
	log.Printf("D! Statsd.Gather().gatherGauges len=%d", len(g))
	a := make([]*accumulator, 0, len(g))
	for key := range g {
		a = append(a, &accumulator{
			measurement: g[key].name,
			fields:      g[key].fields,
			tags:        g[key].tags,
			t:           []time.Time{now},
		})
	}
	if s.DeleteGauges {
		s.gaugesReset <- struct{}{}
	}
	return a
}

func (s *Statsd) gatherCounters(now time.Time) []*accumulator {
	log.Println("D! Statsd.Gather() Starting range of s.counters")
	c, ok := <-s.counters
	if !ok {
		panic("chan cachedcounters is closed")
	}
	log.Printf("D! Statsd.Gather().gatherCounters len=%d", len(c))
	a := make([]*accumulator, 0, len(c))
	for key := range c {
		a = append(a, &accumulator{
			measurement: c[key].name,
			fields:      c[key].fields,
			tags:        c[key].tags,
			t:           []time.Time{now},
		})
	}
	if s.DeleteCounters {
		s.countersReset <- struct{}{}
	}
	return a
}

func (s *Statsd) gatherSets(now time.Time) []*accumulator {
	log.Println("D! Statsd.Gather() Starting range of s.sets")
	sets, ok := <-s.sets
	if !ok {
		panic("chan cachedsets is closed")
	}
	log.Printf("D! Statsd.Gather().gatherSets len=%d", len(sets))
	a := make([]*accumulator, 0, len(sets))
	for key := range sets {
		fields := make(map[string]interface{})
		for field, set := range sets[key].fields {
			fields[field] = int64(len(set))
		}
		a = append(a, &accumulator{
			measurement: sets[key].name,
			fields:      fields,
			tags:        sets[key].tags,
			t:           []time.Time{now},
		})
	}
	if s.DeleteSets {
		s.setsReset <- struct{}{}
	}
	return a
}
