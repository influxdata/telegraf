package openconfig_telemetry

import (
	"io"
	"log"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/inputs/openconfig_telemetry/oc"
	"github.com/influxdata/telegraf/plugins/parsers"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type OpenConfigTelemetry struct {
	Server          string
	Sensors         []string
	SampleFrequency uint32
	CertFile        string
	Debug           bool

	parser parsers.Parser
	sync.Mutex

	// keep the accumulator internally:
	acc            telegraf.Accumulator
	grpcClientConn *grpc.ClientConn
}

var sampleConfig = `
  server = "localhost:1883"

  ## Frequency to get data in milliseconds
  sampleFrequency = 1000

  ## Sensors to subscribe for
  ## A identifier for each sensor can be provided in path by separating with space
  ## Else sensor path will be used as identifier
  sensors = [
   "/oc/firewall/usage",
   "interfaces /oc/interfaces/",
  ]

  ## x509 Certificate to use with TLS connection. If it is not provided, an insecure 
  ## channel will be opened with server
  certFile = "/path/to/x509_cert_file"

  ## To see data being received from gRPC server, set debug to true
  debug = true
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
	log.Print("I! Started OpenConfig Telemetry plugin\n")
	return nil
}

func (m *OpenConfigTelemetry) Stop() {
	m.Lock()
	defer m.Unlock()
	m.grpcClientConn.Close()
}

// Takes in XML path with predicates and returns list of tags+values along with a final XML path without predicates
func spitTagsNPath(xmlpath string) (string, map[string]string) {
	re := regexp.MustCompile("\\/([^\\/]*)\\[([A-Za-z0-9\\-\\/]*)\\=([^\\[]*)\\]")
	subs := re.FindAllStringSubmatch(xmlpath, -1)
	tags := make(map[string]string)

	// Given XML path, this will spit out final path without predicates
	if len(subs) > 0 {
		for _, sub := range subs {
			tagKey := strings.Split(xmlpath, sub[0])[0]
			tagKey += "/" + sub[1] + "/@" + sub[2]
			tagValue := strings.Replace(sub[3], "'", "", -1)

			tags[tagKey] = tagValue
			xmlpath = strings.Replace(xmlpath, sub[0], "/"+sub[1], 1)
		}
	}

	return xmlpath, tags
}

func (m *OpenConfigTelemetry) Gather(acc telegraf.Accumulator) error {
	m.Lock()
	defer m.Unlock()

	m.acc = acc

	// Extract device name / IP
	s := strings.Split(m.Server, ":")
	grpc_server, grpc_port := s[0], s[1]

	var err error

	// If a certificate is provided, open a secure channel. Else open insecure one
	if m.CertFile != "" {
		creds, err := credentials.NewClientTLSFromFile(m.CertFile, "")
		if err != nil {
			log.Fatalf("E! Failed to read certificate: %v", err)
		}
		m.grpcClientConn, err = grpc.Dial(m.Server, grpc.WithTransportCredentials(creds))
	} else {
		m.grpcClientConn, err = grpc.Dial(m.Server, grpc.WithInsecure())
	}
	if err != nil {
		log.Fatalf("E! Failed to connect: %v", err)
	}

	c := telemetry.NewOpenConfigTelemetryClient(m.grpcClientConn)
	if m.Debug {
		log.Printf("I! Opened a new gRPC session to %s on port %s", grpc_server, grpc_port)
	}

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
				&telemetry.SubscriptionRequest{PathList: []*telemetry.Path{&telemetry.Path{Path: sensorPath,
					SampleFrequency: m.SampleFrequency}}})
			if err != nil {
				log.Fatalf("E! Could not subscribe: %v", err)
			}
			for {
				r, err := stream.Recv()
				if err == io.EOF {
					break
				}
				if err != nil {
					log.Fatalf("E! Failed to read: %v", err)
				}

				// Print incoming data as info if debug is set
				if m.Debug {
					log.Printf("I! Received: ", r)
				}

				// Create a point and add to batch
				tags := make(map[string]string)

				if err != nil {
					log.Fatalln("E! Error: %v", err)
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

					// Will search for attribute and will extract
					// example : /junos/interface[name=xe-0/0/0]/test
					// - the name of the element   		(interface)
					// - the name of the attribute 		(name)
					// - the value of the attribute		(xe-0/0/0)
					re := regexp.MustCompile("\\/([^\\/]*)\\[([A-Za-z0-9\\-\\/]*)\\=([^\\[]*)\\]")
					subs := re.FindAllStringSubmatch(prefix, -1)

					if len(subs) > 0 {
						for _, sub := range subs {

							// if the  attribute name is "name",
							// Extract the name of the element as "key" and the value of the attribute as value
							// /junos/interface[name=xe-0/0/0]/test
							if sub[2] == "name" {
								sub[3] = strings.Replace(sub[3], "'", "", -1)
								if sub[1] == "" {
									tags[sub[2]] = sub[3]
								} else {
									tags[sub[1]] = sub[3]
								}
							}
						}
					}
				}

				// Insert additional tags
				tags["device"] = grpc_server

				dgroups := []DataGroup{}

				for _, v := range r.Kv {
					kv := make(map[string]interface{})
					xmlpath, finaltags := spitTagsNPath(v.Key)
					finaltags["device"] = grpc_server

					switch v.Value.(type) {
					case *telemetry.KeyValue_StrValue:
						// If this is actually a integer value but wrongly encoded as string,
						// convert and use it as value to field
						if val, err := strconv.ParseInt(v.GetStrValue(), 10, 64); err == nil {
							kv[xmlpath] = val
						} else {
							tags[v.Key] = v.GetStrValue()
						}
						break
					case *telemetry.KeyValue_DoubleValue:
						kv[xmlpath] = v.GetDoubleValue()
						break
					case *telemetry.KeyValue_IntValue:
						kv[xmlpath] = v.GetIntValue()
						break
					case *telemetry.KeyValue_UintValue:
						kv[xmlpath] = v.GetUintValue()
						break
					case *telemetry.KeyValue_SintValue:
						kv[xmlpath] = v.GetSintValue()
						break
					case *telemetry.KeyValue_BoolValue:
						kv[xmlpath] = v.GetBoolValue()
						break
					case *telemetry.KeyValue_BytesValue:
						kv[xmlpath] = v.GetBytesValue()
						break
					default:
						log.Println(v.GetValue())
					}
					dgroups = CollectionByKeys(dgroups).Insert(finaltags, kv)
				}

				// Print final data collection
				if m.Debug {
					log.Printf("I! Available collection is: ", dgroups)
				}

				tnow := time.Now()
				// Iterate through data groups and add them
				for _, group := range dgroups {
					if len(group.tags) == 0 {
						acc.AddFields(sensorName, group.data, tags, tnow)
					} else {
						acc.AddFields(sensorName, group.data, group.tags, tnow)
					}
				}
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
