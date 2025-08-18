//go:generate ../../../tools/readme_config_includer/generator
package internal

import (
	_ "embed"
	"fmt"
	"runtime"
	"runtime/metrics"
	"strings"

	"github.com/influxdata/telegraf"
	inter "github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/selfstat"
)

//go:embed sample.conf
var sampleConfig string

// Converts /cpu/classes/gc/mark/assist:cpu-seconds to cpu_classes_gc_mark_assist_cpu_seconds
var replacer = strings.NewReplacer("/", "_", ":", "_", "-", "_")

type Internal struct {
	CollectMemstats bool `toml:"collect_memstats"`
	CollectGostats  bool `toml:"collect_gostats"`
	PerInstance     bool `toml:"per_instance"`
}

func (*Internal) SampleConfig() string {
	return sampleConfig
}

func (s *Internal) Gather(acc telegraf.Accumulator) error {
	if s.PerInstance {
		collectIndividualPluginStat(acc)
	} else {
		collectAccumulatedPluginStat(acc)
	}

	if s.CollectMemstats {
		collectMemStat(acc)
	}

	if s.CollectGostats {
		collectGoStat(acc)
	}

	return nil
}

func collectIndividualPluginStat(acc telegraf.Accumulator) {
	for _, m := range selfstat.Metrics() {
		if m.Name() == "internal_agent" {
			m.AddTag("go_version", strings.TrimPrefix(runtime.Version(), "go"))
		}
		m.AddTag("version", inter.Version)
		acc.AddFields(m.Name(), m.Fields(), m.Tags(), m.Time())
	}
}

func collectAccumulatedPluginStat(acc telegraf.Accumulator) {
	accumulated := make(map[uint64]telegraf.Metric)
	for _, m := range selfstat.Metrics() {
		if m.Name() == "internal_agent" {
			m.AddTag("go_version", strings.TrimPrefix(runtime.Version(), "go"))
		}
		m.AddTag("version", inter.Version)
		key := m.HashIDWithFieldsFiltered([]string{"_id"}, nil)
		am, found := accumulated[key]
		if !found {
			accumulated[key] = m.Copy()
			accumulated[key].RemoveTag("_id")
			continue
		}
		for _, f := range m.FieldList() {
			var av int64
			if af, found := am.GetField(f.Key); found {
				av = af.(int64)
			}
			am.AddField(f.Key, av+f.Value.(int64))
		}
	}

	for _, m := range accumulated {
		acc.AddMetric(m)
	}
}

func collectMemStat(acc telegraf.Accumulator) {
	m := &runtime.MemStats{}
	runtime.ReadMemStats(m)
	fields := map[string]any{
		"alloc_bytes":       m.Alloc,      // bytes allocated and not yet freed
		"total_alloc_bytes": m.TotalAlloc, // bytes allocated (even if freed)
		"sys_bytes":         m.Sys,        // bytes obtained from system (sum of XxxSys below)
		"pointer_lookups":   m.Lookups,    // number of pointer lookups
		"mallocs":           m.Mallocs,    // number of mallocs
		"frees":             m.Frees,      // number of frees

		// Main allocation heap statistics.
		"heap_alloc_bytes":    m.HeapAlloc,    // bytes allocated and not yet freed (same as Alloc above)
		"heap_sys_bytes":      m.HeapSys,      // bytes obtained from system
		"heap_idle_bytes":     m.HeapIdle,     // bytes in idle spans
		"heap_in_use_bytes":   m.HeapInuse,    // bytes in non-idle span
		"heap_released_bytes": m.HeapReleased, // bytes released to the OS
		"heap_objects":        m.HeapObjects,  // total number of allocated objects
		"num_gc":              m.NumGC,
	}
	acc.AddFields("internal_memstats", fields, make(map[string]string))
}

func collectGoStat(acc telegraf.Accumulator) {
	descs := metrics.All()
	samples := make([]metrics.Sample, len(descs))
	for i := range samples {
		samples[i].Name = descs[i].Name
	}
	metrics.Read(samples)

	fields := make(map[string]any, len(samples))
	for _, sample := range samples {
		name := sanitizeName(sample.Name)

		switch sample.Value.Kind() {
		case metrics.KindUint64:
			fields[name] = sample.Value.Uint64()
		case metrics.KindFloat64:
			fields[name] = sample.Value.Float64()
		case metrics.KindFloat64Histogram:
			// The histogram may be quite large, so let's just pull out
			// a crude estimate for the median for the sake of this example.
			fields[name] = medianBucket(sample.Value.Float64Histogram())
		default:
			// This may happen as new metrics get added.
			//
			// The safest thing to do here is to simply log it somewhere
			// as something to look into, but ignore it for now.
			// In the worst case, you might temporarily miss out on a new metric.
			fmt.Printf("%s: unexpected metric Kind: %v\n", name, sample.Value.Kind())
		}
	}

	tags := map[string]string{
		"go_version": strings.TrimPrefix(runtime.Version(), "go"),
	}
	acc.AddFields("internal_gostats", fields, tags)
}

func sanitizeName(name string) string {
	return replacer.Replace(strings.TrimPrefix(name, "/"))
}

func medianBucket(h *metrics.Float64Histogram) float64 {
	total := uint64(0)
	for _, count := range h.Counts {
		total += count
	}
	thresh := total / 2
	total = 0
	for i, count := range h.Counts {
		total += count
		if total >= thresh {
			return h.Buckets[i]
		}
	}

	// default value in case something above did not work
	return 0.0
}

func init() {
	inputs.Add("internal", func() telegraf.Input {
		return &Internal{
			CollectMemstats: true,
		}
	})
}
