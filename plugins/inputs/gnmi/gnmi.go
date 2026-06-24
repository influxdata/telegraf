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
	common_gnmi "github.com/influxdata/telegraf/plugins/common/gnmi"
	common_tls "github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type GNMI struct {
	Addresses                     []string                      `toml:"addresses"`
	Subscriptions                 []common_gnmi.Subscription    `toml:"subscription"`
	TagSubscriptions              []common_gnmi.TagSubscription `toml:"tag_subscription"`
	Aliases                       map[string]string             `toml:"aliases"`
	Encoding                      string                        `toml:"encoding"`
	Origin                        string                        `toml:"origin"`
	Prefix                        string                        `toml:"prefix"`
	Target                        string                        `toml:"target"`
	UpdatesOnly                   bool                          `toml:"updates_only"`
	Username                      config.Secret                 `toml:"username"`
	Password                      config.Secret                 `toml:"password"`
	Redial                        config.Duration               `toml:"redial"`
	MaxMsgSize                    config.Size                   `toml:"max_msg_size"`
	Depth                         int32                         `toml:"depth"`
	KeepaliveTime                 config.Duration               `toml:"keepalive_time"`
	KeepaliveTimeout              config.Duration               `toml:"keepalive_timeout"`
	EnforceFirstNamespaceAsOrigin bool                          `toml:"enforce_first_namespace_as_origin"`
	Log                           telegraf.Logger               `toml:"-"`
	common_tls.ClientConfig
	common_gnmi.HandlerConfig

	// Internal state
	handler *common_gnmi.Handler
	cancel  context.CancelFunc
	wg      sync.WaitGroup
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

	// Create a response handler
	var options []common_gnmi.Option
	if c.EnforceFirstNamespaceAsOrigin {
		options = append(options, common_gnmi.WithEnforceFirstNamespaceAsOrigin())
	}
	h, err := c.HandlerConfig.Handler(c.Log, options...)
	if err != nil {
		return fmt.Errorf("creating response handler failed: %w", err)
	}
	c.handler = h

	// Prepare the subscriptions and add corresponding aliases
	for i := range c.Subscriptions {
		s := c.Subscriptions[i]

		// Check the subscription
		if s.Name == "" {
			return fmt.Errorf("empty 'name' found for subscription %d", i+1)
		}
		if s.Path == "" {
			return fmt.Errorf("empty 'path' found for subscription %d", i+1)
		}

		if err := s.Build(c.Origin, c.Prefix, c.Target); err != nil {
			return err
		}

		// Update the handler
		if err := c.handler.AddAliasFromSubscription(s); err != nil {
			return fmt.Errorf("adding alias for subscription %q (%s) failed: %w", s.Name, s.Path, err)
		}
	}

	for i := range c.TagSubscriptions {
		s := &c.TagSubscriptions[i]
		if err := s.Build(c.Origin, c.Prefix, c.Target); err != nil {
			return err
		}
		switch s.Match {
		case "":
			if len(s.Elements) > 0 {
				s.Match = "elements"
			} else {
				s.Match = "name"
			}
		case "unconditional", "name":
		case "elements":
			if len(s.Elements) == 0 {
				return errors.New("tag_subscription must have at least one element")
			}
		default:
			return fmt.Errorf("unknown match type %q for tag-subscription %q", s.Match, s.Name)
		}

		// Update the handler
		if err := c.handler.AddAliasFromSubscription(s.Subscription); err != nil {
			return fmt.Errorf("adding alias for tag-subscription %q (%s) failed: %w", s.Name, s.Path, err)
		}
		c.handler.AddTagSubscription(s)
	}

	// Add the user-specified aliases
	for alias, path := range c.Aliases {
		c.handler.AddAlias(alias, path)
	}

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

			h := subscriber{
				handler:    c.handler,
				host:       host,
				port:       port,
				maxMsgSize: int(c.MaxMsgSize),
				log:        c.Log,
				ClientParameters: keepalive.ClientParameters{
					Time:                time.Duration(c.KeepaliveTime),
					Timeout:             time.Duration(c.KeepaliveTimeout),
					PermitWithoutStream: false,
				},
			}
			for ctx.Err() == nil {
				if err := h.subscribe(ctx, acc, tlscfg, request); err != nil && ctx.Err() == nil {
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
	requests := make([]*gnmi.Subscription, 0, len(c.Subscriptions)+len(c.TagSubscriptions))
	for _, subscription := range c.TagSubscriptions {
		req, err := subscription.Request()
		if err != nil {
			return nil, err
		}
		requests = append(requests, req)
	}
	for _, subscription := range c.Subscriptions {
		req, err := subscription.Request()
		if err != nil {
			return nil, err
		}
		requests = append(requests, req)
	}

	// Construct subscribe request
	gnmiPath, err := common_gnmi.ParsePath(c.Origin, c.Prefix, c.Target)
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
				Subscription: requests,
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
