//go:generate ../../../tools/readme_config_includer/generator
package mavlink

import (
	"context"
	_ "embed"
	"fmt"
	"net"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bluenviron/gomavlib/v3"
	"github.com/bluenviron/gomavlib/v3/pkg/dialects/ardupilotmega"
	"github.com/bluenviron/gomavlib/v3/pkg/frame"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type Mavlink struct {
	URL                    string   `toml:"url"`
	SystemID               uint8    `toml:"system_id"`
	FilterPattern          []string `toml:"filter"`
	StreamRequestFrequency uint16   `toml:"stream_request_frequency"`

	Log telegraf.Logger `toml:"-"`

	filter         filter.Filter
	connection     *gomavlib.Node
	endpointConfig []gomavlib.EndpointConf
	cancel         context.CancelFunc
	wg             sync.WaitGroup
}

//go:embed sample.conf
var sampleConfig string

func (*Mavlink) SampleConfig() string {
	return sampleConfig
}

func (m *Mavlink) Init() error {
	// Set default values
	if m.URL == "" {
		m.URL = "tcp://127.0.0.1:5760"
	}

	// Parse out the Mavlink endpoint.
	// Try to parse the URL
	u, err := url.Parse(m.URL)
	if err != nil {
		return fmt.Errorf("invalid url: %w", err)
	}

	switch u.Scheme {
	case "serial":
		// Serial client
		// Parse serial URL by hand, because it is not a compliant URL.
		baudRate := 57600
		device, rate, found := strings.Cut(strings.TrimPrefix(m.URL, "serial://"), ":")
		if found {
			r, err := strconv.Atoi(rate)
			if err != nil {
				return fmt.Errorf("serial baud rate not valid: %w", err)
			}
			baudRate = r
		}

		m.endpointConfig = []gomavlib.EndpointConf{
			gomavlib.EndpointSerial{
				Device: device,
				Baud:   baudRate,
			},
		}
	case "tcp":
		// Split host and port, and use default port if it was not specified
		host, port, err := net.SplitHostPort(u.Host)
		if err != nil {
			// Use default mavlink TCP port if a port was not provided or was invalid.
			host = u.Host
			port = "5760"
		}

		if host == "" {
			m.endpointConfig = []gomavlib.EndpointConf{
				gomavlib.EndpointTCPServer{
					Address: "0.0.0.0:" + port,
				},
			}
		} else {
			m.endpointConfig = []gomavlib.EndpointConf{
				gomavlib.EndpointTCPClient{
					Address: host + ":" + port,
				},
			}
		}

	case "udp":
		// Split host and port, and use default port if it was not specified
		host, port, err := net.SplitHostPort(u.Host)
		if err != nil {
			// Use default mavlink UDP port if a port was not provided or was invalid.
			host = u.Host
			port = "14550"
		}

		if host == "" {
			m.endpointConfig = []gomavlib.EndpointConf{
				gomavlib.EndpointUDPServer{
					Address: "0.0.0.0:" + port,
				},
			}
		} else {
			m.endpointConfig = []gomavlib.EndpointConf{
				gomavlib.EndpointUDPClient{
					Address: host + ":" + port,
				},
			}
		}

	default:
		return fmt.Errorf("unknown scheme %q", u.Scheme)
	}

	// Compile filter
	m.filter, err = filter.Compile(m.FilterPattern)
	if err != nil {
		return fmt.Errorf("compiling filter failed: %w", err)
	}

	return nil
}

func (m *Mavlink) Start(acc telegraf.Accumulator) error {
	// Start MAVLink endpoint
	connection, err := gomavlib.NewNode(gomavlib.NodeConf{
		Endpoints:              m.endpointConfig,
		Dialect:                ardupilotmega.Dialect,
		OutVersion:             gomavlib.V2,
		OutSystemID:            m.SystemID,
		StreamRequestEnable:    m.StreamRequestFrequency > 0,
		StreamRequestFrequency: int(m.StreamRequestFrequency),
	})
	if err != nil {
		return &internal.StartupError{
			Err:   fmt.Errorf("connecting to mavlink endpoint %s failed: %w", m.URL, err),
			Retry: true,
		}
	}
	m.connection = connection
	ctx, cancelFunc := context.WithCancel(context.Background())
	m.cancel = cancelFunc

	// Start routine to connect to Mavlink and stream out data async
	m.wg.Add(1)
	go func(ctx context.Context) {
		defer m.connection.Close()
		defer m.wg.Done()

		for {
			select {
			case <-ctx.Done():
				return
			case evt := <-m.connection.Events():
				switch evt := evt.(type) {
				case *gomavlib.EventFrame:
					if result := convertFrameToMetric(evt.Frame, m.filter); result != nil {
						result.AddTag("source", m.URL)
						acc.AddMetric(result)
					}
				case *gomavlib.EventChannelOpen:
					m.Log.Tracef("Mavlink channel opened")
				case *gomavlib.EventChannelClose:
					m.Log.Tracef("Mavlink channel closed")
				case *gomavlib.EventParseError:
					m.Log.Tracef("Mavlink parse error: %s", evt.Error.Error())
				case *gomavlib.EventStreamRequested:
					m.Log.Tracef("Issued stream request to system %d, component %d", evt.SystemID, evt.ComponentID)
				default:
					m.Log.Tracef("Unhandled Mavlink event: %T", evt)
				}
			}
		}
	}(ctx)

	return nil
}

func (*Mavlink) Gather(telegraf.Accumulator) error {
	return nil
}

func (m *Mavlink) Stop() {
	m.cancel()
	m.wg.Wait()
}

// Convert a Mavlink frame into a struct containing Metric data.
func convertFrameToMetric(frm frame.Frame, msgFilter filter.Filter) telegraf.Metric {
	m := frm.GetMessage()
	t := reflect.TypeOf(m)
	v := reflect.ValueOf(m)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}

	name := internal.SnakeCase(strings.TrimPrefix(t.Name(), "Message"))

	if msgFilter != nil && !msgFilter.Match(name) {
		return nil
	}

	tags := map[string]string{
		"sys_id": strconv.FormatUint(uint64(frm.GetSystemID()), 10),
	}
	fields := make(map[string]interface{}, t.NumField())

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)
		fieldName := internal.SnakeCase(field.Name)

		if value.Kind() == reflect.Slice || value.Kind() == reflect.Array {
			// Split array types into individual primitive values
			// with _<n> appended to the key
			for j := 0; j < value.Len(); j++ {
				indexedFieldName := fmt.Sprintf("%s_%d", fieldName, j+1)
				fields[indexedFieldName] = value.Index(j).Interface()
			}
		} else {
			fields[fieldName] = value.Interface()
		}
	}

	return metric.New(name, tags, fields, time.Now())
}

func init() {
	inputs.Add("mavlink", func() telegraf.Input {
		return &Mavlink{
			SystemID:               254,
			StreamRequestFrequency: 4,
		}
	})
}
