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
`

type sMap map[string]map[int]string // "https" == services["tcp"][443]

var services sMap

type PortName struct {
	SourceTag       string `toml:"tag"`
	DestTag         string `toml:"dest"`
	DefaultProtocol string `toml:"default_protocol"`
}

func (d *PortName) SampleConfig() string {
	return sampleConfig
}

func (d *PortName) Description() string {
	return "Filter metrics with repeating field values"
}

func readServicesFile() {
	file, err := os.Open(servicesPath())
	if err != nil {
		return
	}
	defer file.Close()

	services = readServices(file)
}

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
			continue
		}

		var port int
		if l > 0 {
			var err error
			port, err = strconv.Atoi(portProtoSlice[0])
			if err != nil {
				// Can't convert port to string
				continue
			}
		}

		proto := d.DefaultProtocol
		if l > 1 && len(portProtoSlice[1]) > 0 {
			proto = portProtoSlice[1]
		}

		protoMap, ok := services[proto]
		if !ok {
			// Unknown protocol
			continue
		}

		service, ok2 := protoMap[port]
		if !ok2 {
			// Unknown port
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
