//go:generate ../../../tools/readme_config_includer/generator
package ifname

import (
	_ "embed"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal/snmp"
	"github.com/influxdata/telegraf/plugins/common/parallel"
	"github.com/influxdata/telegraf/plugins/processors"
)

//go:embed sample.conf
var sampleConfig string

type nameMap map[uint64]string
type keyType = string
type valType = nameMap

type mapFunc func(agent string) (nameMap, error)

type sigMap map[string]chan struct{}

type IfName struct {
	SourceTag string `toml:"tag"`
	DestTag   string `toml:"dest"`
	AgentTag  string `toml:"agent"`

	snmp.ClientConfig

	CacheSize          uint            `toml:"max_cache_entries"`
	MaxParallelLookups int             `toml:"max_parallel_lookups"`
	Ordered            bool            `toml:"ordered"`
	CacheTTL           config.Duration `toml:"cache_ttl"`

	Log telegraf.Logger `toml:"-"`

	ifTable  *snmp.Table
	ifXTable *snmp.Table

	cache    *TTLCache
	lock     sync.Mutex
	parallel parallel.Parallel
	sigs     sigMap

	getMapRemote mapFunc
}

const minRetry = 5 * time.Minute

func (*IfName) SampleConfig() string {
	return sampleConfig
}

func (d *IfName) Init() error {
	d.getMapRemote = d.getMapRemoteNoMock

	c := NewTTLCache(time.Duration(d.CacheTTL), d.CacheSize)
	d.cache = &c

	d.sigs = make(sigMap)

	if _, err := snmp.NewWrapper(d.ClientConfig); err != nil {
		return fmt.Errorf("parsing SNMP client config: %w", err)
	}

	return nil
}

func (d *IfName) addTag(metric telegraf.Metric) error {
	agent, ok := metric.GetTag(d.AgentTag)
	if !ok {
		d.Log.Warn("Agent tag missing.")
		return nil
	}

	numS, ok := metric.GetTag(d.SourceTag)
	if !ok {
		d.Log.Warn("Source tag missing.")
		return nil
	}

	num, err := strconv.ParseUint(numS, 10, 64)
	if err != nil {
		return errors.New("couldn't parse source tag as uint")
	}

	firstTime := true
	for {
		m, age, err := d.getMap(agent)
		if err != nil {
			return fmt.Errorf("couldn't retrieve the table of interface names for %s: %w", agent, err)
		}

		name, found := m[num]
		if found {
			// success
			metric.AddTag(d.DestTag, name)
			return nil
		}

		// We have the agent's interface map but it doesn't contain
		// the interface we're interested in.  If the entry is old
		// enough, retrieve it from the agent once more.
		if age < minRetry {
			return fmt.Errorf("interface number %d isn't in the table of interface names on %s", num, agent)
		}

		if firstTime {
			d.invalidate(agent)
			firstTime = false
			continue
		}

		// not found, cache hit, retrying
		return fmt.Errorf("missing interface but couldn't retrieve table for %v", agent)
	}
}

func (d *IfName) invalidate(agent string) {
	d.lock.Lock()
	d.cache.Delete(agent)
	d.lock.Unlock()
}

func (d *IfName) Start(acc telegraf.Accumulator) error {
	var err error

	d.ifTable, err = makeTable("1.3.6.1.2.1.2.2.1.2")
	if err != nil {
		return fmt.Errorf("preparing ifTable: %w", err)
	}
	d.ifXTable, err = makeTable("1.3.6.1.2.1.31.1.1.1.1")
	if err != nil {
		return fmt.Errorf("preparing ifXTable: %w", err)
	}

	fn := func(m telegraf.Metric) []telegraf.Metric {
		err := d.addTag(m)
		if err != nil {
			d.Log.Debugf("Error adding tag: %v", err)
		}
		return []telegraf.Metric{m}
	}

	if d.Ordered {
		d.parallel = parallel.NewOrdered(acc, fn, 10000, d.MaxParallelLookups)
	} else {
		d.parallel = parallel.NewUnordered(acc, fn, d.MaxParallelLookups)
	}
	return nil
}

func (d *IfName) Add(metric telegraf.Metric, _ telegraf.Accumulator) error {
	d.parallel.Enqueue(metric)
	return nil
}

func (d *IfName) Stop() {
	d.parallel.Stop()
}

