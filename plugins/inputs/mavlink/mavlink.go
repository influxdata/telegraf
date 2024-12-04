//go:generate ../../../tools/readme_config_includer/generator
package mavlink

import (
	_ "embed"
	"fmt"

	"github.com/chrisdalke/gomavlib/v3"
	"github.com/chrisdalke/gomavlib/v3/pkg/dialects/ardupilotmega"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/internal/choice"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// Plugin state
type Mavlink struct {
	// Config param
	FcuURL                 string   `toml:"fcu_url"`
	SystemID               uint8    `toml:"system_id"`
	MessageFilter          []string `toml:"message_filter"`
	StreamRequestEnable    bool     `toml:"stream_request_enable"`
	StreamRequestFrequency int      `toml:"stream_request_frequency"`

	Log telegraf.Logger `toml:"-"`

	// Internal state
	connection     *gomavlib.Node
	endpointConfig []gomavlib.EndpointConf
	terminated     bool
}

//go:embed sample.conf
var sampleConfig string

func (*Mavlink) SampleConfig() string {
	return sampleConfig
}

func (s *Mavlink) Init() error {
	// Parse out the Mavlink endpoint.
	endpointConfig, err := ParseMavlinkEndpointConfig(s.FcuURL)
	if err != nil {
		return err
	}
	s.endpointConfig = endpointConfig

	return nil
}

func (s *Mavlink) Start(acc telegraf.Accumulator) error {
	// Start MAVLink endpoint
	connection, err := gomavlib.NewNode(gomavlib.NodeConf{
		Endpoints:              s.endpointConfig,
		Dialect:                ardupilotmega.Dialect,
		OutVersion:             gomavlib.V2,
		OutSystemID:            s.SystemID,
		StreamRequestEnable:    s.StreamRequestEnable,
		StreamRequestFrequency: s.StreamRequestFrequency,
	})
	if err != nil {
		return &internal.StartupError{
			Err:   fmt.Errorf("mavlink client failed: %w", err),
			Retry: true,
		}
	}
	s.terminated = false
	s.connection = connection

	// Start routine to connect to Mavlink and stream out data async
	go func() {
		defer s.connection.Close()
		if s.terminated {
			return
		}

		// Process MAVLink messages
		// Use reflection to retrieve and handle all message types.
		// (There are several hundred Mavlink message types)
		for evt := range s.connection.Events() {
			if s.terminated {
				return
			}
			switch evt := evt.(type) {
			case *gomavlib.EventFrame:
				result := MavlinkEventFrameToMetric(evt)
				if len(s.MessageFilter) > 0 && !choice.Contains(result.Name(), s.MessageFilter) {
					continue
				}
				result.AddTag("source", s.FcuURL)
				acc.AddMetric(result)

			case *gomavlib.EventChannelOpen:
				s.Log.Debugf("Mavlink channel opened")

			case *gomavlib.EventChannelClose:
				s.Log.Debugf("Mavlink channel closed")
			}
		}
	}()

	return nil
}

func (s *Mavlink) Gather(_ telegraf.Accumulator) error {
	return nil
}

func (s *Mavlink) Stop() {
	s.terminated = true
}

func init() {
	inputs.Add("mavlink", func() telegraf.Input {
		return &Mavlink{
			FcuURL:                 "udp://:14540",
			MessageFilter:          make([]string, 0),
			SystemID:               254,
			StreamRequestEnable:    true,
			StreamRequestFrequency: 4,
		}
	})
}
