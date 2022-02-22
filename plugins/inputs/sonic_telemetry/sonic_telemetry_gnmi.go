package sonic_telemetry_gnmi

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
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
	"encoding/binary"
	"encoding/base64"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/metric"
	internaltls "github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
	jsonparser "github.com/influxdata/telegraf/plugins/parsers/json"
	"github.com/openconfig/gnmi/proto/gnmi"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

// SONiCTelemetryGNMI plugin instance
type SONiCTelemetryGNMI struct {
	Addresses     []string          `toml:"addresses"`
	Subscriptions []Subscription    `toml:"subscription"`
	Aliases       map[string]string `toml:"aliases"`

	// Optional subscription configuration
	Encoding    string
	Origin      string
	Prefix      string
	Target      string
	UpdatesOnly bool `toml:"updates_only"`

	// SONiC Default credentials
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
	// Lookup/device/name/key/value
	lookup map[string]map[string]map[string]map[string]interface{}

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

	// Tag-only identification
	TagOnly bool `toml:"tag_only"`
}

func Float32ValfromBytes(bytes []byte) float32 {
        bits := binary.BigEndian.Uint32(bytes)
        float32Val := math.Float32frombits(bits)
        return float32Val
}

func IsBase64(s string) (float32, bool)  {
    if len(s) != 8  && len(s) != 12 {
        return 0, false
    }
    retStr, err := base64.StdEncoding.DecodeString(s)
    if err != nil {
        fmt.Println("Error during base64 decode. Err: %s", err.Error())
        return 0, err == nil
    }
    return Float32ValfromBytes(retStr), err == nil
}

