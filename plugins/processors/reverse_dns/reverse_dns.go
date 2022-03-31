package reverse_dns

import (
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/parallel"
	"github.com/influxdata/telegraf/plugins/processors"
)

type lookupEntry struct {
	Tag   string `toml:"tag"`
	Field string `toml:"field"`
	Dest  string `toml:"dest"`
}

type ReverseDNS struct {
	reverseDNSCache *ReverseDNSCache
	acc             telegraf.Accumulator
	parallel        parallel.Parallel

	Lookups            []lookupEntry   `toml:"lookup"`
	CacheTTL           config.Duration `toml:"cache_ttl"`
	LookupTimeout      config.Duration `toml:"lookup_timeout"`
	MaxParallelLookups int             `toml:"max_parallel_lookups"`
	Ordered            bool            `toml:"ordered"`
	Log                telegraf.Logger `toml:"-"`
}

func (r *ReverseDNS) Start(acc telegraf.Accumulator) error {
	r.acc = acc
	r.reverseDNSCache = NewReverseDNSCache(
		time.Duration(r.CacheTTL),
		time.Duration(r.LookupTimeout),
		r.MaxParallelLookups, // max parallel reverse-dns lookups
	)
	if r.Ordered {
		r.parallel = parallel.NewOrdered(acc, r.asyncAdd, 10000, r.MaxParallelLookups)
	} else {
		r.parallel = parallel.NewUnordered(acc, r.asyncAdd, r.MaxParallelLookups)
	}
	return nil
}

func (r *ReverseDNS) Stop() error {
	r.parallel.Stop()
	r.reverseDNSCache.Stop()
	return nil
}

func (r *ReverseDNS) Add(metric telegraf.Metric, _ telegraf.Accumulator) error {
	r.parallel.Enqueue(metric)
	return nil
}

func (r *ReverseDNS) asyncAdd(metric telegraf.Metric) []telegraf.Metric {
	for _, lookup := range r.Lookups {
		if len(lookup.Field) > 0 {
			if ipField, ok := metric.GetField(lookup.Field); ok {
				if ip, ok := ipField.(string); ok {
					result, err := r.reverseDNSCache.Lookup(ip)
					if err != nil {
						r.Log.Errorf("lookup error: %v", err)
						continue
					}
					if len(result) > 0 {
						metric.AddField(lookup.Dest, result[0])
					}
				}
			}
		}
		if len(lookup.Tag) > 0 {
			if ipTag, ok := metric.GetTag(lookup.Tag); ok {
				result, err := r.reverseDNSCache.Lookup(ipTag)
				if err != nil {
					r.Log.Errorf("lookup error: %v", err)
					continue
				}
				if len(result) > 0 {
					metric.AddTag(lookup.Dest, result[0])
				}
			}
		}
	}
	return []telegraf.Metric{metric}
}

func init() {
	processors.AddStreaming("reverse_dns", func() telegraf.StreamingProcessor {
		return newReverseDNS()
	})
}

func newReverseDNS() *ReverseDNS {
	return &ReverseDNS{
		CacheTTL:           config.Duration(24 * time.Hour),
		LookupTimeout:      config.Duration(1 * time.Minute),
		MaxParallelLookups: 10,
	}
}
