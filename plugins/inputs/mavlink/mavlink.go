//go:generate ../../../tools/readme_config_includer/generator
package mavlink

import (
	_ "embed"
	"log"
	"reflect"
	"regexp"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/bluenviron/gomavlib/v3"
	"github.com/bluenviron/gomavlib/v3/pkg/dialects/ardupilotmega"
)

// Convert from CamelCase to snake_case
func ConvertToSnakeCase(input string) string {
	re := regexp.MustCompile(`([a-z0-9])([A-Z])`)
	snake := re.ReplaceAllString(input, `${1}_${2}`)
	snake = strings.ToLower(snake)
	return snake
}

//go:embed sample.conf
var sampleConfig string

// Plugin data struct
type Mavlink struct {
	Endpoint string          `toml:"endpoint"`
	Log telegraf.Logger

	// Internal state
	connection *gomavlib.Node
	acc telegraf.Accumulator
}

// S

func (*Mavlink) SampleConfig() string {
	return sampleConfig
}

func (s *Mavlink) Start(acc telegraf.Accumulator) error {
	s.acc = acc
	log.Printf("Starting Mavlink plugin")

	// Start goroutine to connect to Mavlink and stream out data
	go func() {
		// Start MAVLink endpoint
		connection, err := gomavlib.NewNode(gomavlib.NodeConf{
			Endpoints: []gomavlib.EndpointConf{
				// gomavlib.EndpointSerial{
				// 	Device: "/dev/ttyACM0",
				// 	Baud:   57600,
				// },
				// gomavlib.EndpointTCPServer{":5760"},
				gomavlib.EndpointTCPClient{s.Endpoint},
			},
			Dialect:     ardupilotmega.Dialect,
			OutVersion:  gomavlib.V2,
			OutSystemID: 2,
			StreamRequestEnable: true,
		})
		if err != nil {
			return
		}
		s.connection = connection
		defer s.connection.Close()
	
		log.Printf("Connected to MAVLink!")

		// Process MAVLink messages
		// Use reflection to retrieve and handle all message types.
		for evt := range s.connection.Events() {
			if frm, ok := evt.(*gomavlib.EventFrame); ok {
				tags := map[string]string{}
				var fields = make(map[string]interface{})

				m := frm.Message()
				t := reflect.TypeOf(m)
				v := reflect.ValueOf(m)
				if t.Kind() == reflect.Ptr {
					t = t.Elem()
					v = v.Elem()
				}

				for i := 0; i < t.NumField(); i++ {
					field := t.Field(i)
					value := v.Field(i)
					fields[ConvertToSnakeCase(field.Name)] = value.Interface()
				}

				msg_name := ConvertToSnakeCase(t.Name())

				if (strings.HasPrefix(msg_name, "message_")) {
					msg_name = strings.TrimPrefix(msg_name, "message_")
					s.acc.AddFields(msg_name, fields, tags)
				}
			}
		}
		return
	}()

	return nil
}

func (s *Mavlink) Gather(_ telegraf.Accumulator) error {
	return nil
}

func (s *Mavlink) Stop() {
	log.Printf("Stopping Mavlink plugin")
}

func init() {
	inputs.Add("mavlink", func() telegraf.Input {
		return &Mavlink{
			Endpoint: "serial:/dev/ttyACM0",
		}
	})
}
