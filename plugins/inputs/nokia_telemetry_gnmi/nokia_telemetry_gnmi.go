package nokia_telemetry_gnmi

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/fullstorydev/grpcurl"
	"github.com/google/gnxi/utils/xpath"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/metric"
	internaltls "github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
	jsonparser "github.com/influxdata/telegraf/plugins/parsers/json"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	nokiasros "github.com/karimra/sros-dialout"
	"github.com/openconfig/gnmi/proto/gnmi"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

const sampleConfig = `
 ## DIAL-OUT
 ## enable dialout mode
 enable_dialout = false

 ## addr:port of telegraf grpc server
 server_address = ":57400"

 ## max message size
 # max_msg_size = 0
 
 ## enable server side TLS
 # tls_cert = "/etc/telegraf/cert.pem"
 # tls_key = "/etc/telegraf/key.pem"
 # tls_allowed_cacerts = ["/etc/telegraf/clientca.pem"]
 
 ## DIAL-IN
 ## Address and port of the GNMI GRPC server
 addresses = ["192.168.113.11:57400"]

 ## proto file path
 # proto_file = /path/to/proto/files

 ## proto dir path
 # proto_dir = /path/to/proto/dir
 
 ## username/password, the user should have grpc access rights
 username = "grpc"
 password = "Nokia4gnmi"

 ## GNMI encoding requested (one of: "json", "bytes", "json_ietf")
 # encoding = "json"

 ## redial wait time in case of failures
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

 [[inputs.nokia_telemetry_gnmi.subscription]]
  ## Name of the measurement that will be emitted
  name = "portcounters"

  ## Origin and path of the subscription
  ## See: https://github.com/openconfig/reference/blob/master/rpc/gnmi/gnmi-specification.md#222-paths
  ##
  ## origin usually refers to a (YANG) data model implemented by the device
  ## and path to a specific substructure inside it that should be subscribed to (similar to an XPath)
  ## YANG models can be found e.g. here: https://github.com/nokia/YangModels
  # origin = ""
  path = "/state/port[port-id=*]"

  # Subscription mode (one of: "target_defined", "sample", "on_change") and interval
  subscription_mode = "target_defined"
  sample_interval = "10s"

  ## Suppress redundant transmissions when measured values are unchanged
  # suppress_redundant = false

  ## If suppression is enabled, send updates at least every X seconds anyway
  # heartbeat_interval = "60s"
`

var supportedEncodings = []string{"bytes", "json", "json_ietf"}

// NokiaTelemetryGNMI is the plugin running instance
type NokiaTelemetryGNMI struct {
	// Telemetry type determines if this is a dial-in or dial-out subscription
	// in case of dial-out, next configuration parameters have no effect.

	EnableDialout bool `toml:"enable_dialout,omitempty"`

	// dial-out
	ServerAddress string `toml:"server_address,omitempty"`

	MaxMsgSize int `toml:"max_msg_size"`

	ProtoFile []string `toml:"proto_file,omitempty"`
	ProtoDir  []string `toml:"proto_dir,omitempty"`

	listener   net.Listener
	grpcServer *grpc.Server

	rootDesc desc.Descriptor

	internaltls.ServerConfig
	//gnmi.GNMIServer

	// dial-in
	Addresses     []string          `toml:"addresses"`
	Subscriptions []Subscription    `toml:"subscription"`
	Aliases       map[string]string `toml:"aliases"`

	Encoding    string
	Origin      string
	Prefix      string
	Target      string
	UpdatesOnly bool `toml:"updates_only"`

	Username string
	Password string

	Redial internal.Duration

	EnableTLS bool `toml:"enable_tls"`
	internaltls.ClientConfig

	aliases map[string]string
	acc     telegraf.Accumulator
	cancel  context.CancelFunc
	wg      sync.WaitGroup

	Log telegraf.Logger
}

