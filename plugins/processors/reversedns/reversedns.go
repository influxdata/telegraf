package reversedns

import (
	"time"

	"github.com/influxdata/telegraf/plugins/processors/reversedns/parallel"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/processors"
)

const sampleConfig = `
  # For optimal performance, you may want to limit which metrics are passed to this
  # processor. eg:
  # namepass = ["my_metric_*"]

  # cache_ttl is how long the dns entries should stay cached for.
  # generally longer is better, but if you expect a large number of diverse lookups
  # you'll want to consider memory use.
  cache_ttl = "24h"

  # lookup_timeout is how long should you wait for a single dns request to repsond.
  # this is also the maximum acceptable latency for a metric travelling through
  # the reversedns processor. After lookup_timeout is exceeded, a metric will
  # be passed on unaltered.
  # multiple simultaneous resolution requests for the same IP will only make a
  # single rDNS request, and they will all wait for the answer for this long.
  lookup_timeout = "3s"

  max_parallel_lookups = 100

  [[processors.reversedns.lookup]]
    # get the ip from the field "source_ip", and put the result in the field "source_name"
    field = "source_ip"
    dest = "source_name"

  [[processors.reversedns.lookup]]
    # get the ip from the tag "destination_ip", and put the result in the tag 
    # "destination_name".
    tag = "destination_ip"
    dest = "destination_name"

    # If you would prefer destination_name to be a field you can use a subsequent 
    # converter like so:
    #   [[processors.converter.tags]]
    #     string = ["destination_name"]
    #     order = 2 # orders are necessary with multiple processors when order matters

`

type lookupEntry struct {
	Tag   string `toml:"tag"`
	Field string `toml:"field"`
	Dest  string `toml:"dest"`
}

type ReverseDNS struct {
	reverseDNSCache *ReverseDNSCache
	acc             telegraf.MetricStreamAccumulator
	workers         *parallel.Parallel

	Lookups            []lookupEntry   `toml:"lookup"`
	CacheTTL           config.Duration `toml:"cache_ttl"`
	LookupTimeout      config.Duration `toml:"lookup_timeout"`
	MaxParallelLookups int             `toml:"max_parallel_lookups"`
}

func (r *ReverseDNS) SampleConfig() string {
	return sampleConfig
}

func (r *ReverseDNS) Description() string {
	return "ReverseDNS does a reverse lookup on IP addresses to retrieve the DNS name"
}

func (r *ReverseDNS) Init() {
	r.reverseDNSCache = NewReverseDNSCache(
		time.Duration(r.CacheTTL),
		time.Duration(r.LookupTimeout),
		r.MaxParallelLookups, // max parallel reverse-dns lookups
	)
}

func (r *ReverseDNS) Start(acc telegraf.MetricStreamAccumulator) error {
	r.acc = acc
	r.workers = parallel.New(acc, 10000).Ordered()
	return nil
}

func (r *ReverseDNS) Stop() error {
	r.workers.Wait()
	return nil
}

func (r *ReverseDNS) Add(metric telegraf.Metric) {
	r.workers.Parallel(func(acc telegraf.MetricStreamAccumulator) {
		for _, lookup := range r.Lookups {
			if len(lookup.Field) > 0 {
				if ipField, ok := metric.GetField(lookup.Field); ok {
					if ip, ok := ipField.(string); ok {
						metric.AddField(lookup.Dest, first(r.reverseDNSCache.Lookup(ip)))
					}
				}
			}
			if len(lookup.Tag) > 0 {
				if ipTag, ok := metric.GetTag(lookup.Tag); ok {
					metric.AddTag(lookup.Dest, first(r.reverseDNSCache.Lookup(ipTag)))
				}
			}
		}
		acc.PassMetric(metric)
	})
}

func first(s []string) string {
	if len(s) == 0 {
		return ""
	}
	return s[0]
}

func init() {
	processors.AddStreaming("reversedns", func() telegraf.StreamingProcessor {
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
