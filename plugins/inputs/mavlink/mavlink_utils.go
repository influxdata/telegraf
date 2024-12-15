package mavlink

import (
	"fmt"
	"net"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/chrisdalke/gomavlib/v3"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/metric"
)

// Convert a Mavlink event into a struct containing Metric data.
func convertEventFrameToMetric(frm *gomavlib.EventFrame, filter filter.Filter) telegraf.Metric {
	m := frm.Message()
	t := reflect.TypeOf(m)
	v := reflect.ValueOf(m)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}

	name := internal.SnakeCase(strings.TrimPrefix(t.Name(), "Message"))

	if filter != nil && !filter.Match(name) {
		return nil
	}

	tags := map[string]string{
		"sys_id": strconv.FormatUint(uint64(frm.SystemID()), 10),
	}
	fields := make(map[string]interface{}, t.NumField())

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)
		fields[internal.SnakeCase(field.Name)] = value.Interface()
	}

	return metric.New(name, tags, fields, time.Now())
}

// Parse the URL config to setup a mavlib endpoint config
func parseMavlinkEndpointConfig(confUrl string) ([]gomavlib.EndpointConf, error) {
	// Try to parse the URL
	u, err := url.Parse(confUrl)
	if err != nil {
		return nil, fmt.Errorf("invalid url: %w", err)
	}

	// Split host and port, and use default port if it was not specified
	host, port, err := net.SplitHostPort(u.Host)
	if err != nil {
		// Use default port if we could not parse out the port.
		host = u.Host
		port = "14550"
	}

	switch u.Scheme {
	case "serial":
		// Serial client
		// Parse serial URL by hand, because it is not a compliant URL.
		baudRate := 57600
		device, rate, found := strings.Cut(strings.TrimPrefix(confUrl, "serial://"), ":")
		if found {
			r, err := strconv.Atoi(rate)
			if err != nil {
				return nil, fmt.Errorf("serial baud rate not valid: %w", err)
			}
			baudRate = r
		}

		return []gomavlib.EndpointConf{
			gomavlib.EndpointSerial{
				Device: device,
				Baud:   baudRate,
			},
		}, nil
	case "tcp":
		if len(host) > 0 {
			return []gomavlib.EndpointConf{
				gomavlib.EndpointTCPClient{
					Address: host + ":" + port,
				},
			}, nil
		}

		return []gomavlib.EndpointConf{
			gomavlib.EndpointTCPServer{
				Address: ":" + port,
			},
		}, nil
	case "udp":
		if len(host) > 0 {
			return []gomavlib.EndpointConf{
				gomavlib.EndpointUDPClient{
					Address: host + ":" + port,
				},
			}, nil
		}

		return []gomavlib.EndpointConf{
			gomavlib.EndpointUDPServer{
				Address: ":" + port,
			},
		}, nil
	default:
		return nil, fmt.Errorf("could not parse url %s", confUrl)
	}
}