// getMap gets the interface names map either from cache or from the SNMP
// agent
func (d *IfName) getMap(agent string) (entry nameMap, age time.Duration, err error) {
	var sig chan struct{}

	d.lock.Lock()

	// Check cache
	m, ok, age := d.cache.Get(agent)
	if ok {
		d.lock.Unlock()
		return m, age, nil
	}

	// cache miss.  Is this the first request for this agent?
	sig, found := d.sigs[agent]
	if !found {
		// This is the first request.  Make signal for subsequent requests to wait on
		s := make(chan struct{})
		d.sigs[agent] = s
		sig = s
	}

	d.lock.Unlock()

	if found {
		// This is not the first request.  Wait for first to finish.
		<-sig

		// Check cache again
		d.lock.Lock()
		m, ok, age := d.cache.Get(agent)
		d.lock.Unlock()
		if ok {
			return m, age, nil
		}
		return nil, 0, errors.New("getting remote table from cache")
	}

	// The cache missed and this is the first request for this
	// agent. Make the SNMP request
	m, err = d.getMapRemote(agent)

	d.lock.Lock()
	if err != nil {
		// snmp failure.  signal without saving to cache
		close(sig)
		delete(d.sigs, agent)

		d.lock.Unlock()
		return nil, 0, fmt.Errorf("getting remote table: %w", err)
	}

	// snmp success.  Cache response, then signal any other waiting
	// requests for this agent and clean up
	d.cache.Put(agent, m)
	close(sig)
	delete(d.sigs, agent)

	d.lock.Unlock()
	return m, 0, nil
}

func (d *IfName) getMapRemoteNoMock(agent string) (nameMap, error) {
	gs, err := snmp.NewWrapper(d.ClientConfig)
	if err != nil {
		return nil, fmt.Errorf("parsing SNMP client config: %w", err)
	}

	if err = gs.SetAgent(agent); err != nil {
		return nil, fmt.Errorf("parsing agent tag: %w", err)
	}

	if err = gs.Connect(); err != nil {
		return nil, fmt.Errorf("connecting when fetching interface names: %w", err)
	}

	// try ifXtable and ifName first.  if that fails, fall back to
	// ifTable and ifDescr
	var m nameMap
	if m, err = buildMap(gs, d.ifXTable); err == nil {
		return m, nil
	}

	if m, err = buildMap(gs, d.ifTable); err == nil {
		return m, nil
	}

	return nil, fmt.Errorf("fetching interface names: %w", err)
}

func init() {
	processors.AddStreaming("ifname", func() telegraf.StreamingProcessor {
		return &IfName{
			SourceTag:          "ifIndex",
			DestTag:            "ifName",
			AgentTag:           "agent",
			CacheSize:          100,
			MaxParallelLookups: 100,
			ClientConfig:       *snmp.DefaultClientConfig(),
			CacheTTL:           config.Duration(8 * time.Hour),
		}
	})
}

func makeTable(oid string) (*snmp.Table, error) {
	var err error
	tab := snmp.Table{
		Name:       "ifTable",
		IndexAsTag: true,
		Fields: []snmp.Field{
			{Oid: oid, Name: "ifName"},
		},
	}

	err = tab.Init(nil)
	if err != nil {
		// Init already wraps
		return nil, err
	}

	return &tab, nil
}

func buildMap(gs snmp.GosnmpWrapper, tab *snmp.Table) (nameMap, error) {
	var err error

	rtab, err := tab.Build(gs, true)
	if err != nil {
		// Build already wraps
		return nil, err
	}

	if len(rtab.Rows) == 0 {
		return nil, errors.New("empty table")
	}

	t := make(nameMap)
	for _, v := range rtab.Rows {
		iStr, ok := v.Tags["index"]
		if !ok {
			// should always have an index tag because the table should
			// always have IndexAsTag true
			return nil, errors.New("no index tag")
		}
		i, err := strconv.ParseUint(iStr, 10, 64)
		if err != nil {
			return nil, errors.New("index tag isn't a uint")
		}
		nameIf, ok := v.Fields["ifName"]
		if !ok {
			return nil, errors.New("ifName field is missing")
		}
		name, ok := nameIf.(string)
		if !ok {
			return nil, errors.New("ifName field isn't a string")
		}

		t[i] = name
	}
	return t, nil
}