// Subscription //
type Subscription struct {
	Name   string
	Origin string
	Path   string

	// Subscription mode: one of ["target_defined", "sample", "on_change"]
	SubscriptionMode string `toml:"subscription_mode"`

	// SampleInterval in case of target_defined or sample subscription
	SampleInterval internal.Duration `toml:"sample_interval"`

	// SuppressRedundant may be set for a sampled subscription.
	// if true the target SHOULD NOT generate a telemetry update message unless
	// the value of the path being reported on has changed since the last update was generated
	SuppressRedundant bool `toml:"suppress_redundant"`

	// HeartbeatInterval MAY be specified to modify the behavior of suppress_redundant in a sampled subscription.
	// In this case, the target MUST generate one telemetry update per heartbeat interval,
	// regardless of whether the suppress_redundant flag is set to true
	HeartbeatInterval internal.Duration `toml:"heartbeat_interval"`
}

// SampleConfig //
func (n *NokiaTelemetryGNMI) SampleConfig() string {
	return sampleConfig
}

// Description //
func (n *NokiaTelemetryGNMI) Description() string {
	return "Nokia GNMI telemetry input plugin based on GNMI telemetry data produced by Nokia 7750SR"
}

// Gather required to implement Input interface
func (n *NokiaTelemetryGNMI) Gather(acc telegraf.Accumulator) error {
	return nil
}

func init() {
	inputs.Add("nokia_telemetry_gnmi", func() telegraf.Input {
		return &NokiaTelemetryGNMI{
			EnableDialout: false,
			Encoding:      "json",
			Redial:        internal.Duration{Duration: 10 * time.Second},
		}
	})
}

// Start //
func (n *NokiaTelemetryGNMI) Start(acc telegraf.Accumulator) error {
	n.acc = acc
	var ctx context.Context
	ctx, n.cancel = context.WithCancel(context.Background())

	errs := make([]error, 0, 2)
	var err error
	dialoutFailed := true
	dialinFailed := true
	if n.EnableDialout {
		// load proto files
		if len(n.ProtoFile) > 0 || len(n.ProtoDir) > 0 {
			descSource, err := grpcurl.DescriptorSourceFromProtoFiles(n.ProtoDir, n.ProtoFile...)
			if err != nil {
				n.Log.Errorf("failed to load proto files: %v", err)
				return err
			}
			//descSource.FindSymbol()
			n.rootDesc, err = descSource.FindSymbol("Nokia.SROS.root")
			if err != nil {
				n.Log.Errorf("could not get symbol 'Nokia.SROS.root': %v", err)
				return err
			}
			n.Log.Debug("loaded proto files")
		}
		//
		err = n.startDialoutServer(acc)
		if err != nil {
			n.Log.Warnf("failed to start plugin in dialout mode: %v", err)
			errs = append(errs, err)
		} else {
			dialoutFailed = false
		}
	}
	if len(n.Subscriptions) > 0 {
		err := n.startDialInClients(ctx, acc)
		if err != nil {
			n.Log.Warnf("failed to start plugin in dial-in mode: %v", err)
			errs = append(errs, err)
		} else {
			dialinFailed = false
		}
	}
	if dialinFailed && dialoutFailed {
		return fmt.Errorf("both dialout and dial-in modes failed: %v || %v", errs[0], errs[1])
	}
	return nil
}

