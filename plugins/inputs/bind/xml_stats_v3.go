package bind

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"

	"github.com/influxdata/telegraf"
)

// Omitted branches: socketmgr, taskmgr
type v3Stats struct {
	Server struct {
		CounterGroups []v3Counters `xml:"counters"`
	} `xml:"server"`
	Views []struct {
		Name          string       `xml:"name,attr"`
		CounterGroups []v3Counters `xml:"counters"`
		Caches        []struct {
			Name   string `xml:"name,attr"`
			RRSets []struct {
				Name  string `xml:"name"`
				Value int    `xml:"counter"`
			} `xml:"rrset"`
		} `xml:"cache"`
	} `xml:"views>view"`
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

type v3Counters struct {
	Type     string `xml:"type,attr"`
	Counters []struct {
		Name  string `xml:"name,attr"`
		Value int    `xml:",chardata"`
	} `xml:"counter"`
}

// readStatsV3 decodes a BIND9 XML statistics version 3 document
func (b *Bind) readStatsV3(r io.Reader, acc telegraf.Accumulator) error {
	var stats v3Stats

	if err := xml.NewDecoder(r).Decode(&stats); err != nil {
		return fmt.Errorf("Unable to decode XML document: %s", err)
	}

	// Detailed, per-view stats
	if b.GatherViews {
		for _, v := range stats.Views {
			tags := map[string]string{"view": v.Name}
			fields := make(map[string]interface{})

			for _, cg := range v.CounterGroups {
				tags["type"] = cg.Type

				for _, c := range cg.Counters {
					tags["name"] = c.Name
					fields["value"] = c.Value
					acc.AddCounter("bind_counter", fields, tags)
				}
			}
		}
	}

	// Counter groups
	for _, cg := range stats.Server.CounterGroups {
		tags := map[string]string{"type": cg.Type}
		fields := make(map[string]interface{})

		for _, c := range cg.Counters {
			if cg.Type == "opcode" && strings.HasPrefix(c.Name, "RESERVED") {
				continue
			}

			tags["name"] = c.Name
			fields["value"] = c.Value
			acc.AddCounter("bind_counter", fields, tags)
		}
	}

	// Memory stats
	fields := map[string]interface{}{
		"TotalUse":    stats.Memory.Summary.TotalUse,
		"InUse":       stats.Memory.Summary.InUse,
		"BlockSize":   stats.Memory.Summary.BlockSize,
		"ContextSize": stats.Memory.Summary.ContextSize,
		"Lost":        stats.Memory.Summary.Lost,
	}
	acc.AddGauge("bind_memory", fields, nil)

	// Detailed, per-context memory stats
	if b.GatherMemoryContexts {
		for _, c := range stats.Memory.Contexts {
			tags := map[string]string{"id": c.Id, "name": c.Name}
			fields := map[string]interface{}{"Total": c.Total, "InUse": c.InUse}
			acc.AddGauge("bind_memory_context", fields, tags)
		}
	}

	return nil
}
