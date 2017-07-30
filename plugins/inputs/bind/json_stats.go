package bind

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/influxdata/telegraf"
)

type jsonStats struct {
	OpCodes   map[string]int
	QTypes    map[string]int
	NSStats   map[string]int
	SockStats map[string]int
	Views     map[string]jsonView
	Memory    jsonMemory
}

type jsonMemory struct {
	TotalUse    int
	InUse       int
	BlockSize   int
	ContextSize int
	Lost        int
	Contexts    []struct {
		Id    string
		Name  string
		Total int
		InUse int
	}
}

type jsonView struct {
	Resolver map[string]map[string]int
}

// addJsonCounter adds a counter array to a Telegraf Accumulator, with the specified tags.
func addJsonCounter(acc telegraf.Accumulator, commonTags map[string]string, stats map[string]int) {
	for name, value := range stats {
		if commonTags["type"] == "opcode" && strings.HasPrefix(name, "RESERVED") {
			continue
		}

		tags := make(map[string]string)

		// Create local copy of tags since maps are reference types
		for k, v := range commonTags {
			tags[k] = v
		}

		tags["name"] = name
		fields := map[string]interface{}{"value": value}

		acc.AddCounter("bind_counter", fields, tags)
	}
}

// addStatsJson walks a jsonStats struct and adds the values to the telegraf.Accumulator.
func (b *Bind) addStatsJson(stats jsonStats, acc telegraf.Accumulator, urlTag string) {
	tags := map[string]string{"url": urlTag}

	// Opcodes
	tags["type"] = "opcode"
	addJsonCounter(acc, tags, stats.OpCodes)

	// Query RDATA types
	tags["type"] = "qtype"
	addJsonCounter(acc, tags, stats.QTypes)

	// Nameserver stats
	tags["type"] = "nsstat"
	addJsonCounter(acc, tags, stats.NSStats)

	// Socket statistics
	tags["type"] = "sockstat"
	addJsonCounter(acc, tags, stats.SockStats)

	// Memory stats
	fields := map[string]interface{}{
		"TotalUse":    stats.Memory.TotalUse,
		"InUse":       stats.Memory.InUse,
		"BlockSize":   stats.Memory.BlockSize,
		"ContextSize": stats.Memory.ContextSize,
		"Lost":        stats.Memory.Lost,
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
		for vName, view := range stats.Views {
			for cntrType, counters := range view.Resolver {
				for cntrName, value := range counters {
					tags := map[string]string{
						"url":  urlTag,
						"view": vName,
						"type": cntrType,
						"name": cntrName,
					}
					fields := map[string]interface{}{"value": value}

					acc.AddCounter("bind_counter", fields, tags)
				}
			}
		}
	}
}

// readStatsJson takes a base URL to probe, and requests the individual statistics blobs that we
// are interested in. These individual blobs have a combined size which is significantly smaller
// than if we requested everything at once (e.g. taskmgr and socketmgr can be omitted).
func (b *Bind) readStatsJson(addr *url.URL, acc telegraf.Accumulator) error {
	var stats jsonStats

	// Progressively build up full jsonStats struct by parsing the individual HTTP responses
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

		log.Printf("D! Response content length: %d", resp.ContentLength)

		if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
			return fmt.Errorf("Unable to decode JSON blob: %s", err)
		}
	}

	b.addStatsJson(stats, acc, addr.Host)
	return nil
}

// readStatsJsonComplete is similar to readStatsJson, but takes an io.Reader HTTP response body
// as a result of attempting to auto-detect the statistics format of a URL.
func (b *Bind) readStatsJsonComplete(addr *url.URL, acc telegraf.Accumulator, r io.Reader) error {
	var stats jsonStats

	if err := json.NewDecoder(r).Decode(&stats); err != nil {
		return fmt.Errorf("Unable to decode JSON blob: %s", err)
	}

	b.addStatsJson(stats, acc, addr.Host)
	return nil
}
