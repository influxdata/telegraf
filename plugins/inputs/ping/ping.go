package ping

import (
	"errors"
	"fmt"
	"math"
	"net"
	"os/exec"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/go-ping/ping"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const (
	defaultPingDataBytesSize = 56
)

// HostPinger is a function that runs the "ping" function using a list of
// passed arguments. This can be easily switched with a mocked ping function
// for unit test purposes (see ping_test.go)
type HostPinger func(binary string, timeout float64, args ...string) (string, error)

type Ping struct {
	// wg is used to wait for ping with multiple URLs
	wg sync.WaitGroup

	// Pre-calculated interval and timeout
	calcInterval time.Duration
	calcTimeout  time.Duration

	sourceAddress string

	Log telegraf.Logger `toml:"-"`

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

	nativePingFunc NativePingFunc

	// Calculate the given percentiles when using native method
	Percentiles []int

	// Packet size
	Size *int
}

func (*Ping) Description() string {
	return "Ping given url(s) and return statistics"
}

const sampleConfig = `
  ## Hosts to send ping packets to.
  urls = ["example.org"]

  ## Method used for sending pings, can be either "exec" or "native".  When set
  ## to "exec" the systems ping command will be executed.  When set to "native"
  ## the plugin will send pings directly.
  ##
  ## While the default is "exec" for backwards compatibility, new deployments
  ## are encouraged to use the "native" method for improved compatibility and
  ## performance.
  # method = "exec"

  ## Number of ping packets to send per interval.  Corresponds to the "-c"
  ## option of the ping command.
  # count = 1

  ## Time to wait between sending ping packets in seconds.  Operates like the
  ## "-i" option of the ping command.
  # ping_interval = 1.0

  ## If set, the time to wait for a ping response in seconds.  Operates like
  ## the "-W" option of the ping command.
  # timeout = 1.0

  ## If set, the total ping deadline, in seconds.  Operates like the -w option
  ## of the ping command.
  # deadline = 10

  ## Interface or source address to send ping from.  Operates like the -I or -S
  ## option of the ping command.
  # interface = ""

  ## Percentiles to calculate. This only works with the native method.
  # percentiles = [50, 95, 99]

  ## Specify the ping executable binary.
  # binary = "ping"

  ## Arguments for ping command. When arguments is not empty, the command from
  ## the binary option will be used and other options (ping_interval, timeout,
  ## etc) will be ignored.
  # arguments = ["-c", "3"]

  ## Use only IPv6 addresses when resolving a hostname.
  # ipv6 = false

  ## Number of data bytes to be sent. Corresponds to the "-s"
  ## option of the ping command. This only works with the native method.
  # size = 56
`

func (*Ping) SampleConfig() string {
	return sampleConfig
}

func (p *Ping) Gather(acc telegraf.Accumulator) error {
	for _, host := range p.Urls {
		p.wg.Add(1)
		go func(host string) {
			defer p.wg.Done()

			switch p.Method {
			case "native":
				p.pingToURLNative(host, acc)
			default:
				p.pingToURL(host, acc)
			}
		}(host)
	}

	p.wg.Wait()

	return nil
}

type pingStats struct {
	ping.Statistics
	ttl int
}

type NativePingFunc func(destination string) (*pingStats, error)

func (p *Ping) nativePing(destination string) (*pingStats, error) {
	ps := &pingStats{}

	pinger, err := ping.NewPinger(destination)
	if err != nil {
		return nil, fmt.Errorf("failed to create new pinger: %w", err)
	}

	pinger.SetPrivileged(true)

	if p.IPv6 {
		pinger.SetNetwork("ip6")
	}

	if p.Method == "native" {
		pinger.Size = defaultPingDataBytesSize
		if p.Size != nil {
			pinger.Size = *p.Size
		}
	}

	pinger.Source = p.sourceAddress
	pinger.Interval = p.calcInterval

	if p.Deadline > 0 {
		pinger.Timeout = time.Duration(p.Deadline) * time.Second
	}

	// Get Time to live (TTL) of first response, matching original implementation
	once := &sync.Once{}
	pinger.OnRecv = func(pkt *ping.Packet) {
		once.Do(func() {
			ps.ttl = pkt.Ttl
		})
	}

	pinger.Count = p.Count
	err = pinger.Run()
	if err != nil {
		if strings.Contains(err.Error(), "operation not permitted") {
			if runtime.GOOS == "linux" {
				return nil, fmt.Errorf("permission changes required, enable CAP_NET_RAW capabilities (refer to the ping plugin's README.md for more info)")
			}

			return nil, fmt.Errorf("permission changes required, refer to the ping plugin's README.md for more info")
		}
		return nil, fmt.Errorf("%w", err)
	}

	ps.Statistics = *pinger.Statistics()

	return ps, nil
}

