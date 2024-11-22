//go:generate ../../../tools/readme_config_includer/generator
package mavlink

import (
	_ "embed"
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/bluenviron/gomavlib/v3"
	"github.com/bluenviron/gomavlib/v3/pkg/dialects/ardupilotmega"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// Convert from CamelCase to snake_case
func ConvertToSnakeCase(input string) string {
	re := regexp.MustCompile(`([a-z0-9])([A-Z])`)
	snake := re.ReplaceAllString(input, `${1}_${2}`)
	snake = strings.ToLower(snake)
	return snake
}

// Function to check if a string is in a slice
func Contains(slice []string, str string) bool {
	for _, item := range slice {
		if item == str {
			return true
		}
	}
	return false
}

//go:embed sample.conf
var sampleConfig string

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

func (*Mavlink) SampleConfig() string {
	return sampleConfig
}

func (s *Mavlink) Start(acc telegraf.Accumulator) error {
	s.acc = acc

	// Start goroutine to connect to Mavlink and stream out data async
	go func() {
		endpointConfig := []gomavlib.EndpointConf{}
		if strings.HasPrefix(s.FcuUrl, "serial://") {
			tmpStr := strings.TrimPrefix(s.FcuUrl, "serial://")
			tmpStrParts := strings.Split(tmpStr, ":")
			deviceName := tmpStrParts[0]
			baudRate := 57600
			if len(tmpStrParts) == 2 {
				newBaudRate, err := strconv.Atoi(tmpStrParts[1])
				if err != nil {
					log.Printf("Mavlink setup error: serial baud rate not valid!")
					return
				}
				baudRate = newBaudRate
			}

			log.Printf("Mavlink serial client: device %s, baud rate %d", deviceName, baudRate)
			endpointConfig = []gomavlib.EndpointConf{
				gomavlib.EndpointSerial{
					Device: deviceName,
					Baud:   baudRate,
				},
			}
		} else if strings.HasPrefix(s.FcuUrl, "tcp://") {
			// TCP client
			tmpStr := strings.TrimPrefix(s.FcuUrl, "tcp://")
			tmpStrParts := strings.Split(tmpStr, ":")
			if len(tmpStrParts) != 2 {
				log.Printf("Mavlink setup error: TCP requires a port!")
				return
			}

			hostname := tmpStrParts[0]
			port := 14550
			port, err := strconv.Atoi(tmpStrParts[1])
			if err != nil {
				log.Printf("Mavlink setup error: TCP port is invalid!")
				return
			}

			if len(hostname) > 0 {
				log.Printf("Mavlink TCP client: hostname %s, port %d", hostname, port)
				endpointConfig = []gomavlib.EndpointConf{
					gomavlib.EndpointTCPClient{fmt.Sprintf("%s:%d", hostname, port)},
				}
			} else {
				log.Printf("Mavlink TCP server: port %d", port)
				endpointConfig = []gomavlib.EndpointConf{
					gomavlib.EndpointTCPServer{fmt.Sprintf(":%d", port)},
				}
			}
		} else if strings.HasPrefix(s.FcuUrl, "udp://") {
			// UDP client or server
			tmpStr := strings.TrimPrefix(s.FcuUrl, "udp://")
			tmpStrParts := strings.Split(tmpStr, ":")
			if len(tmpStrParts) != 2 {
				log.Printf("Mavlink setup error: UDP requires a port!")
				return
			}

			hostname := tmpStrParts[0]
			port := 14550
			port, err := strconv.Atoi(tmpStrParts[1])
			if err != nil {
				log.Printf("Mavlink setup error: UDP port is invalid!")
				return
			}

			if len(hostname) > 0 {
				log.Printf("Mavlink UDP client: hostname %s, port %d", hostname, port)
				endpointConfig = []gomavlib.EndpointConf{
					gomavlib.EndpointUDPClient{fmt.Sprintf("%s:%d", hostname, port)},
				}
			} else {
				log.Printf("Mavlink UDP server: port %d", port)
				endpointConfig = []gomavlib.EndpointConf{
					gomavlib.EndpointUDPServer{fmt.Sprintf(":%d", port)},
				}
			}
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
		for evt := range s.connection.Events() {
			switch evt.(type) {
			case *gomavlib.EventFrame:
				if frm, ok := evt.(*gomavlib.EventFrame); ok {
					tags := map[string]string{}
					var fields = make(map[string]interface{})
					tags["sys_id"] = fmt.Sprintf("%d", frm.SystemID())
					tags["fcu_url"] = s.FcuUrl

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
					if strings.HasPrefix(msg_name, "message_") {
						msg_name = strings.TrimPrefix(msg_name, "message_")

						if len(s.MessageFilter) > 0 && Contains(s.MessageFilter, msg_name) {
							log.Printf("%s did not match filter\n", msg_name)
							continue
						}
						s.acc.AddFields(msg_name, fields, tags)
					}
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
	// Nothing to do when gathering metrics; fields are accumulated async.
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
