package bind

import (
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/influxdata/telegraf"
)

// XML path: //statistics
// Omitted branches: socketmgr, taskmgr
type v3Stats struct {
	Server v3Server `xml:"server"`
	Views  []v3View `xml:"views>view"`
	Memory v3Memory `xml:"memory"`
}

// XML path: //statistics/memory
type v3Memory struct {
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
}

// XML path: //statistics/server
type v3Server struct {
	CounterGroups []v3CounterGroup `xml:"counters"`
}

// XML path: //statistics/views/view
type v3View struct {
	// Omitted branches: zones
	Name          string           `xml:"name,attr"`
	CounterGroups []v3CounterGroup `xml:"counters"`
	Caches        []struct {
		Name   string `xml:"name,attr"`
		RRSets []struct {
			Name  string `xml:"name"`
			Value int    `xml:"counter"`
		} `xml:"rrset"`
	} `xml:"cache"`
}

// Generic XML v3 doc fragment used in multiple places
type v3CounterGroup struct {
	Type     string `xml:"type,attr"`
	Counters []struct {
		Name  string `xml:"name,attr"`
		Value int    `xml:",chardata"`
	} `xml:"counter"`
}

// addStatsXMLv3 walks a v3Stats struct and adds the values to the telegraf.Accumulator.
func (b *Bind) addStatsXMLv3(stats v3Stats, acc telegraf.Accumulator, urlTag string) {
	// Counter groups
	for _, cg := range stats.Server.CounterGroups {
		for _, c := range cg.Counters {
			if cg.Type == "opcode" && strings.HasPrefix(c.Name, "RESERVED") {
				continue
			}

			tags := map[string]string{"url": urlTag, "type": cg.Type, "name": c.Name}
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
	acc.AddGauge("bind_memory", fields, map[string]string{"url": urlTag})

	// Detailed, per-context memory stats
	if b.GatherMemoryContexts {
		for _, c := range stats.Memory.Contexts {
			tags := map[string]string{"url": urlTag, "id": c.Id, "name": c.Name}
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
						"url":  urlTag,
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
}

// readStatsXMLv3 takes a base URL to probe, and requests the individual statistics documents that
// we are interested in. These individual documents have a combined size which is significantly
// smaller than if we requested everything at once (e.g. taskmgr and socketmgr can be omitted).
func (b *Bind) readStatsXMLv3(addr *url.URL, acc telegraf.Accumulator) error {
	var stats v3Stats

	// Progressively build up full v3Stats struct by parsing the individual HTTP responses
	for _, suffix := range [...]string{"/server", "/net", "/mem"} {
		scrapeUrl := addr.String() + suffix

		resp, err := client.Get(scrapeUrl)
		if err != nil {
			return err
		}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("%s returned HTTP status: %s", scrapeUrl, resp.Status)
		}

		log.Printf("D! HTTP response content length: %d", resp.ContentLength)

		if err := xml.NewDecoder(resp.Body).Decode(&stats); err != nil {
			return fmt.Errorf("Unable to decode XML document: %s", err)
		}
	}

	b.addStatsXMLv3(stats, acc, addr.Host)
	return nil
}
