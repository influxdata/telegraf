package arista_telemetry_gnmi

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/aristanetworks/goarista/gnmi"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	internaltls "github.com/influxdata/telegraf/internal/tls"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/inputs"
	jsonparser "github.com/influxdata/telegraf/plugins/parsers/json"

	pb "github.com/openconfig/gnmi/proto/gnmi"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

// AristaTelemetryGNMI plugin instance
type AristaTelemetryGNMI struct {
	Servers  []string
	Username string
	Password string
	Origin   string
	Paths    []string

	Compression       string
	UpdatesOnly       bool
	Mode              string
	StreamMode        string            `toml:"stream_mode"`
	SampleInterval    internal.Duration `toml:"sample_interval"`
	HeartbeatInterval internal.Duration `toml:"heartbeat_interval"`

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

var (
	// Regex to match and extract data points from path value in received key
	keyPathRegex = regexp.MustCompile("\\/([^\\/]*)\\[([A-Za-z0-9\\-\\/]*\\=[^\\[]*)\\]")
	sampleConfig = `
## Address and port of the GNMI GRPC server
 servers = ["127.0.0.1:6030"]
 ## define credentials
 username = "arista"
 password = "arista"
 ## GNMI encoding requested (one of: "proto", "json", "json_ietf")
 # encoding = "json"
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
 # origin = ""
 # prefix = ""
 # target = ""
 ## Origin and path of the subscription
 ## See: https://github.com/openconfig/reference/blob/master/rpc/gnmi/gnmi-specification.md#222-paths
 ##
 ## origin usually refers to a (YANG) data model implemented by the device
 ## and path to a specific substructe inside it that should be subscribed to (similar to an XPath)
  origin = "openconfig-interfaces"
  paths = ["/interfaces/interface/state/counters"]
  # Stream mode (one of: "target_defined", "sample", "on_change") and interval
  stream_mode = "sample"
  mode = "stream"
  sample_interval = "10s"
  ## Suppress redundant transmissions when measured values are unchanged
  # suppress_redundant = false
  ## If suppression is enabled, send updates at least every X seconds anyway
  # heartbeat_interval = "60s"
`
)

func (a *AristaTelemetryGNMI) SampleConfig() string {
	return sampleConfig
}

func (a *AristaTelemetryGNMI) Description() string {
	return "Read gNMI Telemetry from listed Paths"
}

func (a *AristaTelemetryGNMI) Gather(acc telegraf.Accumulator) error {
	return nil
}

func (a *AristaTelemetryGNMI) Stop() {
	a.cancel()
	a.wg.Wait()
}

func (a *AristaTelemetryGNMI) Start(acc telegraf.Accumulator) error {

	var err error
	var ctx context.Context
	var tlscfg *tls.Config

	a.acc = acc

	ctx, a.cancel = context.WithCancel(context.Background())

	if a.Redial.Duration.Nanoseconds() <= 0 {
		return fmt.Errorf("redial duration must be positive")
	}

	// Parse TLS config
	if a.EnableTLS {
		if tlscfg, err = a.ClientConfig.TLSConfig(); err != nil {
			return err
		}
	}
	if len(a.Username) > 0 {
		ctx = metadata.AppendToOutgoingContext(ctx, "username", a.Username, "password", a.Password)
	}

	for _, server := range a.Servers {
		a.wg.Add(1)
		go func(server string) {
			defer a.wg.Done()
			for ctx.Err() == nil {
				if err := a.CollectData(ctx, server, tlscfg, acc); err != nil && ctx.Err() == nil {
					acc.AddError(fmt.Errorf("failed to collect data from:%s %v", server, err))
				}

				select {
				case <-ctx.Done():
				case <-time.After(a.Redial.Duration):
				}
			}
		}(server)
	}
	return nil

}

// CollectData collects OpenConfig telemetry data from given server
func (a *AristaTelemetryGNMI) CollectData(ctx context.Context, server string, tlscfg *tls.Config, acc telegraf.Accumulator) error {
	var opts []grpc.DialOption

	subscribeoptions := &gnmi.SubscribeOptions{}

	if tlscfg != nil {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlscfg)))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}

	subscribeoptions.SampleInterval = uint64(a.SampleInterval.Duration / time.Millisecond)
	subscribeoptions.HeartbeatInterval = uint64(a.HeartbeatInterval.Duration / time.Millisecond)
	subscribeoptions.Mode = a.Mode
	subscribeoptions.Paths = gnmi.SplitPaths(a.Paths)

	grpcclient, err := grpc.DialContext(ctx, server, opts...)
	if err != nil {
		acc.AddError(fmt.Errorf("failed to dial:%s %v", server, err))
	}
	gnmiclient := pb.NewGNMIClient(grpcclient)
	respChan := make(chan *pb.SubscribeResponse)
	errChan := make(chan error)
	go gnmi.Subscribe(ctx, gnmiclient, subscribeoptions, respChan, errChan)
	for ctx.Err() == nil {
		select {
		case resp, open := <-respChan:
			if !open {
				continue
			}
			response := resp.GetResponse()

			update, ok := response.(*pb.SubscribeResponse_Update)
			if !ok {
				continue
			}
			a.handleSubscribeResponseUpdate(update, server, a.Paths)
		case err := <-errChan:
			acc.AddError(fmt.Errorf("E! Failed to read from %s: %v", server, err))
		}
	}
	return nil
}

