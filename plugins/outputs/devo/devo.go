package devo

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	tlsint "github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/influxdata/telegraf/plugins/serializers"
)

type DevoWriter struct {
  Address         		string
	ContentEncoding 		string `toml:"content_encoding"`
	DefaultHostname			string
	DefaultSeverityCode uint8
	DefaultFacilityCode uint8
	DefaultTag					string
	KeepAlivePeriod 		*internal.Duration
	mapper							*DevoMapper
  encoder 						internal.ContentEncoder
  net.Conn
	serializers.Serializer
	tlsint.ClientConfig
}

func (dw *DevoWriter) Description() string {
	return "Devo writer which formats messages to specified encoding and send to Devo"
}

func (dw *DevoWriter) SampleConfig() string {
	return `
  ## URL to connect to
  # address = "tcp://127.0.0.1:8094"
  # address = "tcp://example.com:http"
  # address = "tcp4://127.0.0.1:8094"

  ## Optional TLS Config
  # tls_ca = "/etc/telegraf/ca.pem"
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false

	## Default severity value. Severity and Facility are used to calculate the
  ## message PRI value (RFC5424#section-6.2.1).  Used when no metric field
  ## with key "severity_code" is defined.  If unset, 5 (notice) is the default
  # default_severity_code = 5

  ## Default facility value. Facility and Severity are used to calculate the
  ## message PRI value (RFC5424#section-6.2.1).  Used when no metric field with
  ## key "facility_code" is defined.  If unset, 1 (user-level) is the default
  # default_facility_code = 1

  ## Period between keep alive probes.
  ## Only applies to TCP sockets.
  ## 0 disables keep alive probes.
  ## Defaults to the OS configuration.
  # keep_alive_period = "5m"

  ## Content encoding for packet-based connections (i.e. UDP, unixgram).
  ## Can be set to "gzip" or to "identity" to apply no encoding.
  ##
  # content_encoding = "identity"

  ## Data format to generate.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  # data_format = "json"

	## Default Devo tag value
	## Used when no metric tag with key "devo_tag" is defined.
	## If unset, "my.app.telegraf.default" is the default
  ## refer here for more information:
  ## https://docs.devo.com/confluence/ndt/parsers-and-collectors/about-devo-tags
	# default_tag = "my.app.telegraf.default"

	## You can also manually set your hostname to identify where these metrics come from
	## if your logs do not have identifiable information attached to them. Otherwise
	## the plugin will try to get the hostname from your OS directly.
	# default_hostname = "unknown"
`
}

func (dw *DevoWriter) SetSerializer(s serializers.Serializer) {
	dw.Serializer = s
}

func (dw *DevoWriter) Connect() error {
	dw.initializeDevoMapper()

	spl := strings.SplitN(dw.Address, "://", 2)
	if len(spl) != 2 {
		return fmt.Errorf("invalid address: %s", dw.Address)
	}

	tlsCfg, err := dw.ClientConfig.TLSConfig()
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

	if err := dw.setKeepAlive(c); err != nil {
		log.Printf("unable to configure keep alive (%s): %s", dw.Address, err)
	}
	//set encoder
	dw.encoder, err = internal.NewContentEncoder(dw.ContentEncoding)
	if err != nil {
		return err
	}

	dw.Conn = c
	return nil
}

func (dw *DevoWriter) setKeepAlive(c net.Conn) error {
	if dw.KeepAlivePeriod == nil {
		return nil
	}
	tcpc, ok := c.(*net.TCPConn)
	if !ok {
		return fmt.Errorf("cannot set keep alive on a %s socket", strings.SplitN(dw.Address, "://", 2)[0])
	}
	if dw.KeepAlivePeriod.Duration == 0 {
		return tcpc.SetKeepAlive(false)
	}
	if err := tcpc.SetKeepAlive(true); err != nil {
		return err
	}
	return tcpc.SetKeepAlivePeriod(dw.KeepAlivePeriod.Duration)
}

// Write writes the given metrics to the destination.
// If an error is encountered, it is up to the caller to retry the same write again later.
// Not parallel safe.
func (dw *DevoWriter) Write(metrics []telegraf.Metric) error {
	if dw.Conn == nil {
		// previous write failed with permanent error and socket was closed.
		if err := dw.Connect(); err != nil {
			return err
		}
	}

	for _, m := range metrics {
		bs, err := dw.Serialize(m)
		if err != nil {
			log.Printf("D! [outputs.devo] Could not serialize metric: %v", err)
			continue
		}

		bs, err = dw.mapper.devoMapper(m, bs)
		if err != nil {
			log.Printf("D! [outputs.devo] Could not devo serialize metric: %v", err)
			continue
		}

		bs, err = dw.encoder.Encode(bs)
		if err != nil {
			log.Printf("D! [outputs.devo] Could not encode metric: %v", err)
			continue
		}

		if _, err := dw.Conn.Write(bs); err != nil {
			//TODO log & keep going with remaining strings
			if err, ok := err.(net.Error); !ok || !err.Temporary() {
				// permanent error. close the connection
				dw.Close()
				dw.Conn = nil
				return fmt.Errorf("closing connection: %v", err)
			}
			return err
		}
	}

	return nil
}

// Close closes the connection. Noop if already closed.
func (dw *DevoWriter) Close() error {
	if dw.Conn == nil {
		return nil
	}
	err := dw.Conn.Close()
	dw.Conn = nil
	return err
}

func (s *DevoWriter) initializeDevoMapper() {
	if s.mapper != nil {
		return
	}
	s.mapper = &DevoMapper{
		DefaultTag: s.DefaultTag,
		DefaultSeverityCode: s.DefaultSeverityCode,
		DefaultFacilityCode: s.DefaultFacilityCode,
		DefaultHostname: s.DefaultHostname,
	}
}

func newDevoWriter() *DevoWriter {
	s, _ := serializers.NewInfluxSerializer()
	return &DevoWriter{
		Serializer: s,
		DefaultTag: "my.app.telegraf.untagged",
		DefaultSeverityCode: uint8(5), // notice
		DefaultFacilityCode: uint8(1), //user-level
		DefaultHostname: "unknown",
	}
}

func init() {
	outputs.Add("devo", func() telegraf.Output { return newDevoWriter() })
}
