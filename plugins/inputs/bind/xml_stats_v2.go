package bind

import (
	"encoding/xml"
	"fmt"
	"io"
)

type v2Root struct {
	XMLName    xml.Name
	Version    string       `xml:"version,attr"`
	Statistics v2Statistics `xml:"bind>statistics"`
}

type v2Statistics struct {
	Version string       `xml:"version,attr"`
	Memory  v2MemoryStat `xml:"memory>summary"`
	Server  v2Server     `xml:"server"`
	Views   []v2View     `xml:"views>view"`
}

type v2MemoryStat struct {
	TotalUse    int
	InUse       int
	BlockSize   int
	ContextSize int
	Lost        int
}

type v2Server struct {
	NSStats     []v2StatCounter `xml:"nsstat"`
	OpCodeStats []v2StatCounter `xml:"requests>opcode"`
	QueryStats  []v2StatCounter `xml:"queries-in>rdtype"`
	SockStats   []v2StatCounter `xml:"sockstat"`
	ZoneStats   []v2StatCounter `xml:"zonestat"`
}

type v2View struct {
	Name          string          `xml:"name"`
	QueryStats    []v2StatCounter `xml:"rdtype"`
	ResolverStats []v2StatCounter `xml:"resstat"`
}

type v2StatCounter struct {
	Name  string `xml:"name"`
	Value int    `xml:"counter"`
}

// dumpStats prints the key-value pairs of a version 2 statistics struct
func (stats *v2Root) dumpStats() {
	fmt.Printf("%#v\n", stats)

	fmt.Println("\nNAMESERVER STATS")
	for _, st := range stats.Statistics.Server.NSStats {
		fmt.Printf("%s => %d\n", st.Name, st.Value)
	}

	fmt.Println("\nOPCODE STATS")
	for _, st := range stats.Statistics.Server.OpCodeStats {
		fmt.Printf("%s => %d\n", st.Name, st.Value)
	}

	fmt.Println("\nQUERY TYPE STATS")
	for _, st := range stats.Statistics.Server.QueryStats {
		fmt.Printf("%s => %d\n", st.Name, st.Value)
	}

	fmt.Println("\nSOCK STATS")
	for _, st := range stats.Statistics.Server.SockStats {
		fmt.Printf("%s => %d\n", st.Name, st.Value)
	}

	fmt.Println("\nZONE STATS")
	for _, st := range stats.Statistics.Server.ZoneStats {
		fmt.Printf("%s => %d\n", st.Name, st.Value)
	}
}

// readStatsV2 decodes a BIND9 XML statistics version 2 document
func readStatsV2(r io.Reader) {
	var stats v2Root

	if err := xml.NewDecoder(r).Decode(&stats); err != nil {
		fmt.Println(err)
		return
	}

	stats.dumpStats()
}
