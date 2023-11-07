//go:generate ../../../tools/readme_config_includer/generator
package snmp_lookup

import (
	_ "embed"
	"fmt"
	"strconv"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal/snmp"
	"github.com/influxdata/telegraf/plugins/common/parallel"
	si "github.com/influxdata/telegraf/plugins/inputs/snmp"
	"github.com/influxdata/telegraf/plugins/processors"

	"github.com/gosnmp/gosnmp"
	"github.com/hashicorp/golang-lru/v2/expirable"
)

//go:embed sample.conf
var sampleConfig string

type signalMap map[string]chan struct{}
type tagMapRows map[string]map[string]string
type tagMap struct {
	created time.Time
	rows    tagMapRows
}

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
	if name != "gosmi" {
		l.Log.Warnf("unsupported agent.snmp_translator value %q, some features might not work", name)
	}

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
	default:
		return fmt.Errorf("invalid agent.snmp_translator value %q", l.Translator)
	}

	return l.initTable()
}

func (l *Lookup) initTable() error {
	l.table.Name = "lookup"
	l.table.IndexAsTag = true
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
	agent, ok := metric.GetTag(l.AgentTag)
	if !ok {
		l.Log.Warn("Agent tag missing")
		return []telegraf.Metric{metric}
	}

	index, ok := metric.GetTag(l.IndexTag)
	if !ok {
		l.Log.Warn("Index tag missing")
		return []telegraf.Metric{metric}
	}

	// Check cache
	tagMap, inCache := l.cache.Peek(agent)
	_, indexExists := tagMap.rows[index]

	// Cache miss
	if !inCache || (!indexExists && time.Since(tagMap.created) > minRetry) {
		gs, err := l.getConnection(metric)
		if err != nil {
			l.Log.Errorf("Could not connect to %q: %v", agent, err)
			return []telegraf.Metric{metric}
		}

		if err = l.loadTagMap(gs, agent); err != nil {
			l.Log.Errorf("Could not load table from %q: %v", agent, err)
			return []telegraf.Metric{metric}
		}
	}

	// Load from cache
	tagMap, inCache = l.cache.Get(agent)
	tags, indexExists := tagMap.rows[index]
	if inCache && indexExists {
		for key, value := range tags {
			metric.AddTag(key, value)
		}
	} else {
		l.Log.Warnf("Could not find index %q on agent %q", index, agent)
	}

	return []telegraf.Metric{metric}
}

// snmpConnection is an interface which wraps a *gosnmp.GoSNMP object.
// We interact through an interface so we can mock it out in tests.
type snmpConnection interface {
	Host() string
	Walk(string, gosnmp.WalkFunc) error
	Get(oids []string) (*gosnmp.SnmpPacket, error)
	Reconnect() error
}

func (l *Lookup) getConnection(metric telegraf.Metric) (snmpConnection, error) {
	clientConfig := l.ClientConfig

	if version, ok := metric.GetTag("version"); ok {
		// inputs.snmp_trap reports like this
		if version == "2c" {
			version = "2"
		}

		v, err := strconv.ParseUint(version, 10, 8)
		if err != nil {
			return nil, fmt.Errorf("parsing version: %w", err)
		}
		clientConfig.Version = uint8(v)
	}

	if community, ok := metric.GetTag("community"); ok {
		clientConfig.Community = community
	}

	if secName, ok := metric.GetTag("sec_name"); ok {
		clientConfig.SecName = secName
	}

	if secLevel, ok := metric.GetTag("sec_level"); ok {
		clientConfig.SecLevel = secLevel
	}

	if authProtocol, ok := metric.GetTag("auth_protocol"); ok {
		clientConfig.AuthProtocol = authProtocol
	}

	if authPassword, ok := metric.GetTag("auth_password"); ok {
		clientConfig.AuthPassword = authPassword
	}

	if privProtocol, ok := metric.GetTag("priv_protocol"); ok {
		clientConfig.PrivProtocol = privProtocol
	}

	if privPassword, ok := metric.GetTag("priv_password"); ok {
		clientConfig.PrivPassword = privPassword
	}

	if contextName, ok := metric.GetTag("context_name"); ok {
		clientConfig.ContextName = contextName
	}

	gs, err := snmp.NewWrapper(clientConfig)
	if err != nil {
		return gs, fmt.Errorf("parsing SNMP client config: %w", err)
	}

	if agent, ok := metric.GetTag(l.AgentTag); ok {
		if err = gs.SetAgent(agent); err != nil {
			return gs, fmt.Errorf("parsing agent tag: %w", err)
		}
	}

	if err := gs.Connect(); err != nil {
		return gs, fmt.Errorf("connecting failed: %w", err)
	}

	return gs, nil
}

func (l *Lookup) loadTagMap(gs snmpConnection, agent string) error {
	table, err := l.table.Build(gs, true, l.translator)
	if err != nil {
		return err
	}

	tagMap := tagMap{
		created: table.Time,
		rows:    make(tagMapRows, len(table.Rows)),
	}

	for _, row := range table.Rows {
		index := row.Tags["index"]
		delete(row.Tags, "index")
		tagMap.rows[index] = row.Tags
	}

	l.cache.Add(agent, tagMap)

	return nil
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
