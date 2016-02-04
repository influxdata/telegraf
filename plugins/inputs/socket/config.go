package socket

import "github.com/influxdata/telegraf/internal/encoding/graphite"

const (
	// DefaultBindAddress is the default binding interface if none is specified.
	DefaultBindAddress = ":2003"

	// DefaultProtocol is the default IP protocol used by the Graphite input.
	DefaultProtocol = "tcp"

	// DefaultUDPReadBuffer is the default buffer size for the UDP listener.
	// Sets the size of the operating system's receive buffer associated with
	// the UDP traffic. Keep in mind that the OS must be able
	// to handle the number set here or the UDP listener will error and exit.
	//
	// DefaultReadBuffer = 0 means to use the OS default, which is usually too
	// small for high UDP performance.
	//
	// Increasing OS buffer limits:
	//     Linux:      sudo sysctl -w net.core.rmem_max=<read-buffer>
	//     BSD/Darwin: sudo sysctl -w kern.ipc.maxsockbuf=<read-buffer>
	DefaultUdpReadBuffer = 0
)

// Config represents the configuration for Graphite endpoints.
type Config struct {
	BindAddress   string
	Protocol      string
	UdpReadBuffer int

	graphite.Config
}

// New Config instance.
func NewConfig(bindAddress, protocol string, udpReadBuffer int, separator string, templates []string) *Config {
	c := &Config{}
	if bindAddress == "" {
		bindAddress = DefaultBindAddress
	}
	if protocol == "" {
		protocol = DefaultProtocol
	}
	if udpReadBuffer < 0 {
		udpReadBuffer = DefaultUdpReadBuffer
	}
	if separator == "" {
		separator = graphite.DefaultSeparator
	}

	c.BindAddress = bindAddress
	c.Protocol = protocol
	c.UdpReadBuffer = udpReadBuffer
	c.Separator = separator
	c.Templates = templates

	return c
}
