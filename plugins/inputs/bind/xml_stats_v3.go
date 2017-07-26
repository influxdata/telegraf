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
		// Omitted branches: zones
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
func (b *Bind) readStatsV3(r io.Reader, acc telegraf.Accumulator, url string) error {
	var stats v3Stats

	if err := xml.NewDecoder(r).Decode(&stats); err != nil {
		return fmt.Errorf("Unable to decode XML document: %s", err)
	}

	// Counter groups
	for _, cg := range stats.Server.CounterGroups {
		for _, c := range cg.Counters {
			if cg.Type == "opcode" && strings.HasPrefix(c.Name, "RESERVED") {
				continue
			}

			tags := map[string]string{"url": url, "type": cg.Type, "name": c.Name}
			fields := map[string]interface{}{"value": c.Value}

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
	acc.AddGauge("bind_memory", fields, map[string]string{"url": url})

	// Detailed, per-context memory stats
	if b.GatherMemoryContexts {
		for _, c := range stats.Memory.Contexts {
			tags := map[string]string{"url": url, "id": c.Id, "name": c.Name}
			fields := map[string]interface{}{"Total": c.Total, "InUse": c.InUse}

			acc.AddGauge("bind_memory_context", fields, tags)
		}
	}

	// Detailed, per-view stats
	if b.GatherViews {
		for _, v := range stats.Views {
			for _, cg := range v.CounterGroups {
				for _, c := range cg.Counters {
					tags := map[string]string{
						"url":  url,
						"view": v.Name,
						"type": cg.Type,
						"name": c.Name,
					}
					fields := map[string]interface{}{"value": c.Value}

					acc.AddCounter("bind_counter", fields, tags)
				}
			}
		}
	}

	return nil
}
