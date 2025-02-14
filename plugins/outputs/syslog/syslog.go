//go:generate ../../../tools/readme_config_includer/generator
package syslog

import (
	"crypto/tls"
	_ "embed"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/leodido/go-syslog/v4/nontransparent"
	"github.com/leodido/go-syslog/v4/rfc5424"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	common_tls "github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/outputs"
)

//go:embed sample.conf
var sampleConfig string

type Syslog struct {
	Address             string
	KeepAlivePeriod     *config.Duration
	DefaultSdid         string
	DefaultSeverityCode uint8
	DefaultFacilityCode uint8
	DefaultAppname      string
	Sdids               []string
	Separator           string `toml:"sdparam_separator"`
	Framing             string `toml:"framing"`
	Trailer             nontransparent.TrailerType
	Log                 telegraf.Logger `toml:"-"`
	net.Conn
	common_tls.ClientConfig
	mapper *SyslogMapper
}

func (*Syslog) SampleConfig() string {
	return sampleConfig
}
func (s *Syslog) Init() error {
	// Check framing and set default
	switch s.Framing {
	case "":
		s.Framing = "octet-counting"
	case "octet-counting", "non-transparent":
	default:
		return fmt.Errorf("invalid 'framing' %q", s.Framing)
	}
	return nil
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
		return &internal.StartupError{Err: err, Retry: true}
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
		if err := s.Connect(); err != nil {
			return err
		}
	}
	for _, metric := range metrics {
		msg, err := s.mapper.MapMetricToSyslogMessage(metric)
		if err != nil {
			s.Log.Errorf("Failed to create syslog message: %v", err)
			continue
		}

		msgBytesWithFraming, err := s.getSyslogMessageBytesWithFraming(msg)
		if err != nil {
			s.Log.Errorf("Failed to convert syslog message with framing: %v", err)
			continue
		}
		if _, err = s.Conn.Write(msgBytesWithFraming); err != nil {
			var netErr net.Error
			if errors.As(err, &netErr) {
				s.Close()
				s.Conn = nil
				return fmt.Errorf("closing connection: %w", netErr)
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

	if s.Framing == "octet-counting" {
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
