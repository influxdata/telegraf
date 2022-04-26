package gnmi

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

	"github.com/google/gnxi/utils/xpath"
	gnmiLib "github.com/openconfig/gnmi/proto/gnmi"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/metric"
	internaltls "github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
	jsonparser "github.com/influxdata/telegraf/plugins/parsers/json"
)

// gNMI plugin instance
type GNMI struct {
	Addresses     []string          `toml:"addresses"`
	Subscriptions []Subscription    `toml:"subscription"`
	Aliases       map[string]string `toml:"aliases"`

	// Optional subscription configuration
	Encoding    string
	Origin      string
	Prefix      string
	Target      string
	UpdatesOnly bool `toml:"updates_only"`

	// gNMI target credentials
	Username string
	Password string

	// Redial
	Redial config.Duration

	// GRPC TLS settings
	EnableTLS bool `toml:"enable_tls"`
	internaltls.ClientConfig

	// Internal state
	internalAliases map[string]string
	acc             telegraf.Accumulator
	cancel          context.CancelFunc
	wg              sync.WaitGroup
	// Lookup/device+name/key/value
	lookup      map[string]map[string]map[string]interface{}
	lookupMutex sync.Mutex

	Log telegraf.Logger
}

// Subscription for a gNMI client
type Subscription struct {
	Name   string
	Origin string
	Path   string

	// Subscription mode and interval
	SubscriptionMode string          `toml:"subscription_mode"`
	SampleInterval   config.Duration `toml:"sample_interval"`

	// Duplicate suppression
	SuppressRedundant bool            `toml:"suppress_redundant"`
	HeartbeatInterval config.Duration `toml:"heartbeat_interval"`

	// Mark this subscription as a tag-only lookup source, not emitting any metric
	TagOnly bool `toml:"tag_only"`
}

// Start the http listener service
func (c *GNMI) Start(acc telegraf.Accumulator) error {
	var err error
	var ctx context.Context
	var tlscfg *tls.Config
	var request *gnmiLib.SubscribeRequest
	c.acc = acc
	ctx, c.cancel = context.WithCancel(context.Background())
	c.lookupMutex.Lock()
	c.lookup = make(map[string]map[string]map[string]interface{})
	c.lookupMutex.Unlock()

	// Validate configuration
	if request, err = c.newSubscribeRequest(); err != nil {
		return err
	} else if time.Duration(c.Redial).Nanoseconds() <= 0 {
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
	c.internalAliases = make(map[string]string, len(c.Subscriptions)+len(c.Aliases))
	for _, subscription := range c.Subscriptions {
		var gnmiLongPath, gnmiShortPath *gnmiLib.Path

		// Build the subscription path without keys
		if gnmiLongPath, err = parsePath(subscription.Origin, subscription.Path, ""); err != nil {
			return err
		}
		if gnmiShortPath, err = parsePath("", subscription.Path, ""); err != nil {
			return err
		}

		longPath, _, err := c.handlePath(gnmiLongPath, nil, "")
		if err != nil {
			return fmt.Errorf("handling long-path failed: %v", err)
		}
		shortPath, _, err := c.handlePath(gnmiShortPath, nil, "")
		if err != nil {
			return fmt.Errorf("handling short-path failed: %v", err)
		}
		name := subscription.Name

		// If the user didn't provide a measurement name, use last path element
		if len(name) == 0 {
			name = path.Base(shortPath)
		}
		if len(name) > 0 {
			c.internalAliases[longPath] = name
			c.internalAliases[shortPath] = name
		}

		if subscription.TagOnly {
			// Create the top-level lookup for this tag
			c.lookupMutex.Lock()
			c.lookup[name] = make(map[string]map[string]interface{})
			c.lookupMutex.Unlock()
		}
	}
	for alias, encodingPath := range c.Aliases {
		c.internalAliases[encodingPath] = alias
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
				case <-time.After(time.Duration(c.Redial)):
				}
			}
		}(addr)
	}
	return nil
}

