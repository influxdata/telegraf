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

var sampleConfig = `
  ## URL to connect to
  ## ex: address = "tcp://127.0.0.1:8094"
  ## ex: address = "tcp4://127.0.0.1:8094"
  ## ex: address = "tcp6://127.0.0.1:8094"
  ## ex: address = "tcp6://[2001:db8::1]:8094"
  ## ex: address = "udp://127.0.0.1:8094"
  ## ex: address = "udp4://127.0.0.1:8094"
  ## ex: address = "udp6://127.0.0.1:8094"
  address = "tcp://127.0.0.1:8094"

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false

  ## Period between keep alive probes.
  ## Only applies to TCP sockets.
  ## 0 disables keep alive probes.
  ## Defaults to the OS configuration.
  # keep_alive_period = "5m"

  ## The framing technique with which it is expected that messages are
  ## transported (default = "octet-counting").  Whether the messages come
  ## using the octect-counting (RFC5425#section-4.3.1, RFC6587#section-3.4.1),
  ## or the non-transparent framing technique (RFC6587#section-3.4.2).  Must
  ## be one of "octet-counting", "non-transparent".
  # framing = "octet-counting"

  ## The trailer to be expected in case of non-transparent framing (default = "LF").
  ## Must be one of "LF", or "NUL".
  # trailer = "LF"

  ## SD-PARAMs settings
  ## Syslog messages can contain key/value pairs within zero or more
  ## structured data sections.  For each unrecognized metric tag/field a
  ## SD-PARAMS is created.
  ##
  ## Example:
  ##   [[outputs.syslog]]
  ##     sdparam_separator = "_"
  ##     default_sdid = "default@32473"
  ##     sdids = ["foo@123", "bar@456"]
  ##
  ##   input => xyzzy,x=y foo@123_value=42,bar@456_value2=84,something_else=1
  ##   output (structured data only) => [foo@123 value=42][bar@456 value2=84][default@32473 something_else=1 x=y]

  ## SD-PARAMs separator between the sdid and tag/field key (default = "_")
  # sdparam_separator = "_"

  ## Default sdid used for tags/fields that don't contain a prefix defined in
  ## the explicit sdids setting below If no default is specified, no SD-PARAMs
  ## will be used for unrecognized field.
  # default_sdid = "default@32473"

  ## List of explicit prefixes to extract from tag/field keys and use as the
  ## SDID, if they match (see above example for more details):
  # sdids = ["foo@123", "bar@456"]

  ## Default severity value. Severity and Facility are used to calculate the
  ## message PRI value (RFC5424#section-6.2.1).  Used when no metric field
  ## with key "severity_code" is defined.  If unset, 5 (notice) is the default
  # default_severity_code = 5

  ## Default facility value. Facility and Severity are used to calculate the
  ## message PRI value (RFC5424#section-6.2.1).  Used when no metric field with
  ## key "facility_code" is defined.  If unset, 1 (user-level) is the default
  # default_facility_code = 1

  ## Default APP-NAME value (RFC5424#section-6.2.5)
  ## Used when no metric tag with key "appname" is defined.
  ## If unset, "Telegraf" is the default
  # default_appname = "Telegraf"
`

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

func (s *Syslog) SampleConfig() string {
	return sampleConfig
}

func (s *Syslog) Description() string {
	return "Configuration for Syslog server to send metrics to"
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
