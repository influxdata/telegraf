package zabbix

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/influxdb/influxdb/client/v2"
	"github.com/influxdb/telegraf/outputs"
)

// Metric class.
type Metric struct {
	Host  string `json:"host"`
	Key   string `json:"key"`
	Value string `json:"value"`
	Clock int64  `json:"clock"`
}

// Metric class constructor.
func NewMetric(host, key, value string, clock ...int64) *Metric {
	m := &Metric{Host: host, Key: key, Value: value}
	// use current time, if `clock` is not specified
	if m.Clock = time.Now().Unix(); len(clock) > 0 {
		m.Clock = int64(clock[0])
	}
	return m
}

// Packet class.
type Packet struct {
	Request string    `json:"request"`
	Data    []*Metric `json:"data"`
	Clock   int64     `json:"clock"`
}

// Packet class cunstructor.
func NewPacket(data []*Metric, clock ...int64) *Packet {
	p := &Packet{Request: `sender data`, Data: data}
	// use current time, if `clock` is not specified
	if p.Clock = time.Now().Unix(); len(clock) > 0 {
		p.Clock = int64(clock[0])
	}
	return p
}

// DataLen Packet class method, return 8 bytes with packet length in little endian order.
func (p *Packet) DataLen() []byte {
	dataLen := make([]byte, 8)
	JSONData, _ := json.Marshal(p)
	binary.LittleEndian.PutUint32(dataLen, uint32(len(JSONData)))
	return dataLen
}

type Zabbix struct {
	Host    string
	Port    int
	Hosttag string

	Debug bool
}

var sampleConfig = `
	# Address of zabbix host
	host = "zabbix.example.com"

	# Port of the Zabbix server
	port = 10051

	# Which tag will be used for measurement hostname
	hosttag = "host"
`

func (z *Zabbix) Connect() error {
	// Test connection to Zabbix server
	// format: hostname:port
	uri := fmt.Sprintf("%s:%d", z.Host, z.Port)
	tcpAddr, err := net.ResolveTCPAddr("tcp", uri)
	if err != nil {
		return fmt.Errorf("Zabbix: TCP address cannot be resolved")
	}

	connection, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return fmt.Errorf("Zabbix: Connection to Zabbix server failed")
	}
	defer connection.Close()
	return nil
}

func (z *Zabbix) Write(points []*client.Point) error {
	if len(points) == 0 {
		return nil
	}
	// Send Data to Zabbix server
	uri := fmt.Sprintf("%s:%d", z.Host, z.Port)
	tcpAddr, _ := net.ResolveTCPAddr("tcp", uri)
	connection, err := net.DialTCP("tcp", nil, tcpAddr)
	defer connection.Close()

	if err != nil {
		return fmt.Errorf("Zabbix: Connection to Zabbix server failed")
	}

	var metrics []*Metric
	var hostname string
	for _, pt := range points {
		if h, ok := pt.Tags()[z.Hosttag]; !ok {
			if h, ok := pt.Tags()["host"]; !ok {
				h, err := os.Hostname()
				if err != nil {
					return fmt.Errorf("Cannot get os.Hostname()")
				}
				hostname = h
			} else {
				hostname = h
			}
		} else {
			hostname = h
		}
		metricValue, buildError := buildValue(pt)
		if buildError != nil {
			fmt.Printf("Zabbix: %s\n", buildError.Error())
		}
		if z.Debug {
			fmt.Printf("%s, %s, %s\n", hostname, pt.Name(), metricValue)
		}

		metrics = append(metrics, NewMetric(hostname, pt.Name(), metricValue))
	}
	packet := NewPacket(metrics)
	dataPacket, _ := json.Marshal(packet)

	buffer := append(z.getHeader(), packet.DataLen()...)
	buffer = append(buffer, dataPacket...)

	if _, err := connection.Write(buffer); err != nil {
		return fmt.Errorf("Zabbix: Sender writing error %s", err.Error())
	}

	return nil
}

func buildValue(pt *client.Point) (string, error) {
	var retv string
	var v = pt.Fields()["value"]
	switch p := v.(type) {
	case int64:
		retv = IntToString(int64(p))
	case uint64:
		retv = UIntToString(uint64(p))
	case float64:
		retv = FloatToString(float64(p))
	default:
		return retv, fmt.Errorf("unexpected type %T with value %v for Zabbix", v, v)
	}

	return retv, nil
}

func IntToString(input_num int64) string {
	return strconv.FormatInt(input_num, 10)
}

func UIntToString(input_num uint64) string {
	return strconv.FormatUint(input_num, 10)
}

func FloatToString(input_num float64) string {
	return strconv.FormatFloat(input_num, 'f', 6, 64)
}

func (z *Zabbix) getHeader() []byte {
	return []byte("ZBXD\x01")
}

func (z *Zabbix) SampleConfig() string {
	return sampleConfig
}

func (z *Zabbix) Description() string {
	return "Configuration for sender to Zabbix server"
}

func (z *Zabbix) Close() error {
	return nil
}

func init() {
	outputs.Add("zabbix", func() outputs.Output {
		return &Zabbix{}
	})
}
