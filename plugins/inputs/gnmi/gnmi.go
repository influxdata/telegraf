//go:generate ../../../tools/readme_config_includer/generator
package gnmi

import (
	"bytes"
	"context"
	"crypto/tls"
	_ "embed"
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
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/metric"
	internaltls "github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
	jsonparser "github.com/influxdata/telegraf/plugins/parsers/json"
)

//go:embed sample.conf
var sampleConfig string

// gNMI plugin instance
type GNMI struct {
	Addresses        []string          `toml:"addresses"`
	Subscriptions    []Subscription    `toml:"subscription"`
	TagSubscriptions []TagSubscription `toml:"tag_subscription"`
	Aliases          map[string]string `toml:"aliases"`

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
	legacyTags      bool

	Log telegraf.Logger
}

type Worker struct {
	address  string
	tagStore *tagNode
}

type tagNode struct {
	elem     *gnmiLib.PathElem
	tagName  string
	value    *gnmiLib.TypedValue
	tagStore map[string][]*tagNode
}

type tagResults struct {
	names  []string
	values []*gnmiLib.TypedValue
}

// Subscription for a gNMI client
type Subscription struct {
	Name   string
	Origin string
	Path   string

	fullPath *gnmiLib.Path

	// Subscription mode and interval
	SubscriptionMode string          `toml:"subscription_mode"`
	SampleInterval   config.Duration `toml:"sample_interval"`

	// Duplicate suppression
	SuppressRedundant bool            `toml:"suppress_redundant"`
	HeartbeatInterval config.Duration `toml:"heartbeat_interval"`

	// Mark this subscription as a tag-only lookup source, not emitting any metric
	TagOnly bool `toml:"tag_only"`
}

// Tag Subscription for a gNMI client
type TagSubscription struct {
	Subscription
	Elements []string
}

func (*GNMI) SampleConfig() string {
	return sampleConfig
}

// Start the http listener service
func (c *GNMI) Start(acc telegraf.Accumulator) error {
	var err error
	var ctx context.Context
	var tlscfg *tls.Config
	var request *gnmiLib.SubscribeRequest
	c.acc = acc
	ctx, c.cancel = context.WithCancel(context.Background())

	for i := len(c.Subscriptions) - 1; i >= 0; i-- {
		subscription := c.Subscriptions[i]
		// Support legacy TagOnly subscriptions
		if subscription.TagOnly {
			tagSub := convertTagOnlySubscription(subscription)
			c.TagSubscriptions = append(c.TagSubscriptions, tagSub)
			// Remove from the original subscriptions list
			c.Subscriptions = append(c.Subscriptions[:i], c.Subscriptions[i+1:]...)
			c.legacyTags = true
			continue
		}
		if err = subscription.buildFullPath(c); err != nil {
			return err
		}
	}
	for idx := range c.TagSubscriptions {
		if err = c.TagSubscriptions[idx].buildFullPath(c); err != nil {
			return err
		}
		if c.TagSubscriptions[idx].TagOnly != c.TagSubscriptions[0].TagOnly {
			return fmt.Errorf("do not mix legacy tag_only subscriptions and tag subscriptions")
		}
		if len(c.TagSubscriptions[idx].Elements) == 0 {
			return fmt.Errorf("tag_subscription must have at least one element")
		}
	}

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
	c.internalAliases = make(map[string]string, len(c.Subscriptions)+len(c.Aliases)+len(c.TagSubscriptions))

	for _, s := range c.Subscriptions {
		if err := s.buildAlias(c.internalAliases); err != nil {
			return err
		}
	}
	for _, s := range c.TagSubscriptions {
		if err := s.buildAlias(c.internalAliases); err != nil {
			return err
		}
	}

	for alias, encodingPath := range c.Aliases {
		c.internalAliases[encodingPath] = alias
	}

	// Create a goroutine for each device, dial and subscribe
	c.wg.Add(len(c.Addresses))
	for _, addr := range c.Addresses {
		worker := Worker{address: addr}
		worker.tagStore = &tagNode{}
		go func(worker Worker) {
			defer c.wg.Done()
			for ctx.Err() == nil {
				if err := c.subscribeGNMI(ctx, &worker, tlscfg, request); err != nil && ctx.Err() == nil {
					acc.AddError(err)
				}

				select {
				case <-ctx.Done():
				case <-time.After(time.Duration(c.Redial)):
				}
			}
		}(worker)
	}
	return nil
}

