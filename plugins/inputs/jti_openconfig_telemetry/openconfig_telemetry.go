package jti_openconfig_telemetry

import (
	"fmt"
	"log"
	"net"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/inputs/jti_openconfig_telemetry/auth"
	"github.com/influxdata/telegraf/plugins/inputs/jti_openconfig_telemetry/oc"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

type OpenConfigTelemetry struct {
	Servers         []string
	Sensors         []string
	Username        string
	Password        string
	ClientID        string            `toml:"client_id"`
	SampleFrequency internal.Duration `toml:"sample_frequency"`
	SSLCert         string            `toml:"ssl_cert"`
	StrAsTags       bool              `toml:"str_as_tags"`
	RetryDelay      internal.Duration `toml:"retry_delay"`

	sensorsConfig   []sensorConfig
	grpcClientConns []*grpc.ClientConn
	wg              *sync.WaitGroup
}

var (
	// Regex to match and extract data points from path value in received key
	keyPathRegex = regexp.MustCompile("\\/([^\\/]*)\\[([A-Za-z0-9\\-\\/]*\\=[^\\[]*)\\]")
	sampleConfig = `
  ## List of device addresses to collect telemetry from
  servers = ["localhost:1883"]

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
  ## reporting rate in Duration at the beginning of sensor paths / collection 
  ## name. For entries without reporting rate, we use configured sample frequency
  sensors = [
   "1000ms customReporting /interfaces /lldp",
   "2000ms collection /components",
   "/interfaces",
  ]

  ## x509 Certificate to use with TLS connection. If it is not provided, an insecure 
  ## channel will be opened with server
  ssl_cert = "/etc/telegraf/cert.pem"

  ## Delay between retry attempts of failed RPC calls or streams. Defaults to 1000ms.
  ## Failed streams/calls will not be retried if 0 is provided
  retry_delay = "1000ms"

  ## To treat all string values as tags, set this to true
  str_as_tags = false
`
)

func (m *OpenConfigTelemetry) SampleConfig() string {
	return sampleConfig
}

func (m *OpenConfigTelemetry) Description() string {
	return "Read JTI OpenConfig Telemetry from listed sensors"
}

func (m *OpenConfigTelemetry) Gather(acc telegraf.Accumulator) error {
	return nil
}

func (m *OpenConfigTelemetry) Stop() {
	for _, grpcClientConn := range m.grpcClientConns {
		grpcClientConn.Close()
	}
	m.wg.Wait()
}

// Takes in XML path with predicates and returns list of tags+values along with a final
// XML path without predicates. If /events/event[id=2]/attributes[key='message']/value
// is given input, this function will emit /events/event/attributes/value as xmlpath and
// { /events/event/@id=2, /events/event/attributes/@key='message' } as tags
func spitTagsNPath(xmlpath string) (string, map[string]string) {
	subs := keyPathRegex.FindAllStringSubmatch(xmlpath, -1)
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

// Takes in a OC response, extracts tag information from keys and returns a
// list of groups with unique sets of tags+values
func (m *OpenConfigTelemetry) extractData(r *telemetry.OpenConfigData, grpcServer string) []DataGroup {
	// Use empty prefix. We will update this when we iterate over key-value pairs
	prefix := ""

	dgroups := []DataGroup{}

	for _, v := range r.Kv {
		kv := make(map[string]interface{})

		if v.Key == "__prefix__" {
			prefix = v.GetStrValue()
			continue
		}

		// Also, lets use prefix if there is one
		xmlpath, finaltags := spitTagsNPath(prefix + v.Key)
		finaltags["device"] = grpcServer

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
		finaltags["system_id"] = r.SystemId
		finaltags["path"] = r.Path

		// Insert derived key and value
		dgroups = CollectionByKeys(dgroups).Insert(finaltags, kv)

		// Insert data from message header
		dgroups = CollectionByKeys(dgroups).Insert(finaltags,
			map[string]interface{}{"_sequence": r.SequenceNumber})
		dgroups = CollectionByKeys(dgroups).Insert(finaltags,
			map[string]interface{}{"_timestamp": r.Timestamp})
		dgroups = CollectionByKeys(dgroups).Insert(finaltags,
			map[string]interface{}{"_component_id": r.ComponentId})
		dgroups = CollectionByKeys(dgroups).Insert(finaltags,
			map[string]interface{}{"_subcomponent_id": r.SubComponentId})
	}

	return dgroups
}

// Structure to hold sensors path list and measurement name
type sensorConfig struct {
	measurementName string
	pathList        []*telemetry.Path
}

// Takes in sensor configuration and converts it into slice of sensorConfig objects
func (m *OpenConfigTelemetry) splitSensorConfig() int {
	var pathlist []*telemetry.Path
	var measurementName string
	var reportingRate uint32

	m.sensorsConfig = make([]sensorConfig, 0)
	for _, sensor := range m.Sensors {
		spathSplit := strings.Fields(sensor)
		reportingRate = uint32(m.SampleFrequency.Duration / time.Millisecond)

		// Extract measurement name and custom reporting rate if specified. Custom
		// reporting rate will be specified at the beginning of sensor list,
		// followed by measurement name like "1000ms interfaces /interfaces"
		// where 1000ms is the custom reporting rate and interfaces is the
		// measurement name. If 1000ms is not given, we use global reporting rate
		// from sample_frequency. if measurement name is not given, we use first
		// sensor name as the measurement name. If first or the word after custom
		// reporting rate doesn't start with /, we treat it as measurement name
		// and exclude it from list of sensors to subscribe
		duration, err := time.ParseDuration(spathSplit[0])
		if err == nil {
			reportingRate = uint32(duration / time.Millisecond)
			spathSplit = spathSplit[1:]
		}

		if len(spathSplit) == 0 {
			log.Printf("E! No sensors are specified")
			continue
		}

		// Word after custom reporting rate is treated as measurement name
		measurementName = spathSplit[0]

		// If our word after custom reporting rate doesn't start with /, we treat
		// it as measurement name. Else we treat it as sensor
		if !strings.HasPrefix(measurementName, "/") {
			spathSplit = spathSplit[1:]
		}

		if len(spathSplit) == 0 {
			log.Printf("E! No valid sensors are specified")
			continue
		}

		// Iterate over our sensors and create pathlist to subscribe
		pathlist = make([]*telemetry.Path, 0)
		for _, path := range spathSplit {
			pathlist = append(pathlist, &telemetry.Path{Path: path,
				SampleFrequency: reportingRate})
		}

		m.sensorsConfig = append(m.sensorsConfig, sensorConfig{
			measurementName: measurementName, pathList: pathlist,
		})

	}

	return len(m.sensorsConfig)
}

// Subscribes and collects OpenConfig telemetry data from given server
func (m *OpenConfigTelemetry) collectData(ctx context.Context,
	grpcServer string, grpcClientConn *grpc.ClientConn,
	acc telegraf.Accumulator) error {
	c := telemetry.NewOpenConfigTelemetryClient(grpcClientConn)
	for _, sensor := range m.sensorsConfig {
		m.wg.Add(1)
		go func(ctx context.Context, sensor sensorConfig) {
			defer m.wg.Done()

			for {
				stream, err := c.TelemetrySubscribe(ctx,
					&telemetry.SubscriptionRequest{PathList: sensor.pathList})
				if err != nil {
					rpcStatus, _ := status.FromError(err)
					// If service is currently unavailable and may come back later, retry
					if rpcStatus.Code() != codes.Unavailable {
						acc.AddError(fmt.Errorf("E! Could not subscribe to %s: %v", grpcServer,
							err))
						return
					} else {
						// Retry with delay. If delay is not provided, use default
						if m.RetryDelay.Duration > 0 {
							log.Printf("D! Retrying %s with timeout %v", grpcServer,
								m.RetryDelay.Duration)
							time.Sleep(m.RetryDelay.Duration)
							continue
						} else {
							return
						}
					}
				}
				for {
					r, err := stream.Recv()
					if err != nil {
						// If we encounter error in the stream, break so we can retry
						// the connection
						acc.AddError(fmt.Errorf("E! Failed to read from %s: %v", err, grpcServer))
						break
					}

					log.Printf("D! Received from %s: %v", grpcServer, r)

					// Create a point and add to batch
					tags := make(map[string]string)

					// Insert additional tags
					tags["device"] = grpcServer

					dgroups := m.extractData(r, grpcServer)

					// Print final data collection
					log.Printf("D! Available collection for %s is: %v", grpcServer, dgroups)

					tnow := time.Now()
					// Iterate through data groups and add them
					for _, group := range dgroups {
						if len(group.tags) == 0 {
							acc.AddFields(sensor.measurementName, group.data, tags, tnow)
						} else {
							acc.AddFields(sensor.measurementName, group.data, group.tags, tnow)
						}
					}
				}
			}
		}(ctx, sensor)
	}

	return nil
}

func (m *OpenConfigTelemetry) Start(acc telegraf.Accumulator) error {
	// Build sensors config
	if m.splitSensorConfig() == 0 {
		return fmt.Errorf("E! No valid sensor configuration available")
	}

	// If SSL certificate is provided, use transport credentials
	var err error
	var transportCredentials credentials.TransportCredentials
	if m.SSLCert != "" {
		transportCredentials, err = credentials.NewClientTLSFromFile(m.SSLCert, "")
		if err != nil {
			return fmt.Errorf("E! Failed to read certificate: %v", err)
		}
	} else {
		transportCredentials = nil
	}

	// Connect to given list of servers and start collecting data
	var grpcClientConn *grpc.ClientConn
	var wg sync.WaitGroup
	ctx := context.Background()
	m.wg = &wg
	for _, server := range m.Servers {
		// Extract device address and port
		grpcServer, grpcPort, err := net.SplitHostPort(server)
		if err != nil {
			log.Printf("E! Invalid server address: %v", err)
			continue
		}

		// If a certificate is provided, open a secure channel. Else open insecure one
		if transportCredentials != nil {
			grpcClientConn, err = grpc.Dial(server, grpc.WithTransportCredentials(transportCredentials))
		} else {
			grpcClientConn, err = grpc.Dial(server, grpc.WithInsecure())
		}
		if err != nil {
			log.Printf("E! Failed to connect to %s: %v", server, err)
		} else {
			log.Printf("D! Opened a new gRPC session to %s on port %s", grpcServer, grpcPort)
		}

		// Add to the list of client connections
		m.grpcClientConns = append(m.grpcClientConns, grpcClientConn)

		if m.Username != "" && m.Password != "" && m.ClientID != "" {
			lc := authentication.NewLoginClient(grpcClientConn)
			loginReply, loginErr := lc.LoginCheck(ctx,
				&authentication.LoginRequest{UserName: m.Username,
					Password: m.Password, ClientId: m.ClientID})
			if loginErr != nil {
				log.Printf("E! Could not initiate login check for %s: %v", server, loginErr)
				continue
			}

			// Check if the user is authenticated. Bail if auth error
			if !loginReply.Result {
				log.Printf("E! Failed to authenticate the user for %s", server)
				continue
			}
		}

		// Subscribe and gather telemetry data
		m.collectData(ctx, grpcServer, grpcClientConn, acc)
	}

	return nil
}

func init() {
	inputs.Add("jti_openconfig_telemetry", func() telegraf.Input {
		return &OpenConfigTelemetry{
			RetryDelay: internal.Duration{Duration: time.Second},
			StrAsTags:  false,
		}
	})
}
