//go:generate ../../../tools/readme_config_includer/generator
package ping

import (
	_ "embed"
	"errors"
	"fmt"
	"math"
	"net"
	"os/exec"
	"runtime"
	"slices"
	"strings"
	"sync"
	"time"

	ping "github.com/prometheus-community/pro-bing"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type Ping struct {
	Urls         []string        `toml:"urls"`          // URLs to ping
	Method       string          `toml:"method"`        // Method defines how to ping (native or exec)
	Count        int             `toml:"count"`         // Number of pings to send (ping -c <COUNT>)
	PingInterval config.Duration `toml:"ping_interval"` // Interval at which to ping (ping -i <INTERVAL>)
	Timeout      config.Duration `toml:"timeout"`       // Per-ping timeout in seconds for the exec method. 0 means no timeout (ping -W <TIMEOUT>)
	Deadline     config.Duration `toml:"deadline"`      // Total ping deadline in seconds. 0 means no deadline (ping -w <DEADLINE>)
	Interface    string          `toml:"interface"`     // Interface or source address to send ping from (ping -I/-S <INTERFACE/SRC_ADDR>)
	Percentiles  []int           `toml:"percentiles"`   // Calculate the given percentiles when using native method
	Binary       string          `toml:"binary"`        // Ping executable binary
	// Arguments for ping command. When arguments are not empty, system binary will be used and other options (ping_interval, timeout, etc.) will be ignored
	Arguments []string        `toml:"arguments"`
	IPv4      bool            `toml:"ipv4"` // Whether to resolve addresses using ipv4 or not.
	IPv6      bool            `toml:"ipv6"` // Whether to resolve addresses using ipv6 or not.
	Size      config.Size     `toml:"size"` // Packet size
	Log       telegraf.Logger `toml:"-"`

	wg             sync.WaitGroup // wg is used to wait for ping with multiple URLs
	calcInterval   time.Duration  // Pre-calculated interval and timeout
	calcTimeout    time.Duration
	sourceAddress  string
	pingHost       hostPingerFunc // host ping function
	nativePingFunc nativePingFunc
}

// hostPingerFunc is a function that runs the "ping" function using a list of
// passed arguments. This can be easily switched with a mocked ping function
// for unit test purposes (see ping_test.go)
type hostPingerFunc func(binary string, timeout float64, args ...string) (string, error)
type nativePingFunc func(destination string) (*pingStats, error)

type pingStats struct {
	ping.Statistics
	ttl int
}

func (*Ping) SampleConfig() string {
	return sampleConfig
}

func (p *Ping) Init() error {
	// Defaults
	if p.Count <= 0 {
		p.Count = 1
	}

	if p.Size == 0 {
		p.Size = 56
	}

	switch p.Method {
	case "":
		p.Method = "exec"
	case "exec", "native":
		// Do nothing, those are valid
	default:
		return fmt.Errorf("invalid 'method' %q", p.Method)
	}

	if p.Binary == "" {
		p.Binary = "ping"
	}

	if p.pingHost == nil {
		p.pingHost = hostPinger
	}

	if p.nativePingFunc == nil {
		p.nativePingFunc = p.nativePing
	}

	// The interval cannot be below 0.2 seconds, matching ping implementation: https://linux.die.net/man/8/ping
	if time.Duration(p.PingInterval) < 200*time.Millisecond {
		p.calcInterval = 200 * time.Millisecond
	} else {
		p.calcInterval = time.Duration(p.PingInterval)
	}

	if p.Method == "native" && p.Timeout > 0 {
		p.Log.Warn(`"timeout" is ignored when method = "native"; use "deadline" to control the total runtime`)
	} else if p.Timeout == 0 {
		p.calcTimeout = 1 * time.Second
	} else {
		p.calcTimeout = time.Duration(p.Timeout)
	}
	return nil
}

func (p *Ping) Gather(acc telegraf.Accumulator) error {
	for _, host := range p.Urls {
		p.wg.Add(1)
		go func(host string) {
			defer p.wg.Done()

			switch p.Method {
			case "native":
				p.pingToURLNative(acc, host)
			default:
				p.pingToURL(acc, host)
			}
		}(host)
	}
	p.wg.Wait()

	return nil
}

func (p *Ping) nativePing(destination string) (*pingStats, error) {
	ps := &pingStats{}

	pinger, err := ping.NewPinger(destination)
	if err != nil {
		return nil, fmt.Errorf("failed to create new pinger: %w", err)
	}

	pinger.SetPrivileged(true)

	if p.IPv4 && p.IPv6 {
		pinger.SetNetwork("ip")
	} else if p.IPv4 {
		pinger.SetNetwork("ip4")
	} else if p.IPv6 {
		pinger.SetNetwork("ip6")
	}

	if p.Method == "native" {
		pinger.Size = int(p.Size)
	}

	// Support either an IP address or interface name
	if p.Interface != "" && p.sourceAddress == "" {
		if addr := net.ParseIP(p.Interface); addr != nil {
			p.sourceAddress = p.Interface
		} else {
			i, err := net.InterfaceByName(p.Interface)
			if err != nil {
				return nil, fmt.Errorf("failed to get interface: %w", err)
			}
			addrs, err := i.Addrs()
			if err != nil {
				return nil, fmt.Errorf("failed to get the address of interface: %w", err)
			}
			if len(addrs) == 0 {
				return nil, fmt.Errorf("no address found for interface %s", p.Interface)
			}
			p.sourceAddress = addrs[0].(*net.IPNet).IP.String()
		}
	}

	pinger.Source = p.sourceAddress
	pinger.Interval = p.calcInterval

	if p.Deadline > 0 {
		pinger.Timeout = time.Duration(p.Deadline)
	}

	// Get Time to live (TTL) of first response, matching original implementation
	once := &sync.Once{}
	pinger.OnRecv = func(pkt *ping.Packet) {
		once.Do(func() {
			ps.ttl = pkt.TTL
		})
	}

	pinger.Count = p.Count
	err = pinger.Run()
	if err != nil {
		if strings.Contains(err.Error(), "operation not permitted") {
			if runtime.GOOS == "linux" {
				return nil, errors.New("permission changes required, enable CAP_NET_RAW capabilities (refer to the ping plugin's README.md for more info)")
			}

			return nil, errors.New("permission changes required, refer to the ping plugin's README.md for more info")
		}
		return nil, err
	}

	ps.Statistics = *pinger.Statistics()

	return ps, nil
}

func (p *Ping) pingToURLNative(acc telegraf.Accumulator, destination string) {
	tags := map[string]string{"url": destination}

	stats, err := p.nativePingFunc(destination)
	if err != nil {
		p.Log.Errorf("ping failed: %s", err.Error())
		fields := make(map[string]interface{}, 1)
		if strings.Contains(err.Error(), "unknown") {
			fields["result_code"] = 1
		} else {
			fields["result_code"] = 2
		}
		acc.AddFields("ping", fields, tags)
		return
	}

	fields := map[string]interface{}{
		"result_code":         0,
		"packets_transmitted": stats.PacketsSent,
		"packets_received":    stats.PacketsRecv,
	}

	if stats.PacketsSent == 0 {
		p.Log.Debug("no packets sent")
		fields["result_code"] = 2
		acc.AddFields("ping", fields, tags)
		return
	}

	if stats.PacketsRecv == 0 {
		p.Log.Debug("no packets received")
		fields["result_code"] = 1
		fields["percent_packet_loss"] = float64(100)
		acc.AddFields("ping", fields, tags)
		return
	}
	slices.Sort(stats.Rtts)

	for _, perc := range p.Percentiles {
		var value = percentile(stats.Rtts, perc)
		var field = fmt.Sprintf("percentile%v_ms", perc)
		fields[field] = float64(value.Nanoseconds()) / float64(time.Millisecond)
	}

	// Set TTL only on supported platform. See golang.org/x/net/ipv4/payload_cmsg.go
	switch runtime.GOOS {
	case "aix", "darwin", "dragonfly", "freebsd", "linux", "netbsd", "openbsd", "solaris":
		fields["ttl"] = stats.ttl
	}

	fields["percent_packet_loss"] = float64(stats.PacketLoss)
	fields["minimum_response_ms"] = float64(stats.MinRtt) / float64(time.Millisecond)
	fields["average_response_ms"] = float64(stats.AvgRtt) / float64(time.Millisecond)
	fields["maximum_response_ms"] = float64(stats.MaxRtt) / float64(time.Millisecond)
	fields["standard_deviation_ms"] = float64(stats.StdDevRtt) / float64(time.Millisecond)

	acc.AddFields("ping", fields, tags)
}

// R7 from Hyndman and Fan (1996), which matches Excel
func percentile(values []time.Duration, perc int) time.Duration {
	if len(values) == 0 {
		return 0
	}
	if perc < 0 {
		perc = 0
	}
	if perc > 100 {
		perc = 100
	}
	percFloat := float64(perc) / 100.0

	count := len(values)
	rank := percFloat * float64(count-1)
	rankInteger := int(rank)
	rankFraction := rank - math.Floor(rank)

	if rankInteger >= count-1 {
		return values[count-1]
	}

	upper := values[rankInteger+1]
	lower := values[rankInteger]
	return lower + time.Duration(rankFraction*float64(upper-lower))
}

func hostPinger(binary string, timeout float64, args ...string) (string, error) {
	bin, err := exec.LookPath(binary)
	if err != nil {
		return "", err
	}
	c := exec.Command(bin, args...)
	out, err := internal.CombinedOutputTimeout(c, time.Second*time.Duration(timeout+5))
	return string(out), err
}

func init() {
	inputs.Add("ping", func() telegraf.Input {
		return &Ping{
			PingInterval: config.Duration(1 * time.Second),
			Deadline:     config.Duration(10 * time.Second),
		}
	})
}
