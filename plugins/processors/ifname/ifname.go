package ifname

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
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
  #ordered = false
`

type mapFunc func(agent string) (nameMap, error)

type IfName struct {
	SourceTag string `toml:"tag"`
	DestTag   string `toml:"dest"`
	AgentTag  string `toml:"agent"`

	snmp.ClientConfig

	CacheSize          uint `toml:"max_cache_entries"`
	MaxParallelLookups int  `toml:"max_parallel_lookups"`
	Ordered            bool `toml:"ordered"`

	Log telegraf.Logger `toml:"-"`

	ifTable  *si.Table `toml:"-"`
	ifXTable *si.Table `toml:"-"`

	rwLock sync.RWMutex `toml:"-"`
	cache  *LRUCache    `toml:"-"`

	parallel parallel.Parallel    `toml:"-"`
	acc      telegraf.Accumulator `toml:"-"`

	getMap mapFunc `toml:"-"`

	gsBase snmp.GosnmpWrapper `toml:"-"`
}

type nameMap map[uint64]string

func (d *IfName) SampleConfig() string {
	return sampleConfig
}

func (d *IfName) Description() string {
	return "Add a tag of the network interface name looked up over SNMP by interface number"
}

func (d *IfName) Init() error {
	d.getMap = d.getMapNoMock
	c := NewLRUCache(d.CacheSize)
	d.cache = &c

	return nil
}

func (d *IfName) addTag(metric telegraf.Metric) error {
	agent, ok := metric.GetTag(d.AgentTag)
	if !ok {
		//agent tag missing
		return nil
	}

	num_s, ok := metric.GetTag(d.SourceTag)
	if !ok {
		//source tag missing
		return nil
	}

	num, err := strconv.ParseUint(num_s, 10, 64)
	if err != nil {
		return fmt.Errorf("couldn't parse source tag as uint")
	}

	m, err := d.getMap(agent)
	if err != nil {
		return fmt.Errorf("couldn't retrieve the table of interface names: %w", err)
	}

	name, found := m[num]
	if !found {
		return fmt.Errorf("interface number %d isn't in the table of interface names", num)
	}
	metric.AddTag(d.DestTag, name)
	return nil
}

func (d *IfName) Start(acc telegraf.Accumulator) error {
	d.acc = acc

	var err error
	d.gsBase, err = snmp.NewWrapper(d.ClientConfig)
	if err != nil {
		return fmt.Errorf("parsing SNMP client config: %w", err)
	}

	d.ifTable, err = makeTable("IF-MIB::ifTable")
	if err != nil {
		return fmt.Errorf("looking up ifTable in local MIB: %w", err)
	}
	d.ifXTable, err = makeTable("IF-MIB::ifXTable")
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
func (d *IfName) getMapNoMock(agent string) (nameMap, error) {
	// Check cache first
	d.rwLock.RLock()
	m, ok := d.cache.Get(agent)
	d.rwLock.RUnlock()
	if ok {
		return m, nil
	}

	// The cache missed so make an SNMP request.  Write the response
	// to cache and return it
	d.rwLock.Lock()
	defer d.rwLock.Unlock()

	// Check cache again while holding write lock to prevent duplicate
	// SNMP requests.
	var err error
	m, ok = d.cache.Get(agent)
	if ok {
		return m, nil
	}

	m, err = d.getMapRemote(agent)
	if err != nil {
		return nil, fmt.Errorf("getting remote table: %w", err)
	}

	d.cache.Put(agent, m)
	return m, nil
}

func (d *IfName) getMapRemote(agent string) (nameMap, error) {
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
	m, err = buildMap(&gs, d.ifXTable, "ifName")
	if err == nil {
		return m, nil
	}

	m, err = buildMap(&gs, d.ifTable, "ifDescr")
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
		}
	})
}

func makeTable(tableName string) (*si.Table, error) {
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

func buildMap(gs *snmp.GosnmpWrapper, tab *si.Table, column string) (nameMap, error) {
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
			continue
		}
		i, err := strconv.ParseUint(i_str, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("index tag isn't a uint")
			continue
		}
		name_if, ok := v.Fields[column]
		if !ok {
			return nil, fmt.Errorf("field %s is missing", column)
			continue
		}
		name, ok := name_if.(string)
		if !ok {
			return nil, fmt.Errorf("field %s isn't a string", column)
			continue
		}

		t[i] = name
	}
	return t, nil
}
