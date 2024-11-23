//go:generate ../../../tools/readme_config_includer/generator
package mavlink

import (
	_ "embed"
	"log"
	"time"

	"github.com/bluenviron/gomavlib/v3"
	"github.com/bluenviron/gomavlib/v3/pkg/dialects/ardupilotmega"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// Plugin state
type Mavlink struct {
	// Config param
	FcuUrl                 string   `toml:"fcu_url"`
	SystemId               uint8    `toml:"system_id"`
	MessageFilter          []string `toml:"message_filter"`
	StreamRequestEnable    bool     `toml:"stream_request_enable"`
	StreamRequestFrequency int      `toml:"stream_request_frequency"`

	// Internal state
	connection *gomavlib.Node
	acc        telegraf.Accumulator
	loading    bool
	terminated bool
}

//go:embed sample.conf
var sampleConfig string

func (*Mavlink) SampleConfig() string {
	return sampleConfig
}

// Container for a parsed Mavlink frame
type MetricFrameData struct {
	name   string
	tags   map[string]string
	fields map[string]any
}

func (s *Mavlink) Start(acc telegraf.Accumulator) error {
	s.acc = acc

	// Start routine to connect to Mavlink and stream out data async
	go func() {
		endpointConfig, err := ParseMavlinkEndpointConfig(s)
		if err != nil {
			log.Printf("%s", err.Error())
			return
		}

		// Start MAVLink endpoint
		s.loading = true
		s.terminated = false
		for s.loading == true {
			connection, err := gomavlib.NewNode(gomavlib.NodeConf{
				Endpoints:              endpointConfig,
				Dialect:                ardupilotmega.Dialect,
				OutVersion:             gomavlib.V2,
				OutSystemID:            s.SystemId,
				StreamRequestEnable:    s.StreamRequestEnable,
				StreamRequestFrequency: s.StreamRequestFrequency,
			})
			if err != nil {
				log.Printf("Mavlink failed to connect (%s), will try again in 5s...", err.Error())
				time.Sleep(5 * time.Second)
				continue
			}
			s.loading = false
			s.connection = connection
		}
		defer s.connection.Close()
		if s.terminated {
			return
		}

		// Process MAVLink messages
		// Use reflection to retrieve and handle all message types.
		// (There are several hundred Mavlink message types)
		for evt := range s.connection.Events() {
			switch evt.(type) {
			case *gomavlib.EventFrame:
				if frm, ok := evt.(*gomavlib.EventFrame); ok {
					result := MavlinkEventFrameToMetric(frm)
					if len(s.MessageFilter) > 0 && Contains(s.MessageFilter, result.name) {
						continue
					}
					result.tags["fcu_url"] = s.FcuUrl
					s.acc.AddFields(result.name, result.fields, result.tags)
				}

			case *gomavlib.EventChannelOpen:
				log.Printf("Mavlink channel opened")

			case *gomavlib.EventChannelClose:
				log.Printf("Mavlink channel closed")
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
	s.loading = false
}

func init() {
	inputs.Add("mavlink", func() telegraf.Input {
		return &Mavlink{
			FcuUrl:                 "udp://:14540",
			MessageFilter:          []string{},
			SystemId:               254,
			StreamRequestEnable:    true,
			StreamRequestFrequency: 4,
		}
	})
}
