package syslog

import (
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/influxdata/go-syslog/v3"
	"github.com/influxdata/go-syslog/v3/nontransparent"
	"github.com/influxdata/go-syslog/v3/octetcounting"
	"github.com/influxdata/go-syslog/v3/rfc3164"
	"github.com/influxdata/go-syslog/v3/rfc5424"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	framing "github.com/influxdata/telegraf/internal/syslog"
	tlsConfig "github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type syslogRFC string

const defaultReadTimeout = time.Second * 5
const ipMaxPacketSize = 64 * 1024
const syslogRFC3164 = "RFC3164"
const syslogRFC5424 = "RFC5424"

// Syslog is a syslog plugin
type Syslog struct {
	tlsConfig.ServerConfig
	Address         string `toml:"server"`
	KeepAlivePeriod *config.Duration
	MaxConnections  int
	ReadTimeout     *config.Duration
	Framing         framing.Framing
	SyslogStandard  syslogRFC
	Trailer         nontransparent.TrailerType
	BestEffort      bool
	Separator       string `toml:"sdparam_separator"`

	now      func() time.Time
	lastTime time.Time

	mu sync.Mutex
	wg sync.WaitGroup
	io.Closer

	isStream      bool
	tcpListener   net.Listener
	tlsConfig     *tls.Config
	connections   map[string]net.Conn
	connectionsMu sync.Mutex

	udpListener net.PacketConn
}

// Gather ...
func (s *Syslog) Gather(_ telegraf.Accumulator) error {
	return nil
}

// Start starts the service.
func (s *Syslog) Start(acc telegraf.Accumulator) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	scheme, host, err := getAddressParts(s.Address)
	if err != nil {
		return err
	}
	s.Address = host

	switch scheme {
	case "tcp", "tcp4", "tcp6", "unix", "unixpacket":
		s.isStream = true
	case "udp", "udp4", "udp6", "ip", "ip4", "ip6", "unixgram":
		s.isStream = false
	default:
		return fmt.Errorf("unknown protocol '%s' in '%s'", scheme, s.Address)
	}

	if scheme == "unix" || scheme == "unixpacket" || scheme == "unixgram" {
		// Accept success and failure in case the file does not exist
		//nolint:errcheck,revive
		os.Remove(s.Address)
	}

	if s.isStream {
		l, err := net.Listen(scheme, s.Address)
		if err != nil {
			return err
		}
		s.Closer = l
		s.tcpListener = l
		s.tlsConfig, err = s.TLSConfig()
		if err != nil {
			return err
		}

		s.wg.Add(1)
		go s.listenStream(acc)
	} else {
		l, err := net.ListenPacket(scheme, s.Address)
		if err != nil {
			return err
		}
		s.Closer = l
		s.udpListener = l

		s.wg.Add(1)
		go s.listenPacket(acc)
	}

	if scheme == "unix" || scheme == "unixpacket" || scheme == "unixgram" {
		s.Closer = unixCloser{path: s.Address, closer: s.Closer}
	}

	return nil
}

// Stop cleans up all resources
func (s *Syslog) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.Closer != nil {
		// Ignore the returned error as we cannot do anything about it anyway
		//nolint:errcheck,revive
		s.Close()
	}
	s.wg.Wait()
}

// getAddressParts returns the address scheme and host
// it also sets defaults for them when missing
// when the input address does not specify the protocol it returns an error
func getAddressParts(a string) (scheme string, host string, err error) {
	parts := strings.SplitN(a, "://", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("missing protocol within address '%s'", a)
	}

	u, err := url.Parse(filepath.ToSlash(a)) //convert backslashes to slashes (to make Windows path a valid URL)
	if err != nil {
		return "", "", fmt.Errorf("could not parse address '%s': %v", a, err)
	}
	switch u.Scheme {
	case "unix", "unixpacket", "unixgram":
		return parts[0], parts[1], nil
	}

	if u.Hostname() != "" {
		host = u.Hostname()
	}
	host += ":"
	if u.Port() == "" {
		host += "6514"
	} else {
		host += u.Port()
	}

	return u.Scheme, host, nil
}

