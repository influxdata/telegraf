//go:generate ../../../tools/readme_config_includer/generator
package chrony

import (
	_ "embed"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"time"

	fbchrony "github.com/facebook/time/ntp/chrony"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type Chrony struct {
	Server    string          `toml:"server"`
	Timeout   config.Duration `toml:"timeout"`
	DNSLookup bool            `toml:"dns_lookup"`
	Log       telegraf.Logger `toml:"-"`

	conn   net.Conn
	client *fbchrony.Client
}

func (*Chrony) SampleConfig() string {
	return sampleConfig
}

func (c *Chrony) Init() error {
	if c.Server != "" {
		// Check the specified server address
		u, err := url.Parse(c.Server)
		if err != nil {
			return fmt.Errorf("parsing server address failed: %w", err)
		}
		switch u.Scheme {
		case "unix":
			// Keep the server unmodified
		case "udp":
			// Check if we do have a port and add the default port if we don't
			if u.Port() == "" {
				u.Host += ":323"
			}
			// We cannot have path elements in an UDP address
			if u.Path != "" {
				return fmt.Errorf("path detected in UDP address %q", c.Server)
			}
			u = &url.URL{Scheme: "udp", Host: u.Host}
		default:
			return errors.New("unknown or missing address scheme")
		}
		c.Server = u.String()
	}

	return nil
}

func (c *Chrony) Start(_ telegraf.Accumulator) error {
	if c.Server != "" {
		// Create a connection
		u, err := url.Parse(c.Server)
		if err != nil {
			return fmt.Errorf("parsing server address failed: %w", err)
		}
		switch u.Scheme {
		case "unix":
			conn, err := net.DialTimeout("unix", u.Path, time.Duration(c.Timeout))
			if err != nil {
				return fmt.Errorf("dialing %q failed: %w", c.Server, err)
			}
			c.conn = conn
		case "udp":
			conn, err := net.DialTimeout("udp", u.Host, time.Duration(c.Timeout))
			if err != nil {
				return fmt.Errorf("dialing %q failed: %w", c.Server, err)
			}
			c.conn = conn
		}
	} else {
		// If no server is given, reproduce chronyc's behavior
		if conn, err := net.DialTimeout("unix", "/run/chrony/chronyd.sock", time.Duration(c.Timeout)); err == nil {
			c.Server = "unix:///run/chrony/chronyd.sock"
			c.conn = conn
		} else if conn, err := net.DialTimeout("udp", "127.0.0.1:323", time.Duration(c.Timeout)); err == nil {
			c.Server = "udp://127.0.0.1:323"
			c.conn = conn
		} else {
			conn, err := net.DialTimeout("udp", "[::1]:323", time.Duration(c.Timeout))
			if err != nil {
				return fmt.Errorf("dialing server failed: %w", err)
			}
			c.Server = "udp://[::1]:323"
			c.conn = conn
		}
	}
	c.Log.Debugf("Connected to %q...", c.Server)

	// Initialize the client
	c.client = &fbchrony.Client{Connection: c.conn}

	return nil
}

func (c *Chrony) Stop() {
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			c.Log.Errorf("Closing connection to %q failed: %v", c.Server, err)
		}
	}
}

func (c *Chrony) Gather(acc telegraf.Accumulator) error {
	req := fbchrony.NewTrackingPacket()
	resp, err := c.client.Communicate(req)
	if err != nil {
		return fmt.Errorf("querying tracking data failed: %w", err)
	}
	tracking, ok := resp.(*fbchrony.ReplyTracking)
	if !ok {
		return fmt.Errorf("got unexpected response type %T while waiting for tracking data", resp)
	}

	// according to https://github.com/mlichvar/chrony/blob/e11b518a1ffa704986fb1f1835c425844ba248ef/ntp.h#L70
	var leapStatus string
	switch tracking.LeapStatus {
	case 0:
		leapStatus = "normal"
	case 1:
		leapStatus = "insert second"
	case 2:
		leapStatus = "delete second"
	case 3:
		leapStatus = "not synchronized"
	}

	tags := map[string]string{
		"leap_status":  leapStatus,
		"reference_id": fbchrony.RefidAsHEX(tracking.RefID),
		"stratum":      strconv.FormatUint(uint64(tracking.Stratum), 10),
	}
	fields := map[string]interface{}{
		"frequency":       tracking.FreqPPM,
		"system_time":     tracking.CurrentCorrection,
		"last_offset":     tracking.LastOffset,
		"residual_freq":   tracking.ResidFreqPPM,
		"rms_offset":      tracking.RMSOffset,
		"root_delay":      tracking.RootDelay,
		"root_dispersion": tracking.RootDispersion,
		"skew":            tracking.SkewPPM,
		"update_interval": tracking.LastUpdateInterval,
	}
	acc.AddFields("chrony", fields, tags)

	return nil
}
func init() {
	inputs.Add("chrony", func() telegraf.Input {
		return &Chrony{Timeout: config.Duration(3 * time.Second)}
	})
}
