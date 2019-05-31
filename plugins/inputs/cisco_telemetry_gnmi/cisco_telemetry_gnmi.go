package cisco_telemetry_gnmi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	internaltls "github.com/influxdata/telegraf/internal/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
	jsonparser "github.com/influxdata/telegraf/plugins/parsers/json"
	"github.com/openconfig/gnmi/proto/gnmi"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

// CiscoTelemetryGNMI plugin instance
type CiscoTelemetryGNMI struct {
	Address       string         `toml:"address"`
	Subscriptions []Subscription `toml:"subscription"`

	// Optional subscription configuration
	Encoding    string
	Origin      string
	Prefix      string
	Target      string
	UpdatesOnly bool `toml:"updates_only"`

	// Cisco IOS XR credentials
	Username string
	Password string

	// Redial
	Redial internal.Duration

	// GRPC TLS settings
	EnableTLS bool `toml:"enable_tls"`
	internaltls.ClientConfig

	// Internal state
	acc    telegraf.Accumulator
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// Subscription for a GNMI client
type Subscription struct {
	Origin string
	Path   string

	// Subscription mode and interval
	SubscriptionMode string            `toml:"subscription_mode"`
	SampleInterval   internal.Duration `toml:"sample_interval"`

	// Duplicate suppression
	SuppressRedundant bool              `toml:"suppress_redundant"`
	HeartbeatInterval internal.Duration `toml:"heartbeat_interval"`
}

// Start the http listener service
func (c *CiscoTelemetryGNMI) Start(acc telegraf.Accumulator) error {
	var err error
	var ctx context.Context
	var opts []grpc.DialOption
	var request *gnmi.SubscribeRequest
	c.acc = acc
	ctx, c.cancel = context.WithCancel(context.Background())

	// Validate configuration
	if request, err = c.newSubscribeRequest(); err != nil {
		return err
	} else if c.Redial.Duration.Nanoseconds() <= 0 {
		return fmt.Errorf("redial duration must be positive")
	}

	if c.EnableTLS {
		tlsConfig, err := c.ClientConfig.TLSConfig()
		if err != nil {
			return err
		}

		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}

	if len(c.Username) > 0 {
		ctx = metadata.AppendToOutgoingContext(ctx, "username", c.Username, "password", c.Password)
	}

	client, err := grpc.DialContext(ctx, c.Address, opts...)
	if err != nil {
		return fmt.Errorf("failed to dial GNMI: %v", err)
	}

	// Dialin client telemetry stream reading routine
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		defer client.Close()

		for ctx.Err() == nil {
			if err := c.subscribeGNMI(ctx, client, request); err != nil {
				acc.AddError(err)
			}

			select {
			case <-ctx.Done():
			case <-time.After(c.Redial.Duration):
			}
		}
	}()
	return nil
}

// Create a new GNMI SubscribeRequest
func (c *CiscoTelemetryGNMI) newSubscribeRequest() (*gnmi.SubscribeRequest, error) {
	// Create subscription objects
	subscriptions := make([]*gnmi.Subscription, len(c.Subscriptions))
	for i, subscription := range c.Subscriptions {
		gnmiPath, err := parsePath(subscription.Origin, subscription.Path, "")
		if err != nil {
			return nil, err
		}
		mode, ok := gnmi.SubscriptionMode_value[strings.ToUpper(subscription.SubscriptionMode)]
		if !ok {
			return nil, fmt.Errorf("invalid subscription mode %s", subscription.SubscriptionMode)
		}
		subscriptions[i] = &gnmi.Subscription{
			Path:              gnmiPath,
			Mode:              gnmi.SubscriptionMode(mode),
			SampleInterval:    uint64(subscription.SampleInterval.Duration.Nanoseconds()),
			SuppressRedundant: subscription.SuppressRedundant,
			HeartbeatInterval: uint64(subscription.HeartbeatInterval.Duration.Nanoseconds()),
		}
	}

	// Construct subscribe request
	gnmiPath, err := parsePath(c.Origin, c.Prefix, c.Target)
	if err != nil {
		return nil, err
	}

	if c.Encoding == "" {
		c.Encoding = "proto"
	} else if c.Encoding != "proto" && c.Encoding != "json" && c.Encoding != "json_ietf" {
		return nil, fmt.Errorf("unsupported encoding %s", c.Encoding)
	}

	return &gnmi.SubscribeRequest{
		Request: &gnmi.SubscribeRequest_Subscribe{
			Subscribe: &gnmi.SubscriptionList{
				Prefix:       gnmiPath,
				Mode:         gnmi.SubscriptionList_STREAM,
				Encoding:     gnmi.Encoding(gnmi.Encoding_value[strings.ToUpper(c.Encoding)]),
				Subscription: subscriptions,
				UpdatesOnly:  c.UpdatesOnly,
			},
		},
	}, nil
}

// SubscribeGNMI and extract telemetry data
func (c *CiscoTelemetryGNMI) subscribeGNMI(ctx context.Context, client *grpc.ClientConn, request *gnmi.SubscribeRequest) error {
	subscribeClient, err := gnmi.NewGNMIClient(client).Subscribe(ctx)
	if err != nil {
		return fmt.Errorf("failed to setup subscription: %v", err)
	}

	if err = subscribeClient.Send(request); err != nil {
		return fmt.Errorf("failed to send subscription request: %v", err)
	}

	log.Printf("D! [inputs.cisco_telemetry_gnmi]: Connection to GNMI device %s established", c.Address)
	defer log.Printf("D! [inputs.cisco_telemetry_gnmi]: Connection to GNMI device %s closed", c.Address)
	for ctx.Err() == nil {
		var reply *gnmi.SubscribeResponse
		if reply, err = subscribeClient.Recv(); err != nil {
			if err != io.EOF && ctx.Err() == nil {
				return fmt.Errorf("aborted GNMI subscription: %v", err)
			}
			break
		}

		c.handleSubscribeResponse(reply)
	}
	return nil
}

