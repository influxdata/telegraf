//go:generate ../../../tools/readme_config_includer/generator
package internet_speed

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"math"
	"net"
	"net/netip"
	"os"
	"time"

	"github.com/showwin/speedtest-go/speedtest"
	"github.com/showwin/speedtest-go/speedtest/transport"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

const (
	measurement    = "internet_speed"
	testModeSingle = "single"
	testModeMulti  = "multi"
)

type InternetSpeed struct {
	ServerIDInclude  []string `toml:"server_id_include"`
	ServerIDExclude  []string `toml:"server_id_exclude"`
	MemorySavingMode bool     `toml:"memory_saving_mode"`
	Cache            bool     `toml:"cache"`
	Connections      int      `toml:"connections"`
	TestMode         string   `toml:"test_mode"`
	LocalAddress     string   `toml:"local_address"`

	Log telegraf.Logger `toml:"-"`

	server       *speedtest.Server // The main(best) server
	servers      speedtest.Servers // Auxiliary servers
	serverFilter filter.Filter
	localAddr    *net.TCPAddr
}

func (*InternetSpeed) SampleConfig() string {
	return sampleConfig
}

func (is *InternetSpeed) Init() error {
	switch is.TestMode {
	case testModeSingle, testModeMulti:
	case "":
		is.TestMode = testModeSingle
	default:
		return fmt.Errorf("unrecognized test mode: %q", is.TestMode)
	}

	var err error
	is.serverFilter, err = filter.NewIncludeExcludeFilterDefaults(is.ServerIDInclude, is.ServerIDExclude, true, false)
	if err != nil {
		return fmt.Errorf("error compiling server ID filters: %w", err)
	}

	// The speedtest library only logs an unusable local address in its debug
	// output and then silently uses the default route, so reject such values
	// here. The address is parsed instead of resolved to keep Init off the
	// network; whether it is assigned is left to the bind during the tests,
	// which is retried every interval.
	if is.LocalAddress != "" {
		addr, err := netip.ParseAddr(is.LocalAddress)
		if err != nil {
			return fmt.Errorf("parsing local address failed: %w", err)
		}
		is.localAddr = net.TCPAddrFromAddrPort(netip.AddrPortFrom(addr, 0))
	}

	return nil
}