// Create a new gNMI SubscribeRequest
func (c *GNMI) newSubscribeRequest() (*gnmiLib.SubscribeRequest, error) {
	// Create subscription objects
	subscriptions := make([]*gnmiLib.Subscription, len(c.Subscriptions))
	for i, subscription := range c.Subscriptions {
		gnmiPath, err := parsePath(subscription.Origin, subscription.Path, "")
		if err != nil {
			return nil, err
		}
		mode, ok := gnmiLib.SubscriptionMode_value[strings.ToUpper(subscription.SubscriptionMode)]
		if !ok {
			return nil, fmt.Errorf("invalid subscription mode %s", subscription.SubscriptionMode)
		}
		subscriptions[i] = &gnmiLib.Subscription{
			Path:              gnmiPath,
			Mode:              gnmiLib.SubscriptionMode(mode),
			SampleInterval:    uint64(time.Duration(subscription.SampleInterval).Nanoseconds()),
			SuppressRedundant: subscription.SuppressRedundant,
			HeartbeatInterval: uint64(time.Duration(subscription.HeartbeatInterval).Nanoseconds()),
		}
	}

	// Construct subscribe request
	gnmiPath, err := parsePath(c.Origin, c.Prefix, c.Target)
	if err != nil {
		return nil, err
	}

	if c.Encoding != "proto" && c.Encoding != "json" && c.Encoding != "json_ietf" && c.Encoding != "bytes" {
		return nil, fmt.Errorf("unsupported encoding %s", c.Encoding)
	}

	return &gnmiLib.SubscribeRequest{
		Request: &gnmiLib.SubscribeRequest_Subscribe{
			Subscribe: &gnmiLib.SubscriptionList{
				Prefix:       gnmiPath,
				Mode:         gnmiLib.SubscriptionList_STREAM,
				Encoding:     gnmiLib.Encoding(gnmiLib.Encoding_value[strings.ToUpper(c.Encoding)]),
				Subscription: subscriptions,
				UpdatesOnly:  c.UpdatesOnly,
			},
		},
	}, nil
}

// SubscribeGNMI and extract telemetry data
func (c *GNMI) subscribeGNMI(ctx context.Context, address string, tlscfg *tls.Config, request *gnmiLib.SubscribeRequest) error {
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

	subscribeClient, err := gnmiLib.NewGNMIClient(client).Subscribe(ctx)
	if err != nil {
		return fmt.Errorf("failed to setup subscription: %v", err)
	}

	if err = subscribeClient.Send(request); err != nil {
		// If io.EOF is returned, the stream may have ended and stream status
		// can be determined by calling Recv.
		if err != io.EOF {
			return fmt.Errorf("failed to send subscription request: %v", err)
		}
	}

	c.Log.Debugf("Connection to gNMI device %s established", address)
	defer c.Log.Debugf("Connection to gNMI device %s closed", address)
	for ctx.Err() == nil {
		var reply *gnmiLib.SubscribeResponse
		if reply, err = subscribeClient.Recv(); err != nil {
			if err != io.EOF && ctx.Err() == nil {
				return fmt.Errorf("aborted gNMI subscription: %v", err)
			}
			break
		}

		c.handleSubscribeResponse(address, reply)
	}
	return nil
}

func (c *GNMI) handleSubscribeResponse(address string, reply *gnmiLib.SubscribeResponse) {
	switch response := reply.Response.(type) {
	case *gnmiLib.SubscribeResponse_Update:
		c.handleSubscribeResponseUpdate(address, response)
	case *gnmiLib.SubscribeResponse_Error:
		c.Log.Errorf("Subscribe error (%d), %q", response.Error.Code, response.Error.Message)
	}
}

// Handle SubscribeResponse_Update message from gNMI and parse contained telemetry data
func (c *GNMI) handleSubscribeResponseUpdate(address string, response *gnmiLib.SubscribeResponse_Update) {
	var prefix, prefixAliasPath string
	grouper := metric.NewSeriesGrouper()
	timestamp := time.Unix(0, response.Update.Timestamp)
	prefixTags := make(map[string]string)

	if response.Update.Prefix != nil {
		var err error
		if prefix, prefixAliasPath, err = c.handlePath(response.Update.Prefix, prefixTags, ""); err != nil {
			c.Log.Errorf("handling path %q failed: %v", response.Update.Prefix, err)
		}
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
			if alias, ok := c.internalAliases[aliasPath]; ok {
				name = alias
			} else {
				c.Log.Debugf("No measurement alias for gNMI path: %s", name)
			}
		}

		// Update tag lookups and discard rest of update
		subscriptionKey := tags["source"] + "/" + tags["name"]
		c.lookupMutex.Lock()
		if _, ok := c.lookup[name]; ok {
			// We are subscribed to this, so add the fields to the lookup-table
			if _, ok := c.lookup[name][subscriptionKey]; !ok {
				c.lookup[name][subscriptionKey] = make(map[string]interface{})
			}
			for k, v := range fields {
				c.lookup[name][subscriptionKey][path.Base(k)] = v
			}
			c.lookupMutex.Unlock()
			// Do not process the data further as we only subscribed here for the lookup table
			continue
		}

		// Apply lookups if present
		for subscriptionName, values := range c.lookup {
			if annotations, ok := values[subscriptionKey]; ok {
				for k, v := range annotations {
					tags[subscriptionName+"/"+k] = fmt.Sprint(v)
				}
			}
		}
		c.lookupMutex.Unlock()

		// Group metrics
		for k, v := range fields {
			key := k
			if len(aliasPath) < len(key) && len(aliasPath) != 0 {
				// This may not be an exact prefix, due to naming style
				// conversion on the key.
				key = key[len(aliasPath)+1:]
			} else if len(aliasPath) >= len(key) {
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

			if err := grouper.Add(name, tags, timestamp, key, v); err != nil {
				c.Log.Errorf("cannot add to grouper: %v", err)
			}
		}

		lastAliasPath = aliasPath
	}

	// Add grouped measurements
	for _, metricToAdd := range grouper.Metrics() {
		c.acc.AddMetric(metricToAdd)
	}
}

