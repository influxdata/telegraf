package bind

import (
	"encoding/xml"
	"fmt"
	"io"

	"github.com/influxdata/telegraf"
)

type v2Root struct {
	XMLName    xml.Name
	Version    string       `xml:"version,attr"`
	Statistics v2Statistics `xml:"bind>statistics"`
}

type v2Statistics struct {
	Version string       `xml:"version,attr"`
	Memory  v2MemoryStat `xml:"memory>summary"`
	Server  v2Server     `xml:"server"`
	Views   []v2View     `xml:"views>view"`
}

type v2MemoryStat struct {
	TotalUse    int
	InUse       int
	BlockSize   int
	ContextSize int
	Lost        int
}

type v2Server struct {
	NSStats     []v2StatCounter `xml:"nsstat"`
	OpCodeStats []v2StatCounter `xml:"requests>opcode"`
	QueryStats  []v2StatCounter `xml:"queries-in>rdtype"`
	SockStats   []v2StatCounter `xml:"sockstat"`
	ZoneStats   []v2StatCounter `xml:"zonestat"`
}

type v2View struct {
	Name          string          `xml:"name"`
	QueryStats    []v2StatCounter `xml:"rdtype"`
	ResolverStats []v2StatCounter `xml:"resstat"`
}

type v2StatCounter struct {
	Name  string `xml:"name"`
	Value int    `xml:"counter"`
}

func makeFieldMap(stats []v2StatCounter) map[string]interface{} {
	fm := make(map[string]interface{})

	for _, st := range stats {
		fm[st.Name] = st.Value
	}

	return fm
}

// readStatsV2 decodes a BIND9 XML statistics version 2 document
func readStatsV2(r io.Reader, acc telegraf.Accumulator) error {
	var stats v2Root

	if err := xml.NewDecoder(r).Decode(&stats); err != nil {
		return fmt.Errorf("Unable to decode XML document: %s", err)
	}

	// Nameserver stats
	tags := map[string]string{}
	fields := makeFieldMap(stats.Statistics.Server.NSStats)
	acc.AddCounter("bind_server", fields, tags)

	// Opcodes
	fields = makeFieldMap(stats.Statistics.Server.OpCodeStats)
	acc.AddCounter("bind_opcodes", fields, tags)

	// Query types
	fields = makeFieldMap(stats.Statistics.Server.QueryStats)
	acc.AddCounter("bind_querytypes", fields, tags)

	// Socket statistics
	fields = makeFieldMap(stats.Statistics.Server.SockStats)
	acc.AddCounter("bind_sockstats", fields, tags)

	// Zone statistics
	tags = map[string]string{"zone": "_global"}
	fields = makeFieldMap(stats.Statistics.Server.ZoneStats)
	acc.AddCounter("bind_zonestats", fields, tags)

	return nil
}
