//go:generate ../../../tools/readme_config_includer/generator
package chrony

import (
	_ "embed"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"syscall"
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
	Metrics   []string        `toml:"metrics"`
	Log       telegraf.Logger `toml:"-"`

	conn   net.Conn
	client *fbchrony.Client
	source string
}

func (*Chrony) SampleConfig() string {
	return sampleConfig
}

func (c *Chrony) Init() error {
	// Use the configured server, if none set, we try to guess it in Start()
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

	// Check the given metrics
	if len(c.Metrics) == 0 {
		c.Metrics = append(c.Metrics, "tracking")
	}
	for _, m := range c.Metrics {
		switch m {
		case "activity", "tracking":
			// Do nothing as those are valid
		default:
			return fmt.Errorf("invalid metric setting %q", m)
		}
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
			c.source = u.Path
		case "udp":
			conn, err := net.DialTimeout("udp", u.Host, time.Duration(c.Timeout))
			if err != nil {
				return fmt.Errorf("dialing %q failed: %w", c.Server, err)
			}
			c.conn = conn
			c.source = u.Host
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
		if err := c.conn.Close(); err != nil && !errors.Is(err, net.ErrClosed) && !errors.Is(err, syscall.EPIPE) {
			c.Log.Errorf("Closing connection to %q failed: %v", c.Server, err)
		}
	}
}

func (c *Chrony) Gather(acc telegraf.Accumulator) error {
	for _, m := range c.Metrics {
		switch m {
		case "activity":
			acc.AddError(c.gatherActivity(acc))
		case "tracking":
			acc.AddError(c.gatherTracking(acc))
		default:
			return fmt.Errorf("invalid metric setting %q", m)
		}
	}

	return nil
}

func (c *Chrony) gatherActivity(acc telegraf.Accumulator) error {
	req := fbchrony.NewActivityPacket()
	r, err := c.client.Communicate(req)
	if err != nil {
		return fmt.Errorf("querying activity data failed: %w", err)
	}
	resp, ok := r.(*fbchrony.ReplyActivity)
	if !ok {
		return fmt.Errorf("got unexpected response type %T while waiting for tracking data", resp)
	}

	tags := map[string]string{}
	if c.source != "" {
		tags["source"] = c.source
	}

	fields := map[string]interface{}{
		"online":        resp.Online,
		"offline":       resp.Offline,
		"burst_online":  resp.BurstOnline,
		"burst_offline": resp.BurstOffline,
		"unresolved":    resp.Unresolved,
	}
	acc.AddFields("chrony_activity", fields, tags)

	return nil
}

func (c *Chrony) gatherTracking(acc telegraf.Accumulator) error {
	req := fbchrony.NewTrackingPacket()
	r, err := c.client.Communicate(req)
	if err != nil {
		return fmt.Errorf("querying tracking data failed: %w", err)
	}
	resp, ok := r.(*fbchrony.ReplyTracking)
	if !ok {
		return fmt.Errorf("got unexpected response type %T while waiting for tracking data", resp)
	}

	// according to https://github.com/mlichvar/chrony/blob/e11b518a1ffa704986fb1f1835c425844ba248ef/ntp.h#L70
	var leapStatus string
	switch resp.LeapStatus {
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
		"reference_id": fbchrony.RefidAsHEX(resp.RefID),
		"stratum":      strconv.FormatUint(uint64(resp.Stratum), 10),
	}
	if c.source != "" {
		tags["source"] = c.source
	}

	fields := map[string]interface{}{
		"frequency":       resp.FreqPPM,
		"system_time":     resp.CurrentCorrection,
		"last_offset":     resp.LastOffset,
		"residual_freq":   resp.ResidFreqPPM,
		"rms_offset":      resp.RMSOffset,
		"root_delay":      resp.RootDelay,
		"root_dispersion": resp.RootDispersion,
		"skew":            resp.SkewPPM,
		"update_interval": resp.LastUpdateInterval,
	}
	acc.AddFields("chrony", fields, tags)

	return nil
}

func init() {
	inputs.Add("chrony", func() telegraf.Input {
		return &Chrony{Timeout: config.Duration(3 * time.Second)}
	})
}
