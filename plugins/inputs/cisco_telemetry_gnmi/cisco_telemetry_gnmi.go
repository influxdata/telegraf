package cisco_telemetry_gnmi

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net"
	"path"
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

// CiscoTelemetryGNMI plugin instance
type CiscoTelemetryGNMI struct {
	Addresses     []string          `toml:"addresses"`
	Subscriptions []Subscription    `toml:"subscription"`
	Aliases       map[string]string `toml:"aliases"`

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
	aliases map[string]string
	acc     telegraf.Accumulator
	cancel  context.CancelFunc
	wg      sync.WaitGroup

	Log telegraf.Logger
}

// Subscription for a GNMI client
type Subscription struct {
	Name   string
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
	var tlscfg *tls.Config
	var request *gnmi.SubscribeRequest
	c.acc = acc
	ctx, c.cancel = context.WithCancel(context.Background())

	// Validate configuration
	if request, err = c.newSubscribeRequest(); err != nil {
		return err
	} else if c.Redial.Duration.Nanoseconds() <= 0 {
		return fmt.Errorf("redial duration must be positive")
	}

	// Parse TLS config
	if c.EnableTLS {
		if tlscfg, err = c.ClientConfig.TLSConfig(); err != nil {
			return err
		}
	}

	if len(c.Username) > 0 {
		ctx = metadata.AppendToOutgoingContext(ctx, "username", c.Username, "password", c.Password)
	}

	// Invert explicit alias list and prefill subscription names
	c.aliases = make(map[string]string, len(c.Subscriptions)+len(c.Aliases))
	for _, subscription := range c.Subscriptions {
		var gnmiLongPath, gnmiShortPath *gnmi.Path

		// Build the subscription path without keys
		if gnmiLongPath, err = parsePath(subscription.Origin, subscription.Path, ""); err != nil {
			return err
		}
		if gnmiShortPath, err = parsePath("", subscription.Path, ""); err != nil {
			return err
		}

		longPath, _ := c.handlePath(gnmiLongPath, nil, "")
		shortPath, _ := c.handlePath(gnmiShortPath, nil, "")
		name := subscription.Name

		// If the user didn't provide a measurement name, use last path element
		if len(name) == 0 {
			name = path.Base(shortPath)
		}
		if len(name) > 0 {
			c.aliases[longPath] = name
			c.aliases[shortPath] = name
		}
	}
	for alias, path := range c.Aliases {
		c.aliases[path] = alias
	}

	// Create a goroutine for each device, dial and subscribe
	c.wg.Add(len(c.Addresses))
	for _, addr := range c.Addresses {
		go func(address string) {
			defer c.wg.Done()
			for ctx.Err() == nil {
				if err := c.subscribeGNMI(ctx, address, tlscfg, request); err != nil && ctx.Err() == nil {
					acc.AddError(err)
				}

				select {
				case <-ctx.Done():
				case <-time.After(c.Redial.Duration):
				}
			}
		}(addr)
	}
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

	if c.Encoding != "proto" && c.Encoding != "json" && c.Encoding != "json_ietf" {
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
func (c *CiscoTelemetryGNMI) subscribeGNMI(ctx context.Context, address string, tlscfg *tls.Config, request *gnmi.SubscribeRequest) error {
	var opt grpc.DialOption
	if tlscfg != nil {
		opt = grpc.WithTransportCredentials(credentials.NewTLS(tlscfg))
	} else {
		opt = grpc.WithInsecure()
	}

	client, err := grpc.DialContext(ctx, address, opt)
	if err != nil {
		return fmt.Errorf("failed to dial: %v", err)
	}
	defer client.Close()

	subscribeClient, err := gnmi.NewGNMIClient(client).Subscribe(ctx)
	if err != nil {
		return fmt.Errorf("failed to setup subscription: %v", err)
	}

	if err = subscribeClient.Send(request); err != nil {
		return fmt.Errorf("failed to send subscription request: %v", err)
	}

	c.Log.Debugf("Connection to GNMI device %s established", address)
	defer c.Log.Debugf("Connection to GNMI device %s closed", address)
	for ctx.Err() == nil {
		var reply *gnmi.SubscribeResponse
		if reply, err = subscribeClient.Recv(); err != nil {
			if err != io.EOF && ctx.Err() == nil {
				return fmt.Errorf("aborted GNMI subscription: %v", err)
			}
			break
		}

		c.handleSubscribeResponse(address, reply)
	}
	return nil
}

// HandleSubscribeResponse message from GNMI and parse contained telemetry data
func (c *CiscoTelemetryGNMI) handleSubscribeResponse(address string, reply *gnmi.SubscribeResponse) {
	// Check if response is a GNMI Update and if we have a prefix to derive the measurement name
	response, ok := reply.Response.(*gnmi.SubscribeResponse_Update)
	if !ok {
		return
	}

	var prefix, prefixAliasPath string
	grouper := metric.NewSeriesGrouper()
	timestamp := time.Unix(0, response.Update.Timestamp)
	prefixTags := make(map[string]string)

	if response.Update.Prefix != nil {
		prefix, prefixAliasPath = c.handlePath(response.Update.Prefix, prefixTags, "")
	}
	prefixTags["source"], _, _ = net.SplitHostPort(address)
	prefixTags["path"] = prefix

	// Parse individual Update message and create measurements
	var name, lastAliasPath string
	for _, update := range response.Update.Update {
		// Prepare tags from prefix
		tags := make(map[string]string, len(prefixTags))
		for key, val := range prefixTags {
			tags[key] = val
		}
		aliasPath, fields := c.handleTelemetryField(update, tags, prefix)

		// Inherent valid alias from prefix parsing
		if len(prefixAliasPath) > 0 && len(aliasPath) == 0 {
			aliasPath = prefixAliasPath
		}

		// Lookup alias if alias-path has changed
		if aliasPath != lastAliasPath {
			name = prefix
			if alias, ok := c.aliases[aliasPath]; ok {
				name = alias
			} else {
				c.Log.Debugf("No measurement alias for GNMI path: %s", name)
			}
		}

		// Group metrics
		for k, v := range fields {
			key := k
			if len(aliasPath) < len(key) {
				// This may not be an exact prefix, due to naming style
				// conversion on the key.
				key = key[len(aliasPath)+1:]
			} else {
				// Otherwise use the last path element as the field key.
				key = path.Base(key)

				// If there are no elements skip the item; this would be an
				// invalid message.
				key = strings.TrimLeft(key, "/.")
				if key == "" {
					c.Log.Errorf("invalid empty path: %q", k)
					continue
				}
			}

			grouper.Add(name, tags, timestamp, key, v)
		}

		lastAliasPath = aliasPath
	}

	// Add grouped measurements
	for _, metric := range grouper.Metrics() {
		c.acc.AddMetric(metric)
	}
}

// HandleTelemetryField and add it to a measurement
func (c *CiscoTelemetryGNMI) handleTelemetryField(update *gnmi.Update, tags map[string]string, prefix string) (string, map[string]interface{}) {
	path, aliasPath := c.handlePath(update.Path, tags, prefix)

	var value interface{}
	var jsondata []byte

	// Make sure a value is actually set
	if update.Val == nil || update.Val.Value == nil {
		c.Log.Infof("Discarded empty or legacy type value with path: %q", path)
		return aliasPath, nil
	}

	switch val := update.Val.Value.(type) {
	case *gnmi.TypedValue_AsciiVal:
		value = val.AsciiVal
	case *gnmi.TypedValue_BoolVal:
		value = val.BoolVal
	case *gnmi.TypedValue_BytesVal:
		value = val.BytesVal
	case *gnmi.TypedValue_DecimalVal:
		value = float64(val.DecimalVal.Digits) / math.Pow(10, float64(val.DecimalVal.Precision))
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
			c.acc.AddError(fmt.Errorf("failed to parse JSON value: %v", err))
		} else {
			flattener := jsonparser.JSONFlattener{Fields: fields}
			flattener.FullFlattenJSON(name, value, true, true)
		}
	}
	return aliasPath, fields
}

// Parse path to path-buffer and tag-field
func (c *CiscoTelemetryGNMI) handlePath(path *gnmi.Path, tags map[string]string, prefix string) (string, string) {
	var aliasPath string
	builder := bytes.NewBufferString(prefix)

	// Prefix with origin
	if len(path.Origin) > 0 {
		builder.WriteString(path.Origin)
		builder.WriteRune(':')
	}

	// Parse generic keys from prefix
	for _, elem := range path.Elem {
		if len(elem.Name) > 0 {
			builder.WriteRune('/')
			builder.WriteString(elem.Name)
		}
		name := builder.String()

		if _, exists := c.aliases[name]; exists {
			aliasPath = name
		}

		if tags != nil {
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
	}

	return builder.String(), aliasPath
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

// Stop listener and cleanup
func (c *CiscoTelemetryGNMI) Stop() {
	c.cancel()
	c.wg.Wait()
}

const sampleConfig = `
 ## Address and port of the GNMI GRPC server
 addresses = ["10.49.234.114:57777"]

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

 ## Define additional aliases to map telemetry encoding paths to simple measurement names
 #[inputs.cisco_telemetry_gnmi.aliases]
 #  ifcounters = "openconfig:/interfaces/interface/state/counters"

 [[inputs.cisco_telemetry_gnmi.subscription]]
  ## Name of the measurement that will be emitted
  name = "ifcounters"

  ## Origin and path of the subscription
  ## See: https://github.com/openconfig/reference/blob/master/rpc/gnmi/gnmi-specification.md#222-paths
  ##
  ## origin usually refers to a (YANG) data model implemented by the device
  ## and path to a specific substructure inside it that should be subscribed to (similar to an XPath)
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
