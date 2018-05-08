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

// NetResponse struct
type NetResponse struct {
	Address     string
	Timeout     internal.Duration
	ReadTimeout internal.Duration
	Send        string
	Expect      string
	Protocol    string
}

var description = "TCP or UDP 'ping' given url and collect response time in seconds"

// Description will return a short string to explain what the plugin does.
func (*NetResponse) Description() string {
	return description
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

// SampleConfig will return a complete configuration example with details about each field.
func (*NetResponse) SampleConfig() string {
	return sampleConfig
}

// TCPGather will execute if there are TCP tests defined in the configuration.
// It will return a map[string]interface{} for fields and a map[string]string for tags
func (n *NetResponse) TCPGather() (tags map[string]string, fields map[string]interface{}) {
	// Prepare returns
	tags = make(map[string]string)
	fields = make(map[string]interface{})
	// Start Timer
	start := time.Now()
	// Connecting
	conn, err := net.DialTimeout("tcp", n.Address, n.Timeout.Duration)
	// Stop timer
	responseTime := time.Since(start).Seconds()
	// Handle error
	if err != nil {
		if e, ok := err.(net.Error); ok && e.Timeout() {
			tags["result_text"] = "timeout"
			fields["result_code"] = 1
			fields["result_type"] = "timeout"
		} else {
			tags["result_text"] = "connection_failed"
			fields["result_code"] = 1
			fields["result_type"] = "connection_failed"
		}
		return tags, fields
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
			tags["result_text"] = "read_failed"
			fields["result_code"] = 1
			fields["string_found"] = false
			tags["result_type"] = "read_failed"
			fields["success"] = 1
		} else {
			// Looking for string in answer
			RegEx := regexp.MustCompile(`.*` + n.Expect + `.*`)
			find := RegEx.FindString(string(data))
			if find != "" {
				tags["result_text"] = "success"
				fields["result_code"] = 0
				fields["result_type"] = "success"
				fields["string_found"] = true
			} else {
				tags["result_text"] = "string_mismatch"
				fields["result_code"] = 1
				fields["result_type"] = "string_mismatch"
				fields["string_found"] = false
			}
		}
	} else {
		tags["result_text"] = "success"
		fields["result_code"] = 0
		fields["result_type"] = "success"
	}
	fields["response_time"] = responseTime
	return tags, fields
}

// UDPGather will execute if there are UDP tests defined in the configuration.
// It will return a map[string]interface{} for fields and a map[string]string for tags
func (n *NetResponse) UDPGather() (tags map[string]string, fields map[string]interface{}) {
	// Prepare returns
	tags = make(map[string]string)
	fields = make(map[string]interface{})
	// Start Timer
	start := time.Now()
	// Resolving
	udpAddr, err := net.ResolveUDPAddr("udp", n.Address)
	LocalAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	// Connecting
	conn, err := net.DialUDP("udp", LocalAddr, udpAddr)
	// Handle error
	if err != nil {
		tags["result_text"] = "connection_failed"
		fields["result_code"] = 1
		fields["result_type"] = "connection_failed"
		return tags, fields
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
		tags["result_text"] = "read_failed"
		fields["result_code"] = 1
		fields["result_type"] = "read_failed"
		return tags, fields
	}

	// Looking for string in answer
	RegEx := regexp.MustCompile(`.*` + n.Expect + `.*`)
	find := RegEx.FindString(string(buf))
	if find != "" {
		tags["result_text"] = "success"
		fields["result_code"] = 0
		fields["result_type"] = "success"
		fields["string_found"] = true
	} else {
		tags["result_text"] = "string_mismatch"
		fields["result_code"] = 1
		fields["result_type"] = "string_mismatch"
		fields["string_found"] = false
	}

	fields["response_time"] = responseTime

	return tags, fields
}

// Gather is called by telegraf when the plugin is executed on its interval.
// It will call either UDPGather or TCPGather based on the configuration and
// also fill an Accumulator that is supplied.
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
	var returnTags map[string]string
	// Gather data
	if n.Protocol == "tcp" {
		returnTags, fields = n.TCPGather()
		tags["protocol"] = "tcp"
	} else if n.Protocol == "udp" {
		returnTags, fields = n.UDPGather()
		tags["protocol"] = "udp"
	} else {
		return errors.New("Bad protocol")
	}
	for key, value := range returnTags {
		tags[key] = value
	}
	// Merge the tags
	for k, v := range returnTags {
		tags[k] = v
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
