//go:generate ../../../tools/readme_config_includer/generator
package chrony

import (
	"bytes"
	_ "embed"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"os/user"
	"path"
	"strconv"
	"sync"
	"syscall"
	"time"

	fbchrony "github.com/facebook/time/ntp/chrony"
	"github.com/google/uuid"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type Chrony struct {
	Server      string          `toml:"server"`
	Timeout     config.Duration `toml:"timeout"`
	DNSLookup   bool            `toml:"dns_lookup"`
	SocketGroup string          `toml:"socket_group"`
	SocketPerms string          `toml:"socket_perms"`
	Metrics     []string        `toml:"metrics"`
	Log         telegraf.Logger `toml:"-"`

	conn   net.Conn
	client *fbchrony.Client
	source string
	local  string

	// clientMu protects concurrent access to the chrony client.
	// This prevents race conditions when multiple Gather() calls
	// access the same client instance concurrently.
	sync.Mutex
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
		case "unixgram":
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
		case "activity", "tracking", "serverstats", "sources", "sourcestats":
			// Do nothing as those are valid
		default:
			return fmt.Errorf("invalid metric setting %q", m)
		}
	}

	if c.SocketGroup == "" {
		c.SocketGroup = "chrony"
	}

	if c.SocketPerms == "" {
		c.SocketPerms = "0660"
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
		case "unixgram":
			conn, err := c.dialUnix(u.Path)
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
		if conn, err := c.dialUnix("/run/chrony/chronyd.sock"); err == nil {
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

func (c *Chrony) Gather(acc telegraf.Accumulator) error {
	// Protect concurrent access to the chrony client to prevent race conditions
	// when multiple Gather() calls occur simultaneously (e.g., when a previous
	// gather hasn't completed before the next interval triggers).
	c.Lock()
	defer c.Unlock()

	for _, m := range c.Metrics {
		switch m {
		case "activity":
			acc.AddError(c.gatherActivity(acc))
		case "tracking":
			acc.AddError(c.gatherTracking(acc))
		case "serverstats":
			acc.AddError(c.gatherServerStats(acc))
		case "sources":
			acc.AddError(c.gatherSources(acc))
		case "sourcestats":
			acc.AddError(c.gatherSourceStats(acc))
		default:
			return fmt.Errorf("invalid metric setting %q", m)
		}
	}

	return nil
}

func (c *Chrony) Stop() {
	if c.conn != nil {
		if err := c.conn.Close(); err != nil && !errors.Is(err, net.ErrClosed) && !errors.Is(err, syscall.EPIPE) {
			c.Log.Errorf("Closing connection to %q failed: %v", c.Server, err)
		}
	}
	if c.local != "" {
		if err := os.Remove(c.local); err != nil {
			c.Log.Errorf("Removing temporary socket %q failed: %v", c.local, err)
		}
	}
}

// dialUnix opens an unixgram connection with chrony
func (c *Chrony) dialUnix(address string) (*net.UnixConn, error) {
	dir := path.Dir(address)
	c.local = path.Join(dir, fmt.Sprintf("chrony-telegraf-%s.sock", uuid.New().String()))
	conn, err := net.DialUnix("unixgram",
		&net.UnixAddr{Name: c.local, Net: "unixgram"},
		&net.UnixAddr{Name: address, Net: "unixgram"},
	)

	if err != nil {
		return nil, err
	}

	filemode, err := strconv.ParseUint(c.SocketPerms, 8, 32)
	if err != nil {
		return nil, fmt.Errorf("parsing file mode %q failed: %w", c.SocketPerms, err)
	}

	if err := os.Chmod(c.local, os.FileMode(filemode)); err != nil {
		return nil, fmt.Errorf("changing file mode of %q failed: %w", c.local, err)
	}

	group, err := user.LookupGroup(c.SocketGroup)
	if err != nil {
		return nil, fmt.Errorf("looking up group %q failed: %w", c.SocketGroup, err)
	}
	gid, err := strconv.Atoi(group.Gid)
	if err != nil {
		return nil, fmt.Errorf("parsing group ID %q failed: %w", group.Gid, err)
	}
	if err := os.Chown(c.local, os.Getuid(), gid); err != nil {
		return nil, fmt.Errorf("changing group of %q failed: %w", c.local, err)
	}

	return conn, nil
}

func (c *Chrony) gatherActivity(acc telegraf.Accumulator) error {
	req := fbchrony.NewActivityPacket()
	r, err := c.client.Communicate(req)
	if err != nil {
		return fmt.Errorf("querying activity data failed: %w", err)
	}
	resp, ok := r.(*fbchrony.ReplyActivity)
	if !ok {
		return fmt.Errorf("got unexpected response type %T while waiting for activity data", r)
	}

	tags := make(map[string]string, 1)
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
		return fmt.Errorf("got unexpected response type %T while waiting for tracking data", r)
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

func (c *Chrony) gatherServerStats(acc telegraf.Accumulator) error {
	req := fbchrony.NewServerStatsPacket()
	r, err := c.client.Communicate(req)
	if err != nil {
		return fmt.Errorf("querying server statistics failed: %w", err)
	}

	tags := make(map[string]string, 1)
	if c.source != "" {
		tags["source"] = c.source
	}

	var fields map[string]interface{}
	switch resp := r.(type) {
	case *fbchrony.ReplyServerStats:
		fields = map[string]interface{}{
			"ntp_hits":  resp.NTPHits,
			"ntp_drops": resp.NTPDrops,
			"cmd_hits":  resp.CMDHits,
			"cmd_drops": resp.CMDDrops,
			"log_drops": resp.LogDrops,
		}
	case *fbchrony.ReplyServerStats2:
		fields = map[string]interface{}{
			"ntp_hits":      resp.NTPHits,
			"ntp_drops":     resp.NTPDrops,
			"ntp_auth_hits": resp.NTPAuthHits,
			"cmd_hits":      resp.CMDHits,
			"cmd_drops":     resp.CMDDrops,
			"log_drops":     resp.LogDrops,
			"nke_hits":      resp.NKEHits,
			"nke_drops":     resp.NKEDrops,
		}
	case *fbchrony.ReplyServerStats3:
		fields = map[string]interface{}{
			"ntp_hits":             resp.NTPHits,
			"ntp_drops":            resp.NTPDrops,
			"ntp_auth_hits":        resp.NTPAuthHits,
			"ntp_interleaved_hits": resp.NTPInterleavedHits,
			"ntp_timestamps":       resp.NTPTimestamps,
			"ntp_span_seconds":     resp.NTPSpanSeconds,
			"cmd_hits":             resp.CMDHits,
			"cmd_drops":            resp.CMDDrops,
			"log_drops":            resp.LogDrops,
			"nke_hits":             resp.NKEHits,
			"nke_drops":            resp.NKEDrops,
		}
	case *fbchrony.ReplyServerStats4:
		fields = map[string]interface{}{
			"ntp_hits":                   resp.NTPHits,
			"ntp_drops":                  resp.NTPDrops,
			"ntp_auth_hits":              resp.NTPAuthHits,
			"ntp_interleaved_hits":       resp.NTPInterleavedHits,
			"ntp_timestamps":             resp.NTPTimestamps,
			"ntp_span_seconds":           resp.NTPSpanSeconds,
			"cmd_hits":                   resp.CMDHits,
			"cmd_drops":                  resp.CMDDrops,
			"log_drops":                  resp.LogDrops,
			"nke_hits":                   resp.NKEHits,
			"nke_drops":                  resp.NKEDrops,
			"ntp_daemon_rx_timestamps":   resp.NTPDaemonRxtimestamps,
			"ntp_daemon_tx_timestamps":   resp.NTPDaemonTxtimestamps,
			"ntp_kernel_rx_timestamps":   resp.NTPKernelRxtimestamps,
			"ntp_kernel_tx_timestamps":   resp.NTPKernelTxtimestamps,
			"ntp_hardware_rx_timestamps": resp.NTPHwRxTimestamps,
			"ntp_hardware_tx_timestamps": resp.NTPHwTxTimestamps,
		}
	default:
		return fmt.Errorf("got unexpected response type %T while waiting for server statistics", r)
	}

	acc.AddFields("chrony_serverstats", fields, tags)

	return nil
}

func (c *Chrony) getSourceName(ip net.IP) (string, error) {
	sourceNameReq := fbchrony.NewNTPSourceNamePacket(ip)
	sourceNameRaw, err := c.client.Communicate(sourceNameReq)
	if err != nil {
		return "", fmt.Errorf("querying name of source %q failed: %w", ip, err)
	}
	sourceName, ok := sourceNameRaw.(*fbchrony.ReplyNTPSourceName)
	if !ok {
		return "", fmt.Errorf("got unexpected response type %T while waiting for source name", sourceNameRaw)
	}

	// Cut the string at null termination
	if termidx := bytes.Index([]byte(sourceName.Name), []byte{0}); termidx >= 0 {
		return sourceName.Name[:termidx], nil
	}
	return sourceName.Name, nil
}

func (c *Chrony) gatherSources(acc telegraf.Accumulator) error {
	sourcesReq := fbchrony.NewSourcesPacket()
	sourcesRaw, err := c.client.Communicate(sourcesReq)
	if err != nil {
		return fmt.Errorf("querying sources failed: %w", err)
	}

	sourcesResp, ok := sourcesRaw.(*fbchrony.ReplySources)
	if !ok {
		return fmt.Errorf("got unexpected response type %T while waiting for sources", sourcesRaw)
	}

	for idx := int32(0); int(idx) < sourcesResp.NSources; idx++ {
		// Getting the source data
		sourceDataReq := fbchrony.NewSourceDataPacket(idx)
		sourceDataRaw, err := c.client.Communicate(sourceDataReq)
		if err != nil {
			return fmt.Errorf("querying data for source %d failed: %w", idx, err)
		}
		sourceData, ok := sourceDataRaw.(*fbchrony.ReplySourceData)
		if !ok {
			return fmt.Errorf("got unexpected response type %T while waiting for source data", sourceDataRaw)
		}

		// Trying to resolve the source name
		var peer string

		if sourceData.Mode == fbchrony.SourceModeRef && sourceData.IPAddr.To4() != nil {
			// References of local sources (PPS, etc..) are encoded in the bits of the IPv4 address,
			// instead of the RefId. Extract the correct name for those sources.
			ipU32 := binary.BigEndian.Uint32(sourceData.IPAddr.To4())

			peer = fbchrony.RefidToString(ipU32)
		} else {
			peer, err = c.getSourceName(sourceData.IPAddr)
			if err != nil {
				return err
			}
		}

		if peer == "" {
			peer = sourceData.IPAddr.String()
		}

		tags := map[string]string{
			"peer": peer,
		}
		if c.source != "" {
			tags["source"] = c.source
		}

		fields := map[string]interface{}{
			"index":                    idx,
			"ip":                       sourceData.IPAddr.String(),
			"poll":                     sourceData.Poll,
			"stratum":                  sourceData.Stratum,
			"state":                    sourceData.State.String(),
			"mode":                     sourceData.Mode.String(),
			"flags":                    sourceData.Flags,
			"reachability":             sourceData.Reachability,
			"sample":                   sourceData.SinceSample,
			"latest_measurement":       sourceData.LatestMeas,
			"latest_measurement_error": sourceData.LatestMeasErr,
		}
		acc.AddFields("chrony_sources", fields, tags)
	}
	return nil
}

func (c *Chrony) gatherSourceStats(acc telegraf.Accumulator) error {
	sourcesReq := fbchrony.NewSourcesPacket()
	sourcesRaw, err := c.client.Communicate(sourcesReq)
	if err != nil {
		return fmt.Errorf("querying sources failed: %w", err)
	}

	sourcesResp, ok := sourcesRaw.(*fbchrony.ReplySources)
	if !ok {
		return fmt.Errorf("got unexpected response type %T while waiting for sources", sourcesRaw)
	}

	for idx := int32(0); int(idx) < sourcesResp.NSources; idx++ {
		// Getting the source data
		sourceStatsReq := fbchrony.NewSourceStatsPacket(idx)
		sourceStatsRaw, err := c.client.Communicate(sourceStatsReq)
		if err != nil {
			return fmt.Errorf("querying data for source %d failed: %w", idx, err)
		}
		sourceStats, ok := sourceStatsRaw.(*fbchrony.ReplySourceStats)
		if !ok {
			return fmt.Errorf("got unexpected response type %T while waiting for source data", sourceStatsRaw)
		}

		// Trying to resolve the source name
		var peer string

		if sourceStats.IPAddr.String() == "::" {
			// `::` IPs mean a local source directly connected to the server.
			// For example a PPS source or a local GPS receiver.
			// Then the name of the source is encoded in its RefID
			peer = fbchrony.RefidToString(sourceStats.RefID)
		} else {
			peer, err = c.getSourceName(sourceStats.IPAddr)
			if err != nil {
				return err
			}
		}

		if peer == "" {
			peer = sourceStats.IPAddr.String()
		}

		tags := map[string]string{
			"reference_id": fbchrony.RefidAsHEX(sourceStats.RefID),
			"peer":         peer,
		}
		if c.source != "" {
			tags["source"] = c.source
		}

		fields := map[string]interface{}{
			"index":              idx,
			"ip":                 sourceStats.IPAddr.String(),
			"samples":            sourceStats.NSamples,
			"runs":               sourceStats.NRuns,
			"span_seconds":       sourceStats.SpanSeconds,
			"stddev":             sourceStats.StandardDeviation,
			"residual_frequency": sourceStats.ResidFreqPPM,
			"skew":               sourceStats.SkewPPM,
			"offset":             sourceStats.EstimatedOffset,
			"offset_error":       sourceStats.EstimatedOffsetErr,
		}
		acc.AddFields("chrony_sourcestats", fields, tags)
	}
	return nil
}

func init() {
	inputs.Add("chrony", func() telegraf.Input {
		return &Chrony{Timeout: config.Duration(3 * time.Second)}
	})
}
