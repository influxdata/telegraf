package mavlink

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/bluenviron/gomavlib/v3"
)

// Convert a string from CamelCase to snake_case
func ConvertToSnakeCase(input string) string {
	re := regexp.MustCompile(`([a-z0-9])([A-Z])`)
	snake := re.ReplaceAllString(input, `${1}_${2}`)
	snake = strings.ToLower(snake)
	return snake
}

// Check if a string is in a slice
func Contains(slice []string, str string) bool {
	for _, item := range slice {
		if item == str {
			return true
		}
	}
	return false
}

// Convert a Mavlink event into a struct containing Metric data.
func MavlinkEventFrameToMetric(frm *gomavlib.EventFrame) MetricFrameData {
	out := MetricFrameData{}
	out.tags = make(map[string]string)
	out.fields = make(map[string]any)
	out.tags["sys_id"] = strconv.FormatUint(uint64(frm.SystemID()), 10)

	m := frm.Message()
	t := reflect.TypeOf(m)
	v := reflect.ValueOf(m)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)
		out.fields[ConvertToSnakeCase(field.Name)] = value.Interface()
	}

	out.name = ConvertToSnakeCase(t.Name())
	out.name = strings.TrimPrefix(out.name, "message_")
	return out
}

// Parse the FcuURL config to setup a mavlib endpoint config
func ParseMavlinkEndpointConfig(s *Mavlink) ([]gomavlib.EndpointConf, error) {
	if strings.HasPrefix(s.FcuURL, "serial://") {
		tmpStr := strings.TrimPrefix(s.FcuURL, "serial://")
		tmpStrParts := strings.Split(tmpStr, ":")
		deviceName := tmpStrParts[0]
		baudRate := 57600
		if len(tmpStrParts) == 2 {
			newBaudRate, err := strconv.Atoi(tmpStrParts[1])
			if err != nil {
				return nil, errors.New("mavlink setup error: serial baud rate not valid")
			}
			baudRate = newBaudRate
		}

		return []gomavlib.EndpointConf{
			gomavlib.EndpointSerial{
				Device: deviceName,
				Baud:   baudRate,
			},
		}, nil
	} else if strings.HasPrefix(s.FcuURL, "tcp://") {
		// TCP client
		tmpStr := strings.TrimPrefix(s.FcuURL, "tcp://")
		tmpStrParts := strings.Split(tmpStr, ":")
		if len(tmpStrParts) != 2 {
			return nil, errors.New("mavlink setup error: TCP requires a port")
		}

		hostname := tmpStrParts[0]
		port, err := strconv.Atoi(tmpStrParts[1])
		if err != nil {
			return nil, errors.New("mavlink setup error: TCP port is invalid")
		}

		if len(hostname) > 0 {
			return []gomavlib.EndpointConf{
				gomavlib.EndpointTCPClient{
					Address: fmt.Sprintf("%s:%d", hostname, port),
				},
			}, nil
		}

		return []gomavlib.EndpointConf{
			gomavlib.EndpointTCPServer{
				Address: fmt.Sprintf(":%d", port),
			},
		}, nil
	} else if strings.HasPrefix(s.FcuURL, "udp://") {
		// UDP client or server
		tmpStr := strings.TrimPrefix(s.FcuURL, "udp://")
		tmpStrParts := strings.Split(tmpStr, ":")
		if len(tmpStrParts) != 2 {
			return nil, errors.New("mavlink setup error: UDP requires a port")
		}

		hostname := tmpStrParts[0]
		port, err := strconv.Atoi(tmpStrParts[1])
		if err != nil {
			return nil, errors.New("mavlink setup error: UDP port is invalid")
		}

		if len(hostname) > 0 {
			return []gomavlib.EndpointConf{
				gomavlib.EndpointUDPClient{
					Address: fmt.Sprintf("%s:%d", hostname, port),
				},
			}, nil
		}

		return []gomavlib.EndpointConf{
			gomavlib.EndpointUDPServer{
				Address: fmt.Sprintf(":%d", port),
			},
		}, nil
	}

	return nil, errors.New("mavlink setup error: invalid fcu_url")
}
