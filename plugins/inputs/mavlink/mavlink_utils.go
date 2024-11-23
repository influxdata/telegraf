package mavlink

import (
	_ "embed"
	"errors"
	"fmt"
	"log"
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
	out.tags = map[string]string{}
	out.fields = make(map[string]interface{})
	out.tags["sys_id"] = fmt.Sprintf("%d", frm.SystemID())

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
	if strings.HasPrefix(out.name, "message_") {
		out.name = strings.TrimPrefix(out.name, "message_")
	}

	return out
}

// Parse the FcuUrl config to setup a mavlib endpoint config
func ParseMavlinkEndpointConfig(s *Mavlink) ([]gomavlib.EndpointConf, error) {
	if strings.HasPrefix(s.FcuUrl, "serial://") {
		tmpStr := strings.TrimPrefix(s.FcuUrl, "serial://")
		tmpStrParts := strings.Split(tmpStr, ":")
		deviceName := tmpStrParts[0]
		baudRate := 57600
		if len(tmpStrParts) == 2 {
			newBaudRate, err := strconv.Atoi(tmpStrParts[1])
			if err != nil {
				return nil, errors.New("Mavlink setup error: serial baud rate not valid!")
			}
			baudRate = newBaudRate
		}

		log.Printf("Mavlink serial client: device %s, baud rate %d", deviceName, baudRate)
		return []gomavlib.EndpointConf{
			gomavlib.EndpointSerial{
				Device: deviceName,
				Baud:   baudRate,
			},
		}, nil
	} else if strings.HasPrefix(s.FcuUrl, "tcp://") {
		// TCP client
		tmpStr := strings.TrimPrefix(s.FcuUrl, "tcp://")
		tmpStrParts := strings.Split(tmpStr, ":")
		if len(tmpStrParts) != 2 {
			return nil, errors.New("Mavlink setup error: TCP requires a port!")
		}

		hostname := tmpStrParts[0]
		port := 14550
		port, err := strconv.Atoi(tmpStrParts[1])
		if err != nil {
			return nil, errors.New("Mavlink setup error: TCP port is invalid!")
		}

		if len(hostname) > 0 {
			log.Printf("Mavlink TCP client: hostname %s, port %d", hostname, port)
			return []gomavlib.EndpointConf{
				gomavlib.EndpointTCPClient{fmt.Sprintf("%s:%d", hostname, port)},
			}, nil
		} else {
			log.Printf("Mavlink TCP server: port %d", port)
			return []gomavlib.EndpointConf{
				gomavlib.EndpointTCPServer{fmt.Sprintf(":%d", port)},
			}, nil
		}
	} else if strings.HasPrefix(s.FcuUrl, "udp://") {
		// UDP client or server
		tmpStr := strings.TrimPrefix(s.FcuUrl, "udp://")
		tmpStrParts := strings.Split(tmpStr, ":")
		if len(tmpStrParts) != 2 {
			return nil, errors.New("Mavlink setup error: UDP requires a port!")
		}

		hostname := tmpStrParts[0]
		port := 14550
		port, err := strconv.Atoi(tmpStrParts[1])
		if err != nil {
			return nil, errors.New("Mavlink setup error: UDP port is invalid!")
		}

		if len(hostname) > 0 {
			log.Printf("Mavlink UDP client: hostname %s, port %d", hostname, port)
			return []gomavlib.EndpointConf{
				gomavlib.EndpointUDPClient{fmt.Sprintf("%s:%d", hostname, port)},
			}, nil
		} else {
			log.Printf("Mavlink UDP server: port %d", port)
			return []gomavlib.EndpointConf{
				gomavlib.EndpointUDPServer{fmt.Sprintf(":%d", port)},
			}, nil
		}
	} else {
		return nil, errors.New("Mavlink setup error: invalid fcu_url!")
	}
}
