package udp_listener

import (
	"log"
	"net"
	"sync"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"
)

type UdpListener struct {
	ServiceAddress         string
	UDPPacketSize          int `toml:"udp_packet_size"`
	AllowedPendingMessages int
	sync.Mutex

	in   chan []byte
	done chan struct{}

	parser parsers.Parser

	// Keep the accumulator in this struct
	acc telegraf.Accumulator
}

const UDP_PACKET_SIZE int = 1500

var dropwarn = "ERROR: Message queue full. Discarding line [%s] " +
	"You may want to increase allowed_pending_messages in the config\n"

const sampleConfig = `
  ## Address and port to host UDP listener on
  service_address = ":8092"

  ## Number of UDP messages allowed to queue up. Once filled, the
  ## UDP listener will start dropping packets.
  allowed_pending_messages = 10000

  ## UDP packet size for the server to listen for. This will depend
  ## on the size of the packets that the client is sending, which is
  ## usually 1500 bytes, but can be as large as 65,535 bytes.
  udp_packet_size = 1500

  ## Data format to consume. This can be "json", "influx" or "graphite"
  ## Each data format has it's own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "influx"
`

func (u *UdpListener) SampleConfig() string {
	return sampleConfig
}

func (u *UdpListener) Description() string {
	return "Generic UDP listener"
}

// All the work is done in the Start() function, so this is just a dummy
// function.
func (u *UdpListener) Gather(_ telegraf.Accumulator) error {
	return nil
}

func (u *UdpListener) SetParser(parser parsers.Parser) {
	u.parser = parser
}

func (u *UdpListener) Start(acc telegraf.Accumulator) error {
	u.Lock()
	defer u.Unlock()

	u.acc = acc
	u.in = make(chan []byte, u.AllowedPendingMessages)
	u.done = make(chan struct{})

	go u.udpListen()
	go u.udpParser()

	log.Printf("Started UDP listener service on %s\n", u.ServiceAddress)
	return nil
}

func (u *UdpListener) Stop() {
	u.Lock()
	defer u.Unlock()
	close(u.done)
	close(u.in)
	log.Println("Stopped UDP listener service on ", u.ServiceAddress)
}

func (u *UdpListener) udpListen() error {
	address, _ := net.ResolveUDPAddr("udp", u.ServiceAddress)
	listener, err := net.ListenUDP("udp", address)
	if err != nil {
		log.Fatalf("ERROR: ListenUDP - %s", err)
	}
	defer listener.Close()
	log.Println("UDP server listening on: ", listener.LocalAddr().String())

	for {
		select {
		case <-u.done:
			return nil
		default:
			buf := make([]byte, u.UDPPacketSize)
			n, _, err := listener.ReadFromUDP(buf)
			if err != nil {
				log.Printf("ERROR: %s\n", err.Error())
			}

			select {
			case u.in <- buf[:n]:
			default:
				log.Printf(dropwarn, string(buf[:n]))
			}
		}
	}
}

func (u *UdpListener) udpParser() error {
	for {
		select {
		case <-u.done:
			return nil
		case packet := <-u.in:
			metrics, err := u.parser.Parse(packet)
			if err == nil {
				u.storeMetrics(metrics)
			} else {
				log.Printf("Malformed packet: [%s], Error: %s\n", packet, err)
			}
		}
	}
}

func (u *UdpListener) storeMetrics(metrics []telegraf.Metric) error {
	u.Lock()
	defer u.Unlock()
	for _, m := range metrics {
		u.acc.AddFields(m.Name(), m.Fields(), m.Tags(), m.Time())
	}
	return nil
}

func init() {
	inputs.Add("udp_listener", func() telegraf.Input {
		return &UdpListener{
			UDPPacketSize: UDP_PACKET_SIZE,
		}
	})
}
