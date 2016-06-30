package openconfig_telemetry

import (
	"io"
	"log"
	"strings"
	"sync"
	"time"
	"regexp"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/inputs/openconfig_telemetry/oc"
	"github.com/influxdata/telegraf/plugins/parsers"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type OpenConfigTelemetry struct {
	Server          string
	Sensors         []string
	SampleFrequency uint32

	parser parsers.Parser
	sync.Mutex

	// keep the accumulator internally:
	acc            telegraf.Accumulator
	grpcClientConn *grpc.ClientConn
}

var sampleConfig = `
  server = "localhost:1883"

  ## Frequency to get data in seconds
  sampleFrequency = 1

  ## Sensors to subscribe for
  ## A identifier for each sensor can be provided in path by separating with space
  ## Else sensor path will be used as identifier
  sensors = [
   "/oc/firewall/usage",
   "interfaces /oc/interfaces/",
  ]

  ## Data format to consume.
  ## Each data format has it's own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "influx"
`

func (m *OpenConfigTelemetry) SampleConfig() string {
	return sampleConfig
}

func (m *OpenConfigTelemetry) Description() string {
	return "Read OpenConfig Telemetry from listed sensors"
}

func (m *OpenConfigTelemetry) SetParser(parser parsers.Parser) {
	m.parser = parser
}

func (m *OpenConfigTelemetry) Start(acc telegraf.Accumulator) error {
	log.Print("Started OpenConfig Telemetry plugin\n")
	return nil
}

func (m *OpenConfigTelemetry) Stop() {
	m.Lock()
	defer m.Unlock()
	m.grpcClientConn.Close()
}

func (m *OpenConfigTelemetry) Gather(acc telegraf.Accumulator) error {
	m.Lock()
	defer m.Unlock()

	m.acc = acc

	acc.SetDebug(true)

	// Extract device name / IP
	s := strings.Split(m.Server, ":")
  grpc_server, grpc_port := s[0], s[1]

	var err error
	m.grpcClientConn, err = grpc.Dial(m.Server, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	c := Telemetry.NewOpenConfigTelemetryClient(m.grpcClientConn)
	log.Printf("Opened a new gRPC session to %s on port %s", grpc_server, grpc_port)

	wg := new(sync.WaitGroup)

	for _, sensor := range m.Sensors {
		wg.Add(1)
		go func(sensor string, acc telegraf.Accumulator) {
			defer wg.Done()
			spathSplit := strings.SplitN(sensor, " ", 2)
			var sensorName string
			var sensorPath string
			if len(spathSplit) > 1 {
				sensorName = spathSplit[0]
				sensorPath = spathSplit[1]
			} else {
				sensorName = sensor
				sensorPath = sensor
			}
			stream, err := c.TelemetrySubscribe(context.Background(),
				&Telemetry.SubscriptionRequest{PathList: []*Telemetry.Path{&Telemetry.Path{Path: sensorPath, SampleFrequency: m.SampleFrequency}}})
			if err != nil {
				log.Fatalf("Could not subscribe: %v", err)
			}
			for {
				r, err := stream.Recv()
				if err == io.EOF {
					break
				}
				if err != nil {
					log.Fatalf("Failed to read: %v", err)
				}

				// Create a point and add to batch
				tags := make(map[string]string)
				fields := make(map[string]interface{})

				if err != nil {
					log.Fatalln("Error: ", err)
				}

				// variables initialization
				var prefix string

				// Search for Prefix if exist
				for _, v := range r.Kv {
					if v.Key == "__prefix__" {
						prefix = v.GetStrValue()
					}
				}

				// Extract information from prefix if Exist
				// TODO make this block a function
				if prefix != "" {

					// WIll search for attribute and will extract
					// example : /junos/interface[name=xe-0/0/0]/test
					// - the name of the element   		(interface)
					// - the name of the attribute 		(name)
					// - the value of the attribute		(xe-0/0/0)
					re := regexp.MustCompile("\\/([^\\/]*)\\[([A-Za-z0-9\\-]*)\\=([^\\[]*)\\]")
					subs := re.FindAllStringSubmatch(prefix, -1)

					if len(subs) > 0 {
						for _, sub := range subs {

							// if the  attribute name is "name"
							// Extract the name of the element as "key" and the value of the attribute as value
							// /junos/interface[name=xe-0/0/0]/test
							if sub[2] == "name" {
								sub[3]= strings.Replace(sub[3], "'", "", -1)
								tags[sub[1]] = sub[3]
							}
						}
					}
				}

				// Insert additional tags
				tags["device"] = grpc_server

				for _, v := range r.Kv {
					switch v.Value.(type) {
					case *Telemetry.KeyValue_StrValue:
						tags[v.Key] = v.GetStrValue()
						break
					case *Telemetry.KeyValue_DoubleValue:
						fields[v.Key] = v.GetDoubleValue()
						break
					case *Telemetry.KeyValue_IntValue:
						fields[v.Key] = v.GetIntValue()
						break
					case *Telemetry.KeyValue_UintValue:
						fields[v.Key] = v.GetUintValue()
						break
					case *Telemetry.KeyValue_SintValue:
						fields[v.Key] = v.GetSintValue()
						break
					case *Telemetry.KeyValue_BoolValue:
						fields[v.Key] = v.GetBoolValue()
						break
					case *Telemetry.KeyValue_BytesValue:
						fields[v.Key] = v.GetBytesValue()
						break
					default:
						fields[v.Key] = v.Value
						log.Println(v.GetValue())
					}
				}

				acc.AddFields(sensorName, fields, tags, time.Now())
			}
		}(sensor, acc)

	}
	wg.Wait()

	return nil
}

func init() {
	inputs.Add("openconfig_telemetry", func() telegraf.Input {
		return &OpenConfigTelemetry{}
	})
}
