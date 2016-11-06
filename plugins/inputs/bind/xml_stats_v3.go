package bind

import (
	"encoding/xml"
	"fmt"
	"io"
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

// dumpStats prints the key-value pairs of a version 3 statistics struct
func (stats *v3Stats) dumpStats() {
	for _, cg := range stats.Server.CounterGroups {
		fmt.Printf("COUNTER GROUP - %s\n", cg.Type)

		for _, c := range cg.Counters {
			fmt.Printf("%s => %d\n", c.Name, c.Value)
		}

		fmt.Println()
	}
}

// readStatsV2 decodes a BIND9 XML statistics version 3 document
func readStatsV3(r io.Reader) {
	var stats v3Stats

	if err := xml.NewDecoder(r).Decode(&stats); err != nil {
		fmt.Println(err)
		return
	}

	stats.dumpStats()
}
