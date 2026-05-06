//go:generate ../../../tools/config_includer/generator
//go:generate ../../../tools/readme_config_includer/generator
package gnmi

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/openconfig/gnmi/proto/gnmi"
	"github.com/openconfig/gnmi/proto/gnmi_ext"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal/choice"
	common_tls "github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/common/yangmodel"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

// Currently supported GNMI Extensions
var supportedExtensions = []string{"juniper_header"}

// Define the warning to show if we cannot get a metric name.
const emptyNameWarning = `Got empty metric-name for response (field %q), usually
indicating configuration issues as the response cannot be related to any
subscription.Please open an issue on https://github.com/influxdata/telegraf
including your device model and the following response data:
%+v
This message is only printed once.`

type GNMI struct {
	Addresses                     []string          `toml:"addresses"`
	Subscriptions                 []subscription    `toml:"subscription"`
	TagSubscriptions              []tagSubscription `toml:"tag_subscription"`
	Aliases                       map[string]string `toml:"aliases"`
	Encoding                      string            `toml:"encoding"`
	Origin                        string            `toml:"origin"`
	Prefix                        string            `toml:"prefix"`
	Target                        string            `toml:"target"`
	UpdatesOnly                   bool              `toml:"updates_only"`
	EmitDeleteMetrics             bool              `toml:"emit_delete_metrics"`
	VendorSpecific                []string          `toml:"vendor_specific"`
	Username                      config.Secret     `toml:"username"`
	Password                      config.Secret     `toml:"password"`
	Redial                        config.Duration   `toml:"redial"`
	MaxMsgSize                    config.Size       `toml:"max_msg_size"`
	Depth                         int32             `toml:"depth"`
	Trace                         bool              `toml:"dump_responses"`
	CanonicalFieldNames           bool              `toml:"canonical_field_names"`
	TrimFieldNames                bool              `toml:"trim_field_names"`
	PrefixTagKeyWithPath          bool              `toml:"prefix_tag_key_with_path"`
	GuessPathStrategy             string            `toml:"path_guessing_strategy"`
	KeepaliveTime                 config.Duration   `toml:"keepalive_time"`
	KeepaliveTimeout              config.Duration   `toml:"keepalive_timeout"`
	YangModelPaths                []string          `toml:"yang_model_paths"`
	EnforceFirstNamespaceAsOrigin bool              `toml:"enforce_first_namespace_as_origin"`
	Log                           telegraf.Logger   `toml:"-"`
	common_tls.ClientConfig

	// Internal state
	internalAliases map[*pathInfo]string
	decoder         *yangmodel.Decoder
	cancel          context.CancelFunc
	wg              sync.WaitGroup
}

func (*GNMI) SampleConfig() string {
	return sampleConfig
}

func (c *GNMI) Init() error {
	// Check options
	switch c.Encoding {
	case "":
		c.Encoding = "proto"
	case "proto", "json", "json_ietf", "bytes":
		// Do nothing, those are valid
	default:
		return fmt.Errorf("unsupported encoding %s", c.Encoding)
	}

	if time.Duration(c.Redial) <= 0 {
		return errors.New("redial duration must be positive")
	}

	// Check vendor_specific options configured by user
	if err := choice.CheckSlice(c.VendorSpecific, supportedExtensions); err != nil {
		return fmt.Errorf("unsupported vendor_specific option: %w", err)
	}

	// Check path guessing and handle deprecated option
	switch c.GuessPathStrategy {
	case "", "none", "common path", "subscription":
	default:
		return fmt.Errorf("invalid 'path_guessing_strategy' %q", c.GuessPathStrategy)
	}

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

		if err := subscription.buildFullPath(c.Origin, c.Prefix, c.Target); err != nil {
			return err
		}
	}
	for idx := range c.TagSubscriptions {
		if err := c.TagSubscriptions[idx].buildFullPath(c.Origin, c.Prefix, c.Target); err != nil {
			return err
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
		info, name, err := s.buildAlias(c.EnforceFirstNamespaceAsOrigin)
		if err != nil {
			return err
		}
		if info != nil && name != "" {
			c.internalAliases[info] = name
		}
	}
	for _, s := range c.TagSubscriptions {
		info, name, err := s.buildAlias(c.EnforceFirstNamespaceAsOrigin)
		if err != nil {
			return err
		}
		if info != nil && name != "" {
			c.internalAliases[info] = name
		}
	}
	for alias, encodingPath := range c.Aliases {
		path := newInfoFromString(encodingPath)
		if c.EnforceFirstNamespaceAsOrigin {
			path.enforceFirstNamespaceAsOrigin()
		}
		c.internalAliases[path] = alias
	}
	c.Log.Debugf("Internal alias mapping: %+v", c.internalAliases)

	// Use the new TLS option for enabling
	// Honor deprecated option
	enable := c.ClientConfig.Enable != nil && *c.ClientConfig.Enable
	c.ClientConfig.Enable = &enable

	// Warn about configures insecure cipher suites
	insecure := common_tls.InsecureCiphers(c.ClientConfig.TLSCipherSuites)
	if len(insecure) > 0 {
		c.Log.Warnf("Configured insecure cipher suites: %s", strings.Join(insecure, ","))
	}

	// Check the TLS configuration
	if _, err := c.ClientConfig.TLSConfig(); err != nil {
		if errors.Is(err, common_tls.ErrCipherUnsupported) {
			secure, insecure := common_tls.Ciphers()
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

			host, port, err := net.SplitHostPort(addr)
			if err != nil {
				acc.AddError(fmt.Errorf("unable to parse address %s: %w", addr, err))
				return
			}
			h := handler{
				host:                          host,
				port:                          port,
				aliases:                       c.internalAliases,
				tagsubs:                       c.TagSubscriptions,
				maxMsgSize:                    int(c.MaxMsgSize),
				vendorExt:                     c.VendorSpecific,
				emitDeleteMetrics:             c.EmitDeleteMetrics,
				tagStore:                      newTagStore(c.TagSubscriptions),
				trace:                         c.Trace,
				canonicalFieldNames:           c.CanonicalFieldNames,
				trimSlash:                     c.TrimFieldNames,
				tagPathPrefix:                 c.PrefixTagKeyWithPath,
				guessPathStrategy:             c.GuessPathStrategy,
				decoder:                       c.decoder,
				enforceFirstNamespaceAsOrigin: c.EnforceFirstNamespaceAsOrigin,
				log:                           c.Log,
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

func (c *GNMI) Stop() {
	c.cancel()
	c.wg.Wait()
}

func (*GNMI) Gather(telegraf.Accumulator) error {
	return nil
}

// Create a new gNMI SubscribeRequest
func (c *GNMI) newSubscribeRequest() (*gnmi.SubscribeRequest, error) {
	// Create subscription objects
	subscriptions := make([]*gnmi.Subscription, 0, len(c.Subscriptions)+len(c.TagSubscriptions))
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

	var extensions []*gnmi_ext.Extension
	if c.Depth > 0 {
		extensions = []*gnmi_ext.Extension{
			{
				Ext: &gnmi_ext.Extension_Depth{
					Depth: &gnmi_ext.Depth{
						Level: uint32(c.Depth),
					},
				},
			},
		}
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
		Extension: extensions,
	}, nil
}

func init() {
	inputs.Add("gnmi", func() telegraf.Input {
		return &GNMI{
			Redial:                        config.Duration(10 * time.Second),
			EnforceFirstNamespaceAsOrigin: true,
		}
	})
}
