//go:generate ../../../tools/readme_config_includer/generator
package snmp_lookup

import (
	_ "embed"
	"fmt"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal/snmp"
	si "github.com/influxdata/telegraf/plugins/inputs/snmp"
	"github.com/influxdata/telegraf/plugins/processors"

	"github.com/gosnmp/gosnmp"
)

//go:embed sample.conf
var sampleConfig string

// snmpConnection is an interface which wraps a *gosnmp.GoSNMP object.
// We interact through an interface so we can mock it out in tests.
type snmpConnection interface {
	Host() string
	Walk(string, gosnmp.WalkFunc) error
	Get(oids []string) (*gosnmp.SnmpPacket, error)
	Reconnect() error
}

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

	CacheSize             int             `toml:"max_cache_entries"`
	ParallelLookups       int             `toml:"max_parallel_lookups"`
	Ordered               bool            `toml:"ordered"`
	CacheTTL              config.Duration `toml:"cache_ttl"`
	MinTimeBetweenUpdates config.Duration `toml:"min_time_between_updates"`

	Log telegraf.Logger `toml:"-"`

	translator si.Translator
	table      si.Table

	cache             *store
	backlog           *backlog
	getConnectionFunc func(string) (snmpConnection, error)
}

const (
	defaultCacheSize             = 100
	defaultCacheTTL              = config.Duration(8 * time.Hour)
	defaultParallelLookups       = 100
	defaultMinTimeBetweenUpdates = config.Duration(5 * time.Minute)
	orderedQueueSize             = 10_000
)

func (*Lookup) SampleConfig() string {
	return sampleConfig
}

func (l *Lookup) Init() (err error) {
	// Check the SNMP configuration
	if _, err = snmp.NewWrapper(l.ClientConfig); err != nil {
		return fmt.Errorf("parsing SNMP client config: %w", err)
	}

	// Setup the GOSMI translator
	l.translator, err = si.NewGosmiTranslator(l.Path, l.Log)
	if err != nil {
		return fmt.Errorf("loading translator: %w", err)
	}

	// Preparing connection-builder function
	l.getConnectionFunc = l.getConnection

	// Initialize the table
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
	l.backlog = newBacklog(acc, l.Log, l.Ordered)

	cacheTTL := time.Duration(l.CacheTTL)
	minUpdateInterval := time.Duration(l.MinTimeBetweenUpdates)
	l.cache = newStore(l.CacheSize, cacheTTL, l.ParallelLookups, minUpdateInterval)
	l.cache.update = l.updateAgent
	l.cache.notify = l.backlog.resolve

	return nil
}

func (l *Lookup) Stop() {
	// Stop resolving
	l.cache.destroy()
	l.cache.purge()

	// Adding unresolved metrics to avoid data loss
	if n := l.backlog.destroy(); n > 0 {
		l.Log.Warnf("Added %d unresolved metrics due to processor stop!", n)
	}
}

func (l *Lookup) Add(m telegraf.Metric, acc telegraf.Accumulator) error {
	agent, found := m.GetTag(l.AgentTag)
	if !found {
		l.Log.Warn("Agent tag missing")
		acc.AddMetric(m)
		return nil
	}

	index, found := m.GetTag(l.IndexTag)
	if !found {
		l.Log.Warn("Index tag missing")
		acc.AddMetric(m)
		return nil
	}

	// Add the metric to the backlog before trying to resolve it
	l.backlog.push(agent, index, m)

	// Try to lookup the information from cache.
	l.cache.lookup(agent, index)

	return nil
}

// Default update function
func (l *Lookup) updateAgent(agent string) *tagMap {
	// Initialize connection to agent
	conn, err := l.getConnectionFunc(agent)
	if err != nil {
		l.Log.Errorf("Getting connection for %q failed: %v", agent, err)
		return nil
	}
	if err := conn.Reconnect(); err != nil {
		l.Log.Errorf("Connecting to %q failed:%v", agent, err)
		return nil
	}

	// Query table including translation
	table, err := l.table.Build(conn, true, l.translator)
	if err != nil {
		l.Log.Errorf("Building table for %q failed: %v", agent, err)
		return nil
	}

	// Copy tags for all rows
	tm := &tagMap{
		created: table.Time,
		rows:    make(tagMapRows, len(table.Rows)),
	}
	for _, row := range table.Rows {
		index := row.Tags["index"]
		delete(row.Tags, "index")
		tm.rows[index] = row.Tags
	}

	return tm
}

func (l *Lookup) getConnection(agent string) (snmpConnection, error) {
	conn, err := snmp.NewWrapper(l.ClientConfig)
	if err != nil {
		return conn, fmt.Errorf("parsing SNMP client config: %w", err)
	}

	if err := conn.SetAgent(agent); err != nil {
		return conn, fmt.Errorf("parsing agent tag: %w", err)
	}

	if err := conn.Connect(); err != nil {
		return conn, fmt.Errorf("connecting failed: %w", err)
	}

	return conn, nil
}

func init() {
	processors.AddStreaming("snmp_lookup", func() telegraf.StreamingProcessor {
		return &Lookup{
			AgentTag:              "source",
			IndexTag:              "index",
			ClientConfig:          *snmp.DefaultClientConfig(),
			CacheSize:             defaultCacheSize,
			CacheTTL:              defaultCacheTTL,
			MinTimeBetweenUpdates: defaultMinTimeBetweenUpdates,
			ParallelLookups:       defaultParallelLookups,
		}
	})
}
