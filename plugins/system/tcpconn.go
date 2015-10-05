package system

import (
	"fmt"

	"github.com/influxdb/telegraf/plugins"
)

type TCPConnectionStats struct {
	ps PS
}

func (_ *TCPConnectionStats) Description() string {
	return "Read metrics about TCP status such as established, time wait etc"
}

var tcpstatSampleConfig = ""

func (_ *TCPConnectionStats) SampleConfig() string {
	return tcpstatSampleConfig
}

func (s *TCPConnectionStats) Gather(acc plugins.Accumulator) error {
	netconns, err := s.ps.NetConnections()
	if err != nil {
		return fmt.Errorf("error getting net connections info: %s", err)
	}
	counts := make(map[string]int)

	// TODO: add family to tags or else
	tags := map[string]string{}
	for _, netcon := range netconns {
		c, ok := counts[netcon.Status]
		if !ok {
			counts[netcon.Status] = 0
		}
		counts[netcon.Status] = c + 1
	}

	acc.Add("established", counts["ESTABLISHED"], tags)
	acc.Add("syn_sent", counts["SYN_SENT"], tags)
	acc.Add("syn_recv", counts["SYN_RECV"], tags)
	acc.Add("fin_wait1", counts["FIN_WAIT1"], tags)
	acc.Add("fin_wait2", counts["FIN_WAIT2"], tags)
	acc.Add("time_wait", counts["TIME_WAIT"], tags)
	acc.Add("close", counts["CLOSE"], tags)
	acc.Add("close_wait", counts["CLOSE_WAIT"], tags)
	acc.Add("last_ack", counts["LAST_ACK"], tags)
	acc.Add("listen", counts["LISTEN"], tags)
	acc.Add("closing", counts["CLOSING"], tags)
	acc.Add("none", counts["NONE"], tags)

	return nil
}

func init() {
	plugins.Add("tcpconn", func() plugins.Plugin {
		return &TCPConnectionStats{ps: &systemPS{}}
	})
}
