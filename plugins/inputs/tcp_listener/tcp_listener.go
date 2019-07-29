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
	"github.com/influxdata/telegraf/selfstat"
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

	MaxConnections     selfstat.Stat
	CurrentConnections selfstat.Stat
	TotalConnections   selfstat.Stat
	PacketsRecv        selfstat.Stat
	BytesRecv          selfstat.Stat
}

var dropwarn = "E! Error: tcp_listener message queue full. " +
	"We have dropped %d messages so far. " +
	"You may want to increase allowed_pending_messages in the config\n"

var malformedwarn = "E! tcp_listener has received %d malformed packets" +
	" thus far."

const sampleConfig = `
  # DEPRECATED: the TCP listener plugin has been deprecated in favor of the
  # socket_listener plugin
  # see https://github.com/influxdata/telegraf/tree/master/plugins/inputs/socket_listener
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

	log.Println("W! DEPRECATED: the TCP listener plugin has been deprecated " +
		"in favor of the socket_listener plugin " +
		"(https://github.com/influxdata/telegraf/tree/master/plugins/inputs/socket_listener)")

	tags := map[string]string{
		"address": t.ServiceAddress,
	}
	t.MaxConnections = selfstat.Register("tcp_listener", "max_connections", tags)
	t.MaxConnections.Set(int64(t.MaxTCPConnections))
	t.CurrentConnections = selfstat.Register("tcp_listener", "current_connections", tags)
	t.TotalConnections = selfstat.Register("tcp_listener", "total_connections", tags)
	t.PacketsRecv = selfstat.Register("tcp_listener", "packets_received", tags)
	t.BytesRecv = selfstat.Register("tcp_listener", "bytes_received", tags)

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
	log.Println("I! TCP server listening on: ", t.listener.Addr().String())

	t.wg.Add(2)
	go t.tcpListen()
	go t.tcpParser()

	log.Printf("I! Started TCP listener service on %s\n", t.ServiceAddress)
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
	log.Println("I! Stopped TCP listener service on ", t.ServiceAddress)
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
	log.Printf("I! Refused TCP Connection from %s", conn.RemoteAddr())
	log.Printf("I! WARNING: Maximum TCP Connections reached, you may want to" +
		" adjust max_tcp_connections")
}

// handler handles a single TCP Connection
func (t *TcpListener) handler(conn *net.TCPConn, id string) {
	t.CurrentConnections.Incr(1)
	t.TotalConnections.Incr(1)
	// connection cleanup function
	defer func() {
		t.wg.Done()
		conn.Close()
		// Add one connection potential back to channel when this one closes
		t.accept <- true
		t.forget(id)
		t.CurrentConnections.Incr(-1)
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
			t.BytesRecv.Incr(int64(n))
			t.PacketsRecv.Incr(1)
			bufCopy := make([]byte, n+1)
			copy(bufCopy, scanner.Bytes())
			bufCopy[n] = '\n'

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
			// drain input packets before finishing:
			if len(t.in) == 0 {
				return nil
			}
		case packet = <-t.in:
			if len(packet) == 0 {
				continue
			}
			metrics, err = t.parser.Parse(packet)
			if err == nil {
				for _, m := range metrics {
					t.acc.AddFields(m.Name(), m.Fields(), m.Tags(), m.Time())
				}
			} else {
				t.malformed++
				if t.malformed == 1 || t.malformed%1000 == 0 {
					log.Printf(malformedwarn, t.malformed)
				}
			}
		}
	}
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
		return &TcpListener{
			ServiceAddress:         ":8094",
			AllowedPendingMessages: 10000,
			MaxTCPConnections:      250,
		}
	})
}
