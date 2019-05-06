package ping

import (
	"time"

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
	results := &pingResults{}

	pinger.OnRecv = func(pkt *ping.Packet) {
		// fmt.Printf("packet: %#v \n", pkt)
		results.ttl = pkt.Ttl
	}
	pinger.OnFinish = func(stats *ping.Statistics) {
		results.received = stats.PacketsRecv
		results.transmitted = stats.PacketsSent
		results.pktLoss = stats.PacketLoss
		results.min = float64(stats.MinRtt.Nanoseconds()) / float64(time.Millisecond)
		results.avg = float64(stats.AvgRtt.Nanoseconds()) / float64(time.Millisecond)
		results.max = float64(stats.MaxRtt.Nanoseconds()) / float64(time.Millisecond)
		results.stddev = float64(stats.StdDevRtt.Nanoseconds()) / float64(time.Millisecond)
	}
	pinger.DoPing(conn)

	return results, nil
}
