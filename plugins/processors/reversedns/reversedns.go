package reversedns

import (
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/processors"
)

const sampleConfig = `
[[processors.reversedns]]
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
	exitWg          *sync.WaitGroup
	reverseDNSCache *ReverseDNSCache

	Lookups            []lookupEntry     `toml:"lookup"`
	CacheTTL           internal.Duration `toml:"cache_ttl"`
	LookupTimeout      internal.Duration `toml:"lookup_timeout"`
	MaxParallelLookups int               `toml:"max_parallel_lookups"`
}

func (r *ReverseDNS) SampleConfig() string {
	return sampleConfig
}

func (r *ReverseDNS) Description() string {
	return "ReverseDNS does a reverse lookup on IP addresses to retrieve the DNS name"
}

type futureMetric func() telegraf.Metric

func (r *ReverseDNS) Start(acc telegraf.MetricStream) error {
	r.exitWg.Add(2)

	r.reverseDNSCache = NewReverseDNSCache(r.CacheTTL.Duration, r.LookupTimeout.Duration)

	orderedOutChannel := make(chan futureMetric, 10000)

	go r.metricReader(acc, orderedOutChannel)
	go r.orderedWriter(acc, orderedOutChannel)

	r.exitWg.Wait()
	return nil
}

func (r *ReverseDNS) Stop() {
	r.exitWg.Wait()
}

func (r *ReverseDNS) metricReader(acc telegraf.MetricStream, outChannel chan<- futureMetric) {
	defer r.exitWg.Done()

	for {
		m := acc.GetNextMetric()
		if m == nil {
			if acc.IsStreamClosed() {
				close(outChannel)
				return
			}
		}

		workChan := make(chan telegraf.Metric)

		go func(m telegraf.Metric, workChan chan telegraf.Metric) {
			// todo: these are not in parallel. Does that matter? more promises?
			for _, lookup := range r.Lookups {
				if len(lookup.Field) > 0 {
					if ipField, ok := m.GetField(lookup.Field); ok {
						if ip, ok := ipField.(string); ok {
							m.AddField(lookup.Dest, first(r.reverseDNSCache.Lookup(ip)))
						}
					}
				}
				if len(lookup.Tag) > 0 {
					if ipTag, ok := m.GetTag(lookup.Tag); ok {
						m.AddTag(lookup.Dest, first(r.reverseDNSCache.Lookup(ipTag)))
					}
				}
			}
			workChan <- m
		}(m, workChan)

		outChannel <- func() telegraf.Metric {
			finishedMetric := <-workChan
			return finishedMetric
		}

	}
}

func first(s []string) string {
	if len(s) == 0 {
		return ""
	}
	return s[0]
}

func (r *ReverseDNS) orderedWriter(acc telegraf.MetricStream, outChannel <-chan futureMetric) {
	defer r.exitWg.Done()

	for futureMetricFunc := range outChannel {
		acc.PassMetric(futureMetricFunc())
	}
}

func init() {
	processors.AddStreaming("reversedns", func() telegraf.StreamingProcessor {
		return newReverseDNS()
	})
}

func newReverseDNS() *ReverseDNS {
	return &ReverseDNS{
		exitWg:             &sync.WaitGroup{},
		CacheTTL:           internal.Duration{Duration: 24 * time.Hour},
		LookupTimeout:      internal.Duration{Duration: 1 * time.Minute},
		MaxParallelLookups: 10,
	}
}
