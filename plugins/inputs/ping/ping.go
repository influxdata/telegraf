package ping

import (
	"context"
	"fmt"
	"net"
	"os/exec"
	"sync"
	"time"

	ping "github.com/sparrc/go-ping"
	"golang.org/x/net/icmp"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// HostPinger is a function that runs the "ping" function using a list of
// passed arguments. This can be easily switched with a mocked ping function
// for unit test purposes (see ping_test.go)
type HostPinger func(binary string, timeout float64, args ...string) (string, error)

type Ping struct {
	wg sync.WaitGroup

	// Interval at which to ping (ping -i <INTERVAL>)
	PingInterval float64 `toml:"ping_interval"`

	// Number of pings to send (ping -c <COUNT>)
	Count int

	// Per-ping timeout, in seconds. 0 means no timeout (ping -W <TIMEOUT>)
	Timeout float64

	// Ping deadline, in seconds. 0 means no deadline. (ping -w <DEADLINE>)
	Deadline int

	// Interface or source address to send ping from (ping -I/-S <INTERFACE/SRC_ADDR>)
	Interface string

	// URLs to ping
	Urls []string

	// Ping executable binary
	Binary string

	// Arguments for ping command. When arguments is not empty, system binary will be used and
	// other options (ping_interval, timeout, etc) will be ignored
	Arguments []string

	// host ping function
	pingHost HostPinger

	// listenAddr is the address associated with the interface defined.
	listenAddr string

	ctx context.Context
}

func (_ *Ping) Description() string {
	return "Ping given url(s) and return statistics"
}

const sampleConfig = `
  ## List of urls to ping
  urls = ["example.org"]

  ## Number of pings to send per collection (ping -c <COUNT>)
  # count = 1

  ## Interval, in s, at which to ping. 0 == default (ping -i <PING_INTERVAL>)
  # ping_interval = 1.0

  ## Per-ping timeout, in s. 0 == no timeout (ping -W <TIMEOUT>)
  # timeout = 1.0

  ## Total-ping deadline, in s. 0 == no deadline (ping -w <DEADLINE>)
  # deadline = 10

  ## Interface or source address to send ping from (ping -I[-S] <INTERFACE/SRC_ADDR>)
  # interface = ""

  ## Specify the ping executable binary, default is "ping"
  # binary = "ping"

  ## Arguments for ping command. When arguments is not empty, system binary will be used and
  ## other options (ping_interval, timeout, etc) will be ignored
  # arguments = ["-c", "3"]
`

func (_ *Ping) SampleConfig() string {
	return sampleConfig
}

func (p *Ping) Gather(acc telegraf.Accumulator) error {
	if p.Interface != "" && p.listenAddr != "" {
		p.listenAddr = getAddr(p.Interface)
	}

	for _, url := range p.Urls {
		_, err := net.LookupHost(url)
		if err != nil {
			acc.AddFields("ping", map[string]interface{}{"result_code": 1}, map[string]string{"url": url})
			// todo: return err?
			acc.AddError(err)
			return nil
		}

		if len(p.Arguments) > 0 {
			p.wg.Add(1)

			go p.pingToURL(url, acc)
		} else {
			pinger, err := ping.NewPinger(url)
			if err != nil {
				acc.AddError(fmt.Errorf("%v: %s", err, url))
				acc.AddFields("ping", map[string]interface{}{"result_code": 2}, map[string]string{"url": url})
				continue
			}

			pinger.Count = p.Count

			if p.PingInterval < 0.2 {
				p.PingInterval = 1
			}
			pinger.Interval = time.Nanosecond * time.Duration(p.PingInterval*1000000000)

			if p.Deadline > 0 {
				pinger.Timeout = time.Duration(p.Deadline) * time.Second
			}

			if p.Timeout <= 0 {
				p.Timeout = 1
			}
			pinger.Deadline = time.Nanosecond * time.Duration(p.Timeout*1000000000)

			pinger.Size = 64

			// TODO: determine need for privileged flag
			pinger.SetPrivileged(true)

			conn, err := pinger.Listen(p.listenAddr)
			if err != nil {
				return err
			}

			p.wg.Add(1)

			go p.pingToURLNative(url, pinger, conn, acc)
		}
	}

	p.wg.Wait()

	return nil
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

func hostPinger(binary string, timeout float64, args ...string) (string, error) {
	bin, err := exec.LookPath(binary)
	if err != nil {
		return "", err
	}
	c := exec.Command(bin, args...)
	out, err := internal.CombinedOutputTimeout(c,
		time.Second*time.Duration(timeout+5))
	return string(out), err
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

// func getGID() uint64 {
// 	b := make([]byte, 64)
// 	b = b[:runtime.Stack(b, false)]
// 	b = bytes.TrimPrefix(b, []byte("goroutine "))
// 	b = b[:bytes.IndexByte(b, ' ')]
// 	n, _ := strconv.ParseUint(string(b), 10, 64)
// 	return n
// }

func (p *Ping) pingToURLNative(u string, pinger *ping.Pinger, conn *icmp.PacketConn, acc telegraf.Accumulator) {

	defer p.wg.Done()
	tags := map[string]string{"url": u}
	fields := map[string]interface{}{"result_code": 0}

	results, err := p.pingHostNative(pinger, conn)
	if err != nil {
		acc.AddError(fmt.Errorf("%s: %s", err, u))
		fields["result_code"] = 2
		acc.AddFields("ping", fields, tags)
		return
	}

	fields["packets_transmitted"] = results.transmitted
	fields["packets_received"] = results.received
	fields["percent_packet_loss"] = results.pktLoss
	if results.ttl > 0 {
		fields["ttl"] = results.ttl
	}
	if results.min >= 0 {
		fields["minimum_response_ms"] = results.min
	}
	if results.avg >= 0 {
		fields["average_response_ms"] = results.avg
	}
	if results.max >= 0 {
		fields["maximum_response_ms"] = results.max
	}
	if results.stddev >= 0 {
		fields["standard_deviation_ms"] = results.stddev
	}

	acc.AddFields("ping", fields, tags)
}

func (p *Ping) pingHostNative(pinger *ping.Pinger, conn *icmp.PacketConn) (*pingResults, error) {
	results := &pingResults{}

	pinger.OnRecv = func(pkt *ping.Packet) {
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

	return results, pinger.DoPing(p.ctx, conn)
}

func init() {
	inputs.Add("ping", func() telegraf.Input {
		return &Ping{
			pingHost:     hostPinger,
			PingInterval: 1.0,
			Count:        1,
			Timeout:      1.0,
			Deadline:     10,
			Binary:       "ping",
			Arguments:    []string{},
			ctx:          context.Background(),
		}
	})
}