// Start the http listener service
func (c *SONiCTelemetryGNMI) Start(acc telegraf.Accumulator) error {
	var err error
	var ctx context.Context
	var tlscfg *tls.Config
	var request *gnmi.SubscribeRequest
	c.acc = acc
	ctx, c.cancel = context.WithCancel(context.Background())
	c.lookup = make(map[string]map[string]map[string]map[string]interface{})

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

		if subscription.TagOnly {
			// Create the top-level lookup for this tag
			c.lookup[name] = make(map[string]map[string]map[string]interface{})
		}
	}
	for alias, path := range c.Aliases {
		c.aliases[path] = alias
	}

	// Create a goroutine for each device, dial and subscribe
	c.wg.Add(len(c.Addresses))
	for _, addr := range c.Addresses {
		// Update the lookup table with this address
		for lu := range c.lookup {
			hostname, _, _ := net.SplitHostPort(addr)
			c.lookup[lu][hostname] = make(map[string]map[string]interface{})
		}
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
func (c *SONiCTelemetryGNMI) newSubscribeRequest() (*gnmi.SubscribeRequest, error) {
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
func (c *SONiCTelemetryGNMI) subscribeGNMI(ctx context.Context, address string, tlscfg *tls.Config, request *gnmi.SubscribeRequest) error {
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
func (c *SONiCTelemetryGNMI) handleSubscribeResponse(address string, reply *gnmi.SubscribeResponse) {
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

		// Update tag lookups and discard rest of update
		if lu, ok := c.lookup[name]; ok {
			if err := updateLookups(lu[tags["source"]], tags, fields); err != nil {
				c.Log.Debugf("Error updating lookups")
			}
			continue
		}

		// Apply lookups if present
		for k, v := range c.lookup {
			if t, ok := v[tags["source"]][tags["name"]]; ok {
				for name, val := range t {
					tagName := strings.Replace(fmt.Sprintf("%s_%s", k, name), ":", "", -1)
					tags[tagName] = val.(string)
				}
			}
		}

		// Group metrics
		for k, v := range fields {
			key := k
			if len(aliasPath) < len(key) {
				// This may not be an exact prefix, due to naming style
				// conversion on the key.
				key = key[len(aliasPath)+1:]
				parts := strings.Split(key, ":")
				key = parts[len(parts)-1]

				key = strings.Replace(key, "-", "_", -1)
			} else {
				// Otherwise use the last path element as the field key.
				key = path.Base(key)
				parts := strings.Split(key, ":")
				key = parts[len(parts)-1]

				key = strings.Replace(key, "-", "_", -1)
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

func updateLookups(lu map[string]map[string]interface{}, tags map[string]string, fields map[string]interface{}) error {
	name, ok := lu[tags["name"]]
	if !ok {
		name = make(map[string]interface{})
		lu[tags["name"]] = name
	}
	for k, v := range fields {
		shortName := path.Base(k)
		name[shortName] = v
	}
	return nil
}

// HandleTelemetryField and add it to a measurement
func (c *SONiCTelemetryGNMI) handleTelemetryField(update *gnmi.Update, tags map[string]string, prefix string) (string, map[string]interface{}) {
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

			for k1, v1 := range fields {
                                if reflect.TypeOf(v1).Kind() != reflect.String {
                                        continue
                                }
                                if (strings.Contains(k1, "transceiver_dom")) {
                                    if fVal, retVal := IsBase64(v1.(string)); retVal {
                                        fields[k1] = fVal
                                    }
                                } else {
                                        if i64, err := strconv.ParseInt(v1.(string), 10, 64); err == nil {
                                                fields[k1] = i64
                                        } else if f64, err := strconv.ParseFloat(v1.(string), 64); err == nil {
                                                fields[k1] = f64
                                        }
                                }
			}
		}
	}
	return aliasPath, fields
}

// Parse path to path-buffer and tag-field
func (c *SONiCTelemetryGNMI) handlePath(path *gnmi.Path, tags map[string]string, prefix string) (string, string) {
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
func (c *SONiCTelemetryGNMI) Stop() {
	c.cancel()
	c.wg.Wait()
}

const sampleConfig = `
 ## Address and port of the GNMI GRPC server
 addresses = ["localhost:8080"]

 ## define credentials
 username = "admin"
 password = "YourPaSsWoRd"

 ## GNMI encoding requested (one of: "proto", "json", "json_ietf")
 encoding = "json_ietf"

 ## redial in case of failures after
 redial = "10s"

 ## enable client-side TLS and define CA to authenticate the device
 enable_tls = true
 # tls_ca = "/etc/telegraf/ca.pem"
 insecure_skip_verify = true

 ## define client-side TLS certificate & key to authenticate to the device
 # tls_cert = "/etc/telegraf/cert.pem"
 # tls_key = "/etc/telegraf/key.pem"

 ## GNMI subscription prefix (optional, can usually be left empty)
 ## See: https://github.com/openconfig/reference/blob/master/rpc/gnmi/gnmi-specification.md#222-paths
 # origin = ""
 # prefix = ""
 target = "OC-YANG"

 ## Define additional aliases to map telemetry encoding paths to simple measurement names
 #[inputs.sonic_telemetry_gnmi.aliases]
 #  ifcounters = "openconfig:/interfaces/interface/state/counters"

 [[inputs.sonic_telemetry_gnmi.subscription]]
  ## Name of the measurement that will be emitted
  name = "ifcounters"

  ## Origin and path of the subscription
  ## See: https://github.com/openconfig/reference/blob/master/rpc/gnmi/gnmi-specification.md#222-paths
  ##
  ## origin usually refers to a (YANG) data model implemented by the device
  ## and path to a specific substructe inside it that should be subscribed to (similar to an XPath)
  ## YANG models can be found e.g. here: https://sonic_mgmt_ip_address/ui
  origin = ""
  path = "/openconfig-interfaces:interfaces/interface/[name=Ethernet0]state/counters"

  # Subscription mode (one of: "target_defined", "sample", "on_change") and interval
  subscription_mode = "target_defined"
  sample_interval = "20s"

  ## Suppress redundant transmissions when measured values are unchanged
  # suppress_redundant = false

  ## If suppression is enabled, send updates at least every X seconds anyway
  # heartbeat_interval = "60s"

  [[inputs.sonic_telemetry_gnmi.subscription]]
   name = "descr"
   origin = "openconfig-interfaces"
   path = "/interfaces/interface/state/description"
   subscription_mode = "on_change"
   ## If tag_only is set, the subscription in question will be utilized to maintain a map of
   ## tags to apply to other measurements emitted by the plugin, by matching path keys
   ## All fields from the tag-only subscription will be applied as tags to other readings,
   ## in the format <name>_<fieldBase>.
   tag_only = true
`

// SampleConfig of plugin
func (c *SONiCTelemetryGNMI) SampleConfig() string {
	return sampleConfig
}

// Description of plugin
func (c *SONiCTelemetryGNMI) Description() string {
	return "SONiC GNMI telemetry input plugin based on GNMI telemetry data"
}

// Gather plugin measurements (unused)
func (c *SONiCTelemetryGNMI) Gather(_ telegraf.Accumulator) error {
	return nil
}

func init() {
	inputs.Add("sonic_telemetry_gnmi", func() telegraf.Input {
		return &SONiCTelemetryGNMI{
			Encoding: "proto",
			Redial:   internal.Duration{Duration: 10 * time.Second},
		}
	})
	inputs.Add("sonic_frr", func() telegraf.Input {
		return &SonicFRR{}
	})
}
