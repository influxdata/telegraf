package ifname

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/internal/snmp"
	si "github.com/influxdata/telegraf/plugins/inputs/snmp"
	"github.com/influxdata/telegraf/plugins/processors"
	"github.com/influxdata/telegraf/plugins/processors/reverse_dns/parallel"
)

var sampleConfig = `
  ## Name of tag holding the interface number
  # tag = "ifIndex"

  ## Name of output tag where service name will be added
  # dest = "ifName"

  ## Name of tag of the SNMP agent to request the interface name from
  # agent = "agent"

  ## Timeout for each request.
  # timeout = "5s"

  ## SNMP version; can be 1, 2, or 3.
  # version = 2

  ## SNMP community string.
  # community = "public"

  ## Number of retries to attempt.
  # retries = 3

  ## The GETBULK max-repetitions parameter.
  # max_repetitions = 10

  ## SNMPv3 authentication and encryption options.
  ##
  ## Security Name.
  # sec_name = "myuser"
  ## Authentication protocol; one of "MD5", "SHA", or "".
  # auth_protocol = "MD5"
  ## Authentication password.
  # auth_password = "pass"
  ## Security Level; one of "noAuthNoPriv", "authNoPriv", or "authPriv".
  # sec_level = "authNoPriv"
  ## Context Name.
  # context_name = ""
  ## Privacy protocol used for encrypted messages; one of "DES", "AES" or "".
  # priv_protocol = ""
  ## Privacy password used for encrypted messages.
  # priv_password = ""

  ## max_parallel_lookups is the maximum number of SNMP requests to
  ## make at the same time.
  # max_parallel_lookups = 100

  ## ordered controls whether or not the metrics need to stay in the
  ## same order this plugin received them in. If false, this plugin
  ## may change the order when data is cached.  If you need metrics to
  ## stay in order set this to true.  keeping the metrics ordered may
  ## be slightly slower
  # ordered = false

  ## cache_ttl is the amount of time interface names are cached for a
  ## given agent.  After this period elapses if names are needed they
  ## will be retrieved again.
  # cache_ttl = "8h"
`

type nameMap map[uint64]string
type keyType = string
type valType = nameMap

type mapFunc func(agent string) (nameMap, error)
type makeTableFunc func(string) (*si.Table, error)

type sigMap map[string](chan struct{})

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

	ifTable  *si.Table `toml:"-"`
	ifXTable *si.Table `toml:"-"`

	rwLock sync.RWMutex `toml:"-"`
	cache  *TTLCache    `toml:"-"`

	parallel parallel.Parallel    `toml:"-"`
	acc      telegraf.Accumulator `toml:"-"`

	getMapRemote mapFunc       `toml:"-"`
	makeTable    makeTableFunc `toml:"-"`

	gsBase snmp.GosnmpWrapper `toml:"-"`

	sigs sigMap `toml:"-"`
}

const minRetry time.Duration = 5 * time.Minute

func (d *IfName) SampleConfig() string {
	return sampleConfig
}

func (d *IfName) Description() string {
	return "Add a tag of the network interface name looked up over SNMP by interface number"
}

func (d *IfName) Init() error {
	d.getMapRemote = d.getMapRemoteNoMock
	d.makeTable = makeTableNoMock

	c := NewTTLCache(time.Duration(d.CacheTTL), d.CacheSize)
	d.cache = &c

	d.sigs = make(sigMap)

	return nil
}

func (d *IfName) addTag(metric telegraf.Metric) error {
	agent, ok := metric.GetTag(d.AgentTag)
	if !ok {
		d.Log.Warn("Agent tag missing.")
		return nil
	}

	num_s, ok := metric.GetTag(d.SourceTag)
	if !ok {
		d.Log.Warn("Source tag missing.")
		return nil
	}

	num, err := strconv.ParseUint(num_s, 10, 64)
	if err != nil {
		return fmt.Errorf("couldn't parse source tag as uint")
	}

	firstTime := true
	for {
		m, age, err := d.getMap(agent)
		if err != nil {
			return fmt.Errorf("couldn't retrieve the table of interface names: %w", err)
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
			return fmt.Errorf("interface number %d isn't in the table of interface names", num)
		}

		if firstTime {
			d.invalidate(agent)
			firstTime = false
			continue
		}

		// not found, cache hit, retrying
		return fmt.Errorf("missing interface but couldn't retrieve table")
	}
}

func (d *IfName) invalidate(agent string) {
	d.rwLock.RLock()
	d.cache.Delete(agent)
	d.rwLock.RUnlock()
}

