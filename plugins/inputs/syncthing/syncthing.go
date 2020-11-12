package syncthing

import (
	"context"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/pkg/errors"
)

const (
	SysconfigEndpoint    = "/rest/system/config"
	ConnectionsEndpoint  = "/rest/system/connections"
	SystemStatusEndpoint = "/rest/system/status"
	NeedEndpoint         = "/rest/db/need"
)

type Syncthing struct {
	URL       string `toml:"url"`
	TokenFile string `toml:"token_file"`
	Token     string `toml:"token"`
	tls.ClientConfig

	Timeout internal.Duration `toml:"timeout"`

	client *http.Client

	// The parser will automatically be set by Telegraf core code because
	// this plugin implements the ParserInput interface (i.e. the SetParser method)
	parser parsers.Parser
}

const (
	AuthHeader = "X-API-KEY"

	sampleConfig = `
	  ## Syncthing host
	  url = "http://localhost:8384"

	  # token_file = "/path/to/file"
	  ## OR
	  # token = "<api-access-token>"

	  ## Optional TLS Config
	  # tls_ca = "/etc/telegraf/ca.pem"
	  # tls_cert = "/etc/telegraf/cert.pem"
	  # tls_key = "/etc/telegraf/key.pem"
	  ## Use TLS but skip chain & host verification
	  # insecure_skip_verify = false

	  ## Amount of time allowed to complete the HTTP request
	  # timeout = "5s"
	`
)

// SampleConfig returns the default configuration of the Input
func (*Syncthing) SampleConfig() string {
	return sampleConfig
}

// Description returns a one-sentence description on the Input
func (*Syncthing) Description() string {
	return "Read syncthing metrics from one or more endpoints"
}

func (s *Syncthing) Init() error {
	tlsCfg, err := s.ClientConfig.TLSConfig()
	if err != nil {
		return err
	}

	if s.TokenFile != "" {
		b, err := ioutil.ReadFile(s.TokenFile)
		if err != nil {
			return errors.Wrap(err, "failed to read token file")
		}
		s.Token = string(b)
	}

	s.Token = strings.TrimSpace(s.Token)
	if s.Token == "" {
		return errors.New("required token was not provided")
	}

	s.client = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig:     tlsCfg,
			Proxy:               http.ProxyFromEnvironment,
			TLSHandshakeTimeout: s.Timeout.Duration,
			Dial: (&net.Dialer{
				Timeout: s.Timeout.Duration,
			}).Dial,
		},
		Timeout: s.Timeout.Duration,
	}

	return nil
}

// Gather takes in an accumulator and adds the metrics that the Input
// gathers. This is called every "interval"
func (s *Syncthing) Gather(acc telegraf.Accumulator) error {
	ctx, cancel := context.WithTimeout(context.Background(), s.Timeout.Duration)
	defer cancel()
	sysstatus, err := s.SystemStatus(ctx, s.URL)
	if err != nil {
		return errors.Wrap(err, "failed to gather system status")
	}

	sysconfig, err := s.SystemConfig(ctx, s.URL)
	if err != nil {
		return errors.Wrap(err, "failed to gather system config")
	}

	conns, err := s.SystemConnections(ctx, s.URL)
	if err != nil {
		return errors.Wrap(err, "failed to gather system connections")
	}

	myID := sysstatus.MyID
	now := time.Now()

	devices := sysconfig.Devices[:0]
	for _, dev := range sysconfig.Devices {
		if dev.ID != myID {
			devices = append(devices, dev)
		}
	}
	sysconfig.Devices = devices

	addConnections(acc, sysconfig, conns, now)

	s.addFolders(ctx, s.URL, acc, sysconfig, now)

	return nil
}

// SetParser takes the data_format from the config and finds the right parser for that format
func (s *Syncthing) SetParser(parser parsers.Parser) {
	s.parser = parser
}

func (s *Syncthing) addFolders(ctx context.Context, host string, acc telegraf.Accumulator, sysconfig *SystemConfig, t time.Time) {
	for _, folder := range sysconfig.Folders {
		n, err := s.Need(ctx, host, folder.ID)
		if err != nil {
			acc.AddError(errors.Wrapf(err, "failed to lookup folder needs for %q:%q", folder.ID, folder.Label))
			continue
		}
		fields := map[string]interface{}{
			"paused": folder.Paused,
			"need":   n.Total,
		}
		tags := map[string]string{
			"label": folder.Label,
			"id":    folder.ID,
			"path":  folder.Path,
		}
		acc.AddFields("syncthing_folder", fields, tags, t)
	}
}

func addConnections(acc telegraf.Accumulator, sysconfig *SystemConfig, conns *SystemConnections, t time.Time) {
	for id, con := range conns.Connections {
		dev := sysconfig.DeviceByID(id)
		if dev == nil {
			// self
			continue
		}

		tags := map[string]string{
			"device_id": id,
			"name":      dev.Name,
		}
		fields := map[string]interface{}{
			"client_version":  con.ClientVersion,
			"address":         con.Address,
			"connected":       con.Connected,
			"crypto":          con.Crypto,
			"in_bytes_total":  con.InBytesTotal,
			"out_bytes_total": con.OutBytesTotal,
			"paused":          con.Paused,
		}
		acc.AddFields("syncthing_connection", fields, tags, t)
	}
}

func init() {
	inputs.Add("syncthing", func() telegraf.Input {
		return &Syncthing{
			URL:     "http://localhost:8384",
			Timeout: internal.Duration{Duration: time.Second * 5},
		}
	})
}
