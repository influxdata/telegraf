package system

import (
	"fmt"
	"syscall"

	"github.com/influxdb/telegraf/plugins"
)

type NetStats struct {
	ps PS
}

func (_ *NetStats) Description() string {
	return "Read metrics about TCP status such as established, time wait etc and UDP sockets counts."
}

var tcpstatSampleConfig = ""

func (_ *NetStats) SampleConfig() string {
	return tcpstatSampleConfig
}

func (s *NetStats) Gather(acc plugins.Accumulator) error {
	netconns, err := s.ps.NetConnections()
	if err != nil {
		return fmt.Errorf("error getting net connections info: %s", err)
	}
	counts := make(map[string]int)
	counts["UDP"] = 0

	// TODO: add family to tags or else
	tags := map[string]string{}
	for _, netcon := range netconns {
		if netcon.Type == syscall.SOCK_DGRAM {
			counts["UDP"] += 1
			continue // UDP has no status
		}
		c, ok := counts[netcon.Status]
		if !ok {
			counts[netcon.Status] = 0
		}
		counts[netcon.Status] = c + 1
	}
	acc.Add("tcp_established", counts["ESTABLISHED"], tags)
	acc.Add("tcp_syn_sent", counts["SYN_SENT"], tags)
	acc.Add("tcp_syn_recv", counts["SYN_RECV"], tags)
	acc.Add("tcp_fin_wait1", counts["FIN_WAIT1"], tags)
	acc.Add("tcp_fin_wait2", counts["FIN_WAIT2"], tags)
	acc.Add("tcp_time_wait", counts["TIME_WAIT"], tags)
	acc.Add("tcp_close", counts["CLOSE"], tags)
	acc.Add("tcp_close_wait", counts["CLOSE_WAIT"], tags)
	acc.Add("tcp_last_ack", counts["LAST_ACK"], tags)
	acc.Add("tcp_listen", counts["LISTEN"], tags)
	acc.Add("tcp_closing", counts["CLOSING"], tags)
	acc.Add("tcp_none", counts["NONE"], tags)
	acc.Add("udp_socket", counts["UDP"], tags)

	return nil
}

func init() {
	plugins.Add("netstat", func() plugins.Plugin {
		return &NetStats{ps: &systemPS{}}
	})
}
