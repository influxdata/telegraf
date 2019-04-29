/**
 * Copyright (c) 2018 Cisco Systems
 * Author: Steven Barth <stbarth@cisco.com>
 */

package cisco_telemetry_gnmi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
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
	ServiceAddress string         `toml:"service_address"`
	Subscriptions  []Subscription `toml:"subscription"`

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
	TLS bool
	internaltls.ClientConfig

	// Internal state
	acc    telegraf.Accumulator
	cancel context.CancelFunc
	ctx    context.Context
	wg     sync.WaitGroup
}

// Subscription for a GNMI client
type Subscription struct {
	Origin string
	Path   string
	Target string

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
	var opts []grpc.DialOption
	c.acc = acc
	c.ctx, c.cancel = context.WithCancel(context.Background())

	if c.TLS {
		tlsConfig, err := c.ClientConfig.TLSConfig()
		if err != nil {
			return err
		}

		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}

	if len(c.Username) > 0 {
		c.ctx = metadata.AppendToOutgoingContext(c.ctx, "username", c.Username, "password", c.Password)
	}

	client, err := grpc.Dial(c.ServiceAddress, opts...)
	if err != nil {
		return fmt.Errorf("E! Failed to dial GNMI: %v", err)
	}

	// Dialin client telemetry stream reading routine
	c.wg.Add(1)
	go c.subscribeGNMI(client)

	log.Printf("I! Started Cisco GNMI service for %s", c.ServiceAddress)

	return nil
}

// SubscribeGNMI and extract telemetry data
func (c *CiscoTelemetryGNMI) subscribeGNMI(client *grpc.ClientConn) {
	for c.ctx.Err() == nil {
		// Create subscription objects
		subscriptions := make([]*gnmi.Subscription, len(c.Subscriptions))
		for i, subscription := range c.Subscriptions {
			subscriptions[i] = &gnmi.Subscription{
				Path:              parsePath(subscription.Origin, subscription.Path, subscription.Target),
				Mode:              gnmi.SubscriptionMode(gnmi.SubscriptionMode_value[strings.ToUpper(subscription.SubscriptionMode)]),
				SampleInterval:    uint64(subscription.SampleInterval.Duration.Nanoseconds()),
				SuppressRedundant: subscription.SuppressRedundant,
				HeartbeatInterval: uint64(subscription.HeartbeatInterval.Duration.Nanoseconds()),
			}
		}

		// Construct subscribe request
		request := &gnmi.SubscribeRequest{
			Request: &gnmi.SubscribeRequest_Subscribe{
				Subscribe: &gnmi.SubscriptionList{
					Prefix:       parsePath(c.Origin, c.Prefix, c.Target),
					Mode:         gnmi.SubscriptionList_STREAM,
					Encoding:     gnmi.Encoding(gnmi.Encoding_value[strings.ToUpper(c.Encoding)]),
					Subscription: subscriptions,
					UpdatesOnly:  c.UpdatesOnly,
				},
			},
		}

		subscribeClient, err := gnmi.NewGNMIClient(client).Subscribe(c.ctx)
		if err != nil {
			c.acc.AddError(fmt.Errorf("E! GNMI subscription setup failed: %v", err))
		} else {
			err = subscribeClient.Send(request)
		}

		if err != nil {
			c.acc.AddError(fmt.Errorf("E! GNMI subscription setup failed: %v", err))
		} else {
			log.Printf("D! Connection to GNMI device %s established", c.ServiceAddress)
			for {
				reply, err := subscribeClient.Recv()

				if err != nil {
					if err != io.EOF && c.ctx.Err() == nil {
						c.acc.AddError(fmt.Errorf("E! GNMI subscription aborted: %v", err))
					}
					break
				}

				// Check for Update message, if not skip (e.g. Sync message)
				response, ok := reply.Response.(*gnmi.SubscribeResponse_Update)
				if !ok {
					continue
				}

				timestamp := time.Unix(0, response.Update.Timestamp)
				fields := make(map[string]interface{})
				tags := make(map[string]string)

				var builder bytes.Buffer

				if len(response.Update.Prefix.Origin) > 0 {
					builder.WriteString(response.Update.Prefix.Origin)
					builder.WriteRune(':')
				}
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

				tags["Producer"] = c.ServiceAddress
				tags["Target"] = response.Update.Prefix.Target
				builder.Truncate(builder.Len() - 1)
				prefix := builder.String()

				// Parse individual Update message and create measurement
				for _, update := range response.Update.Update {
					builder.Reset()
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

					switch update.Val.Value.(type) {
					case *gnmi.TypedValue_AsciiVal:
						value = update.Val.Value.(*gnmi.TypedValue_AsciiVal).AsciiVal
					case *gnmi.TypedValue_BoolVal:
						value = update.Val.Value.(*gnmi.TypedValue_BoolVal).BoolVal
					case *gnmi.TypedValue_BytesVal:
						value = update.Val.Value.(*gnmi.TypedValue_BytesVal).BytesVal
					case *gnmi.TypedValue_DecimalVal:
						value = update.Val.Value.(*gnmi.TypedValue_DecimalVal).DecimalVal
					case *gnmi.TypedValue_FloatVal:
						value = update.Val.Value.(*gnmi.TypedValue_FloatVal).FloatVal
					case *gnmi.TypedValue_IntVal:
						value = update.Val.Value.(*gnmi.TypedValue_IntVal).IntVal
					case *gnmi.TypedValue_StringVal:
						value = update.Val.Value.(*gnmi.TypedValue_StringVal).StringVal
					case *gnmi.TypedValue_UintVal:
						value = update.Val.Value.(*gnmi.TypedValue_UintVal).UintVal
					case *gnmi.TypedValue_JsonIetfVal:
						jsondata = update.Val.Value.(*gnmi.TypedValue_JsonIetfVal).JsonIetfVal
					case *gnmi.TypedValue_JsonVal:
						jsondata = update.Val.Value.(*gnmi.TypedValue_JsonVal).JsonVal
					}

					if value != nil {
						fields[builder.String()] = value
					} else if jsondata != nil {
						if err = json.Unmarshal(jsondata, &value); err != nil {
							c.acc.AddError(fmt.Errorf("W! GNMI JSON data is invalid: %v", err))
							continue
						}

						flattener := jsonparser.JSONFlattener{Fields: fields}
						flattener.FullFlattenJSON(builder.String(), value, true, true)
					}
				}

				// Finally add measurements
				c.acc.AddFields(prefix, fields, tags, timestamp)
			}

			log.Printf("D! Connection to GNMI device %s closed", c.ServiceAddress)
		}

		if c.Redial.Duration.Nanoseconds() <= 0 {
			break
		}

		select {
		case <-c.ctx.Done():
		case <-time.After(c.Redial.Duration):
		}
	}

	client.Close()
	c.wg.Done()
}

