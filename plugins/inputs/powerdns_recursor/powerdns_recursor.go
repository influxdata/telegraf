//go:generate ../../../tools/readme_config_includer/generator
package powerdns_recursor

import (
	_ "embed"
	"fmt"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"strconv"
	"time"
)

// DO NOT REMOVE THE NEXT TWO LINES! This is required to embed the sampleConfig data.
//go:embed sample.conf
var sampleConfig string

type PowerdnsRecursor struct {
	UnixSockets            []string `toml:"unix_sockets"`
	SocketDir              string   `toml:"socket_dir"`
	SocketMode             string   `toml:"socket_mode"`
	ControlProtocolVersion int      `toml:"control_protocol_version"`

	Log telegraf.Logger `toml:"-"`

	mode uint32
}

var defaultTimeout = 5 * time.Second

func (*PowerdnsRecursor) SampleConfig() string {
	return sampleConfig
}

func (p *PowerdnsRecursor) Init() error {
	if p.SocketMode != "" {
		mode, err := strconv.ParseUint(p.SocketMode, 8, 32)
		if err != nil {
			return fmt.Errorf("could not parse socket_mode: %v", err)
		}

		p.mode = uint32(mode)
	}

	if p.ControlProtocolVersion == 0 {
		p.ControlProtocolVersion = 1
	}

	if p.ControlProtocolVersion < 1 || p.ControlProtocolVersion > 3 {
		return fmt.Errorf("unknown control protocol version '%v', allowed values are 1, 2, 3", p.ControlProtocolVersion)
	}

	if len(p.UnixSockets) == 0 {
		p.UnixSockets = []string{"/var/run/pdns_recursor.controlsocket"}
	}

	return nil
}

func (p *PowerdnsRecursor) Gather(acc telegraf.Accumulator) error {
	for _, serverSocket := range p.UnixSockets {
		if err := p.gatherFromServer(serverSocket, acc); err != nil {
			acc.AddError(err)
		}
	}

	return nil
}

func (p *PowerdnsRecursor) gatherFromServer(address string, acc telegraf.Accumulator) error {
	if p.ControlProtocolVersion == 1 {
		return p.gatherFromV1Server(address, acc)
	}

	if p.ControlProtocolVersion == 2 {
		return p.gatherFromV2Server(address, acc)
	}

	if p.ControlProtocolVersion == 3 {
		return p.gatherFromV3Server(address, acc)
	}

	return fmt.Errorf("unknown powerdns recursor protocol version '%v'", p.ControlProtocolVersion)
}

func init() {
	inputs.Add("powerdns_recursor", func() telegraf.Input {
		return &PowerdnsRecursor{
			mode: uint32(0666),
		}
	})
}
