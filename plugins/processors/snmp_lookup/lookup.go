//go:generate ../../../tools/readme_config_includer/generator
package snmp_lookup

import (
	_ "embed"
	"fmt"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal/snmp"
	"github.com/influxdata/telegraf/plugins/common/parallel"
	si "github.com/influxdata/telegraf/plugins/inputs/snmp"
	"github.com/influxdata/telegraf/plugins/processors"

	"github.com/hashicorp/golang-lru/v2/expirable"
)

//go:embed sample.conf
var sampleConfig string

type signalMap map[string]chan struct{}
type tagMap map[string]map[string]string

type Lookup struct {
	AgentTag string     `toml:"agent_tag"`
	IndexTag string     `toml:"index_tag"`
	Tags     []si.Field `toml:"tag"`

	snmp.ClientConfig

	CacheSize       int             `toml:"max_cache_entries"`
	ParallelLookups int             `toml:"max_parallel_lookups"`
	Ordered         bool            `toml:"ordered"`
	CacheTTL        config.Duration `toml:"cache_ttl"`

	Log telegraf.Logger `toml:"-"`

	cache    *expirable.LRU[string, tagMap]
	parallel parallel.Parallel
	sigs     signalMap
	table    si.Table

	translator si.Translator
}

const (
	defaultCacheSize       = 100
	defaultCacheTTL        = config.Duration(8 * time.Hour)
	defaultParallelLookups = 100
	minRetry               = 5 * time.Minute
	orderedQueueSize       = 10_000
)

func (*Lookup) SampleConfig() string {
	return sampleConfig
}

func (l *Lookup) SetTranslator(name string) {
	l.Translator = name
}

func (l *Lookup) Init() (err error) {
	l.sigs = make(signalMap)

	if _, err = snmp.NewWrapper(l.ClientConfig); err != nil {
		return fmt.Errorf("parsing SNMP client config: %w", err)
	}

	switch l.Translator {
	case "", "gosmi":
		if l.translator, err = si.NewGosmiTranslator(l.Path, l.Log); err != nil {
			return fmt.Errorf("loading translator: %w", err)
		}
	case "netsnmp":
		l.translator = si.NewNetsnmpTranslator()
		l.Log.Warnf("unsupported agent.snmp_translator value %q, some features might not work", l.Translator)
	default:
		return fmt.Errorf("invalid agent.snmp_translator value %q", l.Translator)
	}

	return l.initTable()
}

func (l *Lookup) initTable() error {
	l.table.Name = "lookup"
	l.table.Fields = make([]si.Field, len(l.Tags))
	for i, tag := range l.Tags {
		tag.IsTag = true
		l.table.Fields[i] = tag
	}

	return l.table.Init(l.translator)
}

func (l *Lookup) Start(acc telegraf.Accumulator) error {
	l.cache = expirable.NewLRU[string, tagMap](l.CacheSize, nil, time.Duration(l.CacheTTL))
	if l.Ordered {
		l.parallel = parallel.NewOrdered(acc, l.addAsync, orderedQueueSize, l.ParallelLookups)
	} else {
		l.parallel = parallel.NewUnordered(acc, l.addAsync, l.ParallelLookups)
	}
	return nil
}

func (l *Lookup) Add(metric telegraf.Metric, _ telegraf.Accumulator) error {
	l.parallel.Enqueue(metric)
	return nil
}

func (l *Lookup) addAsync(metric telegraf.Metric) []telegraf.Metric {
	// TODO: lookup
	return []telegraf.Metric{metric}
}

func (l *Lookup) Stop() {
	l.parallel.Stop()
}

func init() {
	processors.AddStreaming("snmp_lookup", func() telegraf.StreamingProcessor {
		return &Lookup{
			AgentTag:        "source",
			IndexTag:        "index",
			CacheSize:       defaultCacheSize,
			ParallelLookups: defaultParallelLookups,
			ClientConfig:    *snmp.DefaultClientConfig(),
			CacheTTL:        defaultCacheTTL,
		}
	})
}