//ParsePath from XPath-like string to GNMI path structure
func parsePath(origin string, path string, target string) *gnmi.Path {
	gnmiPath := gnmi.Path{Origin: origin, Target: target}

	elem := &gnmi.PathElem{}
	start, name, value, end := 0, -1, -1, -1

	path = path + "/"

	for i := 0; i < len(path); i++ {
		switch path[i] {
		case '[':
			if end < 0 {
				end = i
				elem.Key = make(map[string]string)
			}
			if name < 0 {
				name = i + 1
			}

		case '=':
			if name > 0 && value < 0 {
				value = i + 1
			}

		case ']':
			if name > 0 && value > name {
				elem.Key[path[name:value-1]] = strings.Trim(path[value:i], "'\"")
			}
			name, value = -1, -1

		case '/':
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

	return &gnmiPath
}

// Stop listener and cleanup
func (c *CiscoTelemetryGNMI) Stop() {
	c.cancel()
	c.wg.Wait()

	log.Println("I! Stopped GNMI service on ", c.ServiceAddress)
}

const sampleConfig = `
  ## Address and port of the GNMI GRPC server
  service_address = "10.49.234.114:57777"

  ## define credentials
  username = "cisco"
  password = "cisco"

  ## redial in case of failures after
  redial = "10s"

  ## enable client-side TLS and define CA to authenticate the device
  # tls = true
  # tls_ca = "/etc/telegraf/ca.pem"
  # insecure_skip_verify = true

  ## define client-side TLS certificate & key to authenticate to the device
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key = "/etc/telegraf/key.pem"

  [[inputs.cisco_telemetry_gnmi.subscription]]
	origin = "Cisco-IOS-XR-infra-statsd-oper"
	path = "infra-statistics/interfaces/interface/latest/generic-counters"

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