// HandleSubscribeResponseUpdate message from GNMI and parse contained telemetry data
func (a *AristaTelemetryGNMI) handleSubscribeResponseUpdate(update *pb.SubscribeResponse_Update, server string, paths []string) {
	var prefix string
	var measurementname string
	timestamp := time.Unix(0, update.Update.Timestamp)
	grouper := metric.NewSeriesGrouper()
	prefixTags := make(map[string]string)

	if update.Update.Prefix != nil {
		prefix = a.handlePath(update.Update.Prefix, prefixTags, "")
	}
	prefixTags["source"], _, _ = net.SplitHostPort(server)

	// Parse individual Update message and create measurements
	for _, update := range update.Update.Update {
		// Prepare tags from prefix
		tags := make(map[string]string, len(prefixTags)+1)
		for key, val := range prefixTags {
			tags[key] = val
		}
		fields := a.handleTelemetryField(update, tags, prefix)
		prefix = a.handlePath(update.Path, prefixTags, "")
		for _, path := range paths {
			newpath := spitPath(path)
			if strings.HasPrefix(prefix, newpath) {
				measurementname = newpath
				tags["path"] = newpath
			}
		}

		// Group metrics
		for key, val := range fields {
			grouper.Add(measurementname, tags, timestamp, key[len(measurementname)+1:], val)
		}

	}

	// Add grouped measurements
	for _, metric := range grouper.Metrics() {
		a.acc.AddMetric(metric)
	}
}

// Takes in path with predicates and returns a path without predicates.
func spitPath(xpath string) string {
	subs := keyPathRegex.FindAllStringSubmatch(xpath, -1)

	// Given XML path, this will spit out final path without predicates
	if len(subs) > 0 {
		for _, sub := range subs {
			xpath = strings.Replace(xpath, sub[0], "/"+strings.TrimSpace(sub[1]), 1)
		}
	}
	xpath = strings.TrimSuffix(xpath, "/")

	return xpath
}

// HandleTelemetryField and add it to a measurement
func (a *AristaTelemetryGNMI) handleTelemetryField(update *pb.Update, tags map[string]string, prefix string) map[string]interface{} {
	path := a.handlePath(update.Path, tags, prefix)

	var value interface{}
	var jsondata []byte

	switch val := update.Val.Value.(type) {
	case *pb.TypedValue_AsciiVal:
		value = val.AsciiVal
	case *pb.TypedValue_BoolVal:
		value = val.BoolVal
	case *pb.TypedValue_BytesVal:
		value = val.BytesVal
	case *pb.TypedValue_DecimalVal:
		value = val.DecimalVal
	case *pb.TypedValue_FloatVal:
		value = val.FloatVal
	case *pb.TypedValue_IntVal:
		value = val.IntVal
	case *pb.TypedValue_StringVal:
		value = val.StringVal
	case *pb.TypedValue_UintVal:
		value = val.UintVal
	case *pb.TypedValue_JsonIetfVal:
		jsondata = val.JsonIetfVal
	case *pb.TypedValue_JsonVal:
		jsondata = val.JsonVal
	}

	name := strings.Replace(path, "-", "_", -1)
	fields := make(map[string]interface{})
	if value != nil {
		fields[name] = value
	} else if jsondata != nil {
		if err := json.Unmarshal(jsondata, &value); err != nil {
			a.acc.AddError(fmt.Errorf("failed to parse JSON value: %v", err))
		} else {
			flattener := jsonparser.JSONFlattener{Fields: fields}
			flattener.FullFlattenJSON(name, value, true, true)
		}
	}
	return fields
}

// Parse path to path-buffer and tag-field
func (a *AristaTelemetryGNMI) handlePath(path *pb.Path, tags map[string]string, prefix string) string {
	builder := bytes.NewBufferString(prefix)

	// Prefix with origin
	if len(path.Origin) > 0 {
		builder.WriteString(path.Origin)
		builder.WriteRune(':')
	}

	// Parse generic keys from prefix
	for _, elem := range path.Elem {
		builder.WriteRune('/')
		builder.WriteString(elem.Name)
		name := builder.String()

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

	return builder.String()
}
func init() {
	inputs.Add("arista_telemetry_gnmi", func() telegraf.Input {
		return &AristaTelemetryGNMI{
			Redial: internal.Duration{Duration: 10 * time.Second},
		}
	})
}
