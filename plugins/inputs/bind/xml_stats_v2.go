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

// Omitted branches: socketmgr, taskmgr
type v2Statistics struct {
	Version string `xml:"version,attr"`
	Views   []struct {
		// Omitted branches: zones
		Name     string      `xml:"name"`
		RdTypes  []v2Counter `xml:"rdtype"`
		ResStats []v2Counter `xml:"resstat"`
		Caches   []struct {
			Name   string      `xml:"name,attr"`
			RRSets []v2Counter `xml:"rrset"`
		} `xml:"cache"`
	} `xml:"views>view"`
	Server struct {
		OpCodes   []v2Counter `xml:"requests>opcode"`
		RdTypes   []v2Counter `xml:"queries-in>rdtype"`
		NSStats   []v2Counter `xml:"nsstat"`
		ZoneStats []v2Counter `xml:"zonestat"`
		ResStats  []v2Counter `xml:"resstat"`
		SockStats []v2Counter `xml:"sockstat"`
	} `xml:"server"`
	Memory struct {
		Contexts []struct {
			// Omitted nodes: references, maxinuse, blocksize, pools, hiwater, lowater
			Id    string `xml:"id"`
			Name  string `xml:"name"`
			Total int    `xml:"total"`
			InUse int    `xml:"inuse"`
		} `xml:"contexts>context"`
		Summary struct {
			TotalUse    int
			InUse       int
			BlockSize   int
			ContextSize int
			Lost        int
		} `xml:"summary"`
	} `xml:"memory"`
}

// BIND statistics v2 counter struct used throughout
type v2Counter struct {
	Name  string `xml:"name"`
	Value int    `xml:"counter"`
}

func makeFieldMap(stats []v2Counter) map[string]interface{} {
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

	// Memory stats
	tags := map[string]string{}
	fields := map[string]interface{}{
		"TotalUse":    stats.Statistics.Memory.Summary.TotalUse,
		"InUse":       stats.Statistics.Memory.Summary.InUse,
		"BlockSize":   stats.Statistics.Memory.Summary.BlockSize,
		"ContextSize": stats.Statistics.Memory.Summary.ContextSize,
		"Lost":        stats.Statistics.Memory.Summary.Lost,
	}
	acc.AddCounter("bind_memory", fields, tags)

	// Nameserver stats
	fields = makeFieldMap(stats.Statistics.Server.NSStats)
	acc.AddCounter("bind_server", fields, tags)

	// Opcodes
	fields = makeFieldMap(stats.Statistics.Server.OpCodes)
	acc.AddCounter("bind_opcodes", fields, tags)

	// RDATA types
	fields = makeFieldMap(stats.Statistics.Server.RdTypes)
	acc.AddCounter("bind_rdtypes", fields, tags)

	// Socket statistics
	fields = makeFieldMap(stats.Statistics.Server.SockStats)
	acc.AddCounter("bind_sockstats", fields, tags)

	// Zone statistics
	tags = map[string]string{"zone": "_global"}
	fields = makeFieldMap(stats.Statistics.Server.ZoneStats)
	acc.AddCounter("bind_zonestats", fields, tags)

	return nil
}
