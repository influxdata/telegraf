//go:generate ../../../tools/readme_config_includer/generator
package mavlink

import (
	_ "embed"
	"log"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/bluenviron/gomavlib/v3"
	"github.com/bluenviron/gomavlib/v3/pkg/dialects/ardupilotmega"
)

//go:embed sample.conf
var sampleConfig string

// Plugin data struct
type Mavlink struct {
	Endpoint string          `toml:"endpoint"`

	connection *gomavlib.Node
}

func (*Mavlink) SampleConfig() string {
	return sampleConfig
}

func (s *Mavlink) Gather(acc telegraf.Accumulator) error {
	if s.connection == nil {
		// Start MAVLink endpoint
		connection, err := gomavlib.NewNode(gomavlib.NodeConf{
			Endpoints: []gomavlib.EndpointConf{
				gomavlib.EndpointSerial{
					Device: "/dev/ttyACM0",
					Baud:   57600,
				},
			},
			Dialect:     ardupilotmega.Dialect,
			OutVersion:  gomavlib.V2,
			OutSystemID: 2,
			StreamRequestEnable: true,
		})
		if err != nil {
			return err
		}
		s.connection = connection
		defer s.connection.Close()
	
		log.Printf("Connected to MAVLink!")
	}

	// Process MAVLink messages
	for evt := range s.connection.Events() {
		if frm, ok := evt.(*gomavlib.EventFrame); ok {
			log.Printf("received: id=%d, %+v\n", frm.Message().GetID(), frm.Message())
		}
	}
	return nil
}

func init() {
	inputs.Add("mavlink", func() telegraf.Input {
		return &Mavlink{
			Endpoint: "serial:/dev/ttyACM0",
		}
	})
}