func (d *IfName) Start(acc telegraf.Accumulator) error {
	d.acc = acc

	var err error
	d.gsBase, err = snmp.NewWrapper(d.ClientConfig)
	if err != nil {
		return fmt.Errorf("parsing SNMP client config: %w", err)
	}

	d.ifTable, err = d.makeTable("IF-MIB::ifTable")
	if err != nil {
		return fmt.Errorf("looking up ifTable in local MIB: %w", err)
	}
	d.ifXTable, err = d.makeTable("IF-MIB::ifXTable")
	if err != nil {
		return fmt.Errorf("looking up ifXTable in local MIB: %w", err)
	}

	fn := func(m telegraf.Metric) []telegraf.Metric {
		err := d.addTag(m)
		if err != nil {
			d.Log.Debugf("Error adding tag %v", err)
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

func (d *IfName) Add(metric telegraf.Metric, acc telegraf.Accumulator) error {
	d.parallel.Enqueue(metric)
	return nil
}

func (d *IfName) Stop() error {
	d.parallel.Stop()
	return nil
}

// getMap gets the interface names map either from cache or from the SNMP
// agent
func (d *IfName) getMap(agent string) (entry nameMap, age time.Duration, err error) {
	var sig chan struct{}

	// Check cache
	d.rwLock.RLock()
	m, ok, age := d.cache.Get(agent)
	d.rwLock.RUnlock()
	if ok {
		return m, age, nil
	}

	// Is this the first request for this agent?
	d.rwLock.Lock()
	sig, found := d.sigs[agent]
	if !found {
		s := make(chan struct{})
		d.sigs[agent] = s
		sig = s
	}
	d.rwLock.Unlock()

	if found {
		// This is not the first request.  Wait for first to finish.
		<-sig
		// Check cache again
		d.rwLock.RLock()
		m, ok, age := d.cache.Get(agent)
		d.rwLock.RUnlock()
		if ok {
			return m, age, nil
		} else {
			return nil, 0, fmt.Errorf("getting remote table from cache")
		}
	}

	// The cache missed and this is the first request for this
	// agent.

	// Make the SNMP request
	m, err = d.getMapRemote(agent)
	if err != nil {
		//failure.  signal without saving to cache
		d.rwLock.Lock()
		close(sig)
		delete(d.sigs, agent)
		d.rwLock.Unlock()

		return nil, 0, fmt.Errorf("getting remote table: %w", err)
	}

	// Cache it, then signal any other waiting requests for this agent
	// and clean up
	d.rwLock.Lock()
	d.cache.Put(agent, m)
	close(sig)
	delete(d.sigs, agent)
	d.rwLock.Unlock()

	return m, 0, nil
}

func (d *IfName) getMapRemoteNoMock(agent string) (nameMap, error) {
	gs := d.gsBase
	err := gs.SetAgent(agent)
	if err != nil {
		return nil, fmt.Errorf("parsing agent tag: %w", err)
	}

	err = gs.Connect()
	if err != nil {
		return nil, fmt.Errorf("connecting when fetching interface names: %w", err)
	}

	//try ifXtable and ifName first.  if that fails, fall back to
	//ifTable and ifDescr
	var m nameMap
	m, err = buildMap(gs, d.ifXTable, "ifName")
	if err == nil {
		return m, nil
	}

	m, err = buildMap(gs, d.ifTable, "ifDescr")
	if err == nil {
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
			ClientConfig: snmp.ClientConfig{
				Retries:        3,
				MaxRepetitions: 10,
				Timeout:        internal.Duration{Duration: 5 * time.Second},
				Version:        2,
				Community:      "public",
			},
			CacheTTL: config.Duration(8 * time.Hour),
		}
	})
}

func makeTableNoMock(tableName string) (*si.Table, error) {
	var err error
	tab := si.Table{
		Oid:        tableName,
		IndexAsTag: true,
	}

	err = tab.Init()
	if err != nil {
		//Init already wraps
		return nil, err
	}

	return &tab, nil
}

func buildMap(gs snmp.GosnmpWrapper, tab *si.Table, column string) (nameMap, error) {
	var err error

	rtab, err := tab.Build(gs, true)
	if err != nil {
		//Build already wraps
		return nil, err
	}

	if len(rtab.Rows) == 0 {
		return nil, fmt.Errorf("empty table")
	}

	t := make(nameMap)
	for _, v := range rtab.Rows {
		i_str, ok := v.Tags["index"]
		if !ok {
			//should always have an index tag because the table should
			//always have IndexAsTag true
			return nil, fmt.Errorf("no index tag")
		}
		i, err := strconv.ParseUint(i_str, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("index tag isn't a uint")
		}
		name_if, ok := v.Fields[column]
		if !ok {
			return nil, fmt.Errorf("field %s is missing", column)
		}
		name, ok := name_if.(string)
		if !ok {
			return nil, fmt.Errorf("field %s isn't a string", column)
		}

		t[i] = name
	}
	return t, nil
}
