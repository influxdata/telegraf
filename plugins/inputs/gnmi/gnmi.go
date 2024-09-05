//go:generate ../../../tools/config_includer/generator
//go:generate ../../../tools/readme_config_includer/generator
package gnmi

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/gnxi/utils/xpath"
	gnmiLib "github.com/openconfig/gnmi/proto/gnmi"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal/choice"
	internaltls "github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/common/yangmodel"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

// Define the warning to show if we cannot get a metric name.
const emptyNameWarning = `Got empty metric-name for response (field %q), usually
indicating configuration issues as the response cannot be related to any
subscription.Please open an issue on https://github.com/influxdata/telegraf
including your device model and the following response data:
%+v
This message is only printed once.`

// Currently supported GNMI Extensions
var supportedExtensions = []string{"juniper_header"}

// gNMI plugin instance
type GNMI struct {
	Addresses            []string          `toml:"addresses"`
	Subscriptions        []Subscription    `toml:"subscription"`
	TagSubscriptions     []TagSubscription `toml:"tag_subscription"`
	Aliases              map[string]string `toml:"aliases"`
	Encoding             string            `toml:"encoding"`
	Origin               string            `toml:"origin"`
	Prefix               string            `toml:"prefix"`
	Target               string            `toml:"target"`
	UpdatesOnly          bool              `toml:"updates_only"`
	VendorSpecific       []string          `toml:"vendor_specific"`
	Username             config.Secret     `toml:"username"`
	Password             config.Secret     `toml:"password"`
	Redial               config.Duration   `toml:"redial"`
	MaxMsgSize           config.Size       `toml:"max_msg_size"`
	Trace                bool              `toml:"dump_responses"`
	CanonicalFieldNames  bool              `toml:"canonical_field_names"`
	TrimFieldNames       bool              `toml:"trim_field_names"`
	PrefixTagKeyWithPath bool              `toml:"prefix_tag_key_with_path"`
	GuessPathTag         bool              `toml:"guess_path_tag" deprecated:"1.30.0;1.35.0;use 'path_guessing_strategy' instead"`
	GuessPathStrategy    string            `toml:"path_guessing_strategy"`
	EnableTLS            bool              `toml:"enable_tls" deprecated:"1.27.0;1.35.0;use 'tls_enable' instead"`
	KeepaliveTime        config.Duration   `toml:"keepalive_time"`
	KeepaliveTimeout     config.Duration   `toml:"keepalive_timeout"`
	YangModelPaths       []string          `toml:"yang_model_paths"`
	Log                  telegraf.Logger   `toml:"-"`
	internaltls.ClientConfig

	// Internal state
	internalAliases map[*pathInfo]string
	decoder         *yangmodel.Decoder
	cancel          context.CancelFunc
	wg              sync.WaitGroup
}

// Subscription for a gNMI client
type Subscription struct {
	Name              string          `toml:"name"`
	Origin            string          `toml:"origin"`
	Path              string          `toml:"path"`
	SubscriptionMode  string          `toml:"subscription_mode"`
	SampleInterval    config.Duration `toml:"sample_interval"`
	SuppressRedundant bool            `toml:"suppress_redundant"`
	HeartbeatInterval config.Duration `toml:"heartbeat_interval"`
	TagOnly           bool            `toml:"tag_only" deprecated:"1.25.0;1.35.0;please use 'tag_subscription's instead"`

	fullPath *gnmiLib.Path
}

// Tag Subscription for a gNMI client
type TagSubscription struct {
	Subscription
	Match    string   `toml:"match"`
	Elements []string `toml:"elements"`
}

func (*GNMI) SampleConfig() string {
	return sampleConfig
}

