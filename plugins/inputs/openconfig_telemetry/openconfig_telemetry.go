package openconfig_telemetry

import (
	"io"
	"log"
	"strings"
	"sync"
	"time"

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

	var err error
	m.grpcClientConn, err = grpc.Dial(m.Server, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	c := telemetry.NewOpenConfigTelemetryClient(m.grpcClientConn)

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
				&telemetry.SubscriptionRequest{PathList: []*telemetry.Path{&telemetry.Path{Path: sensorPath, SampleFrequency: m.SampleFrequency}}})
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

				for _, v := range r.Kv {
					switch v.Value.(type) {
					case *telemetry.KeyValue_StrValue:
						tags[v.Key] = v.GetStrValue()
						break
					case *telemetry.KeyValue_DoubleValue:
						fields[v.Key] = v.GetDoubleValue()
						break
					case *telemetry.KeyValue_IntValue:
						fields[v.Key] = v.GetIntValue()
						break
					case *telemetry.KeyValue_UintValue:
						fields[v.Key] = v.GetUintValue()
						break
					case *telemetry.KeyValue_SintValue:
						fields[v.Key] = v.GetSintValue()
						break
					case *telemetry.KeyValue_BoolValue:
						fields[v.Key] = v.GetBoolValue()
						break
					case *telemetry.KeyValue_BytesValue:
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
