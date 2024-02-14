package socket

import (
	"bufio"
	"crypto/tls"
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
)

type listener interface {
	address() net.Addr
	listen()
	close() error
}

type lengthFieldSpec struct {
	Offset       int64  `toml:"offset"`
	Bytes        int64  `toml:"bytes"`
	Endianness   string `toml:"endianness"`
	HeaderLength int64  `toml:"header_length"`
	converter    func([]byte) int
}

type CallbackData func([]byte)
type CallbackError func(error)

type Config struct {
	MaxConnections       int              `toml:"max_connections"`
	ReadBufferSize       config.Size      `toml:"read_buffer_size"`
	ReadTimeout          config.Duration  `toml:"read_timeout"`
	KeepAlivePeriod      *config.Duration `toml:"keep_alive_period"`
	SocketMode           string           `toml:"socket_mode"`
	ContentEncoding      string           `toml:"content_encoding"`
	MaxDecompressionSize config.Size      `toml:"max_decompression_size"`
	SplittingStrategy    string           `toml:"splitting_strategy"`
	SplittingDelimiter   string           `toml:"splitting_delimiter"`
	SplittingLength      int              `toml:"splitting_length"`
	SplittingLengthField lengthFieldSpec  `toml:"splitting_length_field"`
	tlsint.ServerConfig
}

type Socket struct {
	Config

	url           *url.URL
	interfaceName string
	tlsCfg        *tls.Config
	log           telegraf.Logger

	splitter bufio.SplitFunc
	wg       sync.WaitGroup

	listener listener
}

