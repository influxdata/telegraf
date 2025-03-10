//go:generate ../../../tools/readme_config_includer/generator
package huebridge

import (
	_ "embed"
	"errors"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type RemoteClientConfig struct {
	RemoteClientID     string `toml:"remote_client_id"`
	RemoteClientSecret string `toml:"remote_client_secret"`
	RemoteCallbackURL  string `toml:"remote_callback_url"`
	RemoteTokenDir     string `toml:"remote_token_dir"`
}

type HueBridge struct {
	BridgeUrls      []string          `toml:"bridges"`
	RoomAssignments map[string]string `toml:"room_assignments"`
	Timeout         config.Duration   `toml:"timeout"`
	Log             telegraf.Logger   `toml:"-"`
	RemoteClientConfig
	tls.ClientConfig

	bridges []*bridge
}

func (*HueBridge) SampleConfig() string {
	return sampleConfig
}

func (h *HueBridge) Init() error {
	tlsCfg, err := h.ClientConfig.TLSConfig()
	if err != nil {
		return fmt.Errorf("creating TLS configuration failed: %w", err)
	}

	h.bridges = make([]*bridge, 0, len(h.BridgeUrls))
	for _, b := range h.BridgeUrls {
		u, err := url.Parse(b)
		if err != nil {
			return fmt.Errorf("failed to parse bridge URL %s: %w", b, err)
		}

		switch u.Scheme {
		case "address", "cloud", "mdns":
			// Do nothing, those are valid
		case "remote":
			// Remote scheme also requires a configured rcc
			if h.RemoteClientID == "" || h.RemoteClientSecret == "" || h.RemoteTokenDir == "" {
				return errors.New("missing remote application credentials and/or token director not configured")
			}
		default:
			return fmt.Errorf("unrecognized scheme %s in URL %s", u.Scheme, b)
		}

		// All schemes require a password in the URL
		if _, set := u.User.Password(); !set {
			return fmt.Errorf("missing password in URL %s", u)
		}

		h.bridges = append(h.bridges, &bridge{
			url:                   u,
			configRoomAssignments: h.RoomAssignments,
			remoteCfg:             &h.RemoteClientConfig,
			tlsCfg:                tlsCfg,
			timeout:               time.Duration(h.Timeout),
			log:                   h.Log,
		})
	}

	return nil
}

func (h *HueBridge) Gather(acc telegraf.Accumulator) error {
	var wg sync.WaitGroup
	for _, bridge := range h.bridges {
		wg.Add(1)
		go func() {
			defer wg.Done()
			acc.AddError(bridge.process(acc))
		}()
	}
	wg.Wait()
	return nil
}

func init() {
	inputs.Add("huebridge", func() telegraf.Input {
		return &HueBridge{Timeout: config.Duration(10 * time.Second)}
	})
}
