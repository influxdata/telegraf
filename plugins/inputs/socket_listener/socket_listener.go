//go:generate ../../../tools/readme_config_includer/generator
package socket_listener

import (
	"bufio"
	_ "embed"
	"encoding/binary"
	"encoding/hex"
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

type lengthFieldSpec struct {
	Offset       int64  `toml:"offset"`
	Bytes        int64  `toml:"bytes"`
	Endianess    string `toml:"endianess"`
	HeaderLength int64  `toml:"header_length"`
	converter    func([]byte) int
}

type SocketListener struct {
	ServiceAddress       string           `toml:"service_address"`
	MaxConnections       int              `toml:"max_connections"`
	ReadBufferSize       config.Size      `toml:"read_buffer_size"`
	ReadTimeout          config.Duration  `toml:"read_timeout"`
	KeepAlivePeriod      *config.Duration `toml:"keep_alive_period"`
	SocketMode           string           `toml:"socket_mode"`
	ContentEncoding      string           `toml:"content_encoding"`
	SplittingStrategy    string           `toml:"splitting_strategy"`
	SplittingDelimiter   string           `toml:"splitting_delimiter"`
	SplittingLength      int              `toml:"splitting_length"`
	SplittingLengthField lengthFieldSpec  `toml:"splitting_length_field"`
	Log                  telegraf.Logger  `toml:"-"`
	tlsint.ServerConfig

	wg       sync.WaitGroup
	parser   parsers.Parser
	splitter bufio.SplitFunc

	listener listener
}

func (*SocketListener) SampleConfig() string {
	return sampleConfig
}

func (sl *SocketListener) Init() error {
	switch sl.SplittingStrategy {
	case "", "newline":
		sl.splitter = bufio.ScanLines
	case "null":
		sl.splitter = scanNull
	case "delimiter":
		re := regexp.MustCompile(`(\s*0?x)`)
		d := re.ReplaceAllString(strings.ToLower(sl.SplittingDelimiter), "")
		delimiter, err := hex.DecodeString(d)
		if err != nil {
			return fmt.Errorf("decoding delimiter failed: %w", err)
		}
		sl.splitter = createScanDelimiter(delimiter)
	case "fixed length":
		sl.splitter = createScanFixedLength(sl.SplittingLength)
	case "variable length":
		// Create the converter function
		var order binary.ByteOrder
		switch strings.ToLower(sl.SplittingLengthField.Endianess) {
		case "", "be":
			order = binary.BigEndian
		case "le":
			order = binary.LittleEndian
		default:
			return fmt.Errorf("invalid 'endianess' %q", sl.SplittingLengthField.Endianess)
		}

		switch sl.SplittingLengthField.Bytes {
		case 1:
			sl.SplittingLengthField.converter = func(b []byte) int {
				return int(b[0])
			}
		case 2:
			sl.SplittingLengthField.converter = func(b []byte) int {
				return int(order.Uint16(b))
			}
		case 4:
			sl.SplittingLengthField.converter = func(b []byte) int {
				return int(order.Uint32(b))
			}
		case 8:
			sl.SplittingLengthField.converter = func(b []byte) int {
				return int(order.Uint64(b))
			}
		default:
			sl.SplittingLengthField.converter = func(b []byte) int {
				buf := make([]byte, 8)
				start := 0
				if order == binary.BigEndian {
					start = 8 - len(b)
				}
				for i := 0; i < len(b); i++ {
					buf[start+i] = b[i]
				}
				return int(order.Uint64(buf))
			}
		}

		// Check if we have enough bytes in the header
		sl.splitter = createScanVariableLength(sl.SplittingLengthField)
	default:
		return fmt.Errorf("unknown 'splitting_strategy' %q", sl.SplittingStrategy)
	}
	return nil
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
			Splitter:        sl.splitter,
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
			Splitter:        sl.splitter,
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
