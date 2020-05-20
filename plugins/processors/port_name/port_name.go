package portname

import (
	"bufio"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

var sampleConfig = `
[[processors.port_name]]
  ## Name of tag holding the port number
  # tag = "port"

  ## Name of output tag where service name will be added
  # dest = "service"

  ## Default tcp or udp
  # default_protocol = "tcp"
`

type sMap map[string]map[int]string // "https" == services["tcp"][443]

var services sMap

type PortName struct {
	SourceTag       string `toml:"tag"`
	DestTag         string `toml:"dest"`
	DefaultProtocol string `toml:"default_protocol"`

	Log telegraf.Logger `toml:"-"`
}

func (d *PortName) SampleConfig() string {
	return sampleConfig
}

func (d *PortName) Description() string {
	return "Given a tag of a TCP or UDP port number, add a tag of the service name looked up in the system services file"
}

func readServicesFile() {
	file, err := os.Open(servicesPath())
	if err != nil {
		return
	}
	defer file.Close()

	services = readServices(file)
}

// Read the services file into a map.
//
// This function takes a similar approach to parsing as the go
// standard library (see src/net/port_unix.go in golang source) but
// maps protocol and port number to service name, not protocol and
// service to port number.
func readServices(r io.Reader) sMap {
	services = make(sMap)
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		// "http 80/tcp www www-http # World Wide Web HTTP"
		if i := strings.Index(line, "#"); i >= 0 {
			line = line[:i]
		}
		f := strings.Fields(line)
		if len(f) < 2 {
			continue
		}
		service := f[0]   // "http"
		portProto := f[1] // "80/tcp"
		portProtoSlice := strings.SplitN(portProto, "/", 2)
		if len(portProtoSlice) < 2 {
			continue
		}
		port, err := strconv.Atoi(portProtoSlice[0]) // "80"
		if err != nil || port <= 0 {
			continue
		}
		proto := portProtoSlice[1] // "tcp"
		proto = strings.ToLower(proto)

		protoMap, ok := services[proto]
		if !ok {
			protoMap = make(map[int]string)
			services[proto] = protoMap
		}
		protoMap[port] = service
	}
	return services
}

func (d *PortName) Apply(metrics ...telegraf.Metric) []telegraf.Metric {
	for _, m := range metrics {
		portProto, ok := m.GetTag(d.SourceTag)
		if !ok {
			// Nonexistent tag
			continue
		}
		portProtoSlice := strings.SplitN(portProto, "/", 2)
		l := len(portProtoSlice)

		if l == 0 {
			// Empty tag
			d.Log.Errorf("empty port tag: %v", d.SourceTag)
			continue
		}

		var port int
		if l > 0 {
			var err error
			val := portProtoSlice[0]
			port, err = strconv.Atoi(val)
			if err != nil {
				// Can't convert port to string
				d.Log.Errorf("error converting port to integer: %v", val)
				continue
			}
		}

		proto := d.DefaultProtocol
		if l > 1 && len(portProtoSlice[1]) > 0 {
			proto = portProtoSlice[1]
		}
		proto = strings.ToLower(proto)

		protoMap, ok := services[proto]
		if !ok {
			// Unknown protocol
			//
			// Protocol is normally tcp or udp.  The services file
			// normally has entries for both, so our map does too.  If
			// not, it's very likely the source tag or the services
			// file doesn't make sense.
			d.Log.Errorf("protocol not found in services map: %v", proto)
			continue
		}

		service, ok := protoMap[port]
		if !ok {
			// Unknown port
			//
			// Not all ports are named so this isn't an error, but
			// it's helpful to know when debugging.
			d.Log.Debugf("port not found in services map: %v", port)
			continue
		}

		m.AddTag(d.DestTag, service)
	}

	return metrics
}

func (h *PortName) Init() error {
	services = make(sMap)
	readServicesFile()
	return nil
}

func init() {
	processors.Add("port_name", func() telegraf.Processor {
		return &PortName{
			SourceTag:       "port",
			DestTag:         "service",
			DefaultProtocol: "tcp",
		}
	})
}
