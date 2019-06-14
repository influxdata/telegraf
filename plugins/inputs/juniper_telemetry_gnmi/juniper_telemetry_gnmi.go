package juniper_telemetry_gnmi

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	internaltls "github.com/influxdata/telegraf/internal/tls"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/inputs"
	jsonparser "github.com/influxdata/telegraf/plugins/parsers/json"

	"github.com/openconfig/gnmi/proto/gnmi"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

// JuniperTelemetryGNMI plugin instance
type JuniperTelemetryGNMI struct {
	Servers           []string `toml:"servers"`
	Username          string
	Password          string
	Origin            string
	Paths             []string
	SubscriptionMode  string            `toml:"subscription_mode"`
	SampleInterval    internal.Duration `toml:"sample_interval"`
	SuppressRedundant bool              `toml:"suppress_redundant"`

	// Optional subscription configuration
	Encoding    string
	Prefix      string
	Target      string
	UpdatesOnly bool `toml:"updates_only"`

	// Redial
	Redial internal.Duration

	// GRPC settings
	EnableTLS bool `toml:"enable_tls"`
	internaltls.ClientConfig

	// Internal state
	acc    telegraf.Accumulator
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

const sampleConfig = `
 ## Address and port of the GNMI GRPC server
 servers = ["127.0.0.1:50051"]
 ## define credentials
 username = "juniper"
 password = "juniper"
 ## GNMI encoding requested (one of: "proto", "json", "json_ietf")
 # encoding = "proto"
 ## redial in case of failures after
 redial = "10s"
 ## enable client-side TLS and define CA to authenticate the device
 # enable_tls = true
 # tls_ca = "/etc/telegraf/ca.pem"
 # insecure_skip_verify = true
 ## define client-side TLS certificate & key to authenticate to the device
 # tls_cert = "/etc/telegraf/cert.pem"
 # tls_key = "/etc/telegraf/key.pem"
 ## GNMI subscription prefix (optional, can usually be left empty)
 ## See: https://github.com/openconfig/reference/blob/master/rpc/gnmi/gnmi-specification.md#222-paths
 # origin = ""
 # prefix = ""
 # target = ""
 ## Define additional aliases to map telemetry encoding paths to simple measurement names
 #[inputs.juniper_telemetry_gnmi.aliases]
 #  ifcounters = "openconfig:/interfaces/interface/state/counters"
  ## Name of the measurement that will be emitted
  name = "ifcounters"
  ## Origin and path of the subscription
  ## See: https://github.com/openconfig/reference/blob/master/rpc/gnmi/gnmi-specification.md#222-paths
  ##
  ## origin usually refers to a (YANG) data model implemented by the device
  ## and path to a specific substructe inside it that should be subscribed to (similar to an XPath)
  ## YANG models can be found e.g. here: https://github.com/YangModels/yang/tree/master/vendor/juniper
  origin = "openconfig-interfaces"
  paths = ["/interfaces/interface/state/counters"]
  # Subscription mode (one of: "target_defined", "sample", "on_change") and interval
  subscription_mode = "sample"
  sample_interval = "10s"
  ## Suppress redundant transmissions when measured values are unchanged
  # suppress_redundant = false
  ## If suppression is enabled, send updates at least every X seconds anyway
  # heartbeat_interval = "60s"
`

func (j *JuniperTelemetryGNMI) SampleConfig() string {
	return sampleConfig
}

func (j *JuniperTelemetryGNMI) Description() string {
	return "Juniper GNMI telemetry input plugin based on GNMI telemetry data produced in Junos"
}

func (j *JuniperTelemetryGNMI) Gather(acc telegraf.Accumulator) error {
	return nil
}

// Stop all gRPC Connections and cleanup
func (j *JuniperTelemetryGNMI) Stop() {
	j.cancel()
	j.wg.Wait()
}

func (j *JuniperTelemetryGNMI) Start(acc telegraf.Accumulator) error {

	var err error
	var ctx context.Context
	var tlscfg *tls.Config

	j.acc = acc

	ctx, j.cancel = context.WithCancel(context.Background())

	if j.Redial.Duration.Nanoseconds() <= 0 {
		return fmt.Errorf("redial duration must be positive")
	}

	// Parse TLS config
	if j.EnableTLS {
		if tlscfg, err = j.ClientConfig.TLSConfig(); err != nil {
			return err
		}
	}

	if len(j.Username) > 0 {
		ctx = metadata.AppendToOutgoingContext(ctx, "username", j.Username, "password", j.Password)
	}

	// Create a goroutine for each device, dial and subscribe
	for _, server := range j.Servers {
		j.wg.Add(1)
		go func(server string) {
			defer j.wg.Done()
			for ctx.Err() == nil {
				if err := j.CollectData(ctx, server, tlscfg, acc); err != nil && ctx.Err() == nil {
					acc.AddError(fmt.Errorf("failed to collect data from:%s %v", server, err))
				}

				select {
				case <-ctx.Done():
				case <-time.After(j.Redial.Duration):
				}
			}
		}(server)
	}
	return nil
}

// Create a new GNMI SubscribeRequest
func (j *JuniperTelemetryGNMI) newSubscribeRequest() (*gnmi.SubscribeRequest, error) {
	// Create subscription objects
	mode, ok := gnmi.SubscriptionMode_value[strings.ToUpper(j.SubscriptionMode)]
	if !ok {
		return nil, fmt.Errorf("invalid subscription mode %s", j.SubscriptionMode)
	}

	//prefixPath, _ := parsePath(j.Origin,j.Prefix,j.Target)    ---- Juniper does not support Prefix

	subList := &gnmi.SubscriptionList{
		Subscription: make([]*gnmi.Subscription, len(j.Paths)),
		Mode:         gnmi.SubscriptionList_STREAM,
		Encoding:     gnmi.Encoding(gnmi.Encoding_value[strings.ToUpper(j.Encoding)]),
		UpdatesOnly:  j.UpdatesOnly,
		//Prefix:       prefixPath,
	}
	for i, path := range j.Paths {
		gnmiPath, err := parsePath(j.Origin, path, j.Target)
		if err != nil {
			return nil, err
		}
		subList.Subscription[i] = &gnmi.Subscription{
			Path:              gnmiPath,
			Mode:              gnmi.SubscriptionMode(mode),
			SampleInterval:    uint64(j.SampleInterval.Duration.Nanoseconds()),
			SuppressRedundant: j.SuppressRedundant,
		}
	}

	//Juniper only supports proto as the encoding for gNMI
	if j.Encoding != "proto" {
		return nil, fmt.Errorf("unsupported encoding %s", j.Encoding)
	}

	return &gnmi.SubscribeRequest{
		Request: &gnmi.SubscribeRequest_Subscribe{
			Subscribe: subList}}, nil
}

// CollectData collects OpenConfig telemetry data from given server
func (j *JuniperTelemetryGNMI) CollectData(ctx context.Context, server string, tlscfg *tls.Config, acc telegraf.Accumulator) error {
	var opts []grpc.DialOption

	if tlscfg != nil {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlscfg)))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}

	grpcclient, err := grpc.DialContext(ctx, server, opts...)
	if err != nil {
		acc.AddError(fmt.Errorf("failed to dial:%s %v", server, err))
	}
	gnmiclient := gnmi.NewGNMIClient(grpcclient)
	respChan := make(chan *gnmi.SubscribeResponse)
	errChan := make(chan error)
	go j.Subscribe(ctx, gnmiclient, respChan, errChan)
	for ctx.Err() == nil {
		select {
		case resp, open := <-respChan:
			if !open {
				continue
			}
			response := resp.GetResponse()
			update, ok := response.(*gnmi.SubscribeResponse_Update)
			if !ok {
				continue
			}
			j.handleSubscribeResponseUpdate(update, server)
		case err := <-errChan:
			acc.AddError(fmt.Errorf("E! Failed to read from %s: %v", server, err))
		}
	}
	return nil
}