// HandleSubscribeResponse message from GNMI and parse contained telemetry data
func (c *CiscoTelemetryGNMI) handleSubscribeResponse(reply *gnmi.SubscribeResponse) {
	// Check if response is a GNMI Update and if we have a prefix to derive the measurement name
	response, ok := reply.Response.(*gnmi.SubscribeResponse_Update)
	if !ok || response.Update.Prefix == nil {
		return
	}

	timestamp := time.Unix(0, response.Update.Timestamp)
	fields := make(map[string]interface{})
	tags := make(map[string]string)

	var builder bytes.Buffer
	builder.WriteRune('/')

	// Parse generic keys from prefix
	for _, elem := range response.Update.Prefix.Elem {
		builder.WriteString(elem.Name)
		builder.WriteRune('/')

		for key, val := range elem.Key {
			// Use short-form of key if possible
			if _, exists := tags[key]; exists {
				tags[builder.String()+key] = val
			} else {
				tags[key] = val
			}
		}
	}

	tags["source"], _, _ = net.SplitHostPort(c.Address)
	builder.Truncate(builder.Len() - 1)
	prefix := builder.String()

	// Parse individual Update message and create measurement
	for _, update := range response.Update.Update {
		c.handleTelemetryField(fields, update)
	}

	// Finally add measurements
	if len(response.Update.Prefix.Origin) > 0 {
		tags["path"] = prefix
		c.acc.AddFields(response.Update.Prefix.Origin, fields, tags, timestamp)
	} else {
		c.acc.AddFields(prefix, fields, tags, timestamp)
	}
}

// HandleTelemetryField and add it to a measurement
func (c *CiscoTelemetryGNMI) handleTelemetryField(fields map[string]interface{}, update *gnmi.Update) {
	var builder bytes.Buffer
	if len(update.Path.Origin) > 0 {
		builder.WriteString(update.Path.Origin)
		builder.WriteRune(':')
	}

	parts := update.Path.Elem

	// Compatibility with old GNMI
	if len(parts) == 0 {
		parts = make([]*gnmi.PathElem, len(update.Path.Element))
		for i, part := range update.Path.Element {
			parts[i] = &gnmi.PathElem{Name: part}
		}
	}

	for i, elem := range parts {
		builder.WriteString(elem.Name)

		var keys []string
		for key, val := range elem.Key {
			keys = append(keys, "["+key+"="+val+"]")
		}
		sort.Strings(keys)
		for _, key := range keys {
			builder.WriteString(key)
		}

		if i < len(parts)-1 {
			builder.WriteRune('/')
		}
	}

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

	if value != nil {
		fields[builder.String()] = value
	} else if jsondata != nil {
		if err := json.Unmarshal(jsondata, &value); err != nil {
			c.acc.AddError(fmt.Errorf("GNMI JSON data is invalid: %v", err))
			return
		}

		flattener := jsonparser.JSONFlattener{Fields: fields}
		flattener.FullFlattenJSON(builder.String(), value, true, true)
	}
}

//ParsePath from XPath-like string to GNMI path structure
func parsePath(origin string, path string, target string) (*gnmi.Path, error) {
	var err error
	gnmiPath := gnmi.Path{Origin: origin, Target: target}

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

// Stop listener and cleanup
func (c *CiscoTelemetryGNMI) Stop() {
	c.cancel()
	c.wg.Wait()
}

const sampleConfig = `
## Address and port of the GNMI GRPC server
address = "10.49.234.114:57777"

## define credentials
username = "cisco"
password = "cisco"

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


[[inputs.cisco_telemetry_gnmi.subscription]]
  ## Origin and path of the subscription
  ## See: https://github.com/openconfig/reference/blob/master/rpc/gnmi/gnmi-specification.md#222-paths
  ##
  ## origin usually refers to a (YANG) data model implemented by the device
  ## and path to a specific substructe inside it that should be subscribed to (similar to an XPath)
  ## YANG models can be found e.g. here: https://github.com/YangModels/yang/tree/master/vendor/cisco/xr
  origin = "openconfig-interfaces"
  path = "/interfaces/interface/state/counters"

  # Subscription mode (one of: "target_defined", "sample", "on_change") and interval
  subscription_mode = "sample"
  sample_interval = "10s"

  ## Suppress redundant transmissions when measured values are unchanged
  # suppress_redundant = false

  ## If suppression is enabled, send updates at least every X seconds anyway
  # heartbeat_interval = "60s"
`

// SampleConfig of plugin
func (c *CiscoTelemetryGNMI) SampleConfig() string {
	return sampleConfig
}

// Description of plugin
func (c *CiscoTelemetryGNMI) Description() string {
	return "Cisco GNMI telemetry input plugin based on GNMI telemetry data produced in IOS XR"
}

// Gather plugin measurements (unused)
func (c *CiscoTelemetryGNMI) Gather(_ telegraf.Accumulator) error {
	return nil
}

func init() {
	inputs.Add("cisco_telemetry_gnmi", func() telegraf.Input {
		return &CiscoTelemetryGNMI{
			Encoding: "proto",
			Redial:   internal.Duration{Duration: 10 * time.Second},
		}
	})
}