func (p *Ping) pingToURLNative(destination string, acc telegraf.Accumulator) {
	tags := map[string]string{"url": destination}
	fields := map[string]interface{}{}

	stats, err := p.nativePingFunc(destination)
	if err != nil {
		p.Log.Errorf("ping failed: %s", err.Error())
		if strings.Contains(err.Error(), "unknown") {
			fields["result_code"] = 1
		} else {
			fields["result_code"] = 2
		}
		acc.AddFields("ping", fields, tags)
		return
	}

	fields = map[string]interface{}{
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

	sort.Sort(durationSlice(stats.Rtts))
	for _, perc := range p.Percentiles {
		var value = percentile(durationSlice(stats.Rtts), perc)
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

type durationSlice []time.Duration

func (p durationSlice) Len() int           { return len(p) }
func (p durationSlice) Less(i, j int) bool { return p[i] < p[j] }
func (p durationSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// R7 from Hyndman and Fan (1996), which matches Excel
func percentile(values durationSlice, perc int) time.Duration {
	if len(values) == 0 {
		return 0
	}
	if perc < 0 {
		perc = 0
	}
	if perc > 100 {
		perc = 100
	}
	var percFloat = float64(perc) / 100.0

	var count = len(values)
	var rank = percFloat * float64(count-1)
	var rankInteger = int(rank)
	var rankFraction = rank - math.Floor(rank)

	if rankInteger >= count-1 {
		return values[count-1]
	}

	upper := values[rankInteger+1]
	lower := values[rankInteger]
	return lower + time.Duration(rankFraction*float64(upper-lower))
}

// Init ensures the plugin is configured correctly.
func (p *Ping) Init() error {
	if p.Count < 1 {
		return errors.New("bad number of packets to transmit")
	}

	// The interval cannot be below 0.2 seconds, matching ping implementation: https://linux.die.net/man/8/ping
	if p.PingInterval < 0.2 {
		p.calcInterval = time.Duration(.2 * float64(time.Second))
	} else {
		p.calcInterval = time.Duration(p.PingInterval * float64(time.Second))
	}

	// If no timeout is given default to 5 seconds, matching original implementation
	if p.Timeout == 0 {
		p.calcTimeout = time.Duration(5) * time.Second
	} else {
		p.calcTimeout = time.Duration(p.Timeout) * time.Second
	}

	// Support either an IP address or interface name
	if p.Interface != "" {
		if addr := net.ParseIP(p.Interface); addr != nil {
			p.sourceAddress = p.Interface
		} else {
			i, err := net.InterfaceByName(p.Interface)
			if err != nil {
				return fmt.Errorf("failed to get interface: %w", err)
			}
			addrs, err := i.Addrs()
			if err != nil {
				return fmt.Errorf("failed to get the address of interface: %w", err)
			}
			p.sourceAddress = addrs[0].(*net.IPNet).IP.String()
		}
	}

	return nil
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

func init() {
	inputs.Add("ping", func() telegraf.Input {
		p := &Ping{
			pingHost:     hostPinger,
			PingInterval: 1.0,
			Count:        1,
			Timeout:      1.0,
			Deadline:     10,
			Method:       "exec",
			Binary:       "ping",
			Arguments:    []string{},
			Percentiles:  []int{},
		}
		p.nativePingFunc = p.nativePing
		return p
	})
}