// Subscribe sends a SubscribeRequest to the given client.
func (j *JuniperTelemetryGNMI) Subscribe(ctx context.Context, client gnmi.GNMIClient, respChan chan<- *gnmi.SubscribeResponse, errChan chan<- error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	defer close(respChan)

	stream, err := client.Subscribe(ctx)
	if err != nil {
		errChan <- err
		return
	}
	req, err := j.newSubscribeRequest()
	if err != nil {
		errChan <- err
		return
	}
	if err := stream.Send(req); err != nil {
		errChan <- err
		return
	}

	for {
		resp, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				return
			}
			errChan <- err
			return
		}
		respChan <- resp

		// For POLL subscriptions, initiate a poll request by pressing ENTER
		if j.SubscriptionMode == "poll" {
			switch resp.Response.(type) {
			case *gnmi.SubscribeResponse_SyncResponse:
				fmt.Print("Press ENTER to send a poll request: ")
				reader := bufio.NewReader(os.Stdin)
				reader.ReadString('\n')

				pollReq := &gnmi.SubscribeRequest{
					Request: &gnmi.SubscribeRequest_Poll{
						Poll: &gnmi.Poll{},
					},
				}
				if err := stream.Send(pollReq); err != nil {
					errChan <- err
					return
				}
			}
		}
	}
}

// HandleSubscribeResponse message from GNMI and parse contained telemetry data
func (j *JuniperTelemetryGNMI) handleSubscribeResponseUpdate(update *gnmi.SubscribeResponse_Update, server string) {

	var prefix string
	grouper := metric.NewSeriesGrouper()
	timestamp := time.Unix(0, update.Update.Timestamp)
	prefixTags := make(map[string]string)
	if update.Update.Prefix != nil {
		prefix = j.handlePath(update.Update.Prefix, prefixTags, "")
	}
	prefixTags["source"], _, _ = net.SplitHostPort(server)
	prefixTags["path"] = prefix

	// Parse individual Update message and create measurements
	for _, update := range update.Update.Update {
		// Prepare tags from prefix
		tags := make(map[string]string, len(prefixTags))
		for key, val := range prefixTags {
			tags[key] = val
		}
		fields := j.handleTelemetryField(update, tags, prefix)
		// Group metrics
		for key, val := range fields {
			grouper.Add(prefix, tags, timestamp, key[len(prefix)+1:], val)
		}

	}

	// Add grouped measurements
	for _, metric := range grouper.Metrics() {
		j.acc.AddMetric(metric)
	}
}

