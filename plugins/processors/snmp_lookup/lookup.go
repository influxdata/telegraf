//go:generate ../../../tools/readme_config_includer/generator
package snmp_lookup

import (
	_ "embed"
	"fmt"
	"strconv"
	"sync"
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

// snmpConnection is an interface which wraps a *gosnmp.GoSNMP object.
// We interact through an interface so we can mock it out in tests.
type snmpConnection interface {
	Host() string
	Walk(string, gosnmp.WalkFunc) error
	Get(oids []string) (*gosnmp.SnmpPacket, error)
	Reconnect() error
}

type getConnectionFunc func(metric telegraf.Metric) (snmpConnection, error)
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

	VersionTag      string `toml:"version_tag"`
	CommunityTag    string `toml:"community_tag"`
	SecNameTag      string `toml:"sec_name_tag"`
	SecLevelTag     string `toml:"sec_level_tag"`
	AuthProtocolTag string `toml:"auth_protocol_tag"`
	AuthPasswordTag string `toml:"auth_password_tag"`
	PrivProtocolTag string `toml:"priv_protocol_tag"`
	PrivPasswordTag string `toml:"priv_password_tag"`
	ContextNameTag  string `toml:"context_name_tag"`

	CacheSize       int             `toml:"max_cache_entries"`
	ParallelLookups int             `toml:"max_parallel_lookups"`
	Ordered         bool            `toml:"ordered"`
	CacheTTL        config.Duration `toml:"cache_ttl"`

	Log telegraf.Logger `toml:"-"`

	cache    *expirable.LRU[string, tagMap]
	parallel parallel.Parallel
	sigs     signalMap
	lock     sync.Mutex
	table    si.Table

	getConnectionFunc getConnectionFunc

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

func (l *Lookup) Init() (err error) {
	l.sigs = make(signalMap)
	l.getConnectionFunc = l.getConnection

	if _, err = snmp.NewWrapper(l.ClientConfig); err != nil {
		return fmt.Errorf("parsing SNMP client config: %w", err)
	}

	switch l.Translator {
	case "", "gosmi":
		if l.translator, err = si.NewGosmiTranslator(l.Path, l.Log); err != nil {
			return fmt.Errorf("loading translator: %w", err)
		}
	default:
		return fmt.Errorf("invalid translator %q", l.Translator)
	}

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
	if !metric.HasTag(l.AgentTag) {
		l.Log.Warn("Agent tag missing")
		return []telegraf.Metric{metric}
	}

	index, ok := metric.GetTag(l.IndexTag)
	if !ok {
		l.Log.Warn("Index tag missing")
		return []telegraf.Metric{metric}
	}

	gs, err := l.getConnectionFunc(metric)
	if err != nil {
		l.Log.Errorf("Could not prepare connection: %v", err)
		return []telegraf.Metric{metric}
	}

	// Prepare cache
	if err := l.prepareCache(gs, index); err != nil {
		l.Log.Warnf("Could not prepare cache for %q: %v", gs.Host(), err)
		return []telegraf.Metric{metric}
	}

	// Load from cache
	tagMap, inCache := l.cache.Get(gs.Host())
	tags, indexExists := tagMap.rows[index]
	if inCache && indexExists {
		for key, value := range tags {
			metric.AddTag(key, value)
		}
	} else {
		l.Log.Warnf("Could not find index %q on agent %q", index, gs.Host())
	}

	return []telegraf.Metric{metric}
}

// prepareCache prepares the cache if needed (index does not exist yet, or agent not yet cached)
func (l *Lookup) prepareCache(gs snmpConnection, index string) error {
	agent := gs.Host()

	// Check cache
	l.lock.Lock()
	tagMap, inCache := l.cache.Peek(agent)
	_, indexExists := tagMap.rows[index]

	// Cache miss or non existing index and not recently refreshed the table
	if !inCache || (!indexExists && time.Since(tagMap.created) > minRetry) {
		// Check if another process is alreay loading the table, wait until done if so
		if done, busy := l.sigs[agent]; busy {
			l.lock.Unlock()
			<-done
		} else {
			// No other process is already loading the table, let others know by creating a channel
			l.sigs[agent] = make(chan struct{})
			l.lock.Unlock()

			if err := gs.Reconnect(); err != nil {
				l.signalAgentReady(agent)
				return fmt.Errorf("could not connect: %w", err)
			}

			// build the table for the configured tags
			if err := l.loadTagMap(gs); err != nil {
				l.signalAgentReady(agent)
				return fmt.Errorf("could not load table: %w", err)
			}

			// Done loading, inform other processes
			l.signalAgentReady(agent)
		}
	} else {
		l.lock.Unlock()
	}

	return nil
}

func (l *Lookup) signalAgentReady(agent string) {
	l.lock.Lock()
	close(l.sigs[agent])
	delete(l.sigs, agent)
	l.lock.Unlock()
}

// getConnection prepares a snmpConnection from the given metric tags (if present)
func (l *Lookup) getConnection(metric telegraf.Metric) (snmpConnection, error) {
	clientConfig := l.ClientConfig

	if version, ok := metric.GetTag(l.VersionTag); ok {
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

	if community, ok := metric.GetTag(l.CommunityTag); ok {
		clientConfig.Community = community
	}

	if secName, ok := metric.GetTag(l.SecNameTag); ok {
		clientConfig.SecName = secName
	}

	if secLevel, ok := metric.GetTag(l.SecLevelTag); ok {
		clientConfig.SecLevel = secLevel
	}

	if authProtocol, ok := metric.GetTag(l.AuthProtocolTag); ok {
		clientConfig.AuthProtocol = authProtocol
	}

	if authPassword, ok := metric.GetTag(l.AuthPasswordTag); ok {
		clientConfig.AuthPassword = authPassword
	}

	if privProtocol, ok := metric.GetTag(l.PrivProtocolTag); ok {
		clientConfig.PrivProtocol = privProtocol
	}

	if privPassword, ok := metric.GetTag(l.PrivPasswordTag); ok {
		clientConfig.PrivPassword = privPassword
	}

	if contextName, ok := metric.GetTag(l.ContextNameTag); ok {
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

// loadTagMap gathers the configured table from the snmp agent and
// stores all tags into the cache
func (l *Lookup) loadTagMap(gs snmpConnection) error {
	agent := gs.Host()
	l.Log.Debugf("Building lookup table for %q", agent)
	table, err := l.table.Build(gs, true, l.translator)
	if err != nil {
		return err
	}

	tagMap := tagMap{
		created: table.Time,
		rows:    make(tagMapRows, len(table.Rows)),
	}

	// Copy tags for all rows in the tagMap
	for _, row := range table.Rows {
		index := row.Tags["index"]
		delete(row.Tags, "index")
		tagMap.rows[index] = row.Tags
	}

	// Add the found tags for all indexes on this agent to the cache
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
			ClientConfig:    *snmp.DefaultClientConfig(),
			VersionTag:      "version",
			CommunityTag:    "community",
			SecNameTag:      "sec_name",
			SecLevelTag:     "sec_level",
			AuthProtocolTag: "auth_protocol",
			AuthPasswordTag: "auth_password",
			PrivProtocolTag: "priv_protocol",
			PrivPasswordTag: "priv_password",
			ContextNameTag:  "context_name",
			CacheSize:       defaultCacheSize,
			CacheTTL:        defaultCacheTTL,
			ParallelLookups: defaultParallelLookups,
		}
	})
}
