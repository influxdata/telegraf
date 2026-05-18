//go:generate ../../../tools/config_includer/generator
//go:generate ../../../tools/readme_config_includer/generator
package gnmilistener

import (
	_ "embed"
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/influxdata/telegraf"
	common_gnmi "github.com/influxdata/telegraf/plugins/common/gnmi"
	common_tls "github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/inputs/gnmi_listener/nokia"
)

//go:embed sample.conf
var sampleConfig string

type serverImplementation interface {
	Register(*grpc.Server)
}

type GNMIListener struct {
	Address  string          `toml:"address"`
	Protocol string          `toml:"protocol"`
	Log      telegraf.Logger `toml:"-"`
	common_gnmi.HandlerConfig
	common_tls.ServerConfig

	options []grpc.ServerOption
	server  *grpc.Server
	handler *common_gnmi.Handler
	addr    string
}

func (*GNMIListener) SampleConfig() string {
	return sampleConfig
}

func (g *GNMIListener) Init() error {
	// Defaults
	if g.Address == "" {
		g.Address = "localhost:57400"
	}

	// Check user settings
	switch g.Protocol {
	case "":
		g.Protocol = "nokia"
	case "nokia":
		// Do nothing, those are valid
	default:
		return fmt.Errorf("invalid 'protocol' %q", g.Protocol)
	}

	// Fill the server options depending on the user settings
	if tlsCfg, err := g.ServerConfig.TLSConfig(); err != nil {
		return fmt.Errorf("creating TLS configuration failed: %w", err)
	} else if tlsCfg != nil {
		g.options = append(g.options, grpc.Creds(credentials.NewTLS(tlsCfg)))
	}

	if g.Log.Level().Includes(telegraf.Trace) {
		g.options = append(g.options, grpc.InTapHandle(g.logCalls))
	}

	// Create a response handler
	h, err := g.HandlerConfig.Handler(g.Log, common_gnmi.WithDefaultName("gnmi"))
	if err != nil {
		return fmt.Errorf("creating response handler failed: %w", err)
	}
	g.handler = h

	return nil
}

func (g *GNMIListener) Start(acc telegraf.Accumulator) error {
	// Create the protocol implementation
	var impl serverImplementation
	switch g.Protocol {
	case "nokia":
		impl = nokia.New(acc, g.handler, g.Log)
	default:
		return fmt.Errorf("invalid 'protocol' %q", g.Protocol)
	}

	// Create a listener or wrap it for debugging
	listener, err := net.Listen("tcp", g.Address)
	if err != nil {
		return fmt.Errorf("listening on %q failed: %w", g.Address, err)
	}
	g.addr = listener.Addr().String()

	// Start the server
	g.server = grpc.NewServer(g.options...)
	impl.Register(g.server)
	go func() {
		if err := g.server.Serve(listener); err != nil {
			g.Log.Errorf("Stopping GRPC server on %q due to error: %v", g.addr, err)
		}
	}()

	return nil
}

func (g *GNMIListener) Stop() {
	if g.server != nil {
		g.server.GracefulStop()
	}
}

func (*GNMIListener) Gather(telegraf.Accumulator) error {
	return nil
}

func init() {
	inputs.Add("gnmi_listener", func() telegraf.Input {
		return &GNMIListener{}
	})
}
