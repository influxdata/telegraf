package jti_openconfig_telemetry

import (
	"fmt"
	"io"
	"log"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/inputs/jti_openconfig_telemetry/auth"
	"github.com/influxdata/telegraf/plugins/inputs/jti_openconfig_telemetry/oc"
	"github.com/influxdata/telegraf/plugins/parsers"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type OpenConfigTelemetry struct {
	Server          string
	Sensors         []string
	Username        string
	Password        string
	ClientID        string            `toml:"client_id"`
	SampleFrequency internal.Duration `toml:"sample_frequency"`
	SSLCert         string            `toml:"ssl_cert"`
	StrAsTags       bool              `toml:"str_as_tags"`

	parser         parsers.Parser
	grpcClientConn *grpc.ClientConn
	wg             *sync.WaitGroup
}

var sampleConfig = `
  ## Device address to collect telemetry from
  server = "localhost:1883"

  ## Authentication details. Username and password are must if device expects 
  ## authentication. Client ID must be unique when connecting from multiple instances 
  ## of telegraf to the same device
  username = "user"
  password = "pass"
  client_id = "telegraf"

  ## Frequency to get data
  sample_frequency = "1000ms"

  ## Sensors to subscribe for
  ## A identifier for each sensor can be provided in path by separating with space
  ## Else sensor path will be used as identifier
  ## When identifier is used, we can provide a list of space separated sensors. 
  ## A single subscription will be created with all these sensors and data will 
  ## be saved to measurement with this identifier name
  sensors = [
   "/interfaces/",
   "collection /components/ /lldp",
  ]

  ## We allow specifying sensor group level reporting rate. To do this, specify the 
  ## reporting rate in milliseconds at the beginning of sensor paths / collection 
  ## name. For entries without reporting rate, we use configured sample frequency
  sensors = [
   "1000 customReporting /interfaces /lldp",
   "2000 collection /components",
   "/interfaces",
  ]

  ## x509 Certificate to use with TLS connection. If it is not provided, an insecure 
  ## channel will be opened with server
  ssl_cert = "/etc/telegraf/cert.pem"

  ## To treat all string values as tags, set this to true
  str_as_tags = false
`

func (m *OpenConfigTelemetry) SampleConfig() string {
	return sampleConfig
}

func (m *OpenConfigTelemetry) Description() string {
	return "Read JTI OpenConfig Telemetry from listed sensors"
}

func (m *OpenConfigTelemetry) SetParser(parser parsers.Parser) {
	m.parser = parser
}

func (m *OpenConfigTelemetry) Gather(acc telegraf.Accumulator) error {
	return nil
}

func (m *OpenConfigTelemetry) Stop() {
	m.grpcClientConn.Close()
	m.wg.Wait()
}

// Takes in XML path with predicates and returns list of tags+values along with a final
// XML path without predicates. If /events/event[id=2]/attributes[key='message']/value
// is given input, this function will emit /events/event/attributes/value as xmlpath and
// { /events/event/@id=2, /events/event/attributes/@key='message' } as tags
func spitTagsNPath(xmlpath string) (string, map[string]string) {
	re := regexp.MustCompile("\\/([^\\/]*)\\[([A-Za-z0-9\\-\\/]*\\=[^\\[]*)\\]")
	subs := re.FindAllStringSubmatch(xmlpath, -1)
	tags := make(map[string]string)

	// Given XML path, this will spit out final path without predicates
	if len(subs) > 0 {
		for _, sub := range subs {
			tagKey := strings.Split(xmlpath, sub[0])[0] + "/" + strings.TrimSpace(sub[1]) + "/@"

			// If we have multiple keys in give path like /events/event[id=2 and type=3]/,
			// we must emit multiple tags
			for _, kv := range strings.Split(sub[2], " and ") {
				key := tagKey + strings.TrimSpace(strings.Split(kv, "=")[0])
				tagValue := strings.Replace(strings.Split(kv, "=")[1], "'", "", -1)
				tags[key] = tagValue
			}

			xmlpath = strings.Replace(xmlpath, sub[0], "/"+strings.TrimSpace(sub[1]), 1)
		}
	}

	return xmlpath, tags
}

func (m *OpenConfigTelemetry) Start(acc telegraf.Accumulator) error {
	log.Print("D! Started JTI OpenConfig Telemetry plugin\n")

	// Extract device name / IP
	s := strings.Split(m.Server, ":")
	grpc_server, grpc_port := s[0], s[1]

	var err error
	var wg sync.WaitGroup
	m.wg = &wg

	// If a certificate is provided, open a secure channel. Else open insecure one
	if m.SSLCert != "" {
		creds, err := credentials.NewClientTLSFromFile(m.SSLCert, "")
		if err != nil {
			return fmt.Errorf("E! Failed to read certificate: %v", err)
		}
		m.grpcClientConn, err = grpc.Dial(m.Server, grpc.WithTransportCredentials(creds))
	} else {
		m.grpcClientConn, err = grpc.Dial(m.Server, grpc.WithInsecure())
	}
	if err != nil {
		return fmt.Errorf("E! Failed to connect: %v", err)
	}

	log.Printf("D! Opened a new gRPC session to %s on port %s", grpc_server, grpc_port)

	// If username, password and clientId are provided, authenticate user before subscribing for data
	if m.Username != "" && m.Password != "" && m.ClientID != "" {
		lc := authentication.NewLoginClient(m.grpcClientConn)
		loginReply, loginErr := lc.LoginCheck(context.Background(),
			&authentication.LoginRequest{UserName: m.Username, Password: m.Password,
				ClientId: m.ClientID})
		if loginErr != nil {
			return fmt.Errorf("E! Could not initiate login check: %v", err)
		}

		// Check if the user is authenticated. Bail if auth error
		if !loginReply.Result {
			return fmt.Errorf("E! Failed to authenticate the user")
		}
	}

	c := telemetry.NewOpenConfigTelemetryClient(m.grpcClientConn)

	for _, sensor := range m.Sensors {
		wg.Add(1)
		go func(sensor string, acc telegraf.Accumulator) {
			defer wg.Done()

			spathSplit := strings.SplitN(sensor, " ", -1)
			var sensorName string
			var pathlist []*telemetry.Path
			customReportingRate := false
			var reportingRate uint64

			reportingRate = uint64(m.SampleFrequency.Duration.Nanoseconds() / int64(time.Millisecond))

			if len(spathSplit) > 1 {
				for i, path := range spathSplit {
					// We allow custom reporting rate per sensor by specifying rate in milliseconds
					// followed by measurement name or sensor path
					if i == 0 {
						sensorName = path
						reportingRate, err = strconv.ParseUint(path, 10, 32)
						if err == nil {
							customReportingRate = true
						} else {
							reportingRate = uint64(m.SampleFrequency.Duration.Nanoseconds() / int64(time.Millisecond))
							// If the first word is not an integer and starts with /, we treat it as both sensor and a measurement name
							if strings.HasPrefix(sensorName, "/") {
								pathlist = append(pathlist, &telemetry.Path{Path: path, SampleFrequency: uint32(reportingRate)})
							}
						}
					} else if customReportingRate && i == 1 {
						// If our first word is reporting rate for this list of sensors, we treat second word as measurement name and if it has /
						// we treat it as another sensor to be monitored
						sensorName = path
						if strings.HasPrefix(path, "/") {
							pathlist = append(pathlist, &telemetry.Path{Path: path, SampleFrequency: uint32(reportingRate)})
						}
					} else {
						pathlist = append(pathlist, &telemetry.Path{Path: path, SampleFrequency: uint32(reportingRate)})
					}
				}
			} else {
				sensorName = sensor
				pathlist = append(pathlist, &telemetry.Path{Path: sensor, SampleFrequency: uint32(reportingRate)})
			}
			stream, err := c.TelemetrySubscribe(context.Background(),
				&telemetry.SubscriptionRequest{PathList: pathlist})
			if err != nil {
				acc.AddError(fmt.Errorf("E! Could not subscribe: %v", err))
				return
			}
			for {
				r, err := stream.Recv()
				if err == io.EOF {
					break
				}
				if err != nil {
					acc.AddError(fmt.Errorf("E! Failed to read: %v", err))
					return
				}

				log.Printf("D! Received: %v", r)

				// Create a point and add to batch
				tags := make(map[string]string)

				if err != nil {
					acc.AddError(fmt.Errorf("E! Error: %v", err))
					return
				}

				// Use empty prefix. We will update this when we iterate over key-value pairs
				prefix := ""

				// Insert additional tags
				tags["device"] = grpc_server

				dgroups := []DataGroup{}

				for _, v := range r.Kv {
					kv := make(map[string]interface{})

					if v.Key == "__prefix__" {
						prefix = v.GetStrValue()
						continue
					}

					// Also, lets use prefix if there is one
					xmlpath, finaltags := spitTagsNPath(prefix + v.Key)
					finaltags["device"] = grpc_server

					switch v.Value.(type) {
					case *telemetry.KeyValue_StrValue:
						// If StrAsTags is set, we treat all string values as tags
						if m.StrAsTags {
							finaltags[xmlpath] = v.GetStrValue()
						} else {
							kv[xmlpath] = v.GetStrValue()
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
					}

					// Insert other tags from message
					finaltags["_system_id"] = r.SystemId
					finaltags["_path"] = r.Path

					// Insert derived key and value
					dgroups = CollectionByKeys(dgroups).Insert(finaltags, kv)

					// Insert data from message header
					dgroups = CollectionByKeys(dgroups).Insert(finaltags, map[string]interface{}{"_sequence": r.SequenceNumber})
					dgroups = CollectionByKeys(dgroups).Insert(finaltags, map[string]interface{}{"_timestamp": r.Timestamp})
					dgroups = CollectionByKeys(dgroups).Insert(finaltags, map[string]interface{}{"_component_id": r.ComponentId})
					dgroups = CollectionByKeys(dgroups).Insert(finaltags, map[string]interface{}{"_subcomponent_id": r.SubComponentId})
				}

				// Print final data collection
				log.Printf("D! Available collection is: %v", dgroups)

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

	return nil
}

func init() {
	inputs.Add("jti_openconfig_telemetry", func() telegraf.Input {
		return &OpenConfigTelemetry{}
	})
}
