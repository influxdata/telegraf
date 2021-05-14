package bind

import (
	"encoding/xml"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
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
			ID    string `xml:"id"`
			Name  string `xml:"name"`
			Total int64  `xml:"total"`
			InUse int64  `xml:"inuse"`
		} `xml:"contexts>context"`
		Summary struct {
			TotalUse    int64
			InUse       int64
			BlockSize   int64
			ContextSize int64
			Lost        int64
		} `xml:"summary"`
	} `xml:"memory"`
}

// BIND statistics v2 counter struct used throughout
type v2Counter struct {
	Name  string `xml:"name"`
	Value int    `xml:"counter"`
}

// addXMLv2Counter adds a v2Counter array to a Telegraf Accumulator, with the specified tags
func addXMLv2Counter(acc telegraf.Accumulator, commonTags map[string]string, stats []v2Counter) {
	grouper := metric.NewSeriesGrouper()
	ts := time.Now()
	for _, c := range stats {
		tags := make(map[string]string)

		// Create local copy of tags since maps are reference types
		for k, v := range commonTags {
			tags[k] = v
		}

		if err := grouper.Add("bind_counter", tags, ts, c.Name, c.Value); err != nil {
			acc.AddError(fmt.Errorf("adding field %q to group failed: %v", c.Name, err))
		}
	}

	//Add grouped metrics
	for _, groupedMetric := range grouper.Metrics() {
		acc.AddMetric(groupedMetric)
	}
}

// readStatsXMLv2 decodes a BIND9 XML statistics version 2 document. Unlike the XML v3 statistics
// format, the v2 format does not support broken-out subsets.
func (b *Bind) readStatsXMLv2(addr *url.URL, acc telegraf.Accumulator) error {
	var stats v2Root

	resp, err := b.client.Get(addr.String())
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s returned HTTP status: %s", addr, resp.Status)
	}

	if err := xml.NewDecoder(resp.Body).Decode(&stats); err != nil {
		return fmt.Errorf("unable to decode XML document: %s", err)
	}

	tags := map[string]string{"url": addr.Host}
	host, port, _ := net.SplitHostPort(addr.Host)
	tags["source"] = host
	tags["port"] = port

	// Opcodes
	tags["type"] = "opcode"
	addXMLv2Counter(acc, tags, stats.Statistics.Server.OpCodes)

	// Query RDATA types
	tags["type"] = "qtype"
	addXMLv2Counter(acc, tags, stats.Statistics.Server.RdTypes)

	// Nameserver stats
	tags["type"] = "nsstat"
	addXMLv2Counter(acc, tags, stats.Statistics.Server.NSStats)

	// Zone stats
	tags["type"] = "zonestat"
	addXMLv2Counter(acc, tags, stats.Statistics.Server.ZoneStats)

	// Socket statistics
	tags["type"] = "sockstat"
	addXMLv2Counter(acc, tags, stats.Statistics.Server.SockStats)

	// Memory stats
	fields := map[string]interface{}{
		"total_use":    stats.Statistics.Memory.Summary.TotalUse,
		"in_use":       stats.Statistics.Memory.Summary.InUse,
		"block_size":   stats.Statistics.Memory.Summary.BlockSize,
		"context_size": stats.Statistics.Memory.Summary.ContextSize,
		"lost":         stats.Statistics.Memory.Summary.Lost,
	}
	acc.AddGauge("bind_memory", fields, map[string]string{"url": addr.Host, "source": host, "port": port})

	// Detailed, per-context memory stats
	if b.GatherMemoryContexts {
		for _, c := range stats.Statistics.Memory.Contexts {
			tags := map[string]string{"url": addr.Host, "id": c.ID, "name": c.Name, "source": host, "port": port}
			fields := map[string]interface{}{"total": c.Total, "in_use": c.InUse}

			acc.AddGauge("bind_memory_context", fields, tags)
		}
	}

	// Detailed, per-view stats
	if b.GatherViews {
		for _, v := range stats.Statistics.Views {
			tags := map[string]string{"url": addr.Host, "view": v.Name}

			// Query RDATA types
			tags["type"] = "qtype"
			addXMLv2Counter(acc, tags, v.RdTypes)

			// Resolver stats
			tags["type"] = "resstats"
			addXMLv2Counter(acc, tags, v.ResStats)
		}
	}

	return nil
}
