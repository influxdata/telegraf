package connection

import (
	"bufio"
	"errors"
	"net"
	"net/textproto"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf/plugins/inputs"
)

// Connections struct
type Connection struct {
	Tcps []Tcp `toml:"tcp"`
	Udps []Udp `toml:"udp"`
}

// Tcp connection struct
type Tcp struct {
	Address     string
	Timeout     float64
	ReadTimeout float64
	Send        string
	Expect      string
}

// Udp connection struct
type Udp struct {
	Address     string
	Timeout     float64
	ReadTimeout float64
	Send        string
	Expect      string
}

func (_ *Connection) Description() string {
	return "Ping given url(s) and return statistics"
}

var sampleConfig = `
  [[inputs.connection.tcp]]
    // Server address (default IP localhost)
    address = "github.com:80"
    // Set timeout (default 1.0)
    timeout = 1.0
    // Set read timeout (default 1.0)
    read_timeout = 1.0
    // String sent to the server
    send = "ssh"
    // Expected string in answer
    expect = "ssh"

  [[inputs.connection.tcp]]
    address = ":80"

  [[inputs.connection.udp]]
    // Server address (default IP localhost)
    address = "github.com:80"
    // Set timeout (default 1.0)
    timeout = 1.0
    // Set read timeout (default 1.0)
    read_timeout = 1.0
    // String sent to the server
    send = "ssh"
    // Expected string in answer
    expect = "ssh"

  [[inputs.connection.udp]]
    address = "localhost:161"
    timeout = 2.0
`

func (_ *Connection) SampleConfig() string {
	return sampleConfig
}

func (t *Tcp) Gather() (map[string]interface{}, error) {
	// Prepare fields
	fields := make(map[string]interface{})
	// Start Timer
	start := time.Now()
	// Resolving
	tcpAddr, err := net.ResolveTCPAddr("tcp", t.Address)
	// Connecting
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	// Stop timer
	responseTime := time.Since(start).Seconds()
	// Handle error
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	// Send string if needed
	if t.Send != "" {
		msg := []byte(t.Send)
		conn.Write(msg)
		conn.CloseWrite()
		// Stop timer
		responseTime = time.Since(start).Seconds()
	}
	// Read string if needed
	if t.Expect != "" {
		// Set read timeout
		conn.SetReadDeadline(time.Now().Add(time.Duration(t.ReadTimeout) * time.Second))
		// Prepare reader
		reader := bufio.NewReader(conn)
		tp := textproto.NewReader(reader)
		// Read
		data, err := tp.ReadLine()
		// Stop timer
		responseTime = time.Since(start).Seconds()
		// Handle error
		if err != nil {
			fields["string_found"] = false
		} else {
			// Looking for string in answer
			RegEx := regexp.MustCompile(`.*` + t.Expect + `.*`)
			find := RegEx.FindString(string(data))
			if find != "" {
				fields["string_found"] = true
			} else {
				fields["string_found"] = false
			}
		}

	}
	fields["response_time"] = responseTime
	return fields, nil
}

func (u *Udp) Gather() (map[string]interface{}, error) {
	// Prepare fields
	fields := make(map[string]interface{})
	// Start Timer
	start := time.Now()
	// Resolving
	udpAddr, err := net.ResolveUDPAddr("udp", u.Address)
	LocalAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	// Connecting
	conn, err := net.DialUDP("udp", LocalAddr, udpAddr)
	defer conn.Close()
	// Handle error
	if err != nil {
		return nil, err
	}
	// Send string
	msg := []byte(u.Send)
	conn.Write(msg)
	// Read string
	// Set read timeout
	conn.SetReadDeadline(time.Now().Add(time.Duration(u.ReadTimeout) * time.Second))
	// Read
	buf := make([]byte, 1024)
	_, _, err = conn.ReadFromUDP(buf)
	// Stop timer
	responseTime := time.Since(start).Seconds()
	// Handle error
	if err != nil {
		return nil, err
	} else {
		// Looking for string in answer
		RegEx := regexp.MustCompile(`.*` + u.Expect + `.*`)
		find := RegEx.FindString(string(buf))
		if find != "" {
			fields["string_found"] = true
		} else {
			fields["string_found"] = false
		}
	}
	fields["response_time"] = responseTime
	return fields, nil
}

func (c *Connection) Gather(acc inputs.Accumulator) error {

	var wg sync.WaitGroup
	errorChannel := make(chan error, (len(c.Tcps)+len(c.Udps))*2)

	// Spin off a go routine for each TCP
	for _, tcp := range c.Tcps {
		wg.Add(1)
		go func(tcp Tcp, acc inputs.Accumulator) {
			defer wg.Done()
			// Set default Tcp values
			if tcp.Timeout == 0 {
				tcp.Timeout = 1.0
			}
			if tcp.ReadTimeout == 0 {
				tcp.ReadTimeout = 1.0
			}
			// Prepare host and port
			host, port, err := net.SplitHostPort(tcp.Address)
			if err != nil {
				errorChannel <- err
				return
			}
			if host == "" {
				tcp.Address = "localhost:" + port
			}
			if port == "" {
				errorChannel <- errors.New("Bad port")
				return
			}
			// Gather data
			fields, err := tcp.Gather()
			if err != nil {
				errorChannel <- err
				return
			}
			tags := map[string]string{"server": tcp.Address}
			// Add metrics
			acc.AddFields("tcp_connection", fields, tags)
		}(tcp, acc)
	}

	// Spin off a go routine for each UDP
	for _, udp := range c.Udps {
		wg.Add(1)
		go func(udp Udp, acc inputs.Accumulator) {
			defer wg.Done()
			// Check send and expected string
			if udp.Send == "" {
				errorChannel <- errors.New("Send string cannot be empty")
				return
			}
			if udp.Expect == "" {
				errorChannel <- errors.New("Expected string cannot be empty")
				return
			}
			// Set default Tcp values
			if udp.Timeout == 0 {
				udp.Timeout = 1.0
			}
			if udp.ReadTimeout == 0 {
				udp.ReadTimeout = 1.0
			}
			// Prepare host and port
			host, port, err := net.SplitHostPort(udp.Address)
			if err != nil {
				errorChannel <- err
				return
			}
			if host == "" {
				udp.Address = "localhost:" + port
			}
			if port == "" {
				errorChannel <- errors.New("Bad port")
				return
			}
			// Gather data
			fields, err := udp.Gather()
			if err != nil {
				errorChannel <- err
				return
			}
			tags := map[string]string{"server": udp.Address}
			// Add metrics
			acc.AddFields("udp_connection", fields, tags)
		}(udp, acc)
	}

	wg.Wait()
	close(errorChannel)

	// Get all errors and return them as one giant error
	errorStrings := []string{}
	for err := range errorChannel {
		errorStrings = append(errorStrings, err.Error())
	}

	if len(errorStrings) == 0 {
		return nil
	}
	return errors.New(strings.Join(errorStrings, "\n"))
}

func init() {
	inputs.Add("connection", func() inputs.Input {
		return &Connection{}
	})
}
