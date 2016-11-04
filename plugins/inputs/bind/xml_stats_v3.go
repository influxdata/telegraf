package bind

import (
	"encoding/xml"
	"fmt"
	"io"
)

type v3Stats struct {
	Server v3Server `xml:"server"`
}

type v3Server struct {
	CounterGroups []v3Counters `xml:"counters"`
}

type v3Counters struct {
	Type     string      `xml:"type,attr"`
	Counters []v3Counter `xml:"counter"`
}

type v3Counter struct {
	Name  string `xml:"name,attr"`
	Value int    `xml:",chardata"`
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