func (is *InternetSpeed) Gather(acc telegraf.Accumulator) error {
	// If not caching, go find the closest server each time.
	// We will find the best server as the main server. And
	// the remaining servers will be auxiliary candidates.
	if !is.Cache || is.server == nil {
		if err := is.findClosestServer(); err != nil {
			return fmt.Errorf("unable to find closest server: %w", err)
		}
	}

	err := is.server.PingTest(nil)
	if err != nil {
		return fmt.Errorf("ping test failed: %w", err)
	}

	// The analyzer creates its own dialers and does not use the source address
	// of the client, so set the dialers explicitly to keep the packet-loss test
	// on the same link as the other measurements. The sending timeout is set to
	// the library default as it is also the timeout of those dialers.
	options := &speedtest.PacketLossAnalyzerOptions{
		PacketSendingInterval: time.Millisecond * 100,
		PacketSendingTimeout:  time.Second * 5,
		SamplingDuration:      time.Second * 15,
	}
	if is.localAddr != nil {
		options.TCPDialer = &net.Dialer{Timeout: options.PacketSendingTimeout, LocalAddr: is.localAddr}
		options.UDPDialer = &net.Dialer{
			Timeout:   options.PacketSendingTimeout,
			LocalAddr: &net.UDPAddr{IP: is.localAddr.IP, Zone: is.localAddr.Zone},
		}
	}
	analyzer := speedtest.NewPacketLossAnalyzer(options)

	var pLoss *transport.PLoss

	if is.TestMode == testModeMulti {
		err = is.server.MultiDownloadTestContext(context.Background(), is.servers)
		if err != nil {
			return fmt.Errorf("download test failed: %w", err)
		}
		err = is.server.MultiUploadTestContext(context.Background(), is.servers)
		if err != nil {
			return fmt.Errorf("upload test failed: %w", err)
		}
		// Not all servers are applicable for packet loss testing.
		// If err != nil, we skip it and just report a warning.
		pLoss, err = analyzer.RunMulti(is.servers.Hosts())
		if err != nil {
			is.Log.Warnf("packet loss test failed: %s", err)
		}
	} else {
		err = is.server.DownloadTest()
		if err != nil {
			return fmt.Errorf("download test failed: %w", err)
		}
		err = is.server.UploadTest()
		if err != nil {
			return fmt.Errorf("upload test failed: %w", err)
		}
		// Not all servers are applicable for packet loss testing.
		// If err != nil, we skip it and just report a warning.
		err = analyzer.Run(is.server.Host, func(pl *transport.PLoss) {
			pLoss = pl
		})
		if err != nil {
			is.Log.Warnf("packet loss test failed: %s", err)
		}
	}

	packetLoss := -1.0
	if pLoss != nil {
		packetLoss = pLoss.LossPercent()
	}

	fields := map[string]any{
		"download":    is.server.DLSpeed.Mbps(),
		"upload":      is.server.ULSpeed.Mbps(),
		"latency":     timeDurationMillisecondToFloat64(is.server.Latency),
		"jitter":      timeDurationMillisecondToFloat64(is.server.Jitter),
		"packet_loss": packetLoss,
		"location":    is.server.Name,
	}
	tags := map[string]string{
		"server_id": is.server.ID,
		"source":    is.server.Host,
		"test_mode": is.TestMode,
	}
	// Recycle the history of each test to prevent data backlog.
	is.server.Context.Reset()
	acc.AddFields(measurement, fields, tags)
	return nil
}

func (is *InternetSpeed) findClosestServer() error {
	proto := speedtest.HTTP
	if os.Getegid() <= 0 {
		proto = speedtest.ICMP
	}

	client := speedtest.New(speedtest.WithUserConfig(&speedtest.UserConfig{
		UserAgent:  internal.ProductToken(),
		PingMode:   proto,
		SavingMode: is.MemorySavingMode,
		Source:     is.LocalAddress,
	}))
	if is.Connections > 0 {
		client.SetNThread(is.Connections)
	}

	var err error
	is.servers, err = client.FetchServers()
	if err != nil {
		return fmt.Errorf("fetching server list failed: %w", err)
	}

	if len(is.servers) < 1 {
		return errors.New("no servers found")
	}

	return is.selectServer()
}

func (is *InternetSpeed) selectServer() error {
	// Select the server with the lowest latency matching the filter.
	// If no server has latency info, use the first match (closest by distance).
	var minLatency int64 = math.MaxInt64
	selectIndex := -1
	for index, server := range is.servers {
		if !is.serverFilter.Match(server.ID) {
			continue
		}
		if selectIndex == -1 {
			// Select the first server we found and store the latency if it is valid
			selectIndex = index
			if server.Latency > 0 {
				minLatency = server.Latency.Milliseconds()
			}
		} else if server.Latency > 0 && server.Latency.Milliseconds() < minLatency {
			// Select the server if it has a lower latency than the previous one
			selectIndex = index
			minLatency = server.Latency.Milliseconds()
		}
	}

	if selectIndex == -1 {
		return errors.New("no server set: filter excluded all servers or no available server found")
	}

	is.server = is.servers[selectIndex]
	is.Log.Debugf("using server %s in %s (%s)\n", is.server.ID, is.server.Name, is.server.Host)
	return nil
}

func timeDurationMillisecondToFloat64(d time.Duration) float64 {
	return float64(d) / float64(time.Millisecond)
}

func init() {
	inputs.Add("internet_speed", func() telegraf.Input {
		return &InternetSpeed{}
	})
}