// HandleTelemetryField and add it to a measurement
func (j *JuniperTelemetryGNMI) handleTelemetryField(update *gnmi.Update, tags map[string]string, prefix string) map[string]interface{} {
	path := j.handlePath(update.Path, tags, prefix)

	var value interface{}
	var jsondata []byte

	switch val := update.Val.Value.(type) {
	case *gnmi.TypedValue_AsciiVal:
		value = val.AsciiVal
	case *gnmi.TypedValue_BoolVal:
		value = val.BoolVal
	case *gnmi.TypedValue_BytesVal:
		value = val.BytesVal
	case *gnmi.TypedValue_DecimalVal:
		value = val.DecimalVal
	case *gnmi.TypedValue_FloatVal:
		value = val.FloatVal
	case *gnmi.TypedValue_IntVal:
		value = val.IntVal
	case *gnmi.TypedValue_StringVal:
		value = val.StringVal
	case *gnmi.TypedValue_UintVal:
		value = val.UintVal
	case *gnmi.TypedValue_JsonIetfVal:
		jsondata = val.JsonIetfVal
	case *gnmi.TypedValue_JsonVal:
		jsondata = val.JsonVal
	}

	name := strings.Replace(path, "-", "_", -1)
	fields := make(map[string]interface{})
	if value != nil {
		fields[name] = value
	} else if jsondata != nil {
		if err := json.Unmarshal(jsondata, &value); err != nil {
			j.acc.AddError(fmt.Errorf("failed to parse JSON value: %v", err))
		} else {
			flattener := jsonparser.JSONFlattener{Fields: fields}
			flattener.FullFlattenJSON(name, value, true, true)
		}
	}
	return fields
}

// Parse path to path-buffer and tag-field
func (j *JuniperTelemetryGNMI) handlePath(path *gnmi.Path, tags map[string]string, prefix string) string {
	builder := bytes.NewBufferString(prefix)

	// Prefix with origin
	if len(path.Origin) > 0 {
		builder.WriteString(path.Origin)
		builder.WriteRune(':')
	}

	// Parse generic keys from prefix
	for _, elem := range path.Elem {
		builder.WriteRune('/')
		builder.WriteString(elem.Name)
		name := builder.String()

		for key, val := range elem.Key {
			key = strings.Replace(key, "-", "_", -1)

			// Use short-form of key if possible
			if _, exists := tags[key]; exists {
				tags[name+"/"+key] = val
			} else {
				tags[key] = val
			}
		}
	}

	return builder.String()
}

//ParsePath from XPath-like string to GNMI path structure
func parsePath(origin string, path string, target string) (*gnmi.Path, error) {
	var err error
	gnmiPath := gnmi.Path{Origin: origin, Target: target}

	if len(path) > 0 && path[0] != '/' {
		return nil, fmt.Errorf("path does not start with a '/': %s", path)
	}

	elem := &gnmi.PathElem{}
	start, name, value, end := 0, -1, -1, -1

	path = path + "/"

	for i := 0; i < len(path); i++ {
		if path[i] == '[' {
			if name >= 0 {
				break
			}
			if end < 0 {
				end = i
				elem.Key = make(map[string]string)
			}
			name = i + 1
		} else if path[i] == '=' {
			if name <= 0 || value >= 0 {
				break
			}
			value = i + 1
		} else if path[i] == ']' {
			if name <= 0 || value <= name {
				break
			}
			elem.Key[path[name:value-1]] = strings.Trim(path[value:i], "'\"")
			name, value = -1, -1
		} else if path[i] == '/' {
			if name < 0 {
				if end < 0 {
					end = i
				}

				if end > start {
					elem.Name = path[start:end]
					gnmiPath.Elem = append(gnmiPath.Elem, elem)
					gnmiPath.Element = append(gnmiPath.Element, path[start:i])
				}

				start, name, value, end = i+1, -1, -1, -1
				elem = &gnmi.PathElem{}
			}
		}
	}

	if name >= 0 || value >= 0 {
		err = fmt.Errorf("Invalid GNMI path: %s", path)
	}

	if err != nil {
		return nil, err
	}

	return &gnmiPath, nil
}

func init() {
	inputs.Add("juniper_telemetry_gnmi", func() telegraf.Input {
		return &JuniperTelemetryGNMI{
			Encoding: "proto", //Juniper only supports proto as the encoding for gNMI
			Redial:   internal.Duration{Duration: 10 * time.Second},
		}
	})
}