func (s *Syslog) listenPacket(acc telegraf.Accumulator) {
	defer s.wg.Done()
	b := make([]byte, ipMaxPacketSize)
	var p syslog.Machine
	switch {
	case !s.BestEffort && s.SyslogStandard == syslogRFC5424:
		p = rfc5424.NewParser()
	case s.BestEffort && s.SyslogStandard == syslogRFC5424:
		p = rfc5424.NewParser(rfc5424.WithBestEffort())
	case !s.BestEffort && s.SyslogStandard == syslogRFC3164:
		p = rfc3164.NewParser(rfc3164.WithYear(rfc3164.CurrentYear{}))
	case s.BestEffort && s.SyslogStandard == syslogRFC3164:
		p = rfc3164.NewParser(rfc3164.WithYear(rfc3164.CurrentYear{}), rfc3164.WithBestEffort())
	}
	for {
		n, _, err := s.udpListener.ReadFrom(b)
		if err != nil {
			if !strings.HasSuffix(err.Error(), ": use of closed network connection") {
				acc.AddError(err)
			}
			break
		}

		message, err := p.Parse(b[:n])
		if message != nil {
			acc.AddFields("syslog", fields(message, s), tags(message), s.currentTime())
		}
		if err != nil {
			acc.AddError(err)
		}
	}
}

func (s *Syslog) listenStream(acc telegraf.Accumulator) {
	defer s.wg.Done()

	s.connections = map[string]net.Conn{}

	for {
		conn, err := s.tcpListener.Accept()
		if err != nil {
			if !strings.HasSuffix(err.Error(), ": use of closed network connection") {
				acc.AddError(err)
			}
			break
		}
		var tcpConn, _ = conn.(*net.TCPConn)
		if s.tlsConfig != nil {
			conn = tls.Server(conn, s.tlsConfig)
		}

		s.connectionsMu.Lock()
		if s.MaxConnections > 0 && len(s.connections) >= s.MaxConnections {
			s.connectionsMu.Unlock()
			if err := conn.Close(); err != nil {
				acc.AddError(err)
			}
			continue
		}
		s.connections[conn.RemoteAddr().String()] = conn
		s.connectionsMu.Unlock()

		if err := s.setKeepAlive(tcpConn); err != nil {
			acc.AddError(fmt.Errorf("unable to configure keep alive (%s): %s", s.Address, err))
		}

		go s.handle(conn, acc)
	}

	s.connectionsMu.Lock()
	for _, c := range s.connections {
		if err := c.Close(); err != nil {
			acc.AddError(err)
		}
	}
	s.connectionsMu.Unlock()
}

func (s *Syslog) removeConnection(c net.Conn) {
	s.connectionsMu.Lock()
	delete(s.connections, c.RemoteAddr().String())
	s.connectionsMu.Unlock()
}

func (s *Syslog) handle(conn net.Conn, acc telegraf.Accumulator) {
	defer func() {
		s.removeConnection(conn)
		// Ignore the returned error as we cannot do anything about it anyway
		//nolint:errcheck,revive
		conn.Close()
	}()

	var p syslog.Parser

	emit := func(r *syslog.Result) {
		s.store(*r, acc)
		if s.ReadTimeout != nil && time.Duration(*s.ReadTimeout) > 0 {
			if err := conn.SetReadDeadline(time.Now().Add(time.Duration(*s.ReadTimeout))); err != nil {
				acc.AddError(fmt.Errorf("setting read deadline failed: %v", err))
			}
		}
	}

	// Create parser options
	opts := []syslog.ParserOption{
		syslog.WithListener(emit),
	}
	if s.BestEffort {
		opts = append(opts, syslog.WithBestEffort())
	}

	// Select the parser to use depending on transport framing
	if s.Framing == framing.OctetCounting {
		// Octet counting transparent framing
		p = octetcounting.NewParser(opts...)
	} else {
		// Non-transparent framing
		opts = append(opts, nontransparent.WithTrailer(s.Trailer))
		p = nontransparent.NewParser(opts...)
	}

	p.Parse(conn)

	if s.ReadTimeout != nil && time.Duration(*s.ReadTimeout) > 0 {
		if err := conn.SetReadDeadline(time.Now().Add(time.Duration(*s.ReadTimeout))); err != nil {
			acc.AddError(fmt.Errorf("setting read deadline failed: %v", err))
		}
	}
}

