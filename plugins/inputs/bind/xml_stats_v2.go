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

func accumulate(acc telegraf.Accumulator, mSuffix string, tags map[string]string, stats []v2Counter) {
	measurement := fmt.Sprintf("bind_%s", mSuffix)
	fields := make(map[string]interface{})

	for _, st := range stats {
		tags[mSuffix] = st.Name
		fields["counter"] = st.Value
		acc.AddCounter(measurement, fields, tags)
	}
}

// readStatsV2 decodes a BIND9 XML statistics version 2 document
func (b *Bind) readStatsV2(r io.Reader, acc telegraf.Accumulator) error {
	var stats v2Root

	if err := xml.NewDecoder(r).Decode(&stats); err != nil {
		return fmt.Errorf("Unable to decode XML document: %s", err)
	}

	// Detailed, per-view stats
	if b.GatherViews {
		for _, v := range stats.Statistics.Views {
			// Query RDATA types
			accumulate(acc, "qtype", map[string]string{"view": v.Name}, v.RdTypes)

			// Resolver stats
			accumulate(acc, "resstats", map[string]string{"view": v.Name}, v.ResStats)
		}
	}

	// Opcodes
	accumulate(acc, "opcode", map[string]string{}, stats.Statistics.Server.OpCodes)

	// Query RDATA types
	accumulate(acc, "qtype", map[string]string{}, stats.Statistics.Server.RdTypes)

	// Nameserver stats
	accumulate(acc, "nsstat", map[string]string{}, stats.Statistics.Server.NSStats)

	// Zone stats
	accumulate(acc, "zonestat", map[string]string{}, stats.Statistics.Server.ZoneStats)

	// Socket statistics
	accumulate(acc, "sockstat", map[string]string{}, stats.Statistics.Server.SockStats)

	// Memory stats
	fields := map[string]interface{}{
		"TotalUse":    stats.Statistics.Memory.Summary.TotalUse,
		"InUse":       stats.Statistics.Memory.Summary.InUse,
		"BlockSize":   stats.Statistics.Memory.Summary.BlockSize,
		"ContextSize": stats.Statistics.Memory.Summary.ContextSize,
		"Lost":        stats.Statistics.Memory.Summary.Lost,
	}
	acc.AddGauge("bind_memory", fields, nil)

	// Detailed, per-context memory stats
	if b.GatherMemoryContexts {
		for _, c := range stats.Statistics.Memory.Contexts {
			tags := map[string]string{"id": c.Id, "name": c.Name}
			fields := map[string]interface{}{"Total": c.Total, "InUse": c.InUse}
			acc.AddGauge("bind_memory_context", fields, tags)
		}
	}

	return nil
}
