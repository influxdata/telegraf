package bind

import (
	"encoding/json"
	"fmt"
	"io"
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

// addCounter adds a counter array to a Telegraf Accumulator, with the specified tags
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

// readStatsJson decodes a BIND9 JSON statistics blob
func (b *Bind) readStatsJson(r io.Reader, acc telegraf.Accumulator, url string) error {
	var stats jsonStats

	if err := json.NewDecoder(r).Decode(&stats); err != nil {
		return fmt.Errorf("Unable to decode JSON blob: %s", err)
	}

	tags := map[string]string{"url": url}

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
		for vName, view := range stats.Views {
			for cntrType, counters := range view.Resolver {
				for cntrName, value := range counters {
					tags := map[string]string{
						"url":  url,
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

	return nil
}
