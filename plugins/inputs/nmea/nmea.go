package nmea

import (
	"io"
	"log"
	"net"
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/adrianmo/go-nmea"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/jacobsa/go-serial/serial"
)

type Serial struct {
	Options serial.OpenOptions
}

type Nmea struct {
	Serial      Serial
	Address     string
	Measurement string
	Include     []string
	UseGpsTime  bool

	port         io.ReadWriteCloser
	conn         net.Conn
	acc          telegraf.Accumulator
	configParsed bool
}

func (_ *Nmea) Description() string {
	return "Ping given url(s) and return statistics"
}

const sampleConfig = `
  ## The port to connect to read Nmea data
  port = COM8
`

func (_ *Nmea) SampleConfig() string {
	return sampleConfig
}

func (n *Nmea) ProcessNmeaPacket() error {
	return nil
}

func Contains(slice interface{}, item interface{}) bool {
	s := reflect.ValueOf(slice)

	if s.Kind() != reflect.Slice {
		panic("SliceExists() given a non-slice type")
	}

	for i := 0; i < s.Len(); i++ {
		if s.Index(i).Interface() == item {
			return true
		}
	}

	return false
}

func IsEmptyTime(nt nmea.Time) bool {
	return nt.Hour == 0 && nt.Minute == 0 && nt.Second == 0 && nt.Millisecond == 0
}

func IsEmptyDate(nd nmea.Date) bool {
	return nd.YY == 0 && nd.MM == 0 && nd.DD == 0
}

func (n *Nmea) Collect() {
	for true {
		var length int
		var err error

		sentence := make([]byte, 1024)
		// This is a blocking call. No need to have a sleep.
		if n.conn != nil {
			length, err = n.conn.Read(sentence)
			if err != nil {
				log.Println("ERROR [net.Read]:", err)
			}
		} else if n.port != nil {
			length, err = n.port.Read(sentence)
			if err != nil {
				log.Println("ERROR [serial.Read]:", err)
			}
		}

		if length > 0 {
			if sentence[length-1] == '\n' {
				sentence = sentence[:length-2]
			}
			s, err := nmea.Parse(string(sentence))
			if err != nil {
				log.Println("ERROR [nmea.Parse]:", err)
			}

			if s != nil {
				v := reflect.ValueOf(s)

				fields := make(map[string]interface{})
				tags := make(map[string]string)
				timestamp := time.Now()
				var tm nmea.Time
				var dt nmea.Date

				for i := 1; i < v.NumField(); i++ {
					sType := strings.TrimPrefix(v.Type().String(), "nmea.")
					if Contains(n.Include, sType) {
						typeName := v.Type().Field(i).Type.Name()
						if typeName == "Date" {
							dt = v.Field(i).Interface().(nmea.Date)
						}
						if typeName == "Time" {
							tm = v.Field(i).Interface().(nmea.Time)
						}
						fields[v.Type().Field(i).Name] = v.Field(i).Interface()
					}
				}

				if !IsEmptyDate(dt) && !IsEmptyTime(tm) && n.UseGpsTime {
					// We have a replacement time based on the GPS time. Use it!
					y := 1900 + dt.YY
					if dt.YY < 70 {
						y = y + 100
					}

					// We have milliseconds, but time.Date is expecting nanoseconds. Times by 1000000 to get milliseconds.
					timestamp = time.Date(y, time.Month(dt.MM), dt.DD, tm.Hour, tm.Minute, tm.Second, tm.Millisecond*1000000, time.UTC)
				}

				n.acc.AddFields(n.Measurement, fields, tags, timestamp.Local())
			}
		}
	}
}

func (n *Nmea) ParseConfig(pAcc *telegraf.Accumulator) error {
	var err error

	url, err := url.Parse(n.Address)
	if err == nil {
		n.conn, err = net.Dial(url.Scheme, url.Host)
		if err != nil {
			return err
		}
	} else {
		// If we didn't a URL successfully, don't worry about the serial port
		n.port, err = serial.Open(n.Serial.Options)
		if err != nil {
			return err
		}

		return err
	}

	// Set the accumulator for sending data.
	n.acc = *pAcc

	// Start up the data collection
	go n.Collect()

	return nil
}

func (n *Nmea) Gather(acc telegraf.Accumulator) error {
	if !n.configParsed {
		err := n.ParseConfig(&acc)
		if err != nil {
			return err
		}
		n.configParsed = true
	}

	return nil
}

func init() {
	inputs.Add("nmea", func() telegraf.Input {
		return &Nmea{configParsed: false}
	})
}
