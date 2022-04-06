package port_name

import (
	"bufio"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/processors"
)

type sMap map[string]map[int]string // "https" == services["tcp"][443]

var services sMap

type PortName struct {
	SourceTag       string `toml:"tag"`
	SourceField     string `toml:"field"`
	Dest            string `toml:"dest"`
	DefaultProtocol string `toml:"default_protocol"`
	ProtocolTag     string `toml:"protocol_tag"`
	ProtocolField   string `toml:"protocol_field"`

	Log telegraf.Logger `toml:"-"`
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

func (pn *PortName) Apply(metrics ...telegraf.Metric) []telegraf.Metric {
	for _, m := range metrics {
		var portProto string
		var fromField bool

		if len(pn.SourceTag) > 0 {
			if tag, ok := m.GetTag(pn.SourceTag); ok {
				portProto = tag
			}
		}
		if len(pn.SourceField) > 0 {
			if field, ok := m.GetField(pn.SourceField); ok {
				switch v := field.(type) {
				default:
					pn.Log.Errorf("Unexpected type %t in source field; must be string or int", v)
					continue
				case int64:
					portProto = strconv.FormatInt(v, 10)
				case uint64:
					portProto = strconv.FormatUint(v, 10)
				case string:
					portProto = v
				}
				fromField = true
			}
		}

		if len(portProto) == 0 {
			continue
		}

		portProtoSlice := strings.SplitN(portProto, "/", 2)
		l := len(portProtoSlice)

		if l == 0 {
			// Empty tag
			pn.Log.Errorf("empty port tag: %v", pn.SourceTag)
			continue
		}

		var port int
		if l > 0 {
			var err error
			val := portProtoSlice[0]
			port, err = strconv.Atoi(val)
			if err != nil {
				// Can't convert port to string
				pn.Log.Errorf("error converting port to integer: %v", val)
				continue
			}
		}

		proto := pn.DefaultProtocol
		if l > 1 && len(portProtoSlice[1]) > 0 {
			proto = portProtoSlice[1]
		}
		if len(pn.ProtocolTag) > 0 {
			if tag, ok := m.GetTag(pn.ProtocolTag); ok {
				proto = tag
			}
		}
		if len(pn.ProtocolField) > 0 {
			if field, ok := m.GetField(pn.ProtocolField); ok {
				switch v := field.(type) {
				default:
					pn.Log.Errorf("Unexpected type %t in protocol field; must be string", v)
					continue
				case string:
					proto = v
				}
			}
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
			pn.Log.Errorf("protocol not found in services map: %v", proto)
			continue
		}

		service, ok := protoMap[port]
		if !ok {
			// Unknown port
			//
			// Not all ports are named so this isn't an error, but
			// it's helpful to know when debugging.
			pn.Log.Debugf("port not found in services map: %v", port)
			continue
		}

		if fromField {
			m.AddField(pn.Dest, service)
		} else {
			m.AddTag(pn.Dest, service)
		}
	}

	return metrics
}

func (pn *PortName) Init() error {
	services = make(sMap)
	readServicesFile()
	return nil
}

func init() {
	processors.Add("port_name", func() telegraf.Processor {
		return &PortName{
			SourceTag:       "port",
			SourceField:     "port",
			Dest:            "service",
			DefaultProtocol: "tcp",
			ProtocolTag:     "proto",
			ProtocolField:   "proto",
		}
	})
}
