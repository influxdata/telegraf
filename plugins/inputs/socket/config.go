package socket

import "github.com/influxdata/telegraf/internal/encoding/graphite"

const (
	// DefaultBindAddress is the default binding interface if none is specified.
	DefaultBindAddress = ":2003"

	// DefaultProtocol is the default IP protocol used by the Graphite input.
	DefaultProtocol = "tcp"

	// DefaultSeparator is the default join character to use when joining multiple
	// measurment parts in a template.
	DefaultSeparator = "."

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

// WithDefaults takes the given config and returns a new config with any required
// default values set.
func (c *Config) WithDefaults() *Config {
	d := *c
	if d.BindAddress == "" {
		d.BindAddress = DefaultBindAddress
	}
	if d.Protocol == "" {
		d.Protocol = DefaultProtocol
	}
	if d.Separator == "" {
		d.Separator = DefaultSeparator
	}
	if d.UdpReadBuffer == 0 {
		d.UdpReadBuffer = DefaultUdpReadBuffer
	}
	return &d
}

// New Config instance.
func NewConfig(bindAddress, protocol string, udpReadBuffer int, separator string, tags []string, templates []string) *Config {
	c := &Config{}
	c.BindAddress = bindAddress
	c.Protocol = protocol
	c.UdpReadBuffer = udpReadBuffer

	c.Separator = separator
	c.Tags = tags
	c.Templates = templates

	return c
}