func (c *GNMI) Init() error {
	// Check options
	if time.Duration(c.Redial) <= 0 {
		return errors.New("redial duration must be positive")
	}

	// Check vendor_specific options configured by user
	if err := choice.CheckSlice(c.VendorSpecific, supportedExtensions); err != nil {
		return fmt.Errorf("unsupported vendor_specific option: %w", err)
	}

	// Check path guessing and handle deprecated option
	if c.GuessPathTag {
		if c.GuessPathStrategy == "" {
			c.GuessPathStrategy = "common path"
		}
		if c.GuessPathStrategy != "common path" {
			return errors.New("conflicting settings between 'guess_path_tag' and 'path_guessing_strategy'")
		}
	}
	switch c.GuessPathStrategy {
	case "", "none", "common path", "subscription":
	default:
		return fmt.Errorf("invalid 'path_guessing_strategy' %q", c.GuessPathStrategy)
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
			return errors.New("do not mix legacy tag_only subscriptions and tag subscriptions")
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
				return errors.New("tag_subscription must have at least one element")
			}
		default:
			return fmt.Errorf("unknown match type %q for tag-subscription %q", c.TagSubscriptions[idx].Match, c.TagSubscriptions[idx].Name)
		}
	}

	// Invert explicit alias list and prefill subscription names
	c.internalAliases = make(map[*pathInfo]string, len(c.Subscriptions)+len(c.Aliases)+len(c.TagSubscriptions))
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
		c.internalAliases[newInfoFromString(encodingPath)] = alias
	}
	c.Log.Debugf("Internal alias mapping: %+v", c.internalAliases)

	// Warn about configures insecure cipher suites
	insecure := internaltls.InsecureCiphers(c.ClientConfig.TLSCipherSuites)
	if len(insecure) > 0 {
		c.Log.Warnf("Configured insecure cipher suites: %s", strings.Join(insecure, ","))
	}

	// Check the TLS configuration
	if _, err := c.ClientConfig.TLSConfig(); err != nil {
		if errors.Is(err, internaltls.ErrCipherUnsupported) {
			secure, insecure := internaltls.Ciphers()
			c.Log.Info("Supported secure ciphers:")
			for _, name := range secure {
				c.Log.Infof("  %s", name)
			}
			c.Log.Info("Supported insecure ciphers:")
			for _, name := range insecure {
				c.Log.Infof("  %s", name)
			}
		}
		return err
	}

	// Load the YANG models if specified by the user
	if len(c.YangModelPaths) > 0 {
		decoder, err := yangmodel.NewDecoder(c.YangModelPaths...)
		if err != nil {
			return fmt.Errorf("creating YANG model decoder failed: %w", err)
		}
		c.decoder = decoder
	}

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

	if !c.Username.Empty() {
		usernameSecret, err := c.Username.Get()
		if err != nil {
			return fmt.Errorf("getting username failed: %w", err)
		}
		username := usernameSecret.String()
		usernameSecret.Destroy()

		passwordSecret, err := c.Password.Get()
		if err != nil {
			return fmt.Errorf("getting password failed: %w", err)
		}
		password := passwordSecret.String()
		passwordSecret.Destroy()

		ctx = metadata.AppendToOutgoingContext(ctx, "username", username, "password", password)
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
				tagPathPrefix:       c.PrefixTagKeyWithPath,
				guessPathStrategy:   c.GuessPathStrategy,
				decoder:             c.decoder,
				log:                 c.Log,
				ClientParameters: keepalive.ClientParameters{
					Time:                time.Duration(c.KeepaliveTime),
					Timeout:             time.Duration(c.KeepaliveTimeout),
					PermitWithoutStream: false,
				},
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
func parsePath(origin, pathToParse, target string) (*gnmiLib.Path, error) {
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

func (s *Subscription) buildAlias(aliases map[*pathInfo]string) error {
	// Build the subscription path without keys
	path, err := parsePath(s.Origin, s.Path, "")
	if err != nil {
		return err
	}
	info := newInfoFromPathWithoutKeys(path)

	// If the user didn't provide a measurement name, use last path element
	name := s.Name
	if name == "" && len(info.segments) > 0 {
		name = info.segments[len(info.segments)-1].id
	}
	if name != "" {
		aliases[info] = name
	}
	return nil
}
