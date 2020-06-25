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

	Log telegraf.Logger `toml:"-"`

	ifTable  *si.Table `toml:"-"`
	ifXTable *si.Table `toml:"-"`

	snmp.ClientConfig
}

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

	return nil
}

func (d *IfName) Apply(metrics ...telegraf.Metric) []telegraf.Metric {
	var out []telegraf.Metric
	for _, metric := range metrics {
		out = append(out, metric)

		num_s, ok := metric.GetTag(d.SourceTag)
		if !ok {
			//source tag missing
			continue
		}

		num, err := strconv.ParseUint(num_s, 10, 64)
		if err != nil {
			//source tag not a uint
			continue
		}

		agent, ok := metric.GetTag(d.AgentTag)
		if !ok {
			//agent tag missing
			continue
		}

		//todo: cache gs by agent
		gs, err := snmp.NewWrapper(d.ClientConfig, agent)
		if err != nil {
			//can't translate toml to gosnmp.GoSNMP
			continue
		}

		err = gs.Connect()
		if err != nil {
			//can't connect (if tcp)
			continue
		}

		var m map[uint64]string

		//try ifXtable and ifName first.  if that fails, fall back to
		//ifTable and ifDescr
		m, err = buildMap(&gs, d.ifXTable, "ifName")
		if err != nil {
			m, err = buildMap(&gs, d.ifTable, "ifDescr")
			if err != nil {
				//couldn't get either table
				continue
			}
		}

		name, found := m[num]
		if !found {
			//interface num isn't in table
			continue
		}
		metric.AddTag(d.DestTag, name)

	}

	return out
}

func init() {
	processors.Add("port_name", func() telegraf.Processor {
		return &IfName{
			SourceTag: "ifIndex",
			DestTag:   "ifName",
			AgentTag:  "agent",
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

func buildMap(gs *snmp.GosnmpWrapper, tab *si.Table, column string) (map[uint64]string, error) {
	var err error

	rtab, err := tab.Build(gs, true)
	if err != nil {
		return nil, err
	}

	if len(rtab.Rows) == 0 {
		return nil, fmt.Errorf("empty table")
	}

	t := make(map[uint64]string)
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
	return t, nil
}
