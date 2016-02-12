package zabbix

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
)

// ZabbixMetric class.
type ZabbixMetric struct {
	Host  string
	Key   string
	Value string
	Timestamp int64
}

// ZabbixMetric class constructor.
func NewZabbixMetric(host, key, value string) *ZabbixMetric {
	m := &ZabbixMetric{
		Host: host,
		Key: key,
		Value: value,
		Timestamp: time.Now().Unix()
	}
	return m
}

// ZabbixPacket class.
type ZabbixPacket struct {
	Request string
	Data    []*ZabbixMetric
	Timestamp   int64
}

// ZabbixPacket class cunstructor.
func NewZabbixPacket(data []*ZabbixMetric) *ZabbixPacket {
	p := &ZabbixPacket{
		Request: `sender data`,
		Data: data,
		Timestamp: time.Now().Unix()
	}
	return p
}

// DataLen ZabbixPacket class method, return 8 bytes with packet length in little endian order.
func (p *ZabbixPacket) DataLen() []byte {
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

func (z *Zabbix) Write(metrics []telegraf.Metric) error {
	if len(metrics) == 0 {
		return nil
	}
	// Send Data to Zabbix server
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

	var zabbixMetrics []*ZabbixMetric
	var hostname string
	for _, m := range metrics {
		if h, ok := m.Tags()[z.Hosttag]; !ok {
			if h, ok := m.Tags()["host"]; !ok {
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
		for fieldName, value := range m.Fields() {
			metricValue, buildError := buildValue(value)
			if buildError != nil {
				fmt.Printf("Zabbix: %s\n", buildError.Error())
			}
			if z.Debug {
				fmt.Printf("%s, %s, %s\n", hostname, fieldName, metricValue)
			}

			zabbixMetrics = append(zabbixMetrics, NewZabbixMetric(hostname, fieldName, metricValue))
		}
	}
	packet := NewZabbixPacket(zabbixMetrics)
	dataPacket, _ := json.Marshal(packet)

	buffer := append(z.getHeader(), packet.DataLen()...)
	buffer = append(buffer, dataPacket...)

	if _, err := connection.Write(buffer); err != nil {
		return fmt.Errorf("Zabbix: Sender writing error %s", err.Error())
	}

	return nil
}

func buildValue(v interface{}) (string, error) {
	var retv string
	switch p := v.(type) {
	case int64:
		retv = IntToString(int64(p))
	case uint64:
		retv = UIntToString(uint64(p))
	case float64:
		retv = FloatToString(float64(p))
	case string:
		retv = p
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
	outputs.Add("zabbix", func() telegraf.Output {
		return &Zabbix{}
	})
}
