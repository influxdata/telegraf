//go:generate ../../../tools/readme_config_includer/generator
package socket_listener

import (
	_ "embed"
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strings"
	"sync"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	tlsint "github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"
)

//go:embed sample.conf
var sampleConfig string

type listener interface {
	listen(acc telegraf.Accumulator)
	addr() net.Addr
	close() error
}

type SocketListener struct {
	ServiceAddress  string           `toml:"service_address"`
	MaxConnections  int              `toml:"max_connections"`
	ReadBufferSize  config.Size      `toml:"read_buffer_size"`
	ReadTimeout     config.Duration  `toml:"read_timeout"`
	KeepAlivePeriod *config.Duration `toml:"keep_alive_period"`
	SocketMode      string           `toml:"socket_mode"`
	ContentEncoding string           `toml:"content_encoding"`
	Log             telegraf.Logger  `toml:"-"`
	tlsint.ServerConfig

	wg     sync.WaitGroup
	parser parsers.Parser

	listener listener
}

func (*SocketListener) SampleConfig() string {
	return sampleConfig
}

func (sl *SocketListener) Gather(_ telegraf.Accumulator) error {
	return nil
}

func (sl *SocketListener) SetParser(parser telegraf.Parser) {
	sl.parser = parser
}

func (sl *SocketListener) Start(acc telegraf.Accumulator) error {
	// Resolve the interface to an address if any given
	var ifname string
	ifregex := regexp.MustCompile(`%([\w\.]+)`)
	if matches := ifregex.FindStringSubmatch(sl.ServiceAddress); len(matches) == 2 {
		ifname := matches[1]
		sl.ServiceAddress = strings.Replace(sl.ServiceAddress, "%"+ifname, "", 1)
	}

	// Preparing TLS configuration
	tlsCfg, err := sl.ServerConfig.TLSConfig()
	if err != nil {
		return fmt.Errorf("getting TLS config failed: %w", err)
	}

	// Setup the network connection
	u, err := url.Parse(sl.ServiceAddress)
	if err != nil {
		return fmt.Errorf("parsing address failed: %w", err)
	}

	switch u.Scheme {
	case "tcp", "tcp4", "tcp6":
		ssl := &streamListener{
			ReadBufferSize:  int(sl.ReadBufferSize),
			ReadTimeout:     sl.ReadTimeout,
			KeepAlivePeriod: sl.KeepAlivePeriod,
			MaxConnections:  sl.MaxConnections,
			Encoding:        sl.ContentEncoding,
			Parser:          sl.parser,
			Log:             sl.Log,
		}

		if err := ssl.setupTCP(u, tlsCfg); err != nil {
			return err
		}
		sl.listener = ssl
	case "unix", "unixpacket":
		ssl := &streamListener{
			ReadBufferSize:  int(sl.ReadBufferSize),
			ReadTimeout:     sl.ReadTimeout,
			KeepAlivePeriod: sl.KeepAlivePeriod,
			MaxConnections:  sl.MaxConnections,
			Encoding:        sl.ContentEncoding,
			Parser:          sl.parser,
			Log:             sl.Log,
		}

		if err := ssl.setupUnix(u, tlsCfg, sl.SocketMode); err != nil {
			return err
		}
		sl.listener = ssl

	case "udp", "udp4", "udp6":
		psl := &packetListener{
			Encoding: sl.ContentEncoding,
			Parser:   sl.parser,
		}
		if err := psl.setupUDP(u, ifname, int(sl.ReadBufferSize)); err != nil {
			return err
		}
		sl.listener = psl
	case "ip", "ip4", "ip6":
		psl := &packetListener{
			Encoding: sl.ContentEncoding,
			Parser:   sl.parser,
		}
		if err := psl.setupIP(u); err != nil {
			return err
		}
		sl.listener = psl
	case "unixgram":
		psl := &packetListener{
			Encoding: sl.ContentEncoding,
			Parser:   sl.parser,
		}
		if err := psl.setupUnixgram(u, sl.SocketMode); err != nil {
			return err
		}
		sl.listener = psl
	default:
		return fmt.Errorf("unknown protocol %q in %q", u.Scheme, sl.ServiceAddress)
	}

	sl.Log.Infof("Listening on %s://%s", u.Scheme, sl.listener.addr())

	sl.wg.Add(1)
	go func() {
		defer sl.wg.Done()
		sl.listener.listen(acc)
	}()

	return nil
}

func (sl *SocketListener) Stop() {
	if sl.listener != nil {
		// Ignore the returned error as we cannot do anything about it anyway
		_ = sl.listener.close()
		sl.listener = nil
	}
	sl.wg.Wait()
}

func init() {
	inputs.Add("socket_listener", func() telegraf.Input {
		return &SocketListener{}
	})
}
