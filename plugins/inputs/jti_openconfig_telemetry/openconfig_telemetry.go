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
	Server          string
	Sensors         []string
	Username        string
	Password        string
	ClientID        string            `toml:"client_id"`
	SampleFrequency internal.Duration `toml:"sample_frequency"`
	SSLCert         string            `toml:"ssl_cert"`
	StrAsTags       bool              `toml:"str_as_tags"`
	RetryDelay      internal.Duration `toml:"retry_delay"`

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
  ## reporting rate in Durati0on at the beginning of sensor paths / collection 
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

// Takes in a OC response, extracts tag information from keys and returns a
// list of groups with unique sets of tags+values
func extractData(r *telemetry.OpenConfigData, grpcServer string, strAsTags bool) []DataGroup {
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
			if strAsTags {
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

func (m *OpenConfigTelemetry) Start(acc telegraf.Accumulator) error {
	// Extract device name / IP
	grpcServer, grpcPort, err := net.SplitHostPort(m.Server)
	if err != nil {
		return fmt.Errorf("E! Invalid server address: %v", err)
	}

	var wg sync.WaitGroup
	m.wg = &wg

	reportingRate := uint32(m.SampleFrequency.Duration / time.Millisecond)

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

	log.Printf("D! Opened a new gRPC session to %s on port %s", grpcServer, grpcPort)

	// If username, password and clientId are provided, authenticate user before subscribing
	// for data
	ctx := context.Background()
	if m.Username != "" && m.Password != "" && m.ClientID != "" {
		lc := authentication.NewLoginClient(m.grpcClientConn)
		loginReply, loginErr := lc.LoginCheck(ctx,
			&authentication.LoginRequest{UserName: m.Username,
				Password: m.Password, ClientId: m.ClientID})
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
		go func(ctx context.Context, sensor string, reportingRate uint32,
			acc telegraf.Accumulator) {
			defer wg.Done()

			spathSplit := strings.Fields(sensor)

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
				acc.AddError(fmt.Errorf("E! No sensors are specified"))
				return
			}

			// Word after custom reporting rate is treated as measurement name
			measurementName := spathSplit[0]

			// If our word after custom reporting rate doesn't start with /, we treat
			// it as measurement name. Else we treat it as sensor
			if !strings.HasPrefix(measurementName, "/") {
				spathSplit = spathSplit[1:]
			}

			if len(spathSplit) == 0 {
				acc.AddError(fmt.Errorf("E! No valid sensors are specified"))
				return
			}

			// Iterate over our sensors and create pathlist to subscribe
			var pathlist []*telemetry.Path
			for _, path := range spathSplit {
				pathlist = append(pathlist, &telemetry.Path{Path: path,
					SampleFrequency: reportingRate})
			}

			for {
				stream, err := c.TelemetrySubscribe(ctx,
					&telemetry.SubscriptionRequest{PathList: pathlist})
				if err != nil {
					rpcStatus, _ := status.FromError(err)
					// If service is currently unavailable and may come back later, retry
					if rpcStatus.Code() != codes.Unavailable {
						acc.AddError(fmt.Errorf("E! Could not subscribe: %v", err))
						return
					} else {
						// Retry with delay. If delay is not provided, use default
						if m.RetryDelay.Duration > 0 {
							log.Printf("D! Retrying with timeout %v", m.RetryDelay.Duration)
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
						acc.AddError(fmt.Errorf("E! Failed to read: %v", err))
						break
					}

					log.Printf("D! Received: %v", r)

					// Create a point and add to batch
					tags := make(map[string]string)

					// Insert additional tags
					tags["device"] = grpcServer

					dgroups := extractData(r, grpcServer, m.StrAsTags)

					// Print final data collection
					log.Printf("D! Available collection is: %v", dgroups)

					tnow := time.Now()
					// Iterate through data groups and add them
					for _, group := range dgroups {
						if len(group.tags) == 0 {
							acc.AddFields(measurementName, group.data, tags, tnow)
						} else {
							acc.AddFields(measurementName, group.data, group.tags, tnow)
						}
					}
				}
			}
		}(ctx, sensor, reportingRate, acc)

	}

	return nil
}

func init() {
	inputs.Add("jti_openconfig_telemetry", func() telegraf.Input {
		return &OpenConfigTelemetry{
			RetryDelay: internal.Duration{Duration: time.Second},
		}
	})
}
