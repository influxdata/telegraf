//go:generate ../../../tools/readme_config_includer/generator
package gnmi

import (
	"context"
	_ "embed"
	"fmt"
	"path"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/google/gnxi/utils/xpath"
	gnmiLib "github.com/openconfig/gnmi/proto/gnmi"
	"google.golang.org/grpc/metadata"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal/choice"
	internaltls "github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

// Regular expression to see if a path element contains an origin
var originPattern = regexp.MustCompile(`^([\w-_]+):`)

// Define the warning to show if we cannot get a metric name.
const emptyNameWarning = `Got empty metric-name for response, usually indicating
configuration issues as the response cannot be related to any subscription.
Please open an issue on https://github.com/influxdata/telegraf including your
device model and the following response data:
%+v
This message is only printed once.`

// Currently supported GNMI Extensions
var supportedExtensions = []string{"juniper_header"}

// gNMI plugin instance
type GNMI struct {
	Addresses           []string          `toml:"addresses"`
	Subscriptions       []Subscription    `toml:"subscription"`
	TagSubscriptions    []TagSubscription `toml:"tag_subscription"`
	Aliases             map[string]string `toml:"aliases"`
	Encoding            string            `toml:"encoding"`
	Origin              string            `toml:"origin"`
	Prefix              string            `toml:"prefix"`
	Target              string            `toml:"target"`
	UpdatesOnly         bool              `toml:"updates_only"`
	VendorSpecific      []string          `toml:"vendor_specific"`
	Username            string            `toml:"username"`
	Password            string            `toml:"password"`
	Redial              config.Duration   `toml:"redial"`
	MaxMsgSize          config.Size       `toml:"max_msg_size"`
	Trace               bool              `toml:"dump_responses"`
	CanonicalFieldNames bool              `toml:"canonical_field_names"`
	TrimFieldNames      bool              `toml:"trim_field_names"`
	EnableTLS           bool              `toml:"enable_tls" deprecated:"1.27.0;use 'tls_enable' instead"`
	Log                 telegraf.Logger   `toml:"-"`
	internaltls.ClientConfig

	// Internal state
	internalAliases map[string]string
	cancel          context.CancelFunc
	wg              sync.WaitGroup
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
	TagOnly bool `toml:"tag_only" deprecated:"1.25.0;2.0.0;please use 'tag_subscription's instead"`
}

// Tag Subscription for a gNMI client
type TagSubscription struct {
	Subscription
	Match    string `toml:"match"`
	Elements []string
}

func (*GNMI) SampleConfig() string {
	return sampleConfig
}

func (c *GNMI) Init() error {
	// Check options
	if time.Duration(c.Redial) <= 0 {
		return fmt.Errorf("redial duration must be positive")
	}

	// Check vendor_specific options configured by user
	if err := choice.CheckSlice(c.VendorSpecific, supportedExtensions); err != nil {
		return fmt.Errorf("unsupported vendor_specific option: %w", err)
	}

	// Use the new TLS option for enabling
	// Honor deprecated option
	enable := (c.ClientConfig.Enable != nil && *c.ClientConfig.Enable) || c.EnableTLS
	c.ClientConfig.Enable = &enable

	// Split the subscriptions into "normal" and "tag" subscription
	// and prepare them.
	for i := len(c.Subscriptions) - 1; i >= 0; i-- {
		subscription := c.Subscriptions[i]

		// Check the subscription
		if subscription.Name == "" {
			return fmt.Errorf("empty 'name' found for subscription %d", i+1)
		}
		if subscription.Path == "" {
			return fmt.Errorf("empty 'path' found for subscription %d", i+1)
		}

		// Support and convert legacy TagOnly subscriptions
		if subscription.TagOnly {
			tagSub := TagSubscription{
				Subscription: subscription,
				Match:        "name",
			}
			c.TagSubscriptions = append(c.TagSubscriptions, tagSub)
			// Remove from the original subscriptions list
			c.Subscriptions = append(c.Subscriptions[:i], c.Subscriptions[i+1:]...)
			continue
		}
		if err := subscription.buildFullPath(c); err != nil {
			return err
		}
	}
	for idx := range c.TagSubscriptions {
		if err := c.TagSubscriptions[idx].buildFullPath(c); err != nil {
			return err
		}
		if c.TagSubscriptions[idx].TagOnly != c.TagSubscriptions[0].TagOnly {
			return fmt.Errorf("do not mix legacy tag_only subscriptions and tag subscriptions")
		}
		switch c.TagSubscriptions[idx].Match {
		case "":
			if len(c.TagSubscriptions[idx].Elements) > 0 {
				c.TagSubscriptions[idx].Match = "elements"
			} else {
				c.TagSubscriptions[idx].Match = "name"
			}
		case "unconditional":
		case "name":
		case "elements":
			if len(c.TagSubscriptions[idx].Elements) == 0 {
				return fmt.Errorf("tag_subscription must have at least one element")
			}
		default:
			return fmt.Errorf("unknown match type %q for tag-subscription %q", c.TagSubscriptions[idx].Match, c.TagSubscriptions[idx].Name)
		}
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
	c.Log.Debugf("Internal alias mapping: %+v", c.internalAliases)

	return nil
}

func (c *GNMI) Start(acc telegraf.Accumulator) error {
	// Validate configuration
	request, err := c.newSubscribeRequest()
	if err != nil {
		return err
	}

	// Generate TLS config if enabled
	tlscfg, err := c.ClientConfig.TLSConfig()
	if err != nil {
		return err
	}

	// Prepare the context, optionally with credentials
	var ctx context.Context
	ctx, c.cancel = context.WithCancel(context.Background())
	if len(c.Username) > 0 {
		ctx = metadata.AppendToOutgoingContext(ctx, "username", c.Username, "password", c.Password)
	}

	// Create a goroutine for each device, dial and subscribe
	c.wg.Add(len(c.Addresses))
	for _, addr := range c.Addresses {
		go func(addr string) {
			defer c.wg.Done()

			h := handler{
				address:             addr,
				aliases:             c.internalAliases,
				tagsubs:             c.TagSubscriptions,
				maxMsgSize:          int(c.MaxMsgSize),
				vendorExt:           c.VendorSpecific,
				tagStore:            newTagStore(c.TagSubscriptions),
				trace:               c.Trace,
				canonicalFieldNames: c.CanonicalFieldNames,
				trimSlash:           c.TrimFieldNames,
				log:                 c.Log,
			}
			for ctx.Err() == nil {
				if err := h.subscribeGNMI(ctx, acc, tlscfg, request); err != nil && ctx.Err() == nil {
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
	subscriptions := make([]*gnmiLib.Subscription, 0, len(c.Subscriptions)+len(c.TagSubscriptions))
	for _, subscription := range c.TagSubscriptions {
		sub, err := subscription.buildSubscription()
		if err != nil {
			return nil, err
		}
		subscriptions = append(subscriptions, sub)
	}
	for _, subscription := range c.Subscriptions {
		sub, err := subscription.buildSubscription()
		if err != nil {
			return nil, err
		}
		subscriptions = append(subscriptions, sub)
	}

	// Construct subscribe request
	gnmiPath, err := parsePath(c.Origin, c.Prefix, c.Target)
	if err != nil {
		return nil, err
	}

	// Do not provide an empty prefix. Required for Huawei NE40 router v8.21
	// (and possibly others). See https://github.com/influxdata/telegraf/issues/12273.
	if gnmiPath.Origin == "" && gnmiPath.Target == "" && len(gnmiPath.Elem) == 0 {
		gnmiPath = nil
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

func (s *Subscription) buildAlias(aliases map[string]string) error {
	// Build the subscription path without keys
	gnmiPath, err := parsePath(s.Origin, s.Path, "")
	if err != nil {
		return err
	}

	origin, spath, _, err := handlePath(gnmiPath, nil, nil, "")
	if err != nil {
		return fmt.Errorf("handling path failed: %w", err)
	}

	// If the user didn't provide a measurement name, use last path element
	name := s.Name
	if name == "" {
		name = path.Base(spath)
	}
	if name != "" {
		aliases[origin+spath] = name
		aliases[spath] = name
	}
	return nil
}
