package syslog

import (
	"crypto/tls"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/influxdata/go-syslog/v3/nontransparent"
	"github.com/influxdata/go-syslog/v3/rfc5424"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	framing "github.com/influxdata/telegraf/internal/syslog"
	tlsint "github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/outputs"
)

type Syslog struct {
	Address             string
	KeepAlivePeriod     *config.Duration
	DefaultSdid         string
	DefaultSeverityCode uint8
	DefaultFacilityCode uint8
	DefaultAppname      string
	Sdids               []string
	Separator           string `toml:"sdparam_separator"`
	Framing             framing.Framing
	Trailer             nontransparent.TrailerType
	Log                 telegraf.Logger `toml:"-"`
	net.Conn
	tlsint.ClientConfig
	mapper *SyslogMapper
}

func (s *Syslog) Connect() error {
	s.initializeSyslogMapper()

	spl := strings.SplitN(s.Address, "://", 2)
	if len(spl) != 2 {
		return fmt.Errorf("invalid address: %s", s.Address)
	}

	tlsCfg, err := s.ClientConfig.TLSConfig()
	if err != nil {
		return err
	}

	var c net.Conn
	if tlsCfg == nil {
		c, err = net.Dial(spl[0], spl[1])
	} else {
		c, err = tls.Dial(spl[0], spl[1], tlsCfg)
	}
	if err != nil {
		return err
	}

	if err := s.setKeepAlive(c); err != nil {
		s.Log.Warnf("unable to configure keep alive (%s): %s", s.Address, err)
	}

	s.Conn = c
	return nil
}

func (s *Syslog) setKeepAlive(c net.Conn) error {
	if s.KeepAlivePeriod == nil {
		return nil
	}
	tcpc, ok := c.(*net.TCPConn)
	if !ok {
		return fmt.Errorf("cannot set keep alive on a %s socket", strings.SplitN(s.Address, "://", 2)[0])
	}
	if *s.KeepAlivePeriod == 0 {
		return tcpc.SetKeepAlive(false)
	}
	if err := tcpc.SetKeepAlive(true); err != nil {
		return err
	}
	return tcpc.SetKeepAlivePeriod(time.Duration(*s.KeepAlivePeriod))
}

func (s *Syslog) Close() error {
	if s.Conn == nil {
		return nil
	}
	err := s.Conn.Close()
	s.Conn = nil
	return err
}

func (s *Syslog) Write(metrics []telegraf.Metric) (err error) {
	if s.Conn == nil {
		// previous write failed with permanent error and socket was closed.
		if err = s.Connect(); err != nil {
			return err
		}
	}
	for _, metric := range metrics {
		var msg *rfc5424.SyslogMessage
		if msg, err = s.mapper.MapMetricToSyslogMessage(metric); err != nil {
			s.Log.Errorf("Failed to create syslog message: %v", err)
			continue
		}
		var msgBytesWithFraming []byte
		if msgBytesWithFraming, err = s.getSyslogMessageBytesWithFraming(msg); err != nil {
			s.Log.Errorf("Failed to convert syslog message with framing: %v", err)
			continue
		}
		if _, err = s.Conn.Write(msgBytesWithFraming); err != nil {
			if netErr, ok := err.(net.Error); !ok || !netErr.Temporary() {
				s.Close() //nolint:revive // There is another error which will be returned here
				s.Conn = nil
				return fmt.Errorf("closing connection: %v", netErr)
			}
			return err
		}
	}
	return nil
}

func (s *Syslog) getSyslogMessageBytesWithFraming(msg *rfc5424.SyslogMessage) ([]byte, error) {
	var msgString string
	var err error
	if msgString, err = msg.String(); err != nil {
		return nil, err
	}
	msgBytes := []byte(msgString)

	if s.Framing == framing.OctetCounting {
		return append([]byte(strconv.Itoa(len(msgBytes))+" "), msgBytes...), nil
	}
	// Non-transparent framing
	trailer, err := s.Trailer.Value()
	if err != nil {
		return nil, err
	}
	return append(msgBytes, byte(trailer)), nil
}

func (s *Syslog) initializeSyslogMapper() {
	if s.mapper != nil {
		return
	}
	s.mapper = newSyslogMapper()
	s.mapper.DefaultFacilityCode = s.DefaultFacilityCode
	s.mapper.DefaultSeverityCode = s.DefaultSeverityCode
	s.mapper.DefaultAppname = s.DefaultAppname
	s.mapper.Separator = s.Separator
	s.mapper.DefaultSdid = s.DefaultSdid
	s.mapper.Sdids = s.Sdids
}

func newSyslog() *Syslog {
	return &Syslog{
		Framing:             framing.OctetCounting,
		Trailer:             nontransparent.LF,
		Separator:           "_",
		DefaultSeverityCode: uint8(5), // notice
		DefaultFacilityCode: uint8(1), // user-level
		DefaultAppname:      "Telegraf",
	}
}

func init() {
	outputs.Add("syslog", func() telegraf.Output { return newSyslog() })
}
