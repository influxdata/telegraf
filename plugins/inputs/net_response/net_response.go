//go:generate ../../../tools/readme_config_includer/generator
package net_response

import (
	"bufio"
	_ "embed"
	"errors"
	"fmt"
	"net"
	"net/textproto"
	"regexp"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal/choice"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type ResultType uint64

const (
	Success          ResultType = 0
	Timeout          ResultType = 1
	ConnectionFailed ResultType = 2
	ReadFailed       ResultType = 3
	StringMismatch   ResultType = 4
)

// NetResponse struct
type NetResponse struct {
	Address     string
	Timeout     config.Duration
	ReadTimeout config.Duration
	Send        string
	Expect      string
	Protocol    string
}

func (*NetResponse) SampleConfig() string {
	return sampleConfig
}

// TCPGather will execute if there are TCP tests defined in the configuration.
// It will return a map[string]interface{} for fields and a map[string]string for tags
func (n *NetResponse) TCPGather() (map[string]string, map[string]interface{}, error) {
	// Prepare returns
	tags := make(map[string]string)
	fields := make(map[string]interface{})
	// Start Timer
	start := time.Now()
	// Connecting
	conn, err := net.DialTimeout("tcp", n.Address, time.Duration(n.Timeout))
	// Stop timer
	responseTime := time.Since(start).Seconds()
	// Handle error
	if err != nil {
		var e net.Error
		if errors.As(err, &e) && e.Timeout() {
			setResult(Timeout, fields, tags, n.Expect)
		} else {
			setResult(ConnectionFailed, fields, tags, n.Expect)
		}
		return tags, fields, nil
	}
	defer conn.Close()
	// Send string if needed
	if n.Send != "" {
		msg := []byte(n.Send)
		if _, gerr := conn.Write(msg); gerr != nil {
			return nil, nil, gerr
		}
		// Stop timer
		responseTime = time.Since(start).Seconds()
	}
	// Read string if needed
	if n.Expect != "" {
		// Set read timeout
		if gerr := conn.SetReadDeadline(time.Now().Add(time.Duration(n.ReadTimeout))); gerr != nil {
			return nil, nil, gerr
		}
		// Prepare reader
		reader := bufio.NewReader(conn)
		tp := textproto.NewReader(reader)
		// Read
		data, err := tp.ReadLine()
		// Stop timer
		responseTime = time.Since(start).Seconds()
		// Handle error
		if err != nil {
			setResult(ReadFailed, fields, tags, n.Expect)
		} else {
			// Looking for string in answer
			regEx := regexp.MustCompile(`.*` + n.Expect + `.*`)
			find := regEx.FindString(data)
			if find != "" {
				setResult(Success, fields, tags, n.Expect)
			} else {
				setResult(StringMismatch, fields, tags, n.Expect)
			}
		}
	} else {
		setResult(Success, fields, tags, n.Expect)
	}
	fields["response_time"] = responseTime
	return tags, fields, nil
}

// UDPGather will execute if there are UDP tests defined in the configuration.
// It will return a map[string]interface{} for fields and a map[string]string for tags
func (n *NetResponse) UDPGather() (map[string]string, map[string]interface{}, error) {
	// Prepare returns
	tags := make(map[string]string)
	fields := make(map[string]interface{})
	// Start Timer
	start := time.Now()
	// Resolving
	udpAddr, err := net.ResolveUDPAddr("udp", n.Address)
	// Handle error
	if err != nil {
		setResult(ConnectionFailed, fields, tags, n.Expect)
		return tags, fields, nil //nolint:nilerr // error encoded in result
	}
	// Connecting
	conn, err := net.DialUDP("udp", nil, udpAddr)
	// Handle error
	if err != nil {
		setResult(ConnectionFailed, fields, tags, n.Expect)
		return tags, fields, nil //nolint:nilerr // error encoded in result
	}
	defer conn.Close()
	// Send string
	msg := []byte(n.Send)
	if _, gerr := conn.Write(msg); gerr != nil {
		return nil, nil, gerr
	}
	// Read string
	// Set read timeout
	if gerr := conn.SetReadDeadline(time.Now().Add(time.Duration(n.ReadTimeout))); gerr != nil {
		return nil, nil, gerr
	}
	// Read
	buf := make([]byte, 1024)
	_, _, err = conn.ReadFromUDP(buf)
	// Stop timer
	responseTime := time.Since(start).Seconds()
	// Handle error
	if err != nil {
		setResult(ReadFailed, fields, tags, n.Expect)
		return tags, fields, nil //nolint:nilerr // error encoded in result
	}

	// Looking for string in answer
	regEx := regexp.MustCompile(`.*` + n.Expect + `.*`)
	find := regEx.FindString(string(buf))
	if find != "" {
		setResult(Success, fields, tags, n.Expect)
	} else {
		setResult(StringMismatch, fields, tags, n.Expect)
	}

	fields["response_time"] = responseTime

	return tags, fields, nil
}

// Init performs one time setup of the plugin and returns an error if the
// configuration is invalid.
func (n *NetResponse) Init() error {
	// Set default values
	if n.Timeout == 0 {
		n.Timeout = config.Duration(time.Second)
	}
	if n.ReadTimeout == 0 {
		n.ReadTimeout = config.Duration(time.Second)
	}
	// Check send and expected string
	if n.Protocol == "udp" && n.Send == "" {
		return errors.New("send string cannot be empty")
	}
	if n.Protocol == "udp" && n.Expect == "" {
		return errors.New("expected string cannot be empty")
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
		return errors.New("bad port in config option address")
	}

	if err := choice.Check(n.Protocol, []string{"tcp", "udp"}); err != nil {
		return fmt.Errorf("config option protocol: %w", err)
	}

	return nil
}

// Gather is called by telegraf when the plugin is executed on its interval.
// It will call either UDPGather or TCPGather based on the configuration and
// also fill an Accumulator that is supplied.
func (n *NetResponse) Gather(acc telegraf.Accumulator) error {
	// Prepare host and port
	host, port, err := net.SplitHostPort(n.Address)
	if err != nil {
		return err
	}

	// Prepare data
	tags := map[string]string{"server": host, "port": port}
	var fields map[string]interface{}
	var returnTags map[string]string

	// Gather data
	switch n.Protocol {
	case "tcp":
		returnTags, fields, err = n.TCPGather()
		if err != nil {
			return err
		}
		tags["protocol"] = "tcp"
	case "udp":
		returnTags, fields, err = n.UDPGather()
		if err != nil {
			return err
		}
		tags["protocol"] = "udp"
	}

	// Merge the tags
	for k, v := range returnTags {
		tags[k] = v
	}
	// Add metrics
	acc.AddFields("net_response", fields, tags)
	return nil
}

func setResult(result ResultType, fields map[string]interface{}, tags map[string]string, expect string) {
	var tag string
	switch result {
	case Success:
		tag = "success"
	case Timeout:
		tag = "timeout"
	case ConnectionFailed:
		tag = "connection_failed"
	case ReadFailed:
		tag = "read_failed"
	case StringMismatch:
		tag = "string_mismatch"
	}

	tags["result"] = tag
	fields["result_code"] = uint64(result)

	// deprecated in 1.7; use result tag
	fields["result_type"] = tag

	// deprecated in 1.4; use result tag
	if expect != "" {
		fields["string_found"] = result == Success
	}
}

func init() {
	inputs.Add("net_response", func() telegraf.Input {
		return &NetResponse{}
	})
}
