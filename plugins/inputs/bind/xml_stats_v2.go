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

// addCounter adds a v2Counter array to a Telegraf Accumulator, with the specified tags
func addCounter(acc telegraf.Accumulator, tags map[string]string, stats []v2Counter) {
	fields := make(map[string]interface{})

	for _, c := range stats {
		tags["name"] = c.Name
		fields["value"] = c.Value
		acc.AddCounter("bind_counter", fields, tags)
	}
}

// readStatsV2 decodes a BIND9 XML statistics version 2 document
func (b *Bind) readStatsV2(r io.Reader, acc telegraf.Accumulator, url string) error {
	var stats v2Root

	if err := xml.NewDecoder(r).Decode(&stats); err != nil {
		return fmt.Errorf("Unable to decode XML document: %s", err)
	}

	tags := map[string]string{"url": url}

	// Opcodes
	tags["type"] = "opcode"
	addCounter(acc, tags, stats.Statistics.Server.OpCodes)

	// Query RDATA types
	tags["type"] = "qtype"
	addCounter(acc, tags, stats.Statistics.Server.RdTypes)

	// Nameserver stats
	tags["type"] = "nsstat"
	addCounter(acc, tags, stats.Statistics.Server.NSStats)

	// Zone stats
	tags["type"] = "zonestat"
	addCounter(acc, tags, stats.Statistics.Server.ZoneStats)

	// Socket statistics
	tags["type"] = "sockstat"
	addCounter(acc, tags, stats.Statistics.Server.SockStats)

	// Memory stats
	fields := map[string]interface{}{
		"TotalUse":    stats.Statistics.Memory.Summary.TotalUse,
		"InUse":       stats.Statistics.Memory.Summary.InUse,
		"BlockSize":   stats.Statistics.Memory.Summary.BlockSize,
		"ContextSize": stats.Statistics.Memory.Summary.ContextSize,
		"Lost":        stats.Statistics.Memory.Summary.Lost,
	}
	acc.AddGauge("bind_memory", fields, map[string]string{"url": url})

	// Detailed, per-context memory stats
	if b.GatherMemoryContexts {
		tags := map[string]string{"url": url}

		for _, c := range stats.Statistics.Memory.Contexts {
			tags["id"] = c.Id
			tags["name"] = c.Name
			fields := map[string]interface{}{"Total": c.Total, "InUse": c.InUse}
			acc.AddGauge("bind_memory_context", fields, tags)
		}
	}

	// Detailed, per-view stats
	if b.GatherViews {
		for _, v := range stats.Statistics.Views {
			tags := map[string]string{"url": url, "view": v.Name}

			// Query RDATA types
			tags["type"] = "qtype"
			addCounter(acc, tags, v.RdTypes)

			// Resolver stats
			tags["type"] = "resstats"
			addCounter(acc, tags, v.ResStats)
		}
	}

	return nil
}
