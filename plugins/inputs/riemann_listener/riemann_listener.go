package riemann_listener

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf/metric"

	"github.com/gogo/protobuf/proto"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	tlsint "github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
	riemanngo "github.com/riemann/riemann-go-client"
	riemangoProto "github.com/riemann/riemann-go-client/proto"
)

type RiemannSocketListener struct {
	ServiceAddress  string             `toml:"service_address"`
	MaxConnections  int                `toml:"max_connections"`
	ReadBufferSize  internal.Size      `toml:"read_buffer_size"`
	ReadTimeout     *internal.Duration `toml:"read_timeout"`
	KeepAlivePeriod *internal.Duration `toml:"keep_alive_period"`
	SocketMode      string             `toml:"socket_mode"`
	tlsint.ServerConfig

	wg sync.WaitGroup

	Log telegraf.Logger

	telegraf.Accumulator
}
type setReadBufferer interface {
	SetReadBuffer(bytes int) error
}

type riemannListener struct {
	net.Listener
	*RiemannSocketListener

	sockType string

	connections    map[string]net.Conn
	connectionsMtx sync.Mutex
}

func (rsl *riemannListener) listen(ctx context.Context) {
	rsl.connections = map[string]net.Conn{}

	wg := sync.WaitGroup{}

	select {
	case <-ctx.Done():
		rsl.closeAllConnections()
		wg.Wait()
		return
	default:
		for {
			c, err := rsl.Accept()
			if err != nil {
				if !strings.HasSuffix(err.Error(), ": use of closed network connection") {
					rsl.Log.Error(err.Error())
				}
				break
			}

			if rsl.ReadBufferSize.Size > 0 {
				if srb, ok := c.(setReadBufferer); ok {
					srb.SetReadBuffer(int(rsl.ReadBufferSize.Size))
				} else {
					rsl.Log.Warnf("Unable to set read buffer on a %s socket", rsl.sockType)
				}
			}

			rsl.connectionsMtx.Lock()
			if rsl.MaxConnections > 0 && len(rsl.connections) >= rsl.MaxConnections {
				rsl.connectionsMtx.Unlock()
				c.Close()
				continue
			}
			rsl.connections[c.RemoteAddr().String()] = c
			rsl.connectionsMtx.Unlock()

			if err := rsl.setKeepAlive(c); err != nil {
				rsl.Log.Errorf("Unable to configure keep alive %q: %s", rsl.ServiceAddress, err.Error())
			}

			wg.Add(1)
			go func() {
				defer wg.Done()
				rsl.read(c)
			}()
		}
		rsl.closeAllConnections()
		wg.Wait()
	}
}

func (rsl *riemannListener) closeAllConnections() {
	rsl.connectionsMtx.Lock()
	for _, c := range rsl.connections {
		c.Close()
	}
	rsl.connectionsMtx.Unlock()
}

func (rsl *riemannListener) setKeepAlive(c net.Conn) error {
	if rsl.KeepAlivePeriod == nil {
		return nil
	}
	tcpc, ok := c.(*net.TCPConn)
	if !ok {
		return fmt.Errorf("cannot set keep alive on a %s socket", strings.SplitN(rsl.ServiceAddress, "://", 2)[0])
	}
	if rsl.KeepAlivePeriod.Duration == 0 {
		return tcpc.SetKeepAlive(false)
	}
	if err := tcpc.SetKeepAlive(true); err != nil {
		return err
	}
	return tcpc.SetKeepAlivePeriod(rsl.KeepAlivePeriod.Duration)
}

func (rsl *riemannListener) removeConnection(c net.Conn) {
	rsl.connectionsMtx.Lock()
	delete(rsl.connections, c.RemoteAddr().String())
	rsl.connectionsMtx.Unlock()
}

//Utilities

/*
readMessages will read Riemann messages in binary format
from the TCP connection. byte Array p size will depend on the size
of the riemann  message as sent by the cleint
*/
func readMessages(r io.Reader, p []byte) error {
	for len(p) > 0 {
		n, err := r.Read(p)
		p = p[n:]
		if err != nil {
			return err
		}
	}
	return nil
}

func checkError(err error) {
	log.Println("The error is")
	if err != nil {
		log.Println(err.Error())
	}
}

func (rsl *riemannListener) read(conn net.Conn) {
	defer rsl.removeConnection(conn)
	defer conn.Close()
	var err error

	for {
		if rsl.ReadTimeout != nil && rsl.ReadTimeout.Duration > 0 {
			conn.SetDeadline(time.Now().Add(rsl.ReadTimeout.Duration))
		}

		messagePb := &riemangoProto.Msg{}
		var header uint32
		// First obtain the size of the riemann event from client and acknowledge
		if err = binary.Read(conn, binary.BigEndian, &header); err != nil {
			if err.Error() != "EOF" {
				rsl.Log.Debugf("Failed to read header")
				riemannReturnErrorResponse(conn, err.Error())
				return
			}
			return
		}
		data := make([]byte, header)

		if err = readMessages(conn, data); err != nil {
			rsl.Log.Debugf("Failed to read body: %s", err.Error())
			riemannReturnErrorResponse(conn, "Failed to read body")
			return
		}
		if err = proto.Unmarshal(data, messagePb); err != nil {
			rsl.Log.Debugf("Failed to unmarshal: %s", err.Error())
			riemannReturnErrorResponse(conn, "Failed to unmarshal")
			return
		}
		riemannEvents := riemanngo.ProtocolBuffersToEvents(messagePb.Events)

		for _, m := range riemannEvents {
			if m.Service == "" {
				riemannReturnErrorResponse(conn, "No Service Name")
				return
			}
			tags := make(map[string]string)
			fieldValues := map[string]interface{}{}
			for _, tag := range m.Tags {
				tags[strings.ReplaceAll(tag, " ", "_")] = tag
			}
			tags["Host"] = m.Host
			tags["Description"] = m.Description
			tags["State"] = m.State
			fieldValues["Metric"] = m.Metric
			fieldValues["TTL"] = m.TTL.Seconds()
			singleMetric, err := metric.New(m.Service, tags, fieldValues, m.Time, telegraf.Untyped)
			if err != nil {
				rsl.Log.Debugf("Could not create metric for service %s at %s", m.Service, m.Time.String())
				riemannReturnErrorResponse(conn, "Could not create metric")
				return
			}

			rsl.AddMetric(singleMetric)
		}
		riemannReturnResponse(conn)

	}

}

