package ifname

import (
	"fmt"
	"strconv"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal/snmp"
	si "github.com/influxdata/telegraf/plugins/inputs/snmp"
	"github.com/influxdata/telegraf/plugins/processors"
)

var sampleConfig = `
[[processors.ifname]]
  ## Name of tag holding the interface number
  # tag = "ifIndex"

  ## Name of output tag where service name will be added
  # dest = "ifName"

  ## Name of tag of the SNMP agent to request the interface name from
  # agent = "agent"

`

type IfName struct {
	SourceTag string `toml:"tag"`
	DestTag   string `toml:"dest"`
	AgentTag  string `toml:"agent"`

	CacheSize uint `toml:"max_cache_entries"`

	Log telegraf.Logger `toml:"-"`

	ifTable  *si.Table `toml:"-"`
	ifXTable *si.Table `toml:"-"`

	snmp.ClientConfig

	cache *LRUCache `toml:"-"`
}

type nameMap map[uint64]string

func (d *IfName) SampleConfig() string {
	return sampleConfig
}

func (d *IfName) Description() string {
	return "Add a tag of the network interface name looked up over SNMP by interface number"
}

func (h *IfName) Init() error {
	var err error
	h.ifTable, err = makeTable("IF-MIB::ifTable")
	if err != nil {
		return err
	}
	h.ifXTable, err = makeTable("IF-MIB::ifXTable")
	if err != nil {
		return err
	}

	c := NewLRUCache(h.CacheSize)
	h.cache = &c

	return nil
}

func (d *IfName) Apply(metrics ...telegraf.Metric) []telegraf.Metric {
	for _, metric := range metrics {
		agent, ok := metric.GetTag(d.AgentTag)
		if !ok {
			//agent tag missing
			continue
		}

		num_s, ok := metric.GetTag(d.SourceTag)
		if !ok {
			//source tag missing
			continue
		}

		num, err := strconv.ParseUint(num_s, 10, 64)
		if err != nil {
			//source tag not a uint
			d.Log.Infof("couldn't parse source tag as uint")
			continue
		}

		m, err := d.getMap(agent)
		if err != nil {
			d.Log.Infof("couldn't retrieve the table of interface names: %w", err)
		}

		name, found := (*m)[num]
		if !found {
			//interface num isn't in table
			d.Log.Infof("interface number %d isn't in the table of interface names", num)
			continue
		}
		metric.AddTag(d.DestTag, name)

	}
	return metrics
}

// getMap gets the interface names map either from cache or from the SNMP
// agent
func (d *IfName) getMap(agent string) (*nameMap, error) {
	m, ok := d.cache.Get(agent)
	if ok {
		return m, nil
	}

	var err error
	m, err = d.getMapRemote(agent)
	if err != nil {
		return nil, err
	}

	d.cache.Put(agent, m)
	return m, err
}

func (d *IfName) getMapRemote(agent string) (*nameMap, error) {
	gs, err := snmp.NewWrapper(d.ClientConfig, agent)
	if err != nil {
		//can't translate toml to gosnmp.GoSNMP
		return nil, err
	}

	err = gs.Connect()
	if err != nil {
		//can't connect (if tcp)
		return nil, err
	}

	//try ifXtable and ifName first.  if that fails, fall back to
	//ifTable and ifDescr
	var m *nameMap
	m, err = buildMap(&gs, d.ifXTable, "ifName")
	if err != nil {
		var err2 error
		m, err2 = buildMap(&gs, d.ifTable, "ifDescr")
		if err2 != nil {
			return nil, err2
		}
	}
	return m, nil
}

func init() {
	processors.Add("port_name", func() telegraf.Processor {
		return &IfName{
			SourceTag: "ifIndex",
			DestTag:   "ifName",
			AgentTag:  "agent",
			CacheSize: 1000,
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
		return nil, err
	}

	return &tab, nil
}

func buildMap(gs *snmp.GosnmpWrapper, tab *si.Table, column string) (*nameMap, error) {
	var err error

	rtab, err := tab.Build(gs, true)
	if err != nil {
		return nil, err
	}

	if len(rtab.Rows) == 0 {
		return nil, fmt.Errorf("empty table")
	}

	t := make(nameMap)
	for _, v := range rtab.Rows {
		//fmt.Printf("tags: %v, fields: %v\n", v.Tags, v.Fields)
		i_str, ok := v.Tags["index"]
		if !ok {
			//should always have an index tag because the table should
			//always have IndexAsTag true
			continue
		}
		i, err := strconv.ParseUint(i_str, 10, 64)
		if err != nil {
			//index value isn't a uint?
			continue
		}
		name_if, ok := v.Fields[column]
		if !ok {
			//column isn't present
			continue
		}
		name, ok := name_if.(string)
		if !ok {
			//column isn't a string
			continue
		}

		t[i] = name
	}
	return &t, nil
}
