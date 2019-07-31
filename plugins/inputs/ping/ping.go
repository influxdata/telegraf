package ping

import (
	"context"
	"errors"
	"math"
	"net"
	"os/exec"
	"runtime"
	"sync"
	"time"

	"github.com/glinton/ping"

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

	// Method defines how to ping (native or exec)
	Method string

	// Ping executable binary
	Binary string

	// Arguments for ping command. When arguments is not empty, system binary will be used and
	// other options (ping_interval, timeout, etc) will be ignored
	Arguments []string

	// Whether to resolve addresses using ipv6 or not.
	IPv6 bool

	// host ping function
	pingHost HostPinger

	// listenAddr is the address associated with the interface defined.
	listenAddr string
}

func (*Ping) Description() string {
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

  ## How to ping. "native" doesn't have external dependencies, while "exec" depends on 'ping'.
  # method = "exec"

  ## Specify the ping executable binary, default is "ping"
	# binary = "ping"

  ## Arguments for ping command. When arguments is not empty, system binary will be used and
  ## other options (ping_interval, timeout, etc) will be ignored.
  # arguments = ["-c", "3"]

  ## Use only ipv6 addresses when resolving hostnames.
  # ipv6 = false
`

func (*Ping) SampleConfig() string {
	return sampleConfig
}

func (p *Ping) Gather(acc telegraf.Accumulator) error {
	if p.Interface != "" && p.listenAddr != "" {
		p.listenAddr = getAddr(p.Interface)
	}

	for _, ip := range p.Urls {
		_, err := net.LookupHost(ip)
		if err != nil {
			acc.AddFields("ping", map[string]interface{}{"result_code": 1}, map[string]string{"ip": ip})
			acc.AddError(err)
			return nil
		}

		if p.Method == "native" {
			p.wg.Add(1)
			go func(ip string) {
				defer p.wg.Done()
				p.pingToURLNative(ip, acc)
			}(ip)
		} else {
			p.wg.Add(1)
			go func(ip string) {
				defer p.wg.Done()
				p.pingToURL(ip, acc)
			}(ip)
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

func (p *Ping) pingToURLNative(destination string, acc telegraf.Accumulator) {
	ctx := context.Background()

	network := "ip4"
	if p.IPv6 {
		network = "ip6"
	}

	host, err := net.ResolveIPAddr(network, destination)
	if err != nil {
		acc.AddFields("ping", map[string]interface{}{"result_code": 1}, map[string]string{"url": destination})
		acc.AddError(err)
		return
	}

	interval := p.PingInterval
	if interval < 0.2 {
		interval = 0.2
	}

	timeout := p.Timeout
	if timeout == 0 {
		timeout = 5
	}

	tick := time.NewTicker(time.Duration(interval * float64(time.Second)))
	defer tick.Stop()

	if p.Deadline > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(p.Deadline)*time.Second)
		defer cancel()
	}

	resps := make(chan *ping.Response)
	rsps := []*ping.Response{}

	r := &sync.WaitGroup{}
	r.Add(1)
	go func() {
		for res := range resps {
			rsps = append(rsps, res)
		}
		r.Done()
	}()

	wg := &sync.WaitGroup{}
	c := ping.Client{}

	var i int
	for i = 0; i < p.Count; i++ {
		select {
		case <-ctx.Done():
			goto finish
		case <-tick.C:
			ctx, cancel := context.WithTimeout(ctx, time.Duration(timeout*float64(time.Second)))
			defer cancel()

			wg.Add(1)
			go func(seq int) {
				defer wg.Done()
				resp, err := c.Do(ctx, &ping.Request{
					Dst: net.ParseIP(host.String()),
					Src: net.ParseIP(p.listenAddr),
					Seq: seq,
				})
				if err != nil {
					acc.AddFields("ping", map[string]interface{}{"result_code": 2}, map[string]string{"url": destination})
					acc.AddError(err)
					return
				}

				resps <- resp
			}(i + 1)
		}
	}

finish:
	wg.Wait()
	close(resps)

	r.Wait()
	tags, fields := onFin(i, rsps, destination)
	acc.AddFields("ping", fields, tags)
}

func onFin(packetsSent int, resps []*ping.Response, destination string) (map[string]string, map[string]interface{}) {
	packetsRcvd := len(resps)

	tags := map[string]string{"url": destination}
	fields := map[string]interface{}{
		"result_code":         0,
		"packets_transmitted": packetsSent,
		"packets_received":    packetsRcvd,
	}

	if packetsSent == 0 {
		return tags, fields
	}

	if packetsRcvd == 0 {
		fields["percent_packet_loss"] = float64(100)
		return tags, fields
	}

	fields["percent_packet_loss"] = float64(packetsSent-packetsRcvd) / float64(packetsSent) * 100
	ttl := resps[0].TTL

	var min, max, avg, total time.Duration
	min = resps[0].RTT
	max = resps[0].RTT

	for _, res := range resps {
		if res.RTT < min {
			min = res.RTT
		}
		if res.RTT > max {
			max = res.RTT
		}
		total += res.RTT
	}

	avg = total / time.Duration(packetsRcvd)
	var sumsquares time.Duration
	for _, res := range resps {
		sumsquares += (res.RTT - avg) * (res.RTT - avg)
	}
	stdDev := time.Duration(math.Sqrt(float64(sumsquares / time.Duration(packetsRcvd))))

	// Set TTL only on supported platform. See golang.org/x/net/ipv4/payload_cmsg.go
	switch runtime.GOOS {
	case "aix", "darwin", "dragonfly", "freebsd", "linux", "netbsd", "openbsd", "solaris":
		fields["ttl"] = ttl
	}

	fields["minimum_response_ms"] = float64(min.Nanoseconds()) / float64(time.Millisecond)
	fields["average_response_ms"] = float64(avg.Nanoseconds()) / float64(time.Millisecond)
	fields["maximum_response_ms"] = float64(max.Nanoseconds()) / float64(time.Millisecond)
	fields["standard_deviation_ms"] = float64(stdDev.Nanoseconds()) / float64(time.Millisecond)

	return tags, fields
}

// Init ensures the plugin is configured correctly.
func (p *Ping) Init() error {
	if p.Count < 1 {
		return errors.New("bad number of packets to transmit")
	}

	return nil
}

func init() {
	inputs.Add("ping", func() telegraf.Input {
		return &Ping{
			pingHost:     hostPinger,
			PingInterval: 1.0,
			Count:        1,
			Timeout:      1.0,
			Deadline:     10,
			Method:       "exec",
			Binary:       "ping",
			Arguments:    []string{},
		}
	})
}