func (n *NokiaTelemetryGNMI) newSubscribeRequest() (*gnmi.SubscribeRequest, error) {
	// Create subscription objects from configuration file
	subscriptions := make([]*gnmi.Subscription, len(n.Subscriptions))
	for i, subscription := range n.Subscriptions {
		if len(subscription.Name) == 0 {
			return nil, fmt.Errorf("subscription index %d has an empty name", i)
		}
		gnmiPath, err := parsePath(subscription.Origin, subscription.Path, "")
		if err != nil {
			return nil, err
		}
		n.Log.Debugf("Subscription '%s' gnmiPath: %+v", subscription.Name, gnmiPath)
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

	// Create subscribe request
	gnmiPath, err := parsePath(n.Origin, n.Prefix, n.Target)
	if err != nil {
		return nil, err
	}
	n.Log.Debugf("Create subscribe request gnmiPath: %+v", gnmiPath)
	if !snl(strings.ToLower(n.Encoding), supportedEncodings) {
		return nil, fmt.Errorf("unsupported encoding %s", n.Encoding)
	}
	return &gnmi.SubscribeRequest{
		Request: &gnmi.SubscribeRequest_Subscribe{
			Subscribe: &gnmi.SubscriptionList{
				Prefix:       gnmiPath,
				Mode:         gnmi.SubscriptionList_STREAM,
				Encoding:     gnmi.Encoding(gnmi.Encoding_value[strings.ToUpper(n.Encoding)]),
				Subscription: subscriptions,
				UpdatesOnly:  n.UpdatesOnly,
			},
		},
	}, nil
}

func (n *NokiaTelemetryGNMI) subscribeGNMI(ctx context.Context, address string, tlscfg *tls.Config, request *gnmi.SubscribeRequest) error {
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

	n.Log.Debugf("Connection to GNMI device %s established", address)
	defer n.Log.Debugf("Connection to GNMI device %s closed", address)
	for ctx.Err() == nil {
		var reply *gnmi.SubscribeResponse
		if reply, err = subscribeClient.Recv(); err != nil {
			if err != io.EOF && ctx.Err() == nil {
				return fmt.Errorf("aborted GNMI subscription: %v", err)
			}
			break
		}

		n.handleSubscribeResponse(address, reply, "")
	}
	return nil
}

func (n *NokiaTelemetryGNMI) handleSubscribeResponse(address string, reply *gnmi.SubscribeResponse, subscriptionName string) {
	n.Log.Debugf("got notification: %+v", reply)
	response, ok := reply.Response.(*gnmi.SubscribeResponse_Update)
	if !ok {
		return
	}
	var prefix, prefixAliasPath string
	grouper := metric.NewSeriesGrouper()
	timestamp := time.Unix(0, response.Update.Timestamp)
	prefixTags := make(map[string]string)

	if response.Update.Prefix != nil {
		prefix, prefixAliasPath = n.handlePath(response.Update.Prefix, prefixTags, "")
	}
	prefixTags["source"], _, _ = net.SplitHostPort(address)
	prefixTags["path"] = prefix

	var name, lastAliasPath string
	for _, update := range response.Update.Update {
		tags := make(map[string]string, len(prefixTags))
		for key, val := range prefixTags {
			tags[key] = val
		}
		aliasPath, fields := n.handleTelemetryField(update, tags, prefix)

		if len(prefixAliasPath) > 0 && len(aliasPath) == 0 {
			aliasPath = prefixAliasPath
		}
		if aliasPath != lastAliasPath {
			name = prefix
			if alias, ok := n.aliases[aliasPath]; ok {
				name = alias
			} else {
				n.Log.Debugf("No measurement alias for GNMI path: %s", name)
			}
		}
		if subscriptionName != "" {
			name = subscriptionName
		}
		if name == "" {
			name = "from_dialout"
		}
		// Group metrics
		for k, v := range fields {
			key := k
			if len(aliasPath) < len(key) {
				key = key[len(aliasPath)+1:]
			} else {
				key = path.Base(key)
				key = strings.TrimLeft(key, "/.")
				if key == "" {
					n.Log.Errorf("invalid empty path: %q", k)
					continue
				}
			}

			grouper.Add(name, tags, timestamp, key, v)
			n.Log.Debugf("measurement_name=%s, tags=%+v, timestamp=%s, key=%s, value=%v", name, tags, timestamp, key, v)
		}

		lastAliasPath = aliasPath
	}

	// Add grouped measurements
	for _, metric := range grouper.Metrics() {
		n.Log.Debugf("metric: %+v", metric)
		n.acc.AddMetric(metric)
	}
}

// HandleTelemetryField
func (n *NokiaTelemetryGNMI) handleTelemetryField(update *gnmi.Update, tags map[string]string, prefix string) (string, map[string]interface{}) {
	path, aliasPath := n.handlePath(update.Path, tags, prefix)

	var value interface{}
	var jsondata []byte

	// Make sure a value is actually set
	if update.Val == nil || update.Val.Value == nil {
		n.Log.Infof("Discarded empty or legacy type value with path: %q", path)
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
		value = val.DecimalVal
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
	case *gnmi.TypedValue_ProtoBytes:
		if n.rootDesc != nil {
			m := dynamic.NewMessage(n.rootDesc.GetFile().FindMessage("Nokia.SROS.root"))
			err := m.Unmarshal(update.Val.GetProtoBytes())
			if err != nil {
				n.Log.Errorf("failed to unmarshal m: %v", err)
			}
			tb, err := m.MarshalText()
			if err != nil {
				n.Log.Errorf("failed to marshal txt dynamic msg: %v", err)
			} else {
				n.Log.Debugf("text format=%s", string(tb))
			}
			jsondata, err = m.MarshalJSON()
			if err != nil {
				n.Log.Errorf("failed to marshal dynamic msg: %v", err)
			} else {
				n.Log.Debugf("json format=%s", string(jsondata))
			}
		} else {
			value = update.Val.GetProtoBytes()
		}
	}

	name := strings.Replace(path, "-", "_", -1)
	fields := make(map[string]interface{})
	if value != nil {
		fields[name] = value
	} else if jsondata != nil {
		if err := json.Unmarshal(jsondata, &value); err != nil {
			n.acc.AddError(fmt.Errorf("failed to parse JSON value: %v", err))
		} else {
			flattener := jsonparser.JSONFlattener{Fields: fields}
			flattener.FullFlattenJSON(name, value, true, true)
		}
	}
	return aliasPath, fields
}

// Parse path to path-buffer and tag-field
func (n *NokiaTelemetryGNMI) handlePath(path *gnmi.Path, tags map[string]string, prefix string) (string, string) {
	if path == nil {
		return "", ""
	}
	var aliasPath string

	builder := strings.Builder{}
	builder.WriteString(prefix)

	if len(path.Origin) > 0 {
		builder.WriteString(path.Origin)
		builder.WriteRune(':')
	}

	for _, elem := range path.Elem {
		if len(elem.Name) > 0 {
			builder.WriteRune('/')
			builder.WriteString(elem.Name)
		}
		name := builder.String()

		if _, exists := n.aliases[name]; exists {
			aliasPath = name
		}

		if tags != nil {
			for key, val := range elem.Key {
				tags[strings.Replace(key, "-", "_", -1)] = val
			}
		}
	}

	return builder.String(), aliasPath
}

func parsePath(origin string, path string, target string) (*gnmi.Path, error) {
	gnmiPath, err := xpath.ToGNMIPath(path)
	if err != nil {
		return nil, err
	}
	gnmiPath.Origin = origin
	gnmiPath.Target = target
	return gnmiPath, nil
}

// Stop //
func (n *NokiaTelemetryGNMI) Stop() {
	if n.grpcServer != nil {
		n.grpcServer.Stop()
	}
	if n.listener != nil {
		n.listener.Close()
	}

	n.cancel()
	n.wg.Wait()
}

func snl(s string, l []string) bool {
	for _, sl := range l {
		if s == sl {
			return true
		}
	}
	return false
}

func (n *NokiaTelemetryGNMI) startDialoutServer(acc telegraf.Accumulator) error {
	var err error
	n.acc = acc
	n.listener, err = net.Listen("tcp", n.ServerAddress)
	if err != nil {
		return err
	}
	n.aliases = make(map[string]string, len(n.Aliases))
	for alias, path := range n.Aliases {
		n.aliases[path] = alias
	}
	var opts []grpc.ServerOption
	tlsConfig, err := n.ServerConfig.TLSConfig()
	if err != nil {
		n.listener.Close()
		return err
	} else if tlsConfig != nil {
		opts = append(opts, grpc.Creds(credentials.NewTLS(tlsConfig)))
	}

	if n.MaxMsgSize > 0 {
		opts = append(opts, grpc.MaxRecvMsgSize(n.MaxMsgSize))
	}

	n.grpcServer = grpc.NewServer(opts...)
	nokiasros.RegisterDialoutTelemetryServer(n.grpcServer, n)
	n.wg.Add(1)
	go func() {
		n.grpcServer.Serve(n.listener)
		n.wg.Done()
	}()
	return nil
}

// Publish //
func (n *NokiaTelemetryGNMI) Publish(stream nokiasros.DialoutTelemetry_PublishServer) error {
	for {
		md, ok := metadata.FromIncomingContext(stream.Context())
		if ok {
			n.Log.Debugf("Publish:::metadata = %+v", md)
		}

		peer, ok := peer.FromContext(stream.Context())
		if ok {
			n.Log.Debugf("connection from peer: %+v", peer)
		}

		subResp, err := stream.Recv()
		n.Log.Debugf("%+v", subResp)
		if err != nil {
			if err != io.EOF {
				n.acc.AddError(fmt.Errorf("GRPC dialout receive error: %v", err))
			}
			break
		}
		err = stream.Send(&nokiasros.PublishResponse{})
		if err != nil {
			n.Log.Debugf("error sending publish response to server: %v", err)
		}
		subName := fmt.Sprintf("from_dialout_%s", peer.Addr.String())
		if sn, ok := md["subscription-name"]; ok {
			if len(sn) > 0 {
				subName = sn[0]
			}
		}
		n.Log.Debugf("subscription-name: %v", subName)
		n.handleSubscribeResponse(peer.Addr.String(), subResp, subName)
	}
	return nil
}

func (n *NokiaTelemetryGNMI) startDialInClients(ctx context.Context, acc telegraf.Accumulator) error {
	tlsCfg := new(tls.Config)
	request, err := n.newSubscribeRequest()
	if err != nil {
		return err
	} else if n.Redial.Duration.Nanoseconds() <= 0 {
		return fmt.Errorf("redial duration must be positive")
	}

	if n.EnableTLS {
		if tlsCfg, err = n.ClientConfig.TLSConfig(); err != nil {
			return err
		}
	}

	if len(n.Username) > 0 {
		ctx = metadata.AppendToOutgoingContext(ctx, "username", n.Username, "password", n.Password)
	}

	n.aliases = make(map[string]string, len(n.Subscriptions)+len(n.Aliases))
	for _, subscription := range n.Subscriptions {
		var gnmiLongPath, gnmiShortPath *gnmi.Path

		if gnmiLongPath, err = parsePath(subscription.Origin, subscription.Path, ""); err != nil {
			return err
		}
		n.Log.Debugf("gnmiLongPath: %+v", gnmiLongPath)
		if gnmiShortPath, err = parsePath("", subscription.Path, ""); err != nil {
			return err
		}
		n.Log.Debugf("gnmiShortPath: %+v", gnmiShortPath)
		longPath, _ := n.handlePath(gnmiLongPath, nil, "")
		shortPath, _ := n.handlePath(gnmiShortPath, nil, "")
		n.Log.Debugf("longPath: %s", longPath)
		n.Log.Debugf("shortPath: %s", shortPath)
		n.aliases[longPath] = subscription.Name
		n.aliases[shortPath] = subscription.Name
	}
	for alias, path := range n.Aliases {
		n.aliases[path] = alias
	}
	n.Log.Debugf("Aliases: %+v", n.aliases)
	n.wg.Add(len(n.Addresses))
	for _, addr := range n.Addresses {
		go func(address string) {
			defer n.wg.Done()
			for ctx.Err() == nil {
				if err := n.subscribeGNMI(ctx, address, tlsCfg, request); err != nil && ctx.Err() == nil {
					acc.AddError(err)
				}

				select {
				case <-ctx.Done():
				case <-time.After(n.Redial.Duration):
				}
			}
		}(addr)
	}
	return nil
}