// HandleTelemetryField and add it to a measurement
func (c *GNMI) handleTelemetryField(update *gnmiLib.Update, tags map[string]string, prefix string) (string, map[string]interface{}) {
	gpath, aliasPath, err := c.handlePath(update.Path, tags, prefix)
	if err != nil {
		c.Log.Errorf("handling path %q failed: %v", update.Path, err)
	}

	var value interface{}
	var jsondata []byte

	// Make sure a value is actually set
	if update.Val == nil || update.Val.Value == nil {
		c.Log.Infof("Discarded empty or legacy type value with path: %q", gpath)
		return aliasPath, nil
	}

	switch val := update.Val.Value.(type) {
	case *gnmiLib.TypedValue_AsciiVal:
		value = val.AsciiVal
	case *gnmiLib.TypedValue_BoolVal:
		value = val.BoolVal
	case *gnmiLib.TypedValue_BytesVal:
		value = val.BytesVal
	case *gnmiLib.TypedValue_DecimalVal:
		value = float64(val.DecimalVal.Digits) / math.Pow(10, float64(val.DecimalVal.Precision))
	case *gnmiLib.TypedValue_FloatVal:
		value = val.FloatVal
	case *gnmiLib.TypedValue_IntVal:
		value = val.IntVal
	case *gnmiLib.TypedValue_StringVal:
		value = val.StringVal
	case *gnmiLib.TypedValue_UintVal:
		value = val.UintVal
	case *gnmiLib.TypedValue_JsonIetfVal:
		jsondata = val.JsonIetfVal
	case *gnmiLib.TypedValue_JsonVal:
		jsondata = val.JsonVal
	}

	name := strings.Replace(gpath, "-", "_", -1)
	fields := make(map[string]interface{})
	if value != nil {
		fields[name] = value
	} else if jsondata != nil {
		if err := json.Unmarshal(jsondata, &value); err != nil {
			c.acc.AddError(fmt.Errorf("failed to parse JSON value: %v", err))
		} else {
			flattener := jsonparser.JSONFlattener{Fields: fields}
			if err := flattener.FullFlattenJSON(name, value, true, true); err != nil {
				c.acc.AddError(fmt.Errorf("failed to flatten JSON: %v", err))
			}
		}
	}
	return aliasPath, fields
}

// Parse path to path-buffer and tag-field
func (c *GNMI) handlePath(gnmiPath *gnmiLib.Path, tags map[string]string, prefix string) (pathBuffer string, aliasPath string, err error) {
	builder := bytes.NewBufferString(prefix)

	// Prefix with origin
	if len(gnmiPath.Origin) > 0 {
		if _, err := builder.WriteString(gnmiPath.Origin); err != nil {
			return "", "", err
		}
		if _, err := builder.WriteRune(':'); err != nil {
			return "", "", err
		}
	}

	// Parse generic keys from prefix
	for _, elem := range gnmiPath.Elem {
		if len(elem.Name) > 0 {
			if _, err := builder.WriteRune('/'); err != nil {
				return "", "", err
			}
			if _, err := builder.WriteString(elem.Name); err != nil {
				return "", "", err
			}
		}
		name := builder.String()

		if _, exists := c.internalAliases[name]; exists {
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

	return builder.String(), aliasPath, nil
}

//ParsePath from XPath-like string to gNMI path structure
func parsePath(origin string, pathToParse string, target string) (*gnmiLib.Path, error) {
	gnmiPath, err := xpath.ToGNMIPath(pathToParse)
	if err != nil {
		return nil, err
	}
	gnmiPath.Origin = origin
	gnmiPath.Target = target
	return gnmiPath, err
}

// Stop listener and cleanup
func (c *GNMI) Stop() {
	c.cancel()
	c.wg.Wait()
}

// Gather plugin measurements (unused)
func (c *GNMI) Gather(_ telegraf.Accumulator) error {
	return nil
}

func New() telegraf.Input {
	return &GNMI{
		Encoding: "proto",
		Redial:   config.Duration(10 * time.Second),
	}
}

func init() {
	inputs.Add("gnmi", New)
	// Backwards compatible alias:
	inputs.Add("cisco_telemetry_gnmi", New)
}
