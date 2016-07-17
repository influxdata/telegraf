package tcp_listener

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/plugins/parsers"
)

type TcpListener struct {
	ServiceAddress         string
	AllowedPendingMessages int
	MaxTCPConnections      int `toml:"max_tcp_connections"`

	sync.Mutex
	// Lock for preventing a data race during resource cleanup
	cleanup sync.Mutex
	wg      sync.WaitGroup

	in   chan []byte
	done chan struct{}
	// accept channel tracks how many active connections there are, if there
	// is an available bool in accept, then we are below the maximum and can
	// accept the connection
	accept chan bool
	// drops tracks the number of dropped metrics.
	drops int
	// malformed tracks the number of malformed packets
	malformed int

	// track the listener here so we can close it in Stop()
	listener *net.TCPListener
	// track current connections so we can close them in Stop()
	conns map[string]*net.TCPConn

	parser parsers.Parser
	acc    telegraf.Accumulator
}

var dropwarn = "ERROR: tcp_listener message queue full. " +
	"We have dropped %d messages so far. " +
	"You may want to increase allowed_pending_messages in the config\n"

var malformedwarn = "WARNING: tcp_listener has received %d malformed packets" +
	" thus far."

const sampleConfig = `
  ## Address and port to host TCP listener on
  service_address = ":8094"

  ## Number of TCP messages allowed to queue up. Once filled, the
  ## TCP listener will start dropping packets.
  allowed_pending_messages = 10000

  ## Maximum number of concurrent TCP connections to allow
  max_tcp_connections = 250

  ## Data format to consume.
  ## Each data format has it's own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/influxdata/telegraf/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "influx"
`

func (t *TcpListener) SampleConfig() string {
	return sampleConfig
}

func (t *TcpListener) Description() string {
	return "Generic TCP listener"
}

// All the work is done in the Start() function, so this is just a dummy
// function.
func (t *TcpListener) Gather(_ telegraf.Accumulator) error {
	return nil
}

func (t *TcpListener) SetParser(parser parsers.Parser) {
	t.parser = parser
}

// Start starts the tcp listener service.
func (t *TcpListener) Start(acc telegraf.Accumulator) error {
	t.Lock()
	defer t.Unlock()

	t.acc = acc
	t.in = make(chan []byte, t.AllowedPendingMessages)
	t.done = make(chan struct{})
	t.accept = make(chan bool, t.MaxTCPConnections)
	t.conns = make(map[string]*net.TCPConn)
	for i := 0; i < t.MaxTCPConnections; i++ {
		t.accept <- true
	}

	// Start listener
	var err error
	address, _ := net.ResolveTCPAddr("tcp", t.ServiceAddress)
	t.listener, err = net.ListenTCP("tcp", address)
	if err != nil {
		log.Fatalf("ERROR: ListenUDP - %s", err)
		return err
	}
	log.Println("TCP server listening on: ", t.listener.Addr().String())

	t.wg.Add(2)
	go t.tcpListen()
	go t.tcpParser()

	log.Printf("Started TCP listener service on %s\n", t.ServiceAddress)
	return nil
}

// Stop cleans up all resources
func (t *TcpListener) Stop() {
	t.Lock()
	defer t.Unlock()
	close(t.done)
	t.listener.Close()

	// Close all open TCP connections
	//  - get all conns from the t.conns map and put into slice
	//  - this is so the forget() function doesnt conflict with looping
	//    over the t.conns map
	var conns []*net.TCPConn
	t.cleanup.Lock()
	for _, conn := range t.conns {
		conns = append(conns, conn)
	}
	t.cleanup.Unlock()
	for _, conn := range conns {
		conn.Close()
	}

	t.wg.Wait()
	close(t.in)
	log.Println("Stopped TCP listener service on ", t.ServiceAddress)
}

// tcpListen listens for incoming TCP connections.
func (t *TcpListener) tcpListen() error {
	defer t.wg.Done()

	for {
		select {
		case <-t.done:
			return nil
		default:
			// Accept connection:
			conn, err := t.listener.AcceptTCP()
			if err != nil {
				return err
			}
			// log.Printf("Received TCP Connection from %s", conn.RemoteAddr())

			select {
			case <-t.accept:
				// not over connection limit, handle the connection properly.
				t.wg.Add(1)
				// generate a random id for this TCPConn
				id := internal.RandomString(6)
				t.remember(id, conn)
				go t.handler(conn, id)
			default:
				// We are over the connection limit, refuse & close.
				t.refuser(conn)
			}
		}
	}
}

// refuser refuses a TCP connection
func (t *TcpListener) refuser(conn *net.TCPConn) {
	// Tell the connection why we are closing.
	fmt.Fprintf(conn, "Telegraf maximum concurrent TCP connections (%d)"+
		" reached, closing.\nYou may want to increase max_tcp_connections in"+
		" the Telegraf tcp listener configuration.\n", t.MaxTCPConnections)
	conn.Close()
	log.Printf("Refused TCP Connection from %s", conn.RemoteAddr())
	log.Printf("WARNING: Maximum TCP Connections reached, you may want to" +
		" adjust max_tcp_connections")
}

// handler handles a single TCP Connection
func (t *TcpListener) handler(conn *net.TCPConn, id string) {
	// connection cleanup function
	defer func() {
		t.wg.Done()
		conn.Close()
		// log.Printf("Closed TCP Connection from %s", conn.RemoteAddr())
		// Add one connection potential back to channel when this one closes
		t.accept <- true
		t.forget(id)
	}()

	var n int
	scanner := bufio.NewScanner(conn)
	for {
		select {
		case <-t.done:
			return
		default:
			if !scanner.Scan() {
				return
			}
			n = len(scanner.Bytes())
			if n == 0 {
				continue
			}
			bufCopy := make([]byte, n)
			copy(bufCopy, scanner.Bytes())

			select {
			case t.in <- bufCopy:
			default:
				t.drops++
				if t.drops == 1 || t.drops%t.AllowedPendingMessages == 0 {
					log.Printf(dropwarn, t.drops)
				}
			}
		}
	}
}

// tcpParser parses the incoming tcp byte packets
func (t *TcpListener) tcpParser() error {
	defer t.wg.Done()

	var packet []byte
	var metrics []telegraf.Metric
	var err error
	for {
		select {
		case <-t.done:
			return nil
		case packet = <-t.in:
			if len(packet) == 0 {
				continue
			}
			metrics, err = t.parser.Parse(packet)
			if err == nil {
				t.storeMetrics(metrics)
			} else {
				t.malformed++
				if t.malformed == 1 || t.malformed%1000 == 0 {
					log.Printf(malformedwarn, t.malformed)
				}
			}
		}
	}
}

func (t *TcpListener) storeMetrics(metrics []telegraf.Metric) error {
	t.Lock()
	defer t.Unlock()
	for _, m := range metrics {
		t.acc.AddFields(m.Name(), m.Fields(), m.Tags(), m.Time())
	}
	return nil
}

// forget a TCP connection
func (t *TcpListener) forget(id string) {
	t.cleanup.Lock()
	defer t.cleanup.Unlock()
	delete(t.conns, id)
}

// remember a TCP connection
func (t *TcpListener) remember(id string, conn *net.TCPConn) {
	t.cleanup.Lock()
	defer t.cleanup.Unlock()
	t.conns[id] = conn
}

func init() {
	inputs.Add("tcp_listener", func() telegraf.Input {
		return &TcpListener{}
	})
}
