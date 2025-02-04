//go:generate ../../../tools/readme_config_includer/generator
package huebridge

import (
	_ "embed"
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
	RemoteClientId     string `toml:"remote_client_id"`
	RemoteClientSecret string `toml:"remote_client_secret"`
	RemoteCallbackUrl  string `toml:"remote_callback_url"`
	RemoteTokenDir     string `toml:"remote_token_dir"`
}

type HueBridge struct {
	Bridges         []string          `toml:"bridges"`
	RoomAssignments map[string]string `toml:"room_assignments"`
	Timeout         config.Duration   `toml:"timeout"`
	Log             telegraf.Logger   `toml:"-"`
	RemoteClientConfig
	tls.ClientConfig

	configuredBridges []*bridge
}

func (*HueBridge) SampleConfig() string {
	return sampleConfig
}

func (h *HueBridge) Init() error {
	h.configuredBridges = make([]*bridge, 0, len(h.Bridges))
	for _, bridgeUrl := range h.Bridges {
		bridge, err := newBridge(bridgeUrl, h.RoomAssignments, &h.RemoteClientConfig, &h.ClientConfig, h.Timeout, h.Log)
		if err != nil {
			h.Log.Warnf("Failed to instantiate bridge for URL %s: %s", bridgeUrl, err)
			continue
		}
		h.configuredBridges = append(h.configuredBridges, bridge)
	}
	return nil
}

func (h *HueBridge) Gather(acc telegraf.Accumulator) error {
	var waitComplete sync.WaitGroup
	for _, bridge := range h.configuredBridges {
		waitComplete.Add(1)
		go func() {
			defer waitComplete.Done()
			acc.AddError(bridge.process(acc))
		}()
	}
	waitComplete.Wait()
	return nil
}

func init() {
	inputs.Add("huebridge", func() telegraf.Input {
		return &HueBridge{Timeout: config.Duration(10 * time.Second)}
	})
}