func (cfg *Config) NewSocket(address string, logger telegraf.Logger) (*Socket, error) {
	s := &Socket{
		Config: *cfg,
		log:    logger,
	}

	switch s.SplittingStrategy {
	case "", "newline":
		s.splitter = bufio.ScanLines
	case "null":
		s.splitter = scanNull
	case "delimiter":
		re := regexp.MustCompile(`(\s*0?x)`)
		d := re.ReplaceAllString(strings.ToLower(s.SplittingDelimiter), "")
		delimiter, err := hex.DecodeString(d)
		if err != nil {
			return nil, fmt.Errorf("decoding delimiter failed: %w", err)
		}
		s.splitter = createScanDelimiter(delimiter)
	case "fixed length":
		s.splitter = createScanFixedLength(s.SplittingLength)
	case "variable length":
		// Create the converter function
		var order binary.ByteOrder
		switch strings.ToLower(s.SplittingLengthField.Endianness) {
		case "", "be":
			order = binary.BigEndian
		case "le":
			order = binary.LittleEndian
		default:
			return nil, fmt.Errorf("invalid 'endianness' %q", s.SplittingLengthField.Endianness)
		}

		switch s.SplittingLengthField.Bytes {
		case 1:
			s.SplittingLengthField.converter = func(b []byte) int {
				return int(b[0])
			}
		case 2:
			s.SplittingLengthField.converter = func(b []byte) int {
				return int(order.Uint16(b))
			}
		case 4:
			s.SplittingLengthField.converter = func(b []byte) int {
				return int(order.Uint32(b))
			}
		case 8:
			s.SplittingLengthField.converter = func(b []byte) int {
				return int(order.Uint64(b))
			}
		default:
			s.SplittingLengthField.converter = func(b []byte) int {
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
		s.splitter = createScanVariableLength(s.SplittingLengthField)
	default:
		return nil, fmt.Errorf("unknown 'splitting_strategy' %q", s.SplittingStrategy)
	}

	// Resolve the interface to an address if any given
	ifregex := regexp.MustCompile(`%([\w\.]+)`)
	if matches := ifregex.FindStringSubmatch(address); len(matches) == 2 {
		s.interfaceName = matches[1]
		address = strings.Replace(address, "%"+s.interfaceName, "", 1)
	}

	// Preparing TLS configuration
	tlsCfg, err := s.ServerConfig.TLSConfig()
	if err != nil {
		return nil, fmt.Errorf("getting TLS config failed: %w", err)
	}
	s.tlsCfg = tlsCfg

	// Parse and check the address
	u, err := url.Parse(address)
	if err != nil {
		return nil, fmt.Errorf("parsing address failed: %w", err)
	}
	s.url = u

	switch s.url.Scheme {
	case "tcp", "tcp4", "tcp6", "unix", "unixpacket",
		"udp", "udp4", "udp6", "ip", "ip4", "ip6", "unixgram", "vsock":
	default:
		return nil, fmt.Errorf("unknown protocol %q in %q", u.Scheme, address)
	}

	return s, nil
}

func (s *Socket) Listen(onData CallbackData, onError CallbackError) error {
	switch s.url.Scheme {
	case "tcp", "tcp4", "tcp6":
		l := &streamListener{
			ReadBufferSize:  int(s.ReadBufferSize),
			ReadTimeout:     s.ReadTimeout,
			KeepAlivePeriod: s.KeepAlivePeriod,
			MaxConnections:  s.MaxConnections,
			Encoding:        s.ContentEncoding,
			Splitter:        s.splitter,
			OnData:          onData,
			OnError:         onError,
			Log:             s.log,
		}

		if err := l.setupTCP(s.url, s.tlsCfg); err != nil {
			return err
		}
		s.listener = l
	case "unix", "unixpacket":
		l := &streamListener{
			ReadBufferSize:  int(s.ReadBufferSize),
			ReadTimeout:     s.ReadTimeout,
			KeepAlivePeriod: s.KeepAlivePeriod,
			MaxConnections:  s.MaxConnections,
			Encoding:        s.ContentEncoding,
			Splitter:        s.splitter,
			OnData:          onData,
			OnError:         onError,
			Log:             s.log,
		}

		if err := l.setupUnix(s.url, s.tlsCfg, s.SocketMode); err != nil {
			return err
		}
		s.listener = l
	case "udp", "udp4", "udp6":
		l := &packetListener{
			Encoding:             s.ContentEncoding,
			MaxDecompressionSize: int64(s.MaxDecompressionSize),
			OnData:               onData,
			OnError:              onError,
		}
		if err := l.setupUDP(s.url, s.interfaceName, int(s.ReadBufferSize)); err != nil {
			return err
		}
		s.listener = l
	case "ip", "ip4", "ip6":
		l := &packetListener{
			Encoding:             s.ContentEncoding,
			MaxDecompressionSize: int64(s.MaxDecompressionSize),
			OnData:               onData,
			OnError:              onError,
		}
		if err := l.setupIP(s.url); err != nil {
			return err
		}
		s.listener = l
	case "unixgram":
		l := &packetListener{
			Encoding:             s.ContentEncoding,
			MaxDecompressionSize: int64(s.MaxDecompressionSize),
			OnData:               onData,
			OnError:              onError,
		}
		if err := l.setupUnixgram(s.url, s.SocketMode); err != nil {
			return err
		}
		s.listener = l
	case "vsock":
		l := &streamListener{
			ReadBufferSize:  int(s.ReadBufferSize),
			ReadTimeout:     s.ReadTimeout,
			KeepAlivePeriod: s.KeepAlivePeriod,
			MaxConnections:  s.MaxConnections,
			Encoding:        s.ContentEncoding,
			Splitter:        s.splitter,
			OnData:          onData,
			OnError:         onError,
			Log:             s.log,
		}

		if err := l.setupVsock(s.url); err != nil {
			return err
		}
		s.listener = l
	default:
		return fmt.Errorf("unknown protocol %q", s.url.Scheme)
	}

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.listener.listen()
	}()

	return nil
}

func (s *Socket) Close() {
	if s.listener != nil {
		// Ignore the returned error as we cannot do anything about it anyway
		if err := s.listener.close(); err != nil {
			s.log.Warnf("Closing socket failed: %v", err)
		}
		s.listener = nil
	}
	s.wg.Wait()
}

func (s *Socket) Address() net.Addr {
	return s.listener.address()
}
