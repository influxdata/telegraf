//go:generate ../../../tools/readme_config_includer/generator
package snmp_lookup

import (
	_ "embed"
	"fmt"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/snmp"
	"github.com/influxdata/telegraf/plugins/processors"
)

//go:embed sample.conf
var sampleConfig string

const (
	defaultCacheSize             = 100
	defaultCacheTTL              = config.Duration(8 * time.Hour)
	defaultParallelLookups       = 16
	defaultMinTimeBetweenUpdates = config.Duration(5 * time.Minute)
)

type SNMPLookup struct {
	AgentTag string       `toml:"agent_tag"`
	IndexTag string       `toml:"index_tag"`
	Tags     []snmp.Field `toml:"tag"`

	snmp.ClientConfig

	CacheSize             int             `toml:"max_cache_entries"`
	ParallelLookups       int             `toml:"max_parallel_lookups"`
	Ordered               bool            `toml:"ordered"`
	CacheTTL              config.Duration `toml:"cache_ttl"`
	MinTimeBetweenUpdates config.Duration `toml:"min_time_between_updates"`

	Log telegraf.Logger `toml:"-"`

	table             snmp.Table
	cache             *store
	backlog           *backlog
	getConnectionFunc func(string) (snmp.Connection, error)
}

type tagMapRows map[string]map[string]string

type tagMap struct {
	created time.Time
	rows    tagMapRows
}

func (*SNMPLookup) SampleConfig() string {
	return sampleConfig
}

func (l *SNMPLookup) Init() (err error) {
	// Check the SNMP configuration
	l.GosnmpDebugLogger = l.Log
	if _, err = snmp.NewWrapper(l.ClientConfig); err != nil {
		return fmt.Errorf("parsing SNMP client config: %w", err)
	}

	// Setup the GOSMI translator
	translator, err := snmp.NewGosmiTranslator(l.Path, l.Log)
	if err != nil {
		return fmt.Errorf("loading translator: %w", err)
	}

	// Preparing connection-builder function
	l.getConnectionFunc = l.getConnection

	// Initialize the table
	l.table.Name = "lookup"
	l.table.IndexAsTag = true
	l.table.Fields = l.Tags
	for i := range l.table.Fields {
		l.table.Fields[i].IsTag = true
	}

	return l.table.Init(translator)
}

func (l *SNMPLookup) Start(acc telegraf.Accumulator) error {
	l.backlog = newBacklog(acc, l.Log, l.Ordered)

	l.cache = newStore(l.CacheSize, l.CacheTTL, l.ParallelLookups, l.MinTimeBetweenUpdates)
	l.cache.update = l.updateAgent
	l.cache.notify = l.backlog.resolve

	return nil
}

func (l *SNMPLookup) Add(m telegraf.Metric, acc telegraf.Accumulator) error {
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

func (l *SNMPLookup) Stop() {
	// Stop resolving
	l.cache.destroy()
	l.cache.purge()

	// Adding unresolved metrics to avoid data loss
	if n := l.backlog.destroy(); n > 0 {
		l.Log.Warnf("Added %d unresolved metrics due to processor stop!", n)
	}
}

// Default update function
func (l *SNMPLookup) updateAgent(agent string) *tagMap {
	tm := &tagMap{created: time.Now()}

	// Initialize connection to agent
	conn, err := l.getConnectionFunc(agent)
	if err != nil {
		l.Log.Errorf("Getting connection for %q failed: %v", agent, err)
		return tm
	}

	// Query table including translation
	table, err := l.table.Build(conn, true)
	if err != nil {
		l.Log.Errorf("Building table for %q failed: %v", agent, err)
		return tm
	}

	// Copy tags for all rows
	tm.created = table.Time
	tm.rows = make(tagMapRows, len(table.Rows))
	for _, row := range table.Rows {
		index := row.Tags["index"]
		delete(row.Tags, "index")
		tm.rows[index] = row.Tags
	}

	return tm
}

func (l *SNMPLookup) getConnection(agent string) (snmp.Connection, error) {
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
		return &SNMPLookup{
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