func riemannReturnResponse(conn net.Conn) {
	t := true
	message := new(riemangoProto.Msg)
	message.Ok = &t
	returnData, err := proto.Marshal(message)
	if err != nil {
		checkError(err)
		return
	}
	b := new(bytes.Buffer)
	if err = binary.Write(b, binary.BigEndian, uint32(len(returnData))); err != nil {
		checkError(err)
	}
	// send the msg length
	if _, err = conn.Write(b.Bytes()); err != nil {
		checkError(err)
	}
	if _, err = conn.Write(returnData); err != nil {
		checkError(err)
	}
}

func riemannReturnErrorResponse(conn net.Conn, errorMessage string) {
	t := false
	message := new(riemangoProto.Msg)
	message.Ok = &t
	message.Error = &errorMessage
	returnData, err := proto.Marshal(message)
	if err != nil {
		checkError(err)
		return
	}
	b := new(bytes.Buffer)
	if err = binary.Write(b, binary.BigEndian, uint32(len(returnData))); err != nil {
		checkError(err)
	}
	// send the msg length
	if _, err = conn.Write(b.Bytes()); err != nil {
		checkError(err)
	}
	if _, err = conn.Write(returnData); err != nil {
		log.Println("Somethign")
		checkError(err)
	}
}

func (rsl *RiemannSocketListener) Description() string {
	return "Riemann protobuff listener."
}

func (rsl *RiemannSocketListener) SampleConfig() string {
	return `
  ## URL to listen on. 
  ## Default is "tcp://:5555"
  # service_address = "tcp://:8094"
  # service_address = "tcp://127.0.0.1:http"
  # service_address = "tcp4://:8094"
  # service_address = "tcp6://:8094"
  # service_address = "tcp6://[2001:db8::1]:8094"

  ## Maximum number of concurrent connections.
  ## 0 (default) is unlimited.
  # max_connections = 1024
  ## Read timeout.
  ## 0 (default) is unlimited.
  # read_timeout = "30s"
  ## Optional TLS configuration.
  # tls_cert = "/etc/telegraf/cert.pem"
  # tls_key  = "/etc/telegraf/key.pem"
  ## Enables client authentication if set.
  # tls_allowed_cacerts = ["/etc/telegraf/clientca.pem"]
  ## Maximum socket buffer size (in bytes when no unit specified).
  # read_buffer_size = "64KiB"
  ## Period between keep alive probes.
  ## 0 disables keep alive probes.
  ## Defaults to the OS configuration.
  # keep_alive_period = "5m"
`
}

func (rsl *RiemannSocketListener) Gather(_ telegraf.Accumulator) error {
	return nil
}

func (rsl *RiemannSocketListener) Start(acc telegraf.Accumulator) error {
	ctx, cancelFunc := context.WithCancel(context.Background())
	go processOsSignals(cancelFunc)
	rsl.Accumulator = acc
	if rsl.ServiceAddress == "" {
		rsl.Log.Warnf("Using default service_address tcp://:5555")
		rsl.ServiceAddress = "tcp://:5555"
	}
	spl := strings.SplitN(rsl.ServiceAddress, "://", 2)
	if len(spl) != 2 {
		return fmt.Errorf("invalid service address: %s", rsl.ServiceAddress)
	}

	protocol := spl[0]
	addr := spl[1]

	switch protocol {
	case "tcp", "tcp4", "tcp6":
		tlsCfg, err := rsl.ServerConfig.TLSConfig()
		if err != nil {
			return err
		}

		var l net.Listener
		if tlsCfg == nil {
			l, err = net.Listen(protocol, addr)
		} else {
			l, err = tls.Listen(protocol, addr, tlsCfg)
		}
		if err != nil {
			return err
		}

		rsl.Log.Infof("Listening on %s://%s", protocol, l.Addr())

		rsl := &riemannListener{
			Listener:              l,
			RiemannSocketListener: rsl,
			sockType:              spl[0],
		}

		rsl.wg = sync.WaitGroup{}
		rsl.wg.Add(1)
		go func() {
			defer rsl.wg.Done()
			rsl.listen(ctx)

		}()
	default:
		return fmt.Errorf("unknown protocol '%s' in '%s'", protocol, rsl.ServiceAddress)
	}

	return nil
}

// Handle cancellations from the process
func processOsSignals(cancelFunc context.CancelFunc) {
	signalChan := make(chan os.Signal)
	signal.Notify(signalChan, os.Interrupt)
	for {
		sig := <-signalChan
		switch sig {
		case os.Interrupt:
			log.Println("Signal SIGINT is received, probably due to `Ctrl-C`, exiting ...")
			cancelFunc()
			return
		}
	}

}

func (rsl *RiemannSocketListener) Stop() {
	rsl.wg.Done()
	rsl.wg.Wait()
}

func newRiemannSocketListener() *RiemannSocketListener {
	return &RiemannSocketListener{}
}

func init() {
	inputs.Add("riemann_listener", func() telegraf.Input { return newRiemannSocketListener() })
}
