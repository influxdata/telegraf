package ping

import (
	"errors"
	"net"
	"runtime"
	"time"

	ping "github.com/glinton/go-ping"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Ping struct {
	PingInterval float64  `toml:"ping_interval"` // Interval at which to ping (ping -i <INTERVAL>)
	Count        int      `toml:"count"`         // Number of pings to send (ping -c <COUNT>)
	Timeout      float64  `toml:"timeout"`       // Per-ping timeout, in seconds. 0 means no timeout (ping -W <TIMEOUT>)
	Deadline     int      `toml:"deadline"`      // Ping deadline, in seconds. 0 means no deadline. (ping -w <DEADLINE>)
	Interface    string   `toml:"interface"`     // Interface or source address to send ping from (ping -I/-S <INTERFACE/SRC_ADDR>)
	Hosts        []string `toml:"hosts"`         // Hosts to ping
	IPV6         bool     `toml:"ipv6"`          // Whether to ping ipv6 addresses

	network    string                 // network is the network to listen on ("udp4", "udp6", "ip4:icmp", "ip6:ip6-icmp")
	listenAddr string                 // listenAddr is the address associated with the interface defined.
	hostCache  []net.Addr             // hosts to ping
	rcvdCache  map[string]pingResults // cache of echo responses received
}

func (*Ping) Description() string {
	return "Ping given host(s) and return statistics"
}

const sampleConfig = `
  ## List of hosts to ping.
  hosts = ["8.8.8.8"]

  ## Number of pings to send per collection.
  # count = 1

  ## Interval, in s, at which to ping.
  # ping_interval = 1.0

  ## Per-ping timeout, in s (0 == no timeout).
  # timeout = 1.0

  ## Total-ping deadline, in s. Set to value equal to or lower than agent interval.
  # deadline = 10

  ## Interface or source address to send ping from.
  # interface = ""

	## Whether to ping ipv6 addresses.
  # ipv6 = false
`

func (*Ping) SampleConfig() string {
	return sampleConfig
}

func getAddr(iface string) string {
	if addr := net.ParseIP(iface); addr != nil {
		return addr.String()
	}

	ifaces, err := net.Interfaces()
	if err != nil {
		return ""
	}

	var ip net.IP
	for i := range ifaces {
		if ifaces[i].Name == iface {
			addrs, err := ifaces[i].Addrs()
			if err != nil {
				return ""
			}
			if len(addrs) > 0 {
				switch v := addrs[0].(type) {
				case *net.IPNet:
					ip = v.IP
				case *net.IPAddr:
					ip = v.IP
				}
				if len(ip) == 0 {
					return ""
				}
				return ip.String()
			}
		}
	}

	return ""
}

func (p *Ping) buildHostCache(acc telegraf.Accumulator) {
	if p.hostCache != nil {
		return
	}

	p.network = "ip4:icmp"
	if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
		p.network = "udp4"
		if p.IPV6 {
			p.network = "udp6"
		}
	} else if p.IPV6 {
		p.network = "ip6:ipv6-icmp"
	}

	p.hostCache = []net.Addr{}

	for _, url := range p.Hosts {
		var addr net.Addr
		var err error
		if p.IPV6 {
			addr, err = net.ResolveIPAddr("ip6", url)
		} else {
			addr, err = net.ResolveIPAddr("ip4", url)
		}

		if a, ok := addr.(*net.IPAddr); ok && runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
			addr = &net.UDPAddr{IP: a.IP, Zone: a.Zone}
		}

		if err != nil {
			acc.AddFields("ping", map[string]interface{}{"result_code": 1}, map[string]string{"url": url})
			acc.AddError(err)
			continue
		}
		if addr == nil {
			continue
		}
		p.hostCache = append(p.hostCache, addr)
	}
}

func (p *Ping) onRecv(pkt *ping.Packet) {
	p.rcvdCache[pkt.IPAddr] = pingResults{ttl: pkt.TTL}
}

func (p *Ping) onFinish(acc telegraf.Accumulator) func(stats *ping.Statistics) {
	return func(stats *ping.Statistics) {
		min := float64(stats.MinRTT.Nanoseconds()) / float64(time.Millisecond)
		avg := float64(stats.AvgRTT.Nanoseconds()) / float64(time.Millisecond)
		max := float64(stats.MaxRTT.Nanoseconds()) / float64(time.Millisecond)
		stddev := float64(stats.StdDevRTT.Nanoseconds()) / float64(time.Millisecond)

		tags := map[string]string{"ip": stats.Addr}
		fields := map[string]interface{}{
			"result_code":         0,
			"packets_transmitted": int(stats.PacketsSent),
			"packets_received":    int(stats.PacketsRecv),
			"percent_packet_loss": stats.PacketLoss,
		}

		if p.rcvdCache[stats.Addr].ttl > 0 {
			fields["ttl"] = p.rcvdCache[stats.Addr].ttl
		}
		if min >= 0 {
			fields["minimum_response_ms"] = min
		}
		if avg >= 0 {
			fields["average_response_ms"] = avg
		}
		if max >= 0 {
			fields["maximum_response_ms"] = max
		}
		if stddev >= 0 {
			fields["standard_deviation_ms"] = stddev
		}

		acc.AddFields("ping", fields, tags)
	}
}

// Gather gathers ping metrics via native ping implementation.
func (p *Ping) Gather(acc telegraf.Accumulator) error {
	if p.Interface != "" && p.listenAddr == "" {
		p.listenAddr = getAddr(p.Interface)
	}

	p.buildHostCache(acc)
	if len(p.hostCache) == 0 {
		return errors.New("no valid hosts to ping")
	}

	conn, err := ping.Listen(p.network, p.listenAddr)
	defer conn.Close()
	if err != nil {
		return err
	}

	if p.Count < 0 {
		p.Count = 0
	}

	var size uint = 56
	if p.IPV6 {
		size = 128
	}

	pinger, err := ping.NewPinger(
		conn,
		ping.WithOnRecieve(p.onRecv),
		ping.WithOnFinish(p.onFinish(acc)),
		ping.WithSize(size),
		ping.WithCount(uint(p.Count)),
		ping.WithInterval(time.Nanosecond*time.Duration(p.PingInterval*1000000000)),
		ping.WithTimeout(time.Nanosecond*time.Duration(p.Timeout*1000000000)),
		ping.WithDeadline(time.Duration(p.Deadline)*time.Second),
	)
	if err != nil {
		return err
	}

	if len(p.hostCache) == 1 {
		pinger.Send(p.hostCache[0])
	} else {
		pinger.Send(p.hostCache[0], p.hostCache[1:]...)
	}

	return nil
}

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

func init() {
	inputs.Add("ping_native", func() telegraf.Input {
		return &Ping{
			PingInterval: 1.0,
			Count:        1,
			Timeout:      1.0,
			Deadline:     10,
			rcvdCache:    make(map[string]pingResults),
		}
	})
}