func (s *Syslog) setKeepAlive(c *net.TCPConn) error {
	if s.KeepAlivePeriod == nil {
		return nil
	}

	if *s.KeepAlivePeriod == 0 {
		return c.SetKeepAlive(false)
	}
	if err := c.SetKeepAlive(true); err != nil {
		return err
	}
	return c.SetKeepAlivePeriod(time.Duration(*s.KeepAlivePeriod))
}

func (s *Syslog) store(res syslog.Result, acc telegraf.Accumulator) {
	if res.Error != nil {
		acc.AddError(res.Error)
	}
	if res.Message != nil {
		acc.AddFields("syslog", fields(res.Message, s), tags(res.Message), s.currentTime())
	}
}

func tags(msg syslog.Message) map[string]string {
	ts := map[string]string{}

	// Not checking assuming a minimally valid message
	ts["severity"] = *msg.SeverityShortLevel()
	ts["facility"] = *msg.FacilityLevel()

	switch m := msg.(type) {
	case *rfc5424.SyslogMessage:
		populateCommonTags(&m.Base, ts)
	case *rfc3164.SyslogMessage:
		populateCommonTags(&m.Base, ts)
	}
	return ts
}

func fields(msg syslog.Message, s *Syslog) map[string]interface{} {
	flds := map[string]interface{}{}

	switch m := msg.(type) {
	case *rfc5424.SyslogMessage:
		populateCommonFields(&m.Base, flds)
		// Not checking assuming a minimally valid message
		flds["version"] = m.Version

		if m.StructuredData != nil {
			for sdid, sdparams := range *m.StructuredData {
				if len(sdparams) == 0 {
					// When SD-ID does not have params we indicate its presence with a bool
					flds[sdid] = true
					continue
				}
				for name, value := range sdparams {
					// Using whitespace as separator since it is not allowed by the grammar within SDID
					flds[sdid+s.Separator+name] = value
				}
			}
		}
	case *rfc3164.SyslogMessage:
		populateCommonFields(&m.Base, flds)
	}

	return flds
}

func populateCommonFields(msg *syslog.Base, flds map[string]interface{}) {
	flds["facility_code"] = int(*msg.Facility)
	flds["severity_code"] = int(*msg.Severity)
	if msg.Timestamp != nil {
		flds["timestamp"] = (*msg.Timestamp).UnixNano()
	}
	if msg.ProcID != nil {
		flds["procid"] = *msg.ProcID
	}
	if msg.MsgID != nil {
		flds["msgid"] = *msg.MsgID
	}
	if msg.Message != nil {
		flds["message"] = strings.TrimRightFunc(*msg.Message, func(r rune) bool {
			return unicode.IsSpace(r)
		})
	}
}

func populateCommonTags(msg *syslog.Base, ts map[string]string) {
	if msg.Hostname != nil {
		ts["hostname"] = *msg.Hostname
	}
	if msg.Appname != nil {
		ts["appname"] = *msg.Appname
	}
}

type unixCloser struct {
	path   string
	closer io.Closer
}

func (uc unixCloser) Close() error {
	err := uc.closer.Close()
	// Accept success and failure in case the file does not exist
	//nolint:errcheck,revive
	os.Remove(uc.path)
	return err
}

func (s *Syslog) currentTime() time.Time {
	t := s.now()
	if t == s.lastTime {
		t = t.Add(time.Nanosecond)
	}
	s.lastTime = t
	return t
}

func getNanoNow() time.Time {
	return time.Unix(0, time.Now().UnixNano())
}

func init() {
	defaultTimeout := config.Duration(defaultReadTimeout)
	inputs.Add("syslog", func() telegraf.Input {
		return &Syslog{
			Address:        ":6514",
			now:            getNanoNow,
			ReadTimeout:    &defaultTimeout,
			Framing:        framing.OctetCounting,
			SyslogStandard: syslogRFC5424,
			Trailer:        nontransparent.LF,
			Separator:      "_",
		}
	})
}
