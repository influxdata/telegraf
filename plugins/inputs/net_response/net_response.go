package net_response

import (
	"bufio"
	"errors"
	"net"
	"net/textproto"
	"regexp"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
)

// NetResponses struct
type NetResponse struct {
	Address     string
	Timeout     internal.Duration
	ReadTimeout internal.Duration
	Send        string
	Expect      string
	Protocol    string
}

func (_ *NetResponse) Description() string {
	return "TCP or UDP 'ping' given url and collect response time in seconds"
}

var sampleConfig = `
  ## Protocol, must be "tcp" or "udp"
  ## NOTE: because the "udp" protocol does not respond to requests, it requires
  ## a send/expect string pair (see below).
  protocol = "tcp"
  ## Server address (default localhost)
  address = "localhost:80"
  ## Set timeout
  timeout = "1s"

  ## Set read timeout (only used if expecting a response)
  read_timeout = "1s"

  ## The following options are required for UDP checks. For TCP, they are
  ## optional. The plugin will send the given string to the server and then
  ## expect to receive the given 'expect' string back.
  ## string sent to the server
  # send = "ssh"
  ## expected string in answer
  # expect = "ssh"
`

func (_ *NetResponse) SampleConfig() string {
	return sampleConfig
}

func (n *NetResponse) TcpGather() (map[string]interface{}, error) {
	// Prepare fields
	fields := make(map[string]interface{})
	// Start Timer
	start := time.Now()
	// Connecting
	conn, err := net.DialTimeout("tcp", n.Address, n.Timeout.Duration)
	// Stop timer
	responseTime := time.Since(start).Seconds()
	// Handle error
	if err != nil {
		if e, ok := err.(net.Error); ok && e.Timeout() {
			fields["result_type"] = "timeout"
		} else {
			fields["result_type"] = "connection_failed"
		}
		return fields, nil
	}
	defer conn.Close()
	// Send string if needed
	if n.Send != "" {
		msg := []byte(n.Send)
		conn.Write(msg)
		// Stop timer
		responseTime = time.Since(start).Seconds()
	}
	// Read string if needed
	if n.Expect != "" {
		// Set read timeout
		conn.SetReadDeadline(time.Now().Add(n.ReadTimeout.Duration))
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
			fields["result_type"] = "read_failed"
		} else {
			// Looking for string in answer
			RegEx := regexp.MustCompile(`.*` + n.Expect + `.*`)
			find := RegEx.FindString(string(data))
			if find != "" {
				fields["result_type"] = "success"
				fields["string_found"] = true
			} else {
				fields["result_type"] = "string_mismatch"
				fields["string_found"] = false
			}
		}
	} else {
		fields["result_type"] = "success"
	}
	fields["response_time"] = responseTime
	return fields, nil
}

func (n *NetResponse) UdpGather() (map[string]interface{}, error) {
	// Prepare fields
	fields := make(map[string]interface{})
	// Start Timer
	start := time.Now()
	// Resolving
	udpAddr, err := net.ResolveUDPAddr("udp", n.Address)
	LocalAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	// Connecting
	conn, err := net.DialUDP("udp", LocalAddr, udpAddr)
	// Handle error
	if err != nil {
		fields["result_type"] = "connection_failed"
		return fields, nil
	}
	defer conn.Close()
	// Send string
	msg := []byte(n.Send)
	conn.Write(msg)
	// Read string
	// Set read timeout
	conn.SetReadDeadline(time.Now().Add(n.ReadTimeout.Duration))
	// Read
	buf := make([]byte, 1024)
	_, _, err = conn.ReadFromUDP(buf)
	// Stop timer
	responseTime := time.Since(start).Seconds()
	// Handle error
	if err != nil {
		fields["result_type"] = "read_failed"
		return fields, nil
	} else {
		// Looking for string in answer
		RegEx := regexp.MustCompile(`.*` + n.Expect + `.*`)
		find := RegEx.FindString(string(buf))
		if find != "" {
			fields["result_type"] = "success"
			fields["string_found"] = true
		} else {
			fields["result_type"] = "string_mismatch"
			fields["string_found"] = false
		}
	}
	fields["response_time"] = responseTime
	return fields, nil
}

func (n *NetResponse) Gather(acc telegraf.Accumulator) error {
	// Set default values
	if n.Timeout.Duration == 0 {
		n.Timeout.Duration = time.Second
	}
	if n.ReadTimeout.Duration == 0 {
		n.ReadTimeout.Duration = time.Second
	}
	// Check send and expected string
	if n.Protocol == "udp" && n.Send == "" {
		return errors.New("Send string cannot be empty")
	}
	if n.Protocol == "udp" && n.Expect == "" {
		return errors.New("Expected string cannot be empty")
	}
	// Prepare host and port
	host, port, err := net.SplitHostPort(n.Address)
	if err != nil {
		return err
	}
	if host == "" {
		n.Address = "localhost:" + port
	}
	if port == "" {
		return errors.New("Bad port")
	}
	// Prepare data
	tags := map[string]string{"server": host, "port": port}
	var fields map[string]interface{}
	// Gather data
	if n.Protocol == "tcp" {
		fields, err = n.TcpGather()
		tags["protocol"] = "tcp"
	} else if n.Protocol == "udp" {
		fields, err = n.UdpGather()
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
