package jti_openconfig_telemetry

import (
	"fmt"
	"net"
	"regexp"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	internaltls "github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
	authentication "github.com/influxdata/telegraf/plugins/inputs/jti_openconfig_telemetry/auth"
	telemetry "github.com/influxdata/telegraf/plugins/inputs/jti_openconfig_telemetry/oc"
)

type OpenConfigTelemetry struct {
	Servers         []string        `toml:"servers"`
	Sensors         []string        `toml:"sensors"`
	Username        string          `toml:"username"`
	Password        string          `toml:"password"`
	ClientID        string          `toml:"client_id"`
	SampleFrequency config.Duration `toml:"sample_frequency"`
	StrAsTags       bool            `toml:"str_as_tags"`
	RetryDelay      config.Duration `toml:"retry_delay"`
	EnableTLS       bool            `toml:"enable_tls"`
	internaltls.ClientConfig

	Log telegraf.Logger

	sensorsConfig   []sensorConfig
	grpcClientConns []*grpc.ClientConn
	wg              *sync.WaitGroup
}

var (
	// Regex to match and extract data points from path value in received key
	keyPathRegex = regexp.MustCompile(`/([^/]*)\[([A-Za-z0-9\-/]*=[^\[]*)]`)
)

func (m *OpenConfigTelemetry) Gather(_ telegraf.Accumulator) error {
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
		case *telemetry.KeyValue_DoubleValue:
			kv[xmlpath] = v.GetDoubleValue()
		case *telemetry.KeyValue_IntValue:
			kv[xmlpath] = v.GetIntValue()
		case *telemetry.KeyValue_UintValue:
			kv[xmlpath] = v.GetUintValue()
		case *telemetry.KeyValue_SintValue:
			kv[xmlpath] = v.GetSintValue()
		case *telemetry.KeyValue_BoolValue:
			kv[xmlpath] = v.GetBoolValue()
		case *telemetry.KeyValue_BytesValue:
			kv[xmlpath] = v.GetBytesValue()
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
		reportingRate = uint32(time.Duration(m.SampleFrequency) / time.Millisecond)

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
			m.Log.Error("No sensors are specified")
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
			m.Log.Error("No valid sensors are specified")
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
func (m *OpenConfigTelemetry) collectData(
	ctx context.Context,
	grpcServer string,
	grpcClientConn *grpc.ClientConn,
	acc telegraf.Accumulator,
) {
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
						acc.AddError(fmt.Errorf("could not subscribe to %s: %v", grpcServer,
							err))
						return
					}

					// Retry with delay. If delay is not provided, use default
					if time.Duration(m.RetryDelay) > 0 {
						m.Log.Debugf("Retrying %s with timeout %v", grpcServer, time.Duration(m.RetryDelay))
						time.Sleep(time.Duration(m.RetryDelay))
						continue
					}
					return
				}
				for {
					r, err := stream.Recv()
					if err != nil {
						// If we encounter error in the stream, break so we can retry
						// the connection
						acc.AddError(fmt.Errorf("failed to read from %s: %s", grpcServer, err))
						break
					}

					m.Log.Debugf("Received from %s: %v", grpcServer, r)

					// Create a point and add to batch
					tags := make(map[string]string)

					// Insert additional tags
					tags["device"] = grpcServer

					dgroups := m.extractData(r, grpcServer)

					// Print final data collection
					m.Log.Debugf("Available collection for %s is: %v", grpcServer, dgroups)

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
}

func (m *OpenConfigTelemetry) Start(acc telegraf.Accumulator) error {
	// Build sensors config
	if m.splitSensorConfig() == 0 {
		return fmt.Errorf("no valid sensor configuration available")
	}

	// Parse TLS config
	var opts []grpc.DialOption
	if m.EnableTLS {
		tlscfg, err := m.ClientConfig.TLSConfig()
		if err != nil {
			return err
		}
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlscfg)))
	} else {
		opts = append(opts, grpc.WithInsecure())
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
			m.Log.Errorf("Invalid server address: %s", err.Error())
			continue
		}

		grpcClientConn, err = grpc.Dial(server, opts...)
		if err != nil {
			m.Log.Errorf("Failed to connect to %s: %s", server, err.Error())
		} else {
			m.Log.Debugf("Opened a new gRPC session to %s on port %s", grpcServer, grpcPort)
		}

		// Add to the list of client connections
		m.grpcClientConns = append(m.grpcClientConns, grpcClientConn)

		if m.Username != "" && m.Password != "" && m.ClientID != "" {
			lc := authentication.NewLoginClient(grpcClientConn)
			loginReply, loginErr := lc.LoginCheck(ctx,
				&authentication.LoginRequest{UserName: m.Username,
					Password: m.Password, ClientId: m.ClientID})
			if loginErr != nil {
				m.Log.Errorf("Could not initiate login check for %s: %v", server, loginErr)
				continue
			}

			// Check if the user is authenticated. Bail if auth error
			if !loginReply.Result {
				m.Log.Errorf("Failed to authenticate the user for %s", server)
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
			RetryDelay: config.Duration(time.Second),
			StrAsTags:  false,
		}
	})
}
