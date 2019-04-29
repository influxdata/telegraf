package ping

import (
	"fmt"

	ping "github.com/sparrc/go-ping"
	"golang.org/x/net/icmp"
)

type pingResults struct {
	transmitted int
	received    int
	pktLoss     float64
	ttl         int
	min         float64
	avg         float64
	max         float64
	stddev      float64
}

// TODO: add privileged flag
func (p *Ping) pingHostNative(pinger *ping.Pinger, conn *icmp.PacketConn) (*pingResults, error) {
	results := pingResults{}

	pinger.OnRecv = func(pkt *ping.Packet) {
		fmt.Printf("packet: %#v \n", pkt)
		results.ttl = pkt.Ttl
	}
	pinger.OnFinish = func(stats *ping.Statistics) {
		results.received = stats.PacketsRecv
		results.transmitted = stats.PacketsSent
		results.pktLoss = stats.PacketLoss
		results.min = stats.MinRtt.Seconds() * 1000
		results.avg = stats.AvgRtt.Seconds() * 1000
		results.max = stats.MaxRtt.Seconds() * 1000
		results.stddev = stats.StdDevRtt.Seconds() * 1000

		fmt.Printf("stats: %#v \n", stats)
	}
	pinger.DoPing(conn)

	return &results, nil
}
