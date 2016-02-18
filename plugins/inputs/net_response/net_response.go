package net_response

import (
	"bufio"
	"errors"
	"net"
	"net/textproto"
	"regexp"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// NetResponses struct
type NetResponse struct {
	Address     string
	Timeout     float64
	ReadTimeout float64
	Send        string
	Expect      string
	Protocol    string
}

func (_ *NetResponse) Description() string {
	return "TCP or UDP 'ping' given url and collect response time in seconds"
}

var sampleConfig = `
  ## Protocol, must be "tcp" or "udp"
  protocol = "tcp"
  ## Server address (default localhost)
  address = "github.com:80"
  ## Set timeout (default 1.0 seconds)
  timeout = 1.0
  ## Set read timeout (default 1.0 seconds)
  read_timeout = 1.0
  ## Optional string sent to the server
  # send = "ssh"
  ## Optional expected string in answer
  # expect = "ssh"
`

func (_ *NetResponse) SampleConfig() string {
	return sampleConfig
}

func (t *NetResponse) TcpGather() (map[string]interface{}, error) {
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

func (u *NetResponse) UdpGather() (map[string]interface{}, error) {
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

func (c *NetResponse) Gather(acc telegraf.Accumulator) error {
	// Set default values
	if c.Timeout == 0 {
		c.Timeout = 1.0
	}
	if c.ReadTimeout == 0 {
		c.ReadTimeout = 1.0
	}
	// Check send and expected string
	if c.Protocol == "udp" && c.Send == "" {
		return errors.New("Send string cannot be empty")
	}
	if c.Protocol == "udp" && c.Expect == "" {
		return errors.New("Expected string cannot be empty")
	}
	// Prepare host and port
	host, port, err := net.SplitHostPort(c.Address)
	if err != nil {
		return err
	}
	if host == "" {
		c.Address = "localhost:" + port
	}
	if port == "" {
		return errors.New("Bad port")
	}
	// Prepare data
	tags := map[string]string{"host": host, "port": port}
	var fields map[string]interface{}
	// Gather data
	if c.Protocol == "tcp" {
		fields, err = c.TcpGather()
		tags["protocol"] = "tcp"
	} else if c.Protocol == "udp" {
		fields, err = c.UdpGather()
		tags["protocol"] = "udp"
	} else {
		return errors.New("Bad protocol")
	}
	if err != nil {
		return err
	}
	// Add metrics
	acc.AddFields("net_response", fields, tags)
	return nil
}

func init() {
	inputs.Add("net_response", func() telegraf.Input {
		return &NetResponse{}
	})
}
