//go:generate ../../../tools/readme_config_includer/generator
package snmp_lookup

import (
	_ "embed"
	"errors"
	"fmt"
	"sync"
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

	CacheSize       int             `toml:"max_cache_entries"`
	ParallelLookups int             `toml:"max_parallel_lookups"`
	Ordered         bool            `toml:"ordered"`
	CacheTTL        config.Duration `toml:"cache_ttl"`

	Log telegraf.Logger `toml:"-"`

	translator si.Translator
	table      si.Table

	acc               telegraf.Accumulator
	cache             *store
	backlog           *backlog
	getConnectionFunc func(string) (snmpConnection, error)

	sync.Mutex
}

const (
	defaultCacheSize       = 100
	defaultCacheTTL        = config.Duration(8 * time.Hour)
	defaultParallelLookups = 100
	minTimeBetweenUpdates  = 5 * time.Minute
	orderedQueueSize       = 10_000
)

func (*Lookup) SampleConfig() string {
	return sampleConfig
}

func (l *Lookup) Init() error {
	// Check the SNMP configuration
	if _, err := snmp.NewWrapper(l.ClientConfig); err != nil {
		return fmt.Errorf("parsing SNMP client config: %w", err)
	}

	// Setup the GOSMI translator
	translator, err := si.NewGosmiTranslator(l.Path, l.Log)
	if err != nil {
		return fmt.Errorf("loading translator: %w", err)
	}
	l.translator = translator

	// Initialize the table
	l.table.Name = "lookup"
	l.table.IndexAsTag = true
	l.table.Fields = make([]si.Field, len(l.Tags))
	for i, tag := range l.Tags {
		tag.IsTag = true
		l.table.Fields[i] = tag
	}

	// Preparing connection-builder function
	l.getConnectionFunc = l.getConnection

	return l.table.Init(l.translator)
}

func (l *Lookup) Start(acc telegraf.Accumulator) error {
	l.acc = acc
	l.backlog = newBacklog(acc, l.Log, l.Ordered)

	l.cache = newStore(l.CacheSize, time.Duration(l.CacheTTL), l.ParallelLookups)
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

func (l *Lookup) Add(m telegraf.Metric, _ telegraf.Accumulator) error {
	agent, found := m.GetTag(l.AgentTag)
	if !found {
		l.Log.Warn("Agent tag missing")
		l.acc.AddMetric(m)
		return nil
	}

	index, found := m.GetTag(l.IndexTag)
	if !found {
		l.Log.Warn("Index tag missing")
		l.acc.AddMetric(m)
		return nil
	}

	// Try to lookup the information from cache. An error ErrNotYetAvailable
	// indicates that the information is not yet cached, but the case will take
	// care to give back the information later via the `notify` callback.
	tags, err := l.cache.lookup(agent, index)
	if err != nil {
		if errors.Is(err, ErrNotYetAvailable) {
			l.Log.Debugf("Adding metric to backlog as data not yet available...")
			l.backlog.push(agent, index, m)
			return nil
		}
		l.Log.Errorf("Looking up %q (%s) failed: %v", agent, index, err)
		l.acc.AddMetric(m)
		return nil
	}

	// If resolving the metric from cache succeeded and we are good to directly
	// release the metrics, we will do so. For ordered cases it might be
	// necessary to add the metric to the backlog despite success...
	if l.Ordered && !l.backlog.empty() {
		// Add metric to backlog for later resolving
		l.Log.Debugf("Adding metric to backlog due to ordering constraints...")
		l.backlog.push(agent, index, m)
		return nil
	}

	l.Log.Debugf("Directly adding metric...")
	for key, value := range tags {
		m.AddTag(key, value)
	}
	l.acc.AddMetric(m)

	return nil
}

// Default update function
func (l *Lookup) updateAgent(agent string) *tagMap {
	// Initialize connection to agent
	l.Log.Debugf("Connecting to %q", agent)
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
	l.Log.Debugf("Building lookup table for %q", agent)
	table, err := l.table.Build(conn, true, l.translator)
	if err != nil {
		l.Log.Errorf("Building table for %q failed: %v", agent, err)
		return nil
	}

	// Copy tags for all rows
	l.Log.Debugf("Got table for %q: %v", agent, table.Rows)
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
			AgentTag:        "source",
			IndexTag:        "index",
			ClientConfig:    *snmp.DefaultClientConfig(),
			CacheSize:       defaultCacheSize,
			CacheTTL:        defaultCacheTTL,
			ParallelLookups: defaultParallelLookups,
		}
	})
}
