package bind

import (
	"encoding/xml"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
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
			Value int64  `xml:"counter"`
		} `xml:"rrset"`
	} `xml:"cache"`
}

// Generic XML v3 doc fragment used in multiple places
type v3CounterGroup struct {
	Type     string `xml:"type,attr"`
	Counters []struct {
		Name  string `xml:"name,attr"`
		Value int64  `xml:",chardata"`
	} `xml:"counter"`
}

// addStatsXMLv3 walks a v3Stats struct and adds the values to the telegraf.Accumulator.
func (b *Bind) addStatsXMLv3(stats v3Stats, acc telegraf.Accumulator, hostPort string) {
	grouper := metric.NewSeriesGrouper()
	ts := time.Now()
	host, port, _ := net.SplitHostPort(hostPort)
	// Counter groups
	for _, cg := range stats.Server.CounterGroups {
		for _, c := range cg.Counters {
			if cg.Type == "opcode" && strings.HasPrefix(c.Name, "RESERVED") {
				continue
			}

			tags := map[string]string{"url": hostPort, "source": host, "port": port, "type": cg.Type}

			if err := grouper.Add("bind_counter", tags, ts, c.Name, c.Value); err != nil {
				acc.AddError(fmt.Errorf("adding tags %q to group failed: %v", tags, err))
			}
		}
	}

	// Memory stats
	fields := map[string]interface{}{
		"total_use":    stats.Memory.Summary.TotalUse,
		"in_use":       stats.Memory.Summary.InUse,
		"block_size":   stats.Memory.Summary.BlockSize,
		"context_size": stats.Memory.Summary.ContextSize,
		"lost":         stats.Memory.Summary.Lost,
	}
	acc.AddGauge("bind_memory", fields, map[string]string{"url": hostPort, "source": host, "port": port})

	// Detailed, per-context memory stats
	if b.GatherMemoryContexts {
		for _, c := range stats.Memory.Contexts {
			tags := map[string]string{"url": hostPort, "source": host, "port": port, "id": c.ID, "name": c.Name}
			fields := map[string]interface{}{"total": c.Total, "in_use": c.InUse}

			acc.AddGauge("bind_memory_context", fields, tags)
		}
	}

	// Detailed, per-view stats
	if b.GatherViews {
		for _, v := range stats.Views {
			for _, cg := range v.CounterGroups {
				for _, c := range cg.Counters {
					tags := map[string]string{
						"url":    hostPort,
						"source": host,
						"port":   port,
						"view":   v.Name,
						"type":   cg.Type,
					}

					if err := grouper.Add("bind_counter", tags, ts, c.Name, c.Value); err != nil {
						acc.AddError(fmt.Errorf("adding tags %q to group failed: %v", tags, err))
					}
				}
			}
		}
	}

	//Add grouped metrics
	for _, groupedMetric := range grouper.Metrics() {
		acc.AddMetric(groupedMetric)
	}
}

// readStatsXMLv3 takes a base URL to probe, and requests the individual statistics documents that
// we are interested in. These individual documents have a combined size which is significantly
// smaller than if we requested everything at once (e.g. taskmgr and socketmgr can be omitted).
func (b *Bind) readStatsXMLv3(addr *url.URL, acc telegraf.Accumulator) error {
	var stats v3Stats

	// Progressively build up full v3Stats struct by parsing the individual HTTP responses
	for _, suffix := range [...]string{"/server", "/net", "/mem"} {
		err := func() error {
			scrapeURL := addr.String() + suffix

			resp, err := b.client.Get(scrapeURL)
			if err != nil {
				return err
			}

			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("%s returned HTTP status: %s", scrapeURL, resp.Status)
			}

			if err := xml.NewDecoder(resp.Body).Decode(&stats); err != nil {
				return fmt.Errorf("unable to decode XML document: %s", err)
			}

			return nil
		}()

		if err != nil {
			return err
		}
	}

	b.addStatsXMLv3(stats, acc, addr.Host)
	return nil
}
