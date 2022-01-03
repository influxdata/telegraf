package reverse_dns

import (
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/parallel"
	"github.com/influxdata/telegraf/plugins/processors"
)

const sampleConfig = `
  ## For optimal performance, you may want to limit which metrics are passed to this
  ## processor. eg:
  ## namepass = ["my_metric_*"]

  ## cache_ttl is how long the dns entries should stay cached for.
  ## generally longer is better, but if you expect a large number of diverse lookups
  ## you'll want to consider memory use.
  cache_ttl = "24h"

  ## lookup_timeout is how long should you wait for a single dns request to repsond.
  ## this is also the maximum acceptable latency for a metric travelling through
  ## the reverse_dns processor. After lookup_timeout is exceeded, a metric will
  ## be passed on unaltered.
  ## multiple simultaneous resolution requests for the same IP will only make a
  ## single rDNS request, and they will all wait for the answer for this long.
  lookup_timeout = "3s"

  ## max_parallel_lookups is the maximum number of dns requests to be in flight
  ## at the same time. Requesting hitting cached values do not count against this
  ## total, and neither do mulptiple requests for the same IP.
  ## It's probably best to keep this number fairly low.
  max_parallel_lookups = 10

  ## ordered controls whether or not the metrics need to stay in the same order
  ## this plugin received them in. If false, this plugin will change the order
  ## with requests hitting cached results moving through immediately and not
  ## waiting on slower lookups. This may cause issues for you if you are
  ## depending on the order of metrics staying the same. If so, set this to true.
  ## keeping the metrics ordered may be slightly slower.
  ordered = false

  [[processors.reverse_dns.lookup]]
    ## get the ip from the field "source_ip", and put the result in the field "source_name"
    field = "source_ip"
    dest = "source_name"

  [[processors.reverse_dns.lookup]]
    ## get the ip from the tag "destination_ip", and put the result in the tag
    ## "destination_name".
    tag = "destination_ip"
    dest = "destination_name"

    ## If you would prefer destination_name to be a field instead, you can use a
    ## processors.converter after this one, specifying the order attribute.
`

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

func (r *ReverseDNS) SampleConfig() string {
	return sampleConfig
}

func (r *ReverseDNS) Description() string {
	return "ReverseDNS does a reverse lookup on IP addresses to retrieve the DNS name"
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
