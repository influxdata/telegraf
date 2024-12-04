package mavlink

import (
	"fmt"
	"net"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/chrisdalke/gomavlib/v3"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

// Convert a string from CamelCase to snake_case
// There is no single convention for Mavlink message names - Sometimes
// they are referenced as CAPITAL_SNAKE_CASE. Gomavlink converts them
// to CamelCase. This plugin takes an opinionated stance and makes the
// message names and field names all lowercase snake_case.
func ConvertToSnakeCase(input string) string {
	re := regexp.MustCompile(`([a-z0-9])([A-Z])`)
	snake := re.ReplaceAllString(input, `${1}_${2}`)
	snake = strings.ToLower(snake)
	return snake
}

// Convert a Mavlink event into a struct containing Metric data.
func MavlinkEventFrameToMetric(frm *gomavlib.EventFrame) telegraf.Metric {
	m := frm.Message()
	t := reflect.TypeOf(m)
	v := reflect.ValueOf(m)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}

	messageName := ConvertToSnakeCase(t.Name())
	messageName = strings.TrimPrefix(messageName, "message_")

	out := metric.New(
		messageName,
		make(map[string]string),
		make(map[string]interface{}),
		time.Unix(0, 0),
	)
	out.AddTag("sys_id", strconv.FormatUint(uint64(frm.SystemID()), 10))

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)
		out.AddField(ConvertToSnakeCase(field.Name), value.Interface())
	}

	return out
}

// Parse the FcuURL config to setup a mavlib endpoint config
func ParseMavlinkEndpointConfig(fcuURL string) ([]gomavlib.EndpointConf, error) {
	// Try to parse the URL
	u, err := url.Parse(fcuURL)
	if err != nil {
		return nil, fmt.Errorf("invalid fcu_url: %w", err)
	}

	// Split host and port, and use default port if it was not specified
	host, port, err := net.SplitHostPort(u.Host)
	if err != nil {
		return nil, fmt.Errorf("could not split fcu_url host and port: %w", err)
	}
	if port == "" {
		port = "14550"
	}

	if u.Scheme == "serial" {
		// Serial client
		// Parse serial URL by hand, because it is not technically a
		// compliant URL format, the URL parser may split the path
		// into parts awkwardly.
		tmpStr := strings.TrimPrefix(fcuURL, "serial://")
		tmpStrParts := strings.Split(tmpStr, ":")
		deviceName := tmpStrParts[0]
		baudRate := 57600
		if len(tmpStrParts) == 2 {
			newBaudRate, err := strconv.Atoi(tmpStrParts[1])
			if err != nil {
				return nil, fmt.Errorf("serial baud rate not valid: %w", err)
			}
			baudRate = newBaudRate
		}

		return []gomavlib.EndpointConf{
			gomavlib.EndpointSerial{
				Device: deviceName,
				Baud:   baudRate,
			},
		}, nil
	} else if u.Scheme == "tcp" {
		if len(host) > 0 {
			return []gomavlib.EndpointConf{
				gomavlib.EndpointTCPClient{
					Address: fmt.Sprintf("%s:%s", host, port),
				},
			}, nil
		}

		return []gomavlib.EndpointConf{
			gomavlib.EndpointTCPServer{
				Address: ":" + port,
			},
		}, nil
	} else if u.Scheme == "udp" {
		if len(host) > 0 {
			return []gomavlib.EndpointConf{
				gomavlib.EndpointUDPClient{
					Address: fmt.Sprintf("%s:%s", host, port),
				},
			}, nil
		}

		return []gomavlib.EndpointConf{
			gomavlib.EndpointUDPServer{
				Address: ":" + port,
			},
		}, nil
	}

	return nil, fmt.Errorf("could not parse fcu_url %s", fcuURL)
}