func (s *Subscription) buildSubscription() (*gnmiLib.Subscription, error) {
	gnmiPath, err := parsePath(s.Origin, s.Path, "")
	if err != nil {
		return nil, err
	}
	mode, ok := gnmiLib.SubscriptionMode_value[strings.ToUpper(s.SubscriptionMode)]
	if !ok {
		return nil, fmt.Errorf("invalid subscription mode %s", s.SubscriptionMode)
	}
	return &gnmiLib.Subscription{
		Path:              gnmiPath,
		Mode:              gnmiLib.SubscriptionMode(mode),
		HeartbeatInterval: uint64(time.Duration(s.HeartbeatInterval).Nanoseconds()),
		SampleInterval:    uint64(time.Duration(s.SampleInterval).Nanoseconds()),
		SuppressRedundant: s.SuppressRedundant,
	}, nil
}

// Create a new gNMI SubscribeRequest
func (c *GNMI) newSubscribeRequest() (*gnmiLib.SubscribeRequest, error) {
	// Create subscription objects
	var err error
	subscriptions := make([]*gnmiLib.Subscription, len(c.Subscriptions)+len(c.TagSubscriptions))
	for i, subscription := range c.TagSubscriptions {
		if subscriptions[i], err = subscription.buildSubscription(); err != nil {
			return nil, err
		}
	}
	for i, subscription := range c.Subscriptions {
		if subscriptions[i+len(c.TagSubscriptions)], err = subscription.buildSubscription(); err != nil {
			return nil, err
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
func (c *GNMI) subscribeGNMI(ctx context.Context, worker *Worker, tlscfg *tls.Config, request *gnmiLib.SubscribeRequest) error {
	var creds credentials.TransportCredentials
	if tlscfg != nil {
		creds = credentials.NewTLS(tlscfg)
	} else {
		creds = insecure.NewCredentials()
	}
	opt := grpc.WithTransportCredentials(creds)

	client, err := grpc.DialContext(ctx, worker.address, opt)
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

	c.Log.Debugf("Connection to gNMI device %s established", worker.address)
	defer c.Log.Debugf("Connection to gNMI device %s closed", worker.address)
	for ctx.Err() == nil {
		var reply *gnmiLib.SubscribeResponse
		if reply, err = subscribeClient.Recv(); err != nil {
			if err != io.EOF && ctx.Err() == nil {
				return fmt.Errorf("aborted gNMI subscription: %v", err)
			}
			break
		}

		c.handleSubscribeResponse(worker, reply)
	}
	return nil
}

func (c *GNMI) handleSubscribeResponse(worker *Worker, reply *gnmiLib.SubscribeResponse) {
	if response, ok := reply.Response.(*gnmiLib.SubscribeResponse_Update); ok {
		c.handleSubscribeResponseUpdate(worker, response)
	}
}

// Handle SubscribeResponse_Update message from gNMI and parse contained telemetry data
func (c *GNMI) handleSubscribeResponseUpdate(worker *Worker, response *gnmiLib.SubscribeResponse_Update) {
	var prefix, prefixAliasPath string
	grouper := metric.NewSeriesGrouper()
	timestamp := time.Unix(0, response.Update.Timestamp)
	prefixTags := make(map[string]string)

	if response.Update.Prefix != nil {
		var err error
		if prefix, prefixAliasPath, err = handlePath(response.Update.Prefix, prefixTags, c.internalAliases, ""); err != nil {
			c.Log.Errorf("handling path %q failed: %v", response.Update.Prefix, err)
		}
	}
	prefixTags["source"], _, _ = net.SplitHostPort(worker.address)
	prefixTags["path"] = prefix

	// Process and remove tag-only updates from the response
	for i := len(response.Update.Update) - 1; i >= 0; i-- {
		update := response.Update.Update[i]
		fullPath := pathWithPrefix(response.Update.Prefix, update.Path)
		for _, tagSub := range c.TagSubscriptions {
			if equalPathNoKeys(fullPath, tagSub.fullPath) {
				worker.storeTags(update, tagSub)
				response.Update.Update = append(response.Update.Update[:i], response.Update.Update[i+1:]...)
			}
		}
	}

	// Parse individual Update message and create measurements
	var name, lastAliasPath string
	for _, update := range response.Update.Update {
		fullPath := pathWithPrefix(response.Update.Prefix, update.Path)

		// Prepare tags from prefix
		tags := make(map[string]string, len(prefixTags))
		for key, val := range prefixTags {
			tags[key] = val
		}
		aliasPath, fields := c.handleTelemetryField(update, tags, prefix)

		if tagOnlyTags := worker.checkTags(fullPath); tagOnlyTags != nil {
			for k, v := range tagOnlyTags {
				if alias, ok := c.internalAliases[k]; ok {
					tags[alias] = fmt.Sprint(v)
				} else {
					tags[k] = fmt.Sprint(v)
				}
			}
		}

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

			grouper.Add(name, tags, timestamp, key, v)
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
	gpath, aliasPath, err := handlePath(update.Path, tags, c.internalAliases, prefix)
	if err != nil {
		c.Log.Errorf("handling path %q failed: %v", update.Path, err)
	}
	fields, err := gnmiToFields(strings.Replace(gpath, "-", "_", -1), update.Val)
	if err != nil {
		c.Log.Errorf("error parsing update value %q: %v", update.Val, err)
	}
	return aliasPath, fields
}

// Parse path to path-buffer and tag-field
func handlePath(gnmiPath *gnmiLib.Path, tags map[string]string, aliases map[string]string, prefix string) (pathBuffer string, aliasPath string, err error) {
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

		if _, exists := aliases[name]; exists {
			aliasPath = name
		}

		if tags != nil {
			for key, val := range elem.Key {
				key = strings.ReplaceAll(key, "-", "_")

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

// ParsePath from XPath-like string to gNMI path structure
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

func convertTagOnlySubscription(s Subscription) TagSubscription {
	t := TagSubscription{Subscription: s, Elements: []string{"interface"}}
	return t
}

// equalPathNoKeys checks if two gNMI paths are equal, without keys
func equalPathNoKeys(a *gnmiLib.Path, b *gnmiLib.Path) bool {
	if len(a.Elem) != len(b.Elem) {
		return false
	}
	for i := range a.Elem {
		if a.Elem[i].Name != b.Elem[i].Name {
			return false
		}
	}
	return true
}

func pathKeys(gpath *gnmiLib.Path) []*gnmiLib.PathElem {
	var newPath []*gnmiLib.PathElem
	for _, elem := range gpath.Elem {
		if elem.Key != nil {
			newPath = append(newPath, elem)
		}
	}
	return newPath
}

func pathWithPrefix(prefix *gnmiLib.Path, gpath *gnmiLib.Path) *gnmiLib.Path {
	if prefix == nil {
		return gpath
	}
	fullPath := new(gnmiLib.Path)
	fullPath.Origin = prefix.Origin
	fullPath.Target = prefix.Target
	fullPath.Elem = append(prefix.Elem, gpath.Elem...)
	return fullPath
}

func (s *Subscription) buildFullPath(c *GNMI) error {
	var err error
	if s.fullPath, err = xpath.ToGNMIPath(s.Path); err != nil {
		return err
	}
	s.fullPath.Origin = s.Origin
	s.fullPath.Target = c.Target
	if c.Prefix != "" {
		prefix, err := xpath.ToGNMIPath(c.Prefix)
		if err != nil {
			return err
		}
		s.fullPath.Elem = append(prefix.Elem, s.fullPath.Elem...)
		if s.Origin == "" && c.Origin != "" {
			s.fullPath.Origin = c.Origin
		}
	}
	return nil
}

func (w *Worker) storeTags(update *gnmiLib.Update, sub TagSubscription) {
	updateKeys := pathKeys(update.Path)
	var foundKey bool
	for _, requiredKey := range sub.Elements {
		foundKey = false
		for _, elem := range updateKeys {
			if elem.Name == requiredKey {
				foundKey = true
			}
		}
		if !foundKey {
			return
		}
	}
	// All required keys present for this TagSubscription
	w.tagStore.insert(updateKeys, sub.Name, update.Val)
}

func (node *tagNode) insert(keys []*gnmiLib.PathElem, name string, value *gnmiLib.TypedValue) {
	if len(keys) == 0 {
		node.value = value
		node.tagName = name
		return
	}
	var found *tagNode
	key := keys[0]
	keyName := key.Name
	if node.tagStore == nil {
		node.tagStore = make(map[string][]*tagNode)
	}
	if _, ok := node.tagStore[keyName]; !ok {
		node.tagStore[keyName] = make([]*tagNode, 0)
	}
	for _, node := range node.tagStore[keyName] {
		if compareKeys(node.elem.Key, key.Key) {
			found = node
			break
		}
	}
	if found == nil {
		found = &tagNode{elem: keys[0]}
		node.tagStore[keyName] = append(node.tagStore[keyName], found)
	}
	found.insert(keys[1:], name, value)
}

func (node *tagNode) retrieve(keys []*gnmiLib.PathElem, tagResults *tagResults) {
	if node.value != nil {
		tagResults.names = append(tagResults.names, node.tagName)
		tagResults.values = append(tagResults.values, node.value)
	}
	for _, key := range keys {
		if elems, ok := node.tagStore[key.Name]; ok {
			for _, node := range elems {
				if compareKeys(node.elem.Key, key.Key) {
					node.retrieve(keys, tagResults)
				}
			}
		}
	}
}

func (w *Worker) checkTags(fullPath *gnmiLib.Path) map[string]interface{} {
	results := &tagResults{}
	w.tagStore.retrieve(pathKeys(fullPath), results)
	tags := make(map[string]interface{})
	for idx := range results.names {
		vals, _ := gnmiToFields(results.names[idx], results.values[idx])
		for k, v := range vals {
			tags[k] = v
		}
	}
	return tags
}

func (s *Subscription) buildAlias(aliases map[string]string) error {
	var err error
	var gnmiLongPath, gnmiShortPath *gnmiLib.Path

	// Build the subscription path without keys
	if gnmiLongPath, err = parsePath(s.Origin, s.Path, ""); err != nil {
		return err
	}
	if gnmiShortPath, err = parsePath("", s.Path, ""); err != nil {
		return err
	}

	longPath, _, err := handlePath(gnmiLongPath, nil, nil, "")
	if err != nil {
		return fmt.Errorf("handling long-path failed: %v", err)
	}
	shortPath, _, err := handlePath(gnmiShortPath, nil, nil, "")
	if err != nil {
		return fmt.Errorf("handling short-path failed: %v", err)
	}

	// If the user didn't provide a measurement name, use last path element
	name := s.Name
	if len(name) == 0 {
		name = path.Base(shortPath)
	}
	if len(name) > 0 {
		aliases[longPath] = name
		aliases[shortPath] = name
	}
	return nil
}

func gnmiToFields(name string, updateVal *gnmiLib.TypedValue) (map[string]interface{}, error) {
	var value interface{}
	var jsondata []byte

	// Make sure a value is actually set
	if updateVal == nil || updateVal.Value == nil {
		return nil, nil
	}

	switch val := updateVal.Value.(type) {
	case *gnmiLib.TypedValue_AsciiVal:
		value = val.AsciiVal
	case *gnmiLib.TypedValue_BoolVal:
		value = val.BoolVal
	case *gnmiLib.TypedValue_BytesVal:
		value = val.BytesVal
	case *gnmiLib.TypedValue_DoubleVal:
		value = val.DoubleVal
	case *gnmiLib.TypedValue_DecimalVal:
		//nolint:staticcheck // to maintain backward compatibility with older gnmi specs
		value = float64(val.DecimalVal.Digits) / math.Pow(10, float64(val.DecimalVal.Precision))
	case *gnmiLib.TypedValue_FloatVal:
		//nolint:staticcheck // to maintain backward compatibility with older gnmi specs
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

	fields := make(map[string]interface{})
	if value != nil {
		fields[name] = value
	} else if jsondata != nil {
		if err := json.Unmarshal(jsondata, &value); err != nil {
			return nil, fmt.Errorf("failed to parse JSON value: %v", err)
		}
		flattener := jsonparser.JSONFlattener{Fields: fields}
		if err := flattener.FullFlattenJSON(name, value, true, true); err != nil {
			return nil, fmt.Errorf("failed to flatten JSON: %v", err)
		}
	}
	return fields, nil
}

func compareKeys(a map[string]string, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if _, ok := b[k]; !ok {
			return false
		}
		if b[k] != v {
			return false
		}
	}
	return true
}
